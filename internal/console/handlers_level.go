package console

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) levelSample(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID string  `json:"station_id"`
		RawMM     float64 `json:"raw_mm"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sample, err := s.levels.Sample(body.StationID, body.RawMM, time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.pumps.HandleLevel(sample); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sample)
}

func (s *Server) levelRate(w http.ResponseWriter, r *http.Request) {
	rate, err := s.levels.Rate(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"rate": rate})
}

func (s *Server) levelDirection(w http.ResponseWriter, r *http.Request) {
	direction, err := s.levels.Direction(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"direction": direction})
}

func (s *Server) levelHistory(w http.ResponseWriter, r *http.Request) {
	history, err := s.levels.History(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (s *Server) levelLast(w http.ResponseWriter, r *http.Request) {
	sample, err := s.levels.LastSample(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sample)
}
