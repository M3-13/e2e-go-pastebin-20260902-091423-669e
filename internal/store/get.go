package store

import "time"

func (s *Store) Get(id string) (Paste, bool) {
	s.mu.RLock()
	p, ok := s.pastes[id]
	s.mu.RUnlock()

	if !ok {
		return Paste{}, false
	}

	if p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now()) {
		s.mu.Lock()
		delete(s.pastes, id)
		s.mu.Unlock()
		return Paste{}, false
	}

	return p, true
}
