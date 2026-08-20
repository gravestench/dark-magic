// Package shell owns renderer-independent interactive shell sessions shared by
// graphical clients, headless game servers, and realms.
package shell

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Policy describes the authority granted to one shell session.
type Policy struct {
	Name         string
	Capabilities []string
	Mutable      bool
}

// Result is a structured evaluator response. Text is suitable for terminals;
// Kind lets graphical adapters select richer presentation without parsing it.
type Result struct {
	Kind string
	Text string
}

// Candidate is one side-effect-free completion result.
type Candidate struct {
	Value  string
	Detail string
}

// Evaluator executes one source submission and inspects completion candidates.
type Evaluator interface {
	Evaluate(context.Context, string) (Result, error)
	Complete(context.Context, string) ([]Candidate, error)
	Close() error
}

// Entry is one immutable transcript event.
type Entry struct {
	At          time.Time
	CompletedAt time.Time
	Source      string
	Result      Result
	Error       string
}

// Session serializes history and transcript state around one explicit runtime.
type Session struct {
	mu                 sync.RWMutex
	evalMu             sync.Mutex
	id                 string
	target             string
	policy             Policy
	evaluator          Evaluator
	history            []string
	transcript         []Entry
	logs               *LogBuffer
	transcriptRevision uint64
	motd               string
	settings           *Settings
	closed             bool
}

// NewSession binds one explicit target runtime and copied authority policy. A
// session never discovers a process-global VM or widens its own capabilities.
func NewSession(id, target string, policy Policy, evaluator Evaluator) (*Session, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(target) == "" {
		return nil, errors.New("shell: session id and target are required")
	}

	if evaluator == nil {
		return nil, errors.New("shell: evaluator is required")
	}

	// Copy the capability slice because callers often reuse policy configuration
	// across sessions with different lifetimes.
	policy.Capabilities = append([]string(nil), policy.Capabilities...)

	return &Session{
		id:        id,
		target:    target,
		policy:    policy,
		evaluator: evaluator,
		motd:      defaultMOTD(target, policy),
		settings:  NewTransientSettings(),
	}, nil
}

// ID returns the stable identity used to distinguish concurrent shell sessions.
func (s *Session) ID() string { return s.id }

// Target identifies the runtime endpoint whose evaluator owns this session.
func (s *Session) Target() string { return s.target }

// Policy returns a defensive authority snapshot so presentation code cannot
// widen the capabilities retained by the session.
func (s *Session) Policy() Policy {
	policy := s.policy
	policy.Capabilities = append([]string(nil), policy.Capabilities...)

	return policy
}

// MOTD returns the immutable introduction generated from the target and copied policy.
func (s *Session) MOTD() string { return s.motd }

// AttachSettings replaces the presentation settings shared by shell adapters.
// Nil is ignored so optional application wiring cannot erase usable defaults.
func (s *Session) AttachSettings(settings *Settings) {
	if settings == nil {
		return
	}

	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()
}

// Settings returns the currently attached settings object. The object owns its
// own synchronization, so callers may retain it after the session lock is released.
func (s *Session) Settings() *Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.settings
}

// defaultMOTD exposes target authority and discovery hints before the first
// command, preventing a read-only or restricted session from appearing unrestricted.
func defaultMOTD(target string, policy Policy) string {
	capabilities := "none registered for this target"
	if len(policy.Capabilities) > 0 {
		capabilities = strings.Join(policy.Capabilities, ", ")
	}

	const message = "Welcome to the Dark Magic Lua shell.\n" +
		"Target: %s | Policy: %s\n" +
		"Root objects: dm (alias: darkmagic)\n" +
		"Capabilities: %s\n" +
		"Try d2legacy.help(), d2legacy.capabilities(), print(...), or printregs(). " +
		"Press F2 for application logs."

	return fmt.Sprintf(message, target, policy.Name, capabilities)
}

// Submit serializes evaluator access, records non-empty commands in history,
// and publishes the completed transcript entry after evaluation finishes.
func (s *Session) Submit(ctx context.Context, source string) Entry {
	source = strings.TrimSpace(source)

	entry := Entry{At: time.Now(), Source: source}
	if source == "" {
		entry.Error = "empty input"

		return entry
	}

	// Acquire evaluator ownership while holding the state lock. Close uses the
	// same order, so no evaluation can begin after the session becomes closed.
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()

		entry.Error = "shell session is closed"

		return entry
	}

	s.history = append(s.history, source)
	s.evalMu.Lock()
	s.mu.Unlock()

	result, err := s.evaluator.Evaluate(ctx, source)
	s.evalMu.Unlock()

	entry.CompletedAt = time.Now()

	entry.Result = result
	if err != nil {
		entry.Error = err.Error()
	}

	s.mu.Lock()
	s.transcript = append(s.transcript, entry)
	s.transcriptRevision++
	s.mu.Unlock()

	return entry
}

// Complete serializes runtime inspection with evaluation and close, then sorts
// candidates so every adapter presents the same deterministic order.
func (s *Session) Complete(ctx context.Context, prefix string) ([]Candidate, error) {
	s.mu.RLock()

	if s.closed {
		s.mu.RUnlock()

		return nil, errors.New("shell: session is closed")
	}

	s.evalMu.Lock()
	s.mu.RUnlock()

	candidates, err := s.evaluator.Complete(ctx, prefix)
	s.evalMu.Unlock()

	if err != nil {
		return nil, err
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Value < candidates[j].Value })

	return candidates, nil
}

// History returns a copied command list so editor navigation cannot rewrite session state.
func (s *Session) History() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]string(nil), s.history...)
}

// Transcript returns a copied event slice; Entry fields are immutable values once published.
func (s *Session) Transcript() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]Entry(nil), s.transcript...)
}

// Close prevents new work before waiting for an active evaluation, then closes
// the evaluator exactly once without holding the session state lock.
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()

		return nil
	}

	s.closed = true
	s.mu.Unlock()

	s.evalMu.Lock()
	defer s.evalMu.Unlock()

	return s.evaluator.Close()
}

// SharedPrefix returns the insertion shared by every candidate.
func SharedPrefix(candidates []Candidate) string {
	if len(candidates) == 0 {
		return ""
	}

	prefix := candidates[0].Value
	for _, candidate := range candidates[1:] {
		for !strings.HasPrefix(candidate.Value, prefix) && prefix != "" {
			prefix = prefix[:len(prefix)-1]
		}
	}

	return prefix
}
