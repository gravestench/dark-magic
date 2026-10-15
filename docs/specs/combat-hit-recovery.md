# Hit recovery and Faster Hit Recovery

**Description**

A connected hit can make a character or monster briefly recoil. Damage, the target's states and competing reactions determine whether that happens. Faster Hit Recovery changes the recovery animation's rate; block and knockback use their own reaction and rate rules.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: HRC-R, HRC-Q, HRC-V, HRC-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-hit-recovery.md)
Open evidence gaps: [gap file](../plans/gaps/combat-hit-recovery.md)

## 1. Scope

This contract covers get-hit requests, reaction selection, damage/state veto, player interruption, monster get-hit entry, recovery rate, and adjacent block/knockback rate and block-cooldown decisions. Get-hit (GH) is player mode 4 or monster mode 3. Hirelings and summons reach the monster branch when supplied as monster units. Damage production/application belongs to [DPO](combat-damage-pipeline-order.md), block/avoidance rolls to [SBA](combat-shield-block-and-avoidance.md), knockback chance/motion to [KNB](combat-knockback.md), and state production to [CSA](combat-status-application.md). Completion, animation selection and unresolved mode effects retain explicit gates within this boundary.

## 2. Trigger and ownership

D2Game's server combat-result consumer runs once per reaction call after damage application (DPO-R26). Melee and missile resolution supply flags through SBA-R13/SBA-R14/SBA-R20. The consumer orders reactions and invokes player/monster mode handling; D2Common owns shared stat/state/mode operations and animation rates. The monster mode table selects separate GH start, tick and end handlers. Exact server completion and client prediction remain HRC-Q02/HRC-Q05. No message identity or presentation timing is established. Full original-binary chains are in the companion.

## 3. Inputs and data

Use NRF-R01/NRF-R02 for seed/draw, NRF-R06a/NRF-R06b/NRF-R07 for arithmetic, NRF-R12a for states and NRF-R12b/NRF-R12c for selected-array/minimum stat reads. Unqualified integer arithmetic is wrapping i32; signed division truncates toward zero. An explicitly unsigned product is reinterpreted as u32 before division. Own-list reads are named separately. A seed is `{lo,hi}` of u32 words.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| R | DPO/NRF damage result | u16 | hit 1, lethal 2, GH 4, KB 8, block 16/32768, avoidance 128/256/512, critical 8192, soft 16384 |
| total, poison, cold, fire, lightning | Applied-damage record | i32 raw 8.8 life | Veto and class selection |
| hitClass k; synthesis suppression | NRF damage-field contract | i32; u8 | Entire class retained; neutral nonzero suppression skips synthesis |
| element cursor | Shared selector state | u8 | Cyclic component start; increment wraps 256 |
| defender/source, room, level | Logical associations | Nullable units/room; i32 level | Source record/permission in town gate |
| kind, mode, class, base | GDF-R05/GDF-R14/GDF-R15/GDF-R17 | i32; BaseId i16 | Player 0, monster 1; dead guards and exceptions |
| requestedMode, context | Request supplied to the local player decision | i32 values | requestedMode is distinct from the defender's current mode; zero requested mode or zero context selects the stated early allowance |
| availability | MonStats2.mGH/mKB, mode 6; GDF-R06 | Boolean | Monster gates and fallback |
| town permission | Source MonStats.inTown; GDF-R18 | Boolean | Permission in fixed level set |
| missile selectors | Missiles.GetHit/SoftHit/KnockBack/HitClass; GDF-R07 | Booleans, u8 chance/class | Request/rewrite and initial class |
| supplied block/avoidance code | SBA-R14 block/avoidance-stage result | i32 semantic result code | Low five bits determine whether HRC-R01c rewrites before the final SBA result mapping; distinct from R |
| M | Layer 0 stat 7 maxhp | i32 raw 8.8 | Selected/minimum max life |
| FHR, FBR, last block | Layer 0 stats 99 item_fastergethitrate, 102 item_fasterblockrate, 95 lastblockframe | i32 | Rates and cooldown |
| concentration value | State 42 list's own stat 164 skill_concentration | i32 | Roll threshold; missing list 0 |
| health sync value | Layer 0 stat 352 last_sent_hp_pct | i32 | Plain-hit monster comparison |
| states / stun list | 1 freeze, 15 concentrate, 21 stunned, 42 concentration, 54 uninterruptable, 65 dodge, 66 avoid, 68 evade, 92 death_delay, 101 holyshield | Active predicates / matching stun list / requests | State 92 name only; death/timer semantics require separate evidence |
| skill context | Current/passed skill; Skills.interrupt/srvdofunc; GDF-R08 | Optional identities, boolean, u16 function index | Player decision |
| frame | Game frame | i32 | Cooldown subtraction |
| animation | Effective kind/mode, base rate B, count, accumulator, total, condition | i32; stored rate i16 | Actual selected inputs/remap gated |
| monster KB base | Supplied signed record word | i16 widened | Loader-column identity GDF-Q01 |
| dependency operations | State/mode/update/event/preparation | Ordered requests | Effects and gates in section 9 |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Stat/state names above | Effective ItemStatCost.txt ID/Stat and States.txt id/state; [named-cell record](../research/spec-evidence/combat-hit-recovery.md#named-table-cells) | Names from data; numeric consumers fixed in code |
| Modes, missile/monster/skill selectors | GDF-R06/GDF-R07/GDF-R08/GDF-R14/GDF-R15/GDF-R18 | Loader-defined fields consumed by engine handlers |
| Divisors, masks, 256 floor, cooldown 15, block bases, player KB 213, path speed 4096 | Engine code | Fixed by original code |
| Diminishing A120, FHR index 1/FBR index 3 | Engine table, NRF-R07 | Fixed table selection |
| Animation B/count/start/remap | Animation data and state selection | Actual GH record unresolved, HRC-Q04 |

Named text cells alone are data-verified; compiled runtime correspondence and custom remapping are not established by those reads.

## 4. Rules (execution order)

**HRC-R01** Request production stages.

- **HRC-R01a** Melee request [verified in 1.13c code (static)]. Consume SBA-R13's post-block/avoid result: retained hit 1 and state 54 absent ORs GH 4. Otherwise preserve R. No direct draw in this OR stage; preceding resolution keeps SBA's draws.
- **HRC-R01b** Missile request [verified in 1.13c code (static)]. Consume SBA-R20: OR hit 1. Within the state-54-absent branch, if GetHit is true OR 4; otherwise, if SoftHit is true OR 16384. State 54 present skips both additions and missile KB chance; existing bits persist. KB production is KNB-R02a/KNB-R02b, not a complete-hit zero-draw claim.
- **HRC-R01c** Missile rewrite [verified in 1.13c code (static)]. Consume SBA-R14: a supplied block/avoid code with any bits 0..4, GetHit false and KnockBack 0 clears hit 1, ORs soft 16384 and discards those code bits before final mapping. Other R bits persist. No direct draw here.
- **HRC-R01d** Other producers [unresolved]. Additional GH-request producers remain HRC-Q03, outside the supplied-R consumer.

**HRC-R02** Entry eligibility.

- **HRC-R02a** Kind/dead guards [verified in 1.13c code (static)]. After preceding town/class/state stages, ordinary reactions accept only players/monsters; monster modes 0/12 return. No local hireling/summon/NPC discriminator.
- **HRC-R02b** Town [verified in 1.13c code (static)]. Consume KNB-R03b before class/state dispatch: present target room with level in `{1,40,75,103,109}` requires monster source with valid MonStats and inTown true; otherwise return. Missing room/outside set continues. Predicate uses unsigned wrapped level-1<=108 then fixed membership. No direct draw.

**HRC-R03** Protected reaction.

- **HRC-R03a** State 54 [verified in 1.13c code (static)]. After town and connected-hit class processing, active 54 bypasses ordinary dispatch. Lethal 2 requests `SetState(defender,92,1)` through NRF-R12a/STL-R07; otherwise no state request. No GH/KB/block/avoidance mode request follows. No direct draw. Routing/membership is known; disguise/update execution retains HRC-Q10. Ordinary death-mode handling is not asserted.
- **HRC-R03b** Monster wrapper guard [unresolved]. A second state 54 test invokes assertion/failure and additional handling in the monster mode wrapper. Complete semantics are HRC-Q06; no silent-refusal default is supplied.

**HRC-R04** Priority.

- **HRC-R04a** Selector [verified in 1.13c code (static)]. Live monster: first consume KNB-R06a unavailable-KB fallback (replace 8 with 4, retain other flags), then KNB-R06c's valid base 78 non-frozen GH-to-KB conversion. Select one terminal branch: lethal > KB > block 16 > GH > soft > plain. Monster block selects reaction synchronization when soft present, class 243/333/705, or mode 6 unavailable; otherwise mode 6. Player: avoidance 128, 256, 512 in that order, then block 16/32768, lethal, KB, GH, soft. Encountering an avoidance flag owns that branch even if its state/skill checks refuse; no fallthrough. Priority draws nothing directly; selected local rules execute normally. External terminal effects retain their gates.
- **HRC-R04b** Avoidance effects [unresolved]. States 65/66/68 and skill data control player avoidance requests. Exact skill-derived inputs/eligibility/start effects are HRC-Q06.

**HRC-R05** Stun/freeze [verified in 1.13c code (static)]. At selected GH, a matching state-21 list found through NRF-R12a bypasses HRC-R06 and requests GH. Otherwise execute HRC-R06; veto selects HRC-R08. Thus a matching stun list with active freeze skips the freeze veto here. Active state 21 alone does not supply that list. No direct draw beyond executed local/dependency operations.

**HRC-R06** Damage veto [verified in 1.13c code (static)]. Return veto/allow without changing mode. In order: state 1 vetoes; nonzero poison==total vetoes; signed total<256 vetoes. Read selected/minimum M stat 7. Entire i32 k selects divisor d:5=>64; 4/8=>32; 2/6/10/11=>8; else 16. Compute signed truncating `T1=M/d,T2=M/(d/2),T4=M/(d/4)`. Total<T1 vetoes without draw. Total<T2 draws defender RandMask(2), zero vetoes immediately. Total<T4 draws defender RandMask(4), zero vetoes. Finally monster without mGH vetoes; others allow. Each draw steps NRF-R01 then masks the new low word per NRF-R02. With d=16, ordinary positive M, the derived pass fractions before monster availability are 0 below M/16, 3/8 through below M/8, 3/4 through below M/4, then 1. A failing first draw skips the second. Negative/zero M uses the same comparisons, without a life floor.

**HRC-R07** Class preparation.

- **HRC-R07a** Synthesis/store [verified in 1.13c code (static)]. On hit 1, if suppression 0 and `(k&240)==0`: read index cursor&3, increment cursor modulo 256, visit four consecutive cyclic entries cold/fire/lightning/poison, choose first signed-positive component's code 48/32/64/80. None positive gives 16 if critical 8192, else 0. Substitute 13 for source k=0; otherwise retain entire k; OR chosen code and store class. Every hit 1 stores resulting class as defender lastHitClass, even when synthesis skipped. Non-hit writes neither. No RNG step; cursor advances even without an element. Missile initial class consumes zero-extended u8 Missiles.HitClass (invalid missile/class/record returns 0); full upstream caller eligibility retains SBA's gates. Nonzero upper bits stay and select HRC-R06's default divisor.
- **HRC-R07b** Melee class producer [unresolved]. HRC-Q03; no candidate table column is assumed to feed k. Supplied k suffices locally.

**HRC-R08** Soft/plain updates [unresolved]. Veto/soft requests update without GH start. Monster handling also requests reaction synchronization and a hit event. Its plain-hit branch requires positive total and a current-life-ratio difference from stat 352 exceeding 4; player hit-only handling returns. Exact ratio/update/event effects and observable timing are HRC-Q06.

**HRC-R09** Monster GH.

- **HRC-R09a** Start [verified in 1.13c code (static)]. At supplied start-handler entry, missing unit/mode 0/mode 12 returns 1 without change; missing mGH returns 0. Otherwise request entry preparation, then mode 3 through NRF-R14a, including its verified mode and modeChanged writes, then return 1. Actual animation preparation/reset inputs and wrapper integration retain NRF-Q17/HRC-Q06. No direct draw in this start decision; complete event draws are not asserted. GH table start/tick/end identities are established.
- **HRC-R09b** Completion [unresolved]. Selected tick traverses an optional associated chain and conditionally calls a helper; no local animation-end predicate. Shared end handler branches by class. Complete server GH completion is HRC-Q02.

**HRC-R10** Player GH.

- **HRC-R10a** Local decision [verified in 1.13c code (static)]. Missing defender, state 54, or current mode 0/17 refuses. After these guards, requested mode 0 or zero context allows. For requested mode 4 with nonzero context, absent current skill or unavailable skill definition allows. Passed skill identical to current skill with srvdofunc 67/76 allows before interrupt checks. Otherwise interrupt false refuses, except current mode 1 requests neutral then allows. Interrupttrue with state 42 runs HRC-R10b; roll<own-list164 refuses with that same mode 1 exception. Then state 15 refuses with the same exception; otherwise allow. Neutral execution, outer validator and completed GH remain HRC-Q06. Missing-skill allowance does not establish idle/walking equivalence. Apart from the stated early requested-mode-0/zero-context allowance, branches for other requested modes are outside this verified GH decision slice.
- **HRC-R10b** Draw [verified in 1.13c code (static)]. After earlier fast allowances, interrupt true+state 42 makes exactly one defender Rand(100), before state 15. Others skip. Own-list164 is uncapped here.
- **HRC-R10c** Entry [unresolved]. Successful outer handling requests neutral entry preparation, mode 4, additional unit handling, current-skill clearing, and further action/path handling in that order. Preparation is not proved removal of all timed events. Actual effects/return handling/cancellation are HRC-Q06.

**HRC-R11** Block cooldown [verified in 1.13c code (static)]. Selected player block with soft 16384 returns without either request. Otherwise wrapped `delta=frame-lastblockframe`, signed `threshold=trunc(FBR/8)+15`; only delta>threshold requests mode 9, then requests a layer-0 stat-95 write carrying the current frame. The stat-write request follows the mode request irrespective of its returned value. This verified boundary establishes the two requests and their arguments/order. Actual mode start and the stat writer's persisted effect remain HRC-Q06. No direct draw; monster block has no cooldown here.

**HRC-R12** Recovery rate.

- **HRC-R12a** Formula [verified in 1.13c code (static)]. Effective player mode 4/monster mode 3 consumes NRF-R07 index 1, A120, stat 99. `D=0` if F=0, else `trunc(wrap32(120*F)/wrap32(120+F))`. Compute `p=wrap32((D+50)*B)`, reinterpret u32,`q=floor(p/100)` under NRF-R06b. Store min(q, 32767), including 0. No generic 15..175 clamp/minimum 1. Denominator 0/division overflow faults under shared arithmetic. No draw. Negative products saturate in ordinary no-overflow examples; general wrapping cases follow p rather than a sign shortcut.
- **HRC-R12b** Effective remap [unresolved]. Ordered state transformations can alter kind/mode before rate dispatch. Actual GH consequences HRC-Q04; supplied effective inputs support arithmetic independently.

**HRC-R13** Animation progress.

- **HRC-R13a** Duration [unresolved]. For positive constant rate and unwrapped `acc0`, a next-step predicate mathematically yields ceil((total-acc0)/rate). Actual GH initialization/stepping/record selection HRC-Q02/HRC-Q04; no per-class duration is established.
- **HRC-R13b** Selected values [unresolved]. GH record, B, frames, start and breakpoints HRC-Q04. File presence does not supply them.
- **HRC-R13c** Server caller [unresolved]. Actual GH accumulator driver/end order HRC-Q02.
- **HRC-R13d** Isolated primitives [verified in 1.13c code (static)]. Advance sign-extends i16 rate, wrap-adds accumulator, subtracts total once if signed sum>=total, else keeps sum. End predicate faults on null; condition!=0 returns total<=0, else returns `wrap32`(acc+signExtend(rate))>=total without writes/draws. A second advance uses threshold=`wrap32`(third*256); if sum>=threshold stores `wrap32`(sum+second*256-threshold), else sum. No draw. These are isolated operations, not a claim GH tick invokes them.

**HRC-R14** Block/KB rates [verified in 1.13c code (static)]. Effective player mode 9/monster mode 6 obtains D through NRF-R07 index 3, A=120, stat 102 FBR: D=0 when FBR=0, otherwise `D=trunc(wrap32(120*FBR)/wrap32(120+FBR))`. Base 100 if kind 0/1/3, stateCount>101, state 101 active; else 50. Form `p=wrap32((base+D)*B)`, reinterpret p as u32 and compute `q=floor(p/100)` through NRF-R06b; zero becomes 1, cap 32767. KB player mode 19/monster mode 13 instead uses base mode rate, unmodified by FHR/FBR/other_animrate: signed<=0 stores 0, positive caps 32767. Player base fixed 213; monster supplied signed word. KB then sets path speed 4096 through NRF-R08b, with its speed/target/reset effects. No direct draws in rate/speed stage; no full-event draw claim.

**HRC-R15** Mode-change elemental processing [unresolved]. Requested monster mode 3 skips the wrapper's elemental helper. Other modes can reach it; reached slot draws consume NRF-R04b (chance 1..99 Rand 100, accepted type 10 Rand 5+1). Upstream eligibility, byte regions, additional helper effects and complete event draws remain HRC-Q06/NRF-Q11. GH skipping this helper does not prove a full zero-draw mode event.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| HRC-R01a | Melee request | verified in 1.13c code (static) | none post-resolution | HRC-V09 |
| HRC-R01b | Missile request | verified in 1.13c code (static) | SBA upstream only | HRC-V10 |
| HRC-R01c | Soft rewrite | verified in 1.13c code (static) | none supplied-result | HRC-V20 |
| HRC-R01d | Other producers | unresolved | HRC-Q03 upstream only | none |
| HRC-R02a | Kind/dead guard | verified in 1.13c code (static) | none after prior stages | HRC-V11 |
| HRC-R02b | Town | verified in 1.13c code (static) | none supplied room/record | HRC-V17 |
| HRC-R03a | Protected state request | verified in 1.13c code (static) | HRC-Q10 disguise/update only | HRC-V12 |
| HRC-R03b | Wrapper state guard | unresolved | HRC-Q06 wrapper | none |
| HRC-R04a | Priority | verified in 1.13c code (static) | HRC-Q06 terminal effects only | HRC-V07, HRC-V21 |
| HRC-R04b | Avoidance effects | unresolved | HRC-Q06 eligibility/start | none |
| HRC-R05 | Stun bypass | verified in 1.13c code (static) | HRC-Q06 completed effects only | HRC-V03 |
| HRC-R06 | Veto | verified in 1.13c code (static) | none supplied stats/class | HRC-V01, HRC-V02, HRC-V05, HRC-V22 |
| HRC-R07a | Class synthesis | verified in 1.13c code (static) | none supplied record | HRC-V16 |
| HRC-R07b | Melee class producer | unresolved | HRC-Q03 upstream only | none |
| HRC-R08 | Soft/plain updates | unresolved | HRC-Q06 full effects | none |
| HRC-R09a | Monster start | verified in 1.13c code (static) | HRC-Q06/NRF-Q17 reset/wrapper | HRC-V13 |
| HRC-R09b | Completion | unresolved | HRC-Q02 end chain | none |
| HRC-R10a | Local player decision | verified in 1.13c code (static) | HRC-Q06 neutral/outer/start only | HRC-V14, HRC-V18 |
| HRC-R10b | Interrupt draw | verified in 1.13c code (static) | none locally | HRC-V14 |
| HRC-R10c | Player entry | unresolved | HRC-Q06 full effects | none |
| HRC-R11 | Cooldown and ordered requests | verified in 1.13c code (static) | HRC-Q06 actual start/stat-write effect only | HRC-V06 |
| HRC-R12a | GH rate | verified in 1.13c code (static) | HRC-Q04 actual input selection only | HRC-V04, HRC-V08 |
| HRC-R12b | Remap | unresolved | HRC-Q04 effective inputs | none |
| HRC-R13a | Duration | unresolved | HRC-Q02/HRC-Q04 actual duration | none |
| HRC-R13b | Data/breakpoints | unresolved | HRC-Q04 selected record | none |
| HRC-R13c | Server driver | unresolved | HRC-Q02 caller/order | none |
| HRC-R13d | Animation primitives | verified in 1.13c code (static) | none isolated; HRC-Q02 GH integration | HRC-V19 |
| HRC-R14 | Block/KB rate | verified in 1.13c code (static) | HRC-Q04 inputs/GDF-Q01 field identity; HRC-Q06 full mode only | HRC-V15 |
| HRC-R15 | Mode-event draw integration | unresolved | HRC-Q06/NRF-Q11 | none |

## 5. Outputs and side effects

| Write / effect | Rule |
|---|---|
| Result flag additions/rewrite, remaining block/avoidance code after rewrite; KNB flag conversions | HRC-R01a/HRC-R01b/HRC-R01c/HRC-R04a |
| Ordinary guard decision: continue or stop without reaction | HRC-R02a/HRC-R02b |
| Protected-branch stop of ordinary dispatch, with the stated state request when lethal | HRC-R03a |
| Damage-veto decision: veto or allow, with no mode change inside the operation | HRC-R06 |
| Monster start-handler return: 0 for missing mGH, otherwise 1 on the stated guarded/start paths | HRC-R09a |
| Local player decision: allow or refuse, including the stated neutral request on its exception path | HRC-R10a |
| Shared cursor, damage class, defender lastHitClass | HRC-R07a |
| State 92 on, routed membership/changed membership, update request | HRC-R03a; HRC-Q10 gates disguise/update execution |
| Selected terminal branch and reaction/neutral requests | HRC-R04a/HRC-R05/HRC-R09a/HRC-R10a; HRC-R04b/HRC-R08/HRC-R10c full effects gated |
| Defender seed updates | HRC-R06/HRC-R10b |
| Monster mode 3/modeChanged via NRF-R14a | HRC-R09a; reset integration gated |
| Mode 9 request, then stat-95 write request carrying frame | HRC-R11; persisted stat-write effect HRC-Q06 |
| Animation rate | HRC-R12a/HRC-R14 |
| Speed/targetSpeed4096, resetMarker 15 when speed changed | HRC-R14/NRF-R08b |
| Isolated accumulator writes/end predicate return | HRC-R13d |
| Fatal arithmetic or required-unit failure with no numeric return on the stated failure paths | HRC-R12a/HRC-R13d |
| Soft/plain update, monster hit event, current-skill clearing | HRC-R08/HRC-R10c, gated |

No persistence format or named network message is established here.

## 6. Branches and edge cases

The read local slices add no separate difficulty, PvP, boss/champion or hireling arithmetic; upstream inputs may differ and full equivalence is not claimed. Evidence is Expansion 1.13c. Reflected damage uses the supplied reflected defender at DPO handoff, whose seed owns its veto draws. Missing stat lists read 0 with NRF's conditional minimum semantics. Missing animation records beyond explicit guards remain HRC-Q04/HRC-Q06. Negative/wrapped FHR/FBR follows stated widths; F=-120 faults. GH can store 0; block forces 1 after division; KB nonpositive input stores 0. Signed stepping with one subtraction is not repeated modulo.

## 7. Open questions

| Question | Status / kind | Missing evidence / blocked boundary |
|---|---|---|
| **HRC-Q02** | open / integration | Actual GH stepping/end chain, tick helper, class-dependent completion and observed timing; not isolated HRC-R13d |
| **HRC-Q03** | open / outside-boundary | Other request producers and melee class source; none locally with supplied R/k |
| **HRC-Q04** | open / integration | Complete remap, selected GH record B/count/start and compiled correspondence; not supplied-input rates |
| **HRC-Q05** | open / outside-boundary | Client prediction/presentation owner and relation to server; none locally |
| **HRC-Q06** | open / behavior and integration | Wrapper/validator, avoidance skill inputs, soft/plain helpers, player entry, persisted block-stamp write and elemental helper; blocks full effects/event draws, not supplied requests/local arithmetic |
| **HRC-Q10** | open / integration | STL-Q07 disguise and NRF-Q17 update queue for state 92; not NRF/STL routing/membership |
| **HRC-Q11** | pending / observation | Matching gameplay cases for every verified behavioral rule, including HRC-O01–HRC-O06 |

Attempts: [Open evidence gaps](../plans/gaps/combat-hit-recovery.md#open-evidence-gaps). No other-version equivalence is asserted.

## 8. Test vectors

### 8.1 Executable vectors

Every case resets independently. Defaults: non-null live player mode 1, no active states or matching state lists, R=5, k=13, suppression 1, cursor 0, total 8000, components 0, M=25600, mGH/mKB/mode 6 true, valid room level 2, source absent, no current/passed skill, valid skill definition whenever a current skill is supplied, class/base 100 with valid records, context 1, requestedMode 4 in local player-decision fixtures independently of the current mode supplied by the case, FHR/FBR/stat 95=0, stateCount 200, state storage present, expiryFlag false, no mode-bound lists, state 92 off/unchanged, modeChanged false. Stats use real NRF-R12b/NRF-R12c selected/minimum reads with valid definitions and no minimum adjustment; unspecified values 0. Rates supply effective kind/mode/B. No draw means seed unchanged. External terminal mode/death/update/event/preparation operations log calls and have no supplied writes/draws. The stat-95 writer is an explicit supplied operation: it logs its request and stores the requested value in the fixture. That supplied stored value verifies no original writer interior; HRC-R11 verifies only request arguments/order. For V03/V11 a supplied monster wrapper dispatches the local HRC-R09a start, with no additional wrapper writes/draws; it is not a no-effect replacement of that local start. SetState executes NRF-R12a/STL-R07 routing/membership while disguise/update execution logs only. Mode write logs its request, executes NRF-R14a including its update request and local mode/modeChanged writes, then logs animation preparation. No mode-bound list is eligible for removal. Actual update/reset effects remain gated. Local HRC calls execute their rules, not no-effect substitutes. Logs are ordered. Completed gated effects are not verified by these supplied operations.

- **HRC-V01** Veto bands (HRC-R06; boundary: isolated rule)
  - Inputs: independent totals 255, 1599, 1600, 3200, 6400; other defaults.
  - Seeds: defender {1, 0} each.
  - Draws: first two none; 1600 mask 2=1, mask 4=1; 3200 mask 4=1; 6400 none.
  - Trace: d=16, T1=1600, T2=3200, T4=6400; floor/below T1 veto; positive enabled draws pass.
  - Outputs: veto, veto, allow, allow, allow; final seeds {1, 0}, {1, 0}, {1721382809, 747178471}, {1791398085, 0}, {1, 0}. No mode write.
  - Discriminates: two-step seed rejects replacing the 3/8 branch with one roll; equal thresholds select next band.
  - runtime_observed: no

- **HRC-V02** Poison (HRC-R06; boundary: isolated rule)
  - Inputs: total=poison 20000, M=25600.
  - Seeds: no draw.
  - Draws: none.
  - Trace: nonzero poison equals total before thresholds.
  - Outputs: veto, no writes/requests.
  - Discriminates: above M/4 still vetoes.
  - runtime_observed: no

- **HRC-V03** Stun/freeze (HRC-R05/HRC-R06; boundary: supplied GH branch)
  - Inputs: monster mode 1, mGH true; (a) matching state-21 list, no active freeze; (b) active state 1, no stun list; (c) active state 1 and matching state-21 list; (d) active state 21 alone, no matching list, active state 1. Total 8000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a/c bypass veto; b/d enter HRC-R06 and freeze-veto.
  - Outputs: a/c log [GH, entryPreparation, mode 3, updateRequest, animationPreparation], mode 3, modeChanged true; b/d log [softUpdate], mode 1, modeChanged false. Reset effects remain gated.
  - Discriminates: c rejects freeze-first; full terminal effects gated HRC-Q06.
  - runtime_observed: no

- **HRC-V04** FHR factors (HRC-R12a; boundary: supplied rate)
  - Inputs: effective player mode 4, B=100; F=0, 10, 30, 60, 120, -30 independently.
  - Seeds: no draw.
  - Draws: none.
  - Trace: D=0, 9, 24, 40, 60, -40; products 5000, 5900, 7400, 9000, 11000, 1000; unsigned/100.
  - Outputs: stored rates 50, 59, 74, 90, 110, 10; no other writes.
  - Discriminates: 10 rejects generic minimum 15; 59 rejects rounding upward.
  - runtime_observed: no

- **HRC-V05** Divisors (HRC-R06; boundary: isolated veto)
  - (a) High-band divisor binding:
    - Inputs: independent k=5, 4, 8, 2, 6, 10, 11, 0, 13, 53; total 12800, M=25600.
    - Seeds: no draw.
    - Draws: none.
    - Trace: d=64 for 5, T1=400; d=32 for 4/8, T1=800; d=8 for 2/6/10/11, T1=3200; others d=16, T1=1600; T2/T4=2/4 times T1, total>=every T4.
    - Outputs: all allow; divisors 64, 32, 32, 8, 8, 8, 8, 16, 16, 16; seeds unchanged.
    - Discriminates: the selected intermediate divisor for k=53 is 16 rather than low-nibble-only 64; case (b) additionally distinguishes their final draw effects.
  - (b) Divisors produce distinct terminal traces:
    - Inputs: player, total 2000, M=25600, no freeze, poison 0; independent k=5, 4, 8, 2, 6, 10, 11, 0, 13, 53. Other shared defaults remain.
    - Seeds: defender {1, 0} independently for every listed k.
    - Draws: k=5 none; k=4/8 one mask-4 draw returning 1; k=2/6/10/11 none; k=0/13/53 mask-2 draw returning 1, then mask-4 draw returning 1.
    - Trace: d=64 gives thresholds 400/800/1600, so 2000 reaches T4; d=32 gives 800/1600/3200, so only the mask-4 branch runs; d=8 gives 3200/6400/12800, so total<T1 vetoes; d=16 gives 1600/3200/6400, so both threshold draws run and pass.
    - Outputs: k=5 allow, seed {1, 0}; k=4/8 allow, seed {1791398085, 0}; k=2/6/10/11 veto, seed {1, 0}; k=0/13/53 allow, seed {1721382809, 747178471}. No writes except the stated stepped seeds and no mode requests.
    - Discriminates: one total separates all four divisors by return/draw/final-seed combination. Treating whole k=53 as low nibble 5 would skip both required draws and fails its final seed.
  - runtime_observed: no

- **HRC-V06** Block cooldown (HRC-R11; boundary: supplied block)
  - Inputs: R=16, frame 100; independent (a)FBR 24, last 82; (b)24, last 81; (c)0, last 85; (d)0, last 84; (e)R=16400, FBR 0, last 0; (f)b with supplied mode 9 return 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: threshold 18/15, equality refuses; soft skips; mode-request return is ignored for the later stat-write request.
  - Outputs: a/c/e no calls/stamp writes, stamp 82/85/0; b/d/f log [mode 9, writeStat 95(100)]. The explicitly supplied writer stores stamp 100 in those fixtures; this does not verify original writer execution. Mode stays 1 under supplied mode 9.
  - Discriminates: strict>, soft gate, and f reject>=, soft animation, or issuing the stat-write request only after successful mode start.
  - runtime_observed: no

- **HRC-V07** Order (HRC-R04a; boundary: supplied selector)
  - Inputs: player; (a)R=12; (b)R=20, frame 100, last 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: KB precedes GH; block precedes GH and runs HRC-R11.
  - Outputs: a log [KB(mode 19)]; b log [mode 9, writeStat 95(100)], stamp 100; mode 1, no GH/veto calls.
  - Discriminates: exact logs reject both reactions firing.
  - runtime_observed: no

- **HRC-V08** Rate edge (HRC-R12a; boundary: supplied rate)
  - Inputs: effective player mode 4, prior rate 7; (a)F=-60, B=100; (b)F=120, B=0; (c)F=-120, B=100.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a D=-120, factor-70, product-7000 as u32=4294960296, quotient 42949602; b product 0; c denominator 0.
  - Outputs: a32767; b0; c division fault, prior rate 7 retained. Factor 0 also stores 0 by product rule.
  - Discriminates: a rejects signed clamp 0; b rejects forced minimum 1.
  - runtime_observed: no

- **HRC-V09** Melee request (HRC-R01a; boundary: post-resolution)
  - Inputs: (a)R=1, state 54 off; (b)R=1, state 54 on; (c)R=16, state 54 off.
  - Seeds: no draw.
  - Draws: none.
  - Trace: retained hit plus absent 54 OR 4.
  - Outputs: R=5, 1, 16; no other effects.
  - Discriminates: blocked c must not gainGH.
  - runtime_observed: no

- **HRC-V10** Missile request (HRC-R01b; boundary: supplied flags)
  - Inputs: incomingR 0, KB 0; (GetHit, SoftHit)=(1, 0), (0, 1), (1, 1), (0, 0); fifth=(1, 1), state 54, KB 100.
  - Seeds: no draw.
  - Draws: none; fifth skips chance stage.
  - Trace: preset hit, GetHit priority, protected skip.
  - Outputs: R=5, 16385, 5, 1, 1; fifth missile seed unchanged.
  - Discriminates: third rejects adding both flags; fifth rejects protected KB draw.
  - runtime_observed: no

- **HRC-V11** Kind/dead (HRC-R02a; boundary: after prior town/class/state stages)
  - Inputs: (a)type 2; (b)monster 0; (c)monster 12; (d)hireling monster 1, mode 1, mGH true. Prior hit-class stage complete.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a/b/c return; d runs HRC-R06 allow then HRC-R09a start.
  - Outputs: a/b/c no reaction calls, modes unchanged; d mode 3, modeChanged true, log [entryPreparation, mode 3, updateRequest, animationPreparation]. Reset integration remains gated.
  - Discriminates: subtype-only exclusion fails d.
  - runtime_observed: no

- **HRC-V12** Protected lethal (HRC-R03a; boundary: supplied branch)
  - Inputs: state 54 on, state 92 off/unchanged, storage present; (a)R=29; (b)R=3.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a nonlethal returns; b SetState 92 on via NRF/STL.
  - Outputs: a no calls, mode 1; b log [SetState(92, 1), activeMembership(92, 1), updateRequest], active 92 true, changed 92 true, mode 1; no death/GH/KB/block request. Disguise/update execution gated.
  - Discriminates: the state request and unchanged mode 1 reject ordinary death dispatch. The state-92 name supplies no timer semantics; full dependent effects remain gated and are not tested by this discriminator.
  - runtime_observed: no

- **HRC-V13** Monster start (HRC-R09a; boundary: supplied start)
  - Inputs: monster; (a)mode 1, mGH true; (b)mode 0; (c)mode 1, mGH false; (d)missing unit; prior modeChanged false.
  - Seeds: no draw.
  - Draws: none directly; supplied preparation adds none.
  - Trace: guards before entry preparation and NRF-R14a mode write.
  - Outputs: a return 1, mode 3, modeChanged true, log [entryPreparation, mode 3, updateRequest, animationPreparation]; b return 1, mode 0, modeChanged false, no calls; c return 0, mode 1, modeChanged false, no calls; d return 1, no calls. Full wrapper draws unasserted.
  - Discriminates: missing mGH failure 0 differs from dead no-op success 1.
  - runtime_observed: no

- **HRC-V14** Interrupt rolls (HRC-R10a/HRC-R10b; boundary: local decision)
  - Inputs: current mode 5, current skill A, passed absent, srvdofunc 0, interrupt true, state 42, own 164=30; (a/b)state 15 off; (c)42off, 15off; (d)interrupt false; (e)42+15on.
  - Seeds: a{0, 29}; b{0, 30}; c/d{0, 29}; e{0, 30}, defender.
  - Draws: a Rand 10029; b30; c/d none; e30 before state 15.
  - Trace: a29<30 refuses; b equality passes; c no roll; d interrupt false refuses; e roll passes then 15 refuses.
  - Outputs: a refuse, seed{29, 0}; b allow, {30, 0}; c allow, {0, 29}; d refuse, {0, 29}; e refuse, {30, 0}; all mode 5, no requests.
  - Discriminates: e rejects testing 15 before draw; a/b distinguish< from<=.
  - runtime_observed: no

- **HRC-V15** Block/KB rates (HRC-R14; boundary: supplied rate/speed)
  - Inputs: B=256, stateCount 200; (a)player mode 9, FBR 0, 101off; (b)same FBR 30; (c)FBR 0, 101on; (d)player mode 9, B=1, FBR 0, 101off; (e)player mode 19, FHR 120, FBR 30; (f)monster mode 13, base 213, FHR 120, FBR 30; e/f speed 256, target 512, reset 7.
  - Seeds: no draw.
  - Draws: none in stage.
  - Trace: a50*256/100; bD 24, 74*256/100; c100*256/100; d=0 forced 1; e/fKB213 no stat factor.
  - Outputs: rates 128, 189, 256, 1, 213, 213; e/f speed 4096, target 4096, reset 15. Full KB event draws unasserted.
  - Discriminates: d separates block/GH minima; e rejects player Run rate 101; exact 4096 rejects unspecified speed.
  - (g) Nonpositive monster KB base:
    - Inputs: effective monster mode 13, signed base 0, FHR 120, FBR 30, speed 256, targetSpeed 512, resetMarker 7.
    - Seeds: no draw.
    - Draws: none in this rate/speed stage.
    - Trace: base<=0 stores rate 0, then the KB branch still requests path speed 4096; the speed differs, so NRF-R08b updates speed/targetSpeed and resetMarker.
    - Outputs: rate 0, speed 4096, targetSpeed 4096, resetMarker 15. Complete KB event draws remain unasserted.
    - Discriminates: rejects returning before the speed setter when rate becomes zero.
  - runtime_observed: no

- **HRC-V16** Cyclic class (HRC-R07a; boundary: supplied record)
  - Inputs: R=1, suppression 0, lastHitClass 7; (a)k=0, cursor 0, cold 10, fire 20; (b)k=0, cursor 1, cold 10, fire 20; (c)k=0, cursor 255, components 0, R=8193; (d)k=65541, cursor 0, cold 10; (e)k=53, cursor 0, cold 10; (f)R=0, k=0, cursor 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a48|13; b32|13; c16|13, byte wraps; d=65541|48; e high nibble skips; f no hit skips.
  - Outputs: a class/last 61, cursor 1; b45, cursor 2; c29, cursor 0; d=65589, cursor 1; e53, last 53, cursor 0; (f) class 0, last 7, cursor 0.
  - Discriminates: b rejects fixed element priority; d rejects word truncation; c rejects no-element cursor freeze.
  - (g) No positive element and no critical indication:
    - Inputs: R=1, k=0, suppression 0, cursor 0, cold/fire/lightning/poison all 0, lastHitClass 7; other shared defaults.
    - Seeds: no draw.
    - Draws: none.
    - Trace: visit all four components, find none positive, choose element code 0, substitute source class 13 and OR zero; increment cursor once.
    - Outputs: damage hitClass 13, defender lastHitClass 13, cursor 1; no other writes.
    - Discriminates: rejects adding critical code 16 without the critical indication; contrasts with case (c)'s class 29.
  - runtime_observed: no

- **HRC-V17** Town (HRC-R02b; boundary: supplied guard)
  - Inputs: (a)level 1, source absent; (b)40, monster valid inTown false; (c)75, monster valid inTown true; (d)103, player; (e)109, monster valid inTown true; (f)110, absent; (g)0, absent; (h)missing room.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a-e members; c/e permitted; f/g bound/nonmembership; h skips.
  - Outputs: a/b/d reject; c/e/f/g/h continue; no writes.
  - Discriminates: membership/permission rejects all-levels1..109 town approximation.
  - runtime_observed: no

- **HRC-V18** Local fast paths (HRC-R10a; boundary: local decision)
  - Inputs: current mode 5, interrupt false; (a)current absent; (b)current skill A, passed skill A, srvdofunc 67; (c)current skill A, passed skill B, srvdofunc 67; (d)c but mode 1; (e)a but state 54; (f)current skill A, passed skill A, srvdofunc 76; (g)a but mode 17.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a missing skill allows; b/f matching exception; c refuses; d neutral then allow; e/g initial guards.
  - Outputs: a/b/f allow, mode 5, no requests; c refuse, mode 5; d allow, log [neutralMode], mode 1; e refuse, mode 5; g refuse, mode 17.
  - Discriminates: c rejects exception by function alone; a rejects equating mode 5 with active uninterruptible skill.
  - runtime_observed: no

- **HRC-V19** Animation primitives (HRC-R13d; boundary: isolated operations)
  - Inputs: (a)advance acc 900, rate 300, total 1000; (b)acc 900, rate 2300, total 1000; (c)end acc 900, rate 100, total 1000, condition 0; (d)c condition 1; (e)second advance acc 900, rate 300, second 2, third 4; (f)advance acc 100, rate -1 (i16), total 1000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) 1200-1000; (b) 3200-1000; (c) next 1000>=1000; (d) total>0; (e) 1200+512-1024; (f) 100-1.
  - Outputs: (a) acc 200; (b) acc 2200; (c) true, acc 900 unchanged; (d) false, acc 900 unchanged; (e) acc 688; (f) acc 99; no other writes.
  - Discriminates: b rejects repeated modulo; d rejects ignoring condition; f rejects unsigned rate.
  - (g) Second advance below threshold:
    - Inputs: accumulator 100, signed rate 300, second 2, third 4; other defaults.
    - Seeds: no draw.
    - Draws: none.
    - Trace: sum=100+300=400, threshold=4*256=1024; signed 400<1024 skips adjustment.
    - Outputs: accumulator 400; no other writes.
    - Discriminates: always applying the adjustment would produce -112 and fails.
  - (h) Null end-predicate input:
    - Inputs: end predicate with absent unit; no animation record is supplied.
    - Seeds: no draw.
    - Draws: none.
    - Trace: the required-unit guard faults before reading animation inputs or producing a predicate result.
    - Outputs: fatal failure, no numeric return, no writes.
    - Discriminates: rejects treating a missing unit as an ended or unfinished animation.
  - (i) Nonzero animation condition with nonpositive total:
    - Inputs: end predicate, present unit, animation condition 1, total 0, accumulator 900, signed rate 100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: condition is nonzero, so the predicate tests total<=0; zero satisfies that test without advancing the accumulator.
    - Outputs: true; accumulator 900 and all other inputs unchanged, no writes.
    - Discriminates: rejects treating every nonzero-condition animation as unfinished; contrasts with case (d)'s positive-total false result.
  - runtime_observed: no

- **HRC-V20** Soft rewrite (HRC-R01c; boundary: post-block stage)
  - Inputs: R=9, GetHit false, SoftHit false, KB 0, code 1; (b)same GetHit true; (c)same KB 1. No chance producer runs here.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a clears 1, OR 16384, discards code 1; b/c map block 16, clear hit 1.
  - Outputs: (a) R=16392, code 0; (b)/(c) R=24, code 1; no other writes.
  - Discriminates: a preserves 8; b/c reject unconditional soft conversion.
  - (d) SoftHit already enabled:
    - Inputs: R=9, GetHit false, SoftHit true, KB 0, supplied block/avoidance code 1. No chance producer runs in this post-block stage.
    - Seeds: no draw.
    - Draws: none.
    - Trace: the rewrite tests GetHit false and KB zero, regardless of SoftHit; clear hit 1, OR 16384, discard code 1.
    - Outputs: R=16392, remaining code 0; no other writes.
    - Discriminates: rejects adding an erroneous SoftHit-false requirement to the rewrite.
  - runtime_observed: no

- **HRC-V21** Priority contrast (HRC-R04a; boundary: supplied selector)
  - Inputs: (a)monster, R=30; (b)player, R=30, frame 100; (c)player, R=148, state 65 off; (d)monster, R=20, class 243; (e)monster, R=20, class 100, mode 6 available.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a lethal first; b block first, runs HRC-R11; c avoidance branch owns return despite missing state; d special sync; e ordinary block.
  - Outputs: a[death]; b[mode 9, writeStat 95(100)], stamp 100; c no requests/no block/GH fallthrough; d[monsterReactionSync]; e[mode 6]; modes 1 under supplied effects.
  - Discriminates: a/b reject common death/block order; c rejects fallthrough; completed avoidance remains gated.
  - runtime_observed: no

- **HRC-V22** Early draws/availability (HRC-R06; boundary: isolated veto)
  - Inputs: total 1600, k=13; (a)player; (b)player; (c)monster, mGH false; (d)monster, mGH true, total 6400; (e)state 1, total 8000.
  - Seeds: a{2, 0}; b{3, 0}; c/d/e{1, 0}, defender.
  - Draws: a mask 2=0; (b) mask 2=1, mask 4=0; (c) mask 2=1, mask 4=1; d/e none.
  - Trace: a first veto; b second veto; c thresholds pass then availability fails; d high band; e freeze.
  - Outputs: (a) veto seed{3582796170, 0}; (b) veto{869181132, 450137329}; (c) veto{1721382809, 747178471}; (d) allow{1, 0}; (e) veto{1, 0}.
  - Discriminates: c rejects availability-before-draw; a rejects unconditional second draw.
  - (f) Zero max life without an invented floor:
    - Inputs: player, total 256, M=0, k=13, no freeze, poison 0; the valid stat definition has no minimum adjustment as in the shared defaults.
    - Seeds: defender {2, 0}.
    - Draws: none.
    - Trace: total equals the 256 floor and passes it; d=16 gives T1=T2=T4=0, so all three signed less-than comparisons fail and the player is allowed.
    - Outputs: allow, defender seed {2, 0} unchanged; no writes or mode requests.
    - Discriminates: an invented max-life floor of 4096 raw would give T2=512, draw mask 2 from seed {2, 0} and veto on zero; the required allow/unchanged-seed result rejects that alternative.
  - runtime_observed: no

### 8.2 Observation vectors

All pending; each procedure records its controlled setup and measured scope.

| Observation | Measurement | Procedure |
|---|---|---|
| **HRC-O01** | Veto floor/bands, ordered draws/final seeds, mGH | [HRC-O01](../research/spec-evidence/combat-hit-recovery.md#hrc-o01) |
| **HRC-O02** | Stun/freeze/state 54, state 92 request | [HRC-O02](../research/spec-evidence/combat-hit-recovery.md#hrc-o02) |
| **HRC-O03** | Cursor/class/divisor and producer mapping | [HRC-O03](../research/spec-evidence/combat-hit-recovery.md#hrc-o03) |
| **HRC-O04** | Rate, selected GH data, actual duration separately | [HRC-O04](../research/spec-evidence/combat-hit-recovery.md#hrc-o04) |
| **HRC-O05** | Cooldown/stamp, block/KB rates/speed vs full events | [HRC-O05](../research/spec-evidence/combat-hit-recovery.md#hrc-o05) |
| **HRC-O06** | Local player decision vs completed entries/end | [HRC-O06](../research/spec-evidence/combat-hit-recovery.md#hrc-o06) |

## 9. Acceptance and dependencies

Partial static acceptance is all verified assertions of HRC-V01–HRC-V22 against named rules/supplied boundaries: exact values, logs, draws, seeds and writes. Probability estimates alone cannot pass seeded cases. Gated terminal effects are excluded from partial acceptance. Full reproduction requires HRC-Q02/HRC-Q04/HRC-Q06/HRC-Q10, applicable producer integration, and matching gameplay for every verified rule, HRC-Q11. Arithmetic alone does not establish duration/cancellation.

**Depends on:** NRF-R01/NRF-R02 seed/draw, NRF-R06a/NRF-R06b arithmetic, NRF-R07 diminishing, NRF-R08b speed, NRF-R12a state routing, NRF-R12b/NRF-R12c selected/minimum and own-list stats, NRF-R14a mode writes with NRF-Q17 limits; GDF-R05/GDF-R06/GDF-R07/GDF-R08/GDF-R11/GDF-R12/GDF-R14/GDF-R15/GDF-R17/GDF-R18 names/schema; SBA-R13/SBA-R14/SBA-R20 production; KNB-R03b town, KNB-R06a/KNB-R06c conversion; DPO-R26 reaction ordering; STL-R07 membership, STL-R07a/STL-Q07 disguise. [MMF](state-movement-modifiers-and-freeze.md) MMF-R06f owns shared per-mode precedence; HRC owns these three rate branches. [Skill dispatch](skill-cast-start-and-do-dispatch.md) SCD-R01/SCD-Q01/SCD-Q10 distinguishes requested entry preparation from its effects; no timed-event removal is inferred for HRC entry preparation.

**Depended on by:** KNB reaction/stun, SBA flag consumer, DPO reaction stage, MMF rate references. Reverse references are not automatic prerequisites for supplied local arithmetic.

## 10. Related specs

- [combat-status-application](combat-status-application.md): stun/freeze production/lifetimes.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): per-mode rate ordering.
- [combat-knockback](combat-knockback.md): chance, conversion, entry/motion.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): result production/mapping.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): record/application/reaction.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): membership/disguise integration.

## Appendix A. Provenance

Original Expansion 1.13c Windows x86 D2Game/D2Common are identified in the [source manifest](../research/source-manifest.md). Compatible instruction readings dated 2026-10-11/12 retain their original limits. C-first availability assessment, targeted instruction/original-image checks and named text-cell reads are dated 2026-10-15. One animation predicate used C view checked against original bytes; newly examined other functions lacked C view checked against original bytes and used disassembly, keeping unread byte/helper regions gated. No gameplay vector was observed. Full original references, correction rationale, profile limits and pending procedures are in the [companion](../research/spec-evidence/combat-hit-recovery.md).
