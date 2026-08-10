// Package item owns authoritative item identity and container placement rules.
// It deliberately knows nothing about pixels, panels, Lua, or Raylib.
package item

// Container names the kind of place holding an item.
type Container string

const (
	// ContainerWorld places an item in authoritative world state.
	ContainerWorld Container = "world"
	// ContainerInventory is the player's backpack grid.
	ContainerInventory Container = "inventory"
	// ContainerStash is durable personal storage.
	ContainerStash Container = "stash"
	// ContainerCube is the Horadric Cube grid.
	ContainerCube Container = "cube"
	// ContainerEquipment contains player body slots.
	ContainerEquipment Container = "equipment"
	// ContainerHireling contains the hireling's independent body slots.
	ContainerHireling Container = "hireling"
	// ContainerBelt contains indexed quick-use slots.
	ContainerBelt Container = "belt"
	// ContainerHeld is persistent in-hand authority, not transient cursor UI.
	ContainerHeld Container = "held"
	// ContainerVendor contains authority-arranged paged stock.
	ContainerVendor Container = "vendor"
	// ContainerQuest contains named quest or service escrow sockets.
	ContainerQuest Container = "quest_service"
)

// Item contains only facts needed to decide where an item can go. Rich loot
// properties remain attached to the item identity elsewhere.
type Item struct {
	ID              string
	Code            string
	Width           int
	Height          int
	BodySlots       []string
	BeltEligible    bool
	BaseCost        int64
	AppliedServices []string
	Presentation    Presentation
}

// Presentation keeps the two original views explicit. World drops may animate
// a generic or item-specific DC6; panels and the held cursor use InventoryDC6.
type Presentation struct {
	WorldDC6      string
	InventoryDC6  string
	WorldAnimated bool
}

// Placement says where one item currently lives. X and Y are inventory cells;
// Slot is a body-slot code or zero-based belt slot.
type Placement struct {
	Container Container
	X         int
	Y         int
	Slot      string
	BeltSlot  int
	WeaponSet int
	Page      int
}

// Grid is a container's cell dimensions, independent of presentation pixels.
type Grid struct{ Width, Height int }

// Layout contains container rules and balances that must move with owner state.
type Layout struct {
	Grids           map[Container]Grid
	BeltCapacity    int
	ActiveWeaponSet int
	VendorGrid      Grid
	Gold            GoldBalance
}

// GoldBalance mirrors Diablo II's distinct carried and stash-bank stats.
type GoldBalance struct{ Carried, Stashed int64 }
