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

// ErrRemoteProfileAdmission deliberately collapses authentication, throttling,
// and offer-validation failures so remote callers cannot probe policy details.
var ErrRemoteProfileAdmission = errors.New("server: remote player-profile admission rejected")

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

// RemoteProfileAdmissions serializes throttling, sequence allocation, ticket
// issuance, and command submission so accepted players retain deterministic
// identities and queue order under concurrent network requests.
type RemoteProfileAdmissions struct {
	mu       sync.Mutex
	host     *gameserver.Host
	tickets  *gameserver.TicketAuthority
	config   RemoteProfileConfig
	sequence uint64
	clients  map[string]profileAdmissionBucket
	now      func() time.Time
}

// NewRemoteProfileAdmissions validates host-owned policy before accepting any
// player-controlled offer. Realm hosts must use the Realm admission path.
func NewRemoteProfileAdmissions(
	host *gameserver.Host,
	tickets *gameserver.TicketAuthority,
	config RemoteProfileConfig,
) (*RemoteProfileAdmissions, error) {
	if err := validateRemoteProfileConfig(host, tickets, config); err != nil {
		return nil, ErrRemoteProfileAdmission
	}

	return &RemoteProfileAdmissions{
		host:    host,
		tickets: tickets,
		config:  config,
		clients: make(map[string]profileAdmissionBucket),
		now:     time.Now,
	}, nil
}

// validateRemoteProfileConfig keeps all trust-bearing identities and spawn
// bounds host-supplied, and rejects configurations that cannot issue tickets.
func validateRemoteProfileConfig(
	host *gameserver.Host,
	tickets *gameserver.TicketAuthority,
	config RemoteProfileConfig,
) error {
	if host == nil || host.Session == nil || host.Mode == gameserver.ModeRealm || tickets == nil {
		return ErrRemoteProfileAdmission
	}

	if !config.AllowDirect && strings.TrimSpace(config.Credential) == "" {
		return ErrRemoteProfileAdmission
	}

	if strings.TrimSpace(config.PrincipalID) == "" || strings.TrimSpace(config.PlayerID) == "" || config.Lifetime <= 0 {
		return ErrRemoteProfileAdmission
	}

	_, err := playeradapter.NewDestination(
		config.Destination.X,
		config.Destination.Y,
		config.Destination.Width,
		config.Destination.Height,
		config.Destination.Act,
		config.Destination.LevelID,
	)
	if err != nil {
		return ErrRemoteProfileAdmission
	}

	return nil
}

// Admit authenticates one bounded selected-character offer, queues it as
// system authority, and returns a one-use ordinary session ticket.
func (admissions *RemoteProfileAdmissions) Admit(ctx context.Context, credential string, offer []byte) (string, error) {
	// The lock covers the complete admission transaction: accepting a later
	// request first would disconnect ticket identity from command queue order.
	admissions.mu.Lock()
	defer admissions.mu.Unlock()

	if !admissions.take(profileAdmissionClient(ctx)) {
		return "", ErrRemoteProfileAdmission
	}

	if !admissions.credentialAccepted(credential) {
		return "", ErrRemoteProfileAdmission
	}

	character, err := d2save.DecodeCharacterOffer(offer)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrRemoteProfileAdmission, err)
	}

	sequence, destination, principal := admissions.reserveAdmission(character)

	return admissions.issueTicketAndQueueEntry(character, sequence, destination, principal)
}

// credentialAccepted uses constant-time comparison for the shared secret;
// direct mode is an explicit host policy that bypasses credential matching.
func (admissions *RemoteProfileAdmissions) credentialAccepted(credential string) bool {
	return admissions.config.AllowDirect ||
		subtle.ConstantTimeCompare([]byte(credential), []byte(admissions.config.Credential)) == 1
}

// reserveAdmission advances the sequence before ticket issuance, preserving
// unique identities even when a later issuance or queue operation fails.
func (admissions *RemoteProfileAdmissions) reserveAdmission(
	character d2save.Character,
) (uint64, playeradapter.Destination, gameserver.Principal) {
	admissions.sequence++
	sequence := admissions.sequence
	destination := admissions.config.Destination

	// Direct/listen players need distinct visible and collision-valid entry
	// positions. Keep the host-selected anchor authoritative while spreading a
	// small deterministic row inside world bounds.
	destination.X = min(destination.X+float64(sequence-1)*directPlayerSpawnSpacing, destination.Width-1)
	principal := gameserver.Principal{
		ID:                  fmt.Sprintf("%s-%d", admissions.config.PrincipalID, sequence),
		CharacterID:         character.ID,
		PlayerID:            fmt.Sprintf("%s-%d", admissions.config.PlayerID, sequence),
		RuntimeIdentityHash: admissions.host.Allocation.IdentityHash,
	}

	return sequence, destination, principal
}

// issueTicketAndQueueEntry revokes the newly minted ticket when command
// submission fails, preventing credentials for a player who cannot enter.
func (admissions *RemoteProfileAdmissions) issueTicketAndQueueEntry(
	character d2save.Character,
	sequence uint64,
	destination playeradapter.Destination,
	principal gameserver.Principal,
) (string, error) {
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
