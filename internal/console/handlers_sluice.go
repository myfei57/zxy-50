package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"drainnet/internal/sluice"
)

func (s *Server) createGate(w http.ResponseWriter, r *http.Request) {
	var gate sluice.Gate
	if err := readJSON(r, &gate); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	registered, err := s.sluices.Register(gate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (s *Server) channelGates(w http.ResponseWriter, r *http.Request) {
	gates, err := s.sluices.Gates(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, gates)
}

func (s *Server) manualGate(w http.ResponseWriter, r *http.Request) {
	gateID := chi.URLParam(r, "id")
	var body struct {
		Target int `json:"target"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.sluices.Manual(gateID, body.Target); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"gate": gateID, "target": formatTarget(body.Target)})
}

func (s *Server) closeGate(w http.ResponseWriter, r *http.Request) {
	gateID := chi.URLParam(r, "id")
	if err := s.sluices.CloseGate(gateID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"gate": gateID, "position": "closed"})
}

func (s *Server) gateHistory(w http.ResponseWriter, r *http.Request) {
	history, err := s.sluices.History(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func formatTarget(value int) string {
	return formatCount(value)
}
