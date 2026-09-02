package store

import (
	"sync"
	"time"
)

type Paste struct {
	ID        string
	Content   string
	Language  string
	CreatedAt time.Time
	ExpiresAt *time.Time
}

type Meta struct {
	ID        string
	Language  string
	CreatedAt time.Time
	ExpiresAt *time.Time
}

type Store struct {
	mu     sync.RWMutex
	pastes map[string]Paste
}

func NewStore() *Store {
	return &Store{
		pastes: make(map[string]Paste),
	}
}

func (s *Store) Insert(p Paste) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pastes[p.ID] = p
}
