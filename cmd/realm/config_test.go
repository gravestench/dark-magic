package main

import "testing"

// TestDefaultMailModePrefersExplicitPolicy verifies that configuration wins over inference.
func TestDefaultMailModePrefersExplicitPolicy(t *testing.T) {
	t.Setenv("DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE", "log")
	t.Setenv("DARK_MAGIC_REALM_SMTP_ADDRESS", "smtp.example.test:25")

	if got := defaultMailMode(); got != "log" {
		t.Fatalf("defaultMailMode() = %q, want log", got)
	}
}

// TestDefaultMailModeInfersSMTP verifies the safe fallback when SMTP is configured.
func TestDefaultMailModeInfersSMTP(t *testing.T) {
	t.Setenv("DARK_MAGIC_REALM_ACCOUNT_MAIL_MODE", "")
	t.Setenv("DARK_MAGIC_REALM_SMTP_ADDRESS", "smtp.example.test:25")

	if got := defaultMailMode(); got != "smtp" {
		t.Fatalf("defaultMailMode() = %q, want smtp", got)
	}
}

// TestValidateLoopbackAddressRejectsPublicBindings protects the private operator boundary.
func TestValidateLoopbackAddressRejectsPublicBindings(t *testing.T) {
	if err := validateLoopbackAddress("127.0.0.1:6113"); err != nil {
		t.Fatalf("validate loopback address: %v", err)
	}

	if err := validateLoopbackAddress("0.0.0.0:6113"); err == nil {
		t.Fatal("validateLoopbackAddress accepted a public binding")
	}
}
