package realm

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrOperatorAuthentication = errors.New("realm: operator authentication failed")

// LoadOrCreateOperatorToken owns the local operator capability. The raw token
// never enters PostgreSQL, logs, browser state, or player sessions.
func LoadOrCreateOperatorToken(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}

	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
			return "", errors.New("realm: operator token must be an owner-only regular file")
		}

		data, readErr := os.ReadFile(path)

		token := strings.TrimSpace(string(data))
		if readErr != nil || len(token) < 32 || len(token) > 4096 {
			return "", errors.New("realm: invalid operator token file")
		}

		return token, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	token := base64.RawURLEncoding.EncodeToString(data)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("realm: create operator token: %w", err)
	}

	if _, err = file.WriteString(token + "\n"); err == nil {
		err = file.Sync()
	}

	if closeErr := file.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		_ = os.Remove(path)
		return "", err
	}

	return token, nil
}
