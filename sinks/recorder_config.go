package sinks

import (
	"github.com/joy-dx/relay/dto"
)

type RecorderConfig struct {
	Level         dto.RelayLevel
	SegmentSize   int
	MaxSegments   int
	InputBuffer   int
	BlockOnFull   bool
	CloneOnRecord bool
}

func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		Level: dto.Info,
	}
}

func (c *RecorderConfig) WithLevel(level dto.RelayLevel) *RecorderConfig {
	c.Level = level
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
