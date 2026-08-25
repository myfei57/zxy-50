package station

import (
	"sync"

	"drainnet/internal/ns"
)

type Mapping struct {
	mu      sync.RWMutex
	entries map[string]string
	frozen  map[string]string
	epoch   int
}

func NewMapping() *Mapping {
	return &Mapping{entries: map[string]string{}, frozen: map[string]string{}}
}

func (m *Mapping) Set(zoneID string, stationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[zoneID] = stationID
	m.frozen[zoneID] = stationID
}

func (m *Mapping) Rebuild(zones []ns.Zone) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	next := map[string]string{}
	for _, zone := range zones {
		if zone.StationID != "" {
			next[zone.ID] = zone.StationID
		}
	}
	m.entries = next
	if len(m.frozen) == 0 {
		m.frozen = map[string]string{}
		for zoneID, stationID := range next {
			m.frozen[zoneID] = stationID
		}
	}
	m.epoch++
	return m.epoch
}

func (m *Mapping) Lookup(zoneID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stationID, ok := m.frozen[zoneID]
	return stationID, ok
}

func (m *Mapping) Default() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, stationID := range m.frozen {
		return stationID
	}
	return ""
}

func (m *Mapping) Version() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.epoch
}
