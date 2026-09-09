package dto

import (
	"time"
)

type EventChannel string
type EventRef string

type RelaySinkConfig struct {
	Ref   string     `yaml:"ref"`
	Level RelayLevel `json:"level"`
}

type EmittedEvent struct {
	Time  time.Time
	Level RelayLevel
	Event RelayEventInterface
}

func (e EmittedEvent) Clone() EmittedEvent {
	out := e
	if cloner, ok := e.Event.(CloneableRelayEvent); ok {
		out.Event = cloner.Clone()
	}
	return out
}
