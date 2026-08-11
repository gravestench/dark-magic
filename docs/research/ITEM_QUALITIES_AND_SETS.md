# Unique, set, rare, magic, crafted, personalized, superior, low-quality, and ethereal item research

Status: implementation-oriented research baseline. Dark Magic M6 already covers much of quality election and affix/property materialization. This document focuses on the remaining runtime state, generation-context differences, special-record semantics, set activation, durability/ethereal/personalization, and transformations that must become first-class item-instance state.

## Executive conclusion

`Quality` is not enough to describe a Diablo II item.

A durable item instance eventually needs to preserve orthogonal facts such as:

```text
base item identity
item level / quality level context
generation seed(s)
quality family
concrete unique/set record, if any
rare/crafted display name parts
magic prefixes/suffixes
automagic/staffmods
rolled property instances
identified state
ethereal state
current/max durability
quantity
socket capacity + inserted items
runeword identity/stat source
personalized name
quest/special flags
version/ladder/source metadata
```

Do not encode these by cloning a base-item definition and permanently mutating it.

## Current Dark Magic M6 coverage

Dark Magic already has:

- `Quality` enum for unique/set/rare/magic/superior/normal/low;
- ItemRatio-based quality chance calculations;
- item-type Normal/Magic/Rare restrictions;
- UniqueItems/SetItems availability by level/version/ladder;
- weighted concrete special-item selection;
- quality fallback when no eligible Unique/Set exists;
- magic/rare prefix/suffix parsing and selection;
- automagic support;
- property-value materialization;
- set/unique table schemas in the typed catalog.

The goal is to extend this into a **complete item instance**, not replace it.

## Quality family versus concrete identity

Separate:

```text
rolled quality = unique
```

from:

```text
concrete UniqueItems row = Stone of Jordan
```

and likewise for set items.

A unique/set quality roll can fail to find an eligible concrete row. That failure requires a verified fallback path. Dark Magic already falls Unique/Set toward Rare when no eligible special record exists; retain the distinction between:

- quality check succeeded;
- concrete record election succeeded;
- final fallback quality.

Do not throw away the diagnostic reason.

## Per-game unique occurrence state

D2MOO 1.10f maintains a game-level bitset recording unique item rows that have already dropped. Unique records with a no-limit flag bypass that restriction.

This means concrete Unique eligibility may depend on mutable **session loot history**, not just static tables and item level.

If Dark Magic adopts this behavior, add a deterministic/checkpointed state participant such as:

```text
UniqueOccurrenceState
  generation/content fingerprint
  set of consumed unique-record IDs
```

It must be reset per game/session, not persisted as a permanent character property.

## Set items

Set item identity has two layers:

1. a concrete SetItems row carried by one item;
2. the Sets row describing group bonuses.

An equipped set item contributes:

- its ordinary item properties;
- possible item-level partial bonuses depending on number of equipped pieces;
- group/set bonuses depending on piece count.

D2Common property research shows state/stat-list categories associated with set-piece activation rather than rewriting the item's base stats permanently.

This fits the layered source model from `ITEM_STATS_AND_AFFIXES.md`:

```text
equipped set item source
+ conditional partial-set source(s)
+ conditional full/group-set source(s)
= effective stats
```

When a piece is removed, only the sources whose conditions no longer hold should disappear.

### Verification needs

- duplicate-equipping semantics where possible;
- partial bonus thresholds;
- `addfunc` variations;
- exact source/target of set bonuses;
- display versus effective stat ordering;
- transformation/color effects.

## Rare items

A rare item is not just a magic item with more affixes.

It has at least:

- rare-name prefix and suffix identity;
- multiple magic affix records;
- item-level/affix-level eligibility;
- group/exclusion rules;
- optional automagic/staffmods;
- identification state.

Dark Magic's current rare affix selection can become the core, but durable item serialization must preserve the exact selected affix and name-part identities rather than rerolling from a seed on load.

Do not rely on regenerated affixes for save/network identity.

## Crafted items

Crafted items combine recipe-fixed properties with rare-like random affixes.

The Cube workstream should produce an ItemGenerationRequest whose generation context is `CraftedOutput`, then materialize:

```text
base output item
+ forced crafted quality
+ recipe fixed properties
+ random affix budget/eligibility
+ output ilvl/affix level policy
```

Keep fixed recipe properties and random affixes as separate stat-source/provenance categories for debugging and display.

Exact crafted affix count and output-level formulas are probe items.

## Magic items

Magic items normally carry prefix/suffix identities plus any automagic/staffmods allowed by the base type.

The current affix machinery should eventually expose an inspectable trace:

```text
item ilvl / qlvl / magic level
effective affix level
eligible groups
prefix roll
suffix roll
forced/rejected groups
rolled property values
```

This is especially useful for Cube rerolls, charms, vendor stock, and gambling.

## Superior and low-quality items

Superior and inferior/low quality are distinct generation paths with base-stat modification semantics, not merely label/color choices.

Future item state should preserve the concrete high/low-quality modifier identity where the original data/rules choose one. Dark Magic already loads `QualityModifiers` and `LowQualityItemNames`; connect them to actual materialization rather than hard-coding a display adjective.

Research needs to pin:

- enhanced damage/defense/durability ranges;
- low-quality penalties;
- low-quality name selection;
- socket interactions;
- repair/durability behavior;
- upgrade recipe preservation/transformation.

## Ethereal is an orthogonal flag

D2MOO's ordinary 1.10f generation path provides strong evidence:

- normal chance is 5% in that path;
- explicit never/always flags can override;
- eligible items must be weapon/armor with durability;
- inferior, set, and quest items are excluded by that path;
- ethereality applies a stat transformation;
- max durability becomes `oldMax / 2 + 1`, then current durability fills to the new max.

Ethereal should therefore be an item-instance flag plus derived stat source, not a separate quality value.

Vendor repair logic also treats ethereal items specially. Preserve that distinction all the way into commerce.

## Durability and broken state

Durability is mutable item state.

The original item path includes:

- chance-based durability loss on qualifying use/hit contexts;
- separate weapon/armor/throwable probabilities;
- current/max durability;
- broken handling;
- replenishing durability scheduled in game frames;
- throwable quantity/durability interactions;
- indestructible state.

Item authority should own this state so equipment effects can disable when broken without destroying the item identity.

Suggested fields:

```text
Durability
MaxDurability
Broken
Indestructible
ReplenishRate/source
```

Whether `Broken` is stored or derived from durability + item flags should be decided from save/runtime evidence.

## Identification

Identification affects presentation and potentially interaction policy, but not whether the underlying rolled properties exist.

Keep:

```text
rolled item properties = authoritative and already present
identified flag = controls which of those facts the player may inspect/use for some UI flows
```

Do not roll properties at identification time.

Cain/vendor identify services should toggle authoritative state through a trusted transaction.

## Personalization

Personalization is another orthogonal durable item flag/string.

D2MOO item save/drop structures reuse item name storage for ear/personalized names depending on item type/flags.

Dark Magic's semantic item should distinguish:

```text
PersonalizedByName
EarOwnerName
RareNameParts
BaseLocalizedName
```

rather than overloading one string internally merely because the legacy file format does.

Personalization must survive trade, stash, sockets, repair, upgrades, save/load, and vendor rules as verified.

## Quantity and stackability

Stackable items carry mutable quantity. Quantity interacts with:

- item placement;
- purchase/sale;
- thrown-weapon use;
- replenishing quantity;
- Cube matching/consumption;
- save serialization;
- ground drops/pickups.

Stack operations should be atomic item-authority transactions. Avoid representing every arrow/potion/rune as an independent entity when the original item is one stack.

## Concrete special records should remain immutable data

A `UniqueItems` or `SetItems` row is a definition. A generated item instance references the row and stores its rolled property values.

Do not mutate catalog rows when:

- the item is identified;
- durability changes;
- set bonuses activate;
- a gem/rune is inserted;
- a runeword activates;
- a service personalizes or sockets the item.

Those are instance state/source layers.

## Quality fallback needs context

Different generation contexts can force or bypass normal quality election:

- TC entries can force concrete unique/set identities;
- quest items can force Normal/special treatment;
- direct item creation can force quality;
- vendor/gamble/Cube output can have their own rules;
- item types can force Normal/Magic/Unique-like results.

Record the `GenerationSource` with every item so debugging can explain why two identical base codes arrived at quality differently.

## Suggested implementation slices

### Q1 — richer item-instance quality state

Extend the item-generation recipe/durable item shape with:

- quality;
- concrete unique/set ID;
- selected affix IDs/name parts;
- identified;
- ethereal;
- durability/max durability;
- personalization;
- generation source/seed/ilvl.

Keep placement behavior unchanged.

### Q2 — special-record materialization

Turn the already selected Unique/Set record into actual immutable rolled property sources on the item instance.

### Q3 — conditional set sources

Integrate equipped-set counting with the shared stat-source model and prove add/remove behavior without duplicating permanent stats.

### Q4 — durability/ethereal lifecycle

Implement synthetic durability, broken/indestructible, replenish scheduling, and verified ethereal transformation.

### Q5 — identification/personalization services

Implement as atomic item-authority mutations with persistence/network snapshots.

## Verification backlog

1. Unique per-game drop occurrence/no-limit rules and concrete fallback.
2. Unique/set concrete-record weighted election and duplicate rows.
3. Set partial/full bonus thresholds and `addfunc` behavior.
4. Set-state/stat-list application and removal ordering.
5. Rare affix count, prefix/suffix balance, group exclusion, name generation.
6. Crafted output ilvl and affix count/eligibility formulas.
7. Magic affix-level formula including qlvl and magic level.
8. Automagic and staffmod ordering versus ordinary affixes.
9. Superior modifier election/ranges.
10. Low-quality modifier/name election and stat penalties.
11. Ethereal eligibility in every generation context, not only ordinary drops.
12. Ethereal damage/defense stat transformation in addition to durability.
13. Durability loss probability by item/use event.
14. Broken item effects/equipment behavior.
15. Replenishing durability/quantity tick math.
16. Throwing stack quantity/durability interaction.
17. Identification restrictions and Cain/vendor service behavior.
18. Personalization preservation across upgrades/socketing/runewords/trade.
19. Repair/recharge eligibility for ethereal, indestructible and charged items.
20. Save/network encoding of all quality-specific fields.

## Primary sources inspected

- Current Dark Magic M6 quality/eligibility/affix/property code and typed quality tables.
- D2MOO pinned 1.10f `Items.cpp`, `D2Items`, `ItemMods`, stat-list and save paths.
- D2MOO set/runeword state/stat-list evidence and per-game unique occurrence state.
