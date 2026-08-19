package clientsession

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
)

// Submit sends one staged or immediate intent under a stable credential. reconnectMu permits
// concurrent streams but prevents credential rotation until every request using the old value ends.
func (session *Session) Submit(ctx context.Context, intent gameserver.CommandIntent) error {
	session.reconnectMu.RLock()
	defer session.reconnectMu.RUnlock()

	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()

		return errors.New("client session: closed")
	}

	if intent.ObservedServerTick == 0 {
		intent.ObservedServerTick = session.timelineLocked(time.Now()).Prediction.Tick
	}

	if intent.TargetTick == 0 {
		intent.TargetTick = intent.ObservedServerTick + 2
	}

	transport, credential := session.transport, session.credential
	session.mu.Unlock()

	// Never retain session.mu across I/O: presentation reads pending input and clock state each frame.
	if err := transport.Submit(ctx, credential, intent); err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return nil
	}

	if session.pending == nil {
		session.pending = make(map[uint64]gameserver.CommandIntent)
	}

	session.pending[intent.Sequence] = intent

	return nil
}

// StageInput makes locally produced input visible to prediction before transport completes. Reusing a
// sequence is idempotent only when every authority-relevant field is identical.
func (session *Session) StageInput(intent gameserver.CommandIntent) error {
	if intent.Sequence == 0 || intent.Kind == "" {
		return errors.New("client session: invalid staged input")
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return errors.New("client session: closed")
	}

	if existing, found := session.pending[intent.Sequence]; found {
		if conflictingInput(existing, intent) {
			return errors.New("client session: conflicting staged input sequence")
		}

		return nil
	}

	intent.Payload = append(json.RawMessage(nil), intent.Payload...)
	session.pending[intent.Sequence] = intent

	return nil
}

// conflictingInput prevents one sequence number from referring to two different deterministic commands.
func conflictingInput(existing, candidate gameserver.CommandIntent) bool {
	return existing.ObservedServerTick != candidate.ObservedServerTick ||
		existing.TargetTick != candidate.TargetTick ||
		existing.Kind != candidate.Kind ||
		!bytes.Equal(existing.Payload, candidate.Payload)
}

// DiscardInput removes a staged command that transport rejected before authority admission.
func (session *Session) DiscardInput(sequence uint64) {
	session.mu.Lock()
	delete(session.pending, sequence)
	session.mu.Unlock()
}

// PendingInputs returns a defensive sequence-ordered history for acknowledgement and local replay.
func (session *Session) PendingInputs() []gameserver.CommandIntent {
	session.mu.Lock()
	defer session.mu.Unlock()

	sequences := make([]uint64, 0, len(session.pending))
	for sequence := range session.pending {
		sequences = append(sequences, sequence)
	}

	sort.Slice(sequences, func(i, j int) bool { return sequences[i] < sequences[j] })

	result := make([]gameserver.CommandIntent, 0, len(sequences))
	for _, sequence := range sequences {
		intent := session.pending[sequence]
		intent.Payload = append(json.RawMessage(nil), intent.Payload...)
		result = append(result, intent)
	}

	return result
}

// discardAcknowledgedLocked removes the contiguous input prefix authority has accepted. Callers hold mu.
func (session *Session) discardAcknowledgedLocked(acknowledged uint64) {
	for sequence := range session.pending {
		if sequence <= acknowledged {
			delete(session.pending, sequence)
		}
	}
}
