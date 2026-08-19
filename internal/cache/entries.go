package cache

import "errors"

// cacheNode is one resident entry in the doubly linked recency list. The head
// is most recently used and the tail is the next entry eligible for eviction.
type cacheNode struct {
	next       *cacheNode
	prev       *cacheNode
	key        string
	value      interface{}
	weight     int
	namespace  string
	generation uint64
}

// Insert adds an unversioned entry. It delegates to the versioned path so both
// APIs preserve the same validation, recency, and eviction behavior.
func (c *Cache) Insert(key string, value interface{}, weight int) error {
	return c.InsertVersioned("", key, 0, value, weight)
}

// InsertVersioned adds an entry tied to one content generation namespace.
// Successful insertion may immediately evict this or older entries when their
// combined caller-declared weight exceeds the current budget.
func (c *Cache) InsertVersioned(namespace, key string, generation uint64, value interface{}, weight int) error {
	if key == "" {
		return errors.New("cache key is required")
	}

	if weight < 0 {
		return errors.New("cache weight cannot be negative")
	}

	storageKey := versionedKey(namespace, key)

	c.mutex.Lock()

	if _, found := c.lookup[storageKey]; found {
		c.mutex.Unlock()
		return errors.New("key already exists in Cache")
	}

	node := &cacheNode{
		key:        storageKey,
		value:      value,
		weight:     weight,
		namespace:  namespace,
		generation: generation,
	}
	c.addToFrontLocked(node)
	c.lookup[storageKey] = node
	c.weight += node.weight

	evicted := c.evictOverBudgetLocked()
	handler := c.onEvict
	c.mutex.Unlock()

	notifyEvictions(handler, evicted...)

	return nil
}

// Retrieve looks up an unversioned entry. A hit promotes the entry through the
// same recency path used by version-aware callers.
func (c *Cache) Retrieve(key string) (interface{}, bool) {
	return c.RetrieveVersioned("", key, 0)
}

// RetrieveVersioned returns only an entry from the requested generation. A
// generation mismatch is both a miss and an eviction, preventing stale content
// from surviving after a caller advances its generation.
func (c *Cache) RetrieveVersioned(namespace, key string, generation uint64) (interface{}, bool) {
	storageKey := versionedKey(namespace, key)

	c.mutex.Lock()

	node, found := c.lookup[storageKey]
	if !found {
		c.misses++
		c.mutex.Unlock()

		return nil, false
	}

	if node.generation != generation {
		c.misses++
		stale := c.removeLocked(node)
		c.evictions++
		handler := c.onEvict
		c.mutex.Unlock()

		notifyEvictions(handler, stale)

		return nil, false
	}

	c.hits++
	c.moveToFrontLocked(node)
	value := node.value
	c.mutex.Unlock()

	return value, true
}

// Remove deletes an unversioned key and notifies the handler after releasing
// the cache mutex, allowing resource cleanup to call back into the cache safely.
func (c *Cache) Remove(key string) (interface{}, bool) {
	return c.RemoveVersioned("", key)
}

// RemoveVersioned deletes a key from one namespace regardless of generation.
// Its return value and eviction callback both receive the exact stored value.
func (c *Cache) RemoveVersioned(namespace, key string) (interface{}, bool) {
	storageKey := versionedKey(namespace, key)

	c.mutex.Lock()

	node, found := c.lookup[storageKey]
	if !found {
		c.mutex.Unlock()

		return nil, false
	}

	value := c.removeLocked(node)
	c.evictions++
	handler := c.onEvict
	c.mutex.Unlock()

	notifyEvictions(handler, value)

	return value, true
}

// addToFrontLocked links node as the most recently used entry. Callers must
// hold the cache mutex so head, tail, and neighboring links change atomically.
func (c *Cache) addToFrontLocked(node *cacheNode) {
	node.prev = nil
	node.next = c.head

	if c.head != nil {
		c.head.prev = node
	}

	c.head = node
	if c.tail == nil {
		c.tail = node
	}
}

// moveToFrontLocked promotes a cache hit without changing lookup membership or
// total weight. Callers must hold the cache mutex while recency links move.
func (c *Cache) moveToFrontLocked(node *cacheNode) {
	if node == c.head {
		return
	}

	c.detachLocked(node)
	c.addToFrontLocked(node)
}

// detachLocked unlinks node from the recency list while leaving membership and
// weight untouched. This separation lets promotion and removal share link logic.
func (c *Cache) detachLocked(node *cacheNode) {
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
}

// removeLocked removes node from every cache index and returns the owned value
// for notification after the mutex is released. Callers must hold the mutex.
func (c *Cache) removeLocked(node *cacheNode) interface{} {
	c.detachLocked(node)
	delete(c.lookup, node.key)
	c.weight -= node.weight

	return node.value
}

// versionedKey preserves the package's NUL-delimited composite-key format so
// every entry operation continues to use the established lookup representation.
func versionedKey(namespace, key string) string {
	return namespace + "\x00" + key
}
