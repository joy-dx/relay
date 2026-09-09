package relay

import (
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
	r.sinks = append(r.sinks, sink)
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

func (r *RelaySvc) emit(level dto.RelayLevel, event dto.RelayEventInterface) {

	emittedEvent := dto.EmittedEvent{
		Time:  time.Now(),
		Level: level,
		Event: event,
	}

	// dispatch to registered sinks
	for _, sink := range r.sinks {
		switch level {
		case dto.Debug:
			sink.Debug(emittedEvent)
		case dto.Info:
			sink.Info(emittedEvent)
		case dto.Warn:
			sink.Warn(emittedEvent)
		case dto.Error:
			sink.Error(emittedEvent)
		case dto.Fatal:
			sink.Fatal(emittedEvent)
		case dto.Meta:
			sink.Meta(emittedEvent)
		}
	}
	// After draining all the sinks, exit if fatal
	if level == dto.Fatal {
		os.Exit(1)
	}
}

func (r *RelaySvc) Debug(e dto.RelayEventInterface) { r.emit(dto.Debug, e) }
func (r *RelaySvc) Info(e dto.RelayEventInterface)  { r.emit(dto.Info, e) }
func (r *RelaySvc) Warn(e dto.RelayEventInterface)  { r.emit(dto.Warn, e) }
func (r *RelaySvc) Error(e dto.RelayEventInterface) { r.emit(dto.Error, e) }
func (r *RelaySvc) Fatal(e dto.RelayEventInterface) { r.emit(dto.Fatal, e) }

// Meta A special handler for
func (r *RelaySvc) Meta(e dto.RelayEventInterface) { r.emit(dto.Meta, e) }
