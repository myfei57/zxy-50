package station

func (r *Registry) SetPumpState(stationID string, ref PumpRef) (PumpRef, error) {
	value, err := r.Get(stationID)
	if err != nil {
		return PumpRef{}, err
	}
	for index := range value.Pumps {
		if value.Pumps[index].ID == ref.ID {
			value.Pumps[index] = ref
			if err := r.Update(value); err != nil {
				return PumpRef{}, err
			}
			return ref, nil
		}
	}
	value.Pumps = append(value.Pumps, ref)
	if err := r.Update(value); err != nil {
		return PumpRef{}, err
	}
	return ref, nil
}

func (r *Registry) StandbysOf(stationID string) ([]PumpRef, error) {
	value, err := r.Get(stationID)
	if err != nil {
		return nil, err
	}
	standbys := make([]PumpRef, 0, len(value.Pumps))
	for _, ref := range value.Pumps {
		if ref.Duty == "standby" {
			standbys = append(standbys, ref)
		}
	}
	return standbys, nil
}

func (r *Registry) LeadOf(stationID string) (PumpRef, error) {
	value, err := r.Get(stationID)
	if err != nil {
		return PumpRef{}, err
	}
	if value.LeadPumpID == "" {
		for _, ref := range value.Pumps {
			if ref.Duty == "lead" {
				return ref, nil
			}
		}
		return PumpRef{}, ErrStationNotFound
	}
	for _, ref := range value.Pumps {
		if ref.ID == value.LeadPumpID {
			return ref, nil
		}
	}
	return PumpRef{}, ErrStationNotFound
}

func (r *Registry) CaptureSnapshot(stationID string) (LeadSnapshot, error) {
	value, err := r.Get(stationID)
	if err != nil {
		return LeadSnapshot{}, err
	}
	state := r.LeadStateFor(stationID)
	leadID := ""
	if state != nil {
		leadID = state.Snapshot()
	}
	breakerID := ""
	for _, ref := range value.Pumps {
		if ref.ID == leadID {
			breakerID = ref.BreakerID
			break
		}
	}
	return LeadSnapshot{LeadID: leadID, BreakerID: breakerID}, nil
}

func (r *Registry) ApplySnapshot(stationID string, snapshot LeadSnapshot) error {
	state := r.LeadStateFor(stationID)
	if state == nil {
		return ErrStationNotFound
	}
	state.ApplySnapshot(snapshot.LeadID)
	return nil
}
