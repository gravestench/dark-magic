// Package entryworld prepares the first authoritative d2legacy game world for
// every topology. Interactive, listen, dedicated, and Realm authorities must
// all consume this adapter so map geometry, initial state, and admission spawn
// cannot drift between composition roots.
package entryworld

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

type Records interface {
	Load(string) ([]map[string]string, error)
	Invalidate(string)
	Loaded(string) bool
}

type Prepared struct {
	Worlds     map[int]*gameworld.Map
	Zones      map[int]*worldgen.Zone
	Spawns     map[int][2]float64
	Seam       gametransition.Seam
	Difficulty int
}

func Build(ctx context.Context, content, d2legacySource fs.FS, records Records, resolver gameworld.ObjectResolver, seed uint64, difficulty int) (*Prepared, error) {
	if content == nil || d2legacySource == nil || records == nil || resolver == nil {
		return nil, errors.New("d2legacy entry world: content, source, records, and object resolver are required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	generated, err := d2mapgen.GenerateEntryWorld(ctx, d2legacySource, records, seed, difficulty)
	if err != nil {
		return nil, err
	}
	town, err := materialize(ctx, content, generated.Town, resolver)
	if err != nil {
		return nil, fmt.Errorf("materialize Act I town: %w", err)
	}
	moor, err := materialize(ctx, content, generated.Wilderness, resolver)
	if err != nil {
		return nil, fmt.Errorf("materialize Blood Moor: %w", err)
	}
	seam, err := gametransition.ResolveSeam(generated.Seam, town, moor)
	if err != nil {
		return nil, fmt.Errorf("join Act I town to Blood Moor: %w", err)
	}
	townX, townY, found := d2mapgen.ResolveTownEntry(ctx, d2legacySource, records, town)
	if !found {
		return nil, errors.New("d2legacy entry world: Act I town has no campfire entry")
	}
	return &Prepared{
		Worlds: map[int]*gameworld.Map{generated.Seam.FirstLevel: town, generated.Seam.SecondLevel: moor},
		Zones:  map[int]*worldgen.Zone{generated.Seam.FirstLevel: generated.Town, generated.Seam.SecondLevel: generated.Wilderness},
		Spawns: map[int][2]float64{
			generated.Seam.FirstLevel:  {townX, townY},
			generated.Seam.SecondLevel: {seam.Wilderness.ArrivalX, seam.Wilderness.ArrivalY},
		},
		Seam:       seam,
		Difficulty: difficulty,
	}, nil
}

func materialize(ctx context.Context, content fs.FS, zone *worldgen.Zone, resolver gameworld.ObjectResolver) (*gameworld.Map, error) {
	materializer, err := gameworld.NewMaterializer(content, zone, resolver)
	if err != nil {
		return nil, err
	}
	for {
		err = materializer.Step(ctx)
		if errors.Is(err, gameworld.ErrMaterializationComplete) {
			break
		}
		if err != nil {
			return nil, err
		}
		progress := materializer.Progress()
		if progress.Completed == progress.Total {
			break
		}
	}
	return materializer.Result()
}

func (world *Prepared) Destination(levelID int) (playeradapter.Destination, error) {
	if world == nil || world.Worlds[levelID] == nil || world.Zones[levelID] == nil {
		return playeradapter.Destination{}, errors.New("d2legacy entry world: destination level is unavailable")
	}
	spawn, found := world.Spawns[levelID]
	if !found {
		return playeradapter.Destination{}, errors.New("d2legacy entry world: destination has no trusted spawn")
	}
	request := world.Zones[levelID].Request()
	gameMap := world.Worlds[levelID]
	return playeradapter.NewDestination(spawn[0], spawn[1], float64(gameMap.WidthSubtiles), float64(gameMap.HeightSubtiles),
		int64(request.Act), int64(request.LevelID))
}

func (world *Prepared) InitialData(owner string, developmentItems bool) map[string]any {
	return map[string]any{
		"d2legacy.game_rules": map[string]any{
			"target": "lod-1.14d", "expansion": true, "difficulty": world.Difficulty,
			"hardcore": false, "ladder": false, "maximum_players": 8,
		},
		"d2legacy.development_items": map[string]any{
			"enabled": developmentItems, "create_empty_containers": !developmentItems,
		},
		"d2legacy.interactions":      InteractionData(world.Worlds, owner, ""),
		"d2legacy.world_transitions": TransitionData(world.Seam),
	}
}

func TransitionData(seam gametransition.Seam) map[string]any {
	endpoint := func(source, destination gametransition.SeamEndpoint) map[string]any {
		return map[string]any{
			"source_level": float64(source.LevelID), "destination_level": float64(destination.LevelID),
			"source_x": source.X, "source_y": source.Y,
			"arrival_x": destination.ArrivalX, "arrival_y": destination.ArrivalY,
			"world_width": destination.Width, "world_height": destination.Height,
		}
	}
	return map[string]any{"seams": []any{
		endpoint(seam.Town, seam.Wilderness),
		endpoint(seam.Wilderness, seam.Town),
	}}
}

func InteractionData(worlds map[int]*gameworld.Map, owner, initial string) map[string]any {
	targets := []any{map[string]any{"id": "act1-akara", "npc": "Akara", "vendor": "Akara", "categories": "armo,misc,weap", "services": "", "x": float64(4096), "y": float64(4096), "radius": float64(160)}}
	levels := make([]int, 0, len(worlds))
	for levelID := range worlds {
		levels = append(levels, levelID)
	}
	sort.Ints(levels)
	for _, levelID := range levels {
		gameMap := worlds[levelID]
		objects := make(map[string]gameworld.Object, len(gameMap.Objects))
		for index, object := range gameMap.Objects {
			objects[fmt.Sprintf("ds1-object:%d:%d:%d", object.Type, object.ID, index)] = object
		}
		for _, selected := range gameMap.Selectables() {
			object := objects[selected.ID]
			name := strings.TrimSpace(object.Description)
			if name == "" {
				name = strings.TrimSpace(object.Class)
			}
			if name != "" {
				targets = append(targets, map[string]any{"id": selected.ID, "npc": name, "vendor": "", "categories": "", "services": "", "x": selected.X, "y": selected.Y, "radius": float64(4)})
			}
		}
	}
	return map[string]any{"owner": owner, "initial_target": initial, "targets": targets}
}

func (world *Prepared) PopulationData(nearby int) map[string]any {
	if world == nil {
		return nil
	}
	levelID := world.Seam.Wilderness.LevelID
	zone, gameMap := world.Zones[levelID], world.Worlds[levelID]
	if zone == nil || gameMap == nil {
		return nil
	}
	request := zone.Request()
	populated := make(map[uint32]bool)
	for _, stamp := range zone.Stamps() {
		populated[stamp.ID] = stamp.Populate
	}
	player := world.Spawns[levelID]
	rooms := make([]any, 0, len(zone.Rooms()))
	for _, room := range zone.Rooms() {
		var anchors [][2]float64
		if nearby > 0 {
			anchors = [][2]float64{{player[0] + 10, player[1]}, {player[0] + 7, player[1] + 7}, {player[0], player[1] + 10}, {player[0] - 7, player[1] + 7}}
		} else {
			centerX, centerY := float64((room.X+room.Width/2)*5)+2, float64((room.Y+room.Height/2)*5)+2
			anchors = [][2]float64{{centerX, centerY}, {centerX + 1, centerY}, {centerX, centerY + 1}, {centerX - 1, centerY}}
		}
		points := make([]any, 0, len(anchors))
		for _, anchor := range anchors {
			if x, y, ok := gameMap.OpenPointNearSubtile(anchor[0], anchor[1]); ok {
				points = append(points, map[string]any{"x": x, "y": y})
			}
		}
		rooms = append(rooms, map[string]any{
			"id": float64(room.ID), "populate": populated[room.StampID], "points": points,
			"x": float64(room.X * 5), "y": float64(room.Y * 5),
			"width": float64(room.Width * 5), "height": float64(room.Height * 5),
		})
	}
	links := make([]any, 0, len(zone.Links()))
	for _, link := range zone.Links() {
		links = append(links, map[string]any{"from": float64(link.From), "to": float64(link.To)})
	}
	return map[string]any{"seed": float64(request.Seed), "act": float64(request.Act), "level_id": float64(request.LevelID), "difficulty": float64(request.Difficulty), "rooms": rooms, "links": links}
}

func (world *Prepared) PopulationCommand(nearby int) (simulation.Command, error) {
	payload, err := json.Marshal(world.PopulationData(nearby))
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: 1, Player: "d2legacy.population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: payload}, nil
}

func (world *Prepared) InstallCollision(ctx context.Context, runtime *modruntime.Runtime) error {
	if world == nil || runtime == nil {
		return errors.New("d2legacy entry world: prepared world and runtime are required")
	}
	for levelID, collision := range world.Worlds {
		if err := modruntime.SetWorldMapForLevel(ctx, runtime, "d2legacy.gameplay.systems.init", "set_collision", levelID, collision); err != nil {
			return err
		}
	}
	return nil
}
