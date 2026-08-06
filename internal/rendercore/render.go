// Package rendercore defines the backend-neutral retained composition model.
package rendercore

import (
	"errors"
	"fmt"
	"image"
	"sort"
	"sync"
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
	Visible  bool
	Clip     *Rect
	Blend    string
	Image    image.Image
}

// Rect is a clipping rectangle.
type Rect struct{ X, Y, Width, Height float64 }

// Change is one ordered renderer-thread command.
type Change struct {
	Kind string
	Node Node
	ID   NodeID
}

// Backend consumes changes on the renderer thread.
type Backend interface {
	Apply(Change) error
}

type slot struct {
	generation uint32
	node       *Node
}

// Composer accepts thread-safe retained-state mutations and queues backend
// changes. Drain must be called by the renderer owner thread.
type Composer struct {
	mu      sync.Mutex
	slots   []slot
	free    []uint32
	pending []Change
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
	update(node)
	c.pending = append(c.pending, Change{Kind: "update", Node: *node, ID: id})
	return nil
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
