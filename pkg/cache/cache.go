package cache

import (
	"errors"
	"sync"
)

type cacheNode struct {
	next   *cacheNode
	prev   *cacheNode
	key    string
	value  interface{}
	weight int
}

// Cache stores arbitrary data for fast retrieval
type Cache struct {
	head    *cacheNode
	tail    *cacheNode
	lookup  map[string]*cacheNode
	weight  int
	budget  int
	mutex   sync.Mutex
	onEvict func(interface{})
}

// New creates an  instance of a Cache
func New(budget int) *Cache {
	return &Cache{lookup: make(map[string]*cacheNode), budget: budget}
}

// GetWeight gets the "weight" of a cache
func (c *Cache) GetWeight() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.weight
}

// GetBudget gets the memory budget of a cache
func (c *Cache) GetBudget() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.budget
}

// SetEvictionHandler installs a callback invoked after values leave the cache.
func (c *Cache) SetEvictionHandler(handler func(interface{})) {
	c.mutex.Lock()
	c.onEvict = handler
	c.mutex.Unlock()
}

// Insert inserts an object into the cache
func (c *Cache) Insert(key string, value interface{}, weight int) error {
	c.mutex.Lock()

	if _, found := c.lookup[key]; found {
		c.mutex.Unlock()
		return errors.New("key already exists in Cache")
	}

	node := &cacheNode{
		key:    key,
		value:  value,
		weight: weight,
		next:   c.head,
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
	for ; c.tail != nil && c.tail != c.head && c.weight > c.budget; c.tail = c.tail.prev {
		c.weight -= c.tail.weight
		c.tail.prev.next = nil
		evicted = append(evicted, c.tail.value)
		delete(c.lookup, c.tail.key)
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
	c.mutex.Lock()
	defer c.mutex.Unlock()

	node, found := c.lookup[key]
	if !found {
		return nil, false
	}

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

	return node.value, true
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
	c.mutex.Lock()
	node, found := c.lookup[key]
	if !found {
		c.mutex.Unlock()
		return nil, false
	}
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
	delete(c.lookup, key)
	c.weight -= node.weight
	handler := c.onEvict
	c.mutex.Unlock()
	if handler != nil {
		handler(node.value)
	}
	return node.value, true
}
