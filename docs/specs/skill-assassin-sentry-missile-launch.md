# Assassin sentry missile launch

**Description**

When a charged-bolt sentry fires, its skill releases a burst of small bolts rather than one missile. The number of bolts it tries to release comes from the skill's own formula, and a bolt is skipped when no usable target point is found for it. Each released bolt gets a short, capped travel time and its own random seed. How the launched missiles then fly, collide and match their observed speed is described only as far as the evidence reaches.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: STT-R, STT-Q, STT-V, STT-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-assassin-sentry-missile-launch.md)
Open evidence gaps: [gap file](../plans/gaps/skill-assassin-sentry-missile-launch.md)

## 1. Scope

Defines the launch side of the existing Assassin trap-family boundary: the Skills text rows that select action 17, the action-17 burst (validation, attempt count, per-index target resolution, conditional creation requests and the initialization callback) and the shared missile-creation speed stage as sentry missiles consume it. The sentry's think, shot budget, target search and range belong to [Assassin sentry firing and range](skill-assassin-sentry-firing-and-range.md). Missile flight/collision, damage/hit resolution and the full skill/missile catalogs belong to their owners. Collision footprint, steering and observed speed are gated below; no fixed angular spread is inferred.

## 2. Trigger and ownership

D2Game computes the action-17 handler, its initialization callback and the creation speed stage; D2Common supplies skill/missile records, positions, frame counts, seed initialization and path requests. Skills.srvdofunc selects the burst through SCD-R06/SCD-R12: the handler runs once each time cast dispatch reaches the skill's action, not automatically once each frame. For a sentry, the think's shot request is SFR-R01f; how that mode request reaches action dispatch is Q04. The initialization callback runs inside the creator for each missile it actually initializes, and the speed stage runs inside each creation request the creator processes (Q03/Q04). No ordering against unrelated frame systems or named client message is established. [Binary chains](../research/spec-evidence/skill-assassin-sentry-missile-launch.md#dispatch-and-function-records) identify the reached handler and its output paths.

## 3. Inputs and data

Arithmetic is signed i32 with low32 wrapping and signed division truncating toward zero unless stated. Seeds consume NRF-R03a; units consume NRF-R10a/NRF-R10b. Supplied dependency results describe local request/return boundaries and do not verify those producers.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| caster, created missile | runtime units/creator results | optional objects | latch owner, creation owner, initialized unit |
| action skill identity and level | the action dispatcher's arguments (SCD-R06) | signed i32 each | R02 validation, `calc1` and creation request |
| skill record count and record | Skills collection | signed i32 count; optional record | R02 skill validation |
| selected missile and record count | the supplied missile-selection operation for the skill and caster; Missiles collection | signed identity; signed i32 count | R02 missile validation |
| calculation | Skills.`calc1` | compiled reference and i32 result | burst attempts; SCX-R11/SLD-R19 |
| burst resolution and creation | per-index target coordinates/result, caster position, creator result/callback | i32 coordinates; boolean result | conditional attempts and creation arguments |
| initialization inputs | the missile's total and current frame counts, its path stored-point X, the creation index | frame counts signed i16; stored-point X u16; index i32 | callback frame clamp, seed and path duration |
| speed values | Missiles.Vel/VelLev, invocation level, explicit velocity/conditions, slow mask, state 87 slow-state identity (GDF-R03), state list stat 161 | Vel/VelLev u8, others i32/booleans | shared speed stage; list-selected layer 0 read per RLI-R02a/RLI-R02b |

| Mapping / constant | Origin | Binding |
|---|---|---|
| action 17, path 10, frame threshold 78 and frame cap 77 | engine code | fixed original handler/constants; the burst count `calc1` has no local cap |
| state 87 test | engine code; GDF-R03 supplies the name | fixed original-code test; the name does not imply remapping |
| Skills.srvdofunc text cells for Id38/186/306/322 | table data; [TD-SKILLS17](../research/spec-evidence/skill-assassin-sentry-missile-launch.md#td-skills17) | text-resolved selectors choose engine-defined handlers; compiled binding Q02 |
| `calc1`, Vel/VelLev, Missiles.Size | table data | supplied prepared-record values; original row values require Q02 records |

## 4. Rules (execution order)

R01 names the text selectors, R02 is the reached action and its initialization callback, R03/R04 are the creation speed stage and R05 the gated footprint.

**STT-R01** Text rows selecting action 17 [verified in 1.13c data]. The effective Skills text rows with srvdofunc=17 are Charged Bolt (Id38), PrimeBolt (Id186), BoltSentry (Id306), ImpBolt (Id322). Each value is a named cell from the linked TD-SKILLS17 record. This establishes text selectors, not their compiled/runtime correspondence or every skill's execution.

**STT-R02** Charged-bolt requests and initialization [verified in 1.13c code (static)]. At reached action selector 17, set the caster do-handled latch when the caster is present, before skill validation. A skill identity <0 or >= the Skills record count, or an absent Skills record, returns 0; so does an invalid selected missile identity. These are the handler's returns; the dispatcher's final action result is SCD-R12's. The action receives the skill identity and skill level from its dispatcher (SCD-R06). The missile identity is the result of the supplied missile-selection operation for that skill and caster; it is invalid when <0 or >= the Missiles record count. With a valid missile identity, evaluate `calc1`(sentry, skill identity, the action's supplied level) to signed count n; nonpositive n makes no creation attempt and returns 1. For each index 0<=i<n, request target-coordinate resolution for the caster; the request carries the caster only (no index, skill or level), and each index receives that call's supplied result (resolver production is Q04). Only a successful resolution with BOTH resolved coordinates nonzero requests missile creation; coordinates returned with a failed resolution are not used. The creation request carries the missile identity, owner caster, source = caster position, target = that index's resolved coordinates, the skill identity and the action's level, the initialization callback and i; creator failure does not stop subsequent indices and does not change the final return 1. With an absent caster the position reads and resolution requests still use the supplied results for that absent caster (Q04). The callback runs only for a missile the creator actually initializes, and that invocation belongs to the creator (Q04); a failed creation initializes nothing. An initialization callback given an absent unit does nothing. On each actually initialized missile, in this order: read its own TOTAL frame count f as signed i16 (the current frame count is not read); if f>=78, set g=77 and write total then current frame count 77, else g=f with no frame write; initialize its seed from wrapping i32(its own path stored-point X, read as u16, +i) under NRF-R03a; request path type 10; request path duration min(u8(g), 77), using g after the clamp, not the original f; then request path preparation. Creation, path preparation and steering remain Q04/Q01; request count is not successful bolt count. No fixed angular spread is computed here.

**STT-R03** Shared missile speed [asm-only]. The creation speed stage is the missile creator's, consumed from [missile creation and seeding](missile-creation-and-seeding.md): MCI-R05 initial speed, then MCI-R06 slow-missiles scale, then MCI-R07 global speed scale. Local consequences for an action-17 burst: the owner those rules read is the creation request's owner unit, the caster. Per RLI-R02a a state-87 list without a stat 161 entry reads 0, giving speed 0, which then skips the MCI-R07 scale; whether RLI-R02b's minimum adjustment applies to this read is not established (Q03). Full output binding/creator applicability remains Q03/Q04; no observed world speed follows from this arithmetic alone.

**STT-R04** Observed sentry missile speed [unresolved]. Matching this creation speed stage to the travel of actual Lightning Sentry and Charged Bolt Sentry missiles remains Q03. No particular Vel table cell is asserted without its data record.

**STT-R05** Collision footprint [unresolved]. A missile's collision footprint, its relation to Missiles.Size and units, and flight steering need the missile-path/collision owner. The sentry's candidate-size range arithmetic does not establish missile collision geometry.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| STT-R01 | Text rows selecting action 17 | verified in 1.13c data | STT-Q02: compiled bindings/other required cells | TD-SKILLS17 |
| STT-R02 | Charged-bolt requests and initialization | verified in 1.13c code (static) | STT-Q04: creator/target/path effects; supplied requests available | STT-V01, STT-V02, STT-V03 |
| STT-R03 | Shared missile speed | asm-only | STT-Q03/STT-Q04: full creator/write integration | STT-V04 |
| STT-R04 | Observed sentry missile speed | unresolved | STT-Q03: observed movement | none |
| STT-R05 | Collision footprint | unresolved | STT-Q01: footprint integration | none |

## 5. Outputs and side effects

| Effect | Rules |
|---|---|
| burst latch, per-index target-resolution requests, conditional creation requests with their stated arguments, callback seed/frame/path-duration/type and path-preparation requests, action return | STT-R02 |
| supplied creator speed intermediate | gated STT-R03 |

Vectors also assert values that are not stored effects: intermediate results of the rules above (attempt counts, clamped frame values, speed intermediates) and the presence, absence, arguments and count of requests to supplied operations (target resolution, creation, evaluations). These are checked through the supplied operations' logs and return values. Dependency operations retain their declared stored effects and draws, including calculation draws and creation seeding. No persistence write or named client message is established here.

## 6. Branches and edge cases

An invalid skill or selected missile returns 0 after the latch is set. Count<=0 returns successful action without a creation request. Both coordinates must be nonzero for a burst creation request; negative coordinates are not rejected by that zero test, and coordinates returned with a failed resolution are ignored. Creation may fail on each attempt without stopping later indices or changing the return 1. A null initialization callback unit does nothing. The initialization uses the missile's own path stored-point X, not an assumed caster X, because the stored point is independent of current position. A frame count of 78 or more is clamped to 77; a negative frame count is not written and its u8 narrowing sets the requested duration. In the speed stage an explicit velocity of 0 is used as 0, a missing state-87 list skips the slow scale and a zero speed skips the 75% scale. No difficulty/PvP special arithmetic exists in these local stages.

## 7. Open questions

Attempted evidence and precise blocked boundaries are in the [gap file](../plans/gaps/skill-assassin-sentry-missile-launch.md).

| Id | Status / kind | Remaining evidence |
|---|---|---|
| **STT-Q01** | unresolved / integration | Missile footprint/collision units and actual charged-path steering; O01 |
| **STT-Q02** | unresolved / integration | Original relevant `calc1`/Vel/VelLev/Size values and the compiled/runtime association of the action-17 selectors; text do 17 cells alone do not establish these |
| **STT-Q03** | unresolved / observation and integration | Shared speed output binding and actual trap travel, with authenticated Vel inputs |
| **STT-Q04** | unresolved / integration | Missile selection, target resolution, creation/path initialization and the creator's speed write; how the sentry's skill-mode request reaches action dispatch |
| **STT-Q05** | unresolved / observation | No gameplay vector has been observed on the original game; observation is possible and pending |

## 8. Test vectors

### 8.1 Executable vectors

All records are synthetic supplied inputs, not archive rows. Each independent case starts with game frame 100; caster monster sentry S at (1000, 1000); action skill identity 20 with Skills count 400 and a present record, action level 3. Target-resolution and creator operations log requests/results; pure getters are omitted from asserted call-count logs. Logs assert effectful requests, expression draws and their meaningful order. The calculation operation returns the stated `calc1` without draws. A burst fixture selects synthetic missile identity 1 with count 2 and a present record; successful target resolution returns (1010, 1001), caster latch starts false, and each initialized missile enters with seed {6, 0x29A}. Successful creator effects in these fixtures invoke only the supplied callback and make no game-seed draw; real creation-seeding draws remain SCD-R13/NRF-R12a. Initializer stored-point X is an unsigned 16-bit field supplied independently of current position; frame count is signed i16. Each frame-count fixture supplies both total/current fields equal to its stated f before the callback; current position does not determine stored-point X. Supplied dependency operations persist their stated writes, log in order and draw nothing unless a case says otherwise. Unmentioned state stays at entry; logs start empty. Every vector states only its reached local or supplied-input assertions; integration gates remain in the Id map.

- **STT-V01** Burst attempt count (STT-R02; boundary: supplied-input stage)
  - Inputs: valid missile; `calc1` result 3; per-index resolve succeeds with(1010, 1001); creator initializes every request at stored-point X = 1000, f = 100.
  - Seeds: no direct draw; each initialized seed is overwritten from the stated value.
  - Draws: none.
  - Trace: indices 0, 1, 2; f = 100>=78 ->77; seed = stored-point X + i.
  - Outputs: 3 creation requests; return 1; latch=true; seed lows 1000, 1001, 1002, each hi0x29A; total/current 77, path duration 77, type 10.
  - Discriminates: Count means 3 conditional requests with successful creators explicitly supplied. This same-X case does not discriminate the initializer's X source; V03 supplies current X700 versus stored-point X40, producing seed 44 for index 4 rather than 704.
  - runtime_observed: no

- **STT-V02** Charged initializer cap (STT-R02; boundary: supplied-input stage)
  - Inputs: `calc1`=3; three successful resolution/creation results; each created missile stored-point X = 1000, frame count 100.
  - Seeds: no direct draw; initialize each seed from values.
  - Draws: none.
  - Trace: seed=1000+i; 100>=78 clamps.
  - Outputs: 3 initialized missiles:seed lows 1000/1001/1002, hi0x29A; current/total frames 77; path duration 77, type 10.
  - Discriminates: This case pins the stated seed values and cap 77, without independently distinguishing equal caster/stored-point X inputs. V03 discriminates stored-point X40 from current X700: seed 42 for index 2 rather than 702, and seed 44 for index 4 rather than 704.
  - runtime_observed: no

- **STT-V03** Per-index creation gates/failure (STT-R02; boundary: supplied-input stage)
  - Inputs: action skill 20 at supplied level 3, so `calc1` is evaluated with (S, 20, 3) and returns 5; index 0 resolution fails while returning coordinates (5, 5); index 1 success(0, 10); index 2 success(10, 0); index 3 success(-2, 9), creator fails; index 4 success(12, 8), creator succeeds with missile M, current X=700, stored-point X=40, total/current frames 77 and seed{6, 0x29A}, invoking its initializer with index 4; independent initialized missile current X = 700, stored-point X = 40, index 2, frames 77; null initializer unit.
  - Seeds: no direct draw; initializer seed replacement only.
  - Draws: none.
  - Trace: indices 3/4 have success and both coordinates nonzero; failure at 3 does not stop index 4. M initialization uses 40+4=44; the independent callback uses 40+2=42; 77<78 retains both frame fields.
  - Outputs: five target-resolution requests, each for caster S only; ordered creation requests for indices 3 then 4, returning failure then M. Request 3 carries missile identity 1, owner S, source (1000, 1000), target (-2, 9), skill 20, level 3, the initialization callback and index 3; request 4 is the same with target (12, 8) and index 4; final action return 1, latch true; M seed={44, 0x29A}, frames 77 with no frame-count write request, duration 77, type 10. The independent initializer yields seed{42, 0x29A}, frames 77, duration 77, type 10; null initializer unit has no effects. Within M's successful initializer, the ordered effect subsequence is seed overwrite{44, 0x29A}, path-kind request 10, path-duration request 77, path-preparation request for M. The independent frame 77/index 2 initializer has the same ordered request sequence with seed{42, 0x29A}; these are request assertions, not assertions about path-preparation internals.
  - (a) Inputs: independent initialized missile, current X=700, stored-point X=40, index=2, signed total/current frame counts=-1. Draws: none. Outputs: seed={42, 0x29A}; total/current frame count=-1 unchanged; requested path duration=77 because u8(-1)=255, capped 77.
  - (b) Inputs: independent initialized missile, current X=700, stored-point X=40, index=2, signed total/current frame counts=-256. Draws: none. Outputs: seed={42, 0x29A}; total/current frame count=-256 unchanged; requested path duration=0 because u8(-256)=0.
  - (c) Inputs: independent action with non-null caster, latch=false, skill identity 400/count 400. Draws: none. Outputs: latch=true before rejection; return 0; no expression or creation request.
  - (d) Inputs: independent valid action, latch=false, `calc1` result 0. Draws: none. Outputs: latch=true, return 1; no target-resolution or creation request.
  - (e) Inputs: independent action with non-null caster, latch=false, selected missile identity 2, count 2; valid skill 20. Draws: none. Outputs: latch=true, return 0; no expression evaluation or creation request.
  - (f) Inputs: independent action with non-null caster, latch=false, selected missile identity=-1, count 2; valid skill 20. Draws: none. Outputs: latch=true, return 0; no expression evaluation or creation request.
  - (g) Inputs: independently reached action with caster absent, valid skill 20, selected missile identity 1/count 2, `calc1` result 0; supplied caster-position reads return(0, 0). Draws: none. Outputs: action return 1; no unit latch exists to write; no creation request. This is the handler stage, without asserting that outer dispatch accepts an absent unit.
  - (h) Inputs: independent initialized missile, stored-point X=40, index=2, signed total/current frame counts=78. Draws: none. Outputs: ordered effects: total frame count write 77, current frame count write 77, seed overwrite{42, 0x29A}, path-kind request 10, path-duration request 77, path-preparation request. A `>78` clamp test would leave the frame counts at 78.
  - (i) Inputs: independent initialized missile, stored-point X=40, index=2, signed total/current frame counts=300. Draws: none. Outputs: total/current frame counts 77; seed={42, 0x29A}; requested path duration=77 (the clamped 77, not u8(300)=44).
  - (j) Inputs: independent valid action, latch=false, `calc1` result -1. Draws: none. Outputs: latch=true, return 1; no target-resolution or creation request. An unsigned or nonzero loop test would make requests.
  - (k) Inputs: independent initialized missile, stored-point X=40, index=2, total frame count 50, current frame count 100. Draws: none. Outputs: no frame-count write; total 50 and current 100 unchanged; seed={42, 0x29A}; path duration 50. Reading the current count (100) would write 77 to both.
  - (l) Inputs: independent valid action, `calc1` result 1; index 0 resolution succeeds with (10, 10); the creator fails. Draws: none. Outputs: one creation request; return 1. Returning the last creator's result would give 0.
  - Discriminates: Stopping on the first creator failure omits the index 4 request and M seed 44. The asserted creation arguments reject passing the caster position as target or swapping source and target. M's frames 77 produce no frame-count write request, rejecting a `>=77` clamp. Ignoring the failed index 0 resolution would request a creation for (5, 5). Count is not successful missiles; negative coordinates still pass the nonzero gate. (h) pins the clamp boundary 78 and frame-write order; (i) rejects a duration computed from the unclamped f.
  - runtime_observed: no

- **STT-V04** Shared speed supplied arithmetic (STT-R03 consuming MCI-R05/MCI-R06/MCI-R07; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a)30*256*75/100=5760; (b)10*256*75/100=1920; (c)missing list skips 50%; (d)2560*50/100=1280 then 75%=960.
  - (a) Inputs: Vel=30, VelLev=0, L=3, no explicit/slow. Draws: none. Outputs: speed intermediate 5760.
  - (b) Inputs: explicit 10, unscaled, no slow. Draws: none. Outputs: speed intermediate 1920.
  - (c) Inputs: explicit 2560, already scaled, slowmask true, state 87 active, list absent. Draws: none. Outputs: speed intermediate 1920.
  - (d) Inputs: explicit 2560, already scaled, slowmask true, state 87 active, list present, selected stat 161/layer 0=50. Draws: none. Outputs: speed intermediate 960.
  - (e) Inputs: explicit velocity 1073741824 (0x40000000), already scaled, no slow. Draws: none. Outputs: wrapped product low32(1073741824*75)=0xC0000000, interpreted as -1073741824; speed intermediate=-10737418 after signed truncation by 100.
  - (f) Inputs: Vel=0, VelLev=1, L=-9, no explicit/slow. Draws: none. Outputs: trunc(-9/8)=-1; shifted speed=-256; speed intermediate=-192.
  - (g) Inputs: explicit velocity 2560, already scaled, slowmask true, state 87 inactive, list present, selected stat 161/layer 0=50. Draws: none. Outputs: speed intermediate 1920; inactive state skips the slow scale despite the matching mask and available list.
  - (h) Inputs: explicit velocity 2560, already scaled, slowmask false, state 87 active, list present, selected stat 161/layer 0=50. Draws: none. Outputs: speed intermediate 1920; false mask match skips the slow scale despite the active state and available list.
  - (i) Inputs: explicit velocity 3, already scaled, slowmask true, state 87 active, state-87 list present with stat 161/layer 0=50. Draws: none. Outputs: slow trunc(3*50/100)=1, then nonzero 1 becomes trunc(1*75/100)=0; speed intermediate 0. Applying 75% first gives 2 then 1.
  - (j) Inputs: explicit velocity 2560, already scaled, slowmask true, state 87 active; the owner's state-87 list has stat 161/layer 0=50 and its base list has stat 161=100. Draws: none. Outputs: speed intermediate 960. Reading the base list would give 1920.
  - (k) Inputs: explicit-velocity condition set with value 0, unscaled, Vel=30, VelLev=0, L=3, no slow. Draws: none. Outputs: 0 shifted is 0; the zero speed skips the 75% scale; speed intermediate 0. Treating value 0 as no explicit velocity would give 5760.
  - Discriminates: State alone must not imply zero from missing list or skip list-selected stat; (k) tests the explicit condition rather than the value; (i) fixes slow-then-75% order and (j) the state list as the stat source.
  - runtime_observed: no

### 8.2 Observation vectors

All are pending; this initial recipe does not exhaust every rule. Procedures are [in the companion](../research/spec-evidence/skill-assassin-sentry-missile-launch.md#observation-procedures).

- **STT-O01** Measure charged-bolt burst count and successive path directions against a stationary target; compare only successful, initialized requests with V01/V02 and retain creator/steering limits.

## 9. Acceptance and dependencies

Partial acceptance reproduces all verified local request/return, initialization and seed assertions in V01, V02 and V03, respecting each vector's supplied boundary and mixed-rule gates. V04 is an explicit supplied arithmetic check for the asm-only R03; passing it does not promote R03. TD-SKILLS17 establishes only the named text cells in R01. Full reproduction additionally requires Q01/Q02/Q03/Q04 producer contracts and original-game observations for every behavioral rule (Q05). O01 and further branch observations are pending; no partial pass claims complete parity.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md), NRF-R03a supplied seed initialization, NRF-R10a/b coordinates, RLI-R02a/RLI-R02b list stat reads and NRF-R12a creation seeding; [data fields](core-game-data-fields.md), GDF-R03 states; [missile creation and seeding](missile-creation-and-seeding.md), MCI-R05/MCI-R06/MCI-R07 speed stage; [calculation expressions](skill-calc-expressions.md), SCX-R11/SCX-R15/SCX-R24 and their producer gates; [leveled skill calculations](skill-leveled-damage-tohit-and-elements.md), SLD-R19 calculation-front-end boundary; [cast dispatch](skill-cast-start-and-do-dispatch.md), SCD-R06/SCD-R12 action selection and SCD-R07/SCD-R13 creation request/seed boundaries. Owners of the cited RLI ids: [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** no known consumer cites these rules.

## 10. Related specs

- [Assassin sentry firing and range](skill-assassin-sentry-firing-and-range.md): the think whose shot request (SFR-R01f) leads to this launch; its target search and range do not depend on these rules.
- [Missile creation and seeding](missile-creation-and-seeding.md): generic creation, seeding and initial speed.
- [Missile movement and collision](missile-movement-and-collision.md): later flight and collision of the launched missiles.
- [Blade Sentinel and Fury](skill-assassin-blade-sentinel-and-fury.md): a separate Assassin summon with its own launch context.

## Appendix A. Provenance

Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies D2Game/D2Common original images. The reused speed-stage reading retains its 2026-10-12 date and limits; the action-17 handler, initializer and speed-branch readings and the narrow effective Skills text cells are dated 2026-10-15. The [companion](../research/spec-evidence/skill-assassin-sentry-missile-launch.md) records exact functions, corrections and table cells. No gameplay observation was performed; table reads establish named text values only.
