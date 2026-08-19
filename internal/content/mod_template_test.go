package content_test

import (
	"context"
	"io/fs"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/modcache"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestModTemplateIsDiscoverableStarterContent proves the bundled starter is complete and executable by mod discovery.
func TestModTemplateIsDiscoverableStarterContent(t *testing.T) {
	t.Parallel()

	source := content.ModTemplate()
	for _, path := range []string{
		"README.md",
		"mod.json",
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
	defer func() {
		_ = runtime.Stop(context.Background())
	}()

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

// TestBundledModsDeclareValidPackageManifests pins both embedded trees to the engine's current package contract.
func TestBundledModsDeclareValidPackageManifests(t *testing.T) {
	for name, source := range map[string]fs.FS{"d2legacy": content.D2Legacy(), "mod_template": content.ModTemplate()} {
		t.Run(name, func(t *testing.T) {
			manifest, err := modcache.ReadManifest(source)
			if err != nil {
				t.Fatal(err)
			}

			if manifest.ID != name || manifest.EngineAPI != modcache.EngineAPI {
				t.Fatalf("manifest = %#v", manifest)
			}
		})
	}
}

// TestD2LegacyArchitectureGuideLinksStarterAndTestDocumentation keeps embedded contributor guidance distributable.
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
