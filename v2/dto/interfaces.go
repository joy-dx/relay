package dto

import (
	"log/slog"
)

// RelayEventInterface values are treated as immutable after emission.
//
// Relay dispatches the same event value to every sink. Callers must not mutate
// the event, or data referenced by the event, while it is being emitted or
// retained by a sink. Events that need to be retained independently should
// implement CloneableRelayEvent.
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

	CurrentRevision() string
	HistoryEnabled() bool
	ReplayAfter(revision string) (ReplayResult, error)
}

type RelaySinkInterface interface {
	Ref() string
	Close() error
	Emit(EmittedEvent)
}

type CloneableRelayEvent interface {
	Clone() RelayEventInterface
}
