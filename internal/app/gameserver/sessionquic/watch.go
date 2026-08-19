package sessionquic

import (
	"context"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/quic-go/quic-go"
)

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

	// Transform lifetime follows the reliable stream, which owns the membership's single watch reservation.
	go server.sendTransforms(stream.Context(), connection, credential)

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
) {
	ticker := time.NewTicker(TransformInterval)
	defer ticker.Stop()

	lastTick := uint64(0)

	for {
		snapshot, err := server.endpoint.Observe(credential)
		if err != nil {
			return
		}

		if snapshot.Tick > lastTick {
			payload, encodeErr := encodeTransformFrame(credential, snapshot)
			if encodeErr != nil || connection.SendDatagram(payload) != nil {
				return
			}

			lastTick = snapshot.Tick
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
