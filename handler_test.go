package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newServer() *http.ServeMux {
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewStore())
	return mux
}

func do(mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestHealthz(t *testing.T) {
	if rec := do(newServer(), "GET", "/healthz", ""); rec.Code != http.StatusOK {
		t.Fatalf("healthz: got %d, want 200", rec.Code)
	}
}

func TestCreateThenGet(t *testing.T) {
	mux := newServer()
	if rec := do(mux, "POST", "/api/v1/notes", `{"text":"halo"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d, want 201", rec.Code)
	}
	rec := do(mux, "GET", "/api/v1/notes/1", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "halo") {
		t.Fatalf("get: got %d %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRejectsEmptyText(t *testing.T) {
	if rec := do(newServer(), "POST", "/api/v1/notes", `{"text":""}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}

func TestUnknownNoteIs404(t *testing.T) {
	if rec := do(newServer(), "GET", "/api/v1/notes/99", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
}
