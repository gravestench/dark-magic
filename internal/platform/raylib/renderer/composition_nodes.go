package raylibRenderer

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// createNode publishes the native node before applying state, preserving error recovery visibility for later changes.
func (b *compositionBackend) createNode(id render.NodeID, state render.Node) error {
	if _, exists := b.nodes[id]; exists {
		return fmt.Errorf("node %v already exists", id)
	}

	node := b.renderer.newNode()
	b.nodes[id] = node

	return b.applyNode(node, state)
}

// updateNode applies a complete retained snapshot only to an existing native node.
func (b *compositionBackend) updateNode(id render.NodeID, state render.Node) error {
	node, exists := b.nodes[id]
	if !exists {
		return fmt.Errorf("node %v does not exist", id)
	}

	return b.applyNode(node, state)
}

// destroyNode detaches hierarchy ownership before clearing textures and semantic resource associations.
func (b *compositionBackend) destroyNode(id render.NodeID) error {
	node, exists := b.nodes[id]
	if !exists {
		return fmt.Errorf("node %v does not exist", id)
	}

	node.Disable()
	node.setParent(nil)
	node.ClearTextures()
	delete(b.nodes, id)
	delete(b.nodeResources, id)
	delete(b.playbacks, id)

	return nil
}

// applyNode mirrors one complete retained state in dependency order: hierarchy, style, resources, then playback.
func (b *compositionBackend) applyNode(node *node, state render.Node) error {
	if err := b.applyNodeParent(node, state.Parent); err != nil {
		return err
	}

	applyNodeTransform(node, state)
	applyNodeTint(node, state.Tint)
	applyNodeClip(node, state.Clip)

	if err := applyNodeBlend(node, state.Blend); err != nil {
		return err
	}

	node.SetZIndex(float32(int(state.Layer)*1_000_000 + state.Z))

	if err := b.applyNodePalette(node, state.Palette); err != nil {
		return err
	}

	if err := b.applyNodeResource(node, state); err != nil {
		return err
	}

	applyNodeVisibility(node, state)
	b.applyAnimationState(node, state)

	return nil
}

// applyNodeParent changes hierarchy only when the retained state names a parent, matching grouping-node semantics.
func (b *compositionBackend) applyNodeParent(node *node, parentID render.NodeID) error {
	if parentID == (render.NodeID{}) {
		return nil
	}

	parent, exists := b.nodes[parentID]
	if !exists {
		return fmt.Errorf("parent node %v does not exist", parentID)
	}

	node.setParent(parent)

	return nil
}

// applyNodeTransform writes local geometry before Z ordering so parent propagation sees one coherent snapshot.
func applyNodeTransform(node *node, state render.Node) {
	node.SetPosition(float32(state.X), float32(state.Y))
	node.setScaleXY(float32(state.ScaleX), float32(state.ScaleY))
	node.SetRotation(float32(state.Rotation))
	node.SetOrigin(state.OriginX, state.OriginY)
}

// applyNodeTint defaults an absent tint to white and keeps RGB independent from the separate opacity channel.
func applyNodeTint(node *node, tint color.RGBA) {
	if tint.A == 0 {
		tint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	node.SetTint(rl.NewColor(tint.R, tint.G, tint.B, 255))
}

// applyNodeClip copies retained clip values into Raylib storage so no caller-owned pointer reaches the native node.
func applyNodeClip(node *node, clipState *render.Rect) {
	if clipState == nil {
		node.SetClip(nil)
		return
	}

	clip := rl.NewRectangle(
		float32(clipState.X),
		float32(clipState.Y),
		float32(clipState.Width),
		float32(clipState.Height),
	)
	node.SetClip(&clip)
}

// applyNodeBlend maps retained blend names to Raylib modes while reserving custom mode for Diablo screen blending.
func applyNodeBlend(node *node, blend string) error {
	switch blend {
	case "", "alpha":
		node.SetBlendMode(rl.BlendAlpha)
	case "additive":
		node.SetBlendMode(rl.BlendAdditive)
	case "screen":
		node.SetBlendMode(rl.BlendCustom)
	case "multiply":
		node.SetBlendMode(rl.BlendMultiplied)
	case "add-colors":
		node.SetBlendMode(rl.BlendAddColors)
	case "subtract-colors":
		node.SetBlendMode(rl.BlendSubtractColors)
	default:
		return fmt.Errorf("unsupported blend mode %q", blend)
	}

	return nil
}

// applyNodePalette lazily allocates one GPU effect per palette resource and shares it across attached nodes.
func (b *compositionBackend) applyNodePalette(node *node, paletteID render.ResourceID) error {
	if paletteID == (render.ResourceID{}) {
		node.SetShader(nil, nil, 0)
		return nil
	}

	effect := b.paletteEffects[paletteID]
	if effect == nil {
		resource, exists := b.resources[paletteID]
		if !exists || resource.Kind != render.ResourcePalette {
			return fmt.Errorf("palette resource %v is unavailable", paletteID)
		}

		var err error

		effect, err = newGPUPaletteEffect(resource.Payload.(color.Palette))
		if err != nil {
			return err
		}

		b.paletteEffects[paletteID] = effect
	}

	node.SetShader(&effect.shader, &effect.texture, effect.textureLocation)

	return nil
}

// applyNodeResource validates drawable state before replacing a node's texture or animation ownership.
func (b *compositionBackend) applyNodeResource(node *node, state render.Node) error {
	if state.Resource == (render.ResourceID{}) {
		return nil
	}

	resource, exists := b.resources[state.Resource]
	if !exists {
		return fmt.Errorf("resource %v does not exist", state.Resource)
	}

	decoded, err := b.drawableImage(resource)
	if err != nil {
		return err
	}

	if b.nodeResources[state.ID] == state.Resource {
		return nil
	}

	node.ClearTextures()

	if resource.Kind == render.ResourceAnimation {
		if err := b.attachAnimation(state.ID, node, resource); err != nil {
			return err
		}
	} else {
		node.SetImageResource(decoded, resource.TextureKey)
		delete(b.playbacks, state.ID)
	}

	b.nodeResources[state.ID] = state.Resource

	return nil
}

// applyNodeVisibility keeps resource-less grouping transforms disabled so Raylib never draws their default texture.
func applyNodeVisibility(node *node, state render.Node) {
	if state.Visible && state.Resource != (render.ResourceID{}) {
		node.Enable()
	} else {
		node.Disable()
	}

	node.setVisible(state.Visible)
}

// applyAnimationState applies pause before revision-gated seek, preserving playback timing and explicit seek order.
func (b *compositionBackend) applyAnimationState(node *node, state render.Node) {
	playback := b.playbacks[state.ID]
	if playback == nil {
		return
	}

	playback.player.SetPaused(state.AnimationPaused)

	if playback.seekRevision == state.AnimationSeekRevision {
		return
	}

	frame, changed := playback.player.Seek(state.AnimationSeek)

	playback.seekRevision = state.AnimationSeekRevision
	if changed {
		b.setAnimationFrame(node, playback.frames[frame], playback.keys[frame], frame)
	}
}
