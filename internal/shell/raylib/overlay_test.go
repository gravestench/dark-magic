package raylibshell

import (
	"context"
	"strings"
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
	overlay.cursor = 3
	overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{"tab": {Pressed: true}}})
	if overlay.input != "print" {
		t.Fatalf("completion = %q", overlay.input)
	}
	overlay.input = "hé"
	overlay.cursor = 2
	overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{"backspace": {Pressed: true}}})
	if overlay.input != "h" {
		t.Fatalf("backspace = %q", overlay.input)
	}
	overlay.input, overlay.cursor = "ab", 1
	overlay.Handle(context.Background(), inputstate.Frame{Text: "X", Actions: map[string]inputstate.ActionState{}})
	if overlay.input != "aXb" || overlay.cursor != 2 || overlay.inputWithCaret() != "aX|b" {
		t.Fatalf("mid-line insert = %q cursor=%d", overlay.input, overlay.cursor)
	}
	overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{"delete": {Pressed: true}}})
	if overlay.input != "aX" {
		t.Fatalf("mid-line delete = %q", overlay.input)
	}
}

func TestOverlayModalViewsKeepLogsOutOfLua(t *testing.T) {
	session, _ := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, evaluator{})
	logs := shell.NewLogBuffer(4)
	logs.Append(shell.LogEntry{At: time.Now(), Level: "info", Message: "visible log"})
	session.AttachLogs(logs)
	session.Submit(context.Background(), "lua value")
	overlay := New(session)
	overlay.open = true
	if lines := overlay.timeline(800); len(lines) != 2 || strings.Contains(lines[0].text, "visible log") {
		t.Fatalf("lua lines = %#v", lines)
	}
	overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{"shell_logs": {Pressed: true}}})
	if overlay.view != viewLogs {
		t.Fatal("F2 did not select logs")
	}
	if lines := overlay.timeline(800); len(lines) != 1 || !strings.Contains(lines[0].text, "visible log") {
		t.Fatalf("log lines = %#v", lines)
	}
	overlay.Handle(context.Background(), inputstate.Frame{Text: "ignored", Actions: map[string]inputstate.ActionState{}})
	if overlay.input != "" {
		t.Fatalf("log view edited Lua input: %q", overlay.input)
	}
}

func TestWrapTranscriptPreservesStyleAndUnicode(t *testing.T) {
	lines := wrapTranscript([]transcriptLine{{text: "héllo", result: true}}, 3)
	if len(lines) != 2 || lines[0].text != "hél" || lines[1].text != "lo" || !lines[1].result {
		t.Fatalf("wrapped = %#v", lines)
	}
}

func TestOverlayAnimatesAndCapturesThroughClose(t *testing.T) {
	session, _ := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, evaluator{})
	overlay := New(session)
	started := time.Unix(100, 0)
	overlay.setOpen(true, started)
	overlay.updateAnimation(started.Add(openDuration / 2))
	if overlay.progress != 0.5 {
		t.Fatalf("opening progress = %v", overlay.progress)
	}
	_, opacity := overlay.presentation()
	if opacity != 0.875 {
		t.Fatalf("opening opacity = %v", opacity)
	}
	overlay.setOpen(false, started.Add(openDuration/2))
	position, opacity := overlay.presentation()
	if position != 0.5 || opacity != 0.125 {
		t.Fatalf("closing presentation = position %v, opacity %v", position, opacity)
	}
	if !overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{}}) {
		t.Fatal("closing overlay released scene input before leaving the screen")
	}
	overlay.updateAnimation(started.Add(openDuration/2 + closeDuration))
	if overlay.progress != 0 {
		t.Fatalf("closing progress = %v", overlay.progress)
	}
	if overlay.Handle(context.Background(), inputstate.Frame{Actions: map[string]inputstate.ActionState{}}) {
		t.Fatal("closed overlay still captures scene input")
	}
}
