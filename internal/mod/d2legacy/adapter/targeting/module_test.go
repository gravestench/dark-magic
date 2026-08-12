package targeting

import (
	"context"
	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"testing"
	"testing/fstest"
)

func TestTargetingModuleReturnsCopiedSpawnedFacts(t *testing.T) {
	engine := gameecs.New()
	positions, _ := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.position", Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	selectables, _ := akara.RegisterSchema(engine.World(), Schema())
	entity := engine.World().MustCreateEntity()
	_, _ = positions.Set(entity, map[string]any{"x": 4.0, "y": 5.0})
	_, _ = selectables.Set(entity, map[string]any{"id": "fallen:1", "kind": KindHostile, "label": "Fallen", "owner": "", "radius": 1.0, "priority": int64(2)})
	runtime := modruntime.New()
	if err := runtime.RegisterModule(Module(New(engine))); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	script := fstest.MapFS{"test.lua": {Data: []byte(`local t=require("d2legacy.targeting/v1").selectable_at(4,5);assert(t and t.id=="fallen:1" and t.kind=="hostile" and t.label=="Fallen")`)}}
	if err := runtime.Execute(ctx, script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
