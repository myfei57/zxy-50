package station

import (
	"drainnet/internal/audit"
	"drainnet/internal/ns"
)

func (r *Registry) RefreshMapping(stationID string, zones []ns.Zone) error {
	value, err := r.Get(stationID)
	if err != nil {
		return err
	}
	r.mu.Lock()
	value.MappingEpoch = r.mapping.Rebuild(zones)
	r.mu.Unlock()
	if err := r.Update(value); err != nil {
		return err
	}
	return r.audit.Record(audit.Event{Type: "station.mapping", EntityID: stationID, Message: "zone mapping refreshed"})
}

func (r *Registry) StationForZone(zoneID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if resolved, ok := r.mapping.Lookup(zoneID); ok {
		return resolved, nil
	}
	if resolved := r.mapping.Default(); resolved != "" {
		return resolved, nil
	}
	return "", ErrStationNotFound
}

func (r *Registry) SetLead(stationID string, pumpID string) error {
	value, err := r.Get(stationID)
	if err != nil {
		return err
	}
	found := false
	for _, ref := range value.Pumps {
		if ref.ID == pumpID {
			found = true
			break
		}
	}
	if !found {
		return ErrStationNotFound
	}
	r.mu.Lock()
	state := r.lead[stationID]
	if state == nil {
		state = NewLeadState()
		r.lead[stationID] = state
	}
	state.SetLead(pumpID)
	r.mu.Unlock()
	value.LeadPumpID = pumpID
	if err := r.Update(value); err != nil {
		return err
	}
	return r.audit.Record(audit.Event{Type: "station.lead", EntityID: pumpID, Message: "lead pump changed"})
}

func (r *Registry) LeadStateFor(stationID string) *LeadState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lead[stationID]
}

func (r *Registry) AllocateBreaker(stationID string, pumpID string) (string, error) {
	value, err := r.Get(stationID)
	if err != nil {
		return "", err
	}
	r.mu.Lock()
	pool := r.pools[stationID]
	r.mu.Unlock()
	if pool == nil {
		return "", ErrStationNotFound
	}
	breakerID, err := pool.Allocate(pumpID)
	if err != nil {
		return "", err
	}
	for index := range value.Breakers {
		if value.Breakers[index].ID == breakerID {
			value.Breakers[index].PumpID = pumpID
		}
	}
	if err := r.Update(value); err != nil {
		return "", err
	}
	return breakerID, nil
}

func (r *Registry) ReleaseBreaker(stationID string, breakerID string) error {
	value, err := r.Get(stationID)
	if err != nil {
		return err
	}
	r.mu.Lock()
	pool := r.pools[stationID]
	r.mu.Unlock()
	if pool != nil {
		pool.Release(breakerID)
	}
	for index := range value.Breakers {
		if value.Breakers[index].ID == breakerID {
			value.Breakers[index].PumpID = ""
		}
	}
	for index := range value.Pumps {
		if value.Pumps[index].BreakerID == breakerID {
			value.Pumps[index].BreakerID = ""
		}
	}
	return r.Update(value)
}

func (r *Registry) BreakerHolder(stationID string, breakerID string) string {
	r.mu.RLock()
	pool := r.pools[stationID]
	r.mu.RUnlock()
	if pool == nil {
		return ""
	}
	return pool.Holder(breakerID)
}

func (r *Registry) PoolStats(stationID string) (int, map[string]string, error) {
	r.mu.RLock()
	pool := r.pools[stationID]
	r.mu.RUnlock()
	if pool == nil {
		return 0, map[string]string{}, ErrStationNotFound
	}
	return pool.FreeCount(), pool.Held(), nil
}

func (r *Registry) Delete(stationID string) error {
	r.mu.Lock()
	delete(r.lead, stationID)
	delete(r.pools, stationID)
	r.mu.Unlock()
	if err := r.store.Delete(Kind, stationID); err != nil {
		return err
	}
	return r.audit.Record(audit.Event{Type: "station.delete", EntityID: stationID, Message: "station removed"})
}

func (r *Registry) MappingVersion() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mapping.Version()
}
