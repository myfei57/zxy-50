package quota

import (
	"errors"
	"sync"

	"drainnet/internal/store"
)

const Kind = "quotas"

var ErrQuotaExceeded = errors.New("daily discharge quota exceeded")

const defaultDailyLimit = 1000000

type Quota struct {
	StationID  string  `json:"station_id"`
	DailyLimit float64 `json:"daily_limit"`
	Used       float64 `json:"used"`
	Day        string  `json:"day"`
}

type Service struct {
	store *store.Store
	mu    sync.RWMutex
	cache map[string]Quota
}

func NewService(st *store.Store) *Service {
	return &Service{store: st, cache: map[string]Quota{}}
}

func (s *Service) Get(stationID string) (Quota, error) {
	s.mu.RLock()
	quota, ok := s.cache[stationID]
	s.mu.RUnlock()
	if ok {
		return quota, nil
	}
	var stored Quota
	err := s.store.ReadJSON(Kind, stationID, &stored)
	if err != nil {
		stored = Quota{StationID: stationID, DailyLimit: defaultDailyLimit}
	}
	s.mu.Lock()
	s.cache[stationID] = stored
	s.mu.Unlock()
	return stored, nil
}
