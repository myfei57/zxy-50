package console

import (
	"net/http"

	"drainnet/internal/policy"
)

func (s *Server) getRules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.policies.Rules())
}

func (s *Server) setRules(w http.ResponseWriter, r *http.Request) {
	var rules policy.Rules
	if err := readJSON(r, &rules); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.policies.SetRules(rules); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.policies.Rules())
}

func (s *Server) checkReverse(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StationID string `json:"station_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	reverse, err := s.policies.CheckReverse(body.StationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"reverse": reverse})
}

func (s *Server) loadRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.policies.LoadRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (s *Server) gateRule(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"rule": s.policies.GateRule()})
}
