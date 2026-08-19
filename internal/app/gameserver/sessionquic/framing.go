package sessionquic

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/quic-go/quic-go"
)

// quicConfig bounds idle lifetime, stream concurrency, receive memory, and initial packet size
// identically at both ends.
func quicConfig() *quic.Config {
	return &quic.Config{
		HandshakeIdleTimeout:       5 * time.Second,
		MaxIdleTimeout:             30 * time.Second,
		InitialPacketSize:          InitialPacketSize,
		MaxIncomingStreams:         16,
		MaxIncomingUniStreams:      -1,
		MaxStreamReceiveWindow:     MaxFrameBytes,
		MaxConnectionReceiveWindow: 2 * MaxFrameBytes,
		EnableDatagrams:            true,
	}
}

// serverTLS clones caller state before pinning the session ALPN and TLS floor.
func serverTLS(config *tls.Config) *tls.Config {
	return sessionTLS(config)
}

// clientTLS applies the same ALPN and TLS floor so negotiation cannot silently select another application protocol.
func clientTLS(config *tls.Config) *tls.Config {
	return sessionTLS(config)
}

// sessionTLS prevents transport policy from mutating TLS configuration shared with another listener or client.
func sessionTLS(config *tls.Config) *tls.Config {
	clone := config.Clone()
	clone.NextProtos = []string{ALPN}
	clone.MinVersion = tls.VersionTLS13

	return clone
}

// writeFrame serializes one JSON value behind a fixed-width length prefix after enforcing the allocation ceiling.
func writeFrame(writer io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if len(data) == 0 || len(data) > MaxFrameBytes {
		return ErrWire
	}

	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(data)))

	if err := writeAll(writer, size[:]); err != nil {
		return err
	}

	return writeAll(writer, data)
}

// writeAll treats zero-progress and impossible writer counts as short writes instead of spinning forever.
func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := writer.Write(data)
		if err != nil {
			return err
		}

		if written <= 0 || written > len(data) {
			return io.ErrShortWrite
		}

		data = data[written:]
	}

	return nil
}

// readFrame bounds the declared body, rejects unknown fields, and requires exactly one complete JSON value.
func readFrame(reader io.Reader, value any) error {
	buffered := bufio.NewReaderSize(reader, 4<<10)

	var size [4]byte
	if _, err := io.ReadFull(buffered, size[:]); err != nil {
		return err
	}

	length := binary.BigEndian.Uint32(size[:])
	if length == 0 || length > MaxFrameBytes {
		return ErrWire
	}

	limited := &io.LimitedReader{R: buffered, N: int64(length)}
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("%w: %v", ErrWire, err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) || limited.N != 0 {
		return ErrWire
	}

	return nil
}
