// Package mapgen defines the deterministic boundary between Diablo II world
// generation and the systems that consume its result.
//
// A generator decides what the world is. Asset loading and presentation decide
// how that world looks. Keeping those jobs separate lets a server generate and
// verify a zone without Raylib, textures, or even the original MPQ files.
package mapgen

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

// Definition is the mutable input accepted by NewZone. The constructor copies,
// validates, and canonicalizes it before exposing an immutable Zone.
type Definition struct {
	Request Request
	Kind    Kind
	Bounds  Bounds
	Stamps  []Stamp
	Rooms   []Room
	Links   []Link
	Warps   []Warp
	Spawns  []Spawn
	Trace   []string
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
	return nil
}
