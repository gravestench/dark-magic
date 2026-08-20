package envconfig

import (
	"os"
	"testing"
)

// preserveEnvironment prevents process-global environment mutations from leaking
// across tests and making later precedence assertions order-dependent.
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
