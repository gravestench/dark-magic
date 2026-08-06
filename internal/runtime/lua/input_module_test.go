package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/inputstate"
	lua "github.com/yuin/gopher-lua"
)

func TestInputModuleReadsLogicalFrameSnapshot(t *testing.T) {
	t.Parallel()

	var input inputstate.Store
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"confirm": {Pressed: true}}, Text: "hé", CursorX: 12, CursorY: 34})
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
entered = input.text()
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("confirmed") != lua.LTrue || state.GetGlobal("cursor_x").String() != "12" || state.GetGlobal("cursor_y").String() != "34" || state.GetGlobal("entered").String() != "hé" {
			t.Fatalf("input globals = %s/%s/%s/%s", state.GetGlobal("confirmed"), state.GetGlobal("cursor_x"), state.GetGlobal("cursor_y"), state.GetGlobal("entered"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
