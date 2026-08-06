package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/shell"
)

type evaluator struct{}

func (evaluator) Evaluate(_ context.Context, source string) (shell.Result, error) {
	return shell.Result{Kind: "value", Text: source}, nil
}

func (evaluator) Complete(context.Context, string) ([]shell.Candidate, error) {
	return []shell.Candidate{{Value: "print", Detail: "global"}}, nil
}

func (evaluator) Close() error { return nil }

func TestModelShowsAuthorityAndEditsCompletionToken(t *testing.T) {
	session, err := shell.NewSession("test", "realm", shell.Policy{
		Name: "operator", Capabilities: []string{"status/v1"}, Mutable: false,
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
