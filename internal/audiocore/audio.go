// Package audiocore defines the renderer-thread command boundary for audio.
package audiocore

import (
	"errors"
	"fmt"
	"sync"
)

// SoundID is a generation-checked native sound handle.
type SoundID struct {
	Slot       uint32
	Generation uint32
}

// Command is one ordered backend operation.
type Command struct {
	Kind   string
	ID     SoundID
	Format string
	Data   []byte
	Volume float32
}

// Backend consumes audio commands on the native audio owner thread.
type Backend interface{ Apply(Command) error }

type slot struct {
	generation uint32
	active     bool
}

// Mixer accepts concurrent sound requests and queues backend commands.
type Mixer struct {
	mu      sync.Mutex
	slots   []slot
	free    []uint32
	pending []Command
}

// Play queues a decoded-by-backend sound asset.
func (m *Mixer) Play(format string, data []byte) (SoundID, error) {
	if format == "" || len(data) == 0 {
		return SoundID{}, errors.New("audiocore: format and sound data are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var index uint32
	if len(m.free) != 0 {
		index = m.free[len(m.free)-1]
		m.free = m.free[:len(m.free)-1]
	} else {
		index = uint32(len(m.slots))
		m.slots = append(m.slots, slot{generation: 1})
	}
	entry := &m.slots[index]
	entry.active = true
	id := SoundID{Slot: index, Generation: entry.generation}
	m.pending = append(m.pending, Command{Kind: "play", ID: id, Format: format, Data: append([]byte(nil), data...), Volume: 1})
	return id, nil
}

// SetVolume queues a volume change in the range 0..1.
func (m *Mixer) SetVolume(id SoundID, volume float32) error {
	if volume < 0 || volume > 1 {
		return errors.New("audiocore: volume must be between 0 and 1")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.valid(id); err != nil {
		return err
	}
	m.pending = append(m.pending, Command{Kind: "volume", ID: id, Volume: volume})
	return nil
}

// Stop invalidates id and queues native stop/destruction.
func (m *Mixer) Stop(id SoundID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.valid(id); err != nil {
		return err
	}
	entry := &m.slots[id.Slot]
	entry.active = false
	entry.generation++
	if entry.generation == 0 {
		entry.generation = 1
	}
	m.free = append(m.free, id.Slot)
	m.pending = append(m.pending, Command{Kind: "stop", ID: id})
	return nil
}

// Drain applies pending commands in order, retaining failed and later commands.
func (m *Mixer) Drain(backend Backend) error {
	if backend == nil {
		return errors.New("audiocore: nil backend")
	}
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

func (m *Mixer) valid(id SoundID) error {
	if int(id.Slot) >= len(m.slots) {
		return fmt.Errorf("audiocore: invalid sound slot %d", id.Slot)
	}
	entry := m.slots[id.Slot]
	if !entry.active || entry.generation != id.Generation {
		return fmt.Errorf("audiocore: stale sound %v", id)
	}
	return nil
}
