package serverapp

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

var ErrRemoteProfileAdmission = errors.New("server: remote player-profile admission rejected")

const maxRemoteProfileAttempts = 8

// directPlayerSpawnSpacing is measured in DS1 subtiles. A two-subtile offset
// projects to only 16 by 8 pixels and leaves full player composites visually
// superimposed. Eight subtiles keeps a joining party nearby while separating
// their roughly character-width retained sprites at the initial camera frame.
const directPlayerSpawnSpacing = 8.0

// RemoteProfileConfig is explicit self-host policy. PlayerID and destination
// come from the host; neither is accepted from the remote character offer.
type RemoteProfileConfig struct {
	Credential  string
	AllowDirect bool
	PrincipalID string
	PlayerID    string
	Destination playeradapter.Destination
	Lifetime    time.Duration
}

type RemoteProfileAdmissions struct {
	mu       sync.Mutex
	host     *gameserver.Host
	tickets  *gameserver.TicketAuthority
	config   RemoteProfileConfig
	sequence uint64
	attempts int
}

func NewRemoteProfileAdmissions(host *gameserver.Host, tickets *gameserver.TicketAuthority, config RemoteProfileConfig) (*RemoteProfileAdmissions, error) {
	if host == nil || host.Session == nil || host.Mode == gameserver.ModeRealm || tickets == nil ||
		(!config.AllowDirect && strings.TrimSpace(config.Credential) == "") || strings.TrimSpace(config.PrincipalID) == "" ||
		strings.TrimSpace(config.PlayerID) == "" || config.Lifetime <= 0 {
		return nil, ErrRemoteProfileAdmission
	}
	if _, err := playeradapter.NewDestination(config.Destination.X, config.Destination.Y, config.Destination.Width,
		config.Destination.Height, config.Destination.Act, config.Destination.LevelID); err != nil {
		return nil, ErrRemoteProfileAdmission
	}
	return &RemoteProfileAdmissions{host: host, tickets: tickets, config: config}, nil
}

// Admit authenticates one bounded selected-character offer, queues it as
// system authority, and returns a one-use ordinary session ticket.
func (admissions *RemoteProfileAdmissions) Admit(_ context.Context, credential string, offer []byte) (string, error) {
	admissions.mu.Lock()
	defer admissions.mu.Unlock()
	admissions.attempts++
	if admissions.attempts > maxRemoteProfileAttempts {
		return "", ErrRemoteProfileAdmission
	}
	if !admissions.config.AllowDirect && subtle.ConstantTimeCompare([]byte(credential), []byte(admissions.config.Credential)) != 1 {
		return "", ErrRemoteProfileAdmission
	}
	character, err := d2save.DecodeCharacterOffer(offer)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrRemoteProfileAdmission, err)
	}
	admissions.sequence++
	sequence := admissions.sequence
	destination := admissions.config.Destination
	// Direct/listen players need distinct visible and collision-valid entry
	// positions. Keep the host-selected anchor authoritative while spreading a
	// small deterministic row inside world bounds.
	destination.X = min(destination.X+float64(sequence-1)*directPlayerSpawnSpacing, destination.Width-1)
	principal := gameserver.Principal{ID: fmt.Sprintf("%s-%d", admissions.config.PrincipalID, admissions.sequence), CharacterID: character.ID,
		PlayerID: fmt.Sprintf("%s-%d", admissions.config.PlayerID, admissions.sequence), RuntimeIdentityHash: admissions.host.Allocation.IdentityHash}
	ticket, err := admissions.tickets.Issue(principal, admissions.config.Lifetime)
	if err != nil {
		return "", err
	}
	err = admissions.host.Session.SubmitNext(func(tick uint64) (simulation.Command, error) {
		return playeradapter.AdmissionCommand(character, principal.PlayerID, destination,
			"self-host:remote-profile", sequence, tick, simulation.AuthoritySystem)
	})
	if err != nil {
		_ = admissions.tickets.Revoke(ticket)
		return "", fmt.Errorf("%w: submit entry: %v", ErrRemoteProfileAdmission, err)
	}
	return ticket, nil
}
