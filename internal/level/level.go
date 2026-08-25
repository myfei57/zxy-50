package level

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"drainnet/internal/store"
)

const Kind = "level_samples"

type Sample struct {
	ID         string    `json:"id"`
	StationID  string    `json:"station_id"`
	RawMM      float64   `json:"raw_mm"`
	SmoothedMM float64   `json:"smoothed_mm"`
	Rate       float64   `json:"rate"`
	Direction  string    `json:"direction"`
	At         time.Time `json:"at"`
}

type point struct {
	MM float64   `json:"mm"`
	At time.Time `json:"at"`
}

type History struct {
	ID        string    `json:"id"`
	StationID string    `json:"station_id"`
	Raw       []point   `json:"raw"`
	Smoothed  []float64 `json:"smoothed"`
}

type Service struct {
	store        *store.Store
	mu           sync.RWMutex
	history      map[string]*History
	windows      map[string]*RateWindow
	dirs         map[string]*Direction
	reverseDelta float64
}

func NewService(st *store.Store) *Service {
	return &Service{
		store:        st,
		history:      map[string]*History{},
		windows:      map[string]*RateWindow{},
		dirs:         map[string]*Direction{},
		reverseDelta: 8,
	}
}

func (s *Service) SetReverseDelta(delta float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reverseDelta = delta
}

func (s *Service) ReverseDelta() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reverseDelta
}

func (s *Service) Sample(stationID string, rawMM float64, at time.Time) (Sample, error) {
	threshold := s.ReverseDelta()
	s.mu.Lock()
	history := s.history[stationID]
	if history == nil {
		history = &History{ID: uuid.NewString(), StationID: stationID}
		s.history[stationID] = history
	}
	history.Raw = append(history.Raw, point{MM: rawMM, At: at})
	smoothed := smooth(history.Raw)
	history.Smoothed = append(history.Smoothed, smoothed)
	direction := s.computeDirection(history.Raw, threshold)
	s.directionFor(stationID).Update(direction)
	window := s.windowFor(stationID)
	window.PushSample(smoothed, rawMM, at)
	rate := window.Rate()
	s.mu.Unlock()
	sample := Sample{
		ID:         uuid.NewString(),
		StationID:  stationID,
		RawMM:      rawMM,
		SmoothedMM: smoothed,
		Rate:       rate,
		Direction:  direction,
		At:         at,
	}
	if err := s.store.WriteJSON(Kind, sample.ID, sample); err != nil {
		return Sample{}, err
	}
	if err := s.store.WriteJSON("level_history", history.ID, history); err != nil {
		return Sample{}, err
	}
	return sample, nil
}

func (s *Service) computeDirection(raw []point, threshold float64) string {
	if len(raw) < 2 {
		return "stable"
	}
	delta := raw[len(raw)-1].MM - raw[len(raw)-2].MM
	if delta >= threshold {
		return "in"
	}
	if delta <= -threshold {
		return "out"
	}
	return "stable"
}

func (s *Service) windowFor(stationID string) *RateWindow {
	window := s.windows[stationID]
	if window == nil {
		window = NewRateWindow()
		s.windows[stationID] = window
	}
	return window
}

func (s *Service) directionFor(stationID string) *Direction {
	direction := s.dirs[stationID]
	if direction == nil {
		direction = NewDirection()
		s.dirs[stationID] = direction
	}
	return direction
}

func smooth(raw []point) float64 {
	start := 0
	if len(raw) > 5 {
		start = len(raw) - 5
	}
	values := make([]float64, 0, len(raw)-start)
	for _, value := range raw[start:] {
		values = append(values, value.MM)
	}
	if len(values) == 0 {
		return 0
	}
	sort.Float64s(values)
	return values[(len(values)-1)/2]
}
