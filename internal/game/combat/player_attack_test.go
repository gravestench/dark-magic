package combat

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
)

func TestRepeatedAttackClickPreservesSameTargetApproach(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	casts, receipts, controls, approaches, attacks, _, _, _, _, _, err := registerPlayerAttackStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	caster := engine.World().MustCreateEntity()
	if _, err := controls.Set(caster, map[string]any{"player": "hero"}); err != nil {
		t.Fatal(err)
	}
	if _, err := approaches.Set(caster, map[string]any{
		"target_id": "monster:fallen", "request_tick": int64(3),
		"goal_x": 12.0, "goal_y": 8.0, "waypoint_x": 10.0, "waypoint_y": 7.0, "has_waypoint": true,
	}); err != nil {
		t.Fatal(err)
	}
	event := engine.World().MustCreateEntity()
	if _, err := casts.Set(event, map[string]any{
		"kind": gameskill.EventSkillEffect, "tick": int64(9), "player": "hero",
		"skill_id": int64(0), "skill_level": int64(1), "behavior": gameskill.BehaviorBasicMelee,
		"target_x": 12.0, "target_y": 8.0, "target_id": "monster:fallen", "reason": "",
	}); err != nil {
		t.Fatal(err)
	}
	commands := akara.NewCommandBuffer()
	if err := translatePlayerAttacks([]akara.Entity{event}, commands, 0, nil, casts, receipts, controls, approaches, attacks, nil, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := commands.Apply(); err != nil {
		t.Fatal(err)
	}
	pending, _ := approaches.Get(caster)
	if tick, _ := pending.Get("request_tick"); tick != int64(3) {
		t.Fatalf("same-target click reset request tick to %v", tick)
	}
	if waypoint, _ := pending.Get("waypoint_x"); waypoint != 10.0 {
		t.Fatalf("same-target click reset cached waypoint to %v", waypoint)
	}
	if !receipts.Has(event) {
		t.Fatal("repeated click event was not acknowledged")
	}
}

func TestRepeatedAttackClickDoesNotRestartSameTargetSwing(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	casts, receipts, controls, approaches, attacks, _, _, _, _, _, err := registerPlayerAttackStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	caster := engine.World().MustCreateEntity()
	_, _ = controls.Set(caster, map[string]any{"player": "hero"})
	_, _ = attacks.Set(caster, map[string]any{
		"target_id": "monster:fallen", "start_tick": int64(4), "frames": int64(8), "speed": int64(256),
		"impact_frame": int64(4), "progress": int64(768), "impact_fired": false,
	})
	event := engine.World().MustCreateEntity()
	_, _ = casts.Set(event, map[string]any{
		"kind": gameskill.EventSkillEffect, "tick": int64(9), "player": "hero", "skill_id": int64(0),
		"skill_level": int64(1), "behavior": gameskill.BehaviorBasicMelee,
		"target_x": 12.0, "target_y": 8.0, "target_id": "monster:fallen", "reason": "",
	})
	commands := akara.NewCommandBuffer()
	if err := translatePlayerAttacks([]akara.Entity{event}, commands, 0, nil, casts, receipts, controls, approaches, attacks, nil, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := commands.Apply(); err != nil {
		t.Fatal(err)
	}
	active, _ := attacks.Get(caster)
	if progress, _ := active.Get("progress"); progress != int64(768) {
		t.Fatalf("same-target click reset swing progress to %v", progress)
	}
	if approaches.Has(caster) {
		t.Fatal("same-target click created an approach during the swing")
	}
}
