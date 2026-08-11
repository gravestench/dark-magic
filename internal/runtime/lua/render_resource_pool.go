package modruntime

import (
	"fmt"
	"image"
	"sync"

	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// renderResourcePool owns one managed texture resource for every immutable
// semantic picture. Placement nodes borrow that resource; the final borrower
// destroys it. This saves both retained RGBA memory and duplicate GPU uploads.
type renderResourcePool struct {
	mu       sync.Mutex
	composer *render.Composer
	entries  map[string]*pooledRenderResource
}

type pooledRenderResource struct {
	id   render.ResourceID
	refs int
}

func newRenderResourcePool(composer *render.Composer) *renderResourcePool {
	return &renderResourcePool{composer: composer, entries: make(map[string]*pooledRenderResource)}
}

func (pool *renderResourcePool) acquire(key string, pixels image.Image) (render.ResourceID, func() error, error) {
	if pool == nil || pool.composer == nil || key == "" || pixels == nil {
		return render.ResourceID{}, nil, fmt.Errorf("render resource pool requires composer, key, and pixels")
	}
	pool.mu.Lock()
	entry := pool.entries[key]
	if entry == nil {
		id, err := pool.composer.CreateTexture(pixels, key)
		if err != nil {
			pool.mu.Unlock()
			return render.ResourceID{}, nil, err
		}
		entry = &pooledRenderResource{id: id}
		pool.entries[key] = entry
	}
	entry.refs++
	id := entry.id
	pool.mu.Unlock()
	var once sync.Once
	var releaseErr error
	return id, func() error {
		once.Do(func() { releaseErr = pool.release(key, id) })
		return releaseErr
	}, nil
}

func (pool *renderResourcePool) release(key string, id render.ResourceID) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	entry := pool.entries[key]
	if entry == nil || entry.id != id {
		return nil
	}
	entry.refs--
	if entry.refs > 0 {
		return nil
	}
	delete(pool.entries, key)
	return pool.composer.DestroyResource(id)
}
