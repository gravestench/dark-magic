package realm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	GameDirectoryVersion        = "RealmGameDirectory/v1"
	maximumGameNameBytes        = 32
	maximumGameDescriptionBytes = 255
	maximumGamePasswordBytes    = 64
	maximumGamePlayers          = 8
)

var (
	ErrGameDirectoryInput = errors.New("realm: invalid game directory input")
	ErrGamePassword       = errors.New("realm: invalid game password")
	ErrGameFull           = errors.New("realm: game is full")
	ErrGameLevelRange     = errors.New("realm: character level is outside the game range")
)

type GameDifficulty string

const (
	DifficultyNormal    GameDifficulty = "normal"
	DifficultyNightmare GameDifficulty = "nightmare"
	DifficultyHell      GameDifficulty = "hell"
)

type CreateGameRequest struct {
	Name                string         `json:"name"`
	Password            string         `json:"password,omitempty"`
	Description         string         `json:"description,omitempty"`
	Difficulty          GameDifficulty `json:"difficulty"`
	Maximum             int            `json:"maximum_players"`
	CharacterDifference int            `json:"character_difference,omitempty"`
	Expansion           bool           `json:"expansion"`
	Hardcore            bool           `json:"hardcore"`
}

type GamePlayer struct {
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Level       int    `json:"level"`
}

type GameDirectoryEntry struct {
	Version             string         `json:"version"`
	Revision            uint64         `json:"revision"`
	GameID              string         `json:"game_id"`
	Name                string         `json:"name"`
	Description         string         `json:"description,omitempty"`
	CreatedBy           string         `json:"created_by"`
	Difficulty          GameDifficulty `json:"difficulty"`
	Players             int            `json:"players"`
	MaximumPlayers      int            `json:"maximum_players"`
	CharacterDifference int            `json:"character_difference,omitempty"`
	PasswordRequired    bool           `json:"password_required"`
	Expansion           bool           `json:"expansion"`
	Hardcore            bool           `json:"hardcore"`
	CreatedAt           time.Time      `json:"created_at"`
}

type GameDetail struct {
	Entry   GameDirectoryEntry `json:"entry"`
	Players []GamePlayer       `json:"players"`
}

type GameFilter struct {
	Difficulty *GameDifficulty `json:"difficulty,omitempty"`
	Expansion  *bool           `json:"expansion,omitempty"`
	Hardcore   *bool           `json:"hardcore,omitempty"`
}

type directoryGame struct {
	entry        GameDirectoryEntry
	state        string
	ownerAccount string
	passwordHash []byte
	players      []GamePlayer
	reservations map[string]GamePlayer
}

type GamePlayerReservation struct {
	GameID string
	Token  string
}

// GameRepository is the durable Realm directory boundary. Production uses
// PostgreSQL; GameDirectory remains the deterministic in-memory test adapter.
// Reservation tokens are private capabilities and must be persisted only by
// digest by durable implementations.
type GameRepository interface {
	Create(context.Context, AuthenticatedPrincipal, CreateGameRequest) (GameDetail, error)
	List(context.Context, GameFilter) ([]GameDirectoryEntry, error)
	Detail(context.Context, string) (GameDetail, error)
	admissionDetail(context.Context, string) (GameDetail, error)
	ResolveJoin(context.Context, string, string) (string, error)
	ReservePlayer(context.Context, string, GamePlayer) (GamePlayerReservation, error)
	CommitPlayer(context.Context, GamePlayerReservation) (GameDetail, error)
	CancelPlayer(context.Context, GamePlayerReservation) error
	SetPlayers(context.Context, string, []GamePlayer) error
	BeginDrain(context.Context, string) error
	RemovePlayer(context.Context, string, string) (GameDetail, error)
	Remove(context.Context, string) error
	gameIDs(context.Context) ([]string, error)
}

// GameDirectory owns discoverable realm game names separately from opaque
// GameIDs and gameplay endpoints. It never exposes worker addresses or password
// material in browser projections.
type GameDirectory struct {
	mu     sync.RWMutex
	now    func() time.Time
	byID   map[string]*directoryGame
	byName map[string]string
}

func NewGameDirectory() *GameDirectory {
	return &GameDirectory{now: time.Now, byID: make(map[string]*directoryGame), byName: make(map[string]string)}
}

func (directory *GameDirectory) Create(ctx context.Context, principal AuthenticatedPrincipal, request CreateGameRequest) (GameDetail, error) {
	request.Expansion = true
	if err := contextErr(ctx); err != nil {
		return GameDetail{}, err
	}
	if directory == nil || !principal.valid() {
		return GameDetail{}, ErrGameDirectoryInput
	}
	displayName, normalizedName, err := normalizeGameName(request.Name)
	if err != nil || validateCreateGame(request) != nil {
		return GameDetail{}, ErrGameDirectoryInput
	}
	var passwordHash []byte
	if request.Password != "" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			return GameDetail{}, err
		}
	}
	if err := contextErr(ctx); err != nil {
		return GameDetail{}, err
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	if _, exists := directory.byName[normalizedName]; exists {
		return GameDetail{}, ErrGameExists
	}
	gameID := uuid.New().String()
	entry := GameDirectoryEntry{Version: GameDirectoryVersion, Revision: 1, GameID: gameID, Name: displayName,
		Description: strings.TrimSpace(request.Description), CreatedBy: principal.name, Difficulty: request.Difficulty,
		MaximumPlayers: request.Maximum, CharacterDifference: request.CharacterDifference,
		PasswordRequired: len(passwordHash) != 0, Expansion: request.Expansion,
		Hardcore: request.Hardcore, CreatedAt: directory.now().UTC()}
	game := &directoryGame{entry: entry, state: activeRealmGameState, ownerAccount: principal.accountID, passwordHash: passwordHash,
		reservations: make(map[string]GamePlayer)}
	directory.byID[gameID], directory.byName[normalizedName] = game, gameID
	return gameDetail(game), nil
}

func (directory *GameDirectory) List(ctx context.Context, filter GameFilter) ([]GameDirectoryEntry, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if directory == nil {
		return nil, ErrGameDirectoryInput
	}
	directory.mu.RLock()
	defer directory.mu.RUnlock()
	result := make([]GameDirectoryEntry, 0, len(directory.byID))
	for _, game := range directory.byID {
		if game.state != activeRealmGameState {
			continue
		}
		// Password-protected legacy realm games are joined manually by exact
		// name and do not disclose themselves in the public browser.
		if game.entry.PasswordRequired ||
			filter.Difficulty != nil && game.entry.Difficulty != *filter.Difficulty ||
			filter.Expansion != nil && game.entry.Expansion != *filter.Expansion ||
			filter.Hardcore != nil && game.entry.Hardcore != *filter.Hardcore {
			continue
		}
		result = append(result, game.entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].GameID < result[j].GameID
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (directory *GameDirectory) Detail(ctx context.Context, reference string) (GameDetail, error) {
	if err := contextErr(ctx); err != nil {
		return GameDetail{}, err
	}
	if directory == nil {
		return GameDetail{}, ErrGameDirectoryInput
	}
	directory.mu.RLock()
	defer directory.mu.RUnlock()
	game := directory.resolveLocked(reference)
	if game == nil || game.state != activeRealmGameState || game.entry.PasswordRequired {
		return GameDetail{}, ErrGameNotFound
	}
	return gameDetail(game), nil
}

func (directory *GameDirectory) admissionDetail(ctx context.Context, gameID string) (GameDetail, error) {
	if err := contextErr(ctx); err != nil {
		return GameDetail{}, err
	}
	directory.mu.RLock()
	defer directory.mu.RUnlock()
	game := directory.byID[strings.TrimSpace(gameID)]
	if game == nil || game.state != activeRealmGameState {
		return GameDetail{}, ErrGameNotFound
	}
	return gameDetail(game), nil
}

// ResolveJoin converts a player-entered name or selected opaque GameID into the
// canonical GameID after password and best-effort capacity preflight.
// Admissions and the selected worker must revalidate capacity atomically while
// issuing the subsequent character lease and worker ticket.
func (directory *GameDirectory) ResolveJoin(ctx context.Context, reference, password string) (string, error) {
	if err := contextErr(ctx); err != nil {
		return "", err
	}
	if directory == nil || len(password) > maximumGamePasswordBytes {
		return "", ErrGameDirectoryInput
	}
	directory.mu.RLock()
	game := directory.resolveLocked(reference)
	if game == nil || game.state != activeRealmGameState {
		directory.mu.RUnlock()
		return "", ErrGameNotFound
	}
	gameID, full := game.entry.GameID, len(game.players)+len(game.reservations) >= game.entry.MaximumPlayers
	passwordHash := append([]byte(nil), game.passwordHash...)
	directory.mu.RUnlock()
	if len(passwordHash) == 0 {
		if password != "" {
			return "", ErrGamePassword
		}
	} else if bcrypt.CompareHashAndPassword(passwordHash, []byte(password)) != nil {
		return "", ErrGamePassword
	}
	if full {
		return "", ErrGameFull
	}
	return gameID, nil
}

// ReservePlayer atomically claims capacity before character admission. Pending
// reservations are private and never inflate the public player roster.
func (directory *GameDirectory) ReservePlayer(ctx context.Context, gameID string, player GamePlayer) (GamePlayerReservation, error) {
	if err := contextErr(ctx); err != nil {
		return GamePlayerReservation{}, err
	}
	if directory == nil {
		return GamePlayerReservation{}, ErrGameDirectoryInput
	}
	if _, err := validateGamePlayers([]GamePlayer{player}); err != nil {
		return GamePlayerReservation{}, err
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	game := directory.byID[strings.TrimSpace(gameID)]
	if game == nil || game.state != activeRealmGameState {
		return GamePlayerReservation{}, ErrGameNotFound
	}
	if len(game.players)+len(game.reservations) >= game.entry.MaximumPlayers {
		return GamePlayerReservation{}, ErrGameFull
	}
	for _, existing := range game.players {
		if existing.CharacterID == player.CharacterID {
			return GamePlayerReservation{}, ErrCharacterLeased
		}
	}
	for _, existing := range game.reservations {
		if existing.CharacterID == player.CharacterID {
			return GamePlayerReservation{}, ErrCharacterLeased
		}
	}
	token := uuid.New().String()
	game.reservations[token] = player
	return GamePlayerReservation{GameID: game.entry.GameID, Token: token}, nil
}

func (directory *GameDirectory) CommitPlayer(ctx context.Context, reservation GamePlayerReservation) (GameDetail, error) {
	if err := contextErr(ctx); err != nil {
		return GameDetail{}, err
	}
	if directory == nil || strings.TrimSpace(reservation.Token) == "" {
		return GameDetail{}, ErrGameDirectoryInput
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	game := directory.byID[strings.TrimSpace(reservation.GameID)]
	if game == nil || game.state != activeRealmGameState {
		return GameDetail{}, ErrGameNotFound
	}
	player, found := game.reservations[reservation.Token]
	if !found {
		return GameDetail{}, ErrGameDirectoryInput
	}
	delete(game.reservations, reservation.Token)
	game.players = append(game.players, player)
	game.entry.Players = len(game.players)
	game.entry.Revision++
	return gameDetail(game), nil
}

func (directory *GameDirectory) CancelPlayer(ctx context.Context, reservation GamePlayerReservation) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if directory == nil || strings.TrimSpace(reservation.Token) == "" {
		return ErrGameDirectoryInput
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	game := directory.byID[strings.TrimSpace(reservation.GameID)]
	if game == nil {
		return ErrGameNotFound
	}
	if _, found := game.reservations[reservation.Token]; !found {
		return ErrGameDirectoryInput
	}
	delete(game.reservations, reservation.Token)
	return nil
}

// SetPlayers is a trusted realm/game-worker projection update. Public clients
// can only observe the resulting copied detail; they cannot call this through a
// player capability.
func (directory *GameDirectory) SetPlayers(ctx context.Context, gameID string, players []GamePlayer) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if directory == nil {
		return ErrGameDirectoryInput
	}
	cloned, err := validateGamePlayers(players)
	if err != nil {
		return err
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	game := directory.byID[strings.TrimSpace(gameID)]
	if game == nil {
		return ErrGameNotFound
	}
	if len(cloned) > game.entry.MaximumPlayers {
		return ErrGameFull
	}
	game.players = cloned
	game.entry.Players = len(cloned)
	game.entry.Revision++
	return nil
}

// BeginDrain atomically closes discovery and admission before canonical player
// departure begins. It is idempotent so an interrupted operator drain can be
// retried without reopening the game.
func (directory *GameDirectory) BeginDrain(ctx context.Context, gameID string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if directory == nil {
		return ErrGameDirectoryInput
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	game := directory.byID[strings.TrimSpace(gameID)]
	if game == nil {
		return ErrGameNotFound
	}
	if game.state == drainingRealmGameState {
		return nil
	}
	if game.state != activeRealmGameState {
		return ErrGameNotFound
	}
	game.state = drainingRealmGameState
	game.reservations = make(map[string]GamePlayer)
	game.entry.Revision++
	return nil
}

// RemovePlayer commits the public roster half of a trusted Realm departure.
// Character persistence is completed through Admissions before this method is
// called; public clients cannot mutate this roster directly.
func (directory *GameDirectory) RemovePlayer(ctx context.Context, gameID, characterID string) (GameDetail, error) {
	if err := contextErr(ctx); err != nil {
		return GameDetail{}, err
	}
	if directory == nil || strings.TrimSpace(characterID) == "" {
		return GameDetail{}, ErrGameDirectoryInput
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	game := directory.byID[strings.TrimSpace(gameID)]
	if game == nil {
		return GameDetail{}, ErrGameNotFound
	}
	for index, player := range game.players {
		if player.CharacterID != characterID {
			continue
		}
		game.players = append(game.players[:index:index], game.players[index+1:]...)
		game.entry.Players = len(game.players)
		game.entry.Revision++
		return gameDetail(game), nil
	}
	return GameDetail{}, ErrCharacterNotFound
}

func (directory *GameDirectory) Remove(ctx context.Context, gameID string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if directory == nil {
		return ErrGameDirectoryInput
	}
	directory.mu.Lock()
	defer directory.mu.Unlock()
	game := directory.byID[strings.TrimSpace(gameID)]
	if game == nil {
		return ErrGameNotFound
	}
	delete(directory.byID, game.entry.GameID)
	delete(directory.byName, strings.ToLower(game.entry.Name))
	return nil
}

func (directory *GameDirectory) gameIDs(ctx context.Context) ([]string, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if directory == nil {
		return nil, ErrGameDirectoryInput
	}
	directory.mu.RLock()
	defer directory.mu.RUnlock()
	result := make([]string, 0, len(directory.byID))
	for gameID := range directory.byID {
		result = append(result, gameID)
	}
	sort.Strings(result)
	return result, nil
}

func (directory *GameDirectory) resolveLocked(reference string) *directoryGame {
	reference = strings.TrimSpace(reference)
	if game := directory.byID[reference]; game != nil {
		return game
	}
	return directory.byID[directory.byName[strings.ToLower(strings.Join(strings.Fields(reference), " "))]]
}

func normalizeGameName(name string) (string, string, error) {
	display := strings.Join(strings.Fields(name), " ")
	if display == "" || len(display) > maximumGameNameBytes || !utf8.ValidString(display) {
		return "", "", ErrGameDirectoryInput
	}
	for _, value := range display {
		if value < 0x20 || value == 0x7f {
			return "", "", ErrGameDirectoryInput
		}
	}
	return display, strings.ToLower(display), nil
}

func validateCreateGame(request CreateGameRequest) error {
	if len(request.Description) > maximumGameDescriptionBytes || !utf8.ValidString(request.Description) ||
		len(request.Password) > maximumGamePasswordBytes || request.Maximum < 1 || request.Maximum > maximumGamePlayers ||
		request.CharacterDifference < 0 || request.CharacterDifference > 99 {
		return ErrGameDirectoryInput
	}
	switch request.Difficulty {
	case DifficultyNormal, DifficultyNightmare, DifficultyHell:
		return nil
	default:
		return ErrGameDirectoryInput
	}
}

func validateGamePlayers(players []GamePlayer) ([]GamePlayer, error) {
	result := append([]GamePlayer(nil), players...)
	seen := make(map[string]struct{}, len(result))
	for _, player := range result {
		if strings.TrimSpace(player.CharacterID) == "" || strings.TrimSpace(player.Name) == "" || strings.TrimSpace(player.Class) == "" || player.Level < 1 {
			return nil, ErrGameDirectoryInput
		}
		if _, exists := seen[player.CharacterID]; exists {
			return nil, ErrGameDirectoryInput
		}
		seen[player.CharacterID] = struct{}{}
	}
	return result, nil
}

func gameDetail(game *directoryGame) GameDetail {
	return GameDetail{Entry: game.entry, Players: append([]GamePlayer(nil), game.players...)}
}
