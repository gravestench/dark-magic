# Baal monster skills

**Description**

The corpse action processes selected dead monsters one at a time through the shared corpse-explosion action. The search and the explosion have separate eligibility checks, so selecting a corpse does not guarantee an explosion. Complete cold-missile and tentacle effects still require investigation.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-game-data-fields](core-game-data-fields.md)
Ids: BMS-R, BMS-Q, BMS-V, BMS-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-baal.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-baal.md)

## 1. Scope

The server skill actions selected by Skills.srvdofunc values 139, 140, 141.
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
| Enumeration base and increment | Skills.Param5 and Skills.Param6 | signed 32-bit | Local linear search-parameter operands | runtime input | supplied record fields |
| Skill lookup validity | Supplied record lookup result | boolean | Whether selected record exists | runtime input | supplied input |
| Enumeration operation | Supplied continuation | logical operation | Logs request search parameter and returns normally; recipient effects gated | runtime input | supplied operation |
| Default-radius record | Skills record for identity 285; GDA-R11a | optional record; signed i32 Param1 and Param2 | Used only when the raw search parameter is zero |
| Default Skills count | loaded Skills collection | signed i32 | Exclusive bound for the fixed identity 285 |
| Candidate position and mode | candidate unit and its position storage; core unit identities | u32 kind; i32 mode; integer path coordinates | Inputs to the fixed corpse-query predicate |
| Search center and radius | prepared query | signed i32 path units | Center and effective radius for the candidate test |
| Ordered collected recipients | supplied successful collection phase | stable ordered list of readable unit references | Actual room traversal and list production remain BMS-Q05 |
| Selected target | actor's target relation | optional unit reference with kind and identity | Replaced before every explosion-action invocation |
| Explosion action | DSC-R01a through DSC-R14b | operation on game, actor, skill identity and level | Same shared action, preserving all of DSC's producer and effect gates |
| Effective skill parameters | Skills record | unresolved | Parameters, calculations and effect selectors, gated by BMS-Q01 | unresolved | unresolved |

The selector mappings are code-table observations, not verified archive cell claims.
The effective text-file selector cells are established by BMS-R03; the named identity/start/action selector correspondence is corroborated in the limited observed profile; other compiled fields remain gated.

| Mapping or constant | Origin | Binding |
|---|---|---|
| Default-radius and explosion skill identity 285 | engine code | Fixed by the corpse-action request and callback, independently of the invocation identity |
| Default-radius Param1 and Param2 | table data | Prepared values from the record selected by that fixed identity; compiled correspondence remains BMS-Q01 |
| Monster kind 1 and dead monster mode 12 in the query | engine code; names follow GDF-R05a | Fixed predicate; a different Skills.aurafilter does not replace it |

## 4. Rules (execution order)

**BMS-R01** Negative skill index [verified in 1.13c code (static)]

For action selector 139 only, an index less than zero returns failure (integer zero)
before accessing a skill record or calling a dependency. This isolated entry branch
does not draw RNG, write actor state, or emit an effect. The index is signed 32-bit.
This rule covers only a negative index, not any other invalid-record case.

**BMS-R02** Remaining action effects [unresolved]

The complete cold-missile and tentacle actions retain BMS-Q03/BMS-Q04.
For the corpse action, actual room/candidate production and the shared
explosion's resulting damage and state effects retain BMS-Q05 and DSC's gates.
Its local radius, predicate and ordered callback behavior are BMS-R04 through
BMS-R07, rather than an unverified recipient-effect default.

**BMS-R03** Original text-file action bindings [verified in 1.13c data]

The named cells below are in effective `data/global/excel/skills.txt`, selected
by the `skill` column and read by the `srvdofunc` header. Origin: table data.
Binding: selected by data from engine-defined handlers. They establish text-file
contents. The named identity/start/action selectors additionally match the loaded records in the limited source profile recorded in the companion; other field correspondence remains BMS-Q01.

| Row identity (`skill`) | `srvdofunc` raw decimal value |
|---|---|
| Baal Corpse Explode | 141 |
| Baal Tentacle | 140 |
| Baal Cold Missiles | 139 |

Source: [table-data record](../research/spec-evidence/skill-monster-baal.md#table-data-record).

**BMS-R04** Enumeration search parameter [verified in 1.13c code (static)]

For action selector 141, at the calculation stage with skill lookup validity,
level and two signed record values supplied, choose request search parameter zero if
lookup is invalid or level is nonpositive. Otherwise compute
`low32(base + low32(low32(level - 1) * increment))`, interpreted as signed
32-bit. There is no clamp, calculation callback or random draw in this stage.
Place that value in the enumeration request's search parameter and issue that request.
This raw search parameter is a radius override; BMS-R05 specifies the zero fallback.
This stage alone does not assert recipients or effects. After the supplied enumeration operation
returns normally, return 1 regardless of its return value. Its effects and RNG
draws are supplied dependency behavior, not established by this local rule.
Base is Skills.Param5 and increment is Skills.Param6 under the GDA-R11a field schema. Supplied compiled values establish this arithmetic; text-to-compiled correspondence remains BMS-Q01.

The named Baal Corpse Explode row also has raw Param5 cell 40 and an empty Param6 cell in the effective text capture. The empty cell is not promoted to a numeric compiled value. These data-only facts belong to BMS-R03.

**BMS-R05** Effective corpse-query radius [verified in 1.13c code (static)]

The corpse-action request fixes the default-radius skill identity to 285 and
passes the invocation level unchanged. For a present monster actor, its dynamic
path's unsigned 16-bit integer X/Y parts form the query center; an absent dynamic
path supplies (0,0). Other unit-kind position production is not established by
this rule. A nonzero raw search parameter is the effective radius unchanged,
including negative values; it makes no default-radius lookup. For raw zero,
look up identity 285. A missing/out-of-range record or a nonpositive invocation
level gives radius 0. Otherwise use signed i32 values from that record:
`radius = low32(Param1 + low32(low32(level - 1) * Param2))`.
No clamp or random draw occurs locally. The invocation identity's record does
not substitute for the fixed default-radius record. Position and original
compiled-record production retain their integration limits; the supplied
record calculation is available independently.

**BMS-R06** Fixed corpse-query predicate [verified in 1.13c code (static)]

For the corpse action's fixed query, a readable candidate is accepted exactly
when it is a monster, its mode is 12, and the following signed i32 comparison
passes. Let `dx=low32(candidateX-centerX)` and
`dy=low32(candidateY-centerY)`. Compute
`distanceSquared=low32(low32(dx*dx)+low32(dy*dy))` and
`radiusSquared=low32(radius*radius)`; accept distance when signed
`distanceSquared <= radiusSquared`. Equality is included; neither side is
widened or saturated. Negative radius is not converted to a default here.

For a monster, X and Y are the unsigned 16-bit integer parts of its dynamic
position. An absent dynamic path supplies (0,0) in this predicate. Other unit
kinds require readable position storage and are rejected by the fixed type
filter; this is not a safe-null-storage contract for them. This predicate draws
nothing and writes no unit state. It does not test hostility, ownership,
corpse-selectability or the corpse-explosion action's excluded-state mask.
Those last eligibility checks belong to DSC-R01b/DSC-Q04. Actual room selection
and production of candidates remain BMS-Q05; a supplied candidate test does not
establish that the world query would visit that candidate.

**BMS-R07** Ordered corpse callbacks and target hand-off [verified in 1.13c code (static)]

This stage begins after a successful collection phase with a stable ordered
recipient list and a present monster actor with a readable target path. For
each recipient in list order, replace the actor's selected target with that
recipient, including its kind and identity, then directly invoke the shared
corpse-explosion action with `(game, actor, skill identity 285, invocation level)`.
The selected target at that action's entry is the current recipient. The
invocation identity does not replace the fixed 285. Consume DSC-R01a through
DSC-R14b with these arguments and their existing labels and gates; do not add
another skill-dispatch validation, resource or delay stage around this direct
invocation.

When each action returns normally, continue to the next recipient regardless
of its numeric return. There is no local early stop or restoration of the
actor's earlier target. Any target changes made by the shared action remain
that action's effects until a later callback replaces the target. After the
last callback, or immediately for an empty list, request release of the
collection. Release-produced memory effects remain BMS-Q05. The outer corpse
action then returns 1 under BMS-R04. Target replacement and this traversal make
no direct random draw; all explosion/dependency draws retain DSC's accounting.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| BMS-R01 | Negative-index rejection | verified in 1.13c code (static) | none | BMS-V01 |
| BMS-R02 | Action effects | unresolved | BMS-Q03, BMS-Q04, BMS-Q05 block their respective handler effects; Q01 compiled integration | none |
| BMS-R03 | Named text-file selectors | verified in 1.13c data | none for named text cells and observed selector correspondence; BMS-Q01 gates other compiled fields | table-data record |
| BMS-R04 | Local supplied-input stage | verified in 1.13c code (static) | none at supplied-input boundary; Q01 retains downstream integration | BMS-V02, BMS-V03, BMS-V04, BMS-V05 |
| BMS-R05 | Zero-override radius fallback | verified in 1.13c code (static) | BMS-Q01 original compiled operands and BMS-Q05 position/query production only | BMS-V06 |
| BMS-R06 | Fixed monster-corpse predicate | verified in 1.13c code (static) | BMS-Q05 actual candidate production only; DSC-Q04 subsequent corpse eligibility | BMS-V07 |
| BMS-R07 | Target replacement and ordered direct explosion invocations | verified in 1.13c code (static) | BMS-Q05 actual collection/release and DSC gates for explosion-produced effects | BMS-V08, BMS-V09 |

## 5. Outputs and side effects

| Output | Rule | Evidence boundary |
|---|---|---|
| Failure result = 0, unchanged actor and seed | BMS-R01 | Negative-index branch only |
| Effect requests and final result | BMS-R02 | Gated; no concrete effect or final result asserted |
| Enumeration search parameter request and return 1 | BMS-R04 | Supplied lookup/values and normally returning operation; recipient effects excluded |
| Selected target replaced before each direct explosion invocation | BMS-R07 | Valid monster actor/path; shared-action target changes remain dependency effects |
| Ordered direct explosion-action and collection-release requests | BMS-R07 | Supplied stable recipient list and normally returning operations |

## 6. Branches and edge cases

Full action integration for successful effects, actual record production,
unit types, target eligibility, difficulty differences and effect-helper
failures retains BMS-Q01 and the handler-specific questions below. These
integration limits do not gate BMS-R04 with supplied lookup validity and
parameters: invalid lookup or nonpositive level supplies search parameter 0;
valid positive level uses the signed wrapping expression, including negative
results. BMS-R04 also establishes its normally returning continuation's result
1 independently of the supplied operation's return value. Client visuals do
not establish the remaining downstream effects. No Classic behavior is asserted.

The query's squared-distance comparison can accept a supplied distant candidate
after signed overflow; this does not establish that room collection would
produce that candidate. A dead monster passing BMS-R06 can still fail DSC's
corpse checks, or be the actor itself and fail the shared action's self-target
guard. Such a shared-action failure does not stop later callbacks or restore
the earlier selected target.

## 7. Open questions

**BMS-Q01** [unresolved]: Which remaining compiled effect, state, stat and parameter fields correspond to the named text records and supplied effect fields? The named identity/start/action selector correspondence is established for the observed profile. Resolve the loader
schema and retain narrow loaded-record comparisons. This is an integration gate
on original-profile record production, not the supplied-input arithmetic or
record preparation. AI selection remains outside this boundary.

**BMS-Q02** [unresolved]: Which executable cases match an owned, identified
Expansion 1.13c game? Pending observation does not gate the verified local stages,
but does prevent a claim of runtime confirmation.

**BMS-Q03** [unresolved]: 139 target, missile and direction. Which missile record, target resolution, placement and direction writes establish the cold-missile action?

**BMS-Q04** [unresolved]: 140 summon placement and lifetime. Which class and spawn parameters are selected, how do attempted coordinates resolve into rooms, and what lifetime follows successful creation?

**BMS-Q05** [unresolved]: 141 world-query and shared-effect integration. Which
rooms and ordered unit chains are visited, including the query's room-exclusion
condition, and what are the collection allocation/release effects? The fixed
candidate predicate, target replacement and direct shared-action invocation are
established by BMS-R05 through BMS-R07. Actual corpse validity, damage, states,
events and reaction retain DSC-Q01 through DSC-Q05; passing the query predicate
does not close those dependencies.

## 8. Test vectors

### 8.1 Executable vectors

- **BMS-V01** Reject a negative index (BMS-R01; boundary: isolated rule)
  - Inputs: signed skill index = -1; action selector = 139; all other inputs are uninspected by this branch. The fixture records every attempted skill-record read in a lookup log initially empty; this is a logical access observer, not a required API.
  - Seeds: no draw.
  - Draws: none.
  - Trace: -1 is less than zero, so the entry branch returns before record lookup.
  - Outputs: result = 0; record-lookup log empty; dependency call log empty; actor state unchanged.
  - Discriminates: clamping a negative index to zero and continuing would perform record lookup; the empty lookup/dependency log rejects that reading.
  - runtime_observed: no

- **BMS-V02** Positive enumeration search parameter (BMS-R04; boundary: supplied-input stage)
  - Inputs: action selector = 141; lookup valid; level = 4; base = 7; increment = -3; other enumeration request fields supplied unchanged. Enumeration operation logs search parameter, draws nothing, changes no state and returns integer 0 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: level minus one = 3; 3 times -3 = -9; 7 plus -9 = -2; request enumeration with search parameter -2; ignore its return 0 and return 1.
  - Outputs: request log = [enumeration(searchParameter=-2)]; final result = 1; no local seed update.
  - Discriminates: clamping negative search parameters fails -2; using level instead of level minus one gives -5; forwarding enumeration return fails final result 1.
  - runtime_observed: no

- **BMS-V03** Zero search parameter on invalid lookup (BMS-R04; boundary: supplied-input stage)
  - Inputs: action selector = 141; lookup valid = false; level = 4; base = 7; increment = -3; other enumeration request fields supplied unchanged. Enumeration operation logs search parameter, draws nothing, changes no state and returns integer 0 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: invalid lookup selects search parameter 0 without evaluating the linear expression; issue enumeration request, ignore its return 0, return 1.
  - Outputs: request log = [enumeration(searchParameter=0)]; final result = 1; no local seed update.
  - Discriminates: evaluating the unchecked linear expression gives -2 instead of 0; suppressing the request loses its log entry; forwarding its return produces 0 instead of 1.
  - runtime_observed: no

- **BMS-V04** Zero search parameter on nonpositive level (BMS-R04; boundary: supplied-input stage)
  - Inputs: action selector = 141; lookup valid = true; level = 0; base = 7; increment = -3; other enumeration request fields supplied unchanged. Enumeration operation logs search parameter, draws nothing, changes no state and returns integer 0 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: nonpositive level selects search parameter 0 without evaluating the linear expression; issue enumeration request, ignore its return 0, return 1.
  - Outputs: request log = [enumeration(searchParameter=0)]; final result = 1; no local seed update.
  - Discriminates: evaluating the unchecked linear expression gives 10 instead of 0; suppressing the request loses its log entry; forwarding its return produces 0 instead of 1.
  - runtime_observed: no

- **BMS-V05** Signed wrapping search parameter (BMS-R04; boundary: supplied-input stage)
  - Inputs: action selector = 141; lookup valid = true; level = 2; base = 2147483647; increment = 1; other enumeration request fields supplied unchanged. Enumeration operation logs search parameter, draws nothing, changes no state and returns integer 0 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: level minus one = 1; product = 1; exact sum 2147483648 retains low 32 bits and is interpreted as signed -2147483648; request enumeration with that value; ignore its return 0 and return 1.
  - Outputs: request log = [enumeration(searchParameter=-2147483648)]; final result = 1; no local seed update.
  - Discriminates: saturating signed arithmetic yields 2147483647 and unbounded exact arithmetic yields 2147483648, both failing the request-value assertion.
  - runtime_observed: no

- **BMS-V06** Effective-radius fallback (BMS-R05; boundary: supplied-input stage)
  - Inputs: invocation identity = 999; invocation level = 4; fixed default identity = 285; default Skills count = 357; its prepared record is present with `Param1` = 7 and `Param2` = 2. Present monster actor M has a dynamic path with unsigned 16.16 X = 6619135 and Y = 13107201, whose integer parts are (100,200). Each case starts with an empty default-helper entry log and observes radius and center before room collection.
  - Seeds: no draw.
  - Draws: none.
  - Cases:
    - (a) Inputs: raw search parameter = 40. Trace: nonzero bypasses the default helper. Outputs: radius = 40; helper log empty.
    - (b) Inputs: raw search parameter = 0. Trace: enter default helper with (285,4); 7 + (4-1)*2 = 13. Outputs: radius = 13; helper log = [(285,4)].
    - (c) Inputs: raw search parameter = 0; level = 0. Trace: enter helper (285,0); nonpositive level selects 0. Outputs: radius = 0; helper log = [(285,0)].
    - (d) Inputs: raw search parameter = -3. Trace: negative is nonzero and bypasses the helper. Outputs: radius = -3; helper log empty.
    - (e) Inputs: raw search parameter = 0; level = 2; default `Param1` = 2147483647, `Param2` = 1. Trace: 2147483647 + 1 wraps to -2147483648. Outputs: radius = -2147483648; helper log = [(285,2)].
    - (f) Inputs: raw search parameter = 0; default Skills count = 285. Trace: fixed identity 285 equals the exclusive upper bound. Outputs: radius = 0; helper log = [(285,4)]; no default-record field read.
    - (g) Inputs: raw search parameter = 40; actor dynamic path absent. Trace: both guarded position reads supply 0; nonzero radius bypasses helper. Outputs: radius = 40; center = (0,0); helper log empty.
    - (h) Inputs: raw search parameter = 0; default Skills count = 357; the prepared default record is absent in a synthetic malformed-record fixture. Trace: helper (285,4) rejects the absent record before its operands are read. Outputs: radius = 0; helper log = [(285,4)]; no default-record field read. This does not claim an absent slot is reachable in GDA's normal prepared collection.
  - Outputs: exactly the case's radius and helper log; center = (100,200) in (a)-(f) and (h), with (g)'s independent center as stated; all unit and seed state unchanged. These are synthetic prepared records, not original profile values.
  - Discriminates: treating all nonpositive overrides as defaults fails (d); ignoring zero fallback fails (b); using invocation identity 999 fails the helper identity; saturating arithmetic fails (e); inclusive record bound fails (f); fractional coordinate rounding fails (a)-(f), and invoking a fatal null-path getter instead of the guarded zero path fails (g).
  - runtime_observed: no

- **BMS-V07** Fixed corpse-query candidate test (BMS-R06; boundary: isolated predicate)
  - Inputs: each case has a fresh readable candidate; defaults are monster kind 1, mode 12, dynamic path present with integer position (103,104), center (100,100), radius 5. Other unit kinds in these fixtures have readable position storage. No room collector or explosion action runs.
  - Seeds: no draw.
  - Draws: none.
  - Cases:
    - (a) Trace: dx=3, dy=4, squared distance=25 and radius squared=25. Outputs: predicate result = 1.
    - (b) Inputs: candidate position = (104,104). Trace: 16+16=32 > 25. Outputs: result = 0.
    - (c) Inputs: monster mode = 1. Trace: distance passes but monster is not in mode 12. Outputs: result = 0.
    - (d) Inputs: kind = player 0; mode = 17. Trace: distance passes but fixed type filter excludes players. Outputs: result = 0.
    - (e) Inputs: monster dynamic path absent; center = (0,0); radius = 0. Trace: absent dynamic path supplies (0,0); 0 <= 0. Outputs: result = 1.
    - (f) Inputs: radius = -5. Trace: low32((-5)*(-5))=25; distance 25 passes. Outputs: result = 1.
    - (g) Inputs: candidate position = (50000,0); center = (0,0); radius = 1. Trace: low32(50000*50000) interpreted signed = -1794967296; -1794967296 <= 1. Outputs: result = 1.
    - (h) Inputs: candidate position = (0,0); center = (-2147483648,0); radius = 0. Trace: low32(0-(-2147483648)) is signed -2147483648; low32(dx*dx)=0; distance and radius squares are both 0. Outputs: result = 1. This is an isolated supplied-center fixture, not a claim that a monster actor's unsigned integer position produces this center.
    - (i) Inputs: candidate position = (100,100); center = (100,100); radius = 50000. Trace: distance square = 0; low32(50000*50000) is signed -1794967296; signed 0 > -1794967296. Outputs: result = 0.
  - Outputs: only the case's predicate result; no unit-state or seed change. Cases (g)/(h) are isolated synthetic predicate cases, not claims that real room collection visits these candidate/center combinations.
  - Discriminates: strict-radius comparison fails (a); distance omission fails (b); accepting live monsters or player corpses fails (c)/(d); fatal-null-path substitution fails (e); negative-radius rejection fails (f); widened or unsigned squared comparison fails (g); saturating the subtraction to 2147483647 gives squared distance 1 and fails (h); widened or saturated radius squaring wrongly accepts (i).
  - runtime_observed: no

- **BMS-V08** Process every collected corpse without restoring the old target (BMS-R07 and BMS-R04 normal-return continuation; boundary: supplied-input stage)
  - Inputs: game G; present monster actor M with readable target path, selected target Z; invocation identity 999 and level 7; stable collection K contains monster references A(kind 1,id 201) then B(kind 1,id 202). The shared explosion operation logs its entry arguments/current selected target, returns 0 for A and 7 for B, and makes no other writes or draws in this fixture. Collection release logs its request and returns normally. Target replacement executes BMS-R07, rather than a supplied no-op. Ordered observation log starts empty.
  - Seeds: no draw.
  - Draws: none.
  - Trace: select A; invoke explosion with fixed identity 285 and level 7; ignore 0; select B; invoke explosion with 285/7; ignore 7; request release; outer normal-return continuation yields 1.
  - Outputs: log = [selectTarget(M,A), explosion(G,M,285,7,target=A), selectTarget(M,B), explosion(G,M,285,7,target=B), release(K)]; final selected target = B with kind 1/id 202; corpse-action result = 1; all seeds unchanged.
  - Independent empty-list case: Inputs: fresh selected target Z; collection K contains no recipients; same normally returning release operation. Trace: skip callbacks, request release, outer continuation yields 1. Outputs: log = [release(K)]; selected target = Z; result = 1; seeds unchanged.
  - Discriminates: stopping after explosion return 0 loses B's entries; forwarding 7 fails final result 1; using invocation identity 999 fails both action entries; restoring Z fails final target B; skipping empty-list release fails its log.
  - runtime_observed: no

- **BMS-V09** Shared-action target changes are neither restored nor allowed to suppress the next selection (BMS-R07 and BMS-R04 continuation; boundary: supplied-input stage)
  - Inputs: game G; monster actor M with readable target path and initial selected target Z; invocation identity 999, level 7; stable collection K = [A(kind 1,id 201), B(kind 1,id 202)]. At the first shared-action entry the fixture logs the incoming target, changes the selected target to T(kind 1,id 301), then returns 0 normally. At the second entry it logs the target, changes it to U(kind 1,id 302), then returns 7 normally. These are declared supplied-action effects, not a claim that the original shared action always makes those changes. Release logs and returns normally; no supplied operation draws.
  - Seeds: no draw.
  - Draws: none.
  - Trace: select A; shared action sees A and leaves T; next callback replaces T with B; second shared action sees B and leaves U; release K; outer continuation returns 1 without a target restoration.
  - Outputs: log = [selectTarget(M,A), explosion(G,M,285,7,target=A), selectTarget(M,B), explosion(G,M,285,7,target=B), release(K)]; final selected target = U with kind 1/id 302; result = 1; all seeds unchanged.
  - Discriminates: preserving T instead of selecting B fails the second action's entry target; resetting target to B after the action fails final U; restoring the original Z also fails final U; stopping on first return 0 loses the second call.
  - runtime_observed: no

### 8.2 Observation vectors

**BMS-O01** Observe the selected action and its final return/effects on an
owned Expansion 1.13c game after establishing the missing bindings. Expected
values for gated effects are deliberately unspecified; the
[procedure](../research/spec-evidence/skill-monster-baal.md#observation-procedure)
defines what evidence must be retained. This is a pending investigation, not a
matching observation.

## 9. Acceptance and dependencies

BMS-V01 passing establishes only BMS-R01 for its named negative-index case.
BMS-V02 through BMS-V05 additionally establish only BMS-R04, with the supplied inputs and operation results stated there.
BMS-V06 through BMS-V09 establish the supplied-record radius fallback,
fixed candidate predicate and supplied-collection target/callback stages of
BMS-R05 through BMS-R07. They do not establish real room/candidate production,
collection storage, or any damage/stat/event result produced by the shared
corpse-explosion action.
Full reproduction requires resolving BMS-Q03, BMS-Q04, BMS-Q05 and BMS-Q01, specifying executable vectors for
BMS-R02, and retaining attributable observations under BMS-Q02. A partial result
does not establish parity for these actions.

**Depends on:** [skill dispatch](skill-cast-start-and-do-dispatch.md) SCD-R06, SCD-R07 and SCD-R08 for selecting
the action from the server selector and invoking its handler; successful effect
integration remains BMS-Q01. No damage, missile or monster-creation outcome is
credited through an unnamed dependency.

**Depended on by:** no established consumer.

[Game-data records](core-game-data-table-records.md) GDA-R11a supplies the Param5/Param6 record mapping used by BMS-R04.
GDA-R11a also supplies the Param1/Param2 default-radius mapping.
[Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md)
DSC-R01a through DSC-R14b defines the shared explosion-action entry invoked by
BMS-R07, with actor M, fixed skill identity 285 and the invocation level. Its
corpse-validity, spatial, state, damage and event producer gates remain in force;
the shared entry does not import trap AI or trap-specific caller conditions.

## 10. Related specs

[Skill dispatch](skill-cast-start-and-do-dispatch.md) supplies the action context.
[Numeric foundations](core-numeric-and-rng-foundations.md) owns integer and seed
operations when their consumption is established.
[Shared corpse-explosion action](skill-assassin-death-sentry-corpse-damage.md)
owns the action invoked for each corpse; its selection/AI contract is not reused.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c D2Game, identified by
[source manifest](../research/source-manifest.md), was examined through static
instruction listings on 2026-10-29. The
[companion](../research/spec-evidence/skill-monster-baal.md) identifies the dispatch slots,
read boundary and unresolved callees. Named selector text cells have the companion record; no gameplay case has been confirmed.
