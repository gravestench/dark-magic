package clientapp

import (
	"errors"
	"fmt"
	"os"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// loadPlayerProfile chooses ephemeral fixtures before disk state so development runs cannot overwrite real saves.
func loadPlayerProfile(path string, fixtures []d2save.Character) (*d2save.Store, string, error) {
	// Development fixtures are ephemeral and must never replace a real player
	// profile merely because a capture or test scene was launched.
	if len(fixtures) > 0 {
		return d2save.New(fixtures...), "", nil
	}
	if path == "" {
		return d2save.New(), "", nil
	}

	profile, err := d2save.ReadProfileFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return d2save.New(), path, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("load player profile: %w", err)
	}

	store, err := d2save.NewFromProfile(profile)
	if err != nil {
		return nil, "", fmt.Errorf("restore player profile: %w", err)
	}

	return store, path, nil
}

// persistOfflineCharacter projects canonical authority state back into the
// selected save without touching remote sessions.
func persistOfflineCharacter(saves *d2save.Store, session *gamesession.Session, playerID string) error {
	if saves == nil || session == nil {
		return nil
	}
	baseline, selected := saves.Selected()
	if !selected {
		return nil
	}

	checkpoint, err := session.CanonicalCheckpoint()
	if err != nil {
		return fmt.Errorf("persist single-player character: %w", err)
	}

	updated, err := playeradapter.ProjectCharacter(playerID, baseline, checkpoint)
	if err != nil {
		return fmt.Errorf("persist single-player character: %w", err)
	}
	if err := saves.UpdateSelected(updated); err != nil {
		return fmt.Errorf("persist single-player character: %w", err)
	}

	return nil
}
