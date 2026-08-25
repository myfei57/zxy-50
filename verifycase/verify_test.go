package verifycase

import (
	"strings"
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

func TestDnSluiceOpenOrder(t *testing.T) {
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
	upstream, err := sluices.Register(sluice.Gate{ChannelID: "ch1", Name: "上游闸", Side: "upstream"})
	if err != nil {
		t.Fatal(err)
	}
	downstream, err := sluices.Register(sluice.Gate{ChannelID: "ch1", Name: "下游闸", Side: "downstream"})
	if err != nil {
		t.Fatal(err)
	}
	if err := dispatcher.OpenForPreDrain("ch1"); err != nil {
		t.Fatal(err)
	}
	events, err := audits.List()
	if err != nil {
		t.Fatal(err)
	}
	opens := make([]audit.Event, 0, 2)
	for _, event := range events {
		if event.Type == "sluice.open" {
			opens = append(opens, event)
		}
	}
	if len(opens) == 0 {
		t.Fatal("no sluice open order was recorded")
	}
	order := strings.Split(opens[0].Meta["plan"], ",")
	if len(order) < 2 || order[0] != downstream.ID {
		t.Fatalf("downstream gate must open before upstream, plan=%v upstream=%s downstream=%s", order, upstream.ID, downstream.ID)
	}
}
