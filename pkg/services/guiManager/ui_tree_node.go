package guiManager

import (
	"image"
	"image/color"

	"github.com/fogleman/gg"
	"github.com/google/uuid"
)

// YieldsImage represents an interface for nodes that yield an image.
type YieldsImage interface {
	Image() image.Image
}

// TreeNode represents a node in the GUI tree.
type TreeNode struct {
	uuid uuid.UUID

	X, Y     float64
	Canvas   *gg.Context
	children []node

	parent  node
	enabled bool

	inputVector  InputState
	InputHandler func()

	imageFunc  func() image.Image
	updateFunc func()
}

// NewTreeNode creates a new tree node.
func NewTreeNode(x, y float64) *TreeNode {
	canvas := gg.NewContext(int(x), int(y))
	return &TreeNode{
		X:            x,
		Y:            y,
		Canvas:       canvas,
		children:     []node{},
		enabled:      true,
		InputHandler: nil,
		uuid:         uuid.New(),
	}
}

func (n *TreeNode) UUID() uuid.UUID {
	return n.uuid
}

// Implement inputHandler methods
func (n *TreeNode) InputVector() InputState {
	return n.inputVector
}

func (n *TreeNode) SetInputVector(state InputState) {
	n.inputVector = state
}

func (n *TreeNode) ClearInputVectors() {
	n.inputVector = InputState{}
}

func (n *TreeNode) Handler() func() {
	return n.InputHandler
}

func (n *TreeNode) SetHandler(handler func()) {
	n.InputHandler = handler
}

// Implement element methods
func (n *TreeNode) Enable() {
	n.enabled = true
}

func (n *TreeNode) Disable() {
	n.enabled = false
}

func (n *TreeNode) IsEnabled() bool {
	return n.enabled
}

func (n *TreeNode) Position() image.Point {
	return image.Point{int(n.X), int(n.Y)}
}

func (n *TreeNode) SetPosition(point image.Point) {
	n.X = float64(point.X)
	n.Y = float64(point.Y)
}

// Other element methods (Scale, Opacity, Origin, Rotation) can be implemented similarly.

// Implement node methods
func (n *TreeNode) Parent() node {
	return n.parent
}

func (n *TreeNode) SetParent(parent node) {
	n.parent = parent
}

func (n *TreeNode) Children() []node {
	return n.children
}

func (n *TreeNode) AddChild(child node) {
	n.children = append(n.children, child)
	child.SetParent(n)
}

func (n *TreeNode) RemoveChild(child node) {
	for i, c := range n.children {
		if c == child {
			n.children = append(n.children[:i], n.children[i+1:]...)
			return
		}
	}
}

func (n *TreeNode) LayerIndexOf(child node) int {
	for i, c := range n.children {
		if c == child {
			return i
		}
	}
	return -1
}

func (n *TreeNode) SetLayerIndexOf(child node, index int) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex == -1 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(n.children) {
		index = len(n.children) - 1
	}
	n.children = append(n.children[:currentIndex], n.children[currentIndex+1:]...)
	n.children = append(n.children[:index], append([]node{child}, n.children[index:]...)...)
}

// BringToTop moves the specified child node to the top of the z-order.
func (n *TreeNode) BringToTop(child node) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex == -1 {
		return
	}
	n.RemoveChild(child)
	n.AddChild(child)
}

// BringToBottom moves the specified child node to the bottom of the z-order.
func (n *TreeNode) BringToBottom(child node) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex == -1 {
		return
	}
	n.RemoveChild(child)
	n.children = append([]node{child}, n.children...)
}

// Raise moves the specified child node one layer up in the z-order.
func (n *TreeNode) Raise(child node) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex > 0 {
		n.SetLayerIndexOf(child, currentIndex-1)
	}
}

// Lower moves the specified child node one layer down in the z-order.
func (n *TreeNode) Lower(child node) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex < len(n.children)-1 {
		n.SetLayerIndexOf(child, currentIndex+1)
	}
}

// Implement HandleInput method
func (n *TreeNode) HandleInput(state InputState) bool {
	// Handle input here
	// You can access state.Keys, state.ModifierKeys, state.MouseButtons, and state.MouseCursor
	return false // Example: Return true to indicate termination
}

// Image composites child images on top of its own canvas.
func (n *TreeNode) Image() image.Image {
	n.Canvas.SetColor(color.Transparent)
	n.Canvas.Clear()

	for _, child := range n.children {
		if child.ImageFunc() == nil {
			continue
		}

		childImage := child.ImageFunc()()
		if childImage != nil {
			p := child.Position()
			n.Canvas.DrawImage(childImage, p.X, p.Y)
		}
	}

	return n.Canvas.Image()
}

func (n *TreeNode) ImageFunc() func() image.Image {
	return n.imageFunc
}

func (n *TreeNode) SetImageFunc(f func() image.Image) {
	n.imageFunc = f
}

func (n *TreeNode) Update() {
	if n.updateFunc != nil {
		n.updateFunc()
	}

	for _, child := range n.children {
		child.Update()
	}
}

func (n *TreeNode) UpdateFunc() func() {
	return n.updateFunc
}

func (n *TreeNode) SetUpdateFunc(f func()) {
	n.updateFunc = f
}

// GetRelativePosition calculates the relative position of a given node within the subtree.
func (n *TreeNode) GetRelativePosition(target node) (relativePosition image.Point, found bool) {
	// Check if the target node is the current node
	if n == target {
		return image.Point{0, 0}, true
	}

	// Recursively search for the target node in children
	for _, child := range n.children {
		if childRelativePos, found := child.GetRelativePosition(target); found {
			// If the target node is found in a child subtree, calculate the relative position
			childPosition := child.Position()
			return image.Point{
				childRelativePos.X + childPosition.X - int(n.X),
				childRelativePos.Y + childPosition.Y - int(n.Y),
			}, true
		}
	}

	// If the target node is not found in the subtree, return (0,0) and false
	return image.Point{}, false
}
