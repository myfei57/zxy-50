package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"drainnet/internal/audit"
	"drainnet/internal/config"
	"drainnet/internal/console"
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

func main() {
	cfg := config.Load()
	st := store.New(cfg.DataDir)
	audits := audit.NewService(st)
	nsSvc := ns.NewService(st, audits)
	stations := station.NewRegistry(st, audits)
	levels := level.NewService(st)
	levels.SetReverseDelta(cfg.ReverseFlowDelta)
	pumps := pump.NewController(st, stations, audits, levels)
	sluices := sluice.NewRegistry(st, audits)
	rains := rain.NewService(st)
	policies := policy.NewService(st, levels)
	if err := policies.SetRules(policy.Rules{
		PeakRainThreshold: cfg.PeakRainThreshold,
		PeakWindow:        cfg.PeakWindow,
		FullSpeedRate:     cfg.FullSpeedRate,
		ReverseFlowDelta:  cfg.ReverseFlowDelta,
		GateRule:          "downstream-first",
	}); err != nil {
		log.Fatalf("apply policy rules: %v", err)
	}
	quotas := quota.NewService(st)
	dispatcher := dispatch.NewDispatcher(st, nsSvc, stations, pumps, sluices, rains, levels, policies, quotas, audits)
	server := console.NewServer(nsSvc, stations, pumps, sluices, rains, levels, policies, quotas, audits, dispatcher)
	httpServer := &http.Server{Addr: cfg.Addr, Handler: server.Handler()}
	errorsC := make(chan error, 1)
	go func() {
		log.Printf("drainnet listening on %s", cfg.Addr)
		errorsC <- httpServer.ListenAndServe()
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errorsC:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	case sig := <-signals:
		log.Printf("received signal %s, shutting down", sig)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
