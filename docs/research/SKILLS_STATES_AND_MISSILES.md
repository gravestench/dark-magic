# Skills, states, missiles, and combat actions research

> Architecture note: this document's recovered behavior, data joins, and test
> vectors remain valid. Its proposed Go package layout describes the current
> transitional implementation, not the permanent owner. Under
> [the engine/`d2legacy` boundary](../ARCHITECTURE.md), D2 skill and missile
> policy migrates to authoritative Lua while justified generic scheduling,
> collision, movement, and data-decoding mechanisms may remain in Go.

Status: implementation-oriented research baseline. Runtime findings are primarily from D2MOO's Diablo II 1.10f reconstruction and are version-labeled accordingly. This document describes boundaries and evidence; it is not yet a complete skill-by-skill compatibility specification.

This workstream builds directly on:

- [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md);
- [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md);
- [COMBAT_DAMAGE_AND_DEATH.md](COMBAT_DAMAGE_AND_DEATH.md);
- the existing Dark Magic `player.use_skill` authoritative command;
- the typed `Skills`, `SkillDesc`, `Missiles`, `States`, and calculation-table catalogs.

## Executive conclusion

Dark Magic already has the correct **input/authority edge** for skills: presentation selects a side and semantic target; the fixed-tick session validates that the selected skill is learned/allowed and records authoritative intent.

The missing layer is the actual deterministic skill transaction:

```text
player.use_skill command
        |
        v
resolve assigned skill + learned level
        |
        v
validate use
  alive / mode / town / target / range / LOS / item / corpse / resource / delay
        |
        v
SkillCast transaction
  start tick
  source snapshot
  target snapshot/semantic target
  cost
  action/mode
  behavior family
        |
        +--> immediate effect
        +--> spawn missile
        +--> apply state/stat-list source
        +--> summon/owned unit
        +--> move/teleport/charge/leap
        +--> operate corpse/object/item
        `--> schedule continuation/events
```

Do not make each skill a bespoke UI callback. Do not require the renderer's current animation frame to decide whether gameplay fires.

## Current Dark Magic state

`internal/game/session/skills.go` currently provides:

- `player.assign_skills`;
- `player.use_skill`;
- semantic left/right assignment validation;
- learned-skill/side validation;
- point target `(x,y)` plus optional semantic target ID;
- an authoritative `d2.player.skill_intent` ECS record.

This is a good admission boundary, but the intent currently stops before:

- mana/resource cost;
- target/range/LOS validation;
- cast/action timing;
- cooldown/delay;
- server behavior dispatch;
- missile creation;
- state application;
- damage resolution;
- summoning;
- item/corpse/world-object operations.

The next milestone should consume the existing intent rather than replacing `player.use_skill`.

## Skills.txt dispatch is strongly data-driven

D2MOO's `Skills.cpp` reconstructs large server start/do dispatch tables. `Skills.txt` records identify behavior families by numeric server-function IDs, and those IDs route many named skills through shared implementations.

Examples of shared behavior families visible in the 1.10f tables include:

- ordinary attack and left-hand swing;
- kick/jab/impale and related melee actions;
- projectile firing and throwing;
- curses and buffs;
- auras;
- teleport;
- charge, leap, leap attack, whirlwind;
- corpse explosion and corpse-consuming skills;
- golems, skeletons, revives, druid summons and assassin shadows;
- sentries/traps;
- walls/prisons;
- monster skills and hireling missile behavior;
- stateful charge-up/release skills.

This is important architecture evidence: **skill identity and behavior family are separate concepts**.

Dark Magic should preserve the data-driven relationship while using independent semantic handler names/interfaces. A possible normalized layer is:

```text
SkillDefinition
  stable skill ID
  class / requirements / skill tab
  serverStartBehavior
  serverDoBehavior
  target policy
  cost formula
  delay policy
  calc references
  missile references
  state references
```

The normalized behavior registry must be stable/fingerprinted for replay. Changing the implementation bound to a behavior ID without changing the replay/content generation must not silently produce a different simulation.

## Start versus effect phases

The separate server-start and server-do function families are evidence that a skill can have distinct phases:

1. validate/initiate action;
2. enter animation/action mode;
3. fire an authoritative event at the appropriate action point;
4. perform the effect;
5. possibly schedule further effects;
6. complete/recover.

Dark Magic's future animation-event system should be able to align presentation with these semantic action events, but **gameplay timing must remain authoritative even without presentation assets**.

A headless server must be able to cast every supported skill.

## Target policies

The original runtime contains numerous target checks in shared skills and monster-spawn helpers. A future normalized target policy should distinguish at least:

- self;
- unit;
- hostile unit;
- friendly unit;
- point/ground;
- corpse;
- item;
- object/portal;
- direction/vector;
- area around caster;
- area around target;
- summon placement;
- movement destination.

Target validation may additionally need:

- range;
- line of sight;
- collision/footprint clearance;
- town restriction;
- target state/alive/dead restrictions;
- owner/party/hostility;
- item type/equipment state;
- corpse eligibility;
- summon limit;
- skill-specific exceptions.

The optional client `TargetID` is useful presentation context, but the server must re-resolve or validate targets from authoritative world state just as current interaction code already does.

## Mana and resource costs

Costs should be committed as part of the cast transaction, not deducted by Lua.

Research/implementation needs to distinguish:

```text
base skill cost
level scaling
calculation expression
cost modifiers
minimum/maximum behavior
resource availability
when the cost is charged
whether failed/interrupted actions refund
```

Some skills use non-mana resources or consume items/corpses/charges. A generic `ResourceCost` vocabulary should support those without pretending everything is mana.

Exact formula and rounding behavior remain a probe item.

## Skill delays and cooldowns

Diablo II contains skill delay/cooldown behavior distinct from animation duration. The engine should model it as authoritative future-tick eligibility.

Do not express cooldown as "wait until animation node finishes."

Suggested state:

```text
SkillCooldown
  skill/group identity
  availableTick
  source
```

Whether delay is per-skill, shared by a group, transformed by states, or modified by patch-specific rules must be verified before compatibility claims.

## Calculation expressions and synergies

The typed catalog already includes skill calculation records. Skills and missiles reference formula/calculation slots extensively.

Dark Magic needs one deterministic calculation boundary that can evaluate formulas from explicit inputs such as:

- skill level;
- base stats;
- effective stats;
- character level;
- synergies and prerequisite skill levels;
- missile level;
- difficulty;
- target stats when allowed;
- parameters from the data row.

Do not implement formulas ad hoc in every skill handler. The evaluator should be bounded, deterministic, diagnostic, and testable with source-located errors.

Unknown or unsupported formulas should fail at admission/definition validation where practical rather than halfway through a fight.

## Missile creation evidence

D2MOO's `MISSILES_CreateMissileFromParams` exposes a rich missile state model.

Observed inputs/derived properties include:

- owner and optional origin unit;
- missile class from `Missiles.txt`;
- spawn position and target position/unit;
- velocity from base velocity plus per-level scaling;
- fixed-point velocity representation;
- slow-missile state interaction;
- range plus per-level range and sub-loop adjustments;
- activation frame;
- collision type and separate footprint/movement collision masks;
- acceleration and maximum velocity;
- skill identity and skill level;
- source damage snapshot copied into missile stats;
- owner identity;
- pierce count determined from owner stats/RNG;
- to-hit bonus;
- damage frame rate;
- optional client synchronization facts.

This strongly argues that missiles should be **authoritative entities/state**, not particles.

A future Dark Magic missile entity likely needs:

```text
identity/class
owner/source
skill + skill level
position + fractional motion
path/movement type
velocity/acceleration/max velocity
remaining lifetime/range
activation tick
collision policy
hit memory / pierce budget
damage/stat snapshot or explicit dynamic-source policy
presentation identity
```

## Snapshot versus live-source missile damage

The 1.10f creation path calculates damage data when the missile is created and writes damage stats onto the missile. This is significant evidence that at least substantial portions of missile damage are **snapshotted at creation**.

Dark Magic should not automatically recompute all projectile damage from the caster's current equipment on impact. For each effect family, determine what is snapshotted and what remains dynamically queried.

This is especially important for:

- weapon swaps after projectile launch;
- buffs expiring in flight;
- skill-level changes;
- owner death/despawn;
- minion/trap attribution.

## Missile collision is not generic sprite overlap

Missile creation assigns:

- a footprint collision mask;
- a move-test collision mask based on `CollideType`;
- target/path information;
- optional destructibility;
- specialized path types for some missile families.

Missile collision belongs in authoritative world/collision code and needs semantic contact events.

A useful abstraction:

```text
MissileStep
  -> path movement
  -> terrain/object collision
  -> entity collision candidates
  -> collision policy
       ignore / hit / pierce / bounce / explode / stop / spawn child
  -> CombatResolution or behavior callback
```

Rendering the missile can lag/interpolate independently.

## Pierce

The 1.10f path derives a finite pierce count from owner pierce stats using deterministic RNG and caps the observed index loop at four increments.

Do not implement pierce as an infinite boolean. Preserve an explicit remaining-hit/pierce policy and verify skill-specific exceptions.

## States and stat-list sources

`States.txt` is only the vocabulary/metadata. Runtime state also needs an active source record.

D2MOO's state/stat-list behavior and the combat stun path provide evidence for state instances carrying:

- state ID;
- owner/source identity;
- attached stat modifications;
- expiration game frame;
- removal callback/behavior;
- possible stacking/refresh relationship;
- state mask semantics;
- persistence/death behavior.

Conceptually:

```text
UnitStateInstance
  ID
  Source
  AppliedTick
  ExpiresTick
  StatSourceID
  Parameters
```

A unit may need several sources contributing to the same effective stat while only one visible logical state is toggled. Do not collapse active states into `map[stateID]bool` if source/expiration semantics require more.

## State refresh and stacking

The combat stun path demonstrates refresh behavior: if a stun state already exists, its expiration may be updated/rescheduled instead of adding a duplicate.

Other states may stack, replace, keep the strongest source, keep the longest duration, or allow multiple independent sources. This must be researched by state/skill family rather than applying one universal rule.

The state engine should therefore support a policy layer instead of hard-coding "latest wins" globally.

## Auras and continuously applied states

Auras are better modeled as **owned periodic/proximity sources** than as permanent edits to every nearby entity.

Potential flow:

```text
AuraOwner state
  -> authoritative spatial query at defined ticks
  -> add/refresh source-tagged state/stat source on eligible units
  -> source disappears or unit leaves range
  -> source-tagged contribution expires/removes
```

This design also works for monster auras and some environmental effects.

Exact aura pulse cadence and update ordering remain probe items.

## Summons and owned units

Skill behavior should not create an unowned generic monster. Summoning must enter the owned-unit system described in [HIRELINGS_AND_OWNED_UNITS.md](HIRELINGS_AND_OWNED_UNITS.md): owner, pet type, limit/replacement policy, inherited/snapshotted stats, AI, and lifecycle are authoritative.

## Movement skills

Charge, Leap, Leap Attack, Whirlwind, Dragon Flight, teleport and similar skills should not bypass the world system by directly setting presentation coordinates.

They need explicit movement/effect policies that can:

- validate destination;
- select collision/path type;
- move over several ticks or atomically teleport;
- apply contact/hit rules;
- handle interruption;
- update authoritative facing/mode;
- emit presentation events.

The path research document records the original path-type vocabulary.

## Proposed Dark Magic boundaries

Possible package ownership:

```text
internal/game/skill
    catalog.go       normalized skill definitions / behavior registry
    cast.go          authoritative cast state machine
    calc.go          deterministic formula adapter
    target.go        target-policy validation

internal/game/state
    state.go         active source-tagged timed states
    policy.go        stack/refresh/removal rules

internal/game/missile
    state.go         authoritative projectile state
    movement.go      path/collision update
    impact.go        behavior/combat dispatch
```

These are candidate package names, not mandatory layout. Repository review should prefer focused existing owners where equivalent abstractions already exist.

## Suggested implementation slices

### S1 — consume `d2.player.skill_intent`

Build a deterministic cast-request stage that resolves the current authoritative assignment/learned level and clears or acknowledges intent exactly once.

No actual damage is required yet.

### S2 — generic cast transaction

Add target-policy validation, mana cost, start/complete tick state, and one basic melee or instant skill behavior.

### S3 — behavior registry

Normalize `SrvStFunc`/`SrvDoFunc`-style data dispatch into explicit trusted behavior handlers with coverage diagnostics. Unsupported behavior IDs must be visible in tooling.

### S4 — state engine

Implement timed source-tagged state application with expiration in authoritative ticks and checkpoint/replay coverage. Start with one refreshable state such as stun/chill.

### S5 — missile vertical slice

Implement one straight projectile end-to-end:

```text
cast -> missile entity -> collision -> combat -> removal
```

Then add pierce, acceleration, child/spawn behavior, and special movement families incrementally.

## Verification backlog

1. Exact mana-cost formulas, integer scale, and charge timing.
2. Exact skill-delay/cooldown semantics and shared delay groups.
3. Cast-speed/action-frame relationship versus server effect timing.
4. Interrupt rules from hit recovery, stun, block, knockback, death, movement.
5. Complete target policy flags in Skills.txt.
6. Formula evaluator opcode/parameter semantics and overflow/rounding.
7. Synergy ordering and soft-point/hard-point distinctions.
8. Weapon contribution and alternate-weapon snapshot behavior.
9. Aura update cadence, range metric, stacking, and source removal.
10. Curse overwrite/priority behavior.
11. Passive skill/stat source lifetime.
12. State persistence through death, level transition, save, shapeshift, and dispel.
13. Missile velocity/range fixed-point stepping and exact lifetime boundaries.
14. Missile collision-type table and collision-mask semantics.
15. Pierce count and hit-memory behavior.
16. Splash/explosion area queries and multi-hit prevention.
17. Homing/guided missile target loss behavior.
18. Missile damage snapshot versus dynamic owner lookup by family.
19. Trap/minion ownership and kill/proc attribution.
20. Corpse skill eligibility and corpse-consumption transaction ordering.
21. Summon limit/replacement policies by PetType.
22. Teleport/charge/leap/whirlwind collision and path-type rules.

## Primary sources inspected

- D2MOO pinned 1.10f `source/D2Game/src/SKILLS/Skills.cpp` and class/monster skill files.
- D2MOO `source/D2Game/src/MISSILES/Missiles.cpp`, `MissMode.cpp`, and missile/path structures.
- D2MOO `D2States.cpp`, `D2StatList.cpp`, and combat state application paths.
- Current Dark Magic `internal/game/session/skills.go`, movement/targeting code, and typed game-data catalog.
