package session

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// ErrCompatibility identifies any runtime-identity mismatch that makes persisted or client state unsafe to use.
var ErrCompatibility = errors.New("game session: incompatible runtime")

// PredictionTier describes how much local prediction the authoritative runtime contract permits.
type PredictionTier string

const (
	PredictionNone        PredictionTier = "none"
	PredictionLimited     PredictionTier = "limited_movement_presentation"
	PredictionSharedRules PredictionTier = "shared_authoritative_rules"
)

// Allocation pins one session to an exact runtime identity and client prediction policy.
type Allocation struct {
	SessionID    string                     `json:"session_id"`
	Identity     simulation.RuntimeIdentity `json:"identity"`
	IdentityHash string                     `json:"identity_hash"`
	Prediction   PredictionTier             `json:"prediction"`
}

// AdmissionToken records the compatibility evidence a reconnecting character must present.
type AdmissionToken struct {
	SessionID    string         `json:"session_id"`
	CharacterID  string         `json:"character_id"`
	IdentityHash string         `json:"identity_hash"`
	Prediction   PredictionTier `json:"prediction"`
}

// DurableCompatibility stores the runtime identity fields required to safely reopen a character.
type DurableCompatibility struct {
	CharacterID     string `json:"character_id"`
	ModID           string `json:"mod_id"`
	ContractVersion string `json:"contract_version"`
	IdentityHash    string `json:"identity_hash"`
}

// StateMigration explicitly authorizes one persisted identity to move to the allocation's current identity.
type StateMigration struct{ FromIdentityHash, ToIdentityHash string }

// Allocate validates and pins an immutable runtime digest before any expensive session state is created.
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

	return Allocation{
		SessionID:    sessionID,
		Identity:     identity,
		IdentityHash: hash,
		Prediction:   prediction,
	}, nil
}

// ValidatePredictionTier rejects client contracts the authority does not
// explicitly support. Hosts use it before allocating expensive runtime state.
func ValidatePredictionTier(prediction PredictionTier) error {
	if prediction != PredictionNone && prediction != PredictionLimited && prediction != PredictionSharedRules {
		return fmt.Errorf("%w: unknown prediction tier", ErrCompatibility)
	}

	return nil
}

// Admit binds a character to the allocation only when the client's complete runtime identity matches the authority.
func (allocation Allocation) Admit(
	characterID string,
	client simulation.RuntimeIdentity,
) (AdmissionToken, error) {
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

	return AdmissionToken{
		SessionID:    allocation.SessionID,
		CharacterID:  characterID,
		IdentityHash: hash,
		Prediction:   allocation.Prediction,
	}, nil
}

// ValidateReconnect prevents a valid token from being replayed into another session or under another runtime recipe.
func (allocation Allocation) ValidateReconnect(token AdmissionToken, client simulation.RuntimeIdentity) error {
	hash, err := client.Digest()
	if err != nil {
		return err
	}

	sessionDiffers := token.SessionID != allocation.SessionID
	tokenIdentityDiffers := token.IdentityHash != allocation.IdentityHash
	clientIdentityDiffers := hash != allocation.IdentityHash

	characterMissing := token.CharacterID == ""
	if sessionDiffers || tokenIdentityDiffers || clientIdentityDiffers || characterMissing {
		return fmt.Errorf("%w: reconnect identity differs", ErrCompatibility)
	}

	return nil
}

// Durable creates the compact identity evidence persisted beside character data for later compatibility checks.
func (allocation Allocation) Durable(characterID string) DurableCompatibility {
	return DurableCompatibility{
		CharacterID:     characterID,
		ModID:           allocation.Identity.Recipe.Packages.Base.ID,
		ContractVersion: allocation.Identity.Recipe.EngineAPI,
		IdentityHash:    allocation.IdentityHash,
	}
}

// ValidateDurable rejects character data produced by another base mod, engine contract, or composed package set.
func (allocation Allocation) ValidateDurable(value DurableCompatibility) error {
	modDiffers := value.ModID != allocation.Identity.Recipe.Packages.Base.ID
	contractDiffers := value.ContractVersion != allocation.Identity.Recipe.EngineAPI

	identityDiffers := value.IdentityHash != allocation.IdentityHash
	if modDiffers || contractDiffers || identityDiffers {
		return fmt.Errorf("%w: durable character identity differs", ErrCompatibility)
	}

	return nil
}

// ValidateCheckpoint verifies that a checkpoint belongs to this allocation or an explicitly approved migration.
func (allocation Allocation) ValidateCheckpoint(checkpoint simulation.Checkpoint, migration *StateMigration) error {
	identity, err := simulation.RuntimeIdentityFromParticipants(checkpoint.Participants)
	if err != nil {
		return err
	}

	return allocation.validateRestored(identity, migration)
}

// ValidateReplay checks both the replay origin and every checkpoint, preventing mixed-runtime replay evidence.
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

		// RuntimeIdentityFromParticipants has already validated the embedded identity before this comparison.
		digest, _ := checkpointIdentity.Digest()
		if digest != allocation.IdentityHash {
			return fmt.Errorf("%w: replay checkpoint identity differs", ErrCompatibility)
		}
	}

	return nil
}

// validateRestored centralizes exact-match and explicit-migration policy for every persisted-state entry point.
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
