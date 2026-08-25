package dispatch

import (
	"time"

	"drainnet/internal/audit"
	"drainnet/internal/station"
)

type ZoneRouter struct {
	stations *station.Registry
}

func NewZoneRouter(stations *station.Registry) *ZoneRouter {
	return &ZoneRouter{stations: stations}
}

func (z *ZoneRouter) Resolve(zoneID string) (string, error) {
	return z.stations.StationForZone(zoneID)
}

func (d *Dispatcher) DispatchRain(zoneID string, mm float64, at time.Time) error {
	stationID, err := d.router.Resolve(zoneID)
	if err != nil {
		return err
	}
	if err := d.ReserveQuota(stationID, mm); err != nil {
		return err
	}
	gauges, err := d.rains.GaugesForStation(stationID)
	if err != nil {
		return err
	}
	started := false
	for _, gauge := range gauges {
		if _, err := d.rains.Accumulate(gauge.ID, mm, at); err != nil {
			return err
		}
		peak, err := d.rains.PeakExceeded(gauge.ID, d.policies.Rules().PeakRainThreshold, at, d.policies.Rules().PeakWindow)
		if err != nil {
			return err
		}
		if peak {
			lead, err := d.stations.LeadOf(stationID)
			if err != nil {
				return err
			}
			if err := d.pumps.StartPump(stationID, lead.ID, "rain"); err != nil {
				return err
			}
			started = true
		}
	}
	return d.audits.Record(audit.Event{
		Type:     "dispatch.rain",
		EntityID: zoneID,
		Message:  "rain dispatched",
		Meta:     map[string]string{"station": stationID, "mm": formatFloat(mm), "started": formatBool(started)},
	})
}

func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
