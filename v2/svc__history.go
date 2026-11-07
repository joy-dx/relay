package relay

import (
	"fmt"
	"strconv"

	"github.com/joy-dx/relay/v2/dto"
)

// appendHistoryLocked requires r.mu to be held.
func (r *RelaySvc) appendHistoryLocked(event dto.EmittedEvent) {
	if len(r.history) == 0 {
		return
	}

	if r.count < len(r.history) {
		index := (r.start + r.count) % len(r.history)
		r.history[index] = event
		r.count++
		return
	}

	r.history[r.start] = event
	r.start = (r.start + 1) % len(r.history)
}

// CurrentRevision returns a decimal string to avoid JavaScript
// integer precision loss.
func (r *RelaySvc) CurrentRevision() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return strconv.FormatUint(r.revision, 10)
}

func (r *RelaySvc) HistoryEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.history) > 0
}

// ReplayAfter returns events in (revision, Through], in ascending
// revision order.
//
// Gap indicates that the requested range cannot be satisfied.
// No events are emitted to sinks during replay.
//
// When history is disabled, every request returns ErrReplayDisabled.
func (r *RelaySvc) ReplayAfter(
	revision string,
) (dto.ReplayResult, error) {
	result, err := r.replayAfterSnapshot(revision)
	if err != nil {
		return dto.ReplayResult{}, err
	}

	// Clone outside r.mu: Clone is application-defined code.
	for i := range result.Events {
		result.Events[i] = result.Events[i].Clone()
	}

	return result, nil
}

func (r *RelaySvc) replayAfterSnapshot(
	revision string,
) (dto.ReplayResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.history) == 0 {
		return dto.ReplayResult{}, dto.ErrReplayDisabled
	}

	after, err := strconv.ParseUint(revision, 10, 64)
	if err != nil {
		return dto.ReplayResult{}, fmt.Errorf(
			"invalid relay revision %q: %w",
			revision,
			err,
		)
	}

	result := dto.ReplayResult{
		Events:  make([]dto.EmittedEvent, 0),
		Through: r.revision,
	}

	if after > r.revision {
		result.Gap = true
		return result, nil
	}

	if after == r.revision {
		return result, nil
	}

	if r.count == 0 {
		result.Gap = true
		return result, nil
	}

	oldest := r.history[r.start].Revision

	// Equivalent to oldest > after+1, without overflowing after+1.
	if oldest > 0 && after < oldest-1 {
		result.Gap = true
		return result, nil
	}

	for i := 0; i < r.count; i++ {
		index := (r.start + i) % len(r.history)
		event := r.history[index]

		if event.Revision > after {
			result.Events = append(result.Events, event)
		}
	}

	return result, nil
}
