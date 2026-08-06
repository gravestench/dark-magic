package luashell

import (
	"context"
	"strings"
	"testing"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func TestEvaluatorPersistsLocalsFormatsValuesAndCompletesWithoutExecution(t *testing.T) {
	runtime := modruntime.New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	evaluator, err := New(runtime)
	if err != nil {
		t.Fatal(err)
	}
	defer evaluator.Close()
	if result, err := evaluator.Evaluate(context.Background(), "answer = 42"); err != nil || result.Text != "ok" {
		t.Fatalf("assignment = %#v, %v", result, err)
	}
	if result, err := evaluator.Evaluate(context.Background(), "answer"); err != nil || result.Text != "42" {
		t.Fatalf("answer = %#v, %v", result, err)
	}
	if result, err := evaluator.Evaluate(context.Background(), `{name="hero", level=2}`); err != nil || !strings.Contains(result.Text, "level=2") {
		t.Fatalf("table = %#v, %v", result, err)
	}
	candidates, err := evaluator.Complete(context.Background(), "pri")
	if err != nil || len(candidates) == 0 || candidates[0].Value != "print" {
		t.Fatalf("completion = %#v, %v", candidates, err)
	}
}
