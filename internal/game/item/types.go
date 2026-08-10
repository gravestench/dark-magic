// Package item owns authoritative item identity and container placement rules.
// It deliberately knows nothing about pixels, panels, Lua, or Raylib.
package item

// Container names the kind of place holding an item.
type Container string

const (
	ContainerWorld     Container = "world"
	ContainerInventory Container = "inventory"
	ContainerStash     Container = "stash"
	ContainerCube      Container = "cube"
	ContainerEquipment Container = "equipment"
	ContainerHireling  Container = "hireling"
	ContainerBelt      Container = "belt"
	ContainerHeld      Container = "held"
	ContainerVendor    Container = "vendor"
	ContainerQuest     Container = "quest_service"
)

// Item contains only facts needed to decide where an item can go. Rich loot
// properties remain attached to the item identity elsewhere.
type Item struct {
	ID           string
	Code         string
	Width        int
	Height       int
	BodySlots    []string
	BeltEligible bool
	Presentation Presentation
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
}

type Grid struct{ Width, Height int }

type Layout struct {
	Grids        map[Container]Grid
	BeltCapacity int
}
