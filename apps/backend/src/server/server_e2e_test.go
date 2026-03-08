package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestHTTPServer_HealthEndpoint_E2E starts a real server on a random port,
// hits /health via HTTP, and validates the response.
func TestHTTPServer_HealthEndpoint_E2E(t *testing.T) {
	srv := testServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use port 0 for random ephemeral port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close() // free it so ListenAndServe can rebind

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, port, 500*time.Millisecond)
	}()

	// Wait for server to be ready
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !waitForReady(healthURL, 100*time.Millisecond, 2*time.Second) {
		cancel()
		t.Fatal("server did not become ready")
	}

	// Hit /health
	resp, err := http.Get(healthURL)
	if err != nil {
		cancel()
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("invalid JSON: %s", body)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q, want %q", result["status"], "ok")
	}

	// Cancel and wait for clean shutdown
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down after cancel")
	}
}

// TestHTTPServer_ActivityKeepsAlive verifies that requests reset the idle
// timer, so the server does NOT shut down while receiving traffic.
func TestHTTPServer_ActivityKeepsAlive(t *testing.T) {
	srv := testServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	idleTimeout := 300 * time.Millisecond

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, port, idleTimeout)
	}()

	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !waitForReady(healthURL, 50*time.Millisecond, 2*time.Second) {
		cancel()
		t.Fatal("server did not become ready")
	}

	// Send requests every 100ms for 600ms (2x idle timeout).
	// Server should NOT shut down because activity resets the timer.
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(600 * time.Millisecond)

	for {
		select {
		case <-deadline:
			// Server survived past 2x idle timeout — success.
			cancel()
			<-errCh
			return
		case <-ticker.C:
			resp, err := http.Get(healthURL)
			if err != nil {
				t.Fatalf("GET /health failed during keep-alive: %v", err)
			}
			resp.Body.Close()
		case err := <-errCh:
			t.Fatalf("server shut down unexpectedly while receiving traffic: %v", err)
		}
	}
}

// TestHTTPServer_IdleShutdown_E2E verifies the server shuts down after idle.
func TestHTTPServer_IdleShutdown_E2E(t *testing.T) {
	srv := testServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	idleTimeout := 200 * time.Millisecond

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, port, idleTimeout)
	}()

	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !waitForReady(healthURL, 50*time.Millisecond, 2*time.Second) {
		cancel()
		t.Fatal("server did not become ready")
	}

	// Do one request, then let it idle
	resp, _ := http.Get(healthURL)
	if resp != nil {
		resp.Body.Close()
	}

	// Server should shut down within idleTimeout + some margin
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Success: server shut down due to idle
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("server did not shut down after idle timeout")
	}

	// Verify server is no longer accepting connections
	_, err = http.Get(healthURL)
	if err == nil {
		t.Error("expected connection refused after shutdown, but got response")
	}
}

// TestHTTPServer_NoIdleTimeout_E2E verifies that passing 0 disables idle
// shutdown (the "start" backwards-compat mode).
func TestHTTPServer_NoIdleTimeout_E2E(t *testing.T) {
	srv := testServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, port, 0) // no idle timeout
	}()

	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !waitForReady(healthURL, 50*time.Millisecond, 2*time.Second) {
		cancel()
		t.Fatal("server did not become ready")
	}

	// Wait 300ms with no requests — server should NOT shut down
	select {
	case err := <-errCh:
		t.Fatalf("server shut down unexpectedly with idleTimeout=0: %v", err)
	case <-time.After(300 * time.Millisecond):
		// Good — still alive
	}

	// Clean up
	cancel()
	select {
	case <-errCh:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down after cancel")
	}
}

func waitForReady(url string, interval, timeout time.Duration) bool {
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return false
		default:
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				return true
			}
			time.Sleep(interval)
		}
	}
}
