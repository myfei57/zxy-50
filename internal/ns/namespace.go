package ns

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"drainnet/internal/audit"
	"drainnet/internal/store"
)

const CatchmentKind = "catchments"

const ZoneKind = "zones"

var ErrCatchmentNotFound = errors.New("catchment not found")

var ErrZoneNotFound = errors.New("zone not found")

type Zone struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CatchmentID string `json:"catchment_id"`
	StationID   string `json:"station_id"`
}

type Catchment struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	SplitEpoch int      `json:"split_epoch"`
	ZoneIDs    []string `json:"zone_ids"`
	StationIDs []string `json:"station_ids"`
}

type Service struct {
	store *store.Store
	audit *audit.Service
	mu    sync.RWMutex
}

func NewService(st *store.Store, a *audit.Service) *Service {
	return &Service{store: st, audit: a}
}

func (s *Service) CreateCatchment(name string) (Catchment, error) {
	catchment := Catchment{
		ID:      uuid.NewString(),
		Name:    name,
		ZoneIDs: []string{},
	}
	if err := s.store.WriteJSON(CatchmentKind, catchment.ID, catchment); err != nil {
		return Catchment{}, err
	}
	if err := s.audit.Record(audit.Event{Type: "ns.catchment", EntityID: catchment.ID, Message: "catchment created"}); err != nil {
		return Catchment{}, err
	}
	return catchment, nil
}

func (s *Service) GetCatchment(id string) (Catchment, error) {
	var catchment Catchment
	err := s.store.ReadJSON(CatchmentKind, id, &catchment)
	if err != nil {
		return Catchment{}, ErrCatchmentNotFound
	}
	return catchment, nil
}

func (s *Service) ListCatchments() ([]Catchment, error) {
	ids, err := s.store.List(CatchmentKind)
	if err != nil {
		return nil, err
	}
	catchments := make([]Catchment, 0, len(ids))
	for _, id := range ids {
		catchment, err := s.GetCatchment(id)
		if err != nil {
			return nil, err
		}
		catchments = append(catchments, catchment)
	}
	return catchments, nil
}

func (s *Service) BindStation(catchmentID string, stationID string) error {
	catchment, err := s.GetCatchment(catchmentID)
	if err != nil {
		return err
	}
	for _, existing := range catchment.StationIDs {
		if existing == stationID {
			return nil
		}
	}
	catchment.StationIDs = append(catchment.StationIDs, stationID)
	if err := s.store.WriteJSON(CatchmentKind, catchment.ID, catchment); err != nil {
		return err
	}
	return s.audit.Record(audit.Event{Type: "ns.bind", EntityID: stationID, Message: "station bound to catchment"})
}
