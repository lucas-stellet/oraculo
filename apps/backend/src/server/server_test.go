package server_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/server"
)

func TestHealth_IncludesProjectName(t *testing.T) {
	srv := server.New(nil, nil, nil, nil, "test-project", "test")
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["project_name"] != "test-project" {
		t.Errorf("expected project_name 'test-project', got %q", body["project_name"])
	}
}

func TestLastActivity_UpdatedOnRequest(t *testing.T) {
	srv := testServer(t)

	before := srv.LastActivity()

	time.Sleep(10 * time.Millisecond)
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	after := srv.LastActivity()
	if !after.After(before) {
		t.Errorf("LastActivity not updated: before=%v after=%v", before, after)
	}
}

func TestListenAndServe_ShutdownOnIdle(t *testing.T) {
	srv := testServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idleTimeout := 200 * time.Millisecond
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, 0, idleTimeout)
	}()

	// Give server time to start then let it idle
	time.Sleep(50 * time.Millisecond)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Server shut down due to idle — success
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("server did not shut down after idle timeout")
	}
}
