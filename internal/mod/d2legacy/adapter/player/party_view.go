package player

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	PartyViewVersion   uint32 = 1
	MaxPartyViewRoster        = 8
)

var ErrPartyView = errors.New("party view: authenticated projection is absent")

// PartyView is the bounded owner-scoped presentation projection derived by
// d2legacy. It intentionally omits other parties' IDs and membership lists,
// invitation timestamps, and every raw authority-state field.
type PartyView struct {
	Version  uint32             `json:"version"`
	Tick     uint64             `json:"tick"`
	Revision uint64             `json:"revision"`
	PartyID  string             `json:"party_id,omitempty"`
	Roster   []PartyRosterEntry `json:"roster"`
}

type PartyRosterEntry struct {
	PlayerID     string `json:"player_id"`
	Name         string `json:"name"`
	Class        string `json:"class"`
	Level        int64  `json:"level"`
	Relationship string `json:"relationship"`
}

// ProjectPartyView selects only the authenticated player's materialized party
// view from the canonical checkpoint. The Lua state machine remains authority.
func ProjectPartyView(playerID string, checkpoint simulation.Checkpoint) (PartyView, error) {
	view := PartyView{Version: PartyViewVersion, Tick: checkpoint.Tick, Roster: []PartyRosterEntry{}}
	if checkpoint.Snapshot == nil || strings.TrimSpace(playerID) == "" {
		return view, ErrPartyView
	}
	identities, found := findComponent(*checkpoint.Snapshot, "d2legacy.player.identity")
	if !found {
		return view, ErrPartyView
	}
	entity, _, found := findString(identities, "player", playerID)
	if !found {
		return view, ErrPartyView
	}
	component, found := findComponent(*checkpoint.Snapshot, "d2legacy.player.party_view")
	if !found {
		return view, ErrPartyView
	}
	fields, found := findInstance(component, entity)
	if !found || intField(fields, "schema_version") != int64(PartyViewVersion) {
		return view, ErrPartyView
	}
	view.Revision = uint64(max(0, intField(fields, "revision")))
	view.PartyID = stringField(fields, "party_id")
	count := intField(fields, "roster_count")
	if count < 1 || count > MaxPartyViewRoster {
		return view, ErrPartyView
	}
	for slot := int64(1); slot <= count; slot++ {
		suffix := fmt.Sprintf("_%d", slot)
		view.Roster = append(view.Roster, PartyRosterEntry{
			PlayerID:     stringField(fields, "player"+suffix),
			Name:         stringField(fields, "name"+suffix),
			Class:        stringField(fields, "class"+suffix),
			Level:        intField(fields, "level"+suffix),
			Relationship: stringField(fields, "relationship"+suffix),
		})
	}
	return view, nil
}
