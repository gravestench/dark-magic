package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/modruntime"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

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
	defer runtime.Stop(ctx)
	source := fstest.MapFS{"component.lua": &fstest.MapFile{Data: []byte(`local r=require("dm.render/v1"); return {id="dynamic",start=function(self) self.node=r.create("world") end}`)}}
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

	source["component.lua"] = &fstest.MapFile{Data: []byte(`local r=require("dm.render/v1"); return {id="dynamic",start=function(self) self.node=r.create("modal") end}`)}
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
