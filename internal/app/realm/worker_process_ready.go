package realm

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
)

const (
	WorkerProcessReadyVersion = "RealmWorkerProcessReady/v2"
	maximumWorkerReadyBytes   = 16 << 10
)

// WorkerProcessReady is the owner-only rendezvous written after a child worker
// has opened both private control and public game transports. It contains no
// bearer token, admission key, or private key.
type WorkerProcessReady struct {
	Version        string       `json:"version"`
	GameID         string       `json:"game_id"`
	AllocationID   string       `json:"allocation_id"`
	ProcessID      int          `json:"process_id"`
	ControlAddress string       `json:"control_address"`
	GameEndpoint   GameEndpoint `json:"game_endpoint"`
}

// WriteWorkerProcessReady emits the canonical worker process ready representation so persisted and transported values
// retain one stable shape.
func WriteWorkerProcessReady(path string, ready WorkerProcessReady) error {
	ready.Version = WorkerProcessReadyVersion
	if err := ready.Validate(); err != nil {
		return err
	}

	return writeRealmJSON(path, ready)
}

// ReadWorkerProcessReady decodes the worker process ready representation at one boundary so malformed data fails
// before it becomes shared state.
func ReadWorkerProcessReady(path string) (WorkerProcessReady, error) {
	file, err := os.Open(path)
	if err != nil {
		return WorkerProcessReady{}, err
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maximumWorkerReadyBytes+1))
	if err != nil || len(data) > maximumWorkerReadyBytes {
		return WorkerProcessReady{}, ErrWorkerProtocol
	}

	var ready WorkerProcessReady

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&ready); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return WorkerProcessReady{}, ErrWorkerProtocol
	}

	if err := ready.Validate(); err != nil {
		return WorkerProcessReady{}, err
	}

	return ready, nil
}

// Validate checks the worker process ready invariant before state changes, keeping invalid values off shared paths.
func (ready WorkerProcessReady) Validate() error {
	if ready.Version != WorkerProcessReadyVersion || strings.TrimSpace(ready.GameID) == "" || len(ready.GameID) > 255 ||
		strings.TrimSpace(ready.AllocationID) == "" || len(ready.AllocationID) > 255 || ready.ProcessID <= 0 ||
		!validLoopbackAddress(ready.ControlAddress) ||
		strings.TrimSpace(ready.GameEndpoint.Address) == "" || !validTLSFingerprint(ready.GameEndpoint.TLSFingerprint) {
		return ErrWorkerProtocol
	}

	return nil
}

// validLoopbackAddress checks the worker process ready invariant before state changes, keeping invalid values off
// shared paths.
func validLoopbackAddress(address string) bool {
	host, port, err := net.SplitHostPort(strings.TrimSpace(address))
	if err != nil || port == "" {
		return false
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}

// validTLSFingerprint checks the worker process ready invariant before state changes, keeping invalid values off
// shared paths.
func validTLSFingerprint(value string) bool {
	const prefix = "sha256:"

	value = strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}

	for _, character := range value[len(prefix):] {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}

	return true
}
