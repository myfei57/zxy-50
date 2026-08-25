package policy

func (s *Service) CheckReverse(stationID string) (bool, error) {
	s.levels.CommitDirection(stationID)
	direction, err := s.levels.Direction(stationID)
	if err != nil {
		return false, err
	}
	return direction == "in", nil
}
