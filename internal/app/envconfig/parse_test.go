package envconfig

import (
	"strings"
	"testing"
)

// TestParseReadableDotEnvSyntax documents the deliberate human-editable subset
// and guards credentials containing hashes, spaces, and quoted escape sequences.
func TestParseReadableDotEnvSyntax(t *testing.T) {
	document := strings.Join([]string{
		"# comment",
		"export ONE=plain",
		"TWO='two words'",
		`THREE="line\nthree"`,
		"FOUR=value # comment",
	}, "\n")

	values, err := Parse(strings.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"ONE":   "plain",
		"TWO":   "two words",
		"THREE": "line\nthree",
		"FOUR":  "value",
	}
	for key, value := range want {
		if values[key] != value {
			t.Errorf("%s = %q, want %q", key, values[key], value)
		}
	}
}
