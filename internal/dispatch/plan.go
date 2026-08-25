package dispatch

import (
	"drainnet/internal/audit"
)

func (d *Dispatcher) Plan(channelID string) ([]string, error) {
	plan, err := d.GatePlan(channelID)
	if err != nil {
		return nil, err
	}
	if err := d.audits.Record(audit.Event{
		Type:     "dispatch.plan",
		EntityID: channelID,
		Message:  "discharge plan reviewed",
		Meta:     map[string]string{"order": joinPlan(plan)},
	}); err != nil {
		return nil, err
	}
	return plan, nil
}

func joinPlan(plan []string) string {
	out := ""
	for index, id := range plan {
		if index > 0 {
			out += ";"
		}
		out += id
	}
	return out
}
