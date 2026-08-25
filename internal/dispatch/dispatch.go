package dispatch

import (
	"time"

	"drainnet/internal/audit"
	"drainnet/internal/level"
	"drainnet/internal/ns"
	"drainnet/internal/policy"
	"drainnet/internal/pump"
	"drainnet/internal/quota"
	"drainnet/internal/rain"
	"drainnet/internal/sluice"
	"drainnet/internal/station"
	"drainnet/internal/store"
)

type Command struct {
	ID        string    `json:"id"`
	StationID string    `json:"station_id"`
	Kind      string    `json:"kind"`
	Mode      string    `json:"mode"`
	At        time.Time `json:"at"`
}

type Dispatcher struct {
	ns       *ns.Service
	stations *station.Registry
	pumps    *pump.Controller
	sluices  *sluice.Registry
	rains    *rain.Service
	levels   *level.Service
	policies *policy.Service
	quotas   *quota.Service
	audits   *audit.Service
	router   *ZoneRouter
	queue    *CommandQueue
	lastGatePosition map[string]int
}

func NewDispatcher(
	st *store.Store,
	nsSvc *ns.Service,
	stations *station.Registry,
	pumps *pump.Controller,
	sluices *sluice.Registry,
	rains *rain.Service,
	levels *level.Service,
	policies *policy.Service,
	quotas *quota.Service,
	audits *audit.Service,
) *Dispatcher {
	return &Dispatcher{
		ns:       nsSvc,
		stations: stations,
		pumps:    pumps,
		sluices:  sluices,
		rains:    rains,
		levels:   levels,
		policies: policies,
		quotas:   quotas,
		audits:   audits,
		router:   NewZoneRouter(stations),
		queue:    NewCommandQueue(st),
		lastGatePosition: map[string]int{},
	}
}

func (d *Dispatcher) ReserveQuota(stationID string, amount float64) error {
	if err := d.quotas.Reserve(stationID, amount); err != nil {
		return d.audits.Record(audit.Event{
			Type:     "dispatch.quota_denied",
			EntityID: stationID,
			Message:  "quota reserve rejected",
			Meta:     map[string]string{"amount": formatFloat(amount)},
		})
	}
	return nil
}

func (d *Dispatcher) HandleLevelSample(stationID string, rawMM float64, at time.Time) error {
	sample, err := d.levels.Sample(stationID, rawMM, at)
	if err != nil {
		return err
	}
	return d.pumps.HandleLevel(sample)
}

func (d *Dispatcher) Status() (map[string]int, error) {
	stations, err := d.stations.List()
	if err != nil {
		return nil, err
	}
	status := map[string]int{"stations": len(stations), "pumps": 0, "running": 0}
	for _, value := range stations {
		summary, err := d.pumps.Summary(value.ID)
		if err != nil {
			return nil, err
		}
		status["pumps"] += summary["running"] + summary["standby"] + summary["tripped"]
		status["running"] += summary["running"]
	}
	return status, nil
}

func formatFloat(value float64) string {
	if value == float64(int64(value)) {
		return strconvInt(int64(value))
	}
	return strconvFloat(value)
}
