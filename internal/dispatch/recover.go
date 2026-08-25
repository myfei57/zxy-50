package dispatch

import (
	"drainnet/internal/audit"
)

func (d *Dispatcher) Recover(stationID string) error {
	if _, err := d.queue.Replay(); err != nil {
		return err
	}
	if err := d.pumps.Recover(stationID); err != nil {
		return err
	}
	return d.audits.Record(audit.Event{
		Type:     "dispatch.recover",
		EntityID: stationID,
		Message:  "post-rain recovery finished",
	})
}
