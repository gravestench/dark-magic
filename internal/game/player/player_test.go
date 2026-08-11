package player

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/persistence"
)

func TestEntryCommandMaterializesAuthoritativePlayerAtomically(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	character := persistence.Character{ID: "amazon-hero", Name: "Hero", Class: "Amazon", Level: 3, Expansion: true, Stats: &persistence.Stats{Experience: 100, Health: 25, MaxHealth: 30, Mana: 12, MaxMana: 15}}
	entry := EntryFromCharacter(character, "player-1", 5, 7, 100, 80)
	entry.Skills = []Skill{{ID: 6, Level: 1, ListRow: 1, LeftAllowed: true, RightAllowed: true}}
	command, err := Command(entry, "server", 1, 1, simulation.AuthoritySystem)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	identity, found := akara.GetDynamicStore(engine.World(), "dm.player.identity")
	if !found || len(identity.Entities()) != 1 {
		t.Fatalf("identity store = %v, %v", identity, found)
	}
	entity := identity.Entities()[0]
	component, _ := identity.Get(entity)
	if player, _ := component.Get("player"); player != "player-1" {
		t.Fatalf("player = %v", player)
	}
	position, _ := akara.GetDynamicStore(engine.World(), "dm.world.position")
	transform, _ := position.Get(entity)
	if x, _ := transform.Get("x"); x != float64(5) {
		t.Fatalf("x = %v", x)
	}
	modes, found := akara.GetDynamicStore(engine.World(), "dm.player.movement_mode")
	if !found {
		t.Fatal("movement mode store was not materialized")
	}
	mode, _ := modes.Get(entity)
	if running, _ := mode.Get("running"); running != false {
		t.Fatalf("initial movement mode = %v, want walking", running)
	}
	appearanceStore, found := akara.GetDynamicStore(engine.World(), "dm.player.appearance")
	if !found {
		t.Fatal("appearance store was not materialized")
	}
	appearance, _ := appearanceStore.Get(entity)
	if token, _ := appearance.Get("token"); token != "AM" {
		t.Fatalf("token = %v, want AM", token)
	}
	animationStore, found := akara.GetDynamicStore(engine.World(), "dm.player.animation")
	if !found {
		t.Fatal("animation store was not materialized")
	}
	animationState, _ := animationStore.Get(entity)
	if animation, _ := animationState.Get("mode"); animation != "NU" {
		t.Fatalf("mode = %v, want NU", animation)
	}
	if weaponClass, _ := appearance.Get("weapon_class"); weaponClass != "HTH" {
		t.Fatalf("weapon class = %v, want HTH", weaponClass)
	}
	beltStore, found := akara.GetDynamicStore(engine.World(), "dm.player.belt")
	if !found {
		t.Fatal("belt store was not materialized")
	}
	belt, _ := beltStore.Get(entity)
	if capacity, _ := belt.Get("capacity"); capacity != int64(4) {
		t.Fatalf("initial belt capacity = %v, want 4", capacity)
	}
	if item, _ := belt.Get("slot_16"); item != "" {
		t.Fatalf("initial belt slot 16 = %v, want empty", item)
	}
	selectableStore, found := akara.GetDynamicStore(engine.World(), "dm.world.selectable")
	if !found {
		t.Fatal("selectable store was not materialized")
	}
	selectable, present := selectableStore.Get(entity)
	if !present {
		t.Fatal("player lacks selectable component")
	}
	if kind, _ := selectable.Get("kind"); kind != "player" {
		t.Fatalf("selectable kind = %v", kind)
	}
	locationStore, found := akara.GetDynamicStore(engine.World(), "dm.world.location")
	if !found {
		t.Fatal("world location store was not materialized")
	}
	location, _ := locationStore.Get(entity)
	if act, _ := location.Get("act"); act != int64(1) {
		t.Fatalf("entry act = %v", act)
	}
	if level, _ := location.Get("level_id"); level != int64(1) {
		t.Fatalf("entry level = %v", level)
	}
	learned, found := akara.GetDynamicStore(engine.World(), "dm.player.learned_skill")
	if !found || learned.Len() != 1 {
		t.Fatalf("learned skills = %v, %v", learned, found)
	}
	if audit := session.Audit(); len(audit) != 1 || audit[0].Kind != EnterCommand {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestEntrySourceAdmitsSelectedCharacterOnce(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	saves := persistence.New(persistence.Character{ID: "amazon-hero", Name: "Hero", Class: "Amazon", Level: 1})
	if err := saves.Select("amazon-hero"); err != nil {
		t.Fatal(err)
	}
	source, err := NewEntrySourceAt(engine, saves, "local-player", 12, 13, 100, 80, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.AdvanceWithSource(time.Second/25, source.Commands); err != nil {
		t.Fatal(err)
	}
	if _, err := session.AdvanceWithSource(time.Second/25, source.Commands); err != nil {
		t.Fatal(err)
	}
	identities, found := akara.GetDynamicStore(engine.World(), "dm.player.identity")
	if !found || identities.Len() != 1 {
		t.Fatalf("identities = %v, %v", identities, found)
	}
	positions, found := akara.GetDynamicStore(engine.World(), "dm.world.position")
	if !found {
		t.Fatal("position store is missing")
	}
	position, _ := positions.Get(identities.Entities()[0])
	if x, _ := position.Get("x"); x != float64(12) {
		t.Fatalf("spawn x = %v, want authored 12", x)
	}
	if y, _ := position.Get("y"); y != float64(13) {
		t.Fatalf("spawn y = %v, want authored 13", y)
	}
}

func TestEntrySourceRecordsServerSelectedTown(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	saves := persistence.New(persistence.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	if err := saves.Select("hero"); err != nil {
		t.Fatal(err)
	}
	source, err := NewEntrySourceAtLocation(engine, saves, "player", 12, 13, 100, 80, 5, 109, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.AdvanceWithSource(time.Second/25, source.Commands); err != nil {
		t.Fatal(err)
	}
	store, found := akara.GetDynamicStore(engine.World(), "dm.world.location")
	if !found {
		t.Fatal("location store is missing")
	}
	location, _ := store.Get(store.Entities()[0])
	if act, _ := location.Get("act"); act != int64(5) {
		t.Fatalf("act = %v", act)
	}
	if level, _ := location.Get("level_id"); level != int64(109) {
		t.Fatalf("level = %v", level)
	}
}

func TestRemoteAdmissionUsesServerSelectedTownDestination(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	destination, err := NewDestination(23, 17, 100, 80, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	character := persistence.Character{ID: "realm-amazon", Name: "RemoteHero", Class: "Amazon", Level: 1}
	command, err := AdmissionCommand(character, "account:42", destination, nil, "system:join", 1, 1, simulation.AuthoritySystem)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	positions, _ := akara.GetDynamicStore(engine.World(), "dm.world.position")
	locations, _ := akara.GetDynamicStore(engine.World(), "dm.world.location")
	entity := positions.Entities()[0]
	position, _ := positions.Get(entity)
	location, _ := locations.Get(entity)
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	level, _ := location.Get("level_id")
	if x != float64(23) || y != float64(17) || level != int64(1) {
		t.Fatalf("remote admission = (%v,%v) level=%v", x, y, level)
	}
	if _, err := AdmissionCommand(character, "account:42", destination, nil, "client", 2, 2, simulation.AuthorityPlayer); err == nil {
		t.Fatal("client minted a trusted player-admission command")
	}
}
