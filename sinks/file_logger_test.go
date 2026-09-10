package sinks

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joy-dx/relay/dto"
)

type fileMsgEvent struct {
	msg string
}

func (e fileMsgEvent) RelayChannel() dto.EventChannel { return "relay" }
func (e fileMsgEvent) RelayType() dto.EventRef        { return "relay.log" }
func (e fileMsgEvent) Message() string                { return e.msg }
func (e fileMsgEvent) ToSlog() []slog.Attr            { return nil }

func TestFileLoggerSink_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		level      dto.RelayLevel
		call       func(s *FileLoggerSink)
		wantOutput string
	}{
		{
			name:  "debug suppressed at info",
			level: dto.Info,
			call: func(s *FileLoggerSink) {
				s.Debug(fileMsgEvent{msg: "debug-msg"})
			},
			wantOutput: "",
		},
		{
			name:  "info printed at info",
			level: dto.Info,
			call: func(s *FileLoggerSink) {
				s.Info(fileMsgEvent{msg: "info-msg"})
			},
			wantOutput: "relay.log: info-msg\n",
		},
		{
			name:  "warn printed at info",
			level: dto.Info,
			call: func(s *FileLoggerSink) {
				s.Warn(fileMsgEvent{msg: "warn-msg"})
			},
			wantOutput: "relay.log: warn-msg\n",
		},
		{
			name:  "error printed when enabled",
			level: dto.Error,
			call: func(s *FileLoggerSink) {
				s.Error(fileMsgEvent{msg: "error-msg"})
			},
			wantOutput: "ERROR: error-msg\n",
		},
		{
			name:  "fatal printed when enabled",
			level: dto.Fatal,
			call: func(s *FileLoggerSink) {
				s.Fatal(fileMsgEvent{msg: "fatal-msg"})
			},
			wantOutput: "FATAL: fatal-msg\n",
		},
		{
			name:  "meta always prints",
			level: dto.Fatal,
			call: func(s *FileLoggerSink) {
				s.Meta(fileMsgEvent{msg: "meta-msg"})
			},
			wantOutput: "META: meta-msg\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")

			cfg := DefaultFileLoggerConfig()
			cfg.WithFilePath(logPath).
				WithLevel(tt.level).
				WithKeyPadding(0)

			sink, err := NewFileLogger(&cfg)
			if err != nil {
				t.Fatalf("failed to create file logger: %v", err)
			}
			defer sink.Close()

			tt.call(sink)

			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("failed reading log file: %v", err)
			}

			got := string(data)

			if got != tt.wantOutput {
				t.Fatalf(
					"\n--- got ---\n%q\n--- want ---\n%q\n",
					got,
					tt.wantOutput,
				)
			}
		})
	}
}

func TestFileLoggerSink_CreatesDirectories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "a", "b", "c", "test.log")

	cfg := DefaultFileLoggerConfig()
	cfg.WithFilePath(nestedPath)

	sink, err := NewFileLogger(&cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer sink.Close()

	if _, err := os.Stat(nestedPath); err != nil {
		t.Fatalf("expected log file to exist")
	}
}

func TestFileLoggerSink_RequiresPath(t *testing.T) {
	t.Parallel()

	cfg := DefaultFileLoggerConfig()

	_, err := NewFileLogger(&cfg)
	if err == nil {
		t.Fatal("expected error when path is empty")
	}
}

func TestFileLoggerSink_Appends(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "append.log")

	cfg := DefaultFileLoggerConfig()
	cfg.WithFilePath(logPath).
		WithLevel(dto.Debug).
		WithKeyPadding(0)

	sink, err := NewFileLogger(&cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sink.Info(fileMsgEvent{msg: "first"})
	sink.Info(fileMsgEvent{msg: "second"})
	sink.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed reading log file: %v", err)
	}

	got := string(data)

	if !strings.Contains(got, "first") || !strings.Contains(got, "second") {
		t.Fatalf("expected appended lines, got: %q", got)
	}
}
