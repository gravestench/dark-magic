package raylibshell

import (
	"context"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/inputstate"
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

func TestOverlayCapturesInputAndSubmitsThroughSharedSession(t *testing.T) {
	session, err := shell.NewSession("test", "client", shell.Policy{Name: "developer", Mutable: true}, evaluator{})
	if err != nil {
		t.Fatal(err)
	}
	overlay := New(session)
	pressed := func(name string) inputstate.Frame {
		return inputstate.Frame{Actions: map[string]inputstate.ActionState{name: {Pressed: true}}}
	}
	if captured := overlay.Handle(context.Background(), pressed("shell_toggle")); !captured || !overlay.Open() {
		t.Fatal("toggle did not open and capture input")
	}
	overlay.Handle(context.Background(), inputstate.Frame{Text: "1 + 2", Actions: map[string]inputstate.ActionState{}})
	overlay.Handle(context.Background(), pressed("confirm"))
	select {
	case entry := <-overlay.finished:
		if entry.Source != "1 + 2" || entry.Result.Text != "1 + 2" {
			t.Fatalf("entry = %#v", entry)
		}
	case <-time.After(time.Second):
		t.Fatal("evaluation did not finish")
	}
	if captured := overlay.Handle(context.Background(), pressed("cancel")); !captured || overlay.Open() {
		t.Fatal("escape did not close and capture input")
	}
}

func TestOverlayCompletesAndEditsUTF8(t *testing.T) {
	session, _ := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, evaluator{})
	overlay := New(session)
	overlay.open = true
	overlay.input = "pri"
	overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{"tab": {Pressed: true}}})
	if overlay.input != "print" {
		t.Fatalf("completion = %q", overlay.input)
	}
	overlay.input = "hé"
	overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{"backspace": {Pressed: true}}})
	if overlay.input != "h" {
		t.Fatalf("backspace = %q", overlay.input)
	}
}

func TestWrapTranscriptPreservesStyleAndUnicode(t *testing.T) {
	lines := wrapTranscript([]transcriptLine{{text: "héllo", result: true}}, 3)
	if len(lines) != 2 || lines[0].text != "hél" || lines[1].text != "lo" || !lines[1].result {
		t.Fatalf("wrapped = %#v", lines)
	}
}
