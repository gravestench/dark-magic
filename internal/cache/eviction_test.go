package cache

import (
	"reflect"
	"testing"
)

func TestCacheEvictionRemovalAndClearReleaseValues(t *testing.T) {
	t.Parallel()

	cache := New(2)
	var evicted []string
	cache.SetEvictionHandler(func(value interface{}) { evicted = append(evicted, value.(string)) })
	if err := cache.Insert("a", "A", 1); err != nil {
		t.Fatal(err)
	}
	if err := cache.Insert("b", "B", 1); err != nil {
		t.Fatal(err)
	}
	if err := cache.Insert("c", "C", 1); err != nil {
		t.Fatal(err)
	}
	if _, exists := cache.Retrieve("a"); exists {
		t.Fatal("oldest entry was not evicted")
	}
	if value, removed := cache.Remove("b"); !removed || value != "B" {
		t.Fatalf("Remove = %v/%v", value, removed)
	}
	cache.Clear()
	if !reflect.DeepEqual(evicted, []string{"A", "B", "C"}) {
		t.Fatalf("evicted = %v", evicted)
	}
	if cache.GetWeight() != 0 {
		t.Fatalf("weight = %d", cache.GetWeight())
	}
}
