package raylibRenderer

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Origin returns the normalized pivot applied to scaled texture dimensions during drawing.
func (n *node) Origin() rl.Vector2 {
	return n.origin
}

// SetOrigin updates the normalized draw pivot without invalidating world transforms.
func (n *node) SetOrigin(x, y float64) {
	n.origin.X = float32(x)
	n.origin.Y = float32(y)
}

// ZIndex returns the local matrix slot used exclusively for stable sibling ordering.
func (n *node) ZIndex() float32 {
	return n.local.M14
}

// SetZIndex invalidates the parent ordering cache only when the effective local Z value changes.
func (n *node) SetZIndex(index float32) {
	if n.local.M14 == index {
		return
	}

	n.local.M14 = index

	n.transformDirty = true
	if n.parent != nil {
		n.parent.childrenSorted = false
	}
}

// Position returns world-space translation after parent transforms have propagated.
func (n *node) Position() (x, y float32) {
	return n.world.M12, n.world.M13
}

// SetPosition updates local translation and defers world propagation until the frame update traversal.
func (n *node) SetPosition(x, y float32) {
	if n.local.M12 == x && n.local.M13 == y {
		return
	}

	n.local.M12 = x
	n.local.M13 = y
	n.transformDirty = true
}

// Rotation returns the cached world-space angle derived during transform propagation.
func (n *node) Rotation() float32 {
	return n.worldRotation
}

// matrixRotation removes X scale before extracting the world Z rotation in degrees.
func matrixRotation(matrix rl.Matrix) float32 {
	scaleX := math.Sqrt(float64(matrix.M0*matrix.M0 + matrix.M1*matrix.M1 + matrix.M2*matrix.M2))
	if scaleX == 0 {
		return 0
	}

	m0Prime := matrix.M0 / float32(scaleX)
	m1Prime := matrix.M1 / float32(scaleX)
	theta := math.Atan2(float64(m1Prime), float64(m0Prime))

	return float32(theta * 180.0 / math.Pi)
}

// SetRotation rebuilds local X/Y rotation while preserving the matrix's existing scale and translation components.
func (n *node) SetRotation(degrees float32) {
	scaleX := math.Sqrt(float64(n.local.M0*n.local.M0 + n.local.M1*n.local.M1 + n.local.M2*n.local.M2))
	scaleY := math.Sqrt(float64(n.local.M4*n.local.M4 + n.local.M5*n.local.M5 + n.local.M6*n.local.M6))
	scaleZ := math.Sqrt(float64(n.local.M8*n.local.M8 + n.local.M9*n.local.M9 + n.local.M10*n.local.M10))

	translationX := n.local.M12
	translationY := n.local.M13
	translationZ := n.local.M14

	radians := float64(degrees) * math.Pi / 180.0
	cosine := math.Cos(radians)
	sine := math.Sin(radians)
	n.local.M0 = float32(cosine) * float32(scaleX)
	n.local.M1 = float32(sine) * float32(scaleX)
	n.local.M4 = float32(-sine) * float32(scaleY)
	n.local.M5 = float32(cosine) * float32(scaleY)

	// Preserve the established Z normalization even though renderer transforms are otherwise two-dimensional.
	n.local.M8 = n.local.M8 / float32(scaleZ)
	n.local.M9 = n.local.M9 / float32(scaleZ)
	n.local.M10 = float32(scaleZ)
	n.local.M12 = translationX
	n.local.M13 = translationY
	n.local.M14 = translationZ
	n.transformDirty = true
}

// Scale returns the historical world M0 component used by uniform-scale callers.
func (n *node) Scale() float32 {
	return n.world.M0
}

// scaleXY returns independent world-axis scale components used by non-uniform retained nodes.
func (n *node) scaleXY() (float32, float32) {
	return n.world.M0, n.world.M5
}

// SetScale applies one uniform local value through the same non-uniform update path.
func (n *node) SetScale(scale float32) {
	n.setScaleXY(scale, scale)
}

// setScaleXY updates local X/Y scale and preserves the 2D matrix's unit Z scale.
func (n *node) setScaleXY(x, y float32) {
	if n.local.M0 == x && n.local.M5 == y {
		return
	}

	n.local.M0 = x
	n.local.M5 = y
	n.local.M10 = 1
	n.transformDirty = true
}

// WorldMatrix returns the last propagated world transform without mutating local state.
func (n *node) WorldMatrix() rl.Matrix {
	return n.world
}

// LocalMatrix returns the node-local transform used by the next propagation pass.
func (n *node) LocalMatrix() rl.Matrix {
	return n.local
}
