package realm_scripts_test

import (
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var publicScripts = []string{
	"up.sh",
	"down.sh",
	"fresh-install.sh",
	"drain-game.sh",
	"mailpit-up.sh",
	"mailpit-down.sh",
	"test-production.sh",
}

// TestRealmScriptsHaveValidShellSyntaxAndHelp keeps every public entry point usable by POSIX sh and documents its
// operator-facing help without requiring external services.
func TestRealmScriptsHaveValidShellSyntaxAndHelp(t *testing.T) {
	for _, name := range publicScripts {
		t.Run(name, func(t *testing.T) {
			for _, arguments := range [][]string{{"-n", name}, {name, "--help"}} {
				command := exec.Command("sh", arguments...)

				command.Env = append(os.Environ(), "DARK_MAGIC_CONFIG_DIR="+t.TempDir())

				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("sh %s: %v\n%s", strings.Join(arguments, " "), err, output)
				}
			}
		})
	}
}

// TestFreshInstallRejectsMissingPostgresWithoutTouchingOperationalState verifies validation precedes every destructive
// filesystem action, so a configuration mistake cannot archive or replace live account data.
func TestFreshInstallRejectsMissingPostgresWithoutTouchingOperationalState(t *testing.T) {
	runtimeDirectory := filepath.Join(t.TempDir(), "runtime")
	dataDirectory := filepath.Join(runtimeDirectory, "data")

	if err := os.MkdirAll(dataDirectory, 0o700); err != nil {
		t.Fatal(err)
	}

	accountFile := filepath.Join(dataDirectory, "accounts.json")
	if err := os.WriteFile(accountFile, []byte("precious"), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("sh", "fresh-install.sh", "--yes", "--no-start")

	command.Env = append(os.Environ(),
		"DARK_MAGIC_CONFIG_DIR="+filepath.Join(t.TempDir(), "config"),
		"DARK_MAGIC_REALM_RUNTIME_DIR="+runtimeDirectory,
		"DARK_MAGIC_REALM_DATA="+dataDirectory,
		"DARK_MAGIC_REALM_POSTGRES_URL=",
		"DARK_MAGIC_REALM_MANAGE_POSTGRES=0",
	)

	output, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "requires DARK_MAGIC_REALM_POSTGRES_URL") {
		t.Fatalf("fresh install error=%v output=%s", err, output)
	}

	data, err := os.ReadFile(accountFile)
	if err != nil || string(data) != "precious" {
		t.Fatalf("operational state changed: data=%q error=%v", data, err)
	}
}

// TestBringupPrintConfigDoesNotExposePostgresURL protects credentials embedded in the database URL while still
// requiring enough output for an operator to confirm the selected storage backend.
func TestBringupPrintConfigDoesNotExposePostgresURL(t *testing.T) {
	runtimeDirectory := filepath.Join(t.TempDir(), "runtime")
	secretURL := "postgres://realm:secret@127.0.0.1:5432/realm?sslmode=disable"
	command := exec.Command("sh", "up.sh", "--print-config")

	command.Env = append(os.Environ(),
		"DARK_MAGIC_CONFIG_DIR="+filepath.Join(t.TempDir(), "config"),
		"DARK_MAGIC_REALM_RUNTIME_DIR="+runtimeDirectory,
		"DARK_MAGIC_REALM_POSTGRES_URL="+secretURL,
	)

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("print config: %v\n%s", err, output)
	}

	if strings.Contains(string(output), secretURL) || !strings.Contains(string(output), "storage=postgresql") {
		t.Fatalf("unsafe or incomplete print config: %s", output)
	}
}

// TestBringupRejectsRealmAlreadyServingWithoutOwnedPID proves endpoint availability is not mistaken for ownership;
// the launcher must never adopt a Realm that was not recorded in its runtime PID file.
func TestBringupRejectsRealmAlreadyServingWithoutOwnedPID(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := statusServer()
	go func() {
		// Closing the server during cleanup is an expected exit path, so Serve's error is intentionally ignored.
		_ = server.Serve(listener)
	}()

	t.Cleanup(func() {
		_ = server.Close()
	})

	runtimeDirectory := filepath.Join(t.TempDir(), "runtime")
	command := exec.Command("sh", "up.sh", "--no-build")

	command.Env = append(os.Environ(),
		"DARK_MAGIC_CONFIG_DIR="+filepath.Join(t.TempDir(), "config"),
		"DARK_MAGIC_REALM_RUNTIME_DIR="+runtimeDirectory,
		"DARK_MAGIC_REALM_LISTEN="+listener.Addr().String(),
		"DARK_MAGIC_REALM_URL=http://"+listener.Addr().String(),
		"DARK_MAGIC_REALM_POSTGRES_URL=postgres://realm@127.0.0.1:5432/realm?sslmode=disable",
		"DARK_MAGIC_REALM_MANAGE_POSTGRES=0",
	)

	output, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "already serving a Realm not owned by this runtime") {
		t.Fatalf("bringup error=%v output=%s", err, output)
	}
}

// statusServer supplies only the readiness route needed to model an unowned Realm without involving application code.
func statusServer() *http.Server {
	return &http.Server{Handler: http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/status" {
			http.NotFound(response, request)

			return
		}

		response.WriteHeader(http.StatusOK)
	})}
}

// TestLifecycleLoaderAcceptsExportedDotenvEntries preserves compatibility with dotenv files that use an optional
// export prefix, while ensuring the file value wins only when the process did not already export the setting.
func TestLifecycleLoaderAcceptsExportedDotenvEntries(t *testing.T) {
	environmentFile := filepath.Join(t.TempDir(), "realm.env")
	environment := strings.Join([]string{
		"export DARK_MAGIC_REALM_LISTEN=127.0.0.1:16222",
		"DARK_MAGIC_REALM_POSTGRES_URL=postgres://realm@127.0.0.1:5432/realm?sslmode=disable",
		"",
	}, "\n")

	if err := os.WriteFile(environmentFile, []byte(environment), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("sh", "-c", `. ./common.sh; printf '%s' "$REALM_LISTEN"`)

	command.Env = append(environmentWithout("DARK_MAGIC_REALM_LISTEN"),
		"DARK_MAGIC_CONFIG_DIR="+filepath.Join(t.TempDir(), "config"),
		"DARK_MAGIC_REALM_ENV_FILE="+environmentFile,
	)

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("load exported dotenv entry: %v\n%s", err, output)
	}

	if string(output) != "127.0.0.1:16222" {
		t.Fatalf("listener=%q, want exported file value", output)
	}
}

// environmentWithout copies the process environment minus one key so tests can exercise dotenv fallback semantics
// without mutating the parent test process or leaking state between cases.
func environmentWithout(name string) []string {
	prefix := name + "="
	result := make([]string, 0, len(os.Environ()))

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, prefix) {
			result = append(result, entry)
		}
	}

	return result
}
