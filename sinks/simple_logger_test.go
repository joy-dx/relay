package sinks

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/joy-dx/relay/v2/dto"
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
		emitLevel  dto.RelayLevel
		msg        string
		wantOutput bool
	}{
		{
			name:       "debug suppressed at info",
			cfgLevel:   dto.Info,
			emitLevel:  dto.Debug,
			msg:        "d",
			wantOutput: false,
		},
		{
			name:       "info printed at info",
			cfgLevel:   dto.Info,
			emitLevel:  dto.Info,
			msg:        "i",
			wantOutput: true,
		},
		{
			name:       "warn printed at info",
			cfgLevel:   dto.Info,
			emitLevel:  dto.Warn,
			msg:        "writer",
			wantOutput: true,
		},
		{
			name:       "error always prints",
			cfgLevel:   dto.Fatal,
			emitLevel:  dto.Error,
			msg:        "e",
			wantOutput: true,
		},
		{
			name:       "fatal always prints",
			cfgLevel:   dto.Fatal,
			emitLevel:  dto.Fatal,
			msg:        "f",
			wantOutput: true,
		},
	}

	fixedTime := time.Date(2026, 9, 9, 7, 0, 0, 0, time.UTC)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			cfg := DefaultSimpleLoggerConfig()
			cfg.WithLevel(tt.cfgLevel).WithKeyPadding(0).WithWriter(&buf)

			sink := NewSimpleLogger(&cfg)

			switch tt.emitLevel {
			case dto.Debug:
				sink.Debug(dto.EmittedEvent{
					Time:  fixedTime,
					Level: tt.emitLevel,
					Event: msgEvent{msg: tt.msg},
				})

			case dto.Info:
				sink.Info(dto.EmittedEvent{
					Time:  fixedTime,
					Level: tt.emitLevel,
					Event: msgEvent{msg: tt.msg},
				})

			case dto.Warn:
				sink.Warn(dto.EmittedEvent{
					Time:  fixedTime,
					Level: tt.emitLevel,
					Event: msgEvent{msg: tt.msg},
				})

			case dto.Error:
				sink.Error(dto.EmittedEvent{
					Time:  fixedTime,
					Level: tt.emitLevel,
					Event: msgEvent{msg: tt.msg},
				})
			case dto.Fatal:
				sink.Fatal(dto.EmittedEvent{
					Time:  fixedTime,
					Level: tt.emitLevel,
					Event: msgEvent{msg: tt.msg},
				})
			}

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
