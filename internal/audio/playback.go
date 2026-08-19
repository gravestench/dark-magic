package audio

import (
	"errors"
	"fmt"
	"time"
)

const defaultSoundBus = "sfx"

// fade tracks deterministic mixer-time interpolation without consulting the wall clock.
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

// Play queues a decoded-by-backend sound with canonical SFX routing and unity gain.
func (m *Mixer) Play(format string, data []byte) (SoundID, error) {
	return m.PlayWithOptions(format, data, PlayOptions{Bus: defaultSoundBus, Volume: 1})
}

// PlayWithOptions validates routing before reserving a slot, so rejected requests leave mixer state untouched.
func (m *Mixer) PlayWithOptions(format string, data []byte, options PlayOptions) (SoundID, error) {
	if format == "" || len(data) == 0 {
		return SoundID{}, errors.New("audiocore: format and sound data are required")
	}

	validatedOptions, err := validatePlayOptions(options)
	if err != nil {
		return SoundID{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureBusVolumesLocked()
	index := m.reserveSlotLocked()
	entry := &m.slots[index]
	entry.active = true
	entry.bus = validatedOptions.Bus
	entry.volume = validatedOptions.Volume
	entry.group = validatedOptions.Group
	entry.fade = nil

	id := SoundID{Slot: index, Generation: entry.generation}
	command := Command{
		Kind:   commandPlay,
		ID:     id,
		Format: format,
		// The queue owns its bytes so callers may immediately reuse their input buffer.
		Data:   append([]byte(nil), data...),
		Volume: validatedOptions.Volume * m.busVolumeLocked(validatedOptions.Bus),
		Pan:    validatedOptions.Pan,
		Loop:   validatedOptions.Loop,
		Stream: validatedOptions.Stream,
	}
	m.pending = append(m.pending, command)

	return id, nil
}

// Fade schedules a deterministic volume transition, using SetVolume for immediate changes to preserve command timing.
func (m *Mixer) Fade(id SoundID, volume float32, duration time.Duration) error {
	if err := validateVolume(volume); err != nil {
		return err
	}

	if duration <= 0 {
		return m.SetVolume(id, volume)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validateSoundLocked(id); err != nil {
		return err
	}

	entry := &m.slots[id.Slot]
	entry.fade = &fade{from: entry.volume, to: volume, duration: duration}

	return nil
}

// Advance progresses every active fade by caller-supplied mixer time.
// Explicit time keeps playback deterministic in tests and games.
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

		id := SoundID{Slot: uint32(index), Generation: entry.generation}

		m.pending = append(m.pending, Command{
			Kind:   commandVolume,
			ID:     id,
			Volume: entry.volume * m.busVolumeLocked(entry.bus),
		})
		if amount == 1 {
			entry.fade = nil
		}
	}
}

// StopGroup retires matching sounds in slot order so emitted stop commands remain deterministic.
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

		m.retireSlotLocked(uint32(index))
	}

	return nil
}

// SetVolume records logical gain separately from bus gain so later bus updates retain the sound's own volume.
func (m *Mixer) SetVolume(id SoundID, volume float32) error {
	if err := validateVolume(volume); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validateSoundLocked(id); err != nil {
		return err
	}

	entry := &m.slots[id.Slot]
	entry.volume = volume
	m.pending = append(m.pending, Command{
		Kind:   commandVolume,
		ID:     id,
		Volume: volume * m.busVolumeLocked(entry.bus),
	})

	return nil
}

// SetPan validates the backend's normalized stereo range before queuing an ordered pan update.
func (m *Mixer) SetPan(id SoundID, pan float32) error {
	if pan < -1 || pan > 1 {
		return errors.New("audiocore: pan must be between -1 and 1")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validateSoundLocked(id); err != nil {
		return err
	}

	m.pending = append(m.pending, Command{Kind: commandPan, ID: id, Pan: pan})

	return nil
}

// SetBusVolume updates matching sounds in slot order so routing changes produce a deterministic command sequence.
func (m *Mixer) SetBusVolume(bus string, volume float32) error {
	if !validBus(bus) {
		return fmt.Errorf("audiocore: unknown bus %q", bus)
	}

	if err := validateVolume(volume); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureBusVolumesLocked()
	m.buses[bus] = volume

	for index, entry := range m.slots {
		if !entry.active || entry.bus != bus {
			continue
		}

		m.pending = append(m.pending, Command{
			Kind: commandVolume,
			ID: SoundID{
				Slot:       uint32(index),
				Generation: entry.generation,
			},
			Volume: entry.volume * volume,
		})
	}

	return nil
}

// validatePlayOptions applies the zero-value bus default before enforcing the public routing and range contract.
func validatePlayOptions(options PlayOptions) (PlayOptions, error) {
	if options.Bus == "" {
		options.Bus = defaultSoundBus
	}

	if !validBus(options.Bus) {
		return PlayOptions{}, fmt.Errorf("audiocore: unknown bus %q", options.Bus)
	}

	if err := validateVolume(options.Volume); err != nil {
		return PlayOptions{}, err
	}

	if options.Pan < -1 || options.Pan > 1 {
		return PlayOptions{}, errors.New("audiocore: pan must be between -1 and 1")
	}

	return options, nil
}

// validateVolume centralizes the inclusive normalized gain contract shared by sounds, fades, and buses.
func validateVolume(volume float32) error {
	if volume < 0 || volume > 1 {
		return errors.New("audiocore: volume must be between 0 and 1")
	}

	return nil
}

// validBus defines the closed routing set so all entry points accept exactly the same names.
func validBus(bus string) bool {
	switch bus {
	case "music", "ui", defaultSoundBus, "ambience", "speech", "cinematic":
		return true
	}

	return false
}

// ensureBusVolumesLocked preserves Mixer zero-value usability by allocating overrides only when a caller sets one.
func (m *Mixer) ensureBusVolumesLocked() {
	if m.buses == nil {
		m.buses = make(map[string]float32)
	}
}

// busVolumeLocked returns unity gain for an unset bus so zero-value routing remains audible.
func (m *Mixer) busVolumeLocked(bus string) float32 {
	if volume, ok := m.buses[bus]; ok {
		return volume
	}

	return 1
}
