package d2legacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

type partyStateFixture struct {
	Revision    int                                 `json:"revision"`
	NextPartyID int                                 `json:"next_party_id"`
	Parties     map[string]partyFixture             `json:"parties"`
	Membership  map[string]string                   `json:"membership"`
	Invites     map[string]map[string]inviteFixture `json:"invites"`
}

type partyFixture struct {
	ID      string   `json:"id"`
	Members []string `json:"members"`
}

type inviteFixture struct {
	Inviter     string `json:"inviter"`
	Target      string `json:"target"`
	CreatedTick int    `json:"created_tick"`
}

func TestPartyAuthorityCommandsRestoreAndDepartureCleanup(t *testing.T) {
	ctx := context.Background()
	authority, engine, session := startPartyFixture(t, nil)
	t.Cleanup(func() {
		_ = authority.Stop(ctx)
		_ = session.Close()
		_ = engine.Close()
	})

	for sequence, player := range []string{"alice", "bob", "carol"} {
		if err := session.Submit(simulation.Command{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
			Sequence: uint64(sequence + 1), Kind: "system.player.enter",
			Payload: generatedPlayerPayload(t, "hero-"+player, player, float64(sequence+1), 1)}); err != nil {
			t.Fatal(err)
		}
	}
	stepPartySession(t, session)

	submitPartyCommand(t, session, 2, "alice", 1, "party.invite", map[string]any{"target": "bob"})
	submitPartyCommand(t, session, 2, "carol", 1, "party.invite", map[string]any{"target": "bob"})
	stepPartySession(t, session)
	state := readPartyState(t, authority)
	if state.Revision != 2 || len(state.Invites["bob"]) != 2 || state.Invites["bob"]["alice"].CreatedTick != 2 {
		t.Fatalf("party invitations = %#v", state)
	}

	submitPartyCommand(t, session, 3, "carol", 2, "party.cancel_invite", map[string]any{"target": "bob"})
	stepPartySession(t, session)
	state = readPartyState(t, authority)
	if state.Revision != 3 || len(state.Invites["bob"]) != 1 || state.Invites["bob"]["alice"].Inviter != "alice" {
		t.Fatalf("party invitation cancellation = %#v", state)
	}

	submitPartyCommand(t, session, 4, "bob", 1, "party.accept", map[string]any{"inviter": "alice"})
	stepPartySession(t, session)
	state = readPartyState(t, authority)
	party := state.Parties["party:1"]
	if state.Revision != 4 || state.NextPartyID != 2 || state.Membership["alice"] != party.ID ||
		state.Membership["bob"] != party.ID || len(party.Members) != 2 || party.Members[0] != "alice" ||
		party.Members[1] != "bob" || state.Invites["bob"] != nil {
		t.Fatalf("accepted party state = %#v", state)
	}
	additional, err := modruntime.Call(ctx, authority.Runtime, "d2legacy.policy.party", "additional_living_members_in_same_level", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if additional != float64(1) {
		t.Fatalf("additional living same-level party members = %v, want 1", additional)
	}

	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	aliceView, err := playeradapter.ProjectPartyView("alice", checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	if aliceView.Version != playeradapter.PartyViewVersion || aliceView.Revision != 4 || aliceView.PartyID != "party:1" ||
		len(aliceView.Roster) != 3 || aliceView.Roster[0].PlayerID != "alice" || aliceView.Roster[0].Relationship != "self" ||
		aliceView.Roster[1].PlayerID != "bob" || aliceView.Roster[1].Relationship != "party" ||
		aliceView.Roster[2].PlayerID != "carol" || aliceView.Roster[2].Relationship != "available" {
		t.Fatalf("alice party projection = %#v", aliceView)
	}
	submitPartyCommand(t, session, 5, "bob", 2, "party.leave", map[string]any{})
	stepPartySession(t, session)
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	state = readPartyState(t, authority)
	if state.Revision != 5 || len(state.Parties) != 0 || len(state.Membership) != 0 {
		t.Fatalf("party leave did not dissolve two-member party: %#v", state)
	}

	restored, restoredEngine, restoredSession := startPartyFixture(t, &checkpoint)
	t.Cleanup(func() {
		_ = restored.Stop(ctx)
		_ = restoredSession.Close()
		_ = restoredEngine.Close()
	})
	submitPartyCommand(t, restoredSession, 5, "bob", 1, "party.leave", map[string]any{})
	stepPartySession(t, restoredSession)
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	continued := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("restored party transition checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
	if got := readPartyState(t, restored); got.Revision != 5 || len(got.Parties) != 0 {
		t.Fatalf("restored party leave state = %#v", got)
	}
	submitPartyCommand(t, session, 6, "alice", 2, "party.invite", map[string]any{"target": "carol"})
	stepPartySession(t, session)
	submitPartyCommand(t, session, 7, "carol", 3, "party.accept", map[string]any{"inviter": "alice"})
	stepPartySession(t, session)
	if err := session.Submit(simulation.Command{Tick: 8, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 4, Kind: "system.player.leave", Payload: json.RawMessage(`{"player":"carol"}`)}); err != nil {
		t.Fatal(err)
	}
	stepPartySession(t, session)
	state = readPartyState(t, authority)
	if state.Revision != 8 || len(state.Parties) != 0 || len(state.Membership) != 0 || len(state.Invites) != 0 {
		t.Fatalf("game departure did not clean party authority: %#v", state)
	}
}

func startPartyFixture(t *testing.T, checkpoint *simulation.Checkpoint) (*Authority, *gameecs.Engine, *gamesession.Session) {
	t.Helper()
	var engine *gameecs.Engine
	var restore []simulation.ParticipantState
	if checkpoint != nil {
		var err error
		engine, err = gameecs.RestoreSnapshot(*checkpoint.Snapshot)
		if err != nil {
			t.Fatal(err)
		}
		restore = checkpoint.Participants
	} else {
		engine = gameecs.New()
	}
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), generatedHostileRecords(), engine, session,
		Config{Seed: 314, Restore: restore})
	if err != nil {
		_ = session.Close()
		_ = engine.Close()
		t.Fatal(err)
	}
	return authority, engine, session
}

func submitPartyCommand(t *testing.T, session *gamesession.Session, tick uint64, player string, sequence uint64, kind string, payload map[string]any) {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(simulation.Command{Tick: tick, Player: player, Authority: simulation.AuthorityPlayer,
		Sequence: sequence, Kind: kind, Payload: encoded}); err != nil {
		t.Fatal(err)
	}
}

func stepPartySession(t *testing.T, session *gamesession.Session) {
	t.Helper()
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
}

func readPartyState(t *testing.T, authority *Authority) partyStateFixture {
	t.Helper()
	registered, found := authority.State.Read("d2legacy.party")
	if !found {
		t.Fatal("party authority state is missing")
	}
	var value partyStateFixture
	if err := json.Unmarshal(registered.Data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
