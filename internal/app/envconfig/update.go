package envconfig

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Update edits only keys declared by the role template and preserves surrounding
// documentation. This makes generated files safe for both tooling and human ownership.
func Update(role string, updates map[string]string) (string, error) {
	path, _, err := Install(role)
	if err != nil {
		return "", err
	}

	allowed, err := templateValues(role)
	if err != nil {
		return "", err
	}

	if err := validateUpdates(role, updates, allowed); err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	updated := updateDocument(data, updates)

	return path, writePrivate(path, updated)
}

// templateValues makes the embedded template the schema for writable settings,
// preventing callers from smuggling arbitrary process variables into the file.
func templateValues(role string) (map[string]string, error) {
	data, err := templates.ReadFile("templates/" + role + ".env")
	if err != nil {
		return nil, err
	}

	return Parse(strings.NewReader(string(data)))
}

// validateUpdates fails the entire request on unknown keys so typos cannot appear
// to persist successfully while the target process silently ignores them.
func validateUpdates(role string, updates, allowed map[string]string) error {
	for key := range updates {
		if _, found := allowed[key]; !found {
			return fmt.Errorf("environment variable %q is not part of the %s template", key, role)
		}
	}

	return nil
}

// updateDocument preserves existing line order and comments, then appends absent
// supported keys in sorted order to keep repeated updates reviewable and stable.
func updateDocument(data []byte, updates map[string]string) []byte {
	remaining := copyValues(updates)

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	for index, line := range lines {
		key := assignmentKey(line)

		value, found := remaining[key]
		if !found {
			continue
		}

		lines[index] = key + "=" + strconv.Quote(value)
		delete(remaining, key)
	}

	for _, key := range sortedKeys(remaining) {
		lines = append(lines, key+"="+strconv.Quote(remaining[key]))
	}

	return []byte(strings.Join(lines, "\n") + "\n")
}

// assignmentKey recognizes only syntactically valid assignments; comments and
// malformed human text are preserved rather than accidentally rewritten.
func assignmentKey(line string) string {
	trimmed := strings.TrimSpace(line)

	separator := strings.IndexByte(trimmed, '=')
	if separator <= 0 {
		return ""
	}

	return strings.TrimSpace(trimmed[:separator])
}

// copyValues gives Update private ownership of its working set, allowing keys to
// be consumed during rewriting without surprising the caller.
func copyValues(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}

	return result
}

// sortedKeys removes Go map iteration order from generated configuration diffs.
func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
