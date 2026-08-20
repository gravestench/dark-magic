package serverapp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// ErrProfileAdmission hides profile parsing and projection details behind the
// stable admission failure contract expected by command startup.
var ErrProfileAdmission = errors.New("server: invalid player-profile admission")

// ProfileAdmission identifies the locally selected character and the
// host-authoritative player identity and spawn destination assigned to it.
type ProfileAdmission struct {
	Path        string
	PlayerID    string
	Destination playeradapter.Destination
}

// AdmitSelectedProfile queues one host-trusted entry from a player-controlled
// profile. It is deliberately separate from realm admission and network join:
// configuring the host is the act that trusts this profile for this game only.
func AdmitSelectedProfile(host *gameserver.Host, config ProfileAdmission) error {
	if strings.TrimSpace(config.Path) == "" {
		return nil
	}

	if err := validateProfileAdmission(host, config); err != nil {
		return err
	}

	_, character, err := readSelectedProfile(config.Path)
	if err != nil {
		return err
	}
	// The host supplies player identity, destination, and system authority;
	// none of those trust decisions come from the player-controlled profile.
	err = host.Session.SubmitNext(func(tick uint64) (simulation.Command, error) {
		return playeradapter.AdmissionCommand(character, config.PlayerID, config.Destination,
			"self-host:profile", 1, tick, simulation.AuthoritySystem)
	})
	if err != nil {
		return fmt.Errorf("%w: submit entry: %v", ErrProfileAdmission, err)
	}

	return nil
}

// PersistSelectedProfile writes the host's canonical character state back to
// the same player-controlled profile. It never uses realm repository APIs.
func PersistSelectedProfile(host *gameserver.Host, config ProfileAdmission) error {
	if strings.TrimSpace(config.Path) == "" {
		return nil
	}

	if err := validateProfileAdmission(host, config); err != nil {
		return err
	}

	profile, baseline, err := readSelectedProfile(config.Path)
	if err != nil {
		return err
	}

	checkpoint, err := host.Session.CanonicalCheckpoint()
	if err != nil {
		return err
	}

	character, err := playeradapter.ProjectCharacter(config.PlayerID, baseline, checkpoint)
	if err != nil {
		return fmt.Errorf("%w: project canonical character: %v", ErrProfileAdmission, err)
	}

	// Replace only the selected roster slot so unrelated local characters and
	// profile metadata retain their original ordering and ownership.
	for index := range profile.Characters {
		if profile.Characters[index].ID == profile.Selected {
			profile.Characters[index] = character
			break
		}
	}

	return d2save.WriteProfileFile(config.Path, profile)
}

// validateProfileAdmission rejects an unusable host before profile parsing;
// an empty path remains an intentional no-op handled by each public operation.
func validateProfileAdmission(host *gameserver.Host, config ProfileAdmission) error {
	if host == nil || host.Session == nil || strings.TrimSpace(config.PlayerID) == "" {
		return ErrProfileAdmission
	}

	return nil
}

// readSelectedProfile applies the same strict roster validation to admission
// and persistence so neither path can silently choose a different character.
func readSelectedProfile(path string) (d2save.Profile, d2save.Character, error) {
	profile, err := d2save.ReadProfileFile(path)
	if err != nil {
		return d2save.Profile{}, d2save.Character{}, fmt.Errorf("%w: %v", ErrProfileAdmission, err)
	}

	store, err := d2save.NewFromProfile(profile)
	if err != nil {
		return d2save.Profile{}, d2save.Character{}, fmt.Errorf("%w: %v", ErrProfileAdmission, err)
	}

	character, selected := store.Selected()
	if !selected {
		return d2save.Profile{}, d2save.Character{}, fmt.Errorf(
			"%w: profile has no selected character",
			ErrProfileAdmission,
		)
	}

	return profile, character, nil
}
