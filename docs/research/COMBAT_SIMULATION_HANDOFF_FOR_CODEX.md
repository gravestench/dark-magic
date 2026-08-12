# Combat/simulation research handoff for Codex

> Architecture note: this document records the implementation sequence and
> package ownership that produced the original Go authority spine. Its
> behavioral evidence and acceptance criteria remain useful, but permanent
> ownership guidance is superseded by
> [the engine/`d2legacy` boundary](../ARCHITECTURE.md). Completed Go systems are
> migration sources for M21.14, not proof that Diablo policy must remain in Go.

Status: implementation handoff derived from the second gameplay research tranche. Read the linked research documents before implementing broad systems.

## Read first

1. [COMBAT_DAMAGE_AND_DEATH.md](COMBAT_DAMAGE_AND_DEATH.md)
2. [SKILLS_STATES_AND_MISSILES.md](SKILLS_STATES_AND_MISSILES.md)
3. [MONSTERS_SPAWNING_AND_AI.md](MONSTERS_SPAWNING_AND_AI.md)
4. [HIRELINGS_AND_OWNED_UNITS.md](HIRELINGS_AND_OWNED_UNITS.md)
5. [MOVEMENT_COLLISION_AND_PATHFINDING.md](MOVEMENT_COLLISION_AND_PATHFINDING.md)
6. [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md)
7. [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md)

## Current-engine facts to preserve

Do not replace these existing boundaries:

- `internal/game/session` remains authoritative fixed-tick command/replay/checkpoint ownership.
- `internal/game/world` remains renderer-neutral static map/collision/coordinate ownership.
- Current deterministic eight-way A* remains a useful generic route planner.
- `player.move` remains the semantic authoritative movement command boundary.
- `player.use_skill` remains the presentation/network admission boundary for a selected skill and semantic target.
- `internal/game/targeting` semantic target kinds remain the starting target vocabulary.
- `internal/game/item` remains item/equipment authority.
- `internal/game/loot` remains loot/item-generation authority.
- typed game-data snapshots remain the source of Skills/Missiles/States/MonStats/Hireling/PetType/etc.
- Lua remains presentation and intent, not combat/AI authority.

## Highest-leverage implementation order

### 1. Shared stat/effect source model

Before broad combat, add the minimal layered stat-source representation needed by:

- equipped item properties;
- skill/passive sources;
- temporary states;
- monster modifiers;
- auras;
- hireling equipment.

Do not flatten every source into permanent final values.

### 2. Fixed-point combat vocabulary

Introduce explicit typed damage channels and fixed-point conversion/rounding helpers. Keep the first slice small and fully deterministic.

### 3. Monster materialization

Materialize one ordinary hostile from typed data with stable seed, effective level-scaled stats, collider, selectable identity, presentation facts and no sophisticated AI.

This creates a real target for the next slices.

### 4. Basic scheduled AI

Implement one deterministic behavior:

```text
idle -> acquire hostile -> path toward -> request basic attack
```

AI selects intent; it does not subtract HP itself.

### 5. Basic melee combat transaction

Implement:

```text
attack request -> legal target -> hit/miss -> physical damage -> HP -> hit/death event
```

Use synthetic verified vectors. Leave advanced critical/absorb/poison/etc. layered for later.

### 6. Consume `d2legacy.player.skill_intent`

Turn the existing use-skill intent into a deterministic cast request that resolves the assigned skill and learned level exactly once.

### 7. Generic skill cast state

Add target validation, resource cost, action start/effect/complete ticks and one basic skill behavior. Gameplay timing must work headlessly without DCC/animation assets.

### 8. Timed state engine

Implement source-tagged state instances and scheduled expiration. Start with one simple refreshable control state; checkpoint/replay it.

### 9. Straight missile vertical slice

Implement:

```text
cast -> authoritative missile -> motion/collision -> combat impact -> removal
```

Snapshot the required source stats at creation according to the documented evidence.

### 10. Monster population slice

Use generated Blood Moor content to build one deterministic population plan and materialize ordinary packs. Keep population trace inspectable.

### 11. Monster death integration

Join death to:

- XP event;
- deterministic loot event;
- quest kill event surface;
- corpse state;
- removal/inactive lifecycle.

### 12. Owned-unit relation

Add generic owner/category/limit/attribution state before implementing broad summons or hirelings.

## Things not to implement as shortcuts

Do not:

- use renderer animation frames as the authoritative attack/cast timer;
- let Lua directly damage units or spawn authoritative missiles;
- recompute projectile damage from current equipment on impact without evidence;
- make all skills one giant name switch;
- make all state effects `map[state]bool`;
- make every movement type ordinary A*;
- treat player/monster/hireling movement speed as the current hard-coded 10/15 forever;
- cull authoritative monsters because their render nodes are off-screen;
- infer owned-unit attribution from proximity;
- implement monster attacks through a private HP-mutation backdoor;
- hard-code community chance-to-hit/block/CB/Open Wounds formulas before the verification queue is answered.

## Cross-system event vocabulary to establish early

A small deterministic event record will prevent subsystems from scanning deleted/mutated ECS state after the fact.

Candidate semantic events:

```text
SkillCastStarted
SkillEffect
MissileSpawned
AttackAttempted
HitResolved
DamageApplied
StateApplied
StateRemoved
UnitKilled
UnitDied
CorpseCreated
MonsterSpawned
OwnedUnitAdded
OwnedUnitRemoved
ZoneActivated
ZoneDeactivated
```

The exact API is not prescribed. The requirement is stable causal data and ordering.

## Replay/checkpoint requirements

The following state affects future results and must not remain hidden in transient Go objects:

- combat/monster/missile RNG streams or equivalent deterministic stream state;
- active skill/cast state;
- scheduled effect ticks;
- active timed state instances and expiration ticks;
- missile position/motion/lifetime/pierce/hit memory;
- AI behavior memory and next-think tick;
- owner/minion/group relationships;
- inactive-room unit archives;
- monster/population identities already consumed from deterministic streams.

## Recommended PR granularity

Prefer small vertical PRs such as:

```text
combat fixed-point types + tests
one hostile monster archetype materialization
basic attack transaction
basic monster chase/attack AI
skill cast transaction shell
one timed state
one straight missile
monster death -> loot/XP/corpse event
owned-unit relation
one summon
one hireling
```

Do not submit "implement combat" or "implement monster AI" as one enormous PR.

## Verification discipline

For every compatibility-sensitive arithmetic path:

1. write the expected inputs/outputs as a vector;
2. record source/version;
3. distinguish source-derived behavior from Dark Magic policy;
4. capture owned-game/original output where possible;
5. add a holdout/boundary vector not used while deriving the implementation;
6. promote the associated research statement from baseline toward validated.

Use [COMBAT_SIMULATION_VERIFICATION_QUEUE.md](COMBAT_SIMULATION_VERIFICATION_QUEUE.md) as the empirical backlog.

## Definition of a successful first vertical slice

A strong near-term demo is:

```text
generated Blood Moor
   |
   +-- one typed ordinary hostile spawned deterministically
   +-- hostile acquires player on scheduled AI tick
   +-- hostile paths through authoritative collision
   +-- player or monster performs basic attack
   +-- fixed-point physical damage transaction resolves
   +-- HP changes in authoritative ECS
   +-- death emits semantic event
   +-- deterministic loot roll is triggered
   +-- replay/checkpoint reproduces identical result
   +-- Lua only displays snapshots/events
```

That proves the architectural spine before hundreds of skills and monster families multiply the number of edge cases.
