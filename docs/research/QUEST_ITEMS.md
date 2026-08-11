# Quest items and scripted item-interaction research

Status: implementation-oriented research baseline. Dark Magic already has quest runtime research, authoritative item identity/containers, semantic world objects, and server-owned quest/service escrow. This document joins them so quest items are not implemented as UI-only exceptions or anonymous quest bits.

## Executive conclusion

A quest item is both:

1. a real authoritative item instance with ordinary placement/identity/save semantics; and
2. a participant in quest/script rules that can restrict generation, possession, pickup, drop, consumption, difficulty reuse, and world progression.

Use explicit semantic events/conditions:

```text
QuestItemGenerated
QuestItemPickedUp
QuestItemDropped
QuestItemMovedToService
QuestItemConsumed
QuestItemTransformed
QuestItemLost/Recovered
```

Quest runtime consumes these events and authoritative item queries. Do not let Lua infer quest completion from tooltip text or delete quest items directly.

## Current Dark Magic foundation

Relevant existing owners:

- `internal/game/item`: stable item identity and all player/world/service placements;
- `ContainerQuest`: named quest/service escrow sockets;
- `ServiceRule`: server-owned target/material consumption with atomic commit;
- quest runtime/controller research: quest bits, events, party credit, world gates;
- semantic DS1/world-object materialization;
- fixed-tick session/replay/checkpoint;
- persistence boundary for durable character/quest state.

This is a strong basis for real quest-item transactions.

## Quest flag versus quest item

Do not collapse these into one concept.

Examples of independent states:

```text
quest bit says objective started
player does not currently possess item
world object still can generate item

player possesses item
quest completion bit not yet set

quest completed in this difficulty
old item from another difficulty may exist in imported save
```

Quest logic must query both quest state and item state as required.

## Quest items participate in item generation differently

D2MOO's item-quality path forces quest items to Normal in the shown generic generation context and excludes quest items from ordinary ethereal generation.

Ground cleanup also exempts quest items from automatic removal in the observed 1.10f path.

These are strong examples of generation/lifecycle policy keyed by quest-item status.

Do not merely mark them "cannot sell" after ordinary random item generation.

## Difficulty provenance

D2MOO's quest-item search logic references quest item difficulty checks and a quest-item difficulty stat. Some Cube input logic also tests quest item difficulty relative to the current game difficulty.

A durable quest item may therefore need explicit difficulty provenance/source semantics.

Possible normalized state:

```text
QuestItemState
  quest identity
  generated difficulty
  generation/source event
  special concrete ID if needed
```

Do not make this visible gameplay state dependent on which save codec encoded a numeric stat.

## Finding quest items

The original runtime contains a helper that can search the player's cursor/inventory for an item code while considering quest flags/difficulty and inventory page.

Dark Magic should provide semantic item queries to quest logic:

```text
HasQuestItem(owner, questItemKind, policy)
FindQuestItems(...)
```

The policy can specify allowed containers:

- held;
- inventory;
- Cube;
- stash;
- equipment;
- corpse/service escrow;
- owned unit.

Exact Diablo rules should be encoded by the quest/item rule, not by hard-coding "quest item means search inventory only" globally.

## Pickup eligibility

Some quest items are globally unique per quest state, party-aware, or restricted if a player already has/has completed the item objective.

Pickup should be an authoritative item transaction with a quest eligibility hook:

```text
world item + actor
 -> ordinary pickup capacity/placement validation
 -> quest-item eligibility
 -> commit placement
 -> QuestItemPickedUp event
```

If eligibility fails, the item remains in world state.

The client should not decide whether it can pick up a quest item.

## Drop eligibility

D2MOO calls a quest hook when an item is dropped to the ground.

That is strong architectural evidence that quest runtime observes item movement rather than only NPC dialogue.

Dark Magic should emit item lifecycle events after atomic placement changes so quest controllers can respond to:

- drop;
- pickup;
- consume;
- sell attempt;
- transform;
- player death/corpse;
- zone transition if relevant.

## Quest-service escrow

Dark Magic's `ContainerQuest` plus `ServiceRule` is a useful generalized model for interactions such as:

- give item(s) to NPC;
- use materials on a quest target;
- socket/personalize/reward item service;
- assemble quest components;
- operate an altar/device with items.

The UI moves named item identities into authority-owned semantic slots; the service rule validates all required identities/quest conditions and commits atomically.

For real quest implementations, extend `ServiceRule` with quest predicates/results rather than moving logic into Lua.

## Quest Cube recipes

Quest items in the Horadric Cube should be handled by CubeMain recipe matching plus quest/difficulty conditions.

The Cube engine emits `QuestItemConsumed/Transformed` and world/player operation requests. Quest runtime updates progression from those semantic results.

Do not implement quest transmutation as a separate hidden recipe engine.

## World-object item interactions

Some quest items operate on specific world objects/set pieces.

Recommended flow:

```text
player interacts with object using semantic item/service intent
 -> server validates object identity/state/range
 -> server validates required item identity/container/quest state
 -> atomic item consumption + object state transition
 -> quest event
 -> presentation/audio cue
```

The object animation/sound occurs after the authoritative transition.

## Quest-item generation sources

Possible source categories:

- fixed DS1/world object;
- monster death;
- special chest/object;
- NPC reward;
- Cube transformation;
- quest script/encounter;
- player body part/ear-like special item where relevant to nonquest systems.

Store generation source so duplicates/recovery/debugging can be explained.

## Duplicate prevention and recovery

Quest systems need explicit duplicate/recovery policy:

```text
MayExistPerPlayer
MayExistPerParty
MayExistPerGame
MayRegenerateIfLost
MayDropAgainAfterCompletion
```

Do not rely solely on searching current inventory: the item could be on ground, in Cube/stash, held, corpse, or another party member depending on the quest.

The exact policy is quest-specific.

## Party credit

Picking up or consuming a quest item may grant quest credit to:

- only the actor;
- nearby/same-level party members;
- all eligible players in the game;
- no one until a later NPC/object event.

Quest runtime already needs party-credit semantics. Item events should carry enough context:

```text
actor
item ID/kind
zone/level
party snapshot/eligible players
source object/monster/recipe
```

## Selling and vendor restrictions

Quest item vendor behavior must be explicit.

Potential rules:

- cannot sell;
- can sell but triggers recovery/quest behavior;
- special quest vendor accepts only through service/dialogue path;
- item value is irrelevant/zero.

Do not treat `BaseCost == 0` as the sole "quest item" restriction.

## Player death and quest items

Death/corpse behavior may differ for quest items.

The death transaction must have an item policy capable of preserving/moving/dropping special items according to verified rules rather than blindly treating all inventory items alike.

This needs owned-game probes for representative quest items.

## Save/exit and difficulty changes

Durable quest item state and quest bits can become inconsistent in imported/modded/corrupt saves.

Import should:

- preserve the item identity/data;
- preserve quest bits independently;
- validate/report contradictions;
- avoid silently deleting data unless a compatibility repair policy explicitly says so.

Difficulty-specific quest items must not accidentally satisfy a new-difficulty quest if original rules reject them.

## Quest items and ground lifetime

Pinned 1.10f evidence shows quest items return no automatic ground-removal time in the inspected path.

Dark Magic should preserve an explicit `GroundLifetimePolicy` from generation/item type. Do not special-case cleanup code by display name.

## Presentation/audio

Quest item visuals/audio are consumers of semantic item/quest events.

Examples:

- quest-item drop/pickup sound;
- object accepts item;
- transmutation success;
- quest reward cue;
- world set-piece changes.

See [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md). Audio must not decide whether item consumption succeeded.

## Suggested implementation slices

### QI1 — quest-item metadata/policy

Add semantic quest-item identity/difficulty/source metadata to item instances or an attached component, without changing ordinary placement.

### QI2 — item lifecycle event stream

Emit stable pickup/drop/consume/transform events from item authority after commits.

### QI3 — quest-aware pickup/service rule

Implement one synthetic quest item whose pickup and NPC service consumption are gated by quest state and replay-safe.

### QI4 — real Act I quest-item vertical slice

Choose one well-bounded item/object/NPC quest interaction and connect generation -> pickup -> quest state -> service/consume -> persistence.

### QI5 — recovery/duplicate policy

Add game/player/party duplicate/recovery rules with deterministic world-item search.

## Verification backlog

1. Quest-item quality/ethereal/socket/vendor restrictions by target version.
2. Quest item difficulty stat generation/check behavior.
3. Which containers count for each quest's possession checks.
4. Pickup denial/duplicate rules for key Act I-V quest items.
5. Regeneration/recovery when a quest item is dropped/lost/removed.
6. Ground cleanup exemption and inactive-room persistence.
7. Drop hooks and quest-state changes on manual/death drops.
8. Party credit rules for pickup/consume/object use.
9. Quest-item vendor/sell/buyback behavior.
10. Quest Cube recipe difficulty/progression gates.
11. World-object use item consumption/return behavior.
12. NPC hand-in service sequencing and dialogue quest bits.
13. Quest items on player corpse/death and save/exit.
14. Imported old-difficulty items in a new difficulty.
15. Multiplayer duplicate items across party members/game state.
16. Quest reward item generation context/quality/seed/placement.
17. Exact item callbacks used by representative quest scripts.
18. Audio/presentation cues tied to semantic quest-item events.

## Primary sources inspected

- Current Dark Magic item authority, `ContainerQuest`, service transactions, quest runtime research and semantic world interactions.
- D2MOO pinned 1.10f `ITEMS_FindQuestItem`, item generation/drop hooks, Cube quest-item checks, and quest runtime call sites.
