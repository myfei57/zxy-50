package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"drainnet/internal/pump"
	"drainnet/internal/station"
)

func (s *Server) startPump(w http.ResponseWriter, r *http.Request) {
	pumpID := chi.URLParam(r, "id")
	var body struct {
		StationID string `json:"station_id"`
		Reason    string `json:"reason"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.pumps.StartPump(body.StationID, pumpID, body.Reason); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pump": pumpID, "status": "running"})
}

func (s *Server) stopPump(w http.ResponseWriter, r *http.Request) {
	pumpID := chi.URLParam(r, "id")
	var body struct {
		StationID string `json:"station_id"`
		Reason    string `json:"reason"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.pumps.StopPump(body.StationID, pumpID, body.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pump": pumpID, "status": "stopped"})
}

func (s *Server) failoverPump(w http.ResponseWriter, r *http.Request) {
	pumpID := chi.URLParam(r, "id")
	var body struct {
		StationID string               `json:"station_id"`
		Snapshot  station.LeadSnapshot `json:"snapshot"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.pumps.Failover(body.StationID, pumpID, body.Snapshot); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pump": pumpID, "status": "failover completed"})
}

func (s *Server) pumpSpeed(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	speed, err := s.pumps.Speed(stationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": stationID, "speed": pump.SpeedLabel(speed)})
}

func (s *Server) pumpTelemetry(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	if err := s.pumps.RecordSnapshot(stationID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	row, err := s.pumps.Telemetry(stationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, row)
}
