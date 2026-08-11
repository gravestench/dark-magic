# Horadric Cube and recipe-engine research

Status: implementation-oriented research baseline. Dark Magic already has an authoritative Cube container but does not yet have a typed CubeMain recipe runtime. D2MOO provides strong 1.10f evidence for the compiled recipe grammar and input matching; output execution still needs focused verification for several operation codes and transformation flags.

## Executive conclusion

The Cube is an **atomic item transformation engine** over the existing authoritative item archive/container system.

Conceptually:

```text
ContainerCube snapshot
+ player/difficulty/class/quest/session context
+ pinned CubeMain recipe catalog
        |
        v
recipe matcher
  enabled/version/ladder/difficulty/class/op gates
  multiset input matching
  item code/type/quality/quantity/grade/socket/ethereal/special/stat predicates
        |
        v
selected recipe
        |
        v
output planner
  create / copy / transform / upgrade / repair / recharge / socket / unsocket
  output ilvl/quality/affixes/properties
  side effects: stats/skills/portal/quest context where applicable
        |
        v
atomic transaction
  consume/modify exact input identities
  create exact output identities
  update placements/stat sources
  emit semantic event
```

Do not implement Cube recipes as Lua scripts that delete arbitrary items and mint replacements.

## Current Dark Magic foundation

Dark Magic already has:

- `ContainerCube` as an authoritative grid;
- stable item identity and placement;
- held-item and item-transaction authority;
- deterministic loot/property/affix helpers;
- service escrow and atomic service transactions;
- typed/generic table loading that can be extended for CubeMain;
- fixed-tick replay/checkpoint authority.

The recipe engine should consume those boundaries.

## CubeMain compiled structure

D2MOO's `HoradricCube.h` reconstructs a compiled row with:

- enabled flag;
- ladder flag;
- minimum difficulty;
- class restriction;
- operation code;
- operation parameter/value;
- number of inputs;
- version;
- up to seven input descriptors;
- up to three output descriptors.

Each input descriptor can carry:

- flags;
- item/item-type identity;
- special row identity;
- quality;
- quantity.

Each output descriptor can carry:

- output flags;
- base/special identity;
- quality;
- quantity;
- type code;
- level/percent-level/item-level parameters;
- up to three forced prefixes and suffixes;
- up to five property/output-mod descriptors with param/min/max/chance.

This is a real declarative transformation language, not a fixed list of hand-written named recipes.

## Input matching is multiset matching

D2MOO's `PLRTRADE_CheckCubeInput` iterates actual Cube-page inventory items and tracks which items have already satisfied input descriptors.

Architectural consequence:

- recipe inputs are a **multiset**, not ordered slots;
- one item identity must not satisfy two descriptors unless quantity/stack semantics explicitly allow it;
- matching must be deterministic when several items satisfy the same descriptor;
- stacked quantities need explicit accounting;
- recipe selection/order can matter if multiple rows match.

Use stable item ordering/IDs during matching. Never depend on Go map iteration order.

## Input predicate vocabulary

Recovered input flags/behavior cover at least:

- any item / exact base item;
- item type;
- no sockets;
- socketed;
- ethereal;
- non-ethereal;
- special concrete item;
- upgraded grade relationship;
- normal/exceptional/elite base grade;
- reject runewords;
- exact quality;
- quantity/stack count;
- special stat comparisons through operation codes;
- concrete file-index comparisons;
- quest-difficulty checks in a special operation path;
- output-specific preconditions such as requiring a socketable target.

Normalize these into explicit predicates with diagnostic failure reasons.

Example:

```text
InputPredicate
  Code/Type
  Quality
  Quantity
  Grade
  SocketState
  EtherealState
  RunewordState
  ConcreteSpecialID
  StatComparison
```

## Upgrade-grade matching

The input matcher can compare an authored base item against its normal/exceptional/elite code chain.

Do not derive upgrades from display names. Preserve normalized base-item relationships from Weapons/Armor/Misc data:

```text
NormalCode
ExceptionalCode
EliteCode
```

Upgrade outputs should explicitly decide which instance properties survive and which are recomputed.

## Stat-comparison operations

The 1.10f matcher contains operation families comparing an item's:

- effective stat value;
- base stat value;
- stat bonus;
- special file index;

against recipe parameters using ItemStatCost `ValShift` conversion.

This means the Cube engine needs access to the same canonical item-stat evaluator used by combat/equipment. Do not invent a separate stat parser just for recipes.

## Output flags

Recovered output flags include concepts for:

- copy mods;
- add/socket output;
- force ethereal;
- special identity;
- unsocket;
- remove;
- force normal/exceptional/elite grade;
- repair;
- recharge;
- plus some D2MOO custom flags that should **not** be assumed original behavior merely because they appear in that header.

Record provenance for every normalized output flag. D2MOO custom extensions must not silently become Diablo compatibility rules.

## Copy/use-item output

Cube output can reference/use an input item rather than always mint a brand-new base item. This is crucial for:

- upgrades;
- socketing;
- unsocketing;
- repair/recharge;
- rerolls that preserve selected fields;
- crafted transformations.

The output planner therefore needs operations over existing stable item IDs:

```text
KeepAndMutate(itemID)
Consume(itemID)
Create(newRecipe)
Duplicate/CopySelectedState(itemID)  // only where verified
```

Avoid implementing every recipe by delete-all-then-reroll because that destroys personalization, sockets, seed/identity, durability, or other fields that some outputs preserve.

## Output level context

Cube outputs can derive item level using authored level/player-level/item-level parameters.

Keep output level policy explicit:

```text
OutputLevelPolicy
  fixed level
  percent of player level
  percent of input item level
  combinations/clamps
```

This level feeds affix eligibility and quality materialization. Exact formula order/rounding needs verified vectors.

## Crafted items

Craft recipes should use the ordinary item-generation/property infrastructure with a distinct context:

```text
CraftedOutput
  forced crafted quality
  preserved/derived ilvl
  recipe fixed output properties
  rare-like random affixes
```

Do not create a second property evaluator.

## Reroll recipes

Magic/rare reroll recipes need explicit policies for:

- preserved base item identity;
- output item level;
- whether personalization is retained;
- whether sockets/ethereal/durability survive;
- new item seed/affix rolls;
- identified state.

Treat each output mode as a transformation policy derived from CubeMain/runtime evidence.

## Socket and unsocket recipes

These should call the socket subsystem through atomic item transformations.

Socket output:

- validate host socketability and current socket state;
- determine count using recipe policy/RNG;
- mutate capacity;
- preserve verified host fields.

Unsocket output:

- remove/destroy socketed children according to recipe semantics;
- remove filler/runeword stat sources;
- preserve host identity/other fields as verified.

## Repair and recharge

Cube repair/recharge outputs should reuse the same authoritative durability/charged-skill logic as vendor repair.

Do not independently set "durability = max" without accounting for quantity, ethereal restrictions, charged skills, and output flags.

## Special operations and portals

The recovered enum includes operations for Cow Portal, stat/skill points, red portals, and later-version Uber operations behind version conditionals.

These prove the Cube can produce **world/player side effects**, not only item outputs.

Architectural boundary:

```text
CubeTransactionResult
  ItemMutations
  CharacterMutations
  WorldTransition/PortalRequests
```

The Cube engine should emit semantic requests to player/world systems. It should not directly create renderer portal nodes.

Target-version support for Uber operations must be explicitly separated from 1.10f/LoD compatibility.

## Quest recipes

Quest Cube recipes should integrate with quest runtime through semantic events/conditions.

Do not use presence of a quest item alone as the full quest state. The same item can have difficulty/version ownership semantics and some quest recipes are gated by current quest bits/progression.

## Recipe selection order

Multiple CubeMain rows may be potentially matchable. Exact runtime selection order must be preserved, likely table order after enabled/version/gate filtering unless verified otherwise.

The recipe catalog must retain stable authored row order. Do not sort recipes alphabetically.

## Transaction safety

A transmute must be all-or-nothing.

Before commit:

- pin the item authority revision;
- match exact input IDs/quantities;
- generate all output plans and random results;
- validate all output placements/side effects;
- ensure consumed items are not simultaneously held/traded/service-locked.

Then commit one authoritative transaction.

On stale revision or validation failure, no item/gold/quest mutation occurs.

## Diagnostics/tooling

Add a headless inspector:

```text
dm cube match --character ... --items ...
dm cube explain <recipe-row>
```

or equivalent dev API exposing:

- eligible recipe rows;
- per-input match reasons;
- selected exact input identities;
- computed output levels;
- planned preserved/removed fields;
- RNG trace;
- side effects.

This will be invaluable for modded CubeMain data.

## Suggested implementation slices

### CUBE1 — typed/normalized CubeMain

Parse rows into immutable normalized predicates/output plans while preserving authored order and raw unsupported fields.

### CUBE2 — deterministic input matcher

Match one simple multi-item recipe against `ContainerCube`, including stack quantity and stable identity assignment.

### CUBE3 — simple create/consume transaction

Implement a rune/gem upgrade-style recipe using existing loot/item creation and atomic consumption.

### CUBE4 — use-item/copy transformation

Implement one item upgrade/reroll preserving verified fields.

### CUBE5 — socket/repair/recharge

Integrate socket and durability/charge subsystems rather than duplicating logic.

### CUBE6 — semantic world/player operations

Implement a synthetic portal/stat/skill side-effect recipe through trusted world/player requests, then verify real quest/special operations.

## Verification backlog

1. Full CubeMain input/output text grammar and parser edge cases.
2. Recipe row ordering and first-match behavior.
3. Stack quantity matching/consumption.
4. Ladder/version/min-difficulty/class gates.
5. Operation codes 1..N and target-version differences.
6. Stat/base-stat/bonus comparison operation semantics and rounding.
7. Output `lvl`/`plvl`/`ilvl` formula and clamp order.
8. Copy-mods exact field preservation.
9. Crafted output affix count and ilvl formula.
10. Magic/rare reroll seed and field-preservation behavior.
11. Normal/exceptional/elite upgrade preservation of ethereal, sockets, personalization, durability and requirements.
12. Cube socket count formulas.
13. Unsocket child destruction and runeword removal.
14. Repair/recharge rules and ethereal exclusions.
15. Output property chance/min/max rolling and property source provenance.
16. Cow Portal eligibility, location, quest/item consumption and failure modes.
17. Quest-specific Cube recipes and difficulty-item checks.
18. Uber-operation separation for later versions.
19. Network/replay behavior for transmute revision conflicts.
20. Legacy save field identity after transformations.

## Primary sources inspected

- D2MOO pinned 1.10f `DataTbls/HoradricCube.h/.cpp` and `PLAYER/PlrTrade.cpp` Cube matcher/transmute paths.
- Current Dark Magic item containers/authority/service framework and M6 generation/property systems.
