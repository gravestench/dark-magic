package d2legacy

import (
	"context"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestDecodedQuestRewardAndDifficultyRestoreIdentically(t *testing.T) {
	ctx := context.Background()
	initial := map[string]any{
		"d2legacy.items": map[string]any{"owner": "alice", "inventory_width": 4.0,
			"inventory_height": 4.0, "belt_capacity": 4.0, "items": []any{}},
		"d2legacy.interactions": map[string]any{"owner": "alice", "targets": []any{
			map[string]any{"id": "npc:akara", "npc": "akara", "x": 10.0, "y": 12.0, "radius": 3.0},
		}},
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, engine, session,
		Config{Seed: 77, InitialData: initial})
	if err != nil {
		t.Fatal(err)
	}
	enter := []byte(`{"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",` +
		`"level":1,"experience":0,"dexterity":20,"defense":0,"health":50,"max_health":50,` +
		`"mana":20,"max_mana":20,"expansion":true,"hardcore":false,"cof":"","palette":"units",` +
		`"direction":0,"mode":"NU","x":10,"y":12,"world_width":100,"world_height":80,` +
		`"act":1,"level_id":1,"skills":[]}`)
	commands := []simulation.Command{
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.player.enter", Payload: enter},
		{Tick: 2, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: 1, Kind: "interaction.open", Payload: []byte(`{"target":"npc:akara"}`)},
		{Tick: 3, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 2, Kind: "system.quest.complete", Payload: []byte(`{"player":"alice","quest_id":1}`)},
	}
	for _, command := range commands {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
		stepSession(t, session, 1)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	continuation := []simulation.Command{
		{Tick: 4, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: 2, Kind: "quest.claim_reward", Payload: []byte(`{}`)},
		{Tick: 5, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 3, Kind: "system.difficulty.advance", Payload: []byte(`{"player":"alice","difficulty":1,"completed_act":5}`)},
	}
	for _, command := range continuation {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	stepSession(t, session, 2)
	originalReplay, _ := session.Replay()
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	assertQuestRewardAndDifficulty(t, engine)
	_ = authority.Stop(ctx)
	_ = session.Close()
	_ = engine.Close()

	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredEngine.Close()
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, restoredEngine, restoredSession,
		Config{Seed: 77, InitialData: initial, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	restoredContinuation := append([]simulation.Command(nil), continuation...)
	restoredContinuation[0].Sequence = 1
	restoredContinuation[1].Sequence = 1
	for _, command := range restoredContinuation {
		if err := restoredSession.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	stepSession(t, restoredSession, 2)
	restoredReplay, _ := restoredSession.Replay()
	continued := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("restored quest/difficulty checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
	assertQuestRewardAndDifficulty(t, restoredEngine)
}

func assertQuestRewardAndDifficulty(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	progress, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.progress")
	value, _ := progress.Get(progress.Entities()[0])
	points, _ := value.Get("unspent_skill_points")
	if points != int64(1) {
		t.Fatalf("rewarded skill points = %v, want 1", points)
	}
	quests, _ := akara.GetDynamicStore(engine.World(), "d2legacy.quest.progress")
	quest, _ := quests.Get(quests.Entities()[0])
	rewarded, _ := quest.Get("rewarded")
	if rewarded != true {
		t.Fatal("quest reward was not committed")
	}
	difficulties, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.difficulty")
	difficulty, _ := difficulties.Get(difficulties.Entities()[0])
	current, _ := difficulty.Get("current")
	if current != int64(1) {
		t.Fatalf("difficulty = %v, want Nightmare", current)
	}
}
