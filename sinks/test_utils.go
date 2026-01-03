package sinks

import (
	"log/slog"

	"github.com/joy-dx/relay/dto"
)

type basicEvent struct {
	ref dto.EventRef
	msg string
}

func (e basicEvent) RelayChannel() dto.EventChannel { return "relay" }
func (e basicEvent) RelayType() dto.EventRef        { return e.ref }
func (e basicEvent) Message() string                { return e.msg }
func (e basicEvent) ToSlog() []slog.Attr            { return nil }
