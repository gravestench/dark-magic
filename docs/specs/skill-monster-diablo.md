# Diablo monster skills

**Description**

These skills control selected special actions used by Diablo. The cold action prepares a combat record, applies it to a refreshed target and requests the resulting reaction. Complete combat effects and the other actions still require investigation of their supporting operations.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-game-data-fields](core-game-data-fields.md)
Ids: DMB-R, DMB-Q, DMB-V, DMB-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-diablo.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-diablo.md)

## 1. Scope

The server skill actions selected by Skills.srvdofunc values 100, 103, 104, 152, including the associated start-to-do handoff. DiabLight uses srvstfunc 53 and DiabRun uses srvstfunc 54 in the same original text records; these start paths remain DMB-Q01.
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
| Result flags / flags / hit-class source values | Selected skill record fields | unsigned 16-bit / unsigned 32-bit / signed 32-bit | Skills.ResultFlags / Skills.HitFlags / Skills.HitClass; loader mapping established locally, original-profile values gated | runtime input | supplied record fields |
| Target context | Present game, actor and dynamic path; selected target, cached kind/id and matching game registry | Logical units, signed kind/id | Normal-entry normalization under MRH-R06; cached kind is supported 0..5 when looked up | runtime input | supplied context |
| Target eligibility | Supplied target stage | distinct non-null unit | Entry precondition for isolated DMB-R04; DMB-R05 produces this locally | runtime input | supplied input |
| Do-handled latch | Actor dispatch state under SCD-R07 | boolean | Set by the valid action-100 entry | engine code | fixed by original code |
| Existing target stats | RLI-R02a/RLI-R02b unit-stat inputs | Logical lists, ordered arrays, definitions | Post-apply stat 6/layer 0, not an own-array-only read | runtime input | dependency contract |
| Overlay input | Skills.srvoverlay and overlay table count | signed 16-bit selector, signed 32-bit count | Caller request guard followed by overlay-entry guard | table-selected input | resolved from supplied data |
| Apply, reaction and overlay operations | Shared combat and overlay contracts | Supplied operations returning normally | Ordered requests may mutate the same combat record and target stats; full effects retain DMB-Q03 | runtime input | dependency contracts |
| Duration operation and fill operation | SLD-R11 / SLD-R15 | supplied/logged operations | Duration and fill execution or supplied results at the isolated preparation boundary | runtime input | dependency contracts |
| Combat record | Shared DamageRecord, [record foundation section 3.1](core-record-layouts-and-ids.md#31-damage-record) | Logical fields with the widths and units defined there | Exact field set cleared and handed off by DMB-R04; cold duration means coldLength | runtime input | shared semantic contract |
| Effective skill parameters | Skills record | unresolved | Parameters, calculations and effect selectors, gated by DMB-Q01 | unresolved | unresolved |

The selector mappings are code-table observations, not verified archive cell claims.
The effective text-file selector cells are established by DMB-R03; the named identity/start/action selector correspondence is corroborated in the limited observed profile; other compiled fields remain gated.

## 4. Rules (execution order)

**DMB-R01** Negative skill index [verified in 1.13c code (static)]

For the selected actions, an index less than zero returns failure (integer zero)
before accessing a skill record or calling a dependency. This isolated entry branch
does not draw RNG, write actor state, or emit an effect. The index is signed 32-bit.
This rule covers only a negative index, not any other invalid-record case.

**DMB-R02** Valid-action execution [unresolved]

Actions 103, 104 and 152 retain unresolved successful effects, targeting, repeated-hit handling, object interactions and downstream creation. For action 100, DMB-R04 through DMB-R07 establish local target selection, ordered requests and final return; completed damage, reaction and valid-overlay effects retain DMB-Q03 and the named shared contracts.

**DMB-R03** Original text-file action bindings [verified in 1.13c data]

The named cells below are in effective `data/global/excel/skills.txt`, selected
by the `skill` column and read by the `srvdofunc` header. Origin: table data.
Binding: selected by data from engine-defined handlers. They establish text-file
contents. The named identity/start/action selectors additionally match the loaded records in the limited source profile recorded in the companion; other field correspondence remains DMB-Q01.

| Row identity (`skill`) | `srvdofunc` raw decimal value |
|---|---|
| DiabLight | 152 |
| DiabCold | 100 |
| DiabRun | 103 |
| DiabPrison | 104 |

The same effective text record has DiabCold.EType raw value `mag`. This is a named text cell only; numeric compiled EType correspondence remains DMB-Q01.

Source: [table-data record](../research/spec-evidence/skill-monster-diablo.md#table-data-record).

**DMB-R04** Cold-action request preparation [verified in 1.13c code (static)]

For action selector 100 after a distinct non-null target is supplied by the
outside-boundary target stage, clear every semantic field of the shared DamageRecord in [core-record-layouts-and-ids section 3.1](core-record-layouts-and-ids.md#31-damage-record) to zero. No binary padding or unnamed extension field is required by this contract. Set
the 16-bit resultFlags to `1 OR suppliedResultFlags`; set the 32-bit flags to
suppliedFlags. If suppliedHitClass is nonzero, set hit class to it, including
negative values; zero retains the cleared value. Request elemental duration
for the same actor, skill and level with condition 1 under SLD-R11; store its
supplied signed return as cold duration. Then request elemental fill under
SLD-R15 on that same record and context. This stage ends at the fill request:
its changes, draws, target production and subsequent damage application are not
asserted here. Local preparation draws nothing; duration/fill computations
retain their dependency contracts and evidence gates.

**DMB-R05** Cold-action record guard, latch and target [verified in 1.13c code (static)]

For action 100, reject signed skill index below zero, at least the skill-record count, or a lookup producing no record: return 0 before latch, target preparation or effect requests. Otherwise, on a present actor set its do-handled latch true before target preparation. The following normal-entry contract requires a present actor, game and dynamic path, and a supported cached kind whenever registry lookup is reached. Execute the target-normalization substage of [MRH-R06](skill-monster-resurrect-and-heal.md): resolve a present selection's cached kind/id in this game, clear selection on identity mismatch or item mode 1/2, and preserve cached kind/id when clearing. Read the resulting selected target. Absent target or target identical to actor returns 1, retaining the latch, without preparing a combat record or calling duration/fill/apply/reaction/overlay. A distinct target proceeds through DMB-R04. The local stage and normalization draw nothing. This action's latch timing and return are specified here; MRH's later healing checks do not supply them.

**DMB-R06** Cold-action damage and reaction sequence [verified in 1.13c code (static)]

On DMB-R05's distinct-target path, complete DMB-R04, then execute SLD-R15 elemental fill on that same record. Fill requests its own duration again, then minimum and maximum, before its range draw and routing under SLD-R16/SLD-R17. The action does not force numeric EType to cold: field writes follow the supplied skill's numeric selector. For example, type 3 adds magic and retains the coldLength stored by DMB-R04. Fill's returned damage is ignored.

Next request direct damage application with `(game, actor, target, same record, scaleRequest=true)` under DPO-R08c/DPO-R13a and the applicable subsequent damage-pipeline stages. After it returns normally, read target stat 6/layer 0 using RLI-R02a/RLI-R02b. If and only if this getter returns exactly 0, OR resultFlags with 2; a negative nonzero value skips the OR. A nonzero result does not clear an existing bit 2. Request combat-result reaction `(game, actor, target, same record)` under HRC-R02 through HRC-R11, preserving each subrule's evidence and gates. Read the overlay selector after reaction and perform DMB-R07. Return 1 after all requested operations return normally, regardless of their returned values. No local RNG step occurs outside the executed shared contracts; their draws and mutations remain theirs. This specifies the caller's request placement and post-apply stat read, not otherwise gated application, reaction or overlay effects.

**DMB-R07** Cold-action overlay request and invalid-entry refusal [verified in 1.13c code (static)]

After reaction, read the skill's signed 16-bit srvoverlay selector. Request overlay `(target, selector, extra=0)` only when selector is strictly positive and at most the current signed overlay-table count. The overlay operation itself first rejects selector below 0 or at least the current overlay count, without unit/stat writes, calls or RNG. Consequently, with a stable count, selector equal to count produces a request that immediately returns without effects. A positive selector below count passes that entry guard; completed list/stat/update effects remain DMB-Q03 / DPO-Q18. There is no local draw in either guard, and the action ignores the overlay operation's return.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| DMB-R01 | Negative-index rejection | verified in 1.13c code (static) | none | DMB-V01 |
| DMB-R02 | Action effects | unresolved | DMB-Q03, DMB-Q04, DMB-Q05, DMB-Q06, DMB-Q07 block their respective handler effects; Q01 compiled integration | none |
| DMB-R03 | Named text-file selectors | verified in 1.13c data | none for named text cells and observed selector correspondence; DMB-Q01 gates other compiled fields | table-data record |
| DMB-R04 | Local supplied-input stage | verified in 1.13c code (static) | none at supplied-input boundary; Q01 retains original-profile compiled integration | DMB-V02, DMB-V03, DMB-V05 |
| DMB-R05 | Record guard, early latch and normalized target | verified in 1.13c code (static) | valid normal-entry context as stated; no effect claim beyond entry | DMB-V04 |
| DMB-R06 | Fill, apply, exact-zero read, reaction, final return | verified in 1.13c code (static) | DMB-Q03 and shared gates for completed supplied apply/reaction effects; Q01 original-profile fields | DMB-V05, DMB-V07 |
| DMB-R07 | Overlay request and invalid-entry refusal | verified in 1.13c code (static) | DMB-Q03/DPO-Q18 only after valid overlay entry | DMB-V06 |

## 5. Outputs and side effects

| Output | Rule | Evidence boundary |
|---|---|---|
| Failure result = 0, unchanged actor and seed | DMB-R01 | Negative-index branch only |
| Other actions' effect requests and final result | DMB-R02 | 103/104/152 gated |
| Cleared and initialized combat record, duration request then fill request | DMB-R04 | Supplied eligible target and duration result; before fill execution |
| Actor do-handled latch, normalized selection, retained cached identity, early/final return 1 | DMB-R05/DMB-R06 | Valid action-100 record and stated normal-entry context |
| Filled combat record and executed fill draws, direct apply then current-life read, optional resultFlags OR 2, reaction request on same record | DMB-R06 | Actual SLD/RLI stages; supplied apply/reaction operations or each owner's effect gates |
| Overlay request with target/selector/zero extra; no overlay effects on invalid entry | DMB-R07 | Completed valid-overlay list/stat/update effects gated |

## 6. Branches and edge cases

Action 100 uses DMB-R05 through DMB-R07, including missing/self target success, exact-zero post-apply life, signed overlay limits and normal-return dependency handling. Invalid actor/game/path context is outside its stated normal-entry boundary. Successful effects and missing-record behavior for the other actions, their unit types, target eligibility, level zero, negative parameters, difficulty differences and effect-helper failures retain their named questions. Neither ordinary successful casts nor client visuals establish these
branches. No Classic behavior is asserted.

## 7. Open questions

**DMB-Q01** [unresolved]: Which remaining compiled effect, state, stat and parameter values correspond to the named text records? The local loader descriptors establish the names ResultFlags, HitFlags, HitClass and srvoverlay for action-100 inputs; this does not establish their loaded values or text-to-numeric EType correspondence. The named identity/start/action selector correspondence is established for the observed profile. Resolve the loader
schema and retain narrow loaded-record comparisons. This is an integration gate
on original-profile record production, not the supplied-input arithmetic or
record preparation, target normalization or local continuation. AI selection remains outside this boundary.

**DMB-Q02** [unresolved]: Which executable cases match an owned, identified
Expansion 1.13c game? Pending observation does not gate the isolated verified branch,
but does prevent a claim of runtime confirmation.

**DMB-Q03** [unresolved]: 100 completed dependency effects. Local normalization, request order, exact-zero read, overlay guards and return are DMB-R05 through DMB-R07. Completed damage application consumes DPO-R08c/DPO-R13a and subsequent stages with their predicate/event/status gates; completed reaction consumes HRC-R02 through HRC-R11 with their wrapper/update gates. Positive in-range overlay effects retain DPO-Q18. Which attributable original-profile casts reproduce these completed effects? Supplied normal-return operations do not close those integrations.

**DMB-Q04** [unresolved]: 103 object interaction and melee branch. Which selected object/target conditions choose object changes versus melee damage, and which exact effects and return values follow?

**DMB-Q05** [unresolved]: 104 prison target and creation. Which target eligibility conditions and creation result produce the action success?

**DMB-Q06** [unresolved]: 152 missile and event continuation. Which missile request and subsequent event transitions determine success and repeat behavior?

**DMB-Q07** [unresolved]: start handoffs. Which state/target preparation and rejection behavior do original start selectors 53 and 54 supply?

## 8. Test vectors

### 8.1 Executable vectors

- **DMB-V01** Reject a negative index (DMB-R01; boundary: isolated rule)
  - Inputs: signed skill index = -1; run once for each action in Scope; all other inputs are uninspected by this branch. The fixture records every attempted skill-record read in a lookup log initially empty; this is a logical access observer, not a required API.
  - Seeds: no draw.
  - Draws: none.
  - Trace: -1 is less than zero, so the entry branch returns before record lookup.
  - Outputs: result = 0; record-lookup log empty; dependency call log empty; actor state unchanged.
  - Discriminates: clamping a negative index to zero and continuing would perform record lookup; the empty lookup/dependency log rejects that reading.
  - runtime_observed: no

- **DMB-V02** Prepare request before elemental fill (DMB-R04; boundary: supplied-input stage)
  - Inputs: action selector = 100; supplied target distinct from actor; valid skill = 194; level = 2; suppliedResultFlags = 4; suppliedFlags = 32; suppliedHitClass = -7; initially every field in the shared section-3.1 semantic DamageRecord set = 99. Duration operation logs (actor,194,2,condition=1), draws nothing and returns 17. Fill operation only logs its input record/context; stop before executing it.
  - Seeds: no draw.
  - Draws: none.
  - Trace: clear record; resultFlags = 1 OR 4 = 5; flags = 32; hitClass = -7; request duration, store returned 17, then request fill.
  - Outputs: ordered log = [duration(actor,194,2,1), elementalFill(actor,194,2,record)]; handed-off record has resultFlags = 5, flags = 32, hitClass = -7, cold duration = 17, every other field in that shared semantic set = 0; actor seed unchanged at this boundary.
  - Discriminates: skipping clear leaves unrelated fields 99; replacing resultFlags instead of OR gives 4; positive-only hitClass replacement gives 0; storing duration after fill fails the record captured by the fill request.
  - runtime_observed: no

- **DMB-V03** Zero hit-class source retains cleared value (DMB-R04; boundary: supplied-input stage)
  - Inputs: action selector = 100; supplied target distinct from actor; valid skill = 194; level = 2; suppliedResultFlags = 4; suppliedFlags = 32; suppliedHitClass = 0; initially every field in the shared section-3.1 semantic DamageRecord set = 99. Duration operation logs (actor,194,2,condition=1), draws nothing and returns 17. Fill operation only logs the input record/context; stop before executing it.
  - Seeds: no draw.
  - Draws: none.
  - Trace: clear the shared field set; resultFlags = 1 OR 4 = 5; flags = 32; source hit class 0 skips replacement, so cleared hitClass remains 0; request duration, store 17, then request fill.
  - Outputs: ordered log = [duration(actor,194,2,1), elementalFill(actor,194,2,record)]; handed-off shared record has resultFlags = 5, flags = 32, hitClass = 0, coldLength = 17, all other fields in the shared semantic set = 0; actor seed unchanged.
  - Discriminates: retaining pre-clear hitClass gives 99; skipping the clear leaves unrelated fields 99; moving the duration write after the fill changes its captured record.
  - runtime_observed: no

- **DMB-V04** Cold-action guard and early latch/target normalization (DMB-R05; boundary: actual local entry plus MRH-R06 normalization)
  - Inputs: each case resets. Action 100, skill count 357, level 2; actor A is a present monster with dynamic path and do-handled latch false, game G has the supplied registry; no other actor flags change. A's cached pair is `(monster,20)`; selected T is a distinct monster with id 20, mode 1. Registry `(monster,20)` resolves to T. The stage stops before DMB-R04 on a distinct target. Duration/fill/apply/reaction/overlay logs start empty and are not executed. All unmentioned state remains supplied and unchanged.
  - Seeds: no draw.
  - Draws: none.
  - Trace and outputs: (a) skill -1; (b) skill 357; (c) skill 194 with lookup yielding no record: each returns 0, latch false, selection T and cached pair unchanged, target normalization/effect logs empty. (d) valid record 194, selected target absent: latch becomes true before normalization, no registry lookup, return 1, cached pair retained. (e) selected A and cached pair `(monster,A.id)` resolving to A: latch true, selection A, return 1. (f) selected stale T with registry resolving to distinct replacement U with the same id: latch true, selection cleared, original pair retained, return 1; U is not selected. (g) select item I, cache `(item,I.id)` resolving to I, mode 1 and separately mode 2: latch true, selection cleared, cache retained, return 1. (h) original distinct T fixture, plus separately item I mode 0 with matching registry: latch true, selection retained and next stage is DMB-R04 with that exact target; no whole-action return asserted at this stopped boundary. A logical observer at normalization entry sees latch true in every valid-record case. Early-return effect logs are empty.
  - Discriminates: late latch, healing-style return 0 on missing/self, substituting U, clearing cached identity, or blanket item rejection fail distinct assertions.
  - runtime_observed: no

- **DMB-V05** Execute fill and post-apply getter in caller order (DMB-R04/DMB-R06; boundary: actual SLD/RLI, supplied apply/reaction/overlay operations)
  - Inputs: action 100 with valid record 194, level 2, game G, actor A and normalized distinct target T; latch true at this continuation. Synthetic supplied skill values are ResultFlags=4, HitFlags=32, HitClass=-7, numeric EType=3, ELen=17, all ELevLen bands=0, empty ELenSymPerCalc, EMin=1, EMax=2, every elemental level-band increment=0, HitShift=8, empty EDmgSymPerCalc, srvoverlay=3, overlay count=4. These numeric fields are test inputs, not original DiabCold compiled-data claims. Combat record initially every semantic field 99. T has available stat-6 definition with fMin false; definition count=359, list usesResolvedStats=true, ownStats=[(6,0,999)] and resolvedStats=[(6,0,256)], each active count 1. Tuple order is `(stat,layer,value)`. The supplied apply operation logs the whole incoming record/context and scale request, then sets record.resultFlags=4, record.total=99 and resolved stat-6 value=0, returning integer 0 with no draws. Supplied reaction logs the same record/context, sets hitClass=123 and returns 0 with no draws. Supplied valid-overlay operation logs target/selector/extra and returns 0 with no draws or other effects. Every unmentioned dependency field takes the stated empty/zero branch; EType 3 requests no elemental mastery stat. Execute SLD-R11, SLD-R15/SLD-R16 and RLI-R02a/RLI-R02b, not supplied fill/getter results.
  - Seeds: A `{low=1, high=666}`; all other seeds unchanged.
  - Draws: two actual durations and both endpoints draw nothing; SLD fill makes one NRF-R01/NRF-R02 range step for range 256: new low=1791398751, high=0, remainder=95, damage=256+95=351. No other operation draws.
  - Trace: clear; resultFlags=5, flags=32, hitClass=-7; first duration stores coldLength=17; fill requests second duration=17, minimum=256, maximum=512, draws and adds magic=351 while retaining coldLength and hitClass. Apply logs this record, mutates resultFlags to 4 and target resolved life to 0; actual unit getter selects resolved value 0 rather than own 999; OR 2 makes resultFlags=6; reaction logs resultFlags=6, total=99, then changes hitClass; overlay request follows it. Ignore returned values; return 1.
  - Outputs: ordered operations `[duration(A,194,2,1), fill.duration(A,194,2,1), fill.minimum(A,194,2,1), fill.maximum(A,194,2,1), range(A,256), apply(G,A,T,record,true), Stat(T,6,0), reaction(G,A,T,record), overlay(T,3,0)]`. Apply's captured record has flags=32, resultFlags=5, magic=351, coldLength=17, hitClass=-7 and every other semantic field=0. Reaction's captured record changes only resultFlags=6 and total=99 from that capture. Final record also has hitClass=123; final resolved life=0, own life=999, A seed low=1791398751/high=0, action result=1, latch stays true. These captured records are values of one shared mutable record at the named times. Completed apply/reaction/valid-overlay effects beyond the explicitly supplied mutations are gated by DMB-Q03.
  - Discriminates: forcing cold damage, skipping the first duration or the fill's second duration, reading life before apply or from ownStats, using a copied pre-apply record, reacting before the OR, or forwarding a dependency's zero return fails the assertions.
  - runtime_observed: no

- **DMB-V06** Overlay caller limit versus actual callee limit (DMB-R07; boundary: local caller guard and invalid overlay entry)
  - Inputs: stable overlay count 4; target T; preceding reaction already returned normally; run separately with signed srvoverlay -1, 0, 3, 4, 5. All other handler work is outside this stage. A supplied observer logs overlay entry and stops on valid entry; it executes the actual initial overlay guard for invalid entry. Unit/stat/update call log initially empty; unit/lists/stats/seeds unchanged at entry.
  - Seeds: no draw.
  - Draws: none.
  - Trace and outputs: -1 and 0 fail positivity, 5 fails upper limit: no overlay request. 3 requests `(T,3,0)` and reaches valid entry; subsequent effects unspecified. 4 requests `(T,4,0)` and actual overlay-entry comparison refuses it: no unit/list/stat writes or calls. The action proceeds to its return 1 on all normally returning cases, including the refused equal-count request. The stage observer's stop on 3 asserts only the reached entry and leaves that action's final return unobserved.
  - Discriminates: caller using strict less-than drops the equal-count request; callee using less-than-or-equal incorrectly reaches effects for 4; zero-as-valid incorrectly requests 0.
  - runtime_observed: no

- **DMB-V07** Exact-zero post-apply test and existing lethal bit (DMB-R06; boundary: V05 actual fill/getter and supplied effects)
  - Inputs: reset all V05 inputs/operations for each case, except apply stores the specified post-apply resolved stat value and resultFlags below; ownStats stat 6=0, definition fMin=false unless stated. Apply still sets total=99, reaction still changes hitClass=123, overlay operation still returns 0, both draw nothing. (a) resolved life=-1, resultFlags=0; (b) resolved life=256, resultFlags=2; (c) resolved life=0, resultFlags=0; (d) resolved life=0, resultFlags=0, fMin=true, MinAccr=1, ValShift=8, associated list owner type=1. Stat arrays each retain active count 1; definition count 359 and available definition as V05.
  - Seeds: each resets A low=1/high=666.
  - Draws: each takes exactly V05's one range draw, new low=1791398751/high=0, remainder=95.
  - Trace: actual unit read selects resolvedStats; returns -1,256,0,256 respectively, the last using RLI-R02b's minimum replacement. Only (c) ORs 2. A nonzero result never clears existing bit 2.
  - Outputs: reaction capture and final record.resultFlags are respectively (a) 0, (b) 2, (c) 2, (d) 0; all other captures/outputs/order/final action return 1 and seed equal V05 with these explicit overrides. Stored resolved life stays respectively -1,256,0,0; getter adjustment in (d) does not write storage.
  - Discriminates: signed nonpositive test sets (a); own-array read sets (a); clearing bit on positive life changes (b); raw stored-life read sets (d).
  - runtime_observed: no

### 8.2 Observation vectors

**DMB-O01** Observe the selected action and its final return/effects on an
owned Expansion 1.13c game after establishing the missing bindings. Expected
values for gated effects are deliberately unspecified; the
[procedure](../research/spec-evidence/skill-monster-diablo.md#observation-procedure)
defines what evidence must be retained. This is a pending investigation, not a
matching observation.

## 9. Acceptance and dependencies

DMB-V01 passing establishes only DMB-R01 for its named negative-index case.
DMB-V02 and DMB-V03 additionally establish only DMB-R04, with the supplied inputs and operation results stated there.
DMB-V04 through DMB-V07 establish the verified action-100 local stages and the actual shared computations explicitly executed there; supplied dependency mutations establish caller sequencing only.
Full reproduction requires resolving DMB-Q03, DMB-Q04, DMB-Q05, DMB-Q06, DMB-Q07 and DMB-Q01, specifying executable vectors for
DMB-R02, and retaining attributable observations under DMB-Q02. A partial result
does not establish parity for these actions.

**Depends on:** [skill dispatch](skill-cast-start-and-do-dispatch.md) SCD-R06, SCD-R07 and SCD-R08 for selecting
the action from the server selector and invoking its handler; successful effect
integration remains DMB-Q01. No damage, missile or monster-creation outcome is
credited through an unnamed dependency.

Target normalization consumes the named normalization substage of [healing](skill-monster-resurrect-and-heal.md) MRH-R06 with its valid-entry preconditions. Existing life reads consume [record/stat foundation](core-record-layouts-and-ids.md) RLI-R02a/RLI-R02b. Direct damage requests consume [damage pipeline](combat-damage-pipeline-order.md) DPO-R08c/DPO-R13a and applicable subsequent stages with their original gates; reaction requests consume [hit recovery](combat-hit-recovery.md) HRC-R02 through HRC-R11, retaining every subrule's labels and effect gates. Overlay completion retains DPO-Q18. None of these references imports a queued melee producer or verifies AI choice.

**Depended on by:** no established consumer.

[Leveled damage](skill-leveled-damage-tohit-and-elements.md) SLD-R11 supplies the duration request/return contract and SLD-R15 supplies the elemental-fill request contract; producer expression gates remain with that owner.

## 10. Related specs

[Skill dispatch](skill-cast-start-and-do-dispatch.md) supplies the action context.
[Numeric foundations](core-numeric-and-rng-foundations.md) owns integer and seed
operations when their consumption is established.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c D2Game, identified by
[source manifest](../research/source-manifest.md), was examined through static
instruction listings on 2026-10-29. The
[companion](../research/spec-evidence/skill-monster-diablo.md) identifies the dispatch slots,
read boundary and unresolved callees. Named selector text cells have the companion record; no gameplay case has been confirmed.
