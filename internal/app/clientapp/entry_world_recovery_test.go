package clientapp

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// Recovery must preserve the Cold Plains checkpoint and produce the same authoritative result after a return crossing.
func TestColdPlainsRecoveryMatchesContinuation(t *testing.T) {
	seedFixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})

	seedTwo, err := d2mapgen.GenerateEntryWorld(t.Context(), content.D2Legacy(), seedFixture.app.records, 2, 0)
	if err != nil {
		t.Fatal(err)
	}

	seedTwoSnapshot, err := seedTwo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	source := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 3,
		applySceneDefaults: true, topologySnapshot: seedTwoSnapshot,
	})
	if err := source.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	source.advanceGame(t, 10)

	bundle, err := source.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if bundle.TopologyChecksum == seedFixture.app.entryWorldChecksum {
		t.Fatal("alternate seed did not change the entry topology")
	}

	savedHUD := recoveryPlayerHUD(t, source)

	target := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 3,
		applySceneDefaults: true, skipPopulationBootstrap: true, recovery: &bundle,
	})
	if err := target.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	restoredHUD := recoveryPlayerHUD(t, target)
	if savedHUD.Location != restoredHUD.Location || savedHUD.Position != restoredHUD.Position ||
		savedHUD.Movement != restoredHUD.Movement || savedHUD.Player != restoredHUD.Player {
		t.Fatalf("restored player facts = %#v, want %#v", restoredHUD, savedHUD)
	}

	if target.app.entryWorldChecksum != bundle.TopologyChecksum || target.app.gameWorlds[3] == nil {
		t.Fatal("recovery did not compose the saved Cold Plains topology")
	}

	actual, err := target.app.offlineSession.RecoveryCheckpoint()
	if err != nil || actual.Checksum != bundle.Session.Checksum {
		t.Fatalf("restored recovery checksum = %s, want %s: %v", actual.Checksum, bundle.Session.Checksum, err)
	}

	seam := source.app.transitionSeams[1].Second
	for _, fixture := range []*realD2LegacyFixture{source, target} {
		if err := fixture.app.playerControl.SetMoveTargetWithRadius(seam.X, seam.Y, 0.5); err != nil {
			t.Fatal(err)
		}
	}

	for range 25 {
		source.advanceGame(t, 1)
		target.advanceGame(t, 1)
	}

	sourceState, err := source.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	targetState, err := target.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if targetState.Tick != sourceState.Tick || targetState.Checksum != sourceState.Checksum {
		t.Fatalf("recovered continuation = tick %d checksum %s, want tick %d checksum %s",
			targetState.Tick, targetState.Checksum, sourceState.Tick, sourceState.Checksum)
	}

	if source.app.activeWorldLevel != 2 || target.app.activeWorldLevel != 2 {
		t.Fatalf("return crossing ended at source/target levels %d/%d, want 2/2",
			source.app.activeWorldLevel, target.app.activeWorldLevel)
	}

	sourceHUD := recoveryPlayerHUD(t, source)

	targetHUD := recoveryPlayerHUD(t, target)
	if sourceHUD.Player != targetHUD.Player || sourceHUD.Location != targetHUD.Location ||
		sourceHUD.Position != targetHUD.Position || sourceHUD.Movement != targetHUD.Movement {
		t.Fatalf("recovered return player facts = %#v, want %#v", targetHUD, sourceHUD)
	}
}

// recoveryPlayerHUD projects the authority's canonical player facts for a recovery parity assertion.
func recoveryPlayerHUD(t *testing.T, fixture *realD2LegacyFixture) playeradapter.HUD {
	t.Helper()

	checkpoint, err := fixture.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	payload, err := playeradapter.ProjectHUD("local-player", checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	var hud playeradapter.HUD
	if err := json.Unmarshal(payload, &hud); err != nil {
		t.Fatal(err)
	}

	return hud
}

// The saved map may match while another rule changes the runtime recipe; that checkpoint must not mutate the target.
func TestColdPlainsRecoveryRejectsDifferentRuntimeIdentity(t *testing.T) {
	source := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 3, applySceneDefaults: true,
	})
	if err := source.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	source.advanceGame(t, 10)

	bundle, err := source.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	target := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 3,
		applySceneDefaults: true, skipPopulationBootstrap: true,
		topologySnapshot: bundle.Topology, gameDifficulty: 1,
	})

	before, err := target.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if err := target.app.restoreOfflineEntryWorldRecovery(bundle); err == nil {
		t.Fatal("different runtime identity restored")
	}

	after, err := target.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if after.Tick != before.Tick || after.Checksum != before.Checksum || target.app.activeWorldLevel != 3 {
		t.Fatal("rejected runtime identity mutated target session or active world")
	}
}

// A saved v1 town/Blood Moor world remains a two-level session after current code composes it.
func TestLegacyEntryWorldSnapshotComposesWithoutColdPlains(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	generated := fixture.app.entryWorldTopology
	legacySeam := generated.Seams[0]
	legacySeam.TriggerRadius = 0
	legacySeam.MaxArrivalDistance = 0
	legacySeam.ArrivalInset = 0
	legacySeam.FirstWarpID = 0
	legacySeam.SecondWarpID = 0
	legacySeam.FirstWarpRole = ""
	legacySeam.SecondWarpRole = ""
	legacy := d2mapgen.EntryWorld{
		Zones: map[int]*worldgen.Zone{1: generated.Zones[1], 2: generated.Zones[2]},
		Seams: []d2mapgen.SeamSpec{legacySeam}, Version: 1,
	}

	snapshot, err := legacy.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	restored := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{topologySnapshot: snapshot})
	if restored.app.entryWorldTopology.Version != 1 || len(restored.app.gameWorlds) != 2 ||
		len(restored.app.transitionSeams) != 1 || restored.app.gameWorlds[3] != nil {
		t.Fatal("v1 composition gained Cold Plains")
	}

	wantChecksum, err := legacy.Checksum()
	if err != nil || restored.app.entryWorldChecksum != wantChecksum {
		t.Fatalf("v1 checksum = %s, want %s: %v", restored.app.entryWorldChecksum, wantChecksum, err)
	}

	if len(restored.app.transitionBootstrapData()["seams"].([]any)) != 2 {
		t.Fatal("v1 entry seam lost reciprocal transitions")
	}
}

// Even a valid alternate topology must fail before a newly composed session restores a tick or participant state.
func TestColdPlainsRecoveryRejectsDifferentTopology(t *testing.T) {
	source := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 3, applySceneDefaults: true,
	})
	if err := source.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	source.advanceGame(t, 10)

	bundle, err := source.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	other, err := d2mapgen.GenerateEntryWorld(t.Context(), content.D2Legacy(), source.app.records, 2, 0)
	if err != nil {
		t.Fatal(err)
	}

	bundle.Topology, err = other.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	bundle.TopologyChecksum, err = other.Checksum()
	if err != nil {
		t.Fatal(err)
	}

	target := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 3,
		applySceneDefaults: true, skipPopulationBootstrap: true,
	})

	before, err := target.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if err := target.app.restoreOfflineEntryWorldRecovery(bundle); err == nil {
		t.Fatal("different valid topology restored into current composition")
	}

	after, err := target.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if after.Tick != before.Tick || after.Checksum != before.Checksum || target.app.activeWorldLevel != 3 {
		t.Fatal("rejected recovery mutated target session or active world")
	}
}
