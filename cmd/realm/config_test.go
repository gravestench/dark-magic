package main

import "testing"

// TestDefaultMailModePrefersExplicitPolicy prevents SMTP environment discovery
// from overriding an operator's deliberate delivery or disablement choice.
func TestDefaultMailModePrefersExplicitPolicy(t *testing.T) {
	t.Setenv("DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE", "log")
	t.Setenv("DARK_MAGIC_REALM_SMTP_ADDRESS", "smtp.example.test:25")

	if got := defaultMailMode(); got != "log" {
		t.Fatalf("defaultMailMode() = %q, want log", got)
	}
}

// TestDefaultMailModeInfersSMTP ensures a configured SMTP server selects actual
// delivery when no explicit mode exists, rather than silently logging account links.
func TestDefaultMailModeInfersSMTP(t *testing.T) {
	t.Setenv("DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE", "")
	t.Setenv("DARK_MAGIC_REALM_SMTP_ADDRESS", "smtp.example.test:25")

	if got := defaultMailMode(); got != "smtp" {
		t.Fatalf("defaultMailMode() = %q, want smtp", got)
	}
}

// TestValidateLoopbackAddressRejectsPublicBindings protects the assumption that
// bearer-token administration is reachable only from the local host.
func TestValidateLoopbackAddressRejectsPublicBindings(t *testing.T) {
	if err := validateLoopbackAddress("127.0.0.1:6113"); err != nil {
		t.Fatalf("validate loopback address: %v", err)
	}

	if err := validateLoopbackAddress("0.0.0.0:6113"); err == nil {
		t.Fatal("validateLoopbackAddress accepted a public binding")
	}
}
