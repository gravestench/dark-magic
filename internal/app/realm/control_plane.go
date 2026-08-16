package realm

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const RealmControlPlaneVersion = "RealmControlPlane/v1"

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

type departureReceipt struct {
	Record        CharacterRecord `json:"record"`
	PlayerID      string          `json:"player_id"`
	WorkerRemoved bool            `json:"worker_removed"`
}

const (
	workerFailureThreshold = 3
)

var ErrGameUnavailable = errors.New("realm: game service unavailable")

type GameHandoff struct {
	Game       GameDetail     `json:"game"`
	Assignment JoinAssignment `json:"assignment"`
}

func (control *ControlPlane) ReconnectGame(ctx context.Context, token, gameID string) (handoff GameHandoff, err error) {
	event := AuditEvent{Operation: AuditGameReconnect, GameID: strings.TrimSpace(gameID)}
	defer func() { control.recordAudit(ctx, event, err) }()
	if control == nil || control.admissions == nil {
		return GameHandoff{}, ErrGameUnavailable
	}
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	detail, err := control.games.admissionDetail(ctx, strings.TrimSpace(gameID))
	if err != nil {
		return GameHandoff{}, err
	}
	assignment, err := control.admissions.ReconnectAssignment(ctx, detail.Entry.GameID, principal.accountID)
	if err != nil {
		return GameHandoff{}, err
	}
	event.CharacterID = ""
	return GameHandoff{Game: detail, Assignment: assignment}, nil
}

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
		admissions, err = NewAdmissionsWithMemberships(config.Allocator, characters, memberships, config.LeaseLifetime, config.TicketLifetime)
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
	return &ControlPlane{version: RealmControlPlaneVersion, accounts: accounts, accountLifecycle: lifecycle,
		channels: NewChannels(config.ChatHistory), games: games, allocations: allocations, membershipStore: memberships,
		checkpoints: checkpoints, characters: characters,
		characterCompatibility: config.CharacterCompatibility, audit: config.Audit, allocator: config.Allocator,
		admissions: admissions, entryDestination: config.EntryDestination,
		healthFailures: make(map[string]int), checkpointInterval: config.CheckpointInterval, presenceTimeout: config.PresenceTimeout,
		lastCheckpoint: make(map[string]time.Time)}, nil
}

func (control *ControlPlane) Signup(ctx context.Context, request SignupRequest) (account Account, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{Operation: AuditAccountSignup, AccountID: account.ID,
			AccountName: firstNonEmpty(account.Name, strings.TrimSpace(request.Name))}, err)
	}()
	if control == nil || control.accountLifecycle == nil {
		return Account{}, ErrAccountInput
	}
	return control.accountLifecycle.Signup(ctx, request)
}

func (control *ControlPlane) VerifyEmail(ctx context.Context, challenge string) (account Account, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{Operation: AuditAccountVerify, AccountID: account.ID,
			AccountName: account.Name}, err)
	}()
	if control == nil || control.accountLifecycle == nil {
		return Account{}, ErrAccountChallenge
	}
	return control.accountLifecycle.VerifyEmail(ctx, challenge)
}

func (control *ControlPlane) BeginPasswordRecovery(ctx context.Context, email string) (err error) {
	defer func() { control.recordAudit(ctx, AuditEvent{Operation: AuditAccountRecoveryBegin}, err) }()
	if control == nil || control.accountLifecycle == nil {
		return ErrAccountInput
	}
	return control.accountLifecycle.BeginPasswordRecovery(ctx, email)
}

func (control *ControlPlane) CompletePasswordRecovery(ctx context.Context, challenge, password string) (err error) {
	defer func() { control.recordAudit(ctx, AuditEvent{Operation: AuditAccountRecoveryComplete}, err) }()
	if control == nil || control.accountLifecycle == nil {
		return ErrAccountChallenge
	}
	return control.accountLifecycle.CompletePasswordRecovery(ctx, challenge, password)
}

func (control *ControlPlane) Version() string {
	if control == nil {
		return ""
	}
	return control.version
}

func (control *ControlPlane) CreateAccount(ctx context.Context, name, password string) (account Account, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{Operation: AuditAccountCreate, AccountID: account.ID,
			AccountName: firstNonEmpty(account.Name, strings.TrimSpace(name))}, err)
	}()
	if control == nil || control.accounts == nil {
		return Account{}, ErrAccountInput
	}
	return control.accounts.Create(ctx, name, password)
}

func (control *ControlPlane) Authenticate(ctx context.Context, name, password string) (session RealmSession, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{Operation: AuditAccountLogin, AccountID: session.Account.ID,
			AccountName: firstNonEmpty(session.Account.Name, strings.TrimSpace(name)), SessionID: session.ID}, err)
	}()
	if control == nil || control.accounts == nil {
		return RealmSession{}, ErrAccountCredentials
	}
	return control.accounts.Authenticate(ctx, name, password)
}

type CreateCharacterRequest struct {
	Name      string `json:"name"`
	Class     string `json:"class"`
	Expansion bool   `json:"expansion"`
	Hardcore  bool   `json:"hardcore"`
}

const maximumRealmCharacters = 18

func (control *ControlPlane) ListCharacters(ctx context.Context, token string) ([]CharacterRecord, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return nil, err
	}
	return control.characters.List(ctx, principal.accountID)
}

// CreateCharacter accepts only player choices. Identity, level, stats,
// appearance, revision, ownership, and compatibility are realm/d2legacy owned.
func (control *ControlPlane) CreateCharacter(ctx context.Context, token string, request CreateCharacterRequest) (record CharacterRecord, err error) {
	event := AuditEvent{Operation: AuditCharacterCreate, CharacterName: strings.TrimSpace(request.Name)}
	defer func() {
		event.AccountID = firstNonEmpty(event.AccountID, record.AccountID)
		event.CharacterID = record.Character.ID
		event.CharacterName = firstNonEmpty(record.Character.Name, event.CharacterName)
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	existing, err := control.characters.List(ctx, principal.accountID)
	if err != nil {
		return CharacterRecord{}, err
	}
	if len(existing) >= maximumRealmCharacters {
		return CharacterRecord{}, ErrCharacterLimit
	}
	wanted := strings.ToLower(strings.TrimSpace(request.Name))
	for _, record := range existing {
		if strings.ToLower(record.Character.Name) == wanted {
			return CharacterRecord{}, ErrCharacterExists
		}
	}
	character, err := d2save.NewCharacter(d2save.CharacterRequest{ID: uuid.New().String(), Name: request.Name,
		Class: request.Class, Expansion: request.Expansion, Hardcore: request.Hardcore})
	if err != nil {
		return CharacterRecord{}, ErrCharacterInput
	}
	record = CharacterRecord{AccountID: principal.accountID, Revision: 1, Character: character,
		Compatibility: control.characterCompatibility}
	if err := control.characters.Create(ctx, record); err != nil {
		return CharacterRecord{}, err
	}
	if err := control.accounts.SelectCharacter(ctx, token, character.ID); err != nil {
		return CharacterRecord{}, err
	}
	return cloneCharacterRecord(record), nil
}

func (control *ControlPlane) DeleteCharacter(ctx context.Context, token, characterID string) (err error) {
	event := AuditEvent{Operation: AuditCharacterDelete, CharacterID: strings.TrimSpace(characterID)}
	defer func() { control.recordAudit(ctx, event, err) }()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	return control.characters.Delete(ctx, principal.accountID, strings.TrimSpace(characterID))
}

func (control *ControlPlane) SelectCharacter(ctx context.Context, token, characterID string) (record CharacterRecord, err error) {
	event := AuditEvent{Operation: AuditCharacterSelect, CharacterID: strings.TrimSpace(characterID)}
	defer func() {
		event.CharacterName = record.Character.Name
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	record, err = control.characters.Get(ctx, principal.accountID, strings.TrimSpace(characterID))
	if err != nil {
		return CharacterRecord{}, err
	}
	if err := control.accounts.SelectCharacter(ctx, token, record.Character.ID); err != nil {
		return CharacterRecord{}, err
	}
	return record, nil
}

func (control *ControlPlane) SelectedCharacter(ctx context.Context, token string) (CharacterRecord, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	return control.characters.Get(ctx, principal.accountID, characterID)
}

func (control *ControlPlane) JoinChannel(ctx context.Context, token, channel string) (view ChannelView, err error) {
	event := AuditEvent{Operation: AuditChannelJoin, Channel: strings.TrimSpace(channel)}
	defer func() {
		event.Channel = firstNonEmpty(view.Name, event.Channel)
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return ChannelView{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return ChannelView{}, err
	}
	record, err := control.characters.Get(ctx, principal.accountID, characterID)
	if err != nil {
		return ChannelView{}, err
	}
	event.CharacterID, event.CharacterName = record.Character.ID, record.Character.Name
	return control.channels.Join(ctx, principal, channel, presenceFromCharacter(record))
}

func (control *ControlPlane) ChannelView(ctx context.Context, token string) (ChannelView, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return ChannelView{}, err
	}
	return control.channels.View(ctx, principal)
}

func (control *ControlPlane) SendChannelMessage(ctx context.Context, token, message string) (chatEvent ChatEvent, err error) {
	event := AuditEvent{Operation: AuditChannelMessage, MessageBytes: len(message)}
	defer func() {
		event.Channel = chatEvent.ChannelID
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return ChatEvent{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	return control.channels.Send(ctx, principal, message)
}

func (control *ControlPlane) ChannelEvents(ctx context.Context, token string, after uint64, limit int) ([]ChatEvent, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return nil, err
	}
	return control.channels.EventsAfter(ctx, principal, after, limit)
}

func (control *ControlPlane) CreateGame(ctx context.Context, token string, request CreateGameRequest) (handoff GameHandoff, err error) {
	event := AuditEvent{Operation: AuditGameCreate, GameName: strings.TrimSpace(request.Name)}
	defer func() {
		event.GameID = handoff.Game.Entry.GameID
		event.GameName = firstNonEmpty(handoff.Game.Entry.Name, event.GameName)
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	if characterID, selectedErr := control.accounts.SelectedCharacter(ctx, token); selectedErr == nil {
		if record, recordErr := control.characters.Get(ctx, principal.accountID, characterID); recordErr == nil {
			event.CharacterID, event.CharacterName = record.Character.ID, record.Character.Name
		}
	}
	if control.allocator == nil || control.admissions == nil {
		return GameHandoff{}, ErrGameUnavailable
	}
	detail, err := control.games.Create(ctx, principal, request)
	if err != nil {
		return GameHandoff{}, err
	}
	allocationID := uuid.New().String()
	if _, err := control.allocations.Request(ctx, detail.Entry.GameID, allocationID); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		return GameHandoff{}, errors.Join(err, control.games.Remove(cleanupCtx, detail.Entry.GameID))
	}
	allocation, err := control.allocator.Allocate(ctx, GameSpec{GameID: detail.Entry.GameID, AllocationID: allocationID})
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		return GameHandoff{}, errors.Join(err, control.allocations.Fail(cleanupCtx, detail.Entry.GameID, err),
			control.games.Remove(cleanupCtx, detail.Entry.GameID))
	}
	if allocation.AllocationID != allocationID || allocation.Worker == nil {
		err = ErrWorker
		return GameHandoff{}, errors.Join(err, control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err))
	}
	description, err := allocation.Worker.Describe(ctx)
	if err != nil {
		return GameHandoff{}, errors.Join(err, control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err))
	}
	if _, err := control.allocations.Ready(ctx, detail.Entry.GameID, allocation.Endpoint, description.Runtime); err != nil {
		return GameHandoff{}, errors.Join(err, control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err))
	}
	if err := control.admissions.RegisterGame(detail.Entry.GameID, allocation.Tickets, allocation.Endpoint); err != nil {
		return GameHandoff{}, errors.Join(err, control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err))
	}
	handoff, err = control.joinResolvedGame(ctx, token, principal, detail)
	if err != nil {
		return GameHandoff{}, errors.Join(err, control.removeAllocatedGame(ctx, detail.Entry.GameID, true, err))
	}
	return handoff, nil
}

func (control *ControlPlane) ListGames(ctx context.Context, token string, filter GameFilter) ([]GameDirectoryEntry, error) {
	if _, err := control.authorize(ctx, token); err != nil {
		return nil, err
	}
	return control.games.List(ctx, filter)
}

func (control *ControlPlane) GameDetail(ctx context.Context, token, reference string) (GameDetail, error) {
	if _, err := control.authorize(ctx, token); err != nil {
		return GameDetail{}, err
	}
	return control.games.Detail(ctx, reference)
}

func (control *ControlPlane) ResolveGameJoin(ctx context.Context, token, reference, password string) (gameID string, err error) {
	event := AuditEvent{Operation: AuditGameResolve, GameReference: strings.TrimSpace(reference)}
	defer func() {
		event.GameID = gameID
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return "", err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	return control.games.ResolveJoin(ctx, reference, password)
}

func (control *ControlPlane) JoinGame(ctx context.Context, token, reference, password string) (handoff GameHandoff, err error) {
	event := AuditEvent{Operation: AuditGameJoin, GameReference: strings.TrimSpace(reference)}
	defer func() {
		event.GameID = handoff.Game.Entry.GameID
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	if characterID, selectedErr := control.accounts.SelectedCharacter(ctx, token); selectedErr == nil {
		if record, recordErr := control.characters.Get(ctx, principal.accountID, characterID); recordErr == nil {
			event.CharacterID, event.CharacterName = record.Character.ID, record.Character.Name
		}
	}
	if control.allocator == nil || control.admissions == nil {
		return GameHandoff{}, ErrGameUnavailable
	}
	gameID, err := control.games.ResolveJoin(ctx, reference, password)
	if err != nil {
		return GameHandoff{}, err
	}
	detail, err := control.games.admissionDetail(ctx, gameID)
	if err != nil {
		return GameHandoff{}, err
	}
	handoff, err = control.joinResolvedGame(ctx, token, principal, detail)
	return handoff, err
}

func (control *ControlPlane) joinResolvedGame(ctx context.Context, token string, principal AuthenticatedPrincipal, detail GameDetail) (GameHandoff, error) {
	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}
	record, err := control.characters.Get(ctx, principal.accountID, characterID)
	if err != nil {
		return GameHandoff{}, err
	}
	if record.Character.Expansion != detail.Entry.Expansion || record.Character.Hardcore != detail.Entry.Hardcore {
		return GameHandoff{}, ErrGameDirectoryInput
	}
	if detail.Entry.CharacterDifference > 0 && len(detail.Players) > 0 {
		difference := record.Character.Level - detail.Players[0].Level
		if difference < 0 {
			difference = -difference
		}
		if difference > detail.Entry.CharacterDifference {
			return GameHandoff{}, ErrGameLevelRange
		}
	}
	playerID := uuid.New().String()
	reservation, err := control.games.ReservePlayer(ctx, detail.Entry.GameID, GamePlayer{CharacterID: record.Character.ID,
		Name: record.Character.Name, Class: record.Character.Class, Level: record.Character.Level})
	if err != nil {
		return GameHandoff{}, err
	}
	assignment, err := control.admissions.Join(ctx, JoinRequest{AccountID: principal.accountID, CharacterID: record.Character.ID,
		PlayerID: playerID, GameID: detail.Entry.GameID, Destination: control.entryDestination})
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		return GameHandoff{}, errors.Join(err, control.games.CancelPlayer(cleanupCtx, reservation))
	}
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	updated, err := control.games.CommitPlayer(cleanupCtx, reservation)
	if err != nil {
		return GameHandoff{}, errors.Join(err, control.admissions.CancelMembership(cleanupCtx, detail.Entry.GameID, playerID),
			control.games.CancelPlayer(cleanupCtx, reservation))
	}
	// Successful admission moves the session out of the public channel at once.
	// This keeps the lobby roster a projection of current channel occupants rather
	// than waiting for the renewable presence lease to expire.
	// Presence cleanup is intentionally subordinate to the already-committed
	// admission. Maintenance will remove it if this best-effort leave fails.
	_ = control.channels.Leave(cleanupCtx, principal)
	return GameHandoff{Game: updated, Assignment: assignment}, nil
}

func (control *ControlPlane) removeAllocatedGame(ctx context.Context, gameID string, registered bool, cause error) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	var result error
	if registered {
		result = errors.Join(result, control.admissions.UnregisterGame(gameID))
	}
	result = errors.Join(result, control.allocator.Release(cleanupCtx, gameID))
	result = errors.Join(result, control.games.Remove(cleanupCtx, gameID))
	if cause == nil && result == nil {
		result = errors.Join(result, control.allocations.Complete(cleanupCtx, gameID))
		if result == nil {
			result = errors.Join(result, control.checkpoints.Remove(cleanupCtx, gameID))
			control.checkpointMu.Lock()
			delete(control.lastCheckpoint, gameID)
			control.checkpointMu.Unlock()
		}
	} else {
		result = errors.Join(result, control.allocations.Fail(cleanupCtx, gameID, errors.Join(cause, result)))
	}
	return result
}

// LeaveGame commits only the worker's canonical projection. The authenticated
// account selects the membership; clients cannot provide player IDs, revisions,
// lease tokens, or replacement character state.
func (control *ControlPlane) LeaveGame(ctx context.Context, token, gameID string) (record CharacterRecord, err error) {
	event := AuditEvent{Operation: AuditGameLeave, GameID: strings.TrimSpace(gameID)}
	defer func() {
		event.CharacterID, event.CharacterName = record.Character.ID, record.Character.Name
		control.recordAudit(ctx, event, err)
	}()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	// Serialize receipt creation with membership consumption so concurrent HTTP
	// retries cannot both pass the pre-commit lookup.
	control.departureFlowMu.Lock()
	defer control.departureFlowMu.Unlock()
	if completed, found, lookupErr := control.departure(ctx, gameID, principal.accountID); lookupErr != nil {
		return CharacterRecord{}, lookupErr
	} else if found {
		return completed.Record, control.completeDeparture(ctx, gameID, principal.accountID, completed)
	}
	if control.allocator == nil || control.admissions == nil {
		return CharacterRecord{}, ErrGameUnavailable
	}
	playerID, baseline, err := control.admissions.AccountMembership(gameID, principal.accountID)
	if err != nil {
		return CharacterRecord{}, err
	}
	event.CharacterID, event.CharacterName = baseline.Character.ID, baseline.Character.Name
	record, commitErr := control.admissions.LeaveCanonicalMembership(ctx, gameID, playerID)
	if record.Character.ID == "" {
		return CharacterRecord{}, commitErr
	}
	membership, receiptErr := control.membershipStore.ByAccount(ctx, gameID, principal.accountID)
	if receiptErr != nil || membership.Departure == nil {
		return record, errors.Join(commitErr, receiptErr, ErrMembership)
	}
	receipt := cloneDepartureReceipt(*membership.Departure)
	return record, control.completeDeparture(ctx, gameID, principal.accountID, receipt)
}

func (control *ControlPlane) completeDeparture(ctx context.Context, gameID, accountID string, receipt departureReceipt) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if !receipt.WorkerRemoved {
		worker, found := control.allocator.Game(gameID)
		if !found {
			return ErrGameNotFound
		}
		if err := worker.RemoveCharacter(cleanupCtx, receipt.PlayerID); err != nil {
			return err
		}
		updated, err := control.membershipStore.MarkWorkerRemoved(cleanupCtx, gameID, receipt.PlayerID)
		if err != nil {
			return err
		}
		receipt = updated
	}
	detail, rosterErr := control.games.RemovePlayer(cleanupCtx, gameID, receipt.Record.Character.ID)
	if errors.Is(rosterErr, ErrGameNotFound) || errors.Is(rosterErr, ErrCharacterNotFound) {
		return nil
	}
	if rosterErr != nil {
		return rosterErr
	}
	if detail.Entry.Players == 0 {
		return control.removeAllocatedGame(cleanupCtx, gameID, true, nil)
	}
	return nil
}

type GameDrainResult struct {
	GameID              string `json:"game_id"`
	CommittedCharacters int    `json:"committed_characters"`
}

// DrainGame is the trusted operator lifecycle path. BeginDrain first closes
// discovery and all new admission atomically. Every active membership then
// follows the ordinary canonical projection, durable departure receipt, worker
// removal, and roster cleanup sequence. A partial failure leaves the durable
// game draining and is safe to retry.
func (control *ControlPlane) DrainGame(ctx context.Context, gameID string) (result GameDrainResult, err error) {
	gameID = strings.TrimSpace(gameID)
	result.GameID = gameID
	event := AuditEvent{Operation: AuditGameDrain, GameID: gameID}
	defer func() { control.recordAudit(ctx, event, err) }()
	if control == nil || ctx == nil || control.allocator == nil || control.admissions == nil || gameID == "" {
		return result, ErrGameUnavailable
	}
	if err := control.games.BeginDrain(ctx, gameID); err != nil {
		return result, err
	}
	players, err := control.membershipStore.DrainPlayerIDs(ctx, gameID)
	if err != nil {
		return result, err
	}
	if len(players) == 0 {
		err = control.removeAllocatedGame(ctx, gameID, true, nil)
		return result, err
	}
	for _, playerID := range players {
		membership, lookupErr := control.membershipStore.ByPlayer(ctx, gameID, playerID)
		if lookupErr != nil {
			return result, lookupErr
		}
		if err := control.reconcileExpiredPlayer(ctx, gameID, playerID); err != nil {
			return result, err
		}
		if membership.State == MembershipActive {
			result.CommittedCharacters++
		}
	}
	return result, nil
}

// ReconcileGames removes allocations whose worker cannot report ready health.
// It preserves the last committed character revision and releases leases; it
// never guesses a checkpoint after the authority is gone.
func (control *ControlPlane) ReconcileGames(ctx context.Context) (int, error) {
	if control == nil || ctx == nil || control.allocator == nil || control.admissions == nil {
		return 0, ErrGameUnavailable
	}
	reconciled := 0
	var result error
	gameIDs, listErr := control.games.gameIDs(ctx)
	if listErr != nil {
		return 0, listErr
	}
	for _, gameID := range gameIDs {
		worker, found := control.allocator.Game(gameID)
		healthy := found
		var status WorkerStatus
		if found {
			var err error
			status, err = worker.Status(ctx)
			healthy = err == nil && status.Ready
		}
		if healthy {
			control.clearHealthFailure(gameID)
			result = errors.Join(result, control.allocations.Healthy(ctx, gameID))
			result = errors.Join(result, control.checkpointGame(ctx, gameID, worker))
			_, renewErr := control.admissions.RenewGameMemberships(ctx, gameID)
			result = errors.Join(result, renewErr)
			for _, playerID := range status.ExpiredPlayers {
				result = errors.Join(result, control.reconcileExpiredPlayer(ctx, gameID, playerID))
			}
			continue
		}
		if control.noteHealthFailure(gameID) < workerFailureThreshold {
			continue
		}
		control.departureFlowMu.Lock()
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		if _, directoryErr := control.games.admissionDetail(cleanupCtx, gameID); errors.Is(directoryErr, ErrGameNotFound) {
			cancel()
			control.departureFlowMu.Unlock()
			control.clearHealthFailure(gameID)
			continue
		}
		restoreCtx, stopRestore := context.WithTimeout(context.WithoutCancel(ctx), 45*time.Second)
		restoreErr := control.restoreAllocatedGame(restoreCtx, gameID)
		stopRestore()
		if restoreErr == nil {
			cancel()
			control.departureFlowMu.Unlock()
			control.recordAudit(ctx, AuditEvent{Operation: AuditGameRestore, GameID: gameID}, nil)
			control.clearHealthFailure(gameID)
			reconciled++
			continue
		}
		err := control.admissions.AbandonGame(cleanupCtx, gameID)
		if releaseErr := control.allocator.Release(cleanupCtx, gameID); releaseErr != nil && !errors.Is(releaseErr, ErrGameNotFound) {
			err = errors.Join(err, releaseErr)
		}
		err = errors.Join(err, control.games.Remove(cleanupCtx, gameID))
		err = errors.Join(err, control.allocations.Fail(cleanupCtx, gameID, errors.Join(ErrWorker, err)))
		cancel()
		control.departureFlowMu.Unlock()
		control.recordAudit(ctx, AuditEvent{Operation: AuditGameReconcile, GameID: gameID}, err)
		control.clearHealthFailure(gameID)
		result = errors.Join(result, err)
		reconciled++
	}
	return reconciled, result
}

func (control *ControlPlane) restoreAllocatedGame(ctx context.Context, gameID string) (err error) {
	restorer, supported := control.allocator.(GameRestorer)
	if !supported {
		return ErrWorker
	}
	allocation, err := control.allocations.Get(ctx, gameID)
	if err != nil || allocation.State != AllocationReady {
		return errors.Join(err, ErrAllocationRecord)
	}
	checkpoint, err := control.checkpoints.Latest(ctx, gameID)
	if err != nil || checkpoint.AllocationID != allocation.AllocationID {
		return errors.Join(err, ErrGameCheckpoint)
	}
	players, err := control.membershipStore.ActivePlayerIDs(ctx, gameID)
	if err != nil || len(players) == 0 {
		return errors.Join(err, ErrMembership)
	}
	recovery, err := NewGameRecovery(checkpoint.Checkpoint, players)
	if err != nil {
		return err
	}
	if releaseErr := control.allocator.Release(ctx, gameID); releaseErr != nil && !errors.Is(releaseErr, ErrGameNotFound) {
		return releaseErr
	}
	replacement, err := restorer.Restore(ctx, GameSpec{GameID: gameID, AllocationID: allocation.AllocationID}, recovery)
	if err != nil {
		return err
	}
	installed := true
	defer func() {
		if err != nil && installed {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			err = errors.Join(err, control.allocator.Release(cleanupCtx, gameID))
			cancel()
		}
	}()
	description, err := replacement.Worker.Describe(ctx)
	if err != nil {
		return err
	}
	wantHash, wantErr := allocation.Runtime.Digest()
	gotHash, gotErr := description.Runtime.Digest()
	if wantErr != nil || gotErr != nil || wantHash != gotHash || gotHash != checkpoint.IdentityHash || description.GameID != gameID ||
		replacement.AllocationID != allocation.AllocationID || replacement.GameID != gameID {
		return ErrWorker
	}
	if _, err = control.allocations.RestoreReady(ctx, gameID, allocation.AllocationID,
		replacement.Endpoint, description.Runtime); err != nil {
		return err
	}
	if err = control.admissions.ReplaceGame(gameID, replacement.Tickets, replacement.Endpoint); err != nil {
		return err
	}
	control.checkpointMu.Lock()
	delete(control.lastCheckpoint, gameID)
	control.checkpointMu.Unlock()
	_ = control.checkpointGame(ctx, gameID, replacement.Worker)
	installed = false
	return nil
}

// checkpointGame captures the worker's canonical simulation state only after
// the durable allocation has been confirmed ready. The allocation generation
// and complete runtime identity are pinned into the record so a stale or
// incompatible worker cannot overwrite a replacement authority's checkpoint.
func (control *ControlPlane) checkpointGame(ctx context.Context, gameID string, worker WorkerClient) error {
	if control == nil || control.allocations == nil || control.checkpoints == nil || worker == nil {
		return ErrGameUnavailable
	}
	now := time.Now().UTC()
	control.checkpointMu.Lock()
	last := control.lastCheckpoint[gameID]
	control.checkpointMu.Unlock()
	if !last.IsZero() && now.Sub(last) < control.checkpointInterval {
		return nil
	}
	allocation, err := control.allocations.Get(ctx, gameID)
	if err != nil || allocation.State != AllocationReady {
		return errors.Join(err, ErrAllocationRecord)
	}
	identityHash, err := allocation.Runtime.Digest()
	if err != nil {
		return ErrAllocationRecord
	}
	checkpoint, err := worker.Checkpoint(ctx)
	if err != nil {
		return err
	}
	record, err := NewGameCheckpoint(gameID, allocation.AllocationID, identityHash, checkpoint)
	if err != nil {
		return err
	}
	_, err = control.checkpoints.Save(ctx, record)
	if err == nil {
		control.checkpointMu.Lock()
		control.lastCheckpoint[gameID] = now
		control.checkpointMu.Unlock()
	}
	return err
}

// RecoverInterruptedGames runs once before the Realm accepts traffic. An
// allocator that supports both fencing and restore must positively stop the
// exact surviving allocation generation before checkpoint replacement. Other
// allocators, incomplete handoffs, and failed fences retain the conservative
// fail-closed cleanup path.
func (control *ControlPlane) RecoverInterruptedGames(ctx context.Context) (int, error) {
	if control == nil || ctx == nil || control.allocations == nil || control.games == nil || control.characters == nil {
		return 0, ErrGameUnavailable
	}
	records, err := control.allocations.Active(ctx)
	if err != nil {
		return 0, err
	}
	recovered := 0
	var result error
	for _, record := range records {
		recoveryCtx, stopRecovery := context.WithTimeout(context.WithoutCancel(ctx), 45*time.Second)
		recoveryErr := control.restoreInterruptedGame(recoveryCtx, record)
		stopRecovery()
		if recoveryErr == nil {
			control.recordAudit(ctx, AuditEvent{Operation: AuditGameRestore, GameID: record.GameID}, nil)
			recovered++
			continue
		}
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		var cleanupErr error
		if control.allocator != nil {
			if _, found := control.allocator.Game(record.GameID); found {
				cleanupErr = errors.Join(cleanupErr, control.allocator.Release(cleanupCtx, record.GameID))
			}
		}
		_, leaseErr := control.characters.ReleaseGame(cleanupCtx, record.GameID)
		cleanupErr = errors.Join(cleanupErr, leaseErr)
		cleanupErr = errors.Join(cleanupErr, control.membershipStore.AbandonGame(cleanupCtx, record.GameID))
		if directoryErr := control.games.Remove(cleanupCtx, record.GameID); directoryErr != nil && !errors.Is(directoryErr, ErrGameNotFound) {
			cleanupErr = errors.Join(cleanupErr, directoryErr)
		}
		cause := errors.Join(errors.New("realm: allocation interrupted by Realm restart"), recoveryErr, cleanupErr)
		cleanupErr = errors.Join(cleanupErr, control.allocations.Fail(cleanupCtx, record.GameID, cause))
		cancel()
		control.recordAudit(ctx, AuditEvent{Operation: AuditGameReconcile, GameID: record.GameID}, cleanupErr)
		result = errors.Join(result, cleanupErr)
		recovered++
	}
	return recovered, result
}

func (control *ControlPlane) restoreInterruptedGame(ctx context.Context, record AllocationRecord) (err error) {
	fencer, canFence := control.allocator.(GameFencer)
	restorer, canRestore := control.allocator.(GameRestorer)
	if !canFence || !canRestore || record.State != AllocationReady {
		return ErrWorker
	}
	checkpoint, err := control.checkpoints.Latest(ctx, record.GameID)
	if err != nil || checkpoint.AllocationID != record.AllocationID {
		return errors.Join(err, ErrGameCheckpoint)
	}
	players, err := control.membershipStore.ActivePlayerIDs(ctx, record.GameID)
	if err != nil || len(players) == 0 {
		return errors.Join(err, ErrMembership)
	}
	recovery, err := NewGameRecovery(checkpoint.Checkpoint, players)
	if err != nil {
		return err
	}
	spec := GameSpec{GameID: record.GameID, AllocationID: record.AllocationID}
	if err := fencer.Fence(ctx, spec); err != nil {
		return err
	}
	replacement, err := restorer.Restore(ctx, spec, recovery)
	if err != nil {
		return err
	}
	installed := true
	defer func() {
		if err != nil && installed {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			err = errors.Join(err, control.allocator.Release(cleanupCtx, record.GameID))
			cancel()
		}
	}()
	description, err := replacement.Worker.Describe(ctx)
	if err != nil {
		return err
	}
	wantHash, wantErr := record.Runtime.Digest()
	gotHash, gotErr := description.Runtime.Digest()
	if wantErr != nil || gotErr != nil || wantHash != gotHash || gotHash != checkpoint.IdentityHash ||
		description.GameID != record.GameID || replacement.GameID != record.GameID || replacement.AllocationID != record.AllocationID {
		return ErrWorker
	}
	if _, err = control.allocations.RestoreReady(ctx, record.GameID, record.AllocationID,
		replacement.Endpoint, description.Runtime); err != nil {
		return err
	}
	if _, err = control.admissions.ResumeGame(ctx, record.GameID, replacement.Tickets, replacement.Endpoint); err != nil {
		return err
	}
	control.checkpointMu.Lock()
	delete(control.lastCheckpoint, record.GameID)
	control.checkpointMu.Unlock()
	_ = control.checkpointGame(ctx, record.GameID, replacement.Worker)
	installed = false
	return nil
}

// reconcileExpiredPlayer handles a trusted worker notification emitted only
// after transport reconnect grace has elapsed. Canonical projection and the
// lease commit still happen before entity, roster, or allocation removal.
func (control *ControlPlane) reconcileExpiredPlayer(ctx context.Context, gameID, playerID string) (err error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return ErrWorker
	}
	control.departureFlowMu.Lock()
	defer control.departureFlowMu.Unlock()
	accountID, baseline, membershipErr := control.admissions.PlayerMembership(gameID, playerID)
	receiptAccountID, receipt, completed, receiptErr := control.departureByPlayer(ctx, gameID, playerID)
	if receiptErr != nil {
		return receiptErr
	}
	if membershipErr != nil && !completed {
		return membershipErr
	}
	if completed {
		accountID = receiptAccountID
	}
	if !completed {
		record, commitErr := control.admissions.LeaveCanonicalMembership(ctx, gameID, playerID)
		if record.Character.ID == "" {
			return commitErr
		}
		membership, durableErr := control.membershipStore.ByPlayer(ctx, gameID, playerID)
		if durableErr != nil || membership.Departure == nil {
			return errors.Join(commitErr, durableErr, ErrMembership)
		}
		receipt = cloneDepartureReceipt(*membership.Departure)
	}
	event := AuditEvent{Operation: AuditGameLeave, GameID: gameID, AccountID: accountID,
		CharacterID:   firstNonEmpty(receipt.Record.Character.ID, baseline.Character.ID),
		CharacterName: firstNonEmpty(receipt.Record.Character.Name, baseline.Character.Name)}
	defer func() { control.recordAudit(ctx, event, err) }()
	return control.completeDeparture(ctx, gameID, accountID, receipt)
}

func (control *ControlPlane) departure(ctx context.Context, gameID, accountID string) (departureReceipt, bool, error) {
	record, err := control.membershipStore.ByAccount(ctx, gameID, accountID)
	if errors.Is(err, ErrMembership) {
		return departureReceipt{}, false, nil
	}
	if err != nil {
		return departureReceipt{}, false, err
	}
	if record.State != MembershipDeparted || record.Departure == nil {
		return departureReceipt{}, false, nil
	}
	return cloneDepartureReceipt(*record.Departure), true, nil
}

func (control *ControlPlane) departureByPlayer(ctx context.Context, gameID, playerID string) (string, departureReceipt, bool, error) {
	record, err := control.membershipStore.ByPlayer(ctx, gameID, playerID)
	if errors.Is(err, ErrMembership) {
		return "", departureReceipt{}, false, nil
	}
	if err != nil {
		return "", departureReceipt{}, false, err
	}
	if record.State != MembershipDeparted || record.Departure == nil {
		return "", departureReceipt{}, false, nil
	}
	return record.AccountID, cloneDepartureReceipt(*record.Departure), true, nil
}

func (control *ControlPlane) noteHealthFailure(gameID string) int {
	control.lifecycleMu.Lock()
	defer control.lifecycleMu.Unlock()
	control.healthFailures[gameID]++
	return control.healthFailures[gameID]
}

func (control *ControlPlane) clearHealthFailure(gameID string) {
	control.lifecycleMu.Lock()
	defer control.lifecycleMu.Unlock()
	delete(control.healthFailures, gameID)
}

// Logout removes channel presence before invalidating the session. Failure to
// find channel membership is harmless; an invalid/expired session still fails.
func (control *ControlPlane) Logout(ctx context.Context, token string) (err error) {
	event := AuditEvent{Operation: AuditAccountLogout}
	defer func() { control.recordAudit(ctx, event, err) }()
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return err
	}
	event.AccountID, event.AccountName, event.SessionID = principal.accountID, principal.name, principal.sessionID
	leaveErr := control.channels.Leave(ctx, principal)
	if errors.Is(leaveErr, ErrChannelMember) {
		leaveErr = nil
	}
	return errors.Join(leaveErr, control.accounts.Logout(ctx, token))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (control *ControlPlane) authorize(ctx context.Context, token string) (AuthenticatedPrincipal, error) {
	if control == nil || control.accounts == nil || control.channels == nil || control.games == nil || control.characters == nil {
		return AuthenticatedPrincipal{}, ErrRealmSession
	}
	if _, err := control.PruneExpiredSessions(ctx); err != nil {
		return AuthenticatedPrincipal{}, err
	}
	return control.accounts.Authorize(ctx, token)
}

// PruneExpiredSessions is safe to call from a maintenance ticker and is also
// invoked before authorization so ordinary realm traffic clears ghost presence.
func (control *ControlPlane) PruneExpiredSessions(ctx context.Context) (int, error) {
	if control == nil || control.accounts == nil || control.channels == nil {
		return 0, ErrRealmSession
	}
	expired, err := control.accounts.PruneExpired(ctx)
	if err != nil {
		return 0, err
	}
	for _, principal := range expired {
		leaveErr := control.channels.Leave(ctx, principal)
		if errors.Is(leaveErr, ErrChannelMember) {
			leaveErr = nil
		}
		control.recordAudit(ctx, AuditEvent{Operation: AuditAccountExpire, AccountID: principal.accountID,
			AccountName: principal.name, SessionID: principal.sessionID}, leaveErr)
		if leaveErr != nil {
			return 0, leaveErr
		}
	}
	return len(expired), nil
}

func (control *ControlPlane) PruneInactivePresence(ctx context.Context) (int, error) {
	if control == nil || control.channels == nil || control.presenceTimeout <= 0 {
		return 0, ErrChannelInput
	}
	return control.channels.PruneInactive(ctx, time.Now().UTC().Add(-control.presenceTimeout))
}

func presenceFromCharacter(record CharacterRecord) CharacterPresence {
	character := record.Character
	return CharacterPresence{CharacterID: character.ID, Name: character.Name, Class: character.Class, Level: character.Level,
		Expansion: character.Expansion, Hardcore: character.Hardcore, Appearance: clonePresence(CharacterPresence{Appearance: character.Appearance}).Appearance}
}
