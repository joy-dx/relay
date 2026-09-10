package relay

import (
	"errors"
)

func (s *RelaySvc) Hydrate() error {
	if s.cfg == nil {
		return errors.New("Relay cfg required but is nil")
	}
	return nil
}
