# Legacy codec follow-up work

This is the handoff for continuing work in Dark Magic's independent codec repositories.
Keep codecs headless and use user-supplied Diablo II assets only for optional
probes and smoke tests.

The general streaming-I/O boundary was completed on 2026-08-10: sequential
formats accept readers/writers, offset formats expose lazy `ReaderAt` views,
and MPQ entry reads are positional and concurrent. The format-semantic,
inspection, and encoder tasks below remain intentionally open.

## General rules for all codec repos

1. Preserve raw fields even when semantic names are uncertain.
2. Keep binary parsing independent of Dark Magic rendering, DRLG, game data, and archive implementations.
3. Add structural JSON `inspect` output before adding engine-specific helpers.
4. Add synthetic redistributable fixtures for malformed/boundary behavior.
5. Add opt-in owned-asset probes controlled by environment variables; never commit Blizzard assets or decoded golden images.
6. Prefer assertions over structural hashes, counts, dimensions, offsets, and stream-consumption totals for owned-asset smoke tests.
7. Distinguish a binary-format validity constraint from an original Diablo II renderer/tool limit.
8. When research sources disagree, expose enough raw data to probe the disagreement rather than choosing one interpretation inside the codec.

## M40 creature-generation prerequisites

The creature asset program in `docs/CREATURE_ASSET_ROADMAP.md` adds an eventual
modern-to-legacy export consumer. Keep this work separate from the Dark Magic
planning change and deliver/tag it in the owning repositories before M40:

- `gravestench/dc6`: deterministic encoder/marshal API accepting indexed frames,
  signed offsets, directions, transparency index, and trailer policy; add
  synthetic round trips and stable structural inspection.
- `gravestench/dcc`: research and specify a deterministic encoder boundary only
  after the direction/cell trace work below is complete; reject unsupported
  inputs explicitly rather than producing plausible corrupt files.
- `gravestench/cof`: construction and strict validation for layers, directions,
  frames, per-frame order, and events, with deterministic marshal round trips.
- `gravestench/pl2`: reusable palette/transform lookup helpers and structural
  validation; Dark Magic still owns material regions and quantizer policy.

Binary layout, compression, and validation stay in codecs. Pose sampling,
component alignment/common origins, quantization policy, modern semantic loss
reports, and exporter orchestration stay in Dark Magic. No codec implementation
belongs in the Blender addon.

## `gravestench/dt1`

### P0: expose complete tile semantics

The engine needs lossless access to:

- both leading DT1 header words;
- tile light direction;
- roof height;
- material flags as raw `uint16` plus interpreted bits;
- signed geometry fields including height-to-bottom;
- orientation/type, main/style, sub/sequence;
- rarity/frame field;
- transparent RGB24 field;
- all 25 raw subtile collision bytes in file order and normalized 5x5 order;
- block offsets/length/count;
- cache index and unresolved header fields;
- block grid coordinates and raw format value;
- computed decoded bounding box.

Do not expose only a decoded image.

### P0: block-format regression suite

Create synthetic fixtures for:

- fixed isometric `0x0001` block with exact 256-byte payload;
- RLE `0x1001` block;
- isometric-RLE `0x2005` block;
- negative block X/Y;
- walls extending above/beyond nominal 160x80 bounds;
- floor/roof Y-origin normalization;
- RLE line reset `(0,0)`;
- overlong/truncated runs and destination overflow.

### P0: owned-asset probe command

`dt1 inspect --json` should emit at least:

```text
header words
tile count
every distinct tile type
material flag histogram
rarity histogram
block format histogram
min/max tile and block coordinates
min/max decoded bounds
unknown header-field value histograms
subtile flag bit histograms
```

Probe representative Act I-V files and record non-proprietary aggregate results in the codec repo.

### P1: semantic helpers

Safe helpers can include a logical key type:

```text
TileKey{Type, Style, Sequence}
```

Do **not** put rarity selection RNG in the codec; that belongs to Dark Magic map composition.

## `gravestench/ds1`

### P0: explicit version matrix in package docs/tests

Add structural fixtures covering every version gate used by Diablo II versions 1-18:

- object section v2;
- dependencies v3;
- explicit wall count v4;
- object flags v6;
- legacy orientation conversion boundary v7;
- Act v8;
- unknown 8-byte section v9-13;
- substitution method/tag v10;
- substitution groups v12;
- group extra word v13;
- paths v14;
- path actions v15;
- floor count v16;
- group preamble word v18.

The decoder should expose raw version-gated unknown words instead of silently skipping them when practical.

### P0: packed-cell API

Expose the raw 32-bit cell and named accessors for the 1.10f-observed fields:

- wall/floor;
- LOS/enclosed/exit;
- layer below/above;
- sequence bits 8-15;
- fill LOS/unwalkable;
- wall layer bits 18-19;
- style bits 20-25;
- reveal hidden/shadow/linkage/object wall/hidden;
- unknown bits 5 and 30.

Do not collapse the cell into just style/sequence/hidden.

### P0: raw orientation preservation

For versions `<7`, retain both serialized orientation value and converted compatibility tile type. The conversion table is an original-format compatibility fact, but callers may need raw values for inspection.

### P0: object/path/group probe

`ds1 inspect --json` should emit version, dimensions, Act, dependency strings, layer counts, orientation histogram, packed-bit histogram, object type/ID/position/flags, substitution groups, and path node actions.

### P1: no DRLG in codec

Do not add DT1 rarity lookup, room-edge wall remapping, warp creation, object-preset table resolution, or `LvlSub` behavior to this repo. Those are engine/data-table responsibilities.

## `gravestench/dc6`

### P0: retain complete raw header/frame metadata

Ensure API/inspector exposes:

- version/flags/encoding;
- four header termination bytes;
- directions and frames per direction;
- global frame offsets;
- `flipped`, width/height, signed offsets;
- unknown word, next-block word, payload length;
- per-frame trailer bytes if present in the file;
- anchor-relative frame bounds.

### P0: RLE boundary tests

Add synthetic tests for:

- transparent runs 1, 127, 128, 254, 255 split into legal commands for any future encoder;
- opaque run 127;
- alternating transparent/opaque spans;
- exactly one scanline and one pixel;
- truncated literal payload;
- missing/extra EOL;
- scanline overflow;
- frame dimensions above 256 to prove decoder safety policy is not a historical renderer limit.

### P0: offset/trailer consistency probe

For owned assets, compare:

```text
global frameOffsets[i+1]
frame.nextBlock
frame.length
bytes consumed including trailer
```

Record whether the 3-byte per-frame trailer is invariant, meaningful, or simply padding.

### P1: page-composition helper boundary

If a 256-page helper exists, keep it in a presentation/conversion package or command, not in core decode validation. Core DC6 must allow dimensions based on configured safe limits.

## `gravestench/dcc`

DCC has the largest remaining research risk.

### P0: direction-level trace inspector

Add a debug/inspect mode able to emit without rendering pixels:

```text
file header + direction offsets
direction uncompressed size
compression flags
all seven 4-bit encoded-width selectors
per-frame variable0/size/signed offsets/extra/compressed/flip
extra-data byte counts
all bitstream sizes
palette-value table
frame and direction bounding boxes
computed 4x4/edge cell geometry
stream bits consumed vs declared
```

### P0: cell reconstruction trace

Optional verbose tracing should print, per changed cell:

```text
frame index
cell coordinate and dimensions
previously touched?
equal-cell bit
pixel mask
raw/displacement encoding choice
pixel-code stack before/after palette-table mapping
selector bit width used during rasterization
copy/clear/fill/reconstruct action
```

This is essential for diagnosing stream-order bugs without comparing proprietary screenshots.

### P0: malformed/boundary fixtures

Cover:

- all encoded-width selectors including 0 and 32 bits;
- signed offset extrema for small selectors;
- frame heights 1-8;
- non-multiple-of-four frame/direction bounds;
- edge cells of every width/height 1-4;
- first-touch implicit `0xF` pixel mask;
- equal-cell reuse;
- changed cell with each 4-bit pixel mask;
- raw-coded stack;
- displacement chain with repeated `0xF` nibbles;
- early stack termination on repeated code;
- 1/2/3/4-color cell rasterization;
- stream sizes around and above 16 KiB;
- transparent/cleared frames;
- consecutive frames with different bottom rows.

### P0: remove folklore limits from validation

Do not reject valid assets based on 16 KiB stream buffers, ~200x200 frame folklore, or 75x75-cell folklore. Use checked arithmetic and configurable memory limits.

### P1: unresolved field research

Probe distributions/correlations for:

- file signature/version/tag/uncompressed-size values;
- frame `variable0`;
- frame `extraBytes` content;
- `compressedBytes` relation to actual stream consumption.

Keep names neutral until evidence supports stronger semantics.

## `gravestench/cof`

### P0: preserve the raw 25-byte header

The current research found a source disagreement over the final four header bytes. The codec should expose:

- layer/frame/direction/version;
- raw four bytes after version;
- signed composite bounds;
- raw final four rate/speed bytes;
- a 32-bit interpretation for probes;
- a low-byte speed interpretation for probes.

Do not discard upper bytes just because one implementation calls them unknown.

### P0: layer raw-byte semantics

Expose the five non-weapon bytes of every 9-byte layer record in raw form alongside helpers:

```text
component
shadow
selectable
byte3
byte4
weaponClass[4]
```

Document the competing interpretations of byte3/byte4 until owned-asset rendering probes resolve them.

### P0: event/priority validation

Tests should cover:

- all keyframe values 0-4;
- unknown keyframe value preservation/error policy;
- per-direction/per-frame changing priority;
- component IDs HD through S8;
- priority value referring to absent component;
- one-layer/one-frame short/special COFs, including the observed 42-byte case.

### P0: inspector

`cof inspect --json` should print raw header bytes, both speed interpretations, layer records, keyframe bytes, and a human-readable priority matrix using component codes.

### P1: keep resolver out of codec

Do not add player inventory, token/mode lookup, weapon-combination resolution, DCC filename construction, PL2 transforms, or rendering to `cof`. The codec owns file structure; Dark Magic owns composite resolution/playback.

## `gravestench/pl2`

The composite research implies future PL2 work may be needed even though PL2 was not a primary format in this task.

### P1: verify transform API is shader-friendly

The engine should be able to obtain a 256-entry palette-index remap without converting source DCC/DC6 frames to RGBA. Keep colormap rows immutable/cacheable and expose tint metadata separately.

### P1: transform probe

Record table counts, row sizes, tint records, and any Act-specific differences using owned files. Avoid encoding higher-level "equipment color" meaning into the PL2 codec.

## Dark Magic-only follow-ups, not codec work

Codex should plan these in `gravestench/dark-magic`, after the codec probes above are available:

- deterministic Diablo seed/RNG type and trace tooling;
- level-table relationships (`Levels`, `LvlTypes`, `LvlPrest`, `LvlMaze`, `LvlSub`, `LvlWarp`);
- DRLG preset/maze/outdoor room graph generation;
- room-edge tile linking/remapping;
- DT1 rarity-weighted tile choice using room seed;
- subtile collision assembly;
- warp and automap semantic tile processing;
- composite token/mode/weapon-class resolver;
- equipment/component appearance resolver;
- direction mapping;
- synchronized composite playback/events;
- indexed shader path for PL2/component transforms;
- shadow, selection, and final composite bounds.

## Suggested sequencing for Codex

```text
1. cof raw-header + inspector probes
2. dcc trace inspector + boundary fixtures
3. dt1 semantic metadata + aggregate inspector
4. ds1 version/packed-cell inspector
5. dc6 trailer/offset + >256 regression tests
6. pl2 indexed-transform API audit
7. only then plan Dark Magic composite runtime and DRLG replacement work
```

The first goal is to make uncertainties observable. Do not start by rewriting Dark Magic's renderer or map generator around assumptions that the codec APIs cannot yet prove.
