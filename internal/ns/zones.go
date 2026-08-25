package ns

import (
	"github.com/google/uuid"

	"drainnet/internal/audit"
)

func (s *Service) AddZone(catchmentID string, zone Zone) (Zone, error) {
	catchment, err := s.GetCatchment(catchmentID)
	if err != nil {
		return Zone{}, err
	}
	if zone.ID == "" {
		zone.ID = uuid.NewString()
	}
	zone.CatchmentID = catchmentID
	catchment.ZoneIDs = append(catchment.ZoneIDs, zone.ID)
	if err := s.store.WriteJSON(ZoneKind, zone.ID, zone); err != nil {
		return Zone{}, err
	}
	if err := s.store.WriteJSON(CatchmentKind, catchment.ID, catchment); err != nil {
		return Zone{}, err
	}
	if err := s.audit.Record(audit.Event{Type: "ns.zone", EntityID: zone.ID, Message: "zone added"}); err != nil {
		return Zone{}, err
	}
	return zone, nil
}

func (s *Service) Zones(catchmentID string) ([]Zone, error) {
	catchment, err := s.GetCatchment(catchmentID)
	if err != nil {
		return nil, err
	}
	zones := make([]Zone, 0, len(catchment.ZoneIDs))
	for _, id := range catchment.ZoneIDs {
		zone, err := s.ZoneByID(id)
		if err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}
	return zones, nil
}

func (s *Service) ZoneByID(id string) (Zone, error) {
	var zone Zone
	err := s.store.ReadJSON(ZoneKind, id, &zone)
	if err != nil {
		return Zone{}, ErrZoneNotFound
	}
	return zone, nil
}

func (s *Service) AssignStation(zoneID string, stationID string) error {
	zone, err := s.ZoneByID(zoneID)
	if err != nil {
		return err
	}
	zone.StationID = stationID
	if err := s.store.WriteJSON(ZoneKind, zone.ID, zone); err != nil {
		return err
	}
	return s.audit.Record(audit.Event{Type: "ns.assign", EntityID: zoneID, Message: "zone assigned to station"})
}
