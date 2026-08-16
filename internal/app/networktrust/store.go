// Package networktrust persists direct-game host identity and client TOFU pins.
package networktrust

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const schemaVersion = 1

type pinsFile struct {
	Version int               `json:"version"`
	Hosts   map[string]string `json:"hosts"`
}

type Store struct {
	mu  sync.Mutex
	dir string
}

func New(dir string) (*Store, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("network trust: directory is required")
	}
	return &Store{dir: dir}, nil
}

func Directory(preferencesPath string) (string, error) {
	if preferencesPath != "" {
		return filepath.Join(filepath.Dir(preferencesPath), "network"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("network trust: user config directory: %w", err)
	}
	return filepath.Join(dir, "dark-magic", "network"), nil
}

func (store *Store) HostTLS() (*tls.Config, *tls.Config, string, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := os.MkdirAll(store.dir, 0o700); err != nil {
		return nil, nil, "", err
	}
	certPath, keyPath := filepath.Join(store.dir, "host-certificate.pem"), filepath.Join(store.dir, "host-identity.pem")
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if errors.Is(err, os.ErrNotExist) {
		certificate, err = generateIdentity(certPath, keyPath)
	}
	if err != nil {
		return nil, nil, "", fmt.Errorf("network trust: load host identity: %w", err)
	}
	if len(certificate.Certificate) == 0 {
		return nil, nil, "", errors.New("network trust: host certificate is empty")
	}
	parsed, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil || time.Now().After(parsed.NotAfter) {
		return nil, nil, "", errors.New("network trust: host certificate is invalid or expired")
	}
	sum := sha256.Sum256(certificate.Certificate[0])
	server := &tls.Config{Certificates: []tls.Certificate{certificate}}
	client := pinnedTLS(sum[:])
	return server, client, "sha256:" + hex.EncodeToString(sum[:]), nil
}

func (store *Store) ClientTLS(address string) (*tls.Config, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	pins, err := store.loadPins()
	if err != nil {
		return nil, err
	}
	expected := pins.Hosts[address]
	return &tls.Config{InsecureSkipVerify: true, VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return errors.New("network trust: server presented no certificate")
		}
		sum := sha256.Sum256(rawCerts[0])
		actual := hex.EncodeToString(sum[:])
		if expected != "" && expected != actual {
			return errors.New("network trust: host identity changed")
		}
		if expected == "" {
			store.mu.Lock()
			defer store.mu.Unlock()
			latest, err := store.loadPins()
			if err != nil {
				return err
			}
			if pinned := latest.Hosts[address]; pinned != "" && pinned != actual {
				return errors.New("network trust: host identity changed")
			}
			latest.Hosts[address] = actual
			return store.savePins(latest)
		}
		return nil
	}}, nil
}

func pinnedTLS(expected []byte) *tls.Config {
	return &tls.Config{InsecureSkipVerify: true, VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return errors.New("network trust: server presented no certificate")
		}
		actual := sha256.Sum256(rawCerts[0])
		if !strings.EqualFold(hex.EncodeToString(actual[:]), hex.EncodeToString(expected)) {
			return errors.New("network trust: host identity changed")
		}
		return nil
	}}
}

// PinnedTLSFingerprint creates a client configuration for a short-lived
// Realm-assigned worker. Unlike direct-host TOFU, it persists nothing: the
// trusted Realm assignment already carries the exact certificate identity.
func PinnedTLSFingerprint(fingerprint string) (*tls.Config, error) {
	const prefix = "sha256:"
	value := strings.ToLower(strings.TrimSpace(fingerprint))
	if !strings.HasPrefix(value, prefix) {
		return nil, errors.New("network trust: invalid TLS fingerprint")
	}
	expected, err := hex.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil || len(expected) != sha256.Size {
		return nil, errors.New("network trust: invalid TLS fingerprint")
	}
	return pinnedTLS(expected), nil
}

func generateIdentity(certPath, keyPath string) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	now := time.Now()
	template := x509.Certificate{SerialNumber: big.NewInt(now.UnixNano()), Subject: pkix.Name{CommonName: "Dark Magic Direct Host"}, NotBefore: now.Add(-time.Minute), NotAfter: now.AddDate(10, 0, 0), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	if err := atomicWrite(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0o600); err != nil {
		return tls.Certificate{}, err
	}
	if err := atomicWrite(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644); err != nil {
		return tls.Certificate{}, err
	}
	return tls.LoadX509KeyPair(certPath, keyPath)
}

func (store *Store) loadPins() (pinsFile, error) {
	path := filepath.Join(store.dir, "known-hosts.json")
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

func (store *Store) savePins(pins pinsFile) error {
	if err := os.MkdirAll(store.dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(pins, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(store.dir, "known-hosts.json"), append(data, '\n'), 0o600)
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".network-trust-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err = file.Chmod(mode); err == nil {
		_, err = file.Write(data)
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
