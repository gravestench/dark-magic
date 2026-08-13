package serverapp

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestAdmitSelectedProfileQueuesSystemEntryWithoutRealmAuthority(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	if err := session.Register(playeradapter.EnterCommand, gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
		Allowed:  []simulation.Authority{simulation.AuthoritySystem},
	}); err != nil {
		t.Fatal(err)
	}
	store := d2save.New(d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 3})
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := d2save.WriteProfileFile(path, store.Profile()); err != nil {
		t.Fatal(err)
	}
	destination, _ := playeradapter.NewDestination(10, 20, 100, 100, 1, 40)
	if err := AdmitSelectedProfile(&gameserver.Host{Session: session}, ProfileAdmission{Path: path, PlayerID: "owner", Destination: destination}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Commands) != 1 || replay.Commands[0].Authority != simulation.AuthoritySystem || replay.Commands[0].Player != "self-host:profile" {
		t.Fatalf("profile command = %#v", replay.Commands)
	}
	var entry playeradapter.Entry
	if err := json.Unmarshal(replay.Commands[0].Payload, &entry); err != nil || entry.CharacterID != "hero" || entry.Player != "owner" || entry.LevelID != 40 {
		t.Fatalf("profile entry = %#v error=%v", entry, err)
	}
	installProfileCheckpoint(t, engine)
	if err := PersistSelectedProfile(&gameserver.Host{Session: session}, ProfileAdmission{Path: path, PlayerID: "owner"}); err != nil {
		t.Fatal(err)
	}
	persisted, err := d2save.ReadProfileFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Characters[0].Level != 4 || persisted.Characters[0].Stats.Health != 22 {
		t.Fatalf("persisted canonical character = %#v", persisted.Characters[0])
	}
}

func installProfileCheckpoint(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	register := func(name string, fields []akara.Field) *akara.DynamicStore {
		store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: name, Version: 1, Fields: fields})
		if err != nil {
			t.Fatal(err)
		}
		return store
	}
	entity := engine.World().MustCreateEntity()
	stores := map[*akara.DynamicStore]map[string]any{
		register("d2legacy.player.identity", []akara.Field{{Name: "character_id", Kind: akara.FieldString}, {Name: "player", Kind: akara.FieldString}, {Name: "name", Kind: akara.FieldString}, {Name: "class", Kind: akara.FieldString}}): {"character_id": "hero", "player": "owner", "name": "Hero", "class": "Amazon"},
		register("d2legacy.player.vitals", []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "mana", Kind: akara.FieldInt64}, {Name: "max_mana", Kind: akara.FieldInt64}}):      {"health": int64(22), "max_health": int64(30), "mana": int64(10), "max_mana": int64(15)},
		register("d2legacy.player.progress", []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}, {Name: "unspent_skill_points", Kind: akara.FieldInt64}}):                                 {"level": int64(4), "experience": int64(200), "unspent_skill_points": int64(0)},
		register("d2legacy.player.combat_stats", []akara.Field{{Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}}):                                                                                {"attack_rating": int64(12), "defense": int64(5)},
		register("d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}):                                                                                                   {"x": 10.0, "y": 20.0},
		register("d2legacy.world.location", []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}):                                                                                              {"act": int64(1), "level_id": int64(40)},
	}
	for store, values := range stores {
		if _, err := store.Set(entity, values); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAdmitSelectedProfileRejectsUnselectedRoster(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := d2save.WriteProfileFile(path, d2save.New(d2save.Character{ID: "hero"}).Profile()); err != nil {
		t.Fatal(err)
	}
	if err := AdmitSelectedProfile(&gameserver.Host{Session: &gamesession.Session{}}, ProfileAdmission{Path: path, PlayerID: "owner"}); err == nil {
		t.Fatal("unselected profile was admitted")
	}
}
