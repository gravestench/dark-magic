package player

import (
	"math"
	"testing"
)

func TestValidateClientViewRejectsUnboundedOrInvalidNetworkState(t *testing.T) {
	valid := validClientView(7)
	if err := ValidateClientView(valid, 7); err != nil {
		t.Fatalf("valid view: %v", err)
	}

	tests := map[string]func(*ClientView){
		"wrong nested tick": func(view *ClientView) { view.Private.Tick++ },
		"non-finite owner":  func(view *ClientView) { view.HUD.Position.X = math.NaN() },
		"excess entities": func(view *ClientView) {
			view.World.Entities = make([]WorldEntity, MaxWorldViewEntities+1)
		},
		"duplicate world ID": func(view *ClientView) {
			view.World.Entities = append(view.World.Entities, view.World.Entities[0])
		},
		"excess private items": func(view *ClientView) {
			view.Private.Items.Items = make([]ItemEntityView, MaxPrivateItems+1)
		},
		"inconsistent interaction": func(view *ClientView) { view.Private.Interaction.Active = true },
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
