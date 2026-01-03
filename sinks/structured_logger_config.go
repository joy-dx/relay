package sinks

import (
	"github.com/joy-dx/relay/dto"
)

type StructuredLoggerConfig struct {
	Level dto.RelayLevel `json:"level" yaml:"level" mapstructure:"level"`
}

func DefaultStructuredLoggerConfig() StructuredLoggerConfig {
	return StructuredLoggerConfig{
		Level: dto.Info,
	}
}

func (c *StructuredLoggerConfig) WithLevel(level dto.RelayLevel) *StructuredLoggerConfig {
	c.Level = level
	return c
}
