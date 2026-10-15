# Movement modifiers and freeze

**Description**

Movement and action speed change when a unit gains speed bonuses or is chilled. Chill slows
movement, attacks and other actions, with different minimum speeds for each; freeze can instead
hold an eligible monster until its effect ends. Repeated chill or freeze can extend the effect,
and some monsters receive chill instead of freeze.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MMF-R, MMF-Q, MMF-V, MMF-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/state-movement-modifiers-and-freeze.md)
Open evidence gaps: [gap file](../plans/gaps/state-movement-modifiers-and-freeze.md)

## 1. Scope

This contract covers movement percentage, velocity, animation rate, attack and other-action rates,
and the movement-facing cold/freeze duration, immunity, refresh, expiry and AI requests. The
cold/freeze application operations consume [combat-status-application](combat-status-application.md)
(CSA); resistance consumes [combat-damage-mitigation](combat-damage-mitigation.md) (DMI), and their
combat placement consumes [combat-damage-pipeline-order](combat-damage-pipeline-order.md) (DPO).
Lengths produced before that stage, death/corpse consequences of shatter, full path creation,
network delivery and actual list-stat propagation are outside the local arithmetic/request
boundary. They remain explicit dependencies or questions. Get-hit, block and knockback rate
calculations belong to [combat-hit-recovery](combat-hit-recovery.md) (HRC).

## 2. Trigger and ownership

D2Common computes the rate dispatcher synchronously when animation preparation or an explicit
animation refresh requests it. A cold application requests refresh after its new list's three
stat-write requests; cold expiry requests it after state requests (CSA-R05, CSA-R13). Complete
animation preparation and effective type/class/mode production remain MMF-Q11. This is not an
assertion that every unit is refreshed every frame.

D2Game's applied-hit status stage requests cold, then freeze, once per hit, after the current-life
write and before the later combat reactions (CSA-R01; DPO-R17, DPO-R26). Freeze can call cold again.
Expiry callbacks are requested through timed-list removal; callback invocation after contribution
removal and actual stored propagation remain STL-R05/STL-R06b/STL-R06c. The callback contracts below
also apply at supplied entry independently. The game frame and difficulty for application come
from the attacker's game; freeze expiry uses the affected unit's game. No direct message is
established here; state-message selection/delivery remains STL-R14. Original dispatch locations
and callback chains are in the [companion](../research/spec-evidence/state-movement-modifiers-and-freeze.md#dispatch-and-function-references).

## 3. Inputs and data

Notation is NRF section 1: wrapping i32 arithmetic unless stated, signed comparisons and signed
division truncating toward zero. `low32` retains 32 bits; `i32`/`u32` interpret those bits.
`Stat(unit,id,0)` consumes NRF-R12b/NRF-R12c's selected-array read and minimum adjustment. It does
not sum arbitrary lists. `D(index)` consumes NRF-R07, including its fault behavior. Frames are i32
absolute game ticks; supplied difficulty is 0..2. A path speed is i32 8.8; step components are i32
16.16 (NRF-R08a, NRF-R11a).

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| effective unit kind, class, mode | prepared dispatcher inputs; actual mode separately retained | i32 | Player 0, monster 1; preparation can remap these (MMF-Q11) |
| movement/action stats | 67 `velocitypercent`, 68 `attackrate`, 69 `other_animrate` | i32, layer 0 | Supplied unit read results, including any already-resolved contributions |
| item speed stats | 96 `item_fastermovevelocity`, 93 `item_fasterattackrate`, 99 `item_fastergethitrate`, 102 `item_fasterblockrate`, 105 `item_fastercastrate` | i32, layer 0 | Diminished FRW, IAS, FHR, FBR, FCR |
| length defenses | 43 `coldresist`, 44 `maxcoldresist`, attacker 335 `passive_cold_pierce`, 153 `item_cannotbefrozen`, 118 `item_halffreezeduration` | i32, layer 0 | DMI-R07/DPO-R10 consume them; all nonzero values trigger the two item defenses |
| player movement base | CharStats.WalkVelocity; RunVelocity belongs to upstream producers | u8 tile-speed | Walk base used for all selected movement modes |
| monster movement base | MonStats.Velocity | signed i16 tile-speed | Shifted by 8 before percentage multiplication |
| monster animation bases | prepared ordinary and run animation values, not text columns | signed i16 | Production GDF-R05b/GDF-Q01; run selects second |
| action base B | selected animation record's rate | i32 | Actual selection MMF-Q11; supplied base is independent |
| classification results | cast, block, get-hit, knockback, movement, attack, base-only predicates | boolean, in dispatch order | Supplied effective predicates; actual skill/table selection MMF-Q11 |
| attack preparation | attack-preparation operation; optional base-override predicate/result | supplied operation; boolean/i32 | Preparation and override production MMF-Q08; local consumer uses a nonzero override |
| paired-item inputs | dual-wield eligibility, two accepted/type-qualified items, each Stat(item,68,0) | boolean flags; i32 values S1/S2 | Caller arithmetic is separate from item/slot/type production (MMF-Q08) |
| path | present/absent, speed, targetSpeed, resetMarker, scale, acceleration, ramp tick, direction components | i32 logical fields | NRF-R08b/NRF-R11a/NRF-R11b; axis direction 4096, equal diagonal 2896 |
| rate outputs at entry | animationRate; actionRate | i16 nonnegative result; i32 companion result | Cast/attack write both; other branches write animationRate |
| application | attacker, defender, attacker game.frame/difficulty, coldLength, freezeLength | optional associations; i32 lengths/frames | Lengths are independent, before strip/mitigation or explicitly after those stages |
| monster cold fields | MonStats.coldeffect, coldeffect(N), coldeffect(H) | signed i8 | Difficulty column selected by the consumer; invalid/unavailable record gives -50 |
| freeze checks | HasState(defender,54), MonStats.boss, monsterFlag8, hireling-class category | boolean/nonzero | CSA-R02/CSA-R07; monsterFlag8 has no wider classification asserted |
| shatter suppression | MonStats.MonStatsEx selects MonStats2.deadCol | signed i16 linked identity; boolean | Invalid/unavailable lookups supply false (CSA-R06) |
| duration divisors | DifficultyLevels.MonsterColdDivisor / MonsterFreezeDivisor | i32 | Cold zero skips division; freeze zero is fatal |
| expiry context | list presence, expiry, own list stats, active states, event-chain nodes with kind/successor | logical values | Supplied lookup/storage operations described below |
| callback context | unit type/mode, aliveExclusionCondition, StayDeath result, unit game, monster data and record | boolean, i32, associations | Alive and stay-death consume CSA-R13/STL-R11/STL-R12a |
| freeze AI delay | MonStats.aidel / aidel(N) / aidel(H), game.modeConditionA/B | u8 values; two neutral booleans | Both conditions false selects Normal; otherwise difficulty column; zero delay becomes 15 |
| seed | defender seed {lo,hi} | two u32 | Cold side draw consumes NRF-R01/NRF-R02 |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Stat names/ids above | ItemStatCost named rows, ID column; GDF-R11a | Names from table; code fixes the selected numeric stat ids |
| States 1 freeze, 11 cold, 54 uninterruptable, 101 holyshield, 107 shatter | States named rows/id; GDF-R12 | Code-fixed selectors; no custom remapping established |
| Player modes 2 Walk, 3 Run, 6 Town Walk, 18 Sequence; monster mode 15 run; neutral mode 1 | PlrMode/MonMode positional ids; GDF-R14a/GDF-R15a | Code-fixed comparisons; supplied effective mode after preparation |
| Diminishing indices 0 IAS, 1 FHR, 2 FCR, 3 FBR, 4 FRW | Engine table; NRF-R07 | Fixed handlers: A=120 for indices 0..3, A=150 for index 4 |
| Difficulty cold and freeze divisors | Effective `data/global/excel/difficultylevels.txt`, rows Name Normal/Nightmare/Hell, respective columns: 1/1, 2/2, 4/4 | Loaded named cells, [GDF records](../research/spec-evidence/core-game-data-fields.md#named-cell-records); consumer policy fixed in code |
| WalkVelocity/Velocity, cold effects, AI delay, boss/deadCol | CharStats/MonStats/MonStats2 named fields; GDF-R03a/GDF-R05a/GDF-R06a | Loaded or derived fields; per-class vector values are supplied fixtures, not claims of named archive cells |
| 25 movement floor; 15..175 action clamp; 32767 rate cap; 213/101 player movement bases; 75/100/100 monster initialization requests | Engine code | Fixed by code; initialization requests do not prove stored totals |
| Timer kinds 12 expiry / 2 AI; draw modulus 100 and shatter threshold 20 | Engine code | Fixed selectors/constants; scheduling consumes STL-R01 |

Dependency operations are semantic contracts: FindStateList/HasState/SetState (NRF-R12a,
STL-R07), list allocation/state/callback/expiry operations (STL-R06a), attachment (STL-R06b,
gated), replacing stat writes (STL-R06c, gated), scheduling (STL-R01), timer-node removal and
refresh requests (CSA-Q02), cold expiry (CSA-R13), StayDeath (STL-R12a), and mode writes
(NRF-R14a). The names impose no API or representation.

## 4. Rules (execution order)

**MMF-R01** Diminishing selection [verified in 1.13c code (static)]. Consume NRF-R07's index
mapping from section 3. Its ordinary nonzero input computes trunc(i32(low32(A*v))/i32(low32(A+v)));
null list/invalid definition gives 0. It draws nothing; v=-A faults. Null unit is not the null-list case.

**MMF-R02** Movement percentage [verified in 1.13c code (static)]. On a selected movement
branch with a path, pct=max(i32(low32(Stat(unit,67,0)+D(4))),25). No upper clamp; same pct for
walk/run/town-walk. Absent path returns before the stat reads and both rate/speed writes. No draw.

**MMF-R03** Movement velocity [verified in 1.13c code (static)]. Select player WalkVelocity or
monster signed Velocity and shift left 8 with low32. With supplied valid class record, compute
velocity=trunc(i32(low32(base*pct))/100), then consume NRF-R08b: set path speed and targetSpeed;
a differing speed first sets resetMarker=15. No clamp, no draw. Invalid class record is an
unguarded read/fault, not a zero-speed fallback. Full run baseline production remains MMF-Q05.

**MMF-R04a** Movement animation [verified in 1.13c code (static)]. Player base is 101 for
effective mode 3, otherwise 213. Monster base is the signed prepared run value for effective
mode 15, otherwise the ordinary value. rate=clamp(trunc(i32(low32(base*pct))/100),0,32767).
Write animationRate only, before MMF-R03's speed request. No draw.

**MMF-R04b** Monster animation-base production [unresolved]. The two prepared signed values
come from initialization; GDF-R05b/GDF-Q01 gates their production. Supplied bases support MMF-R04a.

**MMF-R05a** Attack rate consumer [verified in 1.13c code (static)]. After the attack-preparation
request, obtain v=low32(Stat(unit,68,0)+D(0)); apply MMF-R05b. If the actual unit is a player in
actual mode 18, subtract 30 with wrap. Clamp signed v to 15..175. Use supplied override result
only when the override predicate is true and that result is nonzero; otherwise use B. Form
p=low32(base*v), q=floor(u32(p)/100) (NRF-R06b), rate=min(q,32767). Write animationRate and
actionRate=rate. No local draw; preparation/override effects and their possible draws remain
MMF-Q08, separate from the consumer.

**MMF-R05b** Paired-item rate adjustment [verified in 1.13c code (static)]. When supplied
dual-wield eligibility and both accepted/type-qualified item tests are true, replace v with
low32(v+trunc(i32(low32(S1+S2))/2)-S1). Otherwise leave v. The eligibility helper admits player
classes 4/6 and valid monster classes 417/418, otherwise false, including null. Inventory slot,
type selection and item-test integration remain MMF-Q08; arithmetic with supplied S1/S2 draws nothing.

**MMF-R06a** Default other-action rate [verified in 1.13c code (static)]. If every preceding
predicate is false, v=clamp(Stat(unit,69,0),15,175); p=low32(B*v); rate=min(floor(u32(p)/100),32767).
Write animationRate only. No diminishing and no draw.

**MMF-R06b** Player initialization requests [verified in 1.13c code (static)]. At the reached
player stat-initialization stage request layer-0 writes, in order, 68=100, 67=100, 69=100. These local requests
do not establish stored totals or complete creation (MMF-Q06).

**MMF-R06c** Monster initialization requests [verified in 1.13c code (static)]. At the reached
monster fixed-stat initialization stage request layer-0 writes, in order, 68=100, 67=75, 69=100.
They do not depend on class in this slice. Later class/property/equipment contributions and
actual write propagation remain outside this slice (MMF-Q04/MMF-Q06). No direct draw in either fixed-write slice; write propagation retains MMF-Q06.

**MMF-R06d** Supplied velocity-to-step operation [verified in 1.13c code (static)]. Consume
NRF-R11a with the path's speed, scale, acceleration, targetSpeed, ramp tick and direction. It
performs its five-tick ramp, defaults nonpositive scale to 1024, computes the low32 products and
arithmetic shifts, and writes stepX/stepY. It draws nothing. This does not establish every caller's path inputs.

**MMF-R06e** Direction scale [verified in 1.13c code (static)]. Consume NRF-R11b: axis 4096 and
equal-axis diagonal 2896/2896. Complete direction/path production is NRF-Q04 (MMF-Q05).

**MMF-R06f** Ordered dispatcher [verified in 1.13c code (static)]. After supplied effective
input/animation preparation, first matching predicate wins. Null unit, actual type 4 or actual
type>=5 returns before preparation; other admitted types still depend on valid preparation.
This contract's rate branches use supplied valid player/monster inputs. No direct dispatcher draw.

| Priority | Selected branch | Calculation / writes |
|---|---|---|
| 1 | cast | v=min(i32(low32(100+D(2))),175), no floor; rate=min(floor(u32(low32(B*v))/100),32767); write animationRate and actionRate |
| 2 | block | HRC-R14: base 100 with valid active state 101 and admitted type, else 50; FBR, minimum 1, cap 32767; animationRate |
| 3 | get-hit | HRC-R12a: 50+D(1), no 15..175 clamp, unsigned /100, minimum 0; animationRate |
| 4 | knockback | HRC-R14: unmodified base mode rate, nonpositive to 0, cap 32767; animationRate then speed request 4096 |
| 5 | movement | MMF-R02/MMF-R04a/MMF-R03 |
| 6 | attack | MMF-R05a/MMF-R05b |
| 7 | base-only | animationRate=clamp(B,0,32767), no stat |
| 8 | default | MMF-R06a |

Cast normally selects player mode 10 or monster 7; player Sequence can also select cast when its
selected sequence record identifies cast. Block/get-hit/knockback normal effective pairs are
player 9/4/19 and monster 6/3/13. Player movement includes 2/3/6. These identities do not replace
the ordered table/skill predicates. HRC-R12b and MMF-Q11 gate actual remapping/selection; supplied
predicates independently establish local precedence and arithmetic. Wrapped products govern
extreme signed inputs; a negative ordinary product interpreted unsigned saturates, but this is
not a general sign shortcut after overflow.

**MMF-R06g** Full dispatcher input production [unresolved]. Actual remapping, selected animation
record, sequence/skill/table predicates and animation preparation effects remain MMF-Q11. No
unverified choice of predicate or base is a default for the supplied consumer rules.

**MMF-R07** Cold effect [verified in 1.13c code (static)]. Consume CSA-R05's ColdEffect lookup:
nonmonster or invalid/unavailable monster record gives -50, otherwise signed difficulty cold
effect E. E=0 returns before list/state/timer/stat requests and draw. Positive E is allowed by cold.

**MMF-R08a** Cold before freeze [verified in 1.13c code (static)]. Consume CSA-R01: after the
current-life write request Cold with coldLength, then Freeze with freezeLength; no threshold
conversion. A fallback can make two cold draws in that order. Other status stages retain CSA ownership.

**MMF-R08b** Freeze routing [verified in 1.13c code (static)]. Consume CSA-R07: nonpositive
length returns. Player invokes full cold with this length. Monster first tests state 54: active
returns without cold or freeze. Otherwise boss, monsterFlag8 and hireling category are tested
in that order; first true/nonzero invokes full cold. A monster passing them reaches freeze.
Other types reach freeze requests then CSA-R09's fatal type assertion; they are not successful
freeze targets. Null defender faults at list lookup on the reached freeze path.

**MMF-R08c** Freeze check identities [verified in 1.13c code (static)]. The tests in MMF-R08b
are NRF-R12a's state test of 54, CSA-R02's boss recordGate, condition-word monsterGate with mask
8 and fixed hireling unitGate. Category results 1/2/3/4 correspond to classes
271/338/359/560-or-561. monsterFlag8 is neutral; no champion/unique classification is inferred.

**MMF-R08d** Freeze cold-effect gate [verified in 1.13c code (static)]. On the reached freeze
path, E>=0 returns before list/state/timer requests; no cold fallback and no draw (CSA-R08).

**MMF-R09** Length strip [verified in 1.13c code (static)]. Consume DPO-R10: only if coldLength>0
or freezeLength>0, nonzero defender 153 sets both to 0; otherwise nonzero 118 halves both signed,
once, including a negative partner. No draw. This precedes length resistance.

**MMF-R10a** Length resistance consumer [verified in 1.13c code (static)]. Consume DMI-R06 and
DMI-R09 entries 5/6. Nonpositive entry is stored as 0. Otherwise use DMI-R07's effective r;
bypass makes positive r zero. Reduction slot 0 is zero. With positive value and nonzero r,
clamp r to at most 100 and compute NRF-R05 mulDiv(a=100-r,b=len,c=100) in that argument order.
Store result; no minimum here, no draw. For small unwrapped inputs this is len*(100-r)/100.

**MMF-R10b** Effective cold resistance [verified in 1.13c code (static)]. Consume DMI-R01,
DMI-R05, DMI-R07 with resist 43, max 44, pierce 335. That contract determines the regular-monster
immunity/pierce exception, expansion/classic penalty distinction, floor -100, maximum-resist cap
and positive-resist bypass. Players and hirelings share its nonregular branch; bosses/champions
are not excluded from regular monsters merely by that classification. No independent formula
or player-only penalty is established here. No draw.

**MMF-R10c** No length absorb [verified in 1.13c code (static)]. DMI entries 5/6 have neither
percent nor flat absorb bindings and use zero reduction slot 0, so no absorb/stat-read/heal step
runs for either length (DMI-R02/DMI-R06/DMI-R08). Cold damage has separate bindings. No draw.

**MMF-R11** Cold duration [verified in 1.13c code (static)]. At supplied cold entry len<=0
returns. With E!=0, monster E<0 divides by MonsterColdDivisor only when nonzero; then len=max(len,1).
End=low32(attackerGame.frame+len). Consume CSA-R05's exact request order. Nonmonster/positive E
skips division. This determines duration and local requests, not actual stat propagation.

**MMF-R12** Freeze duration [verified in 1.13c code (static)]. After routing and E<0, a monster
requires nonzero MonsterFreezeDivisor (zero is fatal) and divides signed, without a minimum.
End=low32(attackerGame.frame+len). CSA-R08 schedules expiry even when len becomes 0; STL-R01
queues an at-current-frame request for next frame. No draw.

**MMF-R13a** List/state/stat requests [verified in 1.13c code (static)]. Consume CSA-R05 and
CSA-R08. New cold requests state 11 on, timed list creation, kind-12 expiry, list state/callback,
full attachment, replacing list writes 67/68/69=E at layer 0 in that order, then animation refresh.
New freeze requests state 1 on, timed list creation, state/callback, full attachment and kind-12
expiry; no speed-stat write. Allocation, attachment, write and refresh execution retain their
own dependency limits; these are ordered requests.

**MMF-R13b** Stored contributions [unresolved]. Whether requested list writes/attachment and
removal produce the selected stat totals and restored rate is STL-R05/STL-R06b/STL-R06c integration
(MMF-Q06). NRF-R12b/NRF-R12c read the selected array; they do not establish summation or mutation.

**MMF-R13c** Freeze tail and callback-local effects [verified in 1.13c code (static)]. Consume
CSA-R09: after freeze list/expiry requests, require monster or fail fatally; traverse the supplied
event chain with successor saved before each request, request removal for kind 2, then schedule
kind 2 at low32(frame+newAdjustedLength+1). Actual node removal is CSA-Q02; mode-bound list processing remains NRF-R14b.

At supplied cold callback entry consume CSA-R13: when dead and StayDeath nonzero, return;
otherwise state s off, shatter 107 off only if alive, then animation refresh. Alive means
aliveExclusionCondition is false AND ((type is player AND mode is outside {0,17}) OR
(type is monster AND mode is outside {0,12})); other types are dead.

At supplied freeze callback entry use that same dead/StayDeath gate. Request state s off; null
unit then fails fatally. Request associated-record cleanup, then mode 1 through NRF-R14a. Only
for a monster with a game, monster data and its record: select Normal aidel when both
modeConditionA/B false, otherwise difficulty aidel; unsigned zero becomes 15. Traverse supplied
event chain saving successor, request kind-2 removals, schedule kind 2 at low32(unitGame.frame+delay).
No direct callback draw; cleanup and refresh execution and actual timer removal retain MMF-Q12.

**MMF-R13d** Refresh and stacking [verified in 1.13c code (static)]. Consume CSA-R05/CSA-R08.
Look up first matching state list (NRF-R12a), rather than require globally unique lists. Existing
cold list keeps its stats: E is looked up anew for gates/duration/draw eligibility but not rewritten.
Only a signed later candidate end requests expiry replacement and kind 12; equal/earlier does
neither. Existing freeze uses max-signed(oldEnd,candidateEnd), replaces only when later and
always requests kind 12 at selected end. Freeze's kind-2 request still uses the new adjusted
length, independently of selected list end. Every eligible cold application still makes MMF-R14a's draw.

**MMF-R13e** Actual lifecycle integration [unresolved]. Attachment/write propagation, expired
contribution removal before callback, timer removal, animation refresh and associated-record
cleanup effects are not inferred from local request order. MMF-Q06/MMF-Q12 gate these results.

**MMF-R14a** Cold side draw [verified in 1.13c code (static)]. Consume CSA-R06: every cold
application reaching len>0 and E!=0 advances defender seed once, unsigned low word modulo 100.
If roll<20, defender is monster and deadCol false, request state 107 on; otherwise request off.
An unavailable record lookup yields false deadCol. New list, extension and unchanged expiry all draw; freeze
proper does not. No draw after length strip suppresses both lengths or E=0.

**MMF-R14b** Shatter consequences [unresolved]. Actual death/corpse behavior from active state
107 is outside this movement boundary (MMF-Q07); its name supplies no corpse rule. MMF-R14a's
state request and draw remain available independently.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MMF-R01 | Diminishing selection | verified in 1.13c code (static) | none | MMF-V07, MMF-V08 |
| MMF-R02 | Movement percentage | verified in 1.13c code (static) | MMF-Q11: actual input preparation | MMF-V01, MMF-V02, MMF-V09 |
| MMF-R03 | Velocity writer | verified in 1.13c code (static) | MMF-Q05/MMF-Q11: full run/path/input production | MMF-V01, MMF-V02, MMF-V09 |
| MMF-R04a | Movement rate | verified in 1.13c code (static) | MMF-Q11/GDF-Q01: selected/derived bases only | MMF-V02, MMF-V09 |
| MMF-R04b | Derived monster bases | unresolved | GDF-Q01: actual derived values | none |
| MMF-R05a | Attack consumer | verified in 1.13c code (static) | MMF-Q08/MMF-Q11: preparation, override/input production | MMF-V03, MMF-V10 |
| MMF-R05b | Paired-item adjustment | verified in 1.13c code (static) | MMF-Q08: original item selection/tests | MMF-V23 |
| MMF-R06a | Other actions | verified in 1.13c code (static) | MMF-Q11: branch/base production | MMF-V11 |
| MMF-R06b | Player fixed requests | verified in 1.13c code (static) | MMF-Q06: stored totals | MMF-V12 |
| MMF-R06c | Monster fixed requests | verified in 1.13c code (static) | MMF-Q04/MMF-Q06: later producers/storage | MMF-V12 |
| MMF-R06d | Step operation | verified in 1.13c code (static) | MMF-Q05: full path input production | MMF-V24 |
| MMF-R06e | Direction scale | verified in 1.13c code (static) | NRF-Q04: full direction production | MMF-V24 |
| MMF-R06f | Dispatcher | verified in 1.13c code (static) | MMF-Q11: actual selection | MMF-V22 |
| MMF-R06g | Input production | unresolved | MMF-Q11: actual preparation/predicates | none |
| MMF-R07 | Cold effect | verified in 1.13c code (static) | none at supplied record boundary | MMF-V13, MMF-V14, MMF-V16 |
| MMF-R08a | Cold then freeze | verified in 1.13c code (static) | none at supplied status stage | MMF-V06, MMF-V21 |
| MMF-R08b | Freeze routing | verified in 1.13c code (static) | none locally | MMF-V06, MMF-V15, MMF-V25, MMF-V28 |
| MMF-R08c | Freeze identities | verified in 1.13c code (static) | none locally | MMF-V15, MMF-V25 |
| MMF-R08d | Nonnegative freeze gate | verified in 1.13c code (static) | none locally | MMF-V16 |
| MMF-R09 | Length strip | verified in 1.13c code (static) | none locally | MMF-V04, MMF-V05 |
| MMF-R10a | Length resistance | verified in 1.13c code (static) | none at supplied DMI stage | MMF-V17 |
| MMF-R10b | Resistance dependency | verified in 1.13c code (static) | none at supplied DMI stage | MMF-V26 |
| MMF-R10c | No length absorb | verified in 1.13c code (static) | none at supplied DMI stage | MMF-V17 |
| MMF-R11 | Cold duration | verified in 1.13c code (static) | MMF-Q06: actual stored effects only | MMF-V06, MMF-V13, MMF-V16 |
| MMF-R12 | Freeze duration | verified in 1.13c code (static) | MMF-Q06/MMF-Q12: actual lifecycle only | MMF-V04, MMF-V18 |
| MMF-R13a | Requests | verified in 1.13c code (static) | MMF-Q06/MMF-Q12: actual effects only | MMF-V06, MMF-V13 |
| MMF-R13b | Stored contribution | unresolved | MMF-Q06: propagation/removal | none |
| MMF-R13c | Callbacks and AI | verified in 1.13c code (static) | MMF-Q12: cleanup/refresh/timer removal; MMF-Q06: callback invocation | MMF-V19, MMF-V27 |
| MMF-R13d | Refresh | verified in 1.13c code (static) | MMF-Q06/MMF-Q12: actual effects only | MMF-V20 |
| MMF-R13e | Lifecycle integration | unresolved | MMF-Q06/MMF-Q12: actual lifecycle | none |
| MMF-R14a | Cold draw/state request | verified in 1.13c code (static) | none locally; MMF-Q06 for later propagation | MMF-V21 |
| MMF-R14b | Corpse effects | unresolved | MMF-Q07: outside boundary | none |

## 5. Outputs and side effects

| Output / request | Rule |
|---|---|
| animationRate and optional actionRate | MMF-R04a/MMF-R05a/MMF-R06a/MMF-R06f |
| path speed/targetSpeed/resetMarker; step/ramp fields through NRF operations | MMF-R03/MMF-R06d; HRC-R14 for knockback |
| fixed initialization stat-write requests | MMF-R06b/MMF-R06c |
| coldLength/freezeLength modifications, no absorb-heal addition by length entries | MMF-R09/MMF-R10a/MMF-R10c |
| list allocation/state/callback/expiry, attachment and 67/68/69 write requests | MMF-R13a/MMF-R13d |
| state 11/1/107 requests; active/changed membership when supplied storage supports STL-R07 | MMF-R13a/MMF-R13c/MMF-R14a |
| kind-12 expiry and kind-2 AI schedule/removal requests | MMF-R12/MMF-R13a/MMF-R13c/MMF-R13d |
| callback cleanup/refresh requests and mode 1 write request | MMF-R13c |
| defender seed advancement | MMF-R14a |
| diminishing helper signed return value or division fault | MMF-R01 |
| eligibility helper boolean return value | MMF-R05b |

No message or persistence write is established. Actual lifecycle/cleanup/refresh and update delivery
remain dependency gates; a logged stat-write request does not assert a stored unit total.

## 6. Branches and edge cases

The movement floor differs from attack/default clamps, so equal supplied cold contributions can
produce different minimum rates. Positive monster E applies cold without division but refuses
freeze; zero E skips both. Cold divisor zero skips division; freeze divisor zero is fatal.
A zero adjusted freeze length retains end=current frame and asks scheduling for that frame,
which STL-R01 queues next frame. Earlier/equal cold refresh still draws; freeze keeps its list end
but computes its AI request from the new duration. Wrapped sums/products precede comparisons;
unsigned action division and signed movement division must remain distinct. A missing movement
path writes neither rate nor speed. Invalid movement class records fault. Successful cold/freeze
calls require valid attacker/game and the reached dependency inputs; no null default is invented.

## 7. Open questions

| Question | Status / blocked assertion | Evidence that would settle it |
|---|---|---|
| **MMF-Q01** Length producers | outside-boundary; none locally with supplied lengths | DPO-R06b/R07 identifies cold and conversion producers; other skill/missile producer contracts must establish their supplied fields |
| **MMF-Q03** Per-class cold effect and AI-delay cells | outside-boundary; supplied record values remain available | Named needed class/column read through original loader and compiled correspondence when actual runtime values are claimed; divisor cells already recorded |
| **MMF-Q04** Later monster contributions | outside-boundary; none locally for fixed requests | Trace named class/property/equipment producer only when that integration is requested; MMF-R06c fixes the reached initialization requests |
| **MMF-Q05** Full run/path input production | integration; full movement trajectory/run baseline | Trace all relevant callers producing modifier totals, direction and path inputs; NRF-R11a/R11b already establishes supplied step and scale |
| **MMF-Q06** Stored contribution and removal | integration; actual cold stat totals/restoration and callback invocation | Complete STL-R05/STL-R06b/STL-R06c propagation, removal and selected-array production; supplied read/request assertions remain available |
| **MMF-Q07** Shatter death use | outside-boundary; none locally for draw/state request | Death/corpse consumer contract tracing active state 107 to its observable corpse result |
| **MMF-Q08** Attack preparation and item/base producers | integration; actual attack base/pair selection and preparation effects/draws | Complete selected inventory slots, type-45 relationship preparation, item validity, attack preparation and override producer with all relevant callees |
| **MMF-Q09** Gameplay confirmation | observation; runtime-confirmation claims only | Controlled original-game observations for all verified rules, including division fault, with the linked procedures; no gameplay performed here |
| **MMF-Q11** Dispatcher preparation and predicates | integration; actual effective type/class/mode, record and sequence/skill selection | Complete preparation/remap and predicate paths through data/skill helpers; supplied effective inputs remain independently usable |
| **MMF-Q12** Callback/timer/refresh execution | integration; actual cleanup, refresh and timer removal/scheduling lifecycle | Complete CSA-Q02/STL-Q07 and associated-record cleanup effects, then original expiry-to-callback integration; supplied logs and NRF mode write stay usable |

Every remaining question and the derived-base dependency is tracked in the
[open evidence gaps](../plans/gaps/state-movement-modifiers-and-freeze.md#open-evidence-gaps).

## 8. Test vectors

### 8.1 Executable vectors

All vectors are supplied-input stages or isolated operations, with no gameplay observation.
Unless overridden: player class 0 or monster class 0, valid records, a non-null base stat list
with state storage and supplied selected entries, valid stat definitions, layer 0, no minimum adjustment,
unit stats 67/68/69=100 and item speed stats 93/96/99/102/105=0; all other mentioned stats 0.
Rates start at 7, actionRate at 9; path present with speed=targetSpeed=256 and resetMarker=7;
B=256, player WalkVelocity=6; all category predicates false unless named. Effective and actual
kind/mode agree. Attack preparation logs a call without writes/draws; no dual-wield or override.
State count=185, states off, no existing cold/freeze timed list, defender alive in neutral mode, frame=1000,
difficulty=0, cold E=-50, cold/freeze divisors=1, monster record available, boss/monsterFlag8/hireling
category/deadCol false, monster Velocity=6 and ordinary/run animation bases=213/101.
All unlisted input fields remain unchanged; no message is asserted.

Request fixtures log direct calls from this contract in order; nested dependency bookkeeping
is not appended to that direct-call log. Successful supplied allocation creates an empty
logical list; supplied expiry/state/callback setters retain their fields; supplied list writes
retain only own-list values. Supplied Attach logs without propagating stats. SetState uses
STL-R07 membership with storage present, but update delivery is not asserted. Scheduling logs
requested frames; actual queue behavior is asserted only where STL-R01 is explicitly consumed.
Timer removal, refresh and cleanup log without effects. Callbacks are invoked explicitly after
supplied removal; that does not verify original removal. Fixed initialization writes only log.
For exact logs below, `coldNew(E,end)` means ordered [state 11 on, new timed list(end), schedule 12(end),
listState11, coldCallback, attach(full), write 67(E), write 68(E), write 69(E), refresh];
`freezeNew(end)` means [state 1 on, new timed list(end), listState1, freezeCallback, attach(full),
schedule 12(end)]. The final shatter request is appended after the cold log. Owner of each new list
is the supplied attacker type/id. Supplied attacker is player type 0/id 1; defender id 2.
Every independent cold case starts defender seed {1,0}, draws once modulo 100=85, ends
{1791398085,0}, and appends state 107 off, unless the case explicitly states no draw or a different seed.
A rate/arithmetic case has Seeds no draw and Draws none; it leaves every seed unchanged.
MMF-V01 additionally records animationRate assignments and path-speed requests in one ordered
log, so their relative order is observable without counting internal data queries.

- **MMF-V01** Player chill supplied totals (MMF-R02/MMF-R03/MMF-R04a; boundary: supplied-input stage)
  - Inputs: player mode 2, movement true; selected stat 67=50, stat 96=0, representing a supplied -50 cold contribution to baseline 100; W=6
  - Seeds: no draw
  - Draws: none
  - Trace: pct=50; animation 213*50/100=106 is stored first, then velocity 1536*50/100=768 is requested
  - Outputs: speed=targetSpeed=768, resetMarker=15, animationRate=106, actionRate=9; stored cold propagation unasserted; ordered log[writeAnimationRate(106),setPathSpeed(768)]
  - Discriminates: rate 106 and speed 768 reject treating chilled movement as pct 25; the ordered log rejects requesting speed before writing animationRate
  - runtime_observed: no

- **MMF-V02** Movement floor (MMF-R02/MMF-R03/MMF-R04a; boundary: supplied-input stage)
  - Inputs: player mode 2, movement true, stat 67=-200, W=6
  - Seeds: no draw
  - Draws: none
  - Trace: pct=max(-200,25)=25; 1536*25/100=384; 5325/100=53
  - Outputs: speed=targetSpeed=384, resetMarker=15, animationRate=53, actionRate=9
  - Discriminates: speed 384 rejects permitting negative movement
  - runtime_observed: no

- **MMF-V03** Attack minimum (MMF-R05a; boundary: supplied-input stage)
  - Inputs: player mode 7, attack true, stat 68=-150, B=256
  - Seeds: no draw
  - Draws: none
  - Trace: v=-150, clamp 15; 3840/100=38
  - Outputs: log[attackPreparation]; animationRate=actionRate=38
  - Discriminates: 38 rejects movement floor 25, which would give 64
  - runtime_observed: no

- **MMF-V04** Half freeze followed by divisor (MMF-R09/MMF-R10a/MMF-R12/MMF-R13a; boundary: supplied-input stage)
  - Inputs: monster, coldLength=0, freezeLength=100 before strip, stat 118=1, effective r=0, bypass false, freeze divisor 4
  - Seeds: no draw
  - Draws: none
  - Trace: strip 50, mitigation 50, duration 12, end 1012
  - Outputs: length 50; log freezeNew(1012), schedule 2(1013); no speed-stat writes, seed unchanged
  - Discriminates: end 1012 rejects ignoring Half, which would give end 1025
  - runtime_observed: no

- **MMF-V05** Cannot Be Frozen (MMF-R09/MMF-R08b; boundary: supplied-input stage)
  - Inputs: monster; coldLength=9, freezeLength=100 before strip, stat 153=-1, stat 118=1
  - Seeds: no draw
  - Draws: none
  - Trace: nonzero 153 wins; both lengths 0; both appliers return
  - Outputs: lengths 0/0; no requests, no seed change
  - Discriminates: negative 153 still suppresses; rejects testing positive only or halving first
  - runtime_observed: no

- **MMF-V06** Player freeze is cold (MMF-R07/MMF-R08a/MMF-R08b/MMF-R11/MMF-R13a; boundary: supplied-input stage)
  - Inputs: player; coldLength=0, freezeLength=60 after mitigation, E=-50; supplied MonsterColdDivisor=4 and MonsterFreezeDivisor=2
  - Seeds: defender {1,0}
  - Draws: defender modulo 100=85
  - Trace: Cold(0) returns; Freeze(60) invokes Cold(60); player skips both divisors, end 1060
  - Outputs: log coldNew(-50,1060), state 107 off; list own 67/68/69=-50, active 11 true, active 1 false; final seed{1791398085,0}; unit totals remain 100
  - Discriminates: one draw and state 11 reject applying state 1 to player; end 1060 rejects dividing by the cold divisor (end 1015) or freeze divisor (end 1030)
  - runtime_observed: no

- **MMF-V07** Diminishing division fault (MMF-R01; boundary: isolated rule)
  - Inputs: player movement branch, stat 96=-150, path present, prior rate 7/speed 256
  - Seeds: no draw
  - Draws: none
  - Trace: 150+(-150)=0; D(4) faults before rate/speed writes
  - Outputs: division fault; animationRate=7, speed=targetSpeed=256, no requests
  - Discriminates: rejects zero-denominator fallback 0
  - runtime_observed: no

- **MMF-V08** Diminishing values (MMF-R01; boundary: isolated rule)
  - (a) Ordinary diminishing values:
    - Inputs: independent (index 4,v=50), (index 0,v=100), (index 0,v-20), (index 0,v=0)
    - Seeds: no draw
    - Draws: none
    - Trace: 7500/200=37; 12000/220=54; -2400/100=-24; zero shortcut
    - Outputs: returns 37,54,-24,0 respectively; no writes
    - Discriminates: 37 rejects rounding 37.5 upward; -24 rejects unsigned input
  - (b) Null diminishing unit:
    - Inputs: direct isolated D(4) operation with a null unit, independently of the dispatcher guard.
    - Seeds: no draw.
    - Draws: none before the fault.
    - Trace: diminishing reads the unit's base-list association without a null-unit guard.
    - Outputs: fault; no return value or writes before the fault.
    - Discriminates: rejects treating null unit as the null-list zero-return case.
  - (c) Null diminishing list:
    - Inputs: direct isolated D(4), non-null player with null base stat list.
    - Seeds: no draw.
    - Draws: none.
    - Trace: missing list supplies v=0; diminishing returns 0.
    - Outputs: return 0, no writes.
    - Discriminates: rejects applying the null-unit fault to a missing list.
  - (d) Invalid stat definition:
    - Inputs: direct isolated D(4), non-null player/base list with selected stat 96=50, definition count=96; stat 96 is out of range.
    - Seeds: no draw.
    - Draws: none.
    - Trace: the definition guard supplies v=0, rather than the stored 50; diminishing returns 0.
    - Outputs: return 0, no writes.
    - Discriminates: rejects using the stored stat despite the invalid definition, which would return 37.
  - runtime_observed: no

- **MMF-V09** Movement modes, missing path and wrapping (MMF-R02/MMF-R03/MMF-R04a; boundary: supplied-input stage)
  - (a) Walk:
    - Inputs: player mode 2, movement true, stat 67=100, stat 96=50, W=6
    - Seeds: no draw
    - Draws: none
    - Trace: D=37; pct 137; velocity 210432/100=2104; rate 29181/100=291
    - Outputs: speed=targetSpeed=2104, resetMarker=15, rate 291, actionRate=9
    - Discriminates: rejects diminishing the sum instead of item FRW
  - (b) Run:
    - Inputs: player mode 3, movement true, stat 67=100, stat 96=50, W=6, supplied CharStats.RunVelocity=9
    - Seeds: no draw
    - Draws: none
    - Trace: pct 137; same velocity 2104; 101*137/100=138
    - Outputs: speed=targetSpeed=2104, resetMarker=15, rate 138, actionRate=9
    - Discriminates: same velocity rejects separate RunVelocity base here
  - (c) Missing path:
    - Inputs: player mode 2, movement true, stat 67=100, stat 96=50, W=6, path absent, rate 7
    - Seeds: no draw
    - Draws: none
    - Trace: movement branch returns before stat reads
    - Outputs: rate 7/actionRate=9; no rate/speed writes
    - Discriminates: rejects writing animationRate despite absent path
  - (d) Wrapped products:
    - Inputs: player mode 2, movement true, stat 67=16777216, stat 96=0, W=6
    - Seeds: no draw
    - Draws: none
    - Trace: pct 16777216; low32(1536*pct)=0; i32(low32(213*pct))=-721420288; trunc/100=-7214202, clamp 0
    - Outputs: speed=targetSpeed=0, resetMarker=15, rate 0, actionRate=9
    - Discriminates: rejects unbounded-integer products followed by clamp
  - (e) Monster ordinary base and velocity:
    - Inputs: monster effective/actual mode 2, movement true, stat 67=150, stat 96=0, supplied Velocity=8, ordinary animation base=200 and run base=100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: pct=150; velocity=(8*256)*150/100=3072; ordinary rate=200*150/100=300.
    - Outputs: speed=targetSpeed=3072, resetMarker=15, animationRate=300, actionRate=9.
    - Discriminates: rejects using player W=6 or the monster run animation base; either changes an asserted result.
  - (f) Monster run base:
    - Inputs: monster effective/actual mode 15, movement true, stat 67=150, stat 96=0, supplied Velocity=8, ordinary animation base=200 and run base=100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: pct=150; velocity=(8*256)*150/100=3072; run rate=100*150/100=150.
    - Outputs: speed=targetSpeed=3072, resetMarker=15, animationRate=150, actionRate=9.
    - Discriminates: rejects using the ordinary monster animation base, which would store 300.
  - (g) Invalid velocity record (isolated MMF-R03 operation):
    - Inputs: monster class 1, supplied MonStats count=1, no class record, pct=150; operation starts at velocity-base selection, before any velocity write.
    - Seeds: no draw.
    - Draws: none.
    - Trace: class at count yields no record; the signed Velocity read faults.
    - Outputs: fault; no path write, speed=targetSpeed=256 and resetMarker=7 unchanged.
    - Discriminates: rejects silently substituting a zero velocity for an invalid record.
  - runtime_observed: no

- **MMF-V10** Attack clamp, Sequence and override (MMF-R05a; boundary: supplied-input stage)
  - (a) Ordinary:
    - Inputs: player mode 7, attack true, stat 68=100, stat 93=100, B=256
    - Seeds: no draw
    - Draws: none
    - Trace: D=54, v=154; 39424/100=394
    - Outputs: log[attackPreparation], animationRate=actionRate=394
    - Discriminates: rejects raw IAS=100
  - (b) Sequence:
    - Inputs: player actual/effective mode 18, attack true, cast false, stat 68=100, stat 93=100, B=256
    - Seeds: no draw
    - Draws: none
    - Trace: 154-30=124;31744/100=317
    - Outputs: log[attackPreparation], animationRate=actionRate=317
    - Discriminates: rejects skipping Sequence penalty
  - (c) Upper clamp:
    - Inputs: player mode 7, attack true, stat 68=100, stat 93=1000, B=256
    - Seeds: no draw
    - Draws: none
    - Trace: D=107, v=207 clamps to 175;44800/100=448
    - Outputs: log[attackPreparation], animationRate=actionRate=448
    - Discriminates: rejects cap before adding IAS
  - (d) Nonzero override:
    - Inputs: player mode 7, attack true, stat 68=100, stat 93=100, B=256; supplied override predicate true/result 512
    - Seeds: no draw
    - Draws: none
    - Trace: 154*512/100=788
    - Outputs: log[attackPreparation], animationRate=actionRate=788
    - Discriminates: rejects ignoring supplied nonzero override
  - (e) Zero override:
    - Inputs: player mode 7, attack true, stat 68=100, stat 93=100, B=256; supplied override predicate true/result 0
    - Seeds: no draw
    - Draws: none
    - Trace: zero result retains B=256; rate 394
    - Outputs: log[attackPreparation], animationRate=actionRate=394
    - Discriminates: rejects treating predicate alone as base 0
  - (f) Wrapped attack product:
    - Inputs: player mode 7, attack true, stat 68=100, stat 93=0, B=1073741824; pair and override inactive.
    - Seeds: no draw.
    - Draws: none.
    - Trace: v=100; low32(1073741824*100)=0; unsigned division returns 0, below the cap.
    - Outputs: log[attackPreparation]; animationRate=actionRate=0.
    - Discriminates: rejects an unbounded product followed by division/cap, which would store 32767.
  - runtime_observed: no

- **MMF-V11** Other-action clamp (MMF-R06a; boundary: supplied-input stage)
  - (a) Ordinary factor clamp:
    - Inputs: default branch, B=256, independent stat 69=200,50,5
    - Seeds: no draw
    - Draws: none
    - Trace: clamp 175,50,15; unsigned products 44800,12800,3840 /100
    - Outputs: animationRate=448,128,38 respectively; actionRate=9
    - Discriminates: rejects diminishing or using stat 68
  - (b) Negative product interpreted unsigned:
    - Inputs: default branch, B=-1, stat 69=100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: factor=100; low32(B*100) is -100 as i32 and 4294967196 as u32; floor(unsigned product/100)=42949671, capped to 32767.
    - Outputs: animationRate=32767, actionRate=9 unchanged; no path write.
    - Discriminates: rejects signed division followed by a zero floor, which would store 0.
  - (c) Wrapped default product:
    - Inputs: default branch, B=1073741824, stat 69=100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: factor=100; low32(B*100)=0; unsigned division gives 0.
    - Outputs: animationRate=0, actionRate=9 unchanged; no path write.
    - Discriminates: rejects unbounded multiplication/division followed by cap, which would store 32767.
  - runtime_observed: no

- **MMF-V12** Fixed initialization requests (MMF-R06b/MMF-R06c; boundary: supplied-input stage)
  - (a) Player:
    - Inputs: reached player fixed-stat initialization stage; supplied CharStats.WalkVelocity=6 and RunVelocity=9.
    - Seeds: no draw
    - Draws: none
    - Trace: three fixed write requests carry 100; the competing run/walk derivation would produce trunc(100*9/6)=150 for stat 67.
    - Outputs: log[writeUnit(68,100,layer 0),writeUnit(67,100,layer 0),writeUnit(69,100,layer 0)]; stored totals unasserted
    - Discriminates: stat 67 request 100 rejects the supplied run/walk ratio result 150; no internal table-query count is asserted.
  - (b) Monster:
    - Inputs: reached monster fixed-stat initialization stage, valid monster data and records; prior helper outcomes supplied
    - Seeds: no draw
    - Draws: none
    - Trace: fixed writes do not select by class
    - Outputs: log[writeUnit(68,100,layer 0),writeUnit(67,75,layer 0),writeUnit(69,100,layer 0)]; stored totals unasserted
    - Discriminates: 75 rejects initializing monster movement percentage 100
  - (c) Supplied player baseline after initialization:
    - Inputs: player mode 2, movement true, supplied selected stat 67=100/stat 96=0, W=6; no propagation from either preceding case is assumed.
    - Seeds: no draw.
    - Draws: none.
    - Trace: pct=100, velocity=1536*100/100=1536=W*256; rate=213.
    - Outputs: speed=targetSpeed=1536, resetMarker=15, animationRate=213, actionRate=9.
    - Discriminates: supplied velocity 1536 rejects using monster baseline 75 for a player; the creation-to-total integration remains gated.
  - runtime_observed: no

- **MMF-V13** Monster cold division/minimum (MMF-R07/MMF-R11/MMF-R13a; boundary: supplied-input stage)
  - (a) Divided:
    - Inputs: monster, E=-33, cold length 75, cold divisor 2
    - Seeds: defender{1,0}
    - Draws: defender modulo 100 = 85
    - Trace: 75/2=37;end 1037
    - Outputs: log coldNew(-33,1037), state 107 off; own 67/68/69=-33, active 11 true; final seed{1791398085,0}
    - Discriminates: 37 rejects rounding division upward
  - (b) Minimum:
    - Inputs: monster, E=-33, cold length 1, cold divisor 2
    - Seeds: defender{1,0}
    - Draws: defender modulo 100 = 85
    - Trace: 1/2=0;minimum 1;end 1001
    - Outputs: log coldNew(-33,1001), state 107 off; own 67/68/69=-33; final seed{1791398085,0}
    - Discriminates: minimum 1 rejects zero cold duration
  - (c) Zero divisor:
    - Inputs: monster, E=-33, cold length 75, cold divisor 0
    - Seeds: defender{1,0}
    - Draws: defender modulo 100 = 85
    - Trace: zero divisor skips;end 1075
    - Outputs: log coldNew(-33,1075), state 107 off; final seed{1791398085,0}
    - Discriminates: rejects freeze-style fatal zero divisor
  - (d) Invalid monster class cold fallback:
    - Inputs: monster class 1, supplied MonStats count=1, unavailable class record; cold length=10, divisor=1; other cold-entry inputs valid.
    - Seeds: defender {1,0}.
    - Draws: defender modulo 100 = 85.
    - Trace: ColdEffect lookup yields -50, not 0; 10/1=10, end=1010.
    - Outputs: log coldNew(-50,1010), state 107 off; own list stats 67/68/69=-50; final seed {1791398085,0}.
    - Discriminates: rejects treating an invalid class as E=0 and skipping cold.
  - (e) Negative monster class cold fallback:
    - Inputs: monster class -1, supplied MonStats count=1; the class index is below the valid range, so no class record is selected; cold length=10, divisor=1; other cold-entry inputs valid.
    - Seeds: defender {1,0}.
    - Draws: defender modulo 100 = 85.
    - Trace: the negative class lookup yields E=-50; duration=10, end=1010.
    - Outputs: log coldNew(-50,1010), state 107 off; own list stats 67/68/69=-50; final seed {1791398085,0}.
    - Discriminates: rejects omitting the signed lower-bound class check or substituting E=0 for a negative class.
  - (f) Negative incoming length:
    - Inputs: monster, cold length=-1, valid supplied cold effect=-50; no existing cold list.
    - Seeds: defender {1,0}.
    - Draws: none.
    - Trace: signed len<=0 returns before duration minimum, requests and draw.
    - Outputs: no requests; state 11 remains off, defender seed {1,0} unchanged.
    - Discriminates: rejects raising every incoming length to the cold minimum before the entry guard.
  - (g) Wrapped cold expiry request:
    - Inputs: monster, cold length=20, E=-50, cold divisor=1, attacker game frame=2147483640; new cold list.
    - Seeds: defender {1,0}.
    - Draws: defender modulo 100 = 85.
    - Trace: duration=20; low32(2147483640+20) interpreted i32 is -2147483636.
    - Outputs: log coldNew(-50,-2147483636), state 107 off; supplied list expiry=-2147483636, own stats 67/68/69=-50; final defender seed {1791398085,0}. Only requested timer frame is asserted, not actual queue execution.
    - Discriminates: rejects retaining the unbounded positive expiry 2147483660 or saturating it to 2147483647.
  - runtime_observed: no

- **MMF-V14** Zero cold effect (MMF-R07; boundary: supplied-input stage)
  - Inputs: monster E0, cold length 50
  - Seeds: defender{1,0}
  - Draws: none
  - Trace: E0 returns before requests/draw
  - Outputs: no requests; seed{1,0}; state 11 off
  - Discriminates: rejects drawing before E gate
  - runtime_observed: no

- **MMF-V15** Protected monster freeze (MMF-R08b/MMF-R08c; boundary: supplied-input stage)
  - Inputs: monster state 54 on, boss true, freeze length 50, cold length 0
  - Seeds: defender{1,0}
  - Draws: none
  - Trace: state 54 tested first returns without fallback
  - Outputs: no requests; state 1/state 11 off; state 54 on; seed{1,0}
  - Discriminates: boss would fallback if tested first, so no draw rejects wrong check order
  - runtime_observed: no

- **MMF-V16** Positive cold effect (MMF-R07/MMF-R08d/MMF-R11; boundary: supplied-input stage)
  - (a) Cold:
    - Inputs: monster E+10, cold length 50, cold divisor 2
    - Seeds: defender{1,0}
    - Draws: defender modulo 100 = 85
    - Trace: E>=0 skips division;end 1050
    - Outputs: log coldNew(10,1050), state 107 off; own 67/68/69=10; seed{1791398085,0}
    - Discriminates: rejects dividing positive E duration
  - (b) Freeze:
    - Inputs: monster E+10, freeze length 50, all eligibility gates false
    - Seeds: defender{1,0}
    - Draws: none
    - Trace: freeze E>=0 returns
    - Outputs: no requests; seed{1,0}
    - Discriminates: rejects cold fallback after E gate
  - runtime_observed: no

- **MMF-V17** Length resistance, rounding and no absorb (MMF-R10a/MMF-R10c; boundary: supplied-input stage)
  - (a) Small length/resistance cases:
    - Inputs: independent supplied effective (length,r)=(100,40),(100,-50),(100,100),(100,150),(7,33); bypass false; both length entries have zero reduction/no absorb bindings; absorbHeal9; defender cold-damage absorb stats 148/149=999
    - Seeds: no draw
    - Draws: none
    - Trace: factors 60,150,0,0,67; NRF mulDiv gives 60,150,0,0,4
    - Outputs: stored length 60,150,0,0,4 respectively; absorbHeal9 unchanged, no absorb-stat reads
    - Discriminates: 7 case rejects rounding 4.69 upward; nonzero absorb stats reject cold-damage bindings for length
  - (b) Caller operand order:
    - Inputs: supplied effective length=65537, r=33, bypass false; zero reduction slot and no length absorb bindings; absorbHeal=9.
    - Seeds: no draw.
    - Draws: none.
    - Trace: factor=67; NRF-R05 receives (a=67,b=65537,c=100). Neither b>1048576 nor a>65536 holds, so trunc(67*65537/100)=43909.
    - Outputs: stored length=43909, absorbHeal=9 unchanged.
    - Discriminates: swapped operands (a=65537,b=67,c=100) select divide-first trunc(65537/100)*67=43885, which fails the required 43909.
  - runtime_observed: no

- **MMF-V18** Freeze division and fatal zero (MMF-R12/MMF-R13a; boundary: supplied-input stage)
  - (a) Truncation:
    - Inputs: monster freeze length 7, divisor 4
    - Seeds: no draw
    - Draws: none
    - Trace: duration 1;end 1001
    - Outputs: log freezeNew(1001), schedule 2(1002); no speed-stat writes
    - Discriminates: rejects rounding duration up to 2
  - (b) No minimum:
    - Inputs: monster freeze length 3, divisor 4
    - Seeds: no draw
    - Draws: none
    - Trace: duration 0;end 1000; consume STL-R01 scheduling for requested frame 1000
    - Outputs: log freezeNew(1000), schedule 2(1001); list expiry 1000, kind 12 queue frame 1001; no draw
    - Discriminates: rejects minimum 1: it would make AI request 1002
  - (c) Zero divisor:
    - Inputs: monster freeze length 3, divisor 0
    - Seeds: no draw
    - Draws: none
    - Trace: fatal divisor assertion before list creation
    - Outputs: fatal; no list/state/timer requests, no stat writes
    - Discriminates: rejects cold zero-divisor skip
  - (d) Wrapped freeze expiry request:
    - Inputs: monster, freeze length=20, freeze divisor=1, attacker game frame=2147483640; no prior list or event nodes; eligibility gates false and E=-50.
    - Seeds: no draw.
    - Draws: none.
    - Trace: adjusted length=20; expiry i32(low32(2147483640+20))=-2147483636; AI request i32(low32(2147483640+20+1))=-2147483635.
    - Outputs: log freezeNew(-2147483636), schedule 2(-2147483635); supplied list expiry=-2147483636; no stat writes or draw. Actual queue execution is not asserted.
    - Discriminates: rejects retaining unbounded positive expiry/AI request frames 2147483660/2147483661 or saturating them.
  - runtime_observed: no

- **MMF-V19** Freeze AI and supplied expiry callback (MMF-R13c; boundary: supplied-input stage)
  - Inputs: monster freeze length 20/divisor 1 at 1000; event chain nodes A(kind 2)->B(kind 3)->C(kind 2); then explicitly invoke callback state 1 at frame 1020 after supplied list removal, unit mode 1; aidel Normal=0, conditionsA/B false
  - Seeds: no draw
  - Draws: none
  - Trace: start duration 20,end 1020,AI1021; callback alive clears 1, cleanup, mode 1, delay 0->15
  - Outputs: start log[freezeNew(1020),removeNodeA,removeNodeC,schedule 2(1021)]; callback log[state 1 off,cleanup,mode 1,removeNodeA,removeNodeC,schedule 2(1035)]; active 1 false, mode 1; stats 67/68/69 remain 100; node removal/storage restoration unasserted
  - Discriminates: delay 15 and 1035 reject rescheduling at expiry+1; no stat writes reject freeze applying -50. This vector logs supplied traversal requests; saved-successor behavior with mutating removal is the consumed CSA-R09 contract, discriminated by CSA-V11(a), and is not independently established by this log-only fixture
  - runtime_observed: no

- **MMF-V20** Extend-only refresh uses new duration for AI (MMF-R13d; boundary: supplied-input stage)
  - (a) Cold shorter:
    - Inputs: monster existing cold list end 1050, own 67/68/69=-50, state 11 on, new E=-25, length 40 at 1000
    - Seeds: defender{1,0}
    - Draws: defender modulo 100 = 85
    - Trace: candidate 1040<1050, no expiry write; still draw
    - Outputs: end 1050 and own stats-50; log[state 107 off]; seed{1791398085,0}
    - Discriminates: rejects replacing E or suppressing draw on unchanged list
  - (b) Cold longer:
    - Inputs: monster existing cold list end 1050, own 67/68/69=-50, state 11 on, new E=-25, length 60 at 1000
    - Seeds: defender{1,0}
    - Draws: defender modulo 100 = 85
    - Trace: candidate 1060>1050 extends without rewriting stats
    - Outputs: end 1060, own stats-50; log[setExpiry(1060),schedule 12(1060),state 107 off]; seed{1791398085,0}
    - Discriminates: rejects retaining old end or rewriting E
  - (c) Freeze shorter:
    - Inputs: monster existing freeze list end 1050, state 1 on, new length 20 at 1000, no event nodes
    - Seeds: no draw
    - Draws: none
    - Trace: candidate 1020<=1050 keeps end; schedules expiry 1050 but AI1000+20+1
    - Outputs: end 1050; log[schedule 12(1050),schedule 2(1021)]; no stat writes
    - Discriminates: AI1021 rejects deriving AI from selected list expiry 1050
  - runtime_observed: no

- **MMF-V21** Cold seeded side draw and ordered double call (MMF-R08a/MMF-R14a; boundary: supplied-input stage)
  - (a) On:
    - Inputs: monster cold length 10, E=-50, deadCol false
    - Seeds: defender{7,0}
    - Draws: defender modulo 100=3
    - Trace: step->{3949852003,2};3<20
    - Outputs: log coldNew(-50,1010),state 107 on; seed{3949852003,2}; active 107 true
    - Discriminates: on rejects percent threshold reversed
  - (b) Off:
    - Inputs: monster cold length 10, E=-50, deadCol false
    - Seeds: defender{1,0}
    - Draws: defender modulo 100=85
    - Trace: step->{1791398085,0};85>=20
    - Outputs: log coldNew(-50,1010),state 107 off; seed{1791398085,0}; active 107 false
    - Discriminates: off rejects always setting shatter
  - (c) Two player applications:
    - Inputs: player coldLength=10,freezeLength=10 after mitigation, no cold list at entry; supplied first-call list persists into second call
    - Seeds: defender{7,0}
    - Draws: first modulo 100=3, then modulo 100=73 on stepped seed
    - Trace: cold creates end 1010, player requests off despite 3; freeze invokes cold, same end makes no list/expiry request; second step->{3459745073,1647453129}
    - Outputs: ordered log[coldNew(-50,1010),state 107 off,state 107 off]; seed{3459745073,1647453129}; own 67/68/69=-50, end 1010; two draws, no state 1
    - Discriminates: final seed rejects reuse of entry seed or only one call; second off rejects monster-only threshold test on player
  - (d) Dead-collision suppresses:
    - Inputs: monster cold length 10, deadCol true
    - Seeds: defender{7,0}
    - Draws: defender modulo 100=3
    - Trace: 3<20 but deadCol true
    - Outputs: log coldNew(-50,1010),state 107 off; seed{3949852003,2}
    - Discriminates: rejects ignoring deadCol
  - (e) Unavailable extended record:
    - Inputs: monster cold length 10, unavailable MonStats2 lookup
    - Seeds: defender{7,0}
    - Draws: defender modulo 100=3
    - Trace: lookup false,3<20
    - Outputs: log coldNew(-50,1010),state 107 on; seed{3949852003,2}
    - Discriminates: rejects defaulting unavailable deadCol true
  - runtime_observed: no

- **MMF-V22** Ordered rate branches (MMF-R06f; boundary: supplied-input stage)
  - (a) Cast:
    - Inputs: player mode 10, cast true, stat 105=20, B=256
    - Seeds: no draw
    - Draws: none
    - Trace: D=17;factor 117;29952/100=299
    - Outputs: animationRate=actionRate=299
    - Discriminates: rejects default other rate 256
  - (b) Cast cap:
    - Inputs: player mode 10, cast true, stat 105=200, B=256
    - Seeds: no draw
    - Draws: none
    - Trace: D=75;factor 175;44800/100=448
    - Outputs: animationRate=actionRate=448
    - Discriminates: rejects no cast cap
  - (c) Block:
    - Inputs: player mode 9, block true, stat 102=30, state 101 off
    - Seeds: no draw
    - Draws: none
    - Trace: D=24;base 50;18944/100=189
    - Outputs: animationRate=189, actionRate=9
    - Discriminates: rejects FHR instead of FBR
  - (d) Get-hit:
    - Inputs: player mode 4, get-hit true, stat 99=30
    - Seeds: no draw
    - Draws: none
    - Trace: D=24;factor 74;18944/100=189
    - Outputs: animationRate=189, actionRate=9
    - Discriminates: rejects default action clamp
  - (e) Player knockback:
    - Inputs: player mode 19, knockback true, stat 99=120
    - Seeds: no draw
    - Draws: none
    - Trace: player nonrun base 213;speed 4096
    - Outputs: animationRate=213,actionRate=9,speed=targetSpeed=4096,resetMarker=15
    - Discriminates: rejects player Run base 101 or reading FHR
  - (f) Monster get-hit:
    - Inputs: monster mode 3, get-hit true, stat 99=0
    - Seeds: no draw
    - Draws: none
    - Trace: D=0;50*256/100=128
    - Outputs: animationRate=128, actionRate=9
    - Discriminates: rejects minimum factor 100
  - (g) Precedence:
    - Inputs: player mode 10, cast true AND block true, stat 105=20, stat 102=30
    - Seeds: no draw
    - Draws: none
    - Trace: first cast wins, rate 299
    - Outputs: animationRate=actionRate=299; no block branch
    - Discriminates: actionRate=299 rejects both predicates executing
  - (h) Base-only:
    - Inputs: base-only true, B=40000, stat 69=5
    - Seeds: no draw
    - Draws: none
    - Trace: clamp base 40000 to 32767
    - Outputs: animationRate=32767,actionRate=9
    - Discriminates: rejects default stat 69 scaling
  - (i) Holy Shield:
    - Inputs: player mode 9,block true,state 101 on,stat 102=0
    - Seeds: no draw
    - Draws: none
    - Trace: base 100;25600/100=256
    - Outputs: animationRate=256,actionRate=9
    - Discriminates: rejects base 50 despite active 101
  - (j) Block minimum:
    - Inputs: player mode 9,block true,state 101 off,B=1,stat 102=0
    - Seeds: no draw
    - Draws: none
    - Trace: 50/100=0,block forces 1
    - Outputs: animationRate=1,actionRate=9
    - Discriminates: rejects get-hit zero minimum
  - (k) Early excluded units:
    - Inputs: independent null unit, actual type 4 unit and actual type 5 unit; for the two present units every supplied category predicate is true, animationRate=7, actionRate=9, path speed=targetSpeed=256, resetMarker=7.
    - Seeds: no draw.
    - Draws: none.
    - Trace: the null/type guard returns before preparation and rate dispatch.
    - Outputs: no requests or rate/path writes in all three cases; present units retain animationRate=7, actionRate=9 and the supplied path fields; absent unit has no output fields.
    - Discriminates: rejects falling through to rate calculations for excluded units.
  - (l) Block before get-hit:
    - Inputs: player mode 9; block and get-hit true, all other predicates false; FBR stat 102=30, FHR stat 99=0, state 101 off.
    - Seeds: no draw.
    - Draws: none.
    - Trace: block wins; D(3)=24, (50+24)*256/100=189.
    - Outputs: animationRate=189, actionRate=9, path fields unchanged.
    - Discriminates: get-hit first would produce 128.
  - (m) Get-hit before knockback:
    - Inputs: player mode 4; get-hit and knockback true, all other predicates false; FHR stat 99=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: get-hit wins; 50*256/100=128.
    - Outputs: animationRate=128, actionRate=9, speed=targetSpeed=256 and resetMarker=7 unchanged.
    - Discriminates: knockback first would set rate 213 and speed 4096.
  - (n) Knockback before movement:
    - Inputs: player mode 19; knockback and movement true, all other predicates false.
    - Seeds: no draw.
    - Draws: none.
    - Trace: knockback wins with base 213 and speed request 4096.
    - Outputs: animationRate=213, actionRate=9, speed=targetSpeed=4096, resetMarker=15.
    - Discriminates: movement first would set speed 1536.
  - (o) Movement before attack:
    - Inputs: player mode 2; movement and attack true, all other predicates false.
    - Seeds: no draw.
    - Draws: none.
    - Trace: movement wins; pct=100, rate=213, velocity=1536.
    - Outputs: animationRate=213, actionRate=9, speed=targetSpeed=1536, resetMarker=15; no attackPreparation request.
    - Discriminates: attack first would write both rates 256 and log attackPreparation.
  - (p) Attack before base-only:
    - Inputs: player mode 7; attack and base-only true, all other predicates false.
    - Seeds: no draw.
    - Draws: none.
    - Trace: attack wins; v=100, 256*100/100=256.
    - Outputs: log[attackPreparation]; animationRate=actionRate=256; path unchanged.
    - Discriminates: base-only first would retain actionRate=9 and skip attackPreparation.
  - (q) Base-only before default:
    - Inputs: player mode 1; only base-only true, B=256, stat 69=5.
    - Seeds: no draw.
    - Draws: none.
    - Trace: base-only stores B=256; it does not use the default factor clamped to 15.
    - Outputs: animationRate=256, actionRate=9, path unchanged.
    - Discriminates: default first would store animationRate=38.
  - runtime_observed: no

- **MMF-V23** Paired-item consumer and eligibility (MMF-R05a/MMF-R05b; boundary: supplied-input stage)
  - (a) Average:
    - Inputs: player class 4 mode 7,attack true,paired eligibility true,both accepted/type-qualified,S1=60,S2=20,stat 68=100
    - Seeds: no draw
    - Draws: none
    - Trace: trunc 80/2-60=-20;v=80;20480/100=204
    - Outputs: log[attackPreparation];animationRate=actionRate=204
    - Discriminates: rejects adding average 40 without removing S1
  - (b) Negative odd sum:
    - Inputs: player class 4 mode 7,attack true,paired eligibility true,both accepted/type-qualified,S1=-3,S2=0,stat 68=100
    - Seeds: no draw
    - Draws: none
    - Trace: trunc(-3/2)=-1;adjust+2;v=102;26112/100=261
    - Outputs: log[attackPreparation];animationRate=actionRate=261
    - Discriminates: rejects floor(-3/2)=-2, which gives 258
  - (c) Failed qualification:
    - Inputs: player class 4 mode 7,attack true,paired eligibility true,first qualified/second false,S1=60,S2=20
    - Seeds: no draw
    - Draws: none
    - Trace: no adjustment;v=100
    - Outputs: log[attackPreparation];animationRate=actionRate=256
    - Discriminates: rejects applying average when only one qualifies
  - (d) Eligibility helper:
    - Inputs: independent null, player class 4,player class 6,player class 0,monster class 417/418/419 with MonStats count 734
    - Seeds: no draw
    - Draws: none
    - Trace: fixed class comparisons only
    - Outputs: returns 0,1,1,0,1,1,0 respectively;no writes
    - Discriminates: rejects all players or all monsters qualifying
  - (e) Ineligible paired consumer:
    - Inputs: player class 0 mode 7, attack true; supplied eligibility false, both items accepted/type-qualified with S1=60/S2=20; stat 68=100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: eligibility false skips averaging; v=100, rate=256.
    - Outputs: log[attackPreparation]; animationRate=actionRate=256.
    - Discriminates: ignoring the eligibility result would apply the -20 adjustment and store 204.
  - (f) Wrapped pair sum:
    - Inputs: player class 4 mode 7, attack true, paired eligibility true, both items accepted/type-qualified; S1=S2=2147483647, stat 68=100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: i32(low32(S1+S2))=-2; trunc(-2/2)=-1; adjustment=-1-2147483647=-2147483648; v=-2147483548 clamps to 15; floor(256*15/100)=38.
    - Outputs: log[attackPreparation]; animationRate=actionRate=38.
    - Discriminates: an unbounded pair sum would average to S1, give adjustment 0 and store rate 256.
  - (g) Wrapped pair adjustment:
    - Inputs: player class 4 mode 7, attack true, paired eligibility true, both items accepted/type-qualified; S1=-2147483648, S2=2147483647, stat 68=100.
    - Seeds: no draw.
    - Draws: none.
    - Trace: sum=-1, trunc(-1/2)=0; low32(0-S1) interpreted i32 is -2147483648; adding 100 gives -2147483548, clamps to 15, rate=38.
    - Outputs: log[attackPreparation]; animationRate=actionRate=38.
    - Discriminates: an unbounded adjustment would be positive 2147483648, driving the factor to 175 and rate 448.
  - runtime_observed: no

- **MMF-V24** Supplied path step and direction scale (MMF-R06d/MMF-R06e; boundary: supplied-input stage)
  - (a) Axis:
    - Inputs: speed 256,target 256,scale 0,accel 0,tick 0,dirX=4096,dirY=0
    - Seeds: no draw
    - Draws: none
    - Trace: k=1024;s=262144>>6=4096;stepX=16777216>>12=4096
    - Outputs: stepX=4096,stepY=0;speed 256,target 256,tick 0 unchanged
    - Discriminates: rejects axis scale 1 or division instead of arithmetic shifts
  - (b) Diagonal:
    - Inputs: speed 256,target 256,scale 1024,accel 0,tick 0,dirX=2896,dirY=2896
    - Seeds: no draw
    - Draws: none
    - Trace: s=4096;2896*4096>>12=2896
    - Outputs: stepX=stepY=2896;speed 256,target 256,tick 0 unchanged
    - Discriminates: rejects unnormalized 4096/4096 diagonal
  - runtime_observed: no

- **MMF-V25** Freeze fallback identities (MMF-R08b/MMF-R08c; boundary: supplied-input stage)
  - Inputs: three independent monsters, freeze length 50, state 54 off: (1)boss true; (2)boss false/monsterFlag8 true; (3)boss false/flag 8 false,class 338 category 2
  - Seeds: defender{1,0} independently
  - Draws: one defender modulo 100 = 85 each
  - Trace: first nonzero remaining test invokes cold;duration 50
  - Outputs: each log coldNew(-50,1050),state 107 off;active 11 true/1 false;seed{1791398085,0}
  - Discriminates: state 11 and one draw reject either suppressing or freezing these gates
  - runtime_observed: no

- **MMF-V26** Resistance owner distinctions (MMF-R10a/MMF-R10b; boundary: supplied-input stage)
  - Inputs: five independent DMI stages, length 100,bypass false,coldresist 40,maxcoldresist 0,pierce 0 unless overridden: (1)Expansion player Nightmare penalty-40;(2)Expansion player Hell penalty-100;(3)Classic player Nightmare;(4)regular monster Normal,res 100,pierce 50;(5)hireling monster class 338 Normal,res 100,pierce 50
  - Seeds: no draw
  - Draws: none
  - Trace: DMI r=0,-60,20,100,50 respectively;immune regular monster skips pierce;hireling permits it
  - Outputs: stored length 100,160,80,0,50 respectively;no draws or absorb-heal changes
  - Discriminates: rejects using classic penalties in Expansion or piercing immune regular monster
  - runtime_observed: no

- **MMF-V27** Callback gate and AI column (MMF-R13c; boundary: supplied-input stage)
  - (a) Retained dead:
    - Inputs: cold callback state 11,monster mode 12,StayDeath nonzero,active 11/107 true
    - Seeds: no draw
    - Draws: none
    - Trace: dead and stay-death returns
    - Outputs: no requests;active 11/107 true
    - Discriminates: rejects clearing state despite stay-death
  - (b) Dead without retention:
    - Inputs: cold callback state 11,monster mode 12,StayDeath=0,active 11/107 true
    - Seeds: no draw
    - Draws: none
    - Trace: state 11 off;dead means no 107 clear;refresh
    - Outputs: log[state 11 off,refresh];active 11 false,107 true
    - Discriminates: rejects unconditional shatter clear
  - (c) Normal delay selection:
    - Inputs: freeze callback state 1,monster mode 1,unit game frame 1000/difficulty 2,modeConditions false/false,aidel Normal=7/Hell=2,empty event chain
    - Seeds: no draw
    - Draws: none
    - Trace: alive clears 1;cleanup;mode 1;select Normal 7
    - Outputs: log[state 1 off,cleanup,mode 1,schedule 2(1007)];mode 1
    - Discriminates: rejects always using Hell column
  - (d) Difficulty delay selection:
    - Inputs: freeze callback state 1, monster mode 1, unit game frame 1000/difficulty 2, modeConditions true/false, aidel Normal=7/Hell=2, empty event chain
    - Seeds: no draw
    - Draws: none
    - Trace: select Hell 2
    - Outputs: log[state 1 off,cleanup,mode 1,schedule 2(1002)];mode 1
    - Discriminates: rejects always using Normal column
  - (e) Null freeze callback:
    - Inputs: explicitly invoked freeze callback with null unit, state s=1, state count=185; supplied StayDeath returns 0 and SetState logs its direct call without a stored state for the absent unit.
    - Seeds: no draw.
    - Draws: none before the fatal assertion.
    - Trace: dead/StayDeath gate does not return; request state 1 off, then the null-unit guard is fatal before cleanup/mode/timer operations.
    - Outputs: pre-fatal direct-call log[state 1 off], then fatal assertion; no cleanup, mode or timer request before that assertion; no post-fatal stored state asserted.
    - Discriminates: rejects returning successfully for a null unit or putting cleanup/mode requests before the null guard.
  - runtime_observed: no

- **MMF-V28** Exceptional freeze targets (MMF-R08b; boundary: supplied-input stage)
  - (a) Null defender:
    - Inputs: null defender, freeze length=50, valid attacker/game; lookup dependency is the actual NRF-R12a operation.
    - Seeds: no draw.
    - Draws: none before the fault.
    - Trace: routing treats null as other type, then state-list lookup faults before state/list requests.
    - Outputs: fault; no state/list/timer requests before the fault.
    - Discriminates: rejects a successful null-target no-op return.
  - (b) Other type reaches the fatal tail:
    - Inputs: defender type 2, freeze length=50; state storage present, no freeze list, valid attacker/game and supplied list operations as in the preamble.
    - Seeds: defender {1,0}; no draw.
    - Draws: none before the fatal assertion.
    - Trace: default E=-50; no monster divisor; expiry=1050; list/state requests precede the monster-only assertion.
    - Outputs: pre-fatal log freezeNew(1050), then fatal assertion; no kind-2 request before the assertion. No post-fatal stored state is asserted.
    - Discriminates: rejects returning before the list requests or accepting a successful nonmonster freeze.
  - runtime_observed: no

### 8.2 Observation vectors

No observation has been performed. These controlled families require source/readback records,
including actual class/table inputs, requests, resulting rates and timing; a visual impression
alone does not establish any arithmetic assertion.

| Observation | Measurement | Procedure |
|---|---|---|
| **MMF-O01** | Movement and action extrema, dispatcher branches, fixed initialization, paired-item supplied result versus real selection, and path step | [MMF-O01](../research/spec-evidence/state-movement-modifiers-and-freeze.md#mmf-o01) |
| **MMF-O02** | Cold/freeze defenses, resistance, eligibility, durations, refresh, callbacks and AI requests versus actual lifecycle | [MMF-O02](../research/spec-evidence/state-movement-modifiers-and-freeze.md#mmf-o02) |
| **MMF-O03** | Same supplied cold effect through different movement/action clamps, seeded draws and resulting state requests | [MMF-O03](../research/spec-evidence/state-movement-modifiers-and-freeze.md#mmf-o03) |
| **MMF-O04** | Diminishing zero-denominator behavior | [MMF-O04](../research/spec-evidence/state-movement-modifiers-and-freeze.md#mmf-o04) |

## 9. Acceptance and dependencies

Partial static acceptance is the exact arithmetic, rate/path writes, seed results and request
logs of MMF-V01 through MMF-V28 against the verified rules in the Id map. Supplied category,
base, lookup, list and callback fixtures do not establish their original producers or actual
propagation. MMF-R04b/MMF-R06g/MMF-R13b/MMF-R13e remain gated; MMF-R14b is an outside-boundary
consumer question. Full movement/cold/freeze reproduction requires the applicable MMF-Q05/Q06/
Q08/Q11/Q12 and GDF-Q01 integrations, named actual data values where needed, and matching gameplay
observations for every verified behavioral rule (MMF-Q09). A partial pass does not establish parity.

**Depends on:** NRF-R01/NRF-R02 seed/draw, NRF-R05 multiplication/division, NRF-R06a/NRF-R06b
signed/unsigned arithmetic, NRF-R07 diminishing, NRF-R08b speed writes, NRF-R11a/NRF-R11b step and
scale, NRF-R12a/NRF-R12b/NRF-R12c state/stat reads and requests, NRF-R14a mode writes; GDF-R02/
GDF-R03a/GDF-R05a/GDF-R05b/GDF-R06a/GDF-R11a/GDF-R11b/GDF-R12/GDF-R14a/GDF-R15a names, fields and
derived values; CSA-R01/CSA-R02/CSA-R05/CSA-R06/CSA-R07/CSA-R08/CSA-R09/CSA-R13 cold/freeze local
requests and cold callback; DPO-R10 strip, DPO-R17/DPO-R26 placement; DMI-R01/DMI-R02/DMI-R05/
DMI-R06/DMI-R07/DMI-R08/DMI-R09 mitigation; HRC-R12a/HRC-R14 get-hit/block/knockback; STL-R01/
STL-R04/STL-R05/STL-R06a/STL-R06b/STL-R06c/STL-R07/STL-R11/STL-R12a event/list/state/expiry operations.
Every consumed integration retains its owner's gate. Prior creation helpers consumed in the
monster fixed-write slice are [monster scaling](monster-scaling-by-players-and-difficulty.md)
MSD-R02/MSD-R03; their complete output effects are not inferred here.

**Depended on by:** HRC-R14 for shared dispatcher precedence; CSA's cold/freeze application and
STL-R09/STL-R11 references to movement effects and callback policy; knockback's rate/speed
references. These reverse references are not prerequisites for supplied local arithmetic.

## 10. Related specs

- [combat-hit-recovery](combat-hit-recovery.md): owns get-hit, block and knockback rate branches.
- [combat-knockback](combat-knockback.md): owns displacement/interruption and full knockback execution.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): owns damage-length producers and combat placement.
- [combat-damage-mitigation](combat-damage-mitigation.md): owns effective resistance and per-entry mitigation.
- [combat-status-application](combat-status-application.md): owns cold/freeze application requests and cold callback.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): owns timers, lists, stored contributions and expiry.
- [monster-scaling-by-players-and-difficulty](monster-scaling-by-players-and-difficulty.md): owns surrounding creation and difficulty setup.
- [core-game-data-fields](core-game-data-fields.md): owns names, data fields and monster derived-base production.

## Appendix A. Provenance

Original Lord of Destruction Expansion 1.13c Windows x86 D2Game/D2Common artifacts are identified
in the [source manifest](../research/source-manifest.md). Compatible instruction readings dated
2026-10-12 and dependency records dated 2026-10-12 through 2026-10-15 retain their actual scope.
Targeted instruction, checked-reconstruction and original-listing corroboration on 2026-10-15
supports the corrections and supplied boundaries recorded in the
[companion](../research/spec-evidence/state-movement-modifiers-and-freeze.md). No gameplay was
observed. Named data records are reused from GDF/DMI; no arbitrary per-class table cell is claimed.
