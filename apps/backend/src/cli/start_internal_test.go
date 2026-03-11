package cli

import (
	"net"
	"testing"
)

func TestIsPortInUse_FreePort(t *testing.T) {
	// Find a free port and verify isPortInUse returns false
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("could not open listener: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	if isPortInUse(port) {
		t.Errorf("port %d should be free but isPortInUse returned true", port)
	}
}

func TestIsPortInUse_OccupiedPort(t *testing.T) {
	// Open a listener and verify isPortInUse returns true
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("could not open listener: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	if !isPortInUse(port) {
		t.Errorf("port %d should be occupied but isPortInUse returned false", port)
	}
}
