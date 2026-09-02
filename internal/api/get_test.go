package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-pastebin/internal/store"
)

func newGetRouter(s *store.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /pastes/{id}", GetHandler(s))
	return mux
}

func TestGetKnownID(t *testing.T) {
	s := store.NewStore()
	expires := time.Now().Add(time.Hour)
	s.Insert(store.Paste{
		ID:        "abc123",
		Content:   "hello world",
		Language:  "text",
		CreatedAt: time.Now(),
		ExpiresAt: &expires,
	})

	req := httptest.NewRequest(http.MethodGet, "/pastes/abc123", nil)
	rec := httptest.NewRecorder()
	newGetRouter(s).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["id"] != "abc123" {
		t.Fatalf("expected id 'abc123', got %q", body["id"])
	}
	if body["content"] != "hello world" {
		t.Fatalf("expected content 'hello world', got %q", body["content"])
	}
	if body["language"] != "text" {
		t.Fatalf("expected language 'text', got %q", body["language"])
	}
}

func TestGetUnknownID(t *testing.T) {
	s := store.NewStore()

	req := httptest.NewRequest(http.MethodGet, "/pastes/nope", nil)
	rec := httptest.NewRecorder()
	newGetRouter(s).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("expected error field in response, got %v", body)
	}
}

func TestGetExpiredPaste(t *testing.T) {
	s := store.NewStore()
	expires := time.Now().Add(-time.Hour)
	s.Insert(store.Paste{
		ID:        "expired1",
		Content:   "gone",
		Language:  "text",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: &expires,
	})

	req := httptest.NewRequest(http.MethodGet, "/pastes/expired1", nil)
	rec := httptest.NewRecorder()
	newGetRouter(s).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("expected error field in response, got %v", body)
	}
}
