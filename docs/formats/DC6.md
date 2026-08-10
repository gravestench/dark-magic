# DC6 sprite format

Status: research specification. This document separates binary-format behavior from historical renderer limits.

## Global layout

The independent Riiablo and OpenDiablo2 decoders agree on the structural shape:

```text
u32 version
u32 flags
u32 encoding/format
byte termination[4]
u32 directions
u32 framesPerDirection
u32 frameOffsets[directions * framesPerDirection]
frame records...
```

The four header termination bytes are not the same thing as the per-frame trailer bytes used by some decoders.

## Frame record

Each frame begins with eight 32-bit words:

| Field | Signed | Meaning |
| --- | --- | --- |
| flipped | no | vertical-placement / orientation flag |
| width | no | frame width |
| height | no | frame height |
| xOffset | yes | anchor-relative X |
| yOffset | yes | anchor-relative Y |
| unknown | no | unresolved |
| nextBlock | no | next-frame/file pointer metadata |
| length | no | encoded RLE payload length |

After the payload, OpenDiablo2 consumes three trailer bytes. Riiablo instead slices each frame by the global offset table and therefore does not need to trust that trailer to locate the next frame. Dark Magic should retain/validate both forms of evidence and prefer the global frame-offset table for bounds safety.

## Direction/frame indexing

Logical frame index is:

```text
index = direction * framesPerDirection + frame
```

The offset table has one entry per frame. A decoder may synthesize a sentinel offset equal to file length for safe slicing of the last frame.

## RLE stream

DC6 scanlines are encoded bottom-to-top.

The command byte rules corroborated by Riiablo and OpenDiablo2 are:

```text
0x80           end of scanline
0x81..0xFF     transparent run of (command & 0x7f) pixels
0x00..0x7F     opaque run; copy `command` literal palette-index bytes
```

Decoding state begins at:

```text
x = 0
y = height - 1
```

On `0x80`, reset `x = 0` and decrement `y`.

Transparent pixels are not explicitly stored. The destination image should be initialized to palette index 0 before decoding.

### Encoder implication

A transparent run cannot encode more than 127 pixels in one command. An encoder must split longer transparent spans into legal chunks. This is a format-level command-width rule, not an image-width restriction.

## Vertical inversion and offsets

Because the encoded rows are traversed from the bottom upward, a decoder that writes into a conventional top-down image buffer performs an effective vertical inversion while decoding.

The `flipped` field affects anchor-relative bounds. Riiablo normalizes a frame box as:

```text
xMin = xOffset
if flipped:
    yMin = yOffset
else:
    yMin = yOffset - height
```

with width/height added to derive the maximum corner. The exact historical semantic name of `flipped` should remain separate from this observed placement rule.

## Palette behavior

DC6 stores palette indices, not RGBA colors. Palette choice comes from the resource context.

Dark Magic should preserve indexed pixels as long as practical so palette changes, PL2 transforms, and diagnostic rendering do not require re-decoding compressed frames.

Palette index zero is normally treated as transparent by Diablo II presentation code, but that convention belongs to rendering semantics; the decoder should preserve literal index data.

## Multi-page UI assets

Many UI DC6 files use multiple 256x256-ish frames as pages of a larger logical image. Riiablo contains presentation logic that detects a frame grid and combines pages at runtime.

This is not a second DC6 binary format. It is asset-level composition: directions and frames remain normal DC6 frames, and the caller decides whether a set of them represents animation frames or tiles/pages of one UI surface.

Dark Magic should expose frame metadata and leave page composition to UI/resource metadata or an asset helper.

## File-format limits vs original renderer limits

Historical Diablo II rendering code and reimplementations frequently expose a 256-pixel page size. That does **not** prove the binary format requires `width <= 256` or `height <= 256`.

The decoder contract should instead be:

- dimensions must fit configured memory/safety limits;
- RLE writes must stay within the declared dimensions;
- offsets must remain representable;
- frame offsets/lengths must stay within the file.

Any `256x256` behavior required for original UI page composition should live in the compatibility/presentation layer, not as a decoder rejection rule.

## Bounding boxes

For each decoded frame retain:

```text
width
height
xOffset
yOffset
flipped
anchor-relative bounding box
```

For an animation direction, the union of all frame boxes is useful for stable layout but should not replace per-frame boxes.

## Malformed-input handling

Reject or safely stop on:

- direction/frame multiplication overflow;
- offset table outside the file;
- non-monotonic frame offsets where slicing would overlap backward;
- frame header past its slice;
- declared payload length larger than the frame slice;
- opaque run larger than the remaining scanline;
- transparent run larger than the remaining scanline;
- extra scanline terminators after all rows are consumed;
- payload ending before all required rows terminate;
- dimensions exceeding configured allocation limits.

Do not reject a frame merely because it is larger than the original renderer's usual page size.

## Evidence record

| Fact | Source | Version | Confidence | Notes |
| --- | --- | --- | --- | --- |
| header has version/flags/encoding/4-byte marker/dirs/frames/offset table | Riiablo `Dc6.java`, OpenDiablo2 `d2dc6/dc6.go` | pinned/current | verified | independent parsers agree |
| frame header has flip,width,height,x/y,unknown,next,length | same | pinned/current | verified | x/y signed |
| scanlines decode bottom-to-top | Riiablo `Dc6Decoder.java`, OpenDiablo2 `DecodeFrame` | pinned/current | verified | both initialize y = height-1 |
| `0x80` is end-of-line and high-bit commands are transparent runs | same | pinned/current | verified | transparent count is low 7 bits |
| OpenDiablo2 reads three trailer bytes after frame data | OpenDiablo2 | current | medium | needs owned-asset probe to catalog values |
| 256-page behavior is presentation convention, not proven format maximum | Riiablo page compositor + historical reports | pinned/historical | high | decoder should not inherit limit |

## Probe requirements

See `CODEC_FOLLOWUPS.md`. Capture header marker bytes, frame trailer bytes, all dimension ranges, `flipped` distributions, `nextBlock` consistency with global offsets, and examples of UI page-composed DC6 files.
