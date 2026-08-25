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

func TestDnRainGaugeOffsetFresh(t *testing.T) {
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
		PeakRainThreshold: 70,
		PeakWindow:        10 * time.Minute,
		FullSpeedRate:     5,
		ReverseFlowDelta:  8,
		GateRule:          "downstream-first",
	}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if _, err := rains.Accumulate(gauge.ID, 40, now); err != nil {
		t.Fatal(err)
	}
	if err := rains.Calibrate(gauge.ID, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := rains.Accumulate(gauge.ID, 20, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	accumulated, err := rains.Accumulated(gauge.ID)
	if err != nil {
		t.Fatal(err)
	}
	if accumulated != 70 {
		t.Fatalf("accumulated rainfall must use the calibrated offset, got %.1f", accumulated)
	}
	triggered, err := dispatcher.PreLower(registered.ID, gauge.ID, now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !triggered {
		t.Fatal("pre-lowering must start when the calibrated total reaches the threshold")
	}
}
