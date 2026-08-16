package sessionquic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
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
	pending       sync.WaitGroup
	closing       bool
	writes        int
	dropped       int
	delayed       int
	reordered     int
	injectedDelay time.Duration
	profile       impairmentProfile
}

type impairmentProfile struct {
	dropEvery    int
	delays       []time.Duration
	reorderEvery int
	reorderDelay time.Duration
}

type impairmentStats struct {
	writes, dropped, delayed, reordered int
	injectedDelay                       time.Duration
}

type soakAuthenticator map[string]gameserver.Principal

func (authenticator soakAuthenticator) Authenticate(_ context.Context, credential string) (gameserver.Principal, error) {
	principal, found := authenticator[credential]
	if !found {
		return gameserver.Principal{}, gameserver.ErrAuthentication
	}
	return principal, nil
}

type soakMember struct {
	transport        *Client
	packets          *impairedPacketConn
	credential       gameserver.SessionCredential
	sequence         uint64
	stopStreams      context.CancelFunc
	snapshots        <-chan gameserver.Snapshot
	snapshotErrors   <-chan error
	transforms       <-chan TransformFrame
	transformErrors  <-chan error
	latestCorrection uint64
	latestTransform  uint64
}

func (connection *impairedPacketConn) WriteTo(payload []byte, address net.Addr) (int, error) {
	connection.mu.Lock()
	if connection.closing {
		connection.mu.Unlock()
		return 0, net.ErrClosed
	}
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
	reorder := !drop && connection.profile.reorderEvery > 0 && sequence%connection.profile.reorderEvery == 0
	if reorder {
		connection.reordered++
		delay += connection.profile.reorderDelay
	}
	if delay > 0 {
		connection.delayed++
		connection.injectedDelay += delay
	}
	if reorder {
		// Add while holding the same mutex used by wait. Once wait marks the
		// connection closing, no writer can add work concurrently with Wait.
		connection.pending.Add(1)
	}
	connection.mu.Unlock()

	if reorder {
		copy := append([]byte(nil), payload...)
		go func() {
			defer connection.pending.Done()
			time.Sleep(delay)
			_, _ = connection.PacketConn.WriteTo(copy, address)
		}()
		return len(payload), nil
	}
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
		delayed: connection.delayed, reordered: connection.reordered, injectedDelay: connection.injectedDelay,
	}
}

func (connection *impairedPacketConn) wait() {
	connection.mu.Lock()
	connection.closing = true
	connection.mu.Unlock()
	connection.pending.Wait()
}

func TestReliableSessionRecoversFromDelayJitterAndPacketLoss(t *testing.T) {
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

func TestSustainedMultiClientSessionSurvivesImpairmentAndActiveReconnect(t *testing.T) {
	const memberCount = 3
	activeTicks := networkSoakTicks(t)
	identity := testRuntimeIdentity()
	allocation, err := gamesession.Allocate("soak", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	applied := make(map[string]int)
	if err := session.Register("move", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply: func(_ *gameecs.Engine, command simulation.Command) error {
			applied[command.Player]++
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	authentication := make(soakAuthenticator, memberCount)
	for index := 0; index < memberCount; index++ {
		authentication[fmt.Sprintf("ticket-%d", index)] = gameserver.Principal{
			ID: fmt.Sprintf("account-%d", index), CharacterID: fmt.Sprintf("character-%d", index),
			PlayerID: fmt.Sprintf("player-%d", index), RuntimeIdentityHash: allocation.IdentityHash,
		}
	}
	endpoint, err := gameserver.NewEndpoint(
		&gameserver.Host{Engine: engine, Session: session, Allocation: allocation},
		authentication,
		func(player string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
			view := playeradapter.ClientView{
				Version: playeradapter.ClientViewVersion, Tick: checkpoint.Tick,
				HUD: playeradapter.HUD{
					Version: playeradapter.HUDVersion, Tick: checkpoint.Tick,
					Player:   playeradapter.HUDIdentity{PlayerID: player, CharacterID: "character", Name: player, Class: "Amazon"},
					Vitals:   playeradapter.HUDVitals{Health: 1, MaxHealth: 1, Mana: 1, MaxMana: 1},
					Position: playeradapter.HUDPosition{X: float64(checkpoint.Tick)}, Belt: playeradapter.HUDBelt{Slots: []string{}},
				},
				World: playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: checkpoint.Tick, Entities: []playeradapter.WorldEntity{}},
				Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: checkpoint.Tick,
					Items: playeradapter.ItemView{Items: []playeradapter.ItemEntityView{}}},
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
	serverPackets := &impairedPacketConn{PacketConn: serverSocket, profile: impairmentProfile{
		dropEvery: 17, delays: []time.Duration{0, time.Millisecond, 3 * time.Millisecond},
		reorderEvery: 13, reorderDelay: 8 * time.Millisecond,
	}}
	serverTLS, clientTLS := testTLS(t)
	server, err := ListenPacket(serverPackets, serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	serveContext, stopServer := context.WithCancel(context.Background())
	go func() { _ = server.Serve(serveContext) }()
	t.Cleanup(func() {
		stopServer()
		_ = server.Close()
		serverPackets.wait()
		_ = serverSocket.Close()
	})

	// The soak advances at the production simulation rate, so its deadline must
	// grow with the requested tick count. Keep fixed headroom for connection,
	// impairment, reconnect, and cleanup work around the paced simulation.
	soakDeadline := time.Duration(activeTicks)*session.StepDuration() + 20*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), soakDeadline)
	defer cancel()
	members := make([]*soakMember, memberCount)
	allClientPackets := make([]*impairedPacketConn, 0, memberCount+1)
	for index := range members {
		member, err := dialSoakMember(ctx, t, serverSocket.LocalAddr(), clientTLS, impairmentProfile{
			dropEvery: 19 + index*2, delays: []time.Duration{time.Millisecond, 0, 2 * time.Millisecond},
			reorderEvery: 11 + index*2, reorderDelay: 7 * time.Millisecond,
		})
		if err != nil {
			t.Fatal(err)
		}
		joined, err := member.transport.Join(ctx, gameserver.JoinRequest{
			Version: gameserver.SessionProtocolVersion, Credential: fmt.Sprintf("ticket-%d", index), Identity: identity,
		})
		if err != nil {
			t.Fatal(err)
		}
		member.credential = joined.Credential
		if err := member.startStreams(ctx); err != nil {
			t.Fatal(err)
		}
		members[index] = member
		allClientPackets = append(allClientPackets, member.packets)
	}
	t.Cleanup(func() {
		for _, member := range members {
			member.stop(ctx)
		}
	})

	for step := 0; step < activeTicks; step++ {
		stepStarted := time.Now()
		if step == activeTicks/2 {
			if err := reconnectSoakMember(ctx, members[1], identity); err != nil {
				t.Fatal(err)
			}
		}
		if step == 3*activeTicks/4 {
			oldPackets := members[1].packets
			if err := redialSoakMember(ctx, t, members[1], serverSocket.LocalAddr(), clientTLS, identity); err != nil {
				t.Fatal(err)
			}
			allClientPackets = append(allClientPackets, members[1].packets)
			oldPackets.wait()
		}
		current := session.Status().Tick
		for _, member := range members {
			member.sequence++
			if err := member.transport.Submit(ctx, member.credential, gameserver.CommandIntent{
				ObservedServerTick: current, TargetTick: current + 2, Sequence: member.sequence,
				Kind: "move", Payload: json.RawMessage(`{}`),
			}); err != nil {
				t.Fatalf("tick %d sequence %d: %v", current, member.sequence, err)
			}
		}
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
		if remaining := session.StepDuration() - time.Since(stepStarted); remaining > 0 {
			time.Sleep(remaining)
		}
		for _, member := range members {
			if err := member.drainStreams(); err != nil {
				t.Fatal(err)
			}
		}
	}
	for range 2 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	finalTick := session.Status().Tick
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for !soakCaughtUp(members, finalTick) {
		select {
		case <-deadline.C:
			t.Fatalf("network soak did not converge by tick %d: %#v", finalTick, members)
		case <-time.After(10 * time.Millisecond):
			for _, member := range members {
				if err := member.drainStreams(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	for index, member := range members {
		player := fmt.Sprintf("player-%d", index)
		if applied[player] != activeTicks {
			t.Fatalf("%s applied commands = %d, want %d", player, applied[player], activeTicks)
		}
		stats := member.transport.NetworkStats()
		if stats.SmoothedRTT <= 0 || stats.TransformsReceived == 0 {
			t.Fatalf("member %d network stats = %#v", index, stats)
		}
	}
	assertImpairmentApplied(t, "server soak", serverPackets.profile, serverPackets.stats())
	for index, packets := range allClientPackets {
		assertImpairmentApplied(t, fmt.Sprintf("client soak %d", index), packets.profile, packets.stats())
	}
}

func dialSoakMember(ctx context.Context, t *testing.T, address net.Addr, tlsConfig *tls.Config, profile impairmentProfile) (*soakMember, error) {
	t.Helper()
	socket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	packets := &impairedPacketConn{PacketConn: socket, profile: profile}
	transport, err := DialPacket(ctx, packets, address, tlsConfig)
	if err != nil {
		_ = socket.Close()
		return nil, err
	}
	return &soakMember{transport: transport, packets: packets}, nil
}

func (member *soakMember) startStreams(parent context.Context) error {
	streamContext, stop := context.WithCancel(parent)
	snapshots, snapshotErrors, err := member.transport.Watch(streamContext, member.credential)
	if err != nil {
		stop()
		return err
	}
	transforms, transformErrors, err := member.transport.WatchTransforms(streamContext, member.credential)
	if err != nil {
		stop()
		return err
	}
	if cap(snapshots) != 1 || cap(snapshotErrors) != 1 || cap(transforms) != 1 || cap(transformErrors) != 1 {
		stop()
		return fmt.Errorf("network soak: unbounded stream channels")
	}
	member.stopStreams = stop
	member.snapshots, member.snapshotErrors = snapshots, snapshotErrors
	member.transforms, member.transformErrors = transforms, transformErrors
	return nil
}

func (member *soakMember) drainStreams() error {
	for {
		select {
		case snapshot, open := <-member.snapshots:
			if !open {
				member.snapshots = nil
				break
			}
			member.latestCorrection = max(member.latestCorrection, snapshot.Tick)
		default:
			goto transforms
		}
	}
transforms:
	for {
		select {
		case frame, open := <-member.transforms:
			if !open {
				member.transforms = nil
				break
			}
			member.latestTransform = max(member.latestTransform, frame.Tick)
		default:
			goto streamErrors
		}
	}
streamErrors:
	select {
	case err, open := <-member.snapshotErrors:
		if open && err != nil {
			return err
		}
	default:
	}
	select {
	case err, open := <-member.transformErrors:
		if open && err != nil {
			return err
		}
	default:
	}
	return nil
}

func (member *soakMember) stop(ctx context.Context) {
	if member == nil {
		return
	}
	if member.stopStreams != nil {
		member.stopStreams()
		member.stopStreams = nil
	}
	if member.transport != nil {
		leaveContext, cancel := context.WithTimeout(ctx, time.Second)
		_ = member.transport.Leave(leaveContext, member.credential)
		cancel()
		_ = member.transport.Close()
	}
	if member.packets != nil {
		member.packets.wait()
		_ = member.packets.PacketConn.Close()
	}
}

func reconnectSoakMember(ctx context.Context, member *soakMember, identity simulation.RuntimeIdentity) error {
	if member.stopStreams != nil {
		member.stopStreams()
		member.stopStreams = nil
	}
	reconnected, err := member.transport.Reconnect(ctx, gameserver.ReconnectRequest{
		Credential: member.credential, Identity: identity, Nonce: "0123456789abcdef0123456789abcdef",
	})
	if err != nil {
		return err
	}
	member.credential = reconnected.Credential
	return member.startStreams(ctx)
}

func redialSoakMember(ctx context.Context, t *testing.T, member *soakMember, address net.Addr, tlsConfig *tls.Config, identity simulation.RuntimeIdentity) error {
	t.Helper()
	if member.stopStreams != nil {
		member.stopStreams()
		member.stopStreams = nil
	}
	_ = member.transport.Close()
	_ = member.packets.PacketConn.Close()
	replacement, err := dialSoakMember(ctx, t, address, tlsConfig, member.packets.profile)
	if err != nil {
		return err
	}
	reconnected, err := replacement.transport.Reconnect(ctx, gameserver.ReconnectRequest{
		Credential: member.credential, Identity: identity, Nonce: "fedcba9876543210fedcba9876543210",
	})
	if err != nil {
		replacement.stop(ctx)
		return err
	}
	member.transport, member.packets, member.credential = replacement.transport, replacement.packets, reconnected.Credential
	return member.startStreams(ctx)
}

func soakCaughtUp(members []*soakMember, tick uint64) bool {
	for _, member := range members {
		if member.latestCorrection+4 < tick || member.latestTransform+4 < tick {
			return false
		}
	}
	return true
}

func networkSoakTicks(t *testing.T) int {
	t.Helper()
	value := os.Getenv("DARK_MAGIC_NETWORK_SOAK_TICKS")
	if value == "" {
		return 80
	}
	ticks, err := strconv.Atoi(value)
	if err != nil || ticks < 20 || ticks > 15_000 {
		t.Fatalf("DARK_MAGIC_NETWORK_SOAK_TICKS must be between 20 and 15000, got %q", value)
	}
	return ticks
}

func assertImpairmentApplied(t *testing.T, name string, profile impairmentProfile, stats impairmentStats) {
	t.Helper()
	expectedDelayed, expectedReordered := 0, 0
	var expectedDelay time.Duration
	for sequence := 1; sequence <= stats.writes; sequence++ {
		delay := profile.delays[(sequence-1)%len(profile.delays)]
		drop := profile.dropEvery > 0 && sequence%profile.dropEvery == 0
		if !drop && profile.reorderEvery > 0 && sequence%profile.reorderEvery == 0 {
			delay += profile.reorderDelay
			expectedReordered++
		}
		if delay > 0 {
			expectedDelayed++
			expectedDelay += delay
		}
	}
	if stats.writes == 0 || stats.dropped != stats.writes/profile.dropEvery ||
		stats.delayed != expectedDelayed || stats.injectedDelay != expectedDelay ||
		stats.reordered != expectedReordered {
		t.Fatalf("%s synthetic impairment mismatch: profile=%#v stats=%#v", name, profile, stats)
	}
}
