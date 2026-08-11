# Item stats, affixes, and property evaluation

Status: implementation-oriented research baseline. Dark Magic already parses and materializes many generated item properties in M6; the missing foundation is a complete runtime stat-source/aggregation model with exact legacy scaling, dependency, serialization, and invalidation semantics.

## Executive result

Do not represent a unit's stats as one flat `map[statID]int` with no provenance.

D2MOO's reconstructed `StatList`/`StatListEx` behavior strongly supports a layered model:

```text
unit base stats
+ linked stat sources (equipment, state, aura, skill, temporary list, etc.)
+ metadata-driven derived/op stats
= queried effective stat
```

The canonical runtime needs to preserve **which source contributed which stat** so removing one item/state does not require guessing how to subtract a previously flattened result.

## Current Dark Magic baseline

M6 already has substantial item-generation machinery: typed ItemStatCost/Properties, unique/set/magic/rare/automagic selection, deterministic property rolls, many property functions, and portable generated items. This remains the property-generation front end. The next layer converts modifiers into durable/runtime stat sources and lets combat/progression query effective values.

`internal/game/data/catalog` already indexes `ItemStatsByName` and `PropertiesByCode`. Do not create an independent metadata loader.

## D2MOO stat-list evidence

Pinned D2MOO `D2StatList.cpp` reconstructs:

- ordered arrays keyed by packed stat/layer identity;
- packed identity rather than stat ID alone;
- zero removal unless ItemStatCost says keep zero;
- metadata-gated callbacks;
- extended stat lists accumulating base plus linked lists;
- special handling for damage-related lists;
- ItemStatCost-driven op-stat evaluation;
- `ValShift` use in calculations;
- derived operations that recursively query other stats.

Confidence: **high** for 1.10f reconstruction; exact formulas/rounding still need traces.

## Canonical identities

Define:

```text
StatID
StatParameter / layer
StatKey = (StatID, parameter/layer)
```

Do not discard the parameter. Normalize ItemStatCost into an immutable view:

```text
StatDefinition
  legacy numeric stat ID
  symbolic name
  save/send encoding metadata
  signedness
  value shift / additive bias
  parameter bits
  keep-zero behavior
  op/dependency metadata
  callback/invalidation flags
  display metadata
```

The raw table record remains source evidence.

## Stat sources

Use source provenance:

```text
StatSource
  source ID
  owner entity/item
  kind
  lifetime
  flags
  ordered StatEntry[]
```

Kinds include base unit stats, equipped item, socket content, set bonus, charm, passive skill, aura, state/buff/debuff, temporary combat modifier, quest reward, and difficulty/environment modifier.

A stable source ID makes attach/remove idempotent. Equipping attaches an item source; unequipping removes the same source. State expiration removes the state source. Neither edits an unexplained total.

## Raw, logical, effective, displayed

Use explicit terminology:

- **raw encoded value**: save/network/item bits;
- **logical source value**: decoded after signedness/Add/ValShift rules;
- **accumulated base value**: direct sources combined;
- **effective value**: derived operations/caps applied;
- **display value**: UI formatting/rounding.

Do not reuse one `Value` field for all meanings.

## Fixed point and rounding

For every formula record storage width, sign, save bias, parameter width, fixed-point shift, operation order, truncation direction, cap/floor, and overflow behavior.

Do not make `float64` a compatibility contract because reversed code displays `/100.0`. Build integer vectors against original behavior first.

## Derived/op stats

D2MOO shows ItemStatCost-driven operations that depend on other stats and op bases. Normalize this into a dependency graph. The evaluator needs cycle detection, recursion guards, deterministic order, memoization by revision, and invalidation when a source changes.

A practical modern design is an entity `StatRevision` incremented whenever its active source set changes; cached effective values are tagged with that revision.

## Property application

Keep property interpretation separate from effective-stat resolution:

```text
GeneratedProperty
  property code + rolled params/values
        |
        v
PropertyInterpreter (Properties + ItemStatCost)
        |
        v
StatEntry[] attached to item/source
        |
        v
EffectiveStatResolver
```

This matches the fact that generation starts with property codes while save/network formats commonly serialize resulting stat IDs/values.

## Affix eligibility

Eligibility belongs to generation, not stat aggregation. Preserve item type inheritance/exclusions, qlvl/ilvl/alvl, magic level, frequency/group, spawnability, expansion/ladder/class restrictions, automagic, rare-name rules, and quality-specific counts.

## Local versus global modifiers

Classify whether a modifier changes the item itself or its owning unit. Enhanced weapon damage/defense, durability, sockets, requirements, speed/range, and local elemental damage can be item-local; attributes/resistances/skills can contribute to unit stats.

Do not attach every property directly to the player merely because the item is equipped.

## Location-based activation

A modifier can exist but be inactive:

```text
inventory          -> only inventory-active effects
active equipment   -> equipped/global + item-local combat contribution
inactive weapon set -> preserve item but suppress relevant active-hand source
stash/cube         -> normally inactive
held/trade/vendor/world/corpse -> normally inactive
```

Exact exceptions belong to item/charms/set/runeword research. Activation changes should not regenerate immutable item rolls.

## Temporary states, auras, skills

States/aura sources need stable source/caster ID, state ID, start/expire tick, stat entries, stacking/refresh group, and removal cause. An aura may refresh sources on targets; an effective-stat query should not invoke arbitrary skill script logic.

## Serialization

Legacy item/save/network codecs adapt `StatEntry` plus raw metadata. Exact Save Bits, Save Add, parameter bits, signed behavior, and grouped-property formats need byte-level vectors. Unknown records should survive a lossless save round trip.

Canonical Dark Magic persistence may be semantic and need not mimic legacy bit packing internally.

## Replication

Do not replicate every private source. Separate owner-only canonical stats, nearby/public summaries, item description stats, presentation summaries, and server-only hidden data. Server calculates authoritative effective values.

## Failure/idempotency

- attaching the same source ID twice is replacement or a hard error, never double-apply;
- removing a missing source is defined;
- failed equipment transaction leaves sources unchanged;
- expiration and dispel in one tick cannot subtract twice;
- dependency cycle is diagnosed with stat IDs;
- unsupported op metadata cannot silently return a plausible value;
- unknown save stat records remain preservable.

## Implementation slices

1. normalized `StatDefinition` from ItemStatCost;
2. `StatKey` + `StatSource` model;
3. effective resolver with op dependencies/cycle guard;
4. source revision/invalidation;
5. bridge current generated properties to sources;
6. state/aura timed-source bridge;
7. exact legacy save/network vectors;
8. combat-facing narrow stat views.

## Acceptance criteria

- Same stat ID with different parameters can coexist.
- Equip then unequip returns exactly to prior effective snapshot.
- Replacing the same source is idempotent.
- Inactive weapon-set sources do not leak active-hand effects.
- Source changes invalidate dependent effective stats deterministically.
- Dependency cycles fail with actionable diagnostics.
- Fixed-point vectors avoid unspecified float rounding.
- `KeepZero` behavior is represented where required.
- Unknown save stat records survive lossless round trip.
- Property generation and stat aggregation remain separate layers.

## Verification backlog

1. Enumerate every ItemStatCost op form used by shipped LoD data.
2. Build owned-game vectors for Save Bits/Save Add/ValShift/signed/parameter encoding.
3. Trace local/global enhanced damage/defense and requirements.
4. Trace set bonus thresholds/source replacement.
5. Trace charm activation by container/type.
6. Trace aura refresh/stack/removal.
7. Trace poison/regeneration fixed-point accumulation/display rounding.
8. Trace elemental damage ranges/durations.
9. Trace oskills/class skills/tabs/charges/procs/aura stats through save and gameplay.
10. Determine exact caps/floors/order for resist, absorb, block, speed, leech, MF/GF.
11. Verify overflow/truncation for extreme mod values.
12. Compare blind property-roll vectors against original binary traces.

## Sources

- Dark Magic `internal/game/data/catalog`, M6 roadmap, and current loot/property implementation.
- [D2MOO `D2StatList.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Common/src/D2StatList.cpp).
- [D2MOO `Items.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Common/src/Items/Items.cpp).
- Bundled Data File Guide sections for `ItemStatCost.txt`, `Properties.txt`, and affix tables.
- [libd2 verification](https://github.com/jaenster/libd2/blob/e6cdc4927c6180be8dd309b0423b470f64f1fc6c/docs/VERIFICATION.md).
