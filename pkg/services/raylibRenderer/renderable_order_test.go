package raylibRenderer

import "testing"

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
