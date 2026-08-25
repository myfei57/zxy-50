package sluice

import (
	"drainnet/internal/audit"
)

func (r *Registry) OpenChannel(channelID string, plan []string) ([]Gate, error) {
	gates, err := r.Gates(channelID)
	if err != nil {
		return nil, err
	}
	byID := map[string]Gate{}
	for _, gate := range gates {
		byID[gate.ID] = gate
	}
	opened := make([]Gate, 0, len(plan))
	for _, gateID := range plan {
		gate, ok := byID[gateID]
		if !ok {
			return nil, ErrGateNotFound
		}
		if err := r.SetPosition(gateID, 100); err != nil {
			return nil, err
		}
		opened = append(opened, gate)
	}
	if len(opened) > 0 {
		if err := r.audit.Record(audit.Event{
			Type:     "sluice.open",
			EntityID: channelID,
			Message:  "channel gates opened in order",
			Meta:     map[string]string{"plan": joinIDs(plan)},
		}); err != nil {
			return nil, err
		}
	}
	return opened, nil
}

func joinIDs(ids []string) string {
	out := ""
	for index, id := range ids {
		if index > 0 {
			out += ","
		}
		out += id
	}
	return out
}
