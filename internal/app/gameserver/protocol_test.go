package gameserver

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

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
		Tick: 1, Sequence: 1, Kind: "player.move", Payload: json.RawMessage(`{"x":4}`),
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
	corrected, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	if corrected.Credential == joined.Credential || corrected.Snapshot.Tick != 1 || corrected.Snapshot.Checksum == "" {
		t.Fatalf("reconnect snapshot = %#v", corrected)
	}
	if err := endpoint.Submit(joined.Credential, CommandIntent{}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("stale credential error = %v", err)
	}
	if err := endpoint.Leave(corrected.Credential); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoint.Reconnect(ReconnectRequest{Credential: corrected.Credential, Identity: identity}); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("revoked credential error = %v", err)
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
	mismatch.PackageHash = "different"
	if _, err := endpoint.Reconnect(ReconnectRequest{Credential: joined.Credential, Identity: mismatch}); !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("incompatible reconnect error = %v", err)
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

func testProtocolIdentity() simulation.RuntimeIdentity {
	return simulation.RuntimeIdentity{
		ModID: "d2legacy", ContractVersion: "v1", PackageHash: "package",
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}
}
