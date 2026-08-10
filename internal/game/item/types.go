// Package item owns authoritative item identity and container placement rules.
// It deliberately knows nothing about pixels, panels, Lua, or Raylib.
package item

// Container names the kind of place holding an item.
type Container string

const (
	ContainerWorld     Container = "world"
	ContainerInventory Container = "inventory"
	ContainerEquipment Container = "equipment"
	ContainerBelt      Container = "belt"
	ContainerCursor    Container = "cursor"
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

type Layout struct {
	InventoryWidth  int
	InventoryHeight int
	BeltCapacity    int
}
