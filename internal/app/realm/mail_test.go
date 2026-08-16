package realm

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRenderAccountMailIsBoundedToKnownTemplates(t *testing.T) {
	job := MailJob{Kind: "verify_email", Recipient: "owner@example.test", Payload: map[string]any{
		"account_name": "Alice", "verification_url": "https://accounts.dark-magic.test/verify?token=opaque"}}
	subject, body, err := renderAccountMail(job)
	if err != nil || !strings.Contains(subject, "Verify") || !strings.Contains(body, "token=opaque") {
		t.Fatalf("subject=%q body=%q error=%v", subject, body, err)
	}
	job.Kind = "arbitrary"
	if _, _, err := renderAccountMail(job); !errors.Is(err, ErrMailUnavailable) {
		t.Fatalf("unknown template error = %v", err)
	}
}

func TestRunMailWorkerCompletesSuccessfulDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	outbox := &testMailOutbox{job: MailJob{ID: "mail-1", Kind: "verify_email", Recipient: "owner@example.test",
		Payload: map[string]any{"verification_url": "https://accounts.dark-magic.test/verify?token=opaque"}}, cancel: cancel}
	sender := &testMailSender{}
	done := make(chan struct{})
	go func() {
		RunMailWorker(ctx, outbox, sender, "mailer-1", time.Millisecond, nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("mail worker did not stop")
	}
	if sender.sent != 1 || outbox.completed != "mail-1" || outbox.retried {
		t.Fatalf("sent=%d completed=%q retried=%v", sender.sent, outbox.completed, outbox.retried)
	}
}

func TestRunMailWorkerRetriesFailedDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	outbox := &testMailOutbox{job: MailJob{ID: "mail-2", Kind: "reset_password", Attempts: 1}, cancel: cancel}
	sender := &testMailSender{err: errors.New("smtp unavailable")}
	done := make(chan struct{})
	go func() {
		RunMailWorker(ctx, outbox, sender, "mailer-1", time.Millisecond, nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("mail worker did not stop")
	}
	if !outbox.retried || outbox.retryMessage != "smtp unavailable" || outbox.completed != "" {
		t.Fatalf("retried=%v message=%q completed=%q", outbox.retried, outbox.retryMessage, outbox.completed)
	}
}

func TestDevelopmentMailRequiresExplicitModeAndLoopback(t *testing.T) {
	for address, allowed := range map[string]bool{
		"127.0.0.1:6112": true, "[::1]:6112": true, "localhost:6112": true,
		":6112": false, "0.0.0.0:6112": false, "192.0.2.1:6112": false,
	} {
		if actual := DevelopmentMailAllowed(address); actual != allowed {
			t.Errorf("DevelopmentMailAllowed(%q) = %v", address, actual)
		}
	}
	if _, err := NewDevelopmentMailSender("unknown", nil, nil); !errors.Is(err, ErrMailUnavailable) {
		t.Fatalf("unknown mode error = %v", err)
	}
}

func TestDevelopmentMailCanLogOrAutoVerify(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	job := MailJob{Kind: "verify_email", Recipient: "owner@example.test", Payload: map[string]any{
		"verification_url": "https://accounts.dark-magic.test/verify?token=opaque"}}
	logging, err := NewDevelopmentMailSender("log", nil, logger)
	if err != nil || logging.Send(t.Context(), job) != nil {
		t.Fatalf("log sender error = %v", err)
	}
	lifecycle := &testAccountLifecycle{}
	automatic, err := NewDevelopmentMailSender("auto-verify", lifecycle, logger)
	if err != nil || automatic.Send(t.Context(), job) != nil || lifecycle.verifiedToken != "opaque" {
		t.Fatalf("auto sender error=%v token=%q", err, lifecycle.verifiedToken)
	}
	recovery := MailJob{Kind: "reset_password", Recipient: job.Recipient, Payload: map[string]any{
		"recovery_url": "https://accounts.dark-magic.test/recover?token=recovery"}}
	if err := automatic.Send(t.Context(), recovery); err != nil || lifecycle.verifiedToken != "opaque" {
		t.Fatalf("recovery development mail error=%v token=%q", err, lifecycle.verifiedToken)
	}
}

type testMailOutbox struct {
	mu           sync.Mutex
	job          MailJob
	claimed      bool
	completed    string
	retried      bool
	retryMessage string
	cancel       context.CancelFunc
}

func (outbox *testMailOutbox) ClaimMail(context.Context, string, time.Duration) (MailJob, error) {
	outbox.mu.Lock()
	defer outbox.mu.Unlock()
	if outbox.claimed {
		return MailJob{}, ErrMailUnavailable
	}
	outbox.claimed = true
	return outbox.job, nil
}

func (outbox *testMailOutbox) CompleteMail(_ context.Context, _, jobID string) error {
	outbox.mu.Lock()
	outbox.completed = jobID
	outbox.mu.Unlock()
	outbox.cancel()
	return nil
}

func (outbox *testMailOutbox) RetryMail(_ context.Context, _, _ string, message string, _ time.Time) error {
	outbox.mu.Lock()
	outbox.retried, outbox.retryMessage = true, message
	outbox.mu.Unlock()
	outbox.cancel()
	return nil
}

type testMailSender struct {
	sent int
	err  error
}

func (sender *testMailSender) Send(context.Context, MailJob) error {
	sender.sent++
	return sender.err
}

type testAccountLifecycle struct{ verifiedToken string }

func (*testAccountLifecycle) Signup(context.Context, SignupRequest) (Account, error) {
	return Account{}, nil
}
func (lifecycle *testAccountLifecycle) VerifyEmail(_ context.Context, token string) (Account, error) {
	lifecycle.verifiedToken = token
	return Account{ID: "account-1", EmailVerified: true}, nil
}
func (*testAccountLifecycle) BeginPasswordRecovery(context.Context, string) error { return nil }
func (*testAccountLifecycle) CompletePasswordRecovery(context.Context, string, string) error {
	return nil
}
