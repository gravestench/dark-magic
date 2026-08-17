package world

import "testing"

func TestSelectorUsesPriorityDistanceAndStableID(t *testing.T) {
	selector, err := NewSelector([]Selectable{
		{ID: "far", Kind: "npc", X: 2, Y: 0, Radius: 3},
		{ID: "near", Kind: "npc", X: 0, Y: 0, Radius: 3},
		{ID: "priority", Kind: "item", X: 2, Y: 0, Radius: 3, Priority: 1},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	got, found := selector.Hit(0, 0)
	if !found || got.ID != "priority" {
		t.Fatalf("priority hit = %#v, %v", got, found)
	}
	selector, _ = NewSelector([]Selectable{{ID: "b", Kind: "npc", Radius: 2}, {ID: "a", Kind: "npc", Radius: 2}}, 2)
	got, _ = selector.Hit(0, 0)
	if got.ID != "a" {
		t.Fatalf("stable overlap hit = %q", got.ID)
	}
}

func TestMapLineClearUsesLOSCollision(t *testing.T) {
	world := &Map{WidthSubtiles: 8, HeightSubtiles: 8, flags: make([]Flags, 64)}
	if !world.LineClear(1, 1, 6, 6) {
		t.Fatal("open line reported blocked")
	}
	world.flags[3*8+3] = Flags{BlockLOS: true}
	if world.LineClear(1, 1, 6, 6) {
		t.Fatal("LOS blocker was ignored")
	}
	if !world.LineClear(3, 3, 6, 6) {
		t.Fatal("source endpoint should be allowed")
	}
}

func TestMapBarrierClearUsesJumpCollisionIndependentlyFromLOS(t *testing.T) {
	world := &Map{WidthSubtiles: 8, HeightSubtiles: 8, flags: make([]Flags, 64)}
	world.flags[3*8+3] = Flags{BlockLOS: true}
	if !world.BarrierClear(1, 1, 6, 6) {
		t.Fatal("visual LOS blocker should not become a melee barrier")
	}
	world.flags[3*8+3] = Flags{BlockJump: true}
	if world.BarrierClear(1, 1, 6, 6) {
		t.Fatal("flying/melee barrier was ignored")
	}
	if !world.LineClear(1, 1, 6, 6) {
		t.Fatal("melee barrier should not become a visual LOS blocker")
	}
}

func TestMapObjectsBecomeStableSelectables(t *testing.T) {
	world := &Map{Objects: []Object{{Type: ObjectTypeStatic, ID: 7, X: 10, Y: 20, Resolved: true}, {Type: ObjectTypeDynamic, ID: 2, X: 4, Y: 5, Resolved: true}}}
	got := world.Selectables()
	if len(got) != 2 || got[0].ID != "ds1-object:2:7:0" || got[1].Kind != "dynamic-object" {
		t.Fatalf("selectables = %#v", got)
	}
}
