package item

import (
	"bytes"
	"errors"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestItemCommandsReplayCompleteAuthorityAndDetectConfigurationAndStateDesync(t *testing.T) {
	buildAuthority := func(multiplier int64) *Authority {
		state, err := NewState(Layout{Grids: map[Container]Grid{ContainerInventory: {Width: 4, Height: 4}}, VendorGrid: Grid{Width: 4, Height: 4}, Gold: GoldBalance{Carried: 1000}}, []Item{
			{ID: "sale", Code: "ssd", Width: 1, Height: 2, BaseCost: 100},
			{ID: "stock", Code: "cap", Width: 1, Height: 1, BaseCost: 200},
			{ID: "target", Code: "rin", Width: 1, Height: 1},
			{ID: "material", Code: "gem", Width: 1, Height: 1},
		}, map[string]Placement{
			"sale": {Container: ContainerInventory}, "stock": {Container: ContainerVendor, Slot: "armor"},
			"target": {Container: ContainerQuest, Slot: "target"}, "material": {Container: ContainerQuest, Slot: "material"},
		})
		if err != nil {
			t.Fatal(err)
		}
		authority := NewAuthority()
		authority.SetTradeCatalog(TradeCatalog{"akara": {BuyMultiplier: multiplier, SellMultiplier: 1024, MaxBuy: 1000}})
		authority.SetServiceCatalog(ServiceCatalog{"imbue": {ID: "imbue", TargetSlot: "target", ConsumeSlots: []string{"material"}, GoldCost: 25}})
		if err := authority.Register("alice", state); err != nil {
			t.Fatal(err)
		}
		return authority
	}
	apply := func(authority *Authority, command simulation.Command) error {
		switch command.Kind {
		case MoveCommand:
			payload, err := decodeMove(command.Payload)
			if err != nil {
				return err
			}
			_, err = authority.move(command.Player, payload.ItemID, payload.Destination, payload.PlaceHeld)
			return err
		case VendorSellCommand, VendorBuyCommand:
			payload, err := decodeVendor(command, command.Kind == VendorSellCommand)
			if err != nil {
				return err
			}
			if command.Kind == VendorSellCommand {
				return authority.sellHeld(nil, command.Player, payload.ItemID, payload.Vendor, payload.Category)
			}
			return authority.buyToHeld(nil, command.Player, payload.ItemID, payload.Vendor)
		case ServiceCommand:
			payload, err := decodeService(command.Payload)
			if err != nil {
				return err
			}
			return authority.completeService(nil, command.Player, payload.Service)
		case WeaponSetCommand:
			payload, err := decodeWeaponSet(command.Payload)
			if err != nil {
				return err
			}
			return authority.selectWeaponSet(command.Player, payload.Set)
		default:
			return errors.New("unknown item command")
		}
	}

	original := buildAuthority(512)
	session, err := gamesession.New(gameecs.New(), gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterCommands(session, original); err != nil {
		t.Fatal(err)
	}
	commands := []simulation.Command{}
	move, _ := Command(MovePayload{ItemID: "sale", Destination: Placement{Container: ContainerHeld}}, "alice", 1, 1, simulation.AuthorityPlayer)
	sell, _ := VendorCommand(VendorSellCommand, VendorPayload{ItemID: "sale", Vendor: "akara", Category: "weapons"}, "alice", 2, 2, simulation.AuthorityPlayer)
	buy, _ := VendorCommand(VendorBuyCommand, VendorPayload{ItemID: "stock", Vendor: "akara"}, "alice", 3, 3, simulation.AuthorityPlayer)
	service, _ := ServiceCompletionCommand(ServicePayload{Service: "imbue"}, "alice", 4, 4, simulation.AuthorityPlayer)
	weapons, _ := WeaponSetSelectionCommand(WeaponSetPayload{Set: 1}, "alice", 5, 5, simulation.AuthorityPlayer)
	commands = append(commands, move, sell, buy, service, weapons)
	for _, command := range commands {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	want, err := original.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}

	restored := buildAuthority(512)
	if err := simulation.VerifyReplay(replay, nil, func(_ *gameecs.Engine, command simulation.Command) error {
		return apply(restored, command)
	}, restored); err != nil {
		t.Fatal(err)
	}
	got, err := restored.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("verified replay did not reconstruct final item authority")
	}

	mismatched := buildAuthority(513)
	if err := simulation.VerifyReplay(replay, nil, func(_ *gameecs.Engine, command simulation.Command) error {
		return apply(mismatched, command)
	}, mismatched); err == nil || !errors.Is(err, simulation.ErrReplay) {
		t.Fatalf("configuration mismatch = %v", err)
	}

	tampered := replay
	tampered.Commands = append([]simulation.Command(nil), replay.Commands...)
	tampered.Commands[4] = weapons
	tampered.Commands[4].Payload = []byte(`{"set":0}`)
	desynced := buildAuthority(512)
	var desync *simulation.DesyncError
	err = simulation.VerifyReplay(tampered, nil, func(_ *gameecs.Engine, command simulation.Command) error {
		return apply(desynced, command)
	}, desynced)
	if !errors.As(err, &desync) || desync.Detail != `participant "dm.items/v1" state differs` {
		t.Fatalf("participant desync = %#v, %v", desync, err)
	}
}

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

func TestServiceCommandResolvesServerOwnedSocketsAndReplays(t *testing.T) {
	state, err := NewState(Layout{Gold: GoldBalance{Carried: 50}}, []Item{{ID: "target", Code: "ssd", Width: 1, Height: 3}}, map[string]Placement{"target": {Container: ContainerQuest, Slot: "target"}})
	if err != nil {
		t.Fatal(err)
	}
	authority := NewAuthority()
	authority.SetServiceCatalog(ServiceCatalog{"imbue": {ID: "imbue", TargetSlot: "target", GoldCost: 25}})
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
	command, _ := ServiceCompletionCommand(ServicePayload{Service: "imbue"}, "alice", 1, 1, simulation.AuthorityPlayer)
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	layout, items, _, err := authority.Snapshot("alice")
	if err != nil || layout.Gold.Carried != 25 || items["target"].AppliedServices[0] != "imbue" {
		t.Fatalf("snapshot = %#v %#v %v", layout, items, err)
	}
	replay, err := session.Replay()
	if err != nil || len(replay.Commands) != 1 || replay.Commands[0].Kind != ServiceCommand {
		t.Fatalf("replay = %#v, %v", replay, err)
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
