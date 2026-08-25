package policy

func (s *Service) LoadRules() (Rules, error) {
	var rules Rules
	err := s.store.ReadJSON("policy_rules", "current", &rules)
	if err != nil {
		return s.Rules(), nil
	}
	return rules, nil
}

func (s *Service) GateRule() string {
	return s.Rules().GateRule
}
