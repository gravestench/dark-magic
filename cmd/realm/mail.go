package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

// startMailWorker configures optional account mail and starts its outbox consumer.
func startMailWorker(
	ctx context.Context,
	postgres *realm.Postgres,
	control *realm.ControlPlane,
	config realmConfig,
) error {
	if config.accountMailMode == "disabled" {
		return nil
	}

	mailer, err := buildMailer(control, config)
	if err != nil {
		return err
	}

	workerID := fmt.Sprintf("realm-%d", os.Getpid())
	go realm.RunMailWorker(ctx, postgres.Mail, mailer, workerID, time.Second, logMailResult)

	return nil
}

// buildMailer validates the selected delivery mode and creates its sender.
func buildMailer(control *realm.ControlPlane, config realmConfig) (realm.MailSender, error) {
	switch config.accountMailMode {
	case "smtp":
		return realm.NewSMTPMailer(realm.SMTPConfig{
			Address:    config.smtpAddress,
			From:       config.smtpFrom,
			Username:   config.smtpUsername,
			Password:   config.smtpPassword,
			RequireTLS: config.smtpRequireTLS,
		})
	case "log", "auto-verify":
		if !realm.DevelopmentMailAllowed(config.listenAddress) {
			return nil, fmt.Errorf(
				"development mail mode %q requires an explicit loopback listen address",
				config.accountMailMode,
			)
		}

		slog.Warn(
			"development-only account mail mode enabled",
			"mode", config.accountMailMode,
			"links_may_be_logged", true,
		)

		return realm.NewDevelopmentMailSender(config.accountMailMode, control, nil)
	default:
		return nil, fmt.Errorf("configure Realm account mail mode %q: %w", config.accountMailMode, realm.ErrMailUnavailable)
	}
}

// logMailResult records one asynchronous outbox delivery result.
func logMailResult(result realm.MailWorkerResult) {
	if result.Err != nil {
		slog.Warn(
			"delivering Realm account mail",
			"job_id", result.JobID,
			"kind", result.Kind,
			"error", result.Err,
		)

		return
	}

	slog.Info("delivered Realm account mail", "job_id", result.JobID, "kind", result.Kind)
}
