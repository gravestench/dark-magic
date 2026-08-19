package runtimeapi

import (
	"context"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/host"
)

// TestServerLeavesBlankAddressesDisabled verifies that optional runtime API configuration performs no socket ownership
// or cleanup work, including when the address contains only whitespace.
func TestServerLeavesBlankAddressesDisabled(t *testing.T) {
	server := New(" \t", host.NewManager())

	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("start disabled server: %v", err)
	}

	if server.listen != nil {
		t.Fatalf("disabled server acquired listener %v", server.listen.Addr())
	}

	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("stop disabled server: %v", err)
	}
}

// TestServerReportsListenFailures verifies that address errors remain synchronous and retain the runtimeapi context
// needed to distinguish host startup failures from later serving failures. An invalid address keeps the test
// independent of whether its environment permits loopback socket ownership.
func TestServerReportsListenFailures(t *testing.T) {
	server := New("missing-port", host.NewManager())

	err := server.Start(context.Background())
	if err == nil || !strings.Contains(err.Error(), "runtimeapi: listen:") {
		t.Fatalf("start error=%v", err)
	}
}
