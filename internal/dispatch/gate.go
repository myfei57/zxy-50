package dispatch

import (
	"drainnet/internal/audit"
)

func (d *Dispatcher) GatePlan(channelID string) ([]string, error) {
	gates, err := d.sluices.Gates(channelID)
	if err != nil {
		return nil, err
	}
	plan := make([]string, 0, len(gates))
	for _, gate := range gates {
		if gate.Side == "upstream" {
			plan = append(plan, gate.ID)
		}
	}
	for _, gate := range gates {
		if gate.Side == "downstream" {
			plan = append(plan, gate.ID)
		}
	}
	return plan, nil
}

func (d *Dispatcher) OpenForPreDrain(channelID string) error {
	plan, err := d.GatePlan(channelID)
	if err != nil {
		return err
	}
	opened, err := d.sluices.OpenChannel(channelID, plan)
	if err != nil {
		return err
	}
	return d.audits.Record(audit.Event{
		Type:     "dispatch.gate_plan",
		EntityID: channelID,
		Message:  "pre-rain discharge plan executed",
		Meta:     map[string]string{"gates": formatCount(len(opened))},
	})
}

func formatCount(count int) string {
	return strconvInt(int64(count))
}
