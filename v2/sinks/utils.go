package sinks

import (
	"fmt"

	"github.com/joy-dx/relay/v2/dto"
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

type EventFilterConfig struct {
	ChannelsAllowed []dto.EventChannel `json:"channels_allowed,omitempty"`
	EventsAllowed   []dto.EventRef     `json:"events_allowed,omitempty"`
	Level           dto.RelayLevel     `json:"level" yaml:"level" mapstructure:"level"`
}

func DefaultEventFilterConfig() EventFilterConfig {
	return EventFilterConfig{
		ChannelsAllowed: []dto.EventChannel{"*"},
		EventsAllowed:   []dto.EventRef{"*"},
		Level:           dto.Info,
	}
}
func (c *EventFilterConfig) WithLevel(level dto.RelayLevel) *EventFilterConfig {
	c.Level = level
	return c
}
func (c *EventFilterConfig) WithChannelsAllowed(incList []dto.EventChannel) *EventFilterConfig {
	c.ChannelsAllowed = incList
	return c
}

func (c *EventFilterConfig) WithEventsAllowed(incList []dto.EventRef) *EventFilterConfig {
	c.EventsAllowed = incList
	return c
}

type EventFilter struct {
	LevelEnabled         dto.RelayLevel
	RelayEventsAllowed   map[dto.EventRef]struct{}
	UseEventFilters      bool
	RelayChannelsAllowed map[dto.EventChannel]struct{}
	UseChannelFilters    bool
}

func NewEventFilter(cfg EventFilterConfig) *EventFilter {
	var useEventFilters bool
	eventsAllowed := make(map[dto.EventRef]struct{}, len(cfg.EventsAllowed))
	for _, t := range cfg.EventsAllowed {
		if t != "*" {
			useEventFilters = true
		}
		eventsAllowed[t] = struct{}{}
	}
	var useChannelFilters bool
	channelsAllowed := make(map[dto.EventChannel]struct{}, len(cfg.ChannelsAllowed))
	for _, t := range cfg.ChannelsAllowed {
		if t != "*" {
			useChannelFilters = true
		}
		channelsAllowed[t] = struct{}{}
	}

	return &EventFilter{
		LevelEnabled:         cfg.Level,
		RelayEventsAllowed:   eventsAllowed,
		RelayChannelsAllowed: channelsAllowed,
		UseChannelFilters:    useChannelFilters,
		UseEventFilters:      useEventFilters,
	}
}

// levelEnabled returns true if eventLevel should be emitted
func (f *EventFilter) IsLevelEnabled(eventLevel dto.RelayLevel) bool {
	if f.LevelEnabled == "" {
		return true
	}
	return levelPriority[eventLevel] >= levelPriority[f.LevelEnabled]
}

func (f *EventFilter) IsChannelAllowed(channel dto.EventChannel) bool {
	if !f.UseChannelFilters {
		return true
	}
	_, ok := f.RelayChannelsAllowed[channel]
	return ok
}

func (f *EventFilter) IsEventAllowed(eventType dto.EventRef) bool {
	if !f.UseEventFilters {
		return true
	}
	_, ok := f.RelayEventsAllowed[eventType]
	return ok
}
