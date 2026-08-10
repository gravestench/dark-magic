# COF composite animation format

Status: research specification. COF parsing and composite resource resolution are separate concerns; see `../COMPOSITE_ANIMATION.md`.

## Purpose

A COF describes how independently encoded body/component animations are synchronized and layered. It provides:

- layer/component membership;
- direction count;
- frame count per direction;
- composite bounding metadata;
- animation rate/timing metadata;
- per-layer shadow/selectability/transparency metadata;
- per-frame event/keyframe values;
- per-direction, per-frame component draw ordering.

It does **not** by itself choose which armor appearance code or weapon component file to use.

## Header

Riiablo's later parser consumes:

```text
u8 layerCount
u8 framesPerDirection
u8 directionCount
u8 version
byte unknown[4]
s32 xMin
s32 xMax
s32 yMin
s32 yMax
u32 animationRate
```

OpenDiablo2 parses the same leading counts but interprets the final four bytes of the header differently: three unknown bytes plus a one-byte speed value.

This is currently unresolved. Dark Magic should preserve the full 32-bit field and expose raw bytes until owned-asset probes establish whether the upper 24 bits are always zero/reserved and how the value relates to `AnimData` timing.

## Layer record

Each layer occupies 9 bytes:

```text
u8 component
u8 shadow
u8 selectable
u8 overrideTransLvl
u8 newTransLvl
char weaponClass[4]
```

OpenDiablo2 names bytes 3/4 as transparent + draw effect; Riiablo names them override-transparency-level + new-transparency-level and maps values to blend/transparency behavior in the renderer. The raw bytes agree; the semantic naming differs.

The safe format contract is therefore to expose both raw values and interpreted helpers, not overwrite them with one implementation's naming.

## Component identifiers

The complete observed COF component namespace is 16 entries:

```text
0x00 HD  head
0x01 TR  torso
0x02 LG  legs
0x03 RA  right arm
0x04 LA  left arm
0x05 RH  right hand
0x06 LH  left hand
0x07 SH  shield
0x08 S1  special 1
0x09 S2  special 2
0x0A S3  special 3
0x0B S4  special 4
0x0C S5  special 5
0x0D S6  special 6
0x0E S7  special 7
0x0F S8  special 8
```

Do not hard-code a smaller humanoid-only component set.

## Weapon-class field

Each layer carries a four-byte, normally NUL-padded weapon-class string. This string participates in DCC/DC6 filename resolution for the layer. It is not necessarily the same as the unit's top-level COF weapon class.

## Event/keyframe table

After all layer records comes a frame-event array. Normal assets use one byte per composite frame.

Known values corroborated by Riiablo are:

```text
0 none
1 attack
2 missile
3 sound
4 skill
```

These are gameplay-facing synchronization points, not merely renderer metadata. The animation controller should surface them as events when the composite timeline crosses the corresponding frame.

Riiablo contains a special-case observation for a 42-byte, one-layer, one-direction, one-frame COF where four keyframe bytes are read. This must be probed before becoming a general codec rule.

## Priority / draw-order table

The remaining normal payload is:

```text
layerOrder[directionCount][framesPerDirection][layerCount]
```

Each byte is a **component identifier**, not an index into the file's layer-record array.

For a given direction and frame, iterate the `layerCount` entries in stored order and draw the corresponding component layers in that order.

Therefore priority can change:

- between directions;
- between frames within one direction.

Do not pre-sort layers once per COF.

## Direction mapping

COF/DCC direction ordering is not the same thing as an arbitrary engine-facing 0..N-1 angular index. Riiablo exposes explicit 8/16/32 direction remapping, while OpenDiablo2 uses lookup tables from its 64-direction engine representation to the number of directions in the COF.

Dark Magic should define one canonical world-facing direction representation and one tested compatibility mapping to encoded direction indices. Do not allow each renderer/resource loader to invent its own mapping.

## Timing

Both Riiablo and OpenDiablo2 model Diablo animation timing around 25 updates/frames per second and a 1/256 rate unit:

```text
frameDurationSeconds = 256 / (rate * 25)
```

Riiablo's animation helper exposes this as `setFrameDelta(rate)`.

However, many unit animations also have `AnimData` speed/factor data, and Riiablo's runtime may override the COF-derived frame delta with resolved animation data. Therefore:

- preserve the COF rate exactly;
- treat it as authored timing metadata;
- let the unit animation resolver decide whether external data overrides it.

## Shadow flag

A layer with shadow enabled participates in the composite shadow pass. Riiablo draws a flattened/sheared projection of that component using the same current direction/frame and anchor-relative box.

The exact original shadow transform belongs in renderer compatibility research, not in the COF parser.

## Selectable flag

The layer's selectability byte is rendering/input metadata that can influence hit testing. It does not change DCC decoding.

A correct final composite hit-test implementation should be able to ignore pixels from non-selectable components while still rendering them.

## Transparency / draw effect bytes

Observed rendering behavior includes alpha-like levels and luminosity modes. Riiablo maps `newTransLvl` values roughly as:

```text
0 -> 75% style transparency
1 -> 50%
2 -> 25%
3 -> luminosity
4 -> luminosity-like, uncertain
6 -> luminosity-like, uncertain
```

This mapping is **implementation-correlated but not yet verified original behavior**. Keep it marked medium/uncertain until a Phrozen Keep/original-asset rendering probe corroborates each value.

## Bounding box

The COF header contains composite bounds. Individual DCC/DC6 layers also have direction/frame bounds. Runtime implementations often recompute the active composite bounds as the union of currently loaded component bounds for the current direction/frame.

Dark Magic should retain both:

- authored COF bounds for diagnostics/compatibility;
- actual active-layer bounds for culling and hit testing.

## Synchronization contract

All component layers under a COF share:

- one composite direction;
- one composite frame index/timeline;
- one unit/world anchor;
- the COF priority table for the current encoded direction/frame.

Do not advance each DCC component independently. Their frame numbers are synchronized by the composite controller even though each component is a separately decoded asset.

## Format vs runtime responsibilities

### Codec owns

- exact bytes/field widths;
- layer records;
- event array;
- priority bytes;
- header bounds/rate/raw unknown bytes;
- structural validation.

### Composite runtime owns

- token/mode/weapon/equipment filename resolution;
- direction conversion from world orientation;
- frame clock and external AnimData overrides;
- shadow pass;
- palette/PL2 transforms;
- component suppression/fallback;
- hit testing;
- cache invalidation when equipment changes.

## Malformed-input handling

Reject or safely bound:

- zero/absurd layer, direction, or frame counts where payload arithmetic overflows;
- component IDs outside 0..15;
- truncated layer records;
- event table shorter than the version/asset-family rule requires;
- priority table shorter than `dirs * frames * layers`;
- priority component values not present in the layer set, unless owned assets demonstrate a legitimate suppression convention;
- unterminated/invalid weapon-class bytes only if they violate the fixed four-byte field, not merely because they are unknown strings.

## Evidence record

| Fact | Source | Version | Confidence |
| --- | --- | --- | --- |
| 16 component identifiers HD..S8 | Riiablo `Cof.java`, `Engine.java` | pinned | verified |
| layer record is 9 bytes | Riiablo + OpenDiablo2 | pinned/current | verified |
| priority is dirs x frames x layers and stores component IDs | Riiablo + OpenDiablo2 | pinned/current | verified |
| keyframes 0..4 map none/attack/missile/sound/skill | Riiablo | pinned | high |
| frame duration uses 25Hz and 1/256 rate unit | Riiablo + OpenDiablo2 runtime | pinned/current | verified implementation behavior; original timing high |
| 32-bit rate vs 1-byte speed interpretation | Riiablo vs OpenDiablo2 | pinned/current | uncertain/conflicting |
| transparency-value blend interpretation | Riiablo `Animation.java` | pinned | medium |

## Probe requirements

See `CODEC_FOLLOWUPS.md`. COF probes should dump the raw 25-byte header, both interpretations of the final four bytes, layer raw bytes, event bytes, per-direction/frame priority matrices, and all observed component/weapon-class values across player, monster, and object COFs.
