package guiManager

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

const layerIndexNotFound = -1

type TreeNode struct {
	parent      Node
	children    []Node
	renderable  raylibRenderer.Renderable
	inputVector InputState
	updateFunc  func()
	handler     func()
	enabled     bool
}

func (s *Service) NewTreeNode() *TreeNode {
	return &TreeNode{
		children:   make([]Node, 0),
		renderable: s.renderer.NewRenderable(), // Initialize with nil Renderable
		enabled:    true,
	}
}

func (tn *TreeNode) Parent() Node {
	return tn.parent
}

func (tn *TreeNode) SetParent(parent Node) {
	if parent == tn {
		return
	}

	tn.parent = parent

	if tn.parent == nil {
		return
	}

	if tn.parent.ContainsChild(tn) {
		return
	}

	tn.parent.AddChild(tn)
}

func (tn *TreeNode) Children() []Node {
	return tn.children
}

func (tn *TreeNode) AddChild(child Node) {
	if child == tn || child == nil {
		return
	}

	tn.children = append(tn.children, child)
	child.SetParent(tn)
}

func (tn *TreeNode) RemoveChild(child Node) {
	for i, c := range tn.children {
		if c == child {
			tn.children = append(tn.children[:i], tn.children[i+1:]...)
			return
		}
	}
}

func (tn *TreeNode) LayerIndexOf(child Node) int {
	for i, c := range tn.children {
		if c == child {
			return i
		}
	}
	return -1
}

func (tn *TreeNode) SetLayerIndexOf(child Node, index int) {
	if index < 0 || index >= len(tn.children) {
		return
	}

	currentIndex := tn.LayerIndexOf(child)
	if currentIndex == layerIndexNotFound {
		return // Child not found
	}

	if currentIndex == index {
		return // Child already at the desired layer
	}

	// Remove the child from its current position
	tn.children = append(tn.children[:currentIndex], tn.children[currentIndex+1:]...)

	// Insert the child at the new layer index
	tn.children = append(tn.children[:index], append([]Node{child}, tn.children[index:]...)...)
}

func (tn *TreeNode) ContainsChild(child Node) bool {
	return tn.LayerIndexOf(child) != layerIndexNotFound
}

func (tn *TreeNode) BringToTop(child Node) {
	tn.SetLayerIndexOf(child, len(tn.children)-1)
}

func (tn *TreeNode) BringToBottom(child Node) {
	tn.SetLayerIndexOf(child, 0)
}

func (tn *TreeNode) Raise(child Node) {
	currentIndex := tn.LayerIndexOf(child)
	if currentIndex > 0 {
		tn.SetLayerIndexOf(child, currentIndex-1)
	}
}

func (tn *TreeNode) Lower(child Node) {
	currentIndex := tn.LayerIndexOf(child)
	if currentIndex < len(tn.children)-1 {
		tn.SetLayerIndexOf(child, currentIndex+1)
	}
}

func (tn *TreeNode) HandleInput(input InputState) (terminate bool) {
	if !tn.IsEnabled() {
		return false
	}

	for _, child := range tn.children {
		if child.HandleInput(input) {
			return true
		}
	}

	return false
}

func (tn *TreeNode) Update() {
	if tn.updateFunc != nil {
		tn.updateFunc()
	}

	for _, child := range tn.children {
		child.Update()
	}
}

func (tn *TreeNode) UpdateFunc() func() {
	return tn.updateFunc
}

func (tn *TreeNode) SetUpdateFunc(updateFunc func()) {
	tn.updateFunc = updateFunc
}

func (tn *TreeNode) GetRelativePosition(target Node) (relativePosition image.Point, found bool) {
	if tn == target {
		return image.Point{}, true
	}

	for _, child := range tn.children {
		if pos, ok := child.GetRelativePosition(target); ok {
			return pos, true
		}
	}

	return image.Point{}, false
}

func (tn *TreeNode) Image() image.Image {
	if tn.renderable != nil {
		return tn.renderable.Image()
	}
	return nil
}

func (tn *TreeNode) SetImage(image image.Image) {
	if tn.renderable != nil {
		tn.renderable.SetImage(image)
	}
}

func (tn *TreeNode) UUID() uuid.UUID {
	if tn.renderable != nil {
		return tn.renderable.UUID()
	}
	return uuid.Nil
}

func (tn *TreeNode) Enable() {
	tn.enabled = true
}

func (tn *TreeNode) Disable() {
	tn.enabled = false
}

func (tn *TreeNode) IsEnabled() bool {
	return tn.enabled
}

func (tn *TreeNode) InputVector() InputState {
	return tn.inputVector
}

func (tn *TreeNode) SetInputVector(input InputState) {
	tn.inputVector = input
}

func (tn *TreeNode) ClearInputVectors() {
	tn.inputVector = InputState{}
}

func (tn *TreeNode) Handler() func() {
	if tn.handler == nil {
		return func() {}
	}

	return tn.handler
}

func (tn *TreeNode) SetHandler(handler func()) {
	tn.handler = handler
}

func (tn *TreeNode) ZIndex() float32 {
	return tn.renderable.ZIndex()
}

func (tn *TreeNode) SetZIndex(i float32) {
	tn.renderable.SetZIndex(i)
}

func (tn *TreeNode) Position() (x, y float32) {
	return tn.renderable.Position()
}

func (tn *TreeNode) SetPosition(x, y float32) {
	tn.renderable.SetPosition(x, y)
}

func (tn *TreeNode) Rotation() (degrees float32) {
	return tn.renderable.Rotation()
}

func (tn *TreeNode) SetRotation(degrees float32) {
	tn.renderable.SetRotation(degrees)
}

func (tn *TreeNode) Scale() (scale float32) {
	return tn.renderable.Scale()
}

func (tn *TreeNode) SetScale(scale float32) {
	tn.renderable.SetScale(scale)
}

func (tn *TreeNode) Opacity() (opacity float32) {
	return tn.renderable.Opacity()
}

func (tn *TreeNode) SetOpacity(opacity float32) {
	tn.renderable.SetOpacity(opacity)
}

func (tn *TreeNode) BlendMode() (mode rl.BlendMode) {
	return tn.renderable.BlendMode()
}

func (tn *TreeNode) SetBlendMode(mode rl.BlendMode) {
	tn.renderable.SetBlendMode(mode)
}

func (tn *TreeNode) Texture() rl.Texture2D {
	return tn.renderable.Texture()
}

func (tn *TreeNode) SetTexture(d rl.Texture2D) {
	tn.renderable.SetTexture(d)
}
