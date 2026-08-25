package station

import "sync"

type LeadSnapshot struct {
	LeadID    string
	BreakerID string
}

type LeadState struct {
	mu     sync.RWMutex
	leadID string
	shot   string
}

func NewLeadState() *LeadState {
	return &LeadState{}
}

func (l *LeadState) Current() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.leadID
}

func (l *LeadState) SetLead(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.leadID = id
	l.shot = id
}

func (l *LeadState) Snapshot() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.shot
}

func (l *LeadState) ApplySnapshot(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.shot = id
}
