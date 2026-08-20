package modruntime

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/loading"
	lua "github.com/yuin/gopher-lua"
)

// TestLoadingModuleStartsNamedEngineDependencies protects the loading module starts named engine dependencies
// contract, including its observable ordering and failure behavior.
func TestLoadingModuleStartsNamedEngineDependencies(t *testing.T) {
	coordinator := loading.New(
		map[string]loading.Task{"world": func(context.Context) error { return nil }},
	)
	defer coordinator.Close()

	runtime := New()
	if err := runtime.RegisterModule(LoadingModule(coordinator)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(
			`local loading = require("engine.loading/v1"); loading.begin({"world"})`,
		)
	}); err != nil {
		t.Fatal(err)
	}
}
