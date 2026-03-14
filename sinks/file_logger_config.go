package sinks

import (
	"github.com/joy-dx/relay/dto"
)

type FileLoggerConfig struct {
	AppendLog  bool           `yaml:"append_log," yaml:"append_log" mapstructure:"append_log"`
	Level      dto.RelayLevel `json:"level" yaml:"level" mapstructure:"level"`
	KeyPadding int            `json:"key_padding" yaml:"key_padding" mapstructure:"key_padding"`
	FilePath   string         `json:"file_path" yaml:"file_path" mapstructure:"file_path"`
}

func DefaultFileLoggerConfig() FileLoggerConfig {
	return FileLoggerConfig{
		KeyPadding: 8,
		Level:      dto.Info,
	}
}

func (c *FileLoggerConfig) WithAppendLog(truthy bool) *FileLoggerConfig {
	c.AppendLog = truthy
	return c
}

func (c *FileLoggerConfig) WithLevel(level dto.RelayLevel) *FileLoggerConfig {
	c.Level = level
	return c
}

func (c *FileLoggerConfig) WithKeyPadding(keyPadding int) *FileLoggerConfig {
	c.KeyPadding = keyPadding
	return c
}

func (c *FileLoggerConfig) WithFilePath(filePath string) *FileLoggerConfig {
	c.FilePath = filePath
	return c
}
