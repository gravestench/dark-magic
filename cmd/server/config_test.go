package main

import "testing"

// TestParseDifficultyDocumentsSupportedNames protects the CLI/legacy encoding
// boundary so human-friendly names remain stable while runtime values stay numeric.
func TestParseDifficultyDocumentsSupportedNames(t *testing.T) {
	t.Parallel()

	tests := map[string]int{"normal": 0, " Nightmare ": 1, "HELL": 2}
	for input, want := range tests {
		got, err := parseDifficulty(input)
		if err != nil {
			t.Errorf("parseDifficulty(%q): %v", input, err)
		} else if got != want {
			t.Errorf("parseDifficulty(%q) = %d, want %d", input, got, want)
		}
	}

	if _, err := parseDifficulty("torment"); err == nil {
		t.Fatal("parseDifficulty accepted an unsupported value")
	}
}

// TestServerConfigRequiresCompleteWorkerPolicy ensures a process cannot enter
// Realm-worker mode without every secret, identity, control, and recovery path.
func TestServerConfigRequiresCompleteWorkerPolicy(t *testing.T) {
	t.Parallel()

	config := serverConfig{realmWorker: true, gameMaximumPlayers: 8}
	if err := config.validate(); err == nil {
		t.Fatal("incomplete Realm worker configuration was accepted")
	}

	config.allocationID = "allocation"
	config.workerControlListen = "127.0.0.1:0"
	config.workerControlToken = "token"
	config.workerReadyFile = "ready"
	config.quicListen = "127.0.0.1:0"
	config.tlsCertificate = "certificate"
	config.tlsKey = "key"

	config.admissionKey = "admission"
	if err := config.validate(); err != nil {
		t.Fatalf("complete Realm worker configuration: %v", err)
	}
}

// TestServerConfigRejectsInvalidCapacity prevents invalid session capacity from
// entering runtime identity or being advertised to joining clients.
func TestServerConfigRejectsInvalidCapacity(t *testing.T) {
	t.Parallel()

	for _, capacity := range []int{0, 9} {
		config := serverConfig{gameMaximumPlayers: capacity}
		if err := config.validate(); err == nil {
			t.Errorf("gameMaximumPlayers %d was accepted", capacity)
		}
	}
}
