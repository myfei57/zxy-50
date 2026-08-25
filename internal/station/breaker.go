package station

import (
	"errors"
	"sync"
)

var ErrNoBreaker = errors.New("no free breaker")

type BreakerPool struct {
	mu    sync.Mutex
	names []string
	taken map[string]string
}

func NewBreakerPool(names []string) *BreakerPool {
	return &BreakerPool{names: names, taken: map[string]string{}}
}

func (p *BreakerPool) Allocate(pumpID string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, breaker := range p.names {
		if _, ok := p.taken[breaker]; !ok {
			p.taken[breaker] = pumpID
			return breaker, nil
		}
	}
	return "", ErrNoBreaker
}

func (p *BreakerPool) Release(breakerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.taken, breakerID)
}

func (p *BreakerPool) Holder(breakerID string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.taken[breakerID]
}

func (p *BreakerPool) FreeCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	count := 0
	for _, breaker := range p.names {
		if _, ok := p.taken[breaker]; !ok {
			count++
		}
	}
	return count
}

func (p *BreakerPool) Held() map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	copied := map[string]string{}
	for breaker, pump := range p.taken {
		copied[breaker] = pump
	}
	return copied
}
