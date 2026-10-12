# Movement, collision, pathfinding, and room-streaming research

Status: implementation-oriented research baseline from owned Expansion 1.14d
evidence. Dark Magic targets 1.13c and 1.14d feature parity, but 1.13c movement
equivalence remains unverified. This document distinguishes **Dark Magic's
current deterministic movement policy** from older behavior recovered through
D2MOO. That 1.10f code is secondary evidence only and must be revalidated
where either target patch may differ. The current implementation is already
useful and should be extended rather than replaced wholesale.

## Executive conclusion

Dark Magic now has a good general navigation core:

- authoritative subtile coordinates;
- DT1-derived collision;
- radius-aware walkability;
- deterministic eight-way A*;
- no diagonal corner cutting;
- stopping rings around occupied interaction targets;
- semantic target coordinates admitted through the fixed-tick session;
- trusted zone-transition boundaries;
- explicit invalidation of pointer routes when their world/navigation changes.

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

`internal/mod/d2legacy/adapter/movement/movement.go`, the authoritative
`d2legacy.commands.move_player` handler, and `d2legacy.player.resolve_motion`
currently:

- admits `player.move` commands through the fixed tick;
- normalizes diagonal direct input;
- maintains run/walk intent;
- follows A* waypoints for click movement;
- cancels an unreachable target;
- preserve semantic motion ownership through replay/checkpoint restore;
- derive authoritative velocity and locomotion mode in one ordered resolver;
- uses the player collider radius when planning.

Route targets are scoped to the map that planned them. Cross-level relocation
replaces navigation and atomically drops the client target, accepted waypoints,
pending presentation selection, and held hostile selection. Camera interpolation
also snaps to the relocated player before the destination map accepts pointer
projection. None of those presentation-side facts is authoritative relocation
state.

These are strong engine foundations and should remain generic, renderer-independent tools.

## Current fidelity gaps

Several values in the current player movement path are intentionally simple policy rather than verified Diablo behavior, including:

- target completion threshold around `0.2`;
- waypoint advancement threshold around `0.3`;
- one circle/radius abstraction for footprints;
- A* as the universal click-navigation algorithm;
- no verified acceleration/velocity stepping;
- no special path types for combat skills/missiles/knockback;
- no authoritative dynamic-unit occupancy in ordinary route planning yet.

These should be made data/policy driven as verified rules arrive, not silently treated as compatibility constants.

## Original path system: multiple behavior families

D2MOO's older `Path.h` reconstructs **18 path types** that provide a candidate taxonomy for 1.14d verification:

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

The first motion-resolution boundary is now implemented. `world.collider`
continues to describe continuous footprint radius, while a separate
`world.occupancy.blocks_movement` fact decides whether that footprint blocks
ordinary unit motion. Players and living monsters carry both facts; inactive
monster archive/restore retains them. The Lua movement phase processes stable
ECS order and consults each other same-level unit's current or already-committed
position during both axis steps. Two contenders under this Lua resolver cannot swap through
one another or enter the same footprint, and an explicitly nonblocking unit is
passable. Admission and warp placement can temporarily overlap footprints; a
unit may move only when its candidate position strictly increases separation,
preventing that transient condition from trapping locomotion. Static DT1
collision remains an independent input.

Blocker discovery includes active position/collider/occupancy footprints even
when they lack velocity or movement bounds. This includes stationary bodies
and monsters moved by `engine.world.velocity_mover`; only entities with
velocity and bounds are integrated by the Lua resolver. The native monster
mover still checks terrain without reciprocally reserving unit footprints.
Ordinary A* still plans against static terrain. Town interaction targets are
metadata without spatial bodies, so the NPC observations below do not yet
constitute NPC materialization or category-policy integration.

Remaining target questions include:

- do monsters block players and each other during planning or only during motion resolution?
- can friendly owned units be pushed/passed?
- what are the remaining town NPC subcategory rules beyond the holdouts below?
- are dead bodies blocking?
- do missiles use entity collision in path planning, contact checks, or both?
- when multiple entities attempt to enter the same space on one tick, which stable ordering resolves the conflict?

The remaining planning and category rules must extend this deterministic fixed-
tick policy; they must not introduce mutex/timing winners.

### Owned town NPC holdouts

Unchanged owned LoD 1.14d on Wine 11.18 Staging was observed on 2026-10-11,
with `Game.exe` SHA-256
`631066c1649c4ea9ffe48bf97e24c00bca1f7a6759c21150f1a79982589adaaf`. A
synthetic level-one Sorceress retained the stock Barbarian equipment/quest
tail. Ordinary ground clicks and passive client-unit path reads distinguish
two town NPC cases; raw data and imagery remain private. The immutable receipt
has SHA-256 `e258cd2282ba21cc3f0c5490be013cd97751e067ad65202257a6e5e91686fdc4`.

The stationary Rogue guard, identified by owned `MonStats` index 152 and unit
ID 2, stayed at `(4516.5, 4626.5)`. A running player crossed its center, reaching
a sampled separation of 0.4463 subtiles before continuing beyond it. Kashya,
index 150 and unit ID 3, stayed at `(4498.5, 4633.5)` after an ordinary dialogue
was closed. Ground clicks across her position from both sides produced detours:
both traces reached a closest sampled separation of `sqrt(2)`, with a route
point two cells south of her center. A control crossing five cells south kept
its Y coordinate constant throughout, excluding a systematic projection bend
without excluding every local terrain confound. These are trajectories, not a
recovered planning algorithm or a proof of exact step-collision ordering.

A passive room-grid snapshot showed `0x100` in the four cardinal neighbors of
each NPC. Kashya's center was `0x1100`, while the guard's was `0x2100`; nearby
cells were clear. Those distinct raw masks support the behavioral distinction
without assigning universal meanings to their bits. Both owned rows had
`Align=1`, `soft=1`, and a 2-by-2 size, so those facts alone do not explain the
difference. Other NPCs, actual player/hireling/summon categories, and whether
each category participates in initial planning or later resolution remain
holdouts. A circle-radius approximation must not be presented as the measured
retail footprint.

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

The production locomotion catalog now pins starting Vitality, `WalkVelocity`,
`RunVelocity`, starting stamina, `RunDrain`, `StaminaPerLevel`, and
`StaminaPerVitality` from the mounted Expansion 1.14d `CharStats.txt` generation
for all seven classes. There is no longer a hard-coded player walk/run rate.

Current/max stamina is authoritative 8.8 ECS state. One shared pure rule drives
the Lua authority and connected prediction: wilderness running drains twice the
class RunDrain value per 25 Hz tick before the torso and slower-drain channels;
idle recovery adds max/256 raw units, movement recovery adds max/512, town
running does not drain, and exhaustion changes the active movement to walking.
The equipped-stat graph also exposes generic `velocitypercent`,
`item_fastermovevelocity`, recovery, slower-drain, and armor-penalty sources.
Item FRW uses the recovered 150-point diminishing-return conversion before it
joins the velocity channel. Walk begins at 100%; run begins at the integer
`100 * RunVelocity / WalkVelocity` ratio; the diminished item term and additive
skill/state/armor terms then join before the final 25% floor. This one integer
rule is shared by authority, prediction, and player locomotion presentation.

Raw `CharStats` velocities are not subtiles per second. After resolving the
percentage, the shared rule truncates `WalkVelocity * 256 * percentage / 100`,
then converts that velocity to subtiles per second with `25 / 4096`. For the
owned 6/9 class facts, baseline walk/run is 9.375/14.0625 subtiles per second.
The prior direct 6/9 interpretation moved at 64% of the measured target rate.

### Owned positive-speed holdouts

On 2026-10-11, the unchanged Windows LoD 1.14d reference under Wine was sampled
through ordinary ground movement and Burst of Speed casts. Two private,
synthetic Assassin saves supplied skill ID 258 at ranks 1 and 20, with raw
level/XP and the selected skill chain checked before casting. They retain the
stock Barbarian equipment and quest/item tail. A second candidate unit and
the client agreed on rank-one steps; the candidate's server role remains an
inference. Rank 20 was sampled from the client unit.

The immutable private receipt `speed-evidence-index-001.json` has SHA-256
`ab0d6445f9414d0e6e49aff0cbf50e67dbb948c0d27f87ea03fce92b3721f84b`.
It pins the fixtures, closed watches, ordinary inputs, selected-skill reads,
screenshots, source hashes, exclusions, and observed stat transitions.
Consecutive straight-axis samples had these dominant 16.16 position deltas:

| State | Walk percentage / step | Run percentage / step | attackrate / other_animrate |
| --- | --- | --- | --- |
| Baseline and after expiry | 100 / 24576 | 150 / 36864 | 100 / 100 |
| Burst rank 1 | 123 / 30224 | 173 / 42512 | 121 / 121 |
| Burst rank 20 | 161 / 39552 | 211 / 51840 | 152 / 152 |

Dominant-step intervals were approximately 40 ms. All six deltas fit the
integer velocity conversion above. The run observations distinguish adding
the skill contribution from multiplying the 150% run baseline by it. Raw
speed stats returned to baseline approximately 120 and 348 seconds after the
respective effects began, followed by baseline movement repeats. Passive reads
are non-atomic; observation timestamps do not identify exact instructions.

The corrected rate also requires close-target protection: authority caps a
waypoint step by the remaining distance and actual tick duration. Prediction
re-resolves retained pending target input each replay step, and exhaustion
preserves a smaller target-capped velocity. These are regression safeguards,
not proof of retail arrival or target redirection policy. Prediction seeded
only from a HUD still lacks the authority's retained target, and the connected
client currently supplies a 40 ms step.

This advances #375's positive movement evidence only. It does not enable
Burst of Speed content, prove attack timing, settle extreme modifiers or
cold/freeze policy, or establish bit-exact diagonal/terminal position rounding.

Owned Expansion 1.14d `Armor.txt` rows now pin representative zero/five/ten
speed penalties for both torso armor and shields, and equipment coverage proves
separate equipped pieces stack. A checkpointed generic timed `cold` source
exercises the recovered player `velocitypercent = -50` contribution together
with skill, armor, and item FRW sources and removes it exactly on expiry. This
validates the generic mechanism; it does not yet claim the complete target
runtime cold/freeze policy.

Maximum stamina is now reconstructed in the same shared Go rule used by Lua
authority. `CharStats.stamina` is shifted to 8.8, while per-level and allocated
base-Vitality terms are quarter values (`<< 6`). The target-owned
`ItemStatCost.txt` graph pins bonus Vitality op 9, active/passive skill stamina
percent op 1, and `item_stamina_perlevel` op 2 with level base and parameter 3;
direct `maxstamina` is a whole-point property shifted to 8.8. Op-derived skill
percentages use the direct max value, while bonus Vitality and per-level item
terms remain separate graph contributions.

The world also retains one checkpointed environment cycle per active act.
Normal time advances one tick, Act III night advances eleven, and Act IV
advances sixteen; the recovered cycle boundaries produce the base time consumed
by Properties func 18. Its packed period and signed 10-bit min/max values use
15-unit rounding over a wrapped 360-unit cycle and linear interpolation from the
center maximum to the opposite minimum. Owned Expansion 1.14d tables pin
`item_stamina_bytime` ID 295/op 6 and `stam/time` func 18. The dormant eclipse
facts are not activated until an owned Tainted Sun transition vector exists.

When an active max source changes, positive current stamina is multiplied by
new/old maximum through the recovered double calculation, truncated, and
clamped to `[1,newMax]`; zero remains zero. Level-up is a distinct transaction
and fills the newly derived maximum. The live stamina-progression component
retains admitted base Vitality and the previous progression boundary across
checkpoint/replay. Equipment source activation/removal and the owner-private
HUD therefore observe the same authoritative raw values.

The archive and declarative record identities are target-locked to Expansion
1.14d. Arithmetic is corroborated by recovered executable structure and
independent community measurements. Remaining target work is owned-runtime
boundary vectors for chill/freeze classification and immunity/duration policy,
the Tainted Sun environment transition, and allocation-command integration of
the one-point transition below.

The remaining player locomotion model needs:

- live stat allocation and the Tainted Sun eclipse transition;
- owned 1.14d extreme-modifier and cold/freeze target vectors;
- cold/freeze resistance, immunity, duration, difficulty, and action-rate effects;
- terrain/skill-specific movement modifiers if present.

All of these are authoritative. HUD stamina is a snapshot of this system, not its owner.

## Straight-route waypoint cadence

Player route commands skip only same-direction collinear intermediate stops.
Turns, reversals, the actual fractional start approach, and the final two
nodes remain. The penultimate node anchors sub-cell held-pointer endpoint
updates, so those updates cannot redirect an entire long segment.

`TestStraightRouteMatchesDirectMovementCadence` compares the production command
source, movement resolver, and collision integrator with a direct target at the
same configured rates. Its high-rate synthetic case distinguishes the old
per-cell route, which can reverse after overshooting intermediate nodes.
`TestOwnedPointMovementMaintainsConfiguredRoutePace` checks consecutive full-rate
steps and arrival through ordinary admission, pointer intent, managed Lua
motion, and collision integration with owned data at 25 Hz.

These checks establish route cadence at the configured scale. Matching native
1.14d distance per tick additionally requires the movement-scale conversion and
its target-distance clamp. Historical combined application traces showed
running steps alternating between 0.5625 and 0.4375 subtiles before intermediate
stops were removed. Those traces do not establish native-rate parity on a branch
without that conversion. Native directional observations remain qualified
context, not proof of retail path selection, corner timing, mouse projection,
or bit-exact diagonal/terminal rounding. This advances #375 only partially.

## Movement action versus animation speed

Gameplay distance per tick and sprite/3D animation playback are separate state,
but Expansion locomotion derives both rates from the same effective velocity
percentage. Integrated distance never advances the animation clock.

Recommended split:

```text
LocomotionState
  authoritative world velocity / path policy / facing

AnimationState
  logical mode WL/RN/etc.
  AnimData frame count/events
  WL/RN runtime base 213/101 scaled by effective velocity percentage
```

The retained player playback integrates presentation/network time. A raw
position correction or collision-shortened displacement does not retime it;
FRW or chill does. Rate changes preserve current frame phase rather than
retroactively applying the new rate to all time since mode start.

## Owned one-point Vitality allocation observations

Issue #377's arithmetic boundary was measured on 2026-10-11 in unchanged owned
Windows LoD 1.14d, under Wine 11.18 Staging. The reference `Game.exe` SHA-256 is
`631066c1649c4ea9ffe48bf97e24c00bca1f7a6759c21150f1a79982589adaaf`.
Ordinary mouse/keyboard input spent each point. Passive `ReadProcessMemory`
observations retained raw 8.8 values, base/effective vectors, position, mode,
UTC, and executable/tool identities. Saves, screenshots, and raw recordings
remain in the private corpus. The immutable allocation-index receipt has
SHA-256 `f2fc3d22abe843c9a87625a8b83b9ed48ab27fc11b3d19915d33a84f28328b44`;
it identifies the five completed watches and their observed point transitions.

The local client quantizes stamina and updates it less frequently. A second
player-unit candidate, independently found in committed writable memory,
matched the player identity, progression, position and mode, and exposed the
ordinary per-frame 40-unit run drain. Its server role is a strong inference,
not a proved structure label. Stat records were read twice and headers
rechecked; these observations are not an atomic instruction trace.

These are observed adjacent samples, not pre-adjusted test operands:

| Case | Vitality / unspent points | Current raw, before → after | Maximum raw, before → after | Sampling qualification |
| --- | --- | --- | --- | --- |
| Full, normal life | 25/3 → 26/2 | 23552 → 23808 | 23552 → 23808 | stationary in town |
| Partial | 26/9 → 27/8 | 18168 → 18619 | 23808 → 24064 | same position and running mode |
| Zero | 27/8 → 28/7 | 0 → 303 | 24064 → 24320 | walking; next samples 350, 397 |
| Zero, normal life | 27/1 → 28/0 | 0 → 256 | 24064 → 24320 | walking; next frame 303, then 350 |
| Partial, 30% passive | 25/10 → 26/9 | 29154 → 29652 | 34278 → 34611 | one running frame between positions |
| Partial, normal life | 26/2 → 27/1 | 20855 → 21294 | 23808 → 24064 | one running frame between positions |
| Partial, Assassin | 22/1 → 23/0 | 23340 → 23959 | 24960 → 25280 | same position and running mode |

Fixtures added unearned points to a clone of an owned save. The first partial
and zero cases also used synthetic high life to isolate movement. The passive
case used a synthetic level-12 Barbarian with Increased Stamina rank one.
The normal-life holdout preserved the stock level, life, mana, and equipment,
and independently repeated full, partial, and literal-zero allocation. A second
class holdout loaded an edited level-one Assassin with its authored starting
attributes/resources and three unearned points, retaining the stock Barbarian
equipment/quest tail as a declared synthetic-fixture confound. Its full and
partial cases confirm the 5-quarter class gain (320 raw), including
`trunc(23340 / 24960 * 25280) + 320 = 23959` at an unchanged position.
Loading a save with depleted stamina refilled it; those nominal depleted-save
trials are excluded from the allocation evidence. Depletion above came from
ordinary wilderness running.

An additional class cohort has immutable evidence-index SHA-256
`c11bd3c20fe4f6b8c71c91ac7fbd60d2a156e53038c79d97ec960e605f939794`.
Its level-one fixtures likewise add three unearned points and preserve the
stock Barbarian equipment/quest tail. All seven classes now have observed full
allocation cases, combining this cohort with the earlier Barbarian and Assassin.

| Class / case | Vitality / unspent points | Current raw, before → after | Maximum raw, before → after |
| --- | --- | --- | --- |
| Amazon, full | 20/3 → 21/2 | 21504 → 21760 | 21504 → 21760 |
| Amazon, partial | 21/2 → 22/1 | 16210 → 16656 | 21760 → 22016 |
| Amazon, zero | 22/1 → 23/0 | 0 → 256 | 22016 → 22272 |
| Sorceress, full | 10/3 → 11/2 | 18944 → 19200 | 18944 → 19200 |
| Necromancer, full | 15/3 → 16/2 | 20224 → 20480 | 20224 → 20480 |
| Paladin, full | 25/3 → 26/2 | 22784 → 23040 | 22784 → 23040 |
| Druid, full | 25/3 → 26/2 | 21504 → 21760 | 21504 → 21760 |

The Amazon rows use adjacent samples from a separately found player candidate,
at unchanged positions and modes. Its server role remains inferred. The partial
case distinguishes `trunc(16210 * 22016 / 21760) + 256 = 16656` from adding
before rescaling, which produces 16659. The zero watch observes 256 before later
walking recovery. Wider before/after action receipts include drain or recovery
and do not supply these exact arithmetic operands. The other four full cases
use repeated reads of the same active client unit, followed until current and
maximum settle together. They are not independent authority observations.

The partial same-position vector distinguishes the arithmetic order:

```text
rescale only:             trunc(18168 / 23808 * 24064)       = 18363
rescale, then class gain: trunc(18168 / 23808 * 24064) + 256 = 18619
class gain, then rescale: trunc((18168 + 256) / 23808 * 24064) = 18622
```

The resulting one-point transition is:

```text
newMaximum = maximum_stamina(new base Vitality, active sources)
rescaled = rescale_stamina(oldCurrent, oldMaximum, newMaximum)
newCurrent = min(newMaximum, rescaled + CharStats.StaminaPerVitality * 64)
```

Maximum changes therefore retain the existing ratio/truncation callback, then
allocation adds the class's unmodified quarter-point gain. The passive case
changes maximum by 333 raw but grants 256: after the intervening running drain,
`trunc(29114 / 34278 * 34611) + 256 = 29652`. The normal-life partial similarly
uses 20815 after its intervening drain. The zero case's first observed 303 is
consistent with a 256 allocation gain followed by one 47-unit walking recovery
frame, then the observed +47 sequence. The normal-life zero holdout directly
observed the intermediate 256 before the next walking frame produced 303.
Full stamina caps at the new maximum.

The allocation evidence suite composes the existing production
`d2legacy.movement_rules/v1.maximum_stamina`, `rescale_stamina`, and `class_facts`
capabilities with the observed one-point gain. It checks all seven full cases,
the recorded partial/zero cases, class rounding, and the passive-bonus ordering.
Generic equipment/source changes remain on `rescale_stamina`, where zero stays
zero. The existing calculations remain correct for those roles.
A future allocation transaction must apply the formula exactly once, commit
maximum and current together, and avoid applying the generic callback again.
This is a research foundation for feature parity, not a live allocation API,
earning policy, command, or UI.

The arithmetic order is established for these cases. Exact CPU instruction
order, simultaneous allocation/equipment changes, and bulk allocations remain
separate holdouts. Partial and zero cases for Sorceress, Necromancer, Paladin,
and Druid, and literal zero for Assassin, remain unmeasured.

## Knockback

The original path system has distinct server and client knockback path types. Knockback should therefore be a semantic forced-motion transaction, not `position += vector` inside damage code.

The generic transaction boundary is now implemented. A checkpointed
`world.forced_motion_request` carries a semantic kind plus source point,
policy-owned distance, speed, and request tick. The pre-simulation admission
system derives a destination away from the source; the movement phase advances
that destination through the same static collision, dynamic occupancy, bounds,
and stable entity ordering as ordinary locomotion. Durable events distinguish
completed, partial, blocked, invalid, and replaced outcomes. A newer admitted
request deterministically replaces one active transaction only after recording
the displaced request's exact progress. Forced motion owns velocity while
active, clears it on completion, and ordinary locomotion resumes only from
fresh input. Active progress survives a checkpoint, and presentation may
observe the event without authoring position. Every outcome uses the target's
stable selectable `player:` or `monster:` identity when one exists, so client
playback and replay consumers do not need to correlate transient ECS IDs.

D2MOO's recovered 1.10f/1.13c server path is useful structural evidence: it
selects a dedicated server knockback path, derives its ray away from the target
unit, clips that ray against collision, and applies a dedicated velocity. Those
older constants are deliberately **not** embedded in Dark Magic's target policy.
Expansion 1.14d evidence still needs to define:

- direction source;
- distance/steps;
- collision masks;
- blocked knockback behavior;
- simultaneous movement ordering;
- target mode/hit recovery interaction;
- client interpolation/correction.

Owned Expansion 1.14d records now pin the target facts which do not require
binary inference. `MonStats2.mKB` declares whether a monster has a knockback
mode, `dKB` declares its direction count, and `small`/`large` provide a distinct
size category. Dark Magic joins those columns into a generic checkpointed
monster knockback-target profile; representative small `fallen1`, large
`bighead1`, and mode-incapable normal `gorgon1` rows are asserted directly from
the mounted MPQs. `Properties.knock -> item_knockback` and its
`ItemStatCost` melee/missile event hooks are also pinned. Their event-function
chance arithmetic remains binary-owned and therefore unimplemented.

The owned Expansion 1.14d `Missiles.txt` table independently authors a
`KnockBack` byte. Representative rows pin blank (`firebolt`), `1`
(`leapknockback`, `moltenboulder`), `33` (`amphibiangoo1`), and `75`
(`baal cold maker`). The generic straight-missile decoder now preserves that
raw byte in each checkpointed projectile. It deliberately does not treat the
value as a boolean or percentage and does not emit forced motion: recovered
older code shows a random-threshold/result-flag structure, but only a 1.14d
binary probe may promote that structure into current gameplay policy.

The older recovered source hypotheses are now recorded precisely rather than
as folklore. In D2MOO's 1.10f/1.13c reconstruction, item event function 7 rolls
an attacker-seeded value masked to 0..127 against 128/64/32 thresholds for
small/normal/large targets; the missile path rolls its own seed modulo 100
against the raw `Missiles.KnockBack` byte. The damage-mode path prioritizes
death before knockback, gives players and monsters dedicated KB modes, converts
a monster result to GH when its `MonStats2` record lacks KB, and uses a recovered
five-unit path distance for ordinary players/monsters (with a sand-leaper
exception). Knockback mode installs a dedicated 4096 path velocity. These are
**older recovered hypotheses only**, not Expansion 1.14d implementation values.

`internal/dev/tools/knockback_probe` turns the remaining target-runtime work
into a strict capture contract. It accepts only
`diablo-ii-lod-1.14d-expansion` observations from an owned runtime, requires a
matched neutral control, records target category/record/size/KB-mode support,
separates missed, blocked, lethal, and uninterruptible trials, and normalizes
reaction, displacement, blocked-motion, observed-rate, and Wilson-interval
facts. Older size-weighted and raw-byte-percent rules are emitted only as
explicitly labeled candidates beside an always/nonzero alternative. Cases
without an observable KB mode receive no chance inference. Start from
`docs/research/probes/knockback-lod-114d-expansion.template.json` and run:

```shell
go run ./internal/dev/tools/knockback_probe -input /path/to/sanitized-capture.json
```

The required matrix includes players, KB-capable small/normal/large monsters,
mode-incapable monsters, hirelings, summons, NPCs, and corpses; open-space and
collision-limited pairs; item and representative missile bytes; and
nonlethal/lethal plus interruptible/uninterruptible pairs. Combat emission
remains disabled until owned observations resolve those distinctions.

Once those target rules are pinned, combat will emit the generic request;
movement already owns spatial resolution.

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

The first exact-ID family now does that for Teleport ID 54: a shared
point-relocation system checks current-level bounds, static footprint collision,
and blocking ECS occupancies, then mutates the existing position while stopping
semantic/raw/forced motion. One generic relocation-result ECS entity records
success or the rejected outcome. Owned Levels.txt values are decoded as 0/1/2;
the sole policy-2 level receives a conservative BlockLOS trace. Exact
viewport/range, policy-2 meaning, invalid-target payment and nearest-free
fallback, room-edge order, owned-unit following, and animation timing remain
target-1.14d probes.

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

The older recovered restore path may heal a monster to max HP after a sufficiently long inactive interval. This is unresolved for expansion 1.14d, but demonstrates the kind of behavior that would be lost if room streaming simply serialized `{class,x,y,hp}`.

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

Current implementation status: the population authority evaluates one
deterministic occupied-room-plus-immediate-neighbors set each tick. Its v5 room
plan records world-owned stable resident IDs plus each entity's existing
velocity-mover opt-in while an empty ECS marker excludes the same retained
entity from simulation and presentation. The generic velocity-movement opt-in
is removed and conditionally restored at the activation boundary. The full
component state plus timed-state/stat-source/event references therefore use the
ordinary ECS checkpoint instead of a parallel scalar archive. A checkpoint made
while inactive resumes to the same entity IDs and checksum. This establishes
the first type-independent persistent-identity mechanism; a non-monster resident
without movement opt-in crosses the same checkpoint/reactivation path, while an
equal room ID in another level remains active. Generated room/link IDs are
canonical strings, and production DS1 interaction points acquire residency from
zone room bounds. An active positioned resident resolves those bounds again
before every activation decision; a non-monster regression crosses from one
room to the next and stays active with its current room after the origin room
deactivates. A synthetic owned resident now proves that its direct ECS owner
reference plus category, limit, lifetime, durable identity, and attribution
survive the same inactive checkpoint/reactivation path. Owned-unit lifetime
updates exclude the empty inactive marker and apply an unchanged absolute
expiration on the first active tick. This does not claim that graph distance,
phase ordering, timer aging, healing, corpse policy, or separate presentation-
only residency matches expansion 1.14d.

An ordinary corpse is now a second explicit residency kind. The death
transaction removes AI, collider, selection, and the generic velocity-mover
marker before room deactivation. Its inactive record therefore pins
`velocity_mover=false`, and checkpoint reconstruction/reactivation preserve the
semantic corpse without restoring a movement surface. Corpse lifetime and
skill/object eligibility remain outside this streaming slice.

A production-cast straight projectile is now a third explicit residency kind.
Missile materialization asks the installed population plan to resolve its
authoritative level/position into the canonical room and attaches a stable
world-owned resident ID. The shared room synchronizer and empty inactive tag
then govern it like any other entity; the projectile movement/lifetime query
excludes inactive residents. A checkpoint acceptance test proves unchanged
position, remaining lifetime, and projectile/location/residency components
through deactivate, restore, and reactivate. Its deliberately long lifetime is
synthetic test scaffolding, not evidence for Expansion 1.14d inactive missile
aging or expiry rules.

An imported world-container item is now a fourth residency kind and the first
entity that can predate the room plan. It carries generic position/location and
a transient `d2legacy.world.room_attach` ECS request rather than an authored
room ID. Plan admission resolves the point and replaces that request with the
ordinary stable resident; pickup removes the spatial/residency components and
re-drop resolves them again. The same inactive marker filters item projections
and preserves the original entity across checkpoint/reactivation. The fixture
does not establish public ground ownership, legal drop/pickup policy, or target
1.14d item lifetime.

A synthetic stateful interaction object and its pending-action relationship now
close the first mechanism matrix. Object mode/used/seed/revision facts remain on
the target entity, while the pending action carries a raw ECS target reference
and its own stable room resident ID. Both deactivate/checkpoint/reactivate
together without inferring relationship lifetime or creating an object archive.
The action's due tick is preserved but not executed; target Expansion 1.14d
event timing and inactive aging remain unresolved.

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

### MV1 — locomotion data policy (implemented foundation)

Pinned `CharStats`/stat-driven movement and authoritative stamina drain/recovery
now replace the hard-coded rate. Level/Vitality/direct/skill/item-per-level
maximum stamina and proportional source transitions are also authoritative.
Finish the explicitly listed Expansion 1.14d runtime probes before widening
time-of-day or slow modifier content.

Keep existing A* unchanged.

### MV2 — motion-state abstraction (implemented player foundation)

Client A* owns only replaceable world-scoped waypoints. Admitted directional or
waypoint locomotion and authoritative melee approach claim the checkpointed
`d2legacy.player.motion` fact; one ordered resolver converts its semantic owner,
kind, direction/target, and requested run mode into class/stat-derived velocity,
WL/RN mode, and stamina execution. Explicit locomotion replaces attack approach,
exhaustion downgrades run intent, and relocation clears ownership. Raw velocity
is now an execution output rather than the signal used to infer which player
action owns movement. Player composite playback consumes `AnimData.d2`
frame/event facts and presentation/network time, while WL/RN rate uses the
recovered 213/101 base scaled by the same effective velocity percentage as path
motion. Integrated displacement does not drive that clock; FRW/chill retime it
without resetting phase. Extend the strategy vocabulary for forced motion in
MV4.

### MV3 — dynamic occupancy

Deterministic same-level unit-footprint occupancy/contact resolution now covers
simultaneous contenders, explicit nonblocking policy, archive/restore, and
checkpoint parity, including escape from a pre-existing admission/warp overlap.
Stationary and native-moved blocking footprints are included in Lua-controlled
step checks without acquiring Lua movement ownership.
Extend it with target-verified unit categories and decide which dynamic
footprints affect A* planning before calling MV3 complete.

### MV4 — forced movement

Generic checkpointed forced movement now advances policy-supplied distance and
speed through movement authority with completed, partial, blocked, and invalid
outcomes. Recover the Expansion 1.14d knockback policy that emits it, then extend
the same strategy vocabulary for movement skills without implementing a skill
as a standalone special case.

### MV5 — room inactive residency (persistent-identity slice implemented)

One ordinary generated monster now deactivates/reactivates through an empty ECS
filter tag without losing entity identity, component state, or timed-state/stat-
source/event target references. Room identity itself is world-owned and type-
independent; a non-monster/non-moving resident proves conditional activation-
surface restoration, while a moving non-monster proves current-room membership
replaces spawn-room membership before activation. The path is deterministic,
checkpointed, and renderer-independent. Production resolved DS1 interaction
targets now attach to this boundary. An owned resident's relationship graph and
lifecycle filter also cross it without a second archive. An ordinary corpse
retains its semantic entity while all live movement surfaces remain absent;
a straight projectile attaches at materialization and suspends movement and
lifetime under the same inactive marker without a missile archive. Stateful
object state and a separately resident pending-action relationship now preserve
their graph across the same path. A synthetic imported ground item proves
generic pre-plan attachment plus pickup/re-drop spatial transitions; the public
generated-loot lifecycle and exact target streaming policies remain separate.

### MV6 — streaming policy (synthetic foundation implemented)

The current all-player occupied-room-plus-immediate-neighbors policy activates
and inactivates rooms deterministically and proves replay/checkpoint equality
across the first monster boundary. Its exact 1.14d radius and timing remain a
probe rather than a compatibility claim.

## Verification backlog

1. Original path-type selection by unit/skill/AI family.
2. A*/IDA* neighbor order, score/tie-break behavior and maximum lengths.
3. Straight/toward/wall-follow algorithms and fallback rules.
4. 16-bit fractional position stepping and exact rounding.
5. Velocity/acceleration update cadence and conversion for other unit families.
6. Fixed-point diagonal and terminal path stepping beyond the player
   straight-axis baseline and positive-speed holdouts.
7. Expansion 1.14d holdout vectors for the implemented drain/recovery cadence,
   town exception, torso multiplier, and exhaustion tick.
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

- D2MOO revision `3b21043b99e987bad41cf0f7b49f1f246db52d5c`
  `source/D2Common/include/Path/Path.h`, `PathMisc.cpp`, `Units.cpp`,
  `D2Game/src/UNIT/SUnitDmg.cpp`, `SKILLS/SkillItem.cpp`,
  `MISSILES/MissMode.cpp`, `PLAYER/PlrModes.cpp`, and
  `MONSTER/MonsterMode.cpp` as secondary recovered 1.10f/1.13c evidence, not
  the 1.14d target authority.
- D2MOO missile creation/path configuration.
- D2MOO `source/D2Game/src/UNIT/SUnitInactive.cpp` for room/inactive monster restoration.
- Owned Expansion 1.14d `CharStats.txt`, `ItemStatCost.txt`, and `Properties.txt`
  rows for maximum-stamina units/dependencies and property links.
- D2MOO `D2StatList.cpp`, `PlayerStats.cpp`, and `ItemMode.cpp` for corroborating
  fixed-point dependency, level refill, and max-source current-resource callback
  structure; recovered code remains secondary evidence where a target runtime
  vector is still listed above.
- Current Dark Magic `internal/game/world/navigation.go`, `internal/game/session/movement.go`, generated map/collision/transition architecture and recent world diagnostics.
