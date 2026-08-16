package serverapp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

const maximumRecoveryCheckpointBytes = 32 << 20

// ReadRecoveryCheckpoint reads the owner-only handoff prepared by the Realm
// process allocator. It is never accepted from a gameplay client.
func ReadGameRecovery(path string) (realm.GameRecovery, error) {
	file, err := os.Open(path)
	if err != nil {
		return realm.GameRecovery{}, fmt.Errorf("server: open game recovery: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return realm.GameRecovery{}, fmt.Errorf("server: stat game recovery: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return realm.GameRecovery{}, errors.New("server: game recovery must be an owner-only regular file")
	}
	payload, err := io.ReadAll(io.LimitReader(file, maximumRecoveryCheckpointBytes+1))
	if err != nil || len(payload) == 0 || len(payload) > maximumRecoveryCheckpointBytes {
		return realm.GameRecovery{}, errors.New("server: game recovery exceeds its bounded format")
	}
	var recovery realm.GameRecovery
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&recovery); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return realm.GameRecovery{}, errors.New("server: invalid game recovery encoding")
	}
	if err := realm.ValidateGameRecovery(recovery); err != nil {
		return realm.GameRecovery{}, fmt.Errorf("server: validate game recovery: %w", err)
	}
	return recovery, nil
}
