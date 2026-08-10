package raylibRenderer

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/cache"
)

func TestTextureCacheBudgetIsDeferredUntilOwnerThreadAppliesIt(t *testing.T) {
	service := &Service{cache: cache.New(100)}
	service.SetTextureCacheBudget(512)
	if got := service.cache.GetBudget(); got != 100 {
		t.Fatalf("budget changed from caller thread: %d", got)
	}
	if err := service.applyTextureCacheBudget(); err != nil {
		t.Fatal(err)
	}
	if got := service.cache.GetBudget(); got != 512 {
		t.Fatalf("applied budget = %d", got)
	}
}
