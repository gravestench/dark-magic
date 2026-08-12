package session

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestSkillControllerRejectsInvalidAssignment(t *testing.T) {
	controller := &MovementController{}
	if err := controller.AssignSkill("middle", 1); err == nil {
		t.Fatal("invalid slot was accepted")
	}
	if err := controller.AssignSkill("left", -1); err == nil {
		t.Fatal("negative skill was accepted")
	}
}
