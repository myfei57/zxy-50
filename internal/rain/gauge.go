package rain

func (s *Service) Calibrate(gaugeID string, offset float64) error {
	gauge, err := s.Gauge(gaugeID)
	if err != nil {
		return err
	}
	gauge.Offset = offset
	s.mu.Lock()
	s.gauges[gaugeID] = gauge
	s.mu.Unlock()
	return s.store.WriteJSON(GaugeKind, gaugeID, gauge)
}

func (s *Service) Offset(gaugeID string) (float64, error) {
	gauge, err := s.Gauge(gaugeID)
	if err != nil {
		return 0, err
	}
	return gauge.Offset, nil
}

func (s *Service) StoredOffset(gaugeID string) (float64, error) {
	gauge, err := s.Gauge(gaugeID)
	if err != nil {
		return 0, err
	}
	return gauge.StoredOffset, nil
}
