package networktrust

import (
	"os"
	"path/filepath"
)

// atomicWrite replaces path only after all bytes are written and closed, preserving the previous file on failure.
func atomicWrite(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".network-trust-*")
	if err != nil {
		return err
	}

	temporary := file.Name()
	defer func() {
		// Rename removes the temporary path on success; cleanup failure cannot improve an earlier write error.
		_ = os.Remove(temporary)
	}()

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
