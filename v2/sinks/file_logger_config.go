package sinks

type FileLoggerConfig struct {
	AppendLog         bool              `yaml:"append_log," yaml:"append_log" mapstructure:"append_log"`
	KeyPadding        int               `json:"key_padding" yaml:"key_padding" mapstructure:"key_padding"`
	FilePath          string            `json:"file_path" yaml:"file_path" mapstructure:"file_path"`
	EventFilterConfig EventFilterConfig `json:"event_filter_config" yaml:"event_filter" mapstructure:"event_filter_config"`
}

func DefaultFileLoggerConfig() FileLoggerConfig {
	return FileLoggerConfig{
		KeyPadding:        8,
		EventFilterConfig: DefaultEventFilterConfig(),
	}
}

func (c *FileLoggerConfig) WithAppendLog(truthy bool) *FileLoggerConfig {
	c.AppendLog = truthy
	return c
}

func (c *FileLoggerConfig) WithEventFilterConfig(filter EventFilterConfig) *FileLoggerConfig {
	c.EventFilterConfig = filter
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
