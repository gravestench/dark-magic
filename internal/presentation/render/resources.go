package render

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/color"
)

// CreateResource validates decoded input and queues native creation for the renderer owner thread.
func (c *Composer) CreateResource(kind ResourceKind, payload any) (ResourceID, error) {
	return c.createResource(kind, payload, "")
}

// CreateTexture accepts a semantic identity so asset pipelines can avoid hashing already identified pixels.
func (c *Composer) CreateTexture(pixels image.Image, key string) (ResourceID, error) {
	if key == "" {
		return ResourceID{}, errors.New("rendercore: semantic texture key is required")
	}

	return c.createResource(ResourceTexture, pixels, key)
}

// createResource reserves a generation-checked slot and queues creation while the composer lock protects both states.
func (c *Composer) createResource(kind ResourceKind, payload any, textureKey string) (ResourceID, error) {
	if kind == "" || payload == nil {
		return ResourceID{}, errors.New("rendercore: resource kind and payload are required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.validateResource(kind, payload); err != nil {
		return ResourceID{}, err
	}

	index := c.allocateResourceSlot()
	id := ResourceID{Slot: index, Generation: c.resources[index].generation}

	resource := &Resource{ID: id, Kind: kind, Payload: payload}
	if kind == ResourceTexture {
		resource.TextureKey = textureKey
		if resource.TextureKey == "" {
			resource.TextureKey = TextureKey(payload.(image.Image))
		}
	}

	c.resources[index].resource = resource
	c.pending = append(c.pending, Change{Kind: "resource-create", Resource: *resource, ResourceID: id})
	c.resourceCreates++

	if bytes := resourceTextureBytes(*resource); bytes > 0 {
		c.textureUploads++
		c.textureBytes += bytes
	}

	return id, nil
}

// allocateResourceSlot reuses the most recently freed slot and never exposes a stale generation to callers.
func (c *Composer) allocateResourceSlot() uint32 {
	if len(c.freeResources) > 0 {
		index := c.freeResources[len(c.freeResources)-1]
		c.freeResources = c.freeResources[:len(c.freeResources)-1]

		return index
	}

	index := uint32(len(c.resources))
	c.resources = append(c.resources, resourceSlot{generation: 1})

	return index
}

// TextureKey hashes dimensions and visible RGBA rows so differently shaped or padded images cannot alias.
func TextureKey(pixels image.Image) string {
	if pixels == nil {
		return ""
	}

	bounds := pixels.Bounds()
	hash := sha256.New()

	_, _ = fmt.Fprintf(hash, "%d:%d:", bounds.Dx(), bounds.Dy())
	if rgba, ok := pixels.(*image.RGBA); ok {
		width := bounds.Dx() * 4
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			start := rgba.PixOffset(bounds.Min.X, y)
			_, _ = hash.Write(rgba.Pix[start : start+width])
		}

		return fmt.Sprintf("rgba:%x", hash.Sum(nil))
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := pixels.At(x, y).RGBA()
			hash.Write([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)})
		}
	}

	return fmt.Sprintf("rgba:%x", hash.Sum(nil))
}

// WarmTexture queues optional native residency work without making scene correctness depend on preloading.
func (c *Composer) WarmTexture(pixels image.Image) string {
	return c.WarmTextureKey(TextureKey(pixels), pixels)
}

// WarmTextureKey deduplicates speculative work by semantic identity before it reaches the owner-thread queue.
func (c *Composer) WarmTextureKey(key string, pixels image.Image) string {
	if key == "" {
		return ""
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.warmKeys == nil {
		c.warmKeys = make(map[string]struct{})
	}

	if _, exists := c.warmKeys[key]; exists {
		return key
	}

	c.warmKeys[key] = struct{}{}
	c.warmPending = append(c.warmPending, Change{
		Kind: "texture-warm",
		Resource: Resource{
			Kind:       ResourceTexture,
			Payload:    pixels,
			TextureKey: key,
		},
	})

	return key
}

// validateResource rejects payloads a backend cannot safely interpret and checks animation references under the lock.
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
		return c.validateAnimation(payload)
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

// validateAnimation prevents native playback from observing stale frames, invalid durations, or unknown loop modes.
func (c *Composer) validateAnimation(payload any) error {
	animation, ok := payload.(AnimationData)
	if !ok || len(animation.Frames) == 0 || len(animation.Frames) != len(animation.Durations) {
		return fmt.Errorf("rendercore: invalid animation payload %T", payload)
	}

	if animation.Loop != "" && animation.Loop != "loop" && animation.Loop != "once" &&
		animation.Loop != "ping-pong" {
		return fmt.Errorf("rendercore: invalid animation loop mode %q", animation.Loop)
	}

	for index, frame := range animation.Frames {
		resource, err := c.resource(frame)
		if err != nil || resource.Kind != ResourceTexture || animation.Durations[index] <= 0 {
			return fmt.Errorf("rendercore: invalid animation frame %d", index)
		}
	}

	return nil
}

// UpdateTexture keeps only the newest adjacent update while preserving ordering across other commands for the resource.
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

	// Stop at the first command for this resource: coalescing across it would reorder owner-thread effects.
	for index := len(c.pending) - 1; index >= 0; index-- {
		pending := &c.pending[index]
		if pending.ResourceID != id {
			continue
		}

		if pending.Kind == "resource-update" {
			pending.Resource = *resource
			return nil
		}

		break
	}

	c.pending = append(c.pending, Change{Kind: "resource-update", Resource: *resource, ResourceID: id})
	c.textureUploads++
	c.textureBytes += resourceTextureBytes(*resource)

	return nil
}

// DestroyResource refuses attached inputs before invalidating their generation and queuing native destruction.
func (c *Composer) DestroyResource(id ResourceID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.resource(id); err != nil {
		return err
	}

	if err := c.validateResourceUnused(id); err != nil {
		return err
	}

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

// validateResourceUnused checks node attachments before animation dependencies to retain historical error precedence.
func (c *Composer) validateResourceUnused(id ResourceID) error {
	for _, node := range c.slots {
		if node.node == nil {
			continue
		}

		if node.node.Resource == id {
			return fmt.Errorf("rendercore: resource %v is still attached to node %v", id, node.node.ID)
		}

		if node.node.Palette == id {
			return fmt.Errorf("rendercore: palette %v is still attached to node %v", id, node.node.ID)
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

	return nil
}

// resourceTextureBytes estimates native RGBA residency without traversing or converting pixel data.
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

// ResourceSnapshot returns a checked copy so diagnostics cannot mutate retained resource state.
func (c *Composer) ResourceSnapshot(id ResourceID) (Resource, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resource, err := c.resource(id)
	if err != nil {
		return Resource{}, err
	}

	return *resource, nil
}

// resource resolves a generation-checked handle while callers hold the composer lock.
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
