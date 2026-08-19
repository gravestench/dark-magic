package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gravestench/dark-magic/internal/logging"
)

// serverConfig freezes session identity, admission, transport, and game rules
// before authority starts. A running session must not observe mutable flag state.
type serverConfig struct {
	logLevel            slog.Level
	logLevelName        string
	sessionID           string
	allocationID        string
	quicListen          string
	tlsCertificate      string
	tlsKey              string
	admissionKey        string
	playerProfile       string
	profilePlayer       string
	profileX            float64
	profileY            float64
	profileWidth        float64
	profileHeight       float64
	profileAct          int64
	profileLevel        int64
	remoteProfileKey    string
	realmWorker         bool
	workerControlListen string
	workerControlToken  string
	workerReadyFile     string
	restoreCheckpoint   string
	mods                string
	gameDifficulty      int
	gameHardcore        bool
	gameLadder          bool
	gameMaximumPlayers  int
}

// parseServerConfig rejects incomplete worker/admission combinations before
// content is mounted or an authoritative session is created.
func parseServerConfig(defaultEnvironmentPath string) (serverConfig, error) {
	config := serverConfig{
		logLevelName:       environmentDefault("DARK_MAGIC_SERVER_LOG_LEVEL", "info"),
		sessionID:          "standalone",
		mods:               os.Getenv("DARK_MAGIC_MODS"),
		gameMaximumPlayers: 8,
	}
	gameDifficulty := "normal"
	registerServerFlags(&config, &gameDifficulty, defaultEnvironmentPath)
	flag.Parse()

	level, err := logging.ParseLevel(config.logLevelName)
	if err != nil {
		return serverConfig{}, err
	}

	config.logLevel = level

	config.gameDifficulty, err = parseDifficulty(gameDifficulty)
	if err != nil {
		return serverConfig{}, err
	}

	if err := config.validate(); err != nil {
		return serverConfig{}, err
	}

	return config, nil
}

// registerServerFlags separates session, profile, worker, and game-rule policy
// so each trust boundary can be reviewed without scanning unrelated flags.
func registerServerFlags(
	config *serverConfig,
	gameDifficulty *string,
	defaultEnvironmentPath string,
) {
	_ = flag.String("env-file", defaultEnvironmentPath, "environment file")

	registerSessionFlags(config)
	registerProfileFlags(config)
	registerWorkerFlags(config)
	registerGameRuleFlags(config, gameDifficulty)
}

// registerSessionFlags defines process-wide resources whose values determine
// runtime identity and network compatibility for the entire session.
func registerSessionFlags(config *serverConfig) {
	flag.StringVar(&config.logLevelName, "log-level", config.logLevelName, "log verbosity")
	flag.StringVar(&config.sessionID, "session-id", config.sessionID, "game-session ID")
	flag.StringVar(&config.allocationID, "allocation-id", "", "Realm allocation generation")
	flag.StringVar(&config.quicListen, "quic-listen", "", "authenticated game UDP address")
	flag.StringVar(&config.tlsCertificate, "tls-cert", "", "QUIC certificate")
	flag.StringVar(&config.tlsKey, "tls-key", "", "QUIC private key")
	flag.StringVar(&config.admissionKey, "admission-key", "", "admission HMAC key file")
	flag.StringVar(&config.mods, "mods", config.mods, "temporary comma-separated mod IDs")
}

// registerProfileFlags defines the opt-in self-host identity and spawn location;
// leaving these blank keeps a standalone authority from inventing a player.
func registerProfileFlags(config *serverConfig) {
	flag.StringVar(&config.playerProfile, "player-profile", "", "self-hosted player profile")
	flag.StringVar(&config.profilePlayer, "profile-player", "", "selected profile player ID")
	flag.Float64Var(&config.profileX, "profile-x", 0, "profile spawn X")
	flag.Float64Var(&config.profileY, "profile-y", 0, "profile spawn Y")
	flag.Float64Var(&config.profileWidth, "profile-world-width", 0, "profile world width")
	flag.Float64Var(&config.profileHeight, "profile-world-height", 0, "profile world height")
	flag.Int64Var(&config.profileAct, "profile-act", 0, "profile act")
	flag.Int64Var(&config.profileLevel, "profile-level", 0, "profile level ID")
	flag.StringVar(&config.remoteProfileKey, "remote-profile-key", "", "remote profile key")
}

// registerWorkerFlags defines the private contract between a Realm and its child
// server. The settings are validated as a complete set because partial supervision is unsafe.
func registerWorkerFlags(config *serverConfig) {
	flag.BoolVar(&config.realmWorker, "realm-worker", false, "run as a Realm worker")
	flag.StringVar(&config.workerControlListen, "worker-control-listen", "", "worker control address")
	flag.StringVar(&config.workerControlToken, "worker-control-token", "", "worker bearer token")
	flag.StringVar(&config.workerReadyFile, "worker-ready-file", "", "worker readiness file")
	flag.StringVar(&config.restoreCheckpoint, "restore-checkpoint", "", "Realm recovery checkpoint")
}

// registerGameRuleFlags defines rules embedded in runtime identity and therefore
// immutable after clients begin comparing and joining the session.
func registerGameRuleFlags(config *serverConfig, difficulty *string) {
	flag.StringVar(difficulty, "game-difficulty", *difficulty, "normal, nightmare, or hell")
	flag.BoolVar(&config.gameHardcore, "game-hardcore", false, "use Hardcore rules")
	flag.BoolVar(&config.gameLadder, "game-ladder", false, "use Ladder eligibility rules")
	flag.IntVar(&config.gameMaximumPlayers, "game-maximum-players", 8, "game capacity from 1 through 8")
}

// parseDifficulty contains legacy numeric encoding at the command boundary so
// the rest of startup can reason about a validated integer rule.
func parseDifficulty(value string) (int, error) {
	difficulties := map[string]int{"normal": 0, "nightmare": 1, "hell": 2}

	difficulty, found := difficulties[strings.ToLower(strings.TrimSpace(value))]
	if !found {
		return 0, fmt.Errorf("invalid game difficulty %q", value)
	}

	return difficulty, nil
}

// validate enforces cross-flag invariants that determine which admission and
// supervision model the process is allowed to enter.
func (config serverConfig) validate() error {
	if config.gameMaximumPlayers < 1 || config.gameMaximumPlayers > 8 {
		return errors.New("game-maximum-players must be from 1 through 8")
	}

	if config.workerConfigured() && !config.completeWorkerConfig() {
		return errors.New("realm worker, control, readiness, QUIC, TLS, and admission flags must be set together")
	}

	if config.workerConfigured() && (config.playerProfile != "" || config.remoteProfileKey != "") {
		return errors.New("realm workers cannot admit player-controlled profiles")
	}

	if config.restoreCheckpoint != "" && !config.realmWorker {
		return errors.New("restore-checkpoint is valid only for Realm workers")
	}

	return nil
}

// workerConfigured detects intent to run under Realm supervision. Even one
// worker-only value switches validation to the stricter all-fields-required contract.
func (config serverConfig) workerConfigured() bool {
	return config.realmWorker || config.workerControlListen != "" ||
		config.workerControlToken != "" || config.workerReadyFile != ""
}

// completeWorkerConfig protects the readiness/ticket/recovery handshake from
// starting with missing secrets or rendezvous paths.
func (config serverConfig) completeWorkerConfig() bool {
	return config.realmWorker && config.allocationID != "" &&
		config.workerControlListen != "" && config.workerControlToken != "" &&
		config.workerReadyFile != "" && config.quicListen != "" &&
		config.tlsCertificate != "" && config.tlsKey != "" && config.admissionKey != ""
}

// environmentDefault treats whitespace-only exports as absent so an accidental
// empty assignment cannot erase a safe command default.
func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return fallback
}
