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
	mu        sync.Mutex
	writes    int
	dropped   int
	dropEvery int
	delays    []time.Duration
}

func (connection *impairedPacketConn) WriteTo(payload []byte, address net.Addr) (int, error) {
	connection.mu.Lock()
	connection.writes++
	sequence := connection.writes
	drop := connection.dropEvery > 0 && sequence%connection.dropEvery == 0
	if drop {
		connection.dropped++
	}
	var delay time.Duration
	if len(connection.delays) > 0 {
		delay = connection.delays[(sequence-1)%len(connection.delays)]
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

func (connection *impairedPacketConn) counts() (writes, dropped int) {
	connection.mu.Lock()
	defer connection.mu.Unlock()
	return connection.writes, connection.dropped
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
		PacketConn: serverSocket, dropEvery: 5,
		delays: []time.Duration{0, 2 * time.Millisecond, 5 * time.Millisecond},
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
		PacketConn: clientSocket, dropEvery: 7,
		delays: []time.Duration{3 * time.Millisecond, 0, 7 * time.Millisecond},
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

	clientWrites, clientDrops := clientPackets.counts()
	serverWrites, serverDrops := serverPackets.counts()
	if clientWrites == 0 || serverWrites == 0 || clientDrops == 0 || serverDrops == 0 {
		t.Fatalf("impairment was not exercised: client=%d/%d server=%d/%d", clientDrops, clientWrites, serverDrops, serverWrites)
	}
}
