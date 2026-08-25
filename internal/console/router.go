package console

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

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
)

type Server struct {
	router   chi.Router
	ns       *ns.Service
	stations *station.Registry
	pumps    *pump.Controller
	sluices  *sluice.Registry
	rains    *rain.Service
	levels   *level.Service
	policies *policy.Service
	quotas   *quota.Service
	audits   *audit.Service
	dispatch *dispatch.Dispatcher
}

func NewServer(
	nsSvc *ns.Service,
	stations *station.Registry,
	pumps *pump.Controller,
	sluices *sluice.Registry,
	rains *rain.Service,
	levels *level.Service,
	policies *policy.Service,
	quotas *quota.Service,
	audits *audit.Service,
	dispatcher *dispatch.Dispatcher,
) *Server {
	server := &Server{
		ns:       nsSvc,
		stations: stations,
		pumps:    pumps,
		sluices:  sluices,
		rains:    rains,
		levels:   levels,
		policies: policies,
		quotas:   quotas,
		audits:   audits,
		dispatch: dispatcher,
	}
	server.routes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) routes() {
	router := chi.NewRouter()
	router.Use(requestLog(s.audits))
	router.Use(recoverer)
	router.Get("/api/v1/health", s.health)
	router.Get("/api/v1/status", s.status)
	router.Post("/api/v1/catchments", s.createCatchment)
	router.Get("/api/v1/catchments", s.listCatchments)
	router.Post("/api/v1/catchments/{id}/zones", s.addZone)
	router.Post("/api/v1/catchments/{id}/split", s.splitCatchment)
	router.Post("/api/v1/catchments/bind", s.bindStation)
	router.Post("/api/v1/zones/{id}/assign", s.assignZone)
	router.Post("/api/v1/stations", s.createStation)
	router.Get("/api/v1/stations", s.listStations)
	router.Get("/api/v1/stations/{id}", s.getStation)
	router.Delete("/api/v1/stations/{id}", s.deleteStation)
	router.Get("/api/v1/stations/{id}/speed", s.pumpSpeed)
	router.Get("/api/v1/stations/{id}/telemetry", s.pumpTelemetry)
	router.Get("/api/v1/stations/{id}/pool", s.stationPool)
	router.Get("/api/v1/stations/{id}/mapping-version", s.mappingVersion)
	router.Post("/api/v1/stations/{id}/pumps", s.addPump)
	router.Post("/api/v1/stations/{id}/lead", s.setLead)
	router.Post("/api/v1/stations/{id}/mapping", s.refreshMapping)
	router.Post("/api/v1/stations/{id}/trip", s.tripPump)
	router.Post("/api/v1/pumps/{id}/start", s.startPump)
	router.Post("/api/v1/pumps/{id}/stop", s.stopPump)
	router.Post("/api/v1/pumps/{id}/failover", s.failoverPump)
	router.Post("/api/v1/sluices", s.createGate)
	router.Get("/api/v1/channels/{id}/gates", s.channelGates)
	router.Get("/api/v1/sluices/{id}/history", s.gateHistory)
	router.Post("/api/v1/sluices/{id}/manual", s.manualGate)
	router.Post("/api/v1/sluices/{id}/close", s.closeGate)
	router.Post("/api/v1/rain/gauges", s.createGauge)
	router.Get("/api/v1/rain/gauges/{id}", s.gaugeStatus)
	router.Post("/api/v1/rain/gauges/{id}/accumulate", s.accumulateRain)
	router.Post("/api/v1/rain/gauges/{id}/calibrate", s.calibrateGauge)
	router.Post("/api/v1/rain/gauges/{id}/peak", s.pushPeak)
	router.Post("/api/v1/level/sample", s.levelSample)
	router.Get("/api/v1/level/{id}/last", s.levelLast)
	router.Get("/api/v1/level/{id}/rate", s.levelRate)
	router.Get("/api/v1/level/{id}/direction", s.levelDirection)
	router.Get("/api/v1/level/{id}/history", s.levelHistory)
	router.Post("/api/v1/dispatch/prelower", s.dispatchPreLower)
	router.Post("/api/v1/dispatch/rain", s.dispatchRain)
	router.Post("/api/v1/dispatch/open", s.dispatchOpen)
	router.Post("/api/v1/dispatch/recover", s.dispatchRecover)
	router.Post("/api/v1/dispatch/manual", s.dispatchManual)
	router.Post("/api/v1/dispatch/plan", s.dispatchPlan)
	router.Post("/api/v1/dispatch/queue", s.dispatchQueue)
	router.Get("/api/v1/dispatch/pending", s.dispatchPending)
	router.Post("/api/v1/dispatch/level", s.dispatchLevelSample)
	router.Get("/api/v1/policy/rules", s.getRules)
	router.Post("/api/v1/policy/rules", s.setRules)
	router.Get("/api/v1/policy/rules/load", s.loadRules)
	router.Get("/api/v1/policy/gate-rule", s.gateRule)
	router.Post("/api/v1/policy/reverse", s.checkReverse)
	router.Post("/api/v1/quota/limit", s.setQuotaLimit)
	router.Get("/api/v1/quota/{id}/remaining", s.quotaRemaining)
	router.Post("/api/v1/quota/reset", s.quotaReset)
	router.Post("/api/v1/quota/release", s.quotaRelease)
	router.Get("/api/v1/audit/events", s.auditEvents)
	router.Get("/api/v1/audit/types", s.auditTypes)
	router.Get("/api/v1/audit/recent", s.auditRecent)
	router.Get("/api/v1/audit/between", s.auditBetween)
	s.router = router
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func readJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	status, err := s.dispatch.Status()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	count, err := s.audits.Count()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	status["audit_events"] = count
	writeJSON(w, http.StatusOK, status)
}
