package level

import (
	"drainnet/internal/store"
)

func (s *Service) LastSample(stationID string) (Sample, error) {
	ids, err := s.store.List(Kind)
	if err != nil {
		return Sample{}, err
	}
	for index := len(ids) - 1; index >= 0; index-- {
		var sample Sample
		if err := s.store.ReadJSON(Kind, ids[index], &sample); err != nil {
			return Sample{}, err
		}
		if sample.StationID == stationID {
			return sample, nil
		}
	}
	return Sample{}, store.ErrNotFound
}

func (s *Service) History(stationID string) (History, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	history := s.history[stationID]
	if history == nil {
		return History{}, store.ErrNotFound
	}
	return *history, nil
}
