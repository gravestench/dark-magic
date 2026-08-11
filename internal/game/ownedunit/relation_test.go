package ownedunit

import (
	"testing"

	"github.com/gravestench/akara"
	models "github.com/gravestench/dark-magic/internal/game/data/model"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestRelationEnforcesStableLimitAndSurvivesCheckpoint(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	owner := engine.World().MustCreateEntity()
	first := engine.World().MustCreateEntity()
	second := engine.World().MustCreateEntity()
	third := engine.World().MustCreateEntity()
	category := Category{ID: "skeleton", BaseMax: 2, Replacement: ReplaceOldest}
	attach := func(unit akara.Entity, tick uint64) Decision {
		decision, err := Attach(engine.World(), Relation{Unit: unit, Owner: owner, OwnerID: "player:hero", UltimateOwnerID: "player:hero", CreatedTick: tick}, category)
		if err != nil {
			t.Fatal(err)
		}
		return decision
	}
	attach(first, 10)
	attach(second, 11)
	decision := attach(third, 12)
	if len(decision.Inactivated) != 1 || decision.Inactivated[0] != first {
		t.Fatalf("replacement = %#v", decision)
	}
	active := ActiveFor(engine.World(), owner, "skeleton")
	if len(active) != 2 || active[0] != second || active[1] != third {
		t.Fatalf("active = %v", active)
	}
	attribution, found := ResolveAttribution(engine.World(), third, "monster:skeleton-3")
	if !found || attribution.ImmediateOwnerID != "player:hero" || attribution.UltimateOwnerID != "player:hero" {
		t.Fatalf("attribution = %#v", attribution)
	}
	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := gameecs.RestoreSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	got, err := restored.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	wantChecksum, _ := snapshot.Checksum()
	gotChecksum, _ := got.Checksum()
	if gotChecksum != wantChecksum {
		t.Fatalf("restored checksum = %s, want %s", gotChecksum, wantChecksum)
	}
}

func TestCategoryFromPetTypePreservesAuthoredPolicy(t *testing.T) {
	category, err := CategoryFromPetType(models.PetType{PetType: " raven ", Group: 3, BaseMax: 5, Warp: true, Range: true, Unsummon: true}, ReplaceNewest, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if category.ID != "raven" || category.Group != 3 || category.BaseMax != 5 || !category.WarpWithOwner || !category.RangeLimited || !category.Unsummon || !category.SurvivesOwnerDeath {
		t.Fatalf("category = %#v", category)
	}
}

func TestRelationRejectsLimitAndEnforcesSharedGroup(t *testing.T) {
	world := akara.NewWorld()
	defer world.Close()
	owner := world.MustCreateEntity()
	wolf := world.MustCreateEntity()
	bear := world.MustCreateEntity()
	wolfCategory := Category{ID: "wolf", Group: 7, BaseMax: 1, Replacement: ReplaceOldest}
	if _, err := Attach(world, Relation{Unit: wolf, Owner: owner, OwnerID: "hero", UltimateOwnerID: "hero", CreatedTick: 1}, wolfCategory); err != nil {
		t.Fatal(err)
	}
	bearCategory := Category{ID: "bear", Group: 7, BaseMax: 1, Replacement: ReplaceOldest}
	decision, err := Attach(world, Relation{Unit: bear, Owner: owner, OwnerID: "hero", UltimateOwnerID: "hero", CreatedTick: 2}, bearCategory)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.Inactivated) != 1 || decision.Inactivated[0] != wolf {
		t.Fatalf("group replacement = %#v", decision)
	}
	rejected := world.MustCreateEntity()
	if _, err := Attach(world, Relation{Unit: rejected, Owner: owner, OwnerID: "hero", UltimateOwnerID: "hero", CreatedTick: 3}, Category{ID: "bear", Group: 7, BaseMax: 1, Replacement: Reject}); err == nil {
		t.Fatal("reject policy accepted excess unit")
	}
}
