package dto

import (
	"log/slog"
)

type RelayEventInterface interface {
	RelayChannel() EventChannel
	RelayType() EventRef
	Message() string
	ToSlog() []slog.Attr
}

type RelayInterface interface {
	Close() error
	Emit(level RelayLevel, event RelayEventInterface)
	Debug(event RelayEventInterface)
	Info(event RelayEventInterface)
	Warn(event RelayEventInterface)
	Error(event RelayEventInterface)
	Fatal(event RelayEventInterface)
	Meta(event RelayEventInterface)
	RegisterSink(sink RelaySinkInterface)
	UnregisterSink(sinkRef string) error
}

type RelaySinkInterface interface {
	Ref() string
	Close() error
	Emit(EmittedEvent)
}

type CloneableRelayEvent interface {
	Clone() RelayEventInterface
}
