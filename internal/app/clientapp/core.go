package clientapp

import (
	"context"
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
	gameinteraction "github.com/gravestench/dark-magic/internal/game/interaction"
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
	gameplayer "github.com/gravestench/dark-magic/internal/game/player"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
	loadcore "github.com/gravestench/dark-magic/internal/loading"
	"github.com/gravestench/dark-magic/internal/localization"
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
	session, err := gamesession.New(app.entitySimulation, gamesession.Config{})
	if err != nil {
		_ = app.entitySimulation.Close()
		return wrap("create offline game session", err)
	}
	app.offlineSession = session
	if err := app.buildInteractionAuthority(); err != nil {
		return err
	}
	if err := app.buildItemAuthority(); err != nil {
		return err
	}
	if err := app.registerOfflineCommands(); err != nil {
		return err
	}
	return app.buildLoadingCoordinator()
}

func (app *application) registerOfflineCommands() error {
	for name, register := range map[string]func(*gamesession.Session) error{
		"movement commands": gamesession.RegisterMovement,
		"skill commands":    gamesession.RegisterSkillAssignments,
		"player commands":   gameplayer.Register,
	} {
		if err := register(app.offlineSession); err != nil {
			return wrap("register "+name, err)
		}
	}
	if err := gameinteraction.RegisterCommands(app.offlineSession, app.interactionAuthority); err != nil {
		return wrap("register interaction commands", err)
	}
	if err := gameitem.RegisterCommands(app.offlineSession, app.itemAuthority); err != nil {
		return wrap("register item commands", err)
	}
	movement := &gamesession.MovementController{}
	movementSource, err := gamesession.NewMovementSource(app.entitySimulation, app.inputState, "local-player", "game_world", movement)
	if err != nil {
		return wrap("create offline movement source", err)
	}
	skills, err := gamesession.NewSkillSource(movement, "local-player")
	if err != nil {
		return wrap("create offline skill source", err)
	}
	skillProvider, err := app.skillProvider()
	if err != nil {
		return wrap("build starting skill provider", err)
	}
	entry, err := gameplayer.NewEntrySource(app.entitySimulation, app.saves, "local-player", 4096, 4096, skillProvider)
	if err != nil {
		return wrap("create offline player entry source", err)
	}
	sequencer := simulation.NewLocalSequencer()
	app.commandSource = func(tick uint64) []simulation.Command {
		commands := entry.Commands(tick)
		commands = append(commands, movementSource.Commands(tick)...)
		commands = append(commands, skills.Commands(tick)...)
		commands = append(commands, app.interactionSource.Commands(tick)...)
		commands = append(commands, app.itemSource.Commands(tick)...)
		return sequencer.Assign(commands)
	}
	app.playerControl = movement
	return nil
}

func (app *application) buildInteractionAuthority() error {
	var err error
	app.interactionAuthority, err = gameinteraction.NewAuthority(gameinteraction.Target{
		ID: "act1-akara", NPC: "Akara", Vendor: "Akara",
		Categories: []string{"armo", "weap", "misc"},
		X:          4096, Y: 4096, Radius: 160,
	})
	if err != nil {
		return wrap("create interaction authority", err)
	}
	initialTarget := ""
	if app.options.StartScene == "vendor" {
		initialTarget = "act1-akara"
	}
	if err := app.interactionAuthority.RegisterOwner("local-player", initialTarget); err != nil {
		return wrap("register local interaction owner", err)
	}
	app.interactionControl = &gameinteraction.Controller{}
	app.interactionSource, err = gameinteraction.NewSource(app.interactionControl, "local-player")
	return wrap("create local interaction command source", err)
}

func (app *application) buildItemAuthority() error {
	layout := gameitem.Layout{Grids: map[gameitem.Container]gameitem.Grid{
		gameitem.ContainerInventory: {Width: 10, Height: 4},
		gameitem.ContainerStash:     {Width: 6, Height: 8},
		gameitem.ContainerCube:      {Width: 3, Height: 4},
	}, BeltCapacity: 4, VendorGrid: gameitem.Grid{Width: 10, Height: 10}, Gold: gameitem.GoldBalance{Carried: 10000}}
	items, placements := app.developmentItems()
	state, err := gameitem.NewState(layout, items, placements)
	if err != nil {
		return wrap("create local item state", err)
	}
	app.itemAuthority = gameitem.NewAuthority()
	catalogSnapshot, err := app.gameData.Snapshot()
	if err != nil {
		return wrap("load vendor trade terms", err)
	}
	trades := make(gameitem.TradeCatalog, len(catalogSnapshot.NPCTradesByID))
	for vendor, record := range catalogSnapshot.NPCTradesByID {
		trades[vendor] = gameitem.TradeTerms{BuyMultiplier: int64(record.BuyMult), SellMultiplier: int64(record.SellMult), MaxBuy: int64(record.MaxBuy)}
	}
	app.itemAuthority.SetTradeCatalog(trades)
	app.itemAuthority.SetInteractionPolicy(app.interactionAuthority)
	if err := app.itemAuthority.Register("local-player", state); err != nil {
		return wrap("register local item state", err)
	}
	app.itemControl = &gameitem.Controller{}
	app.itemSource, err = gameitem.NewSource(app.itemControl, "local-player")
	return wrap("create local item command source", err)
}

func (app *application) developmentItems() ([]gameitem.Item, map[string]gameitem.Placement) {
	if app.options.FixtureCharacters <= 0 {
		return nil, nil
	}
	snapshot, err := app.gameData.Snapshot()
	if err != nil {
		return nil, nil
	}
	items := make([]gameitem.Item, 0, 6)
	placements := make(map[string]gameitem.Placement)
	if weapon, found := snapshot.WeaponsByCode["ssd"]; found {
		weaponPresentation := gameitem.Presentation{InventoryDC6: itemAsset(weapon.InvFile), WorldDC6: itemAsset(weapon.FlippyFile), WorldAnimated: true, Composite: compositeRecipe(weapon.Component, weapon.AlternateGfx), WeaponClass: strings.ToUpper(weapon.WeaponClass)}
		items = append(items, gameitem.Item{ID: "fixture-short-sword", Code: weapon.Code, Width: weapon.InvWidth, Height: weapon.InvHeight, BaseCost: int64(weapon.Cost), BodySlots: []string{"rarm", "larm"}, Presentation: weaponPresentation})
		placements["fixture-short-sword"] = gameitem.Placement{Container: gameitem.ContainerInventory, X: 0, Y: 0}
		items = append(items, gameitem.Item{ID: "fixture-vendor-short-sword", Code: weapon.Code, Width: weapon.InvWidth, Height: weapon.InvHeight, BaseCost: int64(weapon.Cost), BodySlots: []string{"rarm", "larm"}, Presentation: weaponPresentation})
		placements["fixture-vendor-short-sword"] = gameitem.Placement{Container: gameitem.ContainerVendor, Slot: "weap", Page: 0}
	}
	if armor, found := snapshot.ArmorByCode["cap"]; found {
		armorPresentation := gameitem.Presentation{InventoryDC6: itemAsset(armor.InvFile), WorldDC6: itemAsset(armor.FlippyFile), WorldAnimated: true, Composite: compositeRecipe(strconv.Itoa(armor.Component), armor.AlternateGfx)}
		items = append(items, gameitem.Item{ID: "fixture-hireling-cap", Code: armor.Code, Width: armor.InvWidth, Height: armor.InvHeight, BaseCost: int64(armor.Cost), BodySlots: []string{"head"}, Presentation: armorPresentation})
		placements["fixture-hireling-cap"] = gameitem.Placement{Container: gameitem.ContainerHireling, Slot: "head"}
		items = append(items, gameitem.Item{ID: "fixture-vendor-cap", Code: armor.Code, Width: armor.InvWidth, Height: armor.InvHeight, BaseCost: int64(armor.Cost), BodySlots: []string{"head"}, Presentation: armorPresentation})
		placements["fixture-vendor-cap"] = gameitem.Placement{Container: gameitem.ContainerVendor, Slot: "armo", Page: 0}
	}
	for index, code := range []string{"hp1", "mp1"} {
		if misc, found := snapshot.MiscByCode[code]; found {
			id := "fixture-" + code
			items = append(items, gameitem.Item{ID: id, Code: code, Width: misc.InvWidth, Height: misc.InvHeight, BaseCost: int64(misc.Cost), BeltEligible: true, Presentation: gameitem.Presentation{InventoryDC6: itemAsset(misc.InvFile), WorldDC6: itemAsset(misc.FlippyFile), WorldAnimated: true}})
			if code == "mp1" {
				placements[id] = gameitem.Placement{Container: gameitem.ContainerBelt, BeltSlot: 0}
			} else {
				placements[id] = gameitem.Placement{Container: gameitem.ContainerInventory, X: 2 + index, Y: 0}
			}
			if code == "hp1" {
				vendorID := "fixture-vendor-" + code
				items = append(items, gameitem.Item{ID: vendorID, Code: code, Width: misc.InvWidth, Height: misc.InvHeight, BaseCost: int64(misc.Cost), BeltEligible: true, Presentation: gameitem.Presentation{InventoryDC6: itemAsset(misc.InvFile), WorldDC6: itemAsset(misc.FlippyFile), WorldAnimated: true}})
				placements[vendorID] = gameitem.Placement{Container: gameitem.ContainerVendor, Slot: "misc", Page: 0}
			}
		}
	}
	for _, fixture := range []struct {
		code         string
		container    gameitem.Container
		beltEligible bool
	}{
		{code: "rvs", container: gameitem.ContainerStash, beltEligible: true},
		{code: "tsc", container: gameitem.ContainerCube},
	} {
		if misc, found := snapshot.MiscByCode[fixture.code]; found {
			id := "fixture-" + fixture.code
			items = append(items, gameitem.Item{ID: id, Code: fixture.code, Width: misc.InvWidth, Height: misc.InvHeight, BaseCost: int64(misc.Cost), BeltEligible: fixture.beltEligible, Presentation: gameitem.Presentation{InventoryDC6: itemAsset(misc.InvFile), WorldDC6: itemAsset(misc.FlippyFile), WorldAnimated: true}})
			placements[id] = gameitem.Placement{Container: fixture.container, X: 0, Y: 0}
		}
	}
	return items, placements
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
	case "game_world", "game_loading", "inventory", "character", "skills", "vendor":
		return true
	default:
		return false
	}
}
