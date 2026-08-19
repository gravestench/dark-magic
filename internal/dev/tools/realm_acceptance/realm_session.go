package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

// acceptanceResult is the stable machine-readable record consumed by the acceptance scripts.
type acceptanceResult struct {
	AccountID         string `json:"account_id"`
	CharacterID       string `json:"character_id"`
	CharacterRevision uint64 `json:"character_revision"`
	GameID            string `json:"game_id,omitempty"`
	PlayerID          string `json:"player_id,omitempty"`
	Mode              string `json:"mode"`
}

// runSessionAcceptance proves the complete Realm lifecycle and emits a result only after canonical persistence.
func runSessionAcceptance(
	ctx context.Context,
	client *realm.RealmClient,
	session realm.RealmSession,
	config acceptanceConfig,
) error {
	created, handoff, err := createAcceptanceGame(ctx, client, config.characterName)
	if err != nil {
		return err
	}

	connected, err := connectWorker(ctx, handoff)
	if err != nil {
		return err
	}

	closed := false
	defer func() {
		if !closed {
			// The acceptance deadline may already be exhausted, so fallback cleanup cannot reuse it.
			_ = connected.Close(context.Background())
		}
	}()

	playerID, err := admittedPlayerID(connected, created.Character.ID)
	if err != nil {
		return err
	}

	if err := exerciseWorkerSession(ctx, client, connected, handoff.Game.Entry.GameID); err != nil {
		return err
	}

	// Leaving commits the worker-owned character before the transport is closed and its final state becomes unavailable.
	committed, err := client.LeaveGame(ctx, handoff.Game.Entry.GameID)
	if err != nil {
		return fmt.Errorf("commit canonical character: %w", err)
	}

	_ = connected.Close(ctx)
	closed = true

	if err := validateCommittedCharacter(created, committed); err != nil {
		return err
	}

	if err := ensureGameDeparted(ctx, client, handoff.Game.Entry.GameID); err != nil {
		return err
	}

	return writeAcceptanceResult(acceptanceResult{
		AccountID:         session.Account.ID,
		CharacterID:       created.Character.ID,
		CharacterRevision: committed.Revision,
		GameID:            handoff.Game.Entry.GameID,
		PlayerID:          playerID,
		Mode:              config.mode,
	})
}

// createAcceptanceGame establishes the character, lobby presence, and named game in their required Realm order.
func createAcceptanceGame(
	ctx context.Context,
	client *realm.RealmClient,
	characterName string,
) (realm.CharacterSummary, realm.GameHandoff, error) {
	created, err := client.CreateCharacter(ctx, realm.CreateCharacterRequest{
		Name:      characterName,
		Class:     "Amazon",
		Expansion: true,
	})
	if err != nil {
		return realm.CharacterSummary{}, realm.GameHandoff{}, fmt.Errorf("create character: %w", err)
	}

	if _, err := client.SelectCharacter(ctx, created.Character.ID); err != nil {
		return realm.CharacterSummary{}, realm.GameHandoff{}, fmt.Errorf("select character: %w", err)
	}

	channel, err := client.JoinChannel(ctx, "Diablo II")
	if err != nil || channel.Name == "" {
		return realm.CharacterSummary{}, realm.GameHandoff{}, fmt.Errorf("join channel: %w", err)
	}

	if _, err := client.SendMessage(ctx, "Realm acceptance joined the channel."); err != nil {
		return realm.CharacterSummary{}, realm.GameHandoff{}, fmt.Errorf("send channel message: %w", err)
	}

	handoff, err := client.CreateGame(ctx, realm.CreateGameRequest{
		Name:       "Acceptance Game",
		Difficulty: realm.DifficultyNormal,
		Maximum:    8,
		Expansion:  true,
	})
	if err != nil {
		return realm.CharacterSummary{}, realm.GameHandoff{}, fmt.Errorf("create named game: %w", err)
	}

	return created, handoff, nil
}

// validateCommittedCharacter ensures the worker advanced exactly one revision without changing character identity.
func validateCommittedCharacter(created, committed realm.CharacterSummary) error {
	if committed.Revision != created.Revision+1 || committed.Character.ID != created.Character.ID {
		return fmt.Errorf("unexpected committed character revision %d", committed.Revision)
	}

	return nil
}

// ensureGameDeparted confirms that committing the final player also removed the now-empty game from discovery.
func ensureGameDeparted(ctx context.Context, client *realm.RealmClient, gameID string) error {
	games, err := client.ListGames(ctx)
	if err != nil {
		return fmt.Errorf("list games after departure: %w", err)
	}

	if containsGame(games, gameID) {
		return errors.New("completed game remains discoverable")
	}

	return nil
}

// containsGame checks identity rather than display name because Realm game IDs are the lifecycle authority.
func containsGame(games []realm.GameDirectoryEntry, gameID string) bool {
	for _, game := range games {
		if game.GameID == gameID {
			return true
		}
	}

	return false
}

// verifyPersisted confirms that a prior session's committed revision survived a Realm restart.
func verifyPersisted(
	ctx context.Context,
	client *realm.RealmClient,
	session realm.RealmSession,
	characterName string,
) error {
	minimum, err := strconv.ParseUint(
		strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_MIN_REVISION")),
		10,
		64,
	)
	if err != nil || minimum == 0 {
		return errors.New("verify mode requires a positive minimum revision")
	}

	characters, err := client.ListCharacters(ctx)
	if err != nil {
		return err
	}

	for _, character := range characters {
		if character.Character.Name == characterName && character.Revision >= minimum {
			return writeAcceptanceResult(acceptanceResult{
				AccountID:         session.Account.ID,
				CharacterID:       character.Character.ID,
				CharacterRevision: character.Revision,
				Mode:              "verify",
			})
		}
	}

	return errors.New("committed character did not survive Realm restart")
}

// writeAcceptanceResult writes exactly one JSON record so shell callers can parse success without log filtering.
func writeAcceptanceResult(result acceptanceResult) error {
	return json.NewEncoder(os.Stdout).Encode(result)
}
