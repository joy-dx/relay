package sinks

import (
	"github.com/joy-dx/relay/dto"
)

type FilteredLoggerConfig struct {
	Level      dto.RelayLevel `json:"level" yaml:"level" mapstructure:"level"`
	RelayTypes []dto.EventRef `json:"relay_types" yaml:"relay_types" mapstructure:"relay_types"`
}

func DefaultFilteredLoggerConfig() FilteredLoggerConfig {
	return FilteredLoggerConfig{
		Level: dto.Info,
		RelayTypes: []dto.EventRef{
			"cmd.log",
			"relay.log",
		},
	}
}

func (c *FilteredLoggerConfig) WithLevel(level dto.RelayLevel) *FilteredLoggerConfig {
	c.Level = level
	return c
}

func (c *FilteredLoggerConfig) WithRelay(event dto.EventRef) *FilteredLoggerConfig {
	c.RelayTypes = append(c.RelayTypes, event)
	return c
}

func (c *FilteredLoggerConfig) WithRelays(events []dto.EventRef) *FilteredLoggerConfig {
	c.RelayTypes = events
	return c
}
