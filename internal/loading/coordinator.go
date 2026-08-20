// Package loading coordinates engine-owned work required by scene transitions.
package loading

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Task loads one named dependency and must stop promptly when its context is
// cancelled. Tasks own their actual resources; the coordinator owns only order,
// cancellation, and observable progress.
type Task func(context.Context) error

// Snapshot is a copied, renderer-neutral view of one loading generation.
type Snapshot struct {
	State     string
	Completed int
	Total     int
	Current   string
	Err       error
}

// Progress returns normalized completed work. An idle coordinator has no total
// yet and therefore reports zero rather than pretending to be complete.
func (s Snapshot) Progress() float64 {
	if s.Total == 0 {
		return 0
	}

	return float64(s.Completed) / float64(s.Total)
}

// Coordinator runs explicitly requested dependency sets one generation at a
// time. Beginning new work cancels the previous generation so a stale worker
// cannot overwrite the progress observed by a newer scene transition.
type Coordinator struct {
	mu         sync.RWMutex
	tasks      map[string]Task
	generation uint64
	cancel     context.CancelFunc
	snapshot   Snapshot
}

// New copies the task registry; later caller map edits cannot change the set of
// dependencies admitted by this coordinator.
func New(tasks map[string]Task) *Coordinator {
	copyTasks := make(map[string]Task, len(tasks))
	for id, task := range tasks {
		copyTasks[id] = task
	}

	return &Coordinator{tasks: copyTasks, snapshot: Snapshot{State: "idle"}}
}

// Begin validates and starts one ordered loading generation. Validation happens before the current generation is
// cancelled, so a malformed replacement request cannot interrupt useful work already in progress.
func (c *Coordinator) Begin(parent context.Context, ids []string) error {
	selected, err := c.selectTasks(ids)
	if err != nil {
		return err
	}

	// Swapping the generation while holding the lock makes the new snapshot and cancellation boundary visible as one
	// state change. An older worker may finish later, but publish rejects its obsolete generation number.
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}

	ctx, cancel := context.WithCancel(parent)
	c.cancel = cancel
	c.generation++
	generation := c.generation
	c.snapshot = Snapshot{State: "running", Total: len(ids), Current: ids[0]}
	c.mu.Unlock()

	go c.run(ctx, generation, append([]string(nil), ids...), selected)

	return nil
}

// selectTasks resolves dependency names without mutating coordinator state. Keeping this phase separate preserves the
// currently running generation when the caller supplies an empty or unknown dependency list.
func (c *Coordinator) selectTasks(ids []string) ([]Task, error) {
	if len(ids) == 0 {
		return nil, errors.New("loading: at least one dependency is required")
	}

	selected := make([]Task, len(ids))
	for index, id := range ids {
		task, exists := c.tasks[id]
		if !exists || task == nil {
			return nil, fmt.Errorf("loading: dependency %q is not registered", id)
		}

		selected[index] = task
	}

	return selected, nil
}

// Snapshot returns the latest generation state by value. Callers may retain the result without holding the
// coordinator lock, while Err continues to carry the original task error for errors.Is and errors.As checks.
func (c *Coordinator) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.snapshot
}

// Close cancels active work and advances the generation so a task that ignores cancellation cannot publish stale
// progress after shutdown.
func (c *Coordinator) Close() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	c.generation++
	c.mu.Unlock()
}

// run executes dependencies in request order because later tasks may consume resources prepared by earlier ones. The
// first error stops the generation and leaves Completed at the count of fully successful tasks.
func (c *Coordinator) run(ctx context.Context, generation uint64, ids []string, tasks []Task) {
	for index, task := range tasks {
		if err := task(ctx); err != nil {
			c.publish(generation, Snapshot{
				State:     "failed",
				Completed: index,
				Total:     len(tasks),
				Current:   ids[index],
				Err:       err,
			})

			return
		}

		next := ""
		if index+1 < len(ids) {
			next = ids[index+1]
		}

		c.publish(generation, Snapshot{
			State:     "running",
			Completed: index + 1,
			Total:     len(tasks),
			Current:   next,
		})
	}

	c.publish(generation, Snapshot{State: "complete", Completed: len(tasks), Total: len(tasks)})
}

// publish applies progress only when it belongs to the active generation. This check is the final defense against
// cancelled tasks racing with a replacement Begin call.
func (c *Coordinator) publish(generation uint64, snapshot Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.generation == generation {
		c.snapshot = snapshot
	}
}
