package item

import "testing"

func TestSellHeldSortsAndPacksVendorFootprints(t *testing.T) {
	state, err := NewState(Layout{VendorGrid: Grid{Width: 2, Height: 2}}, []Item{
		{ID: "z-stock", Code: "zzz", Width: 1, Height: 2},
		{ID: "a-sold", Code: "aaa", Width: 1, Height: 2},
	}, map[string]Placement{
		"z-stock": {Container: ContainerVendor, Slot: "weapons"},
		"a-sold":  {Container: ContainerHeld},
	})
	if err != nil {
		t.Fatal(err)
	}
	destination, err := state.sellHeld("a-sold", "weapons")
	if err != nil {
		t.Fatal(err)
	}
	if destination.Page != 0 || destination.X != 0 || destination.Y != 0 {
		t.Fatalf("sold placement = %#v", destination)
	}
	stock, _ := state.Placement("z-stock")
	if stock.Page != 0 || stock.X != 1 || stock.Y != 0 {
		t.Fatalf("repacked stock = %#v", stock)
	}
}

func TestVendorPackingCreatesPagesAndBuyRequiresEmptyHand(t *testing.T) {
	state, err := NewState(Layout{VendorGrid: Grid{Width: 1, Height: 1}}, []Item{
		{ID: "stock", Code: "aaa", Width: 1, Height: 1},
		{ID: "sold", Code: "bbb", Width: 1, Height: 1},
		{ID: "held", Code: "ccc", Width: 1, Height: 1},
	}, map[string]Placement{
		"stock": {Container: ContainerVendor, Slot: "misc"},
		"sold":  {Container: ContainerHeld},
	})
	if err != nil {
		t.Fatal(err)
	}
	destination, err := state.sellHeld("sold", "misc")
	if err != nil {
		t.Fatal(err)
	}
	if destination.Page != 1 {
		t.Fatalf("sold page = %d, want 1", destination.Page)
	}
	if err := state.Move("held", Placement{Container: ContainerHeld}); err != nil {
		t.Fatal(err)
	}
	if err := state.buyToHeld("stock"); err == nil {
		t.Fatal("purchase replaced an occupied hand")
	}
	if placement, _ := state.Placement("stock"); placement.Container != ContainerVendor {
		t.Fatalf("failed purchase moved stock: %#v", placement)
	}
	if err := state.Move("held", Placement{Container: ContainerWorld}); err != nil {
		t.Fatal(err)
	}
	if err := state.buyToHeld("stock"); err != nil {
		t.Fatal(err)
	}
	if placement, _ := state.Placement("sold"); placement.Page != 0 {
		t.Fatalf("remaining stock was not compacted: %#v", placement)
	}
}

func TestOversizedVendorSaleIsAtomic(t *testing.T) {
	state, err := NewState(Layout{VendorGrid: Grid{Width: 1, Height: 1}}, []Item{{ID: "large", Code: "box", Width: 2, Height: 1}}, map[string]Placement{"large": {Container: ContainerHeld}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := state.sellHeld("large", "misc"); err == nil {
		t.Fatal("oversized sale was accepted")
	}
	if placement, _ := state.Placement("large"); placement.Container != ContainerHeld {
		t.Fatalf("failed sale moved item: %#v", placement)
	}
}
