package verifycase

import (
	"sync"
	"testing"

	"drainnet/internal/audit"
	"drainnet/internal/level"
	"drainnet/internal/pump"
	"drainnet/internal/station"
	"drainnet/internal/store"
)

func TestDnConcurrentPumpBreakerAlloc(t *testing.T) {
	dir := t.TempDir()
	st := store.New(dir)
	audits := audit.NewService(st)
	stations := station.NewRegistry(st, audits)
	levels := level.NewService(st)
	pumps := pump.NewController(st, stations, audits, levels)
	registered, err := stations.Register(station.Station{
		Name: "S1",
		Breakers: []station.Breaker{
			{ID: "b1", Name: "B1"},
			{ID: "b2", Name: "B2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	sid := registered.ID
	p1, err := pumps.Register(sid, station.PumpRef{Name: "P1", Duty: "lead", State: "standby"})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := pumps.Register(sid, station.PumpRef{Name: "P2", Duty: "standby", State: "standby"})
	if err != nil {
		t.Fatal(err)
	}
	p3, err := pumps.Register(sid, station.PumpRef{Name: "P3", Duty: "standby", State: "standby"})
	if err != nil {
		t.Fatal(err)
	}
	for round := 0; round < 20; round++ {
		start := make(chan struct{})
		var wg sync.WaitGroup
		errs := make(chan error, 3)
		for _, pumpID := range []string{p1.ID, p2.ID, p3.ID} {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				<-start
				if err := pumps.StartPump(sid, id, "storm"); err != nil && err != station.ErrNoBreaker {
					errs <- err
				}
			}(pumpID)
		}
		close(start)
		wg.Wait()
		close(errs)
		for err := range errs {
			t.Fatal(err)
		}
		value, err := stations.Get(sid)
		if err != nil {
			t.Fatal(err)
		}
		holders := map[string]int{}
		for _, ref := range value.Pumps {
			if ref.BreakerID != "" {
				holders[ref.BreakerID]++
			}
		}
		for breakerID, count := range holders {
			if count > 1 {
				t.Fatalf("breaker %s allocated to %d pumps in round %d", breakerID, count, round)
			}
		}
		for _, pumpID := range []string{p1.ID, p2.ID, p3.ID} {
			_ = pumps.StopPump(sid, pumpID, "reset")
		}
	}
}
