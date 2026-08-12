// Package simulation defines replay-stable primitives shared by offline and
// authoritative game simulation.
package simulation

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var ErrRandomStream = errors.New("simulation: invalid deterministic random stream")

// Stream is one deterministic SplitMix64 random sequence. Streams should be
// derived by stable purpose names so adding a roll in one subsystem cannot
// perturb another subsystem's results.
type Stream struct {
	state uint64
}

// NewStream derives an independent sequence from a world seed and stable name.
func NewStream(seed uint64, name string) *Stream {
	hash := sha256.New()
	var encoded [8]byte
	binary.LittleEndian.PutUint64(encoded[:], seed)
	_, _ = hash.Write(encoded[:])
	_, _ = hash.Write([]byte(name))
	sum := hash.Sum(nil)
	return &Stream{state: binary.LittleEndian.Uint64(sum[:8])}
}

// RestoreStream resumes a stream from a previously captured state.
func RestoreStream(state uint64) *Stream { return &Stream{state: state} }

// State returns the complete serializable stream state.
func (stream *Stream) State() uint64 { return stream.state }

// Uint64 advances and returns the next value.
func (stream *Stream) Uint64() uint64 {
	stream.state += 0x9e3779b97f4a7c15
	value := stream.state
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

// Uint64n returns a value in [0, limit) without modulo bias.
func (stream *Stream) Uint64n(limit uint64) uint64 {
	if limit == 0 {
		return 0
	}
	threshold := -limit % limit
	for {
		value := stream.Uint64()
		if value >= threshold {
			return value % limit
		}
	}
}

// RandomStreams owns all named random sequences used by one authoritative
// session. A gameplay domain gets its own purpose-named stream so adding a loot
// roll cannot change combat results. The registry, rather than the script VM,
// owns the sequence state so checkpoints and replay can restore it exactly.
type RandomStreams struct {
	mu      sync.Mutex
	seed    uint64
	streams map[string]*Stream
}

type randomStreamsArchive struct {
	Seed    uint64            `json:"seed"`
	Streams map[string]uint64 `json:"streams"`
}

func NewRandomStreams(seed uint64) *RandomStreams {
	return &RandomStreams{seed: seed, streams: make(map[string]*Stream)}
}

func (*RandomStreams) StateID() string { return "engine.authoritative_rng/v1" }

// Register declares a stable purpose name before simulation begins. Requiring
// registration catches misspellings instead of quietly creating a second
// random sequence halfway through a session.
func (registry *RandomStreams) Register(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: purpose name is required", ErrRandomStream)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.streams[name]; exists {
		return fmt.Errorf("%w: purpose %q is already registered", ErrRandomStream, name)
	}
	registry.streams[name] = NewStream(registry.seed, name)
	return nil
}

// Uint64n advances one registered stream and returns a value below limit.
func (registry *RandomStreams) Uint64n(name string, limit uint64) (uint64, error) {
	if limit == 0 {
		return 0, fmt.Errorf("%w: limit must be greater than zero", ErrRandomStream)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	stream, found := registry.streams[name]
	if !found {
		return 0, fmt.Errorf("%w: unknown purpose %q", ErrRandomStream, name)
	}
	return stream.Uint64n(limit), nil
}

func (registry *RandomStreams) SnapshotState() ([]byte, error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	states := make(map[string]uint64, len(registry.streams))
	for name, stream := range registry.streams {
		states[name] = stream.State()
	}
	return json.Marshal(randomStreamsArchive{Seed: registry.seed, Streams: states})
}

func (registry *RandomStreams) RestoreState(data []byte) error {
	var archive randomStreamsArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		return fmt.Errorf("%w: decode: %v", ErrRandomStream, err)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if archive.Seed != registry.seed || len(archive.Streams) != len(registry.streams) {
		return fmt.Errorf("%w: seed or registration differs", ErrRandomStream)
	}
	names := make([]string, 0, len(registry.streams))
	for name := range registry.streams {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		state, found := archive.Streams[name]
		if !found {
			return fmt.Errorf("%w: purpose %q is not registered by checkpoint", ErrRandomStream, name)
		}
		registry.streams[name] = RestoreStream(state)
	}
	return nil
}
