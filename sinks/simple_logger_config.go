package sinks

import (
	"io"

	"github.com/joy-dx/relay/dto"
)

type SimpleLoggerConfig struct {
	HideChannel bool           `json:"hide_channel", yaml:"hide_channel" mapstructure:"hide_channel"`
	KeyPadding  int            `json:"key_padding" yaml:"key_padding" mapstructure:"key_padding"`
	Level       dto.RelayLevel `json:"level" yaml:"level" mapstructure:"level"`
	Writer      io.Writer      `json:"-" yaml:"-" mapstructure:"-"`
}

func DefaultSimpleLoggerConfig() SimpleLoggerConfig {
	return SimpleLoggerConfig{
		Level: dto.Info,
	}
}

func (c *SimpleLoggerConfig) WithHideChannel(truthy bool) *SimpleLoggerConfig {
	c.HideChannel = truthy
	return c
}

func (c *SimpleLoggerConfig) WithLevel(level dto.RelayLevel) *SimpleLoggerConfig {
	c.Level = level
	return c
}

func (c *SimpleLoggerConfig) WithKeyPadding(keyPadding int) *SimpleLoggerConfig {
	c.KeyPadding = keyPadding
	return c
}

func (c *SimpleLoggerConfig) WithWriter(w io.Writer) *SimpleLoggerConfig {
	c.Writer = w
	return c
}
