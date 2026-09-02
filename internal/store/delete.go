package store

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.pastes[id]; !ok {
		return false
	}
	delete(s.pastes, id)
	return true
}
