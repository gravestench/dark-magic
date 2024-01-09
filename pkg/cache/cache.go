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
	head   *cacheNode
	tail   *cacheNode
	lookup map[string]*cacheNode
	weight int
	budget int
	mutex  sync.Mutex
}

// New creates an  instance of a Cache
func New(budget int) *Cache {
	return &Cache{lookup: make(map[string]*cacheNode), budget: budget}
}

// GetWeight gets the "weight" of a cache
func (c *Cache) GetWeight() int {
	return c.weight
}

// GetBudget gets the memory budget of a cache
func (c *Cache) GetBudget() int {
	return c.budget
}

// Insert inserts an object into the cache
func (c *Cache) Insert(key string, value interface{}, weight int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, found := c.lookup[key]; found {
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

	for ; c.tail != nil && c.tail != c.head && c.weight > c.budget; c.tail = c.tail.prev {
		c.weight -= c.tail.weight
		c.tail.prev.next = nil

		delete(c.lookup, c.tail.key)
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

// Delete removes an object from the cache by key
func (c *Cache) Delete(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	node, found := c.lookup[key]
	if !found {
		return false
	}

	if node == c.head {
		c.head = node.next
		if c.head != nil {
			c.head.prev = nil
		}
	} else {
		if node.prev != nil {
			node.prev.next = node.next
		}

		if node.next != nil {
			node.next.prev = node.prev
		}
	}

	delete(c.lookup, key)
	c.weight -= node.weight

	return true
}

// Clear removes all cache entries
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.head = nil
	c.tail = nil
	c.lookup = make(map[string]*cacheNode)
	c.weight = 0
}
