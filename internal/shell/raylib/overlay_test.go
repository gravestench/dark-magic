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

// Evaluate echoes source so overlay submission tests can observe the exact editor value.
func (evaluator) Evaluate(_ context.Context, source string) (shell.Result, error) {
	return shell.Result{Kind: "value", Text: source}, nil
}

// Complete returns one deterministic candidate for editing behavior tests.
func (evaluator) Complete(context.Context, string) ([]shell.Candidate, error) {
	return []shell.Candidate{{Value: "print", Detail: "global"}}, nil
}

// Close is a no-op because the overlay fixture owns no runtime resources.
func (evaluator) Close() error { return nil }

type completionEvaluator struct {
	candidates []shell.Candidate
}

// Evaluate is unused by completion tests and returns the source to satisfy the shared shell contract.
func (fixture completionEvaluator) Evaluate(_ context.Context, source string) (shell.Result, error) {
	return shell.Result{Kind: "value", Text: source}, nil
}

// Complete returns the fixture order so the Session can apply its stable candidate sorting contract.
func (fixture completionEvaluator) Complete(context.Context, string) ([]shell.Candidate, error) {
	return fixture.candidates, nil
}

// Close is a no-op because the completion fixture owns no runtime resources.
func (completionEvaluator) Close() error { return nil }

// actionFrame creates one pressed action without obscuring tests with the nested input map schema.
func actionFrame(name string) inputstate.Frame {
	return inputstate.Frame{
		Actions: map[string]inputstate.ActionState{name: {Pressed: true}},
	}
}

// TestOverlayCapturesInputAndSubmitsThroughSharedSession verifies modal capture,
// asynchronous submission, and close behavior through the public input boundary.
func TestOverlayCapturesInputAndSubmitsThroughSharedSession(t *testing.T) {
	session, err := shell.NewSession("test", "client", shell.Policy{Name: "developer", Mutable: true}, evaluator{})
	if err != nil {
		t.Fatal(err)
	}

	overlay := New(session)

	if captured := overlay.Handle(context.Background(), actionFrame("shell_toggle")); !captured || !overlay.Open() {
		t.Fatal("toggle did not open and capture input")
	}

	overlay.Handle(context.Background(), inputstate.Frame{Text: "1 + 2", Actions: map[string]inputstate.ActionState{}})
	overlay.Handle(context.Background(), actionFrame("confirm"))

	select {
	case entry := <-overlay.finished:
		if entry.Source != "1 + 2" || entry.Result.Text != "1 + 2" {
			t.Fatalf("entry = %#v", entry)
		}
	case <-time.After(time.Second):
		t.Fatal("evaluation did not finish")
	}

	if captured := overlay.Handle(context.Background(), actionFrame("cancel")); !captured || overlay.Open() {
		t.Fatal("escape did not close and capture input")
	}
}

// TestOverlayCompletesAndEditsUTF8 protects rune-safe cursor edits and completion insertion.
func TestOverlayCompletesAndEditsUTF8(t *testing.T) {
	session, _ := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, evaluator{})
	overlay := New(session)

	overlay.open = true
	overlay.input = "pri"
	overlay.cursor = 3
	overlay.Handle(context.Background(), actionFrame("tab"))

	if overlay.input != "print" {
		t.Fatalf("completion = %q", overlay.input)
	}

	overlay.input = "hé"
	overlay.cursor = 2
	overlay.Handle(context.Background(), actionFrame("backspace"))

	if overlay.input != "h" {
		t.Fatalf("backspace = %q", overlay.input)
	}

	overlay.input, overlay.cursor = "ab", 1
	overlay.Handle(context.Background(), inputstate.Frame{Text: "X", Actions: map[string]inputstate.ActionState{}})

	if overlay.input != "aXb" || overlay.cursor != 2 || overlay.inputWithCaret() != "aX|b" {
		t.Fatalf("mid-line insert = %q cursor=%d", overlay.input, overlay.cursor)
	}

	overlay.Handle(context.Background(), actionFrame("delete"))

	if overlay.input != "aX" {
		t.Fatalf("mid-line delete = %q", overlay.input)
	}
}

// TestOverlayCompletionCyclesCandidatesInSessionOrder protects forward Tab from
// skipping the first sorted candidate when the shared prefix matches the input.
func TestOverlayCompletionCyclesCandidatesInSessionOrder(t *testing.T) {
	fixture := completionEvaluator{candidates: []shell.Candidate{
		{Value: "print", Detail: "global"},
		{Value: "pairs", Detail: "global"},
	}}

	session, err := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, fixture)
	if err != nil {
		t.Fatal(err)
	}

	overlay := New(session)
	overlay.input = "p"
	overlay.cursor = 1

	overlay.complete(context.Background(), false)

	if overlay.input != "pairs" || overlay.candidateAt != 0 {
		t.Fatalf("first completion = %q at %d", overlay.input, overlay.candidateAt)
	}

	overlay.complete(context.Background(), false)

	if overlay.input != "print" || overlay.candidateAt != 1 {
		t.Fatalf("second completion = %q at %d", overlay.input, overlay.candidateAt)
	}
}

// TestOverlayModalViewsKeepLogsOutOfLua verifies each tab uses its own timeline
// and the log tab cannot mutate the Lua editor.
func TestOverlayModalViewsKeepLogsOutOfLua(t *testing.T) {
	session, _ := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, evaluator{})
	logs := shell.NewLogBuffer(4)

	logs.Append(shell.LogEntry{At: time.Now(), Level: "info", Message: "visible log"})
	session.AttachLogs(logs)
	session.Submit(context.Background(), "lua value")

	overlay := New(session)
	overlay.open = true
	lines := overlay.timeline(800)

	var luaValue, motd, processLog bool

	for _, line := range lines {
		luaValue = luaValue || strings.Contains(line.text, "lua value")
		motd = motd || strings.Contains(line.text, "Dark Magic Lua shell")
		processLog = processLog || strings.Contains(line.text, "visible log")
	}

	if !luaValue || !motd || processLog {
		t.Fatalf("lua lines = %#v", lines)
	}

	overlay.Handle(context.Background(), actionFrame("shell_logs"))

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

// TestWrapTranscriptPreservesStyleAndUnicode ensures rune wrapping copies semantic flags.
func TestWrapTranscriptPreservesStyleAndUnicode(t *testing.T) {
	lines := wrapTranscript([]transcriptLine{{text: "héllo", result: true}}, 3)

	if len(lines) != 2 || lines[0].text != "hél" || lines[1].text != "lo" || !lines[1].result {
		t.Fatalf("wrapped = %#v", lines)
	}
}

// TestOverlayUsesLiveFontSizeSetting verifies cached wrap columns invalidate after a setting change.
func TestOverlayUsesLiveFontSizeSetting(t *testing.T) {
	session, _ := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, evaluator{})
	settings := shell.NewTransientSettings()
	overlay := New(session, settings)

	overlay.timeline(800)
	defaultColumns := overlay.displayColumns

	if err := settings.SetFontSize(32); err != nil {
		t.Fatal(err)
	}

	overlay.timeline(800)

	if overlay.displayColumns >= defaultColumns {
		t.Fatalf("font size did not reduce columns: default=%d large=%d", defaultColumns, overlay.displayColumns)
	}
}

// TestOverlayLuaViewHasIndependentPageScrollback protects per-tab scroll offsets.
func TestOverlayLuaViewHasIndependentPageScrollback(t *testing.T) {
	session, _ := shell.NewSession("test", "client", shell.Policy{Name: "developer"}, evaluator{})
	overlay := New(session)

	overlay.open = true
	overlay.Handle(context.Background(), actionFrame("page_up"))

	if overlay.luaOffset != 10 || overlay.logOffset != 0 {
		t.Fatalf("offsets lua=%d logs=%d", overlay.luaOffset, overlay.logOffset)
	}

	overlay.Handle(context.Background(), actionFrame("shell_logs"))
	overlay.Handle(context.Background(), actionFrame("page_up"))

	if overlay.logOffset != 10 || overlay.luaOffset != 10 {
		t.Fatalf("offsets lua=%d logs=%d", overlay.luaOffset, overlay.logOffset)
	}
}

// TestOverlayAnimatesAndCapturesThroughClose pins animation progress, easing,
// and input capture until the closing panel leaves the screen.
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

	if !overlay.Handle(context.Background(), inputstate.Frame{}) {
		t.Fatal("closing overlay released scene input before leaving the screen")
	}

	overlay.updateAnimation(started.Add(openDuration/2 + closeDuration))

	if overlay.progress != 0 {
		t.Fatalf("closing progress = %v", overlay.progress)
	}

	if overlay.Handle(context.Background(), inputstate.Frame{}) {
		t.Fatal("closed overlay still captures scene input")
	}
}
