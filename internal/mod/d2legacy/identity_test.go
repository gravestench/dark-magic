package d2legacy

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
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
	if left.ConfigurationHash != same.ConfigurationHash {
		t.Fatal("map insertion order changed configuration identity")
	}
	if left.ConfigurationHash == different.ConfigurationHash {
		t.Fatal("changed gameplay configuration retained identity")
	}
}

func TestIdentitySeparatesPackageAndAuthoritativeLuaHashes(t *testing.T) {
	base := fstest.MapFS{
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
	if first.PackageHash == presentationChanged.PackageHash {
		t.Fatal("package hash ignored non-authoritative bytes")
	}
	if first.AuthoritativeHash != presentationChanged.AuthoritativeHash {
		t.Fatal("presentation bytes changed authoritative Lua hash")
	}
	base["lua/d2legacy/game.lua"] = &fstest.MapFile{Data: []byte("return 2")}
	scriptChanged, err := Identity(base)
	if err != nil {
		t.Fatal(err)
	}
	if presentationChanged.AuthoritativeHash == scriptChanged.AuthoritativeHash {
		t.Fatal("changed Lua retained authoritative hash")
	}
	if len(scriptChanged.Dependencies) == 0 || len(scriptChanged.CapabilityVersions) == 0 {
		t.Fatal("identity omitted dependency or capability graph")
	}
}
