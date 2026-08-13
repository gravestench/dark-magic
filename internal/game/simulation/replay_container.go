package simulation

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	ReplayContainerFormat  = "dark-magic-replay"
	ReplayContainerVersion = 1
)

var (
	ErrReplayContainer = errors.New("simulation: invalid replay container")
	ErrReplayMigration = errors.New("simulation: replay container migration required")
)

// ReplayContainer is the versioned persistence envelope around deterministic
// replay state. Manifests pin external session facts; events retain optional
// semantic evidence without changing replay verification authority.
type ReplayContainer struct {
	Format    string                    `json:"format"`
	Version   uint32                    `json:"version"`
	Manifests map[string]ReplayManifest `json:"manifests,omitempty"`
	Replay    Replay                    `json:"replay"`
	Events    []ReplayEvent             `json:"events,omitempty"`
	Integrity string                    `json:"integrity"`
}

type ReplayManifest struct {
	Schema string          `json:"schema"`
	Data   json.RawMessage `json:"data"`
}

type ReplayEvent struct {
	Tick    uint64          `json:"tick"`
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// ReplayContainerLimits bounds untrusted on-disk input before callers admit it
// to a session. Zero values receive conservative defaults.
type ReplayContainerLimits struct {
	MaxBytes       int64
	MaxManifests   int
	MaxCommands    int
	MaxCheckpoints int
	MaxEvents      int
}

// ReplayContainerMigration transforms exactly one older envelope version into
// the next version. The result is decoded and validated again; a migration can
// never bypass current bounds or schema checks.
type ReplayContainerMigration func(json.RawMessage) (json.RawMessage, error)

type replayContainerHeader struct {
	Format  string `json:"format"`
	Version uint32 `json:"version"`
}

func (limits ReplayContainerLimits) withDefaults() ReplayContainerLimits {
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 64 << 20
	}
	if limits.MaxManifests == 0 {
		limits.MaxManifests = 64
	}
	if limits.MaxCommands == 0 {
		limits.MaxCommands = 1_000_000
	}
	if limits.MaxCheckpoints == 0 {
		limits.MaxCheckpoints = 100_000
	}
	if limits.MaxEvents == 0 {
		limits.MaxEvents = 1_000_000
	}
	return limits
}

func NewReplayContainer(replay Replay) ReplayContainer {
	return ReplayContainer{
		Format:  ReplayContainerFormat,
		Version: ReplayContainerVersion,
		Replay:  replay,
	}
}

func EncodeReplayContainer(destination io.Writer, container ReplayContainer) error {
	if destination == nil {
		return fmt.Errorf("%w: destination is required", ErrReplayContainer)
	}
	limits := ReplayContainerLimits{}.withDefaults()
	if err := validateReplayContainer(container, limits); err != nil {
		return err
	}
	integrity, err := replayContainerIntegrity(container)
	if err != nil {
		return err
	}
	container.Integrity = integrity
	encoded, err := json.Marshal(container)
	if err != nil {
		return fmt.Errorf("%w: encode: %v", ErrReplayContainer, err)
	}
	if int64(len(encoded)+1) > limits.MaxBytes {
		return fmt.Errorf("%w: encoded input exceeds %d bytes", ErrReplayContainer, limits.MaxBytes)
	}
	if _, err := destination.Write(append(encoded, '\n')); err != nil {
		return fmt.Errorf("%w: write: %v", ErrReplayContainer, err)
	}
	return nil
}

func DecodeReplayContainer(source io.Reader, limits ReplayContainerLimits) (ReplayContainer, error) {
	return DecodeReplayContainerWithMigrations(source, limits, nil)
}

func DecodeReplayContainerWithMigrations(source io.Reader, limits ReplayContainerLimits,
	migrations map[uint32]ReplayContainerMigration,
) (ReplayContainer, error) {
	if source == nil {
		return ReplayContainer{}, fmt.Errorf("%w: source is required", ErrReplayContainer)
	}
	limits = limits.withDefaults()
	if limits.MaxBytes <= 0 || limits.MaxManifests < 0 || limits.MaxCommands < 0 ||
		limits.MaxCheckpoints < 0 || limits.MaxEvents < 0 {
		return ReplayContainer{}, fmt.Errorf("%w: limits must be positive", ErrReplayContainer)
	}

	data, err := io.ReadAll(io.LimitReader(source, limits.MaxBytes+1))
	if err != nil {
		return ReplayContainer{}, fmt.Errorf("%w: read: %v", ErrReplayContainer, err)
	}
	if int64(len(data)) > limits.MaxBytes {
		return ReplayContainer{}, fmt.Errorf("%w: input exceeds %d bytes", ErrReplayContainer, limits.MaxBytes)
	}
	data, err = migrateReplayContainer(data, limits, migrations)
	if err != nil {
		return ReplayContainer{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var container ReplayContainer
	if err := decoder.Decode(&container); err != nil {
		return ReplayContainer{}, fmt.Errorf("%w: decode: %v", ErrReplayContainer, err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return ReplayContainer{}, fmt.Errorf("%w: trailing data", ErrReplayContainer)
	}
	if err := validateReplayContainer(container, limits); err != nil {
		return ReplayContainer{}, err
	}
	want, err := replayContainerIntegrity(container)
	if err != nil {
		return ReplayContainer{}, err
	}
	if container.Integrity == "" || container.Integrity != want {
		return ReplayContainer{}, fmt.Errorf("%w: integrity mismatch", ErrReplayContainer)
	}
	return container, nil
}

func replayContainerIntegrity(container ReplayContainer) (string, error) {
	container.Integrity = ""
	encoded, err := json.Marshal(container)
	if err != nil {
		return "", fmt.Errorf("%w: integrity encoding: %v", ErrReplayContainer, err)
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)), nil
}

func migrateReplayContainer(data []byte, limits ReplayContainerLimits,
	migrations map[uint32]ReplayContainerMigration,
) ([]byte, error) {
	var header replayContainerHeader
	headerDecoder := json.NewDecoder(bytes.NewReader(data))
	if err := headerDecoder.Decode(&header); err != nil {
		return nil, fmt.Errorf("%w: header: %v", ErrReplayContainer, err)
	}
	if headerDecoder.Decode(&struct{}{}) != io.EOF {
		return nil, fmt.Errorf("%w: trailing data", ErrReplayContainer)
	}
	if header.Format != ReplayContainerFormat {
		return nil, fmt.Errorf("%w: format %q", ErrReplayContainer, header.Format)
	}
	if header.Version > ReplayContainerVersion {
		return nil, fmt.Errorf("%w: future version %d", ErrReplayMigration, header.Version)
	}
	for header.Version < ReplayContainerVersion {
		migration := migrations[header.Version]
		if migration == nil {
			return nil, fmt.Errorf("%w: no migration from version %d", ErrReplayMigration, header.Version)
		}
		migrated, err := migration(append(json.RawMessage(nil), data...))
		if err != nil {
			return nil, fmt.Errorf("%w: version %d: %v", ErrReplayMigration, header.Version, err)
		}
		if int64(len(migrated)) > limits.MaxBytes {
			return nil, fmt.Errorf("%w: migrated input exceeds %d bytes", ErrReplayContainer, limits.MaxBytes)
		}
		var next replayContainerHeader
		if err := json.Unmarshal(migrated, &next); err != nil {
			return nil, fmt.Errorf("%w: migrated header: %v", ErrReplayContainer, err)
		}
		if next.Format != ReplayContainerFormat || next.Version != header.Version+1 {
			return nil, fmt.Errorf("%w: migration from version %d must produce version %d",
				ErrReplayMigration, header.Version, header.Version+1)
		}
		data, header = append([]byte(nil), migrated...), next
	}
	return data, nil
}

func validateReplayContainer(container ReplayContainer, limits ReplayContainerLimits) error {
	if container.Format != ReplayContainerFormat {
		return fmt.Errorf("%w: format %q", ErrReplayContainer, container.Format)
	}
	if container.Version != ReplayContainerVersion {
		return fmt.Errorf("%w: version %d to %d", ErrReplayMigration,
			container.Version, ReplayContainerVersion)
	}
	if container.Replay.Version != ReplayVersion || container.Replay.StepNanos <= 0 {
		return fmt.Errorf("%w: embedded replay header", ErrReplayContainer)
	}
	if len(container.Manifests) > limits.MaxManifests {
		return fmt.Errorf("%w: %d manifests exceed %d", ErrReplayContainer,
			len(container.Manifests), limits.MaxManifests)
	}
	if len(container.Replay.Commands) > limits.MaxCommands {
		return fmt.Errorf("%w: %d commands exceed %d", ErrReplayContainer,
			len(container.Replay.Commands), limits.MaxCommands)
	}
	if len(container.Replay.Checkpoints) > limits.MaxCheckpoints {
		return fmt.Errorf("%w: %d checkpoints exceed %d", ErrReplayContainer,
			len(container.Replay.Checkpoints), limits.MaxCheckpoints)
	}
	if len(container.Events) > limits.MaxEvents {
		return fmt.Errorf("%w: %d events exceed %d", ErrReplayContainer,
			len(container.Events), limits.MaxEvents)
	}
	for name, manifest := range container.Manifests {
		if strings.TrimSpace(name) == "" || strings.TrimSpace(manifest.Schema) == "" || !json.Valid(manifest.Data) {
			return fmt.Errorf("%w: manifest %q", ErrReplayContainer, name)
		}
	}
	for index, event := range container.Events {
		if strings.TrimSpace(event.Kind) == "" || !json.Valid(event.Payload) {
			return fmt.Errorf("%w: event %d", ErrReplayContainer, index)
		}
		if index > 0 && event.Tick < container.Events[index-1].Tick {
			return fmt.Errorf("%w: event ticks are not ordered", ErrReplayContainer)
		}
	}
	return nil
}
