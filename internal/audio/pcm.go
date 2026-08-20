package audio

import (
	"errors"
	"time"
)

// OpenPCMStream reserves a checked cinematic-bus stream for continuously fed signed-16-bit audio.
func (m *Mixer) OpenPCMStream(sampleRate, channels int) (SoundID, error) {
	if sampleRate <= 0 || channels <= 0 {
		return SoundID{}, errors.New("audiocore: PCM sample rate and channels must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	index := m.reserveSlotLocked()
	entry := &m.slots[index]
	entry.active = true
	entry.bus = "cinematic"
	entry.volume = 1
	entry.pcmRate = sampleRate
	entry.pcmFrames = 0

	id := SoundID{Slot: index, Generation: entry.generation}
	m.pending = append(m.pending, Command{
		Kind:     commandPCMOpen,
		ID:       id,
		Rate:     sampleRate,
		Channels: channels,
		Volume:   m.busVolumeLocked("cinematic"),
	})

	return id, nil
}

// WritePCM queues an owned copy of complete signed-16-bit samples so callers may reuse their input buffer immediately.
func (m *Mixer) WritePCM(id SoundID, pcm []byte) error {
	if len(pcm) == 0 || len(pcm)%2 != 0 {
		return errors.New("audiocore: PCM data must contain complete signed-16-bit samples")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validateSoundLocked(id); err != nil {
		return err
	}

	m.pending = append(m.pending, Command{
		Kind: commandPCMWrite,
		ID:   id,
		Data: append([]byte(nil), pcm...),
	})

	return nil
}

// ReportPCMFrames advances only a live PCM handle after the native owner confirms that frames were consumed.
func (m *Mixer) ReportPCMFrames(id SoundID, frames int) error {
	if frames <= 0 {
		return errors.New("audiocore: reported PCM frames must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validateSoundLocked(id); err != nil {
		return err
	}

	entry := &m.slots[id.Slot]
	if entry.pcmRate <= 0 {
		return errors.New("audiocore: sound is not a PCM stream")
	}

	entry.pcmFrames += int64(frames)

	return nil
}

// PCMTime reports device-consumed media time only after a live stream has processed at least one frame.
func (m *Mixer) PCMTime(id SoundID) (time.Duration, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.validateSoundLocked(id) != nil {
		return 0, false
	}

	entry := m.slots[id.Slot]
	if entry.pcmRate <= 0 || entry.pcmFrames == 0 {
		return 0, false
	}

	return time.Duration(entry.pcmFrames) * time.Second / time.Duration(entry.pcmRate), true
}
