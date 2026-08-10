package item

import "testing"

func TestCommerceUsesNPCMultipliersCapsAndCarriedGold(t *testing.T) {
	state, err := NewState(Layout{VendorGrid: Grid{Width: 2, Height: 2}, Gold: GoldBalance{Carried: 1000, Stashed: 5000}}, []Item{
		{ID: "held", Code: "ssd", Width: 1, Height: 2, BaseCost: 400},
		{ID: "stock", Code: "cap", Width: 1, Height: 1, BaseCost: 300},
	}, map[string]Placement{"held": {Container: ContainerHeld}, "stock": {Container: ContainerVendor, Slot: "armor"}})
	if err != nil {
		t.Fatal(err)
	}
	terms := TradeTerms{BuyMultiplier: 1024, SellMultiplier: 2048, MaxBuy: 250}
	price, err := state.sellHeldForGold("held", "weapons", terms)
	if err != nil || price != 250 {
		t.Fatalf("sale = %d, %v", price, err)
	}
	layout, _, _ := state.Snapshot()
	if layout.Gold.Carried != 1250 || layout.Gold.Stashed != 5000 {
		t.Fatalf("gold after sale = %#v", layout.Gold)
	}
	price, err = state.buyToHeldForGold("stock", terms)
	if err != nil || price != 600 {
		t.Fatalf("purchase = %d, %v", price, err)
	}
	layout, _, _ = state.Snapshot()
	if layout.Gold.Carried != 650 || layout.Gold.Stashed != 5000 {
		t.Fatalf("gold after purchase = %#v", layout.Gold)
	}
}

func TestFailedCommerceChangesNeitherGoldNorItems(t *testing.T) {
	state, err := NewState(Layout{VendorGrid: Grid{Width: 1, Height: 1}, Gold: GoldBalance{Carried: 10}}, []Item{{ID: "stock", Code: "cap", Width: 1, Height: 1, BaseCost: 100}}, map[string]Placement{"stock": {Container: ContainerVendor, Slot: "armor"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := state.buyToHeldForGold("stock", TradeTerms{SellMultiplier: 1024}); err == nil {
		t.Fatal("unaffordable purchase was accepted")
	}
	layout, _, placements := state.Snapshot()
	if layout.Gold.Carried != 10 || placements["stock"].Container != ContainerVendor {
		t.Fatalf("failed purchase mutated state: %#v %#v", layout.Gold, placements)
	}
}

func TestTradeCatalogNormalizesVendorNames(t *testing.T) {
	catalog := TradeCatalog{"akara": {BuyMultiplier: 512, SellMultiplier: 1024}}
	if _, err := catalog.Terms("  AKARA "); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Terms("missing"); err == nil {
		t.Fatal("unknown vendor was accepted")
	}
}
