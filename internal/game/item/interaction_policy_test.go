package item

import "testing"

type interactionPolicyStub struct {
	vendor  string
	service string
}

func (policy interactionPolicyStub) CanTrade(_ string, vendor string) bool {
	return vendor == policy.vendor
}
func (policy interactionPolicyStub) CanService(_ string, service string) bool {
	return service == policy.service
}

func TestAuthorityRejectsCommerceOutsideActiveInteraction(t *testing.T) {
	state, err := NewState(Layout{VendorGrid: Grid{Width: 2, Height: 2}, Gold: GoldBalance{Carried: 100}}, []Item{
		{ID: "held", Code: "cap", Width: 1, Height: 1, BaseCost: 10},
		{ID: "stock", Code: "hp1", Width: 1, Height: 1, BaseCost: 10},
	}, map[string]Placement{
		"held":  {Container: ContainerHeld},
		"stock": {Container: ContainerVendor, Slot: "misc"},
	})
	if err != nil {
		t.Fatal(err)
	}
	authority := NewAuthority()
	authority.SetTradeCatalog(TradeCatalog{"akara": {BuyMultiplier: 1024, SellMultiplier: 1024}})
	authority.SetInteractionPolicy(interactionPolicyStub{vendor: "charsi"})
	if err := authority.Register("alice", state); err != nil {
		t.Fatal(err)
	}
	if err := authority.sellHeld("alice", "held", "akara", "misc"); err == nil {
		t.Fatal("sale outside active vendor interaction succeeded")
	}
	if err := authority.buyToHeld("alice", "stock", "akara"); err == nil {
		t.Fatal("purchase outside active vendor interaction succeeded")
	}
	_, _, placements, err := authority.Snapshot("alice")
	if err != nil {
		t.Fatal(err)
	}
	if placements["held"].Container != ContainerHeld || placements["stock"].Container != ContainerVendor {
		t.Fatalf("rejected commerce mutated state: %#v", placements)
	}
}
