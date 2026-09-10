package sinks

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/joy-dx/relay/dto"
)

type msgEvent struct {
	msg string
}

func (e msgEvent) RelayChannel() dto.EventChannel { return "relay" }
func (e msgEvent) RelayType() dto.EventRef        { return "relay.log" }
func (e msgEvent) Message() string                { return e.msg }

func (e msgEvent) ToSlog() []slog.Attr { return nil }

func TestSimpleLoggerSink_LevelGating_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfgLevel   dto.RelayLevel
		call       func(s *SimpleLoggerSink)
		wantOutput bool
	}{
		{
			name:     "debug suppressed at info",
			cfgLevel: dto.Info,
			call: func(s *SimpleLoggerSink) {
				s.Debug(msgEvent{msg: "d"})
			},
			wantOutput: false,
		},
		{
			name:     "info printed at info",
			cfgLevel: dto.Info,
			call: func(s *SimpleLoggerSink) {
				s.Info(msgEvent{msg: "i"})
			},
			wantOutput: true,
		},
		{
			name:     "warn printed at info",
			cfgLevel: dto.Info,
			call: func(s *SimpleLoggerSink) {
				s.Warn(msgEvent{msg: "writer"})
			},
			wantOutput: true,
		},
		{
			name:     "error always prints",
			cfgLevel: dto.Fatal,
			call: func(s *SimpleLoggerSink) {
				s.Error(msgEvent{msg: "e"})
			},
			wantOutput: true,
		},
		{
			name:     "fatal always prints",
			cfgLevel: dto.Fatal,
			call: func(s *SimpleLoggerSink) {
				s.Fatal(msgEvent{msg: "f"})
			},
			wantOutput: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			cfg := DefaultSimpleLoggerConfig()
			cfg.WithLevel(tt.cfgLevel).WithKeyPadding(0).WithWriter(&buf)

			sink := NewSimpleLogger(&cfg)

			tt.call(sink)
			out := buf.String()

			if tt.wantOutput && out == "" {
				t.Fatalf("expected output, got empty")
			}
			if !tt.wantOutput && out != "" {
				t.Fatalf("expected no output, got: %q", out)
			}
		})
	}
}
