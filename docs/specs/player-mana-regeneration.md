# Player Mana Regeneration

**Description**

Mana normally returns a little at a time while a character is alive. The amount depends on maximum mana and recovery bonuses, so a larger pool does not always take exactly the same time to refill. Effects that suppress normal recovery can still leave other sources of mana recovery active.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [Numeric and RNG foundations](core-numeric-and-rng-foundations.md); names: [Game-data fields](core-game-data-fields.md)
Ids: PMR-R, PMR-Q, PMR-V, PMR-O; [id policy](README.md#conventions).
Evidence: [Public companion](../research/spec-evidence/player-mana-regeneration.md).

## 1. Scope

The player's periodic mana tick: handler requests and eligibility, class-based recovery, percent and flat adjustments, full-mana potion-list cleanup requests, and signed clamping of the requested mana delta. Resource values use signed raw 8.8 units: 256 represents one mana point.

Life/stamina tick internals, initial event creation, complete timer teardown, potion creation, mana spending, leech, mana burn and client synchronization belong to other boundaries. The shared handler's call order is included. Complete cleanup and stat-write effects remain explicit integration gates.

## 2. Trigger and ownership

D2Game owns the server handler and mana arithmetic. A player timer event of type 3 selects the shared resource handler. It requests its successor for the next frame before testing eligibility, then requests life, conditional stamina and mana ticks in that order. Initial posting is PMR-Q01; exact inter-group ordering remains PMR-Q02.

D2Common supplies stat/state reads and list/stat mutation. The mana operation itself sends no client message. Complete stat propagation/notification is PMR-Q04. The companion maps semantic operations to original binaries; no binary layout is required.

## 3. Inputs and data

| Input | Width / units | Source and meaning |
|---|---|---|
| unit, type, class identity, mode | Optional unit; i32 identities | Original type 0 player, 1 monster; player death modes 0/17, monster death modes 0/12 |
| tickSuppressed | Boolean | Engine condition tested before type/mode; wider removal lifecycle PMR-Q09 |
| frame, event arguments | i32 frame; two 32-bit values | Successor copies both arguments and uses wrapping frame+1 |
| current mana | Stat 8 mana, layer 0; i32 8.8 | NRF-R12b/c getter result |
| maximum mana | Stat 9 maxmana, layer 0; i32 8.8 | Guarded getter in PMR-R03 |
| flat recovery | Stat 26 manarecovery, layer 0; i32 raw units/frame | Added after class/percent recovery, even under suppression |
| percent recovery | Stat 27 manarecoverybonus, layer 0; i32 percent | Added to 100 with 32-bit wrap |
| stat storage/definitions | NRF-R12b/c inputs | Existing own/resolved arrays, owner, count and selected ItemStatCost definitions; construction outside boundary |
| normal recovery suppression | State 85 nomanaregen | NRF-R12a presence result |
| potion list | State 106 manapot; attached logical lists | NRF-R12a primary-before-secondary lookup |
| class collection/count | CharStats; signed count | Class identity selects a supplied available record |
| ManaRegen | CharStats ManaRegen; u8 seconds | GDF-R03a prepared field; zero selects engine fallback |
| life/stamina operations | Supplied operations | Log order; life result controls later calls; internals outside boundary |
| schedule, detach, free, add-stat operations | Supplied operations | Request contracts are separate from integrated stored effects |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Stats 8, 9, 26, 27 | data/global/excel/itemstatcost.txt, Stat and explicit ID; GDF-R11 | Fixed original caller identities; definition values loaded from data |
| States 85/106 | data/global/excel/states.txt, state and explicit id; GDF-R12 | Fixed original suppression/potion lookup identities |
| Classes and ManaRegen | data/global/excel/charstats.txt, class/ManaRegen; GDF-R03a/b | Class index selects data; Expansion separator excluded |
| Event 3, death modes, factor 25, fallback 7500 | Engine code | Fixed by original code; custom binding is a separate capability |

Named original-profile values used by vectors: CharStats.ManaRegen (Sorceress)=120 and CharStats.ManaRegen (Amazon)=120. [Table records](../research/spec-evidence/player-mana-regeneration.md#table-evidence) distinguish recorded original-profile support from a current effective-file reread and its limits.

## 4. Rules (execution order)

Consume NRF-R01 notation and NRF-R05 mulDiv. All addition/subtraction/negation keeps signed low32 bits; comparisons are signed and division truncates toward zero. No local RNG draw. Operation names impose no API or storage design.

**PMR-R01a** Successor request [verified in 1.13c code (static)]. On handler entry request event type 3 for the same unit at wrapping frame+1, with no private callback and both original event arguments unchanged. Request it before null, suppression, type or mode tests. Scheduler effects for invalid targets/destruction are PMR-Q08/PMR-Q09.

**PMR-R01b** Scheduling integration [unresolved]. Valid-target insertion/exact-frame dispatch consume STL-R01/STL-R02. Initial posting, whole-chain lifetime and inter-group ordering remain PMR-Q01/PMR-Q02/PMR-Q09. No physical queue layout is required.

**PMR-R02a** Dead-player exit [runtime-confirmed (1.13c)]. After the successor request, a non-null not suppressed player in mode 17 exits before life, stamina or mana calls. The matching record covers that exit only, not successor arguments, cadence or arithmetic.

**PMR-R02b** Eligibility and ordered requests [verified in 1.13c code (static)].

```
after successor request:
    if unit absent or unit.tickSuppressed: return
    if unit.type == 0:
        if unit.mode == 0 or unit.mode == 17: return
    else if unit.type == 1:
        if unit.mode == 0 or unit.mode == 12: return
    else: return
    if LifeTick(unit) == 0: return
    StaminaTick(unit)
    ManaTick(unit)
```

This owns request order. Supplied life/stamina effects do not verify those producers. Natural monster entry to this player handler is PMR-Q05.

**PMR-R03** Read mana inputs [verified in 1.13c code (static)]. Null ManaTick unit returns without a read. Otherwise read current mana, guarded maximum mana, then state 85. Stat reads consume NRF-R12b/c, including conditional minimum replacement. The maximum getter returns 0 for null unit/list, definition count<=9 or absent selected definition; otherwise it reads stat 9 with the same reader. No independent maximum clamp occurs here.

**PMR-R04** Class recovery [verified in 1.13c code (static)]. State 85 active gives base=0 and skips class selection and percent read. Otherwise:

```
if classIdentity < 0 or classIdentity >= classCount: return from ManaTick
record = selected class record
if record absent: return from ManaTick
divisor = unsignedByte(record.ManaRegen) * 25
if divisor == 0: divisor = 7500
base = maximumMana / divisor
if base < 1: base = 1
percent = Stat(unit, 27, 0)
base = mulDiv(i32(100 + percent), base, 100)
```

Operand order is percent factor first/class base second. Early return skips flat recovery too. This logical record contract requires valid supplied collections; arbitrary invalid backing-storage arithmetic is not defined here.

**PMR-R05** Flat recovery [verified in 1.13c code (static)]. Read stat 26 and set regen=i32(base+flatRecovery). This still applies under state 85; negative results can request a drain.

**PMR-R06a** Full-mana cleanup requests [verified in 1.13c code (static)]. When maximumMana<=currentMana and regen>0, lookup the state 106 list through NRF-R12a. If L exists request Detach(unit, L), then Free(L). Continue clamping with the current/maximum values captured in PMR-R03. Otherwise skip lookup or cleanup as appropriate.

**PMR-R06b** Cleanup effects [unresolved]. Actual contribution removal, callbacks, list lifetime and state 106 membership require the owning lifecycle effects. STL-R05 and PMR-Q03 gate these, independently of verified request condition/order.

**PMR-R07a** Clamp and additive request [verified in 1.13c code (static)].

```
room = i32(maximumMana - currentMana)
if regen > room: regen = room
lower = i32(-currentMana)
if regen < lower: regen = lower
if regen == 0: return
AddStat(unit, 8, regen, 0)
```

Upper test precedes lower test. Negative delta requests subtraction. This is not a saturating clamp over arbitrary i32 inputs: subtraction and negation can wrap.

**PMR-R07b** Stored write and notification [unresolved]. Full stat propagation, indirect callbacks and client synchronization require the complete mutation operation (PMR-Q04). Supplied additive fixtures do not verify integrated effects.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| PMR-R01a | Successor before gates | verified in 1.13c code (static) | PMR-Q08/Q09: scheduler effects only | PMR-V10, PMR-V12, PMR-V16 |
| PMR-R01b | Scheduling integration | unresolved | PMR-Q01/Q02/Q09 | none |
| PMR-R02a | Observed mode 17 exit | runtime-confirmed (1.13c) | none for the exit | PMR-V11 |
| PMR-R02b | Gates and ordered calls | verified in 1.13c code (static) | PMR-Q05: natural monster entry only | PMR-V10, PMR-V12, PMR-V16 |
| PMR-R03 | Getter selection | verified in 1.13c code (static) | none with supplied inputs | PMR-V01, PMR-V05, PMR-V13, PMR-V16 |
| PMR-R04 | Class recovery and percent | verified in 1.13c code (static) | none with supplied inputs | PMR-V01..PMR-V04, PMR-V08, PMR-V09, PMR-V14 |
| PMR-R05 | Flat add | verified in 1.13c code (static) | none with supplied inputs | PMR-V05, PMR-V14 |
| PMR-R06a | Cleanup request condition/order | verified in 1.13c code (static) | none with supplied lookup | PMR-V06, PMR-V07, PMR-V15 |
| PMR-R06b | Cleanup stored effects | unresolved | PMR-Q03 | PMR-V06 gated effects |
| PMR-R07a | Clamp and additive request | verified in 1.13c code (static) | none with supplied inputs | PMR-V01..PMR-V08, PMR-V14 |
| PMR-R07b | Integrated mutation | unresolved | PMR-Q04 | supplied model only |

## 5. Outputs and side effects

| Output / request | Rule |
|---|---|
| Type 3 successor request, frame+1, no private callback, unchanged arguments | PMR-R01a |
| Life, conditional stamina and mana calls in that order | PMR-R02b |
| State 106 lookup; detach then free requests | PMR-R06a |
| Gated list/contribution/callback/state effects | PMR-R06b |
| Stat 8, layer 0 addition request, nonzero signed delta | PMR-R07a |
| Gated stored totals and notification | PMR-R07b |
| No message directly sent / no local RNG change | PMR-R03..PMR-R07a |

## 6. Branches and edge cases

Suppression skips class and percent but allows flat recovery. Invalid class exits before flat recovery. ManaRegen 0 uses 7500; the class quotient floors to 1 before percent scaling. Missing stat storage reads 0 through dependencies, so with valid class the floor 1 is clamped back to 0 and makes no write request.

For ordinary nonnegative current/maximum values, the signed ordered clamp prevents a negative final balance and can reduce an above-maximum balance in one tick. Arbitrary i32 inputs preserve wrap semantics. Full mana can request potion cleanup before regen clamps to 0; pre-clamp regen<=0 skips cleanup. Difficulty, expansion, PvP and attack type are not read here.

## 7. Open questions

| Id | Kind | Blocked assertion / evidence that would settle it |
|---|---|---|
| PMR-Q01 | integration | Initial player type 3 event: complete creation/join path to scheduler |
| PMR-Q02 | integration | Whole recurring chain and group ordering: complete player dispatch/lifecycle and STL-R02a |
| PMR-Q03 | integration | Actual potion cleanup/state 106 effects: complete lifecycle detach/free and potion callback |
| PMR-Q04 | integration | Stored mutation/propagation and notification: complete add/propagation/synchronization paths |
| PMR-Q05 | outside-boundary | Natural monster entry: identify a producer for the accepted type case |
| PMR-Q06 | observation | Recovery, exact scheduling arguments and cadence: attributable living-player/scheduler captures |
| PMR-Q08 | integration | Null-target scheduling: complete guarded insertion contract for that input |
| PMR-Q09 | outside-boundary | Removed unit's chain cancellation: teardown and event removal paths |

[Open evidence gaps](../plans/gaps/player-mana-regeneration.md) records attempts. Pending observation does not gate verified supplied-input requests.

## 8. Test vectors

### 8.1 Executable vectors

Each case starts independently. Defaults: non-null player type 0, class 1 Sorceress, mode 1, tickSuppressed=false; class count 7, selected ManaRegen 120; stat definition count 359/list present. Getter results are the explicitly supplied mana/maximum/flat/percent; no minimum adjustment changes them. State 85 off, state 106 lookup absent, flat/percent 0. ManaTick is entered directly, except explicitly named handler cases. All use no draw, ordered draws none, RNG unchanged.

Dependencies log ordered requests. A synthetic AddStat also adds the requested delta to a supplied balance with i32 wrap; those balances preserve the examples but original integrated effects remain PMR-R07b. For V06, synthetic Detach removes L from its supplied chain and synthetic Free releases L; both log calls. Actual original removal/free/state effects remain PMR-R06b. Captured clamp inputs do not change during cleanup.

- **PMR-V01** Ordinary recovery (PMR-R03/R04/R05/R07a; boundary: supplied-input stage)
  - Inputs: maximum 10240, current 4096.
  - Seeds: no draw
  - Draws: none
  - Trace: divisor 3000; base 3; mulDiv(100, 3, 100)=3; regen 3; room 6144; lower -4096.
  - Outputs: AddStat(8, 3, 0); model balance 4099; no lookup/cleanup.
  - (a) Inputs: current mana starts at 0; maximum 10240 and all other base inputs remain constant. Invoke ManaTick directly 3414 times, each call reading the preceding synthetic balance as current mana; no scheduling or gameplay cadence executes. Draws: none. Trace: 3413*3=10239, then room 1 clamps the final request to 1. Outputs: exactly 3413 successive AddStat(8,3,0) requests followed by one AddStat(8,1,0); synthetic balance after 3000 calls is 9000 and after 3414 calls is 10240. No lookup/cleanup occurs. At a separately supplied cadence of 25 calls/second, 3414 calls represent 136.56 seconds.
  - Discriminates: Per-tick truncation gives 3 raw units until the final clamp. Retaining fractional recovery to refill in exactly 120 seconds would already reach 10240 at call 3000 and fails (a)'s request sequence and balances.
  - runtime_observed: no

- **PMR-V02** Base floor (PMR-R04/R07a; boundary: supplied-input stage)
  - Inputs: maximum 2560, current 1280.
  - Seeds: no draw
  - Draws: none
  - Trace: 2560/3000 truncates 0; floor 1; regen 1.
  - Outputs: AddStat(8, 1, 0); model balance 1281. From 0, full after 2560 ticks=102.4 seconds at 25 fps.
  - Discriminates: Without floor there is no recovery.
  - runtime_observed: no

- **PMR-V03** Percent bonus (PMR-R04; boundary: supplied-input stage)
  - Inputs: class 0 Amazon, ManaRegen 120, maximum 25600, current 0, percent 50.
  - Seeds: no draw
  - Draws: none
  - Trace: base 8; mulDiv(150, 8, 100)=12.
  - Outputs: AddStat(8, 12, 0); model balance 12.
  - Discriminates: Omitting percent gives 8.
  - runtime_observed: no

- **PMR-V04** Negative percent and lower clamp (PMR-R04/R07a; boundary: supplied-input stage)
  - Inputs: class 0, ManaRegen 120, maximum 25600, percent -130; independent current 256, 1, 0.
  - Seeds: no draw
  - Draws: none
  - Trace: base 8; mulDiv(-30, 8, 100)=-2; apply lower clamp.
  - Outputs: Respectively AddStat(8, -2, 0)/balance 254; AddStat(8, -1, 0)/balance 0; no request/balance 0.
  - Discriminates: Flooring instead of truncation gives -3 in first case; omitting lower clamp underflows second case.
  - runtime_observed: no

- **PMR-V05** Suppression preserves flat recovery (PMR-R03/R04/R05; boundary: supplied-input stage)
  - Inputs: state 85 on, maximum 25600, percent 999, selected class record unavailable. Independent(a) current 1000/flat 5; (b) current 1000/flat 0; (c) current 2/flat -3.
  - Seeds: no draw
  - Draws: none
  - Trace: No class/percent read; base 0; flat supplies regen; clamp.
  - Outputs: (a) AddStat(8, 5, 0)/balance 1005; (b) no request/balance 1000; (c) AddStat(8, -2, 0)/balance 0. No class/percent read in any case.
  - Discriminates: Suppressing flat fails(a)/(c); unnecessary class access fails all cases.
  - runtime_observed: no

- **PMR-V06** Cleanup before zero clamp (PMR-R06a/R06b/R07a; boundary: supplied-input stage)
  - Inputs: maximum=current 25600; base list has state storage; lookup finds L, state 106, nonextended, no callback; supplied chain [L].
  - Seeds: no draw
  - Draws: none
  - Trace: regen 8; full condition true; lookup L; Detach then Free; room 0 clamps regen to 0.
  - Outputs: Ordered [FindStateList(106), Detach(unit, L), Free(L)]; no AddStat; model balance 25600. Gated original effects: L no longer attached/released; state 106 membership remains PMR-Q03.
  - Discriminates: Clamp-before-cleanup omits calls; Free-before-Detach reverses log.
  - runtime_observed: no

- **PMR-V07** Above maximum (PMR-R06a/R07a; boundary: supplied-input stage)
  - Inputs: maximum 25600, current 28672, lookup returns absent.
  - Seeds: no draw
  - Draws: none
  - Trace: regen 8; lookup; room -3072; lower -28672; request -3072.
  - Outputs: FindStateList(106), then AddStat(8, -3072, 0); no detach/free; model balance 25600.
  - Discriminates: Skipping recovery when already full leaves excess unchanged.
  - runtime_observed: no

- **PMR-V08** Zero class constant (PMR-R04; boundary: supplied-input stage)
  - Inputs: supplied ManaRegen 0, maximum 25600, current 0.
  - Seeds: no draw
  - Draws: none
  - Trace: divisor 0 becomes 7500; base 3.
  - Outputs: AddStat(8, 3, 0); model balance 3.
  - Discriminates: Zero-as-no-recovery omits request.
  - runtime_observed: no

- **PMR-V09** Class range exit (PMR-R04; boundary: supplied-input stage)
  - Inputs: count 7, maximum 25600, current 0, flat 50, state 85 off. Independent cases: (a) class 7; (b) class -1; (c) class 0 with selected record absent.
  - Seeds: no draw
  - Draws: none
  - Trace: Cases (a)/(b) fail the class-range test; (c) passes the range test but finds no selected record. All return before flat recovery.
  - Outputs: No flat read/lookup/cleanup/AddStat; model balance 0.
  - Discriminates: Applying flat despite either exit requests +50; omitting the absent-record guard attempts a field read in (c).
  - runtime_observed: no

- **PMR-V10** Dead player successor request (PMR-R01a/R02b; boundary: supplied-input stage)
  - Inputs: Handler entry, player mode 17, tickSuppressed=false, frame 1000, args 11/22, resources life 0/mana 2560/stamina 100; scheduler logs only.
  - Seeds: no draw
  - Draws: none
  - Trace: Schedule first, then dead exit.
  - Outputs: One schedule(same unit, type 3, frame 1001, no private callback, args 11/22); no life/stamina/mana calls; resources unchanged. Event insertion not asserted.
  - Discriminates: Gate-before-schedule omits request.
  - runtime_observed: no

- **PMR-V11** Observed dead exit (PMR-R02a; boundary: isolated rule)
  - Inputs: Boundary starts after successor request; non-null not suppressed player, class 4, mode 17.
  - Seeds: no draw
  - Draws: none
  - Trace: Type 0/mode 17 selects dead exit.
  - Outputs: Return at dead-mode exit before any resource call. No scheduler arguments, event, balance or cadence assertion.
  - Discriminates: Treating mode 17 as eligible reaches resource calls.
  - runtime_observed: yes (1.13c, 2026-10-12, [narrow dead-exit record](../research/spec-evidence/player-mana-regeneration.md))

- **PMR-V12** Handler gates and order (PMR-R01a/R02b; boundary: supplied-input stage)
  - Inputs: Handler entry, frame 1000, args 11/22; scheduler/ticks log only. Cases(a) absent unit; (b) player suppressed/mode 1; (c) player mode 0; (d) monster mode 12; (e) object mode 1; (f) player mode 1 LifeTick returns 0; (g) player mode 1 LifeTick returns 1; (h) monster mode 0. Other cases not suppressed.
  - Seeds: no draw
  - Draws: none
  - Trace: Successor first; then eligibility, then life-result test.
  - Outputs: Every case first schedules(type 3, frame 1001, no private callback, args 11/22) for supplied target, including absent target(a). (a)..(e) and (h) no ticks; (f) LifeTick only; (g) LifeTick, StaminaTick, ManaTick in exact order. Stored scheduler effects excluded.
  - Discriminates: (a)/(b) reject early gate; (f) reject ignored return; (g) reject swapped calls; (h) rejects treating monster mode 0 as eligible.
  - runtime_observed: no

- **PMR-V13** Getter guards (PMR-R03; boundary: supplied-input stage)
  - Inputs: Cases(a) absent ManaTick unit; (b) unit without a stat list, valid class; (c) unit/list present, definition count 9, no definition 9, current getter 100, state 85 on, flat getter 0.
  - Seeds: no draw
  - Draws: none
  - Trace: (a) return; (b) current/max 0, base floor 1, room clamp 0; (c) maxgetter 0, regen 0, room -100, lower -100.
  - Outputs: (a) no reads/requests; (b) no write request; (c) AddStat(8, -100, 0), model balance 0, no cleanup.
  - Discriminates: Reading missing definition 9 fails(c); early return merely because regen 0 incorrectly omits(c) clamp.
  - runtime_observed: no

- **PMR-V14** Operand order and wrap (PMR-R04/R05/R07a; boundary: supplied-input stage)
  - Inputs: Cases(a) ManaRegen 1, maximum 25000000, current 0, percent 65437, flat 0; (b) maximum 25600, current 100, percent 0, flat 2147483647.
  - Seeds: no draw
  - Draws: none
  - Trace: (a) base 1000000; factor 65537; mulDiv(65537, 1000000, 100)=(65537/100)*1000000=655000000; upper clamp 25000000. (b)8+2147483647 wraps-2147483641, lowerclamp-100.
  - Outputs: (a) scaled intermediate 655000000, AddStat(8, 25000000, 0), balance 25000000; (b) AddStat(8, -100, 0), balance 0.
  - (c) Inputs: maximum 25600, current 0, percent 2147483647, flat 0, ManaRegen 120. Draws: none. Trace: class base 8; i32(100+2147483647)=-2147483549; NRF-R05 takes its final branch, low32(-2147483549*8)=792, then 792/100=7; room 25600/lower 0 leave 7 unchanged. Outputs: scaled intermediate 7, AddStat(8, 7, 0), model balance 7; no cleanup.
  - (d) Inputs: maximum 1, current -2147483648, percent 0, flat 0, ManaRegen 120. Draws: none. Trace: class quotient 0 floors to 1; regen 1; room=i32(1-(-2147483648))=-2147483647, so upper test sets regen=-2147483647; lower=i32(-(-2147483648))=-2147483648, so lower test leaves regen unchanged. Outputs: AddStat(8, -2147483647, 0), model balance i32(-2147483648-2147483647)=1; no cleanup.
  - (e) Inputs: maximum -100, current 100, state 85 on, flat 0; supplied getter values are not replaced by minimum adjustment. Draws: none. Trace: suppression sets base 0, flat leaves regen 0, so cleanup is skipped. Upper room=-100-100=-200 sets regen=-200; lower=-100 then sets regen=-100. Outputs: AddStat(8,-100,0), synthetic balance 0; no class/percent read or cleanup.
  - Discriminates: Swapping mulDiv arguments gives 655370000 intermediate in(a), although final clamp hides the difference. Saturating add in(b) gives a positive request. A nonwrapping factor or product fails(c); a mathematical clamp without i32 subtraction/negation wrap fails(d). Reversing the two clamp tests in(e) requests -200 and leaves balance -100, failing the exact request and balance.
  - runtime_observed: no

- **PMR-V15** Cleanup guards and captured values (PMR-R06a/R07a; boundary: supplied-input stage)
  - Inputs: Cases(a) current 25599, maximum 25600, listpresent; (b) current=maximum 25600, state 85 on, flat 0, listpresent; (c) current=maximum 25600, normal recovery, lookup L, cleanup changes subsequent live getter maximum to 0 but not captured values.
  - Seeds: no draw
  - Draws: none
  - Trace: (a) not full; (b) not positive; (c) cleanup then clamp uses captured25600-25600.
  - Outputs: (a) AddStat(8, 1, 0), no lookup/cleanup; (b) no lookup/cleanup/write; (c) lookup, Detach, Free only, no AddStat, captured current/maximum 25600. No actual cleanup effects asserted.
  - Discriminates: Cleanup near full fails(a); cleanup nonpositive regen fails(b); rereading maximum for clamp fails(c).
  - runtime_observed: no

- **PMR-V16** Request and read boundaries (PMR-R01a/R02b/R03; boundary: supplied-input stage)
  - Inputs: Independent cases. (a) Handler entry, frame 2147483647, player mode 17, suppression false, arguments 11/22, scheduler logs only. (b) Handler entry, frame 1000, monster mode 1, suppression false, arguments 11/22; LifeTick returns 1 and all ticks log only. (c) Direct ManaTick entry, current and maximum getter results 10/10, state 85 on, flat getter 0, selected class record unavailable. In (c) only, getter operations log their top-level reads as well as returning supplied values; internal getter implementation is outside that read log.
  - Seeds: no draw
  - Draws: none
  - Trace: (a) i32(2147483647+1)=-2147483648, then dead gate. (b) Type 1/mode 1 passes; life result 1 permits later ticks. (c) Read current, maximum, state; suppression skips class/percent; read flat; regen and room are 0.
  - Outputs: (a) One schedule request for event type 3 at frame -2147483648, no private callback, same target, arguments 11/22; no ticks. Stored negative-frame scheduling is not asserted. (b) Exact log: schedule(type 3,frame 1001,no private callback,args 11/22),LifeTick,StaminaTick,ManaTick. Natural monster entry is not asserted. (c) Exact read log: Stat(8,0),MaxMana(9,0),HasState(85),Stat(26,0); no class/percent read, cleanup or AddStat; balance 10.
  - Discriminates: (a) Rejects saturating frame addition. (b) Rejects dropping the handler's eligible monster branch. (c) Rejects swapping the three initial reads or reading percent under suppression.
  - runtime_observed: no

### 8.2 Observation vectors

- **PMR-O01** Measure successor arguments, per-unit cadence and dead/alive entries. The narrow recorded exit does not complete these assertions. [Procedure](../research/spec-evidence/player-mana-regeneration.md#observation-procedures).
- **PMR-O02** Measure living-player recovery from actual getter inputs, percent/flat changes and clamps. Pending. [Procedure](../research/spec-evidence/player-mana-regeneration.md#observation-procedures).
- **PMR-O03** Observe full-mana potion cleanup, state membership and contributions. Pending. [Procedure](../research/spec-evidence/player-mana-regeneration.md#observation-procedures).

## 9. Acceptance and dependencies

Verified supplied-input acceptance: PMR-R01a, PMR-R02a/b, PMR-R03, PMR-R04, PMR-R05, PMR-R06a, PMR-R07a and their executable assertions in PMR-V01..PMR-V16. Synthetic balances/cleanup effects do not establish integrated parity. PMR-R01b, PMR-R06b and PMR-R07b require explicit gates resolved for full reproduction. PMR-O01..PMR-O03 define remaining observation scope.

Consumed contracts: NRF-R01/R05/R08a/R09a arithmetic, units and frames; NRF-R12a/b/c state/list/stat reads; GDF-R03a/b, GDF-R11, GDF-R12, GDF-R14/GDF-R15 identities/fields; STL-R01/STL-R02 timer operations and STL-R05 lifecycle effects with their gates. Life/stamina internals are supplied at the handler boundary.

Reverse consumers: regeneration, consumables and mana-spending behaviors can consume the tick. Shared caller relationships do not alone impose plan prerequisites.

## 10. Related specs

- [Numeric and RNG foundations](core-numeric-and-rng-foundations.md), [Game-data fields](core-game-data-fields.md), [Timed-state lifecycle](state-timed-state-lifecycle.md): consumed contracts.
- [Running stamina drain](player-stamina-drain-while-running.md), [Leech and drain](combat-leech-and-status-effects.md), [Status application](combat-status-application.md): related resource behaviors.
- [Coverage](coverage.md): frame/event queues, life/stamina regeneration, consumables and mana-spending candidates.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. [Source manifest](../research/source-manifest.md) identifies original D2Game.dll/D2Common.dll analysis specimens by SHA-256. [Public companion](../research/spec-evidence/player-mana-regeneration.md) maps rules to module-relative functions, resolved imports, dated readings, table evidence and observation procedures.

Static evidence dates are 2026-10-12 and targeted source/dependency/table assessment 2026-10-14. The recorded runtime record confirms only the stated dead-player exit. Recovery, cadence and complete mutation/synchronization remain unconfirmed; the spec remains draft. Public provenance does not require a private reconstruction project.
