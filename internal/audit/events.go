package audit

import (
	"sort"
	"time"
)

func (s *Service) Between(from time.Time, to time.Time) ([]Event, error) {
	events, err := s.List()
	if err != nil {
		return nil, err
	}
	filtered := make([]Event, 0, len(events))
	for _, event := range events {
		if !event.At.Before(from) && !event.At.After(to) {
			filtered = append(filtered, event)
		}
	}
	return filtered, nil
}

func (s *Service) Types() ([]string, error) {
	events, err := s.List()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, event := range events {
		seen[event.Type] = true
	}
	types := make([]string, 0, len(seen))
	for eventType := range seen {
		types = append(types, eventType)
	}
	sort.Strings(types)
	return types, nil
}
