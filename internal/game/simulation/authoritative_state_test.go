package simulation

import (
	"bytes"
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
	identity := RuntimeIdentity{ModID: "d2legacy", ContractVersion: "v1", PackageHash: "package-a", AuthoritativeHash: "rules-a", ConfigurationHash: "config-a", Dependencies: map[string]string{"base": "one"}, CapabilityVersions: map[string]string{"dm.ecs": "v1"}}
	participant, err := NewIdentityParticipant(identity)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := participant.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	other := identity
	other.AuthoritativeHash = "rules-b"
	different, err := NewIdentityParticipant(other)
	if err != nil {
		t.Fatal(err)
	}
	if err := different.RestoreState(snapshot); err == nil {
		t.Fatal("different authoritative code identity was accepted")
	}
}
