package pump

import (
	"drainnet/internal/audit"
	"drainnet/internal/station"
)

func (c *Controller) StartPump(stationID string, pumpID string, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	ref, err := c.StateOf(stationID, pumpID)
	if err != nil {
		return err
	}
	if ref.State == "running" {
		return ErrPumpAlreadyRunning
	}
	free, _, err := c.stations.PoolStats(stationID)
	if err != nil {
		return err
	}
	if free <= 0 {
		return station.ErrNoBreaker
	}
	breakerID, err := c.stations.AllocateBreaker(stationID, pumpID)
	if err != nil {
		return err
	}
	ref.State = "running"
	ref.BreakerID = breakerID
	ref.Speed = "low"
	if _, err := c.stations.SetPumpState(stationID, ref); err != nil {
		c.stations.ReleaseBreaker(stationID, breakerID)
		return err
	}
	return c.audits.Record(audit.Event{
		Type:     "pump.start",
		EntityID: pumpID,
		Message:  "pump started",
		Meta:     map[string]string{"station": stationID, "breaker": breakerID, "reason": reason},
	})
}

func (c *Controller) StopPump(stationID string, pumpID string, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	ref, err := c.StateOf(stationID, pumpID)
	if err != nil {
		return err
	}
	if ref.State != "running" {
		return nil
	}
	if ref.BreakerID != "" {
		if err := c.stations.ReleaseBreaker(stationID, ref.BreakerID); err != nil {
			return err
		}
	}
	ref.State = "standby"
	ref.BreakerID = ""
	ref.Speed = ""
	if _, err := c.stations.SetPumpState(stationID, ref); err != nil {
		return err
	}
	return c.audits.Record(audit.Event{
		Type:     "pump.stop",
		EntityID: pumpID,
		Message:  "pump stopped",
		Meta:     map[string]string{"station": stationID, "reason": reason},
	})
}

func (c *Controller) startStandby(stationID string, ref station.PumpRef, breakerID string, reason string) error {
	ref.State = "running"
	ref.BreakerID = breakerID
	ref.Speed = "low"
	if _, err := c.stations.SetPumpState(stationID, ref); err != nil {
		return err
	}
	return c.audits.Record(audit.Event{
		Type:     "pump.start",
		EntityID: ref.ID,
		Message:  "standby pump started",
		Meta:     map[string]string{"station": stationID, "breaker": breakerID, "reason": reason},
	})
}
