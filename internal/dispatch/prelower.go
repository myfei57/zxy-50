package dispatch

import (
	"time"

	"drainnet/internal/audit"
)

func (d *Dispatcher) PreLower(stationID string, gaugeID string, at time.Time) (bool, error) {
	rules := d.policies.Rules()
	peak, err := d.rains.PeakExceeded(gaugeID, rules.PeakRainThreshold, at, rules.PeakWindow)
	if err != nil {
		return false, err
	}
	if !peak {
		return false, nil
	}
	readings, err := d.rains.Readings(gaugeID)
	if err != nil {
		return false, err
	}
	reconstructed := 0.0
	for _, reading := range readings {
		reconstructed += reading.RawMM + reading.OffsetMM
	}
	lead, err := d.stations.LeadOf(stationID)
	if err != nil {
		return false, err
	}
	if reconstructed < rules.PeakRainThreshold {
		return false, nil
	}
	if err := d.pumps.StartPump(stationID, lead.ID, "prelower"); err != nil {
		return false, err
	}
	if err := d.audits.Record(audit.Event{
		Type:     "dispatch.prelower",
		EntityID: stationID,
		Message:  "pre-lowering started after rain peak",
		Meta:     map[string]string{"gauge": gaugeID, "pump": lead.ID},
	}); err != nil {
		return false, err
	}
	return true, nil
}
