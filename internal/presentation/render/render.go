// Package render defines the backend-neutral retained composition model.
package render

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"sort"
	"sync"
	"time"
)

// Layer is a deterministic top-level composition layer.
type Layer uint8

const (
	LayerWorld Layer = iota
	LayerHUD
	LayerModal
	LayerCursor
	LayerDebug
	LayerTransition
)

// NodeID is a generation-checked render-node handle.
type NodeID struct {
	Slot       uint32
	Generation uint32
}

// ResourceID is a generation-checked handle to renderer-owned resource input.
type ResourceID struct {
	Slot       uint32
	Generation uint32
}

type ResourceKind string

const (
	ResourceTexture      ResourceKind = "texture"
	ResourcePalette      ResourceKind = "palette"
	ResourceFont         ResourceKind = "font"
	ResourceAnimation    ResourceKind = "animation"
	ResourceRenderTarget ResourceKind = "render-target"
)

// Resource carries decoded CPU-side input across the renderer command
// boundary. Backends upload and destroy native resources while applying these
// commands on their owner thread.
type Resource struct {
	ID      ResourceID
	Kind    ResourceKind
	Payload any
}

// FontData is decoded font input; native font creation remains a backend task.
type FontData struct {
	Bytes  []byte
	Format string
	Size   int
}

// AnimationData references managed texture frames with per-frame durations.
type AnimationData struct {
	Frames    []ResourceID
	Durations []time.Duration
	Loop      string
}

// RenderTargetData describes a renderer-owned offscreen target.
type RenderTargetData struct{ Width, Height int }

// Node is the backend-neutral retained state of one renderable.
type Node struct {
	ID                    NodeID
	Parent                NodeID
	Layer                 Layer
	Z                     int
	X                     float64
	Y                     float64
	ScaleX                float64
	ScaleY                float64
	Rotation              float64
	Visible               bool
	Clip                  *Rect
	Blend                 string
	Resource              ResourceID
	AnimationPaused       bool
	AnimationSeek         time.Duration
	AnimationSeekRevision uint64
}

// Rect is a clipping rectangle.
type Rect struct{ X, Y, Width, Height float64 }

// Change is one ordered renderer-thread command.
type Change struct {
	Kind       string
	Node       Node
	ID         NodeID
	Resource   Resource
	ResourceID ResourceID
}

// Backend consumes changes on the renderer thread.
type Backend interface {
	Apply(Change) error
}

type slot struct {
	generation uint32
	node       *Node
}

type resourceSlot struct {
	generation uint32
	resource   *Resource
}

// Composer accepts thread-safe retained-state mutations and queues backend
// changes. Drain must be called by the renderer owner thread.
type Composer struct {
	mu               sync.Mutex
	slots            []slot
	free             []uint32
	pending          []Change
	resources        []resourceSlot
	freeResources    []uint32
	textureUploads   uint64
	textureBytes     uint64
	resourceCreates  uint64
	resourceDestroys uint64
}

// Diagnostics summarizes retained and queued renderer state for leak checks.
type Diagnostics struct {
	ActiveNodes, ActiveResources, Pending, NodeSlots, ResourceSlots int
	RetainedTextureBytes                                            uint64
	TextureUploads, TextureUploadBytes                              uint64
	ResourceCreates, ResourceDestroys                               uint64
}

// Diagnostics returns a consistent composer snapshot without exposing payloads.
func (c *Composer) Diagnostics() Diagnostics {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := Diagnostics{Pending: len(c.pending), NodeSlots: len(c.slots), ResourceSlots: len(c.resources)}
	for _, entry := range c.slots {
		if entry.node != nil {
			result.ActiveNodes++
		}
	}
	for _, entry := range c.resources {
		if entry.resource != nil {
			result.ActiveResources++
			result.RetainedTextureBytes += resourceTextureBytes(*entry.resource)
		}
	}
	result.TextureUploads = c.textureUploads
	result.TextureUploadBytes = c.textureBytes
	result.ResourceCreates = c.resourceCreates
	result.ResourceDestroys = c.resourceDestroys
	return result
}

// CreateResource queues decoded input for renderer-thread native creation.
func (c *Composer) CreateResource(kind ResourceKind, payload any) (ResourceID, error) {
	if kind == "" || payload == nil {
		return ResourceID{}, errors.New("rendercore: resource kind and payload are required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.validateResource(kind, payload); err != nil {
		return ResourceID{}, err
	}
	var index uint32
	if len(c.freeResources) > 0 {
		index = c.freeResources[len(c.freeResources)-1]
		c.freeResources = c.freeResources[:len(c.freeResources)-1]
	} else {
		index = uint32(len(c.resources))
		c.resources = append(c.resources, resourceSlot{generation: 1})
	}
	id := ResourceID{Slot: index, Generation: c.resources[index].generation}
	resource := &Resource{ID: id, Kind: kind, Payload: payload}
	c.resources[index].resource = resource
	c.pending = append(c.pending, Change{Kind: "resource-create", Resource: *resource, ResourceID: id})
	c.resourceCreates++
	if bytes := resourceTextureBytes(*resource); bytes > 0 {
		c.textureUploads++
		c.textureBytes += bytes
	}
	return id, nil
}

func (c *Composer) validateResource(kind ResourceKind, payload any) error {
	switch kind {
	case ResourceTexture:
		if _, ok := payload.(image.Image); !ok {
			return fmt.Errorf("rendercore: texture payload is %T, want image.Image", payload)
		}
	case ResourcePalette:
		palette, ok := payload.(color.Palette)
		if !ok || len(palette) == 0 {
			return fmt.Errorf("rendercore: palette payload is %T or empty", payload)
		}
	case ResourceFont:
		font, ok := payload.(FontData)
		if !ok || len(font.Bytes) == 0 || font.Size <= 0 {
			return fmt.Errorf("rendercore: invalid font payload %T", payload)
		}
	case ResourceAnimation:
		animation, ok := payload.(AnimationData)
		if !ok || len(animation.Frames) == 0 || len(animation.Frames) != len(animation.Durations) {
			return fmt.Errorf("rendercore: invalid animation payload %T", payload)
		}
		if animation.Loop != "" && animation.Loop != "loop" && animation.Loop != "once" && animation.Loop != "ping-pong" {
			return fmt.Errorf("rendercore: invalid animation loop mode %q", animation.Loop)
		}
		for index, frame := range animation.Frames {
			resource, err := c.resource(frame)
			if err != nil || resource.Kind != ResourceTexture || animation.Durations[index] <= 0 {
				return fmt.Errorf("rendercore: invalid animation frame %d", index)
			}
		}
	case ResourceRenderTarget:
		target, ok := payload.(RenderTargetData)
		if !ok || target.Width <= 0 || target.Height <= 0 {
			return fmt.Errorf("rendercore: invalid render-target payload %T", payload)
		}
	default:
		return fmt.Errorf("rendercore: unknown resource kind %q", kind)
	}
	return nil
}

// UpdateTexture replaces the CPU-side pixels for a managed texture and queues
// their upload on the renderer owner thread. Video decoders and other streaming
// producers can call this safely without touching a native graphics API.
func (c *Composer) UpdateTexture(id ResourceID, pixels image.Image) error {
	if pixels == nil {
		return errors.New("rendercore: nil texture update")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	resource, err := c.resource(id)
	if err != nil {
		return err
	}
	if resource.Kind != ResourceTexture {
		return fmt.Errorf("rendercore: resource %v is %q, want texture", id, resource.Kind)
	}
	resource.Payload = pixels
	c.pending = append(c.pending, Change{Kind: "resource-update", Resource: *resource, ResourceID: id})
	c.textureUploads++
	c.textureBytes += resourceTextureBytes(*resource)
	return nil
}

// DestroyResource invalidates a managed resource and queues native destruction.
func (c *Composer) DestroyResource(id ResourceID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, err := c.resource(id)
	if err != nil {
		return err
	}
	for _, node := range c.slots {
		if node.node != nil && node.node.Resource == id {
			return fmt.Errorf("rendercore: resource %v is still attached to node %v", id, node.node.ID)
		}
	}
	for _, candidate := range c.resources {
		if candidate.resource == nil || candidate.resource.Kind != ResourceAnimation {
			continue
		}
		animation := candidate.resource.Payload.(AnimationData)
		for _, frame := range animation.Frames {
			if frame == id {
				return fmt.Errorf("rendercore: resource %v is used by animation %v", id, candidate.resource.ID)
			}
		}
	}
	_ = entry
	slot := &c.resources[id.Slot]
	slot.resource = nil
	slot.generation++
	if slot.generation == 0 {
		slot.generation = 1
	}
	c.freeResources = append(c.freeResources, id.Slot)
	c.pending = append(c.pending, Change{Kind: "resource-destroy", ResourceID: id})
	c.resourceDestroys++
	return nil
}

func resourceTextureBytes(resource Resource) uint64 {
	if resource.Kind != ResourceTexture {
		return 0
	}
	pixels, ok := resource.Payload.(image.Image)
	if !ok {
		return 0
	}
	bounds := pixels.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return 0
	}
	return uint64(bounds.Dx()) * uint64(bounds.Dy()) * 4
}

// Create reserves a node and queues its creation.
func (c *Composer) Create(parent NodeID, layer Layer) (NodeID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if parent != (NodeID{}) {
		if _, err := c.node(parent); err != nil {
			return NodeID{}, fmt.Errorf("rendercore: parent: %w", err)
		}
	}
	var index uint32
	if len(c.free) != 0 {
		index = c.free[len(c.free)-1]
		c.free = c.free[:len(c.free)-1]
	} else {
		index = uint32(len(c.slots))
		c.slots = append(c.slots, slot{generation: 1})
	}
	id := NodeID{Slot: index, Generation: c.slots[index].generation}
	node := &Node{ID: id, Parent: parent, Layer: layer, ScaleX: 1, ScaleY: 1, Visible: true, Blend: "alpha"}
	c.slots[index].node = node
	c.pending = append(c.pending, Change{Kind: "create", Node: *node, ID: id})
	return id, nil
}

// Update mutates a node and queues its complete new state.
func (c *Composer) Update(id NodeID, update func(*Node)) error {
	if update == nil {
		return errors.New("rendercore: nil update")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	node, err := c.node(id)
	if err != nil {
		return err
	}
	candidate := *node
	update(&candidate)
	if candidate.Resource != (ResourceID{}) {
		resource, err := c.resource(candidate.Resource)
		if err != nil {
			return fmt.Errorf("rendercore: node resource: %w", err)
		}
		if resource.Kind != ResourceTexture && resource.Kind != ResourceAnimation && resource.Kind != ResourceRenderTarget {
			return fmt.Errorf("rendercore: resource kind %q is not drawable", resource.Kind)
		}
	}
	*node = candidate
	c.pending = append(c.pending, Change{Kind: "update", Node: *node, ID: id})
	return nil
}

func (c *Composer) resource(id ResourceID) (*Resource, error) {
	if int(id.Slot) >= len(c.resources) {
		return nil, fmt.Errorf("rendercore: invalid resource slot %d", id.Slot)
	}
	entry := &c.resources[id.Slot]
	if entry.resource == nil || entry.generation != id.Generation {
		return nil, fmt.Errorf("rendercore: stale resource %v", id)
	}
	return entry.resource, nil
}

// Destroy invalidates id and recursively queues child destruction first.
func (c *Composer) Destroy(id NodeID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.node(id); err != nil {
		return err
	}
	c.destroy(id)
	return nil
}

// Drain applies all currently queued changes in order. A failed change and all
// later changes remain queued for a retry.
func (c *Composer) Drain(backend Backend) error {
	if backend == nil {
		return errors.New("rendercore: nil backend")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for index, change := range c.pending {
		if err := backend.Apply(change); err != nil {
			c.pending = append([]Change(nil), c.pending[index:]...)
			return fmt.Errorf("rendercore: apply %s for node %v: %w", change.Kind, change.ID, err)
		}
	}
	c.pending = nil
	return nil
}

// Snapshot returns visible nodes in deterministic composition order.
func (c *Composer) Snapshot() []Node {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]Node, 0, len(c.slots))
	for _, entry := range c.slots {
		if entry.node != nil && entry.node.Visible {
			result = append(result, *entry.node)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Layer != result[j].Layer {
			return result[i].Layer < result[j].Layer
		}
		if result[i].Z != result[j].Z {
			return result[i].Z < result[j].Z
		}
		return result[i].ID.Slot < result[j].ID.Slot
	})
	return result
}

// ResourceSnapshot returns a checked copy for diagnostics and headless tests.
func (c *Composer) ResourceSnapshot(id ResourceID) (Resource, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	resource, err := c.resource(id)
	if err != nil {
		return Resource{}, err
	}
	return *resource, nil
}

func (c *Composer) node(id NodeID) (*Node, error) {
	if int(id.Slot) >= len(c.slots) {
		return nil, fmt.Errorf("rendercore: invalid node slot %d", id.Slot)
	}
	entry := &c.slots[id.Slot]
	if entry.node == nil || entry.generation != id.Generation {
		return nil, fmt.Errorf("rendercore: stale node %v", id)
	}
	return entry.node, nil
}

func (c *Composer) destroy(id NodeID) {
	for _, entry := range c.slots {
		if entry.node != nil && entry.node.Parent == id {
			c.destroy(entry.node.ID)
		}
	}
	entry := &c.slots[id.Slot]
	entry.node = nil
	entry.generation++
	if entry.generation == 0 {
		entry.generation = 1
	}
	c.free = append(c.free, id.Slot)
	c.pending = append(c.pending, Change{Kind: "destroy", ID: id})
}
