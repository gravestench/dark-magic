package serverapp

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

var ErrRemoteProfileAdmission = errors.New("server: remote player-profile admission rejected")

const (
	remoteProfileBurst = 8.0
	remoteProfileRate  = 1.0
)

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
	clients  map[string]profileAdmissionBucket
	now      func() time.Time
}

type profileAdmissionBucket struct {
	tokens  float64
	updated time.Time
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
	return &RemoteProfileAdmissions{host: host, tickets: tickets, config: config,
		clients: make(map[string]profileAdmissionBucket), now: time.Now}, nil
}

// Admit authenticates one bounded selected-character offer, queues it as
// system authority, and returns a one-use ordinary session ticket.
func (admissions *RemoteProfileAdmissions) Admit(ctx context.Context, credential string, offer []byte) (string, error) {
	admissions.mu.Lock()
	defer admissions.mu.Unlock()
	if !admissions.take(profileAdmissionClient(ctx)) {
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

func (admissions *RemoteProfileAdmissions) take(client string) bool {
	now := admissions.now()
	bucket, found := admissions.clients[client]
	if !found {
		bucket = profileAdmissionBucket{tokens: remoteProfileBurst, updated: now}
	}
	if now.Before(bucket.updated) {
		now = bucket.updated
	}
	bucket.tokens = min(remoteProfileBurst, bucket.tokens+now.Sub(bucket.updated).Seconds()*remoteProfileRate)
	bucket.updated = now
	if bucket.tokens < 1 {
		admissions.clients[client] = bucket
		return false
	}
	bucket.tokens--
	admissions.clients[client] = bucket
	return true
}

type profileAdmissionClientKey struct{}

// WithProfileAdmissionClient lets transports bind admission throttling to a
// normalized remote IP without expanding the profile interface or trusting a
// client-supplied identifier.
func WithProfileAdmissionClient(ctx context.Context, address string) context.Context {
	host := address
	if parsed, err := netip.ParseAddrPort(address); err == nil {
		host = parsed.Addr().Unmap().String()
	}
	return context.WithValue(ctx, profileAdmissionClientKey{}, host)
}

func (admissions *RemoteProfileAdmissions) WithClient(ctx context.Context, address string) context.Context {
	return WithProfileAdmissionClient(ctx, address)
}

func profileAdmissionClient(ctx context.Context) string {
	if ctx != nil {
		if value, ok := ctx.Value(profileAdmissionClientKey{}).(string); ok && value != "" {
			return value
		}
	}
	return "unknown"
}
