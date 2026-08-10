# DCC compressed animation format

Status: research specification. DCC is direction-local delta compression over palette-indexed frames. A working decoder is not sufficient documentation: the important contract is how its bitstreams reconstruct cells and frames over time.

## File header

The pinned Riiablo parser reads:

```text
u8  signature
u8  version
u8  directionCount
u32 framesPerDirection
u32 tag
u32 uncompressedSize
u32 directionOffsets[directionCount]
```

A safe decoder can append a synthetic sentinel direction offset equal to file length.

Each direction is independently compressed. Frame-to-frame reuse exists within a direction; do not carry delta state from one direction into another.

## Encoded integer width table

DCC frame fields use a 4-bit selector into this bit-width table:

```text
selector:  0  1  2  3  4  5   6   7   8   9  10  11  12  13  14  15
bits:      0  1  2  4  6  8  10  12  14  16  20  24  26  28  30  32
```

X and Y offsets are signed values read using the selected width. Width, height, extra-byte count, compressed-byte count, and the unresolved `variable0` field are unsigned.

## Direction header

Within each direction:

```text
u32 uncompressedSize
bits[2] compressionFlags
bits[4] variable0Bits selector
bits[4] widthBits selector
bits[4] heightBits selector
bits[4] xOffsetBits selector
bits[4] yOffsetBits selector
bits[4] extraBytesBits selector
bits[4] compressedBytesBits selector
frame headers...
[byte align]
optional per-frame extra data...
stream size words...
palette-value presence table...
bitstreams...
```

Known compression flags from Riiablo:

```text
bit 0 HasRawPixelEncoding
bit 1 CompressEqualCells
```

Unknown flag combinations should be rejected only when they make stream interpretation impossible.

## Frame header

For every frame, consume the selected-width values in this order:

```text
variable0
width
height
xOffset (signed)
yOffset (signed)
extraBytes
compressedBytes
flipY (1 bit)
```

The anchor-relative frame box is normalized similarly to DC6:

```text
xMin = xOffset
if flipY:
    yMin = yOffset
else:
    yMin = yOffset - height
```

The direction bounding box is the union of its frame boxes.

`variable0`, `extraBytes`, and the semantic purpose of `compressedBytes` beyond stream accounting need owned-asset probes before Dark Magic assigns stronger names.

## Direction stream ordering

After all frame headers and optional extra-data blocks, the direction contains sizes for optional/required streams:

```text
[if CompressEqualCells] 20-bit equalCellBitStreamSize
20-bit pixelMaskBitStreamSize
[if HasRawPixelEncoding] 20-bit encodingTypeBitStreamSize
[if HasRawPixelEncoding] 20-bit rawPixelCodesBitStreamSize
256 one-bit palette-value-presence flags

Equal-cell bitstream
Pixel-mask bitstream
Encoding-type bitstream
Raw-pixel-code bitstream
Pixel-code/displacement bitstream (remainder of direction)
```

The 256 presence bits build a compact table mapping DCC pixel codes to actual palette indices. Presence bits are scanned in palette-index order; each set bit appends that palette index to the table.

## Why multiple streams exist

DCC separates structural decisions from literal pixel-index data:

- **equal-cell stream**: says a previously touched direction cell is unchanged and can reuse earlier state.
- **pixel-mask stream**: identifies which entries of a cell's four-entry pixel-code stack are updated.
- **encoding-type stream**: chooses literal/raw pixel codes vs displacement-coded values for a changed cell.
- **raw-pixel-code stream**: supplies 8-bit DCC pixel codes for raw-coded cells.
- **pixel-code/displacement stream**: supplies delta-coded pixel-code values and later the per-pixel selector bits used when rasterizing a cell.

Keeping these streams as separate bounded readers makes malformed-input diagnosis much easier.

## Cell model

The direction's union bounding box is partitioned into nominal 4x4 cells. Edge cells can be 1-4 pixels wide/high because neither the direction box nor individual frame boxes must align to multiples of four.

A frame uses only the cells intersecting its own bounding box. Its cell origin within the direction grid is derived from the frame box relative to the direction box.

The decoder maintains direction-level prior-cell state across frames.

## First pass: decode cell definitions

For each frame cell in row-major order:

1. If the corresponding direction cell has never been touched, force pixel mask `0xF` because all stack entries must be initialized.
2. Otherwise, if equal-cell compression exists, consume one equal-cell bit.
3. If equal-cell is true, this frame carries no new definition for the cell; reconstruction later reuses the previous cell image/state.
4. If not equal, consume a 4-bit pixel mask.
5. The mask determines how many pixel-stack entries are updated.
6. If at least one entry changes and raw encoding exists, consume one encoding-type bit.
7. Decode changed pixel codes either literally from the raw stream or as displacement-coded values from the displacement stream.
8. Store the new stack definition in a queue associated with this direction cell/frame.

### Displacement-coded pixel values

For a changed stack entry, begin from the last decoded pixel code. Read 4-bit displacement chunks and add them. A displacement of `0xF` means continue with another nibble. Stop when a nibble smaller than 15 is read.

A decoded code equal to the previous code terminates the stack early.

After this first pass, map every DCC pixel code through the direction's palette-value table to obtain real palette indices.

## Second pass: reconstruct raster frames

Maintain a previous-frame bitmap covering the whole direction box.

For each frame:

1. swap/advance previous and current bitmap buffers;
2. iterate that frame's cells in row-major order;
3. if there was no new cell definition:
   - if current and previous cell dimensions differ, clear the cell;
   - otherwise copy the prior cell image from its previous position;
4. if there is a new cell definition:
   - a one-color stack fills the whole cell;
   - a two-color stack consumes one selector bit per destination pixel;
   - a 3/4-color stack consumes two selector bits per destination pixel;
5. record this cell's position/dimensions as the new prior state.

The remaining selector bits come from the same pixel-code/displacement stream after its stack-definition nibbles have been consumed.

This explains why decoding DCC cannot be implemented as independent frame decompression: the stream and cell state are deliberately shared across frames in one direction.

## Transparency

Palette index zero is normally transparent in the presentation pipeline. A cleared cell therefore becomes zero-filled indexed pixels.

Do not bake RGBA alpha into the DCC decoder. Preserve the palette-index image and anchor-relative bounds.

## Direction independence

Reset all of the following at each direction boundary:

- direction bounding box and 4x4 cell grid;
- touched/prior cell state;
- prior bitmap;
- all stream readers;
- palette-value table.

No delta state should flow from one direction to another.

## Decoder invariants

A strict debug/probe decoder should verify that after a direction is reconstructed:

- equal-cell stream is fully consumed when present;
- pixel-mask stream is fully consumed;
- encoding-type stream is fully consumed when present;
- raw-pixel-code stream is fully consumed when present;
- pixel-code/displacement stream has no unexpected remainder;
- no cell write exceeded the direction bitmap;
- no frame cell count overflowed configured limits.

## Historical size limits

Do not codify historical fixed-buffer assumptions as format maxima. Mature tooling has documented valid DCC direction boxes wider/taller than old folklore limits, and historical tool bugs included streams larger than 16 KiB and very short missile frames.

Dark Magic's codec limits should be memory/safety limits, not original software-buffer dimensions.

## Decode diagram

```text
file
 └─ direction offset
     ├─ direction header + field-width selectors
     ├─ frame headers ───────────────┐
     ├─ optional frame extra bytes   │ establish frame boxes
     ├─ stream sizes                 │
     ├─ 256-bit palette presence     │
     ├─ equal-cell stream ───────────┤
     ├─ pixel-mask stream ───────────┤
     ├─ encoding-type stream ────────┤ first pass: cell definitions
     ├─ raw-code stream ─────────────┤
     └─ displacement/selector stream ┘
                    │
                    v
          direction cell-state queue
                    │
                    v
          second pass: raster frames
                    │
                    v
       palette-indexed frames + boxes
```

## Evidence record

| Fact | Source | Version | Confidence |
| --- | --- | --- | --- |
| header and direction offsets | Riiablo `Dcc.java` | pinned commit | high |
| encoded-width selector table | Riiablo `Dcc.java` | pinned commit | high |
| stream order and conditional stream sizes | Riiablo `Dcc.java` | pinned commit | high |
| 4x4 direction cells, edge cells smaller | Riiablo `DccDecoder.java` | pinned commit | high |
| equal-cell reuse and pixel masks | Riiablo `DccDecoder.java` | pinned commit | high |
| frame-to-frame cell copy/clear behavior | Riiablo `DccDecoder.java` | pinned commit | high |
| streams >16 KiB and short missile-frame edge cases exist | Paul/SixDice research ledger already in Dark Magic | historical | medium-high; probe requested |

## Probe requirements

See `CODEC_FOLLOWUPS.md`. DCC needs the most aggressive trace tooling: dump field selectors, direction/frame boxes, stream sizes, palette tables, per-cell masks/equal flags, stack values, and final stream-consumption counts for representative owned character, monster, object, and missile assets.
