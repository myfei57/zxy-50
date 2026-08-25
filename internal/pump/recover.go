package pump

import (
	"drainnet/internal/audit"
)

func (c *Controller) Recover(stationID string) error {
	value, err := c.stations.Get(stationID)
	if err != nil {
		return err
	}
	for _, ref := range value.Pumps {
		if ref.State == "running" {
			if err := c.StopPump(stationID, ref.ID, "recover"); err != nil {
				return err
			}
		}
	}
	return c.audits.Record(audit.Event{
		Type:     "pump.recover",
		EntityID: stationID,
		Message:  "pumps returned to normal mode",
	})
}
