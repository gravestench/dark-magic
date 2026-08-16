package session

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func compatibilityIdentity(packageHash string) simulation.RuntimeIdentity {
	digestBytes := sha256.Sum256([]byte(packageHash))
	packageHash = "sha256:" + hex.EncodeToString(digestBytes[:])
	return simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID),
		Packages:             simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0", Digest: packageHash, Size: 1, Redistributable: true}},
		AuthoritativeHash:    packageHash, ConfigurationHash: "config",
		CapabilityVersions: map[string]string{"engine.ecs": "v1"},
	}}
}

func identityState(t *testing.T, identity simulation.RuntimeIdentity) []simulation.ParticipantState {
	t.Helper()
	participant, err := simulation.NewIdentityParticipant(identity)
	if err != nil {
		t.Fatal(err)
	}
	data, err := participant.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	return []simulation.ParticipantState{{ID: participant.StateID(), Data: data}}
}

func TestHeadlessRealmAndGameServerPinSessionCompatibility(t *testing.T) {
	server := compatibilityIdentity("package-a")
	allocation, err := Allocate("realm-session-7", server, PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	token, err := allocation.Admit("character-alice", server)
	if err != nil {
		t.Fatal(err)
	}
	if token.Prediction != PredictionLimited {
		t.Fatalf("prediction tier = %q", token.Prediction)
	}
	if err := allocation.ValidateReconnect(token, server); err != nil {
		t.Fatal(err)
	}
	if err := allocation.ValidateDurable(allocation.Durable("character-alice")); err != nil {
		t.Fatal(err)
	}

	mismatch := compatibilityIdentity("package-b")
	if _, err := allocation.Admit("character-bob", mismatch); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched client error = %v", err)
	}
	if err := allocation.ValidateReconnect(token, mismatch); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched reconnect error = %v", err)
	}
	badDurable := allocation.Durable("character-alice")
	badDurable.IdentityHash = "different"
	if err := allocation.ValidateDurable(badDurable); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched durable error = %v", err)
	}

	checkpoint := simulation.Checkpoint{Tick: 1, Participants: identityState(t, server)}
	replay := simulation.Replay{Version: simulation.ReplayVersion, InitialParticipants: identityState(t, server),
		Checkpoints: []simulation.Checkpoint{checkpoint}}
	if err := allocation.ValidateCheckpoint(checkpoint, nil); err != nil {
		t.Fatal(err)
	}
	if err := allocation.ValidateReplay(replay, nil); err != nil {
		t.Fatal(err)
	}
	badCheckpoint := checkpoint
	badCheckpoint.Participants = identityState(t, mismatch)
	if err := allocation.ValidateCheckpoint(badCheckpoint, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched checkpoint error = %v", err)
	}
	badReplay := replay
	badReplay.InitialParticipants = identityState(t, mismatch)
	if err := allocation.ValidateReplay(badReplay, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched replay error = %v", err)
	}
}

func TestChangedPackageAppliesOnlyToNewAllocationWithoutExplicitMigration(t *testing.T) {
	oldIdentity, newIdentity := compatibilityIdentity("package-old"), compatibilityIdentity("package-new")
	oldAllocation, _ := Allocate("old-session", oldIdentity, PredictionNone)
	newAllocation, _ := Allocate("new-session", newIdentity, PredictionSharedRules)
	oldCheckpoint := simulation.Checkpoint{Tick: 4, Participants: identityState(t, oldIdentity)}
	if err := newAllocation.ValidateCheckpoint(oldCheckpoint, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("implicit migration error = %v", err)
	}
	oldHash, _ := oldIdentity.Digest()
	newHash, _ := newIdentity.Digest()
	migration := &StateMigration{FromIdentityHash: oldHash, ToIdentityHash: newHash}
	if err := newAllocation.ValidateCheckpoint(oldCheckpoint, migration); err != nil {
		t.Fatal(err)
	}
	if err := oldAllocation.ValidateCheckpoint(oldCheckpoint, nil); err != nil {
		t.Fatal(err)
	}
}

func TestOrderedExtensionRecipeIsPinnedAcrossEveryCompatibilitySurface(t *testing.T) {
	server := compatibilityIdentity("base")
	server.Recipe.Packages.Extensions = []simulation.RuntimePackage{
		{ID: "foundation", Version: "1", Digest: runtimePackageDigest("foundation"), Size: 10, Redistributable: true},
		{ID: "feature", Version: "1", Digest: runtimePackageDigest("feature"), Size: 20, Redistributable: true},
	}
	allocation, err := Allocate("extensions", server, PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	token, err := allocation.Admit("character", server)
	if err != nil {
		t.Fatal(err)
	}
	mismatch := server
	mismatch.Recipe.Packages.Extensions = append([]simulation.RuntimePackage(nil), server.Recipe.Packages.Extensions...)
	mismatch.Recipe.Packages.Extensions[0], mismatch.Recipe.Packages.Extensions[1] = mismatch.Recipe.Packages.Extensions[1], mismatch.Recipe.Packages.Extensions[0]
	if _, err := allocation.Admit("character", mismatch); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("extension-order admission error = %v", err)
	}
	if err := allocation.ValidateReconnect(token, mismatch); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("extension-order reconnect error = %v", err)
	}
	checkpoint := simulation.Checkpoint{Participants: identityState(t, mismatch)}
	if err := allocation.ValidateCheckpoint(checkpoint, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("extension-order checkpoint error = %v", err)
	}
	replay := simulation.Replay{Version: simulation.ReplayVersion, InitialParticipants: identityState(t, mismatch)}
	if err := allocation.ValidateReplay(replay, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("extension-order replay error = %v", err)
	}
	durable := allocation.Durable("character")
	differentAllocation, _ := Allocate("different", mismatch, PredictionLimited)
	if err := differentAllocation.ValidateDurable(durable); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("extension-order durable error = %v", err)
	}
}

func runtimePackageDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}
