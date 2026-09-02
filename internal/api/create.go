package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-pastebin/internal/store"
)

const (
	maxBodyBytes         = 1 << 20
	defaultExpirySeconds = 86400
	defaultLanguage      = "text"
)

type createRequest struct {
	Content          string `json:"content"`
	Language         string `json:"language"`
	ExpiresInSeconds *int   `json:"expires_in_seconds"`
}

func CreateHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var req createRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			WriteError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.Content == "" {
			WriteError(w, http.StatusBadRequest, "content must not be empty")
			return
		}

		if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds < 0 {
			WriteError(w, http.StatusBadRequest, "expires_in_seconds must not be negative")
			return
		}

		language, expires := resolveDefaults(req.Language, req.ExpiresInSeconds)

		id := s.Create(req.Content, language, expires)

		WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
	}
}

func resolveDefaults(language string, expiresInSeconds *int) (string, int) {
	lang := language
	if lang == "" {
		lang = defaultLanguage
	}
	expires := defaultExpirySeconds
	if expiresInSeconds != nil && *expiresInSeconds > 0 {
		expires = *expiresInSeconds
	}
	return lang, expires
}
