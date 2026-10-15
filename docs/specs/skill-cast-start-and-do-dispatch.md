# Skill cast start and do dispatch

**Description**

Using a skill has a preparation step and an action step. The server checks whether the skill can start, runs its action when the animation calls for it, and requests resource costs at the appropriate step. Some skills also request a missile or repeat their action later. These shared checks determine whether a use proceeds before the skill's own effect is considered.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: SCD-R, SCD-Q, SCD-V, SCD-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-cast-start-and-do-dispatch.md)
Open evidence gaps: [gap file](../plans/gaps/skill-cast-start-and-do-dispatch.md)

## 1. Scope

Defines shared server start and action dispatch: selector bounds and empty slots, ordered validation, callback arguments and returns, mode entry and frame events, missile requests, resource-request timing, periodic scheduling, and other action-table callers. Individual skill effects belong to per-skill specs. Actual mana, quantity, delay, skill level, targeting, animation production, aura state and command interpretation require their own contracts. Supplied results are explicit inputs here; unresolved integration effects remain gated.

## 2. Trigger and ownership

D2Game owns the dispatch decisions. Entering player modes 7, 8, 10–16 or 18 reaches start validation. Their animation events reach actions, once per accepted event rather than automatically each frame. Type-8 server events can repeat actions. Immediate callers and missile functions also enter dispatch or its action table; their eligibility remains SCD-Q04/SCD-Q09. Mode identities follow GDF-R14a.

Within a cast: target/mode-entry preparation → start validation/callback → animation event → action callback → missile request → consuming resource requests → periodic scheduling. The point-target and unit-target entry orders are distinct under SCD-R01. No relative ordering against unrelated server systems or named client message is established here. Client command meanings and mode selection remain SCD-Q02/SCD-Q07. Binary chains are in the companion.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| game, unit, room, record presence | runtime objects; GDF-R08a | optional objects; supplied room lookup result per call | Required objects and table bounds; repeated lookups need not be replaced by a cached result |
| skill identity/count and invocation level | selected skill and call inputs | i32 | Callback identity/count bound and level passed into the dispatcher |
| bonus-inclusive skill level | supplied skill-level operation(unit,skill node,bonus=true) | i32 | Validator/frame call level or action ownership check; independent of an action invocation's already supplied level |
| unit kind and mode | unit; GDF-R14a | i32 | Player 0, monster 1, other kinds 2–5 |
| current/owned skill, base level, charge marker/count | skill list/node | optional nodes; i32 | Marker -1 means ordinary skill |
| start/action selectors | Skills.srvstfunc, srvdofunc; GDF-R08a | signed 16-bit reads | Select callbacks |
| additional selectors | Skills.ItemEffect, `srvprgfunc1..3`; States.srvactivefunc | signed 16-bit reads | Alternative action entries |
| missile/animation selectors | Skills.srvmissile, anim | signed 16-bit; u8 | Missile and Sequence animation 18 |
| flags | Skills.decquant, lob, aura, periodic, InTown, TargetAlly, ItemCheckStart, usemanaondo | booleans | Gates below; GDF-R08a |
| collision selector | Skills.lineofsight | u8 | Start collision mask |
| period/delay expressions | Skills.perdelay, delay; SCX-R11 | expression reference; i32 result | Evaluated with (unit,id,level) |
| missile action selectors | Missiles.Param2, sHitPar1 | numeric selector | Direct table index |
| target/positions | target and position operations | optional unit with kind/identity; i32 coordinate requests | Selected target unit, stored target point, unit position and resolved target point |
| target-coordinate resolution | supplied resolution(game,unit) | i32 pair `(tx,ty)` always supplied, plus success = both nonzero | One operation used by SCD-R04 (success consulted) and SCD-R07 (success ignored); actual target/point source SCD-Q10 |
| target refresh | supplied target-refresh request(game,unit) | request; effects declared per case | Neutral request before the validator's and exceptional frame route's target/predicate work; actual effects SCD-Q03/SCD-Q10 |
| eligibility results | target eligibility, party rejection, activation, bonus level, collision | supplied boolean/i32 | Actual integrations SCD-Q03/SCD-Q06/SCD-Q10 |
| mode-entry request | requested mode; point-target point (point route) or unit target (unit route) | i32 mode; i32 pair; optional unit | Caller inputs to SCD-R01 |
| frame events | sequence of (code, argument) pairs delivered to SCD-R09a | i32 pairs | Supplied ordinary events in vectors |
| controls | force, consume, forced, itemTarget | i32 zero/nonzero | Independent start/action controls |
| callbacks | selected start/action operation | i32 return; declared effects/draws | Receives (game,unit,id,level) |
| do-handled latch | unit animation condition | boolean | Entry clears, missile branch sets |
| skill conditions A/B and special frame predicate | flags on the current (or owned) skill node; predicate | booleans; i32 predicate | Neutral names; wider meaning SCD-Q01 |
| event and completion | code, argument, frame-finished predicate | i32; argument narrowed to 24 bits | Sequence-event value and handler completion |
| periodic state | game frame, pending events, repeat kind | i32 frame; list; zero/nonzero | Due frame and replacement set |
| dependency operations | entry preparation(game,unit), mode/target/mode-state preparation, mana/quantity, delay application(game,unit,value), event storage, creator | supplied operations/effects | Entry preparation is a neutral request before mode setting; delay application receives the evaluated signed delay value; actual effects remain SCD-Q10 |
| command check | length, target kind/id, range result, selected slot, stat 328 | i32; unsigned kind | Command boundary, SCD-Q02 |

Vector records are synthetic supplied records, not archive rows. GDF-R07/GDF-R08a own field schemas; no archive-cell value is asserted here.

| Mapping/constant | Origin | Binding |
|---|---|---|
| Start bound 91, populated 1–29/31–65, null 0/30/66–90 | engine code/original table | Fixed original profile; data selects handler |
| Action bound 191, populated 1–152, null 0/153–190 | engine code/original table | Fixed original profile; data selects handler |
| Line selectors 1–5 → 0x4/0x1C09/0x180/0x804/0x805 | engine code | Data selects engine-defined mask |
| Skill modes 7/8/10–16/18, reset mode 5 | engine code; GDF-R14a names | Fixed uses of mode identities |
| Minimum period 5, event type 8, command length 9/range 50 | engine code | Fixed by original code |

## 4. Rules (execution order)

Verified local stages consume supplied dependency results. This verifies request order, arguments, returns and local writes; actual dependency execution is separately gated. Calls between rules of this spec execute those rules and retain their effects; they are not replaceable dependency results. Individual skill callbacks are deliberate outside-boundary operations with declared returns/effects/draws.

**SCD-R01** Mode entry [verified in 1.13c code (static)]. Listed player modes use two distinct entry orders:

1. Point-target entry: request setting the target point; request entry preparation(game,unit); request the mode change; request setting the unit target to absent.
2. Unit-target entry: request entry preparation(game,unit); request the mode change; then test the supplied target. An absent target asserts fatally here, after those two requests, with no numeric return and no later work. Otherwise request setting that unit target. This route makes no point-target setter request and does not clear the stored target point through its unit-target setter.

On either surviving route, request mode-state reset, pending-event removal and animation preparation in that order; clear the latch; execute SCD-R02. Do not branch on its result or undo its writes: a validator-requested player reset still changes the mode. Point selection clears unit-target selection; selecting a unit leaves the stored point unchanged through that target setter. Actual entry/mode/animation preparation effects remain SCD-Q01/SCD-Q10; supplied-operation vectors do not establish their internal bookkeeping.

**SCD-R02** Start validator [verified in 1.13c code (static)]. In order:

```
if current skill absent: return 0
id = current skill identity
if supplied target eligibility fails: reset player; return 0
level = supplied bonus-inclusive level
if charge marker != -1:
    level = base level
    if charges == 0: return 0      # only exactly zero; other counts continue
if id < 0 or id >= skill count: return 0
request target refresh(game,unit)
target = current selected target (read after the refresh)
if target exists, differs from unit, TargetAlly clear,
   and supplied party rejection(game,unit,id) nonzero: return 0
if InTown clear and supplied town test(room lookup result) true:
    clear current skill; reset player; return 0
r = start dispatch(id,game,unit,level,force=0)
if player and r == 0: reset player
return r
```

Target eligibility receives (game,unit,id); party rejection receives (game,unit,id) and no level. The validator requests a room lookup and passes its result, even absent, to the supplied town test; that test result alone decides (actual absent-room behavior SCD-Q10). Reset player requests mode 5, force 1 for a player only. Missing skill, zero charges, invalid identity and party rejection do not themselves reset. Actual target-predicate and target-refresh side effects are SCD-Q03; level/reset execution is SCD-Q10.

**SCD-R03** Start preconditions [verified in 1.13c code (static)]. In order: absent game or unit returns 0; look up the unit's room and return 0 if absent; reject identity outside `[0,skill count)`, or an absent record, with 0; return 0 when force is zero and supplied activation fails. These failures do not clear current skill or reset mode in this dispatcher. Then, only with InTown clear, look up the room again: an absent or town result on this second lookup clears current skill, resets a player and returns 0. Force bypasses activation only. The two room results are separate supplied inputs; this does not establish that they differ during real gameplay. Actual activation including its selector-116 special branch is SCD-Q06.

**SCD-R04** Start line test [verified in 1.13c code (static)]. Runs after every SCD-R03 check (including the second room lookup) and before SCD-R05a's bound check. If lineofsight=0, make no resolution request and skip testing. Otherwise request target-coordinate resolution; if it fails, skip testing. Otherwise selectors 1–5 select masks 0x4, 0x1C09, 0x180, 0x804, 0x805 respectively; request collision testing with that mask between the unit position and the resolved `(tx,ty)`. False returns 0. Other selectors 6–255 return 0 after successful resolution. This stage does not reset mode; SCD-R02 can. Actual collision/resolution is SCD-Q10.

**SCD-R05a** Start selection/requests [verified in 1.13c code (static)]. Reject signed srvstfunc outside `[0,91)` with 0. Null slot invokes no callback and requests no mana, sets result 1, then continues to the periodic purge below (for any force value). Otherwise call start(game,unit,id,level); zero returns 0 immediately, without mana or purge. After nonzero: request mana(id,unit,game,level) only when usemanaondo clear and force=0. Then, for a null slot or nonzero callback result, for any force value: if periodic, remove every pending unit type-8 event; return the result (1 or the callback result unchanged). Actual resource/event effects are SCD-Q10.

**SCD-R05b** Nonboolean start return [verified in 1.13c code (static)]. Supplied return 7 propagates unchanged; validator failure distinguishes zero/nonzero. Individual handlers' return catalogs are SCD-Q11 and do not gate supplied propagation.

**SCD-R06** Action validation/selection [verified in 1.13c code (static)]. Absent game/unit or out-of-range identity returns 0. For players/monsters: if current skill exists, condition A clear, InTown clear and room absent/town, clear current skill, reset a player, return 0. The town clause above runs for players/monsters whether forced or not; its room handling differs from SCD-R02: an absent room lookup result clears/resets directly, and only a present room is passed to the supplied town test. Unless forced (forced skips only the ownership block below and its level requests): if the current skill has this identity, request its bonus-inclusive level and continue when it is >0 (a current skill of another identity, or no current skill, gets no level request and goes straight to the owned-skill search); otherwise find the owned skill for this identity and return 0 when it is absent or its bonus-inclusive level is <=0 (a matching current skill whose level is <=0 therefore falls through to the owned-skill search). Other kinds skip these checks. For all kinds reject signed srvdofunc outside `[0,191)`. Then players/monsters with usemanaondo set request activation whether forced or not, returning 0 on failure; other kinds and usemanaondo-clear skills make no activation request. Only after that: if itemTarget and forced are both nonzero and signed ItemEffect>1, replace index without another bound check. Initialize result=0; invoke non-null selected action(game,unit,id,level), retain i32 return, continue SCD-R07. Malformed replacement is SCD-Q05; activation/level execution is SCD-Q06/SCD-Q10.

**SCD-R07** Missile request/success [verified in 1.13c code (static)]. After callback, if srvmissile>=0 and its record exists: set latch first; then, only when forced and itemTarget are both nonzero, request target-coordinate resolution, which supplies `(tx,ty)`, and ignore its success result; then request the missile through the lob creator (lob set) or ordinary creator, and set result=1 regardless of callback/creator return. The test below is on the resolved coordinates `tx` and `ty` themselves, not on the displacement. If tx and ty are both nonzero, pass displacement `(tx-ux,ty-uy)` and destination `(tx+(tx-ux),ty+(ty-uy))`. If either is zero, pass the resolved `(tx,ty)` unchanged in the displacement position and `(0,0)` as destination, so a fully zero resolution passes four zeros. Without both controls, pass four zeros and make no resolution request. Coordinate arithmetic wraps i32. Without valid missile, zero result returns 0 before SCD-R08. Creator can fail despite dispatch result=1 (SCD-Q10). Successful creation-seeding entry consumes NRF-R17a; request alone does not establish creation.

**SCD-R08** Consuming requests [verified in 1.13c code (static)]. Consume=0 returns result immediately. Otherwise, in order, for player/monster only: read startTable[srvstfunc] without a new bound; if null or usemanaondo set, request mana(id,unit,game,level); if decquant, request quantity(game,unit). Other kinds make neither request. Every consuming kind then evaluates delay(unit,id,level). If signed delay>0 and player, request delay application(game,unit,delay) unless Skills.anim=18 and sequence-event value nonzero. Return result unchanged. Ordinary successful start plus one consuming action selects one mana request: start when populated slot/usemanaondo clear, action otherwise. Forced starts, nonconsuming paths and pre-charge failures are excluded. Actual resource effects are SCD-Q10; malformed start reads SCD-Q05.

**SCD-R09a** Frame decision [verified in 1.13c code (static)]. Listed modes share the frame handler. No current skill returns 2. Otherwise replace sequence-event value with argument's low 24 bits. If condition A is set: request target refresh(game,unit), then evaluate the supplied frame predicate; if it is 2, set B and run SCD-R09b, skipping the ordinary test. With condition A clear, neither the target refresh nor the frame predicate is requested. Otherwise (A clear, or predicate not 2) run SCD-R09b only when the latch is clear and code=1 or 2. Return 2 when supplied frame-finished predicate true, else 1; return 2 makes dispatcher request completion. Finishing predicate tests null unit, completion condition, other kind, mode 0, player mode 17 or monster mode 12. Actual predicate/completion effects are SCD-Q01/SCD-Q10.

**SCD-R09b** Frame action [verified in 1.13c code (static)]. No current skill does nothing. Otherwise obtain identity/bonus-inclusive level, call action(consume=1,forced=0,itemTarget=0), ignore result, then SCD-R11 with that identity/level and ordinary repeat kind. Action zero does not prevent scheduling.

**SCD-R09c** Frame production/special conditions [unresolved]. Actual animation-frame codes/arguments and complete condition-A/B/predicate effects are SCD-Q01. With supplied ordinary code-1/2 events and no other latch clear, missile dispatch runs at most once per mode entry; a callback leaving latch clear can run each such event. Exceptional condition-A route bypasses the latch gate.

**SCD-R10** Other callers [asm-only]. At the immediate path's reached dispatch branch, ItemCheckStart requests forced start; zero aborts. Action uses consume=0,forced=1,itemTarget=caller input. Target/latch restore saved entry state. Two other paths call action with consume=1,forced=1,itemTarget=0; monster frame handlers also reach frame action. Actual roots/eligibility/roles remain SCD-Q04; no general monster/item trigger follows from labels.

**SCD-R11** Periodic scheduler [verified in 1.13c code (static)]. Invalid identity/record returns 0. Neither periodic nor aura returns 0 without calculation/scheduling. Otherwise evaluate perdelay(unit,id,level), set signed `p=max(value,5)`, then `n=i32(frame+p-1)` and `due=i32(i32(truncTowardZero(n/p)*p)+1)`. Ordinary repeat kind removes all unit type-8 events, schedules type 8 at due carrying (id,level). Nonzero repeat kind removes only unit type-8 events carrying identity -1, schedules (-1,0). Return 1. No direct draws. Actual event writes are SCD-Q10; repeat eligibility SCD-Q09. Positive nonoverflowing scheduling aligns to `k*p+1`; other inputs retain wrapped signed arithmetic.

**SCD-R12** Returns [verified in 1.13c code (static)]. Start: 0 refusal/failure, 1 null slot, otherwise unchanged callback result. Action: 0 refusal or zero callback/no missile; otherwise callback result, overridden to 1 by valid missile request. Mode entry/frame action do not branch on those returns; all writes by the called local stages persist, including a validator's mode reset. Zero does not prove callback had no effects; zero/no-missile skips later charge requests.

**SCD-R13** Draw boundary [verified in 1.13c code (static)]. Validator, start/action dispatch, direct table dispatch and scheduler draw no random numbers directly. Count callbacks/dependencies separately in call order. Successful missile creation seeding steps game seed once (NRF-R17a); later initialization may draw separately. Failed creation can perform no seed step despite dispatch result=1.

**SCD-R14a** Direct action table [verified in 1.13c code (static)]. Supplied `0<=index<191` and non-null slot invokes action(game,unit,id,level); otherwise nothing. No ownership/town/resource/missile/delay stages run. Two missile callers select Param2/sHitPar1 respectively; their eligibility is outside this supplied-index stage.

**SCD-R14b** Progression/active-state calls [asm-only]. Reached progression branches select `srvprgfunc1..3`, requiring `0<index<191`/non-null slot. Active-state path uses aurastate's srvactivefunc with `0<=index<191`/non-null slot. Surrounding eligibility/selection/state effects are SCD-Q09, owned with progression/auras.

**SCD-R15** Command target check [asm-only]. Clear parsed target outputs first. Length!=9 returns 3; unsigned kind>=6 returns 2. Otherwise request target validation with range 50: missing unit returns 1, inventory item 0, different areas 2, otherwise range result. Only result 0 stores parsed kind/id. Eight skill-command callers require one of two slots, increment stat 328 and hand to mode selection. Packet identities, slot/stat meanings, range execution and mode selection are SCD-Q02/SCD-Q07.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SCD-R01 | Mode requests/latch clear | verified in 1.13c code (static) | SCD-Q01/SCD-Q10: preparation effects | SCD-V09 |
| SCD-R02 | Ordered supplied validation | verified in 1.13c code (static) | SCD-Q03/SCD-Q10: actual predicates/level/reset | SCD-V03, SCD-V06, SCD-V09, SCD-V12 |
| SCD-R03 | Start preconditions | verified in 1.13c code (static) | SCD-Q06/SCD-Q10: activation/room | SCD-V06, SCD-V09, SCD-V12, SCD-V17 |
| SCD-R04 | Line selector/mask | verified in 1.13c code (static) | SCD-Q10: collision/target execution | SCD-V07 |
| SCD-R05a | Start/charge/purge | verified in 1.13c code (static) | SCD-Q10: resource/event writes | SCD-V02, SCD-V04, SCD-V09, SCD-V11, SCD-V17 |
| SCD-R05b | Nonboolean propagation | verified in 1.13c code (static) | none locally; SCD-Q11 is producer | SCD-V11 |
| SCD-R06 | Action selection | verified in 1.13c code (static) | SCD-Q05: malformed replacement; SCD-Q06/SCD-Q10: dependencies | SCD-V01, SCD-V08, SCD-V12 |
| SCD-R07 | Missile request/coordinates | verified in 1.13c code (static) | SCD-Q10: actual creator effects | SCD-V01, SCD-V09, SCD-V10, SCD-V13 |
| SCD-R08 | Consuming request order | verified in 1.13c code (static) | SCD-Q05: malformed read; SCD-Q10: writes | SCD-V02, SCD-V04, SCD-V13, SCD-V18 |
| SCD-R09a | Supplied frame decision | verified in 1.13c code (static) | SCD-Q01/SCD-Q10: predicate/completion | SCD-V09, SCD-V14 |
| SCD-R09b | Frame arguments/ignored result | verified in 1.13c code (static) | SCD-Q10: level/event execution | SCD-V05, SCD-V09, SCD-V14 |
| SCD-R09c | Frame production | unresolved | SCD-Q01: animation-to-event integration | SCD-V09 supplied events only; SCD-O01 |
| SCD-R10 | Other caller branches | asm-only | SCD-Q04: eligibility/roots | none |
| SCD-R11 | Due arithmetic/requests | verified in 1.13c code (static) | SCD-Q09/SCD-Q10: repeat/event execution | SCD-V05, SCD-V15 |
| SCD-R12 | Return meanings | verified in 1.13c code (static) | none at supplied stages | SCD-V01, SCD-V03, SCD-V11, SCD-V14 |
| SCD-R13 | Direct draws/seeding | verified in 1.13c code (static) | SCD-Q10: whole dependency draws | SCD-V10, SCD-V13 |
| SCD-R14a | Direct table call | verified in 1.13c code (static) | none for supplied index | SCD-V16 |
| SCD-R14b | Additional users | asm-only | SCD-Q09: progression/aura eligibility | none |
| SCD-R15 | Command check | asm-only | SCD-Q02/SCD-Q07: command/range/mode | none |

## 5. Outputs and side effects

| Effect | Rule |
|---|---|
| Entry-preparation, route-specific target/mode and later preparation requests; current skill cleared/failure reset | SCD-R01, SCD-R02, SCD-R03, SCD-R06 |
| Callback requests/declared effects | SCD-R05a, SCD-R06, SCD-R14a; gated SCD-R10/SCD-R14b |
| Latch, sequence-event value, condition B on the current skill node | SCD-R01, SCD-R07, SCD-R09a; gated SCD-R10 restores |
| Missile request and resulting missile count; supplied successful seeding/creation changes game/missile seed | SCD-R07, SCD-R13, NRF-R17a; full creation gated |
| Mana/quantity/delay requests and declared supplied writes | SCD-R05a, SCD-R08 |
| Type-8 purge/replacement requests with due/payload | SCD-R05a, SCD-R11 |
| Completion request | SCD-R09a |
| Start/action/validator numeric returns; frame-handler 1/2; scheduler 0/1 | SCD-R02, SCD-R03, SCD-R05a, SCD-R05b, SCD-R06, SCD-R07, SCD-R08, SCD-R09a, SCD-R11, SCD-R12 |
| Dependency requests asserted by vectors: room lookup, town test, target eligibility, target refresh, party rejection, bonus-inclusive level, find owned skill, activation, target-coordinate resolution, collision test, frame predicate, frame-finished predicate, perdelay/delay evaluation, delay application | SCD-R02, SCD-R03, SCD-R04, SCD-R06, SCD-R07, SCD-R08, SCD-R09a, SCD-R09b, SCD-R11 |
| Parsed target outputs/stat 328 increment | gated SCD-R15 |

No named client message or persistence write is established by verified dispatch stages.

## 6. Branches and edge cases

Null start succeeds without start charges; null action fails without charges unless missile overrides result. Null target on unit-target mode entry is fatal after entry preparation and the mode request; it is not a normal start rejection. ItemEffect replacement follows original-selector bounds without another check. Other kinds skip action town/ownership/activation and mana/quantity, but consuming delay calculation still runs; only players request positive delay. Charged start uses base level, frame action bonus-inclusive level. Validator tests charges==0, while actual activation's positive-charge test is distinct. Exceptional frame route bypasses latch; argument high bits are discarded. Difficulty/PvP/melee class adds no direct branch. Malformed reads have no guessed safe outcome (SCD-Q05). Arithmetic wraps before signed division without saturation.

## 7. Open questions

Attempts/blocked boundaries are in the [gap file](../plans/gaps/skill-cast-start-and-do-dispatch.md).

- **SCD-Q01** [unresolved] Animation codes/arguments and special frame-predicate/conditions. Read animation accessors/predicate callees through writes, then SCD-O01; supplied decisions remain available.
- **SCD-Q02** [asm-only] Command identities, slot assignment, stat 328 meaning, range behavior/import-argument discrepancy. Resolve packet roots/imports in networking scope.
- **SCD-Q03** [unresolved] Actual target/party predicates, target-refresh request and target-clearing effects. Follow hostile/party/monster-owner/targetability callees and imports.
- **SCD-Q04** [asm-only] Immediate/other caller roots, eligibility, targets and monster mode entry. Trace wrappers and dispatch/event roots.
- **SCD-Q05** [unresolved] Malformed selector reads/replacements; authenticated original-profile Skills data and compiled correspondence needed for prevalence, not safety.
- **SCD-Q06** [unresolved] Actual charged-level/mana activation and selector-116 special predicate; complete dependency contract required.
- **SCD-Q07** [asm-only] Command-selected mode; follow usability, mode lookup, fallback/target redirection. Do not infer from anim alone.
- **SCD-Q09** [asm-only] Repeat/progression/aura eligibility; follow state/node/removal producers through repeat dispatch.
- **SCD-Q10** [unresolved] Actual mode preparation/reset, room/target/collision, level, resources, event storage and creator effects. Verified supplied stages do not establish them; companion names investigated/missing chains.
- **SCD-Q11** [recovered; unverified] Individual start-handler nonboolean return catalog; outside-boundary producer, no gate on SCD-R05b.
- **SCD-Q12** [unresolved] All gameplay confirmation pending; blocks runtime confirmation only. Initial observations below are not exhaustive.

## 8. Test vectors

### 8.1 Executable vectors

Each case independently starts from these synthetic defaults unless overridden: present game/unit/room/Skills record; player kind 0, mode 10; skill id 20,count 400,current/owned id 20,invocation level 3,bonus-inclusive level result 3,base level 2,marker -1,charges 5,no other skills. Every room lookup returns the present non-town room unless a per-call sequence is supplied. InTown=true,TargetAlly=false,no target,target eligibility=true,party rejection=0,activation=true,collision=true; position (10,20); target-coordinate resolution supplies (0,0) and fails unless a case supplies a resolved point. srvstfunc=1/srvdofunc=1 populated, callbacks return 1. ItemEffect=0,srvmissile=-1,lineofsight=0,anim=10; other flags false; delay/perdelay=0,frame=100; force=0,consume=1,forced=0,itemTarget=0,ordinary repeat kind; latch/A/B=false,frame-finished=false,sequence-event value=0; events/logs empty; stat 8 `mana`=2560 raw 8.8,quantity=5.

Calls between this spec's local rules execute those rules; only the named outside-boundary dependency operations and individual skill callbacks are supplied. Supplied operations log calls and declared results, drawing nothing except V10. Callbacks log (game,unit,id,level), return supplied value and have no other effects unless stated. Unless a case explicitly overrides their supplied effects,mana requests subtract cost 256 and quantity requests subtract 1; delay-application requests append their value to the delay log; target-refresh requests only log; reset requests set mode 5. Supplied mode-setting requests persist the requested mode and have no other effects in these fixtures. Supplied point-target setting stores the given point and clears unit-target selection; unit-target setting stores that selection without changing the stored point. Target/skill/latch/condition/event writes persist within a case. Entry preparation and the later mode-state/event/animation preparation operations only log requests in these fixtures. These supplied effects test ordering, with actual integrations SCD-Q10. Unmentioned state stays at input. Shorthand start(20,3)/action(20,3) includes game/unit first; resource requests include the unit and game as specified by the rules. Call lists assert the listed ordered subsequence unless a case says exactly/only; omitted pure getters are not asserted absent. Pure getters are reads of current skill, skill identity, charge marker/count, records, flags, unit kind/mode/position, stored target and sequence-event value. Every other operation in section 5's dependency-request row, every callback and every effect request is an asserted request. Pending-event lists are the unit's own pending events, compared as unordered sets: event storage order is not part of this contract, and another unit's pending events are never touched. No calls excludes pure getters and means no callback, predicate, evaluation or effect request. Every case has runtime_observed: no.

- **SCD-V01** Action bounds/null slots (SCD-R06, SCD-R07, SCD-R12; boundary: supplied-input stage)
  - Inputs: action from defaults, no missile.
  - Seeds: no draw.
  - Draws: none.
  - Trace: bound first; zero/no missile exits before charges.
  - (a) Inputs: srvdofunc=191. Draws: none. Outputs: return 0,mana=2560; exactly one request, the ownership level(current node 20,bonus=true) returning 3 (ownership precedes the bound check); no activation,callback,missile,resource or delay request.
  - (b) Inputs: srvdofunc=160. Draws: none. Outputs: return 0,no callback/resource/delay,mana=2560.
  - (c) Inputs: srvdofunc=152,callback return=7. Draws: none. Outputs: return 7; action(20,3),eval delay(20,3); mana=2560.
  - (d) Inputs: srvdofunc=190 (in-range null slot),srvmissile=7 record present,lob=false; the supplied ordinary creator returns 0 and creates nothing. Draws: none. Outputs: return 1,latch=true,mana=2560,game seed unchanged {1,0x29A}; calls ordinary creator(7,displacement=(0,0),destination=(0,0)),eval delay(20,3); no action callback,resolution request,mana,quantity or delay application.
  - Discriminates: bound/null cannot invoke; return 7 cannot normalize to 1; an action bound of 190 instead of 191 returns 0 in (d).
  - runtime_observed: no

- **SCD-V02** One mana request (SCD-R05a, SCD-R08; boundary: supplied-input stage)
  - Inputs: successful start then action,usemanaondo=false,consume=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: null start defers charge; populated successful start charges and action skips it.
  - (a) Inputs: srvstfunc=0. Draws: none. Outputs: returns (1,1); after activation, calls action(20,3),mana(20,3),eval delay; mana=2304.
  - (b) Inputs: srvstfunc=1. Draws: none. Outputs: returns (1,1); after activation, calls start(20,3),mana(20,3),action(20,3),eval delay; mana=2304.
  - Discriminates: ordered log distinguishes charge stage; 2304 rejects double deduction.
  - runtime_observed: no

- **SCD-V03** Start failure/reset (SCD-R02, SCD-R12; boundary: supplied-input stage)
  - Inputs: validator,start return=0,mode=10.
  - Seeds: no draw.
  - Draws: none.
  - Trace: target/level/activation pass; callback zero skips charge, resets only player.
  - (a) Inputs: kind=0. Draws: none. Outputs: return 0,mode=5,mana=2560; calls target eligibility,level(20,bonus=true),activation,start(20,3),reset(5,1).
  - (b) Inputs: kind=1. Draws: none. Outputs: return 0,mode=10,mana=2560; calls through start as above,no reset.
  - (c) Inputs: kind=0,start return=7. Draws: none. Outputs: return 7,mode=10,mana=2304; calls through start(20,3),then mana(20,3); no reset.
  - Discriminates: failed start cannot charge; monster cannot reset; resetting on any result other than exactly 0 fails (c).
  - runtime_observed: no

- **SCD-V04** usemanaondo/consume (SCD-R05a, SCD-R08; boundary: supplied-input stage)
  - Inputs: usemanaondo=true,successful populated start then action.
  - Seeds: no draw.
  - Draws: none.
  - Trace: no start charge; action activation required; consume controls resources.
  - (a) Inputs: consume=1. Draws: none. Outputs: returns (1,1),mana=2304; activation,start,activation,action,mana,eval delay.
  - (b) Inputs: consume=0. Draws: none. Outputs: returns (1,1),mana=2560; activation,start,activation,action only.
  - Discriminates: ordered logs enforce charge stage and consume=0 suppression.
  - runtime_observed: no

- **SCD-V05** Period alignment/ignored failure (SCD-R11, SCD-R09b; boundary: supplied-input stage)
  - Inputs: execute frame action SCD-R09b and the actual local action dispatcher,logging its (20,3,1,0,0) invocation; srvdofunc=160 (null slot),srvmissile=-1,periodic=true; events (8,30,1,200),(8,-1,0,250),(9,40,2,300), tuples (type,id,level,due). In (a)–(c) and (e) both input type-8 events are removed; only the new type-8 event and the type-9 event remain.
  - Seeds: no draw.
  - Draws: none.
  - Trace: action ownership checks pass; null slot 160 and no missile return 0 before resource/delay work. Frame action ignores this computed zero,then clamps period and replaces type 8 only. Every case has one action-dispatch invocation and no action callback,resource request or delay evaluation. In (a)–(c) the ordered log is action dispatch(20,3,1,0,0)→eval perdelay→remove type 8→schedule; in (d) the action dispatch is the only request; scheduling before the action dispatch fails these orders.
  - (a) Inputs: perdelay=3,frame=101. Draws: none. Outputs: action-dispatch return 0; exactly one eval perdelay(unit,id=20,level=3) returning 3; p=5,n=105,quotient=21,due=106,scheduler return 1; event (8,20,3,106),input type 9 unchanged; no action callback,resource request or delay evaluation.
  - (b) Inputs: perdelay=8,frame=100. Draws: none. Outputs: action-dispatch return 0; exactly one eval perdelay(unit,id=20,level=3) returning 8; p=8,n=107,quotient=13,due=105,scheduler return 1; event (8,20,3,105),type 9 unchanged; no action callback,resource request or delay evaluation.
  - (c) Inputs: perdelay=8,frame=96. Draws: none. Outputs: action-dispatch return 0; exactly one eval perdelay(unit,id=20,level=3) returning 8; p=8,n=103,quotient=12,due=97,scheduler return 1; event (8,20,3,97),type 9 unchanged; no action callback,resource request or delay evaluation.
  - (d) Inputs: periodic=false,aura=false,perdelay=8,frame=96. Draws: none. Outputs: action-dispatch return 0,scheduler return 0; no perdelay evaluation/event replacement; all three input events unchanged; no action callback,resource request or delay evaluation.
  - (e) Inputs: srvdofunc=1 (populated),action callback returns 0,perdelay=8,frame=100. Draws: none. Outputs: action-dispatch return 0 after exactly one action callback(20,3); no mana,quantity,delay evaluation or delay application; exactly one eval perdelay(unit,id=20,level=3) returning 8; due=105,scheduler return 1; events (8,20,3,105),(9,40,2,300).
  - Discriminates: due differs from frame+p; action zero cannot prevent scheduling,whether from a null slot or a populated callback returning 0 (e); charging after a zero callback without missile fails (e); sparing the identity -1 type-8 event fails ordinary replacement.
  - runtime_observed: no

- **SCD-V06** Town start (SCD-R02, SCD-R03; boundary: supplied-input stage)
  - Inputs: validator,town room,mode=10,current skill present.
  - Seeds: no draw.
  - Draws: none.
  - Trace: validator town gate precedes dispatch; allowed case reaches line/start.
  - (a) Inputs: InTown=false. Draws: none. Outputs: return 0,current skill absent,mode=5,mana=2560; exact request log target eligibility→level(20,bonus=true)→target refresh(game,unit)→room lookup→town test→reset(5,1); no activation/start.
  - (b) Inputs: InTown=true. Draws: none. Outputs: return 1,current skill unchanged,mode=10,mana=2304; target eligibility,level,activation,start(20,3),mana.
  - (c) Inputs: InTown=false,monster kind=1. Draws: none. Outputs: return 0,current skill absent,mode=10,mana=2560; target eligibility→level→target refresh→room lookup→town test→clear current skill; no reset,activation or start.
  - Discriminates: forbidden use clears skill and skips callback; allowed use continues; resetting a monster fails (c).
  - runtime_observed: no

- **SCD-V07** Line mask/reset boundary (SCD-R04, SCD-R02; boundary: supplied-input stage)
  - Inputs: direct start,resolved point (30,40),collision=false.
  - Seeds: no draw.
  - Draws: none.
  - Trace: activation before line; valid selector invokes collision; direct failure leaves mode.
  - (a) Inputs: lineofsight=2. Draws: none. Outputs: return 0,mode=10,mana=2560; collision mask 0x1C09 from (10,20) to (30,40); no start/charge.
  - (b) Inputs: lineofsight=6. Draws: none. Outputs: return 0,mode=10,mana=2560; no collision/start/charge.
  - (c) Inputs: lineofsight=0. Draws: none. Outputs: return 1,mana=2304; no collision; activation,start,mana.
  - (d) Inputs: lineofsight=2,validator entry. Draws: none. Outputs: return 0,mode=5,mana=2560; mask/point as (a),then reset(5,1).
  - (e) Inputs: lineofsight=6,target unresolved (resolution supplies (0,0) and fails). Draws: none. Outputs: return 1,mana=2304; no collision; activation,resolution,start,mana.
  - (f) Inputs: collision=true; run once each with lineofsight=1,2,3,4,5. Draws: none. Outputs: each run returns 1,mana=2304,mode=10; one collision request from (10,20) to (30,40) with mask 0x4,0x1C09,0x180,0x804,0x805 respectively,then start(20,3),mana(20,3).
  - (g) Inputs: lineofsight=2,resolution supplies (0,40) and fails. Draws: none. Outputs: return 1,mana=2304; resolution,no collision request,start(20,3),mana(20,3).
  - Discriminates: a resolution success test of either coordinate nonzero fails (g); returning 0 for a valid selector regardless of the collision result,or any wrong selector-to-mask binding,fails (f); unresolved target skips selector 6; validator alone resets line failure.
  - runtime_observed: no

- **SCD-V08** ItemEffect gates (SCD-R06; boundary: supplied-input stage)
  - Inputs: srvdofunc=10,ItemEffect=3,consume=0; callbacks 2/3/10 return 7.
  - Seeds: no draw.
  - Draws: none.
  - Trace: original bound before conditional replacement.
  - (a) Inputs: forced=1,itemTarget=1. Draws: none. Outputs: slot 3 once with (game,unit,20,3),return 7,mana=2560.
  - (b) Inputs: forced=1,itemTarget=0. Draws: none. Outputs: slot 10 once,return 7,mana=2560.
  - (c) Inputs: forced=1,itemTarget=1,ItemEffect=1. Draws: none. Outputs: slot 10 once,return 7,mana=2560.
  - (d) Inputs: forced=0,itemTarget=1. Draws: none. Outputs: slot 10 after ownership level check,return 7,mana=2560.
  - (e) Inputs: srvdofunc=191,forced=1,itemTarget=1,ItemEffect=3. Draws: none. Outputs: return 0,mana=2560,current skill unchanged; no action callback,missile,resource request,delay evaluation or delay application. The invalid original selector rejects before replacement,so slot 3 is not invoked.
  - (f) Inputs: forced=1,itemTarget=1,ItemEffect=2. Draws: none. Outputs: slot 2 once with (game,unit,20,3),return 7,mana=2560; slot 10 not invoked.
  - (g) Inputs: forced=1,itemTarget=1,ItemEffect=-1. Draws: none. Outputs: slot 10 once,return 7,mana=2560.
  - Discriminates: an unsigned ItemEffect comparison fails (g); both controls and ItemEffect>1 necessary; a threshold of >2 fails (f); replacing an invalid original selector before its bound check fails (e). Malformed replacement accesses remain SCD-Q05.
  - runtime_observed: no

- **SCD-V09** Mode entry, validation, ordinary suppression and frame controls (SCD-R01, SCD-R02, SCD-R03, SCD-R05a, SCD-R07, SCD-R08, SCD-R09a, SCD-R09b, SCD-R09c; boundary: supplied-input stage)
  - Inputs: present player unit,initial mode=10,requested mode=10,initial latch=true; initial selected unit target absent,stored point=(50,60). Execute the local validator/start stages and frame/action stages when reached; only their declared outside-boundary operations and skill callbacks are supplied. Entry/later preparation only log. Start slot 1 is populated,start callback returns 1,usemanaondo=false,delay=0,periodic=false,aura=false; eligibility=true,activation=true,effective level=3. Cases (a)/(b) use point-target entry (30,40),initial mana=2816 and events (1,0),(1,0),(2,0),A=false,frame-finished=false. In (a), the supplied creator reaches successful NRF-R17a creation-seeding entry once and creates one missile, with no later initialization draws. Cases (c)/(d) override route,initial mana and event sequence below.
  - Seeds: game={1,0x29A} independently in each case. Case (a) supplies successful NRF-R17a creation-seeding entry and no later initialization draws; other cases create no missile.
  - Draws: (a) one game step with new low 0x6AC6935F; (b–d) none. No direct dispatcher draw.
  - Trace: point entry logs set point(30,40)→entry preparation(game,unit)→mode(10)→set unit target(absent)→mode-state reset→remove events→prepare animation→clear latch→execute validator. In (a)/(b), validator executes target eligibility→level(20,true)→target refresh(game,unit)→room lookup→activation→start(20,3)=1→mana(20,3),so its return is 1 and mana becomes 2816-256=2560. Each reached frame action invokes action dispatch(20,3,1,0,0); populated start slot skips later mana requests, successful actions evaluate delay,and periodic scheduling returns 0 without evaluation/events. In (a), successful creation-seeding steps the game seed to {0x6AC6935F,0} and initializes the supplied missile seed to {0x6AC6935F,666}; this consumes NRF-R17a rather than treating successful creation as draw-free. Unit-target failure and fatal routes are explicit below.
  - (a) Inputs: srvmissile=7,record present,action return=0. Draws: one game step at the supplied creation-seeding entry. Outputs: full point-entry/validator log as above; validator return=1,one start callback and one start mana request; mode=10,point=(30,40),unit target absent,latch false at validator entry. Frame returns (1,1,1); exactly one action-dispatch invocation (20,3,1,0,0),one action callback,one ordinary creator receiving missile=7,displacement=(0,0),destination=(0,0),and one eval delay(20,3), ordered dispatch→callback→ordinary creator→delay evaluation. Both forced and itemTarget are 0,so the stored point (30,40) does not produce nonzero creator coordinates. Action result=1,one missile; game seed={0x6AC6935F,0},missile seed={0x6AC6935F,666}; final latch=true,mana=2560; no action-stage mana/quantity/delay-application or periodic-event requests.
  - (b) Inputs: srvmissile=-1,action return=1,callback leaves latch clear. Draws: none. Outputs: full point-entry/validator log as above; validator return=1,one start callback and one start mana request; mode=10,point=(30,40),unit target absent,latch false at validator entry. Frame returns (1,1,1); exactly three dispatch invocations (20,3,1,0,0),three action callbacks each result=1,three eval delay(20,3), ordered dispatch→callback→delay evaluation per event; no missile. Final latch=false,mana=2560; no action-stage mana/quantity/delay-application or periodic-event requests.
  - (c) Inputs: unit-target entry,target is present monster kind=1,id=99; initial mana=2560,party rejection=0,start callback returns 0,empty frame-event sequence. Draws: none. Outputs: exact entry log entry preparation(game,unit)→mode(10)→set unit target(monster 99)→mode-state reset→remove events→prepare animation→clear latch→execute validator; validator then logs target eligibility→level(20,true)→target refresh(game,unit)→party rejection→room lookup→activation→start(20,3)=0→reset(mode=5,force=1),and returns 0. Final mode=5,unit target=monster 99,point=(50,60),latch=false,current skill unchanged,mana=2560. No point-target setter,assertion,action,creator,mana,quantity or delay request; entry does not restore mode 10 after the reset.
  - (d) Inputs: unit-target entry,target absent; initial mana=2560,empty frame-event sequence; pending event (type=8,id=30,level=1,due=200). Draws: none. Outputs: exact log entry preparation(game,unit)→mode(10)→fatal assertion. No numeric return. Mode=10,unit target absent,point=(50,60),latch=true,current skill unchanged,mana=2560,input event unchanged. No target setter,mode-state reset,event removal,animation preparation,latch clear,validator or action request follows the fatal assertion.
  - Discriminates: full entry logs reject omitted preparation or swapping point/mode requests; fatal entry rejects an early target check or normal zero return. Actual validation must charge once on successful starts and reset mode 5 on failed player start; treating validator return 0 as effect-free or restoring mode 10 fails (c). Unit selection preserves stored point. Exact action logs reject wrong controls; consume=0 omits delay evaluation. Missile overrides callback zero and suppresses later ordinary events. Its supplied successful seeding uses NRF-R17a exactly once. Actual outside-boundary preparation,frame production,creator-entry eligibility and later initialization remain unasserted.
  - runtime_observed: no

- **SCD-V10** Creation seeding/no direct draw (SCD-R07, SCD-R13; boundary: supplied-input stage)
  - Inputs: srvdofunc=160,srvmissile=7 record present,consume=0,target point (30,40); creator succeeds at NRF-R17a missile seeding once,no later initialization draws.
  - Seeds: game={1,0x29A}; missile seed absent.
  - Draws: one game step,no modulus,result 0x6AC6935F; zero direct dispatcher draws.
  - Trace: null slot zero; missile forces 1; NRF-R01 low=0x6AC6935F,high=0; initialize missile from low.
  - Outputs: one missile,return 1,latch=true; game={0x6AC6935F,0}; missile seed={0x6AC6935F,666}; no callback/resource request,mana=2560.
  - Discriminates: extra direct draw changes final game seed; complete initialization draws not asserted.
  - runtime_observed: no

- **SCD-V11** Nonboolean start/purge (SCD-R05a, SCD-R05b; boundary: supplied-input stage)
  - Inputs: direct start return=7,periodic=true; this unit's events (8,30,2,200),(8,-1,0,250),(9,40,1,300), tuples (type,id,level,due); another unit holds pending (8,50,1,500). Every case leaves the other unit's event unchanged; a purge removes both of this unit's type-8 events.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) activation,start=7,charge,purge,retain 7. (b) null slot skips callback/mana but still purges. (c) zero callback returns before charge/purge.
  - (a) Inputs: as above. Draws: none. Outputs: return 7,mana=2304; calls activation,start(20,3),mana(20,3),remove type 8; only (9,40,1,300) remains.
  - (b) Inputs: srvstfunc=90 (in-range null slot). Draws: none. Outputs: return 1,mana=2560; calls activation,remove type 8; no start callback or mana request; only (9,40,1,300) remains.
  - (e) Inputs: srvstfunc=90,force=1. Draws: none. Outputs: return 1,mana=2560; remove type 8 only; no activation,start callback or mana request; only (9,40,1,300) remains.
  - (c) Inputs: start return=0. Draws: none. Outputs: return 0,mana=2560; calls activation,start(20,3); no mana or removal request; both input events remain.
  - (d) Inputs: start return=1,periodic=false. Draws: none. Outputs: return 1,mana=2304; calls activation,start(20,3),mana(20,3); no removal request; both input events remain.
  - Discriminates: boolean normalization/early purge fail return/order assertions (a); an unconditional purge fails (d); a start bound of 90, or a null slot that skips the purge, fails (b); purging after a zero callback fails (c).
  - runtime_observed: no

- **SCD-V12** Validation order/bypass (SCD-R02, SCD-R03, SCD-R06; boundary: supplied-input stage)
  - Inputs: specified stage from defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: first failed gate; charged start uses base level.
  - (a) Inputs: validator,marker=10,charges=0. Draws: none. Outputs: return 0,mode=10,mana=2560; eligibility/level only,no activation/start/reset.
  - (b) Inputs: validator,marker=10,charges=1. Draws: none. Outputs: return 1,mana=2304; eligibility,level,activation,start(20,2),mana(20,2).
  - (c) Inputs: direct start,force=1,activation=false. Draws: none. Outputs: start(20,3),return 1,mana=2560; no activation/mana.
  - (d) Inputs: action,forced=1,usemanaondo=true,activation=false. Draws: none. Outputs: return 0,mana=2560; activation only,no callback/resource.
  - (e) Inputs: validator,target eligibility=false. Draws: none. Outputs: return 0,mode=5,mana=2560; eligibility,reset(5,1); current skill unchanged.
  - (f) Inputs: action,forced=0,current skill id=21; owned skills 20 and 21 exist,find owned skill(20) returns node 20 with effective level 3. Draws: none. Outputs: return 1,mana=2560,current skill id remains 21; calls find owned skill(20),level(node 20,bonus=true),action(20,3),eval delay(20,3); no resource application.
  - (g) Inputs: action,forced=0,current skill id=20 (the owned node 20); every bonus-inclusive level request returns 0. Draws: none. Outputs: return 0,mana=2560,current skill unchanged,mode=10; ordered calls level(current node 20,bonus=true)→find owned skill(20)→level(node 20,bonus=true); no action callback,missile,resource request or delay evaluation.
  - (h) Inputs: action,forced=0,current skill id=20 whose bonus-inclusive level request returns 0; find owned skill(20) returns owned node 20 whose level request returns 5 (supplied per-call results 0 then 5). Draws: none. Outputs: return 1,mana=2560,current skill unchanged; ordered calls level(current node 20,bonus=true)=0→find owned skill(20)→level(node 20,bonus=true)=5→action(20,3)→eval delay(20,3); no resource application. The callback receives invocation level 3,not 5.
  - (i) Inputs: action,forced=1,every bonus-inclusive level request would return 0. Draws: none. Outputs: return 1,mana=2560; calls action(20,3),eval delay(20,3); no level or find-owned-skill request.
  - (j) Inputs: validator,marker=10,charges=-1. Draws: none. Outputs: return 1,mana=2304; eligibility,level,activation,start(20,2),mana(20,2); no reset.
  - (k) Inputs: validator,target present monster kind=1,id=99 (differs from unit),TargetAlly=false,party rejection=1. Draws: none. Outputs: return 0,mode=10,mana=2560,current skill unchanged; exact request log target eligibility→level(20,bonus=true)→target refresh(game,unit)→party rejection(game,unit,20); no room lookup,activation,start or reset.
  - (l) Inputs: validator,target present monster kind=1,id=99,TargetAlly=true,party rejection=1. Draws: none. Outputs: return 1,mana=2304; no party-rejection request; activation,start(20,3),mana(20,3).
  - (m) Inputs: validator,selected target is the unit itself,party rejection=1. Draws: none. Outputs: return 1,mana=2304; no party-rejection request; activation,start(20,3),mana(20,3).
  - (n) Inputs: validator,initial selected target monster kind=1,id=99,TargetAlly=false,party rejection=1; the supplied target refresh sets the selected target to absent. Draws: none. Outputs: return 1,mana=2304,unit target absent; target refresh precedes start; no party-rejection request.
  - (o) Inputs: validator,marker=0,charges=0. Draws: none. Outputs: return 0,mode=10,mana=2560; eligibility,level only; no activation,start or reset.
  - (p) Inputs: action,forced=0,usemanaondo=true,srvdofunc=191,activation=true. Draws: none. Outputs: return 0,mana=2560; exactly one request,level(current node 20,bonus=true); no activation request.
  - (q) Inputs: validator,monster kind=1,target eligibility=false. Draws: none. Outputs: return 0,mode=10,mana=2560; eligibility only; no reset.
  - Discriminates: ignoring TargetAlly fails (l); ignoring the self-target exclusion fails (m); reading the target before the refresh fails (n); a marker test other than !=-1 fails (o); requesting activation before the action bound fails (p); resetting a non-player fails (q); forced running the ownership block fails (i); a charges<=0 test fails (j); resetting or continuing after party rejection fails (k); charge zero does not reset; force skips start activation but forced does not skip action activation; charged start level=2; a mismatched current skill must fall back to the requested owned skill rather than refuse or execute id 21; a matching current skill with level 0 must fall through to the owned-skill search and its second level request rather than return 0 immediately (g), and a positive owned-skill level then proceeds with the invocation level (h); returning 0 after the second request regardless of its result fails (h).
  - runtime_observed: no

- **SCD-V13** Missile coordinates/resource order (SCD-R07, SCD-R08, SCD-R13; boundary: supplied-input stage)
  - Inputs: forced=1,itemTarget=1,ItemEffect=0,srvmissile=7 record present,target=(30,40),lob=true,callback return=7,srvstfunc=30 (in-range null slot),decquant=true,delay=12,anim=18,sequence-event value=0; creator returns 0,creates nothing,and logs the latch value it observes when called.
  - Seeds: game={1,0x29A}; no draw.
  - Draws: none.
  - Trace: in (a), displacement=(30-10,40-20)=(20,20),destination=(30+20,40+20)=(50,60). In (b), the resolution reports failure but supplies (0,40); x zero selects the unchanged-pair path, displacement=(0,40),destination=(0,0). In (d), supplied (30,0) likewise gives displacement=(30,0),destination=(0,0). In (c), signed coordinate operations wrap i32: x displacement is i32(2147483647-(-2147483648))=-1,y displacement=2147483647; destination=(2147483646,i32(2147483647+2147483647))=(2147483646,-2). Every supplied failed creator call still forces 1,then consuming requests.
  - (a) Inputs: common resolved point (30,40). Draws: none. Outputs: return 1,latch=true,mana=2304,quantity=4,delay log=[12],zero missiles; calls action,lob missile(7,displacement=(20,20),destination=(50,60)),mana,quantity,eval delay,delay(12); game seed unchanged {1,0x29A}.
  - (b) Inputs: target-coordinate result (0,40),target resolution fails. Draws: none. Outputs: return 1,latch=true,mana=2304,quantity=4,delay log=[12],zero missiles; calls action,lob missile(7,displacement=(0,40),destination=(0,0)),mana,quantity,eval delay,delay(12); game seed unchanged {1,0x29A}.
  - (c) Inputs: unit position=(-2147483648,0),resolved target=(2147483647,2147483647). Draws: none. Outputs: return 1,latch=true,mana=2304,quantity=4,delay log=[12],zero missiles; calls action,lob missile(7,displacement=(-1,2147483647),destination=(2147483646,-2)),mana,quantity,eval delay,delay(12); game seed unchanged {1,0x29A}.
  - (d) Inputs: target-coordinate result (30,0),target resolution fails. Draws: none. Outputs: return 1,latch=true,mana=2304,quantity=4,delay log=[12],zero missiles; calls action,lob missile(7,displacement=(30,0),destination=(0,0)),mana,quantity,eval delay,delay(12); game seed unchanged {1,0x29A}.
  - (e) Inputs: target-coordinate result (0,0),target resolution fails. Draws: none. Outputs: return 1,latch=true,mana=2304,quantity=4,delay log=[12],zero missiles; calls action,resolution,lob missile(7,displacement=(0,0),destination=(0,0)),mana,quantity,eval delay,delay(12); game seed unchanged {1,0x29A}.
  - (f) Inputs: resolved target (10,20),equal to the unit position; resolution succeeds. Draws: none. Outputs: return 1,latch=true,mana=2304,quantity=4,delay log=[12],zero missiles; calls action,resolution,lob missile(7,displacement=(0,0),destination=(10,20)),mana,quantity,eval delay,delay(12); game seed unchanged {1,0x29A}.
  - (g) Inputs: forced=1,itemTarget=0,common resolved point (30,40) would be available. Draws: none. Outputs: return 1,latch=true,mana=2304,quantity=4,delay log=[12],zero missiles; calls action,lob missile(7,displacement=(0,0),destination=(0,0)),mana,quantity,eval delay,delay(12); no resolution request; game seed unchanged {1,0x29A}.
  - In every case (a)–(g) the creator observes latch=true at its call.
  - Discriminates: resolving with only one control set, or treating either control as sufficient, fails (g); setting the latch after the creator call fails the observed-latch assertion; applying the both-nonzero formula to a fully zero resolution gives (-10,-20)/(-10,-20) and fails (e); testing the displacement instead of the resolved coordinates gives destination (0,0) and fails (f); failed creator cannot suppress result/charges; exact coordinates/order reject reversed displacement or early charge; a partial zero resolution passes its raw pair with zero destination, so substituting four zeros or using the success result fails (b)/(d), and unbounded coordinate arithmetic fails (c). Unchanged game seed rejects direct dispatcher draws even on supplied creation failure. Extreme coordinates test the supplied i32 arithmetic boundary,not actual coordinate-producer eligibility; that integration and real creator eligibility remain SCD-Q10.
  - runtime_observed: no

- **SCD-V14** Exceptional frame/ignored failure (SCD-R09a, SCD-R09b, SCD-R12; boundary: supplied-input stage)
  - Inputs: latch=true,A=true,B=false,frame predicate=2,code=7,arg=0x01ABCDEF,frame-finished=true; execute the local frame/action rules,logging action-dispatch arguments; srvdofunc=160 (null),srvmissile=-1,periodic=true,perdelay=8,frame=100.
  - Seeds: no draw.
  - Draws: none.
  - Trace: in (a), argument narrows to its low 24 bits 0xABCDEF (11259375); exceptional route bypasses code/latch; actual action dispatch returns 0 from null slot 160/no missile before resource/delay work; ignore that zero and schedule due=105. In (b), absent skill returns 2 before argument/predicate/action work,and dispatcher requests completion. Cases (c)/(d) isolate the fallback decision and direct frame-action empty-skill exit.
  - (a) Inputs: current skill present. Draws: none. Outputs: sequence-event value=0xABCDEF,B=true,latch=true,handler return=2,action-dispatch return=0,mana=2560; calls target refresh(game,unit),frame predicate,action dispatch(20,3,1,0,0),eval perdelay,remove type 8,schedule(8,105,20,3),frame-finished predicate,completion; event (type=8,due=105,id=20,level=3); no action callback,resource request or delay evaluation.
  - (b) Inputs: current skill absent; owned node 20 still exists with A=true,B=false. Draws: none. Outputs: sequence-event value=0,owned node 20 B=false,latch=true,handler return=2,mana=2560; completion request only, no frame predicate/argument write/action/scheduler/finished-predicate request; events empty.
  - (c) Inputs: current skill present,frame predicate=1. Draws: none. Outputs: sequence-event value=0xABCDEF,B=false,latch=true,handler return=2,mana=2560; target refresh(game,unit)→frame predicate→frame-finished predicate→completion only; no action dispatch,callback,resource or scheduler request; events empty.
  - (d) Inputs: enter SCD-R09b directly,current skill absent,owned node 20 still exists with A=true,B=false. Draws: none. Outputs: sequence-event value=0,owned node 20 B=false,latch=true,mana=2560; no action,scheduler,predicate,completion or resource requests; events empty. No numeric result is asserted for this direct frame-action call.
  - (e) Inputs: current skill present,A=false,latch=false,code=7. Draws: none. Outputs: sequence-event value=0xABCDEF,B=false,latch=false,handler return=2,mana=2560; frame-finished predicate→completion only; no target refresh,frame predicate,action dispatch,callback,resource or scheduler request; events empty.
  - (f) Inputs: current skill present,frame predicate=3. Draws: none. Outputs: as (c): sequence-event value=0xABCDEF,B=false,handler return=2; target refresh→frame predicate→frame-finished predicate→completion only.
  - (g) Inputs: current skill present,A=false,latch=false; run once each with code 0,3 and -1. Draws: none. Outputs: each as (e): no action dispatch,handler return=2,completion.
  - Discriminates: an exceptional test of predicate >=2 fails (f); a code gate other than exactly 1 or 2 fails (g); code/latch suppression,high-bit retention,8- or 16-bit narrowing,stop-on-zero fail (a); argument/predicate work before missing-skill gate fails (b); treating predicate 1 as exceptional success fails (c); scheduling after an empty direct frame-action call fails (d); an ordinary gate on the latch alone without the code 1/2 requirement, or evaluating the predicate with A clear, fails (e).
  - runtime_observed: no

- **SCD-V15** Special repeat payload (SCD-R11; boundary: supplied-input stage)
  - Inputs: aura=true,repeat kind=1,perdelay=8,frame=100; events (8,-1,0,200),(8,30,2,201),(9,40,1,300),(9,-1,0,400), tuples (type,id,level,due).
  - Seeds: no draw.
  - Draws: none.
  - Trace: all cases set p=8, wrap frame+7 as i32, divide signed toward zero, multiply by 8 and add 1; remove sentinel type 8 only. The default frame 100 gives n=107,quotient=13,due=105; other case computations are explicit below.
  - (a) Inputs: frame=100. Draws: none. Outputs: p=8,n=107,quotient=13,due=105,return 1; events (8,-1,0,105),(8,30,2,201),(9,40,1,300),(9,-1,0,400); exactly one eval perdelay(unit,id=20,level=3) returning 8,then remove type 8 identity -1,schedule(8,105,-1,0).
  - (b) Inputs: frame=-10. Draws: none. Outputs: p=8,n=-3,quotient=0,due=1,return 1; events (8,-1,0,1),(8,30,2,201),(9,40,1,300),(9,-1,0,400); exactly one eval perdelay(unit,id=20,level=3) returning 8,then remove type 8 identity -1,schedule(8,1,-1,0).
  - (c) Inputs: frame=2147483647. Draws: none. Outputs: p=8,n=-2147483642,quotient=-268435455,due=-2147483639,return 1; events (8,-1,0,-2147483639),(8,30,2,201),(9,40,1,300),(9,-1,0,400); exactly one eval perdelay(unit,id=20,level=3) returning 8,then remove type 8 identity -1,schedule(8,-2147483639,-1,0).
  - (d) Inputs: repeat kind=2,frame=100. Draws: none. Outputs: identical to (a).
  - (e) Inputs: repeat kind=-1,frame=100. Draws: none. Outputs: identical to (a).
  - Discriminates: treating only kind 1 as special fails (d)/(e); all-type-8 purge or ordinary payload fails retained-event/payload assertions; removing identity -1 events of any type fails the retained (9,-1,0,400); floor division fails (b), unbounded arithmetic fails (c). Actual event-storage acceptance of unusual due frames remains SCD-Q10; supplied arithmetic/event requests are asserted here.
  - runtime_observed: no

- **SCD-V16** Direct table boundary (SCD-R14a; boundary: isolated rule)
  - Inputs: (game,unit,20,3),slot 10 return=7; usemanaondo=true,decquant=true,delay=12,srvmissile=7 record present.
  - Seeds: no draw.
  - Draws: none.
  - Trace: bound/slot only.
  - (a) Inputs: index=10. Draws: none. Outputs: slot 10 receives (game,unit,20,3) exactly once; no activation/ownership/missile/resource/event calls; mana=2560,quantity=5,latch=false.
  - (b) Inputs: index=191. Draws: none. Outputs: no calls; mana=2560,quantity=5,latch=false.
  - (c) Inputs: index=-1. Draws: none. Outputs: no calls; mana=2560,quantity=5,latch=false.
  - (d) Inputs: index=160 (in-range null slot). Draws: none. Outputs: no calls; mana=2560,quantity=5,latch=false.
  - Discriminates: ordinary dispatch produces forbidden calls; an upper-bound-only check fails (c); a missing null-slot check fails (d); direct helper return not asserted.
  - runtime_observed: no

- **SCD-V17** Room lookup ordering and start bounds (SCD-R03, SCD-R05a; boundary: supplied-input stage)
  - Inputs: direct start,InTown=false,force=0,activation=true; current skill present,mode=10. Room lookup is a supplied operation with the per-call results listed below; other defaults apply.
  - Seeds: no draw.
  - Draws: none.
  - Trace: the first lookup gates entry before activation. Only after valid identity and successful activation does InTown=false cause a second lookup; its failure clears current skill and requests player reset.
  - (a) Inputs: first room result absent. Draws: none. Outputs: return 0,current skill unchanged,mode=10,mana=2560; exactly one room lookup,no activation,callback,clear or reset request.
  - (b) Inputs: first room result present/non-town; second room result absent. Draws: none. Outputs: return 0,current skill absent,mode=5,mana=2560; ordered calls room lookup,activation,room lookup,clear current skill,reset(mode=5,force=1); no callback/resource request.
  - (c) Inputs: first room result present/non-town; activation=false; second room result would be absent if requested. Draws: none. Outputs: return 0,current skill unchanged,mode=10,mana=2560; room lookup then activation only; no second lookup,callback,clear or reset request.
  - (d) Inputs: InTown=true,first room present/non-town,srvstfunc=-1. Draws: none. Outputs: return 0,current skill unchanged,mode=10,mana=2560; room lookup then activation only; no callback/charge/purge/reset.
  - (e) Inputs: InTown=true,first room present/non-town,srvstfunc=91. Draws: none. Outputs: return 0,current skill unchanged,mode=10,mana=2560; room lookup then activation only; no callback/charge/purge/reset.
  - (f) Inputs: InTown=true,first room present/non-town; second room result would be absent if requested. Draws: none. Outputs: return 1,current skill unchanged,mode=10,mana=2304; exactly one room lookup; ordered calls room lookup,activation,start(20,3),mana(20,3); no clear or reset request.
  - (g) Inputs: force=1,InTown=false,first room present/non-town,second room absent. Draws: none. Outputs: return 0,current skill absent,mode=5,mana=2560; ordered room lookup,room lookup,clear current skill,reset(mode=5,force=1); no activation,callback or resource request.
  - Discriminates: force bypassing anything beyond activation fails (g); caching the first room result reaches the callback in (b); treating first absence as the later gate incorrectly clears/resets in (a); checking the second room before activation resets in (c). Invalid start indices must return 0 rather than null-slot success 1 in (d)/(e). Performing the second lookup with InTown set fails (f). Supplied lookup sequences do not assert that real room results change during one invocation.
  - runtime_observed: no

- **SCD-V18** Sequence delay and other-unit resource gates (SCD-R06, SCD-R08; boundary: supplied-input stage)
  - Inputs: action dispatch,consume=1,srvstfunc=0,usemanaondo=true,decquant=true,delay=12,sequence-event value=1; successful action return=1,no missile; other defaults apply.
  - Seeds: no draw.
  - Draws: none.
  - Trace: players reach activation and mana/quantity requests; the delay expression is evaluated even when Sequence suppresses application. Other unit kinds skip ownership/activation and mana/quantity,then evaluate delay but do not apply it. Monster case (d) reaches the resource requests but is not eligible for player-only delay application; its supplied resource operations log without stored changes.
  - (a) Inputs: player kind 0,anim=18. Draws: none. Outputs: return 1,mana=2304,quantity=4,delay log empty; calls activation,action(20,3),mana(20,3),quantity,eval delay(20,3); no delay application.
  - (b) Inputs: kind=2,anim=18,activation=false,no current skill,no owned skills,bonus-inclusive level operation would return 0 if called,invocation level=3,InTown=false,room absent. Draws: none. Outputs: return 1,mana=2560,quantity=5,delay log empty; action(20,3),eval delay(20,3) only; no ownership/activation/level/reset/mana/quantity/delay-application request.
  - (c) Inputs: player kind 0,anim=10. Draws: none. Outputs: return 1,mana=2304,quantity=4,delay log=[12]; calls activation,action(20,3),mana(20,3),quantity,eval delay(20,3),delay(12).
  - (d) Inputs: monster kind 1,anim=10; supplied mana/quantity operations only log and leave stored resources unchanged. Draws: none. Outputs: return 1,mana=2560,quantity=5,delay log empty; calls activation,action(20,3),mana(20,3),quantity,eval delay(20,3); no delay application.
  - Discriminates: the nonzero event value suppresses only Sequence application,not evaluation or all animations; other kinds bypass player/monster checks and resource requests while evaluating delay. Monster still requests mana/quantity but never the player-only delay; the supplied request log distinguishes that branch from both player and other-unit behavior.
  - runtime_observed: no

### 8.2 Observation vectors

- **SCD-O01** Count actions/latch changes with/without server missile, recording real frame codes/arguments; interrupt before action as control. Expected supplied behavior SCD-V09/SCD-V14; production remains SCD-Q01. [Procedure](../research/spec-evidence/skill-cast-start-and-do-dispatch.md#animation-observation). Pending.
- **SCD-O02** Record frame/period/due for periodic scheduler, varying start frame; compare wrapped signed formula. [Procedure](../research/spec-evidence/skill-cast-start-and-do-dispatch.md#periodic-observation). Pending.

## 9. Acceptance and dependencies

SCD-V01–SCD-V18 define partial acceptance of verified supplied stages: selection, local writes, request arguments/order, returns, due arithmetic and separately supplied seeding. They do not verify actual dependency producers. Full reproduction also needs SCD-R09c/SCD-R10/SCD-R14b/SCD-R15 and exact integrations gated in the Id map. Pending observations prevent runtime-confirmation claims; stage vectors do not establish complete skill parity.

**Depends on:** GDF-R07/GDF-R08a/GDF-R14a for schemas/identities; NRF-R01/NRF-R02 for RNG/numeric conventions, NRF-R14a for mode setting, NRF-R17a for seeding; SCX-R11 for expression front-end bounds/context in [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md), with evaluated results supplied. Per-skill callbacks supply effects/draws/returns. Section 8 supplies other operation effects with actual execution gated by SCD-Q03/SCD-Q06/SCD-Q09/SCD-Q10.

**Depended on by:** [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md) for dispatch/latch; [combat-hit-recovery](combat-hit-recovery.md) for action selectors. References do not verify surrounding gameplay.

## 10. Related specs

- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): arithmetic/modes/seeding.
- [core-game-data-fields](core-game-data-fields.md): schemas/names.
- [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md): expressions.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): actions 44/48/latch.
- [combat-hit-recovery](combat-hit-recovery.md): action-selector exceptions.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): basic attack effect.

## Appendix A. Provenance

Original Lord of Destruction Expansion 1.13c Windows x86, D2Game/D2Common identified as LOD113C-D2GAME/LOD113C-D2COMMON in the [source manifest](../research/source-manifest.md). Compatible 2026-10-12 instruction/table readings support local dispatch; targeted 2026-10-15 instruction reading, original-image decoding and table checks assessed missing paths and request/effect boundaries. [Companion records](../research/spec-evidence/skill-cast-start-and-do-dispatch.md) retain dates/methods/corrections/limits. No gameplay observation or new archive-cell verification is claimed.
