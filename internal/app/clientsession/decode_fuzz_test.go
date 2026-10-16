package clientsession

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// TestDecodeViewRejectsUnknownFieldsAndOversizedCollections protects strict schema and memory bounds.
func TestDecodeViewRejectsUnknownFieldsAndOversizedCollections(t *testing.T) {
	view := validNetworkView(9)

	payload, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}

	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}

	object["future_unreviewed_state"] = true

	unknown, _ := json.Marshal(object)
	if _, err := decodeView(gameserver.Snapshot{Tick: 9, Payload: unknown}); err == nil {
		t.Fatal("unknown projection field was accepted")
	}

	view.World.Entities = make([]playeradapter.WorldEntity, playeradapter.MaxWorldViewEntities+1)

	oversized, _ := json.Marshal(view)
	if _, err := decodeView(gameserver.Snapshot{Tick: 9, Payload: oversized}); err == nil {
		t.Fatal("oversized world projection was accepted")
	}
}

// TestDecodeViewVersionsMissileSoundOwners preserves optional ownership within the current schema.
func TestDecodeViewVersionsMissileSoundOwners(t *testing.T) {
	view := validNetworkView(9)
	view.World.Missiles = []playeradapter.WorldMissile{{
		ID: "missile:1", Kind: "projectile", MissileID: "fireball", SoundOwnerID: "projectile:1",
		Act: 1, LevelID: 1, DCC: "fireball.dcc", LogicalDirection: -1, Directions: 16, FramesPerSecond: 16,
	}}

	view.Events.Events = []playeradapter.SemanticEvent{{
		ID: 2, Type: "effect", Tick: 9, Act: 1, LevelID: 1,
		Effect: &playeradapter.SemanticEffectCue{Kind: "missile_spawn", Sound: "fireball", SoundOwnerID: "projectile:1"},
	}}

	for _, owner := range []string{"projectile:1", ""} {
		view.World.Missiles[0].SoundOwnerID = owner
		view.Events.Events[0].Effect.SoundOwnerID = owner

		payload, err := json.Marshal(view)
		if err != nil {
			t.Fatal(err)
		}

		decoded, err := decodeView(gameserver.Snapshot{Tick: 9, Payload: payload})
		if err != nil {
			t.Fatalf("current schema with owner %q: %v", owner, err)
		}

		if decoded.World.Missiles[0].SoundOwnerID != owner || decoded.Events.Events[0].Effect.SoundOwnerID != owner {
			t.Fatal("sound ownership changed during decoding")
		}
	}

	for name, mutate := range map[string]func(*playeradapter.ClientView){
		"previous envelope": func(v *playeradapter.ClientView) { v.Version-- },
		"previous world":    func(v *playeradapter.ClientView) { v.World.Version-- },
		"previous events":   func(v *playeradapter.ClientView) { v.Events.Version-- },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := view
			mutate(&candidate)

			payload, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := decodeView(gameserver.Snapshot{Tick: 9, Payload: payload}); err == nil {
				t.Fatal("obsolete projection version accepted")
			}
		})
	}
}

// FuzzDecodeClientView requires arbitrary payloads to fail safely or produce a validated bounded view.
func FuzzDecodeClientView(f *testing.F) {
	valid, _ := json.Marshal(validNetworkView(3))
	f.Add(uint64(3), valid)
	f.Add(uint64(1), []byte(`{"version":4}`))
	f.Add(uint64(0), []byte{0xff, 0x00, '{'})
	f.Fuzz(func(t *testing.T, tick uint64, payload []byte) {
		view, err := decodeView(gameserver.Snapshot{Tick: tick, Payload: payload})
		if err != nil {
			return
		}

		if err := playeradapter.ValidateClientView(view, tick); err != nil {
			t.Fatalf("decoder accepted invalid projection: %v", err)
		}
	})
}

// validNetworkView supplies the smallest complete wire projection used by decode seeds.
func validNetworkView(tick uint64) playeradapter.ClientView {
	return playeradapter.ClientView{
		Version: playeradapter.ClientViewVersion,
		Tick:    tick,
		HUD: playeradapter.HUD{
			Version: playeradapter.HUDVersion,
			Tick:    tick,
			Player: playeradapter.HUDIdentity{
				PlayerID: "player", CharacterID: "character", Name: "Hero", Class: "Amazon",
			},
			Vitals: playeradapter.HUDVitals{Health: 10, MaxHealth: 10, Mana: 5, MaxMana: 5},
			Belt:   playeradapter.HUDBelt{Slots: []string{}},
		},
		World: playeradapter.WorldView{
			Version:  playeradapter.WorldViewVersion,
			Tick:     tick,
			Entities: []playeradapter.WorldEntity{},
		},
		Private: playeradapter.PrivateView{
			Version: playeradapter.PrivateViewVersion,
			Tick:    tick,
			Items:   playeradapter.ItemView{Items: []playeradapter.ItemEntityView{}},
		},
		Party: playeradapter.PartyView{
			Version: playeradapter.PartyViewVersion,
			Tick:    tick,
			Roster: []playeradapter.PartyRosterEntry{{
				PlayerID: "player", Name: "Hero", Class: "Amazon", Level: 1, Relationship: "self",
			}},
		},
		Events: playeradapter.EventView{
			Version: playeradapter.EventViewVersion, Tick: tick,
			Events: []playeradapter.SemanticEvent{},
		},
	}
}
