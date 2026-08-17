package clientapp

import (
	"fmt"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

func TestConnectedProjectionCreatesDistinctAuthenticatedAndPeerPlayers(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	connected := &clientsession.Session{
		HUD: playeradapter.HUD{
			Version: playeradapter.HUDVersion, Tick: 9,
			Player:   playeradapter.HUDIdentity{PlayerID: "player-2", CharacterID: "barbarian-conan", Name: "Conan", Class: "Barbarian"},
			Vitals:   playeradapter.HUDVitals{Health: 60, MaxHealth: 60, Mana: 10, MaxMana: 10},
			Progress: playeradapter.HUDProgress{Level: 1}, Position: playeradapter.HUDPosition{X: 18, Y: 10},
			Location: playeradapter.HUDLocation{Act: 1, LevelID: 2}, Animation: playeradapter.HUDAnimation{Mode: "NU"},
		},
		World: playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: 9, Entities: []playeradapter.WorldEntity{{
			ID: "player:player-1", Kind: "player", Label: "Natalya", Owner: "player-1",
			Position: playeradapter.HUDPosition{X: 10, Y: 10}, Radius: 0.75, Priority: 10,
			Class: "Assassin", Token: "AI", Mode: "NU",
		}}},
		Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: 9},
		Party: playeradapter.PartyView{Version: playeradapter.PartyViewVersion, Tick: 9, Revision: 3, PartyID: "party:1",
			Roster: []playeradapter.PartyRosterEntry{
				{PlayerID: "player-2", Name: "Conan", Class: "Barbarian", Level: 1, Relationship: "self"},
				{PlayerID: "player-1", Name: "Natalya", Class: "Assassin", Level: 2, Relationship: "party"},
			}},
	}
	if err := app.installRemoteView(connected, true); err != nil {
		t.Fatal(err)
	}
	if err := app.syncRemoteMirrors(connected.World.Entities, connected.HUD.Location); err != nil {
		t.Fatal(err)
	}
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.identity")
	appearances, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.appearance")
	if identities.Len() != 2 {
		t.Fatalf("connected player entities = %d, want 2", identities.Len())
	}
	classes := map[string]string{}
	tokens := map[string]string{}
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		appearance, _ := appearances.Get(entity)
		owner, _ := identity.Get("player")
		class, _ := identity.Get("class")
		token, _ := appearance.Get("token")
		classes[owner.(string)] = class.(string)
		tokens[owner.(string)] = token.(string)
	}
	if classes["player-1"] != "Assassin" || tokens["player-1"] != "AI" ||
		classes["player-2"] != "Barbarian" || tokens["player-2"] != "BA" {
		t.Fatalf("connected roster classes=%v tokens=%v", classes, tokens)
	}
	partyViews, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.party_view")
	if partyViews.Len() != 1 {
		t.Fatalf("owner-scoped party views = %d, want 1", partyViews.Len())
	}
	for _, entity := range partyViews.Entities() {
		view, _ := partyViews.Get(entity)
		partyID, _ := view.Get("party_id")
		peer, _ := view.Get("player_2")
		relationship, _ := view.Get("relationship_2")
		if partyID != "party:1" || peer != "player-1" || relationship != "party" {
			t.Fatalf("installed party view party=%v peer=%v relationship=%v", partyID, peer, relationship)
		}
	}
}

func TestRemoteMirrorStructureAndSampledTransformHaveSeparateLifecycles(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	location := playeradapter.HUDLocation{Act: 1, LevelID: 2}
	initial := playeradapter.WorldEntity{
		ID: "player:peer", Kind: "player", Label: "Peer", Owner: "peer",
		Position: playeradapter.HUDPosition{X: 10, Y: 20}, Class: "Amazon", Token: "AM", Mode: "NU",
		VelocityPercent: -50, ItemFasterMoveVelocity: 100,
	}
	if err := app.syncRemoteMirrors([]playeradapter.WorldEntity{initial}, location); err != nil {
		t.Fatal(err)
	}
	entity := app.remoteMirrors[initial.ID]
	movement, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.movement_stats")
	projected, _ := movement.Get(entity)
	velocityPercent, _ := projected.Get("velocitypercent")
	itemFRW, _ := projected.Get("item_fastermovevelocity")
	if velocityPercent != int64(-50) || itemFRW != int64(100) {
		t.Fatalf("remote movement animation inputs = %v/%v", velocityPercent, itemFRW)
	}
	updated := initial
	updated.Position = playeradapter.HUDPosition{X: 100, Y: 200}
	updated.Mode = "WL"
	if err := app.syncRemoteMirrors([]playeradapter.WorldEntity{updated}, location); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(engine.World(), entity); got != initial.Position {
		t.Fatalf("structural update moved transform to %+v, want %+v", got, initial.Position)
	}
	if err := app.applySampledWorldPositions([]playeradapter.WorldEntity{updated}); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(engine.World(), entity); got != updated.Position {
		t.Fatalf("sampled transform = %+v, want %+v", got, updated.Position)
	}
	if err := app.syncRemoteMirrors(nil, location); err != nil {
		t.Fatal(err)
	}
	if len(app.remoteMirrors) != 0 || len(app.remoteMirrorKeys) != 0 {
		t.Fatalf("removed mirror retained: mirrors=%v keys=%v", app.remoteMirrors, app.remoteMirrorKeys)
	}
}

func TestCorrectionDoesNotMoveAnExistingPresentationMirror(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	entity, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	if _, err := positions.Set(entity, map[string]any{"x": float64(10), "y": float64(20)}); err != nil {
		t.Fatal(err)
	}
	if err := moveMirrorToward(engine.World(), entity, playeradapter.HUDPosition{X: 12, Y: 22}, correctionAlpha(false)); err != nil {
		t.Fatal(err)
	}
	position, _ := positions.Get(entity)
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	if x != float64(10) || y != float64(20) {
		t.Fatalf("correction moved rendered mirror to (%v, %v)", x, y)
	}
}

func TestLocalPredictionUpdatesOnlyAuthenticatedOwnerTransform(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	local, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	peer, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	controls, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.player_control")
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	if _, err := controls.Set(local, map[string]any{"player": "player-1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := positions.Set(local, map[string]any{"x": float64(10), "y": float64(20)}); err != nil {
		t.Fatal(err)
	}
	if _, err := positions.Set(peer, map[string]any{"x": float64(30), "y": float64(40)}); err != nil {
		t.Fatal(err)
	}
	app := &application{clientSimulation: engine}
	if err := app.applyLocalPredictedPosition("player-1", playeradapter.HUDPosition{X: 12, Y: 22}); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(engine.World(), local); got != (playeradapter.HUDPosition{X: 12, Y: 22}) {
		t.Fatalf("local position = %+v", got)
	}
	if got := currentPosition(engine.World(), peer); got != (playeradapter.HUDPosition{X: 30, Y: 40}) {
		t.Fatalf("peer position was changed to %+v", got)
	}
}

func TestAnimationTimelineUsesPredictionForOwnerAndInterpolationForPeer(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.identity")
	animations, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.animation")
	clocks, _ := akara.GetDynamicStore(engine.World(), "d2legacy.presentation.animation_clock")
	local := engine.World().MustCreateEntity()
	peer := engine.World().MustCreateEntity()
	for entity, values := range map[akara.Entity]struct {
		owner string
		start int64
	}{local: {owner: "local", start: 10}, peer: {owner: "peer", start: 9}} {
		if _, err := identities.Set(entity, map[string]any{"player": values.owner, "character_id": values.owner, "name": values.owner, "class": "Amazon"}); err != nil {
			t.Fatal(err)
		}
		if _, err := animations.Set(entity, map[string]any{"direction": int64(0), "mode": "WL", "start_tick": values.start}); err != nil {
			t.Fatal(err)
		}
	}
	app := &application{clientSimulation: engine}
	timeline := networkclock.Timeline{Ready: true, Prediction: networkclock.Moment{Tick: 12, Fraction: .5}, Interpolation: networkclock.Moment{Tick: 10, Fraction: .5}}
	if err := app.applyAnimationTimeline("local", timeline, 40*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	localClock, _ := clocks.Get(local)
	peerClock, _ := clocks.Get(peer)
	localSeconds, _ := localClock.Get("seconds")
	peerSeconds, _ := peerClock.Get("seconds")
	if localSeconds.(float64) != .1 || peerSeconds.(float64) != .06 {
		t.Fatalf("animation seconds local=%v peer=%v", localSeconds, peerSeconds)
	}
}

func TestPrivateProjectionFingerprintChangesOnlyWithPrivateState(t *testing.T) {
	learned := []playeradapter.HUDLearnedSkill{{SkillID: 7, Level: 1}}
	private := playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: 10}
	first, err := privateProjectionFingerprint(learned, private)
	if err != nil {
		t.Fatal(err)
	}
	private.Tick++
	second, err := privateProjectionFingerprint(learned, private)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("transport tick rebuilt unchanged private presentation state")
	}
	private.Items.Layout.CarriedGold++
	third, err := privateProjectionFingerprint(learned, private)
	if err != nil {
		t.Fatal(err)
	}
	if second == third {
		t.Fatal("private content change was not fingerprinted")
	}
}

func registerRemoteViewSchemas(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	field := func(name string, kind akara.FieldKind) akara.Field { return akara.Field{Name: name, Kind: kind} }
	partyFields := []akara.Field{field("schema_version", akara.FieldInt64), field("revision", akara.FieldInt64), field("party_id", akara.FieldString), field("roster_count", akara.FieldInt64)}
	for slot := 1; slot <= playeradapter.MaxPartyViewRoster; slot++ {
		suffix := fmt.Sprintf("_%d", slot)
		partyFields = append(partyFields,
			field("player"+suffix, akara.FieldString), field("name"+suffix, akara.FieldString),
			field("class"+suffix, akara.FieldString), field("level"+suffix, akara.FieldInt64),
			field("relationship"+suffix, akara.FieldString))
	}
	schemas := []akara.Schema{
		{Name: "d2legacy.player.identity", Fields: []akara.Field{field("character_id", akara.FieldString), field("player", akara.FieldString), field("name", akara.FieldString), field("class", akara.FieldString)}},
		{Name: "d2legacy.player.vitals", Fields: []akara.Field{field("health", akara.FieldInt64), field("max_health", akara.FieldInt64), field("mana", akara.FieldInt64), field("max_mana", akara.FieldInt64), field("mana_raw", akara.FieldInt64), field("max_mana_raw", akara.FieldInt64), field("stamina", akara.FieldInt64), field("max_stamina", akara.FieldInt64), field("stamina_raw", akara.FieldInt64), field("max_stamina_raw", akara.FieldInt64)}},
		{Name: "d2legacy.player.progress", Fields: []akara.Field{field("level", akara.FieldInt64), field("experience", akara.FieldInt64), field("unspent_skill_points", akara.FieldInt64)}},
		{Name: "d2legacy.player.combat_stats", Fields: []akara.Field{field("attack_rating", akara.FieldInt64), field("defense", akara.FieldInt64)}},
		{Name: "d2legacy.player.animation", Fields: []akara.Field{field("direction", akara.FieldInt64), field("mode", akara.FieldString), field("start_tick", akara.FieldInt64)}},
		{Name: "d2legacy.presentation.animation_clock", Fields: []akara.Field{field("seconds", akara.FieldFloat64)}},
		{Name: "d2legacy.player.appearance", Fields: []akara.Field{field("cof", akara.FieldString), field("token", akara.FieldString), field("palette", akara.FieldString), field("weapon_class", akara.FieldString)}},
		{Name: "d2legacy.player.movement_mode", Fields: []akara.Field{field("running", akara.FieldBool)}},
		{Name: "d2legacy.player.movement_stats", Fields: []akara.Field{field("run_drain", akara.FieldInt64), field("velocitypercent", akara.FieldInt64), field("item_fastermovevelocity", akara.FieldInt64), field("staminarecoverybonus", akara.FieldInt64), field("item_staminadrainpct", akara.FieldInt64), field("armor_run_drain", akara.FieldInt64)}},
		{Name: "d2legacy.player.skill_assignment", Fields: []akara.Field{field("left", akara.FieldInt64), field("right", akara.FieldInt64)}},
		{Name: "d2legacy.player.learned_skill", Fields: []akara.Field{field("owner", akara.FieldEntity), field("skill_id", akara.FieldInt64), field("level", akara.FieldInt64), field("list_row", akara.FieldInt64), field("left_allowed", akara.FieldBool), field("right_allowed", akara.FieldBool)}},
		{Name: "d2legacy.player.belt", Fields: []akara.Field{field("capacity", akara.FieldInt64)}},
		{Name: "d2legacy.player.party_view", Fields: partyFields},
		{Name: "d2legacy.items.layout", Fields: []akara.Field{field("owner", akara.FieldString), field("inventory_width", akara.FieldInt64), field("inventory_height", akara.FieldInt64), field("stash_width", akara.FieldInt64), field("stash_height", akara.FieldInt64), field("cube_width", akara.FieldInt64), field("cube_height", akara.FieldInt64), field("belt_capacity", akara.FieldInt64), field("active_weapon_set", akara.FieldInt64), field("vendor_width", akara.FieldInt64), field("vendor_height", akara.FieldInt64), field("carried_gold", akara.FieldInt64), field("stashed_gold", akara.FieldInt64)}},
		{Name: "d2legacy.item.identity", Fields: []akara.Field{field("owner", akara.FieldEntity), field("id", akara.FieldString), field("code", akara.FieldString), field("width", akara.FieldInt64), field("height", akara.FieldInt64), field("body_slots", akara.FieldString), field("belt_eligible", akara.FieldBool), field("base_cost", akara.FieldInt64), field("applied_services", akara.FieldString)}},
		{Name: "d2legacy.item.placement", Fields: []akara.Field{field("container", akara.FieldString), field("x", akara.FieldInt64), field("y", akara.FieldInt64), field("slot", akara.FieldString), field("belt_slot", akara.FieldInt64), field("weapon_set", akara.FieldInt64), field("page", akara.FieldInt64)}},
		{Name: "d2legacy.item.presentation", Fields: []akara.Field{field("inventory_dc6", akara.FieldString), field("world_dc6", akara.FieldString), field("world_animated", akara.FieldBool), field("composite", akara.FieldString), field("weapon_class", akara.FieldString)}},
		{Name: "d2legacy.interaction.target", Fields: []akara.Field{field("id", akara.FieldString), field("npc", akara.FieldString), field("vendor", akara.FieldString), field("categories", akara.FieldString), field("services", akara.FieldString), field("x", akara.FieldFloat64), field("y", akara.FieldFloat64), field("radius", akara.FieldFloat64)}},
		{Name: "d2legacy.interaction.context", Fields: []akara.Field{field("owner", akara.FieldString), field("target", akara.FieldEntity)}},
		{Name: "d2legacy.interaction.null_target", Fields: []akara.Field{}},
		{Name: "d2legacy.world.position", Fields: []akara.Field{field("x", akara.FieldFloat64), field("y", akara.FieldFloat64)}},
		{Name: "d2legacy.world.velocity", Fields: []akara.Field{field("x", akara.FieldFloat64), field("y", akara.FieldFloat64)}},
		{Name: "d2legacy.world.facing", Fields: []akara.Field{field("direction", akara.FieldInt64), field("directions", akara.FieldInt64)}},
		{Name: "d2legacy.world.location", Fields: []akara.Field{field("act", akara.FieldInt64), field("level_id", akara.FieldInt64)}},
		{Name: "d2legacy.world.player_control", Fields: []akara.Field{field("player", akara.FieldString)}},
		{Name: "d2legacy.world.bounds", Fields: []akara.Field{field("width", akara.FieldFloat64), field("height", akara.FieldFloat64)}},
		{Name: "d2legacy.world.collider", Fields: []akara.Field{field("radius", akara.FieldFloat64)}},
		{Name: "d2legacy.world.selectable", Fields: []akara.Field{field("id", akara.FieldString), field("kind", akara.FieldString), field("label", akara.FieldString), field("owner", akara.FieldString), field("radius", akara.FieldFloat64), field("priority", akara.FieldInt64)}},
	}
	for _, schema := range schemas {
		if _, err := akara.RegisterSchema(engine.World(), schema); err != nil {
			t.Fatal(err)
		}
	}
}
