# Movement, collision, pathfinding, and room-streaming research

Status: implementation-oriented research baseline. This document deliberately distinguishes **Dark Magic's current deterministic movement policy** from **Diablo II 1.10f compatibility behavior** recovered through D2MOO. The current implementation is already useful and should be extended rather than replaced wholesale.

## Executive conclusion

Dark Magic now has a good general navigation core:

- authoritative subtile coordinates;
- DT1-derived collision;
- radius-aware walkability;
- deterministic eight-way A*;
- no diagonal corner cutting;
- stopping rings around occupied interaction targets;
- semantic target coordinates admitted through the fixed-tick session;
- trusted zone-transition boundaries.

The next movement work is **not "implement pathfinding."** It is to add the movement policies Diablo gameplay needs above that foundation:

```text
ordinary player click path
monster chase/flee/circle
missile path
knockback
leap
charge
whirlwind
wall-follow
teleport
special encounter movement
```

The original engine used multiple path types and collision masks. Treating every behavior as ordinary A* would erase important gameplay semantics.

## Current Dark Magic implementation

`internal/game/world/navigation.go` currently exposes a deterministic `FindPath` over authoritative subtile collision.

Its behavior includes:

- start/goal converted to collision cells;
- radius-aware `walkableCell` checks;
- eight neighbors;
- cardinal cost 10 and diagonal cost 14;
- diagonal rejection when either orthogonal side is blocked;
- deterministic queue tie-breaking;
- goal acceptance within a configurable stopping radius;
- explicit unreachable errors.

`internal/game/session/movement.go` currently:

- admits `player.move` commands through the fixed tick;
- normalizes diagonal direct input;
- maintains run/walk intent;
- follows A* waypoints for click movement;
- cancels an unreachable target;
- derives authoritative facing/mode;
- uses the player collider radius when planning.

These are strong engine foundations and should remain generic, renderer-independent tools.

## Current fidelity gaps

Several values in the current player movement path are intentionally simple policy rather than verified Diablo behavior, including:

- hard-coded walk/run speeds (`10` / `15` in the current handler);
- target completion threshold around `0.2`;
- waypoint advancement threshold around `0.3`;
- one circle/radius abstraction for footprints;
- A* as the universal click-navigation algorithm;
- no stamina drain/recovery integration yet;
- no verified acceleration/velocity stepping;
- no special path types for combat skills/missiles/knockback;
- no authoritative dynamic-unit occupancy in ordinary route planning yet.

These should be made data/policy driven as verified rules arrive, not silently treated as compatibility constants.

## Original path system: multiple behavior families

D2MOO's 1.10f `Path.h` reconstructs **18 path types**:

| Type | Recovered meaning |
| --- | --- |
| IDA* | iterative-deepening A* path |
| A* | A* path |
| toward | direct/toward movement |
| unknown 3 | toward-like, apparently unused/uncertain |
| missile | missile path |
| monster circle CW | clockwise circle behavior |
| monster circle CCW | counter-clockwise circle behavior |
| straight | straight line with close-range A* fallback |
| knockback server | server knockback |
| leap | leap movement |
| charged bolt | special missile path |
| knockback client | client-side counterpart |
| backup/turn | turns on collision before destination |
| toward finish | special direct finish behavior |
| blessed hammer | specialized missile path |
| wall follow | wall-follow algorithm |
| missile stream | apparently unused/uncertain |
| unknown 17 | stubbed/unused uncertainty |

This is one of the most important findings in this workstream.

Dark Magic does **not** need to copy the old path object or force every path through exactly these numeric IDs. But it should model the semantic distinction explicitly.

A modern policy boundary could be:

```text
MovementIntent
  OrdinaryPath
  Direct
  CircleClockwise
  CircleCounterClockwise
  StraightThenNavigate
  Knockback
  Leap
  Charge
  Whirlwind
  Missile(strategy)
  Teleport
```

Each policy can reuse shared collision/coordinate utilities while owning its special update semantics.

## Original dynamic path state is richer than a point list

The reconstructed `D2DynamicPathStrc` stores concepts including:

- 16-bit fractional game position;
- target, previous target, and final target;
- current and previous room;
- current path point/index;
- path type and previous path type;
- unit size/collision pattern;
- footprint collision mask;
- movement-test collision mask;
- collided-with mask;
- optional target unit and identity;
- direction/new direction/direction delta;
- direction and velocity vectors;
- velocity, previous velocity, max velocity;
- acceleration and an acceleration counter;
- distance fields;
- saved steps;
- path-specific scoring fields.

This is strong evidence that movement state should not be reduced to `[]Point` plus a display animation.

Dark Magic can keep its simpler generic `FindPath` result while adding an authoritative **motion state** that owns velocity/path-policy progress separately from the route plan.

## Coordinate precision

The original dynamic path uses 16-bit fractional precision inside one game/subtile coordinate. Existing Dark Magic map research already records `1/65536` fractional units in original runtime terminology.

Dark Magic currently uses `float64` world/subtile positions. That is acceptable as an implementation choice only if determinism and compatibility rounding are controlled.

Before combat-sensitive movement is finalized, decide whether authoritative motion should use:

- explicit fixed-point integer coordinates;
- a deterministic rational/integer step representation;
- or tightly constrained float arithmetic with canonical checkpoint serialization.

Fixed-point has an important advantage: original collision/arrival boundaries become much easier to reproduce and test.

## Collision has two roles: footprint and move-test policy

The original path carries separate:

- footprint collision mask;
- movement-test collision mask;
- collision pattern/unit size;
- collided-with result mask.

This means "what space does this unit occupy?" and "what categories stop this movement?" are separate questions.

Dark Magic should eventually distinguish:

```text
Collider / footprint
  shape/pattern/size
  occupancy categories contributed to world

MovementCollisionPolicy
  terrain bits that block
  unit/object categories that block
  categories that can be crossed but trigger events
```

This is especially important for missiles, doors, flying units, charge/leap, summons, and corpses.

## DT1 collision remains the static terrain foundation

Current Dark Magic is correctly using semantic map collision rather than pixels.

Future movement must continue to consume:

- selected DT1 subtile flags;
- semantic world/object collision;
- dynamically occupied entity footprints;

and must not query rendered tile geometry.

The F3/F4/F5 world diagnostics that recently landed are valuable tools for validating this work visually, but diagnostics are observers, not simulation inputs.

## Dynamic entity occupancy

Ordinary navigation currently checks static map flags plus the mover's radius. Full gameplay needs dynamic occupancy policy.

Questions include:

- do monsters block players and each other during planning or only during motion resolution?
- can friendly owned units be pushed/passed?
- what is the collision behavior of NPCs in town?
- are dead bodies blocking?
- do missiles use entity collision in path planning, contact checks, or both?
- when multiple entities attempt to enter the same space on one tick, which stable ordering resolves the conflict?

Do not solve this with nondeterministic mutex/timing winner behavior. Fixed-tick entity movement must have deterministic conflict policy.

## Doors and objects

D2MOO path/monster-spawn code contains explicit door/path and collision checks, and monster AI has special tactics around world objects.

Dark Magic should model a door as authoritative world/object state affecting collision:

```text
closed door -> blocking semantic collision
open/opening door -> policy changes collision at authoritative tick
AI/player interaction -> explicit operate/open command
path follower -> may replan after state change
```

Do not make a visual door animation itself toggle collision.

## Player walk/run and stamina

`CharStats.txt` already provides class walk/run velocity and run-drain facts. These should eventually replace current hard-coded movement speed policy.

The full player locomotion model needs:

- base class velocity;
- walk/run mode;
- faster run/walk and slow effects;
- stamina amount/max;
- run drain;
- stamina recovery;
- cannot-run fallback when stamina reaches its threshold;
- state effects such as chill/freeze;
- terrain/skill-specific movement modifiers if present.

All of these are authoritative. HUD stamina is a snapshot of this system, not its owner.

## Movement action versus animation speed

Gameplay distance per tick and sprite/3D animation playback speed are related presentation concerns but should not be the same variable.

Recommended split:

```text
LocomotionState
  authoritative world velocity / path policy / facing

AnimationState
  logical mode WL/RN/etc.
  animation timing derived/configured for presentation
```

An animation can be retimed visually without changing server movement distance unless compatibility requires a specific event coupling.

## Knockback

The original path system has distinct server and client knockback path types. Knockback should therefore be a semantic forced-motion transaction, not `position += vector` inside damage code.

Future knockback needs to define:

- direction source;
- distance/steps;
- collision masks;
- blocked knockback behavior;
- simultaneous movement ordering;
- target mode/hit recovery interaction;
- client interpolation/correction.

Combat emits a knockback request; movement owns spatial resolution.

## Leap, charge, whirlwind, teleport

These skills need explicit movement strategies.

### Leap / Leap Attack

Questions:

- what collision is ignored while airborne?
- how is landing position validated?
- can terrain/walls be crossed?
- when does contact damage occur?
- what happens if landing becomes invalid?

### Charge

Likely needs direct/straight movement with target/contact behavior rather than ordinary point A*.

### Whirlwind

Needs continuous/path-following combat contact policy and must not simply teleport between A* cells.

### Teleport

Should perform authoritative destination validation and then atomically relocate the unit/room relationship. It should not consume an ordinary walking path.

These strategies belong above generic world collision APIs.

## Missile movement

The original missile creation path assigns collision type, footprint mask, move-test mask, velocity, acceleration, max velocity, target and specialized path behavior.

Missiles therefore need their own authoritative motion policy. See [SKILLS_STATES_AND_MISSILES.md](SKILLS_STATES_AND_MISSILES.md).

Do not plan a missile route using ordinary player A* unless a particular missile behavior explicitly requires pathfinding.

## AI movement

Monster AI needs movement intents such as:

- approach;
- flee;
- orbit/circle;
- back away/turn;
- follow owner;
- hold position;
- move to scripted map-AI point.

This maps well to the multiple original path types without requiring Dark Magic to expose raw legacy numeric IDs.

AI should select a semantic strategy; movement resolves it deterministically.

## Room activation and inactive units

D2MOO's `SUnitInactive.cpp` is strong evidence that rooms can stop owning fully active monster units while preserving enough state to reconstruct them later.

The inactive-monster record/restoration path preserves or reconstructs facts including:

- unit identity/class;
- position and level/room context;
- dead/alive mode;
- champion/unique/minion type flags;
- unique modifiers and name seed;
- alignment;
- AI control/minion information;
- experience;
- max/current HP;
- special monster parameters;
- selected unit flags;
- timing context such as the frame at inactivation.

The restore path may heal a monster to max HP when it has been inactive for a sufficiently long game-frame interval in the observed 1.10f code. This is exactly the kind of behavior that would be lost if room streaming simply serialized `{class,x,y,hp}`.

### Streaming architecture implication

Dark Magic should separate:

```text
Zone/room existence
Active simulation set
Inactive unit archive
Presentation residency
```

A render chunk becoming invisible must not define whether the authoritative room is active.

Likewise, inactive-room persistence belongs in session/game state and replay/checkpoints, not the VFS or renderer cache.

## Streaming policy should be deterministic

Research still needs the exact original activation radius/call order. Whatever Dark Magic policy chooses, activation/inactivation must not depend on host scheduling races.

A deterministic policy can be based on:

- player occupied rooms;
- graph distance / neighboring room set;
- stable transition sequence;
- fixed tick.

When a room deactivates, all state that affects future results must be captured canonically.

## Network prediction and correction

Current offline/local movement still passes through replayable authoritative commands, which is the right future network boundary.

Networking can layer:

```text
client predicted visual/local path
       |
       v
semantic target/input command
       |
       v
server fixed-tick authority
       |
       v
canonical position/path state
       |
       v
snapshot/correction
```

Do not make client path points trusted server state. The server should reproduce or validate path/motion from semantic intent and authoritative collision/content generation.

Special movement skills and knockback are especially important prediction/correction test cases.

## Recommended architecture

Candidate responsibilities:

```text
internal/game/world
    static collision + coordinate transforms + generic route planning

internal/game/movement
    MotionState
    semantic movement strategies
    deterministic dynamic occupancy resolution
    locomotion speed/stamina
    special forced movement

internal/game/streaming
    active-room policy
    inactive unit archive/restore
```

Package names are suggestions. Keep generic map facts in `world`; do not turn `world.Map` into a god object containing player stamina, monster AI and network prediction.

## Suggested implementation slices

### MV1 — locomotion data policy

Replace hard-coded player walk/run speed with verified `CharStats`/stat-driven movement and authoritative stamina drain/recovery.

Keep existing A* unchanged.

### MV2 — motion-state abstraction

Separate route plan from authoritative velocity/progress/path strategy. Preserve current ordinary movement behavior as one strategy.

### MV3 — dynamic occupancy

Add deterministic unit-footprint occupancy/contact resolution with synthetic multi-unit conflict tests.

### MV4 — forced movement

Implement knockback through movement authority, then one movement skill such as Leap or Charge.

### MV5 — room inactive archive

Prototype deactivation/restoration of one ordinary monster with deterministic state snapshot and no renderer dependency.

### MV6 — streaming policy

Activate/inactivate rooms around players through a deterministic graph policy and prove replay/checkpoint equality across activation boundaries.

## Verification backlog

1. Original path-type selection by unit/skill/AI family.
2. A*/IDA* neighbor order, score/tie-break behavior and maximum lengths.
3. Straight/toward/wall-follow algorithms and fallback rules.
4. 16-bit fractional position stepping and exact rounding.
5. Velocity/acceleration update cadence and unit conversion.
6. `CharStats` walk/run velocity conversion to world distance per game frame.
7. Run stamina drain/recovery cadence and threshold behavior.
8. Faster Run/Walk, slows, chill and state ordering.
9. Collision mask meanings for players, monsters, missiles, doors and objects.
10. Collision-pattern/size mapping from `MonStats2` versus radius approximation.
11. Dynamic unit occupancy and simultaneous conflict ordering.
12. Door open/path replanning behavior.
13. Knockback distance/collision/blocked result.
14. Leap/Leap Attack path and landing validation.
15. Charge path, target loss and collision.
16. Whirlwind travel/contact cadence.
17. Teleport destination validation and room transition semantics.
18. Missile path-type behavior and terrain/entity contact.
19. Monster circle/back-up/wall-follow movement semantics.
20. Hireling/pet owner-follow teleport/leash policy.
21. Original active-room radius and activation/inactivation timing.
22. Complete inactive-unit serialized state by monster/item/object type.
23. Long-inactive monster healing behavior and thresholds across patches.
24. Corpse/inactive-unit lifecycle.
25. Server/client prediction protocol and correction thresholds.

## Primary sources inspected

- D2MOO pinned 1.10f `source/D2Common/include/Path/Path.h` and path algorithm files (`AStar`, `IDAStar`, `PathWF`, `Path`, `PathMisc`, `Step`).
- D2MOO missile creation/path configuration.
- D2MOO `source/D2Game/src/UNIT/SUnitInactive.cpp` for room/inactive monster restoration.
- Current Dark Magic `internal/game/world/navigation.go`, `internal/game/session/movement.go`, generated map/collision/transition architecture and recent world diagnostics.
