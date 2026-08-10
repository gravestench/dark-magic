# Diablo II map composition and procedural generation

Status: research architecture/specification. This document describes Diablo II
1.10f behavior where D2MOO provides reverse-engineered runtime evidence and
separates that behavior from third-party engine architecture. The clean-room
1.14d [libd2](https://github.com/jaenster/libd2) implementation is an additional
high-value source for DRLG rules, call order, and deterministic verification;
version differences and implementation choices must remain explicit.

## Pipeline

```text
Levels / LevelDefs / LvlTypes / LvlPrest / LvlMaze / LvlSub / LvlWarp data
        |
        v
level seed + DRLG type
        |
        +--> preset generator
        +--> maze generator
        `--> outdoor generator (+ Act/level-specific rules)
        |
        v
room graph + room coordinates + preset selection
        |
        v
selected DS1 files + substitutions + preset units
        |
        v
room tile grids
        |
        v
DT1 set selected by level type / masks
        |
        v
(type, style, sequence) lookup + rarity RNG
        |
        v
subtile collision + warps + linked room edges + animated tiles
        |
        v
renderable room/world representation
```

DS1 decoding is only one stage in this pipeline.

## Deterministic RNG

D2MOO reconstructs the 1.10f seed state as two 32-bit values:

```text
lowSeed  = supplied seed
highSeed = 666
next64   = highSeed + 0x6AC690C5 * lowSeed
(lowSeed, highSeed) become the low/high halves of next64
```

Limited random values use either modulo or a power-of-two mask.

At DRLG allocation, the supplied game/map seed initializes the DRLG seed and the runtime immediately rolls one value into `dwStartSeed`.

When a level is initialized, its seed is reinitialized from:

```text
levelSeed = levelId + drlg.dwStartSeed
```

This is a critical compatibility boundary. A Dark Magic deterministic generator should reproduce the original integer arithmetic and call order rather than substituting Go's `math/rand`.

## Level definition data

D2MOO's 1.10f data structures expose the fields that drive generation.

### Level definition / LevelDefs

Important fields include:

- size X/Y per difficulty;
- positional offsets and optional dependency level;
- DRLG type;
- level type;
- substitution type/theme/waypoint/shrine fields;
- up to eight visible destination levels and warp IDs;
- LOS/save-monster behavior.

The level's world position is based on its dependency's position plus its configured offsets. Size comes directly from the difficulty-specific definition.

### LvlTypes

A level type names up to 32 DT1 files. It is the primary logical tile-library family for rooms of that level type.

### LvlPrest

Important fields observed in 1.10f include:

- preset definition ID and level ID;
- populate/logicals/outdoors/animate/kill-edge/fill-blanks flags;
- animation speed;
- preset size;
- automap/scan/pop/poppad metadata;
- file count and up to six DS1 filenames;
- DT1 mask.

### LvlMaze

Each row maps a level ID to:

- desired room count per difficulty;
- room width/height;
- merge behavior.

### LvlSub

Each substitution row contains:

- substitution type;
- DS1 file;
- `CheckAll`;
- border behavior;
- DT1 mask;
- cluster/grid size;
- per-tier probability, trials, and maximum values;
- expansion flag.

### LvlWarp

Warp data includes selection rectangle, exit-walk offsets, render/placement offsets, lit-version behavior, tile count, and direction code.

## DRLG type dispatch

D2MOO's `DRLG_InitLevel` is explicit:

```text
seed level with levelId + startSeed
switch level.drlgType:
    MAZE    -> DRLGMAZE_GenerateLevel
    PRESET  -> DRLGPRESET_GenerateLevel
    OUTDOOR -> DRLGOUTDOORS_GenerateLevel
compute warp info
```

Dark Magic should preserve this conceptual split rather than building a single generic random-map algorithm.

## Preset levels

A preset level selects a `LvlPrest` record and one of its DS1 files. A preset can be built as a single room or subdivided into 8x8-tile rooms.

Room creation preserves the preset's DT1 mask and generated grid flags. Once a room is activated/constructed:

1. the DS1 is loaded if not already cached;
2. floor, wall, tile-type, and shadow grids are sliced to the room;
3. tile counts are determined;
4. room tile arrays are allocated;
5. packed DS1 cells are resolved through DT1 lookup;
6. animated tile grids are created if `LvlPrest.Animate` is enabled;
7. logical-coordinate lists are generated if `LvlPrest.Logicals` is enabled;
8. preset units that fall inside the room are transferred into room-local state.

`KillEdge` controls whether the final row/column of a preset contributes to a room at the preset boundary. `FillBlanks` can turn an otherwise blank floor cell into a hidden logical style-30 floor tile.

## Maze levels

Maze generation is a room-graph algorithm using fixed-size room rectangles from `LvlMaze`.

The generator:

- creates/places adjacent rooms while checking overlap;
- records orthogonal adjacency links;
- uses connection direction bits to select a matching `LvlPrest` chamber definition;
- contains level-type-specific and Act-specific remappings where not every directional chamber exists;
- forces special/quest rooms through dedicated placement logic;
- selects one of a preset definition's DS1 variants;
- then builds each resulting room through the normal preset/DS1 -> DT1 pipeline.

Connection topology therefore influences which DS1 preset is legal. Randomly choosing any DS1 of the same size is not compatible.

## Outdoor levels

Outdoor generation is not one simple noise/grid algorithm. D2MOO shows a shared outdoor grid plus many Act/level-specific generators and hard-coded required placements.

Shared behavior includes:

- creation of a coarse outdoor grid, commonly using 8-tile cells;
- marking level-link/border/path occupancy in packed outdoor-grid flags;
- placing mandatory and optional `LvlPrest` pieces into free grid regions;
- maintaining a selected DS1-file index for a preset area;
- creating paths between exits/links;
- applying substitutions after/beside structural placement;
- converting the resulting grid into room/preset areas.

Act-specific files such as `DrlgOutWild`, `DrlgOutDesr`, `DrlgOutJung`, `DrlgOutSiege`, and generic placement helpers encode materially different rules. Dark Magic should expose Act/level strategy hooks rather than hide these rules behind a generic `OutdoorGenerator` with random callbacks.

## Preset variant selection

`LvlPrest` can name multiple DS1 files. D2MOO uses deterministic level RNG to select file variants. In some outdoor paths it keeps per-preset build state and cycles the selected file after a seeded starting point, reducing immediate repetition.

Variant selection must therefore be part of the deterministic generation trace.

## DT1 file selection and masks

`LvlTypes` defines the candidate DT1 libraries for a level type. Preset and substitution records carry DT1 masks that determine which members of that family are needed by a room/piece.

A room should load only the tile-library set implied by its level type plus applicable masks/special requirements, then build a lookup keyed by tile type/style/sequence.

Do not interpret a DS1 dependency string as the complete authoritative runtime DT1 set. Dependencies are useful metadata, while DRLG data tables and masks control runtime composition.

## Tile variation selection

For a room cell, the original 1.10f runtime queries all loaded DT1 records matching:

```text
tileType
style/mainIndex
sequence/subIndex
```

It sums their DT1 rarity values and performs a room-seeded weighted choice. Selection is therefore deterministic only if the room seed and call order are compatible.

If no tile matches, 1.10f attempts a fallback left-exit tile `(type 10, style 0, sequence 0)`. Treat this as compatibility behavior, not a DT1 decoder rule.

## Blank/default and hidden tiles

D2MOO shows several important cases where "not visibly drawn" does not mean "ignore cell":

- hidden exit tiles can create warps;
- hidden door tiles can spawn door objects;
- style-30 floor cells can act as hidden/default logical floors;
- `FillBlanks` may synthesize a hidden style-30 floor;
- linkage cells on neighboring rooms can create shared tile/link records.

Map composition must preserve semantic cells before deciding what appears in a draw list.

## Wall orientation and room-edge linking

At shared room borders the 1.10f runtime can find a previously created tile from the neighboring room and merge/remap wall orientation according to both rooms' tile types and layer flags.

Consequences:

- room rendering cannot be finalized independently before adjacency/link processing;
- upper/corner wall forms can produce multiple tile records at one cell;
- hidden/layer-above flags alter which border tile remains visible;
- collision may need updating if a linked wall changes type/visibility.

This is a composition rule above the DS1 decoder.

## Shadows, roofs, and upper walls

The DS1 shadow layer is instantiated as the runtime's roof/shadow tile array. COF-style dynamic shadows are unrelated.

Roof/upper-wall behavior also uses packed-cell hidden/reveal/layer flags. Preserve tile type and raw cell metadata in the map representation instead of flattening everything into one ordered sprite list during load.

## Collision

Every selected DT1 tile exposes a 5x5 subtile flag grid. Room/map collision is assembled at subtile precision after tile selection.

A correct map representation needs at least:

```text
world tile coordinate
5x5 subtile collision bytes
map-tile semantic flags
resolved DT1 record identity
room ownership / links
```

The collision layer must not need decoded RGBA pixels.

## Object and monster placement

DS1 preset units are map-authored placement directives. During preset loading the original runtime resolves version/Act-dependent preset IDs, converts positions to room/world subtiles, and conditionally suppresses a few special units based on RNG/quest-specific rules.

Separate later population systems use `Levels` monster/object-group metadata. Do not treat every monster in a generated level as a DS1 object.

## Warps

Warp creation combines:

- `Levels` visibility/warp arrays;
- DS1 exit/special cells;
- `LvlWarp` data;
- room adjacency and world placement.

The runtime records representative warp/waypoint room centers in subtile coordinates after level generation. Exit cells can also create hidden linked floor tiles for lit-version warp behavior.

## Automap implications

Automap data uses level type + tile type/style/sequence ranges. Because hidden, linkage, exit, and roof semantics influence automap visibility, automap generation should consume semantic map tiles rather than screenshots or flattened render batches.

## Coordinate systems

D2MOO documents the core precision hierarchy:

```text
1 room grouping unit ~= 40 subtiles
1 floor tile = 5 subtiles
1 subtile/game coordinate = base collision unit
1 fractional unit = 1/65536 subtile
```

Projection at equal precision is:

```text
clientX = (gameX - gameY) / 2
clientY = (gameX + gameY) / 4
```

At tile precision this yields 160x80 pixels; at subtile precision, 32x16 pixels.

Use strongly named coordinate types in Dark Magic so tile X/Y cannot accidentally be passed to subtile/world APIs.

## Worked example: level ID + seed to rendered room

The exact level/file values depend on owned data, but the compatibility trace should look like this:

```text
input: levelId=L, gameSeed=S, difficulty=D

1. DRLG seed <- InitLowSeed(S)
2. startSeed <- Roll(DRLG seed)
3. load LevelDef[L]
4. level position/size <- LevelDef dependency, offsets, SizeX[D], SizeY[D]
5. level RNG <- InitLowSeed(L + startSeed)
6. dispatch by LevelDef.DrlgType
7. generator creates room graph / selects LvlPrest IDs
8. selected LvlPrest chooses a DS1 file using level RNG
9. load DS1 and slice its grids into room-local grids
10. derive DT1 set from LevelType + LvlPrest.Dt1Mask
11. for each semantic tile cell:
      resolve type/style/sequence
      choose matching DT1 variant using room RNG + rarity
      attach 5x5 subtile flags
12. process room-edge linked tiles, hidden exits/doors, warps, animation grids
13. convert room tile coords to world tile/subtile coords
14. renderer projects semantic tile instances to client pixels and draws by map ordering rules
```

A deterministic debug trace should be able to print every numbered decision without dumping Blizzard-owned pixels.

## Implementation boundary for Dark Magic

Recommended independent architecture:

```text
data records -> GenerationContext(seed, difficulty, tables)
             -> LevelPlan (type, bounds, links)
             -> RoomGraph
             -> PresetInstance(s)
             -> SemanticTileGrid
             -> CollisionGrid + RenderTileInstances + PresetUnits + Warps
```

The raw `ds1` and `dt1` modules should not import DRLG or rendering packages.

## Unresolved work

- exact room-seed derivation/call-order audit for every DRLG subtype;
- full required-preset rules per level/quest;
- every outdoor packed-grid flag meaning;
- full `LvlSub` probability/trial/max algorithm mapping;
- automap population rules from generated tiles;
- version differences after 1.10f, especially 1.13d compatibility.

These should be validated with deterministic probes and D2MOO traces before Dark Magic replaces large map subsystems.

libd2 adds a particularly useful comparison oracle because it reports
cell-for-cell checks against retail-engine captures across all five acts and
keeps blind holdout seeds. Dark Magic should use its verification structure and
observable outputs to design independent tests, not copy embedded fixtures or
assume that every 1.14d behavior is identical to D2MOO's 1.10f reconstruction.
See [`research/libd2.md`](research/libd2.md) for the source boundary and audit
queue.
