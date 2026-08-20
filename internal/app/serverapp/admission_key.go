package serverapp

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const maximumAdmissionKeyBytes = 4096

// ReadAdmissionKey reads an owner-protected shared secret with a strict size
// bound. The single trailing newline allowance supports ordinary secret files
// without weakening the minimum entropy requirement.
func ReadAdmissionKey(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("server: open admission key: %w", err)
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("server: stat admission key: %w", err)
	}
	// Group or world access would expose the credential used to mint admission
	// and worker-control authority.
	if info.Mode().Perm()&0o077 != 0 {
		return nil, errors.New("server: admission key must not be accessible by group or others")
	}

	data, err := io.ReadAll(io.LimitReader(file, maximumAdmissionKeyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("server: read admission key: %w", err)
	}

	if len(data) > maximumAdmissionKeyBytes {
		return nil, errors.New("server: admission key exceeds 4096 bytes")
	}

	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	if len(data) < 32 {
		return nil, errors.New("server: admission key must contain at least 32 bytes")
	}

	return data, nil
}
