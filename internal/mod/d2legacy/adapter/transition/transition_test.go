package transition_test

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

func TestSourceTransitionsPlayerAcrossVerifiedSeamWithoutBounce(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	seam := transition.Seam{
		Town:       transition.SeamEndpoint{LevelID: 1, X: 10, Y: 10, ArrivalX: 5, ArrivalY: 5, Width: 100, Height: 80, Direction: "east"},
		Wilderness: transition.SeamEndpoint{LevelID: 2, X: 0, Y: 40, ArrivalX: 6, ArrivalY: 40, Width: 400, Height: 400, Direction: "west"},
	}
	authority, err := transition.NewAuthority(seam)
	if err != nil {
		t.Fatal(err)
	}
	identity, _ := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	position, _ := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	location, _ := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.location", Version: 1, Fields: []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}})
	entity, _ := engine.World().CreateEntity()
	identity.Set(entity, map[string]any{"player": "player"})
	position.Set(entity, map[string]any{"x": float64(10), "y": float64(10)})
	location.Set(entity, map[string]any{"act": int64(1), "level_id": int64(1)})
	source, err := transition.NewSource(engine, "player", authority)
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(3)
	if len(commands) != 1 {
		t.Fatalf("commands = %#v", commands)
	}
	var payload transition.Payload
	if err := json.Unmarshal(commands[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.SourceLevel != 1 || payload.DestinationLevel != 2 || payload.ArrivalX != 6 || payload.ArrivalY != 40 || payload.WorldWidth != 400 {
		t.Fatalf("transition recipe = %#v", payload)
	}
}
