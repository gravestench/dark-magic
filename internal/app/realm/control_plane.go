package realm

import (
	"errors"
	"sync"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

const RealmControlPlaneVersion = "RealmControlPlane/v1"

// ControlPlaneConfig supplies the durable stores and worker infrastructure that
// the realm service composes. Nil repositories select their in-memory defaults.
type ControlPlaneConfig struct {
	SessionLifetime    time.Duration
	PresenceTimeout    time.Duration
	ChatHistory        int
	Accounts           AccountRepository
	AccountLifecycle   AccountLifecycle
	Characters         CharacterRepository
	Games              GameRepository
	Allocations        AllocationRepository
	Memberships        MembershipRepository
	Checkpoints        CheckpointRepository
	Audit              AuditSink
	Allocator          GameAllocator
	EntryDestination   playeradapter.Destination
	LeaseLifetime      time.Duration
	TicketLifetime     time.Duration
	CheckpointInterval time.Duration
	// CharacterCompatibility pins newly created records to the realm's
	// authoritative package recipe. Workers revalidate it at admission.
	CharacterCompatibility gamesession.DurableCompatibility
}

// ControlPlane is the transport-neutral realm service composition. Network and
// Lua adapters use its authenticated semantic methods; the backing account,
// character, channel, and directory services remain private.
type ControlPlane struct {
	version                string
	accounts               AccountRepository
	accountLifecycle       AccountLifecycle
	channels               *Channels
	games                  GameRepository
	allocations            AllocationRepository
	membershipStore        MembershipRepository
	checkpoints            CheckpointRepository
	characters             CharacterRepository
	characterCompatibility gamesession.DurableCompatibility
	audit                  AuditSink
	allocator              GameAllocator
	admissions             *Admissions
	entryDestination       playeradapter.Destination
	departureFlowMu        sync.Mutex
	lifecycleMu            sync.Mutex
	healthFailures         map[string]int
	checkpointInterval     time.Duration
	presenceTimeout        time.Duration
	checkpointMu           sync.Mutex
	lastCheckpoint         map[string]time.Time
}

var ErrGameUnavailable = errors.New("realm: game service unavailable")

// NewControlPlane composes the realm services and installs conservative
// lifetimes when callers omit them. Worker admission remains disabled unless
// an allocator is explicitly supplied.
func NewControlPlane(config ControlPlaneConfig) (*ControlPlane, error) {
	accounts := config.Accounts

	var err error
	if accounts == nil {
		accounts, err = NewAccounts(config.SessionLifetime)
		if err != nil {
			return nil, err
		}
	}

	lifecycle := config.AccountLifecycle
	if lifecycle == nil {
		lifecycle, _ = accounts.(AccountLifecycle)
	}

	characters := config.Characters
	if characters == nil {
		characters, err = NewMemoryCharacters()
		if err != nil {
			return nil, err
		}
	}

	games := config.Games
	if games == nil {
		games = NewGameDirectory()
	}

	allocations := config.Allocations
	if allocations == nil {
		allocations = NewMemoryAllocations()
	}

	memberships := config.Memberships
	if memberships == nil {
		memberships, err = NewMemoryMemberships(characters)
		if err != nil {
			return nil, err
		}
	}

	checkpoints := config.Checkpoints
	if checkpoints == nil {
		checkpoints = NewMemoryCheckpoints()
	}

	var admissions *Admissions

	if config.Allocator != nil {
		if config.LeaseLifetime <= 0 {
			config.LeaseLifetime = 2 * time.Minute
		}

		if config.TicketLifetime <= 0 {
			config.TicketLifetime = 30 * time.Second
		}

		admissions, err = NewAdmissionsWithMemberships(
			config.Allocator,
			characters,
			memberships,
			config.LeaseLifetime,
			config.TicketLifetime,
		)
		if err != nil {
			return nil, err
		}
	}

	if config.CheckpointInterval <= 0 {
		config.CheckpointInterval = 15 * time.Second
	}

	if config.PresenceTimeout <= 0 {
		config.PresenceTimeout = 30 * time.Second
	}

	return &ControlPlane{
		version:                RealmControlPlaneVersion,
		accounts:               accounts,
		accountLifecycle:       lifecycle,
		channels:               NewChannels(config.ChatHistory),
		games:                  games,
		allocations:            allocations,
		membershipStore:        memberships,
		checkpoints:            checkpoints,
		characters:             characters,
		characterCompatibility: config.CharacterCompatibility,
		audit:                  config.Audit,
		allocator:              config.Allocator,
		admissions:             admissions,
		entryDestination:       config.EntryDestination,
		healthFailures:         make(map[string]int),
		checkpointInterval:     config.CheckpointInterval,
		presenceTimeout:        config.PresenceTimeout,
		lastCheckpoint:         make(map[string]time.Time),
	}, nil
}

// Version identifies the semantic contract implemented by the control plane;
// a nil receiver reports no supported version.
func (control *ControlPlane) Version() string {
	if control == nil {
		return ""
	}

	return control.version
}
