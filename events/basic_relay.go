package events

import (
	"log/slog"

	"github.com/joy-dx/relay/dto"
)

const RELAY_CHANNEL dto.EventChannel = "relay"

const RELAY_LOG dto.EventRef = "relay.log"

type RlyLog struct {
	Msg string `json:"msg" yaml:"msg"`
}

func (e RlyLog) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("msg", e.Msg),
	}
}

func (e RlyLog) Message() string {
	return e.Msg
}

func (e RlyLog) RelayChannel() dto.EventChannel {
	return RELAY_CHANNEL
}

func (e RlyLog) RelayType() dto.EventRef {
	return RELAY_LOG
}
