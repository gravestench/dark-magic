package sessionquic

import (
	"context"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/quic-go/quic-go"
)

// TestTransformWatchStatsObserveAndSend catches missing transform-loop progress or retained credential entries.
func TestTransformWatchStatsObserveAndSend(t *testing.T) {
	fixture := newTransformWatchStatsFixture(t)
	joined := joinTransformStatsClient(t, fixture)
	if _, active := fixture.server.TransformWatchStats(joined.Credential); active {
		t.Fatal("transform stats exist before a watch")
	}

	watchContext, cancelWatch := context.WithCancel(fixture.ctx)
	defer cancelWatch()

	snapshots, _, err := fixture.client.Watch(watchContext, joined.Credential)
	if err != nil {
		t.Fatal(err)
	}

	transforms, _, err := fixture.client.WatchTransforms(watchContext, joined.Credential)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-snapshots:
	case <-fixture.ctx.Done():
		t.Fatal(fixture.ctx.Err())
	}

	waitForTransformStats(t, fixture.ctx, func() bool {
		stats, active := fixture.server.TransformWatchStats(joined.Credential)
		return active && stats.ObserveAttempts > 0
	})

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}

	select {
	case frame := <-transforms:
		if frame.Tick != 1 {
			t.Fatalf("transform tick = %d, want 1", frame.Tick)
		}
	case <-fixture.ctx.Done():
		t.Fatal(fixture.ctx.Err())
	}

	waitForTransformStats(t, fixture.ctx, func() bool {
		stats, active := fixture.server.TransformWatchStats(joined.Credential)
		return active && stats.DatagramsSent == 1
	})
	stats, active := fixture.server.TransformWatchStats(joined.Credential)
	if !active || stats.ObserveAttempts == 0 || stats.ObserveErrors != 0 ||
		stats.LatestObservedTick != 1 || stats.EncodeAttempts != 1 || stats.EncodeErrors != 0 ||
		stats.SendAttempts != 1 || stats.SendErrors != 0 || stats.DatagramsSent != 1 || stats.LatestSentTick != 1 {
		t.Fatalf("successful transform stats = %+v, active = %t", stats, active)
	}

	cancelWatch()
	for range snapshots {
	}

	waitForTransformStats(t, fixture.ctx, func() bool {
		_, active := fixture.server.TransformWatchStats(joined.Credential)
		return !active
	})
}

// TestTransformWatchStatsEncodeFailure distinguishes a valid host observation from an unsendable projection.
func TestTransformWatchStatsEncodeFailure(t *testing.T) {
	fixture := newQUICProtocolFixture(t)
	joined := joinTransformStatsClient(t, fixture)
	watchContext, cancelWatch := context.WithCancel(fixture.ctx)
	defer cancelWatch()

	snapshots, _, err := fixture.client.Watch(watchContext, joined.Credential)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-snapshots:
	case <-fixture.ctx.Done():
		t.Fatal(fixture.ctx.Err())
	}

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}

	waitForTransformStats(t, fixture.ctx, func() bool {
		stats, active := fixture.server.TransformWatchStats(joined.Credential)
		return active && stats.EncodeErrors == 1
	})
	stats, _ := fixture.server.TransformWatchStats(joined.Credential)
	if stats.ObserveAttempts == 0 || stats.ObserveErrors != 0 || stats.LatestObservedTick != 1 ||
		stats.EncodeAttempts != 1 || stats.EncodeErrors != 1 || stats.SendAttempts != 0 ||
		stats.SendErrors != 0 || stats.DatagramsSent != 0 || stats.LatestSentTick != 0 {
		t.Fatalf("failed encode stats = %+v", stats)
	}
}

// TestTransformReceiveStats catches raw/decode conflation and stale active state after receiver completion.
func TestTransformReceiveStats(t *testing.T) {
	ctx, client, peer := newTransformReceiveStatsPair(t)
	credential := gameserver.SessionCredential("transform-stats")
	if client.NetworkStats().TransformWatchActive {
		t.Fatal("transform receiver active before watch")
	}

	watchContext, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()

	frames, receiveErrors, err := client.WatchTransforms(watchContext, credential)
	if err != nil {
		t.Fatal(err)
	}

	if !client.NetworkStats().TransformWatchActive {
		t.Fatal("transform receiver inactive after watch starts")
	}

	if err := peer.SendDatagram([]byte("malformed")); err != nil {
		t.Fatal(err)
	}

	waitForTransformStats(t, ctx, func() bool { return client.NetworkStats().TransformDecodeErrors == 1 })
	stats := client.NetworkStats()
	if stats.DatagramsRead != 1 || stats.TransformDecodeErrors != 1 || stats.TransformsReceived != 0 {
		t.Fatalf("malformed datagram stats = %+v", stats)
	}

	payload, err := encodeTransformFrame(credential, transformSnapshot(t, emptyTransformView(7)))
	if err != nil {
		t.Fatal(err)
	}

	if err := peer.SendDatagram(payload); err != nil {
		t.Fatal(err)
	}

	select {
	case frame := <-frames:
		if frame.Tick != 7 {
			t.Fatalf("decoded transform tick = %d, want 7", frame.Tick)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	stats = client.NetworkStats()
	if stats.DatagramsRead != 2 || stats.TransformDecodeErrors != 1 || stats.TransformsReceived != 1 ||
		!stats.TransformWatchActive {
		t.Fatalf("decoded datagram stats = %+v", stats)
	}

	cancelWatch()
	for range frames {
	}

	if client.NetworkStats().TransformWatchActive {
		t.Fatal("transform receiver active after its output channel closed")
	}

	if err := <-receiveErrors; err != nil {
		t.Fatalf("canceled receiver error = %v", err)
	}

	// A completed receiver releases the single-reader reservation for the next watch.
	nextContext, cancelNext := context.WithCancel(ctx)
	defer cancelNext()

	nextFrames, _, err := client.WatchTransforms(nextContext, credential)
	if err != nil {
		t.Fatal(err)
	}

	cancelNext()
	for range nextFrames {
	}

	if client.NetworkStats().TransformWatchActive {
		t.Fatal("replacement transform receiver stayed active after cancellation")
	}
}

// newTransformWatchStatsFixture reuses the real authority projection with an ordinary, unimpaired QUIC path.
func newTransformWatchStatsFixture(t *testing.T) quicProtocolFixture {
	t.Helper()

	identity := testRuntimeIdentity()
	session, endpoint := newImpairedRecoveryEndpoint(t, identity)
	serverTLS, clientTLS := testTLS(t)
	server, err := Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = server.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	go func() { _ = server.Serve(ctx) }()

	client, err := Dial(ctx, server.Addr(), clientTLS)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = client.Close() })

	return quicProtocolFixture{ctx: ctx, identity: identity, session: session, client: client, server: server}
}

// joinTransformStatsClient authenticates through the production handshake before watch diagnostics are inspected.
func joinTransformStatsClient(t *testing.T, fixture quicProtocolFixture) gameserver.JoinResponse {
	t.Helper()

	joined, err := fixture.client.Join(fixture.ctx, gameserver.JoinRequest{
		Version: gameserver.SessionProtocolVersion, Credential: "realm-ticket", Identity: fixture.identity,
	})
	if err != nil {
		t.Fatal(err)
	}

	return joined
}

// newTransformReceiveStatsPair lets the real QUIC peer send malformed samples without changing production encoders.
func newTransformReceiveStatsPair(t *testing.T) (context.Context, *Client, *quic.Conn) {
	t.Helper()

	serverTLS, clientTLS := testTLS(t)
	listener, err := quic.ListenAddr("127.0.0.1:0", sessionTLS(serverTLS), quicConfig())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = listener.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	client, err := Dial(ctx, listener.Addr().String(), clientTLS)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = client.Close() })

	peer, err := listener.Accept(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = peer.CloseWithError(0, "test complete") })

	return ctx, client, peer
}

// waitForTransformStats waits on observable progress instead of sleeping to arrange goroutine order.
func waitForTransformStats(t *testing.T, ctx context.Context, ready func() bool) {
	t.Helper()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for !ready() {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}
