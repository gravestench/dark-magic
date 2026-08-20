package render

import (
	"errors"
	"fmt"
	"time"
)

// Drain applies queued demand changes in order and retains the failed command plus its successors for retry.
func (c *Composer) Drain(backend Backend) error {
	if backend == nil {
		return errors.New("rendercore: nil backend")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for index, change := range c.pending {
		if err := backend.Apply(change); err != nil {
			// Copy the retry tail so it no longer retains already applied commands through the old backing array.
			c.pending = append([]Change(nil), c.pending[index:]...)

			return fmt.Errorf("rendercore: apply %s for node %v: %w", change.Kind, change.ID, err)
		}
	}

	c.pending = nil

	return nil
}

// DrainWarm uploads optional residency work by byte budget while guaranteeing progress for one oversized texture.
func (c *Composer) DrainWarm(backend Backend, byteBudget uint64) error {
	return c.DrainWarmWithin(backend, byteBudget, 0)
}

// DrainWarmWithin adds a wall-time budget without weakening first-item progress or backend residency admission.
func (c *Composer) DrainWarmWithin(backend Backend, byteBudget uint64, timeBudget time.Duration) error {
	if backend == nil {
		return errors.New("rendercore: nil backend")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	started := time.Now()
	used, count := uint64(0), 0

	for count < len(c.warmPending) {
		if count > 0 && timeBudget > 0 && time.Since(started) >= timeBudget {
			break
		}

		change := c.warmPending[count]

		weight := resourceTextureBytes(change.Resource)
		if admission, ok := backend.(WarmAdmission); ok &&
			!admission.CanWarmTexture(change.Resource.TextureKey, weight) {
			break
		}

		if count > 0 && byteBudget > 0 && used+weight > byteBudget {
			break
		}

		if err := backend.Apply(change); err != nil {
			// Match demand draining: retain the failed item and every later item for an ordered retry.
			c.warmPending = append([]Change(nil), c.warmPending[count:]...)

			return fmt.Errorf("rendercore: warm texture: %w", err)
		}

		delete(c.warmKeys, change.Resource.TextureKey)

		used += weight
		count++
	}

	if count > 0 {
		c.warmPending = append([]Change(nil), c.warmPending[count:]...)
	}

	return nil
}
