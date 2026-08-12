// Package missile owns authoritative projectile creation, motion, contact,
// snapshotted effects, and removal. Rendering only observes its ECS facts.
package missile

import (
	"fmt"
	"hash/fnv"
	"math"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

const (
	SpawnSystemID     = "missile.spawn_from_cast"
	MovementSystemID  = "missile.straight_movement"
	CollisionSystemID = "missile.single_hit_collision"

	Component      = "d2legacy.missile.instance"
	SpawnReceipt   = "d2legacy.missile.spawn_receipt"
	EventComponent = "d2legacy.missile.event"
	EventSpawned   = "missile_spawned"
	EventHit       = "missile_hit"
	EventExpired   = "missile_expired"

	CollisionSingleHit = "single_hit"
)

// Definition is the verified, immutable behavior snapshot source for one skill.
type Definition struct {
	SkillID         int64
	SpeedPerTick    float64
	MaxRange        float64
	LifetimeTicks   uint64
	CollisionRadius float64
	PhysicalDamage  gamecombat.Amount
	DamageChannel   gamecombat.Channel
	MinimumDamage   gamecombat.Amount
	MaximumDamage   gamecombat.Amount
	// Presentation is an immutable copy of the deliberately joined Missiles.txt
	// recipe. Simulation stores these plain facts on the missile entity but never
	// interprets asset paths, animation rates, or sound keys.
	Presentation Presentation
}

type Presentation struct {
	MissileID, DCC, Palette, TravelSound, HitSound string
	Directions, FramesPerSecond                    int64
	Loop                                           bool
	OffsetX, OffsetY, OffsetZ                      float64
}

// Registry rejects projectile definitions that cannot simulate deterministically.
type Registry struct{ definitions map[int64]Definition }

func NewRegistry(definitions ...Definition) (Registry, error) {
	registry := Registry{definitions: make(map[int64]Definition, len(definitions))}
	for _, definition := range definitions {
		definition = normalizeDamageDefinition(definition)
		if definition.SkillID < 0 || !finitePositive(definition.SpeedPerTick) || !finitePositive(definition.MaxRange) || definition.LifetimeTicks == 0 || !finiteNonNegative(definition.CollisionRadius) || definition.MinimumDamage < 0 || definition.MaximumDamage < definition.MinimumDamage {
			return Registry{}, fmt.Errorf("missile: invalid definition for skill %d", definition.SkillID)
		}
		if _, exists := registry.definitions[definition.SkillID]; exists {
			return Registry{}, fmt.Errorf("missile: duplicate definition for skill %d", definition.SkillID)
		}
		registry.definitions[definition.SkillID] = definition
	}
	return registry, nil
}

// Register installs creation, movement, and single-hit collision in their
// authoritative simulation phases.
func Register(engine *gameecs.Engine, registry Registry) error {
	if engine == nil || registry.definitions == nil {
		return fmt.Errorf("missile: engine and registry are required")
	}
	stores, err := registerStores(engine)
	if err != nil {
		return err
	}
	if err := registerSpawn(engine, registry, stores); err != nil {
		return err
	}
	if err := registerMovement(engine, stores); err != nil {
		return err
	}
	return registerCollision(engine, stores)
}

type stores struct {
	casts, receipts, missiles, events, positions, locations, controls, selectables, colliders *akara.DynamicStore
}

func registerStores(engine *gameecs.Engine) (stores, error) {
	schemas := []akara.Schema{
		{Name: gameskill.CastEventComponent, Version: 1, Fields: []akara.Field{{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64}, {Name: "player", Kind: akara.FieldString}, {Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64}, {Name: "behavior", Kind: akara.FieldString}, {Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString}, {Name: "reason", Kind: akara.FieldString}}},
		{Name: SpawnReceipt, Version: 1, Fields: []akara.Field{{Name: "processed", Kind: akara.FieldBool}}},
		{Name: Component, Version: 3, Fields: []akara.Field{{Name: "owner_id", Kind: akara.FieldString}, {Name: "owner_entity", Kind: akara.FieldEntity}, {Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64}, {Name: "created_tick", Kind: akara.FieldInt64}, {Name: "expires_tick", Kind: akara.FieldInt64}, {Name: "velocity_x", Kind: akara.FieldFloat64}, {Name: "velocity_y", Kind: akara.FieldFloat64}, {Name: "previous_x", Kind: akara.FieldFloat64}, {Name: "previous_y", Kind: akara.FieldFloat64}, {Name: "traveled", Kind: akara.FieldFloat64}, {Name: "max_range", Kind: akara.FieldFloat64}, {Name: "collision_policy", Kind: akara.FieldString}, {Name: "collision_radius", Kind: akara.FieldFloat64}, {Name: "damage_channel", Kind: akara.FieldString}, {Name: "damage", Kind: akara.FieldInt64}, {Name: "physical", Kind: akara.FieldInt64}, {Name: "hit_target_id", Kind: akara.FieldString}, {Name: "announced", Kind: akara.FieldBool}, {Name: "missile_id", Kind: akara.FieldString}, {Name: "dcc", Kind: akara.FieldString}, {Name: "palette", Kind: akara.FieldString}, {Name: "travel_sound", Kind: akara.FieldString}, {Name: "hit_sound", Kind: akara.FieldString}, {Name: "directions", Kind: akara.FieldInt64}, {Name: "frames_per_second", Kind: akara.FieldInt64}, {Name: "loop", Kind: akara.FieldBool}, {Name: "offset_x", Kind: akara.FieldFloat64}, {Name: "offset_y", Kind: akara.FieldFloat64}, {Name: "offset_z", Kind: akara.FieldFloat64}}},
		{Name: EventComponent, Version: 3, Fields: []akara.Field{{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64}, {Name: "missile", Kind: akara.FieldEntity}, {Name: "owner_id", Kind: akara.FieldString}, {Name: "target_id", Kind: akara.FieldString}, {Name: "damage_channel", Kind: akara.FieldString}, {Name: "damage", Kind: akara.FieldInt64}, {Name: "physical", Kind: akara.FieldInt64}, {Name: "sound", Kind: akara.FieldString}}},
		{Name: "d2legacy.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "d2legacy.world.location", Version: 1, Fields: []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}},
		{Name: "d2legacy.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}},
		targeting.Schema(),
		{Name: "d2legacy.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}},
	}
	registered := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		store, err := akara.RegisterSchema(engine.World(), schema)
		if err != nil {
			return stores{}, err
		}
		registered[index] = store
	}
	return stores{registered[0], registered[1], registered[2], registered[3], registered[4], registered[5], registered[6], registered[7], registered[8]}, nil
}

func registerSpawn(engine *gameecs.Engine, registry Registry, s stores) error {
	return engine.Register(gameecs.Definition{ID: SpawnSystemID, Phase: gameecs.PhasePreSimulate, After: []string{gameskill.CastLifecycleSystemID}, All: []akara.ComponentType{s.casts}, None: []akara.ComponentType{s.receipts}, Read: []akara.ComponentType{s.casts, s.positions, s.locations, s.controls}, Write: []akara.ComponentType{s.receipts, s.missiles, s.events, s.positions, s.locations}, Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
		for _, eventEntity := range entities {
			event, _ := s.casts.Get(eventEntity)
			kind, _ := event.Get("kind")
			behavior, _ := event.Get("behavior")
			if kind != gameskill.EventSkillEffect || behavior != gameskill.BehaviorStraightMissile {
				continue
			}
			skillID, _ := event.Get("skill_id")
			definition, found := registry.definitions[skillID.(int64)]
			if !found {
				return fmt.Errorf("missile: no definition for skill %d", skillID.(int64))
			}
			player, _ := event.Get("player")
			caster, found := findPlayer(player.(string), s.controls)
			if !found {
				return fmt.Errorf("missile: caster %q does not exist", player.(string))
			}
			position, pok := s.positions.Get(caster)
			location, lok := s.locations.Get(caster)
			if !pok || !lok {
				return fmt.Errorf("missile: caster lacks position or location")
			}
			x, _ := position.Get("x")
			y, _ := position.Get("y")
			targetX, _ := event.Get("target_x")
			targetY, _ := event.Get("target_y")
			dx, dy := targetX.(float64)-x.(float64), targetY.(float64)-y.(float64)
			length := math.Hypot(dx, dy)
			if length == 0 {
				return fmt.Errorf("missile: target must differ from origin")
			}
			act, _ := location.Get("act")
			level, _ := location.Get("level_id")
			skillLevel, _ := event.Get("skill_level")
			presentation := definition.Presentation
			ownerID := combatOwnerID(caster, player.(string), s.selectables)
			damage := rollDefinitionDamage(definition, context.Tick, ownerID)
			physical := int64(0)
			if definition.DamageChannel == gamecombat.Physical {
				physical = damage.Raw()
			}
			commands.AddDynamic(s.receipts, eventEntity, map[string]any{"processed": true})
			commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{
				s.positions: {"x": x, "y": y},
				s.locations: {"act": act, "level_id": level},
				s.missiles:  {"owner_id": ownerID, "owner_entity": caster, "skill_id": skillID, "skill_level": skillLevel, "created_tick": int64(context.Tick), "expires_tick": int64(context.Tick + definition.LifetimeTicks), "velocity_x": dx / length * definition.SpeedPerTick, "velocity_y": dy / length * definition.SpeedPerTick, "previous_x": x, "previous_y": y, "traveled": 0.0, "max_range": definition.MaxRange, "collision_policy": CollisionSingleHit, "collision_radius": definition.CollisionRadius, "damage_channel": definition.DamageChannel.String(), "damage": damage.Raw(), "physical": physical, "hit_target_id": "", "announced": false, "missile_id": presentation.MissileID, "dcc": presentation.DCC, "palette": presentation.Palette, "travel_sound": presentation.TravelSound, "hit_sound": presentation.HitSound, "directions": presentation.Directions, "frames_per_second": presentation.FramesPerSecond, "loop": presentation.Loop, "offset_x": presentation.OffsetX, "offset_y": presentation.OffsetY, "offset_z": presentation.OffsetZ},
			})
		}
		return nil
	}})
}

func combatOwnerID(caster akara.Entity, player string, selectables *akara.DynamicStore) string {
	if selectable, found := selectables.Get(caster); found {
		if value, err := selectable.Get("id"); err == nil && value.(string) != "" {
			return value.(string)
		}
	}
	return "player:" + player
}

func registerMovement(engine *gameecs.Engine, s stores) error {
	return engine.Register(gameecs.Definition{ID: MovementSystemID, Phase: gameecs.PhaseMovement, All: []akara.ComponentType{s.missiles, s.positions}, Read: []akara.ComponentType{s.missiles, s.positions}, Write: []akara.ComponentType{s.missiles, s.positions, s.events}, Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
		for _, entity := range entities {
			missile, _ := s.missiles.Get(entity)
			expires, _ := missile.Get("expires_tick")
			if int64(context.Tick) >= expires.(int64) {
				continue
			}
			position, _ := s.positions.Get(entity)
			x, _ := position.Get("x")
			y, _ := position.Get("y")
			vx, _ := missile.Get("velocity_x")
			vy, _ := missile.Get("velocity_y")
			traveled, _ := missile.Get("traveled")
			maxRange, _ := missile.Get("max_range")
			announced, _ := missile.Get("announced")
			if !announced.(bool) {
				owner, _ := missile.Get("owner_id")
				damage, _ := missile.Get("damage")
				channel, _ := missile.Get("damage_channel")
				travelSound, _ := missile.Get("travel_sound")
				commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{s.events: eventValues(EventSpawned, context.Tick, entity, owner.(string), "", channel.(string), damage.(int64), travelSound.(string))})
				if err := missile.Set("announced", true); err != nil {
					return err
				}
			}
			if err := missile.Set("previous_x", x); err != nil {
				return err
			}
			if err := missile.Set("previous_y", y); err != nil {
				return err
			}
			fullStep := math.Hypot(vx.(float64), vy.(float64))
			step := min(fullStep, max(0, maxRange.(float64)-traveled.(float64)))
			scale := step / fullStep
			if err := position.Set("x", x.(float64)+vx.(float64)*scale); err != nil {
				return err
			}
			if err := position.Set("y", y.(float64)+vy.(float64)*scale); err != nil {
				return err
			}
			if err := missile.Set("traveled", traveled.(float64)+step); err != nil {
				return err
			}
		}
		return nil
	}})
}

func registerCollision(engine *gameecs.Engine, s stores) error {
	return engine.Register(gameecs.Definition{ID: CollisionSystemID, Phase: gameecs.PhaseCombat, All: []akara.ComponentType{s.missiles, s.positions, s.locations}, Read: []akara.ComponentType{s.missiles, s.positions, s.locations, s.selectables, s.colliders}, Write: []akara.ComponentType{s.missiles, s.events}, Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
		for _, entity := range entities {
			missile, _ := s.missiles.Get(entity)
			expires, _ := missile.Get("expires_tick")
			traveled, _ := missile.Get("traveled")
			maxRange, _ := missile.Get("max_range")
			owner, _ := missile.Get("owner_id")
			hitSound, _ := missile.Get("hit_sound")
			if int64(context.Tick) >= expires.(int64) {
				commands.Destroy(engine.World(), entity)
				commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{s.events: eventValues(EventExpired, context.Tick, entity, owner.(string), "", "", 0, "")})
				continue
			}
			target, targetID := firstContact(engine, entity, missile, s)
			if targetID == "" {
				if traveled.(float64) >= maxRange.(float64) {
					commands.Destroy(engine.World(), entity)
					commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{s.events: eventValues(EventExpired, context.Tick, entity, owner.(string), "", "", 0, "")})
				}
				continue
			}
			channelName, _ := missile.Get("damage_channel")
			channel, err := gamecombat.ParseChannel(channelName.(string))
			if err != nil {
				return err
			}
			damage, _ := missile.Get("damage")
			remaining, died, err := gamecombat.ApplyConfirmedDamage(engine, commands, context.Tick, owner.(string), targetID, target, channel, gamecombat.FromRaw(damage.(int64)))
			if err != nil {
				return err
			}
			_ = remaining
			_ = died
			if err := missile.Set("hit_target_id", targetID); err != nil {
				return err
			}
			commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{s.events: eventValues(EventHit, context.Tick, entity, owner.(string), targetID, channelName.(string), damage.(int64), hitSound.(string))})
			commands.Destroy(engine.World(), entity)
		}
		return nil
	}})
}

func firstContact(engine *gameecs.Engine, missileEntity akara.Entity, missile *akara.DynamicComponent, s stores) (akara.Entity, string) {
	owner, _ := missile.Get("owner_entity")
	px, _ := missile.Get("previous_x")
	py, _ := missile.Get("previous_y")
	position, _ := s.positions.Get(missileEntity)
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	radius, _ := missile.Get("collision_radius")
	location, _ := s.locations.Get(missileEntity)
	act, _ := location.Get("act")
	level, _ := location.Get("level_id")
	bestDistance := math.Inf(1)
	var best akara.Entity
	bestID := ""
	for _, candidate := range s.selectables.Entities() {
		if candidate == owner.(akara.Entity) || !hasHealth(engine, candidate) {
			continue
		}
		selectable, _ := s.selectables.Get(candidate)
		id, _ := selectable.Get("id")
		candidatePosition, pok := s.positions.Get(candidate)
		candidateLocation, lok := s.locations.Get(candidate)
		if !pok || !lok {
			continue
		}
		candidateAct, _ := candidateLocation.Get("act")
		candidateLevel, _ := candidateLocation.Get("level_id")
		if candidateAct != act || candidateLevel != level {
			continue
		}
		cx, _ := candidatePosition.Get("x")
		cy, _ := candidatePosition.Get("y")
		candidateRadius := 0.0
		if collider, ok := s.colliders.Get(candidate); ok {
			value, _ := collider.Get("radius")
			candidateRadius = value.(float64)
		}
		distance, along := segmentDistance(px.(float64), py.(float64), x.(float64), y.(float64), cx.(float64), cy.(float64))
		if distance <= radius.(float64)+candidateRadius && along < bestDistance {
			bestDistance, best, bestID = along, candidate, id.(string)
		}
	}
	return best, bestID
}

func hasHealth(engine *gameecs.Engine, entity akara.Entity) bool {
	if monsters, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats"); found && monsters.Has(entity) {
		return true
	}
	if players, found := akara.GetDynamicStore(engine.World(), "d2legacy.player.vitals"); found && players.Has(entity) {
		return true
	}
	return false
}

func segmentDistance(ax, ay, bx, by, px, py float64) (float64, float64) {
	dx, dy := bx-ax, by-ay
	lengthSquared := dx*dx + dy*dy
	t := 0.0
	if lengthSquared > 0 {
		t = ((px-ax)*dx + (py-ay)*dy) / lengthSquared
		t = max(0, min(1, t))
	}
	cx, cy := ax+t*dx, ay+t*dy
	return math.Hypot(px-cx, py-cy), t
}

func findPlayer(player string, controls *akara.DynamicStore) (akara.Entity, bool) {
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		value, _ := control.Get("player")
		if value == player {
			return entity, true
		}
	}
	return 0, false
}
func finitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
func finiteNonNegative(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
func eventValues(kind string, tick uint64, missile akara.Entity, owner, target, channel string, damage int64, sound string) map[string]any {
	physical := int64(0)
	if channel == gamecombat.Physical.String() {
		physical = damage
	}
	return map[string]any{"kind": kind, "tick": int64(tick), "missile": missile, "owner_id": owner, "target_id": target, "damage_channel": channel, "damage": damage, "physical": physical, "sound": sound}
}

func normalizeDamageDefinition(definition Definition) Definition {
	if definition.MinimumDamage == 0 && definition.MaximumDamage == 0 && definition.PhysicalDamage != 0 {
		definition.DamageChannel = gamecombat.Physical
		definition.MinimumDamage, definition.MaximumDamage = definition.PhysicalDamage, definition.PhysicalDamage
	}
	return definition
}

func rollDefinitionDamage(definition Definition, tick uint64, owner string) gamecombat.Amount {
	span := int64(definition.MaximumDamage-definition.MinimumDamage) + 1
	if span <= 1 {
		return definition.MinimumDamage
	}
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(owner))
	_, _ = hash.Write([]byte{byte(definition.SkillID), byte(definition.SkillID >> 8), byte(tick), byte(tick >> 8), byte(tick >> 16), byte(tick >> 24)})
	return definition.MinimumDamage + gamecombat.Amount(hash.Sum64()%uint64(span))
}
