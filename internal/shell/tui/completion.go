package tui

import (
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/shell"
)

// complete loads candidates once per editor state, inserts a shared prefix, and
// cycles one concrete candidate in the requested direction.
func (m *Model) complete(reverse bool) {
	if len(m.candidates) == 0 {
		if !m.loadCompletionCandidates() {
			return
		}
	}

	if len(m.candidates) == 0 {
		m.status = "no completions"

		return
	}

	if reverse {
		m.candidateAt = (m.candidateAt - 1 + len(m.candidates)) % len(m.candidates)
	} else {
		m.candidateAt = (m.candidateAt + 1) % len(m.candidates)
	}

	candidate := m.candidates[m.candidateAt]
	m.replaceToken(candidate.Value)
	m.status = fmt.Sprintf("%d/%d %s", m.candidateAt+1, len(m.candidates), candidate.Detail)
}

// loadCompletionCandidates reports evaluator errors, retains sorted candidates,
// and pauses after extending a shared prefix when multiple choices remain.
func (m *Model) loadCompletionCandidates() bool {
	originalToken := completionToken(m.input.Value())

	candidates, err := m.session.Complete(m.ctx, m.input.Value())
	if err != nil {
		m.status = err.Error()

		return false
	}

	m.candidates = candidates
	m.candidateAt = -1

	prefix := shell.SharedPrefix(candidates)
	if prefix != "" {
		m.replaceToken(prefix)
	}

	if len(candidates) > 1 && prefix != originalToken {
		m.status = fmt.Sprintf("%d candidates", len(candidates))

		return false
	}

	return true
}

// replaceToken replaces only the completion suffix and lets textarea place the
// cursor at the end of the resulting value, matching prior terminal editing.
func (m *Model) replaceToken(value string) {
	source := m.input.Value()
	token := completionToken(source)

	m.input.SetValue(strings.TrimSuffix(source, token) + value)
}

// moveHistory clamps navigation and treats the position after the newest entry
// as an empty draft, without retaining completion candidates from old source.
func (m *Model) moveHistory(delta int) {
	history := m.session.History()
	if len(history) == 0 {
		return
	}

	m.history = max(0, min(len(history), m.history+delta))
	if m.history == len(history) {
		m.input.Reset()
	} else {
		m.input.SetValue(history[m.history])
	}

	m.candidates = nil
}

// completionToken extracts the ASCII identifier/member suffix used by evaluator
// completion while leaving preceding Lua punctuation untouched.
func completionToken(source string) string {
	index := strings.LastIndexFunc(source, func(current rune) bool {
		return !completionCharacter(current)
	})

	return source[index+1:]
}

// completionCharacter defines the compatible identifier and dotted-member set.
func completionCharacter(current rune) bool {
	return current == '_' || current == '.' ||
		current >= 'a' && current <= 'z' ||
		current >= 'A' && current <= 'Z' ||
		current >= '0' && current <= '9'
}
