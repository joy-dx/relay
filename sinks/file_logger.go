package sinks

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joy-dx/relay/v2/dto"
)

const FileLoggerRef = "file"

type FileLoggerSink struct {
	padding int
	file    *os.File
	mu      sync.Mutex
	cfg     *FileLoggerConfig
}

func NewFileLogger(cfg *FileLoggerConfig) (*FileLoggerSink, error) {
	if cfg.FilePath == "" {
		return nil, fmt.Errorf("file logger requires a path")
	}

	// ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
		return nil, err
	}

	var logFile *os.File
	var err error
	if cfg.AppendLog {
		logFile, err = os.OpenFile(
			cfg.FilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
	} else {
		logFile, err = os.Create(cfg.FilePath)
	}
	if err != nil {
		return nil, err
	}

	return &FileLoggerSink{
		cfg:     cfg,
		padding: cfg.KeyPadding,
		file:    logFile,
	}, nil
}

func (s *FileLoggerSink) Ref() string {
	return FileLoggerRef
}

func (s *FileLoggerSink) write(format string, args ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintf(s.file, format, args...)
}

func (s *FileLoggerSink) Debug(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Debug) {
		return
	}
	s.write("%s: %s\n", PadRight(string(ev.Event.RelayType()), s.padding), ev.Event.Message())
}

func (s *FileLoggerSink) Info(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Info) {
		return
	}

	s.write("%s: %s\n", PadRight(string(ev.Event.RelayType()), s.padding), ev.Event.Message())
}

func (s *FileLoggerSink) Warn(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Warn) {
		return
	}

	s.write("%s: %s\n", PadRight(string(ev.Event.RelayType()), s.padding), ev.Event.Message())
}

func (s *FileLoggerSink) Error(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Error) {
		return
	}
	s.write("ERROR: %s\n", ev.Event.Message())
}

func (s *FileLoggerSink) Fatal(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Fatal) {
		return
	}
	s.write("FATAL: %s\n", ev.Event.Message())
}

func (s *FileLoggerSink) Meta(ev dto.EmittedEvent) {
	s.write("META: %s\n", ev.Event.Message())
}

func (s *FileLoggerSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}
