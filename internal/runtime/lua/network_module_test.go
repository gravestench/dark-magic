package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

type testNetworkController struct{ hosted, joined bool }

func (controller *testNetworkController) Host() error { controller.hosted = true; return nil }
func (controller *testNetworkController) Join(address string) error {
	controller.joined = address == "host:4433"
	return nil
}
func (*testNetworkController) Status() map[string]any {
	return map[string]any{"phase": "connected", "mode": "host"}
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
network.join("host:4433")
local status = network.status()
assert(status.phase == "connected" and status.mode == "host")
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if !controller.hosted || !controller.joined {
		t.Fatal("network intents did not reach controller")
	}
}
