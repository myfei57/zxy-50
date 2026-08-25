package level

import "time"

type RateWindow struct {
	values []float64
	times  []time.Time
}

func NewRateWindow() *RateWindow {
	return &RateWindow{}
}

func (w *RateWindow) PushSample(smoothed float64, raw float64, at time.Time) {
	w.values = append(w.values, smoothed)
	w.times = append(w.times, at)
	if len(w.values) > 32 {
		w.values = w.values[len(w.values)-32:]
		w.times = w.times[len(w.times)-32:]
	}
}

func (w *RateWindow) Rate() float64 {
	if len(w.values) < 2 {
		return 0
	}
	first := w.values[0]
	last := w.values[len(w.values)-1]
	span := w.times[len(w.times)-1].Sub(w.times[0])
	if span <= 0 {
		return 0
	}
	return (last - first) / span.Minutes()
}

func (s *Service) Rate(stationID string) (float64, error) {
	s.mu.RLock()
	window := s.windows[stationID]
	s.mu.RUnlock()
	if window == nil {
		return 0, nil
	}
	return window.Rate(), nil
}
