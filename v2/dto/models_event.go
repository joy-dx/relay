package dto

import (
	"time"
)

type EventChannel string
type EventRef string

type RelaySinkConfig struct {
	Ref   string     `json:"ref" yaml:"ref"`
	Level RelayLevel `json:"level" yaml:"level"`
}

type EmittedEvent struct {
	Revision uint64              `json:"revision,string" yaml:"revision" ts_type:"string"`
	Time     time.Time           `json:"time" yaml:"time" ts_type:"string"`
	Channel  EventChannel        `json:"channel" yaml:"channel"`
	Type     EventRef            `json:"type" yaml:"type" ts_type:"string"`
	Level    RelayLevel          `json:"level" yaml:"level"`
	Event    RelayEventInterface `json:"event" yaml:"event"`
}

func (e EmittedEvent) Clone() EmittedEvent {
	out := e
	if cloner, ok := e.Event.(CloneableRelayEvent); ok {
		out.Event = cloner.Clone()
	}
	return out
}
