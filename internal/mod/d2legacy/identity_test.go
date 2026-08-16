package d2legacy

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"reflect"
	"strings"
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

func TestIdentityPinsExplicitGameDataGeneration(t *testing.T) {
	baseline, err := Identity(content.D2Legacy())
	if err != nil {
		t.Fatal(err)
	}
	firstID := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	secondID := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	first, err := IdentityForPackagesAndData(content.D2Legacy(), baseline.Recipe.Packages,
		simulation.EmptyAssetSetID, firstID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := IdentityForPackagesAndData(content.D2Legacy(), baseline.Recipe.Packages,
		simulation.EmptyAssetSetID, secondID)
	if err != nil {
		t.Fatal(err)
	}
	firstDigest, _ := first.Digest()
	secondDigest, _ := second.Digest()
	if firstDigest == secondDigest || first.Recipe.GameDataGenerationID != firstID {
		t.Fatal("explicit game-data generation did not change the runtime identity")
	}
}

func TestAuthoritativeCapabilityIdentityRejectsCompositionDrift(t *testing.T) {
	identity, err := Identity(content.D2Legacy())
	if err != nil {
		t.Fatal(err)
	}
	modules := make([]string, 0, len(identity.Recipe.CapabilityVersions))
	for name, version := range identity.Recipe.CapabilityVersions {
		modules = append(modules, name+"/"+version)
	}
	if err := validateCapabilityIdentity(identity, modules); err != nil {
		t.Fatal(err)
	}
	if err := validateCapabilityIdentity(identity, modules[1:]); err == nil {
		t.Fatal("runtime missing an identity-pinned capability was accepted")
	}
	if err := validateCapabilityIdentity(identity, append(modules, "engine.unpinned/v1")); err == nil {
		t.Fatal("runtime with an unpinned authoritative capability was accepted")
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

func TestIdentityCanonicalizesCheckoutLineEndings(t *testing.T) {
	lf := content.D2Legacy()
	crlf := checkoutWithCRLF(t, lf)
	left, err := Identity(lf)
	if err != nil {
		t.Fatal(err)
	}
	right, err := Identity(crlf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("equivalent text checkouts have different identities:\nLF:   %#v\nCRLF: %#v", left, right)
	}
}

func checkoutWithCRLF(t *testing.T, source fs.FS) fstest.MapFS {
	t.Helper()
	checkout := fstest.MapFS{}
	err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		for _, suffix := range []string{".json", ".lua", ".md", ".txt"} {
			if strings.HasSuffix(strings.ToLower(name), suffix) {
				data = modcache.CanonicalBuiltinSource(name, data)
				data = bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
				break
			}
		}
		checkout[name] = &fstest.MapFile{Data: data, Mode: entry.Type()}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return checkout
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
	if _, err := IdentityForPackages(source, packages, simulation.EmptyAssetSetID); err == nil {
		t.Fatal("built-in package metadata drift was accepted")
	}
}
