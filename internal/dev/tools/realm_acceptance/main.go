// Command realm_acceptance exercises the native Realm-to-worker session path.
// It is intentionally a development acceptance tool, not a player client.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type acceptanceResult struct {
	AccountID         string `json:"account_id"`
	CharacterID       string `json:"character_id"`
	CharacterRevision uint64 `json:"character_revision"`
	GameID            string `json:"game_id,omitempty"`
	PlayerID          string `json:"player_id,omitempty"`
	Mode              string `json:"mode"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "realm acceptance:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	endpoint := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_URL"))
	name := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_ACCOUNT"))
	password, err := acceptancePassword()
	if err != nil {
		return err
	}
	characterName := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_CHARACTER"))
	mode := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_MODE"))
	if mode == "" {
		mode = "session"
	}
	if endpoint == "" || name == "" || password == "" || characterName == "" {
		return errors.New("URL, account, password, and character environment variables are required")
	}
	httpClient, err := localHTTPSClient(endpoint)
	if err != nil {
		return err
	}
	client, err := realm.NewRealmClient(endpoint, httpClient)
	if err != nil {
		return err
	}
	session, err := client.Authenticate(ctx, name, password)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	if mode == "verify" {
		return verifyPersisted(ctx, client, session, characterName)
	}
	if mode != "session" {
		return fmt.Errorf("unsupported mode %q", mode)
	}
	created, err := client.CreateCharacter(ctx, realm.CreateCharacterRequest{
		Name: characterName, Class: "Amazon", Expansion: true,
	})
	if err != nil {
		return fmt.Errorf("create character: %w", err)
	}
	if _, err := client.SelectCharacter(ctx, created.Character.ID); err != nil {
		return fmt.Errorf("select character: %w", err)
	}
	channel, err := client.JoinChannel(ctx, "Diablo II")
	if err != nil || channel.Name == "" {
		return fmt.Errorf("join channel: %w", err)
	}
	if _, err := client.SendMessage(ctx, "Realm acceptance joined the channel."); err != nil {
		return fmt.Errorf("send channel message: %w", err)
	}
	handoff, err := client.CreateGame(ctx, realm.CreateGameRequest{
		Name: "Acceptance Game", Difficulty: realm.DifficultyNormal, Maximum: 8, Expansion: true,
	})
	if err != nil {
		return fmt.Errorf("create named game: %w", err)
	}
	workerTLS, err := networktrust.PinnedTLSFingerprint(handoff.Assignment.Endpoint.TLSFingerprint)
	if err != nil {
		return fmt.Errorf("pin worker identity: %w", err)
	}
	connected, err := clientsession.Connect(ctx, handoff.Assignment, workerTLS)
	if err != nil {
		return fmt.Errorf("join worker over QUIC: %w", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = connected.Close(context.Background())
		}
	}()
	connectedHUD, _ := connected.View()
	if connected.Admission.Admission.CharacterID != created.Character.ID || connectedHUD.Player.PlayerID == "" {
		return errors.New("worker admitted the wrong character identity")
	}
	watchCtx, stopWatch := context.WithCancel(ctx)
	deltas, watchErrors, err := connected.Watch(watchCtx)
	if err != nil {
		return fmt.Errorf("watch worker: %w", err)
	}
	payload, err := json.Marshal(movement.MovePayload{X: 1})
	if err != nil {
		return err
	}
	_, initialWorld := connected.View()
	if err := connected.Submit(ctx, gameserver.CommandIntent{Sequence: 1,
		TargetTick: connected.NextInputTick(time.Now()), Kind: movement.MoveCommand, Payload: payload}); err != nil {
		stopWatch()
		return fmt.Errorf("submit movement: %w", err)
	}
	if err := awaitPlayed(ctx, connected, deltas, watchErrors, initialWorld.Tick); err != nil {
		stopWatch()
		return err
	}
	if err := connected.Reconnect(ctx); err != nil {
		stopWatch()
		return fmt.Errorf("reconnect live QUIC session: %w", err)
	}
	if _, err := connected.Refresh(ctx); err != nil {
		stopWatch()
		return fmt.Errorf("refresh reconnected session: %w", err)
	}
	if err := checkpointBarrier(ctx); err != nil {
		stopWatch()
		return err
	}
	if os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_REALM_RESTART") == "1" {
		reassignment, reconnectErr := client.ReconnectGame(ctx, handoff.Game.Entry.GameID)
		if reconnectErr != nil {
			stopWatch()
			return fmt.Errorf("obtain post-restart assignment: %w", reconnectErr)
		}
		replacementTLS, tlsErr := networktrust.PinnedTLSFingerprint(reassignment.Assignment.Endpoint.TLSFingerprint)
		if tlsErr != nil {
			stopWatch()
			return fmt.Errorf("pin replacement worker identity: %w", tlsErr)
		}
		if reconnectErr = connected.Reassign(ctx, reassignment.Assignment, replacementTLS); reconnectErr != nil {
			stopWatch()
			return fmt.Errorf("reassign post-restart QUIC session: %w", reconnectErr)
		}
		if _, reconnectErr = connected.Refresh(ctx); reconnectErr != nil {
			stopWatch()
			return fmt.Errorf("refresh post-restart session: %w", reconnectErr)
		}
	}
	stopWatch()
	committed, err := client.LeaveGame(ctx, handoff.Game.Entry.GameID)
	if err != nil {
		return fmt.Errorf("commit canonical character: %w", err)
	}
	_ = connected.Close(ctx)
	closed = true
	if committed.Revision != created.Revision+1 || committed.Character.ID != created.Character.ID {
		return fmt.Errorf("unexpected committed character revision %d", committed.Revision)
	}
	if games, err := client.ListGames(ctx); err != nil {
		return fmt.Errorf("list games after departure: %w", err)
	} else if containsGame(games, handoff.Game.Entry.GameID) {
		return errors.New("completed game remains discoverable")
	}
	return json.NewEncoder(os.Stdout).Encode(acceptanceResult{AccountID: session.Account.ID,
		CharacterID: created.Character.ID, CharacterRevision: committed.Revision,
		GameID: handoff.Game.Entry.GameID, PlayerID: connectedHUD.Player.PlayerID, Mode: mode})
}

func acceptancePassword() (string, error) {
	if path := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_PASSWORD_FILE")); path != "" {
		value, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read password file: %w", err)
		}
		password := strings.TrimRight(string(value), "\r\n")
		if password == "" {
			return "", errors.New("password file is empty")
		}
		return password, nil
	}
	password := os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_PASSWORD")
	if password == "" {
		return "", errors.New("password or password file is required")
	}
	return password, nil
}

func localHTTPSClient(endpoint string) (*http.Client, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" {
		return nil, errors.New("acceptance requires an HTTPS Realm URL")
	}
	host := parsed.Hostname()
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return nil, errors.New("development acceptance permits certificate TOFU only on an explicit loopback IP")
	}
	// This disposable loopback test starts with an empty trust store. Worker
	// identity is still strictly pinned from the authenticated Realm handoff.
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} //nolint:gosec
	return &http.Client{Transport: transport, Timeout: 15 * time.Second}, nil
}

func awaitPlayed(ctx context.Context, session *clientsession.Session, deltas <-chan playeradapter.WorldDelta, watchErrors <-chan error, initialTick uint64) error {
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		_, world := session.View()
		if world.Tick > initialTick && len(session.PendingInputs()) == 0 {
			return nil
		}
		select {
		case _, open := <-deltas:
			if !open {
				return errors.New("worker correction stream closed before movement acknowledgement")
			}
		case err := <-watchErrors:
			return fmt.Errorf("worker correction stream: %w", err)
		case <-timer.C:
			return errors.New("worker did not acknowledge movement before timeout")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func checkpointBarrier(ctx context.Context) error {
	ready := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_READY_FILE"))
	proceed := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_CONTINUE_FILE"))
	if ready == "" && proceed == "" {
		return nil
	}
	if ready == "" || proceed == "" {
		return errors.New("ready and continue files must be configured together")
	}
	if err := os.WriteFile(ready, []byte("ready\n"), 0o600); err != nil {
		return fmt.Errorf("publish checkpoint barrier: %w", err)
	}
	for {
		if _, err := os.Stat(proceed); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func containsGame(games []realm.GameDirectoryEntry, gameID string) bool {
	for _, game := range games {
		if game.GameID == gameID {
			return true
		}
	}
	return false
}

func verifyPersisted(ctx context.Context, client *realm.RealmClient, session realm.RealmSession, characterName string) error {
	minimum, err := strconv.ParseUint(strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_MIN_REVISION")), 10, 64)
	if err != nil || minimum == 0 {
		return errors.New("verify mode requires a positive minimum revision")
	}
	characters, err := client.ListCharacters(ctx)
	if err != nil {
		return err
	}
	for _, character := range characters {
		if character.Character.Name == characterName && character.Revision >= minimum {
			return json.NewEncoder(os.Stdout).Encode(acceptanceResult{AccountID: session.Account.ID,
				CharacterID: character.Character.ID, CharacterRevision: character.Revision, Mode: "verify"})
		}
	}
	return errors.New("committed character did not survive Realm restart")
}
