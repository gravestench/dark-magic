# Party experience sharing

**Description**

When a monster dies, the game selects the player credited for the kill and calculates experience for that player, eligible nearby party members and an eligible hireling. Party members share a larger experience pool, with each share adjusted for the member's level and bonuses. Only party members close to the victim on the same map take part in the experience share.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: PXL-R, PXL-Q, PXL-V, PXL-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/player-party-experience-sharing.md)
Open evidence gaps: [gap file](../plans/gaps/player-party-experience-sharing.md)

## 1. Scope

This contract begins with a victim's stored experience and level and a supplied killer. It covers credit selection, solo and party calculations, member collection including the same-level party iterator, member scaling and player/hireling award requests. Level-up and award storage have integration gates. Death experience penalties and the corpse-experience lead remain bounded unanswered questions. [Party gold sharing](player-party-gold-sharing.md) owns gold pickup sharing, the gold cap, excess-gold recreation and drop ownership/pickup timers. [Monster scaling](monster-scaling-by-players-and-difficulty.md) owns difficulty/player-count changes to the victim's stored values and NoDrop; quest experience is not defined here.

## 2. Trigger and ownership

D2Game computes these server results in death/reward and party paths. The reached death path checks victim eligibility, enters the experience wrapper and award router, selects a credited player, attempts hireling routing, then selects solo or party experience. PVP-R09a/PVP-R09b/PVP-R10a in [PvP and kill credit](combat-pvp-damage-and-kill-credit.md) own the shared selector/routing contracts and retain their integration limits. The companion records dispatch roots and output-request locations. Operations run once per reached death, rather than every frame. Complete notification/message delivery inside level-up and hireling updates remains unestablished.

## 3. Inputs and data

Integer notation consumes foundations section 1 and NRF-R05. Unless specified otherwise values are signed 32-bit integers; low32 retains 32 bits, i32 interprets them as signed, u32 as unsigned, and trunc rounds toward zero. Stat reads use layer 0 and RLI-R02a/RLI-R02b; own-array reads are expressly named. Getters do not establish aggregation or mutation.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| killer, victim, credited player, member, hireling, owner | Optional units: type, class, identity, mode | Logical units; type/class/mode i32 | Player type 0, monster type 1; other types rejected where stated |
| victim experience X; victim level v | Stats 13 `experience`, 12 `level` | i32 | Router reads both from the victim own array; positive X required |
| member level m; bonus | Stats 12 `level`, 85 `item_addexperience` | i32 | m is caller-supplied: initial routed player/hireling level is own-array, collected member level is effective; bonus is effective |
| current experience; last gain | Stats 13 `experience`, 29 `lastexp` | i32 bits; cap comparisons u32 | Player/hireling addition reads current experience from own array |
| victim eligibility | MonStats `killable`; victim.skipKillExperience | Booleans | Named monster flag and neutral runtime skip condition |
| party eligibility | member.aliveExclusionCondition; player data, party id, party record, ordered member identities | Boolean; unsigned 16-bit party id; logical records | 65535 means no party; condition producers Q02/Q04. `aliveExclusionCondition` is the shared name from [core-game-data-fields](core-game-data-fields.md) section 3, the same condition MSD-R05/MSD-R21 test; it implies no party-specific meaning |
| rooms/positions | Room presence, room level id, dynamic coordinates from NRF-R10b; supplied static positions | Level id i32; dynamic integer coordinates u16 zero-extended to i32; static coordinates i32 | Same-level enumeration; victim-centered XP distance |
| credit inputs | One-hop owner record; owner-link and supplied game registries; active-state membership/mask; primary redirect lists | PVP-R09a/PVP-R09b inputs | Exact selector including victim null erasure |
| prepared experience data | Per-class maximums, ordered class threshold rows, ratio rows, default shift | i32; threshold comparisons u32 | Supplied prepared records; original text correspondence Q01 |
| hireling inputs | PetType count; active lookup result; entry eligibility; owner level; optional record and thresholdFactor | Count/levels/factor i32; logical records | Record/entry production Q07; supplied values independently usable |
| dependency operations | Replace stat, award, level-up, resource callback, hireling update/notification | Supplied operations with ordered logs | Requests distinct from stored effects; Q07 |
| floating inputs | Stored binary floating point with 32-bit storage quotient, raw shares, rounding environment, alternate division selection | Supplied results/environment | Production Q07; explicit raw shares can enter R05a |
| difficulty and effective player count | Game context | i32 | Varied only with other inputs held fixed for R07 |

| Mapping or constant | Origin | Binding |
|---|---|---|
| Stats 12/13/29/85 and death/dead mode identities | Table names GDF-R02/GDF-R05/GDF-R06; selections engine code | Code selects fixed numeric ids/modes; cited data records supply names |
| MonStats `killable`; PetType count identity | GDA-R16/GDA-R04b | Table meaning/count identity; code fixes tests and signed count >7 |
| Experience class fields and ExpRatio | GDA-R07 | Data fields; code fixes accessors; no original values assumed here |
| Level-factor arrays, 89/256, 86/256, squared radius 6400, eight-member maximum, input cap 8388607 | Engine code | Fixed by original code |
| Hireling thresholdFactor | Supplied record; original table correspondence unresolved | Code fixes multiplication; selection/preparation Q07 |

No vector requires a shipped Experience cell. Maximums, ratios, thresholds and factors supplied below are synthetic inputs.

## 4. Rules (execution order)

### PXL-R01. Experience eligibility

**PXL-R01a** Eligibility/routing [verified in 1.13c code (static)]. At the death entry, an absent victim or any victim type other than player/monster returns without an experience request. A monster victim must not already be mode 0 (`death`) or 12 (`dead`) and must be killable. A player victim must not already be mode 0 (`Death`) or 17 (`Dead`). After those death-entry checks, a present killer/victim with skipKillExperience=false enters award routing. The isolated wrapper itself checks presence/skipKillExperience, without adding the death entry's victim type/mode checks; the isolated router requires killer type player/monster and signed-positive own-array victim.experience, without a further victim-type check. Only after those router guards pass, request credited-player selection through R02a. A wrapper-to-router entry is distinct from this later selection request and does not imply an award. Consume PVP-R09a/PVP-R10a for the shared route. Hireling lookup additionally requires present credited-player data and signed PetType count >7. Rejection stops later selection/award operations at that stage; earlier RouterEntry or other already reached requests remain observable.

**PXL-R01b** Gate meanings [unresolved]. Wider skipKillExperience meaning and purpose of the PetType-count threshold remain Q02; supplied tests are independently available.

### PXL-R02. Credited player

**PXL-R02a** Selection [verified in 1.13c code (static)]. Consume PVP-R09a/PVP-R09b. A player killer returns immediately; absent/other-type killer yields none. A monster starts with its one-hop owner result. An eligible killer supplies its first primary creditRedirect list; an eligible victim replaces that selection, including with absence. A present selected list replaces the result with its owner lookup in the supplied game, including failure. An absent selected list retains the original one-hop owner from the owner-link game. Return only a player result. That player's own class, level and bonus supply its scaling.

**PXL-R02b** Killer production [unresolved]. Complete upstream attribution, including each damage-over-time source, remains outside-boundary Q05. PVP-R07 establishes a separate missile creator-resolution boundary, not every caller's killer argument.

**PXL-R02c** Selector roles [verified in 1.13c code (static)]. PVP-R09b defines creditEligible as active-state membership intersecting the supplied credit-state mask on an extended base list. firstCreditList returns the first matching primary-chain list, excluding base and secondary chain. Generic list existence is insufficient. Actual mask/list/owner production remains Q08/PVP-Q02/PVP-Q09.

**PXL-R03** Solo request [verified in 1.13c code (static)]. Party id 65535 selects R05a with X and credited player, then requests R06a with the result and that player's level. No distance or party-member level test applies. Stored effects remain Q07.

### PXL-R04. Party pool

**PXL-R04a** Enumeration, collection and integer pool [verified in 1.13c code (static)]. The same-level iterator invokes its callback once on the credited player when that player has no room, no usable player party data, or party id 65535. A non-solo id without a matching party record returns without a callback; an empty record member list likewise returns without a callback. Otherwise visit record identities in order, resolve through the game's player registry, and invoke the callback only for found members with a room on the credited player's map level.

The XP callback rejects absent members, aliveExclusionCondition=true, other unit types, player modes 0/17 and monster modes 0/12. With a present victim read victim/member coordinates. For dynamic units, a present path supplies NRF-R10b unsigned 16-bit integer coordinates, zero-extended to i32; an absent path supplies 0. The static-victim route for types 2/4/5 reads supplied i32 coordinates directly. These getter routes do not establish world-unit equivalence for every position producer (Q03). Reject if either X coordinate is 0. Compute dx=i32(low32(memberX-victimX)), dy likewise, d=u32(low32(dx*dx+dy*dy)); reject d>6400. With no victim skip the position test. The ninth passing member causes fatal failure, rather than silent truncation. Store admitted members/levels; n is count and sumL the wrapping i32 level sum.

If n<=0 or sumL<=0 request nothing. For n=1 calculate R05a and request R06a for the original credited player, even if the sole admitted member is someone else. For n>=2 calculate p=i32(low32(low32((n-1)*X)*89)), extra=trunc(p/256), total=i32(low32(X+extra)). Pass total/sumL/levels to R04b, then R05a/R06a for each admitted member in order using its raw share. Effective player count does not cap n.

**PXL-R04b** Floating shares [unresolved]. The located path divides signed total by signed sumL, stores a binary floating point with 32-bit storage quotient, multiplies it by each signed member level and converts the product for the raw share. Exact division environment, alternate-division branch, precision/rounding and out-of-range conversion remain Q07. Section 8's supplied example environment uses nearest-even binary floating point with 32-bit storage quotient, exact multiplication then truncation for finite in-range values; examples do not establish the original environment. Explicit raw shares independently enter verified R05a.

**PXL-R04c** World distance conversion [unresolved]. Raw getter-scale comparison is established. Conversion to yards/screens remains Q03/NRF-Q03 and does not gate supplied coordinates.

### PXL-R05. Per-member scaling

**PXL-R05a** ScaleXP [verified in 1.13c code (static)]. With supplied member, m, victim level v and x:

1. Signed x>8388607 becomes 8388607; otherwise x<=0 immediately returns 1 before maximum-level/type checks.
2. Use player's class for a player, otherwise class 0. If signed m>=classMax(class), return 0.
3. If v<=m choose f=A[min(m-v, 10)]. Else if m>=25 and v>0 set x=NRF-R05(m, x, v). Otherwise choose f=B[min(v-m, 10)]. When f was selected and differs from 256, set x=NRF-R05(f, x, 256); factor 256 leaves x unchanged. A=[256, 256, 256, 256, 256, 256, 207, 159, 110, 61, 13], B=[256, 256, 256, 256, 256, 256, 225, 174, 92, 38, 5], indices 0..10. Differences use wrapping i32; unsafe indices outside original-profile level inputs remain Q07.
4. Read ratio=ExpRatio(m), shift=ExpRatio(0) through R05b. If x>0 and 0<shift<32, limit=2147483647 arithmetic-shift-right ((shift+(ratio arithmetic-shift-right shift)) modulo 32). If signed x>limit set x=i32(low32((x arithmetic-shift-right shift)*ratio)); otherwise x=i32(low32(x*ratio)) arithmetic-shift-right shift. Other inputs leave x unchanged.
5. If bonus!=0 set x=i32(low32(x+NRF-R05(bonus, x, 100))). Argument order matters. NRF-R05 can divide first, wrap narrow products or fault; arbitrary small x/bonus do not justify unbounded multiplication.
6. A player returns x. A monster requires a one-hop player owner and supplied available per-owner record; absent requirements return 0. Read T(m), T(m+1) through R05d; cap=u32(low32(T(m+1)-T(m))) logical-shift-right 6. Return x if u32(x)<=cap, otherwise cap bits. Other/absent units return 0. Actual record/owner production remains Q07/Q08, distinct from the supplied cap calculation.

**PXL-R05b** Prepared experience accessors [verified in 1.13c code (static)]. Class outside signed 0..6 becomes 0. classMax reads that class's prepared header maximum. threshold(class, L) reads that class from prepared row L+1 without a level bounds guard or null-table guard. ExpRatio returns 0 for absent table, header ratio for signed index<=0, 0 for positive index above header class-0 maximum, otherwise ratio from prepared row index+1. Max/threshold require available records; absent storage is not a zero default.

**PXL-R05c** Text/compiled alignment [unresolved]. R05b does not establish each prepared row's text Level label or a particular original default shift. That correspondence remains Q01.

**PXL-R05d** Supplied hireling threshold [verified in 1.13c code (static)]. T(L)=i32(low32(L*(L+1)*L*thresholdFactor)) for the supplied record factor, with each addition/product wrapping at 32 bits. Record selection/preparation remains Q07; arithmetic has no table lookup/draw.

### PXL-R06. Player award requests

**PXL-R06a** Capped request sequence [verified in 1.13c code (static)]. Absent/non-player input returns without requests. Read own-array old experience. new=low32(old+award); cap=threshold(class, classMax(class)-1). If u32(new)>u32(cap) replace new with cap. Request replacing stat 29 with low32(new-old), then stat 13 with new, at layer 0. Read levelFromExp(class, new) through R06b. If different from supplied entry level request level-up, then resource callback event 12 with two zero ancillary arguments; equality requests neither. Actual storage, progression/resources and messages remain Q07.

**PXL-R06b** Prepared level scan [verified in 1.13c code (static)]. Normalize class as R05b. Start result=0 at prepared row 1's class threshold. If u32(experience) is below it return 0. Otherwise, while result<signed classMax, advance one row/increment result; stop when experience is unsigned below that newly read threshold or maximum reached. Return result. Available rows are required; original data correspondence Q01.

**PXL-R07** Local context independence [verified in 1.13c code (static)]. With victim/member inputs and dependency results fixed, local award/scaling/pool/request calculations do not read difficulty or effective player count or recalculate victim values. This excludes unread dependency interiors.

### PXL-R08. Hireling requests

**PXL-R08a** Share, gates and updates [verified in 1.13c code (static)]. Consume PVP-R10a for signed PetType count>7 and supplied eligible monster lookup. Calculate x_m through R05a. Retain it if this hireling is the killer; otherwise use trunc(i32(low32(x_m*86))/256). Reject signed x_m<=0, missing supplied hireling record, or signed hireling level>=owner effective level. Read own-array old hireling experience; new=low32(old+2*x_m). Reject hireling level>=classMax(0)-1; otherwise request stat 13=new with no experience cap, then request an experience-change notification carrying hireling, stat 13, old/new values and the supplied owner delivery context. Starting at existing level, test T(level+1)<=u32(new), advancing while below classMax(0)-1; stop on first failed unsigned threshold. A rise requests hireling progression, owner update, unit notification and resource handling in that order; actual effects Q07. This is additional to player routing.

**PXL-R08b** Entry production [asm-only]. Located iteration tests a neutral entry-exclusion condition before resolving a monster hireling. Complete entry identity/creation/lookup remains Q07/PVP-Q09; no gameplay name is inferred for the condition.

**PXL-R09** NoDrop relationship [verified in 1.13c code (static)]. Consume MSD-R21 for NoDrop eligibility. Shared same-level enumeration omits the XP victim-distance filter: a far living same-level member can count there while excluded from XP. Do not substitute XP n into NoDrop.

### PXL-R10. Death experience/corpse lead

**PXL-R10a** Penalty consumer [unresolved]. GDA-R05 identifies DifficultyLevels.DeathExpPenalty; complete death-time consumer/lost-experience result remains Q06. No absence claim is established.

**PXL-R10b** Corpse lead [recovered; unverified]. A player-creation lead suggests assigning 75 percent of a pending corpse-experience value to a new corpse's stat 13 and clearing pending value. Producer, consumer and arithmetic/store effects remain Q06; this is not an executable penalty/recovery contract.

**PXL-R11** Direct random draws [verified in 1.13c code (static)]. Credit selection, local supplied-input amount calculations, collection and award request routing perform no direct random draw. Dependency results are supplied. Actual level-up/resource/hireling interiors, stat propagation and other death effects are excluded; their draws are not established here. GPS-R06 states the gold-split counterpart.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| PXL-R01a | Eligibility | verified in 1.13c code (static) | none for supplied gates | PXL-V06, PXL-V13 |
| PXL-R01b | Gate meanings | unresolved | Q02: producer meaning | none |
| PXL-R02a | Credit selection | verified in 1.13c code (static) | Q08: registration only | PXL-V07 |
| PXL-R02b | Upstream killer | unresolved | Q05: attribution | none |
| PXL-R02c | Selector roles | verified in 1.13c code (static) | Q08: producers only | PXL-V07 |
| PXL-R03 | Solo request | verified in 1.13c code (static) | Q07: stored effects only | PXL-V06, PXL-V08 |
| PXL-R04a | Collection/integer pool | verified in 1.13c code (static) | Q07: floating producer/stored awards only | PXL-V04, PXL-V09 |
| PXL-R04b | Floating shares | unresolved | Q07: environment/division/conversion | PXL-V04, PXL-V09 supplied examples only |
| PXL-R04c | World conversion | unresolved | Q03: yards/screens | none |
| PXL-R05a | Scaling | verified in 1.13c code (static) | Q01/Q07/Q08: data/record/owner production only | PXL-V01, PXL-V02, PXL-V03, PXL-V04, PXL-V06, PXL-V10, PXL-V13, PXL-V16 |
| PXL-R05b | Prepared accessors | verified in 1.13c code (static) | Q01: original data only | PXL-V10 |
| PXL-R05c | Text alignment | unresolved | Q01: correspondence | none |
| PXL-R05d | Threshold factor | verified in 1.13c code (static) | Q07: record preparation only | PXL-V13 |
| PXL-R06a | Player requests | verified in 1.13c code (static) | Q07: storage/level-up/callback | PXL-V11 |
| PXL-R06b | Level scan | verified in 1.13c code (static) | Q01: original data only | PXL-V11 |
| PXL-R07 | Local context | verified in 1.13c code (static) | none within stage | PXL-V12 |
| PXL-R08a | Hireling requests | verified in 1.13c code (static) | Q07: lookup/storage/update | PXL-V05, PXL-V13 |
| PXL-R08b | Active entry | asm-only | Q07: entry producer/lookup | none |
| PXL-R09 | NoDrop link | verified in 1.13c code (static) | MSD dependency limits | PXL-V14 |
| PXL-R10a | Penalty | unresolved | Q06: consumer/result | none |
| PXL-R10b | Corpse lead | recovered; unverified | Q06: producer/consumer | none |
| PXL-R11 | Direct draw boundary | verified in 1.13c code (static) | Q07: excluded interiors | PXL-V15 |

## 5. Outputs and side effects

| Result or requested effect | Rules |
|---|---|
| Wrapper-to-router entry request and, separately, credit-selection request after the router guards | R01a/R02a; entry does not imply an award |
| Credited player/none; owner lookup arguments | R02a/R02c |
| Ordered collection, count/sum; fatal ninth-member failure | R04a |
| Amount, pool, supplied cap/threshold/level results | R04a/R05a/R05b/R05d/R06b/R08a |
| Replacement requests stat 29 then stat 13, layer 0 | R06a; actual storage Q07 |
| Level-up then resource event-12 requests | R06a; progression/resources/messages Q07 |
| Conditional active-hireling lookup attempt and supplied result | R01a/R08a/PVP-R10a; full lookup/entry production R08b/Q07 |
| Hireling stat-13, experience-change and level/update/notification/resource requests | R08a; actual writes/messages Q07 |
| Dependency-owned NoDrop count before normalization, used only for the count comparison | R09/MSD-R21; no treasure-roll assertion |
| No direct seed update within local stage | R11 |

No persistence or delivered-message guarantee is established for gated interiors.

## 6. Branches and edge cases

Solo ignores distance. Party collection is victim-centered, rejects zero X, compares wrapped distance unsigned, rejects death/dead modes and fails at member nine. Missing party records produce zero callbacks; early no-room/no-party guards invoke one. n=1 credits the original player. Nonpositive ScaleXP returns 1 before maximum-level checks. XP/hireling caps compare unsigned. Classic/custom-data equivalence is not established beyond Expansion 1.13c and supplied-input operations.

## 7. Open questions

| Id | Kind / status | Assertion and evidence needed |
|---|---|---|
| **PXL-Q01** | integration / open | Original Experience Level/header text, compiled preparation and loaded-record correspondence for R05b/R06b. Dated GDF complete file coverage exists; named level/ratio correspondence is still needed. Compare complete named rows/prepared values with consumer indices; partial text maximum alone is insufficient |
| **PXL-Q02** | outside-boundary / open | Wider skipKillExperience/PetType gate purpose: trace producers/qualifying inputs; supplied consumer gates available |
| **PXL-Q03** | outside-boundary / open | Raw scale to yards/screens, NRF-Q03; supplied-coordinate predicate available |
| **PXL-Q04** | outside-boundary / open | Collection-exclusion producer: trace writers; supplied test available. The gold ownerExceptionCondition producer is GPS-Q01 |
| **PXL-Q05** | outside-boundary / open | Full killer selection by source, especially damage over time; consume PVP-R07 and source-specific contacts, then trace remaining producer |
| **PXL-Q06** | behavior / open | Complete DeathExpPenalty consumer and corpse production/recovery. Quest experience has no contract here. Drop policy is GPS-Q03. Root/output anchors, complete callee reading and distinguishing vectors needed |
| **PXL-Q07** | integration / open | Actual stat write propagation, level-up/resources/hireling effects/messages, active lookup/record preparation, extreme scaling indices, floating environment/alternate divide/conversion. Gold storage and excess creation are GPS-Q02. Supplied arithmetic/requests available; read specific interiors and reproduce attributable cases |
| **PXL-Q08** | outside-boundary / open | Owner registration and redirect mask/list producers. PVP-R09 establishes one-hop indirection/selector, not production history |

Every verified behavioral rule still lacks matching gameplay observations for this contract. This observation gap is separate from integration gates and appears in the [gap file](../plans/gaps/player-party-experience-sharing.md).

## 8. Test vectors

### 8.1 Executable vectors

All vectors are supplied-input stage fixtures with Seeds: no draw; Draws: none; runtime_observed: no. Named dependency operations log requests in order; none mutates original storage. A named synthetic store applies requests only within its case. Defaults: player m=v=10, synthetic max=99, ratio=0, shift=0, bonus=0, own/effective values agree, own XP=0, cap=20000000, returned level equals entry level, so no level-up callback. Party inputs have present same-level rooms, alive mode 1 and nonzero X. Floating examples supply nearest-even binary floating point with 32-bit storage quotient and exact multiplication/truncation for finite in-range products, or explicit raw shares. Floating-production assertions are R04b/Q07 gated; collection/integer pool and R05a with explicit raw shares remain verified. Every case starts fresh.

- **PXL-V01** Lower victim (R05a; boundary: supplied-input stage)
  - Inputs: player m=30, v=22, x=1000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: difference 8 selects A[8]=110; trunc(110000/256)=429.
  - Outputs: 429.
  - Discriminates: unpenalized 1000 fails.
  - runtime_observed: no

- **PXL-V02** Higher victim below 25 (R05a; boundary: supplied-input stage)
  - Inputs: m=20, v=27, x=1000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: difference 7 selects B[7]=174; trunc(174000/256)=679.
  - Outputs: 679.
  - Discriminates: B[6]=225 gives 878.
  - runtime_observed: no

- **PXL-V03** Level-ratio branch (R05a; boundary: supplied-input stage)
  - Inputs: m=30, v=33, x=1000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: m>=25 selects NRF-R05(30, 1000, 33)=909.
  - Outputs: 909.
  - Discriminates: the penalty array would retain 1000.
  - runtime_observed: no

- **PXL-V04** Weighted shares (R04a/R04b/R05a; boundary: supplied-input stage)
  - Inputs: admitted A/B levels 20/40; victim v=30, X=1000; credited A; A/victim=(100, 100), B=(101, 100); supplied quotient 22.450000762939453, raw shares 449/898.
  - Seeds: no draw.
  - Draws: none.
  - Trace: sumL=60, p=89000, extra 347, total 1347. Supplied share yields trunc(20*share)=449, trunc(40*share)=898. B[10]=5 yields 8; A[10]=13 yields 45.
  - Outputs: verified n=2, sumL=60, total 1347; verified R05a amounts 8/45 with supplied raw inputs; award requests(A, 8), (B, 45). Original float production gated R04b.
  - Discriminates: flat 35 percent gives total 1350; equal weighting fails 449/898.
  - runtime_observed: no

- **PXL-V05** Non-killer hireling (R08a; boundary: supplied-input stage)
  - Inputs: supplied scaled x_m=1000, hireling not killer, level 20, owner level 30, own XP=0, max 99, record present, thresholdFactor=100, next threshold 970200.
  - Seeds: no draw.
  - Draws: none.
  - Trace: trunc(86000/256)=335; new=0+2*335=670; threshold exceeds new.
  - Outputs: share 335; replace(13, 670), experience-change(13, 0, 670); no rise request.
  - Discriminates: missing doubling gives 335; missing sharing gives 2000.
  - runtime_observed: no

- **PXL-V06** Death gates and routed input sources (R01a/R03/R05a; boundary: supplied-input stage)
  - Inputs: monster victim mode 1, X=500, killable=true, skipKillExperience=false, present player killer. Baseline and cases (a)..(i)/(l) log the post-guard CreditSelection request, not mere router entry; case (d) additionally observes RouterEntry, case (j) observes only wrapper forwarding, and case (k) uses its stated scaling/update log.
  - Seeds: no draw.
  - Draws: none.
  - Trace: apply each fresh rejection independently.
  - Outputs: baseline one CreditSelection request; no stored-write assertion.
  - Discriminates: cases isolate each rejected condition; (d) separately observes wrapper forwarding and the later positive-X guard, so rejecting before router entry or proceeding to credit selection at X=0 fails.
  - (a) Inputs: mode 12. Outputs: no request.
  - (b) Inputs: skipKillExperience=true. Outputs: no request.
  - (c) Inputs: killable=false. Outputs: no request.
  - (d) Inputs: own X=0. Outputs: exactly RouterEntry at wrapper forwarding, then return before CreditSelection; no hireling, solo/party or level-update request. Entering the router is not a successful positive-X check.
  - (e) Inputs: killer type 3. Outputs: no request.
  - (f) Inputs: death-entry victim player type 0, mode 1, X=500, killable=false, skipKillExperience=false, player killer present. Draws: none. Trace: player branch ignores the monster killable test and passes mode eligibility. Outputs: one CreditSelection request.
  - (g) Inputs: death-entry player victim type 0/mode 0, X=500, skipKillExperience=false, player killer present. Draws: none. Trace: player death mode returns before wrapper. Outputs: no experience request.
  - (h) Inputs: death-entry player victim type 0/mode 17, X=500, skipKillExperience=false, player killer present. Draws: none. Trace: player dead mode returns before wrapper. Outputs: no experience request.
  - (i) Inputs: five fresh death-entry runs: victim types 2, 3, 4, 5 and absent victim; present victims have mode 1/X=500/skipKillExperience=false, killer is present player. Draws: none. Trace: each fails the death entry's player/monster type selection. Outputs: no experience request in every run.
  - (j) Inputs: isolated presence/skip wrapper with supplied router-entry logger, present player killer and present type-2 victim mode 1, skipKillExperience=false. Draws: none. Trace: wrapper has no victim-type/mode gate and forwards to router. Outputs: exactly RouterEntry; no claim of death-entry eligibility or complete award.
  - (k) Inputs: isolated award router with player killer/credited P (class 0), party id 65535, present monster victim V, PetType count 7, synthetic class maximum 99/ratio 0/shift 0. Immutable existing arrays with usesResolvedStats=true and minimum adjustment disabled: V.ownStats=[(12,0,10),(13,0,200)], V.resolvedStats=[(12,0,40),(13,0,9999)]; P.ownStats=[(12,0,10),(13,0,0)], P.resolvedStats=[(12,0,30),(13,0,12345),(85,0,0)]. Definitions are present with count 86; level-update operation logs its request. Draws: none. Trace: router reads own victim XP 200/level 10 and own credited level 10, bypasses hireling lookup at count 7, then solo scaling retains 200. Outputs: solo scaling inputs (P,m=10,v=10,x=200), scaled result 200, level-update request carrying P/entry level 10/award 200; no hireling request or stored-effect assertion. Discriminates: selecting effective victim XP/level or credited level changes the scaling inputs and result.
  - (l) Inputs: fresh death-entry monster victim type 1/mode 0, own experience 500, killable=true, skipKillExperience=false, present player killer. Draws: none. Trace: the monster death-mode guard returns before the presence/skip wrapper and router. Outputs: no RouterEntry, CreditSelection, hireling, solo/party or level-update request. Discriminates: retaining only the monster mode-12 guard would wrongly route this case.
  - (m) Inputs: isolated award router, exactly the immutable arrays, player killer/credited P, no-party, PetType count 7 and synthetic records of case (k), except victim V has type 2. Observe CreditSelection in addition to the stated scaling/update log. Draws: none. Trace: this direct router input bypasses death-entry eligibility; player killer and own XP 200 pass the router guards, selection returns P, and own victim/credited levels 10 give amount 200. Outputs: CreditSelection(game,P,V)=P, solo scaling inputs (P,m=10,v=10,x=200), scaled result 200, level-update request carrying P/entry level 10/award 200; no hireling or outer RouterEntry request. The stage begins inside the router and makes no claim that a type-2 victim passes death-entry eligibility.
  - runtime_observed: no

- **PXL-V07** Credit order/list roles (R02a/R02c; boundary: supplied-input stage)
  - Inputs: owner-link game X maps owner 1 to player P; game G maps redirect owner 2 to Q, owner 3 to R. Eligibility inputs: stateCount=32, credit mask state 7, extended lists and active state 7 when called eligible. Primary credit lists as stated; no other matches.
  - Seeds: no draw.
  - Draws: none.
  - Trace: player bypass, one-hop owner, killer selection, victim replacement, list-owner lookup, player-only result.
  - Outputs: per case below; no award effects asserted.
  - Discriminates: absent selection retains owner; a present list with failed lookup erases it.
  - (a) Inputs: killer player P; eligible victim Q list. Outputs: P, no redirect lookup.
  - (b) Inputs: monster linked to P; both units ineligible. Outputs: P, owner lookup in X.
  - (c) Inputs: monster linked to P; eligible victim primary Q list. Outputs: Q, redirect lookup in G.
  - (d) Inputs: monster linked to P; eligible killer R list; eligible victim no match. Outputs: P, no redirect lookup after victim clears selection.
  - (e) Inputs: killer type 3. Outputs: none.
  - (f) Inputs: monster one-hop owner resolves to monster M, no eligible lists; M has owner P. Outputs: none, no second hop.
  - (g) Inputs: monster linked P; extended base marked creditRedirect/owner Q, secondary Q list, no primary match, active state 7. Outputs: P; base/secondary excluded.
  - (h) Inputs: monster linked P; base/primary Q list exists, no intersecting active state. Outputs: P.
  - (i) Inputs: monster linked P; eligible victim primary owner 99 absent in G. Outputs: none, no owner fallback.
  - runtime_observed: no

- **PXL-V08** Solo distance (R03/R05a/R06a; boundary: supplied-input stage)
  - Inputs: credited player m=10, v=10, X=200, party id 65535, squared distance 1000000; own XP=0, cap 20000000, unchanged level result 10; synthetic store.
  - Seeds: no draw.
  - Draws: none.
  - Trace: solo ignores distance; factor 256 retains 200; capped addition 200.
  - Outputs: award 200; replace(29, 200), replace(13, 200); synthetic store holds values, no level/resource request. Original stores Q07.
  - Discriminates: party range incorrectly applied to solo would suppress requests.
  - runtime_observed: no

- **PXL-V09** Party branches (R04a/R04b/R05a; boundary: supplied-input stage)
  - Inputs: credited P and A/B all level 10; victim level 10/X=1000 at(100, 100); P=(100, 100), A=(101, 100), B=(102, 100); record order P/A/B; supplied finite share environment.
  - Seeds: no draw.
  - Draws: none.
  - Trace: collect, sum, choose n branch, calculate pool/amount.
  - Outputs: baseline n=3, sumL=30, total 1695, supplied share 56.5, raw/amount 565 each, requests P/A/B=565. Float production Q07.
  - Discriminates: range, missing-record, credited identity and fatal count cases have distinct outcomes.
  - (a) Inputs: record P/A only, A=(180, 100). Outputs: d=6400 included, n=2, sumL=20, total 1347; supplied share 67.3499984741211, raw/requests 673 each.
  - (b) Inputs: record P/A only, A=(181, 100). Outputs: d=6561 excluded, n=1, P request 1000.
  - (c) Inputs: record P/A only, A mode 17 at victim. Outputs: n=1, P request 1000.
  - (d) Inputs: P has room/player data, party id 1, matching record absent. Outputs: callback count 0, n=0, sumL=0, no award request.
  - (e) Inputs: P has room/player data, party id 1, record present/empty. Outputs: callback count 0, n=0, no request.
  - (f) Inputs: P has no room, mode 1, at victim(100, 100), party id 1. Outputs: one callback P, n=1, P request 1000.
  - (g) Inputs: record A only, P not enumerated; A alive level 10 at victim; credited P level 20, v=10, X=1000. Outputs: collected A, n=1, sumL=10; request P with A[10]=13, amount 50; no A request.
  - (h) Inputs: record P/A, A X=0, Y=100, P at victim. Outputs: A excluded, n=1, P request 1000.
  - (i) Inputs: nine distinct alive level 10 members at victim. Outputs: first eight admitted;ninth fatal failure before awards.
  - (j) Inputs: isolated collection/pool consumer, supplied victim X=1000/v=10 and static victim type 2 with position (65636, 100); record P/A, both player members level 10 at (100, 100). This does not assert that the static victim passes monster-death award eligibility. Draws: none. Trace: each dx=-65536, dy=0, low32(dx*dx)=0. Outputs: wrapped d=0, n=2, sumL=20, total 1347; supplied raw/requests 673 each. Discriminates: unbounded squared distance would exclude both.
  - (k) Inputs: record P/A only, A.aliveExclusionCondition=true; all other baseline inputs. Draws: none. Trace: exclusion rejects A before position/level collection. Outputs: n=1, sumL=10, P request 1000; no A request.
  - (l) Inputs: isolated member callback, empty collection, supplied present victim (100, 100), member at (100, 100) with level 10. Three fresh runs supply member type 3/mode 1, monster type 1/mode 12, and monster type 1/mode 1 respectively. Draws: none. Trace: type rejection, monster dead rejection, then monster acceptance. Outputs: first two runs n=0/sumL=0; third n=1/sumL=10 with that monster stored. No iterator or award eligibility is asserted for these direct callback runs.
  - (m) Inputs: isolated collector/pool stage, victim absent, supplied v=10/X=1000, sole callback member A is an alive player level 10 at (0, 0), credited P level 10. Draws: none. Trace: victim absence skips coordinate checks, including zero X; n=1 credits original P. Outputs: n=1, sumL=10, P request 1000; no A request.
  - (n) Inputs: ordered record P/A/B; P has baseline room/coordinates, A's room has another map level, B has no registry match. Draws: none. Trace: map/lookup guards omit both callbacks. Outputs: callback list [P], n=1, sumL=10, P request 1000.
  - (o) Inputs: record P/A, both admitted levels 0 at victim position, supplied X=1000/v=10. Draws: none. Trace: n=2 but sumL=0 fails the signed sum guard before sharing/scaling. Outputs: n=2, sumL=0, no award requests.
  - (p) Inputs: record P/A; victim and P have baseline dynamic paths at (100, 100); A is an alive player level 10 with dynamic path absent. Draws: none. Trace: member wrapper returns X=0 without calling the path extractor; zero-X guard excludes A. Outputs: n=1, sumL=10, P request 1000; no A request.
  - (q) Inputs: record P/A with valid baseline member paths, present monster victim whose dynamic path is absent, supplied victim X=1000/v=10. Draws: none. Trace: victim coordinate route yields X=0; collector rejects each member before distance. Outputs: n=0, sumL=0, no award request. This is distinct from absent victim V09(m), which skips distance.
  - runtime_observed: no

- **PXL-V10** Scaling/accessor boundaries (R05a/R05b; boundary: supplied-input stage)
  - Inputs: defaults except overrides; (e)-(g), (i) enter named stage with supplied preceding result.
  - Seeds: no draw.
  - Draws: none.
  - Trace: cap, early return, maximum, factors, ratio, bonus in order.
  - Outputs: exact case results below.
  - Discriminates: early return and masked ratio shift cannot become a single unbounded percentage.
  - (a) Inputs: x=9437184, m=v=10. Outputs: capped 8388607.
  - (b) Inputs: x=0, m=99, max 99. Outputs: 1 before max check.
  - (c) Inputs: x=1000, m=99, max 99. Outputs: 0.
  - (d) Inputs: m=40, v=20, x=1000. Outputs: factor 13, result 50.
  - (e) Inputs: ratio-stage x=1000, shift 10, ratio 512. Outputs: limit 2097151, result 500.
  - (f) Inputs: ratio-stage x=3000000, shift 10, ratio 512. Outputs: shift-first 2929*512=1499648.
  - (g) Inputs: bonus-stage x=429, bonus 50. Outputs: bonus 214, result 643.
  - (h) Inputs: present prepared table, header max class 0=3/class 1=4, header ratio 10, row 2 ratio 512. Outputs: classMax(1)=4, classMax(7)=3, ExpRatio(0)=10, ExpRatio(-1)=10, ExpRatio(1)=512, ExpRatio(4)=0. Independent absent-table ratio 1 read returns 0; no max/threshold read attempted on absent table.
  - (i) Inputs: ratio-stage x=1000, shift 10, ratio 23552. Outputs: ratio>>shift 23, masked limit shift(10+23) modulo 32=1, limit 1073741823; product 23552000 then shift gives 23000.
  - (j) Inputs: x=-1, m=99, max=99. Draws: none. Trace: the nonpositive early return precedes maximum-level checks. Outputs: 1.
  - (k) Inputs: m=v=10, x=1000, supplied class-0 max=99; two fresh runs use member type 3 and absent member. Draws: none. Trace: positive amount and max/factor stages pass using class 0, then the final unit-type hook rejects. Outputs: 0 in both runs; no owner lookup.
  - (l) Inputs: player m=v=10, x=65536, bonus=65536, shift=0. Draws: none. Trace: positive input/factor stages retain 65536; NRF-R05(65536,65536,100) takes the narrow branch, low32(product)=0, bonus result=0. Outputs: 65536. Discriminates: widening the caller's narrow bonus product yields a different gain.
  - (m) Inputs: player m=v=10, x=1000, bonus=-200, shift=0. Draws: none. Trace: level/ratio stages retain 1000; bonus trunc(-200000/100)=-2000; sum=-1000; player hook returns it unchanged. Outputs: -1000, without a second nonpositive-input clamp.
  - runtime_observed: no

- **PXL-V11** Cap/level requests (R06a/R06b; boundary: supplied-input stage)
  - Inputs: player class 0, own XP=1000, effective XP=9999, award 500, entry level 10; supplied cap/returned level per case; operations logged.
  - Seeds: no draw.
  - Draws: none.
  - Trace: wrapping sum 1500, unsigned cap, delta from own 1000; delta request before total, then level comparison.
  - Outputs: per case; actual stores Q07.
  - Discriminates: effective XP=9999 or reversed request order fails log.
  - (a) Inputs: cap 1200, returned level 10. Outputs: replace(29, 200), replace(13, 1200), no callbacks.
  - (b) Inputs: own XP=1200, cap 1200, returned level 10. Outputs: replace(29, 0), replace(13, 1200), no callbacks.
  - (c) Inputs: cap 5000, returned level 11. Outputs: replace(29, 500), replace(13, 1500), level-up, resource(event 12, 0, 0), in order.
  - (d) Inputs: own XP=1000, award=-2000, cap 5000, returned level 10. Outputs: sum bits 4294966296 exceeds cap unsigned; replace(29, 4000), replace(13, 5000), no callbacks.
  - (e) Inputs: isolated prepared class 0 max 3, threshold rows 1..4=[0, 100, 300, 600], experience 299. Outputs: scan 2, threshold(0, 1)=100. Independent experience 300 scan 3; class 7 threshold(class 7, 1)=100 by class 0 normalization.
  - (f) Inputs: isolated prepared class 0 max=3, threshold rows 1..4=[100, 200, 300, 400], experience=99. Draws: none. Trace: unsigned 99 is below the initial threshold 100. Outputs: level scan 0; no row advance.
  - (g) Inputs: isolated prepared class 0 max=3, available threshold rows 1..4=[0,100,300,600], experience=999; no later row is supplied. Draws: none. Trace: experience passes each available threshold; result reaches signed maximum 3 and returns before needing another row. Outputs: level scan 3.
  - runtime_observed: no

- **PXL-V12** Context independence (R07; boundary: supplied-input stage)
  - Inputs: V08 inputs; six fresh combinations difficulty 0/1/2, effective count 1/8; victim/member/dependency results fixed.
  - Seeds: no draw.
  - Draws: none.
  - Trace: no local amount input changes.
  - Outputs: six award 200 requests, each replace(29, 200), replace(13, 200).
  - Discriminates: a second difficulty/count multiplier changes results.
  - (a) Inputs: party fixture with P/A/B alive level 10, same-level rooms, all at victim (100, 100), victim X=1000/v=10, record P/A/B, supplied binary floating point quotient 56.5 with 32-bit storage and raw shares 565 each. Two fresh runs set effective player count 1 and 8. Draws: none. Trace: collected n=3 is unchanged, pool 1000+trunc(2*1000*89/256)=1695; supplied shares scale unchanged. Outputs: both runs n=3/sumL=30/total=1695 and requests (P,565),(A,565),(B,565). Original floating production remains Q07. Discriminates: capping collection by effective count 1 would instead choose n=1 and P-only 1000.
  - runtime_observed: no

- **PXL-V13** Hireling gates/cap (R01a/R05a/R05d/R08a; boundary: supplied-input stage)
  - Inputs: PetType count 20, eligible monster lookup, hireling level 20, owner level 30, v=20, X=1000; owner/record present; thresholdFactor=100, T(20)=840000, T(21)=970200, gap=130200, own XP=0, effective XP=9999, max 99.
  - Seeds: no draw.
  - Draws: none.
  - Trace: scale 1000, unsigned hook cap, killer share, positive/record/level/max guards, doubled addition.
  - Outputs: case requests below; no original stored effects asserted.
  - Discriminates: killer bypass, cap before share, doubling, count>7 and owner equality differ.
  - (a) Inputs: hireling killer. Outputs: x_m=1000, replace(13, 2000), experience-change(13, 0, 2000), no rise.
  - (b) Inputs: owner killer. Outputs: x_m=335, replace(13, 670), experience-change(13, 0, 670), no rise.
  - (c) Inputs: hireling level 30, owner level 30. Outputs: no award/update request.
  - (d) Inputs: PetType count 7. Outputs: no hireling lookup/award request.
  - (e) Inputs: supplied hook cap 400 entering the hook-comparison stage, hireling killer; no threshold production asserted. Outputs: x_m=400, replace(13, 800), experience-change(13, 0, 800). Independent non-killer run with cap 400: x_m=134, replace(13, 268), experience-change(13, 0, 268).
  - (f) Inputs: isolated threshold L=20, factor 1. Outputs: T(20)=8400, T(21)=9702, gap 1302, hook cap 20, x=1000 returns 20. Independent hook-stage x=-1, same cap 20: unsigned comparison returns 20.
  - (g) Inputs: supplied x_m=1000 entering the R08a update stage after hook/share, hireling killer, level 20, own XP=9500, factor 1, max 99. Outputs: new 11500, T(21)=9702, T(22)=11132, T(23)=12696, final level 22; replace(13, 11500), experience-change(13, 9500, 11500), then progression/owner-update/unit-notification/resource requests; interiors Q07.
  - (h) Inputs: supplied x_m=0 at the R08a update entry after share, other baseline update inputs. Draws: none. Trace: nonpositive guard returns first. Outputs: no stat, experience-change or level/update request.
  - (i) Inputs: supplied x_m=1000 at update entry, hireling record absent, other baseline update inputs. Draws: none. Trace: record-availability failure precedes writes. Outputs: no stat, experience-change or level/update request.
  - (j) Inputs: hireling killer, hireling/victim level 98, owner level 99, class-0 max=99, thresholdFactor=100, record present, X=1000, own XP=0. Draws: none. Trace: scaling and owner-level guard pass; existing hireling level 98 equals max-1. Outputs: no stat, experience-change or level/update request.
  - (k) Inputs: hireling award stage, hireling killer with level 97, victim level 97/X=1000, owner level 99, class-0 max=99, thresholdFactor=100, owner/record present, own hireling XP=100000000; ratio/shift/bonus=0. Draws: none. Trace: T(97)=92208200, T(98)=95079600, hook cap=44865 retains 1000; owner-level and existing-level maximum guards pass. New XP=100002000; T(98)<=new advances to 98, which equals max-1 and stops. Although T(99)=98010000 also fits the new XP, it does not advance again. Outputs: replace(13,100002000), experience-change(13,100000000,100002000), final level 98, then progression/owner-update/unit-notification/resource requests in order; actual effects remain Q07. Discriminates: scanning beyond max-1 would give level 99.
  - runtime_observed: no

- **PXL-V14** Separate NoDrop count (R09/R04a; boundary: supplied-input stage)
  - Inputs: P/A alive same-level/party, levels 10, X=1000, v=10; victim/P=(100, 100), A=(200, 100); MSD-R21 supplied valid player/party enumeration.
  - Seeds: no draw.
  - Draws: none.
  - Trace: shared iterator visits both; XP excludes A at d=10000; NoDrop alive callback includes both.
  - Outputs: XP n=1, P award request 1000; NoDrop party count 2 before MSD normalization. No treasure roll executed.
  - Discriminates: substituting XP n incorrectly gives party count 1.
  - runtime_observed: no

- **PXL-V15** Local seeds (R11; boundary: supplied-input stage)
  - Inputs: V08 downstream solo inputs; distinct monster killer M with supplied one-hop player owner P, victim V, credited member P and unused hireling H; PetType count 7. Initial seeds M={lo:1, hi:2}, V={lo:3, hi:4}, P={lo:5, hi:6}, H={lo:7, hi:8}, game={lo:9, hi:10}. Supplied dependencies only log. GPS-V03 covers the gold-split run.
  - Seeds: no draw.
  - Draws: none.
  - Trace: no local random operation reached; dependency interiors excluded.
  - Outputs: all ten seed words unchanged, direct draw count 0; no claim inside stat propagation/level/hireling callbacks.
  - Discriminates: a local award roll changes seed or draw count.
  - runtime_observed: no

- **PXL-V16** Divide-first order (R05a/NRF-R05; boundary: supplied-input stage)
  - Inputs: player m=30, v=24, x=8388607, bonus 50, shift 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: factor 207; NRF-R05(207, 8388607, 256) divides first 32767*207=6782769; bonus NRF-R05(50, 6782769, 100)=67827*50=3391350; add 10174119.
  - Outputs: 10174119. Independent x=1000, bonus 0 control returns 808.
  - Discriminates: full-product 6782975 or bonus 3391384 fails result/operand order.
  - runtime_observed: no

### 8.2 Observation vectors

All remain pending; the [companion procedures](../research/spec-evidence/player-party-experience-sharing.md#observation-procedures) state setup and measurements.

- **PXL-O01** Two/three-member kills: expected supplied arithmetic V04/V09; measure collection/floating environment before comparing complete effects. Include range, map-level, death controls.
- **PXL-O02** Solo summon/golem/trap/hireling/damage-over-time kills: measure upstream killer and final credited player; compare only established selectors with actual inputs.
- **PXL-O03** Hireling killer/non-killer and owner-level controls: expect V05/V13 supplied arithmetic; measure actual stat/level/resources separately.
- **PXL-O04** First establish Experience header/text/prepared correspondence, then reproduce a known-level award; no unknown cell has an invented expected value.

## 9. Acceptance and dependencies

Partial acceptance covers verified supplied-input assertions in V01-V16: eligibility/credit, member collection/integer pool/n=1 identity, scaling/prepared accessors, capped player/hireling requests, local context independence, NoDrop count distinction and no direct local draw. Floating examples do not verify R04b production; synthetic stores do not establish original mutation. Full reproduction additionally requires resolving R01b/R02b/R04b/R04c/R05c/R08b/R10a/R10b at their stated boundaries, complete dependency-backed stored award/level/hireling effects and attributable observations for all verified behavioral rules. A verified-subset pass is partial acceptance. User approval is explicit and separate.

**Depends on:** NRF-R05 ordered signed multiply/divide; NRF-R10b dynamic integer coordinate extraction; RLI-R02a/RLI-R02b stat reads, with RLI-Q03/STL-R06c actual-write limits. GDA-R05/GDA-R07/GDF-R02/GDF-R05/GDF-R06/GDA-R16/GDA-R04b supply named fields/stat/mode/type/count identities. PVP-R09a/PVP-R09b/PVP-R10a/PVP-R10b supply shared death/credit/routing/share contracts with PVP-Q02/PVP-Q09 limits. MSD-R21 supplies the distinct NoDrop count. Supporting producer/schema records do not establish complete integration. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** MSD-R21 and GPS-R01 consume same-level enumeration; PVP-R10b refers amount ownership here. Reverse references do not add another prerequisite calculation.

## 10. Related specs

- [Party gold sharing](player-party-gold-sharing.md): gold pickup split over the same-level party iterator.
- [Monster scaling](monster-scaling-by-players-and-difficulty.md): stored victim values and NoDrop.
- [PvP damage and kill credit](combat-pvp-damage-and-kill-credit.md): source contacts/shared selectors.
- [Timed state lifecycle](state-timed-state-lifecycle.md): propagation integration limits.
- [Damage pipeline](combat-damage-pipeline-order.md), [item events](combat-item-event-triggers-and-procs.md), [Open Wounds](combat-open-wounds.md): upstream source/contact boundaries.
- [Hit recovery](combat-hit-recovery.md): separate mode changes; its draws do not define every death effect.

## Appendix A. Provenance

Evidence build: Lord of Destruction Expansion 1.13c. Original D2Game/D2Common identities are in the [source manifest](../research/source-manifest.md). Compatible static readings retain 2026-10-11/12 dates; dependency and targeted original-image/source checks retain their actual 2026-10-14/15 dates and limits in the [public companion](../research/spec-evidence/player-party-experience-sharing.md). No new gameplay observation is claimed. The companion maps semantic operations to original locations, records corrections, data-profile limits and pending procedures. Status remains draft and does not imply approval.
