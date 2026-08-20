package realm

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

type MailSender interface {
	Send(context.Context, MailJob) error
}

type SMTPConfig struct {
	Address    string
	From       string
	Username   string
	Password   string
	ServerName string
	RequireTLS bool
}

type SMTPMailer struct {
	config SMTPConfig
	host   string
	from   string
}

type DevelopmentMailSender struct {
	autoVerify bool
	lifecycle  AccountLifecycle
	logger     *slog.Logger
}

// NewDevelopmentMailSender creates an intentionally unsafe local-only sink.
// Log mode exposes action links; auto-verify consumes verification links and
// still exposes recovery links. Callers must first enforce a loopback listener.
func NewDevelopmentMailSender(
	mode string,
	lifecycle AccountLifecycle,
	logger *slog.Logger,
) (*DevelopmentMailSender, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "log":
		return &DevelopmentMailSender{lifecycle: lifecycle, logger: logger}, nil
	case "auto-verify":
		if lifecycle == nil {
			return nil, ErrMailUnavailable
		}

		return &DevelopmentMailSender{autoVerify: true, lifecycle: lifecycle, logger: logger}, nil
	default:
		return nil, ErrMailUnavailable
	}
}

// DevelopmentMailAllowed contains development mail allowed within the mail boundary so callers do not duplicate its
// domain-specific policy.
func DevelopmentMailAllowed(listenAddress string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(listenAddress))
	if err != nil {
		return false
	}

	if strings.EqualFold(host, "localhost") {
		return true
	}

	address := net.ParseIP(host)

	return address != nil && address.IsLoopback()
}

// Send contains send within the mail boundary so callers do not duplicate its domain-specific policy.
func (sender *DevelopmentMailSender) Send(ctx context.Context, job MailJob) error {
	if sender == nil {
		return ErrMailUnavailable
	}

	field := "verification_url"
	if job.Kind == "reset_password" {
		field = "recovery_url"
	}

	actionURL, _ := job.Payload[field].(string)

	parsed, err := url.Parse(actionURL)
	if err != nil || parsed.Scheme != "https" || parsed.Query().Get("token") == "" {
		return ErrMailUnavailable
	}

	logger := sender.logger
	if logger == nil {
		logger = slog.Default()
	}

	if sender.autoVerify && job.Kind == "verify_email" {
		account, err := sender.lifecycle.VerifyEmail(ctx, parsed.Query().Get("token"))
		if err != nil {
			return err
		}

		logger.Warn("development account email auto-verified", "development_only", true,
			"mail_kind", job.Kind, "recipient", job.Recipient, "account_id", account.ID)

		return nil
	}

	logger.Warn("development account mail action", "development_only", true,
		"mail_kind", job.Kind, "recipient", job.Recipient, "action_url", actionURL)

	return nil
}

// NewSMTPMailer constructs the mail boundary and validates dependencies before callers can publish or mutate shared
// state.
func NewSMTPMailer(config SMTPConfig) (*SMTPMailer, error) {
	host, _, err := net.SplitHostPort(strings.TrimSpace(config.Address))
	if err != nil || host == "" {
		return nil, ErrMailUnavailable
	}

	from, err := mail.ParseAddress(strings.TrimSpace(config.From))
	if err != nil || from.Address == "" {
		return nil, ErrMailUnavailable
	}

	if config.ServerName == "" {
		config.ServerName = host
	}

	return &SMTPMailer{config: config, host: host, from: from.Address}, nil
}

// Send contains send within the mail boundary so callers do not duplicate its domain-specific policy.
func (sender *SMTPMailer) Send(ctx context.Context, job MailJob) error {
	if sender == nil || strings.TrimSpace(job.Recipient) == "" {
		return ErrMailUnavailable
	}

	recipient, err := mail.ParseAddress(job.Recipient)
	if err != nil || recipient.Address != job.Recipient || strings.ContainsAny(job.Recipient, "\r\n") {
		return ErrMailUnavailable
	}

	subject, body, err := renderAccountMail(job)
	if err != nil {
		return err
	}

	dialer := net.Dialer{Timeout: 10 * time.Second}

	connection, err := dialer.DialContext(ctx, "tcp", sender.config.Address)
	if err != nil {
		return fmt.Errorf("realm: dial SMTP: %w", err)
	}

	client, err := smtp.NewClient(connection, sender.host)
	if err != nil {
		_ = connection.Close()
		return fmt.Errorf("realm: open SMTP client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if supported, _ := client.Extension("STARTTLS"); supported {
		if err := client.StartTLS(
			&tls.Config{MinVersion: tls.VersionTLS12, ServerName: sender.config.ServerName},
		); err != nil {
			return fmt.Errorf("realm: start SMTP TLS: %w", err)
		}
	} else if sender.config.RequireTLS {
		return errors.New("realm: SMTP server does not offer STARTTLS")
	}

	if sender.config.Username != "" {
		if err := client.Auth(
			smtp.PlainAuth("", sender.config.Username, sender.config.Password, sender.config.ServerName),
		); err != nil {
			return fmt.Errorf("realm: authenticate SMTP: %w", err)
		}
	}

	if err := client.Mail(sender.from); err != nil {
		return fmt.Errorf("realm: SMTP sender: %w", err)
	}

	if err := client.Rcpt(recipient.Address); err != nil {
		return fmt.Errorf("realm: SMTP recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("realm: SMTP data: %w", err)
	}

	message := "From: " + sender.from + "\r\nTo: " + recipient.Address + "\r\nSubject: " + subject +
		"\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body + "\r\n"
	if _, err := io.WriteString(writer, message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("realm: write SMTP message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("realm: finish SMTP message: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("realm: quit SMTP: %w", err)
	}

	return nil
}

// renderAccountMail contains render account mail within the mail boundary so callers do not duplicate its
// domain-specific policy.
func renderAccountMail(job MailJob) (string, string, error) {
	account, _ := job.Payload["account_name"].(string)
	switch job.Kind {
	case "verify_email":
		link, _ := job.Payload["verification_url"].(string)
		if link == "" {
			return "", "", ErrMailUnavailable
		}

		return "Verify your Dark Magic Realm account",
			fmt.Sprintf(
				"Hello %s,\n\nVerify your Realm account:\n%s\n\nIf you did not request this, ignore this message.",
				account,
				link,
			), nil
	case "reset_password":
		link, _ := job.Payload["recovery_url"].(string)
		if link == "" {
			return "", "", ErrMailUnavailable
		}

		return "Reset your Dark Magic Realm password",
			fmt.Sprintf(
				"Hello %s,\n\nReset your Realm password:\n%s\n\nIf you did not request this, ignore this message.",
				account,
				link,
			), nil
	default:
		return "", "", ErrMailUnavailable
	}
}

type MailWorkerResult struct {
	JobID string
	Kind  string
	Err   error
}

// RunMailWorker continuously drains the leased transactional outbox. Empty
// queues are normal; delivery failures are rescheduled with bounded backoff.
func RunMailWorker(ctx context.Context, outbox MailOutbox, sender MailSender, workerID string,
	pollInterval time.Duration, observe func(MailWorkerResult)) {
	if ctx == nil || outbox == nil || sender == nil || strings.TrimSpace(workerID) == "" {
		return
	}

	if pollInterval <= 0 {
		pollInterval = time.Second
	}

	for {
		job, err := outbox.ClaimMail(ctx, workerID, 30*time.Second)
		if err != nil {
			if !errors.Is(err, ErrMailUnavailable) && observe != nil {
				observe(MailWorkerResult{Err: err})
			}

			if !waitContext(ctx, pollInterval) {
				return
			}

			continue
		}

		err = sender.Send(ctx, job)
		if err == nil {
			err = outbox.CompleteMail(ctx, workerID, job.ID)
		} else {
			delay := time.Duration(1<<min(job.Attempts, 6)) * time.Second
			retryErr := outbox.RetryMail(ctx, workerID, job.ID, err.Error(), time.Now().Add(delay))
			err = errors.Join(err, retryErr)
		}

		if observe != nil {
			observe(MailWorkerResult{JobID: job.ID, Kind: job.Kind, Err: err})
		}

		if ctx.Err() != nil {
			return
		}
	}
}

// waitContext contains wait context within the mail boundary so callers do not duplicate its domain-specific policy.
func waitContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
