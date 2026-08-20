package video

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// Presenter transfers decoded video frames into the retained composition. It
// owns one texture and one transition-layer node for its entire lifetime.
type Presenter struct {
	mu       sync.Mutex
	composer *render.Composer
	texture  render.ResourceID
	node     render.NodeID
	viewport image.Point
	closed   bool
}

// NewPresenter creates a black cinematic surface fitted into viewport. When both
// retained objects exist, failed construction destroys the node before its texture.
func NewPresenter(composer *render.Composer, frameSize, viewport image.Point) (*Presenter, error) {
	if composer == nil {
		return nil, errors.New("videocore: nil composer")
	}

	if frameSize.X <= 0 || frameSize.Y <= 0 || viewport.X <= 0 || viewport.Y <= 0 {
		return nil, fmt.Errorf("videocore: invalid frame %v or viewport %v", frameSize, viewport)
	}

	texture, err := composer.CreateResource(render.ResourceTexture, image.NewRGBA(image.Rectangle{Max: frameSize}))
	if err != nil {
		return nil, err
	}

	node, err := composer.Create(render.NodeID{}, render.LayerTransition)
	if err != nil {
		// The texture exists before the node and must not leak when node creation fails.
		_ = composer.DestroyResource(texture)
		return nil, err
	}

	p := &Presenter{composer: composer, texture: texture, node: node, viewport: viewport}
	if err := p.fit(frameSize, viewport); err != nil {
		// Destroy retained references before their resource handle becomes invalid.
		_ = composer.Destroy(node)
		_ = composer.DestroyResource(texture)

		return nil, err
	}

	return p, nil
}

// Present refits and queues the newest decoded frame while serializing against
// Resize and Close, preventing updates to a destroyed texture or node.
func (p *Presenter) Present(frame image.Image) error {
	if frame == nil {
		return errors.New("videocore: nil frame")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errors.New("videocore: presenter is closed")
	}

	if err := p.fit(frame.Bounds().Size(), p.viewport); err != nil {
		return err
	}

	return p.composer.UpdateTexture(p.texture, frame)
}

// Resize updates letterboxing without changing the decoded frame or resource;
// the existing payload supplies frame dimensions after asynchronous uploads.
func (p *Presenter) Resize(viewport image.Point) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errors.New("videocore: presenter is closed")
	}

	if viewport.X <= 0 || viewport.Y <= 0 {
		return fmt.Errorf("videocore: invalid viewport %v", viewport)
	}

	p.viewport = viewport

	resource, err := p.composer.ResourceSnapshot(p.texture)
	if err != nil {
		return err
	}

	return p.fit(resource.Payload.(image.Image).Bounds().Size(), viewport)
}

// fit centers the frame with a uniform minimum-axis scale, preserving aspect
// ratio and leaving letterbox space on the remaining viewport axis.
func (p *Presenter) fit(frameSize, viewport image.Point) error {
	scaleX := float64(viewport.X) / float64(frameSize.X)
	scaleY := float64(viewport.Y) / float64(frameSize.Y)
	scale := min(scaleX, scaleY)

	return p.composer.Update(p.node, func(node *render.Node) {
		node.Resource = p.texture
		// Raylib composition nodes use a centered origin. Position the center of
		// the fitted frame at the center of the live render viewport.
		node.X = float64(viewport.X) / 2
		node.Y = float64(viewport.Y) / 2
		node.ScaleX = scale
		node.ScaleY = scale
		node.Visible = true
	})
}

// Close removes the cinematic node before invalidating its texture handle and
// joins cleanup failures so neither retained resource is silently leaked.
func (p *Presenter) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	return errors.Join(p.composer.Destroy(p.node), p.composer.DestroyResource(p.texture))
}
