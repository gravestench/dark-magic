package session

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

var ErrCompatibility = errors.New("game session: incompatible runtime")

type PredictionTier string

const (
	PredictionNone        PredictionTier = "none"
	PredictionLimited     PredictionTier = "limited_movement_presentation"
	PredictionSharedRules PredictionTier = "shared_authoritative_rules"
)

type Allocation struct {
	SessionID    string                     `json:"session_id"`
	Identity     simulation.RuntimeIdentity `json:"identity"`
	IdentityHash string                     `json:"identity_hash"`
	Prediction   PredictionTier             `json:"prediction"`
}

type AdmissionToken struct {
	SessionID    string         `json:"session_id"`
	CharacterID  string         `json:"character_id"`
	IdentityHash string         `json:"identity_hash"`
	Prediction   PredictionTier `json:"prediction"`
}

type DurableCompatibility struct {
	CharacterID     string `json:"character_id"`
	ModID           string `json:"mod_id"`
	ContractVersion string `json:"contract_version"`
	IdentityHash    string `json:"identity_hash"`
}

type StateMigration struct{ FromIdentityHash, ToIdentityHash string }

func Allocate(sessionID string, identity simulation.RuntimeIdentity, prediction PredictionTier) (Allocation, error) {
	if strings.TrimSpace(sessionID) == "" {
		return Allocation{}, fmt.Errorf("%w: session ID is required", ErrCompatibility)
	}
	if err := ValidatePredictionTier(prediction); err != nil {
		return Allocation{}, err
	}
	hash, err := identity.Digest()
	if err != nil {
		return Allocation{}, err
	}
	return Allocation{SessionID: sessionID, Identity: identity, IdentityHash: hash, Prediction: prediction}, nil
}

// ValidatePredictionTier rejects client contracts the authority does not
// explicitly support. Hosts use it before allocating expensive runtime state.
func ValidatePredictionTier(prediction PredictionTier) error {
	if prediction != PredictionNone && prediction != PredictionLimited && prediction != PredictionSharedRules {
		return fmt.Errorf("%w: unknown prediction tier", ErrCompatibility)
	}
	return nil
}

func (allocation Allocation) Admit(characterID string, client simulation.RuntimeIdentity) (AdmissionToken, error) {
	if strings.TrimSpace(characterID) == "" {
		return AdmissionToken{}, fmt.Errorf("%w: character ID is required", ErrCompatibility)
	}
	hash, err := client.Digest()
	if err != nil {
		return AdmissionToken{}, err
	}
	if hash != allocation.IdentityHash {
		return AdmissionToken{}, fmt.Errorf("%w: client identity differs", ErrCompatibility)
	}
	return AdmissionToken{SessionID: allocation.SessionID, CharacterID: characterID, IdentityHash: hash, Prediction: allocation.Prediction}, nil
}

func (allocation Allocation) ValidateReconnect(token AdmissionToken, client simulation.RuntimeIdentity) error {
	hash, err := client.Digest()
	if err != nil {
		return err
	}
	if token.SessionID != allocation.SessionID || token.IdentityHash != allocation.IdentityHash || hash != allocation.IdentityHash || token.CharacterID == "" {
		return fmt.Errorf("%w: reconnect identity differs", ErrCompatibility)
	}
	return nil
}

func (allocation Allocation) Durable(characterID string) DurableCompatibility {
	return DurableCompatibility{CharacterID: characterID, ModID: allocation.Identity.ModID,
		ContractVersion: allocation.Identity.ContractVersion, IdentityHash: allocation.IdentityHash}
}

func (allocation Allocation) ValidateDurable(value DurableCompatibility) error {
	if value.ModID != allocation.Identity.ModID || value.ContractVersion != allocation.Identity.ContractVersion || value.IdentityHash != allocation.IdentityHash {
		return fmt.Errorf("%w: durable character identity differs", ErrCompatibility)
	}
	return nil
}

func (allocation Allocation) ValidateCheckpoint(checkpoint simulation.Checkpoint, migration *StateMigration) error {
	identity, err := simulation.RuntimeIdentityFromParticipants(checkpoint.Participants)
	if err != nil {
		return err
	}
	return allocation.validateRestored(identity, migration)
}

func (allocation Allocation) ValidateReplay(replay simulation.Replay, migration *StateMigration) error {
	identity, err := simulation.RuntimeIdentityFromParticipants(replay.InitialParticipants)
	if err != nil {
		return err
	}
	if err := allocation.validateRestored(identity, migration); err != nil {
		return err
	}
	for _, checkpoint := range replay.Checkpoints {
		checkpointIdentity, err := simulation.RuntimeIdentityFromParticipants(checkpoint.Participants)
		if err != nil {
			return err
		}
		if digest, _ := checkpointIdentity.Digest(); digest != allocation.IdentityHash {
			return fmt.Errorf("%w: replay checkpoint identity differs", ErrCompatibility)
		}
	}
	return nil
}

func (allocation Allocation) validateRestored(identity simulation.RuntimeIdentity, migration *StateMigration) error {
	hash, err := identity.Digest()
	if err != nil {
		return err
	}
	if hash == allocation.IdentityHash {
		return nil
	}
	if migration != nil && migration.FromIdentityHash == hash && migration.ToIdentityHash == allocation.IdentityHash {
		return nil
	}
	return fmt.Errorf("%w: restored identity differs", ErrCompatibility)
}
