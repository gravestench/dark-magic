package player

import (
	"math"
	"strings"
	"testing"
)

// TestValidateClientViewRejectsUnboundedOrInvalidNetworkState exercises the
// combined envelope as an untrusted network boundary.
func TestValidateClientViewRejectsUnboundedOrInvalidNetworkState(t *testing.T) {
	valid := validClientView(7)
	if err := ValidateClientView(valid, 7); err != nil {
		t.Fatalf("valid view: %v", err)
	}
	valid.World.States = []WorldState{{TargetID: "player:player", StateID: "poison", Aura: false}}
	if err := ValidateClientView(valid, 7); err != nil {
		t.Fatalf("valid poison state: %v", err)
	}

	for _, kind := range []string{"seam", "warp"} {
		candidate := validClientView(7)
		candidate.HUD.Location = HUDLocation{Act: 1, LevelID: 2, TransitionSequence: 1, TransitionSource: 1, TransitionKind: kind}
		if err := ValidateClientView(candidate, 7); err != nil {
			t.Fatalf("valid relocation %s: %v", kind, err)
		}
	}
	tests := map[string]func(*ClientView){
		"negative relocation sequence": func(view *ClientView) { view.HUD.Location.TransitionSequence = -1 },
		"unknown relocation kind": func(view *ClientView) {
			view.HUD.Location.TransitionSequence = 1
			view.HUD.Location.TransitionSource = 1
			view.HUD.Location.TransitionKind = "other"
		},
		"relocation missing source": func(view *ClientView) {
			view.HUD.Location.TransitionSequence = 1
			view.HUD.Location.TransitionKind = "seam"
		},
		"reason without relocation":       func(view *ClientView) { view.HUD.Location.TransitionKind = "warp" },
		"pre-commerce private schema":     func(view *ClientView) { view.Private.Version = 1 },
		"pre-commerce client schema":      func(view *ClientView) { view.Version = 14 },
		"previous combined client schema": func(view *ClientView) { view.Version = 15 },
		"previous client version":         func(view *ClientView) { view.Version = 13 },
		"previous world version":          func(view *ClientView) { view.World.Version = 6 },
		"oversized ground art path": func(view *ClientView) {
			view.World.Entities[0].Kind = "item"
			view.World.Entities[0].WorldDC6 = strings.Repeat("x", maxWorldAssetBytes+1)
		},
		"wrong nested tick": func(view *ClientView) { view.Private.Tick++ },
		"non-finite owner":  func(view *ClientView) { view.HUD.Position.X = math.NaN() },
		"excess entities": func(view *ClientView) {
			view.World.Entities = make([]WorldEntity, MaxWorldViewEntities+1)
		},
		"duplicate world ID": func(view *ClientView) {
			view.World.Entities = append(view.World.Entities, view.World.Entities[0])
		},
		"excess persistent states": func(view *ClientView) {
			view.World.States = make([]WorldState, MaxWorldViewStates+1)
		},
		"invalid persistent state period": func(view *ClientView) {
			view.World.States = []WorldState{{TargetID: "player:player", StateID: "might", Aura: true}}
		},
		"poison with aura period": func(view *ClientView) {
			view.World.States = []WorldState{{TargetID: "player:player", StateID: "poison", PeriodTicks: 50}}
		},
		"nonpoison timed state": func(view *ClientView) {
			view.World.States = []WorldState{{TargetID: "player:player", StateID: "frozenarmor"}}
		},
		"duplicate persistent state": func(view *ClientView) {
			state := WorldState{TargetID: "player:player", StateID: "might", Aura: true, PeriodTicks: 50}
			view.World.States = []WorldState{state, state}
		},
		"excess private items": func(view *ClientView) {
			view.Private.Items.Items = make([]ItemEntityView, MaxPrivateItems+1)
		},
		"inconsistent interaction":   func(view *ClientView) { view.Private.Interaction.Active = true },
		"negative dialogue revision": func(view *ClientView) { view.Private.Interaction.DialogueRevision = -1 },
		"negative vendor quote": func(view *ClientView) {
			view.Private.Items.Items = []ItemEntityView{{ID: "offer", Code: "tsc", BuyPrice: -1}}
		},
		"excess dialogue line": func(view *ClientView) {
			view.Private.Interaction.DialogueLine = strings.Repeat("x", maxViewTokenBytes+1)
		},
		"duplicate party player": func(view *ClientView) {
			view.Party.Roster = append(view.Party.Roster, view.Party.Roster[0])
		},
		"unknown party relationship": func(view *ClientView) { view.Party.Roster[0].Relationship = "hostile" },
		"missing party owner":        func(view *ClientView) { view.Party.Roster[0].PlayerID = "other" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := validClientView(7)
			mutate(&candidate)

			if err := ValidateClientView(candidate, 7); err == nil {
				t.Fatal("invalid network projection was accepted")
			}
		})
	}
}

// validClientView returns the smallest internally consistent envelope, letting
// each validation case mutate exactly one invariant.
func validClientView(tick uint64) ClientView {
	return ClientView{
		Version: ClientViewVersion,
		Tick:    tick,
		HUD: HUD{
			Version: HUDVersion,
			Tick:    tick,
			Player:  HUDIdentity{PlayerID: "player", CharacterID: "character", Name: "Hero", Class: "Amazon"},
			Vitals:  HUDVitals{Health: 10, MaxHealth: 10, Mana: 5, MaxMana: 5},
			Belt:    HUDBelt{Slots: []string{}},
		},
		World: WorldView{
			Version: WorldViewVersion,
			Tick:    tick,
			Entities: []WorldEntity{{
				ID: "peer", Kind: "player", Position: HUDPosition{X: 1, Y: 2}, Radius: 1,
			}},
		},
		Private: PrivateView{
			Version: PrivateViewVersion,
			Tick:    tick,
			Items: ItemView{Items: []ItemEntityView{{
				ID: "item", Code: "cap", Width: 1, Height: 1,
			}}},
		},
		Party: PartyView{
			Version: PartyViewVersion,
			Tick:    tick,
			Roster: []PartyRosterEntry{{
				PlayerID: "player", Name: "Hero", Class: "Amazon", Level: 1, Relationship: "self",
			}},
		},
		Events: EventView{Version: EventViewVersion, Tick: tick, Events: []SemanticEvent{}},
	}
}
