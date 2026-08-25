package level

import "sync"

type Direction struct {
	mu      sync.RWMutex
	current string
	pending string
}

func NewDirection() *Direction {
	return &Direction{current: "out", pending: "out"}
}

func (d *Direction) Update(next string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.current = next
	d.pending = next
}

func (d *Direction) Commit() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.current = d.pending
}

func (d *Direction) Current() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.current
}

func (s *Service) Direction(stationID string) (string, error) {
	s.mu.RLock()
	direction := s.dirs[stationID]
	s.mu.RUnlock()
	if direction == nil {
		return "out", nil
	}
	return direction.Current(), nil
}

func (s *Service) CommitDirection(stationID string) {
	s.mu.RLock()
	direction := s.dirs[stationID]
	s.mu.RUnlock()
	if direction != nil {
		direction.Commit()
	}
}
