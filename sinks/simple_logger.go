package sinks

import (
	"fmt"
	"io"
	"os"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
	"github.com/joy-dx/relay/output"
)

const SimpleLoggerRef = "simple"

type SimpleLoggerSink struct {
	padding int
	writer  io.Writer
	cfg     *SimpleLoggerConfig
}

func NewSimpleLogger(cfg *SimpleLoggerConfig) *SimpleLoggerSink {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}
	return &SimpleLoggerSink{
		cfg:     cfg,
		padding: cfg.KeyPadding,
		writer:  writer,
	}
}

func (s *SimpleLoggerSink) Ref() string {
	return SimpleLoggerRef
}

func (s *SimpleLoggerSink) formatEvent(ev dto.RelayEventInterface) string {
	if s.cfg.HideChannel {
		return ev.Message()
	}
	return fmt.Sprintf("%s: %s\n", PadRight(string(ev.RelayType()), s.padding), ev.Message())
}

func (s *SimpleLoggerSink) Debug(ev dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Debug) {
		return
	}
	_, _ = io.WriteString(s.writer, s.formatEvent(ev))
}

func (s *SimpleLoggerSink) Info(ev dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Info) {
		return
	}
	_, _ = io.WriteString(s.writer, s.formatEvent(ev))
}

func (s *SimpleLoggerSink) Warn(ev dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Warn) {
		return
	}
	_, _ = io.WriteString(s.writer, s.formatEvent(ev))
}

func (s *SimpleLoggerSink) Error(ev dto.RelayEventInterface) {
	_, _ = io.WriteString(s.writer, s.formatEvent(ev))
}

func (s *SimpleLoggerSink) Fatal(ev dto.RelayEventInterface) {
	_, _ = io.WriteString(s.writer, s.formatEvent(ev))
}

func (s *SimpleLoggerSink) Meta(ev dto.RelayEventInterface) {
	metaCfg, castOk := ev.(events.RlyMeta)
	if !castOk {
		fmt.Fprintln(s.writer, "Could not cast to RlyMeta")
		return
	}

	switch metaCfg.MetaType {
	case "section":
		fmt.Fprintln(s.writer, "")
		fmt.Fprintln(s.writer, "## "+ev.Message())
		fmt.Fprintln(s.writer, "")

	case "failure":
		if _, printErr := output.ErrorColor.Print(" FAILURE "); printErr != nil {
			fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
		}
		fmt.Fprintln(s.writer, " "+ev.Message())

	case "success":
		if _, printErr := output.SuccessColor.Print(" SUCCESS "); printErr != nil {
			fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
		}
		fmt.Fprintln(s.writer, " "+ev.Message())
	}
}

func (s *SimpleLoggerSink) Close() error {
	return nil
}
