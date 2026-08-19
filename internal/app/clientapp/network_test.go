package clientapp

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// TestNetworkControllerDefersHostUntilCharacterSelection verifies that host
// construction cannot begin before the user chooses an offline character.
func TestNetworkControllerDefersHostUntilCharacterSelection(t *testing.T) {
	app := &application{ctx: context.Background(), saves: d2save.New()}
	controller := newNetworkController(app)
	if err := controller.Host(); err != nil {
		t.Fatalf("begin host: %v", err)
	}
	status := controller.Status()
	if status["phase"] != "selecting" || status["mode"] != "host" {
		t.Fatalf("pending host status = %#v", status)
	}
	controller.Cancel()
	if status = controller.Status(); status["phase"] != "frontend" || status["mode"] != "" {
		t.Fatalf("cancelled host status = %#v", status)
	}
}

// TestNetworkControllerSamplesFixedInputClockIndependentlyOfCorrections checks
// that renderer time, rather than correction arrival, drives input sampling.
func TestNetworkControllerSamplesFixedInputClockIndependentlyOfCorrections(t *testing.T) {
	controller := newNetworkController(&application{})

	client := networkTestClient(10)
	if ticks := controller.inputTicks(client, 39*time.Millisecond, time.Now()); len(ticks) != 0 {
		t.Fatalf("premature input ticks = %v", ticks)
	}
	ticks := controller.inputTicks(client, 81*time.Millisecond, time.Now())
	if len(ticks) != 1 || ticks[0] != 12 {
		t.Fatalf("fixed input ticks = %v", ticks)
	}
}

// TestEmptyGeneralIntentMailboxDoesNotConsumeJoiningClientsMovementTick checks
// that an empty generic mailbox cannot advance fixed-step movement sequencing.
func TestEmptyGeneralIntentMailboxDoesNotConsumeJoiningClientsMovementTick(t *testing.T) {
	now := time.Unix(700, 0)
	app := &application{commandIntents: &gamesession.IntentController{}}
	controller := newNetworkController(app)

	client := networkTestClient(10)
	if err := controller.submitPendingIntents(client, now); err != nil {
		t.Fatal(err)
	}
	if controller.lastMovementTick != 0 {
		t.Fatalf("empty intent mailbox consumed movement tick %d", controller.lastMovementTick)
	}
	ticks := controller.inputTicks(client, networkInputStep, now)
	if len(ticks) != 1 || ticks[0] != 12 {
		t.Fatalf("joining-client movement ticks = %v, want [12]", ticks)
	}
}

// TestPendingMovementPredictionReplaysFromCanonicalPosition verifies that
// prediction replays unacknowledged input from the latest authority position.
func TestPendingMovementPredictionReplaysFromCanonicalPosition(t *testing.T) {
	catalog, err := movement.LoadCatalog(predictionMovementRecords{})
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{"x": 1, "y": 0, "running": true})
	hud := playeradapter.HUD{
		Tick:     10,
		Position: playeradapter.HUDPosition{X: 10, Y: 20},
		Player:   playeradapter.HUDIdentity{Class: "Amazon"},
		Vitals: playeradapter.HUDVitals{
			Stamina:       84,
			MaxStamina:    84,
			StaminaRaw:    84 * 256,
			MaxStaminaRaw: 84 * 256,
		},
		Movement: playeradapter.HUDMovement{
			Bounds:   playeradapter.HUDPosition{X: 100, Y: 100},
			Radius:   1,
			RunDrain: 20,
		},
	}
	got := predictPosition(hud, []gameserver.CommandIntent{
		{TargetTick: 11, Sequence: 1, Kind: "player.move", Payload: payload},
		{TargetTick: 12, Sequence: 2, Kind: "player.move", Payload: payload},
	}, networkclock.Moment{Tick: 12}, nil, networkInputStep, catalog)
	if math.Abs(got.X-10.72) > 1e-12 || got.Y != 20 {
		t.Fatalf("predicted position = %#v, want x=10.72 y=20", got)
	}
}

// predictionMovementRecords supplies the minimal movement catalog used here.
type predictionMovementRecords struct{}

// Load returns one deterministic Amazon movement record.
func (predictionMovementRecords) Load(string) ([]map[string]string, error) {
	return []map[string]string{{
		"class": "Amazon", "WalkVelocity": "6", "RunVelocity": "9",
		"vit": "20", "stamina": "84", "RunDrain": "20", "StaminaPerLevel": "4", "StaminaPerVitality": "4",
	}}, nil
}

// TestNetworkControllerActivatesLocalSessionOnlyAfterSelection verifies that
// the frontend becomes local gameplay only after a selected save is accepted.
func TestNetworkControllerActivatesLocalSessionOnlyAfterSelection(t *testing.T) {
	character := d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon"}
	saves := d2save.New(character)
	if err := saves.Select(character.ID); err != nil {
		t.Fatal(err)
	}
	controller := newNetworkController(&application{ctx: context.Background(), saves: saves})
	if controller.Local() {
		t.Fatal("frontend was treated as an active local session")
	}
	if err := controller.StartSelected(); err != nil {
		t.Fatal(err)
	}
	if !controller.Local() || controller.Status()["phase"] != "local" {
		t.Fatalf("local session status = %#v", controller.Status())
	}
}

// TestNetworkControllerAcceptsAuthenticatedRealmCharacterForLoading verifies
// that Realm admission, not offline saves, satisfies the loading dependency.
func TestNetworkControllerAcceptsAuthenticatedRealmCharacterForLoading(t *testing.T) {
	controller := newNetworkController(&application{})
	controller.phase = "connected"
	controller.mode = "realm"
	controller.client = &clientsession.Session{Admission: gameserver.JoinResponse{
		Admission: gamesession.AdmissionToken{CharacterID: "realm-hero"},
	}}
	if !controller.hasSelectedCharacter() {
		t.Fatal("authenticated Realm character was not available to loading")
	}
	controller.client.Admission.Admission.CharacterID = ""
	if controller.hasSelectedCharacter() {
		t.Fatal("Realm connection without an admitted character passed loading")
	}
}

// TestNetworkControllerRealmLoadingDoesNotDependOnTransientHUDProjection checks
// that loading survives an admitted character's temporarily empty HUD view.
func TestNetworkControllerRealmLoadingDoesNotDependOnTransientHUDProjection(t *testing.T) {
	app := &application{saves: d2save.New()}
	controller := newNetworkController(app)
	app.network = controller
	controller.phase = "connected"
	controller.mode = "realm"
	controller.client = &clientsession.Session{Admission: gameserver.JoinResponse{
		Admission: gamesession.AdmissionToken{CharacterID: "realm-hero"},
	}}
	if !controller.hasSelectedCharacter() {
		t.Fatal("authenticated admission was hidden by an empty projected HUD")
	}
	if err := app.buildLoadingCoordinator(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.loading.Close)
	if err := app.loading.Begin(t.Context(), []string{"selected_character"}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for app.loading.Snapshot().State == "running" && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if status := app.loading.Snapshot(); status.State != "complete" {
		t.Fatalf("Realm loading selection = %#v", status)
	}
}

// TestNetworkControllerKeepsStartFailuresAndNormalizesDirectJoin verifies both
// frontend error visibility and default direct-server port selection.
func TestNetworkControllerKeepsStartFailuresAndNormalizesDirectJoin(t *testing.T) {
	app := &application{ctx: context.Background(), saves: d2save.New()}
	controller := newNetworkController(app)
	if err := controller.Host(); err != nil {
		t.Fatal(err)
	}
	if err := controller.StartSelected(); err == nil {
		t.Fatal("starting without selected character was accepted")
	}
	status := controller.Status()

	failedSelection := status["phase"] == "failed" &&
		status["mode"] == "host" &&
		strings.Contains(status["error"].(string), "select")
	if !failedSelection {
		t.Fatalf("start rejection status = %#v", status)
	}
	if err := controller.Join("127.0.0.1"); err != nil {
		t.Fatalf("direct join: %v", err)
	}
	status = controller.Status()

	normalizedJoin := status["phase"] == "selecting" &&
		status["mode"] == "join" &&
		status["address"] == "127.0.0.1:6112"
	if !normalizedJoin {
		t.Fatalf("join selection status = %#v", status)
	}
}

// TestNetworkControllerSamplesMovementOncePerAuthoritativeTick verifies that
// repeated render frames cannot resample the same authority tick.
func TestNetworkControllerSamplesMovementOncePerAuthoritativeTick(t *testing.T) {
	controller := newNetworkController(&application{})
	if !controller.sampleMovement(12) {
		t.Fatal("first tick was not sampled")
	}
	for range 120 {
		if controller.sampleMovement(12) {
			t.Fatal("render frames resampled one authoritative tick")
		}
	}
	if !controller.sampleMovement(13) {
		t.Fatal("next authoritative tick was not sampled")
	}
}

// TestNetworkControllerSendsOneStopAfterActiveMovement verifies that movement
// emits one transition to idle without flooding repeated stop commands.
func TestNetworkControllerSendsOneStopAfterActiveMovement(t *testing.T) {
	controller := newNetworkController(&application{})
	if controller.movementRequired(false) {
		t.Fatal("initial idle state emitted a command")
	}
	if !controller.movementRequired(true) {
		t.Fatal("active movement was suppressed")
	}
	controller.markMovement(true)
	if !controller.movementRequired(true) {
		t.Fatal("active movement samples were suppressed")
	}
	if !controller.movementRequired(false) {
		t.Fatal("first idle sample did not stop authoritative velocity")
	}
	controller.markMovement(false)
	if controller.movementRequired(false) {
		t.Fatal("settled idle state emitted repeated stop commands")
	}
}

// TestMovementCommandActivityDistinguishesValidMovement verifies that malformed
// or unrelated commands remain on the generic submission path.
func TestMovementCommandActivityDistinguishesValidMovement(t *testing.T) {
	payload, err := json.Marshal(movement.MovePayload{X: 1})
	if err != nil {
		t.Fatal(err)
	}

	active, movementCommand := movementCommandActivity(simulation.Command{
		Kind:    movement.MoveCommand,
		Payload: payload,
	})
	if !active || !movementCommand {
		t.Fatal("valid active movement was not recognized")
	}

	_, movementCommand = movementCommandActivity(simulation.Command{
		Kind:    movement.MoveCommand,
		Payload: json.RawMessage(`{`),
	})
	if movementCommand {
		t.Fatal("malformed movement was removed from the generic submission path")
	}
}

// TestNetworkRecipeRejectsDifferentLocalAssetSet verifies that package identity
// cannot hide incompatible external game assets.
func TestNetworkRecipeRejectsDifferentLocalAssetSet(t *testing.T) {
	recipe := simulation.RuntimeRecipe{AssetSetID: simulation.EmptyAssetSetID}
	if err := validateLocalAssetSet(recipe, simulation.EmptyAssetSetID); err != nil {
		t.Fatal(err)
	}

	differentAssetSet := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if err := validateLocalAssetSet(recipe, differentAssetSet); err == nil {
		t.Fatal("client accepted a server recipe for a different external asset set")
	}
}

// networkTestClient returns a client with a deterministic authority clock.
func networkTestClient(tick uint64) *clientsession.Session {
	return &clientsession.Session{
		Admission: gameserver.JoinResponse{
			Snapshot: gameserver.Snapshot{
				Tick:      tick,
				StepNanos: int64(networkInputStep),
			},
		},
	}
}
