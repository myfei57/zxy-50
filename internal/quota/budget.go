package quota

func (s *Service) SetLimit(stationID string, limit float64, day string) error {
	quota, err := s.Get(stationID)
	if err != nil {
		quota = Quota{StationID: stationID}
	}
	quota.DailyLimit = limit
	quota.Day = day
	s.mu.Lock()
	s.cache[stationID] = quota
	s.mu.Unlock()
	return s.store.WriteJSON(Kind, stationID, quota)
}

func (s *Service) Reserve(stationID string, amount float64) error {
	quota, err := s.Get(stationID)
	if err != nil {
		return err
	}
	if quota.Used+amount > quota.DailyLimit {
		return ErrQuotaExceeded
	}
	quota.Used += amount
	s.mu.Lock()
	s.cache[stationID] = quota
	s.mu.Unlock()
	return s.store.WriteJSON(Kind, stationID, quota)
}

func (s *Service) Release(stationID string, amount float64) error {
	quota, err := s.Get(stationID)
	if err != nil {
		return err
	}
	quota.Used -= amount
	if quota.Used < 0 {
		quota.Used = 0
	}
	s.mu.Lock()
	s.cache[stationID] = quota
	s.mu.Unlock()
	return s.store.WriteJSON(Kind, stationID, quota)
}

func (s *Service) Remaining(stationID string) (float64, error) {
	quota, err := s.Get(stationID)
	if err != nil {
		return 0, err
	}
	return quota.DailyLimit - quota.Used, nil
}

func (s *Service) ResetDay(day string) error {
	stations, err := s.store.List(Kind)
	if err != nil {
		return err
	}
	for _, stationID := range stations {
		quota, err := s.Get(stationID)
		if err != nil {
			return err
		}
		quota.Used = 0
		quota.Day = day
		s.mu.Lock()
		s.cache[stationID] = quota
		s.mu.Unlock()
		if err := s.store.WriteJSON(Kind, stationID, quota); err != nil {
			return err
		}
	}
	return nil
}
