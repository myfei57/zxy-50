package verifycase

import (
	"testing"
	"time"

	"drainnet/internal/audit"
	"drainnet/internal/level"
	"drainnet/internal/policy"
	"drainnet/internal/store"
)

func TestDnFlowReverseOrder(t *testing.T) {
	dir := t.TempDir()
	st := store.New(dir)
	audits := audit.NewService(st)
	levels := level.NewService(st)
	policies := policy.NewService(st, levels)
	_ = audits
	base := time.Date(2026, 8, 25, 15, 0, 0, 0, time.UTC)
	samples := []float64{100, 110, 125}
	for index, raw := range samples {
		if _, err := levels.Sample("s1", raw, base.Add(time.Duration(index)*time.Minute)); err != nil {
			t.Fatal(err)
		}
		reverse, err := policies.CheckReverse("s1")
		if err != nil {
			t.Fatal(err)
		}
		if index >= 1 && !reverse {
			t.Fatalf("a rising tide must be judged as reverse flow right after sample %d", index)
		}
	}
}
