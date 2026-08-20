package raylibRenderer

import (
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Enable marks this node drawable while leaving retained visibility and child traversal unchanged.
func (n *node) Enable() {
	n.enabled = true
}

// Disable suppresses this node's draw call without hiding or pruning its child subtree.
func (n *node) Disable() {
	n.enabled = false
}

// IsEnabled reports whether this node itself should draw after visibility and culling checks.
func (n *node) IsEnabled() bool {
	return n.enabled
}

// setVisible controls subtree traversal; invisible nodes prune every descendant regardless of enabled state.
func (n *node) setVisible(visible bool) {
	n.visible = visible
}

// parentNode returns the scene root for an unattached node, retaining the historical public hierarchy view.
func (n *node) parentNode() *node {
	if n.parent == nil {
		return n.renderer.rootNode
	}

	return n.parent
}

// setParent detaches before attaching so each node appears in at most one private native child list.
func (n *node) setParent(parent *node) {
	if n == parent {
		return
	}

	if n.parent != nil {
		n.parent.removeChild(n)
	}

	n.parent = parent

	n.transformDirty = true
	if parent != nil {
		n.parent.addChild(n)
	}
}

// addChild appends in authored order and invalidates the stable Z-order cache.
func (n *node) addChild(child *node) {
	if child == nil {
		return
	}

	n.children = append(n.children, child)
	n.childrenSorted = false
}

// removeChild removes every matching entry from the end so accidental duplicates cannot survive detachment.
func (n *node) removeChild(child *node) {
	if child == nil {
		return
	}

	for index := len(n.children) - 1; index >= 0; index-- {
		if n.children[index] != child {
			continue
		}

		n.children = append(n.children[:index], n.children[index+1:]...)
		n.childrenSorted = false
	}
}

// Children returns the private child slice for legacy renderer-internal callers without allocating a copy.
func (n *node) Children() []*node {
	return n.children
}

// sortedChildren performs one stable in-place Z sort after hierarchy or Z changes, then reuses it without allocation.
func (n *node) sortedChildren() []*node {
	if n.childrenSorted {
		return n.children
	}

	// Stable ordering preserves authored insertion order for siblings with identical Z values.
	sort.SliceStable(n.children, func(first, second int) bool {
		return n.children[first].ZIndex() < n.children[second].ZIndex()
	})
	n.childrenSorted = true

	return n.children
}

// UpdateWorldMatrix propagates dirty ancestry once and caches world rotation for culling and drawing.
func (n *node) UpdateWorldMatrix(parent rl.Matrix, parentDirty bool) {
	dirty := parentDirty || n.transformDirty
	if dirty {
		n.world = rl.MatrixMultiply(n.local, parent)
		n.worldRotation = matrixRotation(n.world)
		n.transformDirty = false
	}

	for index := range n.children {
		n.children[index].UpdateWorldMatrix(n.world, dirty)
	}
}
