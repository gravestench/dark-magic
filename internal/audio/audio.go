// Package audio defines the renderer-thread command boundary for audio.
package audio

import (
	"errors"
	"fmt"
	"sync"
)

// Command kinds form an internal native-backend protocol; their string values remain stable for adapters and tests.
const (
	commandPCMOpen  = "pcm-open"
	commandPCMWrite = "pcm-write"
	commandPlay     = "play"
	commandVolume   = "volume"
	commandPan      = "pan"
	commandStop     = "stop"
)

// SoundID is a generation-checked native sound handle.
type SoundID struct {
	Slot       uint32
	Generation uint32
}

// Command is one ordered backend operation.
type Command struct {
	Kind     string
	ID       SoundID
	Format   string
	Data     []byte
	Volume   float32
	Pan      float32
	Loop     bool
	Stream   bool
	Rate     int
	Channels int
}

// Backend consumes audio commands on the native audio owner thread.
type Backend interface {
	Apply(Command) error
}

// slot holds the generation and mode-specific state associated with one reusable handle slot.
type slot struct {
	generation uint32
	active     bool
	bus        string
	volume     float32
	group      string
	fade       *fade
	pcmRate    int
	pcmFrames  int64
}

// Mixer accepts concurrent sound requests and queues backend commands.
type Mixer struct {
	mu      sync.Mutex
	slots   []slot
	free    []uint32
	pending []Command
	buses   map[string]float32
}

// Diagnostics summarizes checked handles and queued owner-thread work.
type Diagnostics struct {
	Active, Pending, Slots int
	BusVolumes             map[string]float32
}

// Diagnostics returns a defensive snapshot so callers cannot mutate routing state while inspecting mixer health.
func (m *Mixer) Diagnostics() Diagnostics {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := Diagnostics{
		Pending:    len(m.pending),
		Slots:      len(m.slots),
		BusVolumes: make(map[string]float32, len(m.buses)),
	}
	for _, entry := range m.slots {
		if entry.active {
			result.Active++
		}
	}

	for bus, volume := range m.buses {
		result.BusVolumes[bus] = volume
	}

	return result
}

// Stop invalidates id before queuing native destruction so later calls cannot target the retiring sound.
func (m *Mixer) Stop(id SoundID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validateSoundLocked(id); err != nil {
		return err
	}

	m.retireSlotLocked(id.Slot)

	return nil
}

// Drain applies commands in queue order and retains the failed command plus its successors for a later retry.
func (m *Mixer) Drain(backend Backend) error {
	if backend == nil {
		return errors.New("audiocore: nil backend")
	}

	// Holding the lock across Apply prevents concurrent producers from changing the retry suffix or its ordering.
	m.mu.Lock()
	defer m.mu.Unlock()

	for index, command := range m.pending {
		if err := backend.Apply(command); err != nil {
			m.pending = append([]Command(nil), m.pending[index:]...)

			return fmt.Errorf("audiocore: apply %s for sound %v: %w", command.Kind, command.ID, err)
		}
	}

	m.pending = nil

	return nil
}

// reserveSlotLocked takes the most recently freed slot or appends one, preserving deterministic LIFO handle reuse.
func (m *Mixer) reserveSlotLocked() uint32 {
	if len(m.free) == 0 {
		index := uint32(len(m.slots))
		m.slots = append(m.slots, slot{generation: 1})

		return index
	}

	lastFree := len(m.free) - 1
	index := m.free[lastFree]
	m.free = m.free[:lastFree]

	// Reuse does not clear the slot; each playback mode initializes exactly the state it owns.
	return index
}

// retireSlotLocked advances the generation and queues stop after freeing the slot.
// This ordering ensures callers observe handle invalidation before native destruction.
func (m *Mixer) retireSlotLocked(index uint32) {
	entry := &m.slots[index]
	id := SoundID{Slot: index, Generation: entry.generation}

	entry.active = false

	entry.generation++
	if entry.generation == 0 {
		// Generation zero is never issued, so wraparound cannot recreate a zero-value handle.
		entry.generation = 1
	}

	m.free = append(m.free, index)
	m.pending = append(m.pending, Command{Kind: commandStop, ID: id})
}

// validateSoundLocked rejects out-of-range and stale generations so reused slots cannot be controlled by old handles.
func (m *Mixer) validateSoundLocked(id SoundID) error {
	if int(id.Slot) >= len(m.slots) {
		return fmt.Errorf("audiocore: invalid sound slot %d", id.Slot)
	}

	entry := m.slots[id.Slot]
	if !entry.active || entry.generation != id.Generation {
		return fmt.Errorf("audiocore: stale sound %v", id)
	}

	return nil
}
