# DT1 tile library format

Status: research specification. No Dark Magic implementation changes are implied by this document.

## Summary

A DT1 is a palette-indexed tile library. Its important contract is not just pixels: each tile carries a logical identity `(orientation/type, main/style index, sub/sequence index)`, rarity or animation-frame metadata, material metadata, 25 subtile collision bytes, geometry, and one or more encoded pixel blocks.

Dark Magic should keep the binary decoder responsible for lossless file interpretation and expose the semantic metadata required by the map layer. Tile selection and room RNG belong in the engine/map layer.

## Evidence snapshot

| Fact | Source | Version | Confidence | Notes |
| --- | --- | --- | --- | --- |
| Original runtime indexes tiles by type/style/sequence | D2MOO `D2CMP.h`, `DrlgRoomTile.cpp` | 1.10f RE | verified | D2CMP hash node stores these three keys; DRLG requests this tuple. |
| Alternate matching tiles are rarity weighted | D2MOO `DRLGROOMTILE_GetTileCache` | 1.10f RE | verified | room RNG chooses from matching entries using sum of tile rarity. |
| Tile geometry is based on 160x80 floor-tile projection and 5x5 subtiles | D2MOO `Coordinates.md`; Riiablo `Tile.java` | 1.10f / pinned commit | verified | one subtile projects to 32x16 pixels. |
| DT1 tile header carries 25 collision bytes | Riiablo `Dt1Decoder7.java`; D2MOO `D2CMP.h` | pinned / 1.10f RE | verified | D2MOO runtime representation compacts/normalizes these internally. |

## File layout

Observed modern Diablo II DT1 libraries use this high-level layout:

```text
library header
all tile headers
for each tile:
  block headers
  block pixel payloads
```

Riiablo's decoder identifies the first 32-bit value as `0x00000007`, uses a total library header size of `0x114`, tile header size `0x60`, and block header size `0x14`.

Historical tools commonly describe the file as "7.6". Treat the two leading 32-bit words independently until owned-asset probes establish the exact semantics of the second word across asset families.

An owned-asset probe found two Act I Barracks files whose leading words are
`4,1`: `barracks.dt1` and `gargtrap.dt1`. They do **not** use
the modern `0x114` header layout. Bytes following the signature contain compact
counts and pointer tables, so accepting `4,1` and continuing with modern offsets
would be unsafe. They require a distinct, researched legacy decoder before they
can be inspected as tile libraries. The other nine observed Barracks DT1 files
use the modern `7,6` signature.

Barracks DS1 stamps retain historical `.tg1` dependency names. The mounted game
assets provide corresponding modern `.dt1` libraries. A probe of ordinary,
Horadric Malus forge, and secret-room stamps showed that their declarations
resolve to the modern `7,6` files such as `basewall.dt1`, `floor.dt1`, and
`objects.dt1`; they do not declare every DT1 in the neighboring directory.
Consumers must therefore resolve the DS1 declaration table instead of loading
the entire directory. This distinction lets valid quest stamps render while the
unrelated `4,1` files remain explicitly unsupported and available for research.

### Library header

| Offset | Size | Meaning | Confidence |
| --- | ---: | --- | --- |
| `0x000` | 4 | first version word, normally 7 | verified |
| `0x004` | 4 | second header word / flags/version-like value | unresolved naming |
| `0x008` | `0x104` | reserved/library-name area in historical runtime representation | high |
| `0x10C` | 4 | tile count | verified |
| `0x110` | 4 | offset to first tile header | verified |

Do not require unused bytes to be zero unless asset probes show that is a valid malformed-input rule.

## Tile header (`0x60` bytes)

The observed fields are:

| Offset | Size | Signed | Meaning |
| --- | ---: | --- | --- |
| `0x00` | 4 | no | light direction |
| `0x04` | 2 | no | roof height |
| `0x06` | 2 | no | material flags |
| `0x08` | 4 | yes | total height |
| `0x0C` | 4 | yes | width |
| `0x10` | 4 | yes | height-to-bottom / vertical extent metadata |
| `0x14` | 4 | no | orientation / tile type |
| `0x18` | 4 | no | main index / style |
| `0x1C` | 4 | no | sub-index / sequence |
| `0x20` | 4 | no | rarity, or frame index for animated material |
| `0x24` | 4 | no | transparent-color RGB24 field |
| `0x28` | 25 | n/a | 5x5 subtile flags, stored in a file-coordinate order that implementations normalize |
| `0x41` | 7 | n/a | padding/reserved |
| `0x48` | 4 | no | block section offset |
| `0x4C` | 4 | no | total block section length |
| `0x50` | 4 | no | number of blocks |
| `0x54` | 4 | n/a | runtime pointer placeholder / unused on disk |
| `0x58` | 2 | unknown | unresolved |
| `0x5A` | 2 | no | cache index |
| `0x5C` | 4 | unknown | unresolved |

### Material flags

D2MOO identifies the following 16-bit values from the original runtime:

```text
0x0001 other
0x0002 water
0x0004 wood object / blocks reverb
0x0008 indoor stone
0x0010 outdoor stone
0x0020 dirt
0x0040 sand
0x0080 wood
0x0100 lava (special bright + animated behavior)
0x0400 snow
```

Unknown bits must be preserved.

### Subtile flags

Each tile has 25 bytes for its 5x5 collision/light grid. Riiablo names the currently understood bits:

```text
bit 0 block walk
bit 1 block line of sight
bit 2 block jump
bit 3 block player walk
bit 4 unknown
bit 5 block light
bit 6 unknown
bit 7 unknown
```

Do not reduce this to one boolean walkability value. Collision assembly needs the original byte per subtile.

## Block header (`0x14` bytes)

| Offset | Size | Meaning |
| --- | ---: | --- |
| `0x00` | 2 | signed pixel X |
| `0x02` | 2 | signed pixel Y |
| `0x04` | 2 | unknown |
| `0x06` | 1 | grid X |
| `0x07` | 1 | grid Y |
| `0x08` | 2 | format |
| `0x0A` | 4 | encoded payload length |
| `0x0E` | 2 | unknown |
| `0x10` | 4 | payload offset |

Observed format identifiers in the pinned Riiablo decoder are:

- `0x0001`: fixed isometric block, exactly 256 palette indices.
- `0x1001`: RLE block.
- `0x2005`: isometric-RLE variant consumed by the same RLE command stream.

Historical code that collapses these values into `0`/`1` enums is an implementation simplification and must not be treated as the file-format value table.

## Pixel decoding

### Fixed isometric block

A fixed block expands to a 32x16 diamond. The 15 scanline runs are:

```text
skip: 14,12,10,8,6,4,2,0,2,4,6,8,10,12,14
run:   4, 8,12,16,20,24,28,32,28,24,20,16,12, 8, 4
```

The payload is 256 palette indices. The decoder copies each run into the destination tile image at the block's pixel origin.

### RLE block

Commands are byte pairs `(skip, runLength)` followed by `runLength` palette indices.

- `(0,0)` terminates the current scanline and resets X to the block origin.
- otherwise advance X by `skip`, copy `runLength` opaque palette indices, then advance X by the copied length.

Transparent pixels are implicit and remain palette index zero in the normalized output.

## Bounding boxes and vertical placement

Block positions may extend outside the nominal 160x80 diamond, especially for walls/roofs. The tile decoder must calculate a bounding box over its block rectangles rather than forcing every tile into 160x80.

For fixed isometric blocks, the block contributes a 32x16 diamond. RLE blocks use a 32x32 block extent. Riiablo applies an additional negative 80-pixel Y origin adjustment for floor and roof tile classes when normalizing the decoded image.

Treat this as coordinate normalization, not as destructive rewriting of the encoded block coordinates.

## Logical identity and DS1 relationship

A DS1 cell does not point at a physical DT1 record number. It supplies:

```text
tile type/orientation
main/style index
sub/sequence index
```

The loaded DT1 libraries for the room are searched for all records matching that tuple.

D2MOO's 1.10f runtime uses a weighted random choice among all matches:

```text
matches = lookup(type, style, sequence)
total = sum(match.rarity)
r = roomRng(0..total-1) + 1
walk matches subtracting rarity until r <= 0
```

If all rarity values sum to zero, the first match is used. If no matching tuple exists, the 1.10f runtime attempts a fallback lookup for left-wall-exit type, style 0, sequence 0. Dark Magic should reproduce this only in the compatibility map layer; the DT1 decoder itself should simply expose the records.

## Coordinates

Diablo's game grid is dimetric:

```text
clientX = (gameX - gameY) / 2
clientY = (gameX + gameY) / 4
```

At floor-tile precision:

- 1 tile = 5x5 subtiles.
- 1 tile projects to 160x80 pixels.
- 1 subtile projects to 32x16 pixels.

Engine APIs should name the precision explicitly (`Tile`, `Subtile`, `ClientPixel`) to avoid unit mistakes.

## Malformed-input checks

The codec should reject or safely bound:

- impossible tile counts or allocation sizes;
- tile/block sections outside the file;
- block count whose headers overflow the section;
- payload lengths exceeding the declared block section/file;
- fixed ISO payload lengths other than 256 unless verified assets prove another legal case;
- RLE commands that write beyond the computed destination bounds;
- integer overflow while deriving image bounds.

Unknown flags and reserved bytes are not malformed merely because Dark Magic does not interpret them.

## Required probes

See `CODEC_FOLLOWUPS.md`. Highest priority: capture both DT1 header words, raw 25-byte subtile arrays, all observed block format values, rarity distributions, and tile bounding boxes across owned Act I-V libraries.
