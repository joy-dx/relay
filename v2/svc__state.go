package relay

import (
	"errors"
)

func (r *RelaySvc) Hydrate() error {
	if r.cfg == nil {
		return errors.New("Relay cfg required but is nil")
	}
	return nil
}
