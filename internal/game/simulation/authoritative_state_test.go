package simulation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestAuthoritativeStateStoreSnapshotsRestoresAndRejectsSchemaDrift(t *testing.T) {
	store := NewStateStore()
	if err := store.Register("d2legacy.quest", "quest-state/v1", []byte(`{"step":1}`)); err != nil {
		t.Fatal(err)
	}
	initial, err := store.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Replace("d2legacy.quest", "quest-state/v1", []byte(`{"step":2}`)); err != nil {
		t.Fatal(err)
	}
	if err := store.RestoreState(initial); err != nil {
		t.Fatal(err)
	}
	got, found := store.Read("d2legacy.quest")
	if !found || !bytes.Equal(got.Data, []byte(`{"step":1}`)) {
		t.Fatalf("restored state = %#v, %v", got, found)
	}
	if err := store.Replace("d2legacy.quest", "quest-state/v2", nil); err == nil {
		t.Fatal("schema drift was accepted")
	}
}

func TestRuntimeIdentityRejectsDifferentAuthoritativeCode(t *testing.T) {
	identity := runtimeIdentityFixture("package-a")
	participant, err := NewIdentityParticipant(identity)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := participant.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	other := identity
	other.Recipe.AuthoritativeHash = "rules-b"
	different, err := NewIdentityParticipant(other)
	if err != nil {
		t.Fatal(err)
	}
	if err := different.RestoreState(snapshot); err == nil {
		t.Fatal("different authoritative code identity was accepted")
	}
}

func runtimeIdentityFixture(packageDigest string) RuntimeIdentity {
	digest := sha256.Sum256([]byte(packageDigest))
	packageDigest = "sha256:" + hex.EncodeToString(digest[:])
	return RuntimeIdentity{Recipe: RuntimeRecipe{
		Schema: RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: EmptyAssetSetID,
		GameDataGenerationID: GameDataGenerationIDForAssetSet(EmptyAssetSetID),
		Packages:             RuntimePackageSet{Base: RuntimePackage{ID: "d2legacy", Version: "1.0.0", Digest: packageDigest, Size: 1, Redistributable: true}},
		AuthoritativeHash:    packageDigest, ConfigurationHash: "config",
		CapabilityVersions: map[string]string{"engine.ecs": "v1"},
	}}
}
