package cache

import "errors"

// InvalidateNamespace evicts every entry in namespace that does not match the
// requested generation. Entries in other namespaces remain resident.
func (c *Cache) InvalidateNamespace(namespace string, generation uint64) {
	c.mutex.Lock()

	var values []interface{}

	for _, node := range c.lookup {
		if node.namespace != namespace || node.generation == generation {
			continue
		}

		values = append(values, c.removeLocked(node))
		c.evictions++
	}

	handler := c.onEvict
	c.mutex.Unlock()

	notifyEvictions(handler, values...)
}

// SetBudget changes the total weight budget and immediately evicts least
// recently used entries until the cache fits. A rejected negative budget leaves
// both the existing budget and residency untouched.
func (c *Cache) SetBudget(budget int) error {
	if budget < 0 {
		return errors.New("cache budget cannot be negative")
	}

	c.mutex.Lock()
	c.budget = budget
	evicted := c.evictOverBudgetLocked()
	handler := c.onEvict
	c.mutex.Unlock()

	notifyEvictions(handler, evicted...)

	return nil
}

// Clear removes every cache entry while retaining the budget, diagnostics
// counters, and eviction handler. Notifications occur only after the empty state
// is visible to concurrent callers.
func (c *Cache) Clear() {
	c.mutex.Lock()

	values := make([]interface{}, 0, len(c.lookup))
	for _, node := range c.lookup {
		values = append(values, node.value)
	}

	c.head = nil
	c.tail = nil
	c.lookup = make(map[string]*cacheNode)
	c.weight = 0
	c.evictions += uint64(len(values))
	handler := c.onEvict
	c.mutex.Unlock()

	notifyEvictions(handler, values...)
}

// evictOverBudgetLocked removes tail entries in least-recently-used order until
// the total weight fits the budget. Callers must hold the cache mutex and notify
// handlers only after unlocking.
func (c *Cache) evictOverBudgetLocked() []interface{} {
	var values []interface{}
	for c.tail != nil && c.weight > c.budget {
		values = append(values, c.removeLocked(c.tail))
		c.evictions++
	}

	return values
}

// notifyEvictions invokes the snapshotted handler in removal order. Running
// callbacks outside the cache mutex prevents resource cleanup from deadlocking
// when it re-enters cache methods.
func notifyEvictions(handler func(interface{}), values ...interface{}) {
	if handler == nil {
		return
	}

	for _, value := range values {
		handler(value)
	}
}
