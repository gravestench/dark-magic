package cache

import (
	"errors"
	"sync"
)

type cacheNode struct {
	next       *cacheNode
	prev       *cacheNode
	key        string
	value      interface{}
	weight     int
	namespace  string
	generation uint64
}

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

// GetWeight reports the total caller-declared weight of resident entries.
func (c *Cache) GetWeight() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.weight
}

// GetBudget reports the current eviction budget.
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
	if _, exists := c.lookup[versionedKey("", key)]; exists {
		return true
	}
	return weight <= c.budget-c.weight
}

// SetEvictionHandler installs a callback invoked after values leave the cache.
func (c *Cache) SetEvictionHandler(handler func(interface{})) {
	c.mutex.Lock()
	c.onEvict = handler
	c.mutex.Unlock()
}

// Insert inserts an object into the cache
func (c *Cache) Insert(key string, value interface{}, weight int) error {
	return c.InsertVersioned("", key, 0, value, weight)
}

// InsertVersioned inserts an entry tied to one content generation namespace.
func (c *Cache) InsertVersioned(namespace, key string, generation uint64, value interface{}, weight int) error {
	if key == "" {
		return errors.New("cache key is required")
	}
	if weight < 0 {
		return errors.New("cache weight cannot be negative")
	}
	key = versionedKey(namespace, key)
	c.mutex.Lock()

	if _, found := c.lookup[key]; found {
		c.mutex.Unlock()
		return errors.New("key already exists in Cache")
	}

	node := &cacheNode{
		key:        key,
		value:      value,
		weight:     weight,
		namespace:  namespace,
		generation: generation,
		next:       c.head,
	}

	if c.head != nil {
		c.head.prev = node
	}

	c.head = node
	if c.tail == nil {
		c.tail = node
	}

	c.lookup[key] = node
	c.weight += node.weight

	var evicted []interface{}
	for c.tail != nil && c.weight > c.budget {
		node := c.tail
		evicted = append(evicted, c.removeLocked(node))
		c.evictions++
	}
	handler := c.onEvict
	c.mutex.Unlock()
	for _, item := range evicted {
		if handler != nil {
			handler(item)
		}
	}
	return nil
}

// Retrieve gets an object out of the cache
func (c *Cache) Retrieve(key string) (interface{}, bool) {
	return c.RetrieveVersioned("", key, 0)
}

// RetrieveVersioned returns only an entry from the requested generation.
func (c *Cache) RetrieveVersioned(namespace, key string, generation uint64) (interface{}, bool) {
	key = versionedKey(namespace, key)
	c.mutex.Lock()
	node, found := c.lookup[key]
	if !found || node.generation != generation {
		c.misses++
		var stale interface{}
		var handler func(interface{})
		if found {
			stale = c.removeLocked(node)
			c.evictions++
			handler = c.onEvict
		}
		c.mutex.Unlock()
		if handler != nil {
			handler(stale)
		}
		return nil, false
	}
	c.hits++

	if node != c.head {
		if node.next != nil {
			node.next.prev = node.prev
		}

		if node.prev != nil {
			node.prev.next = node.next
		}

		if node == c.tail {
			c.tail = c.tail.prev
		}

		node.next = c.head
		node.prev = nil

		if c.head != nil {
			c.head.prev = node
		}

		c.head = node
	}

	value := node.value
	c.mutex.Unlock()
	return value, true
}

// InvalidateNamespace evicts every entry not matching generation.
func (c *Cache) InvalidateNamespace(namespace string, generation uint64) {
	c.mutex.Lock()
	var values []interface{}
	for _, node := range c.lookup {
		if node.namespace == namespace && node.generation != generation {
			values = append(values, c.removeLocked(node))
			c.evictions++
		}
	}
	handler := c.onEvict
	c.mutex.Unlock()
	for _, value := range values {
		if handler != nil {
			handler(value)
		}
	}
}

// SetBudget changes the byte/weight budget and immediately evicts LRU entries.
func (c *Cache) SetBudget(budget int) error {
	if budget < 0 {
		return errors.New("cache budget cannot be negative")
	}
	c.mutex.Lock()
	c.budget = budget
	var values []interface{}
	for c.tail != nil && c.weight > c.budget {
		values = append(values, c.removeLocked(c.tail))
		c.evictions++
	}
	handler := c.onEvict
	c.mutex.Unlock()
	for _, value := range values {
		if handler != nil {
			handler(value)
		}
	}
	return nil
}

// Diagnostics returns counters and current capacity without exposing entries.
func (c *Cache) Diagnostics() Stats {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return Stats{Entries: len(c.lookup), Weight: c.weight, Budget: c.budget, Hits: c.hits, Misses: c.misses, Evictions: c.evictions}
}

// Clear removes all cache entries
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
	for _, value := range values {
		if handler != nil {
			handler(value)
		}
	}
}

// Remove deletes key and invokes the eviction handler for its value.
func (c *Cache) Remove(key string) (interface{}, bool) {
	return c.RemoveVersioned("", key)
}

// RemoveVersioned removes key from namespace regardless of generation.
func (c *Cache) RemoveVersioned(namespace, key string) (interface{}, bool) {
	key = versionedKey(namespace, key)
	c.mutex.Lock()
	node, found := c.lookup[key]
	if !found {
		c.mutex.Unlock()
		return nil, false
	}
	value := c.removeLocked(node)
	c.evictions++
	handler := c.onEvict
	c.mutex.Unlock()
	if handler != nil {
		handler(value)
	}
	return value, true
}

func (c *Cache) removeLocked(node *cacheNode) interface{} {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}
	delete(c.lookup, node.key)
	c.weight -= node.weight
	return node.value
}

func versionedKey(namespace, key string) string { return namespace + "\x00" + key }
