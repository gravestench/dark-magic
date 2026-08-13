package simulation

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func replayContainerFixture() ReplayContainer {
	container := NewReplayContainer(Replay{
		Version:   ReplayVersion,
		StepNanos: int64(40 * time.Millisecond),
		Commands:  []Command{},
	})
	container.Manifests = map[string]ReplayManifest{
		"session": {Schema: "dark-magic.session/v1", Data: json.RawMessage(`{"mod":"d2legacy"}`)},
	}
	container.Events = []ReplayEvent{{Tick: 1, Kind: "started", Payload: json.RawMessage(`{}`)}}
	return container
}

func TestReplayContainerRoundTripsVersionedEvidence(t *testing.T) {
	var encoded bytes.Buffer
	want := replayContainerFixture()
	if err := EncodeReplayContainer(&encoded, want); err != nil {
		t.Fatal(err)
	}
	got, err := DecodeReplayContainer(&encoded, ReplayContainerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Format != ReplayContainerFormat || got.Version != ReplayContainerVersion ||
		got.Replay.Version != ReplayVersion || len(got.Manifests) != 1 || len(got.Events) != 1 ||
		!strings.HasPrefix(got.Integrity, "sha256:") {
		t.Fatalf("decoded container = %#v", got)
	}
}

func TestReplayContainerRejectsOversizedAndOverpopulatedInput(t *testing.T) {
	encoded, err := json.Marshal(replayContainerFixture())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeReplayContainer(bytes.NewReader(encoded), ReplayContainerLimits{
		MaxBytes: int64(len(encoded) - 1), MaxManifests: 1, MaxCommands: 1,
		MaxCheckpoints: 1, MaxEvents: 1,
	}); !errors.Is(err, ErrReplayContainer) || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized input error = %v", err)
	}
	if _, err := DecodeReplayContainer(bytes.NewReader(encoded), ReplayContainerLimits{
		MaxBytes: int64(len(encoded)), MaxManifests: 1, MaxCommands: 1,
		MaxCheckpoints: 1, MaxEvents: -1,
	}); !errors.Is(err, ErrReplayContainer) {
		t.Fatalf("invalid limits error = %v", err)
	}
}

func TestReplayContainerRequiresExplicitVersionMigration(t *testing.T) {
	container := replayContainerFixture()
	container.Version = ReplayContainerVersion + 1
	encoded, err := json.Marshal(container)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeReplayContainer(bytes.NewReader(encoded), ReplayContainerLimits{}); !errors.Is(err, ErrReplayMigration) {
		t.Fatalf("version error = %v, want migration requirement", err)
	}
}

func TestReplayContainerAppliesExplicitSequentialMigration(t *testing.T) {
	container := replayContainerFixture()
	container.Version = 0
	container.Integrity, _ = replayContainerIntegrity(container)
	encoded, err := json.Marshal(container)
	if err != nil {
		t.Fatal(err)
	}
	migrations := map[uint32]ReplayContainerMigration{
		0: func(input json.RawMessage) (json.RawMessage, error) {
			var value map[string]any
			if err := json.Unmarshal(input, &value); err != nil {
				return nil, err
			}
			value["version"] = float64(1)
			value["integrity"] = ""
			withoutIntegrity, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			var migrated ReplayContainer
			if err := json.Unmarshal(withoutIntegrity, &migrated); err != nil {
				return nil, err
			}
			value["integrity"], err = replayContainerIntegrity(migrated)
			if err != nil {
				return nil, err
			}
			return json.Marshal(value)
		},
	}
	got, err := DecodeReplayContainerWithMigrations(bytes.NewReader(encoded), ReplayContainerLimits{}, migrations)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != ReplayContainerVersion {
		t.Fatalf("migrated version = %d", got.Version)
	}
}

func TestReplayContainerRejectsMigrationThatSkipsVersions(t *testing.T) {
	container := replayContainerFixture()
	container.Version = 0
	encoded, err := json.Marshal(container)
	if err != nil {
		t.Fatal(err)
	}
	migrations := map[uint32]ReplayContainerMigration{
		0: func(input json.RawMessage) (json.RawMessage, error) {
			return bytes.Replace(input, []byte(`"version":0`), []byte(`"version":2`), 1), nil
		},
	}
	if _, err := DecodeReplayContainerWithMigrations(bytes.NewReader(encoded), ReplayContainerLimits{}, migrations); !errors.Is(err, ErrReplayMigration) || !strings.Contains(err.Error(), "must produce version 1") {
		t.Fatalf("skipped migration error = %v", err)
	}
}

func TestReplayContainerRejectsTamperedEvidence(t *testing.T) {
	var encoded bytes.Buffer
	if err := EncodeReplayContainer(&encoded, replayContainerFixture()); err != nil {
		t.Fatal(err)
	}
	tampered := bytes.Replace(encoded.Bytes(), []byte(`"d2legacy"`), []byte(`"different"`), 1)
	if _, err := DecodeReplayContainer(bytes.NewReader(tampered), ReplayContainerLimits{}); !errors.Is(err, ErrReplayContainer) || !strings.Contains(err.Error(), "integrity mismatch") {
		t.Fatalf("tampered evidence error = %v", err)
	}
}

func TestReplayContainerRejectsUnknownOrTrailingInput(t *testing.T) {
	encoded, err := json.Marshal(replayContainerFixture())
	if err != nil {
		t.Fatal(err)
	}
	unknown := bytes.Replace(encoded, []byte(`"format"`), []byte(`"unknown":true,"format"`), 1)
	if _, err := DecodeReplayContainer(bytes.NewReader(unknown), ReplayContainerLimits{}); !errors.Is(err, ErrReplayContainer) {
		t.Fatalf("unknown field error = %v", err)
	}
	if _, err := DecodeReplayContainer(bytes.NewReader(append(encoded, []byte(`{}`)...)), ReplayContainerLimits{}); !errors.Is(err, ErrReplayContainer) || !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("trailing input error = %v", err)
	}
}
