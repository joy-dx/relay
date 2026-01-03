// File: sinks/file_logger_sink_golden_test.go
package sinks

import (
	"bytes"
	"testing"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
)

// --- Tests -------------------------------------------------------------------

func TestFileLoggerSink_LevelGatingAndFormatting_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfgLevel   dto.RelayLevel
		call       func(s *FileLoggerSink, e dto.RelayEventInterface)
		event      dto.RelayEventInterface
		wantOutput string
	}{
		{
			name:     "debug prints at debug",
			cfgLevel: dto.Debug,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Debug(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "d"},
			wantOutput: "cmd.log: d\n",
		},
		{
			name:     "debug suppressed at info",
			cfgLevel: dto.Info,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Debug(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "d"},
			wantOutput: "",
		},
		{
			name:     "info prints at info",
			cfgLevel: dto.Info,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Info(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "i"},
			wantOutput: "cmd.log: i\n",
		},
		{
			name:     "info suppressed at warn",
			cfgLevel: dto.Warn,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Info(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "i"},
			wantOutput: "",
		},
		{
			name:     "warn prints at warn",
			cfgLevel: dto.Warn,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Warn(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "w"},
			wantOutput: "cmd.log: w\n",
		},
		{
			name:     "warn suppressed at error",
			cfgLevel: dto.Error,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Warn(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "w"},
			wantOutput: "",
		},
		{
			name:     "error always prints message only",
			cfgLevel: dto.Fatal,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Error(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "err"},
			wantOutput: "err\n",
		},
		{
			name:     "fatal always prints message only",
			cfgLevel: dto.Fatal,
			call: func(s *FileLoggerSink, e dto.RelayEventInterface) {
				s.Fatal(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "fatal"},
			wantOutput: "fatal\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &FileLoggerConfig{
				Level: tt.cfgLevel,
			}
			s := NewFileLogger(cfg)

			out := CaptureStdout(t, func() {
				tt.call(s, tt.event)
			})

			if out != tt.wantOutput {
				t.Fatalf("output mismatch\nwant: %q\ngot:  %q", tt.wantOutput, out)
			}
		})
	}
}

func TestFileLoggerSink_Meta_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		event     dto.RelayEventInterface
		wantSub   string
		wantEmpty bool
	}{
		{
			name:    "meta non-RlyMeta prints cast error",
			event:   basicEvent{ref: "cmd.log", msg: "nope"},
			wantSub: "Could not cast to RlyMeta",
		},
		{
			name:    "meta section prints header",
			event:   events.RlyMeta{MetaType: "section", Text: "Section Name"},
			wantSub: "## Section Name",
		},
		{
			name:    "meta failure prints message (at least)",
			event:   events.RlyMeta{MetaType: "failure", Text: "Bad"},
			wantSub: "Bad",
		},
		{
			name:    "meta success prints message (at least)",
			event:   events.RlyMeta{MetaType: "success", Text: "Good"},
			wantSub: "Good",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &FileLoggerConfig{Level: dto.Debug}
			s := NewFileLogger(cfg)

			out := CaptureStdout(t, func() {
				s.Meta(tt.event)
			})

			if tt.wantEmpty {
				if out != "" {
					t.Fatalf("expected empty output, got: %q", out)
				}
				return
			}

			if tt.wantSub != "" && !bytes.Contains([]byte(out), []byte(tt.wantSub)) {
				t.Fatalf("expected output to contain %q\noutput:\n%s",
					tt.wantSub, out)
			}
		})
	}
}
