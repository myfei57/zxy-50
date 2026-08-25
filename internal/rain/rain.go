package rain

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"drainnet/internal/store"
)

const GaugeKind = "rain_gauges"

const ReadingKind = "rain_readings"

var ErrGaugeNotFound = errors.New("rain gauge not found")

type Reading struct {
	ID       string    `json:"id"`
	GaugeID  string    `json:"gauge_id"`
	RawMM    float64   `json:"raw_mm"`
	OffsetMM float64   `json:"offset_mm"`
	TotalMM  float64   `json:"total_mm"`
	At       time.Time `json:"at"`
}

type Gauge struct {
	ID           string  `json:"id"`
	StationID    string  `json:"station_id"`
	Name         string  `json:"name"`
	Offset       float64 `json:"offset"`
	StoredOffset float64 `json:"stored_offset"`
	Accumulated  float64 `json:"accumulated"`
}

type point struct {
	MM float64   `json:"mm"`
	At time.Time `json:"at"`
}

type PeakWindow struct {
	points []point
}

func NewPeakWindow() *PeakWindow {
	return &PeakWindow{}
}

func (w *PeakWindow) Push(mm float64, at time.Time) {
	w.points = append(w.points, point{MM: mm, At: at})
}

func (w *PeakWindow) Points() []point {
	return w.points
}

type Service struct {
	store  *store.Store
	mu     sync.RWMutex
	gauges map[string]Gauge
	peaks  map[string]*PeakWindow
}

func NewService(st *store.Store) *Service {
	return &Service{
		store:  st,
		gauges: map[string]Gauge{},
		peaks:  map[string]*PeakWindow{},
	}
}

func (s *Service) RegisterGauge(gauge Gauge) (Gauge, error) {
	if gauge.ID == "" {
		gauge.ID = uuid.NewString()
	}
	gauge.StoredOffset = gauge.Offset
	s.mu.Lock()
	s.gauges[gauge.ID] = gauge
	s.peaks[gauge.ID] = NewPeakWindow()
	s.mu.Unlock()
	if err := s.store.WriteJSON(GaugeKind, gauge.ID, gauge); err != nil {
		return Gauge{}, err
	}
	return gauge, nil
}

func (s *Service) Gauge(gaugeID string) (Gauge, error) {
	s.mu.RLock()
	gauge, ok := s.gauges[gaugeID]
	s.mu.RUnlock()
	if !ok {
		var stored Gauge
		err := s.store.ReadJSON(GaugeKind, gaugeID, &stored)
		if err != nil {
			return Gauge{}, ErrGaugeNotFound
		}
		s.mu.Lock()
		s.gauges[gaugeID] = stored
		s.mu.Unlock()
		return stored, nil
	}
	return gauge, nil
}

func (s *Service) Accumulate(gaugeID string, rawMM float64, at time.Time) (Reading, error) {
	gauge, err := s.Gauge(gaugeID)
	if err != nil {
		return Reading{}, err
	}
	offset := gauge.StoredOffset
	total := rawMM + offset
	gauge.Accumulated += total
	s.mu.Lock()
	s.gauges[gaugeID] = gauge
	s.peaks[gaugeID].Push(total, at)
	s.mu.Unlock()
	if err := s.store.WriteJSON(GaugeKind, gaugeID, gauge); err != nil {
		return Reading{}, err
	}
	reading := Reading{
		ID:       uuid.NewString(),
		GaugeID:  gaugeID,
		RawMM:    rawMM,
		OffsetMM: offset,
		TotalMM:  total,
		At:       at,
	}
	if err := s.store.WriteJSON(ReadingKind, reading.ID, reading); err != nil {
		return Reading{}, err
	}
	return reading, nil
}

func (s *Service) Accumulated(gaugeID string) (float64, error) {
	gauge, err := s.Gauge(gaugeID)
	if err != nil {
		return 0, err
	}
	return gauge.Accumulated, nil
}

func (s *Service) Readings(gaugeID string) ([]Reading, error) {
	ids, err := s.store.List(ReadingKind)
	if err != nil {
		return nil, err
	}
	readings := make([]Reading, 0, len(ids))
	for _, id := range ids {
		var reading Reading
		if err := s.store.ReadJSON(ReadingKind, id, &reading); err != nil {
			return nil, err
		}
		if reading.GaugeID == gaugeID {
			readings = append(readings, reading)
		}
	}
	return readings, nil
}

func (s *Service) GaugesForStation(stationID string) ([]Gauge, error) {
	ids, err := s.store.List(GaugeKind)
	if err != nil {
		return nil, err
	}
	gauges := make([]Gauge, 0, len(ids))
	for _, id := range ids {
		gauge, err := s.Gauge(id)
		if err != nil {
			return nil, err
		}
		if gauge.StationID == stationID {
			gauges = append(gauges, gauge)
		}
	}
	return gauges, nil
}
