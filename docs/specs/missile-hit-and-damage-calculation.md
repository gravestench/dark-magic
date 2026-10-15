# Missile Hit and Damage Calculation

**Description**

A missile checks whether a contacted unit can be hit, then prepares damage from the values the missile carries. Some missiles continue through a hit, while others end or run a special hit effect. The server passes the prepared result to the shared combat rules that determine damage and reactions.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MHD-R, MHD-Q, MHD-V, MHD-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/missile-hit-and-damage-calculation.md)
Open evidence gaps: [gap file](../plans/gaps/missile-hit-and-damage-calculation.md)

## 1. Scope

This contract defines unit contact, contact callbacks, the hit-time damage fill, the damage callback and direct combat hand-off, plus the separate optional no-target callback before removal. Creation-time values belong to [missile-damage-parameters](missile-damage-parameters.md). Contact discovery and removal belong to [missile-movement-and-collision](missile-movement-and-collision.md). Individual hit/damage callback bodies, stat lifecycle propagation, complete collision-map clearing and downstream combat effects retain their own boundaries and gates. Callback invocation accepts supplied operations; it does not reproduce every missile's special effect.

## 2. Trigger and ownership

D2Game's server missile step calls this operation for a contacted unit, expiry, or another supplied contact request. Within that call, the order is target gates, pierce decision, optional to-hit, cooldown/event callbacks, hit callback, damage fill, damage callback, direct combat hand-off, armor-write request, and stop decision. A miss takes its separate branch. The no-target callback/removal helper is a separate entry, not the damage stage. Original locations and caller-specific return handling are in the companion. This contract emits no direct client message; downstream combat owns messages and reactions. The wider server-frame cadence belongs to the movement owner.

## 3. Inputs and data

Integer additions, subtractions, products and shifts wrap at 32 bits unless an operation explicitly widens them. Numeric mask 1, 2 or 4 means testing `(value AND mask)!=0`; these numbers are mask values, not bit indices. Signed division truncates toward zero. Damage uses raw 8.8 quantities; lengths are signed frame counts. `Stat(unit,id)` is layer 0 through RLI-R02a/RLI-R02b, with existing ordered arrays and definitions supplied. Names consume GDF-R02/GDF-R03 and the companion's [named cells](../research/spec-evidence/missile-hit-and-damage-calculation.md#named-table-cells).

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| missile, target, game | Logical units and associations | Nullable associations; kind/class/id i32 | Missile kind 3; target may be absent |
| owner reference / registry | Owner-reference-valid condition, owner kind/id; registered unit chains | Boolean; i32 identities; ordered chains | Current creator lookup, distinct from hostility's later owner hops |
| missile contact selection | Missile contact-type value | u32 | 0 rejects all, 1 player, 2 monster, other values accept all |
| last collision | Missile's saved target kind/id and associated missile data | i32 identities | Same-target exclusion |
| prepared missile definition | Missiles.LastCollide, Explosion, Pierce; CollideKill, CollideFriend, NextHit, NextDelay, ToHit, AlwaysExplode, Size, HitClass, KnockBack, GetHit, SoftHit | Bits; other listed selectors u8 | Branches, cooldown, footprint and reaction requests |
| hit/damage selectors | Missiles.pSrvHitFunc, pSrvDmgFunc | Signed i16 | Indices into engine-defined handler tables |
| physical / attack | Stats 19 tohit, 21 mindamage, 22 maxdamage, 25 damagepercent | i32 | Hit bonus and physical fill |
| elemental pairs | 48/49 firemindam/firemaxdam, 52/53 magicmindam/magicmaxdam, 50/51 lightmindam/lightmaxdam, 54/55 coldmindam/coldmaxdam, 57/58 poisonmindam/poisonmaxdam, 316/317 burningmin/burningmax | i32 raw | Six fill pairs, in this order |
| masteries | 329 passive_fire_mastery, 357 passive_mag_mastery, 330 passive_ltng_mastery, 331 passive_cold_mastery, 332 passive_pois_mastery | i32 percent | Burn reuses fire mastery |
| lengths and drains | 56 coldlength, 59 poisonlength, 315 firelength, 66 stunlength; 60 lifedrainmindam, 62 manadrainmindam, 64 stamdrainmindam; 326 poison_count | i32 | Direct copies; poison-length divisor |
| physical target terms | 121 item_demondamage_percent, 122 item_undeaddamage_percent, 180 damage_vs_montype | i32; up to 128 ordered (u16 layer,i32 value) entries; missile list usesResolvedStats and resolvedStats under NRF section 3.8 | Add conditional bonuses to physical percent |
| target definition | MonStats.demon, lUndead, hUndead; signed monster-type index; supplied type-count and relation matrix | Booleans, i16 index, i32 count, bits | Demon/undead predicates and type bonus matching |
| critical/bypass | 141 item_deadlystrike; 103 skill_bypass_undead, 104 skill_bypass_demons, 106 skill_bypass_beasts | i32 nonzero tests | Deterministic double and record flags |
| pierce | Owner 166 skill_pierce / 156 item_pierce; missile 328 pierce_idx | i32 | Nonzero owner eligibility and remaining contact counter |
| armor / apply | Missile 120 item_damagetargetac, 327 damage_framerate; target 31 armorclass; missile's two source flags | i32; booleans | Armor request, reductionScale and applied flags |
| target state / clock | 54 uninterruptable, 86 justhit; missile-associated game's frame | Booleans, i32 frame | Contact exclusion and reaction/cooldown gates |
| record | NRF section 3.2 DamageRecord semantic fields | i32, u32 flags, u16 resultFlags, u8 reservedCondition | Shared mutable record |
| callbacks/dependencies | Hit, damage, event 0, hit-roll, block/avoidance, double, combat process/results, stat-write, cooldown allocation/attachment, footprint-clear, removal | Supplied operations and results | Log arguments/order; any declared mutations persist for following operations |
| seeds / footprint | Missile seed; owner/target seeds for dependency-owned draws; supplied room,x,y,size | Pairs of u32; i32 coordinates, u8 size | Local fill and missile knockback versus separate dependency budgets |

| Mapping | Origin | Binding |
|---|---|---|
| Handler table counts 71/31, null slots, result masks 1/2/4, contact-type selection, stat/state consumer ids, flags and floor -90 | Engine code | Fixed by original code; handlers selected by definition data |
| Named stats/states and prepared definition columns | Table data | Names from named cells / GDF descriptors; prepared values supplied |
| Poison-count division and type-relation matrix consumer | Engine code | Fixed consumer operations; matrix producer correspondence remains MHD-Q06 |

## 4. Rules (execution order)

### Contact

**MHD-R01** Entry and creator.

- **MHD-R01a** Entry [verified in 1.13c code (static)]. Resolve the supplied missile class's definition using GDF's bounds contract. An absent missile or invalid class yields no definition, followed by an unguarded definition read: this is a fault, not a graceful return. Initialize `hitResult=0` when Explosion is set, otherwise 2; initialize `stopMask=3`. Owner starts absent and is looked up only when the owner-reference-valid condition is set. No direct draw.
- **MHD-R01b** Registered creator lookup [verified in 1.13c code (static)]. With a present game and owner kind 0..4, select that kind's registered chain using `ownerId AND 127`; kind 5 selects its separate chain. Follow in order until the first id match, else return absent. No ownership recursion or kind recheck on matched entries. Invalid kind faults on the selected chain read; missing game is fatal. Chain population/lifecycle is MHD-Q05; supplied chains verify lookup independently. No direct draw.

**MHD-R02** Contact return [verified in 1.13c code (static)]. Return 1 means continue; return 2 requests removal by a caller that consumes the result. Early target rejection, callback result mask 4, ordinary nonstopping success and pierce-suppressed stopping all return 1. The miss branch returns 2 unless its invoked hit callback returns mask 4. The normal final branch returns 2 only after `(hitResult mask 1 OR force!=0) AND stopMask mask 1`; the footprint request can be skipped for an absent room without changing that return. Movement owns whether a particular caller propagates or overrides this result.

**MHD-R03** Target gates, only with a present target, in this order.

- **MHD-R03a** Cooldown [verified in 1.13c code (static)]. Nonzero NextHit and active state 86 return 1 immediately.
- **MHD-R03b** Last collision [verified in 1.13c code (static)]. For a missile of kind 3 with a valid definition, LastCollide set and associated missile data present, equal saved target kind/id returns 1. Otherwise, under those same storage guards, save the target kind/id before later filters. The isolated setter with absent target changes saved id to -1 and retains saved kind. No direct draw.
- **MHD-R03c** Kind filter [verified in 1.13c code (static)]. Contact selection 0 refuses every target; 1 accepts kind 0 only; 2 accepts kind 1 only; every other u32 value accepts all kinds. Refusal returns 1 after the last-collision write. Creation's producer is MCI's contract, not a local inferred mode conversion.
- **MHD-R03d** Relation [verified in 1.13c code (static)]. With a creator, consume PVP-R05a/PVP-R07's hostility predicate. A false result with CollideFriend=0 returns 1; a nonzero CollideFriend passes. With no creator, skip hostility. PVP's actual relation/tier production gates remain dependency integration gates.

**MHD-R04** Pierce decision [verified in 1.13c code (static)]. Require creator, target and Pierce. Read owner stat 166 first; read 156 only when that result is zero. If either is nonzero, read missile stat 328: nonzero `n` requests replacement `(328,layer 0,wrap32(n-1))` and sets stopMask=2; zero leaves stopMask=3. Negative n is nonzero. This local decision makes no draw. Actual counter persistence remains MHD-Q07; supplied mutable writer fixtures can establish subsequent reads.

**MHD-R05** Definition stopping [verified in 1.13c code (static)]. Nonzero CollideKill ORs mask 1 into hitResult before to-hit and hit callbacks. A later hit callback replaces the complete hitResult, so it can clear this bit. No direct draw.

**MHD-R06** Hit/miss branch [verified in 1.13c code (static)]. With a target and nonzero ToHit, read missile stat 19 and request HIT-R01 through HIT-R08's hit roll with attacker=creator, defender=target, bonus=that stat and isMissile=1. An absent attacker or target returns miss before any hit-roll draw. With a nonzero hit result continue normally. On miss, request R08's cooldown then event 0 even independently of stopMask; only nonzero AlwaysExplode enables the valid-index hit callback. Test only `(callbackResult AND 4)!=0`: true returns 1, otherwise return 2. No fill, damage callback, direct hand-off or footprint request occurs on this miss path. Hit-roll hooks and draws belong to SBA.

**MHD-R07** No target [verified in 1.13c code (static)]. Without a target, nonzero force or AlwaysExplode reaches the normal side-effect/callback stage. Otherwise skip directly to R13. No target means no fill or damage hand-off regardless of hitResult mask 2.

**MHD-R08** Cooldown and contact event [verified in 1.13c code (static)]. On the normal path, stopMask mask 2 enables these ordered operations; R06 also invokes them on a miss. First request cooldown only when NextHit is nonzero and target is present. Its requested expiry is `wrap32`(missile-associated game.frame+u8 NextDelay). Allocate a type-2 list with target owner kind/id and that expiry under STL-R06a. Allocation failure ends cooldown preparation before any scheduling/tag/attachment/state request; it does not stop the caller's following event dispatch. On success: schedule target event type 12 at requested expiry (STL-R01 may store a later frame); set list state 86; install callback that requests state 86 off; request full attachment; request active state 86 on. Then dispatch event 0 on the target with other=missile and absent damage record, under IEP-R07, ignoring its final result. Absent target has no event list. Actual attach, expiry/removal and update effects remain MHD-Q08. No direct draw in request placement; callbacks own theirs.

**MHD-R09** Hit callback selection.

- **MHD-R09a** Result use [verified in 1.13c code (static)]. Only signed index `0<h<71` is eligible. On the normal path, invoke after R08 and replace complete hitResult by the returned i32. Mask 4 returns 1 immediately; otherwise mask 2 permits R11 with a target, and mask 1 participates in R13. On a miss only R06's mask-4 test applies. No direct draw outside the callback.
- **MHD-R09b** Table/null slots [verified in 1.13c code (static)]. Invoke the selected operation with game, missile and nullable target. Slot 0 and slots 30,34,41,42,46,49,60..70 are null; index 0 is skipped, eligible null slots attempt a null call and fault. There is no null fallback. Actual row-to-slot bindings and individual callback bodies remain MHD-Q01's outside-boundary questions.

**MHD-R10** Fill the hit-time damage record.

- **MHD-R10a** Reset/physical [verified in 1.13c code (static)]. Zero every semantic field of the record. Read missile max stat 22; nonpositive gives physical 0. Otherwise read min stat 21; nonpositive gives 0. With both positive, swap if min>max, then physical=`wrap32`(min+NRF-R02 Rand(missile,`wrap32`(max-min))). Equal endpoints do not draw.
- **MHD-R10b** Six elemental rolls [verified in 1.13c code (static)]. In order fire(48,49,329), magic(52,53,357), lightning(50,51,330), cold(54,55,331), poison(57,58,332), burn(316,317,329), where each tuple is (min stat,max stat,pct stat). Each roller first reads max, then min, returning 0 without a draw if either read is nonpositive; swap the original endpoints if needed. If pct!=0, independently add NRF-R05 mulDiv(pct,endpoint,100) to each endpoint, with that argument order and 32-bit wrap. Do not swap or guard again. Return `wrap32`(min+Rand(missile,`wrap32`(max-min))). A nonpositive wrapped range skips the draw; adjusted values need not be positive or lie within the original range.
- **MHD-R10c** Placement [verified in 1.13c code (static)]. Store physical before calling the fire roller; fire precedes magic, then lightning, cold, poison and burn. Each active positive range advances the same missile seed once. These seven ranges contribute 0..7 draws to this fill; this is not the whole contact/callback budget.
- **MHD-R10d** Nonrandom fields [verified in 1.13c code (static)]. Copy coldLength=stat 56 and poisonLength=stat 59; when stat 326>1 divide poisonLength by it, signed. Copy manaBurn=stat 62, lifeLeech=stat 60, staminaBurn=stat 64, burnLength=stat 315, stunLength=stat 66. No draw.
- **MHD-R10e** Physical target adjustment [verified in 1.13c code (static)]. With target present start pct=stat 25, add stat 121 for a valid demon definition, then stat 122 for either undead definition bit. For a valid monster definition with positive signed monster-type index, enumerate up to 128 stat 180 entries from the missile's resolvedStats array only when its list exists and usesResolvedStats is true; otherwise enumerate none. Copy matching stat entries in stored order without RLI-R02b's minimum adjustment. For each layer and index both strictly between 0 and supplied typeCount with relation(targetTypeIndex,entryLayer) true, add its value with 32-bit wrap. This is a relation test, not numeric layer equality. Floor final signed pct at -90. Let prod=`wrap32`(physical*pct); add trunc(prod/100) to physical, wrapping. This uses the wrapped product, not NRF-R05's large-operand branches. Without a target skip the entire adjustment. The relation-matrix producer is MHD-Q06; supplied matrix/entries are independently executable.
- **MHD-R10f** Critical/bypass [verified in 1.13c code (static)]. Nonzero stat 141 doubles physical with a wrapping shift and ORs resultFlags critical 8192, without a new chance roll. Nonzero stats 103/104/106 OR flags 256/512/1024 respectively. Creation-time critical selection is MDP-R08, not an additional hit-time draw.

**MHD-R11** Damage callback and hand-off.

- **MHD-R11a** Ordering [verified in 1.13c code (static)]. On the normal path with target and hitResult mask 2: R10 fill, optional R11b damage callback, then R12 direct apply and R11c armor stage. Callback record mutations persist. Ignore its return value. Then R13 decides the final contact return.
- **MHD-R11b** Damage selection [verified in 1.13c code (static)]. Only signed `0<d<31` invokes callback(game,missile,target,shared record). Slots 0 and 15..30 are null: 0 is skipped, eligible null slots fault before hand-off. Invalid indices skip. Callback bodies are supplied under MHD-Q01, not invented defaults.
- **MHD-R11c** Armor-write request [verified in 1.13c code (static)]. After R12 returns, including its no-owner early return, only a present monster target outside hireling classes 271,338,359,560,561 continues. Nonzero missile stat 120 computes a=`wrap32`(Stat(target,31)+stat 120), then floors signed a<0 to 0 and requests replacement of target stat 31/layer 0 by a. No local draw. Actual persistence/propagation is MHD-Q07.

**MHD-R12** Direct apply [verified in 1.13c code (static)]. Require present missile kind 3, valid owner reference, found creator and valid definition; otherwise return without changing the incoming record or invoking apply/results. After success:

1. OR hit 1, then consume HRC-R01b/SBA-R11 and KNB-R02a/KNB-R02b's state 54/get-hit/soft-hit/missile-knockback preparation.
2. Request SBA's defender block/avoidance with attacker=creator, defender=target, blockEnable=(physical!=0), missile argument 1 and game expansion passed unchanged; consume HRC-R01c/SBA-R06's rewrite and result mapping.
3. Set hitClass=u8 HitClass OR (old hitClass AND 240); if the resulting low byte has any of 240, set reservedCondition=1, otherwise retain that field.
4. Request DPO-R22's monster-double operation with attacker=creator and defender=target; it can mutate this same record and draw from the creator.
5. Source flags A/B OR flags 32/128 respectively. Set reductionScale=missile stat 327.
6. If hit 1 remains, request DPO-R08d/DPO-R13's combat process with attacker=creator, defender=target, supplied record, processing flag 1. Then request the combat-result consumer with that creator/target/record regardless of the current hit flag; HRC owns reaction execution.

This establishes ordered calls and shared-record mutations under supplied operations, not actual life loss, status persistence, reaction completion or full owner registry integration. Downstream operations can change flags, fields and seeds; test them under their owners' gates.

**MHD-R13** Final stop decision [verified in 1.13c code (static)]. On the normal path, if neither hitResult mask 1 nor nonzero force, return 1. If stopMask mask 1 is clear, return 1. Otherwise, with a room request footprint-clear(room,x,y,Size,mask 64); with no room skip that request. Return 2 in both cases. The request's map writes remain MHD-Q03. This stage does not itself remove the missile.

**MHD-R14** Separate no-target callback/removal [verified in 1.13c code (static)]. At this separate entry, when doCall is nonzero, fetch the definition and invoke eligible hit callback with target absent, using R09b's table. Ignore the entire callback result, including mask 4. Then request removal unconditionally. With doCall=0, skip even definition access and request removal. With doCall nonzero and missing/invalid definition, fault before removal. Real caller reachability is MHD-Q02; supplied entry assertions stand independently of that question.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MHD-R01a | Entry and initial result | verified in 1.13c code (static) | none | MHD-V01, MHD-V05 |
| MHD-R01b | Creator lookup | verified in 1.13c code (static) | MHD-Q05 registry production only | MHD-V16 |
| MHD-R02 | Continue/removal return | verified in 1.13c code (static) | none | MHD-V01, MHD-V03, MHD-V04, MHD-V13 |
| MHD-R03a | NextHit exclusion | verified in 1.13c code (static) | none | MHD-V02, MHD-V06 |
| MHD-R03b | Last target read/write | verified in 1.13c code (static) | none | MHD-V02 |
| MHD-R03c | Kind filter | verified in 1.13c code (static) | none | MHD-V16 |
| MHD-R03d | Hostility consumption | verified in 1.13c code (static) | PVP relation/tier producer integration only | MHD-V02 |
| MHD-R04 | Pierce decision/write request | verified in 1.13c code (static) | MHD-Q07 actual stat persistence only | MHD-V03 |
| MHD-R05 | CollideKill result bit | verified in 1.13c code (static) | none | MHD-V03, MHD-V13 |
| MHD-R06 | Hit/miss branch | verified in 1.13c code (static) | SBA producer/hook integration only | MHD-V04 |
| MHD-R07 | No-target routing | verified in 1.13c code (static) | none | MHD-V01, MHD-V05 |
| MHD-R08 | Cooldown/event requests | verified in 1.13c code (static) | MHD-Q08 actual attach/expiry/update effects only | MHD-V04, MHD-V06 |
| MHD-R09a | Hit callback gate/result | verified in 1.13c code (static) | MHD-Q01 actual callback bodies only | MHD-V05, MHD-V13 |
| MHD-R09b | Hit table/null call | verified in 1.13c code (static) | MHD-Q01 actual row bindings only | MHD-V13 |
| MHD-R10a | Zero record/physical | verified in 1.13c code (static) | none | MHD-V12 |
| MHD-R10b | Element rolls | verified in 1.13c code (static) | none | MHD-V07, MHD-V12 |
| MHD-R10c | Seven-slot draw order | verified in 1.13c code (static) | none | MHD-V07, MHD-V12 |
| MHD-R10d | Copies/division | verified in 1.13c code (static) | none | MHD-V08 |
| MHD-R10e | Physical target percent | verified in 1.13c code (static) | MHD-Q06 matrix producer only | MHD-V12 |
| MHD-R10f | Critical/bypass | verified in 1.13c code (static) | none | MHD-V12 |
| MHD-R11a | Fill/callback/apply order | verified in 1.13c code (static) | dependency effects only | MHD-V09, MHD-V14 |
| MHD-R11b | Damage table/null call | verified in 1.13c code (static) | MHD-Q01 actual callback bodies only | MHD-V09 |
| MHD-R11c | Armor request | verified in 1.13c code (static) | MHD-Q07 actual persistence only | MHD-V10, MHD-V11 |
| MHD-R12 | Direct apply order | verified in 1.13c code (static) | MHD-Q05 creator registry; downstream dependency effects only | MHD-V11, MHD-V14 |
| MHD-R13 | Final stop/request | verified in 1.13c code (static) | MHD-Q03 actual map clearing only | MHD-V01, MHD-V03, MHD-V13 |
| MHD-R14 | Separate no-target remove entry | verified in 1.13c code (static) | MHD-Q02 real callers only | MHD-V15 |

## 5. Outputs and side effects

| Output / request | Rule |
|---|---|
| Contact continue 1/removal 2, or fault | MHD-R01/MHD-R02/MHD-R09/MHD-R11b/MHD-R13 |
| Saved last-target kind/id | MHD-R03b |
| Pierce stat 328 replacement request | MHD-R04 |
| Cooldown allocation; event 12, list state/callback, full attachment, active state requests; event 0 dispatch | MHD-R08 |
| Hit/damage callback invocation and supplied mutations | MHD-R09/MHD-R11/MHD-R14 |
| Zeroed and filled DamageRecord; physical/elements/lengths/drains/critical/bypass | MHD-R10 |
| Missile seed updates from fill and knockback; creator/target dependency seed updates | MHD-R10/MHD-R12 |
| Footprint-clear and separate removal requests | MHD-R13/MHD-R14 |
| Armor stat 31 replacement request | MHD-R11c |
| Prepared resultFlags, hitClass, reservedCondition, flags and reductionScale; process/results requests | MHD-R12 |

Actual list/stat/map/registry lifecycle effects retain the named integration gates. Combat messages, resource writes and credited death belong to DPO/PVP/CSA/HRC.

## 6. Branches and edge cases

A missing creator bypasses the contact hostility gate, disables pierce, makes ToHit miss without a draw, and makes direct apply return; damage filling and the later armor request can still occur when ToHit is disabled. Explosion starts with no damage bit. Equal positive damage endpoints do not draw; initial nonpositive endpoints return zero before mastery. Negative percentages can invert adjusted ranges, and nonpositive wrapped ranges do not draw. Contact callbacks replace the result; force does not override an early rejection or mask 4 abort. No local difficulty or classic/expansion branch changes the fill; defender block consumes the supplied expansion value. A valid class index with a loaded table selects its definition; it cannot simultaneously be modeled as a missing-definition fixture.

## 7. Open questions

| Id | Status | Evidence that would settle it |
|---|---|---|
| **MHD-Q01** | outside-boundary | Named data rows selecting null slots and complete hit/damage callback bodies. Full table image proves which slots are null, not which prepared rows reach them. No local gate on supplied callback dispatch. |
| **MHD-Q02** | outside-boundary | Establish actual callers of the separate no-target callback/removal entry; its supplied-entry sequence is verified. |
| **MHD-Q03** | integration | Complete footprint shape/room map writes behind the clear request; no gate on its arguments/order or return 2. |
| **MHD-Q04** | observation | Reproduce every verified behavioral rule's applicable assertions on an owned 1.13c game, including O01..O04; no gameplay observation occurred for this contract. |
| **MHD-Q05** | outside-boundary | Full registered-unit creation/removal/chain lifecycle; no gate on supplied-chain lookup or supplied creator consumers. |
| **MHD-Q06** | outside-boundary | Establish original prepared monster-type relation matrix and entries' producer correspondence; no gate on supplied relation/count physical arithmetic. |
| **MHD-Q07** | integration | Complete replacing-stat lifecycle propagation for pierce and armor writes under RLI-Q03/STL-R06c. Request arithmetic is verified. |
| **MHD-Q08** | integration | Complete cooldown list attachment, due-list freeing/callback and active/update propagation under STL-R05/STL-R06b/STL-R07. Ordered requests are verified. |

## 8. Test vectors

### 8.1 Executable vectors

Defaults apply independently to each case: missile kind 3, valid supplied definition/class 0/count 1, associated missile data, owner-reference clear, absent target, no saved last target, contact selection 3, no flags or callback selectors, all stats absent/zero except listed overrides, force 0, no room, frame 1000. A present target is monster kind 1/id 7/class 5 unless overridden. Supplied stat reads return the named values under NRF's existing-array contract (definitions valid, no minimum clamp). Supplied writer logs and persists replacements within a case; actual writer integration is gated. Allocator success supplies an empty list; failure supplies absent. Supplied operations log their arguments and explicit mutations persist; output sequences assert ordered subsequences unless marked complete; ordinary definition lookup and stat/state reads are not logged unless named; all unspecified dependency operations return 0 without mutations/draws. An explicitly supplied hit result or block code is a stage input, not a random result. Unspecified seeds are unchanged; “no draw” counts only the stated stage and its supplied operations. Every vector has runtime_observed:no.

- **MHD-V01** No target/no stop (MHD-R01a/MHD-R02/MHD-R07/MHD-R13; boundary: supplied-input stage)
  - (a)
    - Inputs: defaults: no target, force 0, Explosion=0, AlwaysExplode=0, CollideKill=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: hitResult=2, stopMask=3; no-target branch skips callbacks; no stop bit/force.
    - Outputs: return 1; no callbacks, fill, footprint or removal request; record absent and seeds unchanged.
  - (b)
    - Inputs: contact entry only, missile present with class -1 and loaded definition count 1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: class is invalid; definition resolution returns absent; initial unguarded definition read faults.
    - Outputs: fault before callbacks, fill or contact return.
  - (c)
    - Inputs: contact entry only, missile absent and definition count 1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: absent missile produces no definition, then the same unguarded read faults.
    - Outputs: fault before callbacks, fill or contact return.
  - Discriminates: final 1 rejects returning 2 merely because contact ran.
  - runtime_observed: no

- **MHD-V02** Ordered exclusions (MHD-R03a/MHD-R03b/MHD-R03d; boundary: supplied-input stage)
  - (a)
    - Inputs: target monster id 7, NextHit=1, active 86=true, LastCollide=1, saved kind 0/id 9.
    - Seeds: no draw.
    - Draws: none.
    - Trace: cooldown exclusion precedes saved-target test/write.
    - Outputs: return 1, saved kind 0/id 9 unchanged, no later calls.
  - (b)
    - Inputs: target monster id 7, NextHit=1, active 86=false, LastCollide=1, saved kind 1/id 7.
    - Seeds: no draw.
    - Draws: none.
    - Trace: same-target predicate passes exclusion.
    - Outputs: return 1, saved kind 1/id 7 unchanged, no later calls.
  - (c)
    - Inputs: target monster id 8, LastCollide=1, saved kind 1/id 7; owner reference valid and supplied creator player C; hostility supplied false, CollideFriend=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: new kind 1/id 8 saved before false hostility refuses.
    - Outputs: saved kind 1/id 8, return 1; hostility called with creator C/target 8; no hit callback or fill.
  - (d)
    - Inputs: isolated last-collision setter, valid missile, LastCollide=1, associated data, saved kind 1/id 7; target absent.
    - Seeds: no draw.
    - Draws: none.
    - Trace: the setter's absent-target branch replaces the saved id only.
    - Outputs: saved kind 1/id -1; no other write.
  - (e)
    - Inputs: target monster id 7; owner reference valid, supplied creator player C, supplied hostility false; CollideFriend=1, ToHit=0. Supplied direct hand-off logs without changes.
    - Seeds: no draw.
    - Draws: none.
    - Trace: false hostility is permitted by CollideFriend; normal event and damage stage run with hitResult=2.
    - Outputs: ordered hostility(C,target 7), event 0, fill with all fields 0, hand-off; final return 1; no footprint request.
  - Discriminates: case c rejects rolling back or postponing the saved-target write until relation acceptance; case d retains kind while clearing id, and e rejects ignoring CollideFriend.
  - runtime_observed: no

- **MHD-V03** Pierce and final return (MHD-R04/MHD-R05/MHD-R13/MHD-R02; boundary: supplied-input stage)
  - (a)
    - Inputs: target monster id 7; creator C present, supplied hostility true; owner 166=1; Pierce=1, CollideKill=1; missile 328=2; no callbacks, ToHit=0; room R,x=4,y=5,Size=2. Supplied writer persists; direct apply supplied as no-op for this contact-routing boundary.
    - Seeds: no draw.
    - Draws: none.
    - Trace: request 328=1; stopMask=2; hitResult=3; damage hand-off reached; final stopMask lacks mask 1.
    - Outputs: missile 328=1 in fixture; log replace 328/1 then event 0 then fill then hand-off; return 1; no footprint request; owner 156 not read.
  - (b)
    - Inputs: case a inputs independently, except missile 328=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: no replacement; stopMask=3; hitResult=3 reaches final stopping request.
    - Outputs: missile 328=0; event 0 then fill then hand-off then footprint-clear(R,4,5,2,64); return 2.
  - (c)
    - Inputs: target monster id 7; owner reference valid, creator C, supplied hostility true; owner stats 166=0,156=1; Pierce=1,CollideKill=1; missile stat 328=-1; supplied writer persists and direct apply logs without changes.
    - Seeds: no draw.
    - Draws: none.
    - Trace: zero stat 166 enables stat 156 read; nonzero -1 counter requests -2 and sets stopMask=2; stopping is suppressed.
    - Outputs: stat-read order owner stat 166 then owner stat 156 then missile stat 328; replacement(328,layer 0,-2), event 0, fill, hand-off; fixture missile stat 328=-2; final return 1; no footprint request.
  - Discriminates: case a's final 1 rejects unconditional 2 after damage and rejects clearing the footprint before pierce suppression; c rejects testing counter>0 instead of counter!=0 and skipping the item-pierce fallback.
  - runtime_observed: no

- **MHD-V04** Miss routing and absent attacker (MHD-R06/MHD-R08/MHD-R02; boundary: supplied-input stage)
  - (a)
    - Inputs: target monster id 7; owner absent; ToHit=1, missile 19=50; NextHit=1, NextDelay=10; allocator failure; event 0 supplied returns 99; AlwaysExplode=0, CollideKill=1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: hit roll called with absent attacker, defender 7, bonus 50,isMissile=1 ->0 before any draw; cooldown allocation fails; event 0 runs next, result ignored.
    - Outputs: return 2; ordered log hit-roll, allocate(type 2,owner 1/7,expiry 1010), event 0(target 7,other missile,record absent); no timer/state/attach request, no fill/callback/footprint.
  - (b)
    - Inputs: case a independently with AlwaysExplode=1, hit selector 1 and supplied hit callback returns 4.
    - Seeds: no draw.
    - Draws: none.
    - Trace: same miss/failed allocation/event; callback sees target 7 and returns mask 4.
    - Outputs: same preliminary log then hit 1; final return 1; no fill or footprint.
  - (c)
    - Inputs: target monster id 7; owner reference valid and supplied creator player C; supplied hostility true; ToHit=1, missile stat 19=50, NextHit=0, CollideKill=0. Supplied hit-roll returns 1 without draws; direct apply logs without changes.
    - Seeds: no draw.
    - Draws: none in this supplied-result stage.
    - Trace: successful supplied hit-roll leaves the miss branch, then event and fill/hand-off run; no stop requested.
    - Outputs: ordered hit-roll(C,target 7,bonus 50,isMissile=1), event 0, fill with all fields 0, hand-off; return 1. No cooldown allocation, footprint or removal request.
  - Discriminates: callback 4 changes the final result to 1; allocation failure still reaches event 0; no-owner hit roll does not borrow the missile seed.
  - runtime_observed: no

- **MHD-V05** No-target force and selector bounds (MHD-R01a/MHD-R07/MHD-R09a; boundary: supplied-input stage)
  - (a)
    - Inputs: target absent, force 1, hit selector 0, room absent, Explosion=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: forced callback stage; selector 0 skipped; hitResult=2; force reaches stop decision with stopMask=3.
    - Outputs: return 2, no hit callback, fill or footprint request.
  - (b)
    - Inputs: case a independently with hit selector 71, Explosion=1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: initial result 0, out-of-range selector skipped; force still selects stop.
    - Outputs: return 2, no hit callback or damage stage.
  - (c)
    - Inputs: target absent, force 0, AlwaysExplode=1, NextHit=1, NextDelay=10, hit selector 1; supplied hit callback returns 0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: AlwaysExplode enters normal callbacks; absent target skips cooldown allocation and the event dispatcher invokes no registered callback; hit callback replaces result by 0.
    - Outputs: event-0 dispatch request with owner absent,other missile,record absent returns 0; then hit callback 1(game,missile,target absent); final return 1. No cooldown allocation, fill or footprint request.
  - Discriminates: force 1 reaches stop even with result 0; selector 71 is skipped rather than invoked; case c reaches the hit callback via AlwaysExplode with no target or cooldown allocation.
  - runtime_observed: no

- **MHD-V06** Cooldown request ordering (MHD-R08/MHD-R03a; boundary: supplied-input stage)
  - (a)
    - Inputs: target monster id 7, NextHit=1, NextDelay=10, frame 1000; allocator succeeds with empty list L; supplied attachment/state operations persist membership 86; event 0 logs and returns 4; hit selector 1 callback observes membership 86 then returns 0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: requested expiry 1010 is future; schedule then tag/callback/attach/state; event precedes hit 1; hit result replaced 0 gives nonstopping exit.
    - Outputs: ordered requests allocate(type 2,owner 1/7,expiry 1010), schedule(target 7,type 12,frame 1010), listState(L,86), callback(L,state-off), attach(target 7,L,full 1), state(target 7,86,on 1), event 0(target 7,other missile,record absent), hit 1; return 1. List expiry 1010 and state 86=true in fixture. Event return 4 does not abort hit 1.
  - (b)
    - Inputs: fresh contact at frame 1005, target 7 with active 86=true, NextHit=1, NextDelay=10; existing supplied list expiry 1010.
    - Seeds: no draw.
    - Draws: none.
    - Trace: R03a rejects before allocation/event.
    - Outputs: return 1; no new request; supplied expiry/membership unchanged.
  - (c)
    - Inputs: cooldown stage only, target monster id 7, NextHit=1, NextDelay=0, frame 1000, successful empty list L.
    - Seeds: no draw.
    - Draws: none.
    - Trace: requested expiry 1000, scheduler clamps due frame to 1001; list retains requested expiry.
    - Outputs: list expiry 1000; scheduled event 12 at 1001; same listState/callback/attach/state request order as a.
  - Discriminates: callback sees already requested state 86; list expiry 1000 differs from scheduler 1001 for zero delay. Actual expiry freeing is gated.
  - runtime_observed: no

- **MHD-V07** Magic/cold seed order (MHD-R10b/MHD-R10c; boundary: isolated rule)
  - (a)
    - Inputs: fill only; target absent; stats 52=256,53=512,54=100,55=300; all other pairs/masteries zero.
    - Seeds: missile {1,0x29A}.
    - Draws: missile/256/95, then missile/200/131.
    - Trace: first lo 0x6AC6935F gives magic 351; second lo 0x2F2ED81B gives cold 231.
    - Outputs: magic 351,cold 231, all other damage/length/drain/flags 0; seed {0x2F2ED81B,0x2C890AFD}.
  - (b)
    - Inputs: case a independently, except 357=50.
    - Seeds: missile {1,0x29A}.
    - Draws: missile/384/351, then missile/200/131.
    - Trace: magic endpoints 384/768,384+351=735; cold 231.
    - Outputs: magic 735,cold 231, other fields 0; seed {0x2F2ED81B,0x2C890AFD}.
  - Discriminates: magic 735 rejects scaling the already rolled 351; ordered seed outputs reject cold-first evaluation.
  - runtime_observed: no

- **MHD-V08** Poison divide/copies (MHD-R10d; boundary: isolated rule)
  - (a)
    - Inputs: fill only, stats 59=200,326=3,56=12,60=20,62=30,64=40,315=50,66=60.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 200/3 truncates 66; other terms copied.
    - Outputs: poisonLength=66,coldLength=12,lifeLeech=20,manaBurn=30,staminaBurn=40,burnLength=50,stunLength=60; all damage/flags 0.
  - (b)
    - Inputs: fill only, stats 59=200,326=1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: divisor not greater than 1.
    - Outputs: poisonLength=200, other fields 0.
  - (c)
    - Inputs: fill only, stats 59=-7,326=2.
    - Seeds: no draw.
    - Draws: none.
    - Trace: -7/2 truncates toward 0.
    - Outputs: poisonLength -3, other fields 0.
  - Discriminates: case c rejects floor division producing -4.
  - runtime_observed: no

- **MHD-V09** Damage callback bounds and mutable record (MHD-R11a/MHD-R11b; boundary: supplied-input stage starting at damage fill after contact callbacks)
  - (a)
    - Inputs: target monster 7, damage selector 0, physical pair 21=22=100; supplied direct hand-off logs record and changes nothing.
    - Seeds: no draw.
    - Draws: none.
    - Trace: fill 100, selector skipped, hand-off.
    - Outputs: log fill then hand-off(physical 100), return 1, no damage callback.
  - (b)
    - Inputs: case a independently with damage selector 15.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 15 is positive and below 31 but null.
    - Outputs: physical 100 filled, null-call fault before hand-off/final return.
  - (c)
    - Inputs: case a independently with damage selector 31.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 31 out of range.
    - Outputs: same as a, no callback.
  - (d)
    - Inputs: case a independently with selector 1, supplied callback sets physical 123 and returns 4.
    - Seeds: no draw.
    - Draws: none.
    - Trace: callback receives filled 100 and changes shared physical 123; return ignored.
    - Outputs: log fill 100, damage 1(game,missile,target 7,record 100), hand-off(physical 123); final 1.
  - Discriminates: case d rejects testing damage callback return 4 as abort and rejects passing a stale record copy; case b rejects null fallback.
  - runtime_observed: no

- **MHD-V10** Armor request (MHD-R11c; boundary: supplied-input stage)
  - (a)
    - Inputs: armor stage after direct apply; target monster class 5, stat 31=100; missile 120=-30; supplied writer persists.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 100-30=70, nonnegative.
    - Outputs: replace(target,31,layer 0,value 70); fixture armor 70.
  - (b)
    - Inputs: case a independently except 120=-150.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 100-150=-50 floored 0.
    - Outputs: replace 31/0; fixture armor 0.
  - (c)
    - Inputs: fresh target monster class 271,armor 100,missile 120=-30.
    - Seeds: no draw.
    - Draws: none.
    - Trace: hireling skips.
    - Outputs: no write; armor 100.
  - (d)
    - Inputs: fresh player target,armor 100,missile 120=-30.
    - Seeds: no draw.
    - Draws: none.
    - Trace: non-monster skips.
    - Outputs: no write; armor 100.
  - Discriminates: case b rejects negative armor persistence; case c rejects applying to hirelings. Persisted effects outside the supplied writer are gated.
  - runtime_observed: no

- **MHD-V11** Missing owner after fill (MHD-R12/MHD-R11c; boundary: supplied-input stage)
  - Inputs: contact target monster class 5,armor 100; owner reference valid, registered chain empty at contact and direct apply; ToHit=0,damage selector 0; stats 21=22=100,120=-30,KnockBack=100; supplied stat writer persists.
  - Seeds: no draw.
  - Draws: none.
  - Trace: fill physical 100; direct apply finds no owner and returns before hit preset/knockback/results; armor stage still computes 70.
  - Outputs: record physical 100,resultFlags=0; no knockback, combat-process or combat-result call; replace armor 70, final 1.
  - Discriminates: armor request after no-owner apply rejects returning from the entire hand-off on owner loss. No whole-path immunity beyond these calls is implied.
  - runtime_observed: no

- **MHD-V12** Fill boundaries, relation and critical (MHD-R10a/MHD-R10b/MHD-R10d/MHD-R10e/MHD-R10f/MHD-R10c; boundary: supplied-input stage)
  - (a)
    - Inputs: fill only, target absent, every pair 21/22,48/49,52/53,50/51,54/55,57/58,316/317=1/101; incoming record all semantic fields 99; masteries/other stats 0.
    - Seeds: missile {1,0}.
    - Draws: missile/100/85,9,72,62,66,6,47 in physical/fire/magic/lightning/cold/poison/burn order.
    - Trace: each pair adds 1 to its draw; incoming fields zeroed before fill.
    - Outputs: physical 86,fire 10,magic 73,lightning 63,cold 67,poison 7,burnMagnitude=48; all other fields 0; seed {0xEEA0F793,0x5A0CC847}.
  - (b)
    - Inputs: fill only, physical 21=22=100; target valid monster with demon and undead true, monsterTypeIndex=2, typeCount=5; missile 25=-200,121=20,122=30; missile list usesResolvedStats=true, resolvedStats stat 180 entries(layer 1,value 10),(layer 2,value 90); relation(2,1)=true,relation(2,2)=false; missile 141=1,103=104=106=1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: pct=-200+20+30+10=-140 -> -90; physical 100-90=10, critical doubles 20.
    - Outputs: physical 20,resultFlags=8192,flags 1792; other fields 0.
  - (c)
    - Inputs: fill only, target absent, fire 48=5,49=9,329=-200; other pairs 0.
    - Seeds: missile {1,0}.
    - Draws: none.
    - Trace: adjusted min 5-10=-5,max 9-18=-9,range -4; Rand returns 0.
    - Outputs: fire -5,other fields 0; seed {1,0} unchanged.
  - (d)
    - Inputs: fill only, target present with no demon/undead/type bonus; physical 21=22=1073741824,25=4.
    - Seeds: no draw.
    - Draws: none.
    - Trace: prod=low32(1073741824*4)=0, adjustment 0.
    - Outputs: physical 1073741824,other fields 0.
  - (e)
    - Inputs: fill only, target absent; physical stats 21=12,22=8; fire stats 48=9,49=5; all other stats 0.
    - Seeds: missile {1,0}.
    - Draws: missile/4/1 for physical, then missile/4/1 for fire.
    - Trace: original physical endpoints swap to 8/12 ->9; fire endpoints swap to 5/9 ->6.
    - Outputs: physical 9,fire 6,other fields 0; seed {0x669A3799,0x2C8909E7}.
  - (f)
    - Inputs: fill only, target absent; physical stats 21=0,22=10; fire stats 48=0,49=9; all other stats 0.
    - Seeds: missile {1,0}.
    - Draws: none.
    - Trace: each positive max is read, then the zero min refuses the range before any draw.
    - Outputs: every record field 0; seed {1,0} unchanged.
  - (g)
    - Inputs: fill only; physical stats 21=22=100; target monster with no demon/undead bonus, monsterTypeIndex=2,typeCount=3; missile list usesResolvedStats=true; resolvedStats stat 180 entries(layer 0,value 90),(layer 1,value 10),(layer 3,value 100); relation(2,1)=true, other valid relation cells false.
    - Seeds: no draw.
    - Draws: none.
    - Trace: layer 0 and layer 3 fail strict positive/count bounds; layer 1 contributes 10; physical=100+100*10/100.
    - Outputs: physical 110,other record fields 0.
  - (h)
    - Inputs: fill only; physical stats 21=22=100; target monster with no demon/undead bonus, monsterTypeIndex=1,typeCount=130; missile list usesResolvedStats=true; resolvedStats contains exactly 129 stat 180 entries: entry i has layer i,value 1 for every integer i from 1 through 129; relation(1,i)=true for those 129 layers.
    - Seeds: no draw.
    - Draws: none.
    - Trace: enumerate the first 128 entries only; pct=128; physical=100+128.
    - Outputs: physical 228,other record fields 0; no contribution from entry 129.
  - (i)
    - Inputs: same supplied target/type/entry values as case h, independently, except missile list usesResolvedStats=false; ownStats and resolvedStats both contain stats 21/22 with value 100 plus those same 129 stat 180 entries.
    - Seeds: no draw.
    - Draws: none.
    - Trace: enumeration wrapper refuses a non-resolved list; pct=0 despite entries in both supplied arrays.
    - Outputs: physical 100,other record fields 0.
  - Discriminates: case a rejects element reordering and missed zeroing; b rejects equality matching, floor-before-bonus and new critical RNG; c rejects post-percent swapping; d rejects using widened product or NRF-R05's large-value branch; e/f reject omitted endpoint swap/min guards; g/h/i distinguish strict type bounds, the 128-entry cap and resolved-list eligibility.
  - runtime_observed: no

- **MHD-V13** Hit result replacement/final dispatcher result (MHD-R02/MHD-R05/MHD-R09/MHD-R13; boundary: supplied-input stage)
  - (a)
    - Inputs: target 7,CollideKill=1,hit selector 1; supplied hit callback returns 2; direct hand-off no-op; room R,x=4,y=5,Size=2.
    - Seeds: no draw.
    - Draws: none.
    - Trace: initial result 3 replaced 2; damage bit set, stop bit clear.
    - Outputs: hit 1 then fill then hand-off, final 1, no footprint.
  - (b)
    - Inputs: case a independently with hit callback returns 3.
    - Seeds: no draw.
    - Draws: none.
    - Trace: result 3 retains damage and stop; stopMask=3.
    - Outputs: hit 1,fill,hand-off,footprint-clear(R,4,5,2,64),final 2.
  - (c)
    - Inputs: case a independently with callback returns 4,force 1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: mask 4 exits before damage/final-force test.
    - Outputs: hit 1,final 1; no fill,hand-off,footprint.
  - (d)
    - Inputs: target 7,hit selector 30.
    - Seeds: no draw.
    - Draws: none.
    - Trace: eligible null hit slot.
    - Outputs: event 0 placement then null-call fault; no fill or final return.
  - Discriminates: a rejects ORing returned 2 into initial 3; c rejects force overriding callback abort; d rejects null-as-no-handler.
  - runtime_observed: no

- **MHD-V14** Shared direct-apply ordering (MHD-R11a/MHD-R12; boundary: supplied-input stage)
  - (a)
    - Inputs: direct apply only, owner reference valid/creator C found; target 7 state 54=false; definition GetHit=1,KnockBack=0,HitClass=3; incoming physical 100,resultFlags=8192,hitClass=32,reservedCondition=0,flags 0; source A/B=true; missile 327=7. Supplied block/avoid operation logs preset flags and returns 0; double logs record, sets physical 200; combat process logs then clears hit; result consumer logs without further mutation.
    - Seeds: no draw.
    - Draws: none.
    - Trace: preset result 8197; defender code 0 preserves hit; hitClass=3 OR32=35, reservedCondition=1; double physical 200; source flags 160; reductionScale=7; process sees hit, then clears it; result call still occurs.
    - Outputs: log block(owner C,target 7,blockEnable=1,missile 1,word 8197),double(C,target 7,physical 100,hitClass=35,reservedCondition=1),process(C,target 7,flag 1,physical 200,flags 160,reductionScale=7,word 8197),results(C,target 7,word 8196); final physical 200,resultFlags=8196,hitClass=35,reservedCondition=1,flags 160,reductionScale=7.
  - (b)
    - Inputs: direct apply only, owner reference valid/creator C found,target 7,state 54=false; GetHit=0,SoftHit=0,KnockBack=0,HitClass=0; incoming physical 100,resultFlags=0,hitClass=0,reservedCondition=9; source flags false,327=0; supplied defender code 1; double no mutation; result consumer logs.
    - Seeds: no draw.
    - Draws: none.
    - Trace: preset 1; code 1 with GetHit=0/KnockBack=0 rewrites to soft 16384 and discards code; class 0 leaves reservedCondition=9; process skipped.
    - Outputs: resultFlags=16384,hitClass=0,reservedCondition=9,physical 100; ordered block,double,results; no process.
  - Discriminates: a rejects results skipped after process clears hit or stale record copies; b rejects ORing block 16 after the soft rewrite and rejects clearing reservedCondition when class has no high-nibble bit.
  - runtime_observed: no

- **MHD-V15** Separate no-target entry (MHD-R14; boundary: supplied-input stage)
  - (a)
    - Inputs: valid missile definition,doCall=1,hit selector 1; supplied callback logs absent target and returns 4; removal logs request.
    - Seeds: no draw.
    - Draws: none.
    - Trace: eligible callback, ignored result, unconditional removal request.
    - Outputs: ordered hit 1(game,missile,target absent),remove(game,missile); callback 4 does not select continue; doCall=0 skips invalid-definition access, while doCall=1 faults before removal.
  - (b)
    - Inputs: separate no-target entry, doCall=0; missile present with invalid class -1 and definition count 1; hit selector unavailable; removal operation logs.
    - Seeds: no draw.
    - Draws: none.
    - Trace: doCall=0 skips definition access and callback selection completely.
    - Outputs: removal request(game,missile); no callback or definition fault.
  - (c)
    - Inputs: separate no-target entry, doCall=1; missile present with invalid class -1 and definition count 1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: requested callback preparation resolves no definition, then unguarded selector read faults.
    - Outputs: fault before hit callback or removal request.
  - Discriminates: actual final removal request rejects applying normal contact mask 4 semantics to this entry.
  - runtime_observed: no

- **MHD-V16** Creator chain/filter (MHD-R01b/MHD-R03c; boundary: supplied-input stage)
  - (a)
    - Inputs: lookup only, game present,owner kind 1,id 130; selected bucket 2 chain[(id 2,A),(id 130,B),(id 130,C)].
    - Seeds: no draw.
    - Draws: none.
    - Trace: 130 AND 127=2; first node mismatches, second matches.
    - Outputs: creator B; no chain mutation/draw.
  - (b)
    - Inputs: target monster 7,contact selection 1,LastCollide=1,saved 0/9.
    - Seeds: no draw.
    - Draws: none.
    - Trace: save 1/7 then players-only filter refuses monster.
    - Outputs: saved 1/7,return 1,no cooldown/event/fill.
  - (c)
    - Inputs: target monster 7,contact selection 4294967295,ToHit=0; supplied direct hand-off logs no-op.
    - Seeds: no draw.
    - Draws: none.
    - Trace: unsigned selection exceeds 3, accepts; normal fill/hand-off; no stop.
    - Outputs: event 0,fill,hand-off,return 1.
  - (d)
    - Inputs: lookup only, game present, owner kind 5,id 130; separate kind-5 chain[(id 2,A),(id 130,B)]; all kind 0..4 hash chains empty.
    - Seeds: no draw.
    - Draws: none.
    - Trace: kind 5 selects its separate chain, then the second entry is the first matching id.
    - Outputs: creator B, no chain mutation.
  - (e)
    - Inputs: isolated kind filter, target monster kind 1/id 7, contact selection 0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: selection 0 refuses even this ordinary monster.
    - Outputs: accepted=false; caller contact result 1 before later callbacks/fill.
  - (f)
    - Inputs: isolated kind filter, contact selection 2, target monster kind 1/id 7.
    - Seeds: no draw.
    - Draws: none.
    - Trace: selection 2 accepts kind 1.
    - Outputs: accepted=true; this isolated stage does not decide the caller's final return.
  - (g)
    - Inputs: isolated kind filter, contact selection 2, target player kind 0/id 7.
    - Seeds: no draw.
    - Draws: none.
    - Trace: selection 2 refuses kind 0.
    - Outputs: accepted=false; caller contact result 1 before later callbacks/fill.
  - Discriminates: a rejects last-match lookup; c rejects signed-negative filter refusal; d rejects hashing kind 5, and e/f/g distinguish the explicit filter selectors.
  - runtime_observed: no

### 8.2 Observation vectors

No matching gameplay observation is recorded. Procedures and eligibility limits are in the [companion](../research/spec-evidence/missile-hit-and-damage-calculation.md#observation-procedures).

- **MHD-O01** Measure the missile seed and filled damage for known positive ranged elements; expect physical/fire/magic/lightning/cold/poison/burn order and one step per positive wrapped range, under MHD-R10.
- **MHD-O02** Follow a Pierce missile through aligned monsters; expect a stat 328 decrement request per eligible nonzero counter and continue 1 while stopMask=2, under MHD-R04/MHD-R13.
- **MHD-O03** Observe a missile with nonzero stat 120 against an ordinary monster and a hireling; expect the ordinary monster's floored armor-write request and no such request for the latter, under MHD-R11c.
- **MHD-O04** Remove a missile's creator during flight; expect no direct damage/results through MHD-R12 when lookup is absent. Armor requests and independent callback effects remain separate.

## 9. Acceptance and dependencies

Partial supplied-input acceptance requires MHD-V01..MHD-V16's verified assertions, covering MHD-R01..MHD-R14 and their labeled sub-rules. Fault assertions need a controlled semantic failure fixture, not a real invalid pointer. Supplied callbacks and persistent fixture writers verify local argument/order/return contracts only. No vector closes actual registry production, relation-matrix production, callback bodies/row bindings, footprint writes, stat propagation or cooldown expiry gates. Full reproduction additionally needs those owners' integrated evidence and matching applicable observations under MHD-Q04; a partial pass does not establish complete parity.

**Depends on:** NRF-R01/NRF-R02/NRF-R05/NRF-R06a/NRF-R08a/RLI-R03/RLI-R02a/RLI-R02b for RNG, wrapping/rounding, units and existing stat/state reads; GDA-R10/GDF-R02/GDF-R03 for prepared missile fields and names; MDP-R08/MDP-R12 for creation critical and stored damage stats; MCI's owner/contact/pierce initialization contracts; PVP-R05a/PVP-R07 for supplied hostility/owner roles; HIT-R01 through HIT-R08/SBA-R06/SBA-R11 for hit and defender checks; KNB-R02a/KNB-R02b and HRC-R01b/HRC-R01c for reaction preparation; IEP-R07 for contact-event dispatch; STL-R01/STL-R06a/STL-R06b/STL-R07 for cooldown requests and their gates; DPO-R08d/DPO-R13/DPO-R22 and HRC's result consumer for direct process/results. Their integration gates remain in force. Owners of the cited HIT, GDA, RLI ids: [combat-hit-roll](combat-hit-roll.md) (HIT); [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** missile-movement-and-collision for contact return and no-target requests; Blade Sentinel/Fury for shared filling/cooldown; combat-knockback and combat-hit-recovery for incoming missile result preparation. DPO's separate queued reaction placement is not this spec's direct-contact execution.

## 10. Related specs

- [missile-damage-parameters](missile-damage-parameters.md): creation-time damage values.
- [missile-creation-and-seeding](missile-creation-and-seeding.md) and [missile-initialization-and-registry](missile-initialization-and-registry.md): creator/contact/counter initialization.
- [missile-movement-and-collision](missile-movement-and-collision.md): contact trigger and caller return handling.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): shared base fill and cooldown.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): process/apply and monster double.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): hostility, scaling and owner roles.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): hit and defender checks.
- [combat-knockback](combat-knockback.md): missile chance and later motion.
- [combat-hit-recovery](combat-hit-recovery.md): request flags and result reaction.
- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md): event 0 traversal.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): list/event lifecycle.

## Appendix A. Provenance

Original Lord of Destruction Expansion 1.13c Windows x86, using [source manifest](../research/source-manifest.md) artifacts LOD113C-D2GAME/LOD113C-D2COMMON. Compatible static records retain their 2026-10-12 dates; targeted original-code readings and original-image corroboration date 2026-10-15. Named cells reuse the coordinated 2026-10-15 Storm-loader record with its profile limits. The [companion](../research/spec-evidence/missile-hit-and-damage-calculation.md) maps semantic rules to original locations, argument proofs, corrections and pending observation procedures. No gameplay observation confirms this contract; metadata does not imply user approval.
