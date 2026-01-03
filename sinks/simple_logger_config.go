package sinks

import (
	"github.com/joy-dx/relay/dto"
)

type SimpleLoggerConfig struct {
	KeyPadding int            `json:"key_padding" yaml:"key_padding" mapstructure:"key_padding"`
	Level      dto.RelayLevel `json:"level" yaml:"level" mapstructure:"level"`
}

func DefaultSimpleLoggerConfig() SimpleLoggerConfig {
	return SimpleLoggerConfig{
		Level: dto.Info,
	}
}

func (c *SimpleLoggerConfig) WithLevel(level dto.RelayLevel) *SimpleLoggerConfig {
	c.Level = level
	return c
}

func (c *SimpleLoggerConfig) WithKeyPadding(keyPadding int) *SimpleLoggerConfig {
	c.KeyPadding = keyPadding
	return c
}
