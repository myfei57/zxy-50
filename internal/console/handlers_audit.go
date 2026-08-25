package console

import (
	"net/http"
)

func (s *Server) auditEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.audits.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) auditTypes(w http.ResponseWriter, r *http.Request) {
	types, err := s.audits.Types()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, types)
}

func (s *Server) auditRecent(w http.ResponseWriter, r *http.Request) {
	events, err := s.audits.Recent(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) auditBetween(w http.ResponseWriter, r *http.Request) {
	events, err := s.audits.Between(beginOfDay(), endOfDay())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}
