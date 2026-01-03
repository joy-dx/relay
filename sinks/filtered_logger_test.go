// File: sinks/filtered_logger_sink_golden_test.go
package sinks

import (
	"bytes"
	"testing"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
)

// --- Golden-table tests -------------------------------------------------------

func TestNewFilteredLogger_BuildsTypeSet_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		relayTypes []dto.EventRef
		checkRef   dto.EventRef
		wantOK     bool
	}{
		{
			name:       "type present",
			relayTypes: []dto.EventRef{"cmd.log", "relay.log"},
			checkRef:   "cmd.log",
			wantOK:     true,
		},
		{
			name:       "type absent",
			relayTypes: []dto.EventRef{"cmd.log"},
			checkRef:   "relay.log",
			wantOK:     false,
		},
		{
			name:       "empty config",
			relayTypes: []dto.EventRef{},
			checkRef:   "cmd.log",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &FilteredLoggerConfig{
				Level:      dto.Debug,
				RelayTypes: tt.relayTypes,
			}
			s := NewFilteredLogger(cfg)

			_, ok := s.relayEvents[tt.checkRef]
			if ok != tt.wantOK {
				t.Fatalf("set membership mismatch for %q want=%v got=%v",
					tt.checkRef, tt.wantOK, ok)
			}
		})
	}
}

func TestFilteredLoggerSink_LevelAndTypeFiltering_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfgLevel   dto.RelayLevel
		allowed    []dto.EventRef
		call       func(s *FilteredLoggerSink, e dto.RelayEventInterface)
		event      dto.RelayEventInterface
		wantOutput string // exact output (including newline) or "" for none
	}{
		{
			name:     "debug prints only when level allows and type allowed",
			cfgLevel: dto.Debug,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Debug(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "hello"},
			wantOutput: "hello\n",
		},
		{
			name:     "debug suppressed when configured at info",
			cfgLevel: dto.Info,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Debug(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "hello"},
			wantOutput: "",
		},
		{
			name:     "debug suppressed when type not allowed",
			cfgLevel: dto.Debug,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Debug(e)
			},
			event:      basicEvent{ref: "relay.log", msg: "nope"},
			wantOutput: "",
		},
		{
			name:     "info prints when level allows and type allowed",
			cfgLevel: dto.Info,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Info(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "info-msg"},
			wantOutput: "info-msg\n",
		},
		{
			name:     "warn prints when level allows and type allowed",
			cfgLevel: dto.Warn,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Warn(e)
			},
			event:      basicEvent{ref: "cmd.log", msg: "warn-msg"},
			wantOutput: "warn-msg\n",
		},
		{
			name:     "info suppressed when type not allowed",
			cfgLevel: dto.Info,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Info(e)
			},
			event:      basicEvent{ref: "relay.log", msg: "nope"},
			wantOutput: "",
		},
		{
			name:     "warn suppressed when type not allowed",
			cfgLevel: dto.Warn,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Warn(e)
			},
			event:      basicEvent{ref: "relay.log", msg: "nope"},
			wantOutput: "",
		},
		{
			name:     "error always prints regardless of allowed types",
			cfgLevel: dto.Fatal,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Error(e)
			},
			event:      basicEvent{ref: "relay.log", msg: "err"},
			wantOutput: "err\n",
		},
		{
			name:     "fatal always prints regardless of allowed types",
			cfgLevel: dto.Fatal,
			allowed:  []dto.EventRef{"cmd.log"},
			call: func(s *FilteredLoggerSink, e dto.RelayEventInterface) {
				s.Fatal(e)
			},
			event:      basicEvent{ref: "relay.log", msg: "fatal"},
			wantOutput: "fatal\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			cfg := &FilteredLoggerConfig{
				Level:      tt.cfgLevel,
				RelayTypes: tt.allowed,
				Writer:     &buf,
			}
			s := NewFilteredLogger(cfg)

			tt.call(s, tt.event)
			out := buf.String()

			if out != tt.wantOutput {
				t.Fatalf("output mismatch\nwant: %q\ngot:  %q", tt.wantOutput, out)
			}
		})
	}
}

func TestFilteredLoggerSink_Meta_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		event      dto.RelayEventInterface
		wantSubstr string
		wantEmpty  bool
	}{
		{
			name: "meta non-RlyMeta prints cast error",
			event: basicEvent{
				ref: "cmd.log",
				msg: "not meta",
			},
			wantSubstr: "Could not cast to RlyMeta",
		},
		{
			name: "meta section prints header with blank lines",
			event: events.RlyMeta{
				MetaType: "section",
				Text:     "My Section",
			},
			// exact string is: "\n## My Section\n\n"
			wantSubstr: "## My Section",
		},
		{
			name: "meta failure prints message (at least)",
			event: events.RlyMeta{
				MetaType: "failure",
				Text:     "Bad news",
			},
			wantSubstr: "Bad news",
		},
		{
			name: "meta success prints message (at least)",
			event: events.RlyMeta{
				MetaType: "success",
				Text:     "Good news",
			},
			wantSubstr: "Good news",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			cfg := &FilteredLoggerConfig{
				Level:      dto.Debug,
				RelayTypes: []dto.EventRef{"cmd.log"},
				Writer:     &buf,
			}
			s := NewFilteredLogger(cfg)

			s.Meta(tt.event)
			out := buf.String()

			if tt.wantEmpty {
				if out != "" {
					t.Fatalf("expected empty output, got: %q", out)
				}
				return
			}

			if tt.wantSubstr != "" && !bytes.Contains([]byte(out), []byte(tt.wantSubstr)) {
				t.Fatalf("expected output to contain %q\noutput:\n%s",
					tt.wantSubstr, out)
			}
		})
	}
}
