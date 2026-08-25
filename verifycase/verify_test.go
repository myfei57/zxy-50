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

func TestDnStationMappingFresh(t *testing.T) {
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
	catchment, err := nsSvc.CreateCatchment("东城")
	if err != nil {
		t.Fatal(err)
	}
	z1, err := nsSvc.AddZone(catchment.ID, ns.Zone{Name: "东一区"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := stations.Register(station.Station{Name: "一号泵站", CatchmentID: catchment.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := nsSvc.AssignStation(z1.ID, s1.ID); err != nil {
		t.Fatal(err)
	}
	zones, err := nsSvc.Zones(catchment.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := stations.RefreshMapping(s1.ID, zones); err != nil {
		t.Fatal(err)
	}
	s2, err := stations.Register(station.Station{Name: "二号泵站", CatchmentID: catchment.ID})
	if err != nil {
		t.Fatal(err)
	}
	catchment2, err := nsSvc.SplitCatchment(catchment.ID, ns.Zone{Name: "东二区", StationID: s2.ID})
	if err != nil {
		t.Fatal(err)
	}
	_ = catchment2
	zonesAfter, err := nsSvc.Zones(catchment.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := stations.RefreshMapping(s1.ID, zonesAfter); err != nil {
		t.Fatal(err)
	}
	if err := stations.RefreshMapping(s2.ID, zonesAfter); err != nil {
		t.Fatal(err)
	}
	var z2 ns.Zone
	for _, zone := range zonesAfter {
		if zone.StationID == s2.ID {
			z2 = zone
			break
		}
	}
	if z2.ID == "" {
		t.Fatal("split zone was not assigned to the new station")
	}
	if z2.StationID != s2.ID {
		t.Fatalf("test zone must be assigned to the new station, got %s", z2.StationID)
	}
	if err := dispatcher.DispatchRain(z2.ID, 30, time.Now()); err != nil {
		t.Fatal(err)
	}
	events, err := audits.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == "dispatch.rain" && event.EntityID == z2.ID {
			if event.Meta["station"] != s2.ID {
				t.Fatalf("rain for the split zone must dispatch to %s, got %s", s2.ID, event.Meta["station"])
			}
			return
		}
	}
	t.Fatal("rain dispatch was not recorded")
}
