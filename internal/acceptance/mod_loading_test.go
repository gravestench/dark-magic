package acceptance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The product distribution may choose bundled defaults. Generic content and
// executable composition must consume the resolved package set instead of
// silently mounting one named game mod.
func TestGenericContentStartupDoesNotInjectD2Legacy(t *testing.T) {
	root := repositoryRoot(t)
	for _, relative := range []string{"internal/content/fs.go", "cmd/client/main.go", "cmd/server/main.go"} {
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(string(data), "content.D2Legacy()") || strings.Contains(string(data), "FS: D2Legacy()") {
			t.Errorf("%s injects d2legacy instead of consuming the resolved mod set", relative)
		}
	}

	contentSource, err := os.ReadFile(filepath.Join(root, "internal/content/fs.go"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(contentSource), "DARK_MAGIC_MOD_DIRECTORY") {
		t.Fatal("generic content startup still permits an unverified directory to bypass the resolved mod lock")
	}

	distribution, err := os.ReadFile(filepath.Join(root, "internal/distribution/mods.go"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(distribution), "content.D2Legacy()") {
		t.Fatal("product distribution no longer reconciles the bundled d2legacy package")
	}
}
