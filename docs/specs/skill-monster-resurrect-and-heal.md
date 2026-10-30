# Monster resurrection, healing and blood sucking

**Description**

Some monsters heal another creature, bring a corpse back into play, or regain life through a blood-sucking attack. These use different server actions. The healing action chooses a percentage of the target's maximum life, with a cap at maximum life; the complete resurrection and blood-sucking effects remain gated.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [numeric and RNG foundations](core-numeric-and-rng-foundations.md), [record layouts and ids](core-record-layouts-and-ids.md); names: [game-data fields](core-game-data-fields.md), [game-data table records](core-game-data-table-records.md)
Ids: MRH-R, MRH-Q, MRH-V, MRH-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-resurrect-and-heal.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-resurrect-and-heal.md)

## 1. Scope

The server callbacks for ZakarumHeal/Bestow, Resurrect/Resurrect2, Self-resurrect, Mosquito and Suck Blood in the Expansion profile. This owns their local skill decisions and requested effects, including the valid healing callback's target normalization, compiled calculation requests, existing-stat reads and replacing-write request. AI choosing a skill, generic stat storage, combat execution and corpse lifecycle are dependencies. Successful resurrection and both blood-sucking action families remain MRH-Q02/MRH-Q03 rather than being inferred from callback names.

## 2. Trigger and ownership

D2Game owns these monster skill handlers. The [shared skill dispatcher](skill-cast-start-and-do-dispatch.md) selects the start/action callbacks through `Skills.srvstfunc`/`srvdofunc`; MRH-R01 gives the observed text associations. The common dispatch boundary is SCD-R03–SCD-R09c, retaining that spec's AI/frame gates. No fixed AI think cadence or client-message identity is established here. The supplied-input arithmetic begins after healing target/preparation and both skill calculations have completed.

## 3. Inputs and data

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| skill id / count | callback id / loaded Skills count | signed i32 | Guard input for MRH-R02 |
| caster, target | selected units | logical units or absent | Healing selection is normalized by MRH-R06 |
| level | callback skill level | signed i32 | Calculation input |
| lower / upper calculation results | Skills.`calc1` / Skills.`calc2` evaluation results | signed i32 | Supplied percentage endpoints, in evaluation order upper then lower |
| caster seed | caster RNG state | two u32 words | One bounded draw when normalized upper > lower |
| target life | stat 6 `hitpoints`, layer 0, read through RLI-R02a/RLI-R02b | signed i32, 8 fractional bits | Current raw life; the selected array can be resolved stats |
| target maximum life | stat 7 `maxhp`, layer 0, through RLI-R02a/RLI-R02b | signed i32, 8 fractional bits | Percentage basis and final upper clamp |
| life-write operation | supplied log of stat writes | target, stat id, layer, raw value | Logs request only; actual storage is a dependency |
| resurrection / blood inputs | corpse, collision, owner, mode and combat results | logical values | Successful integration remains MRH-Q02/MRH-Q03 |
| selected-target state | caster's path-associated selected unit and cached target type/id | logical unit or absent; signed i32 type/id | Actual selection normalized by MRH-R06; clearing selection leaves cached type/id unchanged |
| game registry | registered units for each supported type, indexed by unit id | logical identity mapping | Canonical lookup for the cached target identity; valid registry identities are unique |
| target item mode | registered item's current mode | i32 | Mode ids 1 and 2 clear selection during normalization |
| do-handled latch | caster animation condition of SCD-R07/SCD-R09a | boolean | Set after a distinct nonempty healing target survives normalization |
| calculation environment | loaded Skills record, `calc1`/`calc2` references and skill expression pool | u32 references/count; compiled bytes | Actual SCX-R11 requests with context (caster, skill id, level); compiler/producer gates belong to SCX |
| stat-read environment | target stat list, own/resolved arrays, associated owner and definitions | RLI section 3.7 values | Supplies actual stat reads for MRH-R07, rather than only supplied getter results |

| Mapping / constant | Origin | Binding | Evidence |
|---|---|---|---|
| MRH-R01 text selectors | table data | Data selects engine-defined handlers | MRH-D01 |
| Percentage cap 100, life floor raw 1, stat 6/layer 0 request | engine code | Fixed by original code | MRH-F01/MRH-F02 |
| ZakarumHeal.`calc1` = `15+5*lvl`, .`calc2` = `50` | table data | Data expression evaluated by shared machinery | MRH-D01; not a claim that compiled evaluation ran |
| Target item-mode exclusions 1/2, maximum stat 7/layer 0, evaluation/read order | engine code | Fixed by the original heal callback and preparation helper | MRH-F08–MRH-F10 |

## 4. Rules (execution order)

**MRH-R01** **Original text selectors** [verified in 1.13c data]. In effective `data/global/excel/skills.txt`, exact named cells are ZakarumHeal.`srvdofunc` = 96; Bestow.`srvdofunc` = 96; Resurrect.`srvdofunc` = 97; Resurrect2.`srvdofunc` = 97; Self-resurrect.`srvstfunc` = 61; Mosquito.`srvstfunc` = 55 and `.srvdofunc` = 107; Suck Blood.`srvstfunc` = 60 and `.srvdofunc` = 127. ZakarumHeal.`calc1` is `15+5*lvl` and `.calc2` is `50`. These named text values do not prove compiled fields, AI eligibility or calculation execution; see MRH-D01.

**MRH-R02** **Invalid healing/resurrection action id** [verified in 1.13c code (static)]. At the heal, resurrect or blood-suck action callback entry, reject a signed skill id below 0 or at least the loaded Skills count with signed result 0. This branch precedes helper calls, RNG and gameplay writes. Valid loaded table storage is assumed; other malformed-storage and target cases are not covered by this guard.

**MRH-R03** **Healing percentage and write request** [verified in 1.13c code (static)]. At the supplied-input stage, use the upper calculation result first: `upper = clamp(calc2,0,100)`. Then `lower = clamp(calc1,0,upper)`. All are signed i32. Let `range = upper-lower`; use the caster's bounded draw of NRF-R02 with modulus range, which returns 0 without stepping for range <= 0. Set `percent = lower+draw`. Thus a nonempty interval excludes its upper endpoint; equal endpoints produce that endpoint without a draw.

Read supplied target maximum life and current life. Set `increase = mulDiv(percent,maximumLife,100)` using NRF-R05 in that exact operand order. Set `candidate = i32(low32(currentLife+increase))`, then if candidate <= 1 set candidate = 1. Select `writtenLife = min(candidate,maximumLife)` using a signed comparison. Request target stat 6 `hitpoints`, layer 0, raw writtenLife, then return signed 1. The raw lower floor is 1/256 life, not one whole life; the subsequent maximum clamp can override it. There are no local RNG draws after the percentage draw. Supplied getter/write operations have no additional draws/effects in this isolated stage. MRH-R06/MRH-R07 define actual preparation, calculation requests and existing-stat reads. Actual replacing-write storage/propagation remains MRH-Q01; this gate does not block the supplied-result arithmetic.

**MRH-R04** **Resurrection and self-resurrection** [unresolved]. The resurrection action validates a selected corpse and invokes revival/mode/room operations before optional target-state handling. Self-resurrection has a separate start path and monster-type gate. Complete eligibility, corpse reuse/removal, restored state, life and ownership are gated by MRH-Q02; no assumption of full-life or player-style resurrection follows from the action name.

**MRH-R05** **Blood sucking** [unresolved]. Suck Blood uses a selected target and a combat-result-dependent path that can request life restoration for a recipient selected from the caster/owner relationship. Mosquito uses a distinct start/action pair, target identity bookkeeping, repeated damage-related draws and a caster life request. The exact recipient, damage conditions, life basis, ordering and later combat effects of each family are gated by MRH-Q03. The heal percentage rule MRH-R03 is not substituted for either action.

**MRH-R06** **Healing target normalization and early rejection** [verified in 1.13c code (static)]. This rule starts after MRH-R02 admits a valid loaded skill record. Its normal-entry boundary requires a nonempty game, caster and caster path; when a target is selected, the cached type is one of the supported unit kinds 0–5 and the game registry is valid. If no unit is currently selected, normalization changes nothing. Otherwise resolve the registered unit for the cached target type/id. If that unit is absent or differs from the selected unit identity, clear selection; do not switch selection to the resolved unit. Independently, if the resolved unit is an item in mode 1 or 2, clear selection. Clearing the selected unit leaves the cached type/id unchanged. No RNG is drawn and no stat is read or written during this normalization.

Read selection after normalization. If it is absent or is the caster itself, return signed 0 before setting the do-handled latch or evaluating either calculation. Otherwise retain that selected target and proceed to MRH-R07. Normalization does not add a friendliness, living-target, distance, difficulty or monster-only target test; any such upstream eligibility belongs to its producer. Invalid game/caster/path or malformed registry inputs are outside this verified normal-entry boundary, rather than successful empty-target defaults.

**MRH-R07** **Successful healing callback through the write boundary** [verified in 1.13c code (static)]. After MRH-R06 retains a distinct nonempty target, set the caster's do-handled latch. Evaluate the loaded skill's `calc2` through SCX-R11 in context (caster, skill id, callback level), normalize its upper endpoint as MRH-R03, then evaluate `calc1` through the same front end and context and normalize its lower endpoint. Both calculations are requested, including when the normalized upper endpoint is zero. Each evaluation completes before the next step. SCX owns execution of the requested compiled expression, its parameter/function dependencies and any seed changes; those draws precede MRH-R03's bounded draw, which uses the caster seed as it stands after both evaluations.

Then perform the percentage draw and arithmetic of MRH-R03, reading the retained target's stat 7/layer 0 first and stat 6/layer 0 second through RLI-R02a/RLI-R02b. The maximum-life front end uses exactly stat 7; it does not read a separate cached maximum or infer a maximum from current life. These reads select own/resolved arrays and apply conditional minimum adjustment according to RLI; neither is necessarily an unmodified base value. Request the target stat 6/layer 0 replacing write with the computed raw result, and, if that operation returns normally, return signed 1 without testing its result. Successful callback return therefore does not establish that storage changed. Actual replacing storage, propagation and notifications consume the unresolved STL-R06c/STL-Q07 and RLI-Q03 boundary identified by MRH-Q01. Expression compiler or unresolved parameter-producer effects retain their SCX gates; explicitly supplied valid compiled expressions and existing stat arrays allow this complete callback-to-request path independently of them.

For healing, execution is MRH-R02 → MRH-R06 → MRH-R07, with MRH-R07 consuming MRH-R03's arithmetic. MRH-R04/MRH-R05 describe separate action families.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MRH-R01 | Named text cells | verified in 1.13c data | none for cells; AI/frame entry is MRH-Q04 | MRH-D01 in companion |
| MRH-R02 | Invalid-id exits | verified in 1.13c code (static) | none locally | MRH-V01 |
| MRH-R03 | Supplied-result heal stage | verified in 1.13c code (static) | MRH-Q01: actual replacing storage/propagation only | MRH-V02–MRH-V06 |
| MRH-R04 | Resurrection integration | unresolved | MRH-Q02: successful resurrection effects | none |
| MRH-R05 | Blood-suck integration | unresolved | MRH-Q03: recipient and combat/life effects | none |
| MRH-R06 | Actual healing selection normalization | verified in 1.13c code (static) | none within valid game/caster/path/registry boundary | MRH-V07 |
| MRH-R07 | Callback calculation/read/write-request order | verified in 1.13c code (static) | MRH-Q01: actual replacing effects; SCX's compiler/unresolved producer gates when requested | MRH-V08 |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| Callback result 0; no helper call, RNG or gameplay write on invalid-id branch | MRH-R02 |
| Updated caster seed when range is positive | MRH-R03 |
| Ordered target life-write request (stat 6, layer 0), followed by callback result 1 | MRH-R03 |
| Revival, corpse/mode/state/room effects, gated | MRH-R04 |
| Recipient life and subsequent combat effects, gated | MRH-R05 |
| Selected-target pointer cleared on stale identity or excluded item mode; cached identity unchanged | MRH-R06 |
| Early result 0 after absent/self selection; no calculation or latch change | MRH-R06 |
| Do-handled latch set; ordered calculation evaluations, existing-stat reads and replacing-write request; normal return 1 | MRH-R07 |

The arithmetic and successful callback stages assert the actual replacing-write request, not its persistence or a client notification. Actual messages, state expiration and propagation remain integration evidence gaps.

## 6. Branches and edge cases

The arithmetic uses raw signed life values and signed clamps; a zero or negative supplied maximum can override the raw lower floor. A positive range always draws once, including range 1. Negative or overlarge percentages are normalized before drawing. The upper percentage is exclusive only for a nonzero interval. NRF-R05 may divide first on large operands, losing precision; replacing it with a mathematically exact product/division changes results. Signed 32-bit addition wraps before the lower and upper clamps. The verified healing callback contains no difficulty branch; level/data and actual AI entry producers retain their SCX/MRH-Q04 limits. Corpse type and collision failure remain MRH-Q02.

## 7. Open questions

| Question | Status | Evidence that settles it |
|---|---|---|
| **MRH-Q01** | integration unresolved | Complete shared replacing-write storage, propagation and notification under STL-R06c/STL-Q07 and RLI-Q03; local normalization, selected calculation requests and existing-stat reads are available under MRH-R06/MRH-R07 |
| **MRH-Q02** | unresolved | Complete action/self-start revival and corpse/mode/room/state helper chains, with failure cases |
| **MRH-Q03** | unresolved | Complete blood-suck start, combat-result and recipient selection/life-write paths |
| **MRH-Q04** | integration unresolved | Establish actual AI/frame dispatch; narrow loaded selector agreement is recorded in MRH-D01 but does not prove execution |
| **MRH-Q05** | observation pending | Matching attributed captures of arithmetic/write and successful revival/blood-suck behavior; does not gate supplied-input arithmetic |

See [attempted gaps](../plans/gaps/skill-monster-resurrect-and-heal.md).

## 8. Test vectors

### 8.1 Executable vectors

MRH-V02–MRH-V06 enter MRH-R03 with a valid selected target T and caster C after preparation/calculation. Supplied getter results are stable, the life-write log starts empty, the write operation logs its request without changing other state or drawing RNG, and the stage returns 1 after that request. MRH-V07/MRH-V08 instead exercise the new boundaries stated in their own inputs. Numeric life inputs/outputs are raw 8-bit-fraction values. Each case is independent. A positive draw steps only C's supplied seed; all other seeds remain unchanged.

- **MRH-V01** Invalid id before target access (MRH-R02; boundary: isolated rule)
  - Inputs: loaded count = 357; caster/target absent; level = 1; helper/write logs empty. For each heal, resurrect and blood-suck action run (a) skill id -1; (b) skill id 357.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: signed lower or exclusive upper test fails before first call.
  - Outputs: each case result = 0; helper/write logs = []; seeds unchanged.
  - Discriminates: rejects inclusive upper bounds and target/preparation work before the guard.
  - runtime_observed: no

- **MRH-V02** Nonempty interval excludes upper endpoint (MRH-R03; boundary: supplied-input stage)
  - Inputs: `calc1` = 20, `calc2` = 50; currentLife = 2560, maximumLife = 25600; life-write log empty.
  - Seeds: C = {lo: 1, hi: 666}.
  - Draws: (C, modulus 30, result 21); final C = {lo: 1791398751, hi: 0}.
  - Trace: lower 20, upper 50, range 30; percent 41; mulDiv(41,25600,100) = 10496; candidate 13056; neither clamp changes it.
  - Outputs: write log = [(T, stat 6, layer 0, raw 13056)]; result = 1; C as above.
  - Discriminates: using an inclusive 31-value interval changes the roll/result; using current life as percentage basis changes the write.
  - runtime_observed: no

- **MRH-V03** Normalization, no-draw equality and raw floor (MRH-R03; boundary: supplied-input stage)
  - Inputs: case-specific calculations/life below; seed C = {lo: 1, hi: 666}; empty write log.
  - Seeds: `no draw` in all cases because normalized endpoints equal.
  - Draws: none.
  - Trace: normalize upper, normalize lower, then apply percentage and ordered clamps.
  - Cases:
    - (a) `calc1` = 120, `calc2` = 150, currentLife = 1, maximumLife = 1000. Outputs: upper/lower = 100/100; percent = 100; candidate = 1001; write = [(T,6,0,1000)]; result = 1.
    - (b) `calc1` = -7, `calc2` = -1, currentLife = 0, maximumLife = 25600. Outputs: upper/lower = 0/0; increase = 0; write = [(T,6,0,1)]; result = 1.
    - (c) `calc1` = 10, `calc2` = 10, currentLife = 0, maximumLife = 0. Outputs: upper/lower = 10/10; increase = 0; floor gives 1 then maximum clamp gives 0; write = [(T,6,0,0)]; result = 1.
  - Outputs: only case-specific writes above; C stays {lo: 1, hi: 666} in every case.
  - Discriminates: (a) rejects failing to cap percentages or clamp to maximum; (b) rejects a floor of raw 256; (c) rejects applying the raw lower floor after the maximum clamp.
  - runtime_observed: no

- **MRH-V04** Large maximum uses divide-first arithmetic (MRH-R03; boundary: supplied-input stage)
  - Inputs: `calc1` = 25, `calc2` = 25; currentLife = 100; maximumLife = 1048581; empty write log.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: equal endpoints give 25; maximum > 1048576 and 100 <= maximum >> 4, so NRF-R05 gives trunc(1048581/100)*25 = 262125; add 100 = 262225.
  - Outputs: write log = [(T,6,0,262225)]; result = 1; seeds unchanged.
  - Discriminates: exact multiply-then-divide gives candidate 262245, which is rejected.
  - runtime_observed: no

- **MRH-V05** A one-value interval still advances the seed (MRH-R03; boundary: supplied-input stage)
  - Inputs: `calc1` = 25, `calc2` = 26; currentLife = 100; maximumLife = 1000; empty write log.
  - Seeds: C = {lo: 1, hi: 666}.
  - Draws: (C, modulus 1, result 0); final C = {lo: 1791398751, hi: 0}.
  - Trace: range 1 draws once; percent 25; increase 250; candidate 350.
  - Outputs: write log = [(T,6,0,350)]; result = 1; C as above.
  - Discriminates: replacing range <= 0 with range <= 1 as a no-draw optimization leaves a wrong seed.
  - runtime_observed: no

- **MRH-V06** Operand order and signed edge values (MRH-R03; boundary: supplied-input stage)
  - Inputs: case-specific calculations and life values below; target T/caster C and empty life-write log as in the section preamble. These are isolated supplied arithmetic inputs; no claim is made that real target preparation admits the overflow or negative-maximum cases.
  - Seeds: `no draw` in every case because calculation endpoints are equal.
  - Draws: none.
  - Trace: normalize endpoints, apply NRF-R05 in percentage/maximum/100 order, wrap the raw addition to signed i32, apply raw floor 1 and then signed maximum clamp.
  - Cases:
    - (a) `calc1` = 25, `calc2` = 25; currentLife = 0; maximumLife = 100007. Outputs: percent = 25; increase = trunc(2500175/100) = 25001; write log = [(T,6,0,25001)]; result = 1.
    - (b) `calc1` = 100, `calc2` = 100; currentLife = 2147483600; maximumLife = 1000. Outputs: increase = 1000; raw addition 2147484600 wraps to -2147482696; floor gives 1; write log = [(T,6,0,1)]; result = 1.
    - (c) `calc1` = 25, `calc2` = 25; currentLife = 100; maximumLife = -1000. Outputs: increase = -250; candidate = -150; floor gives 1 then signed maximum clamp gives -1000; write log = [(T,6,0,-1000)]; result = 1.
  - Outputs: only the case-specific exact life-write requests/result above; all seeds unchanged.
  - Discriminates: (a) swapping the first two mulDiv operands instead selects trunc(100007/100)*25 = 25000 and is rejected; (b) widened or saturating addition writes maximumLife 1000 instead of raw 1 and is rejected; (c) unsigned maximum comparison or floor-after-maximum ordering does not produce raw -1000 and is rejected.
  - runtime_observed: no

- **MRH-V07** Actual selected-target normalization (MRH-R06; boundary: isolated rule)
  - Inputs: valid admitted skill id 179/count 357; nonempty game G, caster C (monster/id 1) and caster path; initial do-handled latch false; cached identity and registered units per independent case. T is monster/id 9; U is a distinct monster/id 9 replacing T in G; I is item/id 9 with the stated mode. Registry entries have unique type/id pairs. Calculations and stat writes have empty logs.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: resolve the cached pair only when a selected unit exists; clear stale/excluded selection without retargeting or replacing the cache; then test absent/self.
  - Cases:
    - (a) selected T, cache (1,9), G registers T. Outputs: selected T and cache (1,9) unchanged; proceed to MRH-R07 with T; no callback result yet.
    - (b) selected T, cache (1,9), G registers U. Outputs: selected unit absent; cache remains (1,9); callback result 0. Selection does not switch to U.
    - (c) selected T, cache (1,9), no registered unit for that pair. Outputs: selected unit absent; cache remains (1,9); callback result 0.
    - (d) selected C, cache (1,1), G registers C. Outputs: selected C and cache (1,1) unchanged; callback result 0.
    - (e) selected I, cache (4,9), G registers I in mode 1. Outputs: selected unit absent; cache remains (4,9); callback result 0.
    - (f) selected I, cache (4,9), G registers I in mode 2. Outputs: selected unit absent; cache remains (4,9); callback result 0.
    - (g) selected I, cache (4,9), G registers I in mode 0. Outputs: selected I and cache (4,9) unchanged; proceed to MRH-R07 with I; no callback result yet.
    - (h) selected unit absent, cache (1,9), G registers T. Outputs: selection remains absent and cache (1,9) unchanged; callback result 0; no registry lookup occurs.
  - Outputs: each case's selection/cache/control result above; latch stays false; calculation and stat-write logs stay empty; all stats and seeds unchanged.
  - Discriminates: (b) rejects retargeting a stale selection to the registry replacement; (c) rejects trusting a dangling selection; (d) rejects self-heal; (e)/(f)/(g) reject a missing, overbroad or reversed item-mode filter; (h) rejects recovering an absent selection solely from the cached pair. None of these outcomes depends on a supplied target-predicate result.
  - runtime_observed: no

- **MRH-V08** Callback through real expressions and stat reads (MRH-R06/MRH-R07/MRH-R03; boundary: integration through replacing-write request)
  - Inputs: valid loaded skill id 179/count 357, callback level 1; G registers caster C (monster/id 1) and target T (monster/id 9); caster path selects T with cached pair (1,9); latch false. Skill `calc2` reference 0, `calc1` reference 3, skill-pool size 12. T has a stat list with usesResolvedStats true, associated owner T, ordered own entries [(6,0,1000),(7,0,10000)] and resolved entries [(6,0,2560),(7,0,25600)]; each array's active count is 2. Stat-definition count 8 with present definitions for 6/7, each fMin true, MinAccr 1, ValShift 8. Supplied replacing-write operation logs its request, returns 0 normally and makes no stored change or draw; only that shared storage boundary is supplied. Operation log starts empty. The expression bytes are synthetic valid compiled input, not a claim about the text compiler or original loaded expression bytes.
  - Seeds: C = {lo: 1, hi: 666}, T = {lo: 9, hi: 17}; per-case draw details below.
  - Draws: calculations have no random tokens; case (a) draws (C, modulus 30, result 21); case (b) draws none.
  - Trace: retain T; set latch; evaluate upper then lower with caster/179/level 1; lower bytes compute 15+5*level = 20; normalize; perform any bounded draw; actual RLI reads select T's resolved maximum/current entries; compute and request replacement; ignore its 0 result.
  - Cases:
    - (a) Pool bytes in hexadecimal = `07 32 00 07 0f 07 05 04 10 12 10 00`. Outputs: ordered log = [latch(C)=true, eval(`calc2`,C,179,1)=50, eval(`calc1`,C,179,1)=20, draw(C,30)=21, read(T,7,0)=25600, read(T,6,0)=2560, replace(T,6,0,13056)=0]; callback result 1; final C seed {lo: 1791398751, hi: 0}; T seed remains {lo: 9, hi: 17}; selection T/cache (1,9) unchanged; both stat arrays unchanged at the supplied write boundary.
    - (b) Pool bytes in hexadecimal = `07 00 00 07 0f 07 05 04 10 12 10 00`. Outputs: ordered log = [latch(C)=true, eval(`calc2`,C,179,1)=0, eval(`calc1`,C,179,1)=20, read(T,7,0)=25600, read(T,6,0)=2560, replace(T,6,0,2560)=0]; normalized upper/lower 0/0; callback result 1; C seed remains {lo: 1, hi: 666}; T seed remains {lo: 9, hi: 17}; selection T/cache (1,9) unchanged; both stat arrays unchanged at the supplied write boundary.
  - Outputs: exact per-case logs, control/latch/selection/cache/seed state and unchanged supplied stat arrays above. No real replacing-storage or notification effect is asserted.
  - Discriminates: the logs reject upper/lower reversal, target instead of caster expression context, wrong skill/level context, latch setting after evaluation, reading own/base entries instead of resolved entries, reversed stat-read order, skipping lower evaluation when upper is zero, and treating the replacing writer's 0 return as callback failure. Actual expression and getter results are computed from the supplied pool/arrays rather than supplied as arbitrary return values.
  - runtime_observed: no

### 8.2 Observation vectors

**MRH-O01** Measure a Normal-difficulty healing monster with a damaged eligible target, capturing the actual two calculation results, caster seed, current/maximum raw life, setter request and final stat. Compare only actual matching inputs to MRH-V02–MRH-V05. [Procedure and collection result](../research/spec-evidence/skill-monster-resurrect-and-heal.md#observation-procedure) distinguish an idle empty capture from execution evidence. Resurrection/blood-suck observations additionally require MRH-Q02/MRH-Q03's missing semantics before expected final effects can be specified.

## 9. Acceptance and dependencies

Partial acceptance is MRH-R01's cells, MRH-V01 for MRH-R02, MRH-V02–MRH-V06 for the supplied-result MRH-R03 arithmetic, MRH-V07 for actual target normalization, and MRH-V08 for the complete valid callback through expression execution and existing-stat reads to the replacing-write request. Full behavior additionally requires MRH-Q01's actual replacing storage/propagation, requested SCX producer gates where applicable, MRH-Q02–MRH-Q04, and executable integration vectors for resurrection and blood sucking. No partial pass establishes complete monster healing or resurrection parity.

**Depends on:** NRF-R01/NRF-R02 for RNG step/bounded draws, NRF-R05 for ordered mulDiv, NRF-R08a for raw life units; SCD-R03–SCD-R09c for callback entry and latch consumption with existing gates; RLI-R02a/RLI-R02b for actual selected-array life/max-life reads; SCX-R01–SCX-R11 for compiled execution/front-end requests, SCX-R13–SCX-R28 for requested bindings/parameters with their stated producer gates (MRH-V08 uses only SCX-R20/SCX-R21's supplied-level selector). [Timed-state lifecycle](state-timed-state-lifecycle.md) STL-R06c/STL-Q07 and RLI-Q03 retain actual replacing storage/propagation. Corpse lifecycle, collision and blood-combat integration remain their named open questions.

**Depended on by:** no known consumed plan, code or tests.

## 10. Related specs

- [Skill cast start and do dispatch](skill-cast-start-and-do-dispatch.md): selected start/action routing.
- [Numeric and RNG foundations](core-numeric-and-rng-foundations.md): arithmetic order and seed ownership primitives.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c artifacts `LOD113C-D2GAME` and `LOD113C-D2COMMON` in the [source manifest](../research/source-manifest.md). Original dispatch slots, local instructions and selected imported targets were read on 2026-10-29; original-image decoding corroborated encoded preparation, lookup and write-wrapper instructions. Byte-equivalent reconstruction supplied the small target accessors and expression front end; compatible SCX/RLI contracts supply expression execution and existing-stat reads. Effective Skills text was read through an owned original-profile archive loader. No gameplay vector matched a capture; an idle healing trace yielded zero records and supplies no execution proof. [Companion evidence](../research/spec-evidence/skill-monster-resurrect-and-heal.md) records the exact supported slice.
