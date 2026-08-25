package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) setQuotaLimit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID  string  `json:"station_id"`
		DailyLimit float64 `json:"daily_limit"`
		Day        string  `json:"day"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.quotas.SetLimit(body.StationID, body.DailyLimit, body.Day); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": body.StationID, "status": "limit set"})
}

func (s *Server) quotaRemaining(w http.ResponseWriter, r *http.Request) {
	remaining, err := s.quotas.Remaining(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"remaining": remaining})
}

func (s *Server) quotaReset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Day string `json:"day"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.quotas.ResetDay(body.Day); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"day": body.Day, "status": "reset"})
}

func (s *Server) quotaRelease(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID string  `json:"station_id"`
		Amount    float64 `json:"amount"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.quotas.Release(body.StationID, body.Amount); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": body.StationID, "status": "released"})
}
