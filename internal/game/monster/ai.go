package monster

import (
	"fmt"
	"math"
	"sort"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/targeting"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const (
	AISystemID           = "monster.basic_ai"
	AIComponent          = "dm.monster.ai"
	BasicAttackComponent = "dm.monster.basic_attack_request"
	AIIdle               = "idle"
	AIChase              = "chase"
	AIAttack             = "attack"
)

// PathFinder keeps tactics dependent on authoritative collision without making
// the AI system own a map. The existing world.Map satisfies this interface.
type PathFinder interface {
	FindPath(gameworld.PathRequest) ([]gameworld.Point, error)
}

func aiSchema() akara.Schema {
	return akara.Schema{Name: AIComponent, Version: 1, Fields: []akara.Field{
		{Name: "behavior", Kind: akara.FieldString}, {Name: "state", Kind: akara.FieldString},
		{Name: "target_id", Kind: akara.FieldString}, {Name: "next_think_tick", Kind: akara.FieldInt64},
		{Name: "think_interval", Kind: akara.FieldInt64}, {Name: "aggro_radius", Kind: akara.FieldFloat64},
		{Name: "attack_range", Kind: akara.FieldFloat64}, {Name: "speed", Kind: akara.FieldFloat64},
	}}
}

func attackSchema() akara.Schema {
	return akara.Schema{Name: BasicAttackComponent, Version: 1, Fields: []akara.Field{
		{Name: "target_id", Kind: akara.FieldString}, {Name: "request_tick", Kind: akara.FieldInt64}, {Name: "range", Kind: akara.FieldFloat64},
	}}
}

// RegisterAI installs a fixed-tick intent chooser. Target memory and its next
// deadline live in ECS, so snapshots cover everything that affects its future.
func RegisterAI(engine *gameecs.Engine, paths PathFinder) error {
	if engine == nil {
		return fmt.Errorf("monster: AI requires an engine")
	}
	ai, attacks, positions, velocities, locations, selectables, err := registerAIStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: AISystemID, Phase: gameecs.PhaseIntent,
		All:   []akara.ComponentType{ai, positions, velocities, locations},
		Read:  []akara.ComponentType{positions, locations, selectables},
		Write: []akara.ComponentType{ai, velocities, attacks},
		Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return updateAI(context, entities, commands, paths, ai, attacks, positions, velocities, locations, selectables)
		},
	})
}

func registerAIStores(engine *gameecs.Engine) (ai, attacks, positions, velocities, locations, selectables *akara.DynamicStore, err error) {
	schemas := []akara.Schema{
		aiSchema(), attackSchema(),
		{Name: "dm.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.location", Version: 1, Fields: []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}},
		targeting.Schema(),
	}
	stores := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		stores[index], err = akara.RegisterSchema(engine.World(), schema)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
	}
	return stores[0], stores[1], stores[2], stores[3], stores[4], stores[5], nil
}

type aiTarget struct {
	id         string
	x, y       float64
	act, level int64
}

func updateAI(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer, paths PathFinder, ai, attacks, positions, velocities, locations, selectables *akara.DynamicStore) error {
	targets := playerTargets(selectables, positions, locations)
	for _, entity := range entities {
		brain, _ := ai.Get(entity)
		next, _ := brain.Get("next_think_tick")
		if next.(int64) > int64(context.Tick) {
			continue
		}
		position, _ := positions.Get(entity)
		location, _ := locations.Get(entity)
		velocity, _ := velocities.Get(entity)
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		act, _ := location.Get("act")
		level, _ := location.Get("level_id")
		remembered, _ := brain.Get("target_id")
		aggro, _ := brain.Get("aggro_radius")
		target, found := chooseTarget(targets, remembered.(string), x.(float64), y.(float64), act.(int64), level.(int64), aggro.(float64))
		interval, _ := brain.Get("think_interval")
		if err := brain.Set("next_think_tick", int64(context.Tick)+interval.(int64)); err != nil {
			return err
		}
		if !found {
			if err := setBrain(brain, AIIdle, ""); err != nil {
				return err
			}
			if err := setVelocity(velocity, 0, 0); err != nil {
				return err
			}
			commands.Remove(attacks, entity)
			continue
		}
		if err := brain.Set("target_id", target.id); err != nil {
			return err
		}
		dx, dy := target.x-x.(float64), target.y-y.(float64)
		attackRange, _ := brain.Get("attack_range")
		if math.Hypot(dx, dy) <= attackRange.(float64) {
			if err := setBrain(brain, AIAttack, target.id); err != nil {
				return err
			}
			if err := setVelocity(velocity, 0, 0); err != nil {
				return err
			}
			commands.AddDynamic(attacks, entity, map[string]any{"target_id": target.id, "request_tick": int64(context.Tick), "range": attackRange.(float64)})
			continue
		}
		if err := setBrain(brain, AIChase, target.id); err != nil {
			return err
		}
		if paths != nil {
			path, err := paths.FindPath(gameworld.PathRequest{Start: gameworld.Point{X: x.(float64), Y: y.(float64)}, Goal: gameworld.Point{X: target.x, Y: target.y}, StopRadius: attackRange.(float64)})
			if err != nil || len(path) < 2 {
				if err := setVelocity(velocity, 0, 0); err != nil {
					return err
				}
				commands.Remove(attacks, entity)
				continue
			}
			dx, dy = path[1].X-x.(float64), path[1].Y-y.(float64)
		}
		speed, _ := brain.Get("speed")
		length := math.Hypot(dx, dy)
		if err := setVelocity(velocity, dx/length*speed.(float64), dy/length*speed.(float64)); err != nil {
			return err
		}
		commands.Remove(attacks, entity)
	}
	return nil
}

func setBrain(brain *akara.DynamicComponent, state, target string) error {
	if err := brain.Set("state", state); err != nil {
		return err
	}
	return brain.Set("target_id", target)
}

func setVelocity(velocity *akara.DynamicComponent, x, y float64) error {
	if err := velocity.Set("x", x); err != nil {
		return err
	}
	return velocity.Set("y", y)
}

func playerTargets(selectables, positions, locations *akara.DynamicStore) []aiTarget {
	result := make([]aiTarget, 0)
	for _, entity := range selectables.Entities() {
		selectable, present := selectables.Get(entity)
		if !present {
			continue
		}
		kind, _ := selectable.Get("kind")
		position, positionPresent := positions.Get(entity)
		location, locationPresent := locations.Get(entity)
		if kind != targeting.KindPlayer || !positionPresent || !locationPresent {
			continue
		}
		id, _ := selectable.Get("id")
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		act, _ := location.Get("act")
		level, _ := location.Get("level_id")
		result = append(result, aiTarget{id: id.(string), x: x.(float64), y: y.(float64), act: act.(int64), level: level.(int64)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].id < result[j].id })
	return result
}

func chooseTarget(targets []aiTarget, remembered string, x, y float64, act, level int64, radius float64) (aiTarget, bool) {
	eligible := func(target aiTarget) bool {
		return target.act == act && target.level == level && math.Hypot(target.x-x, target.y-y) <= radius
	}
	for _, target := range targets {
		if target.id == remembered && eligible(target) {
			return target, true
		}
	}
	best, found, distance := aiTarget{}, false, math.Inf(1)
	for _, target := range targets {
		if !eligible(target) {
			continue
		}
		current := math.Hypot(target.x-x, target.y-y)
		if !found || current < distance {
			best, found, distance = target, true, current
		}
	}
	return best, found
}
