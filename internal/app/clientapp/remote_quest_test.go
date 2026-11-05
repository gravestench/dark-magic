package clientapp

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

func TestConnectedQuestCorrectionAndRemovalPreserveOwnerHero(t *testing.T) {
	engine := gameecs.New()

	t.Cleanup(func() { _ = engine.Close() })
	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	view := connectedProjectionFixture()

	view.Private.Quest = &playeradapter.QuestView{
		QuestID: 1, Stage: 1, Remaining: 2, UnsettledRooms: 1, PopulationAvailable: true,
	}
	if err := app.installRemoteProjection(view.HUD, view.Private, view.Party, true); err != nil {
		t.Fatal(err)
	}

	hero, err := authenticatedHero(engine.World(), view.HUD.Player.PlayerID)
	if err != nil {
		t.Fatal(err)
	}

	store, _ := akara.GetDynamicStore(engine.World(), "d2legacy.quest.progress")
	if store.Len() != 1 {
		t.Fatal("owner quest not installed")
	}

	q, ok := store.Get(hero)
	if !ok {
		t.Fatal("quest attached to wrong owner")
	}

	if remaining, _ := q.Get("remaining"); remaining != int64(2) {
		t.Fatal("count lost", remaining)
	}

	before, err := privateProjectionFingerprint(view.HUD.Skills.Learned, view.Private)
	if err != nil {
		t.Fatal(err)
	}

	view.Private.Quest = &playeradapter.QuestView{
		QuestID: 1, Stage: 5, Completed: true, Rewarded: true, PopulationAvailable: true,
	}

	after, err := privateProjectionFingerprint(view.HUD.Skills.Learned, view.Private)
	if err != nil || before == after {
		t.Fatal("quest correction omitted from fingerprint", err)
	}

	if err := app.installRemoteProjection(view.HUD, view.Private, view.Party, true); err != nil {
		t.Fatal(err)
	}

	q, _ = store.Get(hero)
	if rewarded, _ := q.Get("rewarded"); rewarded != true {
		t.Fatal("reward correction lost")
	}

	view.Private.Quest = nil
	if err := app.installRemoteProjection(view.HUD, view.Private, view.Party, true); err != nil {
		t.Fatal(err)
	}

	if store.Len() != 0 {
		t.Fatal("absent quest retained")
	}

	current, err := authenticatedHero(engine.World(), view.HUD.Player.PlayerID)
	if err != nil || current != hero {
		t.Fatal("private replacement destroyed owner hero", current, hero, err)
	}
}
