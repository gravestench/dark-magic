package d2legacy

import (
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

func TestIdentityIncludesCanonicalGameplayConfiguration(t *testing.T) {
	left, err := Identity(content.D2Legacy(), map[string]any{"difficulty": "normal", "seed": float64(7)})
	if err != nil {
		t.Fatal(err)
	}
	same, err := Identity(content.D2Legacy(), map[string]any{"seed": float64(7), "difficulty": "normal"})
	if err != nil {
		t.Fatal(err)
	}
	different, err := Identity(content.D2Legacy(), map[string]any{"difficulty": "nightmare", "seed": float64(7)})
	if err != nil {
		t.Fatal(err)
	}
	if left.Recipe.ConfigurationHash != same.Recipe.ConfigurationHash {
		t.Fatal("map insertion order changed configuration identity")
	}
	if left.Recipe.ConfigurationHash == different.Recipe.ConfigurationHash {
		t.Fatal("changed gameplay configuration retained identity")
	}
}

func TestIdentitySeparatesPackageAndAuthoritativeLuaHashes(t *testing.T) {
	manifest, _ := json.Marshal(modcache.Manifest{Schema: modcache.ManifestSchema, ID: "d2legacy", Name: "test",
		Version: "1.0.0", Kind: "game", EngineAPI: modcache.EngineAPI, Redistributable: true})
	base := fstest.MapFS{
		"mod.json":              {Data: manifest},
		"lua/d2legacy/game.lua": {Data: []byte("return 1")},
		"presentation.txt":      {Data: []byte("first")},
	}
	first, err := Identity(base)
	if err != nil {
		t.Fatal(err)
	}
	base["presentation.txt"] = &fstest.MapFile{Data: []byte("second")}
	presentationChanged, err := Identity(base)
	if err != nil {
		t.Fatal(err)
	}
	if first.Recipe.Packages.Base.Digest == presentationChanged.Recipe.Packages.Base.Digest {
		t.Fatal("package hash ignored non-authoritative bytes")
	}
	if first.Recipe.AuthoritativeHash != presentationChanged.Recipe.AuthoritativeHash {
		t.Fatal("presentation bytes changed authoritative Lua hash")
	}
	base["lua/d2legacy/game.lua"] = &fstest.MapFile{Data: []byte("return 2")}
	scriptChanged, err := Identity(base)
	if err != nil {
		t.Fatal(err)
	}
	if presentationChanged.Recipe.AuthoritativeHash == scriptChanged.Recipe.AuthoritativeHash {
		t.Fatal("changed Lua retained authoritative hash")
	}
	if len(scriptChanged.Recipe.CapabilityVersions) == 0 {
		t.Fatal("identity omitted capability contract")
	}
}

func TestIdentityForPackagesRejectsBuiltinMetadataDrift(t *testing.T) {
	source := content.D2Legacy()
	builtin, err := modcache.DescribeBuiltin(source)
	if err != nil {
		t.Fatal(err)
	}
	packages := simulation.RuntimePackageSet{Base: simulation.RuntimePackage{
		ID: builtin.Manifest.ID, Version: builtin.Manifest.Version, Digest: builtin.Descriptor.Digest,
		Size: builtin.Descriptor.Size + 1, Redistributable: builtin.Descriptor.Redistributable,
	}}
	if _, err := IdentityForPackages(source, packages); err == nil {
		t.Fatal("built-in package metadata drift was accepted")
	}
}
