package verifycase

import (
	"testing"

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

func TestDnNoReplayAfterRain(t *testing.T) {
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
	if err := dispatcher.EnqueueCommand(dispatch.Command{
		StationID: registered.ID,
		Kind:      "pump_start",
		Mode:      "storm",
	}); err != nil {
		t.Fatal(err)
	}
	if err := dispatcher.Recover(registered.ID); err != nil {
		t.Fatal(err)
	}
	pending, err := dispatcher.PendingCommands()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("storm commands must be cleared after recovery, %d pending", len(pending))
	}
	ref, err := pumps.StateOf(registered.ID, lead.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.State == "running" {
		t.Fatal("recovery must not replay storm commands onto the channel")
	}
}
