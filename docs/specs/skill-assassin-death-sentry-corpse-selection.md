# Death Sentry corpse selection

**Description**

A Death Sentry trap repeatedly decides whether to explode a nearby monster corpse or use its other attack. It looks for a corpse near its current target, never picks the corpse it used last, and takes a close enough corpse without any chance roll. Otherwise it may use its other attack, depending on the target's distance and a random chance. Each successful choice uses up one of the trap's shots.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: DST-R, DST-Q, DST-V, DST-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-assassin-death-sentry-corpse-selection.md)
Open evidence gaps: [gap file](../plans/gaps/skill-assassin-death-sentry-corpse-selection.md)

## 1. Scope

Defines the reached Death Sentry think: budget and skill check, shared target, corpse-search request and first-result consumption, corpse-distance branch, fallback range/chance branch and the remembered corpse. The explosion that the corpse branch requests is owned by [Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md) (DSC), including the corpse-valid producer DSC-R01b consumed by this search. Shared budget/target search, two-unit distance/size and draws consume the named dependency contracts. Actual AI entry, spatial eligibility, scheduler/mode execution and competing corpse claims retain the gates below.

## 2. Trigger and ownership

D2Game computes server thinking. The reached think checks budget and corpse skill, obtains the shared target, searches for a corpse, then requests corpse or fallback skill mode. Actual DeathSentry AI binding is DST-Q01. A mode-9 request does not establish the frame when the explosion executes; that explosion is the separate Skills.srvdofunc 55 action (DSC section 2). No independent interval or client message is established. [Dispatch anchors](../research/spec-evidence/skill-assassin-death-sentry-corpse-selection.md#dispatch-and-functions) distinguish this entry boundary.

## 3. Inputs and data

Arithmetic is signed i32 with low32 wrapping unless stated. Division truncates toward zero. Time is server frames, positions are integer path units (NRF-R10a/NRF-R10b). Seed pairs are u32 low/high (NRF-R01/NRF-R02). Ordinary geometry fixtures use nonnegative representable differences.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| sentry, target, corpse | runtime identities/kinds | optional units | selection/exclusions |
| Skill1/Skill2; L/record | MonStats; skill operations/Skills | signed i16 identities; i32 L; optional record | corpse/fallback skill and progression context |
| difficulty; `aip2`/`aip3`/`aip4` with N/H variants | game; MonStats, GDA-R08a | u8 difficulty=0..2; signed i16 words | delay, fallback chance, fallback range |
| owner/room, shot counter | SFR-R02 | supplied results; i32 counter | budget failure/decrement |
| target/d1 | SFR-R03a/SFR-R08/SFR-R10 | optional target; i32 distance | stored-first shared search result |
| lastCorpse | sentry's remembered corpse | i32 identity | excludes its last corpse only |
| Param3/Param4 | Skills, GDA-R11a | i32 | corpse-distance progression |
| corpse-valid result | DSC-R01b | boolean | first search-callback test |
| positions/sizes/rooms | path/size/spatial operations | i32 | search point and corpse distance |
| sentry seed | sentry unit | u32 pair | fallback chance draw |

| Mapping or constant | Origin | Binding |
|---|---|---|
| search radius 10/filter 4098 | engine code | fixed by original code |
| modes 9 `skill2`/14 sequence | engine code | code-fixed use; names GDF-R06a |
| selected Skills/MonStats values | table data | values from selected record; consumer meanings code-defined |
| mon death sentry Id=312; deathsentry Skill1=mon death sentry, Skill2=death sentry ltng, AI=DeathSentry | table data | [TD-DST named text cells](../research/spec-evidence/skill-assassin-death-sentry-corpse-selection.md#named-table-cells); compiled/AI integration DST-Q01/DST-Q03 |

Other vector values are synthetic inputs.

## 4. Rules (execution order)

**DST-R01** Budget and skill [verified in 1.13c code (static)]. Run SFR-R02 decrement=false; nonzero lost stops. Otherwise obtain sentry Skill1 and bonus-inclusive L. Missing skill or signed L<=0 stops. These branches request no search, chance draw or local delay; budget effects remain owned by SFR. Actual skill/level production is DST-Q03.

**DST-R02** Shared target [verified in 1.13c code (static)]. Request SFR's stored-first target operation. Consume its unit/d1. No target requests the difficulty-selected signed `aip2` delay and stops. This consumer does not verify actual search or due-frame production (DST-Q03).

**DST-R03a** Corpse search/result [verified in 1.13c code (static)]. Missing sentry, L<=0 or missing target returns no corpse without finding units. Otherwise request unit finding from target room/position with radius 10/filter 4098 and DST-R03b predicate. Nonzero result count takes the first returned unit, without sorting; zero count returns none. Request release after reading the result. No local draw.

**DST-R03b** Integrated search [unresolved]. The callback first tests DSC-R01b then the unit-find predicate with candidate/search descriptor; either zero rejects. Complete spatial filtering, result ordering and collection production remain DST-Q03; the corpse-valid producer remains DSC-Q04. A plain dead-monster test does not reproduce these producers.

**DST-R04** Corpse branch [verified in 1.13c code (static)]. Invalid skill/absent record or L<=0 gives maxDist=0; otherwise `maxDist=low32(Param3+low32((L-1)*Param4))`; half=trunc(maxDist/2). For a present corpse different from lastCorpse, per axis subtract trunc(size(target)/2)+trunc(size(corpse)/2) from the absolute coordinate difference, clamp negative results to 0, then distance=trunc(low32(2*max(dx,dy)+min(dx,dy))/2). If signed distance<half, run SFR-R02 decrement=true. Lost stops before the identity write. Success stores corpse identity and requests Skill1/mode 9 on corpse with point(0,0). No local chance/`aip4`/delay. Size-by-kind consumes SFR-R10's size operation; that rule's full single-candidate subtraction is not this two-unit metric.

**DST-R05** Fallback [verified in 1.13c code (static)]. When DST-R04 is not taken, signed d1>=`aip4` requests `aip2` delay without drawing. Otherwise draw Rand(sentry seed,100). Signed roll>=`aip3` requests `aip2`; a lower roll runs decrementing SFR-R02. Lost stops with the advanced seed; success requests Skill2/mode 14 on target/point(0,0), with no local delay. Signed AI words are not clamped.

**DST-R06a** Remembered corpse/timing [verified in 1.13c code (static)]. Only own lastCorpse is compared; write it after successful decrement. State 104 is requested later at explosion DSC-R01a, not at selection. Request timing alone does not prove which eligibility checks consult that state.

**DST-R06b** Competing claims [unresolved]. Selection, mode execution and state request are distinct; whether another sentry can select/explode the same corpse in between remains DST-Q02. No atomic claim is established.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| DST-R01 | Budget/skill | verified in 1.13c code (static) | DST-Q01 entry; DST-Q03 producers | DST-V01 |
| DST-R02 | Target/delay | verified in 1.13c code (static) | DST-Q01 entry; DST-Q03 actual search/schedule | DST-V02 |
| DST-R03a | First search result | verified in 1.13c code (static) | DST-Q03 actual collection | DST-V03 |
| DST-R03b | Integrated search | unresolved | DST-Q03; DSC-Q04 corpse-valid producer | none |
| DST-R04 | Corpse branch | verified in 1.13c code (static) | DST-Q01 entry; DST-Q03 actual mode | DST-V03 |
| DST-R05 | Fallback | verified in 1.13c code (static) | DST-Q01 entry; DST-Q03 actual mode/delay | DST-V04 |
| DST-R06a | Remembered id/timing | verified in 1.13c code (static) | DST-Q03 actual mode execution; DSC-Q04 actual state | DST-V03, DSC-V06 |
| DST-R06b | Competing claims | unresolved | DST-Q02 | none |

## 5. Outputs and side effects

| Effect | Rules |
|---|---|
| budget checks/decrement/owned effects | DST-R01/DST-R04/DST-R05, SFR-R02 |
| signed delay and targeted skill-mode requests | DST-R02/DST-R04/DST-R05 |
| search and release requests | DST-R03a |
| remembered corpse identity | DST-R04/DST-R06a |
| direct sentry seed update | DST-R05 |

No direct local resource write, client message, persistence format or state request is established; the explosion's effects are DSC section 5.

## 6. Branches and edge cases

No target delays; absent/remembered/far corpse falls back. Passing corpse branch ignores `aip3`/`aip4`. Signed nonpositive skill level stops think; maxDist=0 cannot accept ordinary nonnegative distance. Progression wraps under low32, so a wrapped negative half rejects ordinary distances. AI words are signed and unclamped; chance acceptance is strict. A lost second budget check stops after any fallback draw. DSC-R01b's producer rejects players. No other-build equivalence is asserted.

## 7. Open questions

All open assessments also appear in the [gap file](../plans/gaps/skill-assassin-death-sentry-corpse-selection.md).

| Id | Status / kind | Missing evidence and exact limit |
|---|---|---|
| **DST-Q01** | open / integration | DeathSentry text-to-compiled-AI-selector-to-think-table chain. Named text AI cell does not establish a structural entry match. |
| **DST-Q02** | open / behavior | Scheduler/eligibility ordering for two sentries and one corpse; source chain plus matched observation distinguishing competing outcomes. |
| **DST-Q03** | open / integration | Target/skill/level producers, search/spatial filtering/result ordering/collection production, delay/mode effects and compiled correspondence of named think cells. Local supplied stages remain available. |

Gameplay observations for every behavioral rule remain pending; the one recipe is an initial control, not exhaustive branch coverage. No inability to observe the eligible 1.13c behavior has been established.

## 8. Test vectors

### 8.1 Executable vectors

Cases start independently and are synthetic. Valid room/record, nonzero skill, L=1. Seeds: S={6,666} (sentry), C={1,666} (corpse); no draw leaves every seed unchanged. Local rules execute; supplied producers log ordered requests/return listed results with no unspecified writes/draws. Budget defaults counter 3/lost=0; decrement stores 2; a supplied lost=1 has no writes unless listed. Mode/delay requests log only. Search/eligibility results are supplied and do not verify DST-R03b/DSC-R01b. A supplied operation's effects do not verify its original producer. All vectors are not runtime-observed.

- **DST-V01** Think stops (DST-R01; boundary: supplied-input stage)
  - Inputs: initial budget lost=0/Skill1=312 absent.
  - Seeds: no draw.
  - Draws: none.
  - Trace: budget then lookup failure.
  - Outputs: stop/counter 3/no search/delay/mode.
  - (a) Inputs: skill present/bonus-inclusive L=0. Draws: none. Outputs: same stop/unchanged counter/no search/delay/mode.
  - (b) Inputs: initial lost=1/no supplied budget writes. Draws: none. Outputs: stop before skill lookup/counter 3/no search/delay/mode.
  - Discriminates: Missing-skill delay/search or ignoring budget failure fails.
  - runtime_observed: no

- **DST-V02** No-target delay (DST-R02; boundary: supplied-input stage)
  - Inputs: budget passes/Skill1 present/L=1/difficulty=1; `aip2` Normal 2/Nightmare 7/Hell 9; search absent/distance=2147483647.
  - Seeds: no draw.
  - Draws: none.
  - Trace: select signed Nightmare word 7.
  - Outputs: ordered budget/lookup/level/search/delay=7; no corpse-search/decrement/mode/counter 3. No exact due-frame assertion.
  - Discriminates: Normal-word selection or corpse-search-before-delay changes log.
  - runtime_observed: no

- **DST-V03** First corpse/strict distance/remembered id (DST-R03a/DST-R04/DST-R06a; boundary: supplied-input stage)
  - Inputs: target(100,100)/size 2; A id=77(111,100)/size 2; B id=78(104,100)/size 2; returned[A,B]/count 2; Skill1=312/Param3=20/Param4=2/L=1/lastCorpse=0/counter 3.
  - Seeds: no draw.
  - Draws: none.
  - Trace: target-centered radius 10/filter 4098 request; supplied A first; reduced=11-1-1=9; maxDist=20/half=10;9<10; decrement passes.
  - Outputs: consume A then release; lastCorpse=77/counter 2; request Skill1=312/mode 9/targetA/point(0,0); no chance/delay.
  - (a) Inputs: independently lastCorpse=77/d1=8/`aip4`=8/`aip2`=3. Draws: none. Outputs: fallback delay=3/counter 3/lastCorpse=77/no mode.
  - (b) Inputs: independently A(112,100)/lastCorpse=0/d1=8/`aip4`=8/`aip2`=3. Draws: none. Trace: distance=10 fails=10<10. Outputs: delay=3/counter 3/lastCorpse=0/no mode.
  - (c) Inputs: independently L=2/A(112,100)/lastCorpse=0. Draws: none. Outputs: maxDist=22/half=11/distance=10 passes/counter 2/lastCorpse=77/same mode 9 request.
  - (d) Inputs: independently distance=9/second budget lost=1/no budget writes. Draws: none. Outputs: counter 3/lastCorpse=0/no mode/no delay.
  - (e) Inputs: isolated search, absent sentry (or L=0, or absent target). Draws: none. Outputs: no corpse/no unit-find/no release.
  - (f) Inputs: isolated valid search, present sentry/target, L=1, target room/point(100,100), supplied find result count 0 with an empty result collection. Draws: none. Trace: request finding, read count 0, then request release. Outputs: no corpse, ordered find(target room,point(100,100),radius 10,filter 4098,predicate DST-R03b) then release(collection); no seed change.
  - (g) Inputs: independent reached corpse-branch stage, target(100,100)/size 2, corpse A id 77 at(111,100)/size 2, lastCorpse=0, valid Skill1=312/L=2, Param3=2147483647/Param4=1; fallback d1=8/`aip4`=8/`aip2`=3, counter 3. Draws: none. Trace: low32(2147483647+(2-1)*1)=-2147483648; half=-1073741824; distance 9 fails signed 9<half. Outputs: delay 3, counter 3/lastCorpse 0/no decrement/mode, all seeds unchanged.
  - Discriminates: Sorting chooses B; <=half accepts(b); write-before-budget fails(d); unbounded progression arithmetic takes the wrong corpse branch in (g); skipping release on a valid empty result fails (f), while invalid-input (e) must not request release. Actual A search eligibility is not established by the supplied collection.
  - runtime_observed: no

- **DST-V04** Fallback seed owner (DST-R05; boundary: supplied-input stage)
  - Inputs: present target/unusable corpse/d1=5/`aip4`=8/`aip3`=50/`aip2`=3/Skill2=324/counter 3.
  - Seeds: S={1,0x29A}, C={6,0x29A}.
  - Draws: S mod=100->51/S{0x6AC6935F,0}; C unchanged.
  - Trace:5<8 then 51>=50.
  - Outputs: delay=3/no decrement/mode/counter 3/lastCorpse unchanged.
  - (a) Inputs: independently `aip3`=52/same entry seeds. Draws: S mod=100->51/S{0x6AC6935F,0}. Outputs: counter 2/Skill2=324/mode 14/target/point(0,0)/no delay; C unchanged.
  - (b) Inputs: independently d1=8/same entry seeds. Draws: none. Outputs: delay=3/counter 3/S{1,0x29A}/C{6,0x29A} unchanged.
  - (c) Inputs: independently `aip3`=52/second budget lost=1/no writes/same entry seeds. Draws: S mod=100->51/S{0x6AC6935F,0}. Outputs: counter 3/no mode/no delay/C unchanged.
  - (d) Inputs: independently present target/unusable corpse/d1=5/`aip4`=8/`aip3`=51/`aip2`=3/Skill2=324/counter 3; entry S={1,0x29A}, C={6,0x29A}. Draws: S mod 100 ->51, final S={0x6AC6935F,0}; C unchanged. Trace: 5<8, then 51>=51. Outputs: delay 3, counter 3, no decrement/mode, lastCorpse unchanged.
  - Discriminates: Wrong seed, non-strict range or ignoring second-budget return fails; chance acceptance on equality fails (d).
  - runtime_observed: no

The state-request half of DST-R06a's timing (state 104 requested at explosion, not selection) is asserted by DSC-V06.

### 8.2 Observation vectors

All pending. [Tool-neutral procedures](../research/spec-evidence/skill-assassin-death-sentry-corpse-selection.md#observation-procedures) give controls and reading points.

- **DST-O01** Two sentries/one corpse in one frame window: record selections/mode dispatch/state/eligibility; outcome unresolved.

## 9. Acceptance and dependencies

Partial acceptance reproduces every verified assertion in DST-V01..DST-V04 within its supplied boundary. TD-DST proves named effective text cells only. Supplied search/mode operations do not verify producers. DST-R03b/DST-R06b retain their exact Id-map gates. Full reproduction additionally needs DST-Q01..DST-Q03 resolution and gameplay observations of every behavioral rule. No partial pass establishes parity.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md) NRF-R01/NRF-R02 draws, NRF-R10a/NRF-R10b positions; [fields](core-game-data-fields.md) GDA-R08a/GDA-R11a/GDF-R06a; [sentries](skill-assassin-sentry-firing-and-range.md) SFR-R02 budget, SFR-R03a/SFR-R08 target return/gates, SFR-R10 size-by-kind only; [Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md) DSC-R01b corpse-valid producer (gated). Owners of the cited GDA ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA).

**Depended on by:** No downstream behavioral consumer is established. The explosion in [Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md) is reached through the requested mode, whose execution remains DST-Q03; supporting citations are not extra prerequisites to supplied arithmetic.

## 10. Related specs

- [Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md): the explosion requested by the corpse branch.
- [Sentry firing and range](skill-assassin-sentry-firing-and-range.md): shared shot budget and target search.
- [Skill cast start and do dispatch](skill-cast-start-and-do-dispatch.md): mode requests leading to skill actions.

## Appendix A. Provenance

Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies original D2Game/D2Common artifacts. Compatible disassembly/RNG/distance readings retain 2026-10-12 dates/limits; targeted think, search, progression and predicate readings and original-image/table checks date 2026-10-15. The [companion](../research/spec-evidence/skill-assassin-death-sentry-corpse-selection.md) records dispatch, mappings, corrections, named text cells and profile limits. No gameplay observation or other-build analysis was performed.
