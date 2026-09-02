package store

import (
	"sort"
	"time"
)

func (s *Store) List() []Meta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	metas := make([]Meta, 0, len(s.pastes))
	for _, p := range s.pastes {
		if p.ExpiresAt != nil && !p.ExpiresAt.After(now) {
			continue
		}
		metas = append(metas, Meta{
			ID:        p.ID,
			Language:  p.Language,
			CreatedAt: p.CreatedAt,
			ExpiresAt: p.ExpiresAt,
		})
	}

	sort.Slice(metas, func(i, j int) bool {
		return metas[i].CreatedAt.After(metas[j].CreatedAt)
	})

	return metas
}
