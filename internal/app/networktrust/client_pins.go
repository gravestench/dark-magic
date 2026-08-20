package networktrust

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	schemaVersion      = 1
	knownHostsFilename = "known-hosts.json"
)

type pinsFile struct {
	Version int               `json:"version"`
	Hosts   map[string]string `json:"hosts"`
}

// ClientTLS returns a TOFU client configuration for address, retaining the pin observed when the config was created.
func (store *Store) ClientTLS(address string) (*tls.Config, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	pins, err := store.loadPins()
	if err != nil {
		return nil, err
	}

	expected := pins.Hosts[address]

	return &tls.Config{
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			actual, err := certificateFingerprint(rawCerts)
			if err != nil {
				return err
			}

			if expected != "" {
				return verifyFingerprint(expected, actual)
			}

			return store.trustFirstFingerprint(address, actual)
		},
	}, nil
}

// trustFirstFingerprint atomically rechecks and stores a first-use pin so concurrent handshakes cannot overwrite it.
func (store *Store) trustFirstFingerprint(address, actual string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	latest, err := store.loadPins()
	if err != nil {
		return err
	}

	if pinned := latest.Hosts[address]; pinned != "" {
		return verifyFingerprint(pinned, actual)
	}

	latest.Hosts[address] = actual

	return store.savePins(latest)
}

// loadPins returns an empty current-schema store when no pins exist and rejects malformed persisted state.
func (store *Store) loadPins() (pinsFile, error) {
	path := filepath.Join(store.dir, knownHostsFilename)

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return pinsFile{Version: schemaVersion, Hosts: map[string]string{}}, nil
	}

	if err != nil {
		return pinsFile{}, err
	}

	var pins pinsFile
	if err := json.Unmarshal(data, &pins); err != nil || pins.Version != schemaVersion || pins.Hosts == nil {
		return pinsFile{}, errors.New("network trust: malformed known-hosts file")
	}

	return pins, nil
}

// savePins writes a complete snapshot so readers never observe a partially updated trust file.
func (store *Store) savePins(pins pinsFile) error {
	if err := os.MkdirAll(store.dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(pins, "", "  ")
	if err != nil {
		return err
	}

	return atomicWrite(filepath.Join(store.dir, knownHostsFilename), append(data, '\n'), 0o600)
}

// certificateFingerprint hashes the leaf certificate exactly as presented by the peer.
func certificateFingerprint(rawCerts [][]byte) (string, error) {
	if len(rawCerts) == 0 {
		return "", errors.New("network trust: server presented no certificate")
	}

	sum := sha256.Sum256(rawCerts[0])

	return hex.EncodeToString(sum[:]), nil
}

// verifyFingerprint fails closed when the peer certificate no longer matches the trusted identity.
func verifyFingerprint(expected, actual string) error {
	if expected != actual {
		return errors.New("network trust: host identity changed")
	}

	return nil
}
