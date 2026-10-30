# Monster maggots and burrowing

**Description**

Maggots use separate actions for eggs, laying eggs, going underground and coming back up. Other burrowing monsters have their own submerge and emerge actions. The connections between these actions are identified here, while their complete placement and interaction effects remain gated.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [numeric and RNG foundations](core-numeric-and-rng-foundations.md), [record layouts and ids](core-record-layouts-and-ids.md); names: [game-data fields](core-game-data-fields.md), [game-data table records](core-game-data-table-records.md)
Ids: MGB-R, MGB-Q, MGB-V, MGB-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-maggots-and-burrowing.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-maggots-and-burrowing.md)

## 1. Scope

The server callbacks selected by MaggotEgg, MagottUp, MagottDown, MagottLay, Submerge and Emerge in the original Expansion profile. This includes egg placement requests and the burrowing transitions; it excludes monster AI choosing when to use them, generic monster creation, collision machinery and animation production. The [skill dispatch contract](skill-cast-start-and-do-dispatch.md) owns entering start/action callbacks. Full successful gameplay behavior remains gated by MGB-Q01–MGB-Q03.

## 2. Trigger and ownership

D2Game owns these callbacks in the original `SKILLS\SkillMonst` region. A selected skill's `Skills.srvstfunc` enters its start callback; `Skills.srvdofunc` enters its action callback. The original-profile text associations are MGB-R01. The start/action dispatcher contract, including its own unresolved AI/frame integration, is SCD-R03–SCD-R09c. This spec makes no independent claim about an AI think cadence or the animation frame on which an action is invoked. Client messages and rendering for successful transitions remain MGB-Q03.

## 3. Inputs and data

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| action identity | supplied skill id | i32 | Index checked against loaded Skills record count |
| loaded skill count | Skills loader record count | i32 | Exclusive upper bound |
| action kind | selected callback | egg hatch, descend, lay | Selects MGB-R02's applicable guard |
| caster | supplied unit or absent | logical unit | Start callbacks' nullable input; action paths have additional unresolved preconditions |
| skill level | supplied level | i32 | Forwarded to calculation/selection helpers; complete consumption is gated |
| target and placement results | selected target, selected monster class, candidate position, creator result | logical values | Successful egg paths; MGB-Q01/MGB-Q02 |
| transition state | caster interaction, collision, animation and life state | logical values | Wider semantic interpretation is MGB-Q03 |
| interaction conditions A/B/C | caster's three engine-defined interaction conditions, also named in MTP-R12 | booleans | Set/clear assignments are verified; their full selection/combat roles remain MGB-Q03 |
| hatch schedule | desired count, class, initialMode, placement results | i32 count/class/mode; ordered results | Supplied entry to MGB-R04; requests log class, caster, band, mode and option |
| created-unit experience condition | `skipKillExperience` of PVP-R09a | boolean | Set by a successful hatch/lay result; does not define unrelated reward routes |
| dynamic position and room | caster's current path position/room | integer coordinates and logical room | Descent footprint origin; successful emergence reads the resulting position/room |
| rooms and spatial cells | current room, ordered adjacent rooms, rectangles, optional collision grids | signed i32 origin/size; logical per-cell conditions | Valid rectangles/grid coverage; room ends are origin+size, with no coordinate overflow in the verified footprint boundary |
| burrow-cell condition | one fixed original-engine spatial-cell condition | boolean per cell | Neutral name for the condition changed by the burrow callbacks; wider movement/physics meaning is MGB-Q03 |
| relocation result/state | supplied relocation request result and resulting current point/room | i32 result, logical updated path state | MTP-R10 owns the local relocation path; its actual search/collision integration retains MTP-Q02 |
| descent life environment | Skills.`calc1` compiled reference/pool/context; caster stat list/arrays/owner/definitions | SCX inputs and RLI section 3.7 values | Actual expression and current/max-life reads for MGB-R09 |

| Named mapping | Origin | Binding | Evidence |
|---|---|---|---|
| The six Skills row/selector associations in MGB-R01 | table data | Data selects engine-defined handlers | Companion table record MGB-D01 |
| Bounds and callback return values in MGB-R02/MGB-R03 | engine code | Fixed by original code | Companion MGB-F01–MGB-F05 |
| Interaction trio assignments, cross footprint and created-unit experience condition | engine code | Fixed by original callbacks; neutral condition mapping in companion | MGB-F07–MGB-F10; PVP-R09a consumes the experience condition |

## 4. Rules (execution order)

**MGB-R01** **Original text selectors** [verified in 1.13c data]. These are individual cells of effective `data/global/excel/skills.txt`, selected by the exact `skill` row identity: MaggotEgg.`srvstfunc` = 43 and `.srvdofunc` = 84; MagottUp.`srvstfunc` = 44; MagottDown.`srvstfunc` = 45 and `.srvdofunc` = 86; MagottLay.`srvdofunc` = 87; Submerge.`srvstfunc` = 51 and `.srvdofunc` = 94; Emerge.`srvstfunc` = 52. MGB-D01 records raw cells. These text cells do not establish loaded compiled values or AI eligibility.

**MGB-R02** **Invalid action identity** [verified in 1.13c code (static)]. At entry to the egg-hatch, maggot-descend or egg-laying action, if the signed skill id is negative or at least the loaded Skills count, return signed 0 before invoking a helper, drawing RNG or mutating a gameplay value. This is a callback-entry guard, independent of whether the outer dispatcher would admit that request. Loaded table storage is assumed valid for the supplied count; malformed storage is outside this verified branch.

**MGB-R03** **Submerge/emerge start interaction assignments** [verified in 1.13c code (static)]. Submerge start clears caster interaction conditions A/B/C when the caster is present. Emerge start sets all three conditions when the caster is present. Other conditions are unchanged. Both return signed 1, including for an absent caster, with no helper call or RNG draw. There is no Skills lookup, skill-level gate or unit-type test inside these two starts. These are actual condition assignments; their full observable selection/combat/physics roles remain MGB-Q03 rather than inferred from the condition names.

**MGB-R04** **Hatch attempt schedule and newborn marking** [verified in 1.13c code (static)]. This stage starts after a nonnegative monster class and count calculation are supplied, and stops before the final caster operation. Let `desired` be the signed i32 count, `initialMode` the supplied i32 creation mode, `successes = 0`. If `desired <= 0`, issue no placement requests. Otherwise request placement with band -1 and creation mode `initialMode`. A nonempty result first sets that created unit's `skipKillExperience` condition true, preserving its other conditions, then increments successes. Then set `band = 1`, `tries = 0`. While `successes < desired`, `band < 5` and `tries < i32(low32(2*desired))`, request placement with the current band and creation mode 8. A nonempty result receives the same experience-condition assignment before successes increments; an empty result increments band. Increment tries after each repeated request. Comparisons are signed i32 and additions wrap to i32. Each request carries the same supplied class and caster and a final option value 0. The initial request is excluded from tries. These numeric arguments describe the request, without claiming a physical search geometry. The supplied creator logs each request and returns a supplied created-unit identity or empty; it has no other effects in this stage. No RNG is drawn locally. Real creation and the final caster operation remain MGB-Q01; this rule does not assert that the real creator has no draws or that the experience condition governs every reward route.

**MGB-R05** **Egg laying** [unresolved]. The egg-laying callback uses a selected target and a direction-derived placement request to a monster creator. Target eligibility, direction conversion, failed-placement effects and created-unit semantics are gated by MGB-Q02. No fixed species, guaranteed egg count or complete RNG count is asserted.

**MGB-R06** **Remaining burrowing integration** [unresolved]. Wider interaction-condition meanings, animation progression/reset and complete relocation/physics effects remain MGB-Q03. The actual start assignments, direct footprint changes and descent life-write request are separately established by MGB-R03/MGB-R07–MGB-R09. Null-caster behavior outside MGB-R03 is not generalized.

**MGB-R07** **Maggot descent footprint clearing** [verified in 1.13c code (static)]. On the MagottDown start's reached footprint stage for a present monster with a valid dynamic path, obtain the caster's current integer point (x,y) and current room. Clear the burrow-cell condition at (x-1,y), (x,y), (x+1,y), (x,y-1), (x,y+1). Each cell is resolved independently: use the current room when it contains the point; otherwise use the first nonempty containing room in its ordered adjacent-room list. Containment uses signed, lower-inclusive/upper-exclusive rectangle bounds. If no containing room, collision descriptor or grid exists for one point, skip that point and continue the others. All unrelated cell conditions remain unchanged; diagonal cells are untouched. There is no RNG or stat operation in this footprint stage. The start returns signed 1 after the footprint operation, without testing a map-operation result. Its separate interaction-trio clear uses the same assignment as MGB-R03; the effect of that trio on later targeting remains MGB-Q03. This rule establishes the actual cell changes at the stated valid-grid boundary, not that this condition alone determines traversability.

**MGB-R08** **Maggot emergence footprint setting** [verified in 1.13c code (static)]. For a present monster with a valid dynamic path, MagottUp start requests the shared relocation operation at the caster's current point/room with nearby search enabled and ordinary relocation marker, consuming MTP-R10 and its existing integration gate. If it returns 0, return 0 and make no subsequent interaction-trio or footprint change. A nonzero result sets interaction A/B/C true, obtains the resulting current point/room, and sets the burrow-cell condition true at the same five-cell cross defined by MGB-R07, with the same independent room/grid guards. All unrelated cell conditions and diagonals remain unchanged. Return signed 1. The selected footprint helper also sets the center condition again; this is idempotent and imposes no additional stored result. Local code draws nothing; relocation effects/draws remain the relocation owner's boundary. The rule's cell assertions begin with the spatial state after relocation returns; they do not claim the relocation operation itself cannot change cells.

**MGB-R09** **Maggot descent life request** [verified in 1.13c code (static)]. At the valid MagottDown action's life stage, evaluate the loaded skill's `calc1` using SCX-R11 context (caster, skill id, callback level). A result <=0 skips all current/max-life reads and the replacing-write request, then returns signed 1. A positive result P reads caster stat 6/layer 0 as current life C, then stat 7/layer 0 as maximum M through RLI-R02a/RLI-R02b. Calculate `increase = mulDiv(P,C,100)` using NRF-R05 in that operand order; set `candidate = i32(low32(C+increase))`, then use signed `min(candidate,M)` for the replacing stat 6/layer 0 request. There is no lower floor. If the writer returns normally, return 1 without testing its result. No local RNG is drawn; the selected expression owns any dependency draws. This rule is the reached life stage after the callback's latch/animation preparation; those preceding animation effects remain MGB-Q03. Actual stored replacement/propagation/notification consumes STL-R06c/STL-Q07 and RLI-Q03, rather than a supplied no-op claim about the real writer.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MGB-R01 | Text selector cells | verified in 1.13c data | none; AI/frame entry is MGB-Q04 | MGB-D01 in companion |
| MGB-R02 | Invalid action id | verified in 1.13c code (static) | none locally | MGB-V01 |
| MGB-R03 | Submerge/emerge interaction assignments | verified in 1.13c code (static) | MGB-Q03: wider condition meaning only | MGB-V02, MGB-V04 |
| MGB-R04 | Hatch request schedule and experience marking | verified in 1.13c code (static) | MGB-Q01: real creation/final effects only; local schedule/condition write available | MGB-V03 |
| MGB-R05 | Lay integration | unresolved | MGB-Q02: successful creation and placement | none |
| MGB-R06 | Remaining transition integration | unresolved | MGB-Q03: wider interaction/animation/physics and relocation effects only | none |
| MGB-R07 | Actual descent footprint clear | verified in 1.13c code (static) | MGB-Q03: broader movement/physics meaning only | MGB-V05 |
| MGB-R08 | Emergence success gate and footprint set | verified in 1.13c code (static) | MTP-Q02: actual relocation; wider physics meaning MGB-Q03 | MGB-V06 |
| MGB-R09 | Descent expression/read/write-request stage | verified in 1.13c code (static) | MGB-Q03: preceding animation effects; STL-Q07/RLI-Q03: actual replacing effects | MGB-V07 |

## 5. Outputs and side effects

| Output or effect | Rule |
|---|---|
| Callback result 0 with no gameplay effects on invalid-id branch | MGB-R02 |
| Callback result 1 with no gameplay effects on absent-unit start branch | MGB-R03 |
| Interaction A/B/C cleared by Submerge/MagottDown start or set by Emerge/accepted MagottUp | MGB-R03/MGB-R07/MGB-R08 |
| Ordered placement requests and their count at the supplied-results stage | MGB-R04 |
| Actual skipKillExperience assignment on each successful hatch result | MGB-R04 |
| Created monsters and final caster effects, gated | MGB-R04/MGB-R05 |
| Actual five-cell burrow-condition clear/set with unrelated cells/conditions unchanged | MGB-R07/MGB-R08 |
| Descent current-life percentage replacing-write request and callback result | MGB-R09 |
| Wider interaction/animation, deep relocation/physics and actual replacing storage effects, gated | MGB-R06/MGB-R08/MGB-R09 |

No persistence or named client-message contract is established by the verified subset.

## 6. Branches and edge cases

MGB-R02 uses signed comparisons: negative ids fail, and equality with the record count fails. MGB-R03 applies only to Submerge/Emerge starts, even if a different callback also receives an absent caster. The direct interaction and footprint assignments read no difficulty and draw no RNG. MGB-R09 uses current life as its percentage basis, clamps only to maximum and can issue a negative wrapped raw request; it does not share MRH's maximum-life percentage or raw lower floor. Complete obstruction/relocation, animation progression and client effects retain the stated dependency gates. Other game builds and Classic execution are outside this Expansion contract.

## 7. Open questions

| Question | Status | Evidence that settles it |
|---|---|---|
| **MGB-Q01** | unresolved | Complete class-selection/actual creation and final-caster effects, including placement/RNG; the local newborn experience-condition write is available under MGB-R04 |
| **MGB-Q02** | unresolved | Target lookup, direction conversion and creator chain through successful and failed laying |
| **MGB-Q03** | integration unresolved | Wider interaction/burrow-cell meanings, animation preparation/progression, complete relocation/physics and replacing storage effects; direct assignments/grid changes and descent life request remain available under MGB-R03/MGB-R07–MGB-R09 |
| **MGB-Q04** | unresolved | Establish actual AI/frame entry; narrow loaded selector agreement is recorded in MGB-D01 but does not prove execution |
| **MGB-Q05** | observation pending | Matching naturally reachable captures of the MGB-R04 hatch schedule and successful transitions; invalid-id/absent-unit entry may be excluded by earlier dispatcher checks and remains statically specified |

Attempts and exact blocked assertions are in the [gap file](../plans/gaps/skill-monster-maggots-and-burrowing.md).

## 8. Test vectors

### 8.1 Executable vectors

MGB-V01/MGB-V02 enter callbacks directly; other vectors use their stated stage/integration boundaries. No surrounding dispatcher behavior is supplied or asserted. Cases are independent and have fresh empty logs. RNG state is arbitrary and unchanged when Seeds says `no draw`.

- **MGB-V01** Invalid identities (MGB-R02; boundary: isolated rule)
  - Inputs: loaded Skills count = 357; caster = absent; level = 1; empty helper/write logs.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: signed comparison rejects the id before any use of caster.
  - Cases: for each action kind egg hatch, descend and lay, run (a) skill id = -1; (b) skill id = 357.
  - Outputs: in every independent case result = 0; helper log = []; gameplay-write log = []; all seeds unchanged.
  - Discriminates: rejects an unsigned-only lower-bound test, an inclusive upper bound, or caster/helper access before validation.
  - runtime_observed: no

- **MGB-V02** No caster at transition start (MGB-R03; boundary: isolated rule)
  - Inputs: caster = absent; skill id = -1; level = -7; empty helper/write logs; cases (a) Submerge start, (b) Emerge start.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: absent-caster branch skips the unit write and immediately returns 1; id and level are unused.
  - Outputs: in each independent case result = 1; helper log = []; gameplay-write log = []; all seeds unchanged.
  - Discriminates: rejects treating an absent caster as callback failure or imposing an id/level gate on these starts.
  - runtime_observed: no

- **MGB-V03** Hatch scheduling and stopping limits (MGB-R04; boundary: supplied-input stage)
  - Inputs: class = 1 (synthetic supplied class, no named species claim); caster = egg A; initialMode = 7; request log empty. Every supplied placement result is empty unless a case states otherwise; the supplied creator has no RNG or stored effects. Created-unit records B/C/D initially have skipKillExperience false and all unrelated conditions stable.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: initial attempt precedes the repeated-attempt budget; only failures advance the band.
  - Cases:
    - (a) desired = 3; supplied results in call order = [empty, unit B, empty, unit C, unit D]. Outputs: request (band, mode) log = [(-1,7),(1,8),(1,8),(2,8),(2,8)]; successes = 3; tries = 4; band = 2. Every request has class 1, caster egg A, option 0.
    - (b) desired = 1; all results empty. Outputs: request log = [(-1,7),(1,8),(2,8)]; successes = 0; tries = 2; band = 3. Same class/caster/option on each request.
    - (c) desired = 10; all results empty. Outputs: request log = [(-1,7),(1,8),(2,8),(3,8),(4,8)]; successes = 0; tries = 4; band = 5. Same class/caster/option on each request.
    - (d) desired = 0. Outputs: request log = []; successes = 0; no repeated loop entered.
    - (e) desired = 1; supplied result = unit B. Outputs: request log = [(-1,7)]; successes = 1; tries = 0; band = 1. The request has class 1, caster egg A, option 0; no repeated placement is requested.
    - (f) desired = 1073741824; all results empty. Outputs: request log = [(-1,7)]; successes = 0; tries = 0; band = 1. The request has class 1, caster egg A, option 0. The signed repeated-attempt bound is -2147483648, so no repeated placement is requested.
  - Outputs: the case-specific ordered requests and counters above; every actually returned B/C/D has skipKillExperience true, every unreturned record retains its initial false value, and unrelated conditions remain unchanged. No actual creation or final-caster effect is asserted; all seeds unchanged.
  - Discriminates: (a) rejects advancing the band after success; (b) rejects counting the initial attempt against the repeated budget or using a fixed four retries; (c) rejects a fifth repeated band; (d) rejects an unconditional first placement; (e) rejects an unconditional first repeat after immediate success; (f) rejects widened or unsigned repeated-attempt bounds.
  - runtime_observed: no

- **MGB-V04** Nonempty Submerge/Emerge starts (MGB-R03; boundary: isolated rule)
  - Inputs: present caster C, skill id -1, level -7; unrelated condition K true; empty helper/stat logs. Independent cases (a) Submerge with A=true/B=false/C=true; (b) Emerge with A=false/B=false/C=false.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: direct clear or set of the three interaction conditions; no id/level or type gate.
  - Outputs: (a) all three conditions false, K true, result 1; (b) all three conditions true, K true, result 1. Each case has empty helper/stat logs and unchanged seeds.
  - Discriminates: rejects flipping only one condition, toggling instead of assigning, or applying id/level validation to these starts.
  - runtime_observed: no

- **MGB-V05** Actual descent footprint across rooms (MGB-R07; boundary: integration of caller footprint and cell helpers)
  - Inputs: present monster with valid dynamic path at integer point (100,201), current room A; A rectangle origin (100,200), size (2,3), adjacent list [B]; B rectangle origin (99,200), size (1,3). A grid has condition true at center (100,201), right (101,201), up (100,200), down (100,202) and diagonal (101,200); B's left cell (99,201) is true when its grid is present. Each existing cell also has unrelated condition K=true. Room rectangles/grid coverage are valid and nonoverlapping. The separate interaction-trio clear is outside this footprint snapshot.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: left resolves through B; center/right/up/down through A; each cell clears independently.
  - Cases:
    - (a) Both grids present. Outputs: burrow condition false at all five cross cells; diagonal (101,200) remains true; K true in every cell; start result 1.
    - (b) B collision grid absent. Outputs: all four named A cross cells false; A diagonal true; K true in every existing cell; no B cell is created or written; start result 1.
    - (c) Current room absent, with the named A/B grids otherwise present as independent unused data. Outputs: all six named cells retain true burrow condition and K=true; start result 1.
    - (d) Override point to (101,201). A keeps rectangle origin (100,200), size (2,3), but adjacent list is [B,C]. B and C each have rectangle origin (102,200), size (1,3), separate present grids and burrow condition true at (102,201). A's existing six named cells start true as above, and its additional cell (101,202) also starts true; K=true in every provided cell. The adjacent rectangles intentionally overlap in this synthetic case to exercise first-match selection; this is not a naturally generated-room claim. Outputs: A cells (100,201),(101,201),(101,200),(101,202) false; A cells (100,200),(100,202) true; B cell (102,201) false; C cell (102,201) true; K=true everywhere; start result 1.
  - Outputs: only each case's exact cell states/result above; no stat or seed change.
  - Discriminates: (a) rejects a square/diagonal footprint, reversed coordinates or ignoring the adjacent room; (b) rejects aborting the entire cross on one missing grid; (c) rejects inventing a room or grid when the current room is absent; (d) rejects treating A's upper edge x=102 as included or choosing the second overlapping adjacent room instead of the first.
  - runtime_observed: no

- **MGB-V06** Emergence uses the post-relocation point (MGB-R08; boundary: supplied relocation result followed by actual cell helpers)
  - Inputs: present monster with valid path, initial point (101,201), room A; initial interaction A=false/B=true/C=false and unrelated K=true. Supplied relocation logs request (current point (101,201), room A, nearby search enabled, ordinary marker), returns the listed result and supplies resulting point (100,201)/room D. A has rectangle origin (100,200), size (2,3), no collision descriptor and an empty adjacent list. The post-operation snapshot has D rectangle origin (100,200), size (2,3), adjacent list [B]; B rectangle origin (99,200), size (1,3); D/B grids present. D's cells (100,201),(101,201),(100,200),(100,202),(101,200) and B's cell (99,201) all have burrow condition false and unrelated cell K=true. This snapshot does not assert the real relocation had no earlier cell effects.
  - Seeds: `no draw` locally; supplied relocation makes no draw in these cases.
  - Draws: none.
  - Trace: inspect result; only nonzero sets interaction trio and the five-cell cross at the resulting point.
  - Cases:
    - (a) Relocation result 1. Outputs: interaction A/B/C all true, unit K true; B cell (99,201) and D cells (100,201),(101,201),(100,200),(100,202) all true; D diagonal (101,200) false; cell K true everywhere; A remains without a collision descriptor; callback result 1; resulting point (100,201)/room D.
    - (b) Relocation result 0. Outputs: interaction false/true/false and unit K true; all six D/B snapshot burrow conditions false and cell K true; A remains without a collision descriptor; callback result 0; supplied resulting point (100,201)/room D retained without an additional caller footprint change.
  - Outputs: exact case-specific states/result above and one relocation request with the stated controls; stats/seeds unchanged in the supplied-operation boundary.
  - Discriminates: rejects using the pre-relocation point or room for the footprint: using A yields no cell changes instead of the required D/B cross, and using point (101,201) sets D's asserted-false diagonal while missing B's left cell. It also rejects ignoring relocation failure or assuming a failed operation must have restored its earlier point/room.
  - runtime_observed: no

- **MGB-V07** Descent uses current-life percentage and no lower floor (MGB-R09; boundary: actual expression/stat reads through replacing-write request)
  - Inputs: caster M monster/id 1, valid loaded skill id 162/count 357, callback level 1; `calc1` reference 0 and skill pool size 3. The reached life stage follows animation preparation; those preceding effects are not asserted here. M has usesResolvedStats true, associated owner M, own entries [(6,0,1000),(7,0,10000)] and resolved entries [(6,0,2560),(7,0,25600)], both active counts 2. Stat-definition count 8 with present definitions for 6/7, fMin=true/MinAccr=1/ValShift=8. Supplied replacing writer logs request, returns 0 normally and makes no stored effect/draw. Each case has an empty operation log; valid synthetic compiled bytes are not a claim about the original compiler.
  - Seeds: `no draw`; each supplied compiled expression has no random token.
  - Draws: none.
  - Trace: execute the expression, test signed result, then read current before maximum and apply ordered percentage/add/max only when positive.
  - Cases:
    - (a) Pool bytes `07 0a 00`. Outputs: ordered log [eval(`calc1`,M,162,1)=10, read(M,6,0)=2560, read(M,7,0)=25600, replace(M,6,0,2816)=0]; increase 256, candidate 2816; callback result 1.
    - (b) Pool bytes `07 00 00`. Outputs: log [eval(`calc1`,M,162,1)=0]; no stat read/write; callback result 1.
    - (c) Pool bytes `07 ff 00`. Outputs: log [eval(`calc1`,M,162,1)=-1]; no stat read/write; callback result 1.
    - (d) Pool bytes `07 64 00`; resolved entries instead [(6,0,2147483600),(7,0,2147483647)]. Outputs: ordered log [eval(`calc1`,M,162,1)=100, read(M,6,0)=2147483600, read(M,7,0)=2147483647, replace(M,6,0,-96)=0]; increase 2147483600 via NRF-R05, raw sum 4294967200 wraps to -96; callback result 1. This is an isolated supplied-array edge case, not a naturally prepared monster fixture.
    - (e) Pool bytes `07 0a 00`; resolved entries instead [(6,0,2560),(7,0,2600)]. Outputs: ordered log [eval(`calc1`,M,162,1)=10, read(M,6,0)=2560, read(M,7,0)=2600, replace(M,6,0,2600)=0]; increase 256, candidate 2816, signed maximum clamp gives 2600; callback result 1.
  - Outputs: exact case-specific logs/result above; arrays/seeds unchanged at the supplied write boundary; no actual propagation/notification asserted.
  - Discriminates: (a) rejects using maximum life as the percentage basis or reading own instead of resolved entries; (b)/(c) reject stat reads/writes on a nonpositive calculation; (d) rejects a lower floor, saturating/widened addition or treating the writer's 0 return as callback failure; (e) rejects omitting the maximum clamp or applying the wrong maximum value.
  - runtime_observed: no

### 8.2 Observation vectors

**MGB-O01** Observe MaggotEgg, laying, descent and emergence separately on Normal difficulty, recording each action's class, position, level, initial/final state, callback result and creation count. [Reproduction procedure](../research/spec-evidence/skill-monster-maggots-and-burrowing.md#observation-procedure) specifies the evidence needed before any promotion. Expected MGB-V01/MGB-V02 callback results apply only when those exact entry cases are reproduced; successful gameplay outcomes remain gated.

## 9. Acceptance and dependencies

Partial acceptance consists of MGB-R01's named text cells, MGB-V01–MGB-V04 for guards/interaction assignments/hatch schedule and newborn marking, MGB-V05/MGB-V06 for actual footprint changes at their stated relocation/grid boundaries, and MGB-V07 for actual descent expression/stat reads through replacing-write request. Full reproduction additionally needs actual creation/final-caster effects of MGB-Q01, laying MGB-Q02, wider animation/physics/relocation/storage integration MGB-Q03 and actual AI/frame entry MGB-Q04. Passing the partial subset establishes no complete egg or burrowing behavior.

**Depends on:** NRF-R01/NRF-R02 for unchanged seeds/no draw, NRF-R05 for ordered percentage arithmetic, NRF-R08a for raw life and NRF-R10b for prepared dynamic integer positions; SCD-R03–SCD-R09c for shared start/action entry with its gates; RLI-R02a/RLI-R02b for actual current/max existing-array reads; [skill calculations](skill-calc-expressions.md) SCX-R01–SCX-R11 for the front end/compiled execution, with requested producer gates retained; [monster teleport](skill-monster-teleport.md) MTP-R10 for the relocation request/control path with MTP-Q02; [kill credit](combat-pvp-damage-and-kill-credit.md) PVP-R09a for the skipKillExperience consumer; [timed-state lifecycle](state-timed-state-lifecycle.md) STL-R06c/STL-Q07 and RLI-Q03 for replacing-write propagation. Generic creation and broader physics required for full behavior are not established by a label or the direct cell writes.

**Depended on by:** no known consumed plan, code or test contract.

## 10. Related specs

- [Skill cast start and do dispatch](skill-cast-start-and-do-dispatch.md): common callback entry and frame integration.
- [Monster scaling](monster-scaling-by-players-and-difficulty.md): shared downstream monster creation/scaling boundary, with its own evidence limits.

## Appendix A. Provenance

Windows x86 Expansion 1.13c; original `LOD113C-D2GAME` and `LOD113C-D2COMMON` in the [source manifest](../research/source-manifest.md). Dispatch slots, selected cell-helper branches and applicable instructions were read on 2026-10-29, using byte-equivalent reconstruction where available and original-image instructions/tables elsewhere. Compatible SCX/RLI/MTP/PVP contracts retain their recorded methods and limits. Effective Skills text cells were read through the owned original-profile game's archive loader and narrow loaded selectors agreed on that date. No gameplay vector is runtime-confirmed. [Companion records](../research/spec-evidence/skill-monster-maggots-and-burrowing.md) preserve exact boundaries and limitations.
