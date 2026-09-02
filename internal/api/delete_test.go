package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-pastebin/internal/store"
)

func TestDeleteHandlerKnownID(t *testing.T) {
	s := store.NewStore()
	s.Insert(store.Paste{
		ID:        "0123456789abcdef",
		Content:   "hello world",
		Language:  "text",
		CreatedAt: time.Now(),
	})

	req := httptest.NewRequest(http.MethodDelete, "/pastes/0123456789abcdef", nil)
	req.SetPathValue("id", "0123456789abcdef")
	rec := httptest.NewRecorder()
	DeleteHandler(s)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}

	if _, ok := s.Get("0123456789abcdef"); ok {
		t.Fatal("expected paste to be gone after delete")
	}

	req2 := httptest.NewRequest(http.MethodDelete, "/pastes/0123456789abcdef", nil)
	req2.SetPathValue("id", "0123456789abcdef")
	rec2 := httptest.NewRecorder()
	DeleteHandler(s)(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for a second delete, got %d", rec2.Code)
	}
}

func TestDeleteHandlerUnknownID(t *testing.T) {
	s := store.NewStore()

	req := httptest.NewRequest(http.MethodDelete, "/pastes/ffffffffffffffff", nil)
	req.SetPathValue("id", "ffffffffffffffff")
	rec := httptest.NewRecorder()
	DeleteHandler(s)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestStoreDelete(t *testing.T) {
	s := store.NewStore()
	s.Insert(store.Paste{
		ID:        "abc",
		Content:   "x",
		Language:  "text",
		CreatedAt: time.Now(),
	})

	if !s.Delete("abc") {
		t.Fatal("Delete should return true for an existing id")
	}
	if s.Delete("abc") {
		t.Fatal("Delete should return false after the entry was removed")
	}
	if s.Delete("missing") {
		t.Fatal("Delete should return false for an unknown id")
	}
}
