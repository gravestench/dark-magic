package envconfig

import (
	"os"
	"testing"
)

// preserveEnvironment restores variables changed indirectly by loading a file.
func preserveEnvironment(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		value, found := os.LookupEnv(key)
		t.Cleanup(func() {
			if found {
				_ = os.Setenv(key, value)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}
