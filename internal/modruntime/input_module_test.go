package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/inputcore"
	lua "github.com/yuin/gopher-lua"
)

func TestInputModuleReadsLogicalFrameSnapshot(t *testing.T) {
	t.Parallel()

	var input inputcore.Store
	input.Publish(inputcore.Frame{Actions: map[string]inputcore.ActionState{"confirm": {Pressed: true}}, CursorX: 12, CursorY: 34})
	runtime := New()
	if err := runtime.RegisterModule(InputModule(&input)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local input = require("dm.input/v1")
confirmed = input.pressed("confirm")
cursor_x, cursor_y = input.cursor()
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("confirmed") != lua.LTrue || state.GetGlobal("cursor_x").String() != "12" || state.GetGlobal("cursor_y").String() != "34" {
			t.Fatalf("input globals = %s/%s/%s", state.GetGlobal("confirmed"), state.GetGlobal("cursor_x"), state.GetGlobal("cursor_y"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
