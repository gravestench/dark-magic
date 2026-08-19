package sessionquic

import (
	"net"
	"sync"
	"testing"
	"time"
)

// impairedPacketConn applies deterministic loss, delay, and reorder decisions to outgoing test packets.
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

// impairmentProfile defines repeatable packet faults so failures can be reproduced by write sequence.
type impairmentProfile struct {
	dropEvery    int
	delays       []time.Duration
	reorderEvery int
	reorderDelay time.Duration
}

// impairmentStats captures observed faults for proof that a passing test actually exercised its profile.
type impairmentStats struct {
	writes, dropped, delayed, reordered int
	injectedDelay                       time.Duration
}

// impairmentDecision records the immutable action selected for one packet while holding the connection lock.
type impairmentDecision struct {
	drop    bool
	reorder bool
	delay   time.Duration
}

// WriteTo applies one deterministic impairment while reporting accepted drops as successful UDP writes.
func (connection *impairedPacketConn) WriteTo(payload []byte, address net.Addr) (int, error) {
	decision, open := connection.recordWrite()
	if !open {
		return 0, net.ErrClosed
	}

	if decision.reorder {
		connection.queueReorderedWrite(payload, address, decision.delay)

		return len(payload), nil
	}

	if decision.delay > 0 {
		time.Sleep(decision.delay)
	}

	if decision.drop {
		return len(payload), nil
	}

	return connection.PacketConn.WriteTo(payload, address)
}

// recordWrite chooses and accounts for a fault atomically so wait cannot race with new reorder work.
func (connection *impairedPacketConn) recordWrite() (impairmentDecision, bool) {
	connection.mu.Lock()
	defer connection.mu.Unlock()

	if connection.closing {
		return impairmentDecision{}, false
	}

	connection.writes++
	sequence := connection.writes

	decision := impairmentDecision{
		drop: connection.profile.dropEvery > 0 && sequence%connection.profile.dropEvery == 0,
	}
	if decision.drop {
		connection.dropped++
	}

	if len(connection.profile.delays) > 0 {
		decision.delay = connection.profile.delays[(sequence-1)%len(connection.profile.delays)]
	}

	decision.reorder = !decision.drop &&
		connection.profile.reorderEvery > 0 &&
		sequence%connection.profile.reorderEvery == 0
	if decision.reorder {
		connection.reordered++
		decision.delay += connection.profile.reorderDelay
		// Add under the same lock used by wait; once closing is set no writer can race with Wait.
		connection.pending.Add(1)
	}

	if decision.delay > 0 {
		connection.delayed++
		connection.injectedDelay += decision.delay
	}

	return decision, true
}

// queueReorderedWrite owns a payload copy because its caller may reuse the QUIC packet buffer immediately.
func (connection *impairedPacketConn) queueReorderedWrite(
	payload []byte,
	address net.Addr,
	delay time.Duration,
) {
	copy := append([]byte(nil), payload...)

	go func() {
		defer connection.pending.Done()

		time.Sleep(delay)

		_, _ = connection.PacketConn.WriteTo(copy, address)
	}()
}

// stats snapshots counters under lock so assertions cannot observe a partially accounted packet.
func (connection *impairedPacketConn) stats() impairmentStats {
	connection.mu.Lock()
	defer connection.mu.Unlock()

	return impairmentStats{
		writes:        connection.writes,
		dropped:       connection.dropped,
		delayed:       connection.delayed,
		reordered:     connection.reordered,
		injectedDelay: connection.injectedDelay,
	}
}

// wait rejects new writes before waiting, satisfying WaitGroup's requirement that Add cannot race with Wait.
func (connection *impairedPacketConn) wait() {
	connection.mu.Lock()
	connection.closing = true
	connection.mu.Unlock()

	connection.pending.Wait()
}

// assertImpairmentApplied reconstructs expected counters so a green network test proves faults were injected.
func assertImpairmentApplied(
	t *testing.T,
	name string,
	profile impairmentProfile,
	stats impairmentStats,
) {
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

	if stats.writes == 0 ||
		stats.dropped != stats.writes/profile.dropEvery ||
		stats.delayed != expectedDelayed ||
		stats.injectedDelay != expectedDelay ||
		stats.reordered != expectedReordered {
		t.Fatalf("%s synthetic impairment mismatch: profile=%#v stats=%#v", name, profile, stats)
	}
}
