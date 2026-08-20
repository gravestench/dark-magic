package envconfig

import (
	"os"
	"strings"
	"testing"
)

// TestUpdatePreservesTemplateAndRejectsUnknownKeys protects both sides of the
// configuration contract: humans retain documentation and tools cannot invent schema.
func TestUpdatePreservesTemplateAndRejectsUnknownKeys(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("DARK_MAGIC_CONFIG_DIR", directory)

	path, err := Update("client", map[string]string{"MPQ_DIRECTORY": "/Games/Diablo II"})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `MPQ_DIRECTORY="/Games/Diablo II"`) {
		t.Fatalf("updated file does not contain MPQ_DIRECTORY: %s", data)
	}

	if !strings.Contains(string(data), "# Dark Magic client") {
		t.Fatalf("updated file does not preserve the template comment: %s", data)
	}

	if _, err := Update("client", map[string]string{"TYPO": "value"}); err == nil {
		t.Fatal("unknown template key was accepted")
	}
}
