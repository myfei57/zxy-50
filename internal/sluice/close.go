package sluice

import (
	"drainnet/internal/audit"
)

func (r *Registry) CloseGate(gateID string) error {
	if err := r.SetPosition(gateID, 0); err != nil {
		return err
	}
	return r.audit.Record(audit.Event{
		Type:     "sluice.close",
		EntityID: gateID,
		Message:  "gate closed",
	})
}
