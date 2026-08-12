package mapgen

import "github.com/gravestench/dark-magic/internal/game/simulation"

// Random is the small deterministic capability a generator needs. It prevents
// generator code from reaching for wall-clock-seeded global randomness.
type Random interface {
	Uint64() uint64
	Uint64n(uint64) uint64
}

// Streams derives independent sequences by purpose. Adding a decorative roll
// must not alter room topology, preset choice, or spawn placement.
type Streams struct{ seed uint64 }

func NewStreams(seed uint64) Streams { return Streams{seed: seed} }

func (streams Streams) For(purpose string) Random {
	return simulation.NewStream(streams.seed, "mapgen/"+purpose)
}
