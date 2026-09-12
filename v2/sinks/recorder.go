package sinks

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/joy-dx/relay/v2/dto"
)

const RecorderSinkRef = "recorder"

var ErrRecorderClosed = errors.New("recorder sink is closed")

type RecorderSink struct {
	cfg *RecorderConfig

	input                chan dto.EmittedEvent
	wg                   sync.WaitGroup
	closed               atomic.Bool
	relayChannelsAllowed map[dto.EventChannel]struct{}
	relayEventsAllowed   map[dto.EventRef]struct{}

	mu       sync.RWMutex
	segments [][]dto.EmittedEvent

	recorded atomic.Uint64
	dropped  atomic.Uint64
}

func NewRecorderSink(cfg *RecorderConfig) *RecorderSink {
	if cfg == nil {
		cfg = &RecorderConfig{}
	}

	eventsFilterSet := make(map[dto.EventRef]struct{}, len(cfg.EventsAllowed))
	for _, t := range cfg.EventsAllowed {
		eventsFilterSet[t] = struct{}{}
	}
	channelsFilterSet := make(map[dto.EventChannel]struct{}, len(cfg.ChannelsAllowed))
	for _, t := range cfg.ChannelsAllowed {
		channelsFilterSet[t] = struct{}{}
	}

	if cfg.SegmentSize <= 0 {
		cfg.SegmentSize = RECORDER_DEFAULT_SEGMENT_SIZE
	}
	if cfg.InputBuffer <= 0 {
		cfg.InputBuffer = RECORDER_DEFAULT_INPUT_BUFFER_SIZE
	}

	s := &RecorderSink{
		cfg:                  cfg,
		input:                make(chan dto.EmittedEvent, cfg.InputBuffer),
		relayChannelsAllowed: channelsFilterSet,
		relayEventsAllowed:   eventsFilterSet,
		segments:             make([][]dto.EmittedEvent, 0, 8),
	}

	s.wg.Add(1)
	go s.run()

	return s
}

func (s *RecorderSink) Ref() string {
	return RecorderSinkRef
}

func (s *RecorderSink) Debug(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Debug) {
		return
	}
	s.record(dto.Debug, ev)
}

func (s *RecorderSink) Info(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Info) {
		return
	}
	s.record(dto.Info, ev)
}

func (s *RecorderSink) Warn(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Warn) {
		return
	}
	s.record(dto.Warn, ev)
}

func (s *RecorderSink) Error(ev dto.EmittedEvent) {
	s.record(dto.Error, ev)
}

func (s *RecorderSink) Fatal(ev dto.EmittedEvent) {
	s.record(dto.Fatal, ev)
}

func (s *RecorderSink) Meta(ev dto.EmittedEvent) {
	s.record(dto.Meta, ev)
}

func (s *RecorderSink) record(level dto.RelayLevel, ev dto.EmittedEvent) {
	if s.closed.Load() {
		return
	}

	if len(s.relayEventsAllowed) > 0 {
		if _, allEventsAllowed := s.relayEventsAllowed["*"]; !allEventsAllowed {
			if _, ok := s.relayEventsAllowed[ev.Event.RelayType()]; !ok {
				return
			}
		}
	}

	if len(s.relayChannelsAllowed) > 0 {
		if _, allChannelsAllowed := s.relayChannelsAllowed["*"]; !allChannelsAllowed {
			if _, ok := s.relayChannelsAllowed[ev.Event.RelayChannel()]; !ok {
				return
			}
		}
	}

	if s.cfg.Filter != nil && !s.cfg.Filter(ev) {
		return
	}

	preppedEvent := s.prepareEvent(ev)

	if s.cfg.BlockOnFull {
		defer func() {
			if recover() != nil {
				s.dropped.Add(1)
			}
		}()
		s.input <- preppedEvent
		return
	}

	select {
	case s.input <- preppedEvent:
	default:
		s.dropped.Add(1)
	}
}

func (s *RecorderSink) prepareEvent(
	ev dto.EmittedEvent,
) dto.EmittedEvent {
	if !s.cfg.CloneOnRecord {
		return ev
	}

	return ev.Clone()
}

func (s *RecorderSink) run() {
	defer s.wg.Done()

	for re := range s.input {
		s.mu.Lock()

		if len(s.segments) == 0 ||
			len(s.segments[len(s.segments)-1]) >= s.cfg.SegmentSize {
			s.segments = append(
				s.segments,
				make([]dto.EmittedEvent, 0, s.cfg.SegmentSize),
			)

			if s.cfg.MaxSegments > 0 && len(s.segments) > s.cfg.MaxSegments {
				evicted := uint64(len(s.segments[0]))
				s.segments[0] = nil
				s.segments = s.segments[1:]
				if evicted > 0 {
					s.dropped.Add(evicted)
				}
			}
		}

		last := len(s.segments) - 1
		s.segments[last] = append(s.segments[last], re)
		s.recorded.Add(1)

		s.mu.Unlock()
	}
}

func (s *RecorderSink) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(s.input)
	s.wg.Wait()
	return nil
}

func (s *RecorderSink) Snapshot() []dto.EmittedEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, seg := range s.segments {
		total += len(seg)
	}

	out := make([]dto.EmittedEvent, 0, total)
	for _, seg := range s.segments {
		out = append(out, seg...)
	}

	return out
}

func (s *RecorderSink) Replay(fn func(dto.EmittedEvent) error) error {
	if fn == nil {
		return errors.New("replay callback is nil")
	}

	s.mu.RLock()
	segments := make([][]dto.EmittedEvent, len(s.segments))
	for i, seg := range s.segments {
		cp := make([]dto.EmittedEvent, len(seg))
		copy(cp, seg)
		segments[i] = cp
	}
	s.mu.RUnlock()

	for _, seg := range segments {
		for _, re := range seg {
			if err := fn(re); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *RecorderSink) RecordedCount() uint64 {
	return s.recorded.Load()
}

func (s *RecorderSink) DroppedCount() uint64 {
	return s.dropped.Load()
}

func (s *RecorderSink) SegmentCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.segments)
}

func (s *RecorderSink) EventCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, seg := range s.segments {
		total += len(seg)
	}
	return total
}

func (s *RecorderSink) Reset() error {
	if s.closed.Load() {
		return fmt.Errorf("cannot reset closed recorder sink")
	}

	s.mu.Lock()
	for i := range s.segments {
		s.segments[i] = nil
	}
	s.segments = s.segments[:0]
	s.mu.Unlock()

	return nil
}
