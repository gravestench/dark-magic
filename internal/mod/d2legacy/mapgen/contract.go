// Package mapgen contains Diablo II's world-generation policy.
//
// The engine-owned worldgen package deliberately contains only the immutable
// recipe and validation mechanism. These aliases keep the policy code readable
// while making ownership explicit: d2legacy chooses recipes; the engine merely
// validates and materializes them.
package mapgen

import worldgen "github.com/gravestench/dark-magic/internal/game/worldgen"

const ContractVersion = worldgen.ContractVersion

var (
	ErrRequest = worldgen.ErrRequest
	ErrZone    = worldgen.ErrZone
	NewZone    = worldgen.NewZone
)

type Difficulty = worldgen.Difficulty

const (
	Normal    = worldgen.Normal
	Nightmare = worldgen.Nightmare
	Hell      = worldgen.Hell
)

type Request = worldgen.Request
type Generator = worldgen.Generator
type Kind = worldgen.Kind

const (
	Preset  = worldgen.Preset
	Maze    = worldgen.Maze
	Outdoor = worldgen.Outdoor
)

type Bounds = worldgen.Bounds
type Stamp = worldgen.Stamp
type Room = worldgen.Room
type Link = worldgen.Link
type Warp = worldgen.Warp
type Spawn = worldgen.Spawn
type PathTile = worldgen.PathTile
type StructureTile = worldgen.StructureTile
type Definition = worldgen.Definition
type Zone = worldgen.Zone
