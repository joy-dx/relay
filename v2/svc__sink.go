package relay

import (
	"fmt"

	"github.com/joy-dx/relay/v2/dto"
)

// sinkSnapshot returns the currently registered sinks.
//
// The slice is copied while holding the read lock so callers can dispatch
// without holding r.mu. Sink implementations are application code and must
// not be called while the relay lock is held.
func (r *RelaySvc) sinkSnapshot() []dto.RelaySinkInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sinks := make([]dto.RelaySinkInterface, len(r.sinks))
	copy(sinks, r.sinks)
	return sinks
}

func (r *RelaySvc) RegisterSink(sink dto.RelaySinkInterface) {
	if sink == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sinks = append(r.sinks, sink)
}

func (r *RelaySvc) UnregisterSink(sinkRef string) error {
	r.mu.Lock()

	var sink dto.RelaySinkInterface
	for i, candidate := range r.sinks {
		if candidate.Ref() != sinkRef {
			continue
		}

		sink = candidate

		// Remove the sink while holding the lock.
		copy(r.sinks[i:], r.sinks[i+1:])
		r.sinks[len(r.sinks)-1] = nil // Release references, if Sink is an interface.
		r.sinks = r.sinks[:len(r.sinks)-1]

		break
	}

	r.mu.Unlock()

	if sink == nil {
		return fmt.Errorf("%s: %w", sinkRef, dto.ErrSinkNotFound)
	}

	if err := sink.Close(); err != nil {
		return fmt.Errorf("%s: unable to close sink: %w", sinkRef, err)
	}

	return nil
}
