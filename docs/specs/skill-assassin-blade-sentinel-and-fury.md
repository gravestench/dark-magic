# Blade Sentinel and Blade Fury

**Description**

Blade Sentinel creates a blade creature and records the route requested by the caster. Its attached missile is kept at the position of the unit it belongs to, expected to be the creature, and uses a shared cooldown on each target. Blade Fury attempts a projectile release only after its stored release time has passed. The creature's full movement and the complete effects of changing equipment remain uncertain.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: BSF-R, BSF-Q, BSF-V, BSF-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-assassin-blade-sentinel-and-fury.md)
Open evidence gaps: [gap file](../plans/gaps/skill-assassin-blade-sentinel-and-fury.md)

## 1. Scope

Defines the existing two-skill boundary: Sentinel cast requests, Fury's reached action release gate, the reached Sentinel missile update, per-target contact cooldown routing, creation seeding, and the seven base damage rolls from existing missile stats. Creature AI entry, route/lifetime/destruction, Fury start and held input, creation and stat propagation, hit-dependent bonuses, procs and interruption remain explicit gates or dependency boundaries. Complete damage application belongs to [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md) and the combat specs. No whole-cast random budget or universal equipment snapshot is established.

## 2. Trigger and ownership

D2Game computes the server cast decisions and reached missile/contact stages. Skills.srvdofunc selects action 44 or 48 through SCD-R06/SCD-R12, after that dispatcher's own eligibility checks. The callbacks execute synchronously at that action entry; a callback's return is distinct from the dispatcher's final result. Missile action 20 reaches the owner-follow decision. Contact uses MMC-R11/MMC-R14 and then the damage fill's base-roll stage. The companion records the original dispatch slots and output writers. No ordering against unrelated frame systems is established. The Sentinel cast requests its first type-2 AI event for the next frame; subsequent thinking and animation timing remain Q01/Q02/Q09.

## 3. Inputs and data

All arithmetic below uses signed i32 with low32 wrapping unless explicitly unsigned. Comparisons are signed; division truncates toward zero. Frames/durations use NRF-R09a; positions consume NRF-R10; seeds use NRF-R01/NRF-R02/NRF-R03a. Damage values are raw signed 8.8 amounts (NRF-R08a); the small vector values are raw integers. Supplied dependency results establish a reached local stage, not execution of those dependencies.

| Input | Semantic source | Width / meaning |
|---|---|---|
| caster, summon, missile, target, game, room | optional runtime objects | kinds player 0, monster 1, missile 3; existing storage |
| skill identity, row count, invocation level | Skills collection and action invocation | i32; valid row requires 0<=identity<count |
| summon selection, target selection, creator results | supplied operations | signed summon selection; target success and integer X/Y; optional created unit |
| selected missile and progression result P | skill selection / progression-expression operations | signed i32; P is evaluated for caster, skill and invocation level |
| release time, current frame | existing skill node / game | signed i32 |
| caster marker, summon marker, owner-reference condition | existing unit conditions | booleans; meanings beyond the stated set/clear and owner lookup remain Q09 |
| caster/summon level | stat 12 level, layer 0 | i32; NRF-R12b/NRF-R12c read; replacing write is a request under NRF-Q20 |
| caster position, cast point, owner position | integer coordinate getters / target result | i32 pair; stored command contains start and target |
| owner availability | supplied owner kind/mode/completion condition | i32 kind/mode, boolean condition; R03 tests |
| state membership | 12 inferno, 54 uninterruptable, 86 justhit | booleans, GDF-R12; membership and associated lists are distinct |
| NextHit, NextDelay | Missiles prepared row | u8; nonzero enable, unsigned frame delay |
| contact filter and hit results | MMC-R11/MMC-R14 supplied operations | ordered acceptance/rejection and to-hit result |
| list allocation and side operations | state-list allocation, attach, membership, event, placement, animation refresh | declared results/effects; Q04/Q09 retain complete integration |
| external caster equipment context | fixture weapon identity A/B and its external damage ranges | context only; never an input to isolated R12; actual propagation into missile stats is Q05 |
| missile stats, layer 0 | existing missile stat storage | signed i32 selected/minimum reads through NRF-R12b/NRF-R12c |
| game and missile seeds; initial seed low word | existing seed pairs | each word u32; initialLo u32 |

| Mapping / constant | Origin | Binding |
|---|---|---|
| action 44/48 and missile action 20 | engine dispatch slots | fixed engine-defined handlers selected by prepared data |
| Skills Blade Sentinel Id 257: srvdofunc 44, summon bladecreeper, srvmissilea blade creeper; Blade Fury Id 266: srvstfunc 26, srvdofunc 48, srvmissilea `bladefragment1` | table data, TD-BLADE-SKILLS in companion | text cells; compiled correspondence Q03/Q04 |
| Blade Fury `prgcalc1`=`par4`, `Param4`=5; startmana=6, mana=8, lvlmana=1, manashift=5, usemanaondo=1 | table data, TD-BLADE-SKILLS | text-resolved inputs; complete duration/mana/start execution Q02/Q03 |
| Missiles blade creeper Id 392: pSrvDoFunc 20, pSrvHitFunc 37, NextHit 1, NextDelay 25, Skill Blade Sentinel | table data, TD-BLADE-MISSILES | text cells; actual selected/compiled missile binding Q04 |
| MonStats bladecreeper hcIdx 413, AI=BladeCreeper, `Skill1`=Blade Sentinel | table data, TD-BLADE-MONSTATS | symbolic name mapping; numeric AI 102 and dispatcher binding Q01 |
| state 12/54/86, level 12, damage pairs and mastery stats below; event types 2/12 | engine code; GDF supplies established names | fixed original-code uses; no remapping inference |
| cast/summon marker conditions | engine code | fixed set/clear operations; wider meanings unresolved |

Named text cells are **verified in 1.13c data** only within the companion's recorded source profile. Blank cells remain blank, rather than an inferred numeric zero. Data identities do not prove which unit or missile a real cast creates.

## 4. Rules (execution order)

R01/R10 are Sentinel cast stages; R02/R06/R11 are Fury stages; R03/R05/R09 cover reached follow/interruption decisions; R04 precedes a reached base roll R12. R05/R06/R09/R10/R12 are grouping prefixes.

**BSF-R01** Sentinel action [verified in 1.13c code (static)]. Reject invalid/absent skill row or a negative supplied summon-selection result with return 0 before the caster marker write. Otherwise set the present caster's caster marker, then resolve the target. Target failure returns 0 with that write retained. Request summon creation using the skill, invocation level and resolved target; an absent result returns 0 with the marker retained. With a returned summon, request replacing its level 12 with the caster's layer-0 level 12; request placement with (game, summon, caster room, caster current X, Y, 0, 0); write its route command with start=caster current X, Y and target=resolved cast X, Y. Remove its pending type-2 events, request type 2 at low32(frame+1) with zero extra parameters, clear the summon marker, and return 1. Placement return is ignored. The level, placement, command storage and event operations have the declared supplied effects; complete creation/propagation/placement integration remains Q01/Q04/Q08/Q09. V05/V09 exercise the ordered requests, marker writes and return.

**BSF-R02** Fury action [verified in 1.13c code (static)]. Find the caster's supplied skill node; absence returns 0. Resolve the selected missile; a negative result returns 0. Obtain progression result P and set duration=i32(P-1). If duration<=0, return 0 before clearing the caster marker or resolving a target. Otherwise clear the present caster's caster marker, then resolve the target. Failure returns 0 with the clear retained. Read the node's release time. If release time<current frame, request one missile with an absolute target point, owner=caster, source=caster, selected missile, resolved target, skill identity and invocation level; all remaining creation options are zero/absent. Ignore the creator's return, write release time=i32(frame+duration), and set result 1. Otherwise make no creation/timer write and set result 0. After either gate branch, if caster state 12 is active, request animation refresh(game, caster, 1). Return result. The timer write and return 1 mean an attempted release even when the creator returns no missile. Q09 gates complete animation-refresh effects/draws; V03/V07 exercise this action with supplied external operations.

**BSF-R03** Reached Sentinel missile action [verified in 1.13c code (static)]. For a present missile, test its owner-reference condition and resolve that owner. If condition is false, owner absent, or owner unavailable, request end processing(game, missile, no target, end=true) and return 2 regardless of that operation's result. Unavailable means completion condition true, non-player/non-monster kind, player mode 0/17 or monster mode 0/12. Otherwise request synchronizing the missile to the owner's current room and position, set its signed i16 remaining-lifetime counter to 10, request the plain movement/collision step (MMC-R05) and return 1 regardless of its result. The end request is MMC's contact operation with no target and force set (MMC-R11..R16); the missile dispatcher consumes return 2 as a removal request (MMC-R01). Actual relocation/collision/end effects remain Q04/Q09; V05(c, d) exercise this reached decision. This rule does not move the creature.

**BSF-R04** Contact cooldown routing [verified in 1.13c code (static)]. Consume MMC-R11's ordered contact filters. An active target state 86 with NextHit!=0 rejects before a to-hit roll or arming. Other filter failures also skip arming. At MMC-R14's reached to-hit miss branch, request cooldown arming before the hitbymissile item event and return/end handling; at its reached successful/end hand-off, arm before that event, hit callback and damage fill. Arming with NextHit=0 or absent target does nothing. Otherwise request a type 2 list owned by the target and expiring at requestedExpiry=i32(frame+u8(NextDelay)). Allocation failure stops arming. On success, request event 12 at requestedExpiry, tag the list state 86 and its expiry callback, attach with full application, then request state 86 on. This is shared target membership across missiles/legs. STL-R01 stores a due/past requested event at frame+1; list expiry retains requestedExpiry. Complete attach/free/expiry persistence remains Q04, so an exact real cooldown duration is not inferred from a request alone. V01/V02.

**BSF-R05a** Follow ownership [verified in 1.13c code (static)]. R03 follows the current position of the missile's owner reference, never a separately supplied caster position. With the owner supplied as the creature, the missile follows the creature, not the caster (V05(c)). Which unit the real Sentinel missile's owner reference names is set by its creator and remains Q01/Q04; the assertion "follows the creature" holds only under that supplied owner.

**BSF-R05b** Creature return target [unresolved]. Q01 gates whether and when the creature returns to stored start, current caster position or another point. R01's stored command alone does not establish this behavior.

**BSF-R05c** Creature AI/lifetime [unresolved]. Numeric AI 102 binding, lifetime, destruction and complete route execution remain Q01. No completed absence search is established.

**BSF-R06a** Strict Fury timing [verified in 1.13c code (static)]. Consume R02: equality blocks, only an open gate writes the timer, and that write occurs even on failed creation. With duration 5 and initial timer 99, frames 100 and 106 attempt releases; frames 101..105 do not. V03.

**BSF-R06b** Fury start and held input [unresolved]. The text start selector 26 and its original dispatch target are located, but complete mana requests, timed-state lifecycle, animation refresh and repeat execution remain Q02/Q09. Startmana and `Param3` are distinct named cells; no `Param3`*256 mana rule is established.

**BSF-R07** Creation-time damage capture [unresolved]. Creation requests damage parameter building/storage through MCI-R24 and MDP-R04..R12. Those contracts distinguish source contributions, stored missile stats and possible creation draws. Exact ownership transfer and the captured stat set remain unresolved. Q05 gates an integrated weapon snapshot; R12 remains executable with supplied existing missile stats.

**BSF-R08** Equipment changes [verified in 1.13c code (static)]. At the isolated base-roll stage, R12 reads the missile's existing stats and seed. With those supplied values unchanged, changing a separate caster weapon input changes no R12 result. This does not establish that every real in-flight missile stat, owner/target bonus, proc or final hit is unchanged after equipment swaps; Q05/Q06 gate that larger assertion. V04.

**BSF-R09a** Owner loss/unavailability [verified in 1.13c code (static)]. R03 requests end and returns 2 on the stated missing/unavailable owner paths. The request does not itself prove all destruction/callback effects. V05(d).

**BSF-R09b** Creature/Fury interruption [unresolved]. Q06 gates removal of the creature, Fury states, associated events and complete proc/interruption effects. HRC-R10's interruption decision is a dependency, not proof that these particular effects are canceled.

**BSF-R10a** Reached Sentinel seeding [verified in 1.13c code (static)]. At successful monster creation reaching the seeder, supply the game seed to NRF-R03a: step that source once, copy new lo to the summon initialLo, initialize summon seed={new game lo, 666}. The child seed is not stepped by initialization. This is exactly one game-seed step at that operation, not the whole spawn/cast budget. Earlier failures can bypass it and other creation operations retain their draws. V06.

**BSF-R10b** Summon initialization budget [unresolved]. MSD-R01..R19 supply monster seeding, appearance, life/property/equipment and mode contexts; a particular Sentinel spawn's exact input/path/draw total remains Q08. Type 2 scheduling can additionally draw when state 54 is active. A fresh summon is not assumed to lack that state. The first-event request does not establish a zero-draw whole cast.

**BSF-R11** Reached Fury seeding [verified in 1.13c code (static)]. A missile creator that reaches its seeder supplies the game seed to NRF-R03a: one source step, missile initialLo=new game lo, missile seed={new game lo, 666}, no child step at initialization. A blocked R02 gate makes no creator call and therefore no creation-seeding draw. Complete creator/animation-refresh draws remain Q05/Q09; no whole-cast zero-draw assertion follows. V07.

**BSF-R12a** Per-element base roller [verified in 1.13c code (static)]. Read signed maximum first; if <=0 return 0 with no minimum/percentage read and no draw. Read signed minimum next; if <=0 return 0 with no percentage read and no draw. If minimum>maximum, swap. Apply R12b, then compute n=i32(maximum-minimum), r=Rand(missile seed, n) and return i32(minimum+r). NRF-R02 skips the step when signed n<=0; otherwise it steps once. There is no post-percentage swap, clamp or positivity check. Thus [minimum, maximum) only describes adjusted pairs whose signed difference is positive without endpoint overflow. V08/V10/V11.

**BSF-R12b** Percentage stage [verified in 1.13c code (static)]. After the initial positive guards/swap, use the supplied percentage-stat identity. Identity -1 skips the read. Otherwise read its layer-0 value pct; zero leaves the pair unchanged. For nonzero pct, set minimum=i32(minimum+mulDiv(pct, minimum, 100)), then maximum=i32(maximum+mulDiv(pct, maximum, 100)), consuming NRF-R05 in that operand order. Its branching/truncation and wrapping apply, not an unlimited-precision percent formula. V10.

**BSF-R12c** Base-roll caller order [verified in 1.13c code (static)]. At base-fill entry zero the damage record. Physical first: read maximum 22 then minimum 21, applying the same positive guards/swap/Rand as R12a without a percentage stage; store physical. Then use R12a in order fire, magic, lightning, cold, poison, burn, with the table below. Other field copies may occur between these draws. All seven base rolls use the missile seed, at most one draw per pair and 0..7 total within this base-roll slice. Final target-dependent physical adjustments, critical doubling, bypass/proc processing and damage application belong to MHD-R23c..e and Q06; the base record is not the final hit. V08/V11.

| Order / field | Minimum / maximum stat | Percentage stat |
|---|---|---|
| 1 physical | 21 mindamage /22 maxdamage | none |
| 2 fire | 48 firemindam /49 firemaxdam | 329 passive_fire_mastery |
| 3 magic | 52 magicmindam /53 magicmaxdam | 357 passive_mag_mastery |
| 4 lightning | 50 lightmindam /51 lightmaxdam | 330 passive_ltng_mastery |
| 5 cold | 54 coldmindam /55 coldmaxdam | 331 passive_cold_mastery |
| 6 poison | 57 poisonmindam /58 poisonmaxdam | 332 passive_pois_mastery |
| 7 burn | 316 burningmin /317 burningmax | 329 passive_fire_mastery |

Stat identities follow GDF-R11 and the companion's TD-BLADE-STATS named record; their names do not substitute for the original fixed code pair. In particular the burn slot reads the exact 316/317 pair, regardless of what those data names suggest.

### Id map

| Id | Summary | Status | Gate / executable evidence |
|---|---|---|---|
| BSF-R01 | Sentinel local cast requests and flags | verified in 1.13c code (static) | V05/V09; creation/placement effects Q01/Q04/Q08/Q09 |
| BSF-R02 | Fury attempt, timer and refresh request | verified in 1.13c code (static) | V03/V07; Q02/Q03/Q05/Q09 integration |
| BSF-R03 | Reached owner-follow/end decision | verified in 1.13c code (static) | V05(c, d); Q04/Q09 effects |
| BSF-R04 | Contact filters and cooldown requests | verified in 1.13c code (static) | V01/V02; Q04 attach/free integration |
| BSF-R05a | Owner-reference rather than caster position | verified in 1.13c code (static) | V05(c); real owner identity (creature) Q01/Q04 |
| BSF-R05b | Return target | unresolved | Q01 |
| BSF-R05c | AI entry/lifetime/destruction | unresolved | Q01 |
| BSF-R06a | Strict release timing | verified in 1.13c code (static) | V03 |
| BSF-R06b | Start/mana/repetition | unresolved | Q02/Q09 |
| BSF-R07 | Exact creation-time stat capture | unresolved | Q05 |
| BSF-R08 | Unchanged supplied stats give same base roll | verified in 1.13c code (static) | V04; actual equipment integration Q05/Q06 |
| BSF-R09a | Missing/unavailable owner requests end | verified in 1.13c code (static) | V05(d) |
| BSF-R09b | Complete interruption/removal/procs | unresolved | Q06 |
| BSF-R10a | Reached summon source-seed operation | verified in 1.13c code (static) | V06; total budget Q08 |
| BSF-R10b | Exact summon initialization draws | unresolved | Q08 |
| BSF-R11 | Reached missile source-seed operation | verified in 1.13c code (static) | V07; total creator/refresh budget Q05/Q09 |
| BSF-R12a | Existing-stat per-element roller | verified in 1.13c code (static) | V08/V10/V11 |
| BSF-R12b | Ordered percentage arithmetic | verified in 1.13c code (static) | V10 |
| BSF-R12c | Seven base-roll caller slots/order | verified in 1.13c code (static) | V08/V11; complete fill Q06/Q10 |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| caster caster marker set/cleared, including retained partial failure | R01/R02 |
| summon request; level replacement request; placement request; route command start/target write request; type 2 removals and next-frame request; summon marker clear | R01 |
| absolute-target missile creation request; release-time write; conditional animation-refresh request; action return 0/1 | R02/R06a |
| owner-position synchronization request, remaining-lifetime counter 10 write request, collision/end requests, reached return 1/2 | R03/R05a/R09a |
| target-owned type 2 list allocation with requested expiry; event 12 request; state 86/callback tags, attachment and membership-on requests | R04 |
| source seed step; child initialLo and initialized seed writes at seeder | R10a/R11 |
| base damage record zeroing and seven base component stores; missile seed steps | R12a/R12b/R12c |

Full storage/propagation for declared dependency requests is gated as named above. No named client message or save persistence effect is established locally.

## 6. Branches and edge cases

An invalid skill, absent node, negative selection or nonpositive wrapped duration stops at its stated position. Target failure after a marker change does not roll back that change. Equality blocks Fury; creator failure after the gate does not stop its timer/return 1. State 12 refresh occurs on the blocked branch too. No blanket no-effect or no-draw assertion spans external operations. A null owner, an unavailable owner and a failed end operation all give reached missile return 2; actual cleanup remains external. NextHit 0 never arms a list, and allocation failure makes no event/state/attach request. NextDelay 0 requests a list expiring now but schedules at next frame under STL-R01. Equal positive damage endpoints give a fixed value without a draw; nonpositive original endpoints give 0 before percentage adjustment. Negative percentages can reverse adjusted endpoints; R12 does not swap again. Difficulty, PvP, block, resistance and post-hit effects consume their separate owners.

## 7. Open questions

| Question | Kind / status | Missing evidence / boundary |
|---|---|---|
| **BSF-Q01** | behavior / unresolved | Anchor numeric AI 102 and BladeCreeper's dispatcher consumer; read its complete lifetime, route, return and destruction path. R01 supplied cast outputs remain available |
| **BSF-Q02** | behavior / unresolved | Complete start 26, held-input and mana execution, including timed-state/animation producers and return propagation |
| **BSF-Q03** | integration / unresolved | Establish compiled progression selection/evaluation correspondence to text `prgcalc1`=`par4`, `Param4`=5; supplied P remains available. Text implies P5 and duration 4 only if that correspondence holds |
| **BSF-Q04** | integration / unresolved | Establish selected/compiled missile binding, the unit named by the real Sentinel missile's owner reference (R05a), actual collision/end effects and timed-list attach/free/expiry. Named text NextDelay 25 is known; supplied cooldown requests remain available |
| **BSF-Q05** | integration / unresolved | Exact creation parameter/stat propagation, mutable missile stats, equipment capture and full creator draw budget; supplied stored base-roll inputs remain available |
| **BSF-Q06** | behavior / unresolved | Complete target/owner bonuses, elemental/proc/interruption and creature/Fury state removal. Base-roll outputs alone are not final-hit assertions |
| **BSF-Q07** | outside-boundary / unresolved | Other builds have not been assessed; no 1.14d equivalence assertion or local behavior gate |
| **BSF-Q08** | integration / unresolved | Instantiate Sentinel's monster creation/init/mode inputs to obtain exact extra draw count after the verified seeding operation |
| **BSF-Q09** | integration / unresolved | Complete placement and animation-refresh effects/draws, marker meanings and real request persistence; state 12 blocked action can reach refresh |
| **BSF-Q10** | observation / pending | Reproduce executable base-roll/cast/contact assertions on an owned 1.13c game; full hit integration additionally requires Q06. All O01..O03 remain pending |

The [gap file](../plans/gaps/skill-assassin-blade-sentinel-and-fury.md) records assessed attempts and exact blocked assertions. No pending observation blocks the otherwise verified supplied-input arithmetic.

## 8. Test vectors

### 8.1 Executable vectors

Defaults for each independent case: supplied objects/storage exist, valid skill row, selected missile nonnegative, target selection succeeds, state 12/state 54 false, marker false, completion condition false, no pending events, current frame 100, duration 5 (P=6), caster level 30. Dependency operations log exact ordered requests and return the stated result; omitted external operations neither draw nor change state in these fixtures. A successful supplied level write persists that value; command writes persist the stated points; list allocation creates an empty list, attach records that list, membership writes persist the active bit, and event scheduling uses STL-R01. These supplied effects do not verify actual producers. Creation is an external operation: fixtures explicitly identify whether it reaches the seeder. Every reached local rule executes normally. R12 cases default every listed damage/mastery stat to 0 and use seed {lo=1, hi=0}; values are returned layer-0 stat values. Seven-component tuples are (physical, fire, magic, lightning, cold, poison, burn) at the base-roll boundary; all runtime_observed values are no. A missing optional owner result is stated explicitly; otherwise repeated owner lookups return the same supplied live monster. All reached collision/end/placement operations log calls and have no other fixture effects unless specified.

- **BSF-V01** Shared-target cooldown (R04; boundary: supplied-input stage)
  - Inputs: NextHit 1, NextDelay 25, first contact frame 100, target state 86 off, earlier filters pass, hit supplied true, allocation succeeds; later contact at 110 is from the return leg to the same target with state 86 still on.
  - Seeds: no draw in routing; supplied hit/filter/list operations make no draw.
  - Draws: none.
  - Trace: first arm expiry 100+25=125; schedule 125; second contact rejects state 86 before a hit request.
  - Outputs: one list owner=target, type 2, expiry 125, state 86, expiry callback; event 12 due 125; target state 86 on; ordered allocation, schedule, tag, callback, attach, state-on, hitbymissile; later contact adds none of these and no hit request. Removal at 125 is Q04, not asserted.
  - Discriminates: a separate return-leg cooldown would allow the second contact; its absent hit/arm request rejects that behavior.
  - runtime_observed: no

- **BSF-V02** Miss and arming exits (R04; boundary: supplied-input stage)
  - Inputs: frame 100, initial state 86 off, NextHit 1, NextDelay 25, filters pass; independent cases below.
  - Seeds: no draw in routing; supplied hit result makes no draw.
  - Draws: none in all cases.
  - Trace: reached miss arms before item event; earlier exits skip it; zero delay keeps list expiry 100 but scheduler stores 101.
  - (a) Inputs: to-hit miss, allocation succeeds. Outputs: list expiry 125, state 86 on, event 12 due 125; allocation, schedule, tag, callback, attach, state-on, hitbymissile in order before miss return handling.
  - (b) Inputs: initial state 86 on. Outputs: no to-hit request, list, event, item-event or membership change; state 86 remains on.
  - (c) Inputs: earlier contact type filter rejects. Outputs: no to-hit or arming/item-event request; state 86 off, lists/events empty.
  - (d) Inputs: to-hit miss, allocation fails. Outputs: one allocation request expiry 125, no schedule/tag/attach/state-on; hitbymissile still requested; state 86 off, lists/events empty.
  - (e) Inputs: to-hit miss, NextDelay 0, allocation succeeds. Outputs: list expiry 100, state 86 on, event 12 due 101; same ordered calls as(a), requested event frame 100.
  - (f) Inputs: to-hit miss, NextHit 0. Outputs: hitbymissile requested, no allocation/event/attach/state-on; state 86 off, lists/events empty.
  - Discriminates: miss-only-no-arm fails(a); arm-every-contact fails(b, c); scheduling on failed allocation fails(d); clamping the list as well as the event fails(e).
  - runtime_observed: no

- **BSF-V03** Fury gate, failed creation and partial failure (R02/R06a; boundary: supplied-input stage)
  - Inputs: P=6, duration 5; creator returns absent and draws nothing; each state 12 refresh logs only and draws nothing; marker initially true.
  - Seeds: no draw in supplied operations.
  - Draws: none in all cases.
  - Trace: gate 99<100 opens, writes 105 despite failure; equality 105 blocks; 105<106 opens and writes 111.
  - (a) Inputs: frame 100, timer 99, state 12 false. Outputs: marker false, one creation request (absolute target, owner/source caster, selected missile, skill/level, target; remaining options zero), return 1, timer 105, no missile, no refresh.
  - (b) Inputs: frame 101, timer 105, state 12 true. Outputs: marker false, return 0, timer 105, no creation; exactly one refresh(game, caster, 1), no missile.
  - (c) Inputs: frame 105, timer 105, state 12 false. Outputs: marker false, return 0, timer 105, no creation/refresh/missile.
  - (d) Inputs: frame 106, timer 105, state 12 false. Outputs: marker false, return 1, timer 111, one creation request, no missile/refresh.
  - (e) Inputs: frame 100, timer 99, P=1. Outputs: return 0, timer 99, marker true, no target/creation/refresh request.
  - (f) Inputs: frame 100, timer 99, target selection fails. Outputs: return 0, timer 99, marker false, no creation/refresh request.
  - Discriminates: <= gate fails(c); conditioning timer/return on creator success fails(a); skipping refresh on blocked casts fails(b); rolling back marker fails(f).
  - runtime_observed: no

- **BSF-V04** Existing ranges and an external equipment change (R08/R12a/R12c; boundary: supplied-input stage)
  - Inputs: fire 48=5, 49=9, 329=0; external caster weapon initially A then B with larger external ranges 20..40; supplied missile stats remain 5/9. Each independent before/after base fill starts with the same seed.
  - Seeds: missile {lo=1, hi=0} in each case.
  - Draws: each case (missile, 4, 1), one step.
  - Trace: zero physical skips; fire 5+1=6; other zero maxima skip.
  - Outputs: before and after tuple(0, 6, 0, 0, 0, 0, 0), seed{lo=0x6AC690C5, hi=0}. No external weapon read in this stage; actual missile stat propagation/equipment integration remains Q05.
  - Discriminates: replacing supplied missile fire with B's range yields another output and violates the exact tuple.
  - runtime_observed: no

- **BSF-V05** First think and reached follow/end (R01/R03/R05a/R09a; boundary: supplied-input stage)
  - Inputs: successful summon result, cast point(120, 220), caster point(100, 200), caster room A; supplied writes/placement/event operations have default fixture effects.
  - Seeds: no draw in these supplied operations.
  - Draws: none in all cases.
  - Trace: frame 100+1=101; successful owner branch ignores collision return; end branch ignores end return.
  - (a) Inputs: cast at 100; placement returns failure 0 and has no fixture effect. Outputs: return 1, caster marker true, summon level 30, route start(100, 200)/target(120, 220), summon marker false, event(type 2, due 101, parameters 0/0); ordered creation, level replacement, placement(game, summon, A, 100, 200, 0, 0), command write, remove type 2, schedule type 2, summon-marker clear despite placement failure.
  - (b) Inputs: frame 2147483647; schedule operation has supplied acceptance of wrapped request for this isolated request case. Outputs: return 1, same marker/level/route outputs as(a); requested type 2 frame=-2147483648; actual wrapped-negative scheduling is gated by STL-Q10.
  - (c) Inputs: missile owner condition true, owner=creature, creature live monster mode 1 at(121, 221), room B; caster at(99, 199); collision returns 2. Outputs: synchronize(missile, B, 121, 221), remaining-lifetime counter 10, collision request, return 1; no end request.
  - (d) Inputs: absent owner, or independently owner monster mode 12; end operation returns 1. Outputs: one end(game, missile, no target, true), return 2, no synchronize/remaining-lifetime/collision request.
  - Discriminates: using caster position fails(c); propagating collision/end results fails(c, d); same-frame first event or aborting on placement failure fails(a).
  - runtime_observed: no

- **BSF-V06** Sentinel reached seeder (R10a; boundary: isolated rule)
  - Inputs: monster summon and non-null supplied game seed; no subsequent initialization runs within this vector.
  - Seeds: game{lo=0x00000010, hi=0x00000003}; child overwritten by initializer.
  - Draws: one game source step, no modulus; no child draw.
  - Trace: 0x10*0x6AC690C5+3=0x6_AC690C53.
  - Outputs: game{lo=0xAC690C53, hi=0x00000006}; summon initialLo=0xAC690C53, seed{lo=0xAC690C53, hi=0x0000029A}; exactly one game step in seeding.
  - Discriminates: stepping the child or seeding it from the old game low word fails the exact seed pair.
  - runtime_observed: no

- **BSF-V07** Fury supplied creator with reached seeder (R02/R11; boundary: supplied-input stage)
  - Inputs: frame 100, timer 99, P=6; supplied creator reaches NRF-R03a exactly once, performs no other draws and returns a missile; state 12 false. Second case independently supplies timer 105 and the stepped game seed.
  - Seeds: (a)game{lo=0x10, hi=3}; (b)game{lo=0xAC690C53, hi=6}.
  - Draws: (a)one game step, no modulus; (b)none.
  - Trace: same seeding arithmetic as V06; 105>=101 blocks second creator call.
  - (a) Inputs: frame 100, timer 99. Outputs: return 1, timer 105, marker false, one creation request, game{lo=0xAC690C53, hi=6}; missile initialLo=0xAC690C53, seed{lo=0xAC690C53, hi=666}; no refresh.
  - (b) Inputs: frame 101, timer 105. Outputs: return 0, timer 105, marker false, game unchanged{lo=0xAC690C53, hi=6}; no creation/refresh request and no new missile.
  - Discriminates: seeding every gate check changes(b)'s seed; interpreting the supplied one-step creator as a general creator budget is outside this vector.
  - runtime_observed: no

- **BSF-V08** Physical/fire caller slots (R12a/R12c; boundary: supplied-input stage)
  - Inputs: other pairs/masteries 0; cases start independently.
  - Seeds: missile{lo=1, hi=0}.
  - Draws: (a)(missile, 4, 1); (b)(missile, 4, 1),(missile, 4, 1).
  - Trace: (a)physical 10 fixed; fire 5+1=6. (b)physical 12/8 swaps to 8/12 =>9; next seed step gives low 0x669A3799, high 0x2C8909E7 and fire 5+1=6.
  - (a) Inputs: 21=10, 22=10, 48=5, 49=9. Outputs: tuple(10, 6, 0, 0, 0, 0, 0), seed{lo=0x6AC690C5, hi=0}, one draw.
  - (b) Inputs: 21=12, 22=8, 48=5, 49=9. Outputs: tuple(9, 6, 0, 0, 0, 0, 0), seed{lo=0x669A3799, hi=0x2C8909E7}, two draws.
  - Discriminates: no swap fails(b). V08(b) does not independently discriminate physical/fire order because both modulus-4 results are 1; BSF-V11(e) supplies distinct per-slot moduli and rejects reversed caller order.
  - runtime_observed: no

- **BSF-V09** Sentinel markers, level and failures (R01; boundary: supplied-input stage)
  - Inputs: caster level 30, caster marker false; returned summon initially level 1, summon marker true; caster point(100, 200), cast point(120, 220), room A; supplied level replacement/command storage persist their requests.
  - Seeds: no draw in supplied operations.
  - Draws: none in all cases.
  - Trace: marker is set before target lookup; later failure does not undo it.
  - (a) Inputs: valid row, nonnegative selection, target success, summon present. Outputs: return 1, caster marker true, summon level 30, summon marker false, route start(100, 200)/target(120, 220), event 2 due 101; placement called with summon, not caster.
  - (b) Inputs: target failure. Outputs: return 0, caster marker true, no summon/level/placement/command/event requests.
  - (c) Inputs: spawn absent. Outputs: return 0, caster marker true, one spawn request, no later level/placement/command/event requests.
  - (d) Inputs: negative summon selection. Outputs: return 0, caster marker false, no target/spawn/later requests.
  - Discriminates: clearing the caster's summon marker instead of the summon fails(a); blanket no-effect failures fail(b, c).
  - runtime_observed: no

- **BSF-V10** Percent arithmetic and final roll (R12a/R12b; boundary: isolated rule)
  - Inputs: helper min 5, max 9, pct identity 329; cases independent, not a whole base fill.
  - Seeds: missile{lo=1, hi=0}.
  - Draws: (a)(missile, 6, 3); (b)(missile, 3, 0); (c)none; (d)(missile, 70100, 62685); (e)none.
  - Trace: (a)5+trunc 250/100=7, 9+trunc 450/100=13; (b)5+trunc(-150)/100=4, 9+trunc(-270)/100=7; (c)5-10=-5, 9-18=-9, difference-4 skips draw.
  - (a) Inputs: pct 50. Outputs: adjusted pair 7/13, return 10, seed{lo=0x6AC690C5, hi=0}, one draw.
  - (b) Inputs: pct-30. Outputs: adjusted pair 4/7, return 4, seed{lo=0x6AC690C5, hi=0}, one draw.
  - (c) Inputs: pct-200. Outputs: adjusted pair-5/-9, return-5, seed{lo=1, hi=0}, zero draws.
  - (d) Inputs: minimum 200, maximum 300, pct 70001. Outputs: adjusted pair 140200/210300, return 202885, seed{lo=0x6AC690C5, hi=0}, one draw. Trace: NRF-R05's first operand 70001 selects divide-first (70001/100)*endpoint; additions give 140200/210300, n=70100, roll 62685.
  - (e) Inputs: minimum 1, maximum 2147483647, pct 100. Outputs: adjusted pair 2/-49, return 2, seed{lo=1, hi=0}, zero draws. Trace: maximum's large second operand selects (2147483647/100)*100=2147483600; wrapping endpoint sum=-49, n=-51 skips draw.
  - Discriminates: flooring negative percentages fails(b); a second swap/clamp after percentage fails(c); reversing mulDiv's operands changes(d)'s pair; unlimited-precision multiplication fails(e).
  - runtime_observed: no

- **BSF-V11** Lightning and seven-slot order (R12a/R12c; boundary: supplied-input stage)
  - Inputs: zero defaults except listed cases; each case starts independently.
  - Seeds: missile{lo=1, hi=0}.
  - Draws: (a, b)(missile, 4, 1); (c, d, f)none; (e)ordered draws in the arithmetic table below, all missile-owned.
  - Trace: (a, b)5+1=6; (c)equal endpoints fixed 7; (d)maximum 0 skips; (e)each positive pair consumes one draw in R12c order.
  - (a) Inputs: 50=5, 51=9, 330=0. Outputs: tuple(0, 0, 0, 6, 0, 0, 0), seed{lo=0x6AC690C5, hi=0}, one draw.
  - (b) Inputs: 50=9, 51=5, 330=0. Outputs: tuple(0, 0, 0, 6, 0, 0, 0), seed{lo=0x6AC690C5, hi=0}, one draw.
  - (c) Inputs: 50=7, 51=7. Outputs: tuple(0, 0, 0, 7, 0, 0, 0), seed{lo=1, hi=0}, zero draws.
  - (d) Inputs: 50=7, 51=0, 330=50. Outputs: tuple(0, 0, 0, 0, 0, 0, 0), seed{lo=1, hi=0}, no minimum/mastery read for lightning, zero draws.
  - (e) Inputs: physical 5/9, fire 15/20, magic 25/31, lightning 35/42, cold 45/53, poison 55/64, burn 65/75, masteries 0. Outputs: see exact tuple, seven ordered results and final seed in the immediately following arithmetic table.
  - (f) Inputs: 50=0, 51=9, 330=50. Outputs: tuple(0, 0, 0, 0, 0, 0, 0), seed{lo=1, hi=0}, zero draws and no percentage-stage request for lightning. Trace: positive maximum reaches the nonpositive minimum guard and returns 0 before percentage adjustment.
  - Discriminates: rolling equal or zero endpoints fails(c, d, f); checking only maximum positivity fails(f); altered element order is checked by(e)'s per-slot results.
  - runtime_observed: no

| Slot | Ordered draw (missile, modulus, result) | Component | Seed after draw {lo, hi} |
|---|---|---|---|
| physical | (missile, 4, 1) | 6 | {0x6AC690C5, 0x00000000} |
| fire | (missile, 5, 4) | 19 | {0x669A3799, 0x2C8909E7} |
| magic | (missile, 6, 0) | 25 | {0x3AD1E2A4, 0x2ACB7052} |
| lightning | (missile, 7, 4) | 39 | {0xB0A31886, 0x18888770} |
| cold | (missile, 8, 6) | 51 | {0x3779C68E, 0x49AC8A30} |
| poison | (missile, 9, 4) | 59 | {0xD7E63576, 0x172373C5} |
| burn | (missile, 10, 7) | 72 | {0xEEA0F793, 0x5A0CC847} |

V11(e) exact tuple=(6, 19, 25, 39, 51, 59, 72), final seed={lo=0xEEA0F793, hi=0x5A0CC847}, seven draws.


### 8.2 Observation vectors

- **BSF-O01** Count contacts on outbound/return passes through the same stationary target and compare an off-path control. Record actual selected missile/NextHit/NextDelay, state/list/event frames and miss versus hit. Expect the R04 request order; completed cooldown expiry requires Q04. [Procedure](../research/spec-evidence/skill-assassin-blade-sentinel-and-fury.md#observation-procedures).
- **BSF-O02** Measure held-input Fury action-entry frames, release attempts, creator results, timer and mana. Expect strict timer<frame and the R02 attempt result; do not substitute a guessed fixed held-input interval or mana cost. [Procedure](../research/spec-evidence/skill-assassin-blade-sentinel-and-fury.md#observation-procedures).
- **BSF-O03** Swap between weapons with a large damage difference after launch; measure existing missile stats, seed, base rolls and final hit separately. R08 predicts only unchanged-supplied-stat base rolls; Q05/Q06 retain the integrated comparison. [Procedure](../research/spec-evidence/skill-assassin-blade-sentinel-and-fury.md#observation-procedures).

## 9. Acceptance and dependencies

Partial acceptance is V01..V11's verified local decisions, exact requests, reached seed operation and base-roll outputs. Their supplied operations are explicit fixtures. Passing these establishes neither actual creation/cooldown/stat propagation, AI route/lifetime, held-input/mana, complete equipment/proc/interruption behavior nor final damage. Full reproduction requires Q01..Q06/Q08/Q09 integration and matching observations Q10/O01..O03. Q07 is outside the evidence build.

**Depends on:** NRF-R01/R02/R03a/R05/R08a/R09a/R10/R12a..c for seed, arithmetic, units and existing reads; NRF-Q20 for replacing stat propagation. GDF-R07/R08a/R11/R12 for named schemas/identities. [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md) SCD-R06/R12 for action invocation and final dispatcher semantics; [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md) SCX-R11 for expression requests, with evaluated P supplied. [missile-creation-and-initialization](missile-creation-and-initialization.md) MCI-R10/R24 and [missile-damage-parameters](missile-damage-parameters.md) MDP-R04..R12 own creator seeding/parameters/stores; their gates remain Q05. [missile-movement-and-collision](missile-movement-and-collision.md) MMC-R01/R05 own the missile dispatcher's return handling and the plain step R03 requests, and MMC-R11/R14/R16 own contact filtering/handoff/arming and R03's end contact; [state-timed-state-lifecycle](state-timed-state-lifecycle.md) STL-R01/R06a/R06b/R07/R05 own scheduling/list effects, preserving attach/free gates. [monster-scaling-by-players-and-difficulty](monster-scaling-by-players-and-difficulty.md) MSD-R01..R19 supplies creation/init context; [combat-hit-recovery](combat-hit-recovery.md) HRC-R10 supplies interruption context without complete cancellation. MHD-R23 owns the remainder of hit-time fill.

**Depended on by:** no downstream behavioral consumer is established. Sibling-spec citations are research relationships.

## 10. Related specs

- [skill-assassin-sentry-firing-and-range](skill-assassin-sentry-firing-and-range.md): shared summon, target, speed and creation context; its labels do not verify Blade AI.
- [skill-assassin-death-sentry-corpse-damage](skill-assassin-death-sentry-corpse-damage.md): sibling trap skill.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md), [combat-damage-mitigation](combat-damage-mitigation.md): downstream application and mitigation.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): hit and avoidance operations.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): owner attribution and interaction.

## Appendix A. Provenance

Evidence is original Lord of Destruction Expansion 1.13c Windows x86, D2Game/D2Common identified by the [source manifest](../research/source-manifest.md). Compatible disassembly readings from 2026-10-11/12 are reused within their recorded boundaries; new targeted instructions, original-image decoding, selected matched reconstructions and named loader-cell records were assessed on 2026-10-15. The [companion](../research/spec-evidence/skill-assassin-blade-sentinel-and-fury.md) records dispatch chains, imports, corrections, table-profile limits and observation procedures. No gameplay observation was performed; all executable and observation vectors remain unobserved.
