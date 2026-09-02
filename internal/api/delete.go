package api

import (
	"net/http"

	"go-pastebin/internal/store"
)

func DeleteHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !s.Delete(id) {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
