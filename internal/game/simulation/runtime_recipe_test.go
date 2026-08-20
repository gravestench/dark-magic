package simulation

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestRuntimeRecipeDigestPinsExtensionOrderAndCompleteMetadata keeps every runtime input in session identity.
func TestRuntimeRecipeDigestPinsExtensionOrderAndCompleteMetadata(t *testing.T) {
	const (
		baseDigest   = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		firstDigest  = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		secondDigest = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	)

	recipe := RuntimeRecipe{
		Schema: RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: EmptyAssetSetID,
		GameDataGenerationID: GameDataGenerationIDForAssetSet(EmptyAssetSetID),
		Packages: RuntimePackageSet{
			Base: RuntimePackage{ID: "d2legacy", Version: "1", Digest: baseDigest, Size: 1, Redistributable: true},
			Extensions: []RuntimePackage{
				{ID: "first", Version: "1", Digest: firstDigest, Size: 2, Redistributable: true},
				{ID: "second", Version: "1", Digest: secondDigest, Size: 3, Redistributable: true},
			},
		},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}

	first, err := (RuntimeIdentity{Recipe: recipe}).Digest()
	if err != nil {
		t.Fatal(err)
	}

	recipe.Packages.Extensions[0], recipe.Packages.Extensions[1] =
		recipe.Packages.Extensions[1], recipe.Packages.Extensions[0]

	second, err := (RuntimeIdentity{Recipe: recipe}).Digest()
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Fatal("extension activation order did not change runtime identity")
	}

	recipe.Packages.Extensions[0].Redistributable = false

	third, err := (RuntimeIdentity{Recipe: recipe}).Digest()
	if err != nil {
		t.Fatal(err)
	}

	if second == third {
		t.Fatal("package metadata did not change runtime identity")
	}

	recipe.AssetSetID = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"

	fourth, err := (RuntimeIdentity{Recipe: recipe}).Digest()
	if err != nil {
		t.Fatal(err)
	}

	if third == fourth {
		t.Fatal("external asset set did not change runtime identity")
	}
}

// TestRuntimeRecipeRequiresCanonicalAssetSetIdentity prevents ambiguous external-data provenance.
func TestRuntimeRecipeRequiresCanonicalAssetSetIdentity(t *testing.T) {
	recipe := RuntimeRecipe{
		Schema: RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1",
		Packages: RuntimePackageSet{Base: RuntimePackage{
			ID: "d2legacy", Version: "1",
			Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Size: 1,
		}},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}
	if err := recipe.Validate(); err == nil {
		t.Fatal("runtime recipe without an asset-set identity was accepted")
	}
}

// TestGameDataGenerationPinsAssetBytesAndParserSchema ensures either input invalidates deterministic sessions.
func TestGameDataGenerationPinsAssetBytesAndParserSchema(t *testing.T) {
	first := GameDataGenerationIDForAssetSet(EmptyAssetSetID)
	if err := ValidateGameDataGenerationID(first); err != nil {
		t.Fatal(err)
	}

	other := GameDataGenerationIDForAssetSet("sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
	if first == other {
		t.Fatal("different mounted bytes produced the same game-data generation")
	}

	payload := GameDataGenerationSchema + "\x00" + EmptyAssetSetID + "\x00different-parser/v1"

	digest := sha256.Sum256([]byte(payload))
	if first == "sha256:"+hex.EncodeToString(digest[:]) {
		t.Fatal("different parser schema produced the same game-data generation")
	}
}

// TestRuntimeRecipeRequiresGameDataGenerationIdentity rejects recipes that omit parsed-record provenance.
func TestRuntimeRecipeRequiresGameDataGenerationIdentity(t *testing.T) {
	recipe := runtimeIdentityFixture("generation-required").Recipe

	recipe.GameDataGenerationID = ""
	if err := recipe.Validate(); err == nil {
		t.Fatal("runtime recipe without a game-data generation identity was accepted")
	}
}

// TestRuntimeRecipeRejectsNonCanonicalPackageDigest keeps package identity portable and unambiguous.
func TestRuntimeRecipeRejectsNonCanonicalPackageDigest(t *testing.T) {
	recipe := RuntimeRecipe{
		Schema: RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: EmptyAssetSetID,
		GameDataGenerationID: GameDataGenerationIDForAssetSet(EmptyAssetSetID),
		Packages: RuntimePackageSet{Base: RuntimePackage{
			ID: "d2legacy", Version: "1", Digest: "not-a-digest", Size: 1,
		}},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}
	if err := recipe.Validate(); err == nil {
		t.Fatal("non-canonical package digest was accepted")
	}
}

// TestRuntimeRecipeRejectsOverlappingPackageNamespaces prevents ownership ambiguity between runtime packages.
func TestRuntimeRecipeRejectsOverlappingPackageNamespaces(t *testing.T) {
	const (
		baseDigest      = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		extensionDigest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)

	recipe := RuntimeRecipe{
		Schema: RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: EmptyAssetSetID,
		GameDataGenerationID: GameDataGenerationIDForAssetSet(EmptyAssetSetID),
		Packages: RuntimePackageSet{
			Base:       RuntimePackage{ID: "d2legacy", Version: "1", Digest: baseDigest, Size: 1},
			Extensions: []RuntimePackage{{ID: "d2legacy.feature", Version: "1", Digest: extensionDigest, Size: 1}},
		},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}
	if err := recipe.Validate(); err == nil {
		t.Fatal("overlapping runtime package namespaces were accepted")
	}
}
