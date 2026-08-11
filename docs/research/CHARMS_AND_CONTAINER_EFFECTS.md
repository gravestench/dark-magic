# Charms and inventory/container-dependent item effects research

Status: implementation-oriented research baseline. Dark Magic already owns item placement across inventory, stash, Cube, equipment, belt, held, vendor, quest-service and hireling containers. The missing layer is conditional stat-source activation based on **where the item currently is** and whether the item is usable.

## Executive conclusion

Container location is gameplay state.

Some items contribute effects only in specific containers/states. Charms are the clearest example: their rolled properties exist on the item at all times, but their contribution to the player is conditional.

Use a source-activation rule:

```text
ItemInstance
  rolled property sources
  placement
        |
        v
ContainerEffectPolicy
  does this source contribute to this owner now?
        |
        v
EffectiveStatSources
```

Do not copy charm stats permanently into the character when the charm is picked up.

## Current Dark Magic foundation

`internal/game/item` already has canonical container placement:

- `world`
- `inventory`
- `stash`
- `cube`
- `equipment`
- `hireling`
- `belt`
- `held`
- `vendor`
- `quest_service`

This is exactly the information the future conditional effect evaluator needs.

No new UI-owned inventory state is required.

## Charm usability is an explicit runtime concept

D2MOO exposes `ITEMS_IsCharmUsable` and item-move/equip paths conditionally add/remove item stat lists based on that predicate.

That is strong evidence for the architecture even where exact conditions still need a trace:

- charm properties are ordinary item stat sources;
- moving the item can activate/deactivate those sources;
- broken/no-equip flags participate in usability;
- item movement invokes stat-list update logic rather than rerolling the charm.

The familiar LoD behavior that charms operate from player inventory but not stash/Cube should be promoted only after the exact container/page checks are captured for the targeted runtime.

## Generalize beyond charms

Build a reusable conditional-source mechanism rather than `if item.Type == charm` scattered through inventory commands.

Potential conditions include:

```text
ActiveWhenEquipped
ActiveWhenInInventory
ActiveWhenInBelt
ActiveWhenWeaponSetActive
ActiveWhenSocketed
ActiveWhenSetThresholdMet
ActiveWhenRunewordRecognized
ActiveWhenNotBroken
ActiveWhenOwnerMeetsRequirements
```

These conditions can compose.

Example:

```text
weapon property source
  requires Equipment + active weapon set + not broken + requirements met

charm source
  requires Inventory + charm-usable + requirements if applicable

socketed jewel source
  requires parent host source active
```

## Active weapon set

Dark Magic already tracks `ActiveWeaponSet`. Equipment effects from alternate weapon slots must become inactive without moving/destroying the item.

This belongs in the same conditional-source mechanism as charms.

Do not physically transfer inactive weapon-set items out of equipment just to disable their stats.

## Broken and unusable equipment

D2MOO item stat-list callbacks distinguish broken/no-equip/requirements and actively add/remove transferred properties.

A broken or requirements-invalid equipped item may remain in its equipment slot while some/all gameplay contributions are disabled.

Therefore placement and contribution are separate states:

```text
Placement = Equipment/RHand
ContributionActive = false
Reason = Broken | Requirements | InactiveWeaponSet | ...
```

This distinction is essential for UI red tinting, repair, save compatibility, and deterministic stat recomputation.

## Item-granted skills and auras

The item stat-list callback evidence shows that item sources can alter more than scalar stats.

Observed families include:

- single/non-class skills;
- +class skills;
- skill tabs;
- all skills;
- item aura;
- charged skills;
- explicit state toggles;
- item event registration.

When a source activates/deactivates, the skill/state/event system must react.

This strongly argues for one stat/effect-source activation event rather than recomputing only numeric character-sheet totals.

## Resource maximum changes

D2MOO's item stat-list callback adjusts current HP/mana/stamina proportionally when max values change in certain cases.

That means enabling/disabling a container/equipment source can have immediate authoritative resource consequences.

Dark Magic should not leave max-resource source toggling as a pure read-only derived stat if the original behavior mutates current resources on transition.

Exact rounding needs vectors before implementation.

## Requirements and self-dependency

Equipment requirements can depend on stats granted by equipment, creating ordering/self-dependency edge cases.

A robust evaluator should:

1. build candidate source activation from placement;
2. evaluate requirements under a stable rule;
3. deactivate unusable sources;
4. repeat only if the original semantics require dependency convergence;
5. reject cycles/oscillation deterministically if mods create pathological data.

Do not rely on Go map traversal order to decide which strength-granting item keeps another equipped item valid.

The exact Diablo II requirement-resolution order needs empirical tests.

## Charm-specific generation

D2MOO routes charms through a specialized affix assignment path. Dark Magic's M6 affix machinery should use a generation-context/type policy for charms rather than create a separate persistent item class.

Needed data/probes include:

- small/large/grand charm affix pools and levels;
- whether magic quality is forced by item type/generation path;
- prefix/suffix count and failure behavior;
- unique charms and per-character possession restrictions;
- charm graphics/name/type relationships.

## Unique charm possession rules

Certain unique charms have special possession/pickup restrictions in LoD-era gameplay. Those rules are not ordinary affix eligibility.

Model them as an authoritative inventory admission policy keyed by concrete unique identity and owner/session state.

Do not hide them in UI drag/drop logic.

Exact restrictions for Annihilus, Hellfire Torch and Gheed's Fortune vary by version/content era and need target-version probes.

## Container transitions are transactions

For any item move:

```text
validate destination
compute old active effect sources
compute new active effect sources
validate downstream consequences
commit placement
apply source delta
emit semantic item/stat events
```

Replay/checkpoint must reproduce both placement and effect state.

A simple implementation can recompute all item sources for the owner after each item transaction; optimize incrementally only if profiling justifies it.

## Stash and Cube

Stash/Cube contents remain durable item state but should not automatically contribute inventory-only effects.

The Cube is also a recipe workspace. During transmutation, its input items should not temporarily become active character sources simply because they are being inspected by recipe matching.

## Held item

Dark Magic correctly models the cursor/held item as authoritative `ContainerHeld`. Held items should have explicitly verified contribution rules rather than inheriting whatever container they came from.

This matters during:

- equipment swaps;
- trade/service escrow;
- death/disconnect/save;
- weapon moves;
- charm moves.

## Suggested implementation slices

### CCE1 — source activation evaluator

Given owner + item placement/state, return a stable list of active item source IDs and diagnostic reasons for inactive sources.

Start with equipment active/inactive weapon set and inventory-only charm source.

### CCE2 — atomic source delta on item commands

Integrate the evaluator with move/equip/swap/held transactions and replay/checkpoint state.

### CCE3 — skill/aura/item-event sources

When an item source toggles, add/remove granted skills, charged-skill availability, auras and registered item events through their owning systems.

### CCE4 — requirement/broken-state behavior

Add requirements and broken/no-equip source suppression with deterministic dependency tests.

## Verification backlog

1. Exact `ITEMS_IsCharmUsable` conditions and inventory-page checks in 1.10f.
2. Charm behavior in inventory, Cube, stash, held cursor, corpse, trade and vendor containers.
3. Unique charm possession/drop/pickup rules by target version.
4. Charm affix-level/pool/count behavior.
5. Requirement dependency ordering between equipped items.
6. Broken/no-equip source removal and reactivation.
7. Active/inactive weapon-set stat, skill, aura and proc behavior.
8. Current HP/mana/stamina adjustment when max values change from item source activation.
9. Item aura activation/deactivation timing and ownership.
10. Charged-skill availability/charge state when item moves inactive.
11. Item event registration/unregistration ordering.
12. Socket filler/runeword source dependency on host activity.
13. Set partial/full bonus dependency on active equipped pieces.
14. Multiplayer trade/held/disconnect effect removal timing.

## Primary sources inspected

- Current Dark Magic authoritative item containers/commands/state.
- D2MOO pinned 1.10f `ItemMode.cpp`, D2Common item usability/stat-list code, item events/skill/aura callbacks.
- Dark Magic M6 charm/affix/property foundations.
