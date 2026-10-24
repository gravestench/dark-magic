package sessionquic

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/quic-go/quic-go"
)

// TransformWatchStats is a read-only snapshot retained only while a credential has an active server watch.
type TransformWatchStats struct {
	ObserveAttempts    uint64
	ObserveErrors      uint64
	LatestObservedTick uint64
	EncodeAttempts     uint64
	EncodeErrors       uint64
	SendAttempts       uint64
	SendErrors         uint64
	DatagramsSent      uint64
	LatestSentTick     uint64
}

// transformWatchCounters lets the transform loop publish diagnostics without blocking on readers.
type transformWatchCounters struct {
	observeAttempts    atomic.Uint64
	observeErrors      atomic.Uint64
	latestObservedTick atomic.Uint64
	encodeAttempts     atomic.Uint64
	encodeErrors       atomic.Uint64
	sendAttempts       atomic.Uint64
	sendErrors         atomic.Uint64
	datagramsSent      atomic.Uint64
	latestSentTick     atomic.Uint64
}

// TransformWatchStats copies active-watch diagnostics without retaining credentials after watch completion.
func (server *Server) TransformWatchStats(credential gameserver.SessionCredential) (TransformWatchStats, bool) {
	server.transformWatchMu.Lock()
	defer server.transformWatchMu.Unlock()

	counters, active := server.transformWatches[credential]
	if !active {
		return TransformWatchStats{}, false
	}

	return TransformWatchStats{
		ObserveAttempts:    counters.observeAttempts.Load(),
		ObserveErrors:      counters.observeErrors.Load(),
		LatestObservedTick: counters.latestObservedTick.Load(),
		EncodeAttempts:     counters.encodeAttempts.Load(),
		EncodeErrors:       counters.encodeErrors.Load(),
		SendAttempts:       counters.sendAttempts.Load(),
		SendErrors:         counters.sendErrors.Load(),
		DatagramsSent:      counters.datagramsSent.Load(),
		LatestSentTick:     counters.latestSentTick.Load(),
	}, true
}

// watch streams reliable corrections while a sibling datagram loop publishes disposable transform samples.
func (server *Server) watch(
	ctx context.Context,
	connection *quic.Conn,
	stream *quic.Stream,
	credential gameserver.SessionCredential,
) {
	if err := server.endpoint.BeginWatch(credential); err != nil {
		_ = writeFrame(stream, response{Error: err.Error()})

		return
	}
	defer server.endpoint.EndWatch(credential)

	counters := &transformWatchCounters{}
	server.transformWatchMu.Lock()
	if server.transformWatches == nil {
		server.transformWatches = make(map[gameserver.SessionCredential]*transformWatchCounters)
	}
	server.transformWatches[credential] = counters
	server.transformWatchMu.Unlock()
	defer func() {
		server.transformWatchMu.Lock()
		defer server.transformWatchMu.Unlock()

		delete(server.transformWatches, credential)
	}()

	// Transform lifetime follows the reliable stream, which owns the membership's single watch reservation.
	go server.sendTransforms(stream.Context(), connection, credential, counters)

	ticker := time.NewTicker(CorrectionInterval)
	defer ticker.Stop()

	lastTick, lastChecksum := ^uint64(0), ""

	for {
		snapshot, err := server.endpoint.Observe(credential)
		if err != nil {
			_ = writeFrame(stream, response{Error: err.Error()})

			return
		}

		// Repeated observations at one canonical tick need no duplicate reliable frame.
		if snapshot.Tick != lastTick || snapshot.Checksum != lastChecksum {
			if err := writeFrame(stream, response{Snapshot: &snapshot}); err != nil {
				return
			}

			lastTick, lastChecksum = snapshot.Tick, snapshot.Checksum
		}

		select {
		case <-ctx.Done():
			return
		case <-stream.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

// sendTransforms sends at most one datagram per newer simulation tick and stops instead of retrying stale motion.
func (server *Server) sendTransforms(
	ctx context.Context,
	connection *quic.Conn,
	credential gameserver.SessionCredential,
	counters *transformWatchCounters,
) {
	ticker := time.NewTicker(TransformInterval)
	defer ticker.Stop()

	lastTick := uint64(0)

	for {
		counters.observeAttempts.Add(1)
		snapshot, err := server.endpoint.Observe(credential)
		if err != nil {
			counters.observeErrors.Add(1)

			return
		}

		counters.latestObservedTick.Store(snapshot.Tick)
		if snapshot.Tick > lastTick {
			counters.encodeAttempts.Add(1)
			payload, encodeErr := encodeTransformFrame(credential, snapshot)
			if encodeErr != nil {
				counters.encodeErrors.Add(1)

				return
			}

			counters.sendAttempts.Add(1)
			if connection.SendDatagram(payload) != nil {
				counters.sendErrors.Add(1)

				return
			}

			counters.latestSentTick.Store(snapshot.Tick)
			counters.datagramsSent.Add(1)
			lastTick = snapshot.Tick
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
