package d2legacy

import (
	"context"
	"errors"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestDifficultyAdvanceRecordsUnlockWithoutChangingSessionDifficulty proves
// completing Act V records a character unlock while the active game keeps the
// difficulty chosen by its immutable rules.
func TestDifficultyAdvanceRecordsUnlockWithoutChangingSessionDifficulty(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()

	defer func() { _ = engine.Close() }()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = session.Close() }()

	authority, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, engine, session,
		Config{Seed: 79, InitialData: map[string]any{
			"d2legacy.game_rules": map[string]any{"difficulty": 0},
		}})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = authority.Stop(ctx) }()

	enter := []byte(`{"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",` +
		`"level":1,"experience":0,"dexterity":20,"vitality":20,"defense":0,"health":50,"max_health":50,` +
		`"mana":20,"max_mana":20,"stamina":84,"max_stamina":84,"expansion":true,"hardcore":false,` +
		`"cof":"","palette":"units",` +
		`"direction":0,"mode":"NU","x":10,"y":12,"world_width":100,"world_height":80,` +
		`"act":1,"level_id":1,"skills":[]}`)
	submitAndStep(t, session, simulation.Command{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: enter})
	submitAndStep(t, session, simulation.Command{Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 2, Kind: "system.difficulty.advance",
		Payload: []byte(`{"player":"alice","difficulty":1,"completed_act":5}`)})

	current, highest := playerDifficulty(t, engine)
	if current != 0 || highest != 0 {
		t.Fatalf("after Normal Act V: current=%d highest_completed=%d, want current=0 highest_completed=0",
			current, highest)
	}

	// A second unlock inside the same Normal game would skip Nightmare entirely.
	if err := session.Submit(simulation.Command{Tick: 3, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 3, Kind: "system.difficulty.advance",
		Payload: []byte(`{"player":"alice","difficulty":2,"completed_act":5}`)}); err != nil {
		t.Fatal(err)
	}

	if err := session.Step(); !errors.Is(err, gamesession.ErrCommandApply) {
		t.Fatalf("Hell unlock from a Normal game error = %v, want %v", err, gamesession.ErrCommandApply)
	}

	current, highest = playerDifficulty(t, engine)
	if current != 0 || highest != 0 {
		t.Fatalf("after rejected skip: current=%d highest_completed=%d, want current=0 highest_completed=0",
			current, highest)
	}
}

// submitAndStep admits one command and advances one tick, failing the test
// when either the admission or its application is rejected.
func submitAndStep(t *testing.T, session *gamesession.Session, command simulation.Command) {
	t.Helper()

	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}

	if err := session.Step(); err != nil {
		t.Fatalf("%s: %v", command.Kind, err)
	}
}

// playerDifficulty reads the single player's active and completed difficulty.
func playerDifficulty(t *testing.T, engine *gameecs.Engine) (current, highest int64) {
	t.Helper()

	store, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.difficulty")
	if len(store.Entities()) != 1 {
		t.Fatalf("difficulty entities = %d, want 1", len(store.Entities()))
	}

	value, _ := store.Get(store.Entities()[0])
	currentValue, _ := value.Get("current")
	highestValue, _ := value.Get("highest_completed")

	return currentValue.(int64), highestValue.(int64)
}
