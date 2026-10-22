package clientapp

import (
	"strings"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	d2catalog "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/catalog"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// TestOwnedAkaraDialogueProjectsSelectedLocalizedSpeech exercises authored NPC
// identity, admitted dialogue commands, localization and the connected mirror.
func TestOwnedAkaraDialogueProjectsSelectedLocalizedSpeech(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	createSelectedAcceptanceCharacter(t, fixture)
	fixture.advanceOffline(t, 10)
	targets := requireRealStore(t, fixture, "d2legacy.interaction.target", "interaction targets missing")
	positions := requireRealStore(t, fixture, "d2legacy.world.position", "positions missing")
	identities := requireRealStore(t, fixture, "d2legacy.player.identity", "players missing")

	var target akara.Entity

	for _, entity := range targets.Entities() {
		row, _ := targets.Get(entity)
		if id, _ := row.Get("id"); id == "act1-akara" {
			target = entity
		}
	}

	if target == 0 {
		t.Fatal("authored Akara target missing")
	}

	row, _ := targets.Get(target)
	x, _ := row.Get("x")
	y, _ := row.Get("y")
	// This is controlled test setup beside the real target, not movement proof.
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		if owner, _ := identity.Get("player"); owner == "local-player" {
			position, _ := positions.Get(entity)
			if err := position.Set("x", x); err != nil {
				t.Fatal(err)
			}

			if err := position.Set("y", y); err != nil {
				t.Fatal(err)
			}
		}
	}

	submit := func(kind string, payload map[string]any) {
		t.Helper()

		if err := fixture.app.commandIntents.Submit(kind, payload); err != nil {
			t.Fatal(err)
		}

		fixture.advanceOffline(t, 1)
	}
	submit("interaction.open", map[string]any{"target": "act1-akara"})
	submit("interaction.choose", map[string]any{"target": "act1-akara", "choice": "talk", "revision": 1})
	submit("interaction.choose", map[string]any{"target": "act1-akara", "choice": "gossip", "revision": 2})

	checkpoint, err := fixture.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	view, err := playeradapter.ProjectPrivateView("local-player", checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	if !view.Interaction.Active || view.Interaction.DialoguePage != "text" ||
		view.Interaction.DialogueLine != "akara_act1_gossip_8" || view.Interaction.DialogueRevision != 3 {
		t.Fatalf("selected dialogue = %#v", view.Interaction)
	}

	catalog, err := fixture.app.questCatalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	speech := catalog.SpeechByName[view.Interaction.DialogueLine]
	if speech.StringKey != "AkaraGossip9" {
		t.Fatalf("speech join = %#v", speech)
	}

	localized, err := fixture.app.locale.Text(speech.StringKey)
	if err != nil {
		t.Fatal(err)
	}

	_, body, err := d2catalog.ParseDialogueText(localized)
	if err != nil || strings.TrimSpace(body) == "" {
		t.Fatalf("localized caption unavailable: %v", err)
	}

	peer, err := playeradapter.ProjectPrivateView("unrelated-player", checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	if peer.Interaction.Active || peer.Interaction.DialogueLine != "" {
		t.Fatal("dialogue leaked to another owner")
	}

	engine := gameecs.New()

	t.Cleanup(func() { _ = engine.Close() })
	registerRemoteViewSchemas(t, engine)

	hero, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}

	if err := installPrivateProjection(engine.World(), hero, "local-player", nil, view); err != nil {
		t.Fatal(err)
	}

	contexts, _ := akara.GetDynamicStore(engine.World(), "d2legacy.interaction.context")

	mirror, _ := contexts.Get(contexts.Entities()[0])
	if page, _ := mirror.Get("dialogue_page"); page != "text" {
		t.Fatalf("mirrored page = %v", page)
	}

	if line, _ := mirror.Get("dialogue_line"); line != view.Interaction.DialogueLine {
		t.Fatalf("mirrored speech = %v", line)
	}

	if revision, _ := mirror.Get("dialogue_revision"); revision != int64(3) {
		t.Fatalf("mirrored revision = %v", revision)
	}

	locations, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.location")
	if len(locations.Entities()) != 1 {
		t.Fatal("mirrored interaction lost target location")
	}

	location, _ := locations.Get(locations.Entities()[0])
	if level, _ := location.Get("level_id"); level != int64(1) {
		t.Fatalf("mirrored target level = %v", level)
	}
}
