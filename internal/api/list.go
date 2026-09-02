package api

import (
	"net/http"
	"time"

	"go-pastebin/internal/store"
)

type metaJSON struct {
	ID        string     `json:"id"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func ListHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metas := s.List()
		response := make([]metaJSON, 0, len(metas))
		for _, m := range metas {
			response = append(response, metaJSON{
				ID:        m.ID,
				Language:  m.Language,
				CreatedAt: m.CreatedAt,
				ExpiresAt: m.ExpiresAt,
			})
		}
		WriteJSON(w, http.StatusOK, response)
	}
}
