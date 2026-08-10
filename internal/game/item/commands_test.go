package item

import (
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestMoveCommandIsValidatedAppliedAndAudited(t *testing.T) {
	state := testCommandState(t)
	authority := NewAuthority()
	if err := authority.Register("alice", state); err != nil {
		t.Fatal(err)
	}
	session, err := gamesession.New(gameecs.New(), gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterCommands(session, authority); err != nil {
		t.Fatal(err)
	}
	command, err := Command(MovePayload{ItemID: "potion", Destination: Placement{Container: ContainerHeld}}, "alice", 1, 1, simulation.AuthorityPlayer)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	_, _, placements, err := authority.Snapshot("alice")
	if err != nil || placements["potion"].Container != ContainerHeld {
		t.Fatalf("placement = %#v, %v", placements["potion"], err)
	}
	replay, err := session.Replay()
	if err != nil || len(replay.Commands) != 1 || replay.Commands[0].Kind != MoveCommand {
		t.Fatalf("replay = %#v, %v", replay, err)
	}
}

func TestPlayerCannotTargetAnotherOwner(t *testing.T) {
	payload := MovePayload{Owner: "bob", ItemID: "potion", Destination: Placement{Container: ContainerHeld}}
	command, err := Command(payload, "alice", 1, 1, simulation.AuthorityPlayer)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateMoveCommand(command); err == nil {
		t.Fatal("cross-owner player command was accepted")
	}
}

func TestHeldPlacementAcceptsQuestSocketButNotVendorStock(t *testing.T) {
	for _, test := range []struct {
		container Container
		wantError bool
	}{
		{container: ContainerQuest},
		{container: ContainerVendor, wantError: true},
	} {
		payload := MovePayload{ItemID: "potion", Destination: Placement{Container: test.container, Slot: "input"}, PlaceHeld: true}
		command, err := Command(payload, "alice", 1, 1, simulation.AuthorityPlayer)
		if err != nil {
			t.Fatal(err)
		}
		err = validateMoveCommand(command)
		if (err != nil) != test.wantError {
			t.Fatalf("%s validation error = %v, want error %t", test.container, err, test.wantError)
		}
	}
}

func TestAdministratorMoveNamesOwnerAndAppearsInAudit(t *testing.T) {
	authority := NewAuthority()
	if err := authority.Register("alice", testCommandState(t)); err != nil {
		t.Fatal(err)
	}
	session, err := gamesession.New(gameecs.New(), gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterCommands(session, authority); err != nil {
		t.Fatal(err)
	}
	command, err := Command(MovePayload{Owner: "alice", ItemID: "potion", Destination: Placement{Container: ContainerHeld}}, "admin", 1, 1, simulation.AuthorityAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if audit := session.Audit(); len(audit) != 1 || audit[0].Kind != MoveCommand {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestWeaponSetCommandIsAppliedAndReplayed(t *testing.T) {
	authority := NewAuthority()
	if err := authority.Register("alice", testCommandState(t)); err != nil {
		t.Fatal(err)
	}
	session, err := gamesession.New(gameecs.New(), gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterCommands(session, authority); err != nil {
		t.Fatal(err)
	}
	command, err := WeaponSetSelectionCommand(WeaponSetPayload{Set: 1}, "alice", 1, 1, simulation.AuthorityPlayer)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	layout, _, _, err := authority.Snapshot("alice")
	if err != nil || layout.ActiveWeaponSet != 1 {
		t.Fatalf("active set = %d, %v", layout.ActiveWeaponSet, err)
	}
	replay, err := session.Replay()
	if err != nil || len(replay.Commands) != 1 || replay.Commands[0].Kind != WeaponSetCommand {
		t.Fatalf("replay = %#v, %v", replay, err)
	}
}

func TestPlayerCannotSelectAnotherOwnersWeaponSet(t *testing.T) {
	command, err := WeaponSetSelectionCommand(WeaponSetPayload{Owner: "bob", Set: 1}, "alice", 1, 1, simulation.AuthorityPlayer)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateWeaponSetCommand(command); err == nil {
		t.Fatal("cross-owner weapon selection was accepted")
	}
}

func TestVendorSaleAndPurchaseCommandsAreReplayed(t *testing.T) {
	state, err := NewState(Layout{VendorGrid: Grid{Width: 2, Height: 2}, Gold: GoldBalance{Carried: 1000}}, []Item{
		{ID: "held", Code: "ssd", Width: 1, Height: 2, BaseCost: 100}, {ID: "stock", Code: "cap", Width: 1, Height: 1, BaseCost: 200},
	}, map[string]Placement{"held": {Container: ContainerHeld}, "stock": {Container: ContainerVendor, Slot: "armor"}})
	if err != nil {
		t.Fatal(err)
	}
	authority := NewAuthority()
	authority.SetTradeCatalog(TradeCatalog{"akara": {BuyMultiplier: 512, SellMultiplier: 1024, MaxBuy: 1000}})
	if err := authority.Register("alice", state); err != nil {
		t.Fatal(err)
	}
	session, err := gamesession.New(gameecs.New(), gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterCommands(session, authority); err != nil {
		t.Fatal(err)
	}
	sell, _ := VendorCommand(VendorSellCommand, VendorPayload{ItemID: "held", Vendor: "Akara", Category: "weapons"}, "alice", 1, 1, simulation.AuthorityPlayer)
	if err := session.Submit(sell); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	buy, _ := VendorCommand(VendorBuyCommand, VendorPayload{ItemID: "stock", Vendor: "Akara"}, "alice", 2, 2, simulation.AuthorityPlayer)
	if err := session.Submit(buy); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	_, _, placements, err := authority.Snapshot("alice")
	if err != nil || placements["held"].Container != ContainerVendor || placements["stock"].Container != ContainerHeld {
		t.Fatalf("placements = %#v, %v", placements, err)
	}
	layout, _, _, _ := authority.Snapshot("alice")
	if layout.Gold.Carried != 850 {
		t.Fatalf("carried gold = %d, want 850", layout.Gold.Carried)
	}
	replay, err := session.Replay()
	if err != nil || len(replay.Commands) != 2 || replay.Commands[0].Kind != VendorSellCommand || replay.Commands[1].Kind != VendorBuyCommand {
		t.Fatalf("replay = %#v, %v", replay, err)
	}
}

func TestVendorCommandRejectsClientCoordinatesAndCrossOwner(t *testing.T) {
	move, err := Command(MovePayload{ItemID: "held", Destination: Placement{Container: ContainerVendor, Slot: "weapons", X: 9}}, "alice", 1, 1, simulation.AuthorityPlayer)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateMoveCommand(move); err == nil {
		t.Fatal("client-chosen vendor placement was accepted")
	}
	command, _ := VendorCommand(VendorSellCommand, VendorPayload{Owner: "bob", ItemID: "held", Vendor: "Akara", Category: "weapons"}, "alice", 1, 1, simulation.AuthorityPlayer)
	if _, err := decodeVendor(command, true); err == nil {
		t.Fatal("cross-owner vendor sale was accepted")
	}
}

func testCommandState(t *testing.T) *State {
	t.Helper()
	state, err := NewState(Layout{Grids: map[Container]Grid{ContainerInventory: {Width: 10, Height: 4}}, BeltCapacity: 4}, []Item{{ID: "potion", Code: "hp1", Width: 1, Height: 1, BeltEligible: true}}, map[string]Placement{"potion": {Container: ContainerInventory}})
	if err != nil {
		t.Fatal(err)
	}
	return state
}
