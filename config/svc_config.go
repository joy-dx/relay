package config

import "github.com/joy-dx/relay/v2/dto"

type RelaySvcConfig struct {
	Sinks []dto.RelaySinkInterface `json:"sinks" yaml:"sinks" mapstructure:"sinks"`
}

func DefaultRelaySvcConfig() RelaySvcConfig {
	return RelaySvcConfig{}
}
