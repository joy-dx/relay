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
	Time  time.Time           `json:"time" yaml:"time" ts_type:"string"`
	Level RelayLevel          `json:"level" yaml:"level"`
	Event RelayEventInterface `json:"event" yaml:"event"`
}

func (e EmittedEvent) Clone() EmittedEvent {
	out := e
	if cloner, ok := e.Event.(CloneableRelayEvent); ok {
		out.Event = cloner.Clone()
	}
	return out
}
