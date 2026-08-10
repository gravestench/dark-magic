package cache

import (
	"testing"
)

func TestCacheInsert(t *testing.T) {
	cache := New(1)
	insertError := cache.Insert("A", "", 1)

	if insertError != nil {
		t.Fatalf("Cache insert resulted in unexpected error: %s", insertError)
	}
}

func TestCacheInsertWithinBudget(t *testing.T) {
	cache := New(1)
	insertError := cache.Insert("A", "", 2)

	if insertError != nil {
		t.Fatalf("Cache insert resulted in unexpected error: %s", insertError)
	}
	if cache.GetWeight() != 0 {
		t.Fatal("oversized entry should be evicted")
	}
}

func TestCacheInsertUpdatesWeight(t *testing.T) {
	cache := New(2)
	_ = cache.Insert("A", "", 1)
	_ = cache.Insert("B", "", 1)
	_ = cache.Insert("budget_exceeded", "", 1)

	if cache.GetWeight() != 2 {
		t.Fatal("Cache with budget 2 did not correctly set weight after evicting one of three nodes")
	}
}

func TestCacheInsertDuplicateRejected(t *testing.T) {
	cache := New(2)
	_ = cache.Insert("dupe", "", 1)
	dupeError := cache.Insert("dupe", "", 1)

	if dupeError == nil {
		t.Fatal("Cache insert of duplicate key did not result in any err")
	}
}

func TestCacheInsertEvictsLeastRecentlyUsed(t *testing.T) {
	cache := New(2)
	// with a budget of 2, inserting 3 keys should evict the last
	_ = cache.Insert("evicted", "", 1)
	_ = cache.Insert("A", "", 1)
	_ = cache.Insert("B", "", 1)

	_, foundEvicted := cache.Retrieve("evicted")
	if foundEvicted {
		t.Fatal("Cache insert did not trigger eviction after weight exceedance")
	}

	// double check that only 1 one was evicted and not any extra
	_, foundA := cache.Retrieve("A")
	_, foundB := cache.Retrieve("B")

	if !foundA || !foundB {
		t.Fatal("Cache insert evicted more than necessary")
	}
}

func TestCacheInsertEvictsLeastRecentlyRetrieved(t *testing.T) {
	cache := New(2)
	_ = cache.Insert("A", "", 1)
	_ = cache.Insert("evicted", "", 1)

	// retrieve the oldest node, promoting it head so it is not evicted
	cache.Retrieve("A")

	// insert once more, exceeding weight capacity
	_ = cache.Insert("B", "", 1)
	// now the least recently used key should be evicted
	_, foundEvicted := cache.Retrieve("evicted")
	if foundEvicted {
		t.Fatal("Cache insert did not evict least recently used after weight exceedance")
	}
}

func TestClear(t *testing.T) {
	cache := New(1)
	_ = cache.Insert("cleared", "", 1)
	cache.Clear()
	_, found := cache.Retrieve("cleared")

	if found {
		t.Fatal("Still able to retrieve nodes after cache was cleared")
	}
}

func TestVersionedCacheInvalidationAndDiagnostics(t *testing.T) {
	cache := New(3)
	var evicted []string
	cache.SetEvictionHandler(func(value interface{}) { evicted = append(evicted, value.(string)) })
	if err := cache.InsertVersioned("vfs", "hero", 1, "old", 2); err != nil {
		t.Fatal(err)
	}
	if value, ok := cache.RetrieveVersioned("vfs", "hero", 1); !ok || value != "old" {
		t.Fatalf("value = %v, ok = %v", value, ok)
	}
	cache.InvalidateNamespace("vfs", 2)
	if _, ok := cache.RetrieveVersioned("vfs", "hero", 2); ok {
		t.Fatal("stale generation remained")
	}
	if err := cache.InsertVersioned("vfs", "hero", 2, "new", 2); err != nil {
		t.Fatal(err)
	}
	if err := cache.SetBudget(1); err != nil {
		t.Fatal(err)
	}
	stats := cache.Diagnostics()
	if stats.Entries != 0 || stats.Weight != 0 || stats.Budget != 1 || stats.Hits != 1 || stats.Misses != 1 || stats.Evictions != 2 {
		t.Fatalf("stats = %#v", stats)
	}
	if len(evicted) != 2 || evicted[0] != "old" || evicted[1] != "new" {
		t.Fatalf("evicted = %v", evicted)
	}
}

func TestVersionMismatchEvictsStaleEntry(t *testing.T) {
	cache := New(10)
	if err := cache.InsertVersioned("assets", "font", 4, "font", 1); err != nil {
		t.Fatal(err)
	}
	if _, ok := cache.RetrieveVersioned("assets", "font", 5); ok {
		t.Fatal("version mismatch hit")
	}
	if cache.Diagnostics().Entries != 0 {
		t.Fatal("stale entry was not evicted")
	}
}

func TestCanInsertWithoutEvictionDoesNotTouchResidency(t *testing.T) {
	cache := New(10)
	if err := cache.Insert("active", "scene", 7); err != nil {
		t.Fatal(err)
	}
	if !cache.CanInsertWithoutEviction("active", 100) {
		t.Fatal("resident key was rejected")
	}
	if cache.CanInsertWithoutEviction("warm", 4) {
		t.Fatal("speculative insertion that requires eviction was admitted")
	}
	if !cache.CanInsertWithoutEviction("warm", 3) {
		t.Fatal("speculative insertion that fits was rejected")
	}
	if stats := cache.Diagnostics(); stats.Entries != 1 || stats.Weight != 7 || stats.Hits != 0 || stats.Misses != 0 {
		t.Fatalf("admission changed cache state: %#v", stats)
	}
}
