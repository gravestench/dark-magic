package networktrust

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"strings"
)

// pinnedTLS creates a client configuration that accepts only the exact certificate hash supplied by its caller.
func pinnedTLS(expected []byte) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			actual, err := certificateFingerprint(rawCerts)
			if err != nil {
				return err
			}

			return verifyFingerprint(hex.EncodeToString(expected), actual)
		},
	}
}

// PinnedTLSFingerprint creates a client configuration for a short-lived Realm-assigned worker.
// It persists nothing because the trusted Realm assignment already carries the exact certificate identity.
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
