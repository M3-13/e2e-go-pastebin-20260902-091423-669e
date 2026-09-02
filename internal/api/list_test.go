package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-pastebin/internal/store"
)

func seedListStore(t *testing.T) *store.Store {
	t.Helper()
	s := store.NewStore()
	now := time.Now()

	older := now.Add(-2 * time.Hour)
	middle := now.Add(-1 * time.Hour)
	newer := now

	olderExp := now.Add(time.Hour)
	middleExp := now.Add(time.Hour)
	newerExp := now.Add(time.Hour)

	expiredExp := now.Add(-time.Hour)

	s.Insert(store.Paste{ID: "older", Content: "older content", Language: "text", CreatedAt: older, ExpiresAt: &olderExp})
	s.Insert(store.Paste{ID: "middle", Content: "middle content", Language: "go", CreatedAt: middle, ExpiresAt: &middleExp})
	s.Insert(store.Paste{ID: "newer", Content: "newer content", Language: "md", CreatedAt: newer, ExpiresAt: &newerExp})
	s.Insert(store.Paste{ID: "expired", Content: "expired content", Language: "text", CreatedAt: now, ExpiresAt: &expiredExp})

	return s
}

func getList(t *testing.T, s *store.Store) (*httptest.ResponseRecorder, []map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	rec := httptest.NewRecorder()
	ListHandler(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("expected JSON array, got %v", err)
	}
	return rec, got
}

func TestListReturnsArrayOfAllUnexpired(t *testing.T) {
	s := seedListStore(t)
	_, got := getList(t, s)

	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
}

func TestListEntriesHaveNoContentField(t *testing.T) {
	s := seedListStore(t)
	_, got := getList(t, s)

	for _, entry := range got {
		if _, ok := entry["content"]; ok {
			t.Fatalf("entry %v must not contain content field", entry)
		}
		for _, key := range []string{"id", "language", "created_at", "expires_at"} {
			if _, ok := entry[key]; !ok {
				t.Fatalf("entry %v missing field %q", entry, key)
			}
		}
	}
}

func TestListSortedByCreatedAtDescending(t *testing.T) {
	s := seedListStore(t)
	_, got := getList(t, s)

	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}

	ids := make([]string, len(got))
	for i, entry := range got {
		ids[i] = entry["id"].(string)
	}

	want := []string{"newer", "middle", "older"}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("expected order %v, got %v", want, ids)
		}
	}
}

func TestListOmitsExpiredEntries(t *testing.T) {
	s := seedListStore(t)
	_, got := getList(t, s)

	for _, entry := range got {
		if entry["id"] == "expired" {
			t.Fatalf("expired entry must not appear in list: %v", got)
		}
	}
}

func TestListEmptyReturnsEmptyArray(t *testing.T) {
	s := store.NewStore()
	rec, got := getList(t, s)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got == nil {
		t.Fatal("expected JSON array, got null")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty array, got %d entries", len(got))
	}
}
