package audit

import (
	"time"

	"github.com/google/uuid"

	"drainnet/internal/store"
)

const Kind = "audit"

type Event struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	EntityID string            `json:"entity_id"`
	Message  string            `json:"message"`
	At       time.Time         `json:"at"`
	Meta     map[string]string `json:"meta"`
}

type Service struct {
	store *store.Store
}

func NewService(st *store.Store) *Service {
	return &Service{store: st}
}

func (s *Service) Record(event Event) error {
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.At.IsZero() {
		event.At = time.Now()
	}
	if event.Meta == nil {
		event.Meta = map[string]string{}
	}
	return s.store.WriteJSON(Kind, event.ID, event)
}

func (s *Service) Count() (int, error) {
	ids, err := s.store.List(Kind)
	return len(ids), err
}

func (s *Service) List() ([]Event, error) {
	ids, err := s.store.List(Kind)
	if err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(ids))
	for _, id := range ids {
		var event Event
		if err := s.store.ReadJSON(Kind, id, &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (s *Service) Recent(limit int) ([]Event, error) {
	events, err := s.List()
	if err != nil {
		return nil, err
	}
	if len(events) > limit {
		events = events[len(events)-limit:]
	}
	return events, nil
}
