package sessionquic

import (
	"context"
	"errors"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/quic-go/quic-go"
)

// Watch opens one reliable correction stream whose bounded output channel applies backpressure to QUIC.
func (client *Client) Watch(
	ctx context.Context,
	credential gameserver.SessionCredential,
) (<-chan gameserver.Snapshot, <-chan error, error) {
	stream, err := client.connection.OpenStreamSync(ctx)
	if err != nil {
		return nil, nil, err
	}

	if err := writeFrame(stream, request{Operation: operationWatch, Credential: credential}); err != nil {
		stream.CancelRead(1)
		stream.CancelWrite(1)

		return nil, nil, err
	}

	snapshots := make(chan gameserver.Snapshot, 1)

	errorsOut := make(chan error, 1)
	go client.receiveCorrections(ctx, stream, snapshots, errorsOut)

	return snapshots, errorsOut, nil
}

// receiveCorrections owns stream cleanup and reports transport errors only when cancellation was not requested.
func (client *Client) receiveCorrections(
	ctx context.Context,
	stream *quic.Stream,
	snapshots chan<- gameserver.Snapshot,
	errorsOut chan<- error,
) {
	defer close(snapshots)
	defer close(errorsOut)
	defer func() { _ = stream.Close() }()

	done := make(chan struct{})

	defer close(done)
	go cancelStreamWithContext(ctx, stream, done)

	for {
		var result response
		if err := readFrame(stream, &result); err != nil {
			if ctx.Err() == nil {
				errorsOut <- err
			}

			return
		}

		if result.Error != "" {
			errorsOut <- remoteError(result.Error)

			return
		}

		if result.Snapshot == nil {
			errorsOut <- ErrWire

			return
		}

		select {
		case snapshots <- *result.Snapshot:
		case <-ctx.Done():
			return
		}
	}
}

// cancelStreamWithContext interrupts blocked reads and writes while allowing normal completion to remain graceful.
func cancelStreamWithContext(ctx context.Context, stream *quic.Stream, done <-chan struct{}) {
	select {
	case <-ctx.Done():
		stream.CancelRead(0)
		stream.CancelWrite(0)
	case <-done:
	}
}

// WatchTransforms reserves the connection's single datagram receiver and publishes disposable latest-wins frames.
func (client *Client) WatchTransforms(
	ctx context.Context,
	credential gameserver.SessionCredential,
) (<-chan TransformFrame, <-chan error, error) {
	if err := client.beginTransformWatch(); err != nil {
		return nil, nil, err
	}

	frames := make(chan TransformFrame, 1)

	errorsOut := make(chan error, 1)
	go client.receiveTransforms(ctx, credential, frames, errorsOut)

	return frames, errorsOut, nil
}

// beginTransformWatch prevents two consumers from racing on QUIC's connection-wide datagram queue.
func (client *Client) beginTransformWatch() error {
	client.datagramMu.Lock()
	defer client.datagramMu.Unlock()

	if client.datagramActive {
		return errors.New("game session QUIC: transform watch already active")
	}

	client.datagramActive = true

	return nil
}

// endTransformWatch releases datagram ownership only after the receiver goroutine has stopped reading.
func (client *Client) endTransformWatch() {
	client.datagramMu.Lock()
	defer client.datagramMu.Unlock()

	client.datagramActive = false
}

// receiveTransforms ignores malformed or foreign datagrams because unreliable samples are not session-fatal.
func (client *Client) receiveTransforms(
	ctx context.Context,
	credential gameserver.SessionCredential,
	frames chan TransformFrame,
	errorsOut chan error,
) {
	defer close(frames)
	defer close(errorsOut)
	defer client.endTransformWatch()

	for {
		payload, err := client.connection.ReceiveDatagram(ctx)
		if err != nil {
			if ctx.Err() == nil {
				errorsOut <- err
			}

			return
		}

		frame, err := decodeTransformFrame(credential, payload)
		if err != nil {
			continue
		}

		client.transformsReceived.Add(1)

		if !client.publishLatestTransform(ctx, frames, frame) {
			return
		}
	}
}

// publishLatestTransform replaces an unread sample instead of allowing stale motion to queue behind rendering.
func (client *Client) publishLatestTransform(
	ctx context.Context,
	frames chan TransformFrame,
	frame TransformFrame,
) bool {
	select {
	case frames <- frame:
		return true
	default:
		client.transformsSuperseded.Add(1)
	}

	// The channel is single-slot, but a concurrent consumer may drain it before this non-blocking receive.
	select {
	case <-frames:
	default:
	}

	select {
	case frames <- frame:
		return true
	case <-ctx.Done():
		return false
	}
}
