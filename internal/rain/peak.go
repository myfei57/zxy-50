package rain

import "time"

func (s *Service) PushPeak(gaugeID string, mm float64, at time.Time) error {
	if _, err := s.Gauge(gaugeID); err != nil {
		return err
	}
	s.mu.Lock()
	window := s.peaks[gaugeID]
	if window == nil {
		window = NewPeakWindow()
		s.peaks[gaugeID] = window
	}
	window.Push(mm, at)
	s.mu.Unlock()
	return nil
}

func (s *Service) PeakExceeded(gaugeID string, threshold float64, end time.Time, window time.Duration) (bool, error) {
	if _, err := s.Gauge(gaugeID); err != nil {
		return false, err
	}
	s.mu.RLock()
	peak := s.peaks[gaugeID]
	s.mu.RUnlock()
	if peak == nil {
		return false, nil
	}
	start := end.Truncate(window)
	sum := 0.0
	for _, point := range peak.Points() {
		if !point.At.Before(start) && !point.At.After(end) {
			sum += point.MM
		}
	}
	return sum >= threshold, nil
}
