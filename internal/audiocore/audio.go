// Package audiocore defines the renderer-thread command boundary for audio.
package audiocore

import (
	"errors"
	"fmt"
	"sync"
	"time"
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
	Pan    float32
	Loop   bool
	Stream bool
}

// Backend consumes audio commands on the native audio owner thread.
type Backend interface{ Apply(Command) error }

type slot struct {
	generation uint32
	active     bool
	bus        string
	volume     float32
	group      string
	fade       *fade
}

type fade struct {
	from, to          float32
	elapsed, duration time.Duration
}

// PlayOptions selects routing and playback behavior for one sound.
type PlayOptions struct {
	Bus    string
	Volume float32
	Pan    float32
	Loop   bool
	Group  string
	Stream bool
}

// Mixer accepts concurrent sound requests and queues backend commands.
type Mixer struct {
	mu      sync.Mutex
	slots   []slot
	free    []uint32
	pending []Command
	buses   map[string]float32
}

// Play queues a decoded-by-backend sound asset.
func (m *Mixer) Play(format string, data []byte) (SoundID, error) {
	return m.PlayWithOptions(format, data, PlayOptions{Bus: "sfx", Volume: 1})
}

// PlayWithOptions queues a sound routed through a named bus.
func (m *Mixer) PlayWithOptions(format string, data []byte, options PlayOptions) (SoundID, error) {
	if format == "" || len(data) == 0 {
		return SoundID{}, errors.New("audiocore: format and sound data are required")
	}
	if options.Bus == "" {
		options.Bus = "sfx"
	}
	if !validBus(options.Bus) {
		return SoundID{}, fmt.Errorf("audiocore: unknown bus %q", options.Bus)
	}
	if options.Volume < 0 || options.Volume > 1 {
		return SoundID{}, errors.New("audiocore: volume must be between 0 and 1")
	}
	if options.Pan < -1 || options.Pan > 1 {
		return SoundID{}, errors.New("audiocore: pan must be between -1 and 1")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buses == nil {
		m.buses = make(map[string]float32)
	}
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
	entry.bus = options.Bus
	entry.volume = options.Volume
	entry.group = options.Group
	entry.fade = nil
	id := SoundID{Slot: index, Generation: entry.generation}
	m.pending = append(m.pending, Command{Kind: "play", ID: id, Format: format, Data: append([]byte(nil), data...), Volume: options.Volume * m.busVolume(options.Bus), Pan: options.Pan, Loop: options.Loop, Stream: options.Stream})
	return id, nil
}

// Fade changes a sound's volume over deterministic mixer time.
func (m *Mixer) Fade(id SoundID, volume float32, duration time.Duration) error {
	if volume < 0 || volume > 1 {
		return errors.New("audiocore: volume must be between 0 and 1")
	}
	if duration <= 0 {
		return m.SetVolume(id, volume)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.valid(id); err != nil {
		return err
	}
	entry := &m.slots[id.Slot]
	entry.fade = &fade{from: entry.volume, to: volume, duration: duration}
	return nil
}

// Advance progresses fades without consulting wall-clock time.
func (m *Mixer) Advance(delta time.Duration) {
	if delta <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for index := range m.slots {
		entry := &m.slots[index]
		if !entry.active || entry.fade == nil {
			continue
		}
		entry.fade.elapsed += delta
		amount := float32(entry.fade.elapsed) / float32(entry.fade.duration)
		if amount >= 1 {
			amount = 1
		}
		entry.volume = entry.fade.from + (entry.fade.to-entry.fade.from)*amount
		m.pending = append(m.pending, Command{Kind: "volume", ID: SoundID{Slot: uint32(index), Generation: entry.generation}, Volume: entry.volume * m.busVolume(entry.bus)})
		if amount == 1 {
			entry.fade = nil
		}
	}
}

// StopGroup stops all active sounds carrying group.
func (m *Mixer) StopGroup(group string) error {
	if group == "" {
		return errors.New("audiocore: group is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for index := range m.slots {
		entry := &m.slots[index]
		if !entry.active || entry.group != group {
			continue
		}
		id := SoundID{Slot: uint32(index), Generation: entry.generation}
		entry.active = false
		entry.generation++
		if entry.generation == 0 {
			entry.generation = 1
		}
		m.free = append(m.free, uint32(index))
		m.pending = append(m.pending, Command{Kind: "stop", ID: id})
	}
	return nil
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
	m.slots[id.Slot].volume = volume
	m.pending = append(m.pending, Command{Kind: "volume", ID: id, Volume: volume * m.busVolume(m.slots[id.Slot].bus)})
	return nil
}

// SetPan queues stereo pan in the range -1 (left) through 1 (right).
func (m *Mixer) SetPan(id SoundID, pan float32) error {
	if pan < -1 || pan > 1 {
		return errors.New("audiocore: pan must be between -1 and 1")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.valid(id); err != nil {
		return err
	}
	m.pending = append(m.pending, Command{Kind: "pan", ID: id, Pan: pan})
	return nil
}

// SetBusVolume changes a bus and updates every active sound routed through it.
func (m *Mixer) SetBusVolume(bus string, volume float32) error {
	if !validBus(bus) {
		return fmt.Errorf("audiocore: unknown bus %q", bus)
	}
	if volume < 0 || volume > 1 {
		return errors.New("audiocore: volume must be between 0 and 1")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buses == nil {
		m.buses = make(map[string]float32)
	}
	m.buses[bus] = volume
	for index, entry := range m.slots {
		if entry.active && entry.bus == bus {
			m.pending = append(m.pending, Command{Kind: "volume", ID: SoundID{Slot: uint32(index), Generation: entry.generation}, Volume: entry.volume * volume})
		}
	}
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

func validBus(bus string) bool {
	switch bus {
	case "music", "ui", "sfx", "ambience", "speech":
		return true
	}
	return false
}

func (m *Mixer) busVolume(bus string) float32 {
	if volume, ok := m.buses[bus]; ok {
		return volume
	}
	return 1
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
