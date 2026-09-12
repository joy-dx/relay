package sinks

import (
	"github.com/joy-dx/relay/v2/dto"
)

const RECORDER_DEFAULT_SEGMENT_SIZE = 512
const RECORDER_DEFAULT_INPUT_BUFFER_SIZE = 1024

type RecorderFilter func(dto.EmittedEvent) bool

type RecorderConfig struct {
	Level           dto.RelayLevel     `json:"level" yaml:"level"`
	SegmentSize     int                `json:"segment_size" yaml:"segment_size"`
	MaxSegments     int                `json:"max_segments" yaml:"max_segments"`
	InputBuffer     int                `json:"input_buffer" yaml:"input_buffer"`
	BlockOnFull     bool               `json:"block_on_full,omitempty" yaml:"block_on_full,omitempty"`
	CloneOnRecord   bool               `json:"clone_on_record,omitempty" yaml:"clone_on_record,omitempty"`
	ChannelsAllowed []dto.EventChannel `json:"channels_allowed,omitempty"`
	EventsAllowed   []dto.EventRef     `json:"events_allowed,omitempty"`
	Filter          RecorderFilter     `json:"-" yaml:"-"`
}

func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		Level:           dto.Info,
		ChannelsAllowed: []dto.EventChannel{"*"},
		EventsAllowed:   []dto.EventRef{"*"},
		SegmentSize:     RECORDER_DEFAULT_SEGMENT_SIZE,
		InputBuffer:     RECORDER_DEFAULT_INPUT_BUFFER_SIZE,
	}
}

func (c *RecorderConfig) WithLevel(level dto.RelayLevel) *RecorderConfig {
	c.Level = level
	return c
}

func (c *RecorderConfig) WithChannelsAllowed(incList []dto.EventChannel) *RecorderConfig {
	c.ChannelsAllowed = incList
	return c
}

func (c *RecorderConfig) WithEventsAllowed(incList []dto.EventRef) *RecorderConfig {
	c.EventsAllowed = incList
	return c
}

func (c *RecorderConfig) WithSegmentSize(size int) *RecorderConfig {
	c.SegmentSize = size
	return c
}

func (c *RecorderConfig) WithMaxSegments(segments int) *RecorderConfig {
	c.MaxSegments = segments
	return c
}

func (c *RecorderConfig) WithInputBuffer(size int) *RecorderConfig {
	c.InputBuffer = size
	return c
}

func (c *RecorderConfig) WithBlockOnFull(truthy bool) *RecorderConfig {
	c.BlockOnFull = truthy
	return c
}

func (c *RecorderConfig) WithCloneOnRecord(truthy bool) *RecorderConfig {
	c.CloneOnRecord = truthy
	return c
}

func (c *RecorderConfig) WithFilterFunction(
	filter RecorderFilter,
) *RecorderConfig {
	c.Filter = filter
	return c
}
