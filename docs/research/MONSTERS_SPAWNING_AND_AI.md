# Monsters, spawning, AI, and lifecycle research

Status: implementation-oriented research baseline. Runtime evidence is primarily D2MOO's Diablo II 1.10f reconstruction. Some monster-spawn routines in D2MOO are still represented by incomplete/stubbed decompilation, so this document distinguishes strong surrounding architecture evidence from spawn arithmetic that still needs independent verification.

This workstream connects map generation, game data, targeting, movement, skills, combat, loot, quests, and room activation. A monster is not merely a rendered DCC composite with HP.

## Executive conclusion

Dark Magic should model monsters in four distinct layers:

```text
MonsterDefinition
  MonStats / MonStats2 / MonLvl / MonProp / MonEquip / skills / AI / presentation
        |
        v
PopulationPlan
  which monster families/groups/superuniques belong in a generated zone
        |
        v
SpawnedMonster
  authoritative ECS identity, seed, level-scaled stats, owner/group/AI state
        |
        v
ActiveMonsterLifecycle
  perception -> AI think -> movement/action -> combat -> death/corpse -> inactive/removal
```

Map generation should decide **where and what may be populated**; monster spawning should materialize authoritative units; AI should decide intent; combat/skill systems should execute actions; presentation should render the resulting state.

Do not put these responsibilities into one `Monster` object with renderer handles and ad hoc callbacks.

## Current Dark Magic foundation

The repository already provides much of the substrate:

- typed `MonStats`, `MonStats2`, `MonLvl`, `MonProp`, `MonSounds`, `MonEquip`, `MonPreset`, `MonSeq`, unique-modifier, superunique, skill, missile and state data;
- generated world zones and authoritative semantic DS1 objects;
- stable selectable kinds including `hostile`;
- deterministic world navigation/collision;
- fixed-tick session ordering and replay/checkpoint infrastructure;
- deterministic loot generation;
- player/skill/item authority.

What is missing is the authoritative monster population/materialization/AI/combat lifecycle joining those pieces.

## Monster definitions and level scaling

A normalized definition should preserve relationships rather than flattening every table into one giant struct.

Conceptually:

```text
MonsterDefinition
  class ID / base ID / next-in-class
  type taxonomy
  flags
  mode availability
  size/collision
  movement/velocity
  AI identity + AI params
  base skills
  resist/drain/cold-effect families
  treasure/loot references
  component/presentation references

MonsterDifficultyStats
  level
  HP
  defense
  attack rating
  damage ranges
  resistances
  XP
  scaling mode
```

`MonStats` and `MonLvl` interact. The exact table-driven scaling path depends on monster/table flags and difficulty and should be normalized explicitly, not replaced with a universal `base * level` formula.

The spawn pipeline should retain the effective source rows/IDs for diagnostics and replay fingerprinting.

## Spawn sources are not interchangeable

At least these sources need distinct semantics:

- DS1 preset monster/object directives;
- procedural level population from `Levels.txt` monster lists/group metadata;
- superunique/quest-required placement;
- skill-created summons/spawners;
- scripted encounter waves;
- resurrect/revive behavior;
- traps/nests/eggs/huts that create monsters;
- hirelings/pets, which use monster runtime machinery but have explicit ownership.

A generic `SpawnMonster(class, x, y)` primitive may exist internally, but callers need semantic reason/source metadata.

Recommended spawn request:

```text
MonsterSpawnRequest
  definition ID
  reason/source
  level/difficulty context
  position/room
  owner/group leader
  quality/unique policy
  initial mode
  seed/stream identity
  quest/encounter identity if applicable
```

## D2MOO monster-region evidence

`MonsterRegion.cpp` is useful evidence for the population layer.

Observed behavior includes:

- room-seeded attempts to choose legal coordinates;
- avoidance of warp/spawn regions using `Levels.txt` distance data;
- preset monster IDs that can resolve to ordinary monsters, special semantic pseudo-IDs, champions, or superuniques;
- quest-state substitutions for particular authored monsters;
- difficulty-specific suppression of some preset monsters;
- collision checks and room/town restrictions;
- group/champion spawning around selected monsters;
- level-specific class-chain substitutions.

This shows that a DS1/preset integer is **not always a direct MonStats row index**. The current Dark Magic recovered `MonPreset`/DS1 resolution work should feed a semantic spawn plan instead of hard-coding assumptions in a monster constructor.

### Caution: incomplete spawn decompilation

D2MOO's `D2GAME_SpawnNormalMonster_6FC68E30` body is currently stubbed with a large commented decompilation lead. Do not treat that missing implementation as proof of detailed initialization order.

For exact base-stat materialization, seed setup, group flags, AI initialization, and mode ordering, corroborate with other reconstructed functions, libd2/original traces, and owned-game probes.

## Population should be deterministic and inspectable

For a generated zone, record a population trace such as:

```text
zone seed/fingerprint
population stream identity
room/region
population rule
selected monster family
selected class variant
selected group size
selected quality/unique modifier
spawn positions
suppression reason, if any
```

This makes monster-density and "wrong monster in this room" bugs reproducible without renderer screenshots.

Population RNG should use named/purpose-specific deterministic streams or an equivalently replayable policy. It must not depend on map iteration order or Go map ordering.

## AI control is stateful

D2MOO's `AiGeneral.cpp` reconstructs an explicit AI-control object with:

- owner type/GUID;
- several AI parameters;
- a circular command list/current command;
- map-AI data;
- a minion list;
- flags;
- game/owner relationships.

Commands can be allocated, copied, selected, propagated to minions, and removed. This is strong evidence that original monster behavior cannot be faithfully represented as a stateless `func Think(monster) velocity` callback.

Dark Magic does not need to copy the exact structure, but it should reserve authoritative state for:

```text
AIState
  behavior family
  parameters
  current tactical state/command
  target memory
  owner/leader
  minion/group relationship
  next think tick
  behavior-specific memory
```

All state affecting future decisions belongs in checkpoints/replay.

## AI scheduling

The original runtime uses scheduled AI-think events. Combat state changes can delete/reschedule AI-think events; for example, the cold/freeze removal path schedules the next monster AI think using difficulty-specific `AIdel`, with a fallback delay.

This is an important dependency on `TIMING_RNG_AND_DETERMINISM.md`:

- AI does not necessarily think every simulation tick;
- AI cadence can be definition/difficulty/state dependent;
- state changes can move the next-think time;
- AI timers are authoritative state.

A Dark Magic AI scheduler should use future simulation ticks, not wall clock and not renderer frames.

## AI architecture: behavior family plus tactics

D2MOO separates broad AI/tactics utilities from monster-specific behavior. Source organization includes:

- `AiGeneral` for control/commands/ownership;
- `AiUtil` for queries/helpers;
- `AiTactics` for tactical actions;
- special behavior implementations including Baal and monster-family logic;
- `MonsterAI` for monster/hireling lifecycle helpers.

A useful Dark Magic decomposition:

```text
Perception
  visible/eligible targets, distance, LOS, threat/context
        |
        v
Behavior
  choose intent using definition + AI memory + deterministic RNG
        |
        v
Tactics
  move toward/away/circle
  use skill
  idle/wander
  flee
  operate scripted behavior
        |
        v
existing authoritative command/action systems
```

AI should **request the same movement/skill/combat actions** trusted players use where practical, rather than mutating velocity/HP through a backdoor.

## Target acquisition and aggro

Research must separate:

- candidate discovery radius;
- alignment/hostility;
- line of sight;
- current target retention;
- nearest/priority target selection;
- owner/party/minion relationships;
- taunt/attract/confuse and similar state-driven target policy;
- unreachable target/path failure;
- target loss on room transition/inactivation.

Do not equate presentation selection radius with AI aggro radius.

The current `internal/game/targeting` semantic kinds can be reused as one input, but AI likely needs richer authoritative faction/alignment and threat context.

## Group and minion relationships

AI control and spawn code both expose leader/minion concepts.

Group behavior may influence:

- spawn clustering;
- AI command propagation;
- target selection;
- champion/unique packs;
- death behavior;
- XP/loot quality;
- resurrection/spawner relationships.

Represent leader/owner/minion links explicitly with stable entity IDs. Do not rediscover a pack each tick solely by spatial proximity.

## Monster skills use the same skill engine

D2MOO's server skill tables include many monster skills. Monster AI should choose a skill/action, then submit it through the same normalized skill behavior layer used for player and hireling skills.

Benefits:

- one target-validation vocabulary;
- one missile system;
- one state system;
- one combat pipeline;
- consistent replay events;
- fewer "monster-only" damage shortcuts.

Monster-specific AI may still choose parameters or special behavior families.

## Collision, size, range, and movement

Monster size and collision pattern affect:

- legal spawn coordinates;
- pathfinding;
- door/passability behavior;
- melee range;
- knockback;
- missile contact;
- crowding.

The current Dark Magic circle/radius-aware `FindPath` is a useful generic foundation, but original collision patterns are more specific than one radius in some cases. Preserve the ability to move from a simple collider radius to a collision-pattern/footprint abstraction without changing every AI API.

## Inactive rooms and streaming

D2MOO has explicit inactive-unit/room machinery (`SUnitInactive`) and monster-region lifecycle code. Exact room activation policy needs deeper tracing, but the architectural requirement is already clear:

**presentation visibility is not authoritative monster existence.**

A generated world may have:

- active simulated monsters;
- serialized/inactive monsters in rooms outside the active simulation window;
- permanently dead/removed monsters;
- encounter controllers that remain active while individual visuals are not.

Room streaming must preserve all deterministic AI/combat/state/RNG information needed to resume behavior.

Do not despawn a monster merely because its render node was culled.

Current implementation status: `d2legacy.population.plan/v5` keeps stable
inactive resident records per generated room. The world-owned
`d2legacy.world.room_resident {id, level_id, room_id}` contract no longer depends on
monster identity, and each activation record preserves whether the generic
velocity-mover opt-in existed. Ordinary generated monsters retain their ECS
entity and full component/relationship graph; the deterministic
occupied-room-plus-neighbors policy adds an empty `d2legacy.world.inactive`
filter tag and removes the velocity-movement opt-in rather than copying an
allowlisted scalar record and destroying the entity. Lua simulation systems and
local/remote presentation projections exclude inactive residents. Checkpoint
continuation preserves timed-state, stat-source, and state-event target
references and proves deactivate/restore/reactivate parity; a non-monster,
non-moving fixture proves reactivation does not invent movement capability or
cross levels when room IDs repeat. Production DS1 interaction targets now join
generated room bounds through canonical string IDs. Active positioned residents
also refresh that membership before each activation decision, so movement
across a room boundary does not leave an entity attached to its spawn room.
The first owned-unit graph fixture attaches the generated resident directly to
the player's stable ECS entity and preserves its category, limit, lifetime,
durable identity, and attribution fields through inactive checkpoint restore.
The owned-unit lifetime system now excludes the empty inactive marker and
checks its unchanged absolute expiration when the unit next becomes active;
that deterministic scaffolding does not resolve legacy inactive timer aging.
Corpse/ground-item/projectile residency and stateful object operation/event attachment,
inactive timer-aging policy, and exact 1.14d activation or long-inactive
mutation policy remain unresolved.

## Death, corpses, resurrection

Monster death flows into [COMBAT_DAMAGE_AND_DEATH.md](COMBAT_DAMAGE_AND_DEATH.md).

Monster lifecycle must additionally preserve corpse semantics:

```text
alive monster
  -> lethal transaction
  -> death mode/effects
  -> corpse identity/state
  -> corpse can be:
       ignored
       consumed by skill
       exploded
       horked/searched
       revived/resurrected
       removed after policy/timing
```

The corpse must retain enough source identity for skills/quests/loot rules without keeping the entire live AI object active.

D2MOO's spawn code also derives a resurrection mode from `MonStats2` and may select a configured resurrection skill/mode. This is evidence that resurrection behavior is data-related and not simply `dead=false`.

## Unique, champion, superunique modifiers

The typed catalog already includes unique modifiers, appellations/titles and superuniques. Future spawning should separate:

```text
base monster definition
+ quality (normal/champion/unique/superunique)
+ modifier set
+ generated/name metadata
= effective spawned monster
```

Modifiers may affect:

- stats/resistances/immunities;
- AI/skills;
- aura/state sources;
- minions;
- death effects;
- loot/XP;
- presentation transforms.

Do not bake unique behavior permanently into cloned `MonStats` records.

## XP, loot, and quest credit

Death resolution should emit a stable monster-death event containing at least:

```text
monster definition/runtime ID
quality/modifiers
level/difficulty
killer
owner-attributed killer (pet/trap)
party context
position/level
spawn/encounter source
loot seed/stream identity
```

XP, loot, and quests consume that semantic event. They should not each independently infer who killed a now-deleted ECS entity.

## Recommended Dark Magic ownership

Possible package boundaries:

```text
internal/game/monster
    definition.go
    spawn.go
    state.go
    population.go

internal/game/ai
    state.go
    scheduler.go
    perception.go
    tactics.go
    behavior_registry.go
```

These names are suggestions. If current repository structure prefers monster AI under one owner, preserve the responsibilities even if filenames differ.

AI and monster code must remain renderer/Lua independent.

## Suggested implementation slices

### M-AI1 — monster definition/materialization

Materialize one ordinary hostile from typed data into ECS with:

- stable identity/class;
- seed;
- level-scaled HP/defense/attack/damage/resists;
- position/collider/selectable kind;
- neutral animation/appearance facts;
- no sophisticated AI yet.

### M-AI2 — deterministic population plan

For one generated Blood Moor region, choose and spawn ordinary monster groups from authored level/monster data using an inspectable deterministic trace.

### M-AI3 — basic scheduled AI

Implement idle/wander/acquire/chase/basic-attack behavior using future fixed ticks and current navigation/skill/combat boundaries.

### M-AI4 — group/minion AI

Add explicit leader/minion relationships and one command-sharing/group behavior.

### M-AI5 — quality and lifecycle (partial)

Death/corpse and XP/loot event integration plus the first persistent-identity
ordinary-monster room inactivation/resume path are implemented. Timed-state,
stat-source, and state-event references survive that path. Champion/unique
modifier composition, owned-unit/item/object/projectile activation graphs,
specialized corpses, and broader inactive-unit kinds remain.

## Verification backlog

1. Exact MonStats/MonLvl scaling path by flags/difficulty/player count.
2. Population density/group-size formulas and room eligibility.
3. Level monster list selection and rarity/weight behavior.
4. Preset pseudo-ID resolution across acts and versions.
5. Exact normal monster initialization order currently obscured by D2MOO's stubbed spawn routine.
6. Monster RNG seed derivation and per-purpose consumption order.
7. Champion/unique probability and modifier-selection rules.
8. Superunique minion rules and quest-state substitutions.
9. AI IDs -> behavior family mapping and parameter interpretation.
10. AI think cadence and `AIdel` meaning by game type/difficulty.
11. Target acquisition/retention and LOS/range rules.
12. Fear/flee/taunt/confuse/attract target-policy overrides.
13. Door interaction and path-type selection by monster family.
14. Collision patterns versus simple size/radius.
15. Pack leader/minion command and target propagation.
16. Room activation/inactivation radius and timing.
17. What exact monster state is serialized while inactive.
18. Corpse lifetime and corpse-eligibility flags.
19. Resurrection/revive state restoration and forbidden monsters.
20. XP party/minion ownership attribution.
21. Loot and quest kill-credit ordering relative to death effects/removal.
22. Boss-specific AI/encounter controllers that should remain separate from generic AI.

## Primary sources inspected

- D2MOO pinned 1.10f `MonsterRegion.cpp`, `MonsterSpawn.cpp`, `MonsterAI.cpp`, `Monster.cpp`, `MonsterMode.cpp`, `MonsterUnique.cpp`.
- D2MOO `AiGeneral.cpp`, `AiUtil.cpp`, `AiTactics.cpp`, `SUnitInactive.cpp`.
- D2MOO skills/combat/player-pet code for cross-system ownership and attribution.
- Current Dark Magic typed game-data catalog, generated world, targeting, navigation, session, loot and player systems.
