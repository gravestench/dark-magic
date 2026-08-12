package modruntime

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestInitialDataModuleReturnsValueCopies(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(InitialDataModule(map[string]any{"fixture": map[string]any{"name": "hero", "values": []any{float64(1), float64(2)}}})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	if err := runtime.Run(t.Context(), func(state *lua.LState) error {
		return state.DoString(`local data=require("engine.initial_data/v1"); local a=data.get("fixture"); a.name="changed"; local b=data.get("fixture"); assert(b.name=="hero" and b.values[2]==2)`)
	}); err != nil {
		t.Fatal(err)
	}
}
