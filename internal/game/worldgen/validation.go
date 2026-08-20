package worldgen

import (
	"fmt"
	"strings"
)

// validateDefinition admits a complete zone recipe in dependency order.
// Reference checks follow identity checks so failures point to malformed owners before their dependents.
func validateDefinition(definition Definition) error {
	if err := definition.Request.Validate(); err != nil {
		return err
	}

	if strings.TrimSpace(string(definition.Kind)) == "" {
		return fmt.Errorf("%w: kind is required", ErrZone)
	}

	if !definition.Bounds.valid() {
		return fmt.Errorf("%w: bounds must have positive dimensions", ErrZone)
	}

	stampIDs, err := validateStamps(definition.Stamps)
	if err != nil {
		return err
	}

	roomIDs, err := validateRooms(definition.Rooms, stampIDs)
	if err != nil {
		return err
	}

	if err := validateLinks(definition.Links, roomIDs); err != nil {
		return err
	}

	if err := validateWarps(definition.Warps, definition.Bounds); err != nil {
		return err
	}

	if err := validatePaths(definition.Paths, definition.Bounds); err != nil {
		return err
	}

	return validateStructures(definition.Structures, definition.Bounds)
}

// validateStamps checks authored recipe identities before rooms are allowed to reference them.
// Returning the identity set makes the dependency explicit and avoids rebuilding it during room validation.
func validateStamps(stamps []Stamp) (map[uint32]struct{}, error) {
	identities := make(map[uint32]struct{}, len(stamps))
	for _, stamp := range stamps {
		if stamp.ID == 0 || stamp.Width <= 0 || stamp.Height <= 0 || strings.TrimSpace(stamp.DS1Path) == "" {
			return nil, fmt.Errorf("%w: incomplete stamp %d", ErrZone, stamp.ID)
		}

		if _, duplicate := identities[stamp.ID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate stamp %d", ErrZone, stamp.ID)
		}

		identities[stamp.ID] = struct{}{}
	}

	return identities, nil
}

// validateRooms checks simulation rectangles and rejects references to stamps outside the admitted recipe.
func validateRooms(rooms []Room, stampIDs map[uint32]struct{}) (map[uint32]struct{}, error) {
	identities := make(map[uint32]struct{}, len(rooms))
	for _, room := range rooms {
		if room.ID == 0 || room.Width <= 0 || room.Height <= 0 {
			return nil, fmt.Errorf("%w: incomplete room %d", ErrZone, room.ID)
		}

		if _, duplicate := identities[room.ID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate room %d", ErrZone, room.ID)
		}

		if room.StampID != 0 {
			if _, found := stampIDs[room.StampID]; !found {
				return nil, fmt.Errorf("%w: room %d references stamp %d", ErrZone, room.ID, room.StampID)
			}
		}

		identities[room.ID] = struct{}{}
	}

	return identities, nil
}

// validateLinks ensures every undirected edge connects two distinct admitted rooms.
func validateLinks(links []Link, roomIDs map[uint32]struct{}) error {
	for _, link := range links {
		if link.From == link.To {
			return fmt.Errorf("%w: room %d links to itself", ErrZone, link.From)
		}

		if _, found := roomIDs[link.From]; !found {
			return fmt.Errorf("%w: link references room %d", ErrZone, link.From)
		}

		if _, found := roomIDs[link.To]; !found {
			return fmt.Errorf("%w: link references room %d", ErrZone, link.To)
		}
	}

	return nil
}

// validateWarps checks transition identities, cardinal directions, and placement inside the zone's half-open bounds.
func validateWarps(warps []Warp, bounds Bounds) error {
	identities := make(map[uint32]struct{}, len(warps))
	for _, warp := range warps {
		if warp.ID == 0 || warp.DestinationLevel <= 0 || !contains(bounds, warp.X, warp.Y) {
			return fmt.Errorf("%w: incomplete or out-of-bounds warp %d", ErrZone, warp.ID)
		}

		if !isCardinalDirection(warp.Direction) {
			return fmt.Errorf("%w: warp %d has invalid direction %q", ErrZone, warp.ID, warp.Direction)
		}

		if _, duplicate := identities[warp.ID]; duplicate {
			return fmt.Errorf("%w: duplicate warp %d", ErrZone, warp.ID)
		}

		identities[warp.ID] = struct{}{}
	}

	return nil
}

// validatePaths rejects duplicate authored route cells and any route that escapes authoritative zone bounds.
func validatePaths(paths []PathTile, bounds Bounds) error {
	seen := make(map[PathTile]struct{}, len(paths))
	for _, tile := range paths {
		if !contains(bounds, tile.X, tile.Y) {
			return fmt.Errorf("%w: out-of-bounds path tile %d,%d", ErrZone, tile.X, tile.Y)
		}

		if _, duplicate := seen[tile]; duplicate {
			return fmt.Errorf("%w: duplicate path tile %d,%d", ErrZone, tile.X, tile.Y)
		}

		seen[tile] = struct{}{}
	}

	return nil
}

// validateStructures enforces one named structural meaning per world-tile position.
// Passability is deliberately not interpreted here because it is authoritative mod policy.
func validateStructures(structures []StructureTile, bounds Bounds) error {
	seen := make(map[[2]int]struct{}, len(structures))
	for _, tile := range structures {
		if !contains(bounds, tile.X, tile.Y) {
			return fmt.Errorf("%w: out-of-bounds structure tile %d,%d", ErrZone, tile.X, tile.Y)
		}

		if strings.TrimSpace(tile.Kind) == "" {
			return fmt.Errorf("%w: structure kind is required", ErrZone)
		}

		position := [2]int{tile.X, tile.Y}
		if _, duplicate := seen[position]; duplicate {
			return fmt.Errorf("%w: overlapping structure tile %d,%d", ErrZone, tile.X, tile.Y)
		}

		seen[position] = struct{}{}
	}

	return nil
}

// contains applies half-open bounds consistently to every coordinate-bearing recipe element.
func contains(bounds Bounds, x, y int) bool {
	return x >= bounds.X && y >= bounds.Y && x < bounds.X+bounds.Width && y < bounds.Y+bounds.Height
}

// isCardinalDirection limits warps to the four stable serialized values understood by generic adapters.
func isCardinalDirection(direction string) bool {
	return direction == "north" || direction == "east" || direction == "south" || direction == "west"
}
