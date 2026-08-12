package interaction

import (
	"errors"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestItemCommerceRevalidatesSpatialInteraction(t *testing.T) {
	for _, test := range []struct {
		name    string
		moveOut bool
		wantErr bool
	}{
		{name: "in range purchase"},
		{name: "walked away", moveOut: true, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine := gameecs.New()
			materializeControlledPosition(t, engine, "alice", 10, 10)
			interactions := testAuthority(t)
			items := gameitem.NewAuthority()
			items.SetInteractionPolicy(interactions)
			items.SetTradeCatalog(gameitem.TradeCatalog{"akara": {BuyMultiplier: 1024, SellMultiplier: 1024}})
			state, err := gameitem.NewState(gameitem.Layout{VendorGrid: gameitem.Grid{Width: 2, Height: 2}, Gold: gameitem.GoldBalance{Carried: 100}}, []gameitem.Item{{ID: "stock", Code: "hp1", Width: 1, Height: 1, BaseCost: 10}}, map[string]gameitem.Placement{"stock": {Container: gameitem.ContainerVendor, Slot: "misc"}})
			if err != nil {
				t.Fatal(err)
			}
			if err := items.Register("alice", state); err != nil {
				t.Fatal(err)
			}
			session, err := gamesession.New(engine, gamesession.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			if err := RegisterCommands(session, interactions); err != nil {
				t.Fatal(err)
			}
			if err := gameitem.RegisterCommands(session, items); err != nil {
				t.Fatal(err)
			}
			open, _ := Command(OpenCommand, Payload{Target: "act1-akara"}, "alice", 1, 1, simulation.AuthorityPlayer)
			if err := session.Submit(open); err != nil {
				t.Fatal(err)
			}
			if err := session.Step(); err != nil {
				t.Fatal(err)
			}
			if test.moveOut {
				moveControlledPosition(t, engine, 100, 100)
			}
			buy, _ := gameitem.VendorCommand(gameitem.VendorBuyCommand, gameitem.VendorPayload{ItemID: "stock", Vendor: "Akara"}, "alice", 2, 2, simulation.AuthorityPlayer)
			if err := session.Submit(buy); err != nil {
				t.Fatal(err)
			}
			err = session.Step()
			if test.wantErr && !errors.Is(err, gamesession.ErrCommandApply) {
				t.Fatalf("out-of-range purchase error = %v", err)
			}
			if !test.wantErr && err != nil {
				t.Fatal(err)
			}
			_, _, placements, snapshotErr := items.Snapshot("alice")
			if snapshotErr != nil {
				t.Fatal(snapshotErr)
			}
			want := gameitem.ContainerHeld
			if test.wantErr {
				want = gameitem.ContainerVendor
			}
			if placements["stock"].Container != want {
				t.Fatalf("stock container = %q, want %q", placements["stock"].Container, want)
			}
		})
	}
}

func moveControlledPosition(t *testing.T, engine *gameecs.Engine, x, y float64) {
	t.Helper()
	positions, found := akara.GetDynamicStore(engine.World(), "d2.world.position")
	if !found {
		t.Fatal("position store missing")
	}
	position, found := positions.Get(positions.Entities()[0])
	if !found {
		t.Fatal("controlled position missing")
	}
	if err := position.Set("x", x); err != nil {
		t.Fatal(err)
	}
	if err := position.Set("y", y); err != nil {
		t.Fatal(err)
	}
}
