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
