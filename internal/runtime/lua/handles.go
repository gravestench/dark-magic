package modruntime

import (
	"fmt"
	"sync"
)

// Handle is a stable, generation-checked reference to a native object.
type Handle struct {
	Type       string
	Slot       uint32
	Generation uint32
}

type handleSlot struct {
	typeName   string
	generation uint32
	value      any
}

// Handles stores native objects without exposing their ownership to Lua.
type Handles struct {
	mu    sync.RWMutex
	slots []handleSlot
	free  []uint32
}

// Add stores value and returns a checked handle.
func (h *Handles) Add(typeName string, value any) (Handle, error) {
	if typeName == "" || value == nil {
		return Handle{}, fmt.Errorf("modruntime: handle type and value are required")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var slot uint32
	if len(h.free) != 0 {
		slot = h.free[len(h.free)-1]
		h.free = h.free[:len(h.free)-1]
	} else {
		slot = uint32(len(h.slots))
		h.slots = append(h.slots, handleSlot{generation: 1})
	}

	entry := &h.slots[slot]
	entry.typeName = typeName
	entry.value = value

	return Handle{Type: typeName, Slot: slot, Generation: entry.generation}, nil
}

// Get resolves handle and rejects stale or mistyped references.
func (h *Handles) Get(handle Handle) (any, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if int(handle.Slot) >= len(h.slots) {
		return nil, fmt.Errorf("modruntime: invalid %s handle slot %d", handle.Type, handle.Slot)
	}

	entry := h.slots[handle.Slot]
	if entry.value == nil || entry.generation != handle.Generation ||
		entry.typeName != handle.Type {
		return nil, fmt.Errorf("modruntime: stale or invalid %s handle", handle.Type)
	}

	return entry.value, nil
}

// Release invalidates handle and makes its slot reusable with a new generation.
func (h *Handles) Release(handle Handle) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if int(handle.Slot) >= len(h.slots) {
		return fmt.Errorf("modruntime: invalid %s handle slot %d", handle.Type, handle.Slot)
	}

	entry := &h.slots[handle.Slot]
	if entry.value == nil || entry.generation != handle.Generation ||
		entry.typeName != handle.Type {
		return fmt.Errorf("modruntime: stale or invalid %s handle", handle.Type)
	}

	entry.value = nil
	entry.typeName = ""

	entry.generation++
	if entry.generation == 0 {
		entry.generation = 1
	}

	h.free = append(h.free, handle.Slot)

	return nil
}
