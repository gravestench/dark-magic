// Package simulation defines replay-stable primitives shared by offline and
// authoritative game simulation.
package simulation

import (
	"crypto/sha256"
	"encoding/binary"
)

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
