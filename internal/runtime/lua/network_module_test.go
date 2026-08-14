package modruntime

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"
)

type testNetworkController struct{ hosted, started, cancelled, joined bool }

func (controller *testNetworkController) Host() error          { controller.hosted = true; return nil }
func (controller *testNetworkController) StartSelected() error { controller.started = true; return nil }
func (controller *testNetworkController) Cancel()              { controller.cancelled = true }
func (controller *testNetworkController) Join(address string) error {
	controller.joined = address == "host:4433"
	return nil
}

type rejectingNetworkController struct{}

func (*rejectingNetworkController) Host() error            { return errors.New("not ready") }
func (*rejectingNetworkController) StartSelected() error   { return errors.New("not ready") }
func (*rejectingNetworkController) Cancel()                {}
func (*rejectingNetworkController) Join(string) error      { return errors.New("not ready") }
func (*rejectingNetworkController) Status() map[string]any { return map[string]any{"phase": "failed"} }
func (*testNetworkController) Status() map[string]any {
	return map[string]any{"phase": "connected", "mode": "host"}
}

func TestNetworkModuleTreatsRejectedRequestsAsRecoverableUIState(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(NetworkModule(&rejectingNetworkController{})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
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

func TestNetworkModuleTransportsIntentAndCopiedStatus(t *testing.T) {
	controller := &testNetworkController{}
	runtime := New()
	if err := runtime.RegisterModule(NetworkModule(controller)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local network = require("engine.network/v1")
network.host()
network.start_selected()
network.cancel()
network.join("host:4433")
local status = network.status()
assert(status.phase == "connected" and status.mode == "host")
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if !controller.hosted || !controller.started || !controller.cancelled || !controller.joined {
		t.Fatal("network intents did not reach controller")
	}
}
