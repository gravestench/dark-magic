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

func (auth testAuthenticator) Authenticate(_ context.Context, credential string) (Principal, error) {
	if credential != auth.credential {
		return Principal{}, ErrAuthentication
	}
	return auth.principal, nil
}

func TestEndpointAuthenticatesBindsCommandsAndReconnects(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, err := gamesession.Allocate("game-1", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	if err := session.Register("player.move", gamesession.CommandHandler{
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
	host := &Host{Engine: engine, Session: session, Allocation: allocation}
	endpoint, err := NewEndpoint(host, testAuthenticator{
		credential: "realm-ticket",
		principal:  Principal{ID: "account:7", CharacterID: "character:11", PlayerID: "player:3"},
	}, func(playerID string, _ simulation.Checkpoint) (json.RawMessage, error) {
		return json.Marshal(map[string]string{"player_id": playerID})
	})
	if err != nil {
		t.Fatal(err)
	}

	joined, err := endpoint.Join(context.Background(), JoinRequest{
		Version: SessionProtocolVersion, Credential: "realm-ticket", Identity: identity,
	})
	if err != nil {
		t.Fatal(err)
	}
	if joined.Credential == "" || joined.Admission.CharacterID != "character:11" || string(joined.Snapshot.Payload) != `{"player_id":"player:3"}` {
		t.Fatalf("unexpected join response: %#v", joined)
	}
	if err := endpoint.Submit(joined.Credential, CommandIntent{
		TargetTick: 1, Sequence: 1, Kind: "player.move", Payload: json.RawMessage(`{"x":4}`),
	}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Commands) != 1 || replay.Commands[0].Player != "player:3" || replay.Commands[0].Authority != simulation.AuthorityPlayer {
		t.Fatalf("command identity was not server-bound: %#v", replay.Commands)
	}
	corrected, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: identity, Nonce: testReconnectNonce})
	if err != nil {
		t.Fatal(err)
	}
	if corrected.Credential == joined.Credential || corrected.Snapshot.Tick != 1 || corrected.Snapshot.Checksum == "" {
		t.Fatalf("reconnect snapshot = %#v", corrected)
	}
	replayed, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: identity, Nonce: testReconnectNonce})
	if err != nil || replayed.Credential != corrected.Credential || replayed.Snapshot.Checksum != corrected.Snapshot.Checksum {
		t.Fatalf("replayed reconnect = %#v, %v", replayed, err)
	}
	if _, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: identity, Nonce: "fedcba9876543210fedcba9876543210"}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("stale reconnect with another nonce error = %v", err)
	}
	if err := endpoint.Submit(joined.Credential, CommandIntent{}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("stale credential error = %v", err)
	}
	if err := endpoint.Leave(corrected.Credential); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoint.Reconnect(ReconnectRequest{Credential: corrected.Credential, Identity: identity, Nonce: testReconnectNonce}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("revoked credential error = %v", err)
	}
}

func TestEndpointAcknowledgesExactRetransmitAndRejectsSequenceConflict(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, _ := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	if err := session.Register("player.move", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}
	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation},
		testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player"}},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}
	joined, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	intent := CommandIntent{TargetTick: 1, Sequence: 1, Kind: "player.move", Payload: json.RawMessage(`{"x":1}`)}
	if err := endpoint.Submit(joined.Credential, intent); err != nil {
		t.Fatal(err)
	}
	if err := endpoint.Submit(joined.Credential, intent); err != nil {
		t.Fatalf("exact retransmit = %v", err)
	}
	conflict := intent
	conflict.Payload = json.RawMessage(`{"x":2}`)
	if err := endpoint.Submit(joined.Credential, conflict); !errors.Is(err, gamesession.ErrCommandSequence) {
		t.Fatalf("conflicting retransmit error = %v", err)
	}
}

func TestEndpointRejectsUntrustedAndIncompatibleClients(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, err := gamesession.Allocate("game-1", identity, gamesession.PredictionNone)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	endpoint, err := NewEndpoint(
		&Host{Engine: engine, Session: session, Allocation: allocation},
		testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player"}},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion + 1, Credential: "valid", Identity: identity}); !errors.Is(err, ErrProtocol) {
		t.Fatalf("future protocol error = %v", err)
	}
	if _, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "invalid", Identity: identity}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("authentication error = %v", err)
	}
	joined, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	if err := endpoint.Submit(SessionCredential("forged"), CommandIntent{}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("forged session credential error = %v", err)
	}
	mismatch := identity
	mismatch.Recipe.Packages.Base.Digest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if _, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: mismatch, Nonce: testReconnectNonce}); !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("incompatible reconnect error = %v", err)
	}
}

func TestSessionProtocolVersionMatchesCanonicalRuntimeRecipe(t *testing.T) {
	want := fmt.Sprintf("dark-magic.game-session/v%d", SessionProtocolVersion)
	if simulation.RuntimeNetworkProtocol != want {
		t.Fatalf("runtime network protocol = %q, want %q", simulation.RuntimeNetworkProtocol, want)
	}
}

func TestEndpointRejectsTicketPinnedToAnotherRuntime(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, err := gamesession.Allocate("game-1", identity, gamesession.PredictionNone)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation},
		testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player", RuntimeIdentityHash: "another-runtime"}},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("runtime-bound ticket error = %v", err)
	}
}

func TestEndpointWaitsForTrustedPlayerAdmissionProjection(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionNone)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	pending := errors.New("player pending")
	attempts := 0
	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation},
		testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player"}},
		func(string, simulation.Checkpoint) (json.RawMessage, error) {
			attempts++
			if attempts < 3 {
				return nil, pending
			}
			return json.RawMessage(`{}`), nil
		})
	if err != nil {
		t.Fatal(err)
	}
	endpoint.SetSnapshotPending(func(err error) bool { return errors.Is(err, pending) })
	if _, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity}); err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("projection attempts = %d", attempts)
	}
}

func TestEndpointRateLimitsPerMembershipAndRefills(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionNone)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation},
		testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player"}},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	endpoint.now = func() time.Time { return now }
	joined, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	for range refreshBurst {
		if _, err := endpoint.Refresh(joined.Credential); err != nil {
			t.Fatal(err)
		}
	}
	// Server-paced observation has its own correction ticker and must remain
	// available after the client-paced unary refresh budget is exhausted.
	for range refreshBurst * 3 {
		if _, err := endpoint.Observe(joined.Credential); err != nil {
			t.Fatalf("server-paced observation consumed refresh budget: %v", err)
		}
	}
	if _, err := endpoint.Refresh(joined.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("rate error = %v", err)
	}
	rotated, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: identity, Nonce: testReconnectNonce})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := endpoint.Refresh(rotated.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("rotated credential reset rate budget: %v", err)
	}
	now = now.Add(time.Second)
	for range int(refreshRate) {
		if _, err := endpoint.Refresh(rotated.Credential); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := endpoint.Refresh(rotated.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("refill rate error = %v", err)
	}
}

func TestEndpointRateLimitIsRaceSafeUnderBurst(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionNone)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation}, testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player"}}, func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}
	endpoint.now = func() time.Time { return time.Unix(100, 0) }
	joined, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	results := make(chan error, 64)
	for range 64 {
		wait.Add(1)
		go func() { defer wait.Done(); _, err := endpoint.Refresh(joined.Credential); results <- err }()
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

func TestEndpointAllowsOneCorrectionWatchPerMembership(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, _ := gamesession.Allocate("game", identity, gamesession.PredictionNone)
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation},
		testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player"}},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}
	joined, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	if err := endpoint.BeginWatch(joined.Credential); err != nil {
		t.Fatal(err)
	}
	if err := endpoint.BeginWatch(joined.Credential); !errors.Is(err, ErrRateLimit) {
		t.Fatalf("second watch error = %v", err)
	}
	endpoint.EndWatch(joined.Credential)
	if err := endpoint.BeginWatch(joined.Credential); err != nil {
		t.Fatalf("replacement watch error = %v", err)
	}
}

func TestEndpointKeepsDisconnectedMembershipReconnectableUntilLeaseExpires(t *testing.T) {
	identity := testProtocolIdentity()
	allocation, _ := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	endpoint, err := NewEndpoint(&Host{Engine: engine, Session: session, Allocation: allocation},
		testAuthenticator{credential: "valid", principal: Principal{ID: "account", CharacterID: "character", PlayerID: "player"}},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}
	var expirations []func()
	endpoint.after = func(_ time.Duration, expire func()) { expirations = append(expirations, expire) }
	leaves := 0
	endpoint.SetLeave(func(Principal) error { leaves++; return nil })

	joined, err := endpoint.Join(context.Background(), JoinRequest{Version: SessionProtocolVersion, Credential: "valid", Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	endpoint.Disconnect(joined.Credential)
	if err := endpoint.Submit(joined.Credential, CommandIntent{}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("suspended submit error = %v", err)
	}
	if leaves != 0 || len(expirations) != 1 {
		t.Fatalf("disconnect leaves=%d expirations=%d", leaves, len(expirations))
	}
	reconnected, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: identity, Nonce: testReconnectNonce})
	if err != nil {
		t.Fatal(err)
	}
	expirations[0]()
	if leaves != 0 {
		t.Fatalf("stale lease expiration removed reconnected player")
	}

	endpoint.Disconnect(reconnected.Credential)
	if len(expirations) != 3 {
		t.Fatalf("expirations = %d", len(expirations))
	}
	// Reconnect replay expiry is independent of membership lease expiry.
	expirations[1]()
	expirations[2]()
	if leaves != 1 {
		t.Fatalf("expired disconnect leaves = %d", leaves)
	}
	if _, err := endpoint.Reconnect(ReconnectRequest{Credential: reconnected.Credential, Identity: identity, Nonce: testReconnectNonce}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("expired reconnect error = %v", err)
	}
}

func testProtocolIdentity() simulation.RuntimeIdentity {
	const packageDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1",
		Packages:          simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0", Digest: packageDigest, Size: 1, Redistributable: true}},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}}
}

const testReconnectNonce = "0123456789abcdef0123456789abcdef"
