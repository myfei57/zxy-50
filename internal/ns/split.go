package ns

import (
	"drainnet/internal/audit"
)

func (s *Service) SplitCatchment(catchmentID string, zone Zone) (Catchment, error) {
	catchment, err := s.GetCatchment(catchmentID)
	if err != nil {
		return Catchment{}, err
	}
	zone, err = s.AddZone(catchmentID, zone)
	if err != nil {
		return Catchment{}, err
	}
	catchment, err = s.GetCatchment(catchmentID)
	if err != nil {
		return Catchment{}, err
	}
	catchment.SplitEpoch++
	if err := s.store.WriteJSON(CatchmentKind, catchment.ID, catchment); err != nil {
		return Catchment{}, err
	}
	if err := s.audit.Record(audit.Event{Type: "ns.split", EntityID: zone.ID, Message: "catchment split added zone"}); err != nil {
		return Catchment{}, err
	}
	return catchment, nil
}
