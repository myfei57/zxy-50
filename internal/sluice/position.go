package sluice

import (
	"strconv"
	"sync"

	"drainnet/internal/audit"
)

type Position struct {
	mu   sync.Mutex
	pos  int
	hist []int
}

func NewPosition() *Position {
	return &Position{hist: []int{0}}
}

func (p *Position) Apply(delta int) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pos += delta
	if p.pos < 0 {
		p.pos = 0
	}
	if p.pos > 100 {
		p.pos = 100
	}
	p.hist = append(p.hist, p.pos)
	return p.pos
}

func (p *Position) Set(target int) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if target < 0 {
		target = 0
	}
	if target > 100 {
		target = 100
	}
	p.pos = target
	p.hist = append(p.hist, p.pos)
	return p.pos
}

func (p *Position) Current() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pos
}

func (p *Position) History() []int {
	p.mu.Lock()
	defer p.mu.Unlock()
	copied := append([]int(nil), p.hist...)
	return copied
}

func (r *Registry) positionFor(gateID string) *Position {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pos[gateID]
}

func (r *Registry) Position(gateID string) (int, error) {
	position := r.positionFor(gateID)
	if position == nil {
		return 0, ErrGateNotFound
	}
	return position.Current(), nil
}

func (r *Registry) ApplyDelta(gateID string, delta int) error {
	position := r.positionFor(gateID)
	if position == nil {
		return ErrGateNotFound
	}
	current := position.Apply(delta)
	gate, err := r.Gate(gateID)
	if err != nil {
		return err
	}
	gate.Position = current
	if err := r.persistGate(gate); err != nil {
		return err
	}
	return r.audit.Record(auditEvent(gateID, "sluice.position", current))
}

func (r *Registry) SetPosition(gateID string, target int) error {
	position := r.positionFor(gateID)
	if position == nil {
		return ErrGateNotFound
	}
	current := position.Set(target)
	gate, err := r.Gate(gateID)
	if err != nil {
		return err
	}
	gate.Position = current
	if err := r.persistGate(gate); err != nil {
		return err
	}
	return r.audit.Record(auditEvent(gateID, "sluice.set", current))
}

func (r *Registry) History(gateID string) ([]int, error) {
	position := r.positionFor(gateID)
	if position == nil {
		return nil, ErrGateNotFound
	}
	return position.History(), nil
}

func auditEvent(gateID string, eventType string, position int) audit.Event {
	return audit.Event{
		Type:     eventType,
		EntityID: gateID,
		Message:  "gate position changed",
		Meta:     map[string]string{"position": formatInt(position)},
	}
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}
