package sinks

import (
	"fmt"
	"io"
	"os"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
	"github.com/joy-dx/relay/output"
)

const FileLoggerRef = "simple"

type FileLoggerSink struct {
	level  int
	writer io.Writer
	cfg    *FileLoggerConfig
}

func NewFileLogger(cfg *FileLoggerConfig) *FileLoggerSink {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}
	return &FileLoggerSink{
		cfg:    cfg,
		level:  GetLogLevelIndex(cfg.Level, dto.Levels),
		writer: writer,
	}
}

func (s *FileLoggerSink) Ref() string {
	return FileLoggerRef
}

func (s *FileLoggerSink) Debug(e dto.RelayEventInterface) {
	if s.level <= 3 {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", e.RelayType(), e.Message())

}
func (s *FileLoggerSink) Info(e dto.RelayEventInterface) {
	if s.level <= 2 {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", e.RelayType(), e.Message())
}

func (s *FileLoggerSink) Warn(e dto.RelayEventInterface) {
	if s.level <= 1 {
		return
	}
	fmt.Fprintf(s.writer, "%s: %s\n", e.RelayType(), e.Message())
}

func (s *FileLoggerSink) Error(e dto.RelayEventInterface) {
	fmt.Fprintln(s.writer, e.Message())
}

func (s *FileLoggerSink) Fatal(e dto.RelayEventInterface) {
	fmt.Fprintln(s.writer, e.Message())
}

func (s *FileLoggerSink) Meta(e dto.RelayEventInterface) {
	metaCfg, castOk := e.(events.RlyMeta)
	if !castOk {
		fmt.Fprintln(s.writer, "Could not cast to RlyMeta")
	} else {
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
}
