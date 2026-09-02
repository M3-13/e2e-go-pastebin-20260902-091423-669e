package main

import (
	"log"
	"net/http"

	"go-pastebin/internal/api"
	"go-pastebin/internal/store"
)

func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		api.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func notFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.WriteError(w, http.StatusNotFound, "not found")
	}
}

func newRouter(s *store.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /pastes", api.CreateHandler(s))
	mux.HandleFunc("GET /pastes", api.ListHandler(s))
	mux.HandleFunc("/pastes", methodNotAllowed("GET, POST"))

	mux.HandleFunc("GET /pastes/{id}", api.GetHandler(s))
	mux.HandleFunc("DELETE /pastes/{id}", api.DeleteHandler(s))
	mux.HandleFunc("/pastes/{id}", methodNotAllowed("GET, DELETE"))

	mux.HandleFunc("GET /health", api.HealthHandler())
	mux.HandleFunc("/health", methodNotAllowed("GET"))

	mux.HandleFunc("/", notFound())

	return mux
}

func main() {
	s := store.NewStore()
	handler := newRouter(s)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
