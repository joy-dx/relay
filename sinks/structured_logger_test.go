// File: sinks/structured_logger_golden_test.go
package sinks

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/joy-dx/relay/dto"
)

// --- Test event fixtures ------------------------------------------------------

type slogEvent struct {
	msg   string
	attrs []slog.Attr
}

func (e slogEvent) RelayChannel() dto.EventChannel { return "relay" }
func (e slogEvent) RelayType() dto.EventRef        { return "relay.log" }
func (e slogEvent) Message() string                { return e.msg }
func (e slogEvent) ToSlog() []slog.Attr            { return e.attrs }

// --- Golden-table tests -------------------------------------------------------

func TestStructuredLogger_convertLevel_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   dto.RelayLevel
		want slog.Level
	}{
		{name: "debug", in: dto.Debug, want: slog.LevelDebug},
		{name: "info", in: dto.Info, want: slog.LevelInfo},
		{name: "warn", in: dto.Warn, want: slog.LevelWarn},
		{name: "error", in: dto.Error, want: slog.LevelError},
		// Fatal maps to error in current implementation
		{name: "fatal maps to error", in: dto.Fatal, want: slog.LevelError},
		// Meta/default also map to error in current implementation
		{name: "meta maps to error", in: dto.Meta, want: slog.LevelError},
		{name: "unknown maps to error", in: dto.RelayLevel("nope"), want: slog.LevelError},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := convertLevel(tt.in)
			if got != tt.want {
				t.Fatalf("level mismatch want=%v got=%v", tt.want, got)
			}
		})
	}
}

func TestStructuredLogger_EmitsExpectedTextLog_Golden(t *testing.T) {
	t.Parallel()

	// We avoid NewStructuredLogger() because it hardcodes os.Stdout.
	// Instead, we build the logger with a handler that writes to a buffer.
	//
	// This still tests StructuredLogger’s methods and formatting produced
	// by slog.NewTextHandler.

	type golden struct {
		name      string
		minLevel  slog.Level
		call      func(l *StructuredLogger, e dto.RelayEventInterface)
		event     dto.RelayEventInterface
		wantAny   []string // substrings that must appear
		wantNone  []string // substrings that must NOT appear
		wantEmpty bool     // expect no output
		normalize bool
	}

	mkLogger := func(min slog.Level, buf *bytes.Buffer) *StructuredLogger {
		h := slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: min,
		})
		return &StructuredLogger{
			cfg:    &StructuredLoggerConfig{Level: dto.Info},
			logger: slog.New(h),
		}
	}

	tests := []golden{
		{
			name:     "debug emits msg and attrs when enabled",
			minLevel: slog.LevelDebug,
			call: func(l *StructuredLogger, e dto.RelayEventInterface) {
				l.Debug(e)
			},
			event: slogEvent{
				msg: "hello",
				attrs: []slog.Attr{
					slog.String("k", "v"),
					slog.Int("n", 42),
				},
			},
			wantAny: []string{
				"level=DEBUG",
				`msg=hello`,
				`k=v`,
				`n=42`,
			},
		},
		{
			name:     "debug suppressed when min is info",
			minLevel: slog.LevelInfo,
			call: func(l *StructuredLogger, e dto.RelayEventInterface) {
				l.Debug(e)
			},
			event: slogEvent{
				msg: "hello",
				attrs: []slog.Attr{
					slog.String("k", "v"),
				},
			},
			wantEmpty: true,
		},
		{
			name:     "info emits msg and attrs at info",
			minLevel: slog.LevelInfo,
			call: func(l *StructuredLogger, e dto.RelayEventInterface) {
				l.Info(e)
			},
			event: slogEvent{
				msg: "info-msg",
				attrs: []slog.Attr{
					slog.String("channel", "relay"),
				},
			},
			wantAny: []string{
				"level=INFO",
				`msg=info-msg`,
				`channel=relay`,
			},
		},
		{
			name:     "warn emits at warn",
			minLevel: slog.LevelWarn,
			call: func(l *StructuredLogger, e dto.RelayEventInterface) {
				l.Warn(e)
			},
			event: slogEvent{
				msg: "warn-msg",
				attrs: []slog.Attr{
					slog.String("writer", "1"),
				},
			},
			wantAny: []string{
				"level=WARN",
				`msg=warn-msg`,
				`writer=1`,
			},
		},
		{
			name:     "error emits at error",
			minLevel: slog.LevelError,
			call: func(l *StructuredLogger, e dto.RelayEventInterface) {
				l.Error(e)
			},
			event: slogEvent{
				msg: "error-msg",
				attrs: []slog.Attr{
					slog.String("e", "1"),
				},
			},
			wantAny: []string{
				"level=ERROR",
				`msg=error-msg`,
				`e=1`,
			},
		},
		{
			name:     "fatal uses ERROR level and message literal FATAL (not e.Message())",
			minLevel: slog.LevelDebug,
			call: func(l *StructuredLogger, e dto.RelayEventInterface) {
				l.Fatal(e)
			},
			event: slogEvent{
				msg: "should-not-appear-as-msg",
				attrs: []slog.Attr{
					slog.String("reason", "boom"),
				},
			},
			wantAny: []string{
				"level=ERROR",
				`msg=FATAL`,
				`reason=boom`,
			},
			wantNone: []string{
				`msg=should-not-appear-as-msg`,
			},
		},
		{
			name:     "meta does nothing (no output)",
			minLevel: slog.LevelDebug,
			call: func(l *StructuredLogger, e dto.RelayEventInterface) {
				l.Meta(e)
			},
			event: slogEvent{
				msg: "meta-msg",
				attrs: []slog.Attr{
					slog.String("m", "1"),
				},
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			l := mkLogger(tt.minLevel, &buf)

			tt.call(l, tt.event)

			out := buf.String()

			if tt.wantEmpty {
				if out != "" {
					t.Fatalf("expected empty output, got: %q", out)
				}
				return
			}

			if out == "" {
				t.Fatalf("expected output, got empty")
			}

			for _, sub := range tt.wantAny {
				if !strings.Contains(out, sub) {
					t.Fatalf("expected output to contain %q\noutput:\n%s", sub, out)
				}
			}
			for _, sub := range tt.wantNone {
				if strings.Contains(out, sub) {
					t.Fatalf("expected output NOT to contain %q\noutput:\n%s", sub, out)
				}
			}
		})
	}
}

func TestNewStructuredLogger_WiresHandlerLevel_Golden(t *testing.T) {
	t.Parallel()

	// NewStructuredLogger() writes to os.Stdout, so we can’t easily assert emitted
	// text without capturing stdout. Instead, we assert the created logger has a
	// handler with the expected Enabled behavior by using slog’s Enabled check
	// via logger.Handler().Enabled(ctx, level).
	//
	// This avoids brittle text comparisons and doesn’t need stdout capture.

	tests := []struct {
		name     string
		cfgLevel dto.RelayLevel
		checkLvl slog.Level
		wantOn   bool
	}{
		{
			name:     "info config enables info",
			cfgLevel: dto.Info,
			checkLvl: slog.LevelInfo,
			wantOn:   true,
		},
		{
			name:     "info config disables debug",
			cfgLevel: dto.Info,
			checkLvl: slog.LevelDebug,
			wantOn:   false,
		},
		{
			name:     "debug config enables debug",
			cfgLevel: dto.Debug,
			checkLvl: slog.LevelDebug,
			wantOn:   true,
		},
		{
			name:     "error config disables warn",
			cfgLevel: dto.Error,
			checkLvl: slog.LevelWarn,
			wantOn:   false,
		},
		{
			name:     "fatal config behaves like error (enables error)",
			cfgLevel: dto.Fatal,
			checkLvl: slog.LevelError,
			wantOn:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := DefaultStructuredLoggerConfig()
			cfg.WithLevel(tt.cfgLevel)

			s := NewStructuredLogger(&cfg)

			gotOn := s.logger.Handler().Enabled(
				// context.Background() not required here, but fine
				// (we avoid importing context for a single call)
				nil,
				tt.checkLvl,
			)

			if gotOn != tt.wantOn {
				t.Fatalf("enabled mismatch for %v (cfg=%s): want=%v got=%v",
					tt.checkLvl, tt.cfgLevel, tt.wantOn, gotOn)
			}
		})
	}
}
