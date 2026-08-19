package gameserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type testAuthenticator struct {
	credential string
	principal  Principal
}

// Authenticate accepts one fixture credential so endpoint tests can distinguish authentication from session state.
func (auth testAuthenticator) Authenticate(_ context.Context, credential string) (Principal, error) {
	if credential != auth.credential {
		return Principal{}, ErrAuthentication
	}

	return auth.principal, nil
}

// protocolTestFixture owns the endpoint's ECS and session so every test gets isolated membership and rate-limit state.
type protocolTestFixture struct {
	identity simulation.RuntimeIdentity
	engine   *gameecs.Engine
	session  *gamesession.Session
	endpoint *Endpoint
}

// newProtocolTestFixture constructs the transport-neutral endpoint around a real authoritative session.
func newProtocolTestFixture(
	t *testing.T,
	prediction gamesession.PredictionTier,
	auth testAuthenticator,
	project SnapshotProjector,
) protocolTestFixture {
	t.Helper()

	identity := testProtocolIdentity()

	allocation, err := gamesession.Allocate("game", identity, prediction)
	if err != nil {
		t.Fatal(err)
	}

	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })

	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation}, auth, project)
	if err != nil {
		t.Fatal(err)
	}

	return protocolTestFixture{identity: identity, engine: engine, session: session, endpoint: endpoint}
}

// validProtocolAuthenticator returns the canonical fixture principal bound to one admission credential.
func validProtocolAuthenticator(credential string) testAuthenticator {
	return testAuthenticator{
		credential: credential,
		principal: Principal{
			ID: "account:7", CharacterID: "character:11", PlayerID: "player:3",
		},
	}
}

// emptyProtocolProjector keeps protocol tests focused on admission and lifecycle rather than view schema.
func emptyProtocolProjector(string, simulation.Checkpoint) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

// joinProtocolFixture performs the ordinary handshake used by tests whose interesting behavior starts after admission.
func joinProtocolFixture(t *testing.T, fixture protocolTestFixture, credential string) JoinResponse {
	t.Helper()

	joined, err := fixture.endpoint.Join(context.Background(), JoinRequest{
		Version:    SessionProtocolVersion,
		Credential: credential,
		Identity:   fixture.identity,
	})
	if err != nil {
		t.Fatal(err)
	}

	return joined
}

// TestEndpointAuthenticatesBindsCommandsAndReconnects covers the complete credential lifecycle and command binding.
func TestEndpointAuthenticatesBindsCommandsAndReconnects(t *testing.T) {
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionLimited,
		validProtocolAuthenticator("realm-ticket"),
		func(playerID string, _ simulation.Checkpoint) (json.RawMessage, error) {
			return json.Marshal(map[string]string{"player_id": playerID})
		},
	)
	if err := fixture.session.Register("player.move", gamesession.CommandHandler{
		Validate: func(command simulation.Command) error {
			if !json.Valid(command.Payload) {
				return errors.New("invalid payload")
			}

			return nil
		},
		Apply: func(*gameecs.Engine, simulation.Command) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}

	var connected []string

	fixture.endpoint.SetConnected(func(principal Principal) { connected = append(connected, principal.PlayerID) })

	joined := joinProtocolFixture(t, fixture, "realm-ticket")
	if joined.Credential == "" ||
		joined.Admission.CharacterID != "character:11" ||
		string(joined.Snapshot.Payload) != `{"player_id":"player:3"}` {
		t.Fatalf("unexpected join response: %#v", joined)
	}

	if len(connected) != 1 || connected[0] != "player:3" {
		t.Fatalf("connected observations = %#v", connected)
	}

	if err := fixture.endpoint.Submit(joined.Credential, CommandIntent{
		TargetTick: 1, Sequence: 1, Kind: "player.move", Payload: json.RawMessage(`{"x":4}`),
	}); err != nil {
		t.Fatal(err)
	}

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}

	replay, err := fixture.session.Replay()
	if err != nil {
		t.Fatal(err)
	}

	if len(replay.Commands) != 1 ||
		replay.Commands[0].Player != "player:3" ||
		replay.Commands[0].Authority != simulation.AuthorityPlayer {
		t.Fatalf("command identity was not server-bound: %#v", replay.Commands)
	}

	assertEndpointReconnectLifecycle(t, fixture, joined, &connected)
}

// assertEndpointReconnectLifecycle verifies rotation, idempotent replay, stale rejection, and final revocation.
func assertEndpointReconnectLifecycle(
	t *testing.T,
	fixture protocolTestFixture,
	joined JoinResponse,
	connected *[]string,
) {
	t.Helper()

	corrected, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: joined.Credential, Identity: fixture.identity, Nonce: testReconnectNonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	if corrected.Credential == joined.Credential || corrected.Snapshot.Tick != 1 || corrected.Snapshot.Checksum == "" {
		t.Fatalf("reconnect snapshot = %#v", corrected)
	}

	if len(*connected) != 2 || (*connected)[1] != "player:3" {
		t.Fatalf("reconnect observations = %#v", *connected)
	}

	replayed, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: joined.Credential, Identity: fixture.identity, Nonce: testReconnectNonce,
	})
	if err != nil ||
		replayed.Credential != corrected.Credential ||
		replayed.Snapshot.Checksum != corrected.Snapshot.Checksum {
		t.Fatalf("replayed reconnect = %#v, %v", replayed, err)
	}

	if len(*connected) != 2 {
		t.Fatalf("replayed reconnect emitted another observation: %#v", *connected)
	}

	if _, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: joined.Credential,
		Identity:   fixture.identity,
		Nonce:      "fedcba9876543210fedcba9876543210",
	}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("stale reconnect with another nonce error = %v", err)
	}

	if err := fixture.endpoint.Submit(joined.Credential, CommandIntent{}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("stale credential error = %v", err)
	}

	if err := fixture.endpoint.Leave(corrected.Credential); err != nil {
		t.Fatal(err)
	}

	if _, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: corrected.Credential, Identity: fixture.identity, Nonce: testReconnectNonce,
	}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("revoked credential error = %v", err)
	}
}

// TestEndpointAcknowledgesExactRetransmitAndRejectsSequenceConflict distinguishes retry from equivocation.
func TestEndpointAcknowledgesExactRetransmitAndRejectsSequenceConflict(t *testing.T) {
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionLimited,
		validProtocolAuthenticator("valid"),
		emptyProtocolProjector,
	)
	if err := fixture.session.Register("player.move", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}

	joined := joinProtocolFixture(t, fixture, "valid")

	intent := CommandIntent{TargetTick: 1, Sequence: 1, Kind: "player.move", Payload: json.RawMessage(`{"x":1}`)}
	if err := fixture.endpoint.Submit(joined.Credential, intent); err != nil {
		t.Fatal(err)
	}
	// A network retry must be harmless when every command field is identical.
	if err := fixture.endpoint.Submit(joined.Credential, intent); err != nil {
		t.Fatalf("exact retransmit = %v", err)
	}

	conflict := intent

	conflict.Payload = json.RawMessage(`{"x":2}`)
	if err := fixture.endpoint.Submit(joined.Credential, conflict); !errors.Is(err, gamesession.ErrCommandSequence) {
		t.Fatalf("conflicting retransmit error = %v", err)
	}
}

// TestEndpointRejectsUntrustedAndIncompatibleClients keeps every trust boundary ahead of mutable session state.
func TestEndpointRejectsUntrustedAndIncompatibleClients(t *testing.T) {
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionNone,
		validProtocolAuthenticator("valid"),
		emptyProtocolProjector,
	)
	if _, err := fixture.endpoint.Join(context.Background(), JoinRequest{
		Version:    SessionProtocolVersion + 1,
		Credential: "valid",
		Identity:   fixture.identity,
	}); !errors.Is(err, ErrProtocol) {
		t.Fatalf("future protocol error = %v", err)
	}

	if _, err := fixture.endpoint.Join(context.Background(), JoinRequest{
		Version:    SessionProtocolVersion,
		Credential: "invalid",
		Identity:   fixture.identity,
	}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("authentication error = %v", err)
	}

	joined := joinProtocolFixture(t, fixture, "valid")
	if err := fixture.endpoint.Submit(SessionCredential("forged"), CommandIntent{}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("forged session credential error = %v", err)
	}

	mismatch := fixture.identity

	mismatch.Recipe.Packages.Base.Digest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if _, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: joined.Credential,
		Identity:   mismatch,
		Nonce:      testReconnectNonce,
	}); !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("incompatible reconnect error = %v", err)
	}
}

// TestSessionProtocolVersionMatchesCanonicalRuntimeRecipe prevents the handshake and recipe constants from drifting.
func TestSessionProtocolVersionMatchesCanonicalRuntimeRecipe(t *testing.T) {
	want := fmt.Sprintf("dark-magic.game-session/v%d", SessionProtocolVersion)
	if simulation.RuntimeNetworkProtocol != want {
		t.Fatalf("runtime network protocol = %q, want %q", simulation.RuntimeNetworkProtocol, want)
	}
}

// TestEndpointRejectsTicketPinnedToAnotherRuntime prevents a valid ticket from crossing runtime boundaries.
func TestEndpointRejectsTicketPinnedToAnotherRuntime(t *testing.T) {
	auth := validProtocolAuthenticator("valid")
	auth.principal.RuntimeIdentityHash = "another-runtime"
	fixture := newProtocolTestFixture(t, gamesession.PredictionNone, auth, emptyProtocolProjector)

	if _, err := fixture.endpoint.Join(context.Background(), JoinRequest{
		Version:    SessionProtocolVersion,
		Credential: "valid",
		Identity:   fixture.identity,
	}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("runtime-bound ticket error = %v", err)
	}
}

// TestEndpointWaitsForTrustedPlayerAdmissionProjection keeps admission hidden until its first view can be built.
func TestEndpointWaitsForTrustedPlayerAdmissionProjection(t *testing.T) {
	pending := errors.New("player pending")
	attempts := 0
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionNone,
		validProtocolAuthenticator("valid"),
		func(string, simulation.Checkpoint) (json.RawMessage, error) {
			attempts++
			if attempts < 3 {
				return nil, pending
			}

			return json.RawMessage(`{}`), nil
		},
	)
	fixture.endpoint.SetSnapshotPending(func(err error) bool { return errors.Is(err, pending) })
	joinProtocolFixture(t, fixture, "valid")

	if attempts != 3 {
		t.Fatalf("projection attempts = %d", attempts)
	}
}

// TestEndpointRateLimitsPerMembershipAndRefills proves credential rotation cannot reset a member's shared budget.
func TestEndpointRateLimitsPerMembershipAndRefills(t *testing.T) {
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionNone,
		validProtocolAuthenticator("valid"),
		emptyProtocolProjector,
	)
	now := time.Unix(100, 0)
	fixture.endpoint.now = func() time.Time { return now }
	joined := joinProtocolFixture(t, fixture, "valid")

	for range refreshBurst {
		if _, err := fixture.endpoint.Refresh(joined.Credential); err != nil {
			t.Fatal(err)
		}
	}
	// Server-paced observation has its own correction ticker and must remain
	// available after the client-paced unary refresh budget is exhausted.
	for range refreshBurst * 3 {
		if _, err := fixture.endpoint.Observe(joined.Credential); err != nil {
			t.Fatalf("server-paced observation consumed refresh budget: %v", err)
		}
	}

	if _, err := fixture.endpoint.Refresh(joined.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("rate error = %v", err)
	}

	rotated, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: joined.Credential,
		Identity:   fixture.identity,
		Nonce:      testReconnectNonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fixture.endpoint.Refresh(rotated.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("rotated credential reset rate budget: %v", err)
	}

	now = now.Add(time.Second)

	for range int(refreshRate) {
		if _, err := fixture.endpoint.Refresh(rotated.Credential); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := fixture.endpoint.Refresh(rotated.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("refill rate error = %v", err)
	}
}

// TestEndpointRateLimitIsRaceSafeUnderBurst proves concurrent requests cannot overdraw one membership bucket.
func TestEndpointRateLimitIsRaceSafeUnderBurst(t *testing.T) {
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionNone,
		validProtocolAuthenticator("valid"),
		emptyProtocolProjector,
	)
	fixture.endpoint.now = func() time.Time { return time.Unix(100, 0) }
	joined := joinProtocolFixture(t, fixture, "valid")

	var wait sync.WaitGroup

	results := make(chan error, 64)

	for range 64 {
		wait.Add(1)
		go func() {
			defer wait.Done()

			_, err := fixture.endpoint.Refresh(joined.Credential)
			results <- err
		}()
	}

	wait.Wait()
	close(results)

	accepted := 0

	for err := range results {
		if err == nil {
			accepted++
		} else if !errors.Is(err, ErrRateLimit) {
			t.Fatal(err)
		}
	}

	if accepted != refreshBurst {
		t.Fatalf("accepted %d refreshes, want %d", accepted, refreshBurst)
	}
}

// TestEndpointAllowsOneCorrectionWatchPerMembership prevents duplicate streams from amplifying correction work.
func TestEndpointAllowsOneCorrectionWatchPerMembership(t *testing.T) {
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionNone,
		validProtocolAuthenticator("valid"),
		emptyProtocolProjector,
	)
	joined := joinProtocolFixture(t, fixture, "valid")

	if err := fixture.endpoint.BeginWatch(joined.Credential); err != nil {
		t.Fatal(err)
	}

	if err := fixture.endpoint.BeginWatch(joined.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("second watch error = %v", err)
	}

	fixture.endpoint.EndWatch(joined.Credential)

	if err := fixture.endpoint.BeginWatch(joined.Credential); err != nil {
		t.Fatalf("replacement watch error = %v", err)
	}
}

// TestEndpointKeepsDisconnectedMembershipReconnectableUntilLeaseExpires separates transport loss from departure.
func TestEndpointKeepsDisconnectedMembershipReconnectableUntilLeaseExpires(t *testing.T) {
	fixture := newProtocolTestFixture(
		t,
		gamesession.PredictionLimited,
		validProtocolAuthenticator("valid"),
		emptyProtocolProjector,
	)

	var expirations []func()

	fixture.endpoint.after = func(_ time.Duration, expire func()) {
		expirations = append(expirations, expire)
	}
	leaves := 0

	fixture.endpoint.SetLeave(func(Principal) error {
		leaves++
		return nil
	})

	joined := joinProtocolFixture(t, fixture, "valid")
	fixture.endpoint.Disconnect(joined.Credential)

	if err := fixture.endpoint.Submit(joined.Credential, CommandIntent{}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("suspended submit error = %v", err)
	}

	if leaves != 0 || len(expirations) != 1 {
		t.Fatalf("disconnect leaves=%d expirations=%d", leaves, len(expirations))
	}

	reconnected, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: joined.Credential,
		Identity:   fixture.identity,
		Nonce:      testReconnectNonce,
	})
	if err != nil {
		t.Fatal(err)
	}
	// The old timer must observe its stale generation and leave the replacement credential alone.
	expirations[0]()

	if leaves != 0 {
		t.Fatalf("stale lease expiration removed reconnected player")
	}

	fixture.endpoint.Disconnect(reconnected.Credential)

	if len(expirations) != 3 {
		t.Fatalf("expirations = %d", len(expirations))
	}
	// Reconnect replay expiry is independent of membership lease expiry.
	expirations[1]()
	expirations[2]()

	if leaves != 1 {
		t.Fatalf("expired disconnect leaves = %d", leaves)
	}

	if _, err := fixture.endpoint.Reconnect(ReconnectRequest{
		Credential: reconnected.Credential,
		Identity:   fixture.identity,
		Nonce:      testReconnectNonce,
	}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("expired reconnect error = %v", err)
	}
}

// testProtocolIdentity returns a complete, internally consistent identity for transport-neutral endpoint tests.
func testProtocolIdentity() simulation.RuntimeIdentity {
	const packageDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	return simulation.RuntimeIdentity{
		Recipe: simulation.RuntimeRecipe{
			Schema:               simulation.RuntimeRecipeSchema,
			EngineAPI:            "v1",
			NetworkProtocol:      "test/v1",
			AssetSetID:           simulation.EmptyAssetSetID,
			GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID),
			Packages: simulation.RuntimePackageSet{
				Base: simulation.RuntimePackage{
					ID:              "d2legacy",
					Version:         "1.0.0",
					Digest:          packageDigest,
					Size:            1,
					Redistributable: true,
				},
			},
			AuthoritativeHash: "rules",
			ConfigurationHash: "config",
		},
	}
}

// testReconnectNonce is stable so tests can distinguish idempotent replay from a conflicting reconnect attempt.
const testReconnectNonce = "0123456789abcdef0123456789abcdef"
