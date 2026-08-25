package console

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"drainnet/internal/rain"
)

func (s *Server) createGauge(w http.ResponseWriter, r *http.Request) {
	var gauge rain.Gauge
	if err := readJSON(r, &gauge); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	registered, err := s.rains.RegisterGauge(gauge)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (s *Server) accumulateRain(w http.ResponseWriter, r *http.Request) {
	gaugeID := chi.URLParam(r, "id")
	var body struct {
		RawMM float64 `json:"raw_mm"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	reading, err := s.rains.Accumulate(gaugeID, body.RawMM, time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, reading)
}

func (s *Server) calibrateGauge(w http.ResponseWriter, r *http.Request) {
	gaugeID := chi.URLParam(r, "id")
	var body struct {
		Offset float64 `json:"offset"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.rains.Calibrate(gaugeID, body.Offset); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	offset, err := s.rains.Offset(gaugeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"offset": offset})
}

func (s *Server) gaugeStatus(w http.ResponseWriter, r *http.Request) {
	gaugeID := chi.URLParam(r, "id")
	gauge, err := s.rains.Gauge(gaugeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	accumulated, err := s.rains.Accumulated(gaugeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	readings, err := s.rains.Readings(gaugeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	storedOffset, err := s.rains.StoredOffset(gaugeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"gauge":         gauge,
		"accumulated":   accumulated,
		"stored_offset": storedOffset,
		"readings":      len(readings),
	})
}

func (s *Server) pushPeak(w http.ResponseWriter, r *http.Request) {
	gaugeID := chi.URLParam(r, "id")
	var body struct {
		MM float64 `json:"mm"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.rains.PushPeak(gaugeID, body.MM, time.Now()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"gauge": gaugeID, "status": "peak pushed"})
}
