package sinks

import (
	"fmt"

	"github.com/joy-dx/relay/dto"
)

var levelPriority = map[dto.RelayLevel]int{
	dto.Debug: 0,
	dto.Info:  1,
	dto.Warn:  2,
	dto.Error: 3,
	dto.Fatal: 4,
}

// levelEnabled returns true if eventLevel should be emitted
func levelEnabled(cfgLevel, eventLevel dto.RelayLevel) bool {
	return levelPriority[eventLevel] >= levelPriority[cfgLevel]
}

func PadRight(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}
