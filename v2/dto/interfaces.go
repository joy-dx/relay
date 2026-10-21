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
	DebugWithOpts(event RelayEventInterface, opts RlyOpts)

	Info(event RelayEventInterface)
	InfoWithOpts(event RelayEventInterface, opts RlyOpts)

	Warn(event RelayEventInterface)
	WarnWithOpts(event RelayEventInterface, opts RlyOpts)

	Error(event RelayEventInterface)
	ErrorWithOpts(event RelayEventInterface, opts RlyOpts)

	Fatal(event RelayEventInterface)
	FatalWithOpts(event RelayEventInterface, opts RlyOpts)

	Meta(event RelayEventInterface)
	MetaWithOpts(event RelayEventInterface, opts RlyOpts)

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
