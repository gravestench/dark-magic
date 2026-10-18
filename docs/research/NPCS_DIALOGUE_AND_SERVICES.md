# NPCs, dialogue, gossip, services, and town-state research

Status: implementation-oriented research baseline. Dark Magic already has semantic NPC interaction targets, authoritative vendor/service capability checks, recovered quest/dialogue mappings, localization, item commerce/service authority, hireling research, and gameplay audio support. The missing layer is a general NPC dialogue/service controller with quest-aware availability and persistent per-player town state.

## Executive conclusion

An NPC is an authoritative world actor plus a **capability/dialogue projection** for each player.

Do not make the visible menu the source of truth.

```text
NPCDefinition/Instance
  identity, world position, services, vendor/hireling role, speech data
        |
        v
player interacts
        |
        v
NPC Interaction Context
  current NPC
  range/LOS still valid
  quest/difficulty/player state snapshot
        |
        v
server derives available dialogue/capabilities
  talk/gossip
  quest speech
  trade/gamble
  repair/identify/heal
  hire/resurrect
  quest/service actions
        |
        v
client renders menu/dialogue
        |
        v
semantic choice command
        |
        v
server revalidates and commits downstream transaction
```

## Current Dark Magic interaction authority

`internal/game/interaction` already has several excellent security/architecture properties:

- presentation requests/selects a server-known target;
- target registry is authoritative;
- world selection is rebuilt when the active world changes;
- current range is revalidated for every vendor/service transaction;
- vendor/category/service capabilities come from the target, not client-provided arbitrary names;
- interaction context can be closed/reset;
- UI panel state is not trusted as proof the player is still near the NPC.

Generalize this system rather than introduce a separate NPC panel authority.

## NPC target model

Current targets store `NPC`, optional vendor, categories and services. Future NPC definition/instance should distinguish:

```text
NPCDefinition
  monster/NPC class identity
  dialogue identity
  vendor identity
  hireling/service capabilities
  quest speech hooks
  gossip pool
  presentation/audio roles

NPCInstance
  stable entity ID
  level/position
  active state
  quest/event-specific runtime state
```

Interaction context references the stable instance.

## Dialogue is derived from player + NPC + quest state

The same NPC can show different lines/options depending on:

- quest state;
- difficulty;
- character class;
- whether the player has heard/introduced state;
- quest items;
- party/shared progression;
- reward availability;
- service unlocks;
- game event state.

A dialogue query should be deterministic and side-effect free:

```text
DialogueSnapshot(player, npc, gameState) -> ordered options/lines
```

Choosing an option can then produce a transaction/event.

Do not mutate quest state merely because the dialogue UI was opened unless the original interaction itself triggers a quest event.

## Speech versus text

Keep semantic dialogue line identity separate from:

- localized text;
- localized speech sound;
- portrait/presentation;
- duration/subtitle timing.

Quest/dialogue logic selects a speech-line ID. Localization resolves text; audio resolves the corresponding speech record/asset.

This supports headless server logic and language packs cleanly.

## Recovered Riiablo dialogue data

Dark Magic already carries recovered declarative quest/speech mappings with pinned provenance. Those records are useful for:

- which NPC can speak a line;
- quest/dialogue relationships;
- ordering/activation leads.

They should remain declarative evidence, not become the sole quest controller.

Runtime quest state must still determine eligibility.

## D2MOO NPC architecture evidence

Pinned 1.10f `SUnitNpc` includes or calls behavior for:

- vendor initialization/fill/cache;
- gambling inventory;
- buy/sell;
- repair;
- identify all / identify purchased item;
- heal/purchase heal;
- hire and resurrect mercenary;
- NPC dialogue message handling;
- interaction reset;
- NPC event lists and vendor chains;
- per-NPC trade state and tick/refresh information.

This supports a general NPC service dispatch layer rather than UI-specific commands.

## Service capability model

Normalize capabilities:

```text
Talk
Trade
Gamble
Repair
Identify
Heal
HireMercenary
ResurrectMercenary
QuestService(serviceID)
OpenStash   // object rather than NPC in classic case, but same interaction UI family
```

Capability availability can have a predicate.

The current `interaction.Target.Services` can evolve into a typed capability set without abandoning the range-revalidation pattern.

## Dialogue choices should be semantic IDs

Client command should submit something like:

```text
NPCChoice
  target NPC instance ID
  dialogue/capability choice ID
  optional referenced item/mercenary/etc. stable ID
```

Server re-resolves the choice under current state.

Never accept a client command such as `give_me_quest_reward = true` or a raw price/result.

## Gossip

Gossip is repeatable NPC dialogue outside major quest transitions.

A gossip controller may select from authored candidates based on:

- NPC;
- act/quest progression;
- player introduction/heard bits;
- deterministic/random policy.

If original behavior tracks which lines were heard, keep that as semantic NPC/player state. The `.d2s` research identifies NPC introduction blocks that likely contribute to this family but exact meaning must be mapped with one-field save diffs.

## Introduction/heard state

Persistent NPC state may include:

- introduced/met;
- specific gossip/quest lines heard;
- rewards/services used;
- difficulty-specific state.

Do not overload quest completion bits with every NPC UI memory flag.

The durable character model can preserve a generic NPC-state map/bitset mapped by the legacy save codec.

## Quest dialogue events

Dialogue can trigger or respond to quest events such as:

- quest offered;
- reward available;
- reward claimed;
- item handed in;
- act travel unlocked;
- NPC relocation/state change.

Use semantic quest events through the quest controller. NPC code should not directly set arbitrary legacy bit offsets.

## Vendors/gambling/repair

NPC dialogue opens a capability; actual transaction authority remains in the item/economy system.

Flow:

```text
NPC choice Trade
 -> interaction context exposes vendor snapshot/quote APIs
 -> player buy/sell commands
 -> item authority revalidates NPC context/range
```

Closing the dialogue UI is presentation. Walking away automatically invalidates future transactions because authority re-checks range.

## Heal/identify

These are service transactions:

- Heal mutates player and possibly owned-unit resources/states through player/combat/state APIs;
- Identify mutates item identified state through item authority;
- cost/quest-free behavior is server-owned.

Do not implement them as Lua loops over visible items/resources.

## Hirelings

NPC hire/resurrect choices route to [HIRELINGS_AND_OWNED_UNITS.md](HIRELINGS_AND_OWNED_UNITS.md).

The hireling list/price is a server-derived snapshot; purchase/revival is revalidated/committed by authority.

## Quest-item hand-ins/services

Use `ContainerQuest`/service escrow and quest-item events.

NPC choice can open a named service interaction; exact item identities move through item authority and the commit produces quest events.

This avoids deleting items from a dialogue script.

## NPC town/world state

NPCs may move, disappear, change mode, become available/unavailable or expose different capabilities after quest events.

Keep this in authoritative world/NPC state:

```text
NPCInstanceState
  active/present
  position/path/mode
  interaction enabled
  behavior flags
  quest controller links
```

Presentation cannot assume an NPC always exists at a hard-coded pixel coordinate.

## NPC movement/AI

Some NPCs follow scripted paths or use monster-like AI. Reuse world movement/AI primitives where appropriate but separate:

- combat hostile AI;
- town scripted movement;
- map-AI path points;
- follower/escort behavior.

Interaction selection tracks the authoritative current position.

## Multiplayer dialogue state

Multiple players can interact with the same NPC concurrently.

Separate:

- shared NPC world state;
- per-player interaction context;
- per-player quest/dialogue state;
- shared vendor stock versus per-player gamble stock;
- services whose outcome mutates shared game state.

Do not store "NPC is in dialogue" as a single global lock unless the original service specifically requires one.

## Audio integration

NPC speech uses the semantic line ID and the `speech` bus. Interaction/menu SFX and NPC ambient vocal cues are presentation.

See [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md).

Dialogue logic should not block authoritative game simulation on local audio playback completion unless a specific cinematic/scripted sequence explicitly requires synchronized timing.

## Suggested implementation slices

### NPC1 — typed capability/dialogue snapshot

Generalize current interaction context to expose an ordered server-derived list of semantic capabilities/options for one NPC.

### NPC2 — simple talk/gossip

Select localized text/speech identity for a synthetic NPC with no side effects.

### NPC3 — quest dialogue event

Implement one NPC line/choice whose availability depends on quest state and whose selection emits a quest event.

### NPC4 — service routing

Open existing vendor/repair/identify/heal capability through the same interaction snapshot and range revalidation.

### NPC5 — durable NPC introduction state

Add semantic per-player NPC state and map one small legacy save field/bit through the save codec.

### NPC6 — moving/conditional NPC

Update interaction target position/capabilities when a quest changes NPC world state.

## Verification backlog

1. NPC dialogue table/recovered-data linkage to quest controllers.
2. Speech line IDs -> localized text/audio mapping.
3. Dialogue/gossip ordering and random selection.
4. NPC introduction/heard save bits and one-field `.d2s` diffs.
5. Quest offer/reward dialogue side-effect timing.
6. Class-specific dialogue variations.
7. Party/shared quest dialogue credit.
8. Service unlock restrictions by quest/difficulty.
9. Vendor/gamble interaction lifecycle and refresh on talk/leave town.
10. Heal/identify exact conditions and party/pet effects.
11. Hire/resurrect list/price behavior.
12. Quest item hand-in sequencing.
13. NPC relocation/disappearance and interaction invalidation.
14. NPC scripted movement/map-AI behavior.
15. Multiplayer concurrent interactions/shared NPC state.
16. Speech interruption/StopInst/Solo/subtitle timing.
17. Act-transition NPC travel options and quest gates.

## Primary sources inspected

- Current Dark Magic interaction authority, item/vendor/service/hireling/quest/localization/audio foundations and recovered Riiablo dialogue data.
- D2MOO pinned 1.10f `UNIT/SUnitNpc.*`, player/quest/NPC service paths and NPC event/trade structures.
