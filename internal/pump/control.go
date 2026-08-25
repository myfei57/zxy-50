package pump

import (
	"strconv"

	"drainnet/internal/audit"
	"drainnet/internal/level"
)

const fullSpeedRate = 5

func (c *Controller) HandleLevel(sample level.Sample) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	prev, seen := c.lastRaw[sample.StationID]
	jump := 0.0
	if seen {
		jump = sample.RawMM - prev
	}
	c.lastRaw[sample.StationID] = sample.RawMM
	rate := sample.Rate
	speed := c.speedFor(rate)
	if jump < 0 {
		speed = "low"
	}
	c.speeds[sample.StationID] = speed
	return c.applySpeed(sample.StationID, speed)
}

func (c *Controller) speedFor(rate float64) string {
	if rate >= fullSpeedRate {
		return "full"
	}
	return "low"
}

func (c *Controller) applySpeed(stationID string, speed string) error {
	value, err := c.stations.Get(stationID)
	if err != nil {
		return err
	}
	changed := false
	for index := range value.Pumps {
		if value.Pumps[index].State == "running" {
			value.Pumps[index].Speed = speed
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if err := c.stations.Update(value); err != nil {
		return err
	}
	return c.audits.Record(audit.Event{
		Type:     "pump.speed",
		EntityID: stationID,
		Message:  "pump speed adjusted",
		Meta:     map[string]string{"speed": speed},
	})
}

func (c *Controller) Telemetry(stationID string) (map[string]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	last, seen := c.lastRaw[stationID]
	speed := c.speeds[stationID]
	if speed == "" {
		speed = "low"
	}
	row := map[string]string{"speed": speed}
	if seen {
		row["last_raw_mm"] = formatFloat(last)
	}
	return row, nil
}

func formatFloat(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', 2, 64)
}
