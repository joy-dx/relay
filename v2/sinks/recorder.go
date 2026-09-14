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

	input       chan dto.EmittedEvent
	wg          sync.WaitGroup
	closed      atomic.Bool
	eventFilter *EventFilter

	mu       sync.RWMutex
	segments [][]dto.EmittedEvent

	recorded atomic.Uint64
	dropped  atomic.Uint64
}

func NewRecorderSink(cfg *RecorderConfig) *RecorderSink {
	if cfg == nil {
		cfg = &RecorderConfig{}
	}
	filter := NewEventFilter(cfg.EventFilterConfig)

	if cfg.SegmentSize <= 0 {
		cfg.SegmentSize = RECORDER_DEFAULT_SEGMENT_SIZE
	}
	if cfg.InputBuffer <= 0 {
		cfg.InputBuffer = RECORDER_DEFAULT_INPUT_BUFFER_SIZE
	}

	s := &RecorderSink{
		cfg:         cfg,
		input:       make(chan dto.EmittedEvent, cfg.InputBuffer),
		eventFilter: filter,
		segments:    make([][]dto.EmittedEvent, 0, 8),
	}

	s.wg.Add(1)
	go s.run()

	return s
}

func (s *RecorderSink) Ref() string {
	return RecorderSinkRef
}

func (s *RecorderSink) Emit(ev dto.EmittedEvent) {
	switch ev.Level {
	case dto.Debug, dto.Info, dto.Warn:
		if !s.eventFilter.IsLevelEnabled(ev.Level) {
			return
		}
		if !s.eventFilter.IsChannelAllowed(ev.Event.RelayChannel()) {
			return
		}
		if !s.eventFilter.IsEventAllowed(ev.Event.RelayType()) {
			return
		}
	}
	s.record(ev)
}

func (s *RecorderSink) record(ev dto.EmittedEvent) {
	if s.closed.Load() {
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
