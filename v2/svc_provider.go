package relay

import (
	"fmt"
	"sync"

	"github.com/joy-dx/relay/v2/config"
	"github.com/joy-dx/relay/v2/dto"
)

var (
	service     *RelaySvc
	serviceErr  error
	serviceOnce sync.Once
)

func ProvideRelaySvc(
	cfg *config.RelaySvcConfig,
) (*RelaySvc, error) {
	serviceOnce.Do(func() {
		if cfg == nil {
			serviceErr = fmt.Errorf("relay service config must not be nil")
			return
		}

		if cfg.HistorySize < 0 {
			serviceErr = fmt.Errorf(
				"HistorySize must not be negative: %d",
				cfg.HistorySize,
			)
			return
		}

		service = &RelaySvc{
			cfg:     cfg,
			sinks:   make([]dto.RelaySinkInterface, 0),
			history: make([]dto.EmittedEvent, cfg.HistorySize),
		}
	})

	return service, serviceErr
}
