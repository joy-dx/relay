package sinks

import (
	"context"
	"log/slog"
	"os"

	"github.com/joy-dx/relay/dto"
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

func (s *StructuredLogger) Debug(ev dto.EmittedEvent) {
	s.logger.LogAttrs(context.Background(), slog.LevelDebug, ev.Event.Message(), ev.Event.ToSlog()...)
}
func (s *StructuredLogger) Info(ev dto.EmittedEvent) {
	s.logger.LogAttrs(context.Background(), slog.LevelInfo, ev.Event.Message(), ev.Event.ToSlog()...)
}
func (s *StructuredLogger) Warn(ev dto.EmittedEvent) {
	s.logger.LogAttrs(context.Background(), slog.LevelWarn, ev.Event.Message(), ev.Event.ToSlog()...)
}
func (s *StructuredLogger) Error(ev dto.EmittedEvent) {
	s.logger.LogAttrs(context.Background(), slog.LevelError, ev.Event.Message(), ev.Event.ToSlog()...)
}
func (s *StructuredLogger) Fatal(ev dto.EmittedEvent) {
	s.logger.LogAttrs(context.Background(), slog.LevelError, "FATAL", ev.Event.ToSlog()...)
}

func (s *StructuredLogger) Meta(ev dto.EmittedEvent) {

}

func (s *StructuredLogger) Close() error {
	return nil
}
