package shell

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeEvaluator struct{ closed bool }

func (f *fakeEvaluator) Evaluate(_ context.Context, source string) (Result, error) {
	if source == "bad" {
		return Result{}, errors.New("boom")
	}
	return Result{Kind: "value", Text: source + "!"}, nil
}
func (f *fakeEvaluator) Complete(context.Context, string) ([]Candidate, error) {
	return []Candidate{{Value: "print"}, {Value: "pairs"}}, nil
}
func (f *fakeEvaluator) Close() error { f.closed = true; return nil }

func TestSessionOwnsTranscriptHistoryCompletionAndLifecycle(t *testing.T) {
	evaluator := &fakeEvaluator{}
	session, err := NewSession("local", "client", Policy{Name: "developer", Mutable: true}, evaluator)
	if err != nil {
		t.Fatal(err)
	}
	if got := session.Submit(context.Background(), "  return 1 "); got.Result.Text != "return 1!" {
		t.Fatalf("entry = %#v", got)
	}
	if got := session.Submit(context.Background(), "bad"); got.Error != "boom" {
		t.Fatalf("error entry = %#v", got)
	}
	if got := session.History(); !reflect.DeepEqual(got, []string{"return 1", "bad"}) {
		t.Fatalf("history = %v", got)
	}
	candidates, err := session.Complete(context.Background(), "p")
	if err != nil || SharedPrefix(candidates) != "p" || candidates[0].Value != "pairs" {
		t.Fatalf("completion = %#v, %v", candidates, err)
	}
	if err := session.Close(); err != nil || !evaluator.closed {
		t.Fatalf("close = %v, closed=%v", err, evaluator.closed)
	}
}
