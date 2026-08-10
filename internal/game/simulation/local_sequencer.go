package simulation

import "sync"

// LocalSequencer assigns the single monotonically increasing sequence owned by
// each local player connection. Individual input producers describe intent;
// they do not each own an independent transport sequence.
//
// Remote commands never pass through this helper because their client-assigned
// sequence is part of the untrusted input validated by Admitter.
type LocalSequencer struct {
	mu       sync.Mutex
	sequence map[string]uint64
}

func NewLocalSequencer() *LocalSequencer {
	return &LocalSequencer{sequence: make(map[string]uint64)}
}

// Assign returns a copied slice whose commands are numbered in collection
// order. Payload bytes are immutable command input and can remain shared.
func (sequencer *LocalSequencer) Assign(commands []Command) []Command {
	if len(commands) == 0 {
		return nil
	}
	sequencer.mu.Lock()
	defer sequencer.mu.Unlock()
	result := append([]Command(nil), commands...)
	for index := range result {
		player := result[index].Player
		sequencer.sequence[player]++
		result[index].Sequence = sequencer.sequence[player]
	}
	return result
}
