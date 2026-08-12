// Package worldgen defines the generic, deterministic recipe boundary between
// a mod's world-generation policy and engine-owned world materialization.
//
// A generator decides what the world is. Asset loading and presentation decide
// how that world looks. Keeping those jobs separate lets a server generate and
// verify a zone without Raylib, textures, or even the original MPQ files.
package worldgen

import (
	"errors"
	"fmt"
	"strings"
)

const ContractVersion = 1

var (
	ErrRequest = errors.New("mapgen: invalid request")
	ErrZone    = errors.New("mapgen: invalid zone")
)

// Difficulty uses the ordering authored by the legacy level tables.
type Difficulty uint8

const (
	Normal Difficulty = iota
	Nightmare
	Hell
)

// Request contains every input allowed to influence deterministic generation.
// Version makes future algorithm changes explicit instead of silently changing
// an old seed's result.
type Request struct {
	Version    uint32     `json:"version"`
	Seed       uint64     `json:"seed"`
	Act        uint8      `json:"act"`
	LevelID    int        `json:"level_id"`
	Difficulty Difficulty `json:"difficulty"`
}

func (request Request) Validate() error {
	if request.Version != ContractVersion {
		return fmt.Errorf("%w: unsupported version %d", ErrRequest, request.Version)
	}
	if request.Act < 1 || request.Act > 5 {
		return fmt.Errorf("%w: act %d is outside 1..5", ErrRequest, request.Act)
	}
	if request.LevelID <= 0 {
		return fmt.Errorf("%w: level ID must be positive", ErrRequest)
	}
	if request.Difficulty > Hell {
		return fmt.Errorf("%w: difficulty %d is outside 0..2", ErrRequest, request.Difficulty)
	}
	return nil
}

// Generator turns one complete request into one immutable zone description.
// Implementations may use tables or decoded map facts supplied at construction,
// but Generate must not depend on presentation state or wall-clock time.
type Generator interface {
	Generate(Request) (*Zone, error)
}

// Kind names the legacy generation family used for a zone.
type Kind string

const (
	Preset  Kind = "preset"
	Maze    Kind = "maze"
	Outdoor Kind = "outdoor"
)

// Bounds is a half-open rectangle in world-tile coordinates.
type Bounds struct {
	X, Y, Width, Height int
}

func (bounds Bounds) valid() bool { return bounds.Width > 0 && bounds.Height > 0 }

// Stamp places one authored DS1 recipe. TilePaths are the deterministic DT1
// inputs selected from LvlTypes plus masks; they are asset identities, not
// loaded graphics.
type Stamp struct {
	ID           uint32
	PresetDef    int
	Role         string
	X, Y         int
	Width        int
	Height       int
	DS1Path      string
	TilePaths    []string
	Variant      int
	Populate     bool
	LogicalWalls bool
	Overlay      bool
}

// Room is a simulation-space rectangle. StampID is zero only for an unassigned
// room in a diagnostic topology.
type Room struct {
	ID      uint32
	X, Y    int
	Width   int
	Height  int
	StampID uint32
}

// Link joins two rooms. From is always lower than To in canonical zones.
type Link struct {
	From, To uint32
}

// Warp preserves an authored level transition without creating presentation UI.
type Warp struct {
	ID               uint32
	Role             string
	Direction        string
	X, Y             int
	DestinationLevel int
}

// Spawn is an authoritative placement request recovered from an authored stamp.
type Spawn struct {
	ID       uint32
	Kind     string
	RecordID int
	X, Y     int
}

// PathTile reserves one world-tile cell for a generated semantic route.
// Materialization may realize it with level-specific DT1 floor identities;
// simulation and tests can reason about connectivity before assets are loaded.
type PathTile struct {
	X, Y int
}

// StructureTile reserves one world-tile cell for an outdoor structural layer.
// Kind describes the simulation meaning; presentation later chooses suitable
// DT1 artwork. Passable is authoritative and makes bridge openings explicit.
type StructureTile struct {
	X, Y     int
	Kind     string
	Passable bool
}

// Definition is the mutable input accepted by NewZone. The constructor copies,
// validates, and canonicalizes it before exposing an immutable Zone.
type Definition struct {
	Request    Request
	Kind       Kind
	Bounds     Bounds
	Stamps     []Stamp
	Rooms      []Room
	Links      []Link
	Warps      []Warp
	Spawns     []Spawn
	Paths      []PathTile
	Structures []StructureTile
	Trace      []string
}

func validateDefinition(def Definition) error {
	if err := def.Request.Validate(); err != nil {
		return err
	}
	if def.Kind != Preset && def.Kind != Maze && def.Kind != Outdoor {
		return fmt.Errorf("%w: unknown kind %q", ErrZone, def.Kind)
	}
	if !def.Bounds.valid() {
		return fmt.Errorf("%w: bounds must have positive dimensions", ErrZone)
	}
	stampIDs := make(map[uint32]struct{}, len(def.Stamps))
	for _, stamp := range def.Stamps {
		if stamp.ID == 0 || stamp.Width <= 0 || stamp.Height <= 0 || strings.TrimSpace(stamp.DS1Path) == "" {
			return fmt.Errorf("%w: incomplete stamp %d", ErrZone, stamp.ID)
		}
		if _, duplicate := stampIDs[stamp.ID]; duplicate {
			return fmt.Errorf("%w: duplicate stamp %d", ErrZone, stamp.ID)
		}
		stampIDs[stamp.ID] = struct{}{}
	}
	roomIDs := make(map[uint32]struct{}, len(def.Rooms))
	for _, room := range def.Rooms {
		if room.ID == 0 || room.Width <= 0 || room.Height <= 0 {
			return fmt.Errorf("%w: incomplete room %d", ErrZone, room.ID)
		}
		if _, duplicate := roomIDs[room.ID]; duplicate {
			return fmt.Errorf("%w: duplicate room %d", ErrZone, room.ID)
		}
		if room.StampID != 0 {
			if _, found := stampIDs[room.StampID]; !found {
				return fmt.Errorf("%w: room %d references stamp %d", ErrZone, room.ID, room.StampID)
			}
		}
		roomIDs[room.ID] = struct{}{}
	}
	for _, link := range def.Links {
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
	warpIDs := make(map[uint32]struct{}, len(def.Warps))
	for _, warp := range def.Warps {
		if warp.ID == 0 || warp.DestinationLevel <= 0 || warp.X < def.Bounds.X || warp.Y < def.Bounds.Y || warp.X >= def.Bounds.X+def.Bounds.Width || warp.Y >= def.Bounds.Y+def.Bounds.Height {
			return fmt.Errorf("%w: incomplete or out-of-bounds warp %d", ErrZone, warp.ID)
		}
		if warp.Direction != "north" && warp.Direction != "east" && warp.Direction != "south" && warp.Direction != "west" {
			return fmt.Errorf("%w: warp %d has invalid direction %q", ErrZone, warp.ID, warp.Direction)
		}
		if _, duplicate := warpIDs[warp.ID]; duplicate {
			return fmt.Errorf("%w: duplicate warp %d", ErrZone, warp.ID)
		}
		warpIDs[warp.ID] = struct{}{}
	}
	seenPath := make(map[PathTile]struct{}, len(def.Paths))
	for _, tile := range def.Paths {
		if tile.X < def.Bounds.X || tile.Y < def.Bounds.Y || tile.X >= def.Bounds.X+def.Bounds.Width || tile.Y >= def.Bounds.Y+def.Bounds.Height {
			return fmt.Errorf("%w: out-of-bounds path tile %d,%d", ErrZone, tile.X, tile.Y)
		}
		if _, duplicate := seenPath[tile]; duplicate {
			return fmt.Errorf("%w: duplicate path tile %d,%d", ErrZone, tile.X, tile.Y)
		}
		seenPath[tile] = struct{}{}
	}
	seenStructure := make(map[[2]int]struct{}, len(def.Structures))
	for _, tile := range def.Structures {
		if tile.X < def.Bounds.X || tile.Y < def.Bounds.Y || tile.X >= def.Bounds.X+def.Bounds.Width || tile.Y >= def.Bounds.Y+def.Bounds.Height {
			return fmt.Errorf("%w: out-of-bounds structure tile %d,%d", ErrZone, tile.X, tile.Y)
		}
		if tile.Kind != "river" && tile.Kind != "cliff" && tile.Kind != "bridge" {
			return fmt.Errorf("%w: unknown structure kind %q", ErrZone, tile.Kind)
		}
		position := [2]int{tile.X, tile.Y}
		if _, duplicate := seenStructure[position]; duplicate {
			return fmt.Errorf("%w: overlapping structure tile %d,%d", ErrZone, tile.X, tile.Y)
		}
		seenStructure[position] = struct{}{}
		if tile.Kind != "bridge" && tile.Passable {
			return fmt.Errorf("%w: %s %d,%d cannot be passable", ErrZone, tile.Kind, tile.X, tile.Y)
		}
	}
	return nil
}
