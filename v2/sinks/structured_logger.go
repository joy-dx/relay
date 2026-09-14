package sinks

import (
	"context"
	"log/slog"
	"os"

	"github.com/joy-dx/relay/v2/dto"
)

const StructuredLoggerRef = "structured"

type StructuredLogger struct {
	logger *slog.Logger
	cfg    *StructuredLoggerConfig
}

func (s *StructuredLogger) Ref() string {
	return StructuredLoggerRef
}

func NewStructuredLogger(cfg *StructuredLoggerConfig) *StructuredLogger {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: convertLevel(cfg.Level),
	})
	return &StructuredLogger{
		cfg:    cfg,
		logger: slog.New(h),
	}
}

func convertLevel(l dto.RelayLevel) slog.Level {
	switch l {
	case dto.Debug:
		return slog.LevelDebug
	case dto.Info:
		return slog.LevelInfo
	case dto.Warn:
		return slog.LevelWarn
	case dto.Error:
		return slog.LevelError
	case dto.Fatal:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

func (s *StructuredLogger) Emit(ev dto.EmittedEvent) {
	if ev.Level == dto.Meta {
		return
	}
	s.logger.LogAttrs(
		context.Background(),
		convertLevel(ev.Level),
		ev.Event.Message(),
		ev.Event.ToSlog()...,
	)
}
func (s *StructuredLogger) Close() error {
	return nil
}
