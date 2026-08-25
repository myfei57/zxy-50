package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"drainnet/internal/ns"
	"drainnet/internal/station"
)

func (s *Server) createStation(w http.ResponseWriter, r *http.Request) {
	var value station.Station
	if err := readJSON(r, &value); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	registered, err := s.stations.Register(value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (s *Server) listStations(w http.ResponseWriter, r *http.Request) {
	values, err := s.stations.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (s *Server) getStation(w http.ResponseWriter, r *http.Request) {
	value, err := s.stations.Get(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (s *Server) addPump(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	var ref station.PumpRef
	if err := readJSON(r, &ref); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	registered, err := s.pumps.Register(stationID, ref)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (s *Server) setLead(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	var body struct {
		PumpID string `json:"pump_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.stations.SetLead(stationID, body.PumpID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": stationID, "lead": body.PumpID})
}

func (s *Server) refreshMapping(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	var body struct {
		CatchmentID string `json:"catchment_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	zones, err := s.ns.Zones(body.CatchmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.stations.RefreshMapping(stationID, zones); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": stationID, "status": "mapping refreshed"})
}

func (s *Server) tripPump(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	var body struct {
		PumpID string `json:"pump_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	snapshot, err := s.stations.CaptureSnapshot(stationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.stations.ApplySnapshot(stationID, snapshot); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ref, err := s.pumps.StateOf(stationID, body.PumpID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ref.BreakerID != "" {
		if err := s.stations.ReleaseBreaker(stationID, ref.BreakerID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	ref.State = "tripped"
	ref.BreakerID = ""
	if _, err := s.stations.SetPumpState(stationID, ref); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.pumps.Failover(stationID, body.PumpID, snapshot); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": stationID, "pump": body.PumpID, "status": "tripped"})
}

func (s *Server) createCatchment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	catchment, err := s.ns.CreateCatchment(body.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, catchment)
}

func (s *Server) addZone(w http.ResponseWriter, r *http.Request) {
	catchmentID := chi.URLParam(r, "id")
	var body struct {
		Zone ns.Zone `json:"zone"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	zone, err := s.ns.AddZone(catchmentID, body.Zone)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, zone)
}

func (s *Server) splitCatchment(w http.ResponseWriter, r *http.Request) {
	catchmentID := chi.URLParam(r, "id")
	var body struct {
		Zone ns.Zone `json:"zone"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	catchment, err := s.ns.SplitCatchment(catchmentID, body.Zone)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catchment)
}

func (s *Server) bindStation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CatchmentID string `json:"catchment_id"`
		StationID   string `json:"station_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.ns.BindStation(body.CatchmentID, body.StationID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "bound"})
}

func (s *Server) assignZone(w http.ResponseWriter, r *http.Request) {
	zoneID := chi.URLParam(r, "id")
	var body struct {
		StationID string `json:"station_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.ns.AssignStation(zoneID, body.StationID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"zone": zoneID, "station": body.StationID})
}

func (s *Server) deleteStation(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	if err := s.stations.Delete(stationID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": stationID, "status": "deleted"})
}

func (s *Server) stationPool(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "id")
	free, held, err := s.stations.PoolStats(stationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	value, err := s.stations.Get(stationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	holders := map[string]string{}
	for _, breaker := range value.Breakers {
		holders[breaker.ID] = s.stations.BreakerHolder(stationID, breaker.ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"free":    free,
		"held":    held,
		"holders": holders,
	})
}

func (s *Server) listCatchments(w http.ResponseWriter, r *http.Request) {
	catchments, err := s.ns.ListCatchments()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catchments)
}

func (s *Server) mappingVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]int{"version": s.stations.MappingVersion()})
}
