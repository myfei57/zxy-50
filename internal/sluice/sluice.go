package sluice

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"drainnet/internal/audit"
	"drainnet/internal/store"
)

const Kind = "gates"

var ErrGateNotFound = errors.New("gate not found")

var ErrGateOrder = errors.New("downstream gate must open before upstream gate")

type Gate struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
	Side      string `json:"side"`
	Position  int    `json:"position"`
}

type Registry struct {
	store *store.Store
	audit *audit.Service
	mu    sync.RWMutex
	pos   map[string]*Position
	cache map[string]Gate
}

func NewRegistry(st *store.Store, a *audit.Service) *Registry {
	return &Registry{
		store: st,
		audit: a,
		pos:   map[string]*Position{},
		cache: map[string]Gate{},
	}
}

func (r *Registry) Register(gate Gate) (Gate, error) {
	if gate.ID == "" {
		gate.ID = uuid.NewString()
	}
	r.mu.Lock()
	r.cache[gate.ID] = gate
	r.pos[gate.ID] = NewPosition()
	r.mu.Unlock()
	if err := r.store.WriteJSON(Kind, gate.ID, gate); err != nil {
		return Gate{}, err
	}
	if err := r.audit.Record(audit.Event{Type: "sluice.register", EntityID: gate.ID, Message: "gate registered"}); err != nil {
		return Gate{}, err
	}
	return gate, nil
}

func (r *Registry) Gate(id string) (Gate, error) {
	r.mu.RLock()
	cached, ok := r.cache[id]
	r.mu.RUnlock()
	if ok {
		return cached, nil
	}
	var gate Gate
	err := r.store.ReadJSON(Kind, id, &gate)
	if err != nil {
		return Gate{}, ErrGateNotFound
	}
	r.mu.Lock()
	r.cache[id] = gate
	r.mu.Unlock()
	return gate, nil
}

func (r *Registry) Gates(channelID string) ([]Gate, error) {
	ids, err := r.store.List(Kind)
	if err != nil {
		return nil, err
	}
	gates := make([]Gate, 0, len(ids))
	for _, id := range ids {
		gate, err := r.Gate(id)
		if err != nil {
			return nil, err
		}
		if gate.ChannelID == channelID {
			gates = append(gates, gate)
		}
	}
	return gates, nil
}

func (r *Registry) persistGate(gate Gate) error {
	r.mu.Lock()
	r.cache[gate.ID] = gate
	r.mu.Unlock()
	return r.store.WriteJSON(Kind, gate.ID, gate)
}
