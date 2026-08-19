package clientapp

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// TestConnectedProjectionCreatesDistinctAuthenticatedAndPeerPlayers proves owner-private projection
// attaches only to the authenticated hero while public peers retain separate entities.
func TestConnectedProjectionCreatesDistinctAuthenticatedAndPeerPlayers(t *testing.T) {
	engine := gameecs.New()

	t.Cleanup(func() { _ = engine.Close() })

	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	connected := connectedProjectionFixture()

	if err := app.installRemoteView(connected, true); err != nil {
		t.Fatal(err)
	}

	if err := app.syncRemoteMirrors(connected.World.Entities, connected.HUD.Location); err != nil {
		t.Fatal(err)
	}

	assertConnectedPlayerClasses(t, engine)
	assertConnectedOwnerPartyView(t, engine)
}

// connectedProjectionFixture keeps owner-private HUD facts separate from the peer's public world projection.
func connectedProjectionFixture() *clientsession.Session {
	return &clientsession.Session{
		HUD: playeradapter.HUD{
			Version: playeradapter.HUDVersion, Tick: 9,
			Player: playeradapter.HUDIdentity{
				PlayerID:    "player-2",
				CharacterID: "barbarian-conan",
				Name:        "Conan",
				Class:       "Barbarian",
			},
			Vitals:   playeradapter.HUDVitals{Health: 60, MaxHealth: 60, Mana: 10, MaxMana: 10},
			Progress: playeradapter.HUDProgress{Level: 1}, Position: playeradapter.HUDPosition{X: 18, Y: 10},
			Location: playeradapter.HUDLocation{Act: 1, LevelID: 2}, Animation: playeradapter.HUDAnimation{Mode: "NU"},
		},
		World: playeradapter.WorldView{
			Version: playeradapter.WorldViewVersion,
			Tick:    9,
			Entities: []playeradapter.WorldEntity{{
				ID:       "player:player-1",
				Kind:     "player",
				Label:    "Natalya",
				Owner:    "player-1",
				Position: playeradapter.HUDPosition{X: 10, Y: 10},
				Radius:   0.75,
				Priority: 10,
				Class:    "Assassin",
				Token:    "AI",
				Mode:     "NU",
			}},
		},
		Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: 9},
		Party: playeradapter.PartyView{
			Version: playeradapter.PartyViewVersion, Tick: 9, Revision: 3, PartyID: "party:1",
			Roster: []playeradapter.PartyRosterEntry{
				{PlayerID: "player-2", Name: "Conan", Class: "Barbarian", Level: 1, Relationship: "self"},
				{PlayerID: "player-1", Name: "Natalya", Class: "Assassin", Level: 2, Relationship: "party"},
			},
		},
	}
}

// assertConnectedPlayerClasses proves owner and peer retain independent class-token presentation identities.
func assertConnectedPlayerClasses(t *testing.T, engine *gameecs.Engine) {
	t.Helper()

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
}

// assertConnectedOwnerPartyView ensures party-private projection attaches once to the authenticated player only.
func assertConnectedOwnerPartyView(t *testing.T, engine *gameecs.Engine) {
	t.Helper()

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

// TestRemoteMirrorStructureAndSampledTransformHaveSeparateLifecycles protects the no-cuddling boundary
// between discrete roster changes and per-frame interpolation writes.
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

// TestCorrectionDoesNotMoveAnExistingPresentationMirror ensures packet arrival updates interpolation
// targets without producing an immediate visible teleport.
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

	target := playeradapter.HUDPosition{X: 12, Y: 22}
	if err := moveMirrorToward(engine.World(), entity, target, correctionAlpha(false)); err != nil {
		t.Fatal(err)
	}

	position, _ := positions.Get(entity)
	x, _ := position.Get("x")

	y, _ := position.Get("y")
	if x != float64(10) || y != float64(20) {
		t.Fatalf("correction moved rendered mirror to (%v, %v)", x, y)
	}
}

// TestLocalPredictionUpdatesOnlyAuthenticatedOwnerTransform requires player_control ownership to select
// the predicted entity, preventing local input from moving a peer.
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

// TestAnimationTimelineUsesPredictionForOwnerAndInterpolationForPeer preserves low-latency owner
// animation and delayed smooth peer animation as separate timeline policies.
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
		identity := map[string]any{
			"player":       values.owner,
			"character_id": values.owner,
			"name":         values.owner,
			"class":        "Amazon",
		}
		if _, err := identities.Set(entity, identity); err != nil {
			t.Fatal(err)
		}

		animation := map[string]any{
			"direction":  int64(0),
			"mode":       "WL",
			"start_tick": values.start,
		}
		if _, err := animations.Set(entity, animation); err != nil {
			t.Fatal(err)
		}
	}

	app := &application{clientSimulation: engine}

	timeline := networkclock.Timeline{
		Ready:         true,
		Prediction:    networkclock.Moment{Tick: 12, Fraction: 0.5},
		Interpolation: networkclock.Moment{Tick: 10, Fraction: 0.5},
	}
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

// TestPrivateProjectionFingerprintChangesOnlyWithPrivateState ensures transport metadata cannot churn
// owner graphs while real skill, inventory, or interaction changes do rebuild them.
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
