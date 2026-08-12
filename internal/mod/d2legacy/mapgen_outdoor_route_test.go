package d2legacy_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	. "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func TestD2LegacyLuaOwnsBloodMoorEdgeAndRoutePolicy(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(DeterministicModule()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := fstest.MapFS{"test.lua": {Data: []byte(`
local route=require("d2legacy.mapgen.outdoor_route")
for _,direction in ipairs({"north","east","south","west"}) do
  local plan=route.plan(42,80,80,direction)
  assert(#plan.ordered==10 and #plan.path>10)
  assert(plan.entry.direction==route.opposite(direction))
  assert(plan.exit.direction==direction)
  assert(plan.path[1] and plan.cells)
  for index=2,#plan.ordered do
    local a,b=plan.ordered[index-1],plan.ordered[index]
    assert(math.abs(a.x-b.x)<=1 and math.abs(a.y-b.y)<=1)
  end
end
local east=route.plan(42,80,80,"east")
assert(east.entry.x==0 and east.entry.y==40 and east.exit.x==79 and east.exit.y==40)
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
