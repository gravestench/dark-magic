package raylibRenderer

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TestSortedChildrenCachesAndInvalidatesZOrder protects stable ordering and explicit cache invalidation on Z changes.
func TestSortedChildrenCachesAndInvalidatesZOrder(t *testing.T) {
	parent := &node{}
	back := &node{}
	front := &node{}

	back.SetZIndex(1)
	front.SetZIndex(10)
	parent.addChild(front)
	parent.addChild(back)

	children := parent.sortedChildren()
	if children[0] != back || children[1] != front || !parent.childrenSorted {
		t.Fatalf("children were not sorted: %#v", children)
	}

	front.parent = parent
	front.SetZIndex(0)

	if parent.childrenSorted {
		t.Fatal("changing child Z-index did not invalidate parent ordering")
	}

	children = parent.sortedChildren()
	if children[0] != front || children[1] != back {
		t.Fatalf("children were not re-sorted: %#v", children)
	}
}

// TestAddingAndRemovingChildrenInvalidatesOrdering verifies hierarchy mutations cannot reuse stale sibling order.
func TestAddingAndRemovingChildrenInvalidatesOrdering(t *testing.T) {
	parent := &node{childrenSorted: true}
	child := &node{}
	parent.addChild(child)

	if parent.childrenSorted {
		t.Fatal("adding a child did not invalidate ordering")
	}

	parent.sortedChildren()
	parent.removeChild(child)

	if parent.childrenSorted {
		t.Fatal("removing a child did not invalidate ordering")
	}
}

// TestCachedChildOrderingDoesNotAllocate protects the allocation-free per-frame hierarchy traversal path.
func TestCachedChildOrderingDoesNotAllocate(t *testing.T) {
	parent := &node{}
	for index := 0; index < 16; index++ {
		parent.addChild(&node{local: matrixWithZ(float32(index))})
	}

	parent.sortedChildren()

	if allocations := testing.AllocsPerRun(1000, func() { _ = parent.sortedChildren() }); allocations != 0 {
		t.Fatalf("cached ordering allocations = %v", allocations)
	}
}

// matrixWithZ creates the minimal local transform needed to isolate ordering behavior.
func matrixWithZ(z float32) (matrix rl.Matrix) {
	matrix.M14 = z
	return matrix
}
