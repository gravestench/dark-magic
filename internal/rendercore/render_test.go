package rendercore

import (
	"errors"
	"image"
	"image/color"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestAllManagedResourceKindsUseCheckedHandles(t *testing.T) {
	var composer Composer
	texture, err := composer.CreateResource(ResourceTexture, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if err != nil {
		t.Fatal(err)
	}
	resources := []ResourceID{texture}
	for _, input := range []struct {
		kind    ResourceKind
		payload any
	}{
		{ResourcePalette, color.Palette{color.Black, color.White}},
		{ResourceFont, FontData{Bytes: []byte("font"), Format: "ttf", Size: 16}},
		{ResourceAnimation, AnimationData{Frames: []ResourceID{texture}, Durations: []time.Duration{time.Second / 10}}},
		{ResourceRenderTarget, RenderTargetData{Width: 320, Height: 200}},
	} {
		id, err := composer.CreateResource(input.kind, input.payload)
		if err != nil {
			t.Fatalf("create %s: %v", input.kind, err)
		}
		resources = append(resources, id)
		got, err := composer.ResourceSnapshot(id)
		if err != nil || got.Kind != input.kind {
			t.Fatalf("snapshot %s: %#v, %v", input.kind, got, err)
		}
	}
	if err := composer.DestroyResource(texture); err == nil {
		t.Fatal("texture used by animation was destroyed")
	}
	if err := composer.DestroyResource(resources[3]); err != nil {
		t.Fatal(err)
	}
	if err := composer.DestroyResource(texture); err != nil {
		t.Fatal(err)
	}
	for _, id := range []ResourceID{resources[1], resources[2], resources[4]} {
		if err := composer.DestroyResource(id); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagedResourcePayloadValidation(t *testing.T) {
	var composer Composer
	for _, input := range []struct {
		kind    ResourceKind
		payload any
	}{
		{ResourceTexture, "not an image"}, {ResourcePalette, color.Palette{}},
		{ResourceFont, FontData{}}, {ResourceAnimation, AnimationData{}}, {ResourceRenderTarget, RenderTargetData{}},
	} {
		if _, err := composer.CreateResource(input.kind, input.payload); err == nil {
			t.Errorf("accepted invalid %s", input.kind)
		}
	}
}

func TestAnimationLoopModeValidation(t *testing.T) {
	var composer Composer
	texture, err := composer.CreateResource(ResourceTexture, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"", "loop", "once", "ping-pong"} {
		animation, err := composer.CreateResource(ResourceAnimation, AnimationData{
			Frames: []ResourceID{texture}, Durations: []time.Duration{time.Millisecond}, Loop: mode,
		})
		if err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
		if err := composer.DestroyResource(animation); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := composer.CreateResource(ResourceAnimation, AnimationData{
		Frames: []ResourceID{texture}, Durations: []time.Duration{time.Millisecond}, Loop: "random",
	}); err == nil {
		t.Fatal("accepted unsupported loop mode")
	}
}

type recordingBackend struct {
	changes []Change
	fail    error
}

func TestManagedResourcesAreCheckedAndDrainInOwnershipOrder(t *testing.T) {
	var composer Composer
	resource, err := composer.CreateResource(ResourceTexture, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if err != nil {
		t.Fatal(err)
	}
	node, err := composer.Create(NodeID{}, LayerHUD)
	if err != nil {
		t.Fatal(err)
	}
	if err := composer.Update(node, func(current *Node) { current.Resource = resource }); err != nil {
		t.Fatal(err)
	}
	if err := composer.DestroyResource(resource); err == nil {
		t.Fatal("destroyed attached resource")
	}
	if err := composer.Destroy(node); err != nil {
		t.Fatal(err)
	}
	if err := composer.DestroyResource(resource); err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, change := range backend.changes {
		kinds = append(kinds, change.Kind)
	}
	want := []string{"resource-create", "create", "update", "destroy", "resource-destroy"}
	if !reflect.DeepEqual(kinds, want) {
		t.Fatalf("changes = %v, want %v", kinds, want)
	}
	if _, err := composer.ResourceSnapshot(resource); err == nil {
		t.Fatal("stale resource remained valid")
	}
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

func TestEmptyDrainHotPathDoesNotAllocate(t *testing.T) {
	var composer Composer
	backend := &recordingBackend{}
	allocations := testing.AllocsPerRun(1000, func() {
		if err := composer.Drain(backend); err != nil {
			panic(err)
		}
	})
	if allocations != 0 {
		t.Fatalf("empty drain allocations = %v", allocations)
	}
}
