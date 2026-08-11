# Next gameplay implementation sequence

Status: implementation-planning handoff derived from the merged gameplay research and the current `ROADMAP.md`.

This document is intentionally short-lived compared with the individual research baselines. Reassess it whenever a dependent M21 checkpoint merges. The roadmap remains the implementation-status authority; the system research documents remain the fidelity/evidence authority.

## Current implementation cursor

As of the current repository state, the research program is no longer waiting for a first simulation architecture. The following pieces are already implemented and should be extended rather than replaced:

- M21.0 has a pointer-first authoritative player slice, production-map collision/pathing, a medium player footprint, replayable movement commands, a real legacy composite, and a generated Rogue Encampment/Blood Moor seam.
- M21.1 provides parameterized, provenance-preserving `dm.stats/v1` sources with atomic replace/remove and replay/checkpoint state.
- M21.2 provides explicit eight-fractional-bit combat amounts and typed damage/resource channels without claiming unverified legacy formulas.
- M21.3 materializes one ordinary typed hostile from joined MonStats/MonStats2/MonLvl facts.
- M21.4 gives that hostile scheduled deterministic target acquisition, chase, and semantic attack-request behavior.
- M21.5 resolves one replayable melee request into semantic attempt/hit/damage/death events using explicitly synthetic hit policy.

Do not create parallel stat, combat, monster, AI, item, session, or targeting authorities to implement the remaining work.

## Immediate queue: finish the first simulation spine

These are the active roadmap checkpoints and should remain sequential because each establishes state needed by the next.

### 1. M21.6: consume assigned skill intent

Goal: convert the existing `dm.player.skill_intent` presentation/admission boundary into one immutable authoritative cast request.

Preserve:

- `player.use_skill` as the semantic input command;
- assigned left/right skill identity;
- target unit identity plus target position;
- learned-skill ownership;
- fixed-tick command ordering.

The cast request should snapshot the facts later phases must not re-read ambiguously, especially:

- caster entity/stable player identity;
- skill ID;
- learned/effective level used by this cast;
- side/assignment identity where compatibility needs it;
- target kind/ID/position;
- source tick and purpose-specific deterministic identity.

Do not charge mana, spawn missiles, apply damage, or select renderer animation in this checkpoint.

### 2. M21.7: generic skill cast state

Goal: make casting a deterministic simulation transaction rather than a one-frame callback.

Add the minimum state machine needed for:

```text
request -> validate -> start -> effect -> complete
                     \-> interrupted
```

Required boundaries:

- target policy validation;
- resource-cost validation and authoritative consumption;
- start/effect/complete ticks;
- interruption/cancellation reason;
- semantic cast events;
- one data-selected behavior family;
- replay/checkpoint restoration in the middle of a cast.

Animation may follow cast state, but animation frames must not own effect timing.

### 3. M21.8: timed state engine

Goal: establish the reusable owner for buffs, debuffs, curses, shrine effects, auras, item states, and control effects before individual skills proliferate bespoke timers.

Minimum contract:

- stable state instance identity;
- state/type ID;
- source/owner identity;
- start and expiration tick;
- parameterized stat-source attachment;
- explicit refresh/replace/stack policy;
- apply/remove semantic events;
- deterministic scheduled expiration;
- checkpoint/replay support.

Prove one simple refreshable state. Do not infer universal Diablo stacking behavior from that one fixture.

### 4. M21.9: straight missile slice

Goal: prove an authoritative projectile from cast request through impact/removal.

The first missile should contain enough immutable creation state to avoid later source-state ambiguity:

- owner/source;
- skill and effective level;
- creation tick;
- start position and target/direction;
- velocity/range/lifetime;
- collision policy;
- snapshotted damage/effect inputs required by the selected behavior;
- stable hit memory where needed.

Run movement, collision, impact, combat event emission, and removal on authoritative ticks. Presentation follows copied missile state/events.

### 5. M21.10: Blood Moor population slice

Goal: turn generated-zone content into an inspectable deterministic spawn plan rather than hand-placing the first test monster.

The plan should separate:

```text
zone/population recipe
        -> spawn plan
        -> authoritative materialization
        -> active/inactive residency
        -> presentation
```

Do not make render visibility determine whether a monster exists in simulation.

### 6. M21.11: monster death transaction

Goal: replace the current minimal `unit-died` consequence with one atomic death transaction.

Join, in deterministic order:

- lethal combat resolution;
- kill/death attribution;
- XP consequence;
- deterministic loot event/materialization;
- quest event surface;
- corpse/dead-state creation;
- targetability/collision/mode changes;
- active/inactive lifecycle state;
- semantic audio/presentation cues.

This is the checkpoint that completes the first meaningful hostile gameplay loop.

### 7. M21.12: owned-unit relation

Goal: establish generic ownership before summons, pets, hirelings, traps, or minion-attributed loot/combat each invent a separate model.

At minimum represent:

- owner identity;
- category/PetType-like semantic class;
- limit/accounting policy;
- combat/loot/XP attribution;
- lifetime/death/despawn policy;
- checkpoint/replay state.

Hireling persistence/progression remains a layer above this generic relationship.

## First acceptance loop after M21.12

Before widening feature breadth, prove one generated Blood Moor replay from start to finish:

1. player enters the generated zone;
2. deterministic population materializes a hostile;
3. hostile acquires and paths to the player;
4. player and hostile exchange at least one authoritative action;
5. a skill can enter cast state and one straight missile can resolve;
6. a timed state can apply and expire;
7. lethal resolution produces one atomic death consequence;
8. XP/loot/corpse/quest-event surfaces are produced;
9. copied presentation/audio events follow authority without mutating it;
10. checkpoint restore and replay reproduce the same semantic result.

Only after this loop is stable should broad content coverage become the main queue.

## Next breadth queue after the first simulation loop

The merged research baselines imply the following dependency order. These are implementation themes, not new milestone numbers; use `ROADMAP.md` to assign/checkpoint them.

### A. Combat fidelity and player combat

Extend the basic transaction with evidence-backed rules in small slices:

- player basic attack transaction;
- defense/chance-to-hit;
- block/avoidance;
- resistance and physical/magic reduction;
- absorb;
- critical/deadly/mastery ordering;
- leech/drain;
- regeneration and periodic damage;
- durability consequences;
- player death/corpse/XP-loss policy;
- difficulty/PvP scalars.

Exact formulas stay behind the combat verification queue until supported. Continue to label temporary Dark Magic scaffolding explicitly.

### B. Item effects and generated-item integration

Use `internal/game/stats` as the activation target for equipment, charms, socket fillers, runewords, sets, temporary item states, and charged/aura/proc effects.

Recommended order:

1. richer generated item instance identity/provenance;
2. equip/unequip stat-source attachment;
3. container-dependent charm activation;
4. socket child identity and gem/rune/jewel sources;
5. runeword recognition/source lifecycle;
6. set thresholds/partial bonuses;
7. charged skills/aura/proc hooks;
8. monster/chest death events -> M6 loot -> world item authority;
9. persistence round-trip.

Do not duplicate the completed M6 treasure/quality/affix/property engine.

### C. World interactions and GameRules

Generalize the already-correct interaction and transition authority:

1. target kinds beyond NPCs;
2. one stateful door;
3. one deterministic chest;
4. one shrine using the timed-state engine;
5. generalized authored transition endpoints;
6. waypoint unlock/travel state;
7. dynamic town/quest/Cube portal instances;
8. server-derived NPC dialogue/service snapshots;
9. immutable session `GameRules` for difficulty/content mode;
10. encounter-controller primitive for bosses/special events.

Animation and Lua remain observers of committed object/quest/transition state.

### D. Cube, vendors, quest items, and economy

Build on current authoritative item containers and vendor/service commands:

- declarative CubeMain matcher + atomic outputs;
- item transformation/copy/upgrade/socket/repair/recharge outputs;
- quest-item lifecycle and difficulty provenance;
- vendor stock generation/refresh;
- full quote model beyond `baseCost * multiplier / 1024`;
- repair/recharge;
- gambling;
- identify/heal/quest services.

Pricing formulas and special recipe edge cases remain explicit verification work.

### E. Gameplay and environmental audio

The backend already has the right ownership model. Implement semantics above it in this order:

1. normalize all relevant `Sounds.txt` fields into an immutable playback definition;
2. semantic audio-event bridge from authoritative gameplay events;
3. zone soundscape state driven by level `SoundEnv`, inside/outside, and weather facts;
4. positional/tracking emitters for monsters, objects, missiles, and effects;
5. monster/NPC/object/skill/item/UI cue mapping;
6. music/zone transition policy;
7. priority, duplicate/compound, delay, pitch, fades, falloff, spread, solo/ducking;
8. audio diagnostics showing active semantic emitters and resolved records.

Audio variation must use presentation/audio randomness, never consume gameplay RNG streams.

### F. Versioned UI gameplay projections

Do not let each Lua panel crawl raw ECS stores. Build stable copied/revisioned projections for the contracts in `GAMEPLAY_UI_CONTRACTS.md`, beginning with the screens needed by implemented gameplay:

- HUD vitals/resources/active skills;
- target/monster state;
- item tooltip/equipment sources;
- vendor/Cube/service state;
- quest/waypoint state;
- party/trade state;
- mercenary/owned-unit state;
- death/respawn state.

Lua continues to submit semantic intent and wait for the next projection to confirm authority.

## Networking follows semantic gameplay, not the other way around

M22 should begin once the first authoritative gameplay loop and stable projections exist.

Recommended first networking slices:

1. run the same `Session` through an in-process loopback transport;
2. remote command submission into `Session.Submit` with the same validators as local commands;
3. versioned initial snapshot/projection transfer;
4. per-client incremental authoritative projection/events;
5. reconnect against stable session/player identity;
6. correction/rollback policy only after measured need;
7. latency/loss/duplicate/malformed-command tests.

Do not create a second network-specific gameplay implementation.

## Realm/persistence follows the game-worker contract

M23 should layer on top of the stable game-server boundary:

- Account / Character / SessionPlayer identity separation;
- character lease acquisition and release;
- revision/CAS durable character commits;
- game directory and worker allocation;
- signed admission/reconnect tokens;
- content/mod fingerprint negotiation;
- graceful draining and crash recovery;
- ladder/social/account services as independent realm data.

Future BNCS/MCP/D2GS compatibility belongs in protocol adapters around semantic realm/game services. It must not become the internal simulation API.

## Research/probe work that should run in parallel

Implementation does not need to stop while these are researched, but exact compatibility claims do:

- chance-to-hit, block, avoidance, mitigation, absorb, leech, poison and PvP arithmetic;
- skill action timing, delay rules, interruption, state refresh/stack groups and missile stepping/collision;
- original AI cadence and specialized path types;
- NoDrop/MF/Gold Find rounding and generation-context quality differences;
- runeword/socket and charm/container edge cases;
- Cube operation/output details;
- full vendor pricing/repair/gamble rules;
- object operation IDs/timing, shrine reset/math and portal/waypoint details;
- original player death/corpse/XP-loss behavior;
- `Sounds.txt` pitch/fade/compound/falloff/tracking/solo/block semantics and `Levels.SoundEnv` behavior;
- party reward sharing, hostility/PvP edge cases, trade UX/state and original-client protocol mapping;
- lossless `.d2s` compatibility and TXT->BIN linkage.

Promote findings into implementation only with the confidence/patch labels defined by the research program.

## Rules for Codex sessions

Before starting a gameplay PR:

1. read the relevant roadmap checkpoint;
2. read its research baseline and verification queue;
3. inspect current `main` because another dependent checkpoint may have merged;
4. state **known / inferred / unresolved** before implementing a source-sensitive formula;
5. extend existing owners instead of creating parallel systems;
6. keep each PR to one reviewable behavioral objective;
7. add synthetic tests first, then owned-game/MPQ/save/network probes where the claim requires them;
8. update `ROADMAP.md` in the same PR when objective acceptance is satisfied.

This sequence is planning guidance, not a substitute for the roadmap's checkbox evidence policy.