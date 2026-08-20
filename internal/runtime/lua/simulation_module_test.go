package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/presentation/scene"
)

// TestSimulationModuleMovesPersistentWorld protects the simulation module moves persistent world contract,
// including its observable ordering and failure behavior.
func TestSimulationModuleMovesPersistentWorld(t *testing.T) {
	world := scene.New(7, 100, 80)

	runtime := New()
	if err := runtime.RegisterModule(SimulationModule(NewSimulation(world))); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	if err := runtime.Execute(
		context.Background(),
		fstest.MapFS{
			"test.lua": &fstest.MapFile{
				Data: []byte(
					`local s=require("engine.simulation/v1"); s.move_hero(10,-5); x=s.state().hero_x`,
				),
			},
		},
		"test.lua",
	); err != nil {
		t.Fatal(err)
	}

	if world.Hero.X != 60 || world.Hero.Y != 35 {
		t.Fatalf("world = %#v", world)
	}
}
