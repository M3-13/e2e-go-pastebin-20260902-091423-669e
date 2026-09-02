package store

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func (s *Store) Create(content, language string, expiresInSeconds int) (id string) {
	now := time.Now()
	expires := now.Add(time.Duration(expiresInSeconds) * time.Second)

	for {
		id = newID()
		s.mu.RLock()
		_, exists := s.pastes[id]
		s.mu.RUnlock()
		if !exists {
			break
		}
	}

	s.Insert(Paste{
		ID:        id,
		Content:   content,
		Language:  language,
		CreatedAt: now,
		ExpiresAt: &expires,
	})

	return id
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
