package action

import (
	"testing"

	"github.com/gravestench/akara"
)

func TestExclusiveMatchRequiresSameSkillAndTarget(t *testing.T) {
	world := akara.NewWorld()
	defer world.Close()
	approaches, err := akara.RegisterSchema(world, akara.Schema{Name: AttackApproachComponent, Fields: []akara.Field{
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "target_id", Kind: akara.FieldString},
	}})
	if err != nil {
		t.Fatal(err)
	}
	entity := world.MustCreateEntity()
	if _, err := approaches.Set(entity, map[string]any{"skill_id": int64(0), "target_id": "monster:fallen"}); err != nil {
		t.Fatal(err)
	}
	if !MatchesExclusive(world, entity, 0, "monster:fallen") {
		t.Fatal("identical Attack/target pair did not match active action")
	}
	if MatchesExclusive(world, entity, 47, "monster:fallen") {
		t.Fatal("different skill incorrectly preserved active Attack")
	}
	if MatchesExclusive(world, entity, 0, "monster:zombie") {
		t.Fatal("different target incorrectly preserved active Attack")
	}
}
