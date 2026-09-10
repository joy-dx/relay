package relay

import (
	"sync"

	"github.com/joy-dx/relay/config"
	"github.com/joy-dx/relay/dto"
)

var (
	service     *RelaySvc
	serviceOnce sync.Once
)

func ProvideRelaySvc(cfg *config.RelaySvcConfig) *RelaySvc {
	serviceOnce.Do(func() {
		service = &RelaySvc{
			cfg:   cfg,
			sinks: make([]dto.RelaySinkInterface, 0),
		}
	})
	return service
}
