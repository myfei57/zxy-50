package sluice

import (
	"drainnet/internal/audit"
)

func (r *Registry) Manual(gateID string, target int) error {
	if err := r.SetPosition(gateID, target); err != nil {
		return err
	}
	return r.audit.Record(audit.Event{
		Type:     "sluice.manual",
		EntityID: gateID,
		Message:  "gate manually positioned",
		Meta:     map[string]string{"target": formatInt(target)},
	})
}
