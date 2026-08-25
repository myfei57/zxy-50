package policy

import (
	"sync"
	"time"

	"drainnet/internal/level"
	"drainnet/internal/store"
)

type Rules struct {
	PeakRainThreshold float64       `json:"peak_rain_threshold"`
	PeakWindow        time.Duration `json:"peak_window"`
	FullSpeedRate     float64       `json:"full_speed_rate"`
	ReverseFlowDelta  float64       `json:"reverse_flow_delta"`
	GateRule          string        `json:"gate_rule"`
}

type Service struct {
	store  *store.Store
	levels *level.Service
	mu     sync.RWMutex
	rules  Rules
}

func NewService(st *store.Store, levels *level.Service) *Service {
	return &Service{
		store:  st,
		levels: levels,
		rules: Rules{
			PeakRainThreshold: 50,
			PeakWindow:        10 * time.Minute,
			FullSpeedRate:     5,
			ReverseFlowDelta:  8,
			GateRule:          "downstream-first",
		},
	}
}

func (s *Service) Rules() Rules {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rules
}

func (s *Service) SetRules(rules Rules) error {
	s.mu.Lock()
	s.rules = rules
	s.mu.Unlock()
	s.levels.SetReverseDelta(rules.ReverseFlowDelta)
	return s.store.WriteJSON("policy_rules", "current", rules)
}
