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
	padding     int
	writer      io.Writer
	cfg         *SimpleLoggerConfig
	eventFilter *EventFilter
}

func NewSimpleLogger(cfg *SimpleLoggerConfig) *SimpleLoggerSink {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	filter := NewEventFilter(cfg.EventFilterConfig)

	return &SimpleLoggerSink{
		cfg:         cfg,
		padding:     cfg.KeyPadding,
		writer:      writer,
		eventFilter: filter,
	}
}

func (s *SimpleLoggerSink) Ref() string {
	return SimpleLoggerRef
}

func (s *SimpleLoggerSink) Emit(ev dto.EmittedEvent) {

	var prefix string
	switch ev.Level {
	case dto.Debug, dto.Info:
		if !s.eventFilter.IsLevelEnabled(ev.Level) {
			return
		}
		if !s.eventFilter.IsChannelAllowed(ev.Event.RelayChannel()) {
			return
		}
		if !s.eventFilter.IsEventAllowed(ev.Event.RelayType()) {
			return
		}
	case dto.Warn:
		if !s.eventFilter.IsLevelEnabled(ev.Level) {
			return
		}
		if !s.eventFilter.IsChannelAllowed(ev.Event.RelayChannel()) {
			return
		}
		if !s.eventFilter.IsEventAllowed(ev.Event.RelayType()) {
			return
		}
		prefix = "WARN "
	case dto.Error:
		prefix = "ERROR "
	case dto.Fatal:
		prefix = "FATAL "
	case dto.Meta:
		s.Meta(ev)
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", PadRight(prefix+string(ev.Event.RelayType()), s.padding), ev.Event.Message())
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
