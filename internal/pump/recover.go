package pump

import (
	"drainnet/internal/audit"
)

func (c *Controller) Recover(stationID string) error {
	lead, err := c.stations.LeadOf(stationID)
	if err != nil {
		return err
	}
	if err := c.StartPump(stationID, lead.ID, "recover"); err != nil {
		return err
	}
	return c.audits.Record(audit.Event{
		Type:     "pump.recover",
		EntityID: stationID,
		Message:  "pumps returned to normal mode",
	})
}
