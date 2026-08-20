package raylibRenderer

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/cache"
)

// TestTextureCacheBudgetIsDeferredUntilOwnerThreadAppliesIt protects native eviction from cross-thread setters.
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

// TestRecordFrameWorkPublishesOnlyCurrentFrameDeltas keeps overlay counters distinct from lifetime diagnostics.
func TestRecordFrameWorkPublishesOnlyCurrentFrameDeltas(t *testing.T) {
	service := &Service{}
	service.drawCalls.Store(100)
	service.nodesVisited.Store(200)
	service.subtreesCulled.Store(30)
	service.textureUpdates.Store(40)
	start := service.BackendDiagnostics()

	service.drawCalls.Add(12)
	service.nodesVisited.Add(20)
	service.subtreesCulled.Add(3)
	service.textureUpdates.Add(2)
	service.recordFrameWork(start)

	got := service.BackendDiagnostics()
	if got.LastFrameDrawCalls != 12 || got.LastFrameNodesVisited != 20 ||
		got.LastFrameSubtreesCulled != 3 || got.LastFrameTextureUpdates != 2 {
		t.Fatalf("last-frame diagnostics = %+v", got)
	}
}
