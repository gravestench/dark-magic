package skill

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestIntentConsumerSnapshotsAdmittedSkillAndLearnedLevelOnce(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterIntentConsumer(engine); err != nil {
		t.Fatal(err)
	}
	intents, controls, learned, requests, err := registerIntentStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	owner := engine.World().MustCreateEntity()
	mustSet(t, controls, owner, map[string]any{"player": "alpha"})
	mustSet(t, intents, owner, map[string]any{"side": "left", "skill_id": int64(42), "target_x": 12.5, "target_y": 9.25, "target_id": "monster:fallen"})
	learnedEntity := engine.World().MustCreateEntity()
	mustSet(t, learned, learnedEntity, map[string]any{"owner": owner, "skill_id": int64(42), "level": int64(3), "list_row": int64(1), "left_allowed": true, "right_allowed": true})
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	request, present := requests.Get(owner)
	if !present {
		t.Fatal("cast request was not created")
	}
	skillID, _ := request.Get("skill_id")
	level, _ := request.Get("skill_level")
	tick, _ := request.Get("request_tick")
	targetID, _ := request.Get("target_id")
	if skillID != int64(42) || level != int64(3) || tick != int64(1) || targetID != "monster:fallen" {
		t.Fatalf("cast request = skill %v level %v tick %v target %v", skillID, level, tick, targetID)
	}
	intent, _ := intents.Get(owner)
	side, _ := intent.Get("side")
	intentSkill, _ := intent.Get("skill_id")
	if side != "" || intentSkill != int64(0) {
		t.Fatalf("intent was not cleared: side=%v skill=%v", side, intentSkill)
	}
	learnedComponent, _ := learned.Get(learnedEntity)
	_ = learnedComponent.Set("level", int64(9))
	skillID, _ = request.Get("skill_id")
	level, _ = request.Get("skill_level")
	if skillID != int64(42) || level != int64(3) {
		t.Fatalf("pending request changed to skill %v level %v", skillID, level)
	}
}

func TestIntentConsumerRejectsUnlearnedAssignment(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterIntentConsumer(engine); err != nil {
		t.Fatal(err)
	}
	intents, controls, _, _, _ := registerIntentStores(engine)
	owner := engine.World().MustCreateEntity()
	mustSet(t, controls, owner, map[string]any{"player": "alpha"})
	mustSet(t, intents, owner, map[string]any{"side": "left", "skill_id": int64(42), "target_x": 0.0, "target_y": 0.0, "target_id": ""})
	if err := engine.Update(time.Second / 25); err == nil {
		t.Fatal("unlearned assignment was consumed")
	}
}

func mustSet(t *testing.T, store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	t.Helper()
	if _, err := store.Set(entity, values); err != nil {
		t.Fatal(err)
	}
}
