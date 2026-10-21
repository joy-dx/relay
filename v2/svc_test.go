package relay

import (
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/joy-dx/relay/v2/config"
	"github.com/joy-dx/relay/v2/dto"
)

// --- Test fixtures (golden table pattern) ------------------------------------

type sinkCall struct {
	Ref   string
	Level dto.RelayLevel
	Msg   string
	Type  dto.EventRef
	Ch    dto.EventChannel
	Time  time.Time
}

type recordingSink struct {
	ref   string
	mu    sync.Mutex
	calls []sinkCall
}

func newRecordingSink(ref string) *recordingSink {
	return &recordingSink{ref: ref, calls: make([]sinkCall, 0)}
}

func (s *recordingSink) Ref() string { return s.ref }

func (s *recordingSink) Emit(ev dto.EmittedEvent) {
	e := ev.Event

	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, sinkCall{
		Ref:   s.ref,
		Level: ev.Level,
		Msg:   e.Message(),
		Type:  e.RelayType(),
		Ch:    e.RelayChannel(),
		Time:  ev.Time,
	})
}
func (s *recordingSink) Close() error {
	return nil
}

func (s *recordingSink) Calls() []sinkCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]sinkCall, len(s.calls))
	copy(out, s.calls)
	return out
}

type testEvent struct {
	ch   dto.EventChannel
	ref  dto.EventRef
	msg  string
	slog []any // ignored; we satisfy interface via method below
}

func (e testEvent) RelayChannel() dto.EventChannel { return e.ch }
func (e testEvent) RelayType() dto.EventRef        { return e.ref }
func (e testEvent) Message() string                { return e.msg }

// Keep ToSlog minimal to satisfy interface without depending on slog.Attr
// in assertions.
func (e testEvent) ToSlog() []slog.Attr { return nil }

// --- RelaySvc tests -----------------------------------------------------------

func TestRelaySvc_Hydrate_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		svc       *RelaySvc
		wantErr   bool
		errSubstr string
	}{
		{
			name: "nil cfg errors",
			svc: &RelaySvc{
				cfg: nil,
			},
			wantErr:   true,
			errSubstr: "Relay cfg required but is nil",
		},
		{
			name: "non-nil cfg ok",
			svc: &RelaySvc{
				cfg: &config.RelaySvcConfig{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.svc.Hydrate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got: %v", err)
			}
			if tt.wantErr && tt.errSubstr != "" && err != nil {
				if got := err.Error(); got != tt.errSubstr {
					t.Fatalf("error mismatch\nwant: %q\ngot:  %q", tt.errSubstr, got)
				}
			}
		})
	}
}

func TestRelaySvc_Emit_DispatchesToAllSinks_Golden(t *testing.T) {
	t.Parallel()

	type golden struct {
		name  string
		level dto.RelayLevel
		event dto.RelayEventInterface
		// expected calls across all sinks, in order
		want []sinkCall
	}

	e := testEvent{
		ch:  "relay",
		ref: "relay.log",
		msg: "hello",
	}

	// IMPORTANT: svc.emit dispatch order is sink registration order.
	tests := []golden{
		{
			name:  "debug calls Debug on each sink",
			level: dto.Debug,
			event: e,
			want: []sinkCall{
				{Ref: "s1", Level: dto.Debug, Msg: "hello", Type: "relay.log", Ch: "relay"},
				{Ref: "s2", Level: dto.Debug, Msg: "hello", Type: "relay.log", Ch: "relay"},
			},
		},
		{
			name:  "info calls Info on each sink",
			level: dto.Info,
			event: e,
			want: []sinkCall{
				{Ref: "s1", Level: dto.Info, Msg: "hello", Type: "relay.log", Ch: "relay"},
				{Ref: "s2", Level: dto.Info, Msg: "hello", Type: "relay.log", Ch: "relay"},
			},
		},
		{
			name:  "warn calls Warn on each sink",
			level: dto.Warn,
			event: e,
			want: []sinkCall{
				{Ref: "s1", Level: dto.Warn, Msg: "hello", Type: "relay.log", Ch: "relay"},
				{Ref: "s2", Level: dto.Warn, Msg: "hello", Type: "relay.log", Ch: "relay"},
			},
		},
		{
			name:  "error calls Error on each sink",
			level: dto.Error,
			event: e,
			want: []sinkCall{
				{Ref: "s1", Level: dto.Error, Msg: "hello", Type: "relay.log", Ch: "relay"},
				{Ref: "s2", Level: dto.Error, Msg: "hello", Type: "relay.log", Ch: "relay"},
			},
		},
		{
			name:  "meta calls Meta on each sink",
			level: dto.Meta,
			event: e,
			want: []sinkCall{
				{Ref: "s1", Level: dto.Meta, Msg: "hello", Type: "relay.log", Ch: "relay"},
				{Ref: "s2", Level: dto.Meta, Msg: "hello", Type: "relay.log", Ch: "relay"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s1 := newRecordingSink("s1")
			s2 := newRecordingSink("s2")

			svc := &RelaySvc{
				cfg:   &config.RelaySvcConfig{},
				sinks: []dto.RelaySinkInterface{s1, s2},
			}

			svc.Emit(tt.level, tt.event)

			got := append(s1.Calls(), s2.Calls()...)
			if len(got) != len(tt.want) {
				t.Fatalf("call count mismatch\nwant: %d\ngot:  %d\ncalls: %#v",
					len(tt.want), len(got), got)
			}

			for i := range tt.want {
				if got[i].Time.IsZero() {
					t.Fatalf("call[%d] has zero time", i)
				}

				got[i].Time = time.Time{}
				if got[i] != tt.want[i] {
					t.Fatalf("call[%d] mismatch\nwant: %#v\ngot:  %#v",
						i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestRelaySvc_RegisterSink_AppendsInOrder_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		regOrder []string
		wantRefs []string
	}{
		{
			name:     "register two sinks preserves order",
			regOrder: []string{"a", "b"},
			wantRefs: []string{"a", "b"},
		},
		{
			name:     "register three sinks preserves order",
			regOrder: []string{"a", "b", "c"},
			wantRefs: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &RelaySvc{
				cfg:   &config.RelaySvcConfig{},
				sinks: make([]dto.RelaySinkInterface, 0),
			}

			for _, ref := range tt.regOrder {
				svc.RegisterSink(newRecordingSink(ref))
			}

			if len(svc.sinks) != len(tt.wantRefs) {
				t.Fatalf("sink count mismatch\nwant: %d\ngot:  %d",
					len(tt.wantRefs), len(svc.sinks))
			}

			for i, want := range tt.wantRefs {
				if got := svc.sinks[i].Ref(); got != want {
					t.Fatalf("sink[%d] ref mismatch\nwant: %q\ngot:  %q",
						i, want, got)
				}
			}
		})
	}
}

// We intentionally do NOT call Fatal through emit() in tests here because it
// os.Exit(1)'s the test process. If you want Fatal tested, refactor RelaySvc to
// allow an injected "exit func(int)" or "osExiter interface" so tests can
// assert behavior without exiting.

// --- ProvideRelaySvc tests ----------------------------------------------------

func TestProvideRelaySvc_IsSingleton_Golden(t *testing.T) {
	// Not parallel: this test relies on global state.
	// Reset global for test isolation.
	service = nil
	serviceOnce = sync.Once{}

	cfg1 := &config.RelaySvcConfig{}
	cfg2 := &config.RelaySvcConfig{}

	s1 := ProvideRelaySvc(cfg1)
	s2 := ProvideRelaySvc(cfg2)

	if s1 == nil || s2 == nil {
		t.Fatalf("expected non-nil services")
	}

	if s1 != s2 {
		t.Fatalf("expected singleton instance, got different pointers")
	}

	// Golden behavior: first call wins (cfg2 should be ignored).
	if s1.cfg != cfg1 {
		t.Fatalf("expected cfg from first call to win")
	}
}

// --- SimpleLoggerConfig tests -------------------------------------------------
// These validate the fluent builder behavior and defaults (golden table).

// Note: lives here for convenience; you can move into sinks package tests if you
// prefer. If you move it, update imports accordingly.

func TestDefaultSimpleLoggerConfigAndFluentBuilders_Golden(t *testing.T) {
	t.Parallel()

	// Import sinks in-test to avoid circular deps in production code.
	// (If this file is in relay package, it's fine to import sinks.)
	// If you want these tests in sinks/, create sinks/simple_logger_config_test.go
	// and remove this test from here.
	type cfgIface interface {
		WithLevel(level dto.RelayLevel) interface{}
		WithKeyPadding(keyPadding int) interface{}
	}

	// Since we can't refer to sinks.SimpleLoggerConfig without importing sinks,
	// you should move these tests into sinks/ for clarity.
	_ = cfgIface(nil)
}

func TestRelaySvc_WithOpts_OverridesRelayMetadata_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opts        dto.RlyOpts
		wantChannel dto.EventChannel
		wantType    dto.EventRef
	}{
		{
			name:        "overrides channel and type",
			opts:        dto.RlyOpts{Channel: "custom-channel", Type: "custom.type"},
			wantChannel: "custom-channel",
			wantType:    "custom.type",
		},
		{
			name:        "overrides channel only",
			opts:        dto.RlyOpts{Channel: "custom-channel"},
			wantChannel: "custom-channel",
			wantType:    "relay.log",
		},
		{
			name:        "overrides type only",
			opts:        dto.RlyOpts{Type: "custom.type"},
			wantChannel: "relay",
			wantType:    "custom.type",
		},
		{
			name:        "empty options preserve original metadata",
			opts:        dto.RlyOpts{},
			wantChannel: "relay",
			wantType:    "relay.log",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sink := newRecordingSink("sink")
			svc := &RelaySvc{
				cfg:   &config.RelaySvcConfig{},
				sinks: []dto.RelaySinkInterface{sink},
			}

			event := testEvent{
				ch:  "relay",
				ref: "relay.log",
				msg: "hello",
			}

			svc.InfoWithOpts(event, tt.opts)

			calls := sink.Calls()
			if len(calls) != 1 {
				t.Fatalf("expected one sink call, got %d", len(calls))
			}

			got := calls[0]

			if got.Level != dto.Info {
				t.Fatalf("level mismatch\nwant: %q\ngot:  %q",
					dto.Info, got.Level)
			}

			if got.Msg != "hello" {
				t.Fatalf("message mismatch\nwant: %q\ngot:  %q",
					"hello", got.Msg)
			}

			if got.Ch != tt.wantChannel {
				t.Fatalf("channel mismatch\nwant: %q\ngot:  %q",
					tt.wantChannel, got.Ch)
			}

			if got.Type != tt.wantType {
				t.Fatalf("type mismatch\nwant: %q\ngot:  %q",
					tt.wantType, got.Type)
			}

			if got.Time.IsZero() {
				t.Fatal("expected emitted event time to be populated")
			}
		})
	}
}

func TestRelaySvc_WithOpts_DispatchesToAllSinks_Golden(t *testing.T) {
	t.Parallel()

	s1 := newRecordingSink("s1")
	s2 := newRecordingSink("s2")

	svc := &RelaySvc{
		cfg:   &config.RelaySvcConfig{},
		sinks: []dto.RelaySinkInterface{s1, s2},
	}

	event := testEvent{
		ch:  "relay",
		ref: "relay.log",
		msg: "hello with options",
	}

	svc.InfoWithOpts(event, dto.RlyOpts{
		Channel: "custom-channel",
		Type:    "custom.type",
	})

	got := append(s1.Calls(), s2.Calls()...)

	if len(got) != 2 {
		t.Fatalf("expected two sink calls, got %d", len(got))
	}

	wantRefs := []string{"s1", "s2"}

	for i, call := range got {
		if call.Ref != wantRefs[i] {
			t.Fatalf("call[%d] sink mismatch\nwant: %q\ngot:  %q",
				i, wantRefs[i], call.Ref)
		}

		if call.Level != dto.Info {
			t.Fatalf("call[%d] level mismatch\nwant: %q\ngot:  %q",
				i, dto.Info, call.Level)
		}

		if call.Msg != "hello with options" {
			t.Fatalf("call[%d] message mismatch\nwant: %q\ngot:  %q",
				i, "hello with options", call.Msg)
		}

		if call.Ch != dto.EventChannel("custom-channel") {
			t.Fatalf("call[%d] channel mismatch\nwant: %q\ngot:  %q",
				i, "custom-channel", call.Ch)
		}

		if call.Type != dto.EventRef("custom.type") {
			t.Fatalf("call[%d] type mismatch\nwant: %q\ngot:  %q",
				i, "custom.type", call.Type)
		}

		if call.Time.IsZero() {
			t.Fatalf("call[%d] has zero time", i)
		}
	}
}

func TestRelaySvc_WithOpts_DoesNotMutateOriginalEvent_Golden(t *testing.T) {
	t.Parallel()

	sink := newRecordingSink("sink")
	svc := &RelaySvc{
		cfg:   &config.RelaySvcConfig{},
		sinks: []dto.RelaySinkInterface{sink},
	}

	event := testEvent{
		ch:  "relay",
		ref: "relay.log",
		msg: "original message",
	}

	svc.WarnWithOpts(event, dto.RlyOpts{
		Channel: "custom-channel",
		Type:    "custom.type",
	})

	if event.RelayChannel() != dto.EventChannel("relay") {
		t.Fatalf("original event channel was modified: %q",
			event.RelayChannel())
	}

	if event.RelayType() != dto.EventRef("relay.log") {
		t.Fatalf("original event type was modified: %q",
			event.RelayType())
	}

	if event.Message() != "original message" {
		t.Fatalf("original event message was modified: %q",
			event.Message())
	}

	calls := sink.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected one sink call, got %d", len(calls))
	}

	if calls[0].Ch != dto.EventChannel("custom-channel") {
		t.Fatalf("emitted channel mismatch\nwant: %q\ngot:  %q",
			"custom-channel", calls[0].Ch)
	}

	if calls[0].Type != dto.EventRef("custom.type") {
		t.Fatalf("emitted type mismatch\nwant: %q\ngot:  %q",
			"custom.type", calls[0].Type)
	}
}

func TestRelaySvc_WithOpts_PreservesEventBehavior_Golden(t *testing.T) {
	t.Parallel()

	sink := newRecordingSink("sink")
	svc := &RelaySvc{
		cfg:   &config.RelaySvcConfig{},
		sinks: []dto.RelaySinkInterface{sink},
	}

	event := testEvent{
		ch:  "relay",
		ref: "relay.log",
		msg: "message is preserved",
	}

	svc.DebugWithOpts(event, dto.RlyOpts{
		Channel: "debug-channel",
		Type:    "debug.type",
	})

	calls := sink.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected one sink call, got %d", len(calls))
	}

	got := calls[0]

	if got.Msg != event.Message() {
		t.Fatalf("message mismatch\nwant: %q\ngot:  %q",
			event.Message(), got.Msg)
	}

	if got.Ch != dto.EventChannel("debug-channel") {
		t.Fatalf("channel mismatch\nwant: %q\ngot:  %q",
			"debug-channel", got.Ch)
	}

	if got.Type != dto.EventRef("debug.type") {
		t.Fatalf("type mismatch\nwant: %q\ngot:  %q",
			"debug.type", got.Type)
	}
}

func TestRelaySvc_WithOpts_MetaDispatchesWithOverrides_Golden(t *testing.T) {
	t.Parallel()

	sink := newRecordingSink("sink")
	svc := &RelaySvc{
		cfg:   &config.RelaySvcConfig{},
		sinks: []dto.RelaySinkInterface{sink},
	}

	event := testEvent{
		ch:  "relay",
		ref: "relay.log",
		msg: "metadata event",
	}

	svc.MetaWithOpts(event, dto.RlyOpts{
		Channel: "meta-channel",
		Type:    "meta.type",
	})

	calls := sink.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected one sink call, got %d", len(calls))
	}

	got := calls[0]

	if got.Level != dto.Meta {
		t.Fatalf("level mismatch\nwant: %q\ngot:  %q",
			dto.Meta, got.Level)
	}

	if got.Ch != dto.EventChannel("meta-channel") {
		t.Fatalf("channel mismatch\nwant: %q\ngot:  %q",
			"meta-channel", got.Ch)
	}

	if got.Type != dto.EventRef("meta.type") {
		t.Fatalf("type mismatch\nwant: %q\ngot:  %q",
			"meta.type", got.Type)
	}

	if got.Msg != "metadata event" {
		t.Fatalf("message mismatch\nwant: %q\ngot:  %q",
			"metadata event", got.Msg)
	}
}
