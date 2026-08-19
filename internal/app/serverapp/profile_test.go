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

type selectedProfileFixture struct {
	engine  *gameecs.Engine
	session *gamesession.Session
	config  ProfileAdmission
}

type profileStoreFixture struct {
	name   string
	fields []akara.Field
	values map[string]any
}

// TestAdmitSelectedProfileQueuesSystemEntryWithoutRealmAuthority exercises the
// complete local-profile lifecycle: host-owned admission followed by canonical
// persistence into the same selected roster slot.
func TestAdmitSelectedProfileQueuesSystemEntryWithoutRealmAuthority(t *testing.T) {
	fixture := newSelectedProfileFixture(t)
	host := &gameserver.Host{Session: fixture.session}

	if err := AdmitSelectedProfile(host, fixture.config); err != nil {
		t.Fatal(err)
	}

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}

	assertQueuedProfileEntry(t, fixture.session)

	installProfileCheckpoint(t, fixture.engine)

	if err := PersistSelectedProfile(host, fixture.config); err != nil {
		t.Fatal(err)
	}

	assertPersistedCanonicalCharacter(t, fixture.config.Path)
}

// newSelectedProfileFixture owns the ECS, session, and on-disk profile for one
// test, ensuring failures cannot leak runtime state or reuse mutable fixtures.
func newSelectedProfileFixture(t *testing.T) selectedProfileFixture {
	t.Helper()

	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// Close the session before its engine because the session owns work that may
	// still reference ECS storage.
	t.Cleanup(func() {
		_ = session.Close()
		_ = engine.Close()
	})

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

	destination, err := playeradapter.NewDestination(10, 20, 100, 100, 1, 40)
	if err != nil {
		t.Fatal(err)
	}

	return selectedProfileFixture{
		engine:  engine,
		session: session,
		config: ProfileAdmission{
			Path:        path,
			PlayerID:    "owner",
			Destination: destination,
		},
	}
}

// assertQueuedProfileEntry verifies both the system-authority envelope and the
// host-supplied identity and destination encoded in its payload.
func assertQueuedProfileEntry(t *testing.T, session *gamesession.Session) {
	t.Helper()

	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}

	if len(replay.Commands) != 1 || replay.Commands[0].Authority != simulation.AuthoritySystem ||
		replay.Commands[0].Player != "self-host:profile" {
		t.Fatalf("profile command = %#v", replay.Commands)
	}

	var entry playeradapter.Entry

	err = json.Unmarshal(replay.Commands[0].Payload, &entry)
	if err != nil || entry.CharacterID != "hero" || entry.Player != "owner" || entry.LevelID != 40 {
		t.Fatalf("profile entry = %#v error=%v", entry, err)
	}
}

// assertPersistedCanonicalCharacter confirms projection updates canonical
// fields without making the test depend on the profile encoder's formatting.
func assertPersistedCanonicalCharacter(t *testing.T, path string) {
	t.Helper()

	persisted, err := d2save.ReadProfileFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if persisted.Characters[0].Level != 4 || persisted.Characters[0].Stats.Health != 22 {
		t.Fatalf("persisted canonical character = %#v", persisted.Characters[0])
	}
}

// installProfileCheckpoint creates the minimum canonical component set needed
// by character projection; each schema remains explicit so fixture drift is
// visible during review.
func installProfileCheckpoint(t *testing.T, engine *gameecs.Engine) {
	t.Helper()

	entity := engine.World().MustCreateEntity()
	for _, fixture := range profileCheckpointFixtures() {
		store, err := akara.RegisterSchema(engine.World(), akara.Schema{
			Name:    fixture.name,
			Version: 1,
			Fields:  fixture.fields,
		})
		if err != nil {
			t.Fatal(err)
		}

		if _, err := store.Set(entity, fixture.values); err != nil {
			t.Fatal(err)
		}
	}
}

// profileCheckpointFixtures groups schema and value pairs together so a field
// cannot be added to the projection fixture without its canonical test value.
func profileCheckpointFixtures() []profileStoreFixture {
	return []profileStoreFixture{
		{
			name: "d2legacy.player.identity",
			fields: []akara.Field{
				{Name: "character_id", Kind: akara.FieldString},
				{Name: "player", Kind: akara.FieldString},
				{Name: "name", Kind: akara.FieldString},
				{Name: "class", Kind: akara.FieldString},
			},
			values: map[string]any{
				"character_id": "hero",
				"player":       "owner",
				"name":         "Hero",
				"class":        "Amazon",
			},
		},
		{
			name: "d2legacy.player.vitals",
			fields: []akara.Field{
				{Name: "health", Kind: akara.FieldInt64},
				{Name: "max_health", Kind: akara.FieldInt64},
				{Name: "mana", Kind: akara.FieldInt64},
				{Name: "max_mana", Kind: akara.FieldInt64},
			},
			values: map[string]any{
				"health":     int64(22),
				"max_health": int64(30),
				"mana":       int64(10),
				"max_mana":   int64(15),
			},
		},
		{
			name: "d2legacy.player.progress",
			fields: []akara.Field{
				{Name: "level", Kind: akara.FieldInt64},
				{Name: "experience", Kind: akara.FieldInt64},
				{Name: "unspent_skill_points", Kind: akara.FieldInt64},
			},
			values: map[string]any{
				"level":                int64(4),
				"experience":           int64(200),
				"unspent_skill_points": int64(0),
			},
		},
		{
			name: "d2legacy.player.combat_stats",
			fields: []akara.Field{
				{Name: "attack_rating", Kind: akara.FieldInt64},
				{Name: "defense", Kind: akara.FieldInt64},
			},
			values: map[string]any{
				"attack_rating": int64(12),
				"defense":       int64(5),
			},
		},
		{
			name: "d2legacy.world.position",
			fields: []akara.Field{
				{Name: "x", Kind: akara.FieldFloat64},
				{Name: "y", Kind: akara.FieldFloat64},
			},
			values: map[string]any{
				"x": 10.0,
				"y": 20.0,
			},
		},
		{
			name: "d2legacy.world.location",
			fields: []akara.Field{
				{Name: "act", Kind: akara.FieldInt64},
				{Name: "level_id", Kind: akara.FieldInt64},
			},
			values: map[string]any{
				"act":      int64(1),
				"level_id": int64(40),
			},
		},
	}
}

// TestAdmitSelectedProfileRejectsUnselectedRoster ensures an unselected local
// roster cannot produce an arbitrary or default character admission.
func TestAdmitSelectedProfileRejectsUnselectedRoster(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")

	profile := d2save.New(d2save.Character{ID: "hero"}).Profile()
	if err := d2save.WriteProfileFile(path, profile); err != nil {
		t.Fatal(err)
	}

	host := &gameserver.Host{Session: &gamesession.Session{}}

	config := ProfileAdmission{Path: path, PlayerID: "owner"}
	if err := AdmitSelectedProfile(host, config); err == nil {
		t.Fatal("unselected profile was admitted")
	}
}
