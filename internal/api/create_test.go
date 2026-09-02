package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-pastebin/internal/store"
)

func postPaste(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	s := store.NewStore()
	req := httptest.NewRequest(http.MethodPost, "/pastes", strings.NewReader(body))
	rec := httptest.NewRecorder()
	CreateHandler(s)(rec, req)
	return rec
}

func assertHexID(t *testing.T, id string) {
	t.Helper()
	if len(id) != 16 {
		t.Fatalf("expected 16-char hex id, got %q (len %d)", id, len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		t.Fatalf("id is not valid hex: %q", id)
	}
}

func TestCreateHandlerSuccess(t *testing.T) {
	rec := postPaste(t, `{"content":"hello world","language":"go","expires_in_seconds":3600}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	assertHexID(t, resp.ID)
}

func TestCreateHandlerInvalidJSON(t *testing.T) {
	rec := postPaste(t, `{"content": `)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected error JSON, got %v", err)
	}
	if resp.Error == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestCreateHandlerEmptyContent(t *testing.T) {
	for _, body := range []string{`{"content":""}`, `{}`, `{"language":"go"}`} {
		rec := postPaste(t, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %q: expected status 400, got %d", body, rec.Code)
		}
	}
}

func TestCreateHandlerNegativeExpires(t *testing.T) {
	rec := postPaste(t, `{"content":"x","expires_in_seconds":-1}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateHandlerOversizedBody(t *testing.T) {
	body := `{"content":"` + strings.Repeat("a", maxBodyBytes+1) + `"}`
	rec := postPaste(t, body)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", rec.Code)
	}
}

func TestCreateHandlerDefaults(t *testing.T) {
	for _, body := range []string{`{"content":"x"}`, `{"content":"x","expires_in_seconds":0}`} {
		rec := postPaste(t, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("body %q: expected status 201, got %d", body, rec.Code)
		}

		var resp struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}
		assertHexID(t, resp.ID)
	}
}

func TestResolveDefaults(t *testing.T) {
	lang, exp := resolveDefaults("", nil)
	if lang != "text" || exp != 86400 {
		t.Fatalf("expected defaults text/86400, got %q/%d", lang, exp)
	}

	zero := 0
	lang, exp = resolveDefaults("", &zero)
	if lang != "text" || exp != 86400 {
		t.Fatalf("expected defaults text/86400 for expires 0, got %q/%d", lang, exp)
	}

	pos := 3600
	lang, exp = resolveDefaults("go", &pos)
	if lang != "go" || exp != 3600 {
		t.Fatalf("expected go/3600, got %q/%d", lang, exp)
	}
}
