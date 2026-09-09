package events

import (
	"fmt"
	"log/slog"

	"github.com/joy-dx/relay/v2/dto"
)

// RlyProgress Special handler for tracking task progress
type RlyProgress struct {
	JobName     string `json:"job_name"`
	TaskName    string `json:"task_name"`
	CurrentTask int    `json:"current_task"`
	TaskCount   int    `json:"task_count"`
}

func (e RlyProgress) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("type", string(e.RelayType())),
		slog.String("job_name", e.JobName),
		slog.String("task_name", e.TaskName),
		slog.Int("current_task", e.CurrentTask),
		slog.Int("task_count", e.TaskCount),
	}
}

func (e RlyProgress) Message() string {
	return fmt.Sprintf("%s %s %d/%d", e.JobName, e.TaskName, e.CurrentTask, e.TaskCount)
}

func (e RlyProgress) RelayChannel() dto.EventChannel {
	return "meta"
}

func (e RlyProgress) RelayType() dto.EventRef {
	return "meta.progress"
}
