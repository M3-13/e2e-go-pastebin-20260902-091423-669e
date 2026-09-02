package api

import (
	"net/http"
	"time"

	"go-pastebin/internal/store"
)

func GetHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		p, ok := s.Get(id)
		if !ok {
			WriteError(w, http.StatusNotFound, "paste not found")
			return
		}

		expiresAt := ""
		if p.ExpiresAt != nil {
			expiresAt = p.ExpiresAt.Format(time.RFC3339)
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"id":         p.ID,
			"content":    p.Content,
			"language":   p.Language,
			"created_at": p.CreatedAt.Format(time.RFC3339),
			"expires_at": expiresAt,
		})
	}
}
