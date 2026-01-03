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
	level   int
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
		level:   GetLogLevelIndex(cfg.Level, dto.Levels),
		padding: cfg.KeyPadding,
		writer:  writer,
	}
}

func (s *SimpleLoggerSink) Ref() string {
	return SimpleLoggerRef
}

func (s *SimpleLoggerSink) Debug(e dto.RelayEventInterface) {
	if s.level <= 3 {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *SimpleLoggerSink) Info(e dto.RelayEventInterface) {
	if s.level <= 2 {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *SimpleLoggerSink) Warn(e dto.RelayEventInterface) {
	if s.level <= 1 {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *SimpleLoggerSink) Error(e dto.RelayEventInterface) {
	fmt.Fprintln(s.writer, e.Message())
}

func (s *SimpleLoggerSink) Fatal(e dto.RelayEventInterface) {
	fmt.Fprintln(s.writer, e.Message())
}

func (s *SimpleLoggerSink) Meta(e dto.RelayEventInterface) {
	metaCfg, castOk := e.(events.RlyMeta)
	if !castOk {
		fmt.Fprintln(s.writer, "Could not cast to RlyMeta")
		return
	}

	switch metaCfg.MetaType {
	case "section":
		fmt.Fprintln(s.writer, "")
		fmt.Fprintln(s.writer, "## "+e.Message())
		fmt.Fprintln(s.writer, "")

	case "failure":
		if _, printErr := output.ErrorColor.Print(" FAILURE "); printErr != nil {
			fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
		}
		fmt.Fprintln(s.writer, " "+e.Message())

	case "success":
		if _, printErr := output.SuccessColor.Print(" SUCCESS "); printErr != nil {
			fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
		}
		fmt.Fprintln(s.writer, " "+e.Message())
	}
}
