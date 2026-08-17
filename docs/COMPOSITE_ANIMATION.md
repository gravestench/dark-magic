# Diablo II composite animation resolution and rendering

Status: research architecture/specification. This document separates observable resource naming/rendering behavior from third-party engine structure.

## Overview

A Diablo II composite unit is not one sprite sheet. The runtime resolves a COF for the unit's token/mode/weapon class, then resolves one DCC or DC6 per active COF component using equipment/body-part appearance codes. All components share one composite timeline and anchor. COF priority chooses draw order independently for every encoded direction and frame.

The conceptual pipeline is:

```text
unit class/type
  + token
  + animation mode
  + resolved weapon class
  + component appearance codes
        |
        v
COF resource
        |
        v
COF component list + layer weapon classes
        |
        v
component DCC/DC6 resources
        |
        v
direction mapping + synchronized frame
        |
        v
COF per-frame ordering
        |
        v
palette / PL2 transform / alpha / shadow
        |
        v
final draw calls + bounds + hit-test data
```

## Resource namespaces

OpenDiablo2 and Riiablo independently use these legacy roots:

```text
players:  data/global/chars
monsters: data/global/monsters
objects:  data/global/objects
```

The exact unit-type naming in Dark Magic can differ; compatibility path construction must produce the legacy resource names.

## COF filename construction

The observed COF path pattern is:

```text
<root>/<token>/COF/<token><mode><weaponClass>.COF
```

Example shape:

```text
data/global/chars/AM/COF/AMNUHTH.COF
```

The literal example is illustrative; callers must resolve the actual token/mode/weapon class from game data rather than hard-coding this combination.

## DCC/DC6 component filename construction

For each COF layer/component:

```text
<root>/<token>/<component>/<token><component><appearance><mode><layerWeaponClass>.dcc
```

with `.dc6` as a fallback asset encoding where present.

Important distinctions:

- the COF's top-level filename uses the **resolved unit weapon class**;
- the component animation filename uses the **weapon class stored on that COF layer**;
- the appearance segment is the current component/body-part graphic code (`lit` or equipment-derived alternatives).

Riiablo's later loader constructs the same relationship while allowing dirty components to reload independently.

## Complete component namespace

COF defines 16 component IDs:

```text
HD head
TR torso
LG legs
RA right arm
LA left arm
RH right hand
LH left hand
SH shield
S1..S8 special components
```

A humanoid player normally uses only a subset, but engine state and caches must address the full 16-entry space.

## Component appearance resolution

### Default / missing equipment

Riiablo uses a distinguished default component appearance corresponding to `lit`. A component can also be explicitly suppressed (`NIL`) rather than resolved to `lit`.

Dark Magic should represent at least three states:

```text
DefaultAppearance
SpecificAppearance(code)
Suppressed
```

Do not encode suppression as an empty filename after resource construction has begun.

### Armor

Player body armor can independently select appearance variants for:

- torso;
- legs;
- right arm;
- left arm;
- left shoulder special component;
- right shoulder special component.

Riiablo derives these from armor data and applies one packed color transform to the affected body components.

Headgear resolves the HD appearance independently from body armor.

### Weapons and shield

Right-arm and left-arm inventory slots are resolved into RH, LH, and SH visual components based on whether the equipped item is a weapon or shield and which hand owns it.

A bow-like weapon may visually occupy the LH component even if gameplay inventory semantics present it as one two-handed item. Therefore component ownership is a rendering-resolution rule, not necessarily the same as inventory slot naming.

## Weapon-class resolution

Observed weapon-class codes include:

```text
HTH BOW 1HS 1HT STF 2HS 2HT XBW
1JS 1JT 1SS 1ST HT1 HT2
```

For two one-handed weapons Riiablo resolves combined classes according to left/right swing/thrust categories:

```text
LH 1hs + RH 1hs -> 1SS
LH 1hs + RH 1ht -> 1ST
LH 1ht + RH 1hs -> 1JS
LH 1ht + RH 1ht -> 1JT
```

Two hand-to-hand/claw weapons can resolve to `HT2`. Missile/bow classes use the equipped weapon's authored class. No weapon resolves to `HTH`.

This is mature implementation evidence and should be corroborated against original data/behavior before the resolver is treated as complete, especially for unusual monster/object combinations and class-specific weapons.

## Fallback behavior

Different fallback concepts must remain separate:

1. **component default appearance**: absent body equipment -> `lit` where the COF expects the component;
2. **component suppression**: no weapon/shield -> omit RH/LH/SH as appropriate;
3. **file encoding fallback**: try DCC, then DC6 if the authored component exists in DC6 form;
4. **weapon-class fallback**: compatibility rule when a mode/class COF does not exist;
5. **missing resource error**: a genuinely inconsistent resource set.

OpenDiablo2's implementation fails mode construction when the requested COF path does not exist and defaults an empty component value to `lit`. Riiablo explicitly models dirty/suppressed/default component states. Dark Magic should make fallback policy data-driven and observable rather than silently substituting unrelated animations.

## Caching

Semantically safe cache layers are:

```text
raw file bytes / archive lookup
  -> decoded immutable COF
  -> decoded immutable DCC/DC6 direction/frame data
  -> palette / PL2 tables
```

Mutable unit state should not be cached inside those assets:

```text
current mode
current direction
frame progress
loop/subloop
component selection
component transform/alpha
selection/highlight state
```

Equipment changes should invalidate only the affected component resource descriptors when the COF/mode/weapon class itself remains valid. If the weapon class changes, the COF can change and all COF-defined component descriptors must be reconsidered.

## Direction mapping

Dark Magic should define one canonical world-facing direction type. The compatibility layer maps it to encoded COF/DCC direction indices.

Riiablo's encoded-direction mapping is nontrivial for 8/16/32 directions, and OpenDiablo2 maps a 64-direction world orientation through direction-count-specific lookup tables.

Direction changes should not inherently reset animation frame progress. They select another encoded direction for the same composite timeline unless gameplay state explicitly changes mode.

## Composite timing

All component animations must share one composite frame clock.

The classic timing basis exposed by both Riiablo and OpenDiablo2 is:

```text
25 ticks / second
rate unit = 1/256
frameDuration = 256 / (rate * 25) seconds
```

External animation data can override the COF-authored rate. Preserve both the raw COF rate and the resolved runtime frame delta.

### Equipment changes and progress

Riiablo's later component-loading architecture can change component descriptors without replacing the high-level animation object, which supports preserving frame progress during equipment-only changes.

OpenDiablo2's `Equip` rebuilds its mode object, which may lose some transient mode state. That is an implementation choice, not evidence that Diablo II resets animation progress whenever armor changes.

Dark Magic should preserve direction, normalized frame progress, loop state, and event cursor when only component appearance changes, unless an original-behavior probe establishes a reset.

## World anchor and frame placement

Every component is drawn from the same unit/world anchor.

For a decoded frame box:

```text
drawX = anchorX + box.xMin
drawY = anchorY - box.yMax
```

This is the placement convention used by Riiablo's renderer after normalizing DCC/DC6 frame bounds.

Do not center each component texture independently. The encoded offsets are what align head, torso, limbs, weapons, and shield.

## Per-frame draw ordering

For current encoded direction `d` and frame `f`:

```text
for each componentId in cof.priority[d][f]:
    if component is active:
        draw component frame f at shared anchor
```

Priority entries are component IDs and can vary by frame and direction. This allows a weapon/arm to move in front of or behind torso components during an animation.

## Shadows

COF layers with the shadow flag participate in a separate shadow render pass.

Riiablo projects each flagged component using the same direction/frame but applies a flattened/sheared transform and a dark transparent tint. OpenDiablo2 also performs a separate shadow/non-shadow pass over the COF priority order.

The existence of a separate component shadow pass is high confidence; exact shear, alpha, and lighting interaction still need original rendering probes.

## Palette and transforms

DCC/DC6 pixels are palette indices. Rendering may apply:

- base Act/unit palette;
- PL2 colormap;
- item/equipment transform;
- unit transform;
- component-local transform;
- alpha/transparency/luminosity effect;
- highlight/selection effect.

Riiablo models PL2 as 256-entry index remap tables plus tint records. Equipment transforms can be packed as:

```text
high 3 bits: transform table/category
low 5 bits: color index
```

for component-specific appearance coloration.

Dark Magic's modern shader path should consume indexed textures plus transform parameters rather than permanently recoloring decoded images on the CPU.

## Transform precedence

A practical independent model is:

```text
source palette index
 -> component/equipment PL2 mapping
 -> optional unit/global mapping
 -> palette lookup
 -> COF blend/transparency mode
 -> shadow/highlight pass treatment
```

The exact original precedence between unit and component transforms needs dedicated probes before it becomes a compatibility guarantee.

## Animation events

COF per-frame event values:

```text
1 attack
2 missile
3 sound
4 skill
```

must be surfaced by the timeline controller when frame boundaries are crossed. They are not draw callbacks.

The event dispatcher must account for:

- multiple frames advanced during a long update;
- loops wrapping over frame zero;
- subloops;
- reversed/clamped playback if Dark Magic exposes those modes;
- mode changes cancelling stale events.

## Looping and subloops

Looping/subloop behavior is runtime state, not encoded directly by the DCC frames. OpenDiablo2 exposes loop and subloop controls on component animations. Missile animations often use external missile metadata for direction count, frame count, speed, and looping because they can use DCC without a COF.

Missiles should therefore share the indexed-frame renderer but not be forced through composite/COF resolution.

## Hit testing and selection

COF layer metadata includes a selectability byte. Final selection should be based on rendered component masks/indices at the current frame, honoring:

- component presence;
- current direction/frame offsets;
- COF selectability;
- transparent source pixels;
- optional gameplay selection bounds.

A union rectangle alone is useful for broad-phase hit testing but cannot reproduce per-pixel selection.

## Final composite bounds

For the current direction/frame, union the active component frame boxes. Keep this separate from the authored COF header bounds.

Useful outputs:

```text
activeBounds      current direction/frame and equipment
modeStableBounds  union across frames/directions if needed for culling
cofAuthoredBounds raw file metadata for validation/probes
```

## Worked resolution example

Illustrative player trace:

```text
unit type: player
class -> token AM
mode -> NU
inventory -> no weapon => weapon class HTH

COF path:
  data/global/chars/AM/COF/AMNUHTH.COF

COF says active components include HD/TR/LG/RA/LA/... with layer weapon classes

appearance state:
  HD = helmet alternateGfx or lit
  TR/LG/RA/LA/S1/S2 = armor component codes or lit
  RH/LH/SH = suppressed because nothing equipped

for each active COF component C:
  data/global/chars/AM/<C>/AM<C><appearance><NU><cofLayer.weaponClass>.dcc
  fallback to .dc6 only if that legacy resource actually exists

runtime direction -> encoded COF direction d
composite frame -> f
priority = COF.order[d][f]

for component in priority:
  resolve decoded frame box
  apply component transform/PL2
  draw palette-indexed frame at shared world anchor

also:
  draw COF-shadow-enabled components in shadow pass
  dispatch COF event for frame f if present
```

The important part of this example is the chain of responsibility, not the literal AM/NU/HTH combination.

## Dark Magic architecture recommendation

```text
CompositeDefinition
  token, namespace, mode, resolved weapon class, COF

CompositeAppearance
  [16] component state {suppressed/default/code, transform, alpha}

CompositePlayback
  direction, frameClock, frame, loop/subloop, event cursor

CompositeResources
  immutable COF + independently cached component DCC/DC6 assets

CompositeRenderer
  direction mapping -> COF order -> shared-anchor draw calls
```

This architecture supports a future modern animation system while retaining an adapter that resolves legacy COF/DCC assets to the same high-level component/timeline interface.

## Implemented playable-character boundary

The first Dark Magic adapter now follows the split above:

- immutable COF, DCC, and fully composed direction/mode results are cached;
- `AnimData.d2` supplies typed frame counts, rates, and sparse event bytes;
- one Lua-owned playback clock survives facing and equipment changes, while an
  authoritative mode change starts a new timeline;
- skipped-frame-safe advancement publishes every crossed animation event;
- authoritative equipment slots publish validated component appearance codes
  and active-hand weapon classes, without publishing archive paths;
- COF shadow flags produce a separate tinted shadow draw before their layer.

The composite renderer remains a legacy presentation adapter, but AnimData is
no longer presentation-only input. Its effective binary is pinned with the
authoritative data generation, and a separate renderer-free capability exposes
the same frame count, 24.8 rate, and typed event bytes to gameplay. d2legacy
chooses the actor/mode/weapon key and converts the relevant marker and cursor
wrap into fixed simulation ticks; rendering cannot decide or delay the effect.

## Open questions after the first implementation

- exact weapon-class fallback chain used by original Diablo II when a COF is absent;
- exact COF transparency-byte semantics and blend precedence;
- exact shadow transform and alpha;
- which exceptional mode transitions should retain progress rather than starting
  a new authoritative timeline;
- monster/object component appearance lookup rules across all data-table families;
- selection semantics for non-selectable COF layers;
- missile timing/loop rules by `Missiles.txt` fields.

These are tracked as engine/probe work in `formats/CODEC_FOLLOWUPS.md` where codec exposure is required, and otherwise should be planned as Dark Magic composite-runtime work.
