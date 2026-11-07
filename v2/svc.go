package relay

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/joy-dx/relay/v2/config"
	"github.com/joy-dx/relay/v2/dto"
)

// RelaySvc Is a pseudo logger interface that allows for rich structs to be included alongside messsages. it is
// up to the developer / configuration to subscribe to the relay as sinks
type RelaySvc struct {
	sinks    []dto.RelaySinkInterface
	cfg      *config.RelaySvcConfig
	mu       sync.RWMutex
	revision uint64

	history []dto.EmittedEvent
	start   int
	count   int
}

func (r *RelaySvc) Close() error {
	var closeErr error
	for _, sink := range r.sinkSnapshot() {
		if err := sink.Close(); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}
	return closeErr
}

func (r *RelaySvc) Emit(level dto.RelayLevel, event dto.RelayEventInterface) {

	emittedEvent := dto.EmittedEvent{
		Level:   level,
		Event:   event,
		Channel: event.RelayChannel(),
		Type:    event.RelayType(),
	}

	r.mu.Lock()

	r.revision++

	emittedEvent.Time = time.Now()
	emittedEvent.Revision = r.revision

	r.appendHistoryLocked(emittedEvent)
	r.mu.Unlock()

	for _, sink := range r.sinkSnapshot() {
		sink.Emit(emittedEvent)
	}

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
