package render

import (
	"errors"
	"fmt"
	"image/color"
	"sort"
)

// Create reserves a generation-checked node and queues its complete initial state for backend ownership.
func (c *Composer) Create(parent NodeID, layer Layer) (NodeID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if parent != (NodeID{}) {
		if _, err := c.node(parent); err != nil {
			return NodeID{}, fmt.Errorf("rendercore: parent: %w", err)
		}
	}

	index := c.allocateNodeSlot()
	id := NodeID{Slot: index, Generation: c.slots[index].generation}
	node := &Node{
		ID:      id,
		Parent:  parent,
		Layer:   layer,
		ScaleX:  1,
		ScaleY:  1,
		OriginX: 0.5,
		OriginY: 0.5,
		Visible: true,
		Tint:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Blend:   "alpha",
	}

	c.slots[index].node = node
	c.pending = append(c.pending, Change{Kind: "create", Node: *node, ID: id})
	c.structuralRevision++

	return id, nil
}

// allocateNodeSlot reuses the most recently destroyed slot while retaining its incremented stale-handle guard.
func (c *Composer) allocateNodeSlot() uint32 {
	if len(c.free) != 0 {
		index := c.free[len(c.free)-1]
		c.free = c.free[:len(c.free)-1]

		return index
	}

	index := uint32(len(c.slots))
	c.slots = append(c.slots, slot{generation: 1})

	return index
}

// Update validates a candidate copy before publishing it, so rejected changes cannot partially mutate retained state.
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

	if candidate == *node {
		return nil
	}

	if err := c.validateNodeResources(candidate); err != nil {
		return err
	}

	*node = candidate
	c.pending = append(c.pending, Change{Kind: "update", Node: *node, ID: id})

	return nil
}

// validateNodeResources keeps drawable and palette handles generation-checked before a state update enters the queue.
func (c *Composer) validateNodeResources(candidate Node) error {
	if candidate.Resource != (ResourceID{}) {
		resource, err := c.resource(candidate.Resource)
		if err != nil {
			return fmt.Errorf("rendercore: node resource: %w", err)
		}

		if resource.Kind != ResourceTexture && resource.Kind != ResourceAnimation &&
			resource.Kind != ResourceRenderTarget {
			return fmt.Errorf("rendercore: resource kind %q is not drawable", resource.Kind)
		}
	}

	if candidate.Palette != (ResourceID{}) {
		resource, err := c.resource(candidate.Palette)
		if err != nil {
			return fmt.Errorf("rendercore: node palette: %w", err)
		}

		if resource.Kind != ResourcePalette {
			return fmt.Errorf("rendercore: node palette resource kind is %q", resource.Kind)
		}
	}

	return nil
}

// Destroy recursively queues children before their parent so backends never retain nodes with missing ancestry.
func (c *Composer) Destroy(id NodeID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.node(id); err != nil {
		return err
	}

	c.destroy(id)
	c.structuralRevision++

	return nil
}

// Exists reports whether a retained handle still names the same live generation, enabling idempotent teardown.
func (c *Composer) Exists(id NodeID) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.node(id)

	return err == nil
}

// Snapshot copies visible nodes and sorts the copy without exposing or reordering retained slot storage.
func (c *Composer) Snapshot() []Node {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]Node, 0, len(c.slots))
	for _, entry := range c.slots {
		if entry.node != nil && entry.node.Visible {
			result = append(result, *entry.node)
		}
	}

	// Stable layer, Z, and creation ordering makes backend output deterministic when values tie.
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

// node resolves a generation-checked handle while callers hold the composer lock.
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

// destroy performs child-first invalidation under the caller-held lock and preserves slot-scan ordering.
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
