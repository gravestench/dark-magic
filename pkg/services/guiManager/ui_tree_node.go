package guiManager

import (
	"image"
	"image/color"

	"github.com/fogleman/gg"
	"github.com/google/uuid"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

// YieldsImage represents an interface for nodes that yield an image.
type YieldsImage interface {
	Image() image.Image
}

// TreeNode represents a Node in the GUI tree.
type TreeNode struct {
	raylibRenderer.Renderable

	X, Y     float64
	Canvas   *gg.Context
	children []Node

	parent  Node
	enabled bool

	inputVector  InputState
	InputHandler func()

	imageFunc  func() image.Image
	updateFunc func()
}

// NewTreeNode creates a new tree Node.
func (s *Service) NewTreeNode(x, y float64) *TreeNode {
	canvas := gg.NewContext(int(x), int(y))

	node := &TreeNode{
		Renderable:   s.renderer.NewRenderable(),
		X:            x,
		Y:            y,
		Canvas:       canvas,
		children:     []Node{},
		enabled:      true,
		InputHandler: nil,
	}

	return node
}

func (n *TreeNode) UUID() uuid.UUID {
	return n.Renderable.UUID()
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

// Implement Node methods
func (n *TreeNode) Parent() Node {
	return n.parent
}

func (n *TreeNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *TreeNode) Children() []Node {
	return n.children
}

func (n *TreeNode) AddChild(child Node) {
	n.children = append(n.children, child)
	child.SetParent(n)
}

func (n *TreeNode) RemoveChild(child Node) {
	for i, c := range n.children {
		if c == child {
			n.children = append(n.children[:i], n.children[i+1:]...)
			return
		}
	}
}

func (n *TreeNode) LayerIndexOf(child Node) int {
	for i, c := range n.children {
		if c == child {
			return i
		}
	}
	return -1
}

func (n *TreeNode) SetLayerIndexOf(child Node, index int) {
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
	n.children = append(n.children[:index], append([]Node{child}, n.children[index:]...)...)
}

// BringToTop moves the specified child Node to the top of the z-order.
func (n *TreeNode) BringToTop(child Node) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex == -1 {
		return
	}
	n.RemoveChild(child)
	n.AddChild(child)
}

// BringToBottom moves the specified child Node to the bottom of the z-order.
func (n *TreeNode) BringToBottom(child Node) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex == -1 {
		return
	}
	n.RemoveChild(child)
	n.children = append([]Node{child}, n.children...)
}

// Raise moves the specified child Node one layer up in the z-order.
func (n *TreeNode) Raise(child Node) {
	currentIndex := n.LayerIndexOf(child)
	if currentIndex > 0 {
		n.SetLayerIndexOf(child, currentIndex-1)
	}
}

// Lower moves the specified child Node one layer down in the z-order.
func (n *TreeNode) Lower(child Node) {
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
	if point, found := n.GetRelativePosition(n.Parent()); found {
		n.Renderable.SetPosition(point.X, point.X)
	}

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

// GetRelativePosition calculates the relative position of a given Node within the subtree.
func (n *TreeNode) GetRelativePosition(target Node) (relativePosition image.Point, found bool) {
	// Check if the target Node is the current Node
	if n == target {
		return image.Point{0, 0}, true
	}

	// Recursively search for the target Node in children
	for _, child := range n.children {
		if childRelativePos, found := child.GetRelativePosition(target); found {
			// If the target Node is found in a child subtree, calculate the relative position
			childPosition := child.Position()
			return image.Point{
				childRelativePos.X + childPosition.X - int(n.X),
				childRelativePos.Y + childPosition.Y - int(n.Y),
			}, true
		}
	}

	// If the target Node is not found in the subtree, return (0,0) and false
	return image.Point{}, false
}
