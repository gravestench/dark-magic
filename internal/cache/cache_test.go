package cache

import "testing"

// TestCacheInsert verifies that a valid entry becomes resident with its declared
// value and weight, establishing the basic unversioned cache contract.
func TestCacheInsert(t *testing.T) {
	cache := New(1)

	insertCacheEntry(t, cache, "A", "value", 1)
	requireCachedValue(t, cache, "A", "value")
}

// TestCacheInsertWithinBudget verifies that an oversized entry is accepted and
// immediately evicted rather than leaving the cache over budget.
func TestCacheInsertWithinBudget(t *testing.T) {
	cache := New(1)

	insertCacheEntry(t, cache, "A", "value", 2)

	if got := cache.GetWeight(); got != 0 {
		t.Fatalf("weight = %d, want 0 after oversized entry eviction", got)
	}
}

// TestCacheInsertUpdatesWeight verifies that automatic eviction subtracts only
// the least-recently-used entry's weight from the resident total.
func TestCacheInsertUpdatesWeight(t *testing.T) {
	cache := New(2)

	insertCacheEntry(t, cache, "A", "A", 1)
	insertCacheEntry(t, cache, "B", "B", 1)
	insertCacheEntry(t, cache, "budget_exceeded", "C", 1)

	if got := cache.GetWeight(); got != 2 {
		t.Fatalf("weight = %d, want 2 after evicting one of three entries", got)
	}
}

// TestCacheInsertDuplicateRejected verifies that a second composite key cannot
// replace the resident value or change the package's established error text.
func TestCacheInsertDuplicateRejected(t *testing.T) {
	cache := New(2)
	insertCacheEntry(t, cache, "dupe", "original", 1)

	err := cache.Insert("dupe", "replacement", 1)
	if err == nil || err.Error() != "key already exists in Cache" {
		t.Fatalf("duplicate insert error = %v, want %q", err, "key already exists in Cache")
	}

	requireCachedValue(t, cache, "dupe", "original")
}

// TestCacheInsertRejectsInvalidInput verifies that validation failures preserve
// their public error strings and leave residency unchanged.
func TestCacheInsertRejectsInvalidInput(t *testing.T) {
	cache := New(2)

	err := cache.Insert("", "missing-key", 1)
	if err == nil || err.Error() != "cache key is required" {
		t.Fatalf("empty-key insert error = %v, want %q", err, "cache key is required")
	}

	err = cache.Insert("negative-weight", "value", -1)
	if err == nil || err.Error() != "cache weight cannot be negative" {
		t.Fatalf("negative-weight insert error = %v, want %q", err, "cache weight cannot be negative")
	}

	if stats := cache.Diagnostics(); stats != (Stats{Budget: 2}) {
		t.Fatalf("diagnostics after rejected inserts = %#v, want empty cache", stats)
	}
}

// TestCacheInsertEvictsLeastRecentlyUsed verifies that capacity pressure removes
// exactly the oldest entry and leaves newer entries resident.
func TestCacheInsertEvictsLeastRecentlyUsed(t *testing.T) {
	cache := New(2)

	insertCacheEntry(t, cache, "evicted", "oldest", 1)
	insertCacheEntry(t, cache, "A", "A", 1)
	insertCacheEntry(t, cache, "B", "B", 1)

	requireCacheMiss(t, cache, "evicted")
	requireCachedValue(t, cache, "A", "A")
	requireCachedValue(t, cache, "B", "B")
}

// TestCacheInsertEvictsLeastRecentlyRetrieved verifies that a successful lookup
// promotes its entry, shifting the next eviction to the previously newer key.
func TestCacheInsertEvictsLeastRecentlyRetrieved(t *testing.T) {
	cache := New(2)
	insertCacheEntry(t, cache, "A", "A", 1)
	insertCacheEntry(t, cache, "evicted", "oldest-after-promotion", 1)

	// The lookup makes A most recently used before the insertion exceeds capacity.
	requireCachedValue(t, cache, "A", "A")
	insertCacheEntry(t, cache, "B", "B", 1)

	requireCacheMiss(t, cache, "evicted")
	requireCachedValue(t, cache, "A", "A")
	requireCachedValue(t, cache, "B", "B")
}

// TestClear verifies that clearing makes every resident entry unavailable while
// retaining a usable cache object for subsequent lookups.
func TestClear(t *testing.T) {
	cache := New(1)
	insertCacheEntry(t, cache, "cleared", "value", 1)

	cache.Clear()

	requireCacheMiss(t, cache, "cleared")
}

// TestVersionedCacheInvalidationAndDiagnostics exercises the generation
// lifecycle as one narrative and verifies its observable counters and callbacks.
func TestVersionedCacheInvalidationAndDiagnostics(t *testing.T) {
	cache := New(3)
	recorder := &evictionRecorder{}
	cache.SetEvictionHandler(recorder.record)

	// Establish a hit in the old generation before advancing the namespace.
	insertVersionedCacheEntry(t, cache, "vfs", "hero", 1, "old", 2)

	if value, found := cache.RetrieveVersioned("vfs", "hero", 1); !found || value != "old" {
		t.Fatalf("old generation lookup = %v/%v, want old/true", value, found)
	}

	// Invalidation removes generation one, so generation two cannot see stale data.
	cache.InvalidateNamespace("vfs", 2)

	if value, found := cache.RetrieveVersioned("vfs", "hero", 2); found || value != nil {
		t.Fatalf("invalidated generation lookup = %v/%v, want nil/false", value, found)
	}

	// The replacement is valid but cannot remain after the budget is reduced below its weight.
	insertVersionedCacheEntry(t, cache, "vfs", "hero", 2, "new", 2)

	if err := cache.SetBudget(1); err != nil {
		t.Fatalf("SetBudget(1) error = %v", err)
	}

	wantStats := Stats{Budget: 1, Hits: 1, Misses: 1, Evictions: 2}
	if stats := cache.Diagnostics(); stats != wantStats {
		t.Fatalf("diagnostics = %#v, want %#v", stats, wantStats)
	}

	recorder.requireValues(t, []string{"old", "new"})
}

// TestVersionMismatchEvictsStaleEntry verifies that a mismatched generation is
// counted as a miss and cannot leave stale residency behind.
func TestVersionMismatchEvictsStaleEntry(t *testing.T) {
	cache := New(10)
	insertVersionedCacheEntry(t, cache, "assets", "font", 4, "font", 1)

	if value, found := cache.RetrieveVersioned("assets", "font", 5); found || value != nil {
		t.Fatalf("mismatched generation lookup = %v/%v, want nil/false", value, found)
	}

	wantStats := Stats{Budget: 10, Misses: 1, Evictions: 1}
	if stats := cache.Diagnostics(); stats != wantStats {
		t.Fatalf("diagnostics = %#v, want %#v", stats, wantStats)
	}
}

// TestCanInsertWithoutEvictionDoesNotTouchResidency verifies that admission
// checks neither promote entries nor change diagnostics while assessing capacity.
func TestCanInsertWithoutEvictionDoesNotTouchResidency(t *testing.T) {
	cache := New(10)
	insertCacheEntry(t, cache, "active", "scene", 7)

	if !cache.CanInsertWithoutEviction("active", 100) {
		t.Fatal("resident key was rejected")
	}

	if cache.CanInsertWithoutEviction("warm", 4) {
		t.Fatal("speculative insertion that requires eviction was admitted")
	}

	if !cache.CanInsertWithoutEviction("warm", 3) {
		t.Fatal("speculative insertion that fits was rejected")
	}

	wantStats := Stats{Entries: 1, Weight: 7, Budget: 10}
	if stats := cache.Diagnostics(); stats != wantStats {
		t.Fatalf("admission changed cache state: got %#v, want %#v", stats, wantStats)
	}
}

// insertCacheEntry inserts a fixture through the public API and fails at the
// contract boundary, so later assertions never run against incomplete setup.
func insertCacheEntry(t *testing.T, cache *Cache, key string, value interface{}, weight int) {
	t.Helper()

	if err := cache.Insert(key, value, weight); err != nil {
		t.Fatalf("Insert(%q) error = %v", key, err)
	}
}

// insertVersionedCacheEntry establishes a namespaced fixture and reports all
// identifying fields when setup fails, keeping generation failures diagnostic.
func insertVersionedCacheEntry(
	t *testing.T,
	cache *Cache,
	namespace string,
	key string,
	generation uint64,
	value interface{},
	weight int,
) {
	t.Helper()

	if err := cache.InsertVersioned(namespace, key, generation, value, weight); err != nil {
		t.Fatalf("InsertVersioned(%q, %q, %d) error = %v", namespace, key, generation, err)
	}
}

// requireCachedValue performs a public lookup and asserts its value. Its use is
// intentional when the test also needs Retrieve's hit and promotion side effects.
func requireCachedValue(t *testing.T, cache *Cache, key string, want interface{}) {
	t.Helper()

	value, found := cache.Retrieve(key)
	if !found || value != want {
		t.Fatalf("Retrieve(%q) = %v/%v, want %v/true", key, value, found, want)
	}
}

// requireCacheMiss performs a public lookup and verifies that no value leaks on
// a miss, preserving both halves of Retrieve's result contract.
func requireCacheMiss(t *testing.T, cache *Cache, key string) {
	t.Helper()

	value, found := cache.Retrieve(key)
	if found || value != nil {
		t.Fatalf("Retrieve(%q) = %v/%v, want nil/false", key, value, found)
	}
}
