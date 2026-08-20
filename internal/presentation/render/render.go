// Package render defines the backend-neutral retained composition model.
package render

import (
	"image/color"
	"sync"
	"time"
)

// Layer is a deterministic top-level composition layer.
type Layer uint8

const (
	// Layers are ordered architectural domains, not arbitrary numeric Z bands.
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

// ResourceKind tells a backend how to interpret CPU-side resource input.
type ResourceKind string

const (
	// Resource kinds name semantic lifetimes; native object types remain private to the selected renderer backend.
	ResourceTexture      ResourceKind = "texture"
	ResourcePalette      ResourceKind = "palette"
	ResourceFont         ResourceKind = "font"
	ResourceAnimation    ResourceKind = "animation"
	ResourceRenderTarget ResourceKind = "render-target"
)

// Resource carries decoded CPU-side input across the owner-thread command boundary.
type Resource struct {
	ID      ResourceID
	Kind    ResourceKind
	Payload any
	// TextureKey identifies immutable pixels across nodes and scene lifetimes, allowing backend residency reuse.
	TextureKey string
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
	ID       NodeID
	Parent   NodeID
	Layer    Layer
	Z        int
	X        float64
	Y        float64
	ScaleX   float64
	ScaleY   float64
	Rotation float64
	OriginX  float64
	OriginY  float64
	Visible  bool
	// Tint supports semantic dimming without changing opacity or authored alpha.
	Tint                  color.RGBA
	Clip                  *Rect
	Blend                 string
	Resource              ResourceID
	Palette               ResourceID
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

// WarmAdmission lets a backend reject speculative uploads without affecting demand-resource correctness.
type WarmAdmission interface {
	CanWarmTexture(key string, weight uint64) bool
}

type slot struct {
	generation uint32
	node       *Node
}

type resourceSlot struct {
	generation uint32
	resource   *Resource
}

// Composer accepts thread-safe retained mutations while reserving backend application for the renderer owner thread.
type Composer struct {
	mu                 sync.Mutex
	slots              []slot
	free               []uint32
	pending            []Change
	warmPending        []Change
	warmKeys           map[string]struct{}
	resources          []resourceSlot
	freeResources      []uint32
	textureUploads     uint64
	textureBytes       uint64
	resourceCreates    uint64
	resourceDestroys   uint64
	structuralRevision uint64
}

// Diagnostics summarizes retained and queued renderer state for leak checks.
type Diagnostics struct {
	ActiveNodes, ActiveResources, Pending, WarmPending, NodeSlots, ResourceSlots int
	WarmPendingBytes                                                             uint64
	RetainedTextureBytes                                                         uint64
	TextureUploads, TextureUploadBytes                                           uint64
	ResourceCreates, ResourceDestroys                                            uint64
	StructuralRevision                                                           uint64
}

// Diagnostics returns one locked snapshot, so its counters always describe the same composer state.
func (c *Composer) Diagnostics() Diagnostics {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := Diagnostics{
		Pending:       len(c.pending),
		WarmPending:   len(c.warmPending),
		NodeSlots:     len(c.slots),
		ResourceSlots: len(c.resources),
	}
	for _, change := range c.warmPending {
		result.WarmPendingBytes += resourceTextureBytes(change.Resource)
	}

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
	result.StructuralRevision = c.structuralRevision

	return result
}
