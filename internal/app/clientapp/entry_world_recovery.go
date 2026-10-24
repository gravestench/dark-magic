package clientapp

import (
	"fmt"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
)

// entryWorldRecovery keeps the admitted topology with the gameplay checkpoint so a replacement authority can
// validate its maps and runtime identity before restoring any mutable session state.
type entryWorldRecovery struct {
	Topology         []byte
	TopologyChecksum string
	Session          gamesession.RecoveryCheckpoint
}

// entryWorldRecoveryCheckpoint captures the exact policy topology that produced the authoritative session identity.
func (app *application) entryWorldRecoveryCheckpoint() (entryWorldRecovery, error) {
	if app.offlineSession == nil {
		return entryWorldRecovery{}, fmt.Errorf("entry-world recovery: offline session is unavailable")
	}

	app.worldMu.RLock()
	topology := app.entryWorldTopology
	wantChecksum := app.entryWorldChecksum
	app.worldMu.RUnlock()

	snapshot, err := topology.Snapshot()
	if err != nil {
		return entryWorldRecovery{}, fmt.Errorf("entry-world recovery: snapshot topology: %w", err)
	}

	checksum, err := topology.Checksum()
	if err != nil || checksum != wantChecksum {
		return entryWorldRecovery{}, fmt.Errorf("entry-world recovery: published topology checksum differs: %v", err)
	}

	checkpoint, err := app.offlineSession.RecoveryCheckpoint()
	if err != nil {
		return entryWorldRecovery{}, err
	}

	return entryWorldRecovery{Topology: snapshot, TopologyChecksum: checksum, Session: checkpoint}, nil
}

// prepareOfflineEntryWorldRecovery validates the saved topology before composition. buildEntryWorld then restores
// these exact zones and seams, so its maps and runtime identity are derived from the saved world.
func (app *application) prepareOfflineEntryWorldRecovery(bundle entryWorldRecovery) error {
	if _, err := validateEntryWorldRecovery(bundle); err != nil {
		return err
	}

	app.entryWorldSnapshot = append([]byte(nil), bundle.Topology...)
	copyBundle := bundle
	copyBundle.Topology = append([]byte(nil), bundle.Topology...)
	app.pendingEntryRecovery = &copyBundle

	return nil
}

// validateEntryWorldRecovery authenticates both serialized authorities before either composition or restore uses them.
func validateEntryWorldRecovery(bundle entryWorldRecovery) (string, error) {
	if err := gamesession.ValidateRecoveryCheckpoint(bundle.Session); err != nil {
		return "", fmt.Errorf("entry-world recovery: invalid session checkpoint: %w", err)
	}

	topology, err := d2mapgen.RestoreEntryWorld(bundle.Topology)
	if err != nil {
		return "", fmt.Errorf("entry-world recovery: invalid topology: %w", err)
	}

	checksum, err := topology.Checksum()
	if err != nil || checksum != bundle.TopologyChecksum {
		return "", fmt.Errorf("%w: saved topology checksum differs", gamesession.ErrCompatibility)
	}

	return checksum, nil
}

// preflightOfflineEntryWorldRecovery compares the prepared map and participant recipe before mutating session state.
func (app *application) preflightOfflineEntryWorldRecovery(bundle entryWorldRecovery) error {
	if app.offlineSession == nil {
		return fmt.Errorf("entry-world recovery: offline session is unavailable")
	}

	checksum, err := validateEntryWorldRecovery(bundle)
	if err != nil {
		return err
	}

	app.worldMu.RLock()
	currentChecksum := app.entryWorldChecksum
	app.worldMu.RUnlock()

	if checksum != currentChecksum {
		return fmt.Errorf("%w: saved topology differs from composed maps", gamesession.ErrCompatibility)
	}

	savedIdentity, err := simulation.RuntimeIdentityFromParticipants(bundle.Session.State.Participants)
	if err != nil {
		return fmt.Errorf("entry-world recovery: saved runtime identity: %w", err)
	}

	current, err := app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		return err
	}

	currentIdentity, err := simulation.RuntimeIdentityFromParticipants(current.Participants)
	if err != nil {
		return fmt.Errorf("entry-world recovery: composed runtime identity: %w", err)
	}

	savedDigest, savedErr := savedIdentity.Digest()

	currentDigest, currentErr := currentIdentity.Digest()
	if savedErr != nil || currentErr != nil || savedDigest != currentDigest {
		return fmt.Errorf("%w: saved runtime identity differs from composed runtime", gamesession.ErrCompatibility)
	}

	return nil
}

// restoreOfflineEntryWorldRecovery applies a validated checkpoint after the saved topology has composed this runtime.
func (app *application) restoreOfflineEntryWorldRecovery(bundle entryWorldRecovery) error {
	if err := app.preflightOfflineEntryWorldRecovery(bundle); err != nil {
		return err
	}

	if err := app.offlineSession.RestoreRecoveryCheckpoint(bundle.Session); err != nil {
		return err
	}

	app.syncActiveWorldFromPlayer()

	return nil
}
