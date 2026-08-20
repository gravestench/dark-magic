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
)

const ContractVersion = 1

var (
	ErrRequest = errors.New("mapgen: invalid request")
	ErrZone    = errors.New("mapgen: invalid zone")
)

// Difficulty is an opaque mod-defined variant number. The engine records it so
// recipes can be replayed, but it does not assign meaning to individual values.
type Difficulty uint8

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

// Validate checks the engine-owned replay identity without interpreting opaque mod dimensions.
// Accepting unknown acts and difficulties keeps the contract reusable while version and level checks prevent ambiguity.
func (request Request) Validate() error {
	if request.Version != ContractVersion {
		return fmt.Errorf("%w: unsupported version %d", ErrRequest, request.Version)
	}

	if request.LevelID <= 0 {
		return fmt.Errorf("%w: level ID must be positive", ErrRequest)
	}

	return nil
}

// Generator turns one complete request into one immutable zone description.
// Implementations may use tables or decoded map facts supplied at construction,
// but Generate must not depend on presentation state or wall-clock time.
type Generator interface {
	Generate(Request) (*Zone, error)
}

// Kind is an opaque, stable policy identifier supplied by the mod.
type Kind string

// Bounds is a half-open rectangle in world-tile coordinates.
type Bounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// valid reports whether the half-open rectangle can contain world tiles.
// Coordinates may be negative because only dimensions, not placement policy, belong to this generic contract.
func (bounds Bounds) valid() bool {
	return bounds.Width > 0 && bounds.Height > 0
}

// Stamp places one authored DS1 recipe. TilePaths are the deterministic DT1
// inputs selected from LvlTypes plus masks; they are asset identities, not
// loaded graphics.
type Stamp struct {
	ID           uint32   `json:"id"`
	PresetDef    int      `json:"preset_def,omitempty"`
	Role         string   `json:"role,omitempty"`
	X            int      `json:"x"`
	Y            int      `json:"y"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	DS1Path      string   `json:"ds1_path"`
	TilePaths    []string `json:"tile_paths,omitempty"`
	Variant      int      `json:"variant,omitempty"`
	Populate     bool     `json:"populate,omitempty"`
	LogicalWalls bool     `json:"logical_walls,omitempty"`
	Overlay      bool     `json:"overlay,omitempty"`
}

// Room is a simulation-space rectangle. StampID is zero only for an unassigned
// room in a diagnostic topology.
type Room struct {
	ID      uint32 `json:"id"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	StampID uint32 `json:"stamp_id,omitempty"`
}

// Link joins two rooms. From is always lower than To in canonical zones.
type Link struct {
	From uint32 `json:"from"`
	To   uint32 `json:"to"`
}

// Warp preserves an authored level transition without creating presentation UI.
type Warp struct {
	ID               uint32 `json:"id"`
	Role             string `json:"role,omitempty"`
	Direction        string `json:"direction"`
	X                int    `json:"x"`
	Y                int    `json:"y"`
	DestinationLevel int    `json:"destination_level"`
}

// Spawn is an authoritative placement request recovered from an authored stamp.
type Spawn struct {
	ID       uint32 `json:"id"`
	Kind     string `json:"kind"`
	RecordID int    `json:"record_id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

// PathTile reserves one world-tile cell for a generated semantic route.
// Materialization may realize it with level-specific DT1 floor identities;
// simulation and tests can reason about connectivity before assets are loaded.
type PathTile struct {
	X         int   `json:"x"`
	Y         int   `json:"y"`
	MainIndex int32 `json:"main_index,omitempty"`
	SubIndex  int32 `json:"sub_index,omitempty"`
}

// StructureTile reserves one world-tile cell for an outdoor structural layer.
// Kind describes the simulation meaning; presentation later chooses suitable
// DT1 artwork. Passable is authoritative and makes bridge openings explicit.
type StructureTile struct {
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Kind     string `json:"kind"`
	Passable bool   `json:"passable"`
}

// Definition is the mutable input accepted by NewZone. The constructor copies,
// validates, and canonicalizes it before exposing an immutable Zone.
type Definition struct {
	Request    Request         `json:"request"`
	Kind       Kind            `json:"kind"`
	Bounds     Bounds          `json:"bounds"`
	Stamps     []Stamp         `json:"stamps,omitempty"`
	Rooms      []Room          `json:"rooms,omitempty"`
	Links      []Link          `json:"links,omitempty"`
	Warps      []Warp          `json:"warps,omitempty"`
	Spawns     []Spawn         `json:"spawns,omitempty"`
	Paths      []PathTile      `json:"paths,omitempty"`
	Structures []StructureTile `json:"structures,omitempty"`
	Trace      []string        `json:"trace,omitempty"`
}
