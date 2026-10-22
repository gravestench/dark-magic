package d2legacy

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	lua "github.com/yuin/gopher-lua"
)

// TestPortalLongIdentityProjectsAndInteracts exercises public bounds after real scroll consumption.
func TestPortalLongIdentityProjectsAndInteracts(t *testing.T) {
	fixture := portalProjectionFixture(t)
	owner := strings.Repeat("p", 121)

	field := projectedPortal(t, fixture, owner)
	if len(field.ID) > 128 {
		t.Fatalf("public portal identity is too long: %d", len(field.ID))
	}

	if err := restoreLuaSuiteCheckpoint(t, fixture); err != nil {
		t.Fatal(err)
	}

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}

	if restored := projectedPortal(t, fixture, owner); restored.ID != field.ID {
		t.Fatalf("restored portal identity = %q, want %q", restored.ID, field.ID)
	}

	submitPortalInteraction(t, fixture, owner, field.ID, 1)

	town := projectedPortal(t, fixture, owner)
	if town.ID == field.ID || town.LevelID != 1 {
		t.Fatalf("town endpoint = %#v, want a distinct reciprocal portal in town", town)
	}

	submitPortalInteraction(t, fixture, owner, town.ID, 2)

	view := projectedPortalView(t, fixture, owner)
	for _, entity := range view.World.Entities {
		if entity.Label == "Town Portal" {
			t.Fatalf("return retained a selectable portal: %#v", entity)
		}
	}

	if len(view.Private.Items.Items) != 1 || view.Private.Items.Items[0].ID != strings.Repeat("s", 127)+"b" {
		t.Fatal("portal use did not consume exactly the selected maximum-length item")
	}

	// Host collision handles are rebound after restore, not checkpointed Lua facts.
	if err := runLuaAction(t, fixture, portalProjectionActions(t, fixture).RawGetInt(3).(*lua.LTable)); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{"item_id": strings.Repeat("s", 127) + "b"})
	if err != nil {
		t.Fatal(err)
	}

	if err := fixture.session.Submit(simulation.Command{
		Tick: fixture.engine.Tick() + 1, Sequence: 3, Player: owner,
		Authority: simulation.AuthorityPlayer, Kind: "item.use", Payload: payload,
	}); err != nil {
		t.Fatal(err)
	}

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}

	if replacement := projectedPortal(t, fixture, owner); replacement.ID == field.ID || replacement.ID == town.ID {
		t.Fatal("scroll identities differing only at the last valid byte reused a public portal identity")
	}
}

// portalProjectionFixture shares checked-in Lua setup with policy acceptance instead of embedding Lua in Go.
func portalProjectionFixture(t *testing.T) *luaSuiteFixture {
	t.Helper()

	source := content.D2Legacy()
	metadata := newLuaSuiteFixture(t, source)
	config := luaSuiteConfig{seed: 42, records: fixtureRecords{}}
	path := "lua/d2legacy/commands/item_use_projection_test.lua"

	if err := metadata.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
		suite, err := loadLuaSuite(state, source, path)
		if err != nil {
			return err
		}

		return config.read(suite)
	}); err != nil {
		t.Fatal(err)
	}

	fixture := newLuaSuiteFixture(t, source, config)

	actions := portalProjectionActions(t, fixture)
	for index := 1; index <= actions.Len(); index++ {
		if err := runLuaAction(t, fixture, actions.RawGetInt(index).(*lua.LTable)); err != nil {
			t.Fatal(err)
		}
	}

	return fixture
}

func portalProjectionActions(t *testing.T, fixture *luaSuiteFixture) *lua.LTable {
	t.Helper()

	var actions *lua.LTable

	if err := fixture.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
		suite, err := loadLuaSuite(state, content.D2Legacy(), "lua/d2legacy/commands/item_use_projection_test.lua")
		if err != nil {
			return err
		}

		actions = suite.RawGetString("tests").(*lua.LTable).
			RawGetString("long_item_and_owner_labels_create_a_pair").(*lua.LTable)

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	return actions
}

// projectedPortalView crosses both canonical projection and decoded-client validation.
func projectedPortalView(t *testing.T, fixture *luaSuiteFixture, owner string) playeradapter.ClientView {
	t.Helper()

	replay, err := fixture.session.Replay()
	if err != nil {
		t.Fatal(err)
	}

	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	payload, err := playeradapter.ProjectClientView(owner, checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	var view playeradapter.ClientView
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}

	if err := playeradapter.ValidateClientView(view, checkpoint.Tick); err != nil {
		t.Fatal(err)
	}

	return view
}

func projectedPortal(t *testing.T, fixture *luaSuiteFixture, owner string) playeradapter.WorldEntity {
	t.Helper()

	for _, entity := range projectedPortalView(t, fixture, owner).World.Entities {
		if entity.Label == "Town Portal" {
			return entity
		}
	}

	t.Fatal("no selectable portal in the current player's public projection")

	return playeradapter.WorldEntity{}
}

// submitPortalInteraction uses the projected public identity, not the internal endpoint allocation.
func submitPortalInteraction(t *testing.T, fixture *luaSuiteFixture, owner, target string, sequence uint64) {
	t.Helper()

	payload, err := json.Marshal(map[string]any{"target": target})
	if err != nil {
		t.Fatal(err)
	}

	if err := fixture.session.Submit(simulation.Command{
		Tick: fixture.engine.Tick() + 1, Sequence: sequence, Player: owner,
		Authority: simulation.AuthorityPlayer, Kind: "interaction.open", Payload: payload,
	}); err != nil {
		t.Fatal(err)
	}

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}
}
