package sinks

import (
	"fmt"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
	"github.com/joy-dx/relay/output"
)

const FileLoggerRef = "simple"

type FileLoggerSink struct {
	level int
	cfg   *FileLoggerConfig
}

func NewFileLogger(cfg *FileLoggerConfig) *FileLoggerSink {
	return &FileLoggerSink{
		cfg:   cfg,
		level: GetLogLevelIndex(cfg.Level, dto.Levels),
	}
}

func (s *FileLoggerSink) Ref() string {
	return FileLoggerRef
}

func (s *FileLoggerSink) Debug(e dto.RelayEventInterface) {
	if s.level <= 3 {
		return
	}
	fmt.Printf("%s: %s\n", e.RelayType(), e.Message())

}
func (s *FileLoggerSink) Info(e dto.RelayEventInterface) {
	if s.level <= 2 {
		return
	}
	fmt.Printf("%s: %s\n", e.RelayType(), e.Message())
}

func (s *FileLoggerSink) Warn(e dto.RelayEventInterface) {
	if s.level <= 1 {
		return
	}
	fmt.Printf("%s: %s\n", e.RelayType(), e.Message())
}

func (s *FileLoggerSink) Error(e dto.RelayEventInterface) {
	fmt.Println(e.Message())
}

func (s *FileLoggerSink) Fatal(e dto.RelayEventInterface) {
	fmt.Println(e.Message())
}

func (s *FileLoggerSink) Meta(e dto.RelayEventInterface) {
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
