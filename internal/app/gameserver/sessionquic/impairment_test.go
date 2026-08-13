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
		func(player string, _ simulation.Checkpoint) (json.RawMessage, error) {
			return json.Marshal(map[string]string{"player": player})
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
	stopWatch()
	if err := client.Submit(ctx, joined.Credential, gameserver.CommandIntent{
		Tick: 1, Sequence: 1, Kind: "move", Payload: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	corrected, err := client.Reconnect(ctx, gameserver.ReconnectRequest{
		Credential: joined.Credential, Identity: identity,
	})
	if err != nil {
		t.Fatal(err)
	}
	if corrected.Snapshot.Tick != 1 || corrected.Credential == joined.Credential {
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
