package content

import "sync"

// Change reports that content at Path may resolve differently.
// Generation increases monotonically for the lifetime of its layered filesystem.
type Change struct {
	Path       string
	Generation uint64
}

// Generation returns the latest invalidation generation under the same lock used to publish changes.
func (f *FS) Generation() uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.generation
}

// Invalidate publishes a normalized development-time content change without caching any content bytes itself.
// Consumers use the generation to invalidate decoded records, required Lua modules, and other derived resources.
func (f *FS) Invalidate(name string) (Change, error) {
	clean, err := Normalize(name)
	if err != nil {
		return Change{}, err
	}

	f.mu.Lock()
	f.generation++

	change := Change{Path: clean, Generation: f.generation}
	for _, subscriber := range f.subscribers {
		// Notifications are diagnostic and best-effort; the generation remains authoritative for a slow consumer.
		select {
		case subscriber <- change:
		default:
		}
	}
	f.mu.Unlock()

	return change, nil
}

// Subscribe returns a best-effort change stream and an idempotent cancellation function.
// A minimum capacity of one lets invalidation remain non-blocking even when the caller requests an unbuffered stream.
func (f *FS) Subscribe(buffer int) (<-chan Change, func()) {
	if buffer < 1 {
		buffer = 1
	}

	f.mu.Lock()
	id := f.nextSubID
	f.nextSubID++
	changes := make(chan Change, buffer)
	f.subscribers[id] = changes
	f.mu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			f.removeSubscriber(id, changes)
		})
	}

	return changes, cancel
}

// removeSubscriber deletes and closes a stream under the publication lock so Invalidate cannot send after close.
func (f *FS) removeSubscriber(id uint64, changes chan Change) {
	f.mu.Lock()
	delete(f.subscribers, id)
	close(changes)
	f.mu.Unlock()
}
