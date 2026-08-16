// Command realm is the realm control-plane composition root.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gravestench/dark-magic/internal/app/envconfig"
	"github.com/gravestench/dark-magic/internal/app/headlessshell"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/realmportal"
	portalassets "github.com/gravestench/dark-magic/internal/app/realmportal/assets"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/logging"
	"github.com/gravestench/dark-magic/internal/shell"
)

func main() {
	environment, err := envconfig.Bootstrap("realm", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defaultAccountURL := os.Getenv("DARK_MAGIC_REALM_ACCOUNT_URL")
	if defaultAccountURL == "" {
		defaultAccountURL = "https://accounts.dark-magic.test"
	}
	defaultSMTPFrom := os.Getenv("DARK_MAGIC_REALM_SMTP_FROM")
	if defaultSMTPFrom == "" {
		defaultSMTPFrom = "realm@dark-magic.test"
	}
	defaultMailMode := os.Getenv("DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE")
	if defaultMailMode == "" {
		if os.Getenv("DARK_MAGIC_REALM_SMTP_ADDRESS") != "" {
			defaultMailMode = "smtp"
		} else {
			defaultMailMode = "disabled"
		}
	}
	defaultCheckpointInterval, err := envconfig.Duration("DARK_MAGIC_REALM_CHECKPOINT_INTERVAL", 15*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defaultPresenceTimeout, err := envconfig.Duration("DARK_MAGIC_REALM_PRESENCE_TIMEOUT", 30*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = flag.String("env-file", environment.DefaultPath, "environment file (overrides the default realm.env selection)")
	logLevel := flag.String("log-level", environmentDefault("DARK_MAGIC_REALM_LOG_LEVEL", "info"), "log verbosity: trace, debug, info, warn, or error")
	listenAddress := flag.String("listen", environmentDefault("DARK_MAGIC_REALM_LISTEN", ":6112"), "serve the authenticated realm API on this TCP address")
	dataDirectory := flag.String("data-dir", os.Getenv("DARK_MAGIC_REALM_DATA"), "durable realm data directory")
	postgresURL := flag.String("postgres-url", os.Getenv("DARK_MAGIC_REALM_POSTGRES_URL"), "required PostgreSQL connection string")
	accountBaseURL := flag.String("account-base-url", defaultAccountURL, "public HTTPS origin used in account mail links")
	accountMailMode := flag.String("account-mail-mode", defaultMailMode, "account mail delivery: disabled, smtp, log, or auto-verify")
	smtpAddress := flag.String("smtp-address", os.Getenv("DARK_MAGIC_REALM_SMTP_ADDRESS"), "SMTP host:port for transactional account mail; empty disables delivery")
	smtpFrom := flag.String("smtp-from", defaultSMTPFrom, "transactional mail sender address")
	smtpUsername := flag.String("smtp-username", os.Getenv("DARK_MAGIC_REALM_SMTP_USERNAME"), "SMTP username")
	smtpPassword := os.Getenv("DARK_MAGIC_REALM_SMTP_PASSWORD")
	smtpRequireTLS := flag.Bool("smtp-require-tls", false, "require SMTP STARTTLS (enable outside the local Mailpit profile)")
	workerExecutable := flag.String("worker-executable", os.Getenv("DARK_MAGIC_REALM_WORKER"), "cmd/server executable used for Realm games; empty disables game creation")
	workerControlListen := flag.String("worker-control-listen", environmentDefault("DARK_MAGIC_REALM_WORKER_CONTROL_LISTEN", "127.0.0.1:0"), "loopback address for private worker control")
	workerGameListen := flag.String("worker-game-listen", environmentDefault("DARK_MAGIC_REALM_WORKER_GAME_LISTEN", "0.0.0.0:0"), "UDP listen address assigned to each game worker")
	workerGameAdvertiseHost := flag.String("worker-game-advertise-host", os.Getenv("DARK_MAGIC_REALM_GAME_HOST"), "client-reachable game-worker host; empty uses loopback for local clients")
	workerHealthInterval := flag.Duration("worker-health-interval", 2*time.Second, "interval between Realm session pruning and game-worker health probes")
	checkpointInterval := flag.Duration("checkpoint-interval", defaultCheckpointInterval,
		"minimum interval between canonical game checkpoints")
	presenceTimeout := flag.Duration("presence-timeout", defaultPresenceTimeout,
		"maximum time an unresponsive Realm channel member remains visible")
	operatorTokenFile := flag.String("operator-token-file", os.Getenv("DARK_MAGIC_REALM_OPERATOR_TOKEN_FILE"),
		"owner-only bearer-token file enabling the operator API; empty disables it")
	operatorListen := flag.String("operator-listen", os.Getenv("DARK_MAGIC_REALM_OPERATOR_LISTEN"),
		"explicit loopback address for the private operator API; empty disables it")
	adminShell := flag.Bool("admin-shell", true, "enable the interactive administration shell on standard input")
	flag.Parse()
	level, err := logging.ParseLevel(*logLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	directory, err := realm.DataDirectory(*dataDirectory)
	if err != nil {
		slog.Error("resolving realm data directory", "error", err)
		os.Exit(1)
	}
	if strings.TrimSpace(*postgresURL) == "" {
		slog.Error("Realm requires PostgreSQL", "configuration", "--postgres-url or DARK_MAGIC_REALM_POSTGRES_URL")
		os.Exit(1)
	}
	postgres, err := realm.OpenPostgres(ctx, *postgresURL, 0)
	if err != nil {
		slog.Error("opening Realm PostgreSQL repositories", "error", err)
		os.Exit(1)
	}
	if err := postgres.Accounts.SetAccountBaseURL(*accountBaseURL); err != nil {
		slog.Error("configuring Realm account URL", "error", err)
		os.Exit(1)
	}
	slog.Info("using PostgreSQL Realm repositories")
	var allocator realm.GameAllocator
	var processAllocator *realm.ProcessAllocator
	if *workerExecutable != "" {
		assetSetID, assetSetErr := content.AssetSetIdentityFromEnvironment()
		if assetSetErr != nil {
			slog.Error("identifying worker game assets", "error", assetSetErr)
			os.Exit(1)
		}
		slog.Info("validated worker game asset set", "asset_set_id", assetSetID)
		processes, allocatorErr := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{Executable: *workerExecutable,
			Arguments: []string{"--log-level", *logLevel}, StateDirectory: filepath.Join(directory, "workers"),
			ControlListenAddress: *workerControlListen, GameListenAddress: *workerGameListen,
			GameAdvertiseHost: *workerGameAdvertiseHost, LogWriter: os.Stderr, ExpectedAssetSetID: assetSetID})
		if allocatorErr != nil {
			slog.Error("configuring Realm worker allocator", "error", allocatorErr)
			os.Exit(1)
		}
		allocator, processAllocator = processes, processes
	} else {
		slog.Warn("Realm game creation is disabled; configure --worker-executable with the cmd/server binary")
	}
	audit := realm.AuditSink(realm.NewSlogAuditSink(nil))
	audit = realm.NewAuditFanout(audit, postgres.Audit)
	control, err := realm.NewControlPlane(realm.ControlPlaneConfig{Accounts: postgres.Accounts, Characters: postgres.Characters,
		Games: postgres.Games, Allocations: postgres.Allocations, Memberships: postgres.Memberships, Checkpoints: postgres.Checkpoints,
		Audit: audit, Allocator: allocator, CheckpointInterval: *checkpointInterval, PresenceTimeout: *presenceTimeout})
	if err != nil {
		slog.Error("initializing realm control plane", "error", err)
		os.Exit(1)
	}
	if recovered, recoveryErr := control.RecoverInterruptedGames(ctx); recoveryErr != nil {
		slog.Error("recovering interrupted Realm games", "recovered_games", recovered, "error", recoveryErr)
		os.Exit(1)
	} else if recovered > 0 {
		slog.Warn("failed closed interrupted Realm games", "recovered_games", recovered)
	}
	if *accountMailMode != "disabled" {
		var mailer realm.MailSender
		var mailErr error
		switch *accountMailMode {
		case "smtp":
			mailer, mailErr = realm.NewSMTPMailer(realm.SMTPConfig{Address: *smtpAddress, From: *smtpFrom,
				Username: *smtpUsername, Password: smtpPassword, RequireTLS: *smtpRequireTLS})
		case "log", "auto-verify":
			if !realm.DevelopmentMailAllowed(*listenAddress) {
				slog.Error("development account mail mode requires an explicit loopback --listen address",
					"listen", *listenAddress, "mode", *accountMailMode)
				os.Exit(1)
			}
			mailer, mailErr = realm.NewDevelopmentMailSender(*accountMailMode, control, nil)
			slog.Warn("development-only account mail mode enabled", "mode", *accountMailMode,
				"links_may_be_logged", true)
		default:
			mailErr = realm.ErrMailUnavailable
		}
		if mailErr != nil {
			slog.Error("configuring Realm account mail delivery", "mode", *accountMailMode, "error", mailErr)
			os.Exit(1)
		}
		go realm.RunMailWorker(ctx, postgres.Mail, mailer, fmt.Sprintf("realm-%d", os.Getpid()), time.Second,
			func(result realm.MailWorkerResult) {
				if result.Err != nil {
					slog.Warn("delivering Realm account mail", "job_id", result.JobID, "kind", result.Kind, "error", result.Err)
				} else {
					slog.Info("delivered Realm account mail", "job_id", result.JobID, "kind", result.Kind)
				}
			})
	}
	var portalAssets *portalassets.Cache
	if os.Getenv("MPQ_DIRECTORY") != "" {
		contentFS, contentErr := content.FromEnvironment()
		if contentErr != nil {
			slog.Error("opening Realm portal game assets", "error", contentErr)
			os.Exit(1)
		}
		defer contentFS.Close()
		portalAssets, err = portalassets.New(contentFS, filepath.Join(directory, "cache", "realm-portal"))
		if err != nil {
			slog.Error("configuring Realm portal asset cache", "error", err)
			os.Exit(1)
		}
	}
	if (*operatorTokenFile == "") != (*operatorListen == "") {
		slog.Error("configuring Realm operator API", "error", "operator-token-file and operator-listen must be set together")
		os.Exit(1)
	}
	apiHandler, err := realm.NewHTTPHandler(control)
	if err != nil {
		slog.Error("building realm API", "error", err)
		os.Exit(1)
	}
	handler, err := realmportal.NewHandler(control, apiHandler, portalAssets)
	if err != nil {
		slog.Error("building realm portal", "error", err)
		os.Exit(1)
	}
	trust, err := networktrust.New(filepath.Join(directory, "network"))
	if err != nil {
		slog.Error("loading realm network identity", "error", err)
		os.Exit(1)
	}
	serverTLS, _, fingerprint, err := trust.HostTLS()
	if err != nil {
		slog.Error("loading realm TLS identity", "error", err)
		os.Exit(1)
	}
	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		slog.Error("listening for realm clients", "error", err)
		os.Exit(1)
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 90 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(tls.NewListener(listener, serverTLS)) }()
	slog.Info("serving authenticated realm API", "address", listener.Addr(), "tls_fingerprint", fingerprint, "version", control.Version())
	var operatorServer *http.Server
	var operatorErrors <-chan error
	if *operatorListen != "" {
		host, port, splitErr := net.SplitHostPort(*operatorListen)
		ip := net.ParseIP(host)
		if splitErr != nil || port == "" || ip == nil || !ip.IsLoopback() {
			slog.Error("configuring Realm operator API", "error", "operator-listen must use an explicit loopback IP and port")
			os.Exit(1)
		}
		operatorToken, tokenErr := realm.LoadOrCreateOperatorToken(*operatorTokenFile)
		if tokenErr != nil {
			slog.Error("loading Realm operator credential", "error", tokenErr)
			os.Exit(1)
		}
		operatorHandler, handlerErr := realm.NewOperatorHTTPHandler(control, operatorToken)
		if handlerErr != nil {
			slog.Error("building Realm operator API", "error", handlerErr)
			os.Exit(1)
		}
		operatorListener, listenErr := net.Listen("tcp", *operatorListen)
		if listenErr != nil {
			slog.Error("listening for Realm operators", "error", listenErr)
			os.Exit(1)
		}
		operatorServer = &http.Server{Handler: operatorHandler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
		channel := make(chan error, 1)
		operatorErrors = channel
		go func() { channel <- operatorServer.Serve(tls.NewListener(operatorListener, serverTLS)) }()
		slog.Info("serving private Realm operator API", "address", operatorListener.Addr(), "tls_fingerprint", fingerprint)
	}
	var shellErrors <-chan error
	if *adminShell {
		policy := shell.Policy{Name: "local-realm-admin", Mutable: true}
		channel := make(chan error, 1)
		shellErrors = channel
		go func() { channel <- headlessshell.Run(ctx, "realm", policy, level, os.Stdin, os.Stdout) }()
	}
	go realm.RunMaintenance(ctx, control, processAllocator != nil, *workerHealthInterval, func(result realm.MaintenanceResult) {
		if result.Err != nil {
			slog.Warn("running Realm maintenance", "pruned_sessions", result.PrunedSessions, "pruned_presence", result.PrunedPresence,
				"reconciled_games", result.ReconciledGames, "error", result.Err)
		} else if result.PrunedSessions > 0 || result.PrunedPresence > 0 || result.ReconciledGames > 0 {
			slog.Info("completed Realm maintenance", "pruned_sessions", result.PrunedSessions, "pruned_presence", result.PrunedPresence,
				"reconciled_games", result.ReconciledGames)
		}
	})
	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-shellErrors:
	case runErr = <-serverErrors:
		if errors.Is(runErr, http.ErrServerClosed) {
			runErr = nil
		}
	case runErr = <-operatorErrors:
		if errors.Is(runErr, http.ErrServerClosed) {
			runErr = nil
		}
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownContext)
	if operatorServer != nil {
		shutdownErr = errors.Join(shutdownErr, operatorServer.Shutdown(shutdownContext))
	}
	if processAllocator != nil {
		shutdownErr = errors.Join(shutdownErr, processAllocator.Close(shutdownContext))
	}
	postgres.Close()
	if err := errors.Join(runErr, shutdownErr); err != nil {
		slog.Error("running realm", "error", err)
		os.Exit(1)
	}
}

func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
