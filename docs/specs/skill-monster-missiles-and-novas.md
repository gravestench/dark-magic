# Monster missiles and novas

**Description**

Monster projectile skills can fire toward a target, spread several shots, or surround the caster with a nova. The nova handler issues its own ordered projectile requests and can add a second set between the first directions. Creating each projectile and resolving its later hit remain separate behaviors.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-record-layouts-and-ids](core-record-layouts-and-ids.md); names: [core-game-data-table-records](core-game-data-table-records.md)
Ids: MMS-R, MMS-Q, MMS-V, MMS-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-missiles-and-novas.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-missiles-and-novas.md)

## 1. Scope

This contract owns monster skill action callbacks for targeted spray, poison nova, sustained breath, turret fire and monster missile selection: action selectors 88, 99, 95, 105, 106, 110, 148 and 149. Closely related projectile callbacks 85 and 93 are identified as gated family paths. It specifies nova requests, their position/expression/creator handoff, and the tower callback with its 64-direction launch builder. Sustained-skill state, spread-turret arithmetic and missile-variant selection remain gated. AI selection belongs to its caller. Missile creation, initialization, movement and hits consume their own contracts.

## 2. Trigger and ownership

D2Game owns these `SKILLS` callbacks. [SCD-R06](skill-cast-start-and-do-dispatch.md) chooses the callback from the prepared `srvdofunc` selector and invokes it synchronously. Sustained breath additionally uses a start selector; its complete action/event cadence remains MMS-Q04. No unconditional client message or per-frame frequency is established by this contract. The shared missile contracts own creation and client synchronization requests.

## 3. Inputs and data

All calculations below use signed i32 and wrap each arithmetic operation to 32 bits unless stated otherwise. Signed comparisons and division use [numeric foundations](core-numeric-and-rng-foundations.md).

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| game, caster, identity, invocation level | action callback | references; i32 identity/level | Context passed to selected callback |
| loaded skill/missile collections | GDA-R11a / GDA-R03 | signed counts; records or absent | Lookup and bound inputs |
| missile choice | prepared `Skills.srvmissilea` | signed 16-bit linked identity | First missile for the relevant skill |
| nova loops and spacing | `Skills.calc1`, `Skills.calc2` evaluated with caster, skill identity, level | compiled references; i32 results `loops`, `spacing` | Nova request loop field and second-ring spacing |
| primaryParameter, secondaryParameter | supplied prepared missile `Param1`, `Param2` | i32 | Primary and secondary speed parameters multiplied by 64 before request creation |
| source position | supplied caster X/Y; NRF-R10b for dynamic path reads | i32 path units | Explicit start of each nova request |
| tower speed inputs | selected Missiles `Vel`, `VelLev` | unsigned bytes; i32 level | Whole-unit getter, distinct from nova fixed-point speed |
| target cache/registry | game registry and cached reference/saved key | unit or absent; kind/id | MTP-R09 target refresh |
| compiled expression pool | SCX skill pool and calculation references | bytes and unsigned offsets | Explicit bytecode evaluation; text compilation separate |
| direction constants | section 4's ordered pairs | i32 components | Original code-defined relative aims |
| creation result | MCI/MIR operation | missile or absent | Result deliberately ignored by nova/tower caller |
| targeting, current skill, monster variant values | target refresh, skill-node and monster data operations | logical references, booleans and small integers | Gated targeted, breath and variant paths |
| skill-action condition | caster condition set by this callback | boolean | Neutral local condition; no additional networking or eligibility meaning is asserted |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Named `srvdofunc` cells in MMS-R01 | effective `data/global/excel/skills.txt`, named row and column; T01 in companion | Data selects fixed engine handlers; compiled row correspondence remains MMS-Q01 |
| direction pairs, eight primary calls, scaling factor 64, minimum spacing 1 | original engine instructions and constant arrays | Fixed by original code |
| speed parameters and calculation results | loaded Missiles/Skills fields | Data supplies values; handler fixes their uses |

Synthetic vectors supply prepared values; they do not assert an original missile row's numeric speed parameters.

## 4. Rules (execution order)

**MMS-R01** Named projectile selectors [verified in 1.13c data]. T01 establishes `AndrialSpray.srvdofunc = 88`, `PrimePoisonNova.srvdofunc = 99`, `FetishInferno.srvdofunc = 95`, `DesertTurret.srvdofunc = 105`, `ArcaneTower.srvdofunc = 106`, `MissileSkill1.srvdofunc = 110`, `DoomKnightMissile.srvdofunc = 148` and `NecromageMissile.srvdofunc = 149`. These are named text cells, not gameplay executions.

**MMS-R02** Nova early refusal [verified in 1.13c code (static)]. Selector 99 returns 0 before coordinates, calculations or creation when signed skill identity is negative or at least skill count, the prepared skill record is absent, signed `srvmissilea` is negative or at least missile count, or the prepared missile record is absent. No random draw or unit write occurs on those refusal paths. Complete valid prepared storage supplies records at valid indices; an absent in-range resolution is a synthetic malformed-storage case, not an asserted ordinary archive condition.

**MMS-R03** Nova request stage [verified in 1.13c code (static)]. This verified successful-continuation stage requires a present caster and valid prepared records. Set the caster's skill-action condition. Obtain source X then Y. Evaluate `calc1` then `calc2`, both with the invocation caster, skill identity and level; retain `calc1` as `loops` and set `m=max(spacing,1)` by signed comparison. Dynamic monster coordinates execute NRF-R10b: the integer unsigned high words of 16.16 position; an absent dynamic path gives 0 per coordinate. Calculation references execute SCX-R11 in the skill pool with that context; supplied terminated bytecode executes SCX-R01 through SCX-R10 and the selected parameter/function rules. Compiler production remains SCX-Q01. The stage can also run with supplied coordinate/calculation results, without claiming those producers. A valid-record invocation with absent caster is outside this verified stage and remains MMS-Q09.

Initialize a zeroed creation request. Set owner and source to the caster, missile identity to `srvmissilea`, explicit source point to the supplied position, invocation level to the supplied level, loop count to `loops`, and speed to `i32(primaryParameter*64)`. Set explicit-start, start-relative-aim, speed-override and already-fixed-point-speed controls, and enable adding animation sub-loop duration to range under MIR-R03. The request skill identity remains 0. Other fields retain the zeroed defaults.

Emit eight requests in the following order, changing relative aim for each: `(0,2), (2,2), (2,0), (2,-2), (0,-2), (-2,-2), (-2,0), (-2,2)`. Each emission invokes the whole creation operation once; its result is ignored and never terminates the loop.

Then replace speed with `i32(secondaryParameter*64)`. If `m=1`, emit no secondary request. For `2<=m<=536870911`, use this ordered list of 16 direction pairs, indexed from zero:

`[(0,2),(1,2),(2,2),(2,1),(2,0),(2,-1),(2,-2),(1,-2),(0,-2),(-1,-2),(-2,-2),(-2,-1),(-2,0),(-2,1),(-2,2),(-1,2)]`.

Emit the pairs at indices `1, 1+m, 1+2*m, ...` while the index is at most 15. Each request retains all other fields and ignores creation results. Return exactly 1 after the request sequence, even if every creator invocation returns absent. The stage draws no random numbers directly; calculations and each creation invocation account for their own draws separately. Values of `m` above the stated nonoverflowing bound are MMS-Q02, not a clamped default.

**MMS-R04** Spray launch [unresolved]. Selector 88 requires a valid skill/missile and a current skill node. It obtains two skill-node values and uses target-coordinate resolution if either is zero, then derives a directional starting point and aim correction before one creation request. It returns 1 after the reached creator call without testing that result. Exact directional-table interpretation, target fallback, node fields and dependent draws remain MMS-Q03.

**MMS-R05** Sustained breath [unresolved]. Selector 95 requires a valid skill, current skill and missile. It computes speed from missile velocity fields and invocation level, resolves an aim point, and calls the creator. Missing aim or absent creation reaches a cleanup request and returns 0. Successful creation reaches local follow-up operations; continued breathing can inspect a state and scheduled-frame condition before requesting an event, while another success route requests cleanup and still returns 1. Complete start/action handoff, skill-node bookkeeping, state identity, event schedule and cleanup are MMS-Q04.

**MMS-R06a** Spread turret [unresolved]. Selector 105 resolves target coordinates, rejects coincident aim, derives a signed spread direction and `calc1` count, and emits that many requests when positive; count zero or negative reaches a return of 1 after its earlier gates. Full spread arithmetic and dependent operations remain MMS-Q05.

**MMS-R06b** Tower callback [verified in 1.13c code (static)]. For a present monster with a dynamic path and supplied game registry, validate signed skill identity against the loaded count and resolve its record; reject absent resolution. Read signed `srvmissilea`; negative or at least missile count returns 0. Set the skill-action condition before target refresh, even if a later target test fails. Refresh the caster cache under MTP-R09, then obtain it; absent or self target returns 0 without a launch. A distinct retained target supplies eligibility only; its coordinates are not used by this callback/builder.

Obtain launch speed from the selected missile: invalid identity or absent record produces 0; otherwise use unsigned byte `Vel` and `VelLev` as `speed=i32(Vel+truncTowardZero(i32(VelLev*level)/8))`. Invoke MMS-R10 with game, caster as position unit and owner, missile identity, invocation skill identity/level and that speed. Return exactly 1 after normal builder completion regardless of any creation return. No direct draw; creation owns its draws. This callback does not itself require a nonabsent missile record after the identity bound; the speed helper/creator handle that resolution.

**MMS-R07** Generic monster shot [unresolved]. Selector 110 refreshes a distinct target and obtains a monster-derived missile selector. Values 0 and 31 use the skill's `srvmissilea`; other returned values use the monster-selected missile after lookup. The skill fallback selects one of two launch builders using a skill condition, and the reached builder's result does not decide callback result 1. Complete selector production, condition meaning, builders and final missile fields remain MMS-Q06.

**MMS-R08** Monster variants [unresolved]. Selectors 148 and 149 refresh a distinct target and validate base `srvmissilea`. A monster supplies an unsigned byte variant added to the base; other unit kinds use 0. The adjusted identity must still be below missile count and have a record. The two selectors use different monster bytes; the reached launch builder is selected by a skill condition and its result is ignored before returning 1. The variant bytes' production and meaning, and builder behavior, remain MMS-Q06.

**MMS-R09** Related simple/trap projectile paths [unresolved]. The original action table also points selectors 85 and 93 into the same monster skill family. Named cells associate them with UnHolyBolt/ShamanFire and GargoyleTrap respectively. Their complete target, missile request and return rules remain MMS-Q07; the names alone supply no executable behavior.

**MMS-R10** Tower ring builder [verified in 1.13c code (static)]. With a present position unit, read its X then Y once. Dynamic kinds use NRF-R10b or 0 for an absent path; kinds object/item/tile use prepared static coordinates. Initialize a zeroed whole creation request. Enable explicit start and start-relative aim; set owner, missile identity, source point, skill identity and level from arguments. Source unit remains absent, as do target unit and callback. Loops and every other unmentioned semantic field stay 0/absent. For a nonzero supplied speed, enable speed override and store that speed in whole units; leave the already-fixed-point control clear. A supplied zero leaves speed override clear, so MCI-R05 obtains speed from the missile row instead.

For each direction below issue one whole MCI/MIR creation operation in order, ignoring every return. Let the first quadrant's ordered pairs be

`[(30,0),(29,2),(29,5),(28,8),(27,11),(26,14),(24,16),(23,19),(21,21),(19,23),(16,24),(14,26),(11,27),(8,28),(5,29),(2,29)]`.

Emit those 16 pairs, then their respective transforms `(-y,x)`, then `(-x,-y)`, then `(y,-x)`, for exactly 64 requests. Each pair is a start-relative aim, not a source offset or the selected target's position. No local draw occurs; underlying MCI/MIR controls actual allocation, seeding, range, path, damage and return. The callback's return 1 does not imply a missile exists.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MMS-R01 | Named selector cells | verified in 1.13c data | none for cells; MMS-Q01 for compiled correspondence | T01 in companion |
| MMS-R02 | Nova early refusal | verified in 1.13c code (static) | none for supplied records | MMS-V01 |
| MMS-R03 | Ordered nova requests | verified in 1.13c code (static) | MMS-Q01: original row binding; MMS-Q02: extreme spacing; MMS-Q09: absent caster; MCI/MIR/SCX gates apply only to dependency-produced effects | MMS-V02, MMS-V03, MMS-V04, MMS-V05 |
| MMS-R04 | Spray | unresolved | MMS-Q03: full targeted spray | none |
| MMS-R05 | Sustained breath | unresolved | MMS-Q04: complete start/action/events | none |
| MMS-R06a | Spread turret | unresolved | MMS-Q05: exact spread integration | none |
| MMS-R06b | Tower eligibility and launch | verified in 1.13c code (static) | supplied registry/records; creation effects consume MCI/MIR | MMS-V06, MMS-V07 |
| MMS-R07 | Generic shot | unresolved | MMS-Q06: full selector/builder integration | none |
| MMS-R08 | Variant shots | unresolved | MMS-Q06: variant producer/builders | none |
| MMS-R09 | Simple/trap projectiles | unresolved | MMS-Q07: complete callbacks | none |
| MMS-R10 | Ordered tower ring requests | verified in 1.13c code (static) | complete creation effects retain MCI/MIR limits | MMS-V06, MMS-V07 |

## 5. Outputs and side effects

| Output | Rule | Boundary |
|---|---|---|
| result 0 without calls/writes | MMS-R02 | Verified early refusal |
| skill-action condition true; ordered calculation and creation requests; callback result 1 | MMS-R03 | Verified local request stage with supplied results; dependency-created units are separate |
| targeted creation request/result | MMS-R04 | Gated spray |
| breath creation, cleanup and event requests | MMS-R05 | Gated sustained path |
| spread launch requests | MMS-R06a | Gated spread |
| tower target gate, 64 ordered creation requests and result 1 | MMS-R06b, MMS-R10 | Verified local launch; creation dependencies explicit |
| selected/variant launch requests | MMS-R07, MMS-R08 | Gated shot paths |

Created units, actual game-seed advancement, registration and client messages follow MCI/MIR; a callback result of 1 alone proves none of them.

## 6. Branches and edge cases

Nova identity bounds are signed and exclusive. Its second ring is omitted when `calc2` is at most 1, including negative values. Primary requests always precede secondary requests. Calc1 is forwarded as a loop count without this caller clamping it. All-absent creation results still yield callback result 1. Null caster is not assigned a valid coordinate fallback. Tower target eligibility is tested after its condition write, so a stale target refusal leaves the condition set. Tower zero speed clears the override control and delegates row speed production; it is distinct from the nova fixed-point override. No independent difficulty or player-versus-player branch is established by the local nova/tower arithmetic; table/expression/creator inputs can differ. Extreme spacing and full malformed-input execution remain gated.

## 7. Open questions

**MMS-Q01** [unresolved] Capture compiled Skills/Missiles values and link them to T01; text identities alone do not establish live numeric missile or speed values.

**MMS-Q02** [unresolved] What is the executable outcome when positive `calc2` exceeds 536870911 and the internal signed loop step wraps? Resolve the resulting address/loop safety before specifying an output; do not invent a cap.

**MMS-Q03** [unresolved] Close spray's current-skill values, coordinate fallback and direction-table/callee semantics named in the companion.

**MMS-Q04** [unresolved] Close breath start selector 53, current-skill fields, cleanup, state and event operations through all writes and draw owners.

**MMS-Q05** [unresolved] Close selector 105 sign/quadrant/spread arithmetic and its coordinate/count integration; selector 106 and the ring builder are verified separately.

**MMS-Q06** [unresolved] Close monster missile selection, byte-variant production and both launch builders; specify their exact requests and dependency effects.

**MMS-Q07** [unresolved] Complete callbacks 85 and 93 rather than inferring behavior from names.

**MMS-Q08** [unresolved] Matching gameplay captures remain pending; no callback invocation or gameplay effect was captured. See the companion for the inconclusive observation limits and reproduction procedure.

**MMS-Q09** [unresolved] With valid prepared records and absent caster, what happens when the handler reaches its source-coordinate operation? The local condition write is skipped, but no coordinate fallback or callback result is established. MCI's absent-owner refusal does not answer this earlier coordinate boundary. Resolve the coordinate wrapper's absent-input behavior before extending MMS-R03 to that input.

## 8. Test vectors

### 8.1 Executable vectors

Each fixture is independent. In MMS-V01 to MMS-V04 only, the supplied creator records each complete semantic creation request in order and returns absent without mutating game/caster/seed. MMS-V05 to MMS-V07 execute their explicitly named dependency stages and stop at the stated boundary. Every assertion about a request compares the whole request defined by MCI/MIR: explicitly listed values and controls must match; every unmentioned scalar is 0, every unmentioned unit/reference/callback is absent, and every other logical request condition is clear. In particular target unit is absent, elapsed/activation override/to-hit/range override are 0 and optional callback/opaque argument are absent. This is semantic-field equality, not a comparison of binary padding. In MMS-V02 to MMS-V04, supplied X/Y and `calc1`/`calc2` operations log their order and return the stated values without draws. Thus request logging does not pretend to execute MCI/MIR or SCX integration. The request-control defaults and zeroed fields in MMS-R03 apply to each nova call; tower calls use MMS-R10 instead.

- **MMS-V01** Ordered nova refusal guards (MMS-R02; boundary: supplied-input stage)
  - Inputs: selector 99; game and monster caster present; initial skill-action condition false; identity 192; skill count 400; present skill with srvmissilea 10; missile count 700; present missile record; complete storage. Every case independently overrides these defaults. Record 0 and missile 0 are present with valid fields unless a malformed-storage case explicitly overrides them.
  - Seeds: no draw.
  - Draws: none.
  - Trace: test identity bounds, then resolve the skill record, then read/test its signed missile selector, then resolve the missile record; the first failed guard returns before source-coordinate lookup. Record resolution or its absence is distinct from reading a field of the resolved record.
  - Outputs: result 0; no source-coordinate, calculation or creation calls; caster condition unchanged; no creation requests. The per-case observations below distinguish an attempted record resolution from a record-field read.
  - Discriminates: clamping a negative identity or missile to the valid record 0 would reach source/calculation calls; accepting equality at a count or skipping a null-resolution guard would also continue. Empty downstream logs reject those behaviors. Equivalent unsigned bound formulations need not be distinguished.
  - runtime_observed: no.
  - (a) Inputs: identity -1. Outputs: result 0; no prepared-record resolution or record-field read; no source-coordinate, calculation or creation calls; condition false.
  - (b) Inputs: identity 400. Outputs: result 0; no prepared-record resolution or record-field read; no source-coordinate, calculation or creation calls; condition false.
  - (c) Inputs: identity 0; synthetic malformed skill storage resolves record 0 as absent despite count 400. Outputs: result 0 after that absent resolution; no skill-record field read, including its missile selector; no missile-record resolution; no source-coordinate, calculation or creation calls; condition false.
  - (d) Inputs: identity 192; srvmissilea -1. Outputs: result 0 after the skill missile-selector read; no missile-record resolution or missile-record field read; no source-coordinate, calculation or creation calls; condition false.
  - (e) Inputs: identity 192; srvmissilea 700. Outputs: result 0 after the skill missile-selector read; no missile-record resolution or missile-record field read; no source-coordinate, calculation or creation calls; condition false.
  - (f) Inputs: identity 192; srvmissilea 0; synthetic malformed missile storage resolves record 0 as absent despite count 700. Outputs: result 0 after that absent resolution; no missile-record field read; no source-coordinate, calculation or creation calls; condition false.

- **MMS-V02** Nova keeps both rings after absent creation (MMS-R03; boundary: supplied-input stage)
  - Inputs: selector 99; valid skill/missile records; identity 192; level 3; monster caster C at supplied position (100,200); initial skill-action condition false; srvmissilea 10; supplied `calc1` 2 and `calc2` 2; missile `Param1` 3 and `Param2` 5; all creator results absent.
  - Seeds: no draw in supplied operations or local stage.
  - Draws: none.
  - Trace: X then Y then `calc1` then `calc2`; m=2; primary speed 3*64=192; eight primary requests; secondary speed 5*64=320; indices 1,3,5,7,9,11,13,15 yield eight secondary requests.
  - Outputs: result 1; caster condition true; exactly 16 ordered creator calls. Every call has owner/source C, missile 10, source (100,200), level 3, loops 2, skill identity 0 and R03 controls. Calls 1-8 have speed 192 and relative aims [(0,2),(2,2),(2,0),(2,-2),(0,-2),(-2,-2),(-2,0),(-2,2)]. Calls 9-16 have speed 320 and relative aims [(1,2),(2,1),(2,-1),(1,-2),(-1,-2),(-2,-1),(-2,1),(-1,2)]. No created unit or seed change is asserted.
  - Discriminates: swapping direction axes changes call 1; stopping on an absent result reduces call count; reusing primary speed makes call 9 wrong; using the invocation identity in the zeroed request makes skill identity wrong.
  - runtime_observed: no.

- **MMS-V03** Negative spacing suppresses second ring (MMS-R03; boundary: supplied-input stage)
  - Inputs: selector 99; valid records; identity 192; level 1; monster C at (100,200); initial skill-action condition false; srvmissilea 10; `calc1` -3; `calc2` -8; `Param1` -1; `Param2` 9; creator always absent.
  - Seeds: no draw in supplied operations or local stage.
  - Draws: none.
  - Trace: m=max(-8,1)=1; primary speed i32(-1*64)=-64; primary eight calls still run; replacing speed with 576 emits no secondary call.
  - Outputs: result 1; caster condition true; exactly eight calls in R03 primary order, each with source (100,200), owner/source C, missile 10, level 1, loops -3, skill identity 0, speed -64 and R03 controls; no further calls.
  - Discriminates: clamping loops or speed to positive changes recorded fields; drawing a secondary ring for m=1 changes count.
  - runtime_observed: no.

- **MMS-V04** Secondary stride three and zero loop forwarding (MMS-R03; boundary: supplied-input stage)
  - Inputs: selector 99; valid prepared records; present monster caster C; identity 192; level 1; source (100,200); initial skill-action condition false; srvmissilea 10; supplied `calc1` 0 and `calc2` 3; `Param1` 1 and `Param2` 2; all creator returns absent.
  - Seeds: no draw in supplied operations or local stage.
  - Draws: none.
  - Trace: m=3; primary speed 64 and eight primary requests; secondary speed 128; indices 1,4,7,10,13 are at most 15, next index 16 stops.
  - Outputs: result 1; caster condition true; X,Y,`calc1`,`calc2` call order; exactly 13 creator calls. Each complete semantic request has owner/source C, missile 10, source (100,200), level 1, loops 0, skill identity 0 and R03 controls/defaults. Calls 1-8 have speed 64 and the primary aim order of R03. Calls 9-13 have speed 128 and aims [(1,2),(2,0),(1,-2),(-2,-2),(-2,1)] respectively; all unmentioned semantic request fields match the section preamble.
  - Discriminates: hard-coding odd secondary indices emits eight instead of five secondary requests; starting index 0 changes call 9 to (0,2); clamping loop count to 1 changes every request's loops field.
  - runtime_observed: no.

- **MMS-V05** Nova position, bytecode and first creation prefix (MMS-R03; SCX-R01/R03/R04/R06/R11/R21, NRF-R10b, MCI-R01 to MCI-R11, MIR-R01a; boundary: stop after first metadata writes)
  - Inputs: selector 99; valid skill 192/count 400/level 3, signed missile 10/count 700/present row; monster C/id 9, dynamic FX 6602752/FY 13123584, room R rectangle (90,190,40,40), level 80, no adjacent rooms. Row `Param1` 3/`Param2` 5, `CanSlow` false, `CollideType` 3. Skill pool [4,16,7,1,17,0,7,2,0], count 9; `calc1` reference 0/`calc2` reference 6. Expansion true, counter 41; successful zeroed first-unit U and metadata allocations supplied; act environment E2. Initial caster condition false.
  - Seeds: game {1,666}, caster {123,456}, child initially {0,0}.
  - Draws: one game step to {1791398751,0}; no caster/child draw by this boundary.
  - Trace: integer coordinates (100,200). First expression pushes parameter 16 => level 3, literal 1, subtract => 2; second pushes literal 2. First nova request has owner/source C, missile 10, explicit source (100,200), relative aim (0,2), level 3, loops 2, skill 0, fixed-point override 192 and R03 controls/defaults. Actual MCI selects R, retains (100,202), scales speed 192 to 144, passes distance (0,2), allocates/seeds/numbers U. Stop after MIR-R01a, before event/path/registry requests.
  - Outputs: caster condition true, seed {123,456}; loops 2/spacing 2; game seed {1791398751,0}, counter 42; U type missile/class 10/id 42/mode 3, game reference, act 2/E2, expansion/A/B true, selector 11, auxiliary absent; U seed {1791398751,666}, initial copy 1791398751; metadata unit id 42/source id -1/C/D false. Handoff start R/(100,200), retained (100,202), target unit absent, speed 144. No final creator/callback return or subsequent nova calls asserted.
  - Discriminates: raw fractional coordinates change room/start; wrong expression context or reversed subtraction changes loops; whole-unit treatment of 192 changes speed; stepping child changes its seed.
  - runtime_observed: no.

- **MMS-V06** Tower continues through actual creator room refusal (MMS-R06b, MMS-R10; MTP-R09, MCI-R01/R02/R03; boundary: complete callback)
  - Inputs: selector 106; monster C/id 9 in game G, FX 6602752/FY 13123584, valid skill 193/count 400/level 3, signed missile 10/count 700/present row `Vel` 10/`VelLev` 4; cache distinct T kind 1/id 20, saved target key (1,20), registry maps exactly T, T mode 1. Owner room absent. Initial condition false; counter 41; no allocations. Each creator executes actual MCI validation/start/room refusal, without a supplied arbitrary return. Cases start independently.
  - Seeds: game {1,666}, caster {123,456}.
  - Draws: none across each complete callback.
  - Trace: set condition, refresh preserves T, compute 10+trunc(4*3/8)=11 whole units. Read caster source (100,200), emit all 64 requests; actual MCI refuses each absent owner room. T coordinates are never requested.
  - Outputs: result 1, condition true; seeds/counter unchanged, no allocations. All 64 complete requests have owner C, source unit absent, missile 10, explicit source (100,200), skill 193, level 3, loops 0, speed 11, target unit/callback absent and other zero/absent semantic defaults. Controls: explicit-start, start-relative-aim, whole-unit speed override; already-fixed-point clear. Exact aims: [[30,0],[29,2],[29,5],[28,8],[27,11],[26,14],[24,16],[23,19],[21,21],[19,23],[16,24],[14,26],[11,27],[8,28],[5,29],[2,29],[0,30],[-2,29],[-5,29],[-8,28],[-11,27],[-14,26],[-16,24],[-19,23],[-21,21],[-23,19],[-24,16],[-26,14],[-27,11],[-28,8],[-29,5],[-29,2],[-30,0],[-29,-2],[-29,-5],[-28,-8],[-27,-11],[-26,-14],[-24,-16],[-23,-19],[-21,-21],[-19,-23],[-16,-24],[-14,-26],[-11,-27],[-8,-28],[-5,-29],[-2,-29],[0,-30],[2,-29],[5,-29],[8,-28],[11,-27],[14,-26],[16,-24],[19,-23],[21,-21],[23,-19],[24,-16],[26,-14],[27,-11],[28,-8],[29,-5],[29,-2]].
  - Discriminates: aiming at T, emitting eight requests, swapping axes, fixed-point speed interpretation or stopping on refusal changes complete requests.
  - runtime_observed: no.
  - (a) Inputs: shared defaults. Outputs: complete shared outputs above.
  - (b) Inputs: `Vel` 0/`VelLev` 1/level -1. Trace: trunc(-1/8)=0, speed 0. Outputs: same 64 aims and result 1; request level -1/speed 0, speed override clear and already-fixed-point clear; all other shared fields, unchanged seeds/counter and no allocations.
  - (c) Inputs: saved key resolves different reference H. Trace: refresh clears only cache, preserves key. Outputs: result 0, condition true, cache absent/key (1,20) unchanged; no speed-getter/builder/creator call, no allocation/draw; seeds/counter unchanged.

- **MMS-V07** Tower first creation uses whole-unit speed and real seed prefix (MMS-R06b, MMS-R10; MCI-R01 to MCI-R11, MIR-R01a; boundary: stop after first metadata writes)
  - Inputs: MMS-V06(a), except owner room R rectangle (90,190,80,80), level 80, act environment E2; expansion true; row `CanSlow` false/`CollideType` 3; first U/metadata allocations supplied successful and zeroed. No subsequent event/path operation has run.
  - Seeds: game {1,666}, caster {123,456}, U initially {0,0}.
  - Draws: one game step to {1791398751,0}; no caster/child step by this boundary.
  - Trace: first request source (100,200), relative aim (30,0), owner C/source unit absent, missile 10/skill 193/level 3; override 11 whole units => 2816 fixed-point => 2112 after global scale. Explicit-start permits absent source; select R, retain (130,200), pass distance (30,0), allocate/seed/number U and initialize metadata.
  - Outputs: game seed {1791398751,0}/counter 42; U type missile/class 10/id 42/mode 3/game reference/act 2/E2/expansion/A/B true/selector 11/auxiliary absent; U seed {1791398751,666}, initial copy 1791398751; metadata unit id 42/source id -1/C/D false. Handoff start R/(100,200), retained (130,200), target unit absent, speed 2112. C seed unchanged/condition true/cache T/key unchanged. No final creator/callback return, later requests or whole-ring draw total asserted.
  - Discriminates: target-position aim or source offset changes handoff; treating 11 as fixed-point gives 8 instead of 2112; requiring source despite explicit-start wrongly refuses.
  - runtime_observed: no.

### 8.2 Observation vectors

**MMS-O01** Observe a poison nova action and record ordered creator requests, both speed phases, actual creator returns and final action result. Compare only with a vector using the captured compiled inputs. The companion describes reproduction; pending observation does not verify the gated targeted or sustained handlers.

## 9. Acceptance and dependencies

Partial acceptance comprises MMS-R01's named cells and MMS-V01 to MMS-V07. MMS-R03 is available for a present caster and supplied inputs in the nonoverflowing spacing stage; passing it does not establish the actual number of successfully created missiles. Tower callback/builder stages are available under MMS-R06b/MMS-R10, with actual first-creation prefixes consumed from MCI/MIR rather than inferred from return 1. Full reproduction requires MMS-Q01 to MMS-Q09 to close with executable cases and compatible observations for all in-scope routes.

**Depends on:** MTP-R09 for target refresh and saved-key behavior; SCX-R01 through SCX-R11/SCX-R21 for explicit bytecode and level-parameter evaluation; SCD-R06 for callback selection/arguments; GDA-R11a for supplied skill records; SCX-R11 for skill calculation front-end requests with context and SCX's gated compiler/parameter integration; NRF-R10b for dynamic position components; MCI-R01 to MCI-R11 and MIR-R01a/MIR-R03/MIR-R04/MIR-R07a/MIR-R07b/MIR-R16 for whole creator requests and final outcomes. Their documented gates remain their own integration boundaries.

**Depended on by:** no known plan/code/test consumer.

## 10. Related specs

[Missile movement](missile-movement-and-collision.md) owns later motion and collision; [missile hit calculation](missile-hit-and-damage-calculation.md) owns hit effects. [Monster Diablo](skill-monster-diablo.md) and [monster Baal](skill-monster-baal.md) own distinct boss action handlers even when their data uses a shared breath or teleport callback.

## Appendix A. Provenance

Original LoD Expansion 1.13c Windows x86 D2Game instructions, action table and direction arrays were read on 2026-10-29 against `LOD113C-D2GAME` in the [manifest](../research/source-manifest.md). Existing compatible calculation, coordinate, target-refresh and missile contracts retain their dates and limitations. New 2026-10-29 original-image readings establish the tower callback, speed helper, ring builder and its complete direction constants. The [companion](../research/spec-evidence/skill-monster-missiles-and-novas.md) records named archive cells, original-image decoding, exact request layout mappings, unresolved callees and observation limits. No gameplay rule is runtime-confirmed.
