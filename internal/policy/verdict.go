package policy

func (s *Service) CheckReverse(stationID string) (bool, error) {
	direction, err := s.levels.Direction(stationID)
	if err != nil {
		return false, err
	}
	s.levels.CommitDirection(stationID)
	return direction == "in", nil
}
