package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDashboardStaticFiles(t *testing.T) {
	srv := New(nil, nil, nil, nil, "")

	// Test root path returns HTML
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected text/html, got %s", contentType)
	}
}

func TestDashboard404(t *testing.T) {
	srv := New(nil, nil, nil, nil, "")

	// Test non-existent path returns 404
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
