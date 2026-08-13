package content_test

import (
	"context"
	"io/fs"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func TestModTemplateIsDiscoverableStarterContent(t *testing.T) {
	t.Parallel()

	source := content.ModTemplate()
	for _, path := range []string{
		"README.md",
		"boot.lua",
		"components/example.lua",
		"lua/mod_template.lua",
	} {
		if _, err := fs.Stat(source, path); err != nil {
			t.Fatalf("starter file %q: %v", path, err)
		}
	}

	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(source, "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	definitions, err := modruntime.DiscoverDefinitions(context.Background(), runtime, source)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, len(definitions))
	for index, definition := range definitions {
		ids[index] = definition.ID
	}
	if wanted := []string{"mod_template.boot", "mod_template.example"}; !reflect.DeepEqual(ids, wanted) {
		t.Fatalf("starter definitions = %v, want %v", ids, wanted)
	}
}

func TestD2LegacyArchitectureGuideLinksStarterAndTestDocumentation(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"ARCHITECTURE.md",
		"lua/README.md",
		"lua/d2legacy/README.md",
	} {
		data, err := fs.ReadFile(content.D2Legacy(), path)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			t.Fatalf("documentation %q is empty", path)
		}
	}
}
