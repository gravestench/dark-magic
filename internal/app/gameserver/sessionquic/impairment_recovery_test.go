package sessionquic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// impairedRecoveryFixture owns one client/server path with independent deterministic fault profiles.
type impairedRecoveryFixture struct {
	ctx           context.Context
	identity      simulation.RuntimeIdentity
	session       *gamesession.Session
	client        *Client
	clientPackets *impairedPacketConn
	serverPackets *impairedPacketConn
}

// newImpairedRecoveryFixture runs the production endpoint and QUIC stack over fault-injecting UDP sockets.
func newImpairedRecoveryFixture(t *testing.T) impairedRecoveryFixture {
	t.Helper()

	identity := testRuntimeIdentity()
	session, endpoint := newImpairedRecoveryEndpoint(t, identity)
	serverAddress, serverPackets, clientTLS := startImpairedRecoveryServer(t, endpoint)
	ctx, client, clientPackets := dialImpairedRecoveryClient(t, serverAddress, clientTLS)

	return impairedRecoveryFixture{
		ctx:           ctx,
		identity:      identity,
		session:       session,
		client:        client,
		clientPackets: clientPackets,
		serverPackets: serverPackets,
	}
}

// newImpairedRecoveryEndpoint builds the real authority whose tick progression the test observes.
func newImpairedRecoveryEndpoint(
	t *testing.T,
	identity simulation.RuntimeIdentity,
) (*gamesession.Session, *gameserver.Endpoint) {
	t.Helper()

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
	registerMoveCommand(t, session)

	endpoint, err := gameserver.NewEndpoint(
		&gameserver.Host{Engine: engine, Session: session, Allocation: allocation},
		authenticator{},
		recoveryProjection,
	)
	if err != nil {
		t.Fatal(err)
	}

	return session, endpoint
}

// startImpairedRecoveryServer applies the server fault profile and starts the production accept loop.
func startImpairedRecoveryServer(
	t *testing.T,
	endpoint *gameserver.Endpoint,
) (net.Addr, *impairedPacketConn, *tls.Config) {
	t.Helper()

	serverSocket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	serverPackets := &impairedPacketConn{
		PacketConn: serverSocket,
		profile: impairmentProfile{
			dropEvery: 5,
			delays:    []time.Duration{0, 2 * time.Millisecond, 5 * time.Millisecond},
		},
	}

	serverTLS, clientTLS := testTLS(t)

	server, err := ListenPacket(serverPackets, serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = server.Close() })

	serveContext, stopServer := context.WithCancel(context.Background())
	t.Cleanup(stopServer)

	go func() { _ = server.Serve(serveContext) }()

	return serverSocket.LocalAddr(), serverPackets, clientTLS
}

// dialImpairedRecoveryClient applies the independent client profile and transfers its socket to QUIC.
func dialImpairedRecoveryClient(
	t *testing.T,
	serverAddress net.Addr,
	clientTLS *tls.Config,
) (context.Context, *Client, *impairedPacketConn) {
	t.Helper()

	clientSocket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	clientPackets := &impairedPacketConn{
		PacketConn: clientSocket,
		profile: impairmentProfile{
			dropEvery: 7,
			delays:    []time.Duration{3 * time.Millisecond, 0, 7 * time.Millisecond},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	client, err := DialPacket(ctx, clientPackets, serverAddress, clientTLS)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = client.Close() })

	return ctx, client, clientPackets
}

// recoveryProjection makes authoritative tick movement visible in both correction and transform assertions.
func recoveryProjection(player string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	view := playeradapter.ClientView{
		Version: playeradapter.ClientViewVersion,
		Tick:    checkpoint.Tick,
		HUD: playeradapter.HUD{
			Version:  playeradapter.HUDVersion,
			Tick:     checkpoint.Tick,
			Player:   playeradapter.HUDIdentity{PlayerID: player},
			Position: playeradapter.HUDPosition{X: float64(checkpoint.Tick)},
		},
		World: playeradapter.WorldView{
			Version:  playeradapter.WorldViewVersion,
			Tick:     checkpoint.Tick,
			Entities: []playeradapter.WorldEntity{},
		},
		Private: playeradapter.PrivateView{
			Version: playeradapter.PrivateViewVersion,
			Tick:    checkpoint.Tick,
		},
		Party: playeradapter.PartyView{
			Version: playeradapter.PartyViewVersion,
			Tick:    checkpoint.Tick,
			Roster: []playeradapter.PartyRosterEntry{{
				PlayerID:     player,
				Name:         player,
				Class:        "Amazon",
				Level:        1,
				Relationship: "self",
			}},
		},
		Events: playeradapter.EventView{
			Version: playeradapter.EventViewVersion,
			Tick:    checkpoint.Tick,
			Events:  []playeradapter.SemanticEvent{},
		},
	}

	return json.Marshal(view)
}

// TestReliableSessionRecoversFromDelayJitterAndPacketLoss verifies durable state and reconnect under active faults.
func TestReliableSessionRecoversFromDelayJitterAndPacketLoss(t *testing.T) {
	fixture := newImpairedRecoveryFixture(t)
	joined := joinImpairedRecoveryClient(t, fixture)
	stopWatch, transforms, transformErrors := startImpairedRecoveryWatch(t, fixture, joined.Credential)

	submitAndAdvanceImpairedSession(t, fixture, joined.Credential)
	assertImpairedTransformArrives(t, fixture.ctx, transforms, transformErrors)
	stopWatch()

	assertImpairedReconnectAndLeave(t, fixture, joined)
	assertImpairmentApplied(t, "client", fixture.clientPackets.profile, fixture.clientPackets.stats())
	assertImpairmentApplied(t, "server", fixture.serverPackets.profile, fixture.serverPackets.stats())
}

// joinImpairedRecoveryClient performs the ordinary authenticated handshake over the already impaired path.
func joinImpairedRecoveryClient(t *testing.T, fixture impairedRecoveryFixture) gameserver.JoinResponse {
	t.Helper()

	joined, err := fixture.client.Join(fixture.ctx, gameserver.JoinRequest{
		Version:    gameserver.SessionProtocolVersion,
		Credential: "realm-ticket",
		Identity:   fixture.identity,
	})
	if err != nil {
		t.Fatal(err)
	}

	return joined
}

// startImpairedRecoveryWatch requires the initial reliable correction before returning transform channels.
func startImpairedRecoveryWatch(
	t *testing.T,
	fixture impairedRecoveryFixture,
	credential gameserver.SessionCredential,
) (context.CancelFunc, <-chan TransformFrame, <-chan error) {
	t.Helper()

	watchContext, stopWatch := context.WithCancel(fixture.ctx)

	snapshots, watchErrors, err := fixture.client.Watch(watchContext, credential)
	if err != nil {
		t.Fatal(err)
	}

	transforms, transformErrors, err := fixture.client.WatchTransforms(watchContext, credential)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case snapshot := <-snapshots:
		if snapshot.Tick != 0 {
			t.Fatalf("initial impaired correction tick = %d", snapshot.Tick)
		}
	case streamErr := <-watchErrors:
		t.Fatalf("impaired correction stream: %v", streamErr)
	case <-fixture.ctx.Done():
		t.Fatal(fixture.ctx.Err())
	}

	return stopWatch, transforms, transformErrors
}

// submitAndAdvanceImpairedSession paces ticks so the datagram loop has opportunities between canonical checkpoints.
func submitAndAdvanceImpairedSession(
	t *testing.T,
	fixture impairedRecoveryFixture,
	credential gameserver.SessionCredential,
) {
	t.Helper()

	if err := fixture.client.Submit(fixture.ctx, credential, gameserver.CommandIntent{
		TargetTick: 1,
		Sequence:   1,
		Kind:       "move",
		Payload:    json.RawMessage(`{}`),
	}); err != nil {
		t.Fatal(err)
	}

	for range 8 {
		if err := fixture.session.Step(); err != nil {
			t.Fatal(err)
		}

		time.Sleep(45 * time.Millisecond)
	}
}

// assertImpairedTransformArrives waits through disposable losses until a newer credential-bound sample arrives.
func assertImpairedTransformArrives(
	t *testing.T,
	ctx context.Context,
	transforms <-chan TransformFrame,
	transformErrors <-chan error,
) {
	t.Helper()

	for {
		select {
		case transformed := <-transforms:
			if transformed.Tick == 0 {
				continue
			}

			if transformed.OwnerX != float64(transformed.Tick) {
				t.Fatalf("impaired transform frame = %#v", transformed)
			}

			return
		case streamErr := <-transformErrors:
			t.Fatalf("impaired transform stream: %v", streamErr)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}

// assertImpairedReconnectAndLeave proves reliable rotation catches up after the eight paced authority steps.
func assertImpairedReconnectAndLeave(
	t *testing.T,
	fixture impairedRecoveryFixture,
	joined gameserver.JoinResponse,
) {
	t.Helper()

	corrected, err := fixture.client.Reconnect(fixture.ctx, gameserver.ReconnectRequest{
		Credential: joined.Credential,
		Identity:   fixture.identity,
		Nonce:      "0123456789abcdef0123456789abcdef",
	})
	if err != nil {
		t.Fatal(err)
	}

	if corrected.Snapshot.Tick != 8 || corrected.Credential == joined.Credential {
		t.Fatalf("reconnect after impairment = %#v", corrected)
	}

	if err := fixture.client.Leave(fixture.ctx, corrected.Credential); err != nil {
		t.Fatal(err)
	}
}
