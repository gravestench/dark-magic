package modruntime

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/loadcore"
	lua "github.com/yuin/gopher-lua"
)

func TestLoadingModuleStartsNamedEngineDependencies(t *testing.T) {
	coordinator := loadcore.New(map[string]loadcore.Task{"world": func(context.Context) error { return nil }})
	defer coordinator.Close()
	runtime := New()
	if err := runtime.RegisterModule(LoadingModule(coordinator)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`local loading = require("dm.loading/v1"); loading.begin({"world"})`)
	}); err != nil {
		t.Fatal(err)
	}
}
