package worldobjects

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/model"
	"github.com/gravestench/dark-magic/internal/game/data/recovered"
)

func TestResolverUsesActLocalOrderingForBothObjectKinds(t *testing.T) {
	resolver := New(recovered.Snapshot{MapObjects: []recovered.MapObject{{Act: 1, ID: 3, ObjectID: 108, Description: "Malus"}}}, gamedata.Snapshot{
		MonsterPresets: []models.MonsterPreset{{Act: 1, Place: "fallen"}, {Act: 2, Place: "skeleton"}, {Act: 1, Place: "zombie"}},
	})
	if id, description, found := resolver.ResolveStaticObject(1, 3); !found || id != 108 || description != "Malus" {
		t.Fatalf("static = %d, %q, %v", id, description, found)
	}
	if class, found := resolver.ResolveDynamicObject(1, 1); !found || class != "zombie" {
		t.Fatalf("dynamic = %q, %v", class, found)
	}
	if _, found := resolver.ResolveDynamicObject(2, 1); found {
		t.Fatal("act-local dynamic index leaked across acts")
	}
}
