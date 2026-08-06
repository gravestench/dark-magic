package rendercore

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

type recordingBackend struct {
	changes []Change
	fail    error
}

func (b *recordingBackend) Apply(change Change) error {
	if b.fail != nil {
		err := b.fail
		b.fail = nil
		return err
	}
	b.changes = append(b.changes, change)
	return nil
}

func TestComposerQueuesRetainedChangesAndRejectsStaleNodes(t *testing.T) {
	t.Parallel()

	var composer Composer
	parent, err := composer.Create(NodeID{}, LayerWorld)
	if err != nil {
		t.Fatal(err)
	}
	child, err := composer.Create(parent, LayerHUD)
	if err != nil {
		t.Fatal(err)
	}
	if err := composer.Update(child, func(node *Node) { node.X, node.Y, node.Z = 10, 20, 3 }); err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if got := []string{backend.changes[0].Kind, backend.changes[1].Kind, backend.changes[2].Kind}; !reflect.DeepEqual(got, []string{"create", "create", "update"}) {
		t.Fatalf("changes = %v", got)
	}
	if err := composer.Destroy(parent); err != nil {
		t.Fatal(err)
	}
	if err := composer.Update(child, func(*Node) {}); err == nil {
		t.Fatal("expected child handle to be stale")
	}
	replacement, err := composer.Create(NodeID{}, LayerWorld)
	if err != nil {
		t.Fatal(err)
	}
	if replacement.Generation == parent.Generation && replacement.Slot == parent.Slot {
		t.Fatalf("replacement did not advance generation: %#v", replacement)
	}
}

func TestComposerOrdersLayersThenZThenCreation(t *testing.T) {
	t.Parallel()

	var composer Composer
	modal, _ := composer.Create(NodeID{}, LayerModal)
	worldHigh, _ := composer.Create(NodeID{}, LayerWorld)
	worldLow, _ := composer.Create(NodeID{}, LayerWorld)
	_ = composer.Update(worldHigh, func(node *Node) { node.Z = 10 })
	_ = composer.Update(worldLow, func(node *Node) { node.Z = -2 })
	snapshot := composer.Snapshot()
	got := []NodeID{snapshot[0].ID, snapshot[1].ID, snapshot[2].ID}
	want := []NodeID{worldLow, worldHigh, modal}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestComposerRetainsFailedDrainForRetry(t *testing.T) {
	t.Parallel()

	var composer Composer
	_, _ = composer.Create(NodeID{}, LayerWorld)
	backend := &recordingBackend{fail: errors.New("GPU unavailable")}
	if err := composer.Drain(backend); err == nil {
		t.Fatal("expected drain failure")
	}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.changes) != 1 {
		t.Fatalf("applied changes = %d", len(backend.changes))
	}
}

func TestComposerAcceptsConcurrentSubmission(t *testing.T) {
	t.Parallel()

	var composer Composer
	var wait sync.WaitGroup
	for i := 0; i < 32; i++ {
		wait.Add(1)
		go func(z int) {
			defer wait.Done()
			id, err := composer.Create(NodeID{}, LayerWorld)
			if err != nil {
				t.Errorf("Create: %v", err)
				return
			}
			if err := composer.Update(id, func(node *Node) { node.Z = z }); err != nil {
				t.Errorf("Update: %v", err)
			}
		}(i)
	}
	wait.Wait()
	if got := len(composer.Snapshot()); got != 32 {
		t.Fatalf("node count = %d", got)
	}
}
