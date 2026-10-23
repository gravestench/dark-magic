package sessionquic

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type readinessAuthenticator struct{}

func (readinessAuthenticator) Authenticate(_ context.Context, credential string) (gameserver.Principal, error) {
	if credential != "ready" && credential != "pending" {
		return gameserver.Principal{}, gameserver.ErrAuthentication
	}

	return gameserver.Principal{ID: credential, CharacterID: credential, PlayerID: credential}, nil
}

func newReadinessServer(t *testing.T, project gameserver.SnapshotProjector) (*Server, simulation.RuntimeIdentity) {
	t.Helper()

	identity := testRuntimeIdentity()

	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}

	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })

	endpoint, err := gameserver.NewEndpoint(
		&gameserver.Host{Engine: engine, Session: session, Allocation: allocation}, readinessAuthenticator{}, project)
	if err != nil {
		t.Fatal(err)
	}

	return &Server{endpoint: endpoint}, identity
}

func TestPendingJoinDoesNotBlockOtherSocketMemberships(t *testing.T) {
	pendingError := errors.New("admission pending")
	entered := make(chan struct{})

	var once sync.Once

	server, identity := newReadinessServer(t,
		func(player string, _ simulation.Checkpoint) (json.RawMessage, error) {
			if player == "pending" {
				once.Do(func() { close(entered) })
				return nil, pendingError
			}

			return json.RawMessage(`{}`), nil
		})

	server.endpoint.SetSnapshotPending(func(err error) bool { return errors.Is(err, pendingError) })

	original, pending, replacement := newConnectionMemberships(), newConnectionMemberships(), newConnectionMemberships()
	join := request{Operation: operationJoin, Join: &gameserver.JoinRequest{
		Version: gameserver.SessionProtocolVersion, Credential: "ready", Identity: identity,
	}}

	joined := server.dispatchConnection(t.Context(), join, original)
	if joined.Error != "" || joined.Join == nil {
		t.Fatalf("join = %#v", joined)
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	joinDone := make(chan struct{})

	go func() {
		pendingJoin := *join.Join
		pendingJoin.Credential = "pending"
		server.dispatchConnection(ctx, request{Operation: operationJoin, Join: &pendingJoin}, pending)
		close(joinDone)
	}()

	defer func() { cancel(); <-joinDone }()

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("pending join did not reach readiness")
	}

	recoveryDone := make(chan response, 1)
	go func() {
		recoveryDone <- server.dispatchConnection(t.Context(), request{
			Operation: operationReconnect, Reconnect: &gameserver.ReconnectRequest{
				Credential: joined.Join.Credential, Identity: identity,
				Nonce: "0123456789abcdef0123456789abcdef",
			},
		}, replacement)
	}()

	var recovered response
	select {
	case recovered = <-recoveryDone:
		if recovered.Error != "" || recovered.Join == nil {
			t.Fatalf("reconnect = %#v", recovered)
		}
	case <-time.After(100 * time.Millisecond):
		cancel()
		<-recoveryDone
		t.Fatal("another socket's reconnect waited for pending join readiness")
	}

	cleanupDone := make(chan struct{})

	go func() { server.disconnectMemberships(original); close(cleanupDone) }()

	select {
	case <-cleanupDone:
	case <-time.After(100 * time.Millisecond):
		cancel()
		<-cleanupDone
		t.Fatal("another socket's cleanup waited for pending join readiness")
	}

	if _, err := server.endpoint.Refresh(recovered.Join.Credential); err != nil {
		t.Fatalf("replacement lost membership: %v", err)
	}

	server.disconnectMemberships(replacement)
}

func TestPendingJoinCannotPublishPastItsSocketCleanup(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})

	var once sync.Once

	server, identity := newReadinessServer(t, func(_ string, _ simulation.Checkpoint) (json.RawMessage, error) {
		once.Do(func() { close(entered) })
		<-release

		return json.RawMessage(`{}`), nil
	})
	memberships := newConnectionMemberships()

	joined := make(chan response, 1)
	go func() {
		joined <- server.dispatchConnection(t.Context(), request{Operation: operationJoin,
			Join: &gameserver.JoinRequest{Version: gameserver.SessionProtocolVersion,
				Credential: "pending", Identity: identity}}, memberships)
	}()

	select {
	case <-entered:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("admission did not reach its projector")
	}

	closed := make(chan struct{})

	go func() { server.disconnectMemberships(memberships); close(closed) }()

	select {
	case <-closed:
		close(release)
		<-joined
		t.Fatal("socket cleanup ran before its in-flight admission could be accounted for")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	result := <-joined

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("socket cleanup did not finish after admission")
	}

	if result.Error != "" || result.Join == nil {
		t.Fatalf("join = %#v", result)
	}

	if _, err := server.endpoint.Refresh(result.Join.Credential); err != gameserver.ErrAuthentication {
		t.Fatalf("closed socket left an active published credential: %v", err)
	}
}
