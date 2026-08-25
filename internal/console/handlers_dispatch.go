package console

import (
	"net/http"
	"time"

	"drainnet/internal/dispatch"
)

func (s *Server) dispatchPreLower(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID string `json:"station_id"`
		GaugeID   string `json:"gauge_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	triggered, err := s.dispatch.PreLower(body.StationID, body.GaugeID, time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"triggered": triggered})
}

func (s *Server) dispatchRain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ZoneID string  `json:"zone_id"`
		MM     float64 `json:"mm"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.dispatch.DispatchRain(body.ZoneID, body.MM, time.Now()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"zone": body.ZoneID, "status": "dispatched"})
}

func (s *Server) dispatchOpen(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChannelID string `json:"channel_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.dispatch.OpenForPreDrain(body.ChannelID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"channel": body.ChannelID, "status": "opened"})
}

func (s *Server) dispatchRecover(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID string `json:"station_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.dispatch.Recover(body.StationID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": body.StationID, "status": "recovered"})
}

func (s *Server) dispatchManual(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID string `json:"station_id"`
		GateID    string `json:"gate_id"`
		Target    int    `json:"target"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.dispatch.ManualGate(body.StationID, body.GateID, body.Target); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"gate": body.GateID, "status": "manual applied"})
}

func (s *Server) dispatchPlan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChannelID string `json:"channel_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	plan, err := s.dispatch.Plan(body.ChannelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (s *Server) dispatchQueue(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Command dispatch.Command `json:"command"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	command := body.Command
	if err := s.dispatch.EnqueueCommand(command); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, command)
}

func (s *Server) dispatchPending(w http.ResponseWriter, r *http.Request) {
	commands, err := s.dispatch.PendingCommands()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, commands)
}

func (s *Server) dispatchLevelSample(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID string  `json:"station_id"`
		RawMM     float64 `json:"raw_mm"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.dispatch.HandleLevelSample(body.StationID, body.RawMM, time.Now()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": body.StationID, "status": "level applied"})
}
