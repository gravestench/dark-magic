package render

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

func TestCreateTextureUsesSemanticIdentityWithoutPixelHashing(t *testing.T) {
	var composer Composer
	pixels := image.NewRGBA(image.Rect(0, 0, 2, 2))
	texture, err := composer.CreateTexture(pixels, "player:run:direction-3:frame-7")
	if err != nil {
		t.Fatal(err)
	}
	resource, err := composer.resource(texture)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := resource.TextureKey, "player:run:direction-3:frame-7"; got != want {
		t.Fatalf("texture key = %q, want %q", got, want)
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

func TestPaletteResourcesRemainAliveWhileAttached(t *testing.T) {
	var composer Composer
	palette, err := composer.CreateResource(ResourcePalette, color.Palette{color.Black})
	if err != nil {
		t.Fatal(err)
	}
	node, err := composer.Create(NodeID{}, LayerHUD)
	if err != nil {
		t.Fatal(err)
	}
	if err := composer.Update(node, func(current *Node) { current.Palette = palette }); err != nil {
		t.Fatal(err)
	}
	if err := composer.DestroyResource(palette); err == nil {
		t.Fatal("destroyed a palette while it was attached to a node")
	}
	if err := composer.Update(node, func(current *Node) { current.Palette = ResourceID{} }); err != nil {
		t.Fatal(err)
	}
	if err := composer.DestroyResource(palette); err != nil {
		t.Fatal(err)
	}
}

func TestTextureUpdatesRemainCheckedAndOrdered(t *testing.T) {
	var composer Composer
	texture, err := composer.CreateResource(ResourceTexture, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if err != nil {
		t.Fatal(err)
	}
	replacement := image.NewRGBA(image.Rect(0, 0, 4, 3))
	if err := composer.UpdateTexture(texture, replacement); err != nil {
		t.Fatal(err)
	}
	snapshot, err := composer.ResourceSnapshot(texture)
	if err != nil || snapshot.Payload != replacement {
		t.Fatalf("updated resource = %#v, %v", snapshot, err)
	}
	backend := &recordingBackend{}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if got := []string{backend.changes[0].Kind, backend.changes[1].Kind}; !reflect.DeepEqual(got, []string{"resource-create", "resource-update"}) {
		t.Fatalf("changes = %v", got)
	}
	if err := composer.UpdateTexture(ResourceID{Slot: texture.Slot, Generation: texture.Generation + 1}, replacement); err == nil {
		t.Fatal("updated a stale texture handle")
	}
	font, err := composer.CreateResource(ResourceFont, FontData{Bytes: []byte("font"), Format: "ttf", Size: 12})
	if err != nil {
		t.Fatal(err)
	}
	if err := composer.UpdateTexture(font, replacement); err == nil {
		t.Fatal("updated a non-texture resource")
	}
}

func TestNoOpNodeUpdatesAreNotQueued(t *testing.T) {
	var composer Composer
	node, err := composer.Create(NodeID{}, LayerWorld)
	if err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	backend.changes = nil
	if err := composer.Update(node, func(current *Node) { current.Visible = true }); err != nil {
		t.Fatal(err)
	}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.changes) != 0 {
		t.Fatalf("no-op update queued %d changes", len(backend.changes))
	}
}

func TestStreamingTextureUpdatesKeepOnlyNewestPendingFrame(t *testing.T) {
	var composer Composer
	first := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	texture, err := composer.CreateResource(ResourceTexture, first)
	if err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	backend.changes = nil
	second := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	newest := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	if err := composer.UpdateTexture(texture, second); err != nil {
		t.Fatal(err)
	}
	if err := composer.UpdateTexture(texture, newest); err != nil {
		t.Fatal(err)
	}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.changes) != 1 || backend.changes[0].Resource.Payload != newest {
		t.Fatalf("streaming updates = %#v, want newest frame only", backend.changes)
	}
}

func TestDiagnosticsTrackTextureResidencyAndUploadVolume(t *testing.T) {
	composer := &Composer{}
	texture, err := composer.CreateResource(ResourceTexture, image.NewRGBA(image.Rect(0, 0, 4, 3)))
	if err != nil {
		t.Fatal(err)
	}
	if err := composer.UpdateTexture(texture, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	diagnostics := composer.Diagnostics()
	if diagnostics.RetainedTextureBytes != 16 || diagnostics.TextureUploads != 2 || diagnostics.TextureUploadBytes != 64 || diagnostics.ResourceCreates != 1 {
		t.Fatalf("diagnostics after upload = %#v", diagnostics)
	}
	if err := composer.DestroyResource(texture); err != nil {
		t.Fatal(err)
	}
	diagnostics = composer.Diagnostics()
	if diagnostics.RetainedTextureBytes != 0 || diagnostics.ResourceDestroys != 1 {
		t.Fatalf("diagnostics after destroy = %#v", diagnostics)
	}
}

func BenchmarkComposerNoOpNodeUpdates(b *testing.B) {
	var composer Composer
	nodes := make([]NodeID, 512)
	for index := range nodes {
		var err error
		nodes[index], err = composer.Create(NodeID{}, LayerWorld)
		if err != nil {
			b.Fatal(err)
		}
	}
	backend := &recordingBackend{}
	if err := composer.Drain(backend); err != nil {
		b.Fatal(err)
	}
	backend.changes = nil
	queued := 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, id := range nodes {
			if err := composer.Update(id, func(node *Node) { node.Visible = true }); err != nil {
				b.Fatal(err)
			}
		}
		if err := composer.Drain(backend); err != nil {
			b.Fatal(err)
		}
		queued += len(backend.changes)
		backend.changes = backend.changes[:0]
	}
	b.StopTimer()
	b.ReportMetric(float64(queued)/float64(b.N), "queued-changes/op")
}

func BenchmarkComposerStreamingTextureCoalescing(b *testing.B) {
	var composer Composer
	texture, err := composer.CreateResource(ResourceTexture, image.NewNRGBA(image.Rect(0, 0, 640, 480)))
	if err != nil {
		b.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := composer.Drain(backend); err != nil {
		b.Fatal(err)
	}
	backend.changes = nil
	queued := 0
	frames := []*image.NRGBA{
		image.NewNRGBA(image.Rect(0, 0, 640, 480)),
		image.NewNRGBA(image.Rect(0, 0, 640, 480)),
		image.NewNRGBA(image.Rect(0, 0, 640, 480)),
		image.NewNRGBA(image.Rect(0, 0, 640, 480)),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, frame := range frames {
			if err := composer.UpdateTexture(texture, frame); err != nil {
				b.Fatal(err)
			}
		}
		if err := composer.Drain(backend); err != nil {
			b.Fatal(err)
		}
		queued += len(backend.changes)
		backend.changes = backend.changes[:0]
	}
	b.StopTimer()
	b.ReportMetric(float64(queued)/float64(b.N), "queued-changes/op")
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

func TestComposerExistsRejectsDestroyedGeneration(t *testing.T) {
	var composer Composer
	node, err := composer.Create(NodeID{}, LayerWorld)
	if err != nil {
		t.Fatal(err)
	}
	if !composer.Exists(node) {
		t.Fatal("new node does not exist")
	}
	if err := composer.Destroy(node); err != nil {
		t.Fatal(err)
	}
	if composer.Exists(node) {
		t.Fatal("destroyed node still exists")
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

func TestWarmTextureQueueDeduplicatesAndRespectsBudget(t *testing.T) {
	composer := &Composer{}
	backend := &recordingBackend{}
	first := image.NewRGBA(image.Rect(0, 0, 4, 4))
	second := image.NewRGBA(image.Rect(0, 0, 8, 8))
	first.Pix[0], second.Pix[0] = 1, 2
	if composer.WarmTexture(first) != composer.WarmTexture(first) {
		t.Fatal("stable texture key changed")
	}
	composer.WarmTexture(second)
	if got := composer.Diagnostics().WarmPending; got != 2 {
		t.Fatalf("warm pending = %d", got)
	}
	if err := composer.DrainWarm(backend, 64); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().WarmPending; got != 1 {
		t.Fatalf("warm pending after budget = %d", got)
	}
	if err := composer.DrainWarm(backend, 64); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().WarmPending; got != 0 {
		t.Fatalf("warm pending after drain = %d", got)
	}
}

type refusingWarmBackend struct{ recordingBackend }

func (*refusingWarmBackend) CanWarmTexture(string, uint64) bool { return false }

func TestWarmTextureQueuePreservesWorkRejectedByResidencyPolicy(t *testing.T) {
	composer := &Composer{}
	composer.WarmTexture(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	backend := &refusingWarmBackend{}
	if err := composer.DrainWarm(backend, 1024); err != nil {
		t.Fatal(err)
	}
	if len(backend.changes) != 0 || composer.Diagnostics().WarmPending != 1 {
		t.Fatalf("rejected warm work was applied or discarded: changes=%d diagnostics=%#v", len(backend.changes), composer.Diagnostics())
	}
}

func TestStructuralRevisionIgnoresOrdinaryNodeUpdates(t *testing.T) {
	composer := &Composer{}
	id, err := composer.Create(NodeID{}, LayerModal)
	if err != nil {
		t.Fatal(err)
	}
	created := composer.Diagnostics().StructuralRevision
	if created == 0 {
		t.Fatal("node creation did not advance structural revision")
	}
	if err := composer.Update(id, func(node *Node) { node.X = 12 }); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().StructuralRevision; got != created {
		t.Fatalf("ordinary update changed structural revision: got %d, want %d", got, created)
	}
	resource, err := composer.CreateResource(ResourceTexture, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().StructuralRevision; got != created {
		t.Fatalf("resource refresh changed node topology revision: got %d, want %d", got, created)
	}
	if err := composer.DestroyResource(resource); err != nil {
		t.Fatal(err)
	}
	if err := composer.Destroy(id); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().StructuralRevision; got != created+1 {
		t.Fatalf("destroy revision = %d, want %d", got, created+1)
	}
}
