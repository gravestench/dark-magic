package loading

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCoordinatorReportsDependencyProgress verifies ordered execution and the renderer-facing meaning of Current and
// Completed while a later dependency is still blocked.
func TestCoordinatorReportsDependencyProgress(t *testing.T) {
	release := make(chan struct{})

	coordinator := New(map[string]Task{
		"character": func(context.Context) error { return nil },
		"world": func(context.Context) error {
			<-release
			return nil
		},
	})
	defer coordinator.Close()

	if err := coordinator.Begin(context.Background(), []string{"character", "world"}); err != nil {
		t.Fatal(err)
	}

	waitFor(t, func(snapshot Snapshot) bool {
		return snapshot.Completed == 1 && snapshot.Current == "world"
	}, coordinator)

	close(release)
	waitFor(t, func(snapshot Snapshot) bool {
		return snapshot.State == "complete" && snapshot.Progress() == 1
	}, coordinator)
}

// TestCoordinatorReportsFailure verifies that the first task error is retained in the observable snapshot instead of
// allowing the generation to appear complete.
func TestCoordinatorReportsFailure(t *testing.T) {
	coordinator := New(map[string]Task{
		"world": func(context.Context) error {
			return errors.New("bad map")
		},
	})
	defer coordinator.Close()

	if err := coordinator.Begin(context.Background(), []string{"world"}); err != nil {
		t.Fatal(err)
	}

	waitFor(t, func(snapshot Snapshot) bool {
		return snapshot.State == "failed" && snapshot.Err != nil
	}, coordinator)
}

// waitFor polls asynchronous coordinator state until a test-specific condition is observable. The short deadline
// keeps a stalled worker from hanging the suite while allowing normal scheduler variation.
func waitFor(t *testing.T, predicate func(Snapshot) bool, coordinator *Coordinator) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if predicate(coordinator.Snapshot()) {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("timed out; snapshot = %#v", coordinator.Snapshot())
}
