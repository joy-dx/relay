package sinks

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/joy-dx/relay/dto"
)

type testEventRef string

type testEvent struct {
	typ string
	msg string
}

func (e testEvent) Message() string {
	return e.msg
}

func (e testEvent) RelayChannel() dto.EventChannel {
	return "test"
}

func (e testEvent) RelayType() dto.EventRef {
	return dto.EventRef(e.typ)
}

func (e testEvent) ToSlog() []slog.Attr {
	return []slog.Attr{}
}

type cloneableTestEvent struct {
	testEvent
}

func (e cloneableTestEvent) Clone() dto.RelayEventInterface {
	return cloneableTestEvent{
		testEvent: testEvent{
			typ: e.typ,
			msg: e.msg,
		},
	}
}

type recorderAction struct {
	name string
	run  func(*testing.T, *RecorderSink)
}

type goldenRecordedEvent struct {
	level dto.RelayLevel
	typ   string
	msg   string
}

type recorderGolden struct {
	name              string
	cfg               *RecorderConfig
	actions           []recorderAction
	wantEvents        []goldenRecordedEvent
	wantRecorded      uint64
	wantDroppedMin    uint64
	wantDroppedExact  *uint64
	wantSegments      int
	wantEventCount    int
	wantReplayErr     error
	replayErrAt       int
	afterCloseActions []recorderAction
}

func TestRecorderSink_Golden(t *testing.T) {
	t.Parallel()

	exactZero := uint64(0)

	cases := []recorderGolden{
		{
			name: "records_debug_info_warn_error_fatal_meta_in_order",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 8,
				InputBuffer: 16,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actDebug("dbg-type", "debug message"),
				actInfo("info-type", "info message"),
				actWarn("warn-type", "warn message"),
				actError("err-type", "error message"),
				actFatal("fatal-type", "fatal message"),
				actMeta("meta-type", "meta message"),
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Debug, typ: "dbg-type", msg: "debug message"},
				{level: dto.Info, typ: "info-type", msg: "info message"},
				{level: dto.Warn, typ: "warn-type", msg: "warn message"},
				{level: dto.Error, typ: "err-type", msg: "error message"},
				{level: dto.Fatal, typ: "fatal-type", msg: "fatal message"},
				{level: dto.Meta, typ: "meta-type", msg: "meta message"},
			},
			wantRecorded:     6,
			wantDroppedExact: &exactZero,
			wantSegments:     1,
			wantEventCount:   6,
		},
		{
			name: "respects_level_filter_for_debug_info_warn_but_not_error_fatal_meta",
			cfg: &RecorderConfig{
				Level:       dto.Warn,
				SegmentSize: 8,
				InputBuffer: 16,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actDebug("dbg", "drop debug"),
				actInfo("info", "drop info"),
				actWarn("warn", "keep warn"),
				actError("err", "keep error"),
				actFatal("fatal", "keep fatal"),
				actMeta("meta", "keep meta"),
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Warn, typ: "warn", msg: "keep warn"},
				{level: dto.Error, typ: "err", msg: "keep error"},
				{level: dto.Fatal, typ: "fatal", msg: "keep fatal"},
				{level: dto.Meta, typ: "meta", msg: "keep meta"},
			},
			wantRecorded:     4,
			wantDroppedExact: &exactZero,
			wantSegments:     1,
			wantEventCount:   4,
		},
		{
			name: "creates_multiple_segments_when_segment_size_is_reached",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 2,
				InputBuffer: 16,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actInfo("a", "1"),
				actInfo("b", "2"),
				actInfo("c", "3"),
				actInfo("d", "4"),
				actInfo("e", "5"),
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Info, typ: "a", msg: "1"},
				{level: dto.Info, typ: "b", msg: "2"},
				{level: dto.Info, typ: "c", msg: "3"},
				{level: dto.Info, typ: "d", msg: "4"},
				{level: dto.Info, typ: "e", msg: "5"},
			},
			wantRecorded:     5,
			wantDroppedExact: &exactZero,
			wantSegments:     3,
			wantEventCount:   5,
		},
		{
			name: "evicts_oldest_segments_when_max_segments_is_reached",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 2,
				MaxSegments: 2,
				InputBuffer: 16,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actInfo("a", "1"),
				actInfo("b", "2"),
				actInfo("c", "3"),
				actInfo("d", "4"),
				actInfo("e", "5"),
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Info, typ: "c", msg: "3"},
				{level: dto.Info, typ: "d", msg: "4"},
				{level: dto.Info, typ: "e", msg: "5"},
			},
			wantRecorded:   5,
			wantDroppedMin: 2,
			wantSegments:   2,
			wantEventCount: 3,
		},
		{
			name: "non_blocking_mode_drops_when_input_buffer_is_full",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 8,
				InputBuffer: 1,
				BlockOnFull: false,
			},
			actions: []recorderAction{
				{
					name: "burst writes",
					run: func(t *testing.T, s *RecorderSink) {
						for i := 0; i < 2000; i++ {
							s.Info(testEvent{
								typ: fmt.Sprintf("type-%d", i),
								msg: fmt.Sprintf("msg-%d", i),
							})
						}
					},
				},
			},
			wantRecorded:   0, // checked with eventual assertions below instead
			wantDroppedMin: 1,
		},
		{
			name: "snapshot_returns_independent_copy",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 4,
				InputBuffer: 8,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actInfo("alpha", "one"),
				actInfo("beta", "two"),
				{
					name: "snapshot independence",
					run: func(t *testing.T, s *RecorderSink) {
						waitForRecorderCount(t, s, 2)

						snap1 := s.Snapshot()
						if len(snap1) != 2 {
							t.Fatalf("expected snapshot size 2, got %d", len(snap1))
						}

						s.Info(testEvent{typ: "gamma", msg: "three"})
						waitForRecorderCount(t, s, 3)

						if len(snap1) != 2 {
							t.Fatalf(
								"expected old snapshot to remain size 2, got %d",
								len(snap1),
							)
						}
					},
				},
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Info, typ: "alpha", msg: "one"},
				{level: dto.Info, typ: "beta", msg: "two"},
				{level: dto.Info, typ: "gamma", msg: "three"},
			},
			wantRecorded:     3,
			wantDroppedExact: &exactZero,
			wantSegments:     1,
			wantEventCount:   3,
		},
		{
			name: "replay_visits_all_events_in_order",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 2,
				InputBuffer: 8,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actDebug("d1", "m1"),
				actInfo("i1", "m2"),
				actWarn("w1", "m3"),
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Debug, typ: "d1", msg: "m1"},
				{level: dto.Info, typ: "i1", msg: "m2"},
				{level: dto.Warn, typ: "w1", msg: "m3"},
			},
			wantRecorded:     3,
			wantDroppedExact: &exactZero,
			wantSegments:     2,
			wantEventCount:   3,
		},
		{
			name: "replay_stops_and_returns_callback_error",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 2,
				InputBuffer: 8,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actInfo("a", "1"),
				actInfo("b", "2"),
				actInfo("c", "3"),
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Info, typ: "a", msg: "1"},
				{level: dto.Info, typ: "b", msg: "2"},
				{level: dto.Info, typ: "c", msg: "3"},
			},
			wantRecorded:     3,
			wantDroppedExact: &exactZero,
			wantSegments:     2,
			wantEventCount:   3,
			wantReplayErr:    errReplayStop,
			replayErrAt:      1,
		},
		{
			name: "reset_clears_all_segments_and_events",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 2,
				InputBuffer: 8,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actInfo("a", "1"),
				actInfo("b", "2"),
				{
					name: "reset",
					run: func(t *testing.T, s *RecorderSink) {
						waitForRecorderCount(t, s, 2)
						if err := s.Reset(); err != nil {
							t.Fatalf("reset failed: %v", err)
						}
					},
				},
			},
			wantEvents:        []goldenRecordedEvent{},
			wantRecorded:      2,
			wantDroppedExact:  &exactZero,
			wantSegments:      0,
			wantEventCount:    0,
			afterCloseActions: nil,
		},
		{
			name: "close_is_idempotent_and_post_close_writes_are_ignored",
			cfg: &RecorderConfig{
				Level:       dto.Debug,
				SegmentSize: 4,
				InputBuffer: 8,
				BlockOnFull: true,
			},
			actions: []recorderAction{
				actInfo("before", "close"),
				{
					name: "close twice",
					run: func(t *testing.T, s *RecorderSink) {
						waitForRecorderCount(t, s, 1)

						if err := s.Close(); err != nil {
							t.Fatalf("first close failed: %v", err)
						}
						if err := s.Close(); err != nil {
							t.Fatalf("second close failed: %v", err)
						}
					},
				},
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Info, typ: "before", msg: "close"},
			},
			wantRecorded:     1,
			wantDroppedExact: &exactZero,
			wantSegments:     1,
			wantEventCount:   1,
			afterCloseActions: []recorderAction{
				actInfo("after", "close"),
				actError("after-err", "close"),
			},
		},
		{
			name: "clone_on_record_uses_cloner_when_available",
			cfg: &RecorderConfig{
				Level:         dto.Debug,
				SegmentSize:   4,
				InputBuffer:   8,
				BlockOnFull:   true,
				CloneOnRecord: true,
			},
			actions: []recorderAction{
				{
					name: "record cloneable event",
					run: func(t *testing.T, s *RecorderSink) {
						ev := cloneableTestEvent{
							testEvent: testEvent{
								typ: "clone",
								msg: "original",
							},
						}
						s.Info(ev)
						waitForRecorderCount(t, s, 1)
					},
				},
			},
			wantEvents: []goldenRecordedEvent{
				{level: dto.Info, typ: "clone", msg: "original"},
			},
			wantRecorded:     1,
			wantDroppedExact: &exactZero,
			wantSegments:     1,
			wantEventCount:   1,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sink := NewRecorderSink(tc.cfg)
			defer func() {
				_ = sink.Close()
			}()

			for _, action := range tc.actions {
				action.run(t, sink)
			}

			if tc.name == "non_blocking_mode_drops_when_input_buffer_is_full" {
				waitForCondition(t, 2*time.Second, func() bool {
					return sink.RecordedCount()+sink.DroppedCount() > 0
				})

				recorded := sink.RecordedCount()
				dropped := sink.DroppedCount()

				if recorded == 0 && dropped == 0 {
					t.Fatalf("expected some recorded or dropped events")
				}
				if dropped < tc.wantDroppedMin {
					t.Fatalf(
						"expected dropped >= %d, got %d",
						tc.wantDroppedMin,
						dropped,
					)
				}

				return
			}

			waitUntilDrainedStable(t, sink)

			assertSnapshotMatchesGolden(t, sink.Snapshot(), tc.wantEvents)

			if got := sink.RecordedCount(); got != tc.wantRecorded {
				t.Fatalf(
					"RecordedCount() mismatch: want %d, got %d",
					tc.wantRecorded,
					got,
				)
			}

			if tc.wantDroppedExact != nil {
				if got := sink.DroppedCount(); got != *tc.wantDroppedExact {
					t.Fatalf(
						"DroppedCount() mismatch: want %d, got %d",
						*tc.wantDroppedExact,
						got,
					)
				}
			} else if tc.wantDroppedMin > 0 {
				if got := sink.DroppedCount(); got < tc.wantDroppedMin {
					t.Fatalf(
						"DroppedCount() mismatch: want at least %d, got %d",
						tc.wantDroppedMin,
						got,
					)
				}
			}

			if tc.wantSegments > 0 || len(tc.wantEvents) == 0 {
				if got := sink.SegmentCount(); got != tc.wantSegments {
					t.Fatalf(
						"SegmentCount() mismatch: want %d, got %d",
						tc.wantSegments,
						got,
					)
				}
			}

			if got := sink.EventCount(); got != tc.wantEventCount {
				t.Fatalf(
					"EventCount() mismatch: want %d, got %d",
					tc.wantEventCount,
					got,
				)
			}

			assertReplayBehavior(t, sink, tc)

			if len(tc.afterCloseActions) > 0 {
				before := sink.Snapshot()

				for _, action := range tc.afterCloseActions {
					action.run(t, sink)
				}

				after := sink.Snapshot()
				assertRecordedEventsEqual(t, before, after)
			}
		})
	}
}

func TestRecorderSink_ConcurrentWriters(t *testing.T) {
	t.Parallel()

	sink := NewRecorderSink(&RecorderConfig{
		Level:       dto.Debug,
		SegmentSize: 64,
		InputBuffer: 2048,
		BlockOnFull: true,
	})
	defer func() {
		_ = sink.Close()
	}()

	const goroutines = 16
	const perGoroutine = 250

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		g := g
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				sink.Info(testEvent{
					typ: fmt.Sprintf("g-%d", g),
					msg: fmt.Sprintf("m-%d", i),
				})
			}
		}()
	}

	wg.Wait()
	waitForRecorderCount(t, sink, goroutines*perGoroutine)

	if got := sink.RecordedCount(); got != uint64(goroutines*perGoroutine) {
		t.Fatalf(
			"RecordedCount() mismatch: want %d, got %d",
			goroutines*perGoroutine,
			got,
		)
	}

	if got := sink.DroppedCount(); got != 0 {
		t.Fatalf("DroppedCount() mismatch: want 0, got %d", got)
	}

	if got := sink.EventCount(); got != goroutines*perGoroutine {
		t.Fatalf(
			"EventCount() mismatch: want %d, got %d",
			goroutines*perGoroutine,
			got,
		)
	}
}

func TestRecorderSink_ReplayNilCallback(t *testing.T) {
	t.Parallel()

	sink := NewRecorderSink(&RecorderConfig{
		Level:       dto.Debug,
		SegmentSize: 4,
		InputBuffer: 8,
		BlockOnFull: true,
	})
	defer func() {
		_ = sink.Close()
	}()

	sink.Info(testEvent{typ: "x", msg: "y"})
	waitForRecorderCount(t, sink, 1)

	err := sink.Replay(nil)
	if err == nil {
		t.Fatal("expected error for nil replay callback")
	}
}

func TestRecorderSink_ResetAfterCloseReturnsError(t *testing.T) {
	t.Parallel()

	sink := NewRecorderSink(&RecorderConfig{
		Level:       dto.Debug,
		SegmentSize: 4,
		InputBuffer: 8,
		BlockOnFull: true,
	})

	if err := sink.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	if err := sink.Reset(); err == nil {
		t.Fatal("expected reset after close to fail")
	}
}

var errReplayStop = errors.New("stop replay")

func actDebug(typ string, msg string) recorderAction {
	return recorderAction{
		name: "debug",
		run: func(t *testing.T, s *RecorderSink) {
			s.Debug(testEvent{typ: typ, msg: msg})
		},
	}
}

func actInfo(typ string, msg string) recorderAction {
	return recorderAction{
		name: "info",
		run: func(t *testing.T, s *RecorderSink) {
			s.Info(testEvent{typ: typ, msg: msg})
		},
	}
}

func actWarn(typ string, msg string) recorderAction {
	return recorderAction{
		name: "warn",
		run: func(t *testing.T, s *RecorderSink) {
			s.Warn(testEvent{typ: typ, msg: msg})
		},
	}
}

func actError(typ string, msg string) recorderAction {
	return recorderAction{
		name: "error",
		run: func(t *testing.T, s *RecorderSink) {
			s.Error(testEvent{typ: typ, msg: msg})
		},
	}
}

func actFatal(typ string, msg string) recorderAction {
	return recorderAction{
		name: "fatal",
		run: func(t *testing.T, s *RecorderSink) {
			s.Fatal(testEvent{typ: typ, msg: msg})
		},
	}
}

func actMeta(typ string, msg string) recorderAction {
	return recorderAction{
		name: "meta",
		run: func(t *testing.T, s *RecorderSink) {
			s.Meta(testEvent{typ: typ, msg: msg})
		},
	}
}

func assertSnapshotMatchesGolden(
	t *testing.T,
	got []RecordedEvent,
	want []goldenRecordedEvent,
) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("snapshot length mismatch: want %d, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i].Level != want[i].level {
			t.Fatalf(
				"snapshot[%d].Level mismatch: want %v, got %v",
				i,
				want[i].level,
				got[i].Level,
			)
		}
		if string(got[i].Event.RelayType()) != want[i].typ {
			t.Fatalf(
				"snapshot[%d].EventRef mismatch: want %q, got %q",
				i,
				want[i].typ,
				got[i].Event.RelayType(),
			)
		}
		if got[i].Event.Message() != want[i].msg {
			t.Fatalf(
				"snapshot[%d].Message mismatch: want %q, got %q",
				i,
				want[i].msg,
				got[i].Event.Message(),
			)
		}
	}
}

func assertReplayBehavior(
	t *testing.T,
	sink *RecorderSink,
	tc recorderGolden,
) {
	t.Helper()

	var replayed []goldenRecordedEvent
	idx := 0

	err := sink.Replay(func(re RecordedEvent) error {
		if tc.wantReplayErr != nil && idx == tc.replayErrAt {
			return tc.wantReplayErr
		}

		replayed = append(replayed, goldenRecordedEvent{
			level: re.Level,
			typ:   string(re.Event.RelayType()),
			msg:   re.Event.Message(),
		})
		idx++

		return nil
	})

	if tc.wantReplayErr != nil {
		if !errors.Is(err, tc.wantReplayErr) {
			t.Fatalf(
				"Replay() error mismatch: want %v, got %v",
				tc.wantReplayErr,
				err,
			)
		}

		expectedCount := tc.replayErrAt
		if len(replayed) != expectedCount {
			t.Fatalf(
				"Replay() callback count mismatch before error: want %d, got %d",
				expectedCount,
				len(replayed),
			)
		}

		for i := 0; i < expectedCount; i++ {
			if replayed[i] != tc.wantEvents[i] {
				t.Fatalf(
					"Replay() item mismatch at %d: want %+v, got %+v",
					i,
					tc.wantEvents[i],
					replayed[i],
				)
			}
		}

		return
	}

	if err != nil {
		t.Fatalf("Replay() unexpected error: %v", err)
	}

	if len(replayed) != len(tc.wantEvents) {
		t.Fatalf(
			"Replay() count mismatch: want %d, got %d",
			len(tc.wantEvents),
			len(replayed),
		)
	}

	for i := range tc.wantEvents {
		if replayed[i] != tc.wantEvents[i] {
			t.Fatalf(
				"Replay() item mismatch at %d: want %+v, got %+v",
				i,
				tc.wantEvents[i],
				replayed[i],
			)
		}
	}
}

func assertRecordedEventsEqual(
	t *testing.T,
	want []RecordedEvent,
	got []RecordedEvent,
) {
	t.Helper()

	if len(want) != len(got) {
		t.Fatalf("length mismatch: want %d, got %d", len(want), len(got))
	}

	for i := range want {
		if want[i].Level != got[i].Level {
			t.Fatalf(
				"event[%d].Level mismatch: want %v, got %v",
				i,
				want[i].Level,
				got[i].Level,
			)
		}
		if want[i].Event.Message() != got[i].Event.Message() {
			t.Fatalf(
				"event[%d].Message mismatch: want %q, got %q",
				i,
				want[i].Event.Message(),
				got[i].Event.Message(),
			)
		}
		if want[i].Event.RelayType() != got[i].Event.RelayType() {
			t.Fatalf(
				"event[%d].EventRef mismatch: want %q, got %q",
				i,
				want[i].Event.RelayType(),
				got[i].Event.RelayType(),
			)
		}
	}
}

func waitForRecorderCount(t *testing.T, sink *RecorderSink, want int) {
	t.Helper()

	waitForCondition(t, 2*time.Second, func() bool {
		return sink.EventCount() == want
	})
}

func waitUntilDrainedStable(t *testing.T, sink *RecorderSink) {
	t.Helper()

	var last int
	var stable int

	waitForCondition(t, 2*time.Second, func() bool {
		cur := sink.EventCount()
		if cur == last {
			stable++
		} else {
			stable = 0
		}
		last = cur
		return stable >= 5
	})
}

func waitForCondition(
	t *testing.T,
	timeout time.Duration,
	cond func() bool,
) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition not satisfied before timeout")
}
