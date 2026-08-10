package session

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestSkillSourceAppliesAuthoritativeAssignments(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	assignments, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.player.skill_assignment", Fields: []akara.Field{{Name: "left", Kind: akara.FieldInt64}, {Name: "right", Kind: akara.FieldInt64}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := controls.Set(entity, map[string]any{"player": "alpha"}); err != nil {
		t.Fatal(err)
	}
	assignment, err := assignments.Set(entity, map[string]any{"left": int64(0), "right": int64(0)})
	if err != nil {
		t.Fatal(err)
	}
	session, err := New(engine, Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterSkillAssignments(session); err != nil {
		t.Fatal(err)
	}
	controller := &MovementController{}
	if err := controller.AssignSkill("left", 42); err != nil {
		t.Fatal(err)
	}
	source, err := NewSkillSource(controller, "alpha")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(1)
	if len(commands) != 1 || commands[0].Kind != AssignSkillsCommand || commands[0].Sequence != 1 {
		t.Fatalf("commands = %#v", commands)
	}
	if err := session.Submit(commands[0]); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if left, _ := assignment.Get("left"); left != int64(42) {
		t.Fatalf("left skill = %v, want 42", left)
	}
	if right, _ := assignment.Get("right"); right != int64(0) {
		t.Fatalf("right skill changed = %v", right)
	}
	if commands := source.Commands(2); commands != nil {
		t.Fatalf("drained request emitted again: %#v", commands)
	}
}

func TestSkillControllerRejectsInvalidAssignment(t *testing.T) {
	controller := &MovementController{}
	if err := controller.AssignSkill("middle", 1); err == nil {
		t.Fatal("invalid slot was accepted")
	}
	if err := controller.AssignSkill("left", -1); err == nil {
		t.Fatal("negative skill was accepted")
	}
}
