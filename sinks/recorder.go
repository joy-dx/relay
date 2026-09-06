package sinks

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/joy-dx/relay/dto"
)

const RecorderSinkRef = "recorder"

var ErrRecorderClosed = errors.New("recorder sink is closed")

type RecordedEvent struct {
	Level dto.RelayLevel
	Event dto.RelayEventInterface
}

type RecorderSink struct {
	cfg *RecorderConfig

	input  chan RecordedEvent
	wg     sync.WaitGroup
	closed atomic.Bool

	mu       sync.RWMutex
	segments [][]RecordedEvent

	recorded atomic.Uint64
	dropped  atomic.Uint64
}

func NewRecorderSink(cfg *RecorderConfig) *RecorderSink {
	if cfg == nil {
		cfg = &RecorderConfig{}
	}
	if cfg.SegmentSize <= 0 {
		cfg.SegmentSize = 512
	}
	if cfg.InputBuffer <= 0 {
		cfg.InputBuffer = 1024
	}

	s := &RecorderSink{
		cfg:      cfg,
		input:    make(chan RecordedEvent, cfg.InputBuffer),
		segments: make([][]RecordedEvent, 0, 8),
	}

	s.wg.Add(1)
	go s.run()

	return s
}

func (s *RecorderSink) Ref() string {
	return RecorderSinkRef
}

func (s *RecorderSink) Debug(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Debug) {
		return
	}
	s.record(dto.Debug, e)
}

func (s *RecorderSink) Info(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Info) {
		return
	}
	s.record(dto.Info, e)
}

func (s *RecorderSink) Warn(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Warn) {
		return
	}
	s.record(dto.Warn, e)
}

func (s *RecorderSink) Error(e dto.RelayEventInterface) {
	s.record(dto.Error, e)
}

func (s *RecorderSink) Fatal(e dto.RelayEventInterface) {
	s.record(dto.Fatal, e)
}

func (s *RecorderSink) Meta(e dto.RelayEventInterface) {
	s.record(dto.Meta, e)
}

func (s *RecorderSink) record(level dto.RelayLevel, e dto.RelayEventInterface) {
	if s.closed.Load() {
		return
	}

	re := RecordedEvent{
		Level: level,
		Event: s.prepareEvent(e),
	}

	if s.cfg.BlockOnFull {
		defer func() {
			if recover() != nil {
				s.dropped.Add(1)
			}
		}()
		s.input <- re
		return
	}

	select {
	case s.input <- re:
	default:
		s.dropped.Add(1)
	}
}

func (s *RecorderSink) prepareEvent(
	e dto.RelayEventInterface,
) dto.RelayEventInterface {
	if !s.cfg.CloneOnRecord {
		return e
	}

	if cloner, ok := e.(interface {
		Clone() dto.RelayEventInterface
	}); ok {
		return cloner.Clone()
	}

	return e
}

func (s *RecorderSink) run() {
	defer s.wg.Done()

	for re := range s.input {
		s.mu.Lock()

		if len(s.segments) == 0 ||
			len(s.segments[len(s.segments)-1]) >= s.cfg.SegmentSize {
			s.segments = append(
				s.segments,
				make([]RecordedEvent, 0, s.cfg.SegmentSize),
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

func (s *RecorderSink) Snapshot() []RecordedEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, seg := range s.segments {
		total += len(seg)
	}

	out := make([]RecordedEvent, 0, total)
	for _, seg := range s.segments {
		out = append(out, seg...)
	}

	return out
}

func (s *RecorderSink) Replay(fn func(RecordedEvent) error) error {
	if fn == nil {
		return errors.New("replay callback is nil")
	}

	s.mu.RLock()
	segments := make([][]RecordedEvent, len(s.segments))
	for i, seg := range s.segments {
		cp := make([]RecordedEvent, len(seg))
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
