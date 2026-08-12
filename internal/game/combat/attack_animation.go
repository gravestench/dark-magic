package combat

import (
	"fmt"

	"github.com/gravestench/akara"
	gameaction "github.com/gravestench/dark-magic/internal/game/action"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

const AttackAnimation = gameaction.AttackAnimationComponent

// AttackTiming is the small, simulation-safe subset normalized from
// AnimData.d2. Speed is authored in 1/256 frame per 25 Hz simulation tick.
type AttackTiming struct {
	Frames      int64
	Speed       int64
	ImpactFrame int64
}

// AttackTimingResolver isolates combat from archive lookup and presentation.
// Equipment-specific weapon classes can use the same boundary later.
type AttackTimingResolver interface {
	AttackTiming(token, weaponClass string) (AttackTiming, bool)
}

func (timing AttackTiming) valid() bool {
	return timing.Frames > 0 && timing.Speed > 0 && timing.ImpactFrame >= 0 && timing.ImpactFrame < timing.Frames
}

func attackAnimationSchema() akara.Schema {
	return akara.Schema{Name: AttackAnimation, Version: 1, Fields: []akara.Field{
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "target_id", Kind: akara.FieldString}, {Name: "start_tick", Kind: akara.FieldInt64},
		{Name: "frames", Kind: akara.FieldInt64}, {Name: "speed", Kind: akara.FieldInt64},
		{Name: "impact_frame", Kind: akara.FieldInt64}, {Name: "progress", Kind: akara.FieldInt64},
		{Name: "impact_fired", Kind: akara.FieldBool},
	}}
}

func newAttackAnimation(skillID int64, targetID string, tick uint64, timing AttackTiming) map[string]any {
	return map[string]any{
		"skill_id": skillID, "target_id": targetID, "start_tick": int64(tick), "frames": timing.Frames,
		"speed": timing.Speed, "impact_frame": timing.ImpactFrame,
		"progress": int64(0), "impact_fired": false,
	}
}

func resolveAttackTiming(entity akara.Entity, resolver AttackTimingResolver, appearances *akara.DynamicStore) (AttackTiming, error) {
	if resolver == nil {
		return AttackTiming{}, fmt.Errorf("combat: attack timing resolver is unavailable")
	}
	appearance, present := appearances.Get(entity)
	if !present {
		return AttackTiming{}, fmt.Errorf("combat: attacking player lacks appearance")
	}
	token, _ := appearance.Get("token")
	weaponClass, _ := appearance.Get("weapon_class")
	timing, found := resolver.AttackTiming(token.(string), weaponClass.(string))
	if !found || !timing.valid() {
		return AttackTiming{}, fmt.Errorf("combat: no valid A1 timing for %s/%s", token, weaponClass)
	}
	return timing, nil
}

func registerAttackAnimationSystem(engine *gameecs.Engine, requests, attacks, animations *akara.DynamicStore) error {
	return engine.Register(gameecs.Definition{
		ID: "combat.player_attack_animation", Phase: gameecs.PhasePreSimulate, After: []string{"combat.player_attack_approach"},
		All: []akara.ComponentType{attacks}, Read: []akara.ComponentType{attacks}, Write: []akara.ComponentType{attacks, requests, animations},
		Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return updateAttackAnimations(context, entities, commands, attacks, requests, animations)
		},
	})
}

func updateAttackAnimations(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer, attacks, requests, animations *akara.DynamicStore) error {
	for _, entity := range entities {
		attack, _ := attacks.Get(entity)
		progressValue, _ := attack.Get("progress")
		speedValue, _ := attack.Get("speed")
		progress := progressValue.(int64) + speedValue.(int64)
		if err := attack.Set("progress", progress); err != nil {
			return err
		}
		if err := emitAttackImpact(context, entity, attack, progress, commands, requests); err != nil {
			return err
		}
		framesValue, _ := attack.Get("frames")
		if progress < framesValue.(int64)*256 {
			continue
		}
		commands.Remove(attacks, entity)
		if animation, present := animations.Get(entity); present {
			if err := animation.Set("mode", "NU"); err != nil {
				return err
			}
		}
	}
	return nil
}

func emitAttackImpact(context gameecs.Context, entity akara.Entity, attack *akara.DynamicComponent, progress int64, commands *akara.CommandBuffer, requests *akara.DynamicStore) error {
	impactValue, _ := attack.Get("impact_frame")
	firedValue, _ := attack.Get("impact_fired")
	if firedValue.(bool) || progress/256 < impactValue.(int64) {
		return nil
	}
	targetID, _ := attack.Get("target_id")
	commands.AddDynamic(requests, entity, map[string]any{"target_id": targetID.(string), "request_tick": int64(context.Tick)})
	return attack.Set("impact_fired", true)
}
