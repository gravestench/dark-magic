package raylibshell

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/gravestench/dark-magic/internal/shell"
)

// submit clears the editor before launching serialized evaluation, allowing the
// renderer loop to remain responsive while the buffered channel reports completion.
func (o *Overlay) submit(ctx context.Context) {
	source := strings.TrimSpace(o.input)
	if source == "" || o.busy {
		return
	}

	o.busy = true
	o.input = ""
	o.cursor = 0
	o.resetCompletion()
	o.luaOffset = 0

	go func() {
		o.finished <- o.session.Submit(ctx, source)
	}()
}

// complete loads candidates once per editing state, inserts their shared prefix,
// and cycles the established forward/reverse selection order.
func (o *Overlay) complete(ctx context.Context, reverse bool) {
	if len(o.candidates) == 0 {
		if !o.loadCompletionCandidates(ctx) {
			return
		}
	}

	if reverse {
		o.candidateAt = (o.candidateAt - 1 + len(o.candidates)) % len(o.candidates)
	} else {
		// candidateAt starts before the first candidate so the initial forward
		// cycle selects the first value in the Session's stable sorted order.
		o.candidateAt = (o.candidateAt + 1) % len(o.candidates)
	}

	o.replaceToken(o.candidates[o.candidateAt].Value)
}

// loadCompletionCandidates clamps the rune cursor, requests side-effect-free
// candidates, and inserts a longer common prefix before individual cycling.
func (o *Overlay) loadCompletionCandidates(ctx context.Context) bool {
	runes := []rune(o.input)
	o.cursor = max(0, min(len(runes), o.cursor))

	prefix := string(runes[:o.cursor])
	original := completionToken(prefix)

	candidates, err := o.session.Complete(ctx, prefix)
	if err != nil || len(candidates) == 0 {
		return false
	}

	o.candidates = candidates
	o.candidateAt = -1

	shared := shell.SharedPrefix(candidates)
	if shared != "" {
		o.replaceToken(shared)
	}

	return len(candidates) <= 1 || shared == original
}

// replaceToken substitutes only the identifier suffix before the rune cursor,
// retaining both prior source and text following the cursor.
func (o *Overlay) replaceToken(value string) {
	runes := []rune(o.input)
	o.cursor = max(0, min(len(runes), o.cursor))

	prefix := string(runes[:o.cursor])
	token := completionToken(prefix)
	before := []rune(strings.TrimSuffix(prefix, token))
	replacement := []rune(value)

	o.input = string(append(append(before, replacement...), runes[o.cursor:]...))
	o.cursor = len(before) + len(replacement)
}

// moveHistory clamps navigation to the saved command range and uses the slot
// after the newest command as the empty draft position.
func (o *Overlay) moveHistory(delta int) {
	history := o.session.History()
	if len(history) == 0 {
		return
	}

	o.history = max(0, min(len(history), o.history+delta))
	if o.history == len(history) {
		o.input = ""
	} else {
		o.input = history[o.history]
	}

	o.cursor = utf8.RuneCountInString(o.input)
	o.resetCompletion()
}

// resetCompletion invalidates candidates after any editor or view change.
func (o *Overlay) resetCompletion() {
	o.candidates = nil
	o.candidateAt = -1
}

// insert adds runes at the clamped cursor and advances by inserted rune count,
// preserving UTF-8 boundaries for multi-byte text.
func (o *Overlay) insert(value string) {
	runes := []rune(o.input)
	o.cursor = max(0, min(len(runes), o.cursor))

	inserted := []rune(value)
	o.input = string(append(append(runes[:o.cursor:o.cursor], inserted...), runes[o.cursor:]...))
	o.cursor += len(inserted)
}

// backspace removes the rune before the clamped cursor without splitting UTF-8.
func (o *Overlay) backspace() {
	runes := []rune(o.input)

	o.cursor = max(0, min(len(runes), o.cursor))
	if o.cursor == 0 {
		return
	}

	o.input = string(append(runes[:o.cursor-1], runes[o.cursor:]...))
	o.cursor--
}

// delete removes the rune at the clamped cursor without moving the cursor.
func (o *Overlay) delete() {
	runes := []rune(o.input)

	o.cursor = max(0, min(len(runes), o.cursor))
	if o.cursor == len(runes) {
		return
	}

	o.input = string(append(runes[:o.cursor], runes[o.cursor+1:]...))
}

// inputWithCaret returns a display-only copy with a caret rune inserted at the
// clamped cursor; the editor value itself remains unchanged.
func (o *Overlay) inputWithCaret() string {
	runes := []rune(o.input)
	cursor := max(0, min(len(runes), o.cursor))

	return string(append(append(runes[:cursor:cursor], '|'), runes[cursor:]...))
}

// completionToken extracts the ASCII identifier/member suffix recognized by the
// shared Lua completion contract.
func completionToken(source string) string {
	index := strings.LastIndexFunc(source, func(current rune) bool {
		return !completionCharacter(current)
	})

	return source[index+1:]
}

// completionCharacter keeps completion paths compatible with Lua identifiers
// and dotted member access without accepting surrounding punctuation.
func completionCharacter(current rune) bool {
	return current == '_' || current == '.' ||
		current >= 'a' && current <= 'z' ||
		current >= 'A' && current <= 'Z' ||
		current >= '0' && current <= '9'
}
