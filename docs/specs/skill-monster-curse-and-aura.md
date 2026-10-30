# Monster curse and aura skills

**Description**

The monster curse action makes one random choice from five fixed skills before checking the selected skill record. Which units receive the selected effect and exactly what happens to them still require investigation.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-game-data-fields](core-game-data-fields.md)
Ids: MCA-R, MCA-Q, MCA-V, MCA-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-curse-and-aura.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-curse-and-aura.md)

## 1. Scope

The server skill actions selected by Skills.srvdofunc values 111, 112.
The boundary begins at the selected action entry and ends at its observable
return and effect requests. AI selection, skill dispatch, generic damage,
missile simulation and monster creation are separate contracts; their integration
here remains gated. This document does not establish other boss actions.

## 2. Trigger and ownership

D2Game computes these actions through the server do-function table selected by
Skills.srvdofunc. [Skill dispatch](skill-cast-start-and-do-dispatch.md) owns that
selection under SCD-R06, SCD-R07 and SCD-R08. AI selection is an outside-boundary producer: it does not gate the supplied action context. Frequency and relative frame ordering are not asserted by this contract.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning | Origin | Binding |
|---|---|---|---|---|---|
| Action selector | Skills.srvdofunc | integer | Selects one action in Scope | table data | selected by data from engine-defined handlers |
| Skill index | Dispatcher-supplied skill identity | signed 32-bit | Record lookup input | runtime input | supplied input |
| Skill level | Dispatcher-supplied level | signed 32-bit | Calculation input | runtime input | supplied input |
| Actor and game | Current action context | logical units and game | Owner, targets, difficulty and seed; exact consumption gated | runtime input | supplied input |
| Fixed selected identities | Engine selection mapping | signed 32-bit | 66, 72, 82, 87, 91 in index order | engine code | fixed by original code |
| Actor seed | Actor random state | two unsigned 32-bit halves | Updated by MCA-R01 | runtime input | supplied input |
| Selected-record collection and count | Supplied loaded Skills collection and signed record count | logical collection; signed 32-bit count | V07 supplies an empty collection with count 0; contextual validation input used only after MCA-R01 ends, never read by its selection stage | runtime input | supplied context |
| Selected skill record and loaded counts | Supplied Skills record, signed skill/state counts | logical record, signed 32-bit counts | R04 checks selected identity and target-state range | runtime input | supplied record context |
| Target state | Skills.auratargetstate, selected record; [local field evidence](../research/spec-evidence/skill-monster-curse-and-aura.md#selected-field-names) | signed 16-bit | Target state passed in the request | table data | resolved from supplied loaded record |
| Radius/duration calculations | Selected Skills.aurarangecalc / auralencalc; [local field evidence](../research/spec-evidence/skill-monster-curse-and-aura.md#selected-field-names) | unsigned 32-bit references, signed 32-bit results | Requested in this order under SCX-R11 | table data | resolved from supplied loaded record |
| Parameter 7 | Selected Skills.Param7 | signed 32-bit | Copied into index-4 speed/attack-rate requests | table data | resolved from supplied loaded record |
| Do-handled latch | Actor animation condition, SCD-R01/SCD-R07/SCD-R09 | boolean | Set after valid record/state guards | runtime input | shared actor condition |
| Supplied expression/recipient operations | SCX-R11 and MCA-Q04 recipient boundary | logical logged operations | Expression return/draw behavior and normally returning recipient result supplied by fixture | runtime input | dependency operations |
| Effective skill parameters | Skills record | unresolved | Parameters, calculations and effect selectors, gated by MCA-Q01 | unresolved | unresolved |

The selector mappings are code-table observations, not verified archive cell claims.
The effective text-file selector cells are established by MCA-R03; the named identity/start/action selector correspondence is corroborated in the limited observed profile; other compiled fields remain gated.

## 4. Rules (execution order)

**MCA-R01** Fixed random skill selection [verified in 1.13c code (static)]

For action selector 112 with a supplied non-null actor, advance that actor's
64-bit seed once using NRF-R01 and NRF-R02 in the numeric foundation. Reduce the
unsigned new low 32-bit half modulo five. Map indices 0, 1, 2, 3 and 4 to
skill identities 66, 72, 82, 87 and 91 respectively. These identities are fixed
by engine code, not obtained from the caller's skill index or a table row.
This draw and stored seed update happen before validating the selected record.
The selected identity is an intermediate result; this rule asserts no effect
application, recipient, final return, or caller-skill behavior.

**MCA-R02** Effect selection and application [unresolved]

For selector 111, recipient enumeration and aura effects remain unresolved under Q03. For selector 112, recipient eligibility, application/refresh of the prepared state/stat/event requests and production of the actual processor result remain unresolved under Q04; R04/R05 establish the supplied preparation and return forwarding independently.

**MCA-R03** Original text-file action bindings [verified in 1.13c data]

The named cells below are in effective `data/global/excel/skills.txt`, selected
by the `skill` column and read by the `srvdofunc` header. Origin: table data.
Binding: selected by data from engine-defined handlers. They establish text-file
contents. The named identity/start/action selectors additionally match the loaded records in the limited source profile recorded in the companion; other field correspondence remains MCA-Q01.

| Row identity (`skill`) | `srvdofunc` raw decimal value |
|---|---|
| FetishAura | 111 |
| MonCurseCast | 112 |

Source: [table-data record](../research/spec-evidence/skill-monster-curse-and-aura.md#table-data-record).

**MCA-R04** Selected-record validation and calculation order [verified in 1.13c code (static)]

After R01, reject with result 0 if the selected signed skill identity is
negative, at least the signed loaded skill count, or its record is absent.
Otherwise read that selected record's signed 16-bit `auratargetstate`; reject
with result 0 if it is negative or at least the signed loaded state count.
These rejections make no calculation or recipient request and do not change
the do-handled latch. R01's preceding seed update is retained.

The target-state and two expression field-name mappings are established by the [local loader-and-consumer evidence](../research/spec-evidence/skill-monster-curse-and-aura.md#selected-field-names); GDA-R11a supplies the shared record/Param7 schema, not these locally documented field mappings.

On success set the actor's do-handled latch true, then request calculation of
the selected record's `aurarangecalc`, followed by `auralencalc`, both through
SCX-R11 with context (actor, selected skill identity, caller-supplied level).
Use their signed 32-bit returns as radius and duration without local clamping.
The caller's skill identity does not replace the randomly selected identity.
There is no local draw in R04; expression-owned draws retain SCX's labels and
gates. Supplied results establish this caller's order independently of expression
production and evaluation. No positive-level check is performed here.

**MCA-R05** Five effect requests and recipient handoff [verified in 1.13c code (static)]

With R04's selected record, state, radius and duration supplied, prepare one
effect request containing game, actor, selected skill identity, caller level,
duration and selected target state. Set monster-only eligibility false by
default; prepare no primary stat, no additional stats and no event bindings
unless the index's row below supplies them. All stat values are signed 32-bit;
stat identities and event-binding choices are fixed by engine code. Parameter
7 is the selected record's signed 32-bit `Param7` under GDA-R11a.

| R01 index / selected skill | Primary stat/value | Ordered additional stat/value pairs | Monster-only | Ordered event id / callback-selector bindings |
|---|---|---|---|---|
| 0 / 66 | 36 `damageresist` = -100 | none | false | none |
| 1 / 72 | 25 `damagepercent` = -50 | none | false | none |
| 2 / 82 | none | none | false | (5,4) |
| 3 / 87 | none | none | false | (1,5), (2,5) |
| 4 / 91 | 67 `velocitypercent` = Param7 | (25 `damagepercent`,-50), (36 `damageresist`,-50), (68 `attackrate`,Param7) | true | none |

These are exactly the prepared requests; the skill names or expected player
spell effects do not override this fixed mapping. The event ids and callback
selectors identify requested event bindings, not their callback arithmetic.
Unused stat/event entries produce no requested binding. This semantic request
does not prescribe storage, sentinel values or an array layout.

Request recipient processing with filter 3, actor, supplied radius, the shared
curse-recipient operation and this effect request. The recipient operation
consumes the semantic fields above; its eligibility, state/stat writes and event
registration effects remain MCA-Q04. Forward the normally returning recipient
processor's integer return unchanged as the action result. R05 makes no local
random draw. Dependency-owned draws/effects are outside this supplied-request
stage and cannot be inferred from a prepared request.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MCA-R01 | Fixed random skill selection | verified in 1.13c code (static) | none | MCA-V01, MCA-V02, MCA-V03, MCA-V04, MCA-V05, MCA-V06, MCA-V07 |
| MCA-R02 | Action effects | unresolved | MCA-Q03, MCA-Q04 block their respective handler effects; Q01 compiled integration | none |
| MCA-R03 | Named text-file selectors | verified in 1.13c data | none for named text cells and observed selector correspondence; MCA-Q01 gates other compiled fields | table-data record |
| MCA-R04 | Selected record/state guards and radius/duration order | verified in 1.13c code (static) | MCA-Q01 actual data production; supplied records/results available | MCA-V08 through MCA-V14, MCA-V16 through MCA-V19 |
| MCA-R05 | Fixed request construction and return forwarding | verified in 1.13c code (static) | MCA-Q04 recipient effects only; supplied request/return available | MCA-V08 through MCA-V12, MCA-V15, MCA-V18, MCA-V19, MCA-V20 |

## 5. Outputs and side effects

| Output | Rule | Evidence boundary |
|---|---|---|
| Actor seed advanced once | MCA-R01 | Supplied actor; entry selection stage |
| Actual recipient/effect execution and produced processor result | MCA-R02 | Gated; R05 separately establishes prepared requests and forwarding of a supplied processor result |

| Failure result 0 without expression/recipient calls | MCA-R04 | Invalid selected record or target state; prior R01 seed update retained |
| Actor do-handled latch true; ordered radius/duration requests | MCA-R04 | Valid selected record/state; expression effects belong to SCX |
| Prepared effect request and recipient handoff, forwarded return | MCA-R05 | Recipient writes and event effects excluded |

## 6. Branches and edge cases

The selected-record/state guards and signed supplied calculation returns are
established by R04; the request construction is R05. Full effects, actual
compiled parameter/expression/state values, recipient unit types and helper
failure effects retain Q01/Q03/Q04. These integration limits do not gate the
supplied record/result stages. There is no local positive-level, radius or
duration clamp. No Classic behavior is asserted.

## 7. Open questions

**MCA-Q01** [unresolved]: Which remaining compiled effect, state, stat and parameter fields correspond to the named text records and supplied effect fields? The named identity/start/action selector correspondence is established for the observed profile. Retain narrow loaded-record value comparisons for the remaining fields. The local target-state/radius/duration field names are established; their names are not gated merely because the shared field inventory is partial. This is an integration gate
on original-profile record production, not the supplied-input arithmetic or
record preparation. AI selection remains outside this boundary.

**MCA-Q02** [unresolved]: Which executable cases match an owned, identified
Expansion 1.13c game? Pending observation does not gate the isolated verified branch,
but does prevent a claim of runtime confirmation.

**MCA-Q03** [unresolved]: 111 recipient enumeration and effect. Which target position, eligible monster classes and relationship predicate determine the aura recipients, and which exact effect is applied?

**MCA-Q04** [unresolved]: 112 recipient execution. Which target-coordinate and room/recipient eligibility operations lead the shared processor to invoke the curse recipient operation, and which state/stat/event changes does that operation perform? R04/R05 establish selected-record guards, formula request order, fixed request shapes and forwarding of a supplied processor return; these are not gated by unknown recipient effects. Close the remaining target/room predicates, curse-list producer and event callback registrations to establish complete action effects.

## 8. Test vectors

### 8.1 Executable vectors

- **MCA-V01** Select index 0 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 0}.
  - Draws: (actor, modulus 5, result 0).
  - Trace: 1 times 1791398085 plus 0 = 1791398085; new high = 0; new low modulo 5 = 0.
  - Outputs: actor seed = {lo: 1791398085, hi: 0}; selected skill identity = 66.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V02** Select index 1 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 1}.
  - Draws: (actor, modulus 5, result 1).
  - Trace: 1 times 1791398085 plus 1 = 1791398086; new high = 0; new low modulo 5 = 1.
  - Outputs: actor seed = {lo: 1791398086, hi: 0}; selected skill identity = 72.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V03** Select index 2 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 2}.
  - Draws: (actor, modulus 5, result 2).
  - Trace: 1 times 1791398085 plus 2 = 1791398087; new high = 0; new low modulo 5 = 2.
  - Outputs: actor seed = {lo: 1791398087, hi: 0}; selected skill identity = 82.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V04** Select index 3 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 3}.
  - Draws: (actor, modulus 5, result 3).
  - Trace: 1 times 1791398085 plus 3 = 1791398088; new high = 0; new low modulo 5 = 3.
  - Outputs: actor seed = {lo: 1791398088, hi: 0}; selected skill identity = 87.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V05** Select index 4 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 4}.
  - Draws: (actor, modulus 5, result 4).
  - Trace: 1 times 1791398085 plus 4 = 1791398089; new high = 0; new low modulo 5 = 4.
  - Outputs: actor seed = {lo: 1791398089, hi: 0}; selected skill identity = 91.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V06** Multiply high and carry (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1; stop after selection before record validation.
  - Seeds: actor = {lo: 3, hi: 4294967295}.
  - Draws: (actor, modulus 5, result 3).
  - Trace: 3 times 1791398085 = 5374194255; adding 4294967295 gives 9669161550; split at 4294967296 gives low 1079226958, high 2; low modulo 5 = 3.
  - Outputs: actor seed = {lo: 1079226958, hi: 2}; selected skill identity = 87.
  - Discriminates: retaining old high, writing zero high, or omitting addition carry fails high = 2; reducing the unsplit product fails selected identity.
  - runtime_observed: no

- **MCA-V07** Unsigned reduction before rejected selected record (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1; selected skill collection has count 0, so skill 66 is invalid when validation eventually begins. A logical observer records selected-record validation and record accesses; its log is initially empty. Stop immediately after selection, before validation.
  - Seeds: actor = {lo: 2, hi: 0}.
  - Draws: (actor, modulus 5, result 0).
  - Trace: 2 times 1791398085 plus 0 = 3582796170; split gives low 3582796170, high 0; unsigned low modulo 5 = 0; fixed identity at index 0 is 66; no selected-record validation has yet occurred at this boundary.
  - Outputs: actor seed = {lo: 3582796170, hi: 0}; selected skill identity = 66; selected-record validation/access log empty. No assertion is made about the subsequent return or effects.
  - Discriminates: interpreting low as signed gives -712171126 and a negative signed remainder, failing index 0/identity 66; validating selected record before drawing either logs an early validation or rejects without advancing seed, failing the asserted boundary state.
  - runtime_observed: no

- **MCA-V08** Prepare selected index 0 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 0, selected skill 66; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 66 and level 2; initialize default request and choose row 0; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,66,2), duration(actor,66,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=66, level=2, duration=29, targetState=42, primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V09** Prepare selected index 1 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 1, selected skill 72; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 72 and level 2; initialize default request and choose row 1; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,72,2), duration(actor,72,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=72, level=2, duration=29, targetState=42, primaryStat=25, primaryValue=-50, additionalStats=[], monsterOnly=false, eventBindings=[]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V10** Prepare selected index 2 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 2, selected skill 82; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 82 and level 2; initialize default request and choose row 2; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,82,2), duration(actor,82,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=82, level=2, duration=29, targetState=42, primaryStat=none, primaryValue=none, additionalStats=[], monsterOnly=false, eventBindings=[(5, 4)]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V11** Prepare selected index 3 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 3, selected skill 87; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 87 and level 2; initialize default request and choose row 3; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,87,2), duration(actor,87,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=87, level=2, duration=29, targetState=42, primaryStat=none, primaryValue=none, additionalStats=[], monsterOnly=false, eventBindings=[(1, 5), (2, 5)]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V12** Prepare selected index 4 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 4, selected skill 91; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 91 and level 2; initialize default request and choose row 4; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,91,2), duration(actor,91,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=91, level=2, duration=29, targetState=42, primaryStat=67, primaryValue=-12, additionalStats=[(25, -50), (36, -50), (68, -12)], monsterOnly=true, eventBindings=[]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V13** Reject invalid selected state (MCA-R04; boundary: supplied-input stage)
  - Inputs: entry after R01; index 0, selected identity 66, actor present, latch false, caller level 2, skill count 400; selected record present with auratargetstate=-1; state count 100; expression and recipient logs empty and record reads observed. All later parameters are uninspected.
  - Seeds: no draw inside this post-selection stage; R01 already updated the seed.
  - Draws: none.
  - Trace: selected state is negative, so return before latch, expression or recipient operations.
  - Outputs: result=0, latch false, expression/recipient call logs empty; no additional seed change.
  - Discriminates: proceeding with an invalid state would set latch or issue an operation, violating the asserted unchanged latch/empty logs.
  - runtime_observed: no

- **MCA-V14** Reject invalid selected record (MCA-R04; boundary: supplied-input stage)
  - Inputs: entry after R01; index 0, selected identity 66, actor present, latch false, caller level 2, skill count 400; selected record absent; state count 100; expression and recipient logs empty and record reads observed. All later parameters are uninspected.
  - Seeds: no draw inside this post-selection stage; R01 already updated the seed.
  - Draws: none.
  - Trace: selected record is absent, so return before latch, expression or recipient operations.
  - Outputs: result=0, latch false, expression/recipient call logs empty; no additional seed change.
  - Discriminates: proceeding with an invalid record would set latch or issue an operation, violating the asserted unchanged latch/empty logs.
  - runtime_observed: no

- **MCA-V15** Forward recipient failure result (MCA-R05; boundary: supplied-input stage)
  - Inputs: prepared index-0 context from V08 as independent inputs: actor/game present, skill 66, level 2, state 42, radius 13, duration 29; primary(36,-100), no additional stats/events, monsterOnly=false; recipient operation logs this request, changes no state, draws nothing and returns 0 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: issue recipient request and forward its return 0.
  - Outputs: recipient log contains exactly (filter 3,actor,radius 13,prepared request); final result 0; no local seed change.
  - Discriminates: unconditionally succeeding after recipient processing returns 1 and fails the final-result assertion.
  - runtime_observed: no

- **MCA-V16** Selected identity equals skill count (MCA-R01, MCA-R04; boundary: supplied-input stage)
  - Inputs: action selector = 112; game/actor present; caller skill identity 999, level 2; loaded skill count = 66; actor latch false; record-read, expression and recipient logs initially empty. A logical observer records access to individual skill records, separately from reading the loaded count. All state/effect fields are uninspected.
  - Seeds: actor = {lo: 1, hi: 0}.
  - Draws: (actor, modulus 5, result 0).
  - Trace: 1 times 1791398085 plus 0 gives seed {1791398085,0}; low modulo 5 selects index 0 and identity 66; 66 is at least skill count 66, so return before individual record access.
  - Outputs: actor seed = {lo: 1791398085, hi: 0}; result = 0; latch false; individual-record-read, expression and recipient logs empty.
  - Discriminates: allowing equality at the upper bound reaches individual record lookup; rejecting before drawing leaves the entry seed unchanged. Either fails the boundary assertions.
  - runtime_observed: no

- **MCA-V17** Target state equals state count (MCA-R04; boundary: supplied-input stage)
  - Inputs: entry immediately after R01; actor present, latch false; index 0, selected identity 66; caller level 2; skill count 400 and selected record present; selected auratargetstate = 100, loaded state count = 100; expression and recipient logs empty. Later parameter fields are uninspected.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is already complete.
  - Draws: none.
  - Trace: selected skill record is valid; selected state 100 is at least count 100; return before latch or expression/recipient operations.
  - Outputs: result = 0; latch false; expression and recipient logs empty; no additional seed change.
  - Discriminates: accepting equality at the state bound sets latch or issues an operation, violating the asserted unchanged latch and empty logs.
  - runtime_observed: no

- **MCA-V18** Preserve negative radius and duration (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry after R01; game/actor present; latch false; index 0, selected identity 66; caller identity 999, level 2; skill count 400, record present, selected target state 42, state count 100; selected Param7 = -12; distinct radius/duration references R,D. Calculation operations log their contexts, draw nothing, and return R=-3, D=-7. Recipient operation logs its arguments/request, changes no state, draws nothing and returns 1 normally.
  - Seeds: no draw inside this post-selection stage.
  - Draws: none.
  - Trace: valid record/state; set latch; request radius then duration and retain signed returns -3 and -7; prepare index-0 request and issue recipient processing without clamping or omitting it; forward 1.
  - Outputs: latch true; ordered log = [radius(actor,66,2), duration(actor,66,2), recipient(filter=3,actor,radius=-3,request)]; request game/actor are the supplied context, skill=66, level=2, targetState=42, duration=-7, primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; no other binding; final result=1; no local seed change.
  - Discriminates: clamping signed results to positive values, rejecting nonpositive radius/duration, or omitting the recipient operation fails the exact logged radius/duration and recipient call assertion. No actual recipient behavior is inferred from these supplied returns.
  - runtime_observed: no

- **MCA-V19** Preserve zero radius and duration (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry after R01; game/actor present; latch false; index 0, selected identity 66; caller identity 999, level 2; skill count 400, record present, selected target state 42, state count 100; selected Param7 = -12; distinct radius/duration references R,D. Calculation operations log their contexts, draw nothing, and return R=0, D=0. Recipient operation logs its arguments/request, changes no state, draws nothing and returns 1 normally.
  - Seeds: no draw inside this post-selection stage.
  - Draws: none.
  - Trace: valid record/state; set latch; request radius then duration and retain signed returns 0 and 0; prepare index-0 request and issue recipient processing without clamping or omitting it; forward 1.
  - Outputs: latch true; ordered log = [radius(actor,66,2), duration(actor,66,2), recipient(filter=3,actor,radius=0,request)]; request game/actor are the supplied context, skill=66, level=2, targetState=42, duration=0, primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; no other binding; final result=1; no local seed change.
  - Discriminates: clamping signed results to positive values, rejecting nonpositive radius/duration, or omitting the recipient operation fails the exact logged radius/duration and recipient call assertion. No actual recipient behavior is inferred from these supplied returns.
  - runtime_observed: no

- **MCA-V20** Forward a non-Boolean processor result (MCA-R05; boundary: supplied-input stage)
  - Inputs: prepared index-0 context as independent inputs: actor/game present, selected skill 66, caller level 2, targetState=42, radius=13, duration=29; primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; recipient operation logs this request, changes no state, draws nothing and returns signed 32-bit integer 37 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: issue recipient processing with filter 3, actor, radius 13 and the prepared request; forward its normally returned integer 37 unchanged.
  - Outputs: recipient log contains exactly (filter=3,actor,radius=13,prepared request); final result=37; no local seed or state change.
  - Discriminates: Booleanizing a nonzero result gives 1 and fails final result 37. This supplied fixture checks caller forwarding and makes no claim that the actual processor produces 37.
  - runtime_observed: no

### 8.2 Observation vectors

**MCA-O01** Observe the selected action and its final return/effects on an
owned Expansion 1.13c game after establishing the missing bindings. Expected
values for gated effects are deliberately unspecified; the
[procedure](../research/spec-evidence/skill-monster-curse-and-aura.md#observation-procedure)
defines what evidence must be retained. This is a pending investigation, not a
matching observation.

## 9. Acceptance and dependencies

MCA-V01 through MCA-V07 establish MCA-R01 at the selection stage; MCA-V08 through MCA-V20 establish R04/R05 with their supplied records and operation results; V16 also includes R01. These fixtures assert prepared requests and the action result derived by forwarding a supplied processor return. They do not establish the actual recipient effects or production of the processor return in gameplay.
Full reproduction requires resolving MCA-Q03, MCA-Q04 and MCA-Q01, specifying executable vectors for
MCA-R02, and retaining attributable observations under MCA-Q02. A partial result
does not establish parity for these actions.

**Depends on:** [skill dispatch](skill-cast-start-and-do-dispatch.md) SCD-R06, SCD-R07 and SCD-R08 for selecting
the action from the server selector and invoking its handler; successful effect
integration remains MCA-Q01. No damage, missile or monster-creation outcome is
credited through an unnamed dependency.

**Depended on by:** no established consumer.

[Skill calculations](skill-calc-expressions.md) SCX-R11 supplies the expression front-end contract and retains producer-specific gates; GDA-R11a in [game-data records](core-game-data-table-records.md) supplies the Param7 schema. SCD-R01/SCD-R07/SCD-R09 name the shared do-handled latch.

## 10. Related specs

[Skill dispatch](skill-cast-start-and-do-dispatch.md) supplies the action context.
[Numeric foundations](core-numeric-and-rng-foundations.md) owns integer and seed
operations when their consumption is established.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c D2Game, identified by
[source manifest](../research/source-manifest.md), was examined through static
instruction listings on 2026-10-29. The
[companion](../research/spec-evidence/skill-monster-curse-and-aura.md) identifies the dispatch slots,
read boundary and unresolved callees. Named selector text cells have the companion record; no gameplay case has been confirmed.
