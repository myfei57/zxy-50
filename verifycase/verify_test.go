package verifycase

import (
	"sync"
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

func TestDnConcurrentSluicePosition(t *testing.T) {
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
	registered, err := stations.Register(station.Station{Name: "S1"})
	if err != nil {
		t.Fatal(err)
	}
	gate, err := sluices.Register(sluice.Gate{ChannelID: "ch1", Name: "闸1", Side: "downstream"})
	if err != nil {
		t.Fatal(err)
	}
	for round := 0; round < 50; round++ {
		if err := sluices.SetPosition(gate.ID, 0); err != nil {
			t.Fatal(err)
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			for step := 0; step < 20; step++ {
				_ = sluices.SetPosition(gate.ID, 100)
			}
		}()
		go func() {
			defer wg.Done()
			<-start
			for step := 0; step < 20; step++ {
				_ = dispatcher.ManualGate(registered.ID, gate.ID, 0)
			}
		}()
		close(start)
		wg.Wait()
		position, err := sluices.Position(gate.ID)
		if err != nil {
			t.Fatal(err)
		}
		if position != 0 && position != 100 {
			t.Fatalf("gate must never rest half open, got %d in round %d", position, round)
		}
	}
}
