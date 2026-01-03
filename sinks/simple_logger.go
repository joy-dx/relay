package sinks

import (
	"fmt"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
	"github.com/joy-dx/relay/output"
)

const SimpleLoggerRef = "simple"

type SimpleLoggerSink struct {
	level   int
	padding int
	cfg     *SimpleLoggerConfig
}

func NewSimpleLogger(cfg *SimpleLoggerConfig) *SimpleLoggerSink {
	return &SimpleLoggerSink{
		cfg:     cfg,
		level:   GetLogLevelIndex(cfg.Level, dto.Levels),
		padding: cfg.KeyPadding,
	}
}

func (s *SimpleLoggerSink) Ref() string {
	return SimpleLoggerRef
}

func (s *SimpleLoggerSink) Debug(e dto.RelayEventInterface) {
	if s.level <= 3 {
		return
	}
	fmt.Printf("%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *SimpleLoggerSink) Info(e dto.RelayEventInterface) {
	if s.level <= 2 {
		return
	}
	fmt.Printf("%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *SimpleLoggerSink) Warn(e dto.RelayEventInterface) {
	if s.level <= 1 {
		return
	}
	fmt.Printf("%s: %s\n", PadRight(string(e.RelayType()), s.padding), e.Message())
}

func (s *SimpleLoggerSink) Error(e dto.RelayEventInterface) {
	fmt.Println(e.Message())
}

func (s *SimpleLoggerSink) Fatal(e dto.RelayEventInterface) {
	fmt.Println(e.Message())
}

func (s *SimpleLoggerSink) Meta(e dto.RelayEventInterface) {
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
