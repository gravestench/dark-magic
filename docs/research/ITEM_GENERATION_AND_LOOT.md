# Item generation, treasure classes, drops, and ground-item lifecycle

Status: implementation-oriented research baseline. Dark Magic has a substantial
deterministic M6 loot subsystem. Monster death now supplies distinct actual,
effective, credited-owner living same-level party, and pinned monster counts to
NoDrop and records that context on durable death facts. The exact expansion
1.14d proximity boundary remains a probe. This document records what is strong,
what differs from recovered behavior, and which missing semantics should become
bounded follow-up work rather than a replacement loot engine.

Read with:

- [ITEM_SYSTEM.md](ITEM_SYSTEM.md)
- [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md)
- [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md)

## Executive conclusion

Item generation should be treated as several explicit stages:

```text
LootEvent
  source / killer / owner / area / difficulty / players / seed context
        |
        v
TreasureClass resolution
  recursive picks / NoDrop / nested classes / forced entries
        |
        v
Base item selection
  code / type family / level eligibility
        |
        v
Generation-context quality policy
  monster TC / direct creation / vendor / gamble / Cube / quest / reward
        |
        v
Concrete quality record
  unique/set/rare/magic/etc.
        |
        v
Item-instance materialization
  seed + ilvl + affixes + stats + sockets + ethereal + quantity + durability
        |
        v
Authoritative placement
  world / inventory / vendor / quest-service / etc.
        |
        v
Ground lifetime / ownership / pickup policy
```

The important new conclusion is that **there is not one universal quality-roll entry point**. Diablo II uses different item-creation contexts with forced flags and quality behavior. Dark Magic should keep those contexts explicit.

## What Dark Magic M6 already provides

`internal/game/loot` already has strong, renderer-independent pieces:

- recursive TreasureClass rolls;
- positive Picks, NoDrop, negative Picks, and nested classes;
- cycle/runaway-depth rejection;
- terminal item codes plus traversed TC path;
- TreasureClass quality modifiers;
- ItemType expansion by level;
- Classic/Expansion ItemRatio selection;
- quality chance calculation with level, MF, minimum, and TC adjustments;
- unique/set availability fallback and type restrictions;
- concrete unique/set record election by eligibility/rarity;
- magic/rare affix selection;
- property-value materialization;
- many Properties/ItemStatCost functions;
- ladder eligibility;
- deterministic monster/chest event seeding;
- dedicated loot/quality test applications.

This should remain the basis of the implementation.

### Important Dark Magic policy distinction

The current loot roller uses `splitMix64`. This is deliberate Dark Magic deterministic policy, not Diablo II seed compatibility.

Do not change that casually while adding gameplay completeness. If exact original seed/call-order reproduction becomes a goal for a particular subsystem, make that an explicit compatibility mode or verified migration decision.

## TreasureClassEx runtime evidence

D2MOO's pinned 1.10f `D2GAME_DropTC_6FC51360` reconstructs a stack-based traversal of nested TreasureClassEx records.

### Picks

For the active TC, the runtime tracks an absolute pick count. Negative Picks behave as authored deterministic-entry counts rather than ordinary weighted random picks; positive Picks perform random selection.

Dark Magic already models this distinction.

### Nested classes propagate quality modifiers

Nested TC records carry Magic/Rare/Set/Unique/Superior/Normal modifiers. During nested traversal the observed runtime combines parent and child values by retaining a nonzero parent unless the child's value is larger.

This is more specific than simply accumulating every modifier additively.

Dark Magic currently retains the TC path/modifier chain. Before changing combination semantics, add direct vectors proving the current calculation matches the intended nested rule.

## NoDrop and effective player count

The 1.10f TC path provides strong evidence for multiplayer NoDrop scaling.

Observed inputs include:

- total player count in the game;
- living party members in the same level as the credited player/owner, capped at 8;
- minion ownership, so a pet kill can resolve to its controlling player;
- a monster-side player-count stat that can cap the effective count.

The effective count is reconstructed as:

```text
partyMembers = max(1, same-level living party members)
effectivePlayers = (totalPlayers - partyMembers) / 2 + partyMembers
```

with further monster-side capping in the observed path.

The original then adjusts NoDrop by exponentiating the base NoDrop probability by effective player count and converting it back into a weight.

This should become a dedicated, unit-tested policy function rather than being buried inside TC traversal.

Implementation status: the dedicated policy and monster-death consumer are
active. Ultimate owned-unit attribution selects the credited player, whose
living same-level party members contribute to the count; the monster's pinned
spawn count caps later benefits. This is high-confidence recovered structure,
not yet a verified 1.14d proximity claim.

### Follow-up implications

Dark Magic needs authoritative access to:

- game player count;
- party membership;
- same-level/living status;
- credited source owner;
- monster player-count override/stat.

Do not accept a client-supplied `/players X` value as a trusted drop result. It must be part of server/session state.

## TC variant selection by level

The 1.10f wrapper can select a TreasureClassEx record by TC ID plus an effective monster level, subject to expansion/difficulty and monster flags such as no-ratio/boss behavior.

This supports Dark Magic's existing data-driven TC catalog but highlights a missing integration point: **monster death must pass the verified effective monster/drop level into loot**, not just a TC name and seed.

## Quality is generation-context dependent

This research found an important distinction.

### TC drop path

The expansion TreasureClassEx drop path performs MF/TC-adjusted quality checks in this order:

```text
Unique
Set
Rare
Magic
Superior
Normal
Inferior fallback
```

This matches the broad ordering in Dark Magic's current M6 `RollQuality`.

The TC path also:

- honors item-type flags that force Normal or Unique-like behavior;
- uses level difference against base item level;
- uses ItemRatio divisors/minimums;
- uses TreasureClass quality modifiers in 1024ths;
- obtains Magic Find from the credited unit and, for a minion, adds owner Magic Find;
- uses diminishing-return transformations for Unique/Set/Rare while Magic uses the direct MF divisor family;
- rolls against a `128` success scale.

### Generic/direct item creation path

`ITEMS_RollItemQuality` separately checks:

```text
Unique
Rare
Set
Magic
Superior
Normal
Inferior fallback
```

and can be bypassed by forced quality in the item-drop creation descriptor.

It also immediately forces quest items to Normal in the observed path.

### Architecture consequence

Do not expose one function called `RollQuality(context)` without preserving the generation source.

Use a semantic origin/policy such as:

```text
QualityContextKind
  TreasureClassDrop
  DirectDrop
  VendorStock
  Gamble
  CubeOutput
  QuestItem
  QuestReward
  StartingItem
  ScriptedService
```

These may share low-level ItemRatio helpers but can have different order, forced flags, level source, eligibility, and fallback behavior.

## Magic Find ownership

In the TC path, Magic Find can include both a minion and its owner.

This is an important dependency on the owned-unit model:

```text
immediate killer/source
        +
ultimate player owner if applicable
        -> effective MF
```

Do not erase immediate source identity after resolving the owner; proc/kill semantics may still care whether a pet, trap, hireling, or player performed the kill.

## Gold Find

The same TC drop path applies item gold bonus after a gold item is created and similarly includes owner contribution for minions.

Gold is represented as an item instance with a quantity/stat and uses ordinary world drop placement/lifetime behavior.

Dark Magic's item authority already separates gold balances from items. Future world gold pickup/drop should preserve this source policy and avoid client-side gold arithmetic.

## Unique occurrence tracking

D2MOO maintains per-game unique-drop flags.

Observed behavior:

- a UniqueItems row may carry a no-limit flag;
- otherwise game state can mark that unique record as already dropped;
- quest/special cases participate differently;
- a later Unique result can therefore fail concrete unique election and require fallback behavior.

This state affects future loot rolls and must be part of the authoritative session/checkpoint if Dark Magic reproduces the behavior.

Do not keep it in a renderer or process-global singleton.

## Item creation descriptor

The recovered `D2ItemDropStrc` demonstrates how much context can alter item materialization:

- source unit and RNG seed;
- item level;
- item/base ID;
- ground versus inventory spawn type;
- x/y/room;
- item format/version;
- force flag;
- forced quality;
- quantity;
- min/max durability;
- special item row/index;
- item flags;
- forced seed and item seed;
- ear name/level;
- up to three prefix/suffix overrides;
- additional generation flags.

Dark Magic should not copy this struct, but it needs an explicit **ItemGenerationRequest** capable of representing forced/reward/Cube/vendor contexts without smuggling special cases through global variables.

## Item seed versus event seed

The original item has both a start/item seed and a mutable item RNG seed. Item-generation helpers reset/derive those fields as part of creation.

This is a strong reason to retain stable per-item generation seed metadata even if Dark Magic continues using its own deterministic RNG algorithm.

Useful fields:

```text
GenerationID
GenerationSource
EventSeed
ItemSeed
ItemLevel
GenerationTraceHash
```

The exact runtime RNG state need not remain mutable forever after all instance properties have been materialized, but preserving origin makes debugging and save verification much easier.

## Random sockets during ordinary generation

The observed 1.10f item creation path can add sockets after ordinary special-item setup.

For a suitable socketable, nonstackable item:

- difficulty caps the maximum in this path to 3 Normal, 4 Nightmare, 6 Hell;
- flags can suppress or force socket creation;
- the default chance observed is 33%;
- newer-format socket count can be derived from item start seed modulo maximum plus one;
- older-format logic has additional caps.

This is **random drop-generation socketing**, not Larzuk, Cube socket recipes, or intrinsic socket count. Those contexts need separate policies.

## Ethereal generation

The observed item-generation path:

- limits ethereal generation to weapons/armor with durability;
- rejects inferior, set, and quest items in this path;
- supports explicit never/always ethereal flags;
- otherwise uses a 5% item-seed roll;
- applies ethereal stat changes;
- halves max durability with integer `max/2 + 1` and fills current durability to the new max.

This is strong 1.10f evidence suitable for a focused vector test.

## Charms receive special affix treatment

Charms are routed through a dedicated charm-affix function rather than generic item handling. D2MOO marks a local part of the recovered function as TODO/uncertain, so do not mechanically translate it.

The reliable architectural fact is that **item type can select a specialized materialization policy**.

## Ground placement is authoritative

The original drop path searches for free coordinates near the source unit using collision, then creates a targetable on-ground item with world collision/path state.

Dark Magic should integrate generated item instances with `ContainerWorld` and semantic world targeting. A loot roll is not complete merely because it returned an item code.

Needed transaction:

```text
loot materialization
 -> find legal world drop coordinate
 -> atomically create authoritative item identity + placement
 -> emit ItemDropped event
 -> presentation discovers snapshot
```

## Ground-item cleanup time

D2MOO exposes explicit ground removal times in game frames.

Observed 1.10f examples:

- quest items: no automatic ground removal;
- high qualities in a defined range: longer lifetime (45000 frames);
- gold <= 10000: 15000 frames;
- larger gold: 45000 frames;
- ordinary non-socket-fillers: 15000 frames;
- socket fillers: 30000 frames.

At 25 game frames/sec these correspond to long real durations, but store them as **simulation ticks**, not wall-clock seconds.

Before exact implementation, verify patch behavior and whether interactions reset action stamps.

## Durability and replenishment are later instance lifecycle

The original item system rolls durability loss from unit RNG and can schedule replenishing durability/quantity events from item stats.

This is outside initial loot materialization but belongs to the same authoritative item instance.

Do not regenerate a new item when durability changes.

## Recommended Dark Magic boundaries

Keep `internal/game/loot` focused on deterministic selection/materialization helpers, while `internal/game/item` owns resulting identity and location.

Possible high-level boundary:

```text
loot.Generate(request) -> GeneratedItemRecipe(s)
item.Authority.Create(recipe, placement) -> stable authoritative item(s)
```

`GeneratedItemRecipe` should contain all rolled instance facts, not a pointer back to mutable loot RNG/catalog state.

## Suggested implementation slices

### L1 — generation-context vocabulary

Add explicit source/context metadata to generated item recipes and diagnostics without changing current M6 outcomes.

### L2 — effective player/party NoDrop policy

Implemented as high-confidence recovered structure behind explicit policy and
multiplayer death-path tests, using authoritative session/party/owned-unit state
rather than UI values. Retain the 1.14d proximity probe before marking fidelity
verified.

### L3 — TC quality vector corpus

Generate synthetic and owned-data vectors comparing:

- Dark Magic current chance calculation;
- pinned 1.10f TC path;
- concrete unique/set/rare/magic selection;
- owner/minion MF.

Fix only confirmed differences.

### L4 — item-authority materialization

Connect a loot result to stable authoritative item identity in `ContainerWorld`, legal drop position, ground lifetime, pickup, and replay/checkpoint state.

### L5 — per-game unique occurrence state

Add unique-drop occurrence state if the chosen compatibility target requires it, including checkpoint/replay coverage and fallback tests.

## Verification backlog

1. Exact TCEx nested-class level-up/selection behavior for all level fields.
2. Negative Picks semantics with quantity/weights across real tables.
3. NoDrop player/party formula across 1..8+ players and minion kills.
4. Monster `STAT_MONSTER_PLAYERCOUNT` source and behavior.
5. TC quality-modifier combination across deep nesting.
6. Full MF formulas and integer rounding for Unique/Set/Rare/Magic.
7. Effective MF for hirelings, summons, traps, and party kill attribution.
8. Gold-find owner attribution and rounding.
9. Generic/direct quality path contexts that use Unique->Rare->Set rather than TC ordering.
10. Unique already-dropped behavior, no-limit flag, quest exceptions, and fallback quality.
11. Set-record election/fallback when no eligible set member exists.
12. Rare/magic affix count/order and automagic/staffmod integration after quality election.
13. Random socket chance/count and interaction with base maximum/ilvl/difficulty.
14. Ethereal eligibility and exact stat/durability effects.
15. Ground drop free-coordinate search and collision footprint.
16. Ground lifetime by quality/type/value and whether pickup/drop refreshes it.
17. Drop ownership/visibility rules in multiplayer.
18. Boss/champion/superunique/quest first-kill treasure-class differences.
19. Countess multi-TC behavior and special object/chest drops.
20. Rack/chest/object-specific item-level/base selection.
21. Gambling/vendor/Cube/quest generation paths and their quality policies.
22. Exact original RNG stream and call order if seed-compatible loot becomes a project goal.

## Primary sources inspected

- Current Dark Magic `internal/game/loot/*`, M6 roadmap, and authoritative item state.
- D2MOO pinned 1.10f `source/D2Game/src/ITEMS/Items.cpp` and `include/ITEMS/Items.h`.
- D2MOO item/stat/item-type/data-table helpers reached from TreasureClass and item creation.
- libd2 verification methodology for item-roll traces against the binary.
