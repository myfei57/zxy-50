package dispatch

import (
	"drainnet/internal/audit"
)

func (d *Dispatcher) ManualGate(stationID string, gateID string, target int) error {
	current, err := d.sluices.Position(gateID)
	if err != nil {
		return err
	}
	if err := d.sluices.ApplyDelta(gateID, target-current); err != nil {
		return err
	}
	return d.audits.Record(audit.Event{
		Type:     "dispatch.manual_gate",
		EntityID: gateID,
		Message:  "manual gate command applied",
		Meta:     map[string]string{"station": stationID, "target": formatInt(target)},
	})
}

func formatInt(value int) string {
	if value == 0 {
		return "0"
	}
	if value < 0 {
		return "-" + formatInt(-value)
	}
	return formatCount(value)
}
