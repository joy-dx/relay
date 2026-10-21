package relay

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joy-dx/relay/v2/config"
	"github.com/joy-dx/relay/v2/dto"
)

// RelaySvc Is a pseudo logger interface that allows for rich structs to be included alongside messsages. it is
// up to the developer / configuration to subscribe to the relay as sinks
type RelaySvc struct {
	sinks []dto.RelaySinkInterface
	cfg   *config.RelaySvcConfig
	mu    sync.RWMutex
}

func (r *RelaySvc) RegisterSink(sink dto.RelaySinkInterface) {
	r.mu.Lock()
	r.sinks = append(r.sinks, sink)
	r.mu.Unlock()
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
		return nil
	}

	if err := sink.Close(); err != nil {
		return fmt.Errorf("%s: unable to close sink: %w", sinkRef, err)
	}

	return nil
}

func (r *RelaySvc) Close() error {
	var closeErr error
	for _, sink := range r.sinks {
		if err := sink.Close(); err != nil {
			closeErr = err
		}
	}
	return closeErr
}

func (r *RelaySvc) Emit(level dto.RelayLevel, event dto.RelayEventInterface) {

	emittedEvent := dto.EmittedEvent{
		Time:  time.Now(),
		Level: level,
		Event: event,
		Type:  event.RelayType(),
	}

	// dispatch to registered sinks
	for _, sink := range r.sinks {
		sink.Emit(emittedEvent)
	}
	// After draining all the sinks, exit if fatal
	if level == dto.Fatal {
		os.Exit(1)
	}
}

func (r *RelaySvc) Debug(e dto.RelayEventInterface) { r.Emit(dto.Debug, e) }
func (r *RelaySvc) Info(e dto.RelayEventInterface)  { r.Emit(dto.Info, e) }
func (r *RelaySvc) Warn(e dto.RelayEventInterface)  { r.Emit(dto.Warn, e) }
func (r *RelaySvc) Error(e dto.RelayEventInterface) { r.Emit(dto.Error, e) }
func (r *RelaySvc) Fatal(e dto.RelayEventInterface) { r.Emit(dto.Fatal, e) }

// Meta A special handler for
func (r *RelaySvc) Meta(e dto.RelayEventInterface) { r.Emit(dto.Meta, e) }

func (r *RelaySvc) EmitWithOpts(
	level dto.RelayLevel,
	event dto.RelayEventInterface,
	opts dto.RlyOpts,
) {
	r.Emit(level, dto.ApplyRelayOpts(event, opts))
}

func (r *RelaySvc) DebugWithOpts(
	event dto.RelayEventInterface,
	opts dto.RlyOpts,
) {
	r.EmitWithOpts(dto.Debug, event, opts)
}

func (r *RelaySvc) InfoWithOpts(
	event dto.RelayEventInterface,
	opts dto.RlyOpts,
) {
	r.EmitWithOpts(dto.Info, event, opts)
}

func (r *RelaySvc) WarnWithOpts(
	event dto.RelayEventInterface,
	opts dto.RlyOpts,
) {
	r.EmitWithOpts(dto.Warn, event, opts)
}

func (r *RelaySvc) ErrorWithOpts(
	event dto.RelayEventInterface,
	opts dto.RlyOpts,
) {
	r.EmitWithOpts(dto.Error, event, opts)
}

func (r *RelaySvc) FatalWithOpts(
	event dto.RelayEventInterface,
	opts dto.RlyOpts,
) {
	r.EmitWithOpts(dto.Fatal, event, opts)
}

func (r *RelaySvc) MetaWithOpts(
	event dto.RelayEventInterface,
	opts dto.RlyOpts,
) {
	r.EmitWithOpts(dto.Meta, event, opts)
}
