package shell

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogEntry is the presentation-neutral subset of one structured process log.
type LogEntry struct {
	At         time.Time
	Level      string
	Message    string
	Attributes map[string]any
}

// LogBuffer retains a bounded process-log tail for interactive shell views.
type LogBuffer struct {
	mu       sync.RWMutex
	limit    int
	entries  []LogEntry
	revision uint64
}

// NewLogBuffer creates a bounded tail and clamps invalid limits to one entry so
// Append never needs an unbounded or zero-capacity special case.
func NewLogBuffer(limit int) *LogBuffer {
	if limit < 1 {
		limit = 1
	}

	return &LogBuffer{limit: limit}
}

// Append clones caller-owned attributes before publishing the entry and keeps
// the newest bounded tail while incrementing the observable revision once.
func (b *LogBuffer) Append(entry LogEntry) {
	entry.Attributes = cloneAttributes(entry.Attributes)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.revision++

	if len(b.entries) == b.limit {
		copy(b.entries, b.entries[1:])
		b.entries[len(b.entries)-1] = entry

		return
	}

	b.entries = append(b.entries, entry)
}

// Revision returns the append revision used by adapters to avoid rebuilding unchanged views.
func (b *LogBuffer) Revision() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.revision
}

// Snapshot deep-copies the retained entries and their attribute maps so presentation code cannot mutate the buffer.
func (b *LogBuffer) Snapshot() []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]LogEntry, len(b.entries))
	for index, entry := range b.entries {
		entry.Attributes = cloneAttributes(entry.Attributes)
		result[index] = entry
	}

	return result
}

// cloneAttributes copies the top-level structured attributes carried by slog.
// Nested values remain owned by their producer and are treated as immutable.
func cloneAttributes(attributes map[string]any) map[string]any {
	if len(attributes) == 0 {
		return nil
	}

	result := make(map[string]any, len(attributes))
	for key, value := range attributes {
		result[key] = value
	}

	return result
}

// TimelineEvent merges commands, results, errors, and process logs in the
// order they actually occurred.
type TimelineEvent struct {
	At   time.Time
	Kind string
	Text string
}

// AttachLogs connects the process-log tail used by combined timelines. Nil intentionally detaches logging.
func (s *Session) AttachLogs(logs *LogBuffer) {
	s.mu.Lock()
	s.logs = logs
	s.mu.Unlock()
}

// Logs snapshots the attached buffer without retaining the session lock during the buffer copy.
func (s *Session) Logs() []LogEntry {
	s.mu.RLock()
	logs := s.logs
	s.mu.RUnlock()

	if logs == nil {
		return nil
	}

	return logs.Snapshot()
}

// Timeline stably merges transcript and process-log events by timestamp. Stable
// sorting preserves source order for equal timestamps instead of inventing a tie-breaker.
func (s *Session) Timeline() []TimelineEvent {
	events := append(s.TranscriptTimeline(), s.LogTimeline()...)
	sort.SliceStable(events, func(i, j int) bool { return events[i].At.Before(events[j].At) })

	return events
}

// TranscriptTimeline returns only Lua commands, values, prints, and errors.
func (s *Session) TranscriptTimeline() []TimelineEvent {
	entries := s.Transcript()
	events := make([]TimelineEvent, 0, len(entries)*2+1)

	if s.MOTD() != "" {
		events = append(events, TimelineEvent{Kind: "motd", Text: s.MOTD()})
	}

	for _, entry := range entries {
		events = append(events, TimelineEvent{At: entry.At, Kind: "command", Text: entry.Source})

		if entry.Error != "" {
			events = append(events, TimelineEvent{At: entry.CompletedAt, Kind: "error", Text: entry.Error})
		} else if entry.Result.Text != "" {
			events = append(events, TimelineEvent{At: entry.CompletedAt, Kind: "value", Text: entry.Result.Text})
		}
	}

	return events
}

// LogTimeline returns only structured process logs.
func (s *Session) LogTimeline() []TimelineEvent {
	logs := s.Logs()
	events := make([]TimelineEvent, 0, len(logs))

	for _, entry := range logs {
		text := entry.At.Format("15:04:05.000") + " " + strings.ToUpper(entry.Level) + " " + entry.Message

		encoded, err := json.Marshal(entry.Attributes)
		if err == nil && string(encoded) != "null" && string(encoded) != "{}" {
			text += " " + string(encoded)
		}

		events = append(events, TimelineEvent{At: entry.At, Kind: "log-" + strings.ToLower(entry.Level), Text: text})
	}

	return events
}

// TimelineRevision combines transcript and log revisions so adapters can use a
// single inexpensive invalidation token for merged views.
func (s *Session) TimelineRevision() uint64 {
	s.mu.RLock()
	revision, logs := s.transcriptRevision, s.logs
	s.mu.RUnlock()

	if logs != nil {
		revision += logs.Revision()
	}

	return revision
}
