package player

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestProjectPartyViewSelectsOnlyAuthenticatedOwnerProjection(t *testing.T) {
	fields := []gameecs.FieldSnapshot{
		{Name: "schema_version", Kind: akara.FieldInt64},
		{Name: "revision", Kind: akara.FieldInt64},
		{Name: "party_id", Kind: akara.FieldString},
		{Name: "roster_count", Kind: akara.FieldInt64},
		{Name: "player_1", Kind: akara.FieldString},
		{Name: "name_1", Kind: akara.FieldString},
		{Name: "class_1", Kind: akara.FieldString},
		{Name: "level_1", Kind: akara.FieldInt64},
		{Name: "relationship_1", Kind: akara.FieldString},
	}
	row := func(entity uint64, revision int64, partyID, playerID, name string) gameecs.InstanceSnapshot {
		version, count, level := int64(PartyViewVersion), int64(1), int64(7)
		class, relationship := "Amazon", "self"
		return gameecs.InstanceSnapshot{Entity: entity, Values: []gameecs.ValueSnapshot{
			{Int: &version}, {Int: &revision}, {String: &partyID}, {Int: &count},
			{String: &playerID}, {String: &name}, {String: &class}, {Int: &level}, {String: &relationship},
		}}
	}
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: 11, Components: []gameecs.ComponentSnapshot{
		stringsComponent("d2legacy.player.identity", []string{"player"}, []any{uint64(1), "alice"}, []any{uint64(2), "bob"}),
		{Name: "d2legacy.player.party_view", Version: 1, Fields: fields, Instances: []gameecs.InstanceSnapshot{
			row(1, 3, "", "alice", "Alice"), row(2, 9, "party:secret", "bob", "Bob Secret"),
		}},
	}}
	view, err := ProjectPartyView("alice", simulation.Checkpoint{Tick: 11, Snapshot: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(view)
	if view.Revision != 3 || len(view.Roster) != 1 || view.Roster[0].PlayerID != "alice" ||
		strings.Contains(string(encoded), "party:secret") || strings.Contains(string(encoded), "Bob Secret") {
		t.Fatalf("owner party projection leaked another player view: %s", encoded)
	}
}
