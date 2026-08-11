package monster

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

// SpawnCommand is privileged population/encounter output. Ordinary clients may
// fight a spawned monster but may not manufacture one.
const SpawnCommand = "system.monster.spawn"

// Spawn is the immutable definition snapshot plus world facts needed to create
// one ordinary hostile. No renderer handle or AI decision lives here.
type Spawn struct {
	SpawnID    string     `json:"spawn_id"`
	Definition Definition `json:"definition"`
	Seed       uint64     `json:"seed"`
	X          float64    `json:"x"`
	Y          float64    `json:"y"`
	Act        int64      `json:"act"`
	LevelID    int64      `json:"level_id"`
}

// NewSpawn copies a normalized definition into a stable spawn recipe.
func NewSpawn(id string, definition Definition, seed uint64, x, y float64, act, levelID int64) (Spawn, error) {
	spawn := Spawn{SpawnID: strings.TrimSpace(id), Definition: definition, Seed: seed, X: x, Y: y, Act: act, LevelID: levelID}
	if err := spawn.validate(); err != nil {
		return Spawn{}, err
	}
	return spawn, nil
}

func (spawn Spawn) validate() error {
	if spawn.SpawnID == "" || spawn.Definition.ID == "" || spawn.Definition.LifeMin <= 0 || spawn.Definition.LifeMax < spawn.Definition.LifeMin {
		return fmt.Errorf("monster: spawn identity and valid definition are required")
	}
	if math.IsNaN(spawn.X) || math.IsNaN(spawn.Y) || math.IsInf(spawn.X, 0) || math.IsInf(spawn.Y, 0) || spawn.X < 0 || spawn.Y < 0 || spawn.Act < 1 || spawn.Act > 5 || spawn.LevelID <= 0 {
		return fmt.Errorf("monster: spawn requires a finite position, act, and level")
	}
	if spawn.Definition.ColliderRadius <= 0 || spawn.Definition.SelectRadius <= 0 || len(spawn.Definition.Token) != 2 {
		return fmt.Errorf("monster: spawn definition lacks collision or presentation facts")
	}
	return nil
}

// Command encodes a trusted spawn for deterministic admission and replay.
func Command(spawn Spawn, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	if authority != simulation.AuthoritySystem && authority != simulation.AuthorityAdmin {
		return simulation.Command{}, fmt.Errorf("monster: spawn requires system or admin authority")
	}
	if err := spawn.validate(); err != nil {
		return simulation.Command{}, err
	}
	payload, err := json.Marshal(spawn)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: SpawnCommand, Payload: payload}, nil
}

// Register installs ordinary hostile materialization into session authority.
func Register(session *gamesession.Session) error {
	return session.Register(SpawnCommand, gamesession.CommandHandler{
		Validate: func(command simulation.Command) error {
			_, err := decodeSpawn(command.Payload)
			return err
		},
		Apply:   materialize,
		Allowed: []simulation.Authority{simulation.AuthoritySystem, simulation.AuthorityAdmin},
	})
}

func materialize(engine *gameecs.Engine, command simulation.Command) error {
	spawn, err := decodeSpawn(command.Payload)
	if err != nil {
		return err
	}
	stores, err := registerStores(engine.World())
	if err != nil {
		return err
	}
	for _, entity := range stores.identity.Entities() {
		component, _ := stores.identity.Get(entity)
		id, _ := component.Get("spawn_id")
		if id == spawn.SpawnID {
			return fmt.Errorf("monster: spawn %q already exists", spawn.SpawnID)
		}
	}
	entity, err := engine.World().CreateEntity()
	if err != nil {
		return err
	}
	fail := func(err error) error {
		engine.World().DestroyEntity(entity)
		return err
	}
	life := rollLife(spawn.Definition.LifeMin, spawn.Definition.LifeMax, spawn.Seed)
	seed := make([]byte, 8)
	for index := range seed {
		seed[7-index] = byte(spawn.Seed >> (index * 8))
	}
	components := []struct {
		store  *akara.DynamicStore
		values map[string]any
	}{
		{stores.identity, map[string]any{"spawn_id": spawn.SpawnID, "definition_id": spawn.Definition.ID, "base_id": spawn.Definition.BaseID, "graphics_id": spawn.Definition.GraphicsID, "seed": hex.EncodeToString(seed)}},
		{stores.stats, map[string]any{"level": spawn.Definition.Level, "health": life.Raw(), "max_health": life.Raw(), "defense": spawn.Definition.Defense, "attack_rating": spawn.Definition.AttackRating, "physical_min": spawn.Definition.PhysicalMin.Raw(), "physical_max": spawn.Definition.PhysicalMax.Raw(), "experience": spawn.Definition.Experience}},
		{stores.appearance, map[string]any{"token": spawn.Definition.Token, "mode": "NU", "weapon_class": spawn.Definition.WeaponClass, "name_key": spawn.Definition.NameKey}},
		{stores.ai, map[string]any{"behavior": spawn.Definition.AI, "state": AIIdle, "target_id": "", "next_think_tick": int64(command.Tick), "think_interval": int64(spawn.Definition.ThinkInterval), "aggro_radius": spawn.Definition.AggroRadius, "attack_range": spawn.Definition.AttackRange, "speed": float64(spawn.Definition.Velocity)}},
		{stores.position, map[string]any{"x": spawn.X, "y": spawn.Y}},
		{stores.velocity, map[string]any{"x": 0.0, "y": 0.0}},
		{stores.location, map[string]any{"act": spawn.Act, "level_id": spawn.LevelID}},
		{stores.collider, map[string]any{"radius": spawn.Definition.ColliderRadius}},
		{stores.selectable, map[string]any{"id": "monster:" + spawn.SpawnID, "kind": targeting.KindHostile, "label": spawn.Definition.NameKey, "owner": "", "radius": spawn.Definition.SelectRadius, "priority": int64(20)}},
	}
	for _, component := range components {
		if _, err := component.store.Set(entity, component.values); err != nil {
			return fail(err)
		}
	}
	return nil
}

type stores struct {
	identity, stats, appearance, ai, position, velocity, location, collider, selectable *akara.DynamicStore
}

func registerStores(world *akara.World) (stores, error) {
	schemas := []akara.Schema{
		{Name: "dm.monster.identity", Version: 1, Fields: []akara.Field{{Name: "spawn_id", Kind: akara.FieldString}, {Name: "definition_id", Kind: akara.FieldString}, {Name: "base_id", Kind: akara.FieldString}, {Name: "graphics_id", Kind: akara.FieldString}, {Name: "seed", Kind: akara.FieldString}}},
		{Name: "dm.monster.stats", Version: 1, Fields: []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}, {Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "physical_min", Kind: akara.FieldInt64}, {Name: "physical_max", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}}},
		{Name: "dm.monster.appearance", Version: 1, Fields: []akara.Field{{Name: "token", Kind: akara.FieldString}, {Name: "mode", Kind: akara.FieldString}, {Name: "weapon_class", Kind: akara.FieldString}, {Name: "name_key", Kind: akara.FieldString}}},
		aiSchema(),
		{Name: "dm.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.location", Version: 1, Fields: []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}},
		{Name: "dm.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}},
		targeting.Schema(),
	}
	registered := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		store, err := akara.RegisterSchema(world, schema)
		if err != nil {
			return stores{}, err
		}
		registered[index] = store
	}
	return stores{registered[0], registered[1], registered[2], registered[3], registered[4], registered[5], registered[6], registered[7], registered[8]}, nil
}

func decodeSpawn(encoded []byte) (Spawn, error) {
	var spawn Spawn
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&spawn); err != nil {
		return Spawn{}, fmt.Errorf("monster: decode spawn: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Spawn{}, fmt.Errorf("monster: spawn payload has trailing data")
	}
	spawn.SpawnID = strings.TrimSpace(spawn.SpawnID)
	if err := spawn.validate(); err != nil {
		return Spawn{}, err
	}
	return spawn, nil
}

func rollLife(minimum, maximum gamecombat.Amount, seed uint64) gamecombat.Amount {
	minimumWhole, _ := minimum.Whole(gamecombat.RoundTowardZero)
	maximumWhole, _ := maximum.Whole(gamecombat.RoundTowardZero)
	span := uint64(maximumWhole-minimumWhole) + 1
	state := seed + 0x9E3779B97F4A7C15
	state = (state ^ (state >> 30)) * 0xBF58476D1CE4E5B9
	state = (state ^ (state >> 27)) * 0x94D049BB133111EB
	state ^= state >> 31
	return gamecombat.MustWhole(minimumWhole + int64(state%span))
}
