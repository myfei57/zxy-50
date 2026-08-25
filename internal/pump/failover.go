package pump

import (
	"drainnet/internal/audit"
	"drainnet/internal/station"
)

func (c *Controller) Failover(stationID string, primaryID string, snapshot station.LeadSnapshot) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, err := c.stations.Get(stationID)
	if err != nil {
		return err
	}
	found := false
	for _, ref := range value.Pumps {
		if ref.ID == primaryID {
			found = true
			break
		}
	}
	if !found {
		return ErrPumpNotFound
	}
	state := c.stations.LeadStateFor(stationID)
	if state == nil {
		return station.ErrStationNotFound
	}
	current := state.Current()
	if current != primaryID {
		return nil
	}
	standbys, err := c.stations.StandbysOf(stationID)
	if err != nil {
		return err
	}
	if len(standbys) == 0 {
		return ErrNoStandby
	}
	ref := standbys[0]
	if ref.State == "running" {
		return ErrPumpAlreadyRunning
	}
	breakerID, err := c.stations.AllocateBreaker(stationID, ref.ID)
	if err != nil {
		return err
	}
	if err := c.startStandby(stationID, ref, breakerID, "failover"); err != nil {
		c.stations.ReleaseBreaker(stationID, breakerID)
		return err
	}
	return c.audits.Record(audit.Event{
		Type:     "pump.failover",
		EntityID: ref.ID,
		Message:  "standby took over after primary trip",
		Meta:     map[string]string{"station": stationID, "primary": primaryID, "breaker": breakerID},
	})
}
