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

// ParseListfile accepts slash or backslash paths and preserves the source text
// for diagnostics. Duplicate spelling/case variants collapse deterministically.
func ParseListfile(input io.Reader) ([]ListedPath, error) {
	scanner := bufio.NewScanner(input)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 4*1024*1024)
	seen := make(map[string]struct{})
	result := make([]ListedPath, 0, 4096)
	for scanner.Scan() {
		original := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if original == "" {
			continue
		}
		normalized := path.Clean(strings.TrimPrefix(strings.ReplaceAll(original, "\\", "/"), "/"))
		if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
			continue
		}
		key := strings.ToLower(normalized)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, ListedPath{Original: original, Normalized: normalized})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("asset catalog: read listfile: %w", err)
	}
	return result, nil
}

// AuditListfile checks local availability without treating missing entries as
// invalid knowledge; listfiles commonly span releases, locales, and mods.
func AuditListfile(source fs.FS, entries []ListedPath) ListfileReport {
	report := ListfileReport{Listed: len(entries), Entries: make([]Availability, 0, len(entries))}
	for _, entry := range entries {
		availability := Availability{ListedPath: entry}
		file, err := source.Open(entry.Normalized)
		if err == nil {
			availability.Found = true
			report.Found++
			_ = file.Close()
		} else {
			availability.Error = err.Error()
			report.Missing++
		}
		report.Entries = append(report.Entries, availability)
	}
	return report
}
