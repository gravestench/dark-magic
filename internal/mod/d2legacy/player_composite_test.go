package d2legacy_test

import (
	"context"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	. "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestRealArchivesComposeUnarmedPlayerModes proves archive-backed composition
// selects stable unarmed animation modes across independent runtime scopes.
func TestRealArchivesComposeUnarmedPlayerModes(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the Diablo II MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}

	runtime := New()

	var composer render.Composer

	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}

	capability := NewRenderCapability(runtime, &composer, assets)
	if err := runtime.RegisterModule(capability.Module()); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	scope := &Scope{}
	if err := runtime.ExecuteScoped(
		context.Background(),
		scope,
		content.D2Legacy(),
		"lua/d2legacy/tests/integration/player_composite_real.lua",
	); err != nil {
		t.Fatal(err)
	}

	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}

	before := capability.Diagnostics().DecodeCalls

	secondScope := &Scope{}
	if err := runtime.ExecuteScoped(
		context.Background(),
		secondScope,
		content.D2Legacy(),
		"lua/d2legacy/tests/integration/player_composite_real.lua",
	); err != nil {
		t.Fatal(err)
	}

	if err := secondScope.Close(); err != nil {
		t.Fatal(err)
	}

	if after := capability.Diagnostics().DecodeCalls; after != before {
		t.Fatalf("warm composite recipes decoded again: before=%d after=%d", before, after)
	}
}
