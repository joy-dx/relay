package sinks

import (
	"fmt"
	"io"
	"os"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
	"github.com/joy-dx/relay/output"
)

const FilteredLoggerRef = "filtered"

type FilteredLoggerSink struct {
	relayEvents map[dto.EventRef]struct{}
	writer      io.Writer
	cfg         *FilteredLoggerConfig
}

func NewFilteredLogger(cfg *FilteredLoggerConfig) *FilteredLoggerSink {
	set := make(map[dto.EventRef]struct{}, len(cfg.RelayTypes))
	for _, t := range cfg.RelayTypes {
		set[t] = struct{}{}
	}
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}
	return &FilteredLoggerSink{
		cfg:         cfg,
		relayEvents: set,
		writer:      writer,
	}
}

// FilteredLoggerSink Everyday sink to keep terminal noise down by only outputting cmd.log specific
func (s *FilteredLoggerSink) Ref() string {
	return FilteredLoggerRef
}

func (s *FilteredLoggerSink) Debug(ev dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Debug) {
		return
	}
	_, ok := s.relayEvents[ev.RelayType()]
	if !ok {
		return
	}
	fmt.Fprintln(s.writer, ev.Message())
}

func (s *FilteredLoggerSink) Info(ev dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Info) {
		return
	}
	if _, ok := s.relayEvents[ev.RelayType()]; !ok {
		return
	}
	fmt.Fprintln(s.writer, ev.Message())
}
func (s *FilteredLoggerSink) Warn(ev dto.RelayEventInterface) {
	if !levelEnabled(s.cfg.Level, dto.Warn) {
		return
	}
	if _, ok := s.relayEvents[ev.RelayType()]; !ok {
		return
	}
	fmt.Fprintln(s.writer, ev.Message())
}
func (s *FilteredLoggerSink) Error(ev dto.RelayEventInterface) {
	fmt.Fprintln(s.writer, ev.Message())
}

func (s *FilteredLoggerSink) Fatal(ev dto.RelayEventInterface) {
	fmt.Fprintln(s.writer, ev.Message())
}

func (s *FilteredLoggerSink) Meta(ev dto.RelayEventInterface) {
	metaCfg, castOk := ev.(events.RlyMeta)
	if !castOk {
		fmt.Fprintln(s.writer, "Could not cast to RlyMeta")
	} else {
		switch metaCfg.MetaType {
		case "section":
			fmt.Fprintln(s.writer, "")
			fmt.Fprintln(s.writer, "## "+ev.Message())
			fmt.Fprintln(s.writer, "")
		case "failure":
			if _, printErr := output.ErrorColor.Print(" FAILURE "); printErr != nil {
				fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
				return
			}
			fmt.Fprintln(s.writer, " "+ev.Message())
		case "success":
			if _, printErr := output.SuccessColor.Print(" SUCCESS "); printErr != nil {
				fmt.Fprintln(s.writer, "failure print error: "+printErr.Error())
				return
			}
			fmt.Fprintln(s.writer, " "+ev.Message())
		}
	}
}

func (s *FilteredLoggerSink) Close() error {
	return nil
}
