package sinks

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joy-dx/relay/dto"
)

const FileLoggerRef = "simple"

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

	f, err := os.OpenFile(
		cfg.FilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &FileLoggerSink{
		cfg:     cfg,
		padding: cfg.KeyPadding,
		file:    f,
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

func (s *FileLoggerSink) Debug(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Debug) {
		return
	}
	s.write("%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *FileLoggerSink) Info(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Info) {
		return
	}

	s.write("%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *FileLoggerSink) Warn(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Warn) {
		return
	}

	s.write("%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *FileLoggerSink) Error(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Error) {
		return
	}
	s.write("ERROR: %s\n", e.Message())
}

func (s *FileLoggerSink) Fatal(e dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Fatal) {
		return
	}
	s.write("FATAL: %s\n", e.Message())
}

func (s *FileLoggerSink) Meta(e dto.RelayEventInterface) {
	s.write("META: %s\n", e.Message())
}

func (s *FileLoggerSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}
