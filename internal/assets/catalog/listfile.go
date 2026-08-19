package assetcatalog

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
)

// ListedPath is one normalized, case-insensitively deduplicated listfile entry.
type ListedPath struct {
	Original   string `json:"original"`
	Normalized string `json:"normalized"`
}

// Availability distinguishes community-listed knowledge from local presence.
type Availability struct {
	ListedPath
	Found bool   `json:"found"`
	Error string `json:"error,omitempty"`
}

// ListfileReport summarizes an optional community list against one MPQ stack.
type ListfileReport struct {
	Listed  int            `json:"listed"`
	Found   int            `json:"found"`
	Missing int            `json:"missing"`
	Entries []Availability `json:"entries"`
}

// ParseListfile accepts slash or backslash paths and preserves source text for diagnostics. Entries retain first-seen
// order while unsafe traversal paths and later case-insensitive spelling variants are omitted deterministically.
func ParseListfile(input io.Reader) ([]ListedPath, error) {
	scanner := bufio.NewScanner(input)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 4*1024*1024)

	seen := make(map[string]struct{})
	result := make([]ListedPath, 0, 4096)

	for scanner.Scan() {
		entry, ok := parseListfileEntry(scanner.Text())
		if !ok {
			continue
		}

		key := strings.ToLower(entry.Normalized)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}

		result = append(result, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("asset catalog: read listfile: %w", err)
	}

	return result, nil
}

// parseListfileEntry normalizes one source line without changing the spelling retained for diagnostics. BOM removal is
// deliberately limited to the beginning of a line, and cleaned parent traversal is rejected before filesystem access.
func parseListfileEntry(line string) (ListedPath, bool) {
	original := strings.TrimSpace(strings.TrimPrefix(line, "\ufeff"))
	if original == "" {
		return ListedPath{}, false
	}

	normalized := path.Clean(strings.TrimPrefix(strings.ReplaceAll(original, "\\", "/"), "/"))
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return ListedPath{}, false
	}

	return ListedPath{Original: original, Normalized: normalized}, true
}

// AuditListfile checks local availability without treating missing entries as invalid knowledge. Entries and counters
// stay in caller order because listfiles commonly encode release, locale, and mod knowledge worth preserving as given.
func AuditListfile(source fs.FS, entries []ListedPath) ListfileReport {
	report := ListfileReport{Listed: len(entries), Entries: make([]Availability, 0, len(entries))}

	for _, entry := range entries {
		availability := inspectListfileAvailability(source, entry)
		if availability.Found {
			report.Found++
		} else {
			report.Missing++
		}

		report.Entries = append(report.Entries, availability)
	}

	return report
}

// inspectListfileAvailability opens and immediately closes a listed path because the audit needs presence, not bytes.
// Close errors do not change availability: the original open already proved that the filesystem can resolve the entry.
func inspectListfileAvailability(source fs.FS, entry ListedPath) Availability {
	availability := Availability{ListedPath: entry}

	file, err := source.Open(entry.Normalized)
	if err != nil {
		availability.Error = err.Error()

		return availability
	}

	availability.Found = true
	_ = file.Close()

	return availability
}
