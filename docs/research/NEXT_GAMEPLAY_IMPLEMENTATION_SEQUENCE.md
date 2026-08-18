# Next gameplay implementation sequence

Status: superseded implementation-planning history. The
[Dark Magic Roadmap project](https://github.com/users/gravestench/projects/1),
its linked issues, and GitHub milestones own the current queue. This document
remains useful for the evidence and rationale behind the completed M21 spine,
but its old assumption that M22 networking follows gameplay is obsolete.

Target: expansion-only Diablo II: Lord of Destruction 1.14d. Classic and
earlier-patch behavior are not implementation targets.

This document is historical planning guidance, not an implementation-status
authority. GitHub issues and milestones own status; the system research
documents remain the fidelity/evidence authority.

## Current implementation cursor

The first simulation architecture is no longer hypothetical. Current `main` already contains:

- M21.0 pointer-first authoritative player movement, production-map collision/pathing, a medium player footprint, replayable commands, a real legacy composite, and the Rogue Encampment/Blood Moor seam;
- M21.1 parameterized provenance-preserving `engine.stats/v1` sources with atomic replacement/removal and replay/checkpoint state;
- M21.2 explicit eight-fractional-bit combat amounts and typed physical/fire/lightning/cold/poison/magic/life/mana channels;
- M21.3 one ordinary hostile materialized from joined MonStats/MonStats2/MonLvl facts;
- M21.4 scheduled deterministic hostile target acquisition, chase, and semantic attack requests;
- M21.5 one replayable melee transaction emitting semantic attempt/hit/damage/death events with explicitly synthetic hit policy;
- M21.6 an intent-phase skill consumer that freezes `d2legacy.player.skill_intent` into immutable `d2legacy.skill.cast_request`, preserving admitted skill choice, learned level, target, side, player identity, and request tick exactly once;
- M21.7 a normalized skill-definition registry plus checkpointed cast lifecycle with target/resource validation, start/effect/complete ticks, interruption state, semantic cast events, mana committed once, and a first headless `basic.point_event` behavior family.
- M21.8 a checkpointed timed-state owner with source-tagged identity,
  same-source refresh, independent sources, deterministic expiration, and
  semantic apply/refresh/remove events. Rich stat-source attachment and
  additional stacking/dispel/aura policies remain deliberately deferred.
- M21.9 a headless straight-missile vertical slice with cast-effect creation,
  snapshotted damage and motion facts, swept unit contact, combat-owned impact,
  range/lifetime removal, semantic events, and replay-equivalent restoration.
- M21.10 an inspectable deterministic Blood Moor population plan derived from
  generated-zone and typed monster/level facts, collision-aware placement, and
  privileged replayable materialization independent of renderer residency.
- M21.11 a single effects-phase monster-death commit that consumes lethal
  combat output once, records corpse/lifecycle state, disables live-unit
  behavior, attributes lethal-player XP, rolls the authored treasure class,
  and emits stable kill/loot/quest/presentation facts with replay-equivalent
  checkpoint restoration.
- M21.12 a generic checkpointed owned-unit relation with explicit immediate
  and ultimate owner identity, category/group/limit/replacement policy,
  lifecycle facts, deterministic excess handling, stable queries, and death
  credit propagation to the ultimate player owner.
- the post-networking G4 correction separates immutable admission capacity from
  mutable gameplay population: join/leave drives scaling by default, a
  checkpointed host command supplies `/players X`, and monsters pin the
  effective count at spawn without a skill/subsystem-specific branch.

Do not create parallel stat, combat, monster, AI, skill, item, session, targeting, or transition authorities to implement the remaining work.

## Immediate queue: first simulation acceptance gate

### Completed: M21.8 timed state engine

Goal: create the reusable owner for buffs, debuffs, curses, shrine effects, auras, item states, and control effects before content adds bespoke timers everywhere.

Minimum contract:

- stable state-instance identity;
- state/type ID;
- source and owner identity;
- start/expiration tick;
- parameterized `internal/game/stats` source attachment;
- explicit refresh/replace/stack policy;
- semantic apply/remove events;
- deterministic expiration scheduling;
- checkpoint/replay support.

Prove one simple refreshable state. Do not infer universal legacy stacking behavior from that fixture.

The reusable timed-state foundation is complete. Additional policy modules and
stat-source attachment remain future extensions driven by concrete effects.

### Completed: M21.9 straight missile slice

Goal: prove one authoritative projectile from cast effect through impact/removal.

Creation should freeze enough information that later changes to the caster do not ambiguously rewrite an in-flight projectile:

- owner/source;
- skill ID and effective level;
- creation tick;
- start position and target/direction;
- velocity/range/lifetime;
- collision policy;
- snapshotted damage/effect inputs required by that behavior;
- stable hit memory where needed.

The first synthetic single-hit policy now proves this contract through
checkpointed cast, motion, swept unit collision, combat impact, and removal.
That proof has since been refactored into a definition-driven straight-missile
family: Fire Bolt is configuration/fixture coverage, not a standalone system.
Additional retail definitions and movement/contact/impact families remain
implementation-driven verification work.

### Completed: M21.10 Blood Moor population slice

Goal: replace the hand-created hostile fixture with an inspectable deterministic population/spawn plan derived from generated-zone content.

Keep the layers explicit:

```text
zone population recipe
       -> spawn plan
       -> authoritative materialization
       -> active/inactive simulation residency
       -> presentation
```

The first explicitly synthetic room-density policy now proves this separation
and records every suppression/fallback decision. Exact retail density, room
eligibility, and pack-quality behavior remain verification work.

The generic player action path now also derives `UseAttackRate` timing from
named `attackrate`/`item_fasterattackrate` sources, equipped weapon speed, and
the pinned AnimData record. Attack is only the first exact-ID fixture consuming
that family; subsequent rate-using skills must reuse it and supply their own
target-version action/sequence evidence.

### Completed: M21.11 monster death transaction

Goal: replace the current minimal lethal consequence with one atomic semantic death transaction.

Join, in deterministic order:

- lethal combat resolution;
- killer/source attribution;
- XP consequence;
- deterministic M6 loot event/materialization;
- quest-event surface;
- corpse/dead state;
- targetability/collision/mode changes;
- active/inactive lifecycle state;
- semantic presentation/audio cues.

Death is not merely `hp <= 0`; it is an authoritative transition whose
consequences replay together. The initial policy is now complete. It explicitly
leaves party/owned-unit attribution, corpse expiry/skills, and authoritative
world-item materialization to their existing future owners.

### Completed: M21.12 owned-unit relation

Goal: establish generic ownership before summons, pets, traps, minions, and hirelings each invent separate owner/attribution models.

At minimum represent:

- owner identity;
- category/PetType-like semantic class;
- limit/accounting policy;
- combat/loot/XP attribution;
- lifetime/death/despawn policy;
- checkpoint/replay state.

The generic relation is complete. Hireling persistence, progression, equipment,
revival, and service behavior remain layers above it, as do concrete summon
skills and category-specific transition/death execution.

## Acceptance gate after M21.12

Before breadth becomes the dominant goal, prove one complete generated Blood Moor simulation loop:

1. the player enters the generated zone;
2. deterministic population materializes a hostile;
3. the hostile acquires and paths to the player;
4. player and hostile exchange an authoritative action;
5. an assigned skill runs through deterministic cast state;
6. one straight missile resolves through authoritative movement/collision/impact;
7. one timed state applies, refreshes or expires according to explicit policy;
8. lethal resolution produces one atomic death transaction;
9. XP, loot, corpse, and quest-event surfaces are produced;
10. presentation and audio consume semantic events/snapshots without mutating authority;
11. checkpoint restore and replay reproduce the same semantic result.

This is the first strong gameplay spine. Widen feature coverage after this loop is stable rather than before.

The renderer-independent authority portion is now proven end to end, including
command-admitted entry/population/action, hostile acquisition and fixed-step
movement, melee, cast/missile/state/death consequences, midpoint restoration,
full command-log replay, and checksum-stable observation. Live monster visuals
now consume copied identity, joined MonStats2 component recipes, derived display
mode/facing, and world position; death/missile events enter an observe-once
value queue, and the death cue resolves the authored MonSounds/Sounds record.
Missile definitions now also carry deliberately joined Missiles.txt DCC,
direction, timing, offset, and Sounds.txt keys; retained projectiles consume
copied live positions and observe-once spawn/hit audio cues. The first
production configuration is registered: reviewed Fire Bolt rows normalize
their fixed-point mana and fire damage, straight-missile movement/contact facts,
and presentation/audio recipe into a generic trusted definition. Command
admission, lifecycle, spawning, contact, and damage contain no Fire Bolt branch.
Unknown server behavior does not enter that family merely because its row also
references a missile.

## Breadth queue after the first simulation loop

The merged research baselines imply this dependency order. These are historical
implementation themes, not live milestone assignments. Promote remaining work
through a concrete GitHub issue and milestone before implementation.

### A. Combat fidelity and player combat

Extend the basic transaction with evidence-backed slices:

- player basic attack;
- defense/chance-to-hit;
- block and avoidance;
- resistance, physical/magic reduction and absorb;
- critical/deadly/mastery ordering;
- leech/drain;
- regeneration and periodic damage;
- durability consequences;
- remaining player corpse/equipment, gold, XP-loss/recovery, respawn, and
  Hardcore persistence after the common same-entity death/action-filter
  foundation;
- difficulty and PvP scalars.

Keep exact arithmetic behind `COMBAT_SIMULATION_VERIFICATION_QUEUE.md` until supported. Temporary Dark Magic scaffolding must remain labeled as scaffolding.

### B. Item effects and generated-item integration

Use `internal/game/stats` as the common activation target for equipment, charms, socket fillers, runewords, sets, temporary item states, auras, charges, and procs.

Recommended order:

1. richer generated item-instance identity and provenance;
2. equip/unequip source attachment;
3. container-dependent charm activation;
4. socket child identity and gem/rune/jewel sources;
5. runeword recognition and source lifecycle;
6. set thresholds/partial bonuses;
7. charged skill/aura/proc hooks;
8. monster/chest semantic drop events -> existing M6 loot -> world item authority;
9. durable save/realm round trip.

Do not duplicate the completed M6 treasure/quality/affix/property engine.

### C. World interactions and GameRules

Generalize existing interaction and transition owners:

1. interaction target kinds beyond NPCs;
2. one stateful door;
3. one deterministic chest;
4. one shrine using the timed-state engine;
5. generalized authored transition endpoints;
6. waypoint unlock/travel state;
7. dynamic town/quest/Cube portal instances;
8. server-derived NPC dialogue/service snapshots;
9. immutable session `GameRules` for difficulty/content mode;
10. encounter-controller primitive for bosses/special events.

Object animation and Lua remain observers of committed authoritative state.

### D. Cube, vendors, quest items, and economy

Build on current authoritative Cube/vendor/quest-service containers:

- declarative CubeMain multiset matcher and atomic outputs;
- copy/upgrade/socket/unsocket/repair/recharge transformations;
- quest-item generation/pickup/drop/consume/service/difficulty provenance;
- vendor stock generation and refresh;
- full quote model beyond the current simple base-cost multiplier scaffold;
- repair and recharge;
- gambling;
- identify/heal/quest services.

Exact pricing and special recipe edge cases remain explicit probes.

### E. Gameplay and environmental audio

The existing audio mixer/catalog already has the correct native ownership boundary. Add semantics above it:

1. normalize relevant `Sounds.txt` fields into immutable playback definitions;
2. bridge authoritative semantic gameplay events to audio cues;
3. add zone soundscape state driven by `Levels.SoundEnv`, inside/outside, and weather facts;
4. add positional/tracking emitters for monsters, objects, missiles, and effects;
5. map monster/NPC/object/skill/item/UI cues;
6. add music/zone-transition policy;
7. implement priority, duplicate/compound, delay, pitch, fades, falloff, spread, solo/ducking as verified;
8. expose diagnostics for active semantic emitters and resolved sound records.

Audio variation uses audio/presentation randomness and must never consume gameplay RNG streams.

### F. Versioned gameplay projections for Lua/UI

Do not let every Lua panel independently crawl raw ECS stores. Build copied/revisioned semantic projections from `GAMEPLAY_UI_CONTRACTS.md`, starting with implemented gameplay:

- HUD vitals/resources/active skills and cast state;
- target/monster state;
- item tooltip/equipment source summaries;
- vendor/Cube/service state;
- quest/waypoint state;
- party/trade state;
- mercenary/owned-unit state;
- death/respawn state.

Lua submits semantic intent and waits for the next projection/event to confirm authority.

## M22: networking follows semantic gameplay

Begin networking once the first authoritative gameplay loop and stable view models exist.

Recommended order:

1. run the same `Session` through an in-process loopback transport;
2. submit remote semantic commands into the same `Session.Submit` validators used locally;
3. transfer a versioned initial snapshot/projection;
4. send per-client incremental authoritative projections/events;
5. reconnect using stable SessionPlayer/session identity;
6. add correction/rollback only after measured need;
7. test latency, loss, duplication, malformed commands, reconnect, and soak.

Never create a network-specific gameplay implementation.

## M23: realm and trusted persistence follow the game-worker contract

Layer realm services around the stable game server:

- Account / Character / SessionPlayer identity separation;
- character leases;
- revision/CAS durable character commits;
- game directory and worker allocation;
- signed admission/reconnect tokens;
- content/mod fingerprint negotiation;
- graceful draining and crash recovery;
- ladder/social/account data as separate realm state.

Vanilla BNCS/MCP/D2GS protocols, vanilla save files, and old community-tool
interoperability are outside the supported product boundary.

## Research and probes to run in parallel

Implementation should continue while exact compatibility questions are researched, but unsupported exactness claims should not:

- chance-to-hit, block, avoidance, mitigation, absorb, leech, poison and PvP arithmetic;
- cast timing/delay/interruption, state refresh/stack groups, missile stepping/collision;
- original AI cadence and specialized path types;
- NoDrop/MF/Gold Find rounding and generation-context quality differences;
- runeword/socket/charm/container edge cases;
- Cube operation/output details;
- complete vendor pricing/repair/gambling rules;
- object operation timing, shrine reset/math, portal and waypoint details;
- exact player death DT/DD timing plus corpse/equipment, gold, XP-loss/recovery,
  respawn, multiple-corpse, save, and Hardcore persistence behavior;
- `Sounds.txt` pitch/fade/compound/falloff/tracking/solo/block semantics and `Levels.SoundEnv` behavior;
- party reward sharing, hostility/PvP edge cases, trade UX/state;

## Rules for Codex gameplay PRs

Before starting a gameplay checkpoint:

1. inspect current `main` because dependent work is merging rapidly;
2. read the linked issue, milestone, and acceptance boundary;
3. read the relevant baseline and verification queue;
4. summarize **known / inferred / unresolved** before implementing source-sensitive behavior;
5. extend the existing owner instead of creating a parallel subsystem;
6. keep the PR to one reviewable behavioral objective;
7. add synthetic tests first, then owned-game/MPQ/save/network probes where the claim requires them;
8. update the issue's evidence and status when objective acceptance is
   satisfied, and change repository documentation only when durable technical
   claims changed.

This file is historical planning guidance, not a substitute for live issue and
milestone acceptance.
