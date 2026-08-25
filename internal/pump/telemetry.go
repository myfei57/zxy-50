package pump

func SpeedLabel(speed string) string {
	if speed == "" {
		return "low"
	}
	return speed
}

func (c *Controller) Summary(stationID string) (map[string]int, error) {
	refs, err := c.List(stationID)
	if err != nil {
		return nil, err
	}
	summary := map[string]int{"running": 0, "standby": 0, "tripped": 0}
	for _, ref := range refs {
		summary[ref.State]++
	}
	return summary, nil
}

func (c *Controller) RecordSnapshot(stationID string) error {
	row, err := c.Telemetry(stationID)
	if err != nil {
		return err
	}
	return c.store.WriteJSON("pump_telemetry", stationID, row)
}
