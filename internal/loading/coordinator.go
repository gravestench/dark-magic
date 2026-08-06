// Package loading coordinates engine-owned work required by scene transitions.
package loading

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Task func(context.Context) error

type Snapshot struct {
	State     string
	Completed int
	Total     int
	Current   string
	Err       error
}

func (s Snapshot) Progress() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Completed) / float64(s.Total)
}

type Coordinator struct {
	mu         sync.RWMutex
	tasks      map[string]Task
	generation uint64
	cancel     context.CancelFunc
	snapshot   Snapshot
}

func New(tasks map[string]Task) *Coordinator {
	copyTasks := make(map[string]Task, len(tasks))
	for id, task := range tasks {
		copyTasks[id] = task
	}
	return &Coordinator{tasks: copyTasks, snapshot: Snapshot{State: "idle"}}
}

func (c *Coordinator) Begin(parent context.Context, ids []string) error {
	if len(ids) == 0 {
		return errors.New("loading: at least one dependency is required")
	}
	selected := make([]Task, len(ids))
	for index, id := range ids {
		task, exists := c.tasks[id]
		if !exists || task == nil {
			return fmt.Errorf("loading: dependency %q is not registered", id)
		}
		selected[index] = task
	}

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

func (c *Coordinator) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshot
}

func (c *Coordinator) Close() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.generation++
	c.mu.Unlock()
}

func (c *Coordinator) run(ctx context.Context, generation uint64, ids []string, tasks []Task) {
	for index, task := range tasks {
		if err := task(ctx); err != nil {
			c.publish(generation, Snapshot{State: "failed", Completed: index, Total: len(tasks), Current: ids[index], Err: err})
			return
		}
		next := ""
		if index+1 < len(ids) {
			next = ids[index+1]
		}
		c.publish(generation, Snapshot{State: "running", Completed: index + 1, Total: len(tasks), Current: next})
	}
	c.publish(generation, Snapshot{State: "complete", Completed: len(tasks), Total: len(tasks)})
}

func (c *Coordinator) publish(generation uint64, snapshot Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.generation == generation {
		c.snapshot = snapshot
	}
}
