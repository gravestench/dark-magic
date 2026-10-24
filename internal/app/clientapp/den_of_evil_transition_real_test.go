package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	lua "github.com/yuin/gopher-lua"
)

// Removing selectable click routing, discrete terrain replacement, or camera snap must break this journey.
func TestDenOfEvilPointerRoundTripReal(t *testing.T) {
	for difficulty := range 3 {
		for _, hosted := range []bool{false, true} {
			t.Run(fmt.Sprintf("difficulty_%d/connected_%t", difficulty, hosted), func(t *testing.T) {
				app, input, scene, reconnect := newDenPointerClient(t, hosted, realD2LegacyFixtureConfig{
					fixtureWorldLevel: 2, gameDifficulty: difficulty,
				})
				denFrames(t, app, 10)
				pair := *app.caveWarp
				beforeMaps := maps.Clone(app.gameWorlds)
				beforeSeams := append([]byte(nil), mustDenTopology(t, app)...)

				beforeIDs := checkpointPopulationIDs(t, denCheckpoint(t, app))
				if len(beforeIDs) == 0 {
					t.Fatal("journey has no earlier outdoor residents")
				}

				denRejectRequests(t, app, pair)
				denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
				denFrames(t, app, 12)

				if location := denAuthorityHUD(t, app).Location; app.activeWorldLevel != 2 ||
					location.LevelID != 2 || location.TransitionSequence != 0 {
					t.Fatalf("standing on the cave entrance caused a proximity warp: %v", location)
				}

				scene.assertLevels(t, 1, 2, 3, 4)
				outdoor := denTerrainRoots(t, app)
				denClick(t, app, input, pair.First)
				denAssertArrival(t, app, pair.First, 8)

				for root := range outdoor {
					if app.composer.Exists(root) {
						t.Fatal("outdoor terrain survived the discrete cave warp")
					}
				}

				scene.assertLevels(t, 8)
				indoor := denTerrainRoots(t, app)

				if hosted {
					// Reconnect rotates the credential and refreshes canonical public state in the existing live session.
					reconnect()
					denFrames(t, app, 8)
					denAssertArrival(t, app, pair.First, 8)
				}

				denMoveTo(t, app, pair.Second.X, pair.Second.Y, 0.1)
				denFrames(t, app, 12)

				if location := denAuthorityHUD(t, app).Location; app.activeWorldLevel != 8 ||
					location.LevelID != 8 || location.TransitionSequence != 1 {
					t.Fatalf("standing on the cave exit caused a proximity warp: %v", location)
				}

				denClick(t, app, input, pair.Second)
				denAssertArrival(t, app, pair.Second, 2)

				for root := range indoor {
					if app.composer.Exists(root) {
						t.Fatal("indoor terrain survived the return warp")
					}
				}

				denTerrainRoots(t, app)
				scene.assertLevels(t, 1, 2, 3, 4)
				scene.assertSeams(t, app)

				if !hosted {
					denOutdoorSeams(t, app, scene)
				}

				if !reflect.DeepEqual(beforeSeams, mustDenTopology(t, app)) {
					t.Fatal("cave round trip changed prepared topology")
				}

				for level, world := range beforeMaps {
					if app.gameWorlds[level] != world {
						t.Fatalf("level %d was regenerated", level)
					}
				}

				afterIDs := checkpointPopulationIDs(t, denCheckpoint(t, app))
				for id, entity := range beforeIDs {
					if afterIDs[id] != entity {
						t.Fatalf("outdoor resident %s was recreated", id)
					}
				}
			})
		}
	}
}

// A Den checkpoint must retain the exact authority and permit the same ordinary clicked return.
func TestDenOfEvilPointerCheckpointReturnReal(t *testing.T) {
	for difficulty := range 3 {
		t.Run(fmt.Sprintf("difficulty_%d", difficulty), func(t *testing.T) { testDenCheckpointReturn(t, difficulty) })
	}
}

// testDenCheckpointReturn captures level-8 authority after scene teardown, then verifies a fresh scene can use it.
func testDenCheckpointReturn(t *testing.T, difficulty int) {
	app, input := newPointerWorldClient(t, false, realD2LegacyFixtureConfig{
		fixtureWorldLevel: 2, gameDifficulty: difficulty,
	})
	denFrames(t, app, 10)
	pair := *app.caveWarp
	denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
	denClick(t, app, input, pair.First)
	denAssertArrival(t, app, pair.First, 8)
	// Close the transient scene through its normal lifecycle before saving gameplay authority.
	// Live-scene camera serialization is tracked separately in issue #657.
	if err := app.scenes.Close(t.Context()); err != nil {
		t.Fatal(err)
	}

	bundle, err := app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	restored, restoredInput := newPointerWorldClient(t, false,
		realD2LegacyFixtureConfig{fixtureWorldLevel: 2, gameDifficulty: difficulty, recovery: &bundle},
		func(restored *application) error {
			// Delay scene creation until the restored canonical state has been checked byte for byte.
			restored.options.StartScene = ""
			err := restored.loadScriptComponents()
			restored.options.StartScene = "game_world"

			if err != nil {
				return err
			}

			got, err := restored.offlineSession.RecoveryCheckpoint()
			if err != nil {
				return err
			}

			if got.Checksum != bundle.Session.Checksum {
				return fmt.Errorf("Den checkpoint authority changed at tick %d", got.State.Tick)
			}

			if err := restored.network.StartSelected(); err != nil {
				return err
			}

			return restored.openInitialScene()
		})
	if err := app.openInitialScene(); err != nil {
		t.Fatal(err)
	}

	if restored.entryWorldChecksum != bundle.TopologyChecksum || !reflect.DeepEqual(restored.caveWarp, app.caveWarp) {
		t.Fatal("restored Den topology or warp identities changed")
	}

	denFrames(t, restored, 10)
	denAssertArrival(t, restored, pair.First, 8)

	for _, client := range []struct {
		app   *application
		input *bridgeTestInput
	}{{app, input}, {restored, restoredInput}} {
		denMoveTo(t, client.app, pair.Second.X, pair.Second.Y, 0.1)
		denClick(t, client.app, client.input, pair.Second)
		denAssertArrival(t, client.app, pair.Second, 2)
	}

	originalIDs := checkpointPopulationIDs(t, denCheckpoint(t, app))

	restoredIDs := checkpointPopulationIDs(t, denCheckpoint(t, restored))
	if !reflect.DeepEqual(originalIDs, restoredIDs) {
		t.Fatal("restored return changed outdoor resident identities")
	}
}

// denFrames advances presentation and gives connected correction delivery time between frames.
func denFrames(t *testing.T, app *application, frames int) {
	t.Helper()

	for range frames {
		if app.network.Connected() {
			time.Sleep(40 * time.Millisecond)
		}

		advanceBridgePointerFrame(t, app)
	}
}

// denMoveTo waits for the production movement route to finish before the interaction click.
func denMoveTo(t *testing.T, app *application, x, y, radius float64) {
	t.Helper()

	if err := app.playerControl.SetMoveTargetWithRadius(x, y, radius); err != nil {
		t.Fatal(err)
	}

	for range 2000 {
		denFrames(t, app, 1)

		px, py, ok := app.controlledPlayerPosition()
		if ok && math.Hypot(px-x, py-y) <= math.Max(radius, 0.4) && !app.playerControl.HasMoveTarget() {
			return
		}
	}

	px, py, _ := app.controlledPlayerPosition()
	t.Fatalf("cave approach stopped at %.2f,%.2f, want %.2f,%.2f", px, py, x, y)
}

// denClick uses the scene pointer path and requires exactly one authoritative relocation.
func denClick(t *testing.T, app *application, input *bridgeTestInput, endpoint entryworld.CaveWarpEndpoint) {
	t.Helper()
	before := denAuthorityHUD(t, app)
	camera := bridgeCameraPosition(t, app)

	cursor := app.currentWorld().Map.Coordinates().SubtileToScreen(
		gameworld.Point{X: endpoint.X, Y: endpoint.Y}, camera, gameworld.Point{X: 400, Y: 330})
	if cursor.X < 0 || cursor.X >= 800 || cursor.Y < 0 || cursor.Y >= 480 {
		t.Fatalf("cave selectable outside viewport: %v", cursor)
	}

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true, Down: true}}}

	denFrames(t, app, 1)

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y}

	for range 150 {
		denFrames(t, app, 1)

		if app.activeWorldLevel == endpoint.TargetLevel {
			denFrames(t, app, 5)

			if after := denAuthorityHUD(t, app); after.Location.TransitionSequence != before.Location.TransitionSequence+1 {
				t.Fatal("one cave click did not commit exactly one relocation")
			}

			return
		}
	}

	t.Fatalf("ordinary cave pointer click stayed in level %d, want %d", app.activeWorldLevel, endpoint.TargetLevel)
}

// denAssertArrival checks authority, presentation, and camera against the prepared safe destination.
func denAssertArrival(t *testing.T, app *application, endpoint entryworld.CaveWarpEndpoint, level int) {
	t.Helper()
	denAssertRelocation(t, app, "warp", endpoint.LevelID, level)

	authority := denAuthorityHUD(t, app)

	wantSequence := int64(1)
	if level == 2 {
		wantSequence = 2
	}

	if authority.Location.TransitionSequence != wantSequence {
		t.Fatalf("arrival relocation count %d, want %d", authority.Location.TransitionSequence, wantSequence)
	}

	if int(authority.Location.LevelID) != level || app.activeWorldLevel != level {
		t.Fatalf("authority/presentation level %d/%d, want %d", authority.Location.LevelID, app.activeWorldLevel, level)
	}

	if math.Hypot(authority.Position.X-endpoint.ArrivalX, authority.Position.Y-endpoint.ArrivalY) > .01 {
		t.Fatalf("canonical arrival %v differs from prepared destination", authority.Position)
	}

	x, y, ok := app.controlledPlayerPosition()
	if !ok || math.Hypot(x-endpoint.ArrivalX, y-endpoint.ArrivalY) > .01 {
		t.Fatalf("arrival %.3f,%.3f, want %.3f,%.3f", x, y, endpoint.ArrivalX, endpoint.ArrivalY)
	}

	position := gameworld.Point{X: x, Y: y}
	if !app.currentWorld().Map.WalkableStep(position, position, 1) {
		t.Fatalf("unsafe cave arrival: %v", position)
	}

	camera := bridgeCameraPosition(t, app)
	if math.Hypot(camera.X-x, camera.Y-y) > .01 {
		t.Fatalf("cave camera %v did not snap to committed player %.3f,%.3f", camera, x, y)
	}

	if app.network.Connected() {
		hud, _ := app.network.client.View()
		if int(hud.Location.LevelID) != level || math.Hypot(hud.Position.X-x, hud.Position.Y-y) > .01 {
			t.Fatal("connected public arrival differs from rendered player")
		}
	}

	if app.playerControl.HasMoveTarget() {
		t.Fatal("movement target remains after warp arrival")
	}
}

// denCheckpoint reads the active offline or hosted authority through its canonical checkpoint API.
func denCheckpoint(t *testing.T, app *application) simulation.Checkpoint {
	t.Helper()

	var (
		checkpoint simulation.Checkpoint
		err        error
	)
	if app.network.Connected() {
		checkpoint, err = app.network.host.Session.CanonicalCheckpoint()
	} else {
		checkpoint, err = app.offlineSession.CanonicalCheckpoint()
	}

	if err != nil {
		t.Fatal(err)
	}

	return checkpoint
}

// mustDenTopology captures the durable recipe so the journey can reject accidental regeneration.
func mustDenTopology(t *testing.T, app *application) []byte {
	t.Helper()

	snapshot, err := app.entryWorldTopology.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	return snapshot
}

// Observe retained terrain handles, not a fake renderer's method calls or protected image pixels.
func denTerrainRoots(t *testing.T, app *application) map[render.NodeID]bool {
	t.Helper()

	for range 100 {
		roots := map[render.NodeID]bool{}

		for _, node := range app.composer.Snapshot() {
			if node.Resource == (render.ResourceID{}) {
				continue
			}

			resource, err := app.composer.ResourceSnapshot(node.Resource)
			if err != nil {
				t.Fatal(err)
			}

			if strings.HasPrefix(resource.TextureKey, "world-dt1:") {
				roots[node.Parent] = true
			}
		}

		if len(roots) > 0 {
			return roots
		}

		time.Sleep(time.Millisecond)
		denFrames(t, app, 1)
	}

	t.Fatal("production scene never published terrain nodes")

	return nil
}

// denRejectRequests proves unadmitted far, wrong-level, and stale targets leave authority in Blood Moor.
func denRejectRequests(t *testing.T, app *application, pair entryworld.CaveWarpPair) {
	t.Helper()

	before := denAuthorityHUD(t, app)
	if math.Hypot(before.Position.X-pair.First.X, before.Position.Y-pair.First.Y) <= 3.5 {
		t.Fatal("far request fixture starts in cave interaction range")
	}

	for _, id := range []string{pair.First.ID, pair.Second.ID, "static-warp:stale"} {
		if err := app.commandIntents.Submit("interaction.open", map[string]any{"target": id}); err != nil {
			t.Fatal(err)
		}

		denFrames(t, app, 8)

		after := denAuthorityHUD(t, app)
		if app.activeWorldLevel != 2 || after.Location != before.Location || after.Position != before.Position {
			t.Fatalf("far/wrong-level/stale target %s relocated player", id)
		}
	}
}

// denSceneObservation observes the actual scene's prepared component after its normal update.
// It never supplies layout, positions, or behavior to the scene under test.
type denSceneObservation struct {
	levels  []int
	tiles   map[int]int
	offsets map[int]gameworld.Point
}

// install observes the real scene after its update without supplying layout or changing scene policy.
func (observation *denSceneObservation) install(t *testing.T, app *application) {
	t.Helper()

	err := app.scripts.Run(t.Context(), func(state *lua.LState) error {
		if err := state.CallByParam(lua.P{Fn: state.GetGlobal("require"), NRet: 1, Protect: true},
			lua.LString("d2legacy.screens.game_world")); err != nil {
			return err
		}

		definition := state.Get(-1).(*lua.LTable)
		state.Pop(1)

		update := definition.RawGetString("update")
		definition.RawSetString("update", state.NewFunction(func(state *lua.LState) int {
			self := state.CheckTable(1)

			args := make([]lua.LValue, state.GetTop())
			for i := range args {
				args[i] = state.Get(i + 1)
			}

			if err := state.CallByParam(lua.P{Fn: update, Protect: true}, args...); err != nil {
				state.RaiseError("%v", err)
				return 0
			}

			observation.levels = nil
			observation.tiles = map[int]int{}
			observation.offsets = map[int]gameworld.Point{}

			if layout, ok := self.RawGetString("layout").(*lua.LTable); ok {
				layout.RawGetString("levels").(*lua.LTable).ForEach(func(key, value lua.LValue) {
					id := int(key.(lua.LNumber))
					observation.levels = append(observation.levels, id)
					placement := value.(*lua.LTable)
					observation.offsets[id] = gameworld.Point{
						X: float64(placement.RawGetString("offset_x").(lua.LNumber)),
						Y: float64(placement.RawGetString("offset_y").(lua.LNumber)),
					}
				})
			}

			if adapters, ok := self.RawGetString("adapters").(*lua.LTable); ok {
				adapters.ForEach(func(key, value lua.LValue) {
					id := int(key.(lua.LNumber))

					if nodes, ok := value.(*lua.LTable).RawGetString("nodes").(*lua.LTable); ok {
						nodes.ForEach(func(_, _ lua.LValue) { observation.tiles[id]++ })
					}
				})
			}

			return 0
		}))

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// assertLevels requires the active scene to compose precisely the expected indoor or outdoor maps.
func (observation *denSceneObservation) assertLevels(t *testing.T, want ...int) {
	t.Helper()
	slices.Sort(observation.levels)

	if !reflect.DeepEqual(observation.levels, want) {
		t.Fatalf("live terrain component levels %v, want %v", observation.levels, want)
	}
}

// Own only the watch cancellation so credential rotation happens after its old stream has stopped.
func newDenPointerClient(t *testing.T, hosted bool, config realD2LegacyFixtureConfig,
) (*application, *bridgeTestInput, *denSceneObservation, func()) {
	t.Helper()

	scene := &denSceneObservation{}

	var reconnect func()

	app, input := newPointerWorldClient(t, hosted, config, func(app *application) error {
		scene.install(t, app)

		if !hosted {
			return app.loadScriptComponents()
		}
		// Component discovery and bootstrap precede network composition. Delay direct-start admission
		// until the test owns the watch context, matching the Stony Field reconnect acceptance.
		app.options.StartScene = ""
		err := app.loadScriptComponents()
		app.options.StartScene = "game_world"

		if err != nil {
			return err
		}

		app.network.mu.Lock()
		generation, ctx, _, _ := app.network.beginSelectedNetworkLocked()
		app.network.mu.Unlock()

		runtime := &listenHostRuntime{}
		if err := app.network.prepareListenHost(ctx, runtime); err != nil {
			runtime.close()
			return err
		}

		app.network.startListenServices(ctx, generation, runtime)

		if err := app.network.connectListenPlayer(ctx, runtime); err != nil {
			runtime.close()
			return err
		}

		if !app.network.commitListenHost(generation, runtime) {
			runtime.close()
			return errors.New("listen host commit failed")
		}
		go app.network.send(ctx, runtime.client)

		watchContext, stopWatch := context.WithCancel(ctx)

		watchDone := make(chan struct{})
		go func() { defer close(watchDone); app.network.watch(watchContext, runtime.client) }()

		t.Cleanup(stopWatch)

		reconnect = func() {
			stopWatch()

			select {
			case <-watchDone:
			case <-time.After(5 * time.Second):
				t.Fatal("old correction stream did not stop")
			}

			epoch := app.network.currentConnectionEpoch()
			if err := app.network.recover(ctx, runtime.client, epoch, errors.New("Den acceptance interruption")); err != nil {
				t.Fatal(err)
			}

			if app.network.currentConnectionEpoch() != epoch+1 {
				t.Fatal("reconnect did not advance connection epoch")
			}
			go app.network.watch(ctx, runtime.client)
		}

		if err := app.advanceGame(time.Second / 25); err != nil {
			return err
		}

		return app.openInitialScene()
	})

	return app, input, scene, reconnect
}

// Return traverses every existing seamless join with production movement and the live scene.
func denOutdoorSeams(t *testing.T, app *application, scene *denSceneObservation) {
	t.Helper()

	seams := app.transitionSeams
	if len(seams) != 3 {
		t.Fatalf("outdoor seams = %d, want 3", len(seams))
	}

	for _, crossing := range []struct {
		from, to int
		point    gametransition.SeamEndpoint
	}{
		{2, 1, seams[0].Second}, {1, 2, seams[0].First}, {2, 3, seams[1].First},
		{3, 4, seams[2].First}, {4, 3, seams[2].Second}, {3, 2, seams[1].Second},
	} {
		roots := denTerrainRoots(t, app)
		if err := app.playerControl.SetMoveTargetWithRadius(crossing.point.X, crossing.point.Y, 0.5); err != nil {
			t.Fatal(err)
		}

		reached := false

		for range 2500 {
			denFrames(t, app, 1)

			if app.activeWorldLevel == crossing.to {
				reached = true
				break
			}
		}

		if !reached {
			t.Fatalf("post-cave seamless movement %d -> %d failed", crossing.from, crossing.to)
		}

		scene.assertLevels(t, 1, 2, 3, 4)

		for attempts := 0; attempts < 100 && (scene.tiles[crossing.from] == 0 || scene.tiles[crossing.to] == 0); attempts++ {
			time.Sleep(time.Millisecond)
			denFrames(t, app, 1)
		}

		if scene.tiles[crossing.from] == 0 || scene.tiles[crossing.to] == 0 {
			t.Fatalf("seam %d -> %d lacks terrain on both sides: %v", crossing.from, crossing.to, scene.tiles)
		}

		for root := range roots {
			if !app.composer.Exists(root) {
				t.Fatalf("seam %d -> %d replaced outdoor terrain root", crossing.from, crossing.to)
			}
		}

		denAssertRelocation(t, app, "seam", crossing.from, crossing.to)
	}
}

// denAssertRelocation distinguishes intentional warps from preserved seamless outdoor crossings.
func denAssertRelocation(t *testing.T, app *application, kind string, from, to int) {
	t.Helper()

	location := denAuthorityHUD(t, app).Location
	if location.TransitionKind != kind || location.TransitionSource != int64(from) || location.LevelID != int64(to) {
		t.Fatalf("relocation = %v, want %s %d -> %d", location, kind, from, to)
	}
}

// Project from the canonical checkpoint rather than reading mutable server stores.
func denAuthorityHUD(t *testing.T, app *application) playeradapter.HUD {
	t.Helper()

	owner := "local-player"
	if app.network.Connected() {
		owner = app.clientWorld.lastHUD.Player.PlayerID
	}

	payload, err := playeradapter.ProjectHUD(owner, denCheckpoint(t, app))
	if err != nil {
		t.Fatal(err)
	}

	var hud playeradapter.HUD
	if err := json.Unmarshal(payload, &hud); err != nil {
		t.Fatal(err)
	}

	return hud
}

// The native projection of each prepared join must meet in the actual Lua scene layout.
func (observation *denSceneObservation) assertSeams(t *testing.T, app *application) {
	t.Helper()

	if len(app.transitionSeams) != 3 {
		t.Fatal("return lost an outdoor seam")
	}

	for _, seam := range app.transitionSeams {
		for _, direction := range [][2]gametransition.SeamEndpoint{{seam.First, seam.Second}, {seam.Second, seam.First}} {
			source, destination := direction[0], direction[1]
			from := app.gameWorlds[source.LevelID].Coordinates().SubtileToWorldPixel(gameworld.Point{X: source.X, Y: source.Y})
			to := app.gameWorlds[destination.LevelID].Coordinates().SubtileToWorldPixel(
				gameworld.Point{X: destination.ArrivalX, Y: destination.ArrivalY})

			first, second := observation.offsets[source.LevelID], observation.offsets[destination.LevelID]
			if math.Abs(from.X+first.X-to.X-second.X) > 0.5 || math.Abs(from.Y+first.Y-to.Y-second.Y) > 0.5 {
				t.Fatalf("scene separates outdoor seam %d -> %d", source.LevelID, destination.LevelID)
			}
		}
	}
}
