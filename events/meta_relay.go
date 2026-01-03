package events

import (
	"log/slog"

	"github.com/joy-dx/relay/dto"
)

// RlyMeta Special handler for custom relay events that the developer processes on an as needed basis
type RlyMeta struct {
	MetaType  string      `json:"meta_type"`
	Text      string      `json:"text"`
	ExtraData interface{} `json:"extra_data"`
}

func (e RlyMeta) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("type", string(e.RelayType())),
		slog.String("meta_type", e.MetaType),
		slog.String("text", e.Text),
	}
}

func (e RlyMeta) Message() string {
	return e.Text
}

func (e RlyMeta) RelayChannel() dto.EventChannel {
	return "core"
}

func (e RlyMeta) RelayType() dto.EventRef {
	return "cmd.log"
}
