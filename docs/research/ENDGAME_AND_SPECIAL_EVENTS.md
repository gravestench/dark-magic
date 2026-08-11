# Endgame progression, special areas, bosses, and world-event research

Status: implementation-oriented research baseline. This document deliberately separates the expansion campaign end state and Secret Cow Level from later-patch/realm special content. The pinned D2MOO 1.10f reconstruction is strong evidence for campaign quest/event plumbing but contains compile-time/version-conditional hooks for later Uber Cube operations, so later systems require separate versioned evidence.

## Executive conclusion

"Endgame" is not one subsystem. It is a composition of existing authoritative systems:

```text
campaign/quest progression
+ difficulty unlocks
+ special item/Cube predicates
+ special portal/transition instances
+ generated/preset levels
+ encounter controllers
+ monster/boss spawn and AI
+ loot/quest rewards
+ game/realm event state
= special/endgame content
```

Dark Magic should build generic encounter/event/portal primitives, then describe each special area/event with versioned rules. Do not add a giant `endgame.go` full of special-case global flags.

## Campaign completion is durable progression

D2MOO's Act V Baal quest is strong evidence that completion updates per-difficulty quest state and character progression through the client/player progression boundary.

The recovered Act V quest:

- stores quest data keyed by current difficulty;
- reacts to monster kills, level changes, NPC activation/deactivation, player join/start/leave and scroll/dialogue events;
- marks reward/primary-goal states;
- updates character progression after eligible completion;
- creates the final portal as an authoritative world object in the Worldstone Chamber.

This supports Dark Magic's existing quest-runtime model and the `GameRules` difficulty split.

Campaign completion should emit a semantic progression event, not directly unlock UI buttons.

## Final portal / post-boss transition

The recovered Act V quest dynamically creates a final portal object after the appropriate quest/NPC sequence.

Architectural significance:

```text
boss/quest controller
 -> CreatePortalRequest
 -> authoritative portal object
 -> player operates/enters portal
 -> trusted transition authority
 -> epilogue/credits/front-end presentation as observer
```

Do not teleport the client to credits because the Baal death animation ended.

## Difficulty unlock

Completing the final required act/quest can unlock the next difficulty for the durable character.

This is a progression/persistence mutation committed by quest completion. Game creation later validates it through `DIFFICULTY_AND_GAME_MODES.md`.

The exact Classic versus Expansion completion quest and legacy progression field mapping need dedicated probes.

## Secret Cow Level

The Cube grammar includes an explicit `COWPORTAL` operation, and D2MOO quest/Cube paths reference the special-area rules.

Model Cow Level entry as:

```text
Cube recipe/predicate
  required items
  game difficulty/progression
  origin location/town context
  per-game/per-character special flags
        |
        v
CreatePortalRequest(destination = Cow Level)
        |
        v
special-area encounter/world state
```

Cube owns recipe validation/consumption. Transition/world owns the portal. Quest/game-event state owns repeatability/eligibility.

Do not implement Cow Level as a hard-coded scene accessible from the waypoint list.

## Cow King / repeatability state

Historical Cow Level behavior includes special eligibility affected by killing the Cow King and by game/party context. Exact target-version rules are subtle and changed across later game variants.

Dark Magic should reserve semantic state such as:

```text
SpecialAreaEligibility
  difficulty
  character/game flags
  event-specific completion/lockout facts
```

Do not encode a generic permanent `cow_level_disabled` boolean until owned 1.10f traces establish who receives the flag and under what party/kill-credit conditions.

## Special areas are ordinary levels with extra entry/encounter policy

Once admitted, a special area should still use:

- ordinary map generation/preset realization;
- monster population/AI/combat;
- object interaction;
- loot;
- room streaming;
- soundscape/music;
- network snapshots.

Avoid creating a separate "minigame" engine for Cow/Uber/event levels.

## Encounter controllers

Boss and event encounters need persistent authoritative controller state distinct from individual monster AI.

Example capabilities:

```text
EncounterController
  stable encounter ID
  phase/state
  spawned actor/object IDs
  timers/events
  quest/game predicates
  activation area
  completion state
  reward/portal requests
```

Individual bosses still use monster/AI/skill/combat systems. The controller owns encounter-wide sequencing such as waves, seals, portals, phase changes, scripted spawns or completion gates.

## Existing campaign examples

The quest/runtime source base already provides useful archetypes:

- seal/object-controlled Diablo encounter;
- Ancients activation/completion gate;
- Baal throne/worldstone sequence;
- quest bosses with party credit;
- dynamically created portals/objects.

These should become reusable encounter primitives rather than one-off quest code copied per boss.

## Boss kill credit

Endgame progression/rewards depend on authoritative kill attribution:

```text
immediate killer
ultimate owner
party eligibility/proximity
quest eligibility
current level/encounter
current difficulty
```

Combat emits the kill event; encounter/quest/reward logic consumes it before the monster entity disappears.

## Boss reward/loot policy

Bosses can have:

- ordinary TC drops;
- first-kill/quest drops;
- special quest items;
- guaranteed/event rewards;
- game/realm-wide progression effects.

Represent these as semantic reward policies layered over the standard loot/item system.

Do not put special reward creation inside presentation or boss death animation code.

## Versioned special-content registry

Dark Magic should be able to declare availability by content/runtime target:

```text
Feature
  ID
  IntroducedVersion/content era
  RequiresExpansion
  RequiresLadder/realm/event policy
  Data dependencies
  Entry rules
```

This matters because source trees can contain structures compiled differently for multiple versions.

For example, D2MOO's Horadric Cube enum includes Uber Dungeon/Uber Tristram operation codes only behind `D2_VERSION_HAS_UBERS`. That is evidence these operations are **not part of every target build**.

## Uber portals / Pandemonium-style content

Later-patch Uber portal content should reuse the generic Cube -> portal -> encounter stack:

```text
special keys/organs/items
 -> Cube operation
 -> version/game-mode eligibility
 -> dynamic special portal
 -> special level encounter
 -> boss rewards
```

Exact key drops, portal randomization/nonduplication, organ recipes, boss stats/AI, Hellfire Torch rules and realm/ladder restrictions require a later-version source/probe corpus. Do not infer them from the 1.10f baseline.

## Realm/global world events

Events such as server-wide or realm-coordinated special spawns require a separate **game/realm event input**, not local quest RNG.

Future architecture:

```text
Realm/Event Service
  durable/global counters or schedule
        |
        v
game server receives signed/authoritative event activation
        |
        v
GameEventState
  activation ID/version
  eligible game/level
  replacement/spawn policy
        |
        v
ordinary encounter/monster/reward systems
```

Offline mode can emulate the event with an explicit local policy, but should not pretend to have a realm-global state unless configured.

## Special monster replacement

Some world events replace or augment a normal unique/superunique spawn. Model this in monster population/spawn planning:

```text
SpawnRequest
 + active GameEventState
 -> transformed/replacement definition
```

Do not have the renderer swap a sprite after seeing an event notification.

## Event persistence and reconnect

A game-level special event/encounter may survive player disconnect/reconnect while the game exists.

Checkpoint state needs:

- active event ID/generation;
- encounter controller state;
- spawned unique/boss IDs;
- opened special portals;
- consumed per-game entry flags;
- reward/completion state.

Character durable saves keep only the event-related character progression/items that truly persist beyond the game.

## One-time/repeatable reward restrictions

Unique charms/special rewards may have possession or per-character restrictions. The reward transaction should query item inventory/quest/event state before creation/pickup.

Use the item admission policies from `CHARMS_AND_CONTAINER_EFFECTS.md` and `QUEST_ITEMS.md`.

## Event announcements/audio/UI

Realm/game event activation can produce:

- server/system messages;
- area/boss cues;
- music/soundscape changes;
- special overlays/UI.

Those are derived presentation events. Event activation itself is authoritative game/realm state.

## Mod extensibility

The generic event model should support mods adding:

- new Cube-gated areas;
- boss rushes;
- timed world invasions;
- rotating dungeons;
- custom global/realm counters;
- seasonal content.

Keep extension points semantic and data-driven without weakening authoritative boundaries.

## Suggested implementation slices

### E1 — encounter controller primitive

Implement a synthetic controller that activates from a world/object event, spawns a wave, tracks members and emits completion.

### E2 — campaign boss completion

Route one existing quest boss death through combat kill event -> encounter/quest completion -> durable progression/reward.

### E3 — dynamic final/quest portal

Create and operate an authoritative portal after an encounter completes.

### E4 — Cow Level vertical slice

Implement the target-version Cow recipe/eligibility/portal as a composition of Cube, transition, special-area and game-state rules.

### E5 — versioned feature registry

Make later special content explicitly conditional on content/runtime target and include it in rules/content fingerprints.

### E6 — realm-event interface

Define an input interface and synthetic event that can activate a replacement boss without implementing Battle.net coordination yet.

## Verification backlog

1. Classic/Expansion final quest -> progression/difficulty unlock rules.
2. Baal final portal creation/use and epilogue sequence.
3. Cow Portal Cube recipe predicates, origin level and item consumption.
4. Cow-level access after campaign completion by difficulty.
5. Cow King kill flag/party credit/repeatability rules in 1.10f.
6. Cow Level map/population/superunique/loot special rules.
7. Diablo seals/encounter controller sequencing.
8. Ancients activation/reset/party eligibility.
9. Baal throne waves, portal/chamber transition and quest credit.
10. First-kill/quest boss treasure/reward differences.
11. Later Uber operation introduction/version gates.
12. Key/organ/portal selection and nonduplicate Uber destination rules.
13. Uber boss AI/stats/rewards and unique-charm possession rules.
14. Realm/global event trigger/counter protocol for later patches.
15. Special unique replacement and game eligibility.
16. Event persistence/reconnect and multiplayer announcements.
17. Offline emulation policy for realm-only events.
18. Seasonal/ladder restrictions and content fingerprinting.

## Primary sources inspected

- Current Dark Magic quest/world/Cube/transition/monster/loot/persistence/game-rules architecture.
- D2MOO pinned 1.10f Act V Baal quest, Cube Cow Portal operation and campaign quest/event systems.
- D2MOO multi-version Cube definitions showing Uber operations behind an explicit version feature conditional.

Later-patch Uber/realm event specifics are intentionally left at lower confidence until a pinned later-version/runtime source corpus or owned-game traces are added.
