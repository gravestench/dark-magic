package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/envconfig"
	"github.com/gravestench/dark-magic/internal/logging"
)

// realmConfig freezes public process policy before repositories, listeners, or
// worker supervision start. This prevents long-lived services from consulting
// mutable flag globals while handling requests.
type realmConfig struct {
	logLevel                slog.Level
	logLevelName            string
	listenAddress           string
	dataDirectory           string
	postgresURL             string
	accountBaseURL          string
	accountMailMode         string
	smtpAddress             string
	smtpFrom                string
	smtpUsername            string
	smtpPassword            string
	smtpRequireTLS          bool
	workerExecutable        string
	workerControlListen     string
	workerGameListen        string
	workerGameAdvertiseHost string
	workerHealthInterval    time.Duration
	checkpointInterval      time.Duration
	presenceTimeout         time.Duration
	operatorTokenFile       string
	operatorListen          string
	adminShell              bool
}

// parseRealmConfig establishes the complete operator-facing CLI before resource
// acquisition, allowing invalid cross-flag policy to fail without opening stores or ports.
func parseRealmConfig(defaultEnvironmentPath string) (realmConfig, error) {
	checkpointInterval, err := envconfig.Duration("DARK_MAGIC_REALM_CHECKPOINT_INTERVAL", 15*time.Second)
	if err != nil {
		return realmConfig{}, err
	}

	presenceTimeout, err := envconfig.Duration("DARK_MAGIC_REALM_PRESENCE_TIMEOUT", 30*time.Second)
	if err != nil {
		return realmConfig{}, err
	}

	config := realmConfig{
		accountBaseURL:       environmentDefault("DARK_MAGIC_REALM_ACCOUNT_URL", "https://accounts.dark-magic.test"),
		smtpFrom:             environmentDefault("DARK_MAGIC_REALM_SMTP_FROM", "realm@dark-magic.test"),
		smtpPassword:         os.Getenv("DARK_MAGIC_REALM_SMTP_PASSWORD"),
		checkpointInterval:   checkpointInterval,
		presenceTimeout:      presenceTimeout,
		workerHealthInterval: 2 * time.Second,
		workerControlListen:  environmentDefault("DARK_MAGIC_REALM_WORKER_CONTROL_LISTEN", "127.0.0.1:0"),
		workerGameListen:     environmentDefault("DARK_MAGIC_REALM_WORKER_GAME_LISTEN", "0.0.0.0:0"),
	}
	config.accountMailMode = defaultMailMode()
	registerRealmFlags(&config, defaultEnvironmentPath)
	flag.Parse()

	config.logLevel, err = logging.ParseLevel(config.logLevelName)

	return config, err
}

// registerRealmFlags groups settings by operational responsibility so reviewers
// can reason about public service, mail, worker, and private-operator exposure separately.
func registerRealmFlags(config *realmConfig, defaultEnvironmentPath string) {
	_ = flag.String("env-file", defaultEnvironmentPath, "environment file")

	registerCoreFlags(config)
	registerMailFlags(config)
	registerWorkerFlags(config)
	registerOperatorFlags(config)
}

// registerCoreFlags defines identity and durability policy shared by every Realm
// subsystem; these values must remain stable for the lifetime of issued sessions.
func registerCoreFlags(config *realmConfig) {
	flag.StringVar(
		&config.logLevelName,
		"log-level",
		environmentDefault("DARK_MAGIC_REALM_LOG_LEVEL", "info"),
		"log verbosity",
	)
	flag.StringVar(
		&config.listenAddress,
		"listen",
		environmentDefault("DARK_MAGIC_REALM_LISTEN", ":6112"),
		"authenticated Realm API address",
	)
	flag.StringVar(&config.dataDirectory, "data-dir", os.Getenv("DARK_MAGIC_REALM_DATA"), "durable Realm data directory")
	flag.StringVar(
		&config.postgresURL,
		"postgres-url",
		os.Getenv("DARK_MAGIC_REALM_POSTGRES_URL"),
		"PostgreSQL connection string",
	)
	flag.DurationVar(
		&config.checkpointInterval,
		"checkpoint-interval",
		config.checkpointInterval,
		"minimum checkpoint interval",
	)
	flag.DurationVar(&config.presenceTimeout, "presence-timeout", config.presenceTimeout, "unresponsive presence timeout")
}

// registerMailFlags defines how security-sensitive account links leave the
// process. Keeping delivery policy together makes unsafe mode changes conspicuous.
func registerMailFlags(config *realmConfig) {
	flag.StringVar(&config.accountBaseURL, "account-base-url", config.accountBaseURL, "public account-mail origin")
	flag.StringVar(
		&config.accountMailMode,
		"account-mail-mode",
		config.accountMailMode,
		"disabled, smtp, log, or auto-verify",
	)
	flag.StringVar(&config.smtpAddress, "smtp-address", os.Getenv("DARK_MAGIC_REALM_SMTP_ADDRESS"), "SMTP host and port")
	flag.StringVar(&config.smtpFrom, "smtp-from", config.smtpFrom, "transactional sender address")
	flag.StringVar(&config.smtpUsername, "smtp-username", os.Getenv("DARK_MAGIC_REALM_SMTP_USERNAME"), "SMTP username")
	flag.BoolVar(&config.smtpRequireTLS, "smtp-require-tls", false, "require SMTP STARTTLS")
}

// registerWorkerFlags defines the allocator and health policy that turn Realm
// game records into supervised child processes rather than unmanaged servers.
func registerWorkerFlags(config *realmConfig) {
	flag.StringVar(
		&config.workerExecutable,
		"worker-executable",
		os.Getenv("DARK_MAGIC_REALM_WORKER"),
		"game-worker executable",
	)
	flag.StringVar(
		&config.workerControlListen,
		"worker-control-listen",
		config.workerControlListen,
		"private worker-control address",
	)
	flag.StringVar(&config.workerGameListen, "worker-game-listen", config.workerGameListen, "worker UDP listen address")
	flag.StringVar(
		&config.workerGameAdvertiseHost,
		"worker-game-advertise-host",
		os.Getenv("DARK_MAGIC_REALM_GAME_HOST"),
		"client-reachable worker host",
	)
	flag.DurationVar(
		&config.workerHealthInterval,
		"worker-health-interval",
		config.workerHealthInterval,
		"worker health interval",
	)
}

// registerOperatorFlags defines mutation-capable administration surfaces.
// These are intentionally separate from the public API because their exposure model differs.
func registerOperatorFlags(config *realmConfig) {
	flag.StringVar(
		&config.operatorTokenFile,
		"operator-token-file",
		os.Getenv("DARK_MAGIC_REALM_OPERATOR_TOKEN_FILE"),
		"operator token file",
	)
	flag.StringVar(
		&config.operatorListen,
		"operator-listen",
		os.Getenv("DARK_MAGIC_REALM_OPERATOR_LISTEN"),
		"private operator address",
	)
	flag.BoolVar(&config.adminShell, "admin-shell", true, "enable the administration shell")
}

// defaultMailMode prefers explicit policy and otherwise infers SMTP only when a
// server is configured. The final fallback keeps links local instead of discarding them.
func defaultMailMode() string {
	if configured := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE")); configured != "" {
		return configured
	}

	if os.Getenv("DARK_MAGIC_REALM_SMTP_ADDRESS") != "" {
		return "smtp"
	}

	return "disabled"
}

// environmentDefault treats whitespace-only exports as absent so accidental
// empty shell values cannot erase a documented operational default.
func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return fallback
}
