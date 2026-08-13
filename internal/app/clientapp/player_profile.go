package clientapp

import (
	"errors"
	"fmt"
	"os"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

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
