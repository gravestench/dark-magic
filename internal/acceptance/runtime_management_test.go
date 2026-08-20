package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestScriptRuntimeManagementAndTransactionalReloadDoNotLeak protects node ownership across restart and reload.
func TestScriptRuntimeManagementAndTransactionalReloadDoNotLeak(t *testing.T) {
	ctx := context.Background()

	var composer render.Composer

	runtime := modruntime.New()
	if err := runtime.RegisterModule(modruntime.RenderModule(runtime, &composer)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := runtime.Stop(ctx); err != nil {
			t.Errorf("stop managed-script runtime: %v", err)
		}
	}()

	const worldComponent = `local r=require("engine.render/v1"); return {id="dynamic",` +
		`start=function(self) self.node=r.create("world") end}`

	source := fstest.MapFS{"component.lua": &fstest.MapFile{Data: []byte(worldComponent)}}

	definition, err := modruntime.LoadDefinition(ctx, runtime, source, "component.lua")
	if err != nil {
		t.Fatal(err)
	}

	manager := host.NewManager()
	if err := manager.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}

	if err := manager.Enable(ctx, "dynamic"); err != nil {
		t.Fatal(err)
	}

	assertNodes(t, &composer, 1)

	if err := manager.Restart(ctx, "dynamic"); err != nil {
		t.Fatal(err)
	}

	assertNodes(t, &composer, 1)

	source["component.lua"] = &fstest.MapFile{Data: []byte(`return { id = `)}
	if err := modruntime.ReloadDefinition(ctx, manager, runtime, source, "component.lua"); err == nil {
		t.Fatal("expected failed reload")
	}

	assertNodes(t, &composer, 1)

	status, _ := manager.Status("dynamic")
	if status.State != host.StateEnabled {
		t.Fatalf("status = %#v", status)
	}

	const modalComponent = `local r=require("engine.render/v1"); return {id="dynamic",` +
		`start=function(self) self.node=r.create("modal") end}`

	source["component.lua"] = &fstest.MapFile{Data: []byte(modalComponent)}
	if err := modruntime.ReloadDefinition(ctx, manager, runtime, source, "component.lua"); err != nil {
		t.Fatal(err)
	}

	nodes := composer.Snapshot()
	if len(nodes) != 1 || nodes[0].Layer != render.LayerModal {
		t.Fatalf("nodes after reload = %#v", nodes)
	}

	if err := manager.Disable(ctx, "dynamic"); err != nil {
		t.Fatal(err)
	}

	assertNodes(t, &composer, 0)

	if err := manager.Enable(ctx, "dynamic"); err != nil {
		t.Fatal(err)
	}

	assertNodes(t, &composer, 1)

	if err := manager.Disable(ctx, "dynamic"); err != nil {
		t.Fatal(err)
	}

	assertNodes(t, &composer, 0)
}
