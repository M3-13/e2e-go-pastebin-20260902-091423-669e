package api

import (
	"net/http"

	"go-pastebin/internal/store"
)

func ListHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotImplemented, "not implemented")
	}
}
