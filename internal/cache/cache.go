package cache

import "sync"

// Stats is an immutable cache diagnostics snapshot.
type Stats struct {
	Entries, Weight, Budget int
	Hits, Misses, Evictions uint64
}

// Cache is a mutex-protected weighted LRU shared by resource-owning adapters.
//
// Weight is deliberately caller-defined: decoded pixels, native textures, and
// other resources do not have one universal notion of cost. Namespace and
// generation metadata let a reload invalidate one logical asset family without
// flushing unrelated entries or returning data from an older content generation.
type Cache struct {
	head      *cacheNode
	tail      *cacheNode
	lookup    map[string]*cacheNode
	weight    int
	budget    int
	mutex     sync.Mutex
	onEvict   func(interface{})
	hits      uint64
	misses    uint64
	evictions uint64
}

// New creates an empty cache with the supplied total weight budget. A budget of
// zero keeps bookkeeping active while making every inserted entry immediately
// eligible for eviction.
func New(budget int) *Cache {
	return &Cache{lookup: make(map[string]*cacheNode), budget: budget}
}

// GetWeight reports the total caller-declared weight of resident entries. The
// snapshot is synchronized so callers never observe an eviction halfway through.
func (c *Cache) GetWeight() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.weight
}

// GetBudget reports the current eviction budget. The synchronized snapshot
// prevents callers from racing with an in-progress budget update.
func (c *Cache) GetBudget() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.budget
}

// CanInsertWithoutEviction reports whether key is already resident or adding
// weight fits in currently unused capacity. Optional background work uses this
// to avoid evicting resources that foreground work may still need.
func (c *Cache) CanInsertWithoutEviction(key string, weight int) bool {
	if key == "" || weight < 0 {
		return false
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// A resident entry requires no capacity change, regardless of the proposed weight.
	if _, exists := c.lookup[versionedKey("", key)]; exists {
		return true
	}

	return weight <= c.budget-c.weight
}

// SetEvictionHandler installs the callback used by later removals. Each removal
// snapshots its handler under the mutex, so replacement cannot change an
// already-pending notification.
func (c *Cache) SetEvictionHandler(handler func(interface{})) {
	c.mutex.Lock()
	c.onEvict = handler
	c.mutex.Unlock()
}

// Diagnostics returns counters and current capacity without exposing entries.
// Capturing every field under one lock keeps the snapshot internally consistent.
func (c *Cache) Diagnostics() Stats {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return Stats{
		Entries:   len(c.lookup),
		Weight:    c.weight,
		Budget:    c.budget,
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
	}
}
