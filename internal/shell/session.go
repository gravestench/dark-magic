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
	policy.Capabilities = append([]string(nil), policy.Capabilities...)
	return &Session{id: id, target: target, policy: policy, evaluator: evaluator, motd: defaultMOTD(target, policy), settings: NewTransientSettings()}, nil
}

func (s *Session) ID() string     { return s.id }
func (s *Session) Target() string { return s.target }
func (s *Session) Policy() Policy {
	policy := s.policy
	policy.Capabilities = append([]string(nil), policy.Capabilities...)
	return policy
}

func (s *Session) MOTD() string { return s.motd }

func (s *Session) AttachSettings(settings *Settings) {
	if settings == nil {
		return
	}
	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()
}

func (s *Session) Settings() *Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func defaultMOTD(target string, policy Policy) string {
	capabilities := "none registered for this target"
	if len(policy.Capabilities) > 0 {
		capabilities = strings.Join(policy.Capabilities, ", ")
	}
	return fmt.Sprintf("Welcome to the Dark Magic Lua shell.\nTarget: %s | Policy: %s\nRoot objects: dm (alias: darkmagic)\nCapabilities: %s\nTry d2.help(), d2.capabilities(), print(...), or printregs(). Press F2 for application logs.", target, policy.Name, capabilities)
}

func (s *Session) Submit(ctx context.Context, source string) Entry {
	source = strings.TrimSpace(source)
	entry := Entry{At: time.Now(), Source: source}
	if source == "" {
		entry.Error = "empty input"
		return entry
	}
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

func (s *Session) History() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.history...)
}

func (s *Session) Transcript() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Entry(nil), s.transcript...)
}

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
