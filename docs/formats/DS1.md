# DS1 preset map format

Status: research specification for the preset-map container. DS1 rendering is not the same subsystem as DRLG level generation; see `../MAP_GENERATION.md`.

## Core model

A DS1 stores a rectangular preset in game-tile coordinates, using `width + 1` by `height + 1` cells for its tile grids. It can carry wall cells, wall orientation/type cells, floor cells, a shadow/roof-like cell layer, substitution tags, preset units, substitution groups, and unit paths.

A DS1 tile cell stores logical tile identity and map flags; the matching DT1 graphics record is resolved later.

## Version matrix

The original runtime parser in D2MOO and Riiablo's independent decoder agree on these version gates through version 18.

| Version | Added / changed behavior |
| ---: | --- |
| 1 | base width/height and legacy fixed layers |
| 2 | preset-unit/object section present |
| 3 | dependency/reference string table present |
| 4 | explicit wall-layer count; modern interleaved wall + orientation layout; floor count still fixed to 1 |
| 5 | newer monster/item preset interpretation in original runtime |
| 6 | preset unit spawn/flags word present |
| 7 | orientation values are direct tile types; versions `<7` require legacy orientation-table conversion |
| 8 | Act field present; clamp to Act V for compatibility |
| 9 | two unknown 32-bit words present through version 13 |
| 10 | substitution method field present; substitution/tag layer conditionally present |
| 11 | no new structural field identified in inspected runtime parser |
| 12 | substitution-group list present when substitution is enabled |
| 13 | substitution group gains one extra 32-bit field |
| 14 | path section present |
| 15 | each path node gains an action word; earlier nodes default to action 1 |
| 16 | explicit floor-layer count present; up to two floor layers are expected by mature implementations |
| 17 | no new structural field identified in inspected runtime parser |
| 18 | one extra 32-bit value appears before the substitution-group count |

Supported assets observed by Riiablo are versions 1 through 18. Treat values outside this range as unsupported unless a probe demonstrates a later legitimate Diablo II asset.

## Header

```text
u32 version
u32 width
u32 height
[if version >= 8]  u32 act
[if version >= 10] u32 substitutionMethod
[if version >= 3]  u32 dependencyCount; zero-terminated strings...
[if 9 <= version < 14] 8 unknown bytes
layer-count fields and layers...
```

The stored width/height are not the number of serialized cells per row/column. Tile grids contain `(width + 1) * (height + 1)` cells.

Object positions are in subtiles, not DS1 tile-cell indices. The normal preset bounds therefore span approximately `width * 5` by `height * 5` subtiles for object-placement validation.

## Layer counts and ordering

### Versions below 4

The parser assumes one wall, one floor, one orientation, one tag/substitution grid, then one shadow grid. Historical ordering is unusual and must not be generalized to modern DS1 files.

### Version 4 and later

```text
u32 wallLayerCount
[if version >= 16] u32 floorLayerCount
for each wall layer:
  wall cell grid
  orientation/type grid
for each floor layer:
  floor cell grid
shadow cell grid
[if substitution enabled] substitution tag grid
```

Mature implementations bound the normal asset model at 4 wall layers, 2 floor layers, 1 shadow layer, and 1 substitution/tag layer. These are observed Diablo II conventions, not a reason for unsafe fixed-size parsing: validate counts before allocation and retain an explicit compatibility limit.

## Packed tile cell

D2MOO exposes the 1.10f runtime interpretation of the 32-bit tile-information word:

```text
bit  0      is wall
bit  1      is floor
bit  2      LOS/linking flag generated at room edges
bit  3      enclosed / enclosure-like semantic
bit  4      exit
bit  5      unknown
bit  6      layer below (floor)
bit  7      layer above
bits 8-15   tile sequence / DT1 sub-index
bit 16      fill LOS
bit 17      unwalkable
bits 18-19  wall layer
bits 20-25  tile style / DT1 main index
bit 26      reveal hidden
bit 27      shadow/roof-like flag
bit 28      linkage (waypoint/level/path association)
bit 29      object wall
bit 30      unknown
bit 31      hidden
```

Important: these flags contain both authored DS1 information and code-generated room-link information in the original runtime. Dark Magic should distinguish raw serialized cells from derived runtime tile state.

## Orientation / tile type

For DS1 versions 7 and later, orientation/type cells directly identify the tile type used for DT1 lookup. For versions below 7, the numeric orientation value is mapped through a compatibility table before tile lookup.

The 1.10f runtime tile-type enum includes:

```text
0  floor
1  left wall
2  right wall
3  upper/top right corner wall
4  upper/top left corner wall
5  upper/top right wall
6  bottom left wall
7  bottom right wall
8  left door wall
9  right door wall
10 left exit wall
11 right exit wall
12 column
13 shadow
14 tree
15 roof
16 left wall down
17 right wall down
18 full wall down
19 front wall down
```

Do not infer wall rendering solely from the packed cell's `isWall` bit; the orientation/type layer decides which DT1 orientation is requested.

## DT1 selection

For each logical tile to instantiate, the runtime extracts:

```text
style    = bits 20..25
sequence = bits 8..15
type     = orientation layer (or floor/shadow implicit type)
```

The room's loaded DT1 set is queried for records matching `(type, style, sequence)`, then rarity-weighted selection chooses among alternates. See `DT1.md`.

## Hidden and logical tiles

`hidden` does not mean "ignore the record." Hidden exit/door cells can create warps or preset units and then skip normal tile emission. D2MOO also special-cases floor style 30 sequences 0/1 as hidden/default-fill logical floors.

The original runtime uses linkage, layer-above, hidden, reveal-hidden, fill-LOS, and wall-layer bits when merging border cells between neighboring rooms. A clean-room engine should model this as a room composition/linking pass, not encode those rules into the DS1 file decoder.

## Objects / preset units

For version 2 and later:

```text
u32 count
repeat count:
  u32 type
  u32 id
  u32 x
  u32 y
  [if version >= 6] u32 flags/spawn flag
```

Interpretation of `id` is version- and unit-type-dependent in the original runtime. Monster preset IDs can index Act-specific preset tables; object IDs in later versions may be remapped through object-preset tables. Therefore the binary decoder should expose raw fields and a higher-level compatibility resolver should perform game-data lookup.

Coordinates are subtile coordinates local to the preset.

## Substitution groups

If `version >= 12` and substitution method is enabled:

```text
[if version >= 18] u32 unknown
u32 groupCount
repeat groupCount:
  u32 x
  u32 y
  u32 width
  u32 height
  [if version >= 13] u32 unknown/group metadata
```

The tag grid and these rectangles support DS1-authored substitution selection. They are distinct from `LvlSub.txt`, which also uses DS1-like substitution files as part of outdoor/procedural generation.

## Paths

If `version >= 14`:

```text
u32 pathCount
repeat pathCount:
  u32 nodeCount
  u32 objectX
  u32 objectY
  repeat nodeCount:
    u32 x
    u32 y
    [if version >= 15] u32 action
```

The original runtime associates a path with the preset unit at the matching `(objectX, objectY)` position. If no unit exists at that position it skips the node payload.

## Coordinate spaces

- DS1 tile grids: game-tile precision.
- One tile: 5x5 subtiles.
- Objects and path coordinates: subtile precision.
- World room origin is added by the DRLG/preset builder.
- The final screen projection is dimetric; see `DT1.md` and `../MAP_GENERATION.md`.

Do not store all of these as untyped integer X/Y values in engine APIs.

## Compatibility behavior observed in 1.10f

- Hidden door cells may spawn door objects without emitting a visible map tile.
- Hidden exit cells create/link warp structures and associated floor warp tiles.
- Neighboring-room edges can remap wall orientation and visibility when both rooms contribute a tile at the same border position.
- `FillBlanks` in `LvlPrest.txt` can cause blank preset cells to instantiate a logical style-30 floor rather than remain empty.
- Special map semantics depend on style/sequence pairs in a small number of levels, so preserving packed cell values losslessly is mandatory.

## Malformed-input conditions

Reject or safely bound:

- unsupported version;
- width/height arithmetic overflow when adding one and multiplying;
- layer counts beyond configured compatibility maxima;
- layer payload shorter than the derived cell count;
- unterminated dependency strings;
- object/path/group counts that exceed remaining bytes or sane allocation limits;
- invalid path-node counts;
- orientation-table index outside the legacy conversion table for versions `<7`.

Unknown flag bits and unknown version-gated words must be retained where practical, not rejected.

## Probe requirements

Owned-asset probes should print version, dimensions, dependencies, layer counts, all orientation values, packed-cell bit distributions, object types/IDs, group metadata, and path actions. See `CODEC_FOLLOWUPS.md`.
