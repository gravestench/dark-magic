package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/recovered"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/game/data/worldobjects"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	loadcore "github.com/gravestench/dark-magic/internal/loading"
	"github.com/gravestench/dark-magic/internal/localization"
	d2legacymod "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	gameplayer "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/dark-magic/internal/persistence"
	raylibinput "github.com/gravestench/dark-magic/internal/platform/raylib/input"
	raylibrenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

func (app *application) loadSettings() error {
	path, err := darkpaths.ExpandHost(os.Getenv("DARK_MAGIC_SHELL_CONFIG"))
	if err != nil {
		return wrap("expand shell settings path", err)
	}
	app.shellSettings, err = shell.NewSettings(path)
	if err != nil {
		return wrap("load shell settings", err)
	}
	app.gameSettings, err = preferences.New(os.Getenv("DARK_MAGIC_PREFERENCES"))
	return wrap("load game preferences", err)
}

func (app *application) buildPresentationCore() error {
	profile, err := content.ResolvePresentationProfile(app.options.Content, app.options.PresentationProfileID)
	if err != nil {
		return err
	}
	app.profile = profile
	app.presentation, err = content.LoadPresentationBootstrap(app.options.Content)
	if err != nil {
		return err
	}

	// The renderer is the window. Input reads that window. Everything else talks
	// to backend-neutral stores so game code never needs to know about Raylib.
	app.renderer = &raylibrenderer.Service{}
	app.renderer.SetLogger(slog.Default().With("component", "renderer"))
	app.rendererConfig = raylibrenderer.DefaultConfig()
	app.rendererConfig.Resolution.Width = profile.Width
	app.rendererConfig.Resolution.Height = profile.Height
	app.rendererConfig.Resolution.Fit = app.options.ViewportFit
	app.rendererConfig.Window.Borderless = app.options.BorderlessFullscreen
	app.renderer.Configure(app.rendererConfig)
	if err := app.renderer.ConfigurePaletteQuantization(app.options.Content, app.options.OutputPalette); err != nil {
		return err
	}

	app.input = raylibinput.New(app.renderer)
	app.input.SetLogger(slog.Default().With("component", "input"))
	app.inputState = &inputstate.Store{}
	app.locale = localization.New(app.options.Content, "English")
	app.scripts = modruntime.New()
	app.composer = &render.Composer{}
	app.mixer = &audio.Mixer{}
	app.navigator = navigation.New()
	app.scenes = modruntime.NewScenes(app.scripts, app.navigator)
	app.scenes.SetInputStore(app.inputState)
	if app.options.Profile != nil {
		app.scenes.SetProfiler(app.options.Profile)
	}
	return nil
}

func (app *application) loadGameCatalogs() error {
	app.records = recordstore.New(app.options.Content)
	app.records.SetLogger(slog.Default().With("component", "records"))
	app.gameData = gamedata.New(app.records)
	app.questCatalog = recovered.New(app.options.Content)

	typed, err := app.gameData.Snapshot()
	if err != nil {
		return wrap("load typed game data", err)
	}
	recoveredData, err := app.questCatalog.Snapshot()
	if err != nil {
		return wrap("load recovered game data", err)
	}
	soundNames := make(map[string]struct{}, len(typed.Sounds))
	for _, sound := range typed.Sounds {
		soundNames[strings.ToLower(sound.Sound)] = struct{}{}
	}
	issues := recovered.ValidateReferences(recoveredData, soundNames, app.locale.Text)
	app.worldObjectResolver = worldobjects.New(recoveredData, typed)
	slog.Info("loaded game-data catalogs", "typed_issues", len(typed.Issues),
		"quests", len(recoveredData.Quests), "speech", len(recoveredData.Speech),
		"map_objects", len(recoveredData.MapObjects), "reference_issues", len(issues))
	return nil
}

func (app *application) buildOfflineSession() error {
	fixtures := DevelopmentCharacters(app.options.FixtureCharacters)
	app.saves = persistence.New(fixtures...)
	if len(fixtures) > 0 && fixtureNeedsSelection(app.options.StartScene) {
		if err := app.saves.Select(fixtures[0].ID); err != nil {
			return wrap("select development fixture", err)
		}
	}

	app.entitySimulation = gameecs.New()
	if err := app.buildEntryWorld(); err != nil {
		return err
	}
	session, err := gamesession.New(app.entitySimulation, gamesession.Config{})
	if err != nil {
		_ = app.entitySimulation.Close()
		return wrap("create offline game session", err)
	}
	app.offlineSession = session
	app.authoritativeState = simulation.NewStateStore()
	app.authoritativeRandom, err = d2legacymod.NewRandomStreams(0)
	if err != nil {
		return wrap("register d2legacy random streams", err)
	}
	identity, err := d2legacymod.Identity(app.options.Content)
	if err != nil {
		return wrap("identify d2legacy mod", err)
	}
	if err := session.RegisterAuthoritativeRuntime(identity, app.authoritativeState, app.authoritativeRandom); err != nil {
		return wrap("register d2legacy authoritative runtime", err)
	}
	app.commandIntents = &gamesession.IntentController{}
	app.commandIntentSource, err = gamesession.NewIntentSource(app.commandIntents, "local-player")
	if err != nil {
		return wrap("create local command intent source", err)
	}
	if err := app.buildItemAuthority(); err != nil {
		return err
	}
	if err := d2legacymod.ConfigureRuntime(app.scripts, app.options.Content, app.records, app.entitySimulation, app.offlineSession,
		app.authoritativeState, app.authoritativeRandom, map[string]any{"d2legacy.items": app.itemBootstrapData(), "d2legacy.interactions": app.interactionBootstrapData()}); err != nil {
		return wrap("configure canonical d2legacy runtime", err)
	}
	if err := app.registerOfflineCommands(); err != nil {
		return err
	}
	return app.buildLoadingCoordinator()
}

func (app *application) registerOfflineCommands() error {
	bloodMoor := app.gameWorlds[2]
	if bloodMoor == nil {
		return errors.New("register hostile simulation: Blood Moor world is unavailable")
	}
	if err := gameworld.RegisterVelocityMovement(app.entitySimulation, bloodMoor, gameworld.VelocityComponents{
		Position: "d2legacy.world.position", Velocity: "d2legacy.world.velocity", Collider: "d2legacy.world.collider",
	}); err != nil {
		return wrap("register generic velocity movement", err)
	}
	if err := app.queueEntryPopulation(); err != nil {
		return err
	}
	movement := &d2movement.MovementController{}
	movementSource, err := d2movement.NewMovementSource(app.entitySimulation, app.inputState, "local-player", "game_world", movement)
	if err != nil {
		return wrap("create offline movement source", err)
	}
	app.movementSource = movementSource
	skillProvider, err := app.skillProvider()
	if err != nil {
		return wrap("build starting skill provider", err)
	}
	entryLevel := app.activeWorldLevel
	worldMap := app.gameWorlds[entryLevel]
	if worldMap == nil {
		return errors.New("load offline entry map: session world is unavailable")
	}
	movementSource.SetNavigation(worldMap)
	spawn, found := app.gameWorldSpawns[entryLevel]
	if !found {
		return errors.New("create offline player entry source: world has no trusted spawn subtile")
	}
	spawnX, spawnY := spawn[0], spawn[1]
	request := app.gameWorldZones[entryLevel].Request()
	destination, err := gameplayer.NewDestination(spawnX, spawnY, float64(worldMap.WidthSubtiles), float64(worldMap.HeightSubtiles), int64(request.Act), int64(request.LevelID))
	if err != nil {
		return wrap("create Act I town admission destination", err)
	}
	entry, err := gameplayer.NewEntrySourceForDestination(app.entitySimulation, app.saves, "local-player", destination, skillProvider)
	if err != nil {
		return wrap("create offline player entry source", err)
	}
	app.transitionSource, err = gametransition.NewSource(app.entitySimulation, "local-player", app.transitionAuthority)
	if err != nil {
		return wrap("create zone transition source", err)
	}
	sequencer := simulation.NewLocalSequencer()
	app.commandSource = func(tick uint64) []simulation.Command {
		commands := entry.Commands(tick)
		commands = append(commands, movementSource.Commands(tick)...)
		commands = append(commands, app.commandIntentSource.Commands(tick)...)
		commands = append(commands, app.transitionSource.Commands(tick)...)
		return sequencer.Assign(commands)
	}
	app.playerControl = movement
	return nil
}

func (app *application) queueEntryPopulation() error {
	payload, err := json.Marshal(app.populationBootstrapData())
	if err != nil {
		return wrap("encode entry population geometry", err)
	}
	return wrap("queue d2legacy entry population", app.offlineSession.Submit(simulation.Command{
		Tick: 1, Player: "d2legacy.population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: payload,
	}))
}

func (app *application) populationBootstrapData() map[string]any {
	zone, worldMap := app.gameWorldZones[2], app.gameWorlds[2]
	if zone == nil || worldMap == nil {
		return nil
	}
	request := zone.Request()
	populated := map[uint32]bool{}
	for _, stamp := range zone.Stamps() {
		populated[stamp.ID] = stamp.Populate
	}
	nearby := developmentScenes[app.options.StartScene].nearbyHostiles
	player := app.gameWorldSpawns[2]
	rooms := make([]any, 0, len(zone.Rooms()))
	for _, room := range zone.Rooms() {
		points := make([]any, 0, 8)
		anchors := [][2]float64{}
		if nearby > 0 {
			anchors = [][2]float64{{player[0] + 10, player[1]}, {player[0] + 7, player[1] + 7}, {player[0], player[1] + 10}, {player[0] - 7, player[1] + 7}}
		} else {
			centerX, centerY := float64((room.X+room.Width/2)*5)+2, float64((room.Y+room.Height/2)*5)+2
			anchors = [][2]float64{{centerX, centerY}, {centerX + 1, centerY}, {centerX, centerY + 1}, {centerX - 1, centerY}}
		}
		for _, anchor := range anchors {
			if x, y, ok := worldMap.OpenPointNearSubtile(anchor[0], anchor[1]); ok {
				points = append(points, map[string]any{"x": x, "y": y})
			}
		}
		rooms = append(rooms, map[string]any{"id": float64(room.ID), "populate": populated[room.StampID], "points": points})
	}
	return map[string]any{"seed": float64(request.Seed), "act": float64(request.Act), "level_id": float64(request.LevelID), "difficulty": float64(request.Difficulty), "rooms": rooms}
}

func (app *application) buildItemAuthority() error {
	items, placements := app.developmentItems()
	catalogSnapshot, err := app.gameData.Snapshot()
	if err != nil {
		return wrap("load vendor trade terms", err)
	}
	trades := make(map[string]any, len(catalogSnapshot.NPCTradesByID))
	for vendor, record := range catalogSnapshot.NPCTradesByID {
		trades[vendor] = map[string]any{"buy_multiplier": float64(record.BuyMult), "sell_multiplier": float64(record.SellMult), "max_buy": float64(record.MaxBuy)}
	}
	app.itemInitialData = itemBootstrap(items, placements, trades)
	return nil
}

type bootstrapItem struct {
	id, code                 string
	width, height            int
	bodySlots                []string
	beltEligible             bool
	baseCost                 int64
	inventoryDC6, worldDC6   string
	worldAnimated            bool
	composite                map[string]string
	weaponClass              string
	meleeRange               float64
	physicalMin, physicalMax int64
	meleeWeaponClass         string
}

type bootstrapPlacement struct {
	container, slot                 string
	x, y, beltSlot, weaponSet, page int
}

func (app *application) developmentItems() ([]bootstrapItem, map[string]bootstrapPlacement) {
	if app.options.FixtureCharacters <= 0 {
		return nil, nil
	}
	snapshot, err := app.gameData.Snapshot()
	if err != nil {
		return nil, nil
	}
	items := make([]bootstrapItem, 0, 8)
	placements := make(map[string]bootstrapPlacement)
	if weapon, found := snapshot.WeaponsByCode["ssd"]; found {
		base := bootstrapItem{id: "fixture-short-sword", code: weapon.Code, width: weapon.InvWidth, height: weapon.InvHeight, baseCost: int64(weapon.Cost), bodySlots: []string{"rarm", "larm"}, inventoryDC6: itemAsset(weapon.InvFile), worldDC6: itemAsset(weapon.FlippyFile), worldAnimated: true, composite: compositeRecipe(weapon.Component, weapon.AlternateGfx), weaponClass: strings.ToUpper(weapon.WeaponClass), meleeRange: float64(1 + weapon.RangeAdder), physicalMin: int64(weapon.MinDam) * 256, physicalMax: int64(weapon.MaxDam) * 256, meleeWeaponClass: strings.ToUpper(weapon.WeaponClass)}
		items = append(items, base)
		placements[base.id] = bootstrapPlacement{container: "inventory"}
		base.id = "fixture-vendor-short-sword"
		items = append(items, base)
		placements[base.id] = bootstrapPlacement{container: "vendor", slot: "weap"}
	}
	if armor, found := snapshot.ArmorByCode["cap"]; found {
		base := bootstrapItem{id: "fixture-hireling-cap", code: armor.Code, width: armor.InvWidth, height: armor.InvHeight, baseCost: int64(armor.Cost), bodySlots: []string{"head"}, inventoryDC6: itemAsset(armor.InvFile), worldDC6: itemAsset(armor.FlippyFile), worldAnimated: true, composite: compositeRecipe(strconv.Itoa(armor.Component), armor.AlternateGfx)}
		items = append(items, base)
		placements[base.id] = bootstrapPlacement{container: "hireling", slot: "head"}
		base.id = "fixture-vendor-cap"
		items = append(items, base)
		placements[base.id] = bootstrapPlacement{container: "vendor", slot: "armo"}
	}
	for index, code := range []string{"hp1", "mp1"} {
		if misc, found := snapshot.MiscByCode[code]; found {
			id := "fixture-" + code
			base := bootstrapItem{id: id, code: code, width: misc.InvWidth, height: misc.InvHeight, baseCost: int64(misc.Cost), beltEligible: true, inventoryDC6: itemAsset(misc.InvFile), worldDC6: itemAsset(misc.FlippyFile), worldAnimated: true}
			items = append(items, base)
			if code == "mp1" {
				placements[id] = bootstrapPlacement{container: "belt", beltSlot: 0}
			} else {
				placements[id] = bootstrapPlacement{container: "inventory", x: 2 + index}
			}
			if code == "hp1" {
				vendorID := "fixture-vendor-" + code
				base.id = vendorID
				items = append(items, base)
				placements[vendorID] = bootstrapPlacement{container: "vendor", slot: "misc"}
			}
		}
	}
	for _, fixture := range []struct {
		code         string
		container    string
		beltEligible bool
	}{
		{code: "rvs", container: "stash", beltEligible: true},
		{code: "tsc", container: "cube"},
	} {
		if misc, found := snapshot.MiscByCode[fixture.code]; found {
			id := "fixture-" + fixture.code
			items = append(items, bootstrapItem{id: id, code: fixture.code, width: misc.InvWidth, height: misc.InvHeight, baseCost: int64(misc.Cost), beltEligible: fixture.beltEligible, inventoryDC6: itemAsset(misc.InvFile), worldDC6: itemAsset(misc.FlippyFile), worldAnimated: true})
			placements[id] = bootstrapPlacement{container: fixture.container}
		}
	}
	return items, placements
}

// itemBootstrapData converts the host/import boundary to policy-neutral value
// trees. Lua receives a deep copy and decides what these Diablo item facts mean.
func (app *application) itemBootstrapData() map[string]any {
	return app.itemInitialData
}

func itemBootstrap(items []bootstrapItem, placements map[string]bootstrapPlacement, tradeTerms map[string]any) map[string]any {
	result := map[string]any{
		"owner": "local-player", "belt_capacity": float64(4), "active_weapon_set": float64(0),
		"vendor_width": float64(10), "vendor_height": float64(10), "carried_gold": float64(10000), "stashed_gold": float64(0),
		"inventory_width": float64(10), "inventory_height": float64(4), "stash_width": float64(6), "stash_height": float64(8), "cube_width": float64(3), "cube_height": float64(4),
	}
	sort.Slice(items, func(i, j int) bool { return items[i].id < items[j].id })
	entries := make([]any, 0, len(items))
	for _, item := range items {
		placement := placements[item.id]
		components := make([]string, 0, len(item.composite))
		for component, appearance := range item.composite {
			components = append(components, component+"="+appearance)
		}
		sort.Strings(components)
		entries = append(entries, map[string]any{
			"id": item.id, "code": item.code, "width": float64(item.width), "height": float64(item.height), "body_slots": strings.Join(item.bodySlots, ","), "belt_eligible": item.beltEligible,
			"base_cost": float64(item.baseCost), "applied_services": "", "inventory_dc6": item.inventoryDC6, "world_dc6": item.worldDC6, "world_animated": item.worldAnimated, "composite": strings.Join(components, ","),
			"weapon_class": item.weaponClass, "melee_range": item.meleeRange, "physical_min": float64(item.physicalMin), "physical_max": float64(item.physicalMax), "melee_weapon_class": item.meleeWeaponClass,
			"container": placement.container, "x": float64(placement.x), "y": float64(placement.y), "slot": placement.slot, "belt_slot": float64(placement.beltSlot), "weapon_set": float64(placement.weaponSet), "page": float64(placement.page),
		})
	}
	result["items"] = entries
	result["trade_terms"] = tradeTerms
	return result
}

func (app *application) interactionBootstrapData() map[string]any {
	initial := ""
	if app.options.StartScene == "vendor" {
		initial = "act1-akara"
	}
	targets := []any{map[string]any{"id": "act1-akara", "npc": "Akara", "vendor": "Akara", "categories": "armo,misc,weap", "services": "", "x": float64(4096), "y": float64(4096), "radius": float64(160)}}
	for _, worldMap := range app.gameWorlds {
		objects := make(map[string]gameworld.Object, len(worldMap.Objects))
		for index, object := range worldMap.Objects {
			objects[fmt.Sprintf("ds1-object:%d:%d:%d", object.Type, object.ID, index)] = object
		}
		for _, selected := range worldMap.Selectables() {
			object := objects[selected.ID]
			name := strings.TrimSpace(object.Description)
			if name == "" {
				name = strings.TrimSpace(object.Class)
			}
			if name == "" {
				continue
			}
			targets = append(targets, map[string]any{"id": selected.ID, "npc": name, "vendor": "", "categories": "", "services": "", "x": selected.X, "y": selected.Y, "radius": float64(4)})
		}
	}
	return map[string]any{"owner": "local-player", "initial_target": initial, "targets": targets}
}

func compositeRecipe(component, appearance string) map[string]string {
	tokens := []string{"HD", "TR", "LG", "RA", "LA", "RH", "LH", "SH", "S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8"}
	component = strings.ToUpper(strings.TrimSpace(component))
	appearance = strings.ToUpper(strings.TrimSpace(appearance))
	if appearance == "" {
		return nil
	}
	for _, token := range tokens {
		if component == token {
			return map[string]string{token: appearance}
		}
	}
	index, err := strconv.Atoi(component)
	if err != nil || index < 0 || index >= len(tokens) {
		return nil
	}
	return map[string]string{tokens[index]: appearance}
}

func itemAsset(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return "data/global/items/" + name + ".dc6"
}

// learnedSkills translates character records into the small authoritative
// loadout admitted with a player. Generic actions and the class starting skill
// are available; the later save importer will add actually purchased skills.
func (app *application) skillProvider() (gameplayer.SkillProvider, error) {
	snapshot, err := app.gameData.Snapshot()
	if err != nil {
		return nil, err
	}
	return func(character persistence.Character) []gameplayer.Skill {
		return learnedSkills(snapshot, character)
	}, nil
}

func learnedSkills(snapshot gamedata.Snapshot, character persistence.Character) []gameplayer.Skill {
	start := ""
	for class, record := range snapshot.CharStatsByClass {
		if strings.EqualFold(class, character.Class) {
			start = record.StartSkill
			break
		}
	}
	result := make([]gameplayer.Skill, 0, 8)
	for _, skill := range snapshot.Skills {
		if strings.TrimSpace(skill.General) != "1" && !strings.EqualFold(skill.SkillName, start) {
			continue
		}
		description, found := snapshot.SkillDescByName[skill.SkillDesc]
		id, parseErr := strconv.ParseInt(strings.TrimSpace(skill.ID), 10, 64)
		if !found || parseErr != nil || description.ListRow < 0 || strings.TrimSpace(skill.Passive) == "1" {
			continue
		}
		result = append(result, gameplayer.Skill{ID: id, Level: 1, ListRow: int64(description.ListRow), LeftAllowed: strings.TrimSpace(skill.LeftSkill) == "1", RightAllowed: true})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].ID < result[right].ID })
	return result
}

func (app *application) buildLoadingCoordinator() error {
	app.loading = loadcore.New(map[string]loadcore.Task{
		"selected_character": func(context.Context) error {
			if _, ok := app.saves.Selected(); !ok {
				return errors.New("no character is selected")
			}
			return nil
		},
		"loading_assets": func(context.Context) error {
			for _, name := range app.presentation.LoadingAssets {
				if _, err := fs.Stat(app.options.Content, name); err != nil {
					return fmt.Errorf("load dependency %q: %w", name, err)
				}
			}
			return nil
		},
		"world": func(context.Context) error { return nil },
	})
	return nil
}

func fixtureNeedsSelection(scene string) bool {
	switch scene {
	case "game_world", "game_loading", "combat_lab", "inventory", "character", "skills", "vendor":
		return true
	default:
		return false
	}
}
