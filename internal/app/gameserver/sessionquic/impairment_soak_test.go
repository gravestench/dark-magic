package sessionquic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// soakAuthenticator binds each deterministic ticket to a distinct player and the fixture's exact runtime identity.
type soakAuthenticator map[string]gameserver.Principal

// Authenticate rejects unknown fixture tickets so multi-client assertions cannot accidentally share identity.
func (authenticator soakAuthenticator) Authenticate(
	_ context.Context,
	credential string,
) (gameserver.Principal, error) {
	principal, found := authenticator[credential]
	if !found {
		return gameserver.Principal{}, gameserver.ErrAuthentication
	}

	return principal, nil
}

// soakMember owns one impaired client, its rotating credential, streams, and latest observed authoritative ticks.
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

// networkSoakFixture owns the shared authority, server path, and accounting used by every member.
type networkSoakFixture struct {
	ctx              context.Context
	cancel           context.CancelFunc
	activeTicks      int
	identity         simulation.RuntimeIdentity
	session          *gamesession.Session
	applied          map[string]int
	serverSocket     net.PacketConn
	serverPackets    *impairedPacketConn
	clientTLS        *tls.Config
	members          []*soakMember
	allClientPackets []*impairedPacketConn
}

// newNetworkSoakFixture builds a production-rate authority and a server path with loss, jitter, and reordering.
func newNetworkSoakFixture(t *testing.T, memberCount int, activeTicks int) *networkSoakFixture {
	t.Helper()

	identity, session, applied, endpoint := newNetworkSoakAuthority(t, memberCount)
	serverSocket, serverPackets, clientTLS := startNetworkSoakServer(t, endpoint)

	// The deadline grows with production-paced ticks and retains fixed headroom for connection and cleanup work.
	soakDeadline := time.Duration(activeTicks)*session.StepDuration() + 20*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), soakDeadline)

	return &networkSoakFixture{
		ctx:           ctx,
		cancel:        cancel,
		activeTicks:   activeTicks,
		identity:      identity,
		session:       session,
		applied:       applied,
		serverSocket:  serverSocket,
		serverPackets: serverPackets,
		clientTLS:     clientTLS,
	}
}

// newNetworkSoakAuthority builds distinct authenticated members around one production-rate session.
func newNetworkSoakAuthority(
	t *testing.T,
	memberCount int,
) (simulation.RuntimeIdentity, *gamesession.Session, map[string]int, *gameserver.Endpoint) {
	t.Helper()

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
	registerSoakMoveCommand(t, session, applied)

	endpoint, err := gameserver.NewEndpoint(
		&gameserver.Host{Engine: engine, Session: session, Allocation: allocation},
		soakAuthentication(memberCount, allocation.IdentityHash),
		soakProjection,
	)
	if err != nil {
		t.Fatal(err)
	}

	return identity, session, applied, endpoint
}

// startNetworkSoakServer applies server-side loss, jitter, and reordering before starting the accept loop.
func startNetworkSoakServer(
	t *testing.T,
	endpoint *gameserver.Endpoint,
) (net.PacketConn, *impairedPacketConn, *tls.Config) {
	t.Helper()

	serverSocket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	serverPackets := &impairedPacketConn{
		PacketConn: serverSocket,
		profile: impairmentProfile{
			dropEvery:    17,
			delays:       []time.Duration{0, time.Millisecond, 3 * time.Millisecond},
			reorderEvery: 13,
			reorderDelay: 8 * time.Millisecond,
		},
	}

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

	return serverSocket, serverPackets, clientTLS
}

// registerSoakMoveCommand counts accepted commands by server-bound player identity.
func registerSoakMoveCommand(
	t *testing.T,
	session *gamesession.Session,
	applied map[string]int,
) {
	t.Helper()

	if err := session.Register("move", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply: func(_ *gameecs.Engine, command simulation.Command) error {
			applied[command.Player]++

			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
}

// soakAuthentication creates unique principals so command counts prove every client remained independently live.
func soakAuthentication(memberCount int, identityHash string) soakAuthenticator {
	authentication := make(soakAuthenticator, memberCount)
	for index := range memberCount {
		authentication[fmt.Sprintf("ticket-%d", index)] = gameserver.Principal{
			ID:                  fmt.Sprintf("account-%d", index),
			CharacterID:         fmt.Sprintf("character-%d", index),
			PlayerID:            fmt.Sprintf("player-%d", index),
			RuntimeIdentityHash: identityHash,
		}
	}

	return authentication
}

// soakProjection exposes tick movement and complete empty collections required by the production client schema.
func soakProjection(player string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	view := playeradapter.ClientView{
		Version: playeradapter.ClientViewVersion,
		Tick:    checkpoint.Tick,
		HUD: playeradapter.HUD{
			Version: playeradapter.HUDVersion,
			Tick:    checkpoint.Tick,
			Player: playeradapter.HUDIdentity{
				PlayerID:    player,
				CharacterID: "character",
				Name:        player,
				Class:       "Amazon",
			},
			Vitals:   playeradapter.HUDVitals{Health: 1, MaxHealth: 1, Mana: 1, MaxMana: 1},
			Position: playeradapter.HUDPosition{X: float64(checkpoint.Tick)},
			Belt:     playeradapter.HUDBelt{Slots: []string{}},
		},
		World: playeradapter.WorldView{
			Version:  playeradapter.WorldViewVersion,
			Tick:     checkpoint.Tick,
			Entities: []playeradapter.WorldEntity{},
		},
		Private: playeradapter.PrivateView{
			Version: playeradapter.PrivateViewVersion,
			Tick:    checkpoint.Tick,
			Items:   playeradapter.ItemView{Items: []playeradapter.ItemEntityView{}},
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

// TestSustainedMultiClientSessionSurvivesImpairmentAndActiveReconnect exercises paced traffic and socket replacement.
func TestSustainedMultiClientSessionSurvivesImpairmentAndActiveReconnect(t *testing.T) {
	const memberCount = 3

	fixture := newNetworkSoakFixture(t, memberCount, networkSoakTicks(t))
	defer fixture.cancel()

	connectSoakMembers(t, fixture, memberCount)
	runNetworkSoak(t, fixture)
	advanceSoakTail(t, fixture.session)
	waitForSoakConvergence(t, fixture)
	assertSoakOutcome(t, fixture)
}

// connectSoakMembers gives each member a distinct socket profile, ticket, credential, and pair of watch streams.
func connectSoakMembers(t *testing.T, fixture *networkSoakFixture, memberCount int) {
	t.Helper()

	fixture.members = make([]*soakMember, memberCount)

	fixture.allClientPackets = make([]*impairedPacketConn, 0, memberCount+1)
	for index := range fixture.members {
		member, err := dialSoakMember(
			fixture.ctx,
			t,
			fixture.serverSocket.LocalAddr(),
			fixture.clientTLS,
			impairmentProfile{
				dropEvery:    19 + index*2,
				delays:       []time.Duration{time.Millisecond, 0, 2 * time.Millisecond},
				reorderEvery: 11 + index*2,
				reorderDelay: 7 * time.Millisecond,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		joined, err := member.transport.Join(fixture.ctx, gameserver.JoinRequest{
			Version:    gameserver.SessionProtocolVersion,
			Credential: fmt.Sprintf("ticket-%d", index),
			Identity:   fixture.identity,
		})
		if err != nil {
			t.Fatal(err)
		}

		member.credential = joined.Credential
		if err := member.startStreams(fixture.ctx); err != nil {
			t.Fatal(err)
		}

		fixture.members[index] = member
		fixture.allClientPackets = append(fixture.allClientPackets, member.packets)
	}

	t.Cleanup(func() {
		for _, member := range fixture.members {
			member.stop(fixture.ctx)
		}
	})
}

// runNetworkSoak submits one command per member per paced tick and exercises in-place and new-socket reconnects.
func runNetworkSoak(t *testing.T, fixture *networkSoakFixture) {
	t.Helper()

	for step := range fixture.activeTicks {
		stepStarted := time.Now()

		rotateSoakConnectionAtMilestones(t, fixture, step)

		current := fixture.session.Status().Tick
		for _, member := range fixture.members {
			member.sequence++
			if err := member.transport.Submit(fixture.ctx, member.credential, gameserver.CommandIntent{
				ObservedServerTick: current,
				TargetTick:         current + 2,
				Sequence:           member.sequence,
				Kind:               "move",
				Payload:            json.RawMessage(`{}`),
			}); err != nil {
				t.Fatalf("tick %d sequence %d: %v", current, member.sequence, err)
			}
		}

		if err := fixture.session.Step(); err != nil {
			t.Fatal(err)
		}

		if remaining := fixture.session.StepDuration() - time.Since(stepStarted); remaining > 0 {
			time.Sleep(remaining)
		}

		drainSoakMembers(t, fixture.members)
	}
}

// rotateSoakConnectionAtMilestones proves both credential rotation and active socket replacement preserve membership.
func rotateSoakConnectionAtMilestones(t *testing.T, fixture *networkSoakFixture, step int) {
	t.Helper()

	member := fixture.members[1]
	if step == fixture.activeTicks/2 {
		if err := reconnectSoakMember(fixture.ctx, member, fixture.identity); err != nil {
			t.Fatal(err)
		}
	}

	if step != 3*fixture.activeTicks/4 {
		return
	}

	oldPackets := member.packets
	if err := redialSoakMember(
		fixture.ctx,
		t,
		member,
		fixture.serverSocket.LocalAddr(),
		fixture.clientTLS,
		fixture.identity,
	); err != nil {
		t.Fatal(err)
	}

	fixture.allClientPackets = append(fixture.allClientPackets, member.packets)

	oldPackets.wait()
}

// drainSoakMembers records all currently buffered correction and transform progress after each paced tick.
func drainSoakMembers(t *testing.T, members []*soakMember) {
	t.Helper()

	for _, member := range members {
		if err := member.drainStreams(); err != nil {
			t.Fatal(err)
		}
	}
}

// advanceSoakTail lets commands targeted two ticks ahead become authoritative before convergence checks.
func advanceSoakTail(t *testing.T, session *gamesession.Session) {
	t.Helper()

	for range 2 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
}

// waitForSoakConvergence tolerates bounded delivery lag while continuing to drain one-slot latest-wins channels.
func waitForSoakConvergence(t *testing.T, fixture *networkSoakFixture) {
	t.Helper()

	finalTick := fixture.session.Status().Tick

	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for !soakCaughtUp(fixture.members, finalTick) {
		select {
		case <-deadline.C:
			t.Fatalf("network soak did not converge by tick %d: %#v", finalTick, fixture.members)
		case <-ticker.C:
			drainSoakMembers(t, fixture.members)
		}
	}
}

// assertSoakOutcome proves every command applied and every configured path carried timed transform traffic.
func assertSoakOutcome(t *testing.T, fixture *networkSoakFixture) {
	t.Helper()

	for index, member := range fixture.members {
		player := fmt.Sprintf("player-%d", index)
		if fixture.applied[player] != fixture.activeTicks {
			t.Fatalf(
				"%s applied commands = %d, want %d",
				player,
				fixture.applied[player],
				fixture.activeTicks,
			)
		}

		stats := member.transport.NetworkStats()
		if stats.SmoothedRTT <= 0 || stats.TransformsReceived == 0 {
			t.Fatalf("member %d network stats = %#v", index, stats)
		}
	}

	assertImpairmentApplied(
		t,
		"server soak",
		fixture.serverPackets.profile,
		fixture.serverPackets.stats(),
	)

	for index, packets := range fixture.allClientPackets {
		assertImpairmentApplied(
			t,
			fmt.Sprintf("client soak %d", index),
			packets.profile,
			packets.stats(),
		)
	}
}

// dialSoakMember opens one impaired UDP socket whose ownership transfers to the returned QUIC client.
func dialSoakMember(
	ctx context.Context,
	t *testing.T,
	address net.Addr,
	tlsConfig *tls.Config,
	profile impairmentProfile,
) (*soakMember, error) {
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

// startStreams reserves both watch modes and asserts their application queues remain single-slot bounded.
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

	if cap(snapshots) != 1 ||
		cap(snapshotErrors) != 1 ||
		cap(transforms) != 1 ||
		cap(transformErrors) != 1 {
		stop()

		return fmt.Errorf("network soak: unbounded stream channels")
	}

	member.stopStreams = stop
	member.snapshots, member.snapshotErrors = snapshots, snapshotErrors
	member.transforms, member.transformErrors = transforms, transformErrors

	return nil
}

// drainStreams consumes current progress without blocking the production-paced simulation loop.
func (member *soakMember) drainStreams() error {
	member.drainSnapshots()
	member.drainTransforms()

	if err := currentStreamError(member.snapshotErrors); err != nil {
		return err
	}

	return currentStreamError(member.transformErrors)
}

// drainSnapshots records the newest reliable correction and clears a channel that has closed.
func (member *soakMember) drainSnapshots() {
	for {
		select {
		case snapshot, open := <-member.snapshots:
			if !open {
				member.snapshots = nil

				return
			}

			member.latestCorrection = max(member.latestCorrection, snapshot.Tick)
		default:
			return
		}
	}
}

// drainTransforms records the newest disposable sample and clears a channel that has closed.
func (member *soakMember) drainTransforms() {
	for {
		select {
		case frame, open := <-member.transforms:
			if !open {
				member.transforms = nil

				return
			}

			member.latestTransform = max(member.latestTransform, frame.Tick)
		default:
			return
		}
	}
}

// currentStreamError samples one bounded error channel without stalling healthy traffic.
func currentStreamError(errorsIn <-chan error) error {
	select {
	case err, open := <-errorsIn:
		if open && err != nil {
			return err
		}
	default:
	}

	return nil
}

// stop cancels streams before leave and socket cleanup so no goroutine can retain packet ownership.
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
		_ = member.packets.Close()
	}
}

// reconnectSoakMember rotates credentials on the existing socket and starts fresh watch reservations.
func reconnectSoakMember(
	ctx context.Context,
	member *soakMember,
	identity simulation.RuntimeIdentity,
) error {
	if member.stopStreams != nil {
		member.stopStreams()
		member.stopStreams = nil
	}

	reconnected, err := member.transport.Reconnect(ctx, gameserver.ReconnectRequest{
		Credential: member.credential,
		Identity:   identity,
		Nonce:      "0123456789abcdef0123456789abcdef",
	})
	if err != nil {
		return err
	}

	member.credential = reconnected.Credential

	return member.startStreams(ctx)
}

// redialSoakMember replaces the physical socket before reconnecting the retained logical membership.
func redialSoakMember(
	ctx context.Context,
	t *testing.T,
	member *soakMember,
	address net.Addr,
	tlsConfig *tls.Config,
	identity simulation.RuntimeIdentity,
) error {
	t.Helper()

	if member.stopStreams != nil {
		member.stopStreams()
		member.stopStreams = nil
	}

	_ = member.transport.Close()
	_ = member.packets.Close()

	replacement, err := dialSoakMember(ctx, t, address, tlsConfig, member.packets.profile)
	if err != nil {
		return err
	}

	reconnected, err := replacement.transport.Reconnect(ctx, gameserver.ReconnectRequest{
		Credential: member.credential,
		Identity:   identity,
		Nonce:      "fedcba9876543210fedcba9876543210",
	})
	if err != nil {
		replacement.stop(ctx)

		return err
	}

	member.transport = replacement.transport
	member.packets = replacement.packets
	member.credential = reconnected.Credential

	return member.startStreams(ctx)
}

// soakCaughtUp allows four ticks of bounded delivery lag for both reliable and latest-wins paths.
func soakCaughtUp(members []*soakMember, tick uint64) bool {
	for _, member := range members {
		if member.latestCorrection+4 < tick || member.latestTransform+4 < tick {
			return false
		}
	}

	return true
}

// networkSoakTicks reads the opt-in duration while bounding accidental local or CI runtimes.
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
