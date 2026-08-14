package sessionquic

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type impairedPacketConn struct {
	net.PacketConn
	mu            sync.Mutex
	writes        int
	dropped       int
	delayed       int
	injectedDelay time.Duration
	profile       impairmentProfile
}

type impairmentProfile struct {
	dropEvery int
	delays    []time.Duration
}

type impairmentStats struct {
	writes, dropped, delayed int
	injectedDelay            time.Duration
}

func (connection *impairedPacketConn) WriteTo(payload []byte, address net.Addr) (int, error) {
	connection.mu.Lock()
	connection.writes++
	sequence := connection.writes
	drop := connection.profile.dropEvery > 0 && sequence%connection.profile.dropEvery == 0
	if drop {
		connection.dropped++
	}
	var delay time.Duration
	if len(connection.profile.delays) > 0 {
		delay = connection.profile.delays[(sequence-1)%len(connection.profile.delays)]
	}
	if delay > 0 {
		connection.delayed++
		connection.injectedDelay += delay
	}
	connection.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}
	if drop {
		return len(payload), nil
	}
	return connection.PacketConn.WriteTo(payload, address)
}

func (connection *impairedPacketConn) stats() impairmentStats {
	connection.mu.Lock()
	defer connection.mu.Unlock()
	return impairmentStats{
		writes: connection.writes, dropped: connection.dropped,
		delayed: connection.delayed, injectedDelay: connection.injectedDelay,
	}
}

func TestReliableSessionRecoversFromDelayJitterAndPacketLoss(t *testing.T) {
	identity := simulation.RuntimeIdentity{
		ModID: "d2legacy", ContractVersion: "v1", PackageHash: "package",
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}
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
	if err := session.Register("move", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}
	endpoint, err := gameserver.NewEndpoint(
		&gameserver.Host{Engine: engine, Session: session, Allocation: allocation},
		authenticator{},
		func(player string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
			view := playeradapter.ClientView{
				Version: playeradapter.ClientViewVersion, Tick: checkpoint.Tick,
				HUD: playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: checkpoint.Tick,
					Player: playeradapter.HUDIdentity{PlayerID: player}, Position: playeradapter.HUDPosition{X: float64(checkpoint.Tick)}},
				World:   playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: checkpoint.Tick, Entities: []playeradapter.WorldEntity{}},
				Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: checkpoint.Tick},
			}
			return json.Marshal(view)
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	serverSocket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serverPackets := &impairedPacketConn{
		PacketConn: serverSocket,
		profile: impairmentProfile{dropEvery: 5,
			delays: []time.Duration{0, 2 * time.Millisecond, 5 * time.Millisecond}},
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

	clientSocket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	clientPackets := &impairedPacketConn{
		PacketConn: clientSocket,
		profile: impairmentProfile{dropEvery: 7,
			delays: []time.Duration{3 * time.Millisecond, 0, 7 * time.Millisecond}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := DialPacket(ctx, clientPackets, serverSocket.LocalAddr(), clientTLS)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	joined, err := client.Join(ctx, gameserver.JoinRequest{
		Version: gameserver.SessionProtocolVersion, Credential: "realm-ticket", Identity: identity,
	})
	if err != nil {
		t.Fatal(err)
	}
	watchContext, stopWatch := context.WithCancel(ctx)
	snapshots, watchErrors, err := client.Watch(watchContext, joined.Credential)
	if err != nil {
		t.Fatal(err)
	}
	transforms, transformErrors, err := client.WatchTransforms(watchContext, joined.Credential)
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
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := client.Submit(ctx, joined.Credential, gameserver.CommandIntent{
		TargetTick: 1, Sequence: 1, Kind: "move", Payload: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(45 * time.Millisecond)
	}
	var transformed TransformFrame
	waiting := true
	for waiting {
		select {
		case transformed = <-transforms:
			if transformed.Tick > 0 {
				waiting = false
			}
		case streamErr := <-transformErrors:
			t.Fatalf("impaired transform stream: %v", streamErr)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	if transformed.OwnerX != float64(transformed.Tick) {
		t.Fatalf("impaired transform frame = %#v", transformed)
	}
	stopWatch()
	corrected, err := client.Reconnect(ctx, gameserver.ReconnectRequest{
		Credential: joined.Credential, Identity: identity, Nonce: "0123456789abcdef0123456789abcdef",
	})
	if err != nil {
		t.Fatal(err)
	}
	if corrected.Snapshot.Tick != 8 || corrected.Credential == joined.Credential {
		t.Fatalf("reconnect after impairment = %#v", corrected)
	}
	if err := client.Leave(ctx, corrected.Credential); err != nil {
		t.Fatal(err)
	}

	assertImpairmentApplied(t, "client", clientPackets.profile, clientPackets.stats())
	assertImpairmentApplied(t, "server", serverPackets.profile, serverPackets.stats())
}

func assertImpairmentApplied(t *testing.T, name string, profile impairmentProfile, stats impairmentStats) {
	t.Helper()
	expectedDelayed := 0
	var expectedDelay time.Duration
	for sequence := 1; sequence <= stats.writes; sequence++ {
		delay := profile.delays[(sequence-1)%len(profile.delays)]
		if delay > 0 {
			expectedDelayed++
			expectedDelay += delay
		}
	}
	if stats.writes == 0 || stats.dropped != stats.writes/profile.dropEvery ||
		stats.delayed != expectedDelayed || stats.injectedDelay != expectedDelay {
		t.Fatalf("%s synthetic impairment mismatch: profile=%#v stats=%#v", name, profile, stats)
	}
}
