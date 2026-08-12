package acceptance

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestFiniteMapCameraCoversViewportAtBothEdges(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repositoryRoot(t), "internal/content/shim/lua/d2/presentation/camera_bounds.lua")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	state := lua.NewState()
	defer state.Close()
	chunk, err := state.LoadString(string(source))
	if err != nil {
		t.Fatal(err)
	}
	state.Push(chunk)
	if err := state.PCall(0, 1, nil); err != nil {
		t.Fatal(err)
	}
	module := state.CheckTable(-1)
	clamp, ok := module.RawGetString("clamp_center").(*lua.LFunction)
	if !ok {
		t.Fatal("camera bounds module has no clamp_center function")
	}

	check := func(name string, center, content, viewport, want float64) {
		t.Helper()
		state.Push(clamp)
		state.Push(lua.LNumber(center))
		state.Push(lua.LNumber(content))
		state.Push(lua.LNumber(viewport))
		if err := state.PCall(3, 1, nil); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got := float64(state.CheckNumber(-1))
		state.Pop(1)
		if got != want {
			t.Fatalf("%s: center = %v, want %v", name, got, want)
		}
	}

	check("left edge", 1100, 1600, 800, 800)
	check("right edge", -300, 1600, 800, 0)
	check("interior", 400, 1600, 800, 400)
	check("small canvas", 10, 600, 800, 400)

	anchor, ok := module.RawGetString("anchor_for_center").(*lua.LFunction)
	if !ok {
		t.Fatal("camera bounds module has no anchor_for_center function")
	}
	state.Push(anchor)
	state.Push(lua.LNumber(800))
	state.Push(lua.LNumber(1600))
	state.Push(lua.LNumber(250))
	if err := state.PCall(3, 1, nil); err != nil {
		t.Fatal(err)
	}
	if got := float64(state.CheckNumber(-1)); got != 250 {
		t.Fatalf("effective pointer-projection anchor = %v, want 250", got)
	}
}
