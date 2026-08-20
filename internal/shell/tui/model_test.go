package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/shell"
)

type evaluator struct{}

// Evaluate echoes source so adapter tests can distinguish Lua transcript content.
func (evaluator) Evaluate(_ context.Context, source string) (shell.Result, error) {
	return shell.Result{Kind: "value", Text: source}, nil
}

// Complete returns one stable candidate for terminal editing assertions.
func (evaluator) Complete(context.Context, string) ([]shell.Candidate, error) {
	return []shell.Candidate{{Value: "print", Detail: "global"}}, nil
}

// Close is a no-op because the adapter fixture owns no external runtime.
func (evaluator) Close() error { return nil }

// TestModelShowsAuthorityAndEditsCompletionToken verifies the header exposes
// restriction context and completion replaces only the active token.
func TestModelShowsAuthorityAndEditsCompletionToken(t *testing.T) {
	session, err := shell.NewSession("test", "realm", shell.Policy{
		Name:         "operator",
		Capabilities: []string{"status/v1"},
		Mutable:      false,
	}, evaluator{})
	if err != nil {
		t.Fatal(err)
	}

	model := NewModel(session)
	model.width, model.height = 80, 24
	model.resize()
	model.input.SetValue("return pri")
	model.complete(false)

	if got := model.input.Value(); got != "return print" {
		t.Fatalf("completed input = %q", got)
	}

	view := model.View().Content
	for _, expected := range []string{"realm", "operator", "read-only", "status/v1"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("view does not contain %q", expected)
		}
	}
}

// TestModelSeparatesLuaAndLogViews protects modal source separation so process
// diagnostics never leak into the Lua transcript tab and vice versa.
func TestModelSeparatesLuaAndLogViews(t *testing.T) {
	session, _ := shell.NewSession("test", "realm", shell.Policy{Name: "operator"}, evaluator{})
	logs := shell.NewLogBuffer(2)

	logs.Append(shell.LogEntry{Level: "warn", Message: "realm warning"})
	session.AttachLogs(logs)
	session.Submit(context.Background(), "lua-only")

	model := NewModel(session)
	if strings.Contains(model.output.View(), "realm warning") {
		t.Fatal("Lua view contains process logs")
	}

	model.view = viewLogs
	model.refreshTranscript()

	if output := model.output.View(); !strings.Contains(output, "realm warning") || strings.Contains(output, "lua-only") {
		t.Fatalf("log view = %q", output)
	}
}
