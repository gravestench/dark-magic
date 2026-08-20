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
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	hostCertificateFilename = "host-certificate.pem"
	hostIdentityFilename    = "host-identity.pem"
)

// HostTLS loads or creates the host identity and returns matching server and pinned client configurations.
func (store *Store) HostTLS() (*tls.Config, *tls.Config, string, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	certificate, err := store.loadOrCreateHostIdentity()
	if err != nil {
		return nil, nil, "", err
	}

	if err := validateHostCertificate(certificate); err != nil {
		return nil, nil, "", err
	}

	fingerprint := sha256.Sum256(certificate.Certificate[0])
	server := &tls.Config{Certificates: []tls.Certificate{certificate}}
	client := pinnedTLS(fingerprint[:])

	return server, client, "sha256:" + hex.EncodeToString(fingerprint[:]), nil
}

// loadOrCreateHostIdentity creates an identity only when either identity path does not exist.
// Other load failures remain fatal so corrupt credentials are never silently replaced.
func (store *Store) loadOrCreateHostIdentity() (tls.Certificate, error) {
	if err := os.MkdirAll(store.dir, 0o700); err != nil {
		return tls.Certificate{}, err
	}

	certPath := filepath.Join(store.dir, hostCertificateFilename)
	keyPath := filepath.Join(store.dir, hostIdentityFilename)

	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if errors.Is(err, os.ErrNotExist) {
		certificate, err = generateIdentity(certPath, keyPath)
	}

	if err != nil {
		return tls.Certificate{}, fmt.Errorf("network trust: load host identity: %w", err)
	}

	return certificate, nil
}

// validateHostCertificate rejects empty, malformed, or expired identities before they are used for TLS.
func validateHostCertificate(certificate tls.Certificate) error {
	if len(certificate.Certificate) == 0 {
		return errors.New("network trust: host certificate is empty")
	}

	parsed, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil || time.Now().After(parsed.NotAfter) {
		return errors.New("network trust: host certificate is invalid or expired")
	}

	return nil
}

// generateIdentity creates a ten-year self-signed host identity and persists the private key with restricted access.
func generateIdentity(certPath, keyPath string) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: big.NewInt(now.UnixNano()),
		Subject:      pkix.Name{CommonName: "Dark Magic Direct Host"},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if err := atomicWrite(keyPath, privateKeyPEM, 0o600); err != nil {
		return tls.Certificate{}, err
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER})
	if err := atomicWrite(certPath, certificatePEM, 0o644); err != nil {
		return tls.Certificate{}, err
	}

	return tls.LoadX509KeyPair(certPath, keyPath)
}
