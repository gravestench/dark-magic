package networktrust

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHostIdentityPersistsWithPrivatePermissions verifies repeat loads preserve identity and private-key protection.
func TestHostIdentityPersistsWithPrivatePermissions(t *testing.T) {
	directory := t.TempDir()
	store := newTestStore(t, directory)

	_, _, first, err := store.HostTLS()
	if err != nil {
		t.Fatal(err)
	}

	_, _, second, err := store.HostTLS()
	if err != nil {
		t.Fatal(err)
	}

	if first != second {
		t.Fatalf("host identity changed: %q != %q", first, second)
	}

	info, err := os.Stat(filepath.Join(directory, hostIdentityFilename))
	if err != nil {
		t.Fatal(err)
	}
	// Windows preserves private access through ACLs but does not expose Unix
	// permission bits through FileMode, so the mode assertion is Unix-specific.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("private key mode = %o", info.Mode().Perm())
	}
}

// TestClientTrustOnFirstUsePersistsAndRejectsChangedIdentity exercises the complete direct-host TOFU lifecycle.
func TestClientTrustOnFirstUsePersistsAndRejectsChangedIdentity(t *testing.T) {
	clientStore := newTestStore(t, filepath.Join(t.TempDir(), "client"))
	firstServer := newTestHostTLS(t, filepath.Join(t.TempDir(), "first"))
	secondServer := newTestHostTLS(t, filepath.Join(t.TempDir(), "second"))
	address := "192.0.2.10:6112"

	client, err := clientStore.ClientTLS(address)
	if err != nil {
		t.Fatal(err)
	}

	if err := verify(client, firstServer); err != nil {
		t.Fatalf("first use: %v", err)
	}

	assertAddressWasPinned(t, clientStore, address)

	reloaded := newTestStore(t, clientStore.dir)

	client, err = reloaded.ClientTLS(address)
	if err != nil {
		t.Fatal(err)
	}

	if err := verify(client, firstServer); err != nil {
		t.Fatalf("known identity: %v", err)
	}

	if err := verify(client, secondServer); err == nil || !strings.Contains(err.Error(), "identity changed") {
		t.Fatalf("changed identity error = %v", err)
	}
}

// assertAddressWasPinned confirms a successful first handshake was durably recorded for subsequent clients.
func assertAddressWasPinned(t *testing.T, store *Store, address string) {
	t.Helper()

	pins, err := os.ReadFile(filepath.Join(store.dir, knownHostsFilename))
	if err != nil || !strings.Contains(string(pins), address) {
		t.Fatalf("saved pins=%q error=%v", pins, err)
	}
}

// newTestStore constructs a store and makes setup failures immediately diagnostic to the calling test.
func newTestStore(t *testing.T, directory string) *Store {
	t.Helper()

	store, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}

	return store
}

// newTestHostTLS creates an independent host identity so tests can compare stable and changed peers.
func newTestHostTLS(t *testing.T, directory string) *tls.Config {
	t.Helper()

	server, _, _, err := newTestStore(t, directory).HostTLS()
	if err != nil {
		t.Fatal(err)
	}

	return server
}

// TestMalformedTrustFilesFailClosed ensures corrupt trust state is surfaced instead of replaced or accepted.
func TestMalformedTrustFilesFailClosed(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, knownHostsFilename), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}

	store := newTestStore(t, directory)
	if _, err := store.ClientTLS("localhost:6112"); err == nil {
		t.Fatal("malformed pins accepted")
	}

	if err := os.WriteFile(filepath.Join(directory, hostCertificateFilename), []byte("bad"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(directory, hostIdentityFilename), []byte("bad"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := store.HostTLS(); err == nil {
		t.Fatal("malformed host identity replaced")
	}
}

// TestPinnedTLSFingerprintRejectsMalformedValues verifies only complete SHA-256 assignments produce TLS configs.
func TestPinnedTLSFingerprintRejectsMalformedValues(t *testing.T) {
	valid := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	config, err := PinnedTLSFingerprint(valid)
	if err != nil || config == nil || config.VerifyPeerCertificate == nil {
		t.Fatalf("config=%#v error=%v", config, err)
	}

	for _, value := range []string{"", "sha256:short", "md5:" + valid,
		"sha256:gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg"} {
		if _, err := PinnedTLSFingerprint(value); err == nil {
			t.Fatalf("malformed fingerprint %q was accepted", value)
		}
	}
}

// verify invokes the client callback with the server's leaf chain, avoiding a full network handshake in unit tests.
func verify(client, server *tls.Config) error {
	return client.VerifyPeerCertificate(server.Certificates[0].Certificate, nil)
}
