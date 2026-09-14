package sinks

import (
	"io"
)

type SimpleLoggerConfig struct {
	KeyPadding        int               `json:"key_padding" yaml:"key_padding" mapstructure:"key_padding"`
	Writer            io.Writer         `json:"-" yaml:"-" mapstructure:"-"`
	EventFilterConfig EventFilterConfig `json:"event_filter_config" yaml:"event_filter" mapstructure:"event_filter_config"`
}

func DefaultSimpleLoggerConfig() SimpleLoggerConfig {
	return SimpleLoggerConfig{
		KeyPadding:        8,
		EventFilterConfig: DefaultEventFilterConfig(),
	}
}

func (c *SimpleLoggerConfig) WithEventFilterConfig(filter EventFilterConfig) *SimpleLoggerConfig {
	c.EventFilterConfig = filter
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
