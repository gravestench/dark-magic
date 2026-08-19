package cache

import (
	"reflect"
	"testing"
)

// evictionRecorder captures string-valued callbacks in delivery order so tests
// can verify ownership release without coupling to cache internals.
type evictionRecorder struct {
	values []string
}

// record implements the eviction callback expected by Cache. The type assertion
// deliberately fails a test process if a fixture releases an unexpected value type.
func (r *evictionRecorder) record(value interface{}) {
	r.values = append(r.values, value.(string))
}

// requireValues verifies the complete callback sequence, preventing a test from
// overlooking duplicate, missing, or reordered ownership releases.
func (r *evictionRecorder) requireValues(t *testing.T, want []string) {
	t.Helper()

	if !reflect.DeepEqual(r.values, want) {
		t.Fatalf("evicted values = %v, want %v", r.values, want)
	}
}

// TestCacheEvictionRemovalAndClearReleaseValues verifies that automatic eviction,
// explicit removal, and clearing each release exactly their owned value.
func TestCacheEvictionRemovalAndClearReleaseValues(t *testing.T) {
	t.Parallel()

	cache := New(2)
	recorder := &evictionRecorder{}
	cache.SetEvictionHandler(recorder.record)

	// Inserting c exceeds the budget and releases the least-recently-used value A.
	insertCacheEntry(t, cache, "a", "A", 1)
	insertCacheEntry(t, cache, "b", "B", 1)
	insertCacheEntry(t, cache, "c", "C", 1)
	requireCacheMiss(t, cache, "a")

	// Explicit removal returns and releases B before Clear releases the sole remainder.
	if value, removed := cache.Remove("b"); !removed || value != "B" {
		t.Fatalf("Remove(%q) = %v/%v, want B/true", "b", value, removed)
	}

	cache.Clear()

	recorder.requireValues(t, []string{"A", "B", "C"})

	if got := cache.GetWeight(); got != 0 {
		t.Fatalf("weight = %d, want 0 after Clear", got)
	}
}

// TestSetBudgetRejectsNegativeValue verifies that a failed budget update keeps
// both the existing limit and resident entries intact while preserving error text.
func TestSetBudgetRejectsNegativeValue(t *testing.T) {
	cache := New(2)
	insertCacheEntry(t, cache, "resident", "value", 1)

	err := cache.SetBudget(-1)
	if err == nil || err.Error() != "cache budget cannot be negative" {
		t.Fatalf("SetBudget(-1) error = %v, want %q", err, "cache budget cannot be negative")
	}

	wantStats := Stats{Entries: 1, Weight: 1, Budget: 2}
	if stats := cache.Diagnostics(); stats != wantStats {
		t.Fatalf("diagnostics after rejected budget = %#v, want %#v", stats, wantStats)
	}

	requireCachedValue(t, cache, "resident", "value")
}
