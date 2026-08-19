package envconfig

import (
	"fmt"
	"os"
)

// Load applies file values only when the process has not already exported a key.
func Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open environment file %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()
	values, err := Parse(file)
	if err != nil {
		return fmt.Errorf("parse environment file %q: %w", path, err)
	}
	for _, key := range sortedKeys(values) {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, values[key]); err != nil {
			return fmt.Errorf("set environment variable %q: %w", key, err)
		}
	}
	return nil
}
