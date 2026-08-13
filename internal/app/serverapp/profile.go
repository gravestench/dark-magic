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

var ErrProfileAdmission = errors.New("server: invalid player-profile admission")

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
	if host == nil || host.Session == nil || strings.TrimSpace(config.PlayerID) == "" {
		return ErrProfileAdmission
	}
	profile, err := d2save.ReadProfileFile(config.Path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProfileAdmission, err)
	}
	store, err := d2save.NewFromProfile(profile)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProfileAdmission, err)
	}
	character, selected := store.Selected()
	if !selected {
		return fmt.Errorf("%w: profile has no selected character", ErrProfileAdmission)
	}
	checkpoint, err := host.Session.CanonicalCheckpoint()
	if err != nil {
		return err
	}
	command, err := playeradapter.AdmissionCommand(character, config.PlayerID, config.Destination,
		"self-host:profile", 1, checkpoint.Tick+1, simulation.AuthoritySystem)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProfileAdmission, err)
	}
	if err := host.Session.Submit(command); err != nil {
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
	if host == nil || host.Session == nil || strings.TrimSpace(config.PlayerID) == "" {
		return ErrProfileAdmission
	}
	profile, err := d2save.ReadProfileFile(config.Path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProfileAdmission, err)
	}
	store, err := d2save.NewFromProfile(profile)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProfileAdmission, err)
	}
	baseline, selected := store.Selected()
	if !selected {
		return fmt.Errorf("%w: profile has no selected character", ErrProfileAdmission)
	}
	checkpoint, err := host.Session.CanonicalCheckpoint()
	if err != nil {
		return err
	}
	character, err := playeradapter.ProjectCharacter(config.PlayerID, baseline, checkpoint)
	if err != nil {
		return fmt.Errorf("%w: project canonical character: %v", ErrProfileAdmission, err)
	}
	for index := range profile.Characters {
		if profile.Characters[index].ID == profile.Selected {
			profile.Characters[index] = character
			break
		}
	}
	return d2save.WriteProfileFile(config.Path, profile)
}
