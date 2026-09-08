package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	NewVisionServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	// json.Encoder.Encode terminates every value with a newline.
	if got, want := rec.Body.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

// A wrong method is a 404 rather than a 405: the "/" catch-all matches every
// method, so ServeMux never reaches its method-not-allowed path. What matters is
// that the "/api/" subtree answers, instead of the request falling through to
// the frontend and returning HTML with status 200.
func TestHealthRejectsNonGET(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/api/health", nil)
		rec := httptest.NewRecorder()

		NewVisionServer().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("%s /api/health = %d, want %d", method, rec.Code, http.StatusNotFound)
		}
	}
}

// An unrouted API path must not be answered by the SPA fallback. Returning
// index.html here would give fetch() HTML where it expects JSON, surfacing as a
// parse error rather than as the 404 it really is.
func TestUnknownAPIPathReturns404(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil)
	rec := httptest.NewRecorder()

	NewVisionServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	if strings.Contains(rec.Body.String(), "doctype") {
		t.Errorf("API 404 answered with HTML: %q", rec.Body.String())
	}
}

// The counterpart: a path that looks like a client-side route does get the app,
// so a deep link survives a reload.
func TestUnknownPathServesApp(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/instances/cam0/calibration", nil)
	rec := httptest.NewRecorder()

	NewVisionServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
