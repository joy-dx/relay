package sinks

import (
	"fmt"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
	"github.com/joy-dx/relay/output"
)

const FilteredLoggerRef = "filtered"

type FilteredLoggerSink struct {
	level       int
	relayEvents map[dto.EventRef]struct{}
	cfg         *FilteredLoggerConfig
}

func NewFilteredLogger(cfg *FilteredLoggerConfig) *FilteredLoggerSink {
	set := make(map[dto.EventRef]struct{}, len(cfg.RelayTypes))
	for _, t := range cfg.RelayTypes {
		set[t] = struct{}{}
	}
	return &FilteredLoggerSink{
		cfg:         cfg,
		level:       GetLogLevelIndex(cfg.Level, dto.Levels),
		relayEvents: set,
	}
}

// FilteredLoggerSink Everyday sink to keep terminal noise down by only outputting cmd.log specific
func (s *FilteredLoggerSink) Ref() string {
	return FilteredLoggerRef
}

func (s *FilteredLoggerSink) Debug(e dto.RelayEventInterface) {
	if s.level <= 3 {
		return
	}
	_, ok := s.relayEvents[e.RelayType()]
	if !ok {
		return
	}
	fmt.Println(e.Message())
}

func (s *FilteredLoggerSink) Info(e dto.RelayEventInterface) {
	if s.level <= 2 {
		return
	}
	if _, ok := s.relayEvents[e.RelayType()]; !ok {
		return
	}
	fmt.Println(e.Message())
}
func (s *FilteredLoggerSink) Warn(e dto.RelayEventInterface) {
	if s.level <= 1 {
		return
	}
	if _, ok := s.relayEvents[e.RelayType()]; !ok {
		return
	}
	fmt.Println(e.Message())
}
func (s *FilteredLoggerSink) Error(e dto.RelayEventInterface) {
	fmt.Println(e.Message())
}

func (s *FilteredLoggerSink) Fatal(e dto.RelayEventInterface) {
	fmt.Println(e.Message())
}

func (s *FilteredLoggerSink) Meta(e dto.RelayEventInterface) {
	metaCfg, castOk := e.(events.RlyMeta)
	if !castOk {
		fmt.Println("Could not cast to RlyMeta")
	} else {
		switch metaCfg.MetaType {
		case "section":
			fmt.Println("")
			fmt.Println("## " + e.Message())
			fmt.Println("")
		case "failure":
			if _, printErr := output.ErrorColor.Print(" FAILURE "); printErr != nil {
				fmt.Println("failure print error: " + printErr.Error())
			}
			fmt.Printf(" %s\n", e.Message())
		case "success":
			if _, printErr := output.SuccessColor.Print(" SUCCESS "); printErr != nil {
				fmt.Println("failure print error: " + printErr.Error())
			}
			fmt.Printf(" %s\n", e.Message())
		}
	}
}
