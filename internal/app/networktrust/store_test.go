package networktrust

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHostIdentityPersistsWithPrivatePermissions(t *testing.T) {
	directory := t.TempDir()
	store, _ := New(directory)
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
	info, err := os.Stat(filepath.Join(directory, "host-identity.pem"))
	if err != nil {
		t.Fatal(err)
	}
	// Windows preserves private access through ACLs but does not expose Unix
	// permission bits through FileMode, so the mode assertion is Unix-specific.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("private key mode = %o", info.Mode().Perm())
	}
}

func TestClientTrustOnFirstUsePersistsAndRejectsChangedIdentity(t *testing.T) {
	clientStore, _ := New(filepath.Join(t.TempDir(), "client"))
	firstHost, _ := New(filepath.Join(t.TempDir(), "first"))
	secondHost, _ := New(filepath.Join(t.TempDir(), "second"))
	firstServer, _, _, err := firstHost.HostTLS()
	if err != nil {
		t.Fatal(err)
	}
	secondServer, _, _, err := secondHost.HostTLS()
	if err != nil {
		t.Fatal(err)
	}
	address := "192.0.2.10:6112"
	client, err := clientStore.ClientTLS(address)
	if err != nil {
		t.Fatal(err)
	}
	if err := verify(client, firstServer); err != nil {
		t.Fatalf("first use: %v", err)
	}
	pins, err := os.ReadFile(filepath.Join(clientStore.dir, "known-hosts.json"))
	if err != nil || !strings.Contains(string(pins), address) {
		t.Fatalf("saved pins=%q error=%v", pins, err)
	}
	reloaded, _ := New(clientStore.dir)
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

func TestMalformedTrustFilesFailClosed(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "known-hosts.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	store, _ := New(directory)
	if _, err := store.ClientTLS("localhost:6112"); err == nil {
		t.Fatal("malformed pins accepted")
	}
	if err := os.WriteFile(filepath.Join(directory, "host-certificate.pem"), []byte("bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "host-identity.pem"), []byte("bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.HostTLS(); err == nil {
		t.Fatal("malformed host identity replaced")
	}
}

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

func verify(client, server *tls.Config) error {
	return client.VerifyPeerCertificate(server.Certificates[0].Certificate, nil)
}
