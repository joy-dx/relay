package sinks

import (
	"fmt"
	"io"
	"os"

	"github.com/joy-dx/relay/v2/dto"
	"github.com/joy-dx/relay/v2/events"
	"github.com/joy-dx/relay/v2/output"
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

func (s *SimpleLoggerSink) Debug(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Debug) {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", PadRight(string(ev.Event.RelayType()), s.padding), ev.Event.Message())
}

func (s *SimpleLoggerSink) Info(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Info) {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", PadRight(string(ev.Event.RelayType()), s.padding), ev.Event.Message())
}

func (s *SimpleLoggerSink) Warn(ev dto.EmittedEvent) {
	if !levelEnabled(s.cfg.Level, dto.Warn) {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", PadRight(string(ev.Event.RelayType()), s.padding), ev.Event.Message())
}

func (s *SimpleLoggerSink) Error(ev dto.EmittedEvent) {
	fmt.Fprintln(s.writer, ev.Event.Message())
}

func (s *SimpleLoggerSink) Fatal(ev dto.EmittedEvent) {
	fmt.Fprintln(s.writer, ev.Event.Message())
}

func (s *SimpleLoggerSink) Meta(ev dto.EmittedEvent) {
	metaCfg, castOk := ev.Event.(events.RlyMeta)
	if !castOk {
		fmt.Fprintln(s.writer, "Could not cast to RlyMeta")
		return
	}

	switch metaCfg.MetaType {
	case "section":
		fmt.Fprintln(s.writer, "")
		fmt.Fprintln(s.writer, "## "+ev.Event.Message())
		fmt.Fprintln(s.writer, "")

	case "failure":
		if _, printErr := output.ErrorColor.Print(" FAILURE "); printErr != nil {
			fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
		}
		fmt.Fprintln(s.writer, " "+ev.Event.Message())

	case "success":
		if _, printErr := output.SuccessColor.Print(" SUCCESS "); printErr != nil {
			fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
		}
		fmt.Fprintln(s.writer, " "+ev.Event.Message())
	}
}

func (s *SimpleLoggerSink) Close() error {
	return nil
}
