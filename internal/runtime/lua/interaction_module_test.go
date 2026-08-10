package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	gameinteraction "github.com/gravestench/dark-magic/internal/game/interaction"
)

func TestInteractionModuleExposesCopiesAndQueuesIntent(t *testing.T) {
	authority, err := gameinteraction.NewAuthority(gameinteraction.Target{ID: "act1-akara", NPC: "Akara", Vendor: "Akara", Categories: []string{"misc", "armo"}, Radius: 5})
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.RegisterOwner("alice", "act1-akara"); err != nil {
		t.Fatal(err)
	}
	controller := &gameinteraction.Controller{}
	runtime := New()
	if err := runtime.RegisterModule(InteractionModule(authority, controller, "alice")); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`local i=require("dm.interaction/v1"); local s=i.snapshot(); assert(s.active and s.npc=="Akara" and s.categories[1]=="armo"); i.close(); i.open("act1-akara")`)}}
	if err := runtime.Execute(ctx, script, "test.lua"); err != nil {
		t.Fatal(err)
	}
	source, err := gameinteraction.NewSource(controller, "alice")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(4)
	if len(commands) != 2 || commands[0].Kind != gameinteraction.CloseCommand || commands[1].Kind != gameinteraction.OpenCommand {
		t.Fatalf("unexpected queued commands: %#v", commands)
	}
}
