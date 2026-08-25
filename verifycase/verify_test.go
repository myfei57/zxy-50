package verifycase

import (
	"testing"
	"time"

	"drainnet/internal/audit"
	"drainnet/internal/dispatch"
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

func TestDnRainPeakWindowAccumulate(t *testing.T) {
	dir := t.TempDir()
	st := store.New(dir)
	audits := audit.NewService(st)
	nsSvc := ns.NewService(st, audits)
	stations := station.NewRegistry(st, audits)
	levels := level.NewService(st)
	pumps := pump.NewController(st, stations, audits, levels)
	sluices := sluice.NewRegistry(st, audits)
	rains := rain.NewService(st)
	policies := policy.NewService(st, levels)
	quotas := quota.NewService(st)
	dispatcher := dispatch.NewDispatcher(st, nsSvc, stations, pumps, sluices, rains, levels, policies, quotas, audits)
	registered, err := stations.Register(station.Station{Name: "S1", Breakers: []station.Breaker{{ID: "b1", Name: "B1"}}})
	if err != nil {
		t.Fatal(err)
	}
	lead, err := pumps.Register(registered.ID, station.PumpRef{Name: "P1", Duty: "lead", State: "standby"})
	if err != nil {
		t.Fatal(err)
	}
	_ = lead
	gauge, err := rains.RegisterGauge(rain.Gauge{StationID: registered.ID, Name: "G1", Offset: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := policies.SetRules(policy.Rules{
		PeakRainThreshold: 15,
		PeakWindow:        10 * time.Minute,
		FullSpeedRate:     5,
		ReverseFlowDelta:  8,
		GateRule:          "downstream-first",
	}); err != nil {
		t.Fatal(err)
	}
	morning := time.Date(2026, 8, 25, 9, 59, 0, 0, time.UTC)
	pushes := []time.Time{
		morning.Add(50 * time.Second),
		morning.Add(55 * time.Second),
		morning.Add(60 * time.Second),
		morning.Add(65 * time.Second),
	}
	for _, at := range pushes {
		if _, err := rains.Accumulate(gauge.ID, 5, at); err != nil {
			t.Fatal(err)
		}
	}
	triggered, err := dispatcher.PreLower(registered.ID, gauge.ID, morning.Add(70*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if !triggered {
		t.Fatal("rain peak across a wall-clock boundary must trigger pre-lowering")
	}
}
