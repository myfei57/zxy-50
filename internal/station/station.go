package station

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"drainnet/internal/audit"
	"drainnet/internal/store"
)

const Kind = "stations"

var ErrStationNotFound = errors.New("station not found")

type PumpRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Duty      string `json:"duty"`
	State     string `json:"state"`
	BreakerID string `json:"breaker_id"`
	Speed     string `json:"speed"`
}

type Breaker struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	PumpID string `json:"pump_id"`
}

type Station struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CatchmentID  string    `json:"catchment_id"`
	MappingEpoch int       `json:"mapping_epoch"`
	LeadPumpID   string    `json:"lead_pump_id"`
	Pumps        []PumpRef `json:"pumps"`
	Breakers     []Breaker `json:"breakers"`
}

type Registry struct {
	store   *store.Store
	audit   *audit.Service
	mu      sync.RWMutex
	lead    map[string]*LeadState
	mapping *Mapping
	pools   map[string]*BreakerPool
}

func NewRegistry(st *store.Store, a *audit.Service) *Registry {
	return &Registry{
		store:   st,
		audit:   a,
		lead:    map[string]*LeadState{},
		mapping: NewMapping(),
		pools:   map[string]*BreakerPool{},
	}
}

func (r *Registry) Register(value Station) (Station, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if value.ID == "" {
		value.ID = uuid.NewString()
	}
	if value.Pumps == nil {
		value.Pumps = []PumpRef{}
	}
	if value.Breakers == nil {
		value.Breakers = []Breaker{}
	}
	if err := r.store.WriteJSON(Kind, value.ID, value); err != nil {
		return Station{}, err
	}
	ids := make([]string, 0, len(value.Breakers))
	for _, breaker := range value.Breakers {
		ids = append(ids, breaker.ID)
	}
	r.pools[value.ID] = NewBreakerPool(ids)
	r.lead[value.ID] = NewLeadState()
	if value.LeadPumpID != "" {
		r.lead[value.ID].SetLead(value.LeadPumpID)
	}
	if err := r.audit.Record(audit.Event{Type: "station.register", EntityID: value.ID, Message: "station registered"}); err != nil {
		return Station{}, err
	}
	return value, nil
}

func (r *Registry) Get(id string) (Station, error) {
	var value Station
	err := r.store.ReadJSON(Kind, id, &value)
	if err != nil {
		return Station{}, ErrStationNotFound
	}
	return value, nil
}

func (r *Registry) List() ([]Station, error) {
	ids, err := r.store.List(Kind)
	if err != nil {
		return nil, err
	}
	values := make([]Station, 0, len(ids))
	for _, id := range ids {
		value, err := r.Get(id)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (r *Registry) Update(value Station) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.store.WriteJSON(Kind, value.ID, value)
}

func (r *Registry) AddPump(stationID string, ref PumpRef) (PumpRef, error) {
	value, err := r.Get(stationID)
	if err != nil {
		return PumpRef{}, err
	}
	if ref.ID == "" {
		ref.ID = uuid.NewString()
	}
	for index := range value.Pumps {
		if value.Pumps[index].ID == ref.ID {
			value.Pumps[index] = ref
			if err := r.Update(value); err != nil {
				return PumpRef{}, err
			}
			return ref, nil
		}
	}
	value.Pumps = append(value.Pumps, ref)
	if value.LeadPumpID == "" && ref.Duty == "lead" {
		value.LeadPumpID = ref.ID
		r.lead[stationID].SetLead(ref.ID)
	}
	if err := r.Update(value); err != nil {
		return PumpRef{}, err
	}
	return ref, nil
}
