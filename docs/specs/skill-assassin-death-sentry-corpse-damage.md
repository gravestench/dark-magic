# Death Sentry corpse damage

**Description**

When a Death Sentry trap sets off a monster corpse, the corpse explodes and damages nearby enemies. The explosion uses the corpse's life and level to determine damage, then divides it between physical damage and the skill's element. Its elemental part can reach slightly farther than its physical part.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: DSC-R, DSC-Q, DSC-V, DSC-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-assassin-death-sentry-corpse-damage.md)
Open evidence gaps: [gap file](../plans/gaps/skill-assassin-death-sentry-corpse-damage.md)

## 1. Scope

Defines the Death Sentry corpse explosion action: target and corpse-valid guards, state requests, explosion damage construction, blast geometry, victim copies and damage/reaction hand-off. Choosing the corpse and requesting this action is owned by [Death Sentry corpse selection](skill-assassin-death-sentry-corpse-selection.md) (DST). Expressions, monster base life, element routing and downstream combat consume the named dependency contracts. Spatial eligibility, scheduler/mode execution and complete downstream effects retain the gates below.

## 2. Trigger and ownership

D2Game computes the Skills.srvdofunc 55 explosion. It is reached after the think's mode-9 corpse request (DST-R04); that request does not establish the frame when the explosion executes (DST-Q03). Action 55 prepares/obtains its target before state requests and damage construction; area callbacks apply damage before requesting reaction. No independent interval or client message is established. [Dispatch/output anchors](../research/spec-evidence/skill-assassin-death-sentry-corpse-damage.md#dispatch-and-functions) distinguish this entry boundary.

## 3. Inputs and data

Arithmetic is signed i32 with low32 wrapping unless stated. Division truncates toward zero; `mulDiv(a,b,c)` consumes NRF-R05 in exactly the written order. Damage is raw 8.8 (NRF-R08a), time is server frames, positions are integer path units (NRF-R10a/NRF-R10b). Seed pairs are u32 low/high (NRF-R01/NRF-R02). Ordinary geometry fixtures use nonnegative representable differences/squares; wrapped comparisons are specified below.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| sentry/caster, target, corpse, victim | runtime identities/kinds | optional units; player=0/monster=1 | guards/exclusions/application |
| skill identity; instance/L/record | skill operations/Skills | signed i16 identity; i32 L; optional instance/record | explosion skill and context |
| `calc1`/`calc2`/`calc3`/aurarangecalc | Skills compiled calculations | references; i32 results | life percentages, split, radius R |
| aurafilter/EType | Skills | u32/u8 | area eligibility/element selector |
| ELen/ELevLen1..3/ELenSymPerCalc | Skills | i32 frames/calculation reference | SLD-R11 duration |
| level/maxhp | stats 12/7, layer 0; level reads use the effective array, initial maxhp reads use the base list's own array | i32 level/raw 8.8 life | corpse/caster level and initial own-array life |
| class, selectorByte/selectorFlag, minHP/maxHP | corpse; game; MSD-R10 | i32 class/bounds; zero/nonzero selectors | monster base-life replacement |
| corpse-valid result | dead mode 12, excluded-state overlap, MonStats2.corpseSel | booleans | producer DSC-Q04 |
| corpse state storage | state 104 corpse_nodraw; RLI-R03/STL-R07 | active/changed membership | state/update requests |
| positions/sizes/rooms/ordered chains | path/size/spatial operations | i32; ordered units | area distances |
| record and downstream inputs | NRF section 3.2; PVP/DMI/DPO/HRC/IEP | named record fields/conditions | mutable victim damage and reaction |

| Mapping or constant | Origin | Binding |
|---|---|---|
| divisor 100; monster shift 7 | engine code | fixed by original code |
| state 104; dead mode 12; stats 7/12 | engine code | code-fixed use; names GDF-R02/GDF-R03/GDF-R06a |
| default area filter 1411; type 10 choices/default duration=50 | engine code | area handler/SLD-R16/SLD-R17 |
| selected Skills values | table data | values from selected record; consumer meanings code-defined |
| mon death sentry Id=312/srvdofunc=55 | table data | [TD-DSC named text cells](../research/spec-evidence/skill-assassin-death-sentry-corpse-damage.md#named-table-cells); compiled integration DSC-Q04 |

Other vector values are synthetic inputs. Calculation requests use `(caster,skill,L)` through SCX-R11/SLD-R19, retaining SCX-R15 draws and unresolved producers. A local no-draw statement does not erase dependency draws.

## 4. Rules (execution order)

**DSC-R01a** Explosion guards/state [verified in 1.13c code (static)]. Request target preparation then obtain current target. Missing target, target==caster or zero DSC-R01b result returns 0 without state/damage/local draw. Otherwise request state 104/on 1, then update mark, then skill lookup/check. Skill identity=0 returns 0 after these requests before damage. The nonzero valid-record path continues. Preparation/lookup/update effects retain DSC-Q04; membership consumes RLI-R03/STL-R07 and their limits.

**DSC-R01b** Corpse-valid producer [unresolved]. The located predicate requires monster/dead mode 12, no overlap of active states with an excluded-state mask, and class MonStats2.corpseSel. Complete mask production/class lookup/state effects remain DSC-Q04. DSC-R01a handling of a supplied result is independently available. DST-R03b's search callback consumes the same predicate.

**DSC-R02a** Initial own-array life [verified in 1.13c code (static)]. Read corpse stat 7 `maxhp` at layer 0 from its base list's `ownStats`, even when `usesResolvedStats` is true. An absent corpse/list or invalid/missing stat definition returns 0. For a valid definition, consume RLI-R02a's own-array lookup: a missing entry returns 0 without minimum adjustment; a present entry receives RLI-R02b's conditional minimum adjustment. No fallback to `resolvedStats` or an attached list occurs, and the read changes no array or seed. Non-monsters retain this value; monsters replace it under DSC-R02b. Existing-array lookup/minimum behavior is verified; array construction/aggregation and corpse eligibility retain their producer gates. This isolated consumer does not establish that player corpses pass DSC-R01b.

**DSC-R02b** Monster replacement [verified in 1.13c code (static)]. Clear base output; selector=1 when either game selector condition is nonzero, else 0. Request MSD-R10 for corpse class/difficulty/stat 12 level, HP bounds only; failed lookup leaves cleared bounds 0. Set `maxLife=low32(low32(minHP+maxHP)<<7)`. No stored maxhp, randomized life roll or player bonus substitutes. MSD-R10 owns table guards and base arithmetic; real data retain its limits.

**DSC-R03** Endpoints [verified in 1.13c code (static)]. Evaluate aurarangecalc, then `calc1`, then `calc2` in caster context. Set min=mulDiv(maxLife,`calc1`,100), max=mulDiv(maxLife,`calc2`,100). No local endpoint draw.

**DSC-R04** Corpse roll [verified in 1.13c code (static)]. `n=low32(max-min)`; `dmg=low32(min+Rand(corpse seed,n))` under NRF-R02. Signed n<=0 adds 0 without stepping; n>0 draws once. Nonwrapping ordered endpoints give [min,max-1]. Extreme inputs use wrapped n, not an independent max<=min test.

**DSC-R05** Level scale [verified in 1.13c code (static)]. Read corpse level then caster level. If corpseLevel!=0 and signed casterLevel<corpseLevel, dmg=mulDiv(casterLevel,dmg,corpseLevel), before splitting; otherwise retain dmg. No local draw. NRF-R05's complete divisor/operand conditions apply; dmg>1048576 alone does not prove divide-first for arbitrary corpseLevel.

**DSC-R06** Percent [verified in 1.13c code (static)]. Clear combat record, evaluate `calc3`, clamp its signed result to 0..100 as pct. No local draw beyond evaluation.

**DSC-R07** Blast hand-off [verified in 1.13c code (static)]. Construct DSC-R08a record; pass sentry as area source, corpse coordinates, DSC-R09a outer radius/filter, DSC-R11a callback and extra monster-class exclusion disabled. Ignore area return; return 1 after the call even when no victim is visited.

**DSC-R08a** Split [verified in 1.13c code (static)]. If pct>0 and EType!=0, request SLD-R11 duration(caster,skill,L); route mulDiv(pct,dmg,100) through DSC-R13 with both optional outputs absent; then assign physical=mulDiv(100-pct,dmg,100). Otherwise skip duration/router and assign physical=dmg. Independent truncation can lose one raw unit; physical is not a remainder. Fields begin zero. For a selector routed to physical, the final physical assignment replaces the routed addition. No fire-only restriction is imposed locally.

**DSC-R08b** Units [verified in 1.13c code (static)]. Stat maxhp/damage are raw 8.8. Monster bounds are whole-life i32; summing then shifting 7 gives their average in 8.8 with wrapping, without whole-life rounding/saturation.

**DSC-R09a** Geometry [verified in 1.13c code (static)]. Store corpse point, full=trunc(R/2), outer=trunc(low32(R+1)/2). Area needs signed outer>0/present source room. Both nonzero corpse coordinates give that area center; either coordinate=0 replaces BOTH outer-center coordinates by sentry position. Callback physical distance still uses corpse point. Squared distance=low32(low32(dx*dx)+low32(dy*dy)); squared radii wrap; comparisons are signed <=. A callback beyond full has physical=0, element fields retained. Filter 0 becomes 1411. Use supplied adjacent rooms of sentry room, skip sentry, test distance before target validity. Nonpositive outer/missing room produces no callback.

**DSC-R09b** Actual traversal [unresolved]. Room-list/overlap/unit-chain and target-validity producers determine real callbacks (DSC-Q04). Supplied ordered candidates/predicate results establish local handling independently.

**DSC-R10a** Supplied scaling [verified in 1.13c code (static)]. Callback has no local player multiplier. Its enabled apply consumes PVP-R02a/PVP-R03/PVP-R04 and DPO-R08c with supplied hit-time scaleCondition; scale precedes mitigation. No second local scale or credit award is added.

**DSC-R10b** Actual scale/credit [unresolved]. Sentry hit-time scaleCondition and complete kill credit remain DSC-Q01; supplied condition arithmetic does not verify their producers.

**DSC-R11a** Victim copy/order [verified in 1.13c code (static)]. Each accepted candidate gets a fresh entire explosion record, with physical zeroed under DSC-R09a. Call DPO apply(game,sentry,victim,copy,stageEnable=1), then HRC reaction(game,sentry,victim,same mutable copy), then return 1. Later victims start from the original record. No queued build/new hit roll is inserted locally. Dependency effects retain their gates.

**DSC-R11b** Event effects [unresolved]. StageEnable=1 enables scale/mitigation; it is not an item-event identity. DPO-R13a/DPO-R23/DMI-R03 give conditional event placement; IEP-R07 gives traversal. Complete registered callback effects/suppression/reflection/leech/draws remain DSC-Q02. No universal proc suppression follows from argument 1.

**DSC-R12a** Supplied mitigation [verified in 1.13c code (static)]. With enabled apply and supplied guards reaching mitigation, consume DMI-R02/DMI-R03/DMI-R05..DMI-R09 after PVP scaling. Use their fire bindings/caps/reduction/absorb, adding no local cap/draw.

**DSC-R12b** Actual fire outcome [unresolved]. Actual eligibility/hit-time conditions/callbacks/resources/status/reaction retain DSC-Q03. Supplied mitigation arithmetic is independent.

**DSC-R13** Duration/router [verified in 1.13c code (static)]. Active DSC-R08a branch requests SLD-R11 before SLD-R16/SLD-R17 routing. EType=10 advances sentry seed once, selects {1,2,4,5}[newLow&3] (fire/lightning/cold/poison), defaults signed length<=0 to 50 frames. Other types have no router draw/default. This follows corpse roll and duration request, before physical assignment/area. pct=0/EType=0 skips it. Expression-owned draws can precede it; unresolved requested duration producers retain DSC-Q05.

**DSC-R14a** Conditional victim draws [verified in 1.13c code (static)]. Each accepted victim reaches reaction after its apply in traversal order. If reaction reaches HRC-R06 veto, draws use that victim's seed/conditions. A visit guarantees neither a veto draw nor a fixed count; local explosion never directly steps victim seed.

**DSC-R14b** Complete draw count [unresolved]. Expression/search/mode/item-event dependencies may draw beyond the direct fallback (DST-R05)/corpse/router sites. Complete interleaving/budget remains DSC-Q02/DSC-Q04/DST-Q03.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| DSC-R01a | State requests | verified in 1.13c code (static) | DSC-Q04 actual preparation/update | DSC-V06 |
| DSC-R01b | Corpse-valid producer | unresolved | DSC-Q04 | none |
| DSC-R02a | Stat-life input | verified in 1.13c code (static) | none for supplied stats | DSC-V08 |
| DSC-R02b | Base-life replacement | verified in 1.13c code (static) | MSD producer limits for actual data | DSC-V08 |
| DSC-R03 | Endpoints | verified in 1.13c code (static) | DSC-Q04 actual calculations | DSC-V01 |
| DSC-R04 | Wrapped range | verified in 1.13c code (static) | none for supplied endpoints | DSC-V01, DSC-V09 |
| DSC-R05 | Level scaling | verified in 1.13c code (static) | none for supplied levels | DSC-V02 |
| DSC-R06 | Clamp | verified in 1.13c code (static) | DSC-Q04 actual expression | DSC-V07 |
| DSC-R07 | Area hand-off | verified in 1.13c code (static) | DSC-Q04 actual traversal | DSC-V04, DSC-V10 |
| DSC-R08a | Independent split | verified in 1.13c code (static) | DSC-Q05 actual duration | DSC-V03, DSC-V07, DSC-V09 |
| DSC-R08b | 8.8 units | verified in 1.13c code (static) | none with supplied bounds | DSC-V08 |
| DSC-R09a | Geometry | verified in 1.13c code (static) | DSC-Q04 actual rooms/validity | DSC-V04 |
| DSC-R09b | Actual traversal | unresolved | DSC-Q04 | none |
| DSC-R10a | Supplied scaling | verified in 1.13c code (static) | DSC-Q01 actual scale condition only | DSC-V05 |
| DSC-R10b | Actual scale/credit | unresolved | DSC-Q01 | none |
| DSC-R11a | Copies/call order | verified in 1.13c code (static) | DSC-Q02 dependency effects only | DSC-V10 |
| DSC-R11b | Event effects | unresolved | DSC-Q02 | none |
| DSC-R12a | Supplied mitigation | verified in 1.13c code (static) | DSC-Q03 actual integration only | DSC-V05 |
| DSC-R12b | Actual fire outcome | unresolved | DSC-Q03 | none |
| DSC-R13 | Duration/router | verified in 1.13c code (static) | DSC-Q05 requested producers only | DSC-V09 |
| DSC-R14a | Conditional reaction draws | verified in 1.13c code (static) | HRC actual reaction gates | DSC-V10 |
| DSC-R14b | Whole-event draws | unresolved | DSC-Q02/DSC-Q04/DST-Q03 | none |

## 5. Outputs and side effects

| Effect | Rules |
|---|---|
| state 104/on 1 and update requests; active/changed membership when storage supplied | DSC-R01a, RLI-R03/STL-R07 |
| zeroed record/routed element/duration/hit class/final physical | DSC-R06/DSC-R08a/DSC-R13 |
| direct corpse/sentry seed updates | DSC-R04/DSC-R13 |
| area request, victim copies, ordered enabled apply/reaction, callback/action return 1 | DSC-R07/DSC-R09a/DSC-R11a |
| conditional downstream resource/event/state/mode/seed effects | DSC-R10..DSC-R14, named dependency contracts/gates |

No direct local resource write, client message, persistence format or independent credit award is established. Budget, delay, mode requests and the remembered corpse are DST section 5.

## 6. Branches and edge cases

Scaling skips corpseLevel=0 or casterLevel>=corpseLevel. Nonpositive wrapped range skips draw; max<=min is equivalent only without subtraction overflow. pct=0/EType=0 skips duration/router. Radius/squared overflow uses signed wrapped comparisons; a zero corpse coordinate can separate outer/physical centers. Nonzero-skill damage requires a valid record; no safe invalid-record fallback is established. DSC-R01b's producer rejects players although isolated non-monster stat consumers are testable. MSD/DMI own difficulty/expansion differences; no other-build equivalence is asserted.

## 7. Open questions

All open assessments also appear in the [gap file](../plans/gaps/skill-assassin-death-sentry-corpse-damage.md).

| Id | Status / kind | Missing evidence and exact limit |
|---|---|---|
| **DSC-Q01** | open / outside-boundary | Actual hit-time sentry scaleCondition and final credit; PVP-R02a settles supplied condition consumption, not production. None locally gates supplied factor arithmetic. |
| **DSC-Q02** | open / integration | Complete event callbacks/suppression/effects/draws through apply/reaction, consuming DPO/IEP/HRC. Ordered local calls remain available. |
| **DSC-Q03** | open / integration | Actual eligible target/apply/mitigation/resource/status/reaction result; supplied DMI arithmetic is independent. |
| **DSC-Q04** | open / integration | Preparation/target/skill/level/expression producers, excluded-state mask/class flags, room/target validity, update effects and compiled correspondence of named Skills cells. Local supplied stages remain available. |
| **DSC-Q05** | open / outside-boundary | Actual duration source/argument order is SLD-R11; unresolved expression parameter producers retain SCX gates. None locally gates supplied duration/routing. |

Gameplay observations for every behavioral rule remain pending; the three recipes are initial controls, not exhaustive branch coverage. No inability to observe the eligible 1.13c behavior has been established.

## 8. Test vectors

### 8.1 Executable vectors

Cases start independently and are synthetic. Default positions: sentry(100,100), corpse(110,100); valid room/record, nonzero skill, L=1, levels 40/40. Record fields start at 0. Seeds: S={6,666}, C={1,666}, A={9,666}, B={10,666}; no draw leaves every seed unchanged. Stats use supplied existing lists/definitions and RLI-R02a/RLI-R02b. Unless overridden, each stated stat value is present identically in `ownStats` and `resolvedStats`, `usesResolvedStats` is true, stat definitions are present with count 359, minimum adjustment is disabled, and each array contains exactly its listed entries in the foundations ordering. The initial corpse maxhp read always uses `ownStats`; later level reads select the effective array. Neither read constructs or changes an array. Local rules execute; supplied producers log ordered requests/return listed results with no unspecified writes/draws. State 104 starts inactive/unchanged with stateCount=200/storage present; membership executes RLI-R03/STL-R07, while update/preparation log only. Eligibility/room results are supplied and do not verify DSC-R01b/DSC-R09b. Expressions return listed results with no draw. Apply/reaction log only except DSC-V10 mutations; DSC-V05 executes the cited numeric shared stage. A supplied operation's effects do not verify its original producer. All vectors are not runtime-observed.

- **DSC-V01** Endpoints/exclusive roll (DSC-R03/DSC-R04; boundary: supplied-input stage)
  - Inputs: maxLife=25600, `calc1`=40/`calc2`=80, EType=0, levels 40/40; caster S, skill 312/L=1; supplied `aurarangecalc` result 7. Start at radius evaluation after the life read, then perform endpoint/roll and physical assignment. The supplied evaluator logs each reference/context and its returned value in an initially empty ordered evaluation log, without writes or draws.
  - Seeds: C={1,0x29A}; others default.
  - Draws: C mod=10240 ->2911; C={0x6AC6935F,0}.
  - Trace: min=10240/max=20480/n10240; dmg=10240+2911=13151.
  - Outputs: physical=13151/other fields=0; ordinary range[10240,20479], excluded 20480; ordered evaluation log [(S,312,1,aurarangecalc,7),(S,312,1,`calc1`,40),(S,312,1,`calc2`,80)]. This asserts request order and context, not only a call count.
  - (a) Inputs: independently maxLife=131072/`calc1`=10/`calc2`=30; stop after endpoints. Draws: none. Trace: trunc(131072/100)=1310. Outputs: min=13100/max=39300/all seeds unchanged.
  - Discriminates: Inclusive maximum and wide-product endpoint results 13107/39321 fail assertions; swapping either endpoint request or evaluating radius later fails the ordered log.
  - runtime_observed: no

- **DSC-V02** Operand-ordered scale (DSC-R05; boundary: isolated rule)
  - Inputs: casterLevel=30/corpseLevel=40/dmg=1000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: mulDiv(30,1000,40)=750.
  - Outputs: dmg=750; seeds unchanged.
  - (a) Inputs: dmg=2097152/levels 30/40. Draws: none. Trace: trunc(2097152/40)*30. Outputs:1572840, rejecting 1572864.
  - (b) Inputs: dmg=2097152/levels 40/40. Draws: none. Outputs:2097152 unchanged.
  - (c) Inputs: dmg=1000/casterLevel=30/corpseLevel=0. Draws: none. Outputs:1000 unchanged.
  - (d) Inputs: dmg=2097152/casterLevel=30/corpseLevel=200000. Draws: none. Trace: divisor 200000>2097152>>4. Outputs:314 from exact wide-product branch; unconditional divide-first gives 300.
  - (e) Inputs: independently dmg=1000/casterLevel=50/corpseLevel=40. Draws: none. Outputs: dmg=1000 unchanged; seeds unchanged.
  - Discriminates: Reversing operands or ignoring the divisor threshold changes exact results; scaling a higher caster level fails (e).
  - runtime_observed: no

- **DSC-V03** Independent split (DSC-R08a; boundary: supplied-input stage)
  - Inputs: dmg=101/pct=50/EType=1/duration=0/empty record.
  - Seeds: no draw.
  - Draws: none.
  - Trace: duration request then router fire=50/hitClass=32; physical=50.
  - Outputs: fire=50/physical=50/sum=100, one unit lost; duration requested for (caster,skill,L); router optional outputs absent.
  - (a) Inputs: dmg=13151/pct=25/EType=1/duration=0. Draws: none. Outputs: fire=3287/physical=9863/component sum=13150/hitClass=32.
  - (b) Inputs: dmg=101/pct=50/EType=3/duration=0. Draws: none. Outputs: magic=50/physical=50/hitClass=0.
  - (c) Inputs: dmg=101/pct=50/EType=13/duration=0/empty record; an initially empty ordered routing log records the semantic router request and the record immediately after that operation, then the final physical assignment. The router executes SLD-R16 rather than a no-effect substitute. Draws: none. Trace: request routing type 13/damage 50/length 0 with optional outputs absent; router adds 50 to physical 0; final physical assignment writes 50. Outputs: ordered log [route(type 13,damage 50,length 0,optional outputs absent), post-route physical 50/other damage 0/hitClass 0, final physical assignment 50]; final physical=50/other damage=0/hitClass=0. This checks the intermediate routed write even though the final assignment has the same value.
  - Discriminates: Remainder physical=51 fails default; treating all nonzero selectors as fire fails(b/c); skipping type-13 routing leaves no route request or intermediate physical 50 and fails (c).
  - runtime_observed: no

- **DSC-V04** Rings/rooms/center exception (DSC-R07/DSC-R09a; boundary: supplied-input stage)
  - Inputs: R=7/original physical=1000/fire=500/filter 0; supplied sentry-adjacent ordered candidates A(113,100), B(113,101), C(114,100), D(114,101), validity=1; E(110,100) outside supplied adjacent rooms.
  - Seeds: no draw.
  - Draws: none.
  - Trace: full=3/outer=4; squared thresholds=9/16; candidate distances=9/10/16/17; filter 1411.
  - Outputs: callback A physical=1000/fire=500, B physical=0/fire=500, C physical=0/fire=500; none D/E; other copy fields=0/action 1.
  - (a) Inputs: independent R=6/same candidates. Draws: none. Outputs: full=3/outer=3; only A physical=1000/fire=500/action 1.
  - (b) Inputs: corpse(0,100)/sentry(100,100)/R=7/candidateA(103,100)/validity=1. Draws: none. Trace: outer center(100,100), distance=9; callback corpse distance=10609. Outputs: A physical=0/fire=500/action 1.
  - (c) Inputs: independent R=0/valid room. Draws: none. Outputs: outer=0/no callbacks/action 1.
  - (d) Inputs: isolated blast hand-off with original physical=1000/fire=500, corpse(110,100), sentry(100,100), R=7/filter=0; the supplied area operation logs its request, invokes no callback, makes no writes/draws and returns 0. Draws: none. Trace: ignore supplied area return 0. Outputs: one area request with source sentry/point(110,100)/outer 4/filter 0/extra exclusion disabled, unchanged original record and seeds, no apply/reaction, final action return 1. The area handler, when executed locally, would default filter 0 under DSC-R09a; this hand-off-only case asserts the caller input.
  - (e) Inputs: isolated geometry/physical-filter stage, corpse(1,100), supplied candidate A(65535,100), R=7, outer validity result 1, original physical=1000/fire=500; supplied room-overlap/chain permits the candidate, without claiming real room production. Draws: none. Trace: dx=65534, dy=0; low32(dx*dx)=4294705156, signed -262140; signed -262140<=16 and <=9. Outputs: local outer test accepts A and physical remains 1000/fire 500, all other fields 0; no seed change. An unbounded squared-distance comparison would reject A.
  - (f) Inputs: independent local radius/area stage with R=2147483647, valid room and original physical=1000/fire=500. Draws: none. Trace: full=1073741823; low32(R+1)=-2147483648; outer=-1073741824. Outputs: no room-enumeration/callback request, action return 1, original record and seeds unchanged.
  - Discriminates: One radius, corpse-centered room enumeration or retaining a zero-coordinate outer center fails; forwarding area return 0 fails (d), and unbounded geometry or addition fails (e/f).
  - runtime_observed: no

- **DSC-V05** Supplied PvP/fire caps (DSC-R10a/DSC-R12a; boundary: supplied-input shared stage)
  - Inputs: distinct non-hireling monster sentry/player defender, scaleCondition=false, stage reached; fire=25600/physical=0, expansion Normal penalty=0/regularDefender=0; defender stat 39 fireresist=75/stat 40 maxfireresist=0/stat 35 magicdamagereduction=0/stat 142 item_absorbfire_percent=0/stat 143 item_absorbfire=0; attacker stat 333 passive_fire_pierce=0; bypass=0/event=11 list empty; all other stats/record fields=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: factor=100/cap=75; mulDiv(25,25600,100)=6400; absorb=0.
  - Outputs: fire=6400/absorbHeal=0; no resource/reaction assertion.
  - (a) Inputs: independent same inputs except stat 39=100. Draws: none. Outputs: effectiveResistance=75/fire=6400/absorbHeal=0.
  - (b) Inputs: independent regular-monster defender/fire=25600/stat 39=100/all other mitigation inputs=0/P100. Draws: none. Outputs: effectiveResistance=100/fire=0/absorbHeal=0.
  - (c) Inputs: independent player defender/stat 39=0/scaleCondition=true. Draws: none. Outputs: factor=17/fire=4352/absorbHeal=0; clear condition gives 25600.
  - (d) Inputs: independent supplied shared scale/mitigation stage, distinct non-hireling monster sentry/player defender, scaleCondition=true, fire=71/physical=0; expansion Normal penalty=0/regularDefender=0, defender stat 39 fireresist=75/stat 40 maxfireresist=0/stat 35 magicdamagereduction=0/stat 142 item_absorbfire_percent=0/stat 143 item_absorbfire=0; attacker stat 333 passive_fire_pierce=0; bypass=0/event 11 list empty, all other record/stat fields 0. Draws: none. Trace: factor 17 first gives mulDiv(17,71,100)=12; resistance 75 next gives mulDiv(25,12,100)=3; absorb remains 0. Outputs: pre-mitigation fire 12, final fire 3/absorbHeal 0, all seeds unchanged; no resource/reaction assertion. Reversing stages gives trunc(71*25/100)=17 then trunc(17*17/100)=2 and fails.
  - Discriminates: Uncapped player immunity fails(a); always=17 for sentries fails default; mitigation-before-scaling fails (d). Actual hit-time condition producer is not verified.
  - runtime_observed: no

- **DSC-V06** State-before-skill (DSC-R01a/DST-R06a; boundary: supplied-input stage)
  - Inputs: prepared distinct target/predicate=1/state 104 inactive/unchanged/action skill=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: preparation/target/predicate/state 104-on/update/skill-check.
  - Outputs: state 104 active=true/changed=true/no damage/area. Return 0.
  - (a) Inputs: independently target==caster/state 104 inactive. Draws: none. Outputs: return 0/state inactive/no predicate/state/update/damage.
  - (b) Inputs: independently distinct target/predicate=0/state 104 inactive. Draws: none. Outputs: return 0/state inactive/no state/update/damage.
  - (c) Inputs: independently absent target. Draws: none. Outputs: return 0/no predicate/state/update/damage.
  - Discriminates: Skill-before-state or state-on-reject violates request order/effects.
  - runtime_observed: no

- **DSC-V07** Clamp/disabled element (DSC-R06/DSC-R08a; boundary: supplied-input stage)
  - Inputs: dmg=13151/`calc3`=150/EType=1/duration=0/empty record.
  - Seeds: no draw.
  - Draws: none.
  - Trace: pct=100/duration/router/full element/physical=0.
  - Outputs: fire=13151/physical=0/hitClass=32; duration precedes routing.
  - (a) Inputs: independently `calc3`=-5/EType=1. Draws: none. Outputs: pct=0/physical=13151/other fields=0/no duration/router.
  - (b) Inputs: independently `calc3`=50/EType=0. Draws: none. Outputs: pct=50/physical=13151/other fields=0/no duration/router.
  - Discriminates: Routing before checking disabled element produces an extra request.
  - runtime_observed: no

- **DSC-V08** Maximum-life replacement (DSC-R02a/DSC-R02b/DSC-R08b; boundary: supplied-input stage)
  - Inputs: isolated player-kind consumer (type 0), base list present, `usesResolvedStats=true`, `ownStats=[(7,0,25600)]`, `resolvedStats=[(7,0,51200)]`; stat 7 definition present/count 359, minimum adjustment disabled.
  - Seeds: no draw.
  - Draws: none.
  - Trace: own-array stat 7 read returns 25600 despite resolved value 51200; kind!=1 retains it.
  - Outputs: maxLife=25600, both arrays unchanged, no base request or draw. Player corpse eligibility is not asserted.
  - (a) Inputs: independent monster class 20/difficulty=1/stat 12=40/stored stat 7=25600/selectorByte=0/selectorFlag=1/supplied MSD bounds=10/20. Draws: none. Trace:(10+20)<<7. Outputs: HP-only request(class 20,selector=1,difficulty=1,level=40)/maxLife=3840 (15.0 life).
  - (b) Inputs: independent monster kind 1, class 20/difficulty=1/stat 12=40/stored own stat 7=25600/selectorByte=1/selectorFlag=0/supplied MSD bounds=10/20. Draws: none. Outputs: HP-only request(class 20,selector=1,difficulty=1,level=40), maxLife=3840.
  - (c) Inputs: independent monster kind 1, class 20/difficulty=1/stat 12=40/stored own stat 7=25600/selectorByte=0/selectorFlag=0/failed base lookup/cleared output. Draws: none. Outputs: HP-only request(class 20,selector=0,difficulty=1,level=40), maxLife=0.
  - (d) Inputs: independent isolated player-kind read (type 0), base list present, `usesResolvedStats=true`, associated owner type player 0, `ownStats=[(7,0,-1)]`, `resolvedStats=[(7,0,51200)]`; definition count 359, stat 7 definition `fMin=true`, `MinAccr=1`, `ValShift=8`. Draws: none. Trace: present own value -1 is less than unshifted minimum 1; conditional replacement is low32(1<<8). Outputs: maxLife=256, both arrays unchanged, no base request.
  - (e) Inputs: independent isolated player-kind read (type 0), base list present, `usesResolvedStats=true`, associated owner type player 0, `ownStats=[]`, `resolvedStats=[(7,0,51200)]`; definition count 359, stat 7 definition `fMin=true`, `MinAccr=1`, `ValShift=8`. Draws: none. Trace: missing own entry returns before minimum adjustment. Outputs: maxLife=0, both arrays unchanged, no base request.
  - (f) Inputs: isolated initial-life read with corpse absent. Draws: none. Outputs: getter value 0, no base request, all seeds unchanged; this does not claim the explosion entry accepts an absent corpse.
  - (g) Inputs: isolated player-kind read (type 0), corpse present, base list absent, stat 7 definition present/count 359. Draws: none. Outputs: maxLife=0/no base request/all seeds unchanged.
  - (h) Inputs: isolated player-kind read (type 0), corpse/base list present, `usesResolvedStats=true`, `ownStats=[(7,0,25600)]`, `resolvedStats=[(7,0,51200)]`, definition count 359 but stat 7 definition unavailable. Draws: none. Outputs: maxLife=0, arrays unchanged, no base request, seeds unchanged.
  - Discriminates: Selecting the resolved value gives 51200 instead of 25600 in the default; skipping own-array minimum adjustment fails (d), and applying minimum adjustment to a missing own entry fails (e). Stored monster maxhp gives 25600 rather than 3840 in (a); ignoring one selector condition or adding roll/player bonus also changes assertions.
  - runtime_observed: no

- **DSC-V09** Range/router order (DSC-R04/DSC-R13/DSC-R08a; boundary: supplied-input stage)
  - Inputs: supplied endpoints=10240/20480/levels 40/40/pct=50/EType=10; caster S, skill 312; SLD-R11 ELen=0/increments=0/empty synergy/L=1/valid record; stop before area. For this primary case, an initially empty shared timeline records range requests/returns, actual seed steps, duration requests/returns, router requests/record-at-completion and final physical assignment in one ordered sequence. Timeline observation adds no writes or draws, and local numeric/RNG/router operations execute their cited contracts.
  - Seeds: C={1,0x29A}, S={1,0x29A}; victims default.
  - Draws: C mod=10240->2911/C{0x6AC6935F,0}; after duration request S mask 4->3/S{0x6AC6935F,0}.
  - Trace: dmg=13151/duration=0/element=6575/type 5 poison/default length 50/physical=6575.
  - Outputs: poison=6575/poisonLength=50/physical=6575/hitClass=80/other fields=0; both seeds advanced. Exact shared timeline: [range request(C,10240), draw(C,mod 10240,result 2911,final seed {0x6AC6935F,0}), range return(2911), duration request(S,312,1), duration return(0), route request(S,type 10,damage 6575,length 0,optional outputs absent), draw(S,mask 4,result 3,final seed {0x6AC6935F,0}), route completed(record poison 6575,poisonLength 50,hitClass 80), final physical assignment(6575)]. No area request occurs within this stopped boundary.
  - (a) Inputs: independently min=max=10240/pct=50/EType=1/duration=0/same entry seeds. Draws: none. Outputs: dmg=10240/fire=5120/physical=5120/hitClass=32/S,C unchanged/local count 0.
  - (b) Inputs: independently endpoints=10240/20480/pct=50/EType=1/duration=0/same entry seeds. Draws: C mod=10240->2911/C{0x6AC6935F,0}. Outputs: dmg=13151/fire=6575/physical=6575/hitClass=32/S unchanged{1,0x29A}/local count 1.
  - (c) Inputs: independently min=2147483647/max-2147483648/pct=0/EType=10/same entry seeds. Draws: C mod=1->0/C{0x6AC6935F,0}. Outputs: wrapped n1/dmg=2147483647/physical=2147483647/S unchanged/no duration/router.
  - (d) Inputs: independently min=max=10240/pct=50/EType=10; SLD-R11 ELen=12/increments=0/empty synergy/L=1, yielding duration=12; same entry seeds. Draws: S mask 4->3/S{0x6AC6935F,0}. Outputs: poison=5120/poisonLength=12/physical=5120/hitClass=80/C unchanged{1,0x29A}/local count 1.
  - (e) Inputs: isolated range/split stage with supplied min=-2147483648/max=2147483647, levels 40/40, pct=0/EType=0; entry C={1,0x29A}, S={1,0x29A}. Draws: none. Trace: low32(max-min)=-1; range helper adds 0 without stepping. Outputs: n=-1/dmg=-2147483648/physical=-2147483648/other fields 0, C/S unchanged, no duration/router.
  - Discriminates: Wrong seed owner, general max<=min no-draw test, missing duration context or unconditional default 50 fails; stepping a negative wrapped range fails (e). Reordering the corpse draw, duration request or sentry router draw fails the primary shared timeline even when independent seeds would leave the same final numeric fields.
  - runtime_observed: no

- **DSC-V10** Fresh copies/post-apply reaction (DSC-R07/DSC-R11a/DSC-R14a; boundary: supplied-input callback stage)
  - Inputs: R=7/original physical=1000/fire=500/hitClass=32/other fields=0; accepted A squared=16 then B squared=4. Supplied apply logs input, writes A physical=7/fire=9/total=16/result=0 and B physical=11/fire=13/total=24/result=0; reaction logs same complete record/no writes/draws.
  - Seeds: no draw; A{9,666}/B{10,666}/S{6,666}/C{1,666}.
  - Draws: none locally or in supplied operations; no veto-reached assertion.
  - Trace: copyA/zero physical/applyA/reactionA; fresh copyB/retain physical/applyB/reactionB.
  - Outputs: ordered applyA(input physical=0/fire=500/hitClass=32/others=0,enable 1), reactionA(physical=7/fire=9/total=16/result=0/hitClass=32/others=0), applyB(input physical=1000/fire=500/hitClass=32/others=0,enable 1), reactionB(physical=11/fire=13/total=24/result=0/hitClass=32/others=0). Both callbacks=1/action 1; original unchanged=1000/500/32/others=0; final A copy=7/9/16/result=0/class 32/others=0; final B11/13/24/result=0/class 32/others=0; all seeds unchanged.
  - Discriminates: Reusing A's mutated record contaminates B; reacting on original loses apply writes. Supplied effects verify only copy/call boundary, not dependency interiors.
  - runtime_observed: no

### 8.2 Observation vectors

All pending. [Tool-neutral procedures](../research/spec-evidence/skill-assassin-death-sentry-corpse-damage.md#observation-procedures) give controls and reading points.

- **DSC-O01** Repeat 1000 blasts at one corpse type/level and sentry rank; compare pre-mitigation values to[min,max-1] on a nonwrapping range. Absence of the maximum alone is not proof of exclusion.
- **DSC-O02** Compare fire resistance=75/100 with matched defender classification/max-resist/absorb; DMI distinguishes player cap and monster immunity.
- **DSC-O03** Controlled registered thorns/leech/events: record enabled apply/callback order/mutable record/resources/reaction; conditional placement consumes DPO/IEP, callback interiors remain gated.

## 9. Acceptance and dependencies

Partial acceptance reproduces every verified assertion in DSC-V01..DSC-V10 within its supplied boundary. TD-DSC proves named effective text cells only. Supplied apply/reaction/area operations do not verify producers. DSC-R01b/DSC-R09b/DSC-R10b/DSC-R11b/DSC-R12b/DSC-R14b retain their exact Id-map gates. Full reproduction additionally needs DSC-Q01..DSC-Q05 resolution and gameplay observations of every behavioral rule. No partial pass establishes parity.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md) NRF-R01/NRF-R02 draws, NRF-R04a corpse ownership exception, NRF-R05 arithmetic, NRF-R08a units, NRF-R10a/NRF-R10b positions, RLI-R03 state requests, RLI-R02a/RLI-R02b stats; [fields](core-game-data-fields.md) GDA-R09a/GDA-R11a/GDF-R02/GDF-R03/GDF-R06a; [expressions](skill-calc-expressions.md) SCX-R11/SCX-R15/SCX-R24 and requested producer gates; [leveled values](skill-leveled-damage-tohit-and-elements.md) SLD-R11 duration/SLD-R16/SLD-R17 routing/SLD-R19 front end; [monster scaling](monster-scaling-by-players-and-difficulty.md) MSD-R10 bounds; [lifecycle](state-timed-state-lifecycle.md) STL-R07 membership/STL-R07a/RLI-Q02 limits; [PvP](combat-pvp-damage-and-kill-credit.md) PVP-R02a/PVP-R03/PVP-R04 scaling/PVP-R02b producer gate; [mitigation](combat-damage-mitigation.md) DMI-R02..DMI-R09; [pipeline](combat-damage-pipeline-order.md) DPO-R08c/DPO-R13a/DPO-R23 and downstream gates; [recovery](combat-hit-recovery.md) HRC-R06 conditional victim draws/reaction; [dispatch](skill-cast-start-and-do-dispatch.md) SCD-R06/SCD-R12 action dispatch retaining mode/animation gates. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** [Death Sentry corpse selection](skill-assassin-death-sentry-corpse-selection.md) DST-R03b consumes the gated DSC-R01b corpse-valid producer. No other downstream behavioral consumer is established. Supporting citations are not extra prerequisites to supplied arithmetic.

## 10. Related specs

- [Death Sentry corpse selection](skill-assassin-death-sentry-corpse-selection.md): the think that chooses the corpse and requests this action.
- [Blade Sentinel and Fury](skill-assassin-blade-sentinel-and-fury.md): sibling trap consumers.
- [Shield block and avoidance](combat-shield-block-and-avoidance.md): shared hit resolution outside local construction.
- [Item events](combat-item-event-triggers-and-procs.md): registered downstream callbacks and their limits.

## Appendix A. Provenance

Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies original D2Game/D2Common artifacts. Compatible disassembly/RNG/field readings retain 2026-10-12 dates/limits; targeted instructions, byte-equivalent predicate/distance readings and original-image/table checks date 2026-10-15. The [companion](../research/spec-evidence/skill-assassin-death-sentry-corpse-damage.md) records dispatch, mappings, corrections, named text cells and profile limits. No gameplay observation or other-build analysis was performed.
