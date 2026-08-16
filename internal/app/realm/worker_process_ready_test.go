package realm

import (
	"errors"
	"os"
	"testing"
)

const readyTestFingerprint = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestWorkerProcessReadyRoundTrip(t *testing.T) {
	path := t.TempDir() + "/ready.json"
	want := WorkerProcessReady{GameID: "game", AllocationID: "allocation", ProcessID: os.Getpid(), ControlAddress: "127.0.0.1:1234",
		GameEndpoint: GameEndpoint{Address: "127.0.0.1:4321", TLSFingerprint: readyTestFingerprint}}
	if err := WriteWorkerProcessReady(path, want); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("ready file mode = %o", info.Mode().Perm())
	}
	got, err := ReadWorkerProcessReady(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != WorkerProcessReadyVersion || got.GameID != want.GameID || got.AllocationID != want.AllocationID ||
		got.ProcessID != want.ProcessID || got.ControlAddress != want.ControlAddress || got.GameEndpoint != want.GameEndpoint {
		t.Fatalf("ready = %#v", got)
	}
}

func TestWorkerProcessReadyRejectsUntrustedControlAndMalformedJSON(t *testing.T) {
	for name, payload := range map[string]string{
		"non-loopback":  `{"version":"RealmWorkerProcessReady/v2","game_id":"game","allocation_id":"allocation","process_id":1,"control_address":"192.0.2.1:1234","game_endpoint":{"address":"127.0.0.1:4321","tls_fingerprint":"` + readyTestFingerprint + `"}}`,
		"unknown field": `{"version":"RealmWorkerProcessReady/v2","game_id":"game","allocation_id":"allocation","process_id":1,"control_address":"127.0.0.1:1234","game_endpoint":{"address":"127.0.0.1:4321","tls_fingerprint":"` + readyTestFingerprint + `"},"extra":true}`,
		"trailing data": `{"version":"RealmWorkerProcessReady/v2","game_id":"game","allocation_id":"allocation","process_id":1,"control_address":"127.0.0.1:1234","game_endpoint":{"address":"127.0.0.1:4321","tls_fingerprint":"` + readyTestFingerprint + `"}} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := t.TempDir() + "/ready.json"
			if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := ReadWorkerProcessReady(path); !errors.Is(err, ErrWorkerProtocol) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
