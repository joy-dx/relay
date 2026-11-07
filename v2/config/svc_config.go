package config

import "github.com/joy-dx/relay/v2/dto"

type RelaySvcConfig struct {
	Sinks       []dto.RelaySinkInterface `json:"sinks" yaml:"sinks" mapstructure:"sinks"`
	HistorySize int                      `json:"history_size" yaml:"history_size" mapstructure:"history_size"`
}

func DefaultRelaySvcConfig() RelaySvcConfig {
	return RelaySvcConfig{}
}

func (c *RelaySvcConfig) AddSink(sink dto.RelaySinkInterface) *RelaySvcConfig {
	c.Sinks = append(c.Sinks, sink)
	return c
}
func (c *RelaySvcConfig) WithSinks(sinks []dto.RelaySinkInterface) *RelaySvcConfig {
	c.Sinks = sinks
	return c
}
func (c *RelaySvcConfig) WithHistorySize(historySize int) *RelaySvcConfig {
	c.HistorySize = historySize
	return c
}
