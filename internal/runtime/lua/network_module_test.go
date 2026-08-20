package modruntime

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"
)

type testNetworkController struct{ hosted, started, cancelled, joined bool }

// Host owns the host step at this boundary, keeping its side effects and failure point explicit to callers.
func (controller *testNetworkController) Host() error { controller.hosted = true; return nil }

// StartSelected orders start selected initialization so callers only observe the capability after its dependencies
// are ready.
func (controller *testNetworkController) StartSelected() error { controller.started = true; return nil }

// Cancel owns the cancel step at this boundary, keeping its side effects and failure point explicit to callers.
func (controller *testNetworkController) Cancel() { controller.cancelled = true }

// Join owns the join step at this boundary, keeping its side effects and failure point explicit to callers.
func (controller *testNetworkController) Join(address string) error {
	controller.joined = address == "host:4433"
	return nil
}

type rejectingNetworkController struct{}

// Host owns the host step at this boundary, keeping its side effects and failure point explicit to callers.
func (*rejectingNetworkController) Host() error { return errors.New("not ready") }

// StartSelected orders start selected initialization so callers only observe the capability after its dependencies
// are ready.
func (*rejectingNetworkController) StartSelected() error { return errors.New("not ready") }

// Cancel owns the cancel step at this boundary, keeping its side effects and failure point explicit to callers.
func (*rejectingNetworkController) Cancel() {}

// Join owns the join step at this boundary, keeping its side effects and failure point explicit to callers.
func (*rejectingNetworkController) Join(string) error { return errors.New("not ready") }

// Status returns a stable status observation without exposing mutable runtime state to callers.
func (*rejectingNetworkController) Status() map[string]any { return map[string]any{"phase": "failed"} }

// Status returns a stable status observation without exposing mutable runtime state to callers.
func (*testNetworkController) Status() map[string]any {
	return map[string]any{"phase": "connected", "mode": "host"}
}

// TestNetworkModuleTreatsRejectedRequestsAsRecoverableUIState protects the network module treats rejected requests
// as recoverable uistate contract, including its observable ordering and failure behavior.
func TestNetworkModuleTreatsRejectedRequestsAsRecoverableUIState(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(NetworkModule(&rejectingNetworkController{})); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(t.Context()) }()

	if err := runtime.Execute(t.Context(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local network = require("engine.network/v1")
assert(network.host() == false)
assert(network.start_selected() == false)
assert(network.join("") == false)
assert(network.status().phase == "failed")
`)}}, "test.lua"); err != nil {
		t.Fatalf("recoverable rejection escaped into scene: %v", err)
	}
}

// TestNetworkModuleTransportsIntentAndCopiedStatus protects the network module transports intent and copied status
// contract, including its observable ordering and failure behavior.
func TestNetworkModuleTransportsIntentAndCopiedStatus(t *testing.T) {
	controller := &testNetworkController{}

	runtime := New()
	if err := runtime.RegisterModule(NetworkModule(controller)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(t.Context()) }()

	if err := runtime.Execute(
		context.Background(),
		fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local network = require("engine.network/v1")
network.host()
network.start_selected()
network.cancel()
network.join("host:4433")
local status = network.status()
assert(status.phase == "connected" and status.mode == "host")
`)}},
		"test.lua",
	); err != nil {
		t.Fatal(err)
	}

	if !controller.hosted || !controller.started || !controller.cancelled || !controller.joined {
		t.Fatal("network intents did not reach controller")
	}
}
