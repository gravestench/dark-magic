# Knockback

**Description**

Some hits can push a target away from its attacker and replace its ordinary hit reaction with a knockback reaction. Item knockback is more likely against small monsters than large ones, while missiles use their own chance. A monster's available animations, death and an uninterruptible state can prevent knockback.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: KNB-R, KNB-Q, KNB-V, KNB-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-knockback.md)
Open evidence gaps: [gap file](../plans/gaps/combat-knockback.md)

## 1. Scope

This contract covers item and missile knockback chance, reaction eligibility and selection,
monster knockback start, and the shared path-type-8 endpoint generator. A planned endpoint
is distinct from movement completed over later ticks. Damage and application order belong to
[combat-damage-pipeline-order](combat-damage-pipeline-order.md), block and avoidance to
[combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md), reaction priority
and animation rate to [combat-hit-recovery](combat-hit-recovery.md), and shared movement to
[state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md).
Player displacement, collision clipping, mode-transition integration and completed motion
remain explicit questions within this boundary.

## 2. Trigger and ownership

D2Game's server item-event selector 7 invokes the item chance callback with owner, target,
damage record and stat identity supplied by IEP-R02/IEP-R03/IEP-R07. The missile unit-to-unit
damage path invokes its separate chance check after presetting hit and before block/avoidance
(SBA-R11). Queued-path reaction follows application through DPO-R26. On the missile path,
MHD-R12 owns owner multiplication, hit-conditional direct application and the subsequent
direct reaction, which is called even when the hit bit was cleared. Reaction runs once per call.
Monster mode 13's start dispatch requests path type 8; D2Common's planner dispatch reaches
the endpoint generator. The companion records the dispatch chains and verified boundaries.
Client notification/correction is KNB-Q10; no packet identity is established here.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| Callback arguments | Owner, target, damage record, supplied stat identity | Optional logical references | Missing owner, target or record rejects item callback |
| Item stat | Stat 81 `item_knockback`; callback uses supplied identity rather than fixed 81 | Signed i32 current value, RLI-R02a/RLI-R02b selected-array/minimum contract | Positive enables one draw; magnitude unused |
| Missile chance | `Missiles.KnockBack`, GDA-R10 | u8 | Zero skips; positive compared against remainder out of 100 |
| Result | Per-hit result word, NRF section 3.2 | u16 | Hit 1, lethal 2, GH 4, KB 8, block 16; other flags via HRC-R04a |
| Kind and mode | Player 0, monster 1; monster modes 0 death, 3 gethit, 12 dead, 13 knockback; player modes 4 Get Hit, 19 Knock back | Integer identities, GDF-R05/GDF-R06 | Reaction routing |
| Mode availability | `MonStats2.mGH`, `MonStats2.mKB` | Boolean fields, GDA-R09 | Start and fallback gates; `dKB` unused in these decisions |
| Size | `MonStats2.small`, `MonStats2.large`; record optional | Booleans, GDA-R09 | Item threshold only |
| Town permission | Optional source and optional valid source-monster `MonStats` lookup; `inTown` only from a valid record | Optional references plus Boolean, GDA-R16 | Town reaction requires a monster source and present valid record with permission; a detached flag does not replace the record |
| Mode bookkeeping | `modeChanged`, RLI-R04a | Boolean | V13c supplies the actual equal-mode setter model; engine update integration remains RLI-Q02 |
| Base | Valid monster `MonStats.BaseId` lookup | Signed i16, GDA-R08 | Base 78 exception |
| States | 54 `uninterruptable`, 1 `freeze`; requested state 92 `death_delay` | Predicates / state-on request, GDF-R03/RLI-R03; state 92 text name in companion | Chance/reaction and conversion vetoes; lethal state-54 branch requests state 92 on |
| Room and level | Optional target room and its level | Level as i32 bit pattern | Unsigned bounded town predicate |
| Seeds | Attacker, missile, defender | `{lo,hi}` u32 words | NRF-R01/NRF-R02; separate owners |
| Start | Supplied generator start point | Two u16 path coordinates | Planned path origin |
| Source point | Reference unit's position | Kinds 0/1/3: u16 coordinates; absent dynamic path gives `(0,0)`; kinds 2/4/5: full i32 coordinates | Direction input; other kinds use dynamic accessor branch |
| Explicit destination | Used when reference unit absent | Two u16 path coordinates | Bypasses scaling |
| Distance | Path byte `B` | u8 | Start sets 5 or 10 |
| Path state | Type, reference unit, planned endpoint, first waypoint, waypoint count, actual position | Logical fields; endpoint coordinates u16 | Count may already be nonzero |
| Dependency operations | Clip, planner, mode entry preparation, mode/state/reaction operations | Explicit inputs/results/effects | Supplied-stage boundaries do not prove dependency execution |

NRF-R10 supplies five path units per map tile. Completed knockback displacement has no
matching gameplay observation (KNB-Q04).

| Mapping or constant | Origin | Binding |
|---|---|---|
| `item_knockback` identity 81; `itemeventfunc1/2 = 7` | Effective `data/global/excel/itemstatcost.txt`, row `item_knockback`, explicit ID and zero-based data-row index 81 | Named text cells verified in the [2026-10-15 record](../research/spec-evidence/combat-knockback.md#item-stat-text-cells-2026-10-15); compiled correspondence KNB-Q01 |
| `itemevent1 = domeleedamage`, `itemevent2 = domissiledamage` | Same file/row, named columns | Data chooses event names; numeric binding and execution KNB-Q02 |
| Selector 7 callback | Engine table, IEP-R02/IEP-R06 | Data selects engine-defined handler; registration integration IEP-R04 |
| State 54/1; base 78 `sandleaper1` | Named mappings GDF-R03/GDF-R08 | Numeric identities fixed by code consumers |
| State 92 `death_delay` | Effective `data/global/excel/states.txt`, explicit `id = 92`, zero-based data-row index 92, column `state`; [named-cell record](../research/spec-evidence/combat-knockback.md#state-92-text-name-2026-10-15) | Code fixes request id 92; text supplies its name, without establishing timers or death-mode execution |
| Thresholds 32/64/128; path type 8; byte 5/10; town ids 1/40/75/103/109 | Engine code | Fixed by original code; no remapping capability established |

## 4. Rules (execution order)

**KNB-R01 Item roll** [verified in 1.13c code (static)]. With owner, target and damage
record present, read owner's current supplied stat using RLI-R02a/RLI-R02b's selected-array
lookup and minimum adjustment. A nonpositive value
returns 0 without draw/write. Otherwise step the attacker's seed once through NRF-R01;
`roll = newLow & 127`. Threshold is 32 for a monster with `large`, else 128 for a monster
with `small`, else 64. Missing size record uses 64; large precedes small. If `roll < threshold`,
OR 8 into result and return 1; otherwise return 0 without changing result. These thresholds
accept 32/128, 128/128 and 64/128 possible mask outcomes respectively. Positive stat magnitude
does not change the threshold. Missing owner/target/record returns 0 without draw/write.

**KNB-R02a Missile roll** [verified in 1.13c code (static)]. After KNB-R02b permits it,
a nonzero unsigned chance byte causes one NRF-R02 `Rand(missile,100)` draw. If the unsigned
remainder is less than the byte, OR 8 into result. Zero skips without clearing existing flags.
Bytes 1, 33, 75 cover 1/100, 33/100, 75/100 remainder values; bytes at least 100 always pass.
Surrounding hit/GH/soft flags are SBA-R11/HRC-R01b, outside this isolated chance stage.

**KNB-R02b Missile state gate** [verified in 1.13c code (static)]. State 54, tested through
RLI-R03, skips the missile chance stage: no seed step or KB OR. It does not clear an existing bit.

**KNB-R03b Town gate** [verified in 1.13c code (static)]. At reaction entry, a target room
with level in `{1,40,75,103,109}` requires a monster source with valid supplied `MonStats`
record and `inTown` true; otherwise return without reaction. No room or a level outside
that set passes. The predicate subtracts 1 with u32 wrap, rejects unsigned values above 108,
then consults the fixed table. This gate precedes state/type checks. No direct draw.

**KNB-R05a Uninterruptible reaction** [verified in 1.13c code (static)]. After town gate,
state 54 returns before type-specific dispatch. With lethal set, first request
`SetState(target,92,1)` through RLI-R03 and STL-R07; otherwise request nothing.
This requests state 92 `death_delay` on, using those owners' state-range/storage guards and membership
contract. Only disguise bookkeeping and update-queue integration remain KNB-Q11 through
STL-R07a/STL-Q07 and RLI-Q02. Ordinary death-mode execution is not asserted here. No direct draw.

**KNB-R03a Target dispatch** [verified in 1.13c code (static)]. Once preceding gates pass,
null targets and kinds other than player/monster request no reaction. Monster mode 0 or 12
returns before further reaction. This dispatch contains no hireling/summon/NPC subtype test;
complete subtype equivalence is KNB-Q08. No direct draw.

**KNB-R06a Monster fallback** [verified in 1.13c code (static)]. If monster result has 8
and `mKB` is false, use u16 `result = (result & ~8) | 4`. Retain all other flags. No draw.

**KNB-R06c Base conversion** [verified in 1.13c code (static)]. After fallback, if GH bit 4
is set, target is a monster with valid base lookup 78 and state 1 absent, OR 8 into result.
Invalid lookup does not convert. This may re-add 8 after fallback; there is no second `mKB`
test here. No draw.

**KNB-R04 Death relative to KB** [verified in 1.13c code (static)]. Consume HRC-R04a:
monster lethal precedes KB. Player avoidance/block precede lethal; absent those branches,
lethal precedes KB. The winning branch requests death rather than KB; death's complete
effects are outside this selection stage. Selection has no direct draw.

**KNB-R06b Reaction selection** [verified in 1.13c code (static)]. Consume HRC-R04a after
fallback/conversion: monster death > KB > block > GH > soft > plain; player avoidance > block
> death > KB > GH > soft. KB requests mode 13 for monsters, 19 for players. With both KB and
GH, select only KB. Stun's GH-veto bypass (HRC-R05) is reached only if GH wins and does not
preempt KB. Complete selected-operation effects retain their dependency limits.

**KNB-R06d Monster start** [verified in 1.13c code (static)]. For supplied availability,
return 0 without local path/mode writes unless both `mGH` and `mKB` are true. With both true,
a missing unit or mode 0 returns 1 without these writes. Otherwise set path type 8, set byte
through KNB-R07a, call `planner(target,0)`, call mode entry preparation, request mode 13 through
RLI-R04a, return 1. Ignore planner return. `dKB` is unused. Complete planner/preparation/mode
effects are KNB-Q06/KNB-Q07. No direct RNG instruction in this handler.

**KNB-R07a Distance** [verified in 1.13c code (static)]. Monster start writes unsigned
`B = 10` for valid base 78, else 5. Generator computes `S = (B >> 1) + 1` after unsigned
widening: 6 or 3 path units respectively. These construct the endpoint, not completed travel.

**KNB-R08 Endpoint arithmetic** [verified in 1.13c code (static)]. With reference unit,
widen each u16 start coordinate to i32; source coordinates retain the widths in section 3.
Compute wrapping i32 `v = start - source`, wrapping i32 absolute values (minimum i32 negates
to itself), and signed `m = max(abs(v.x),abs(v.y))`. If `m != 0`, replace each component with
`trunc0(wrap_i32(S * v.component) / m)` using signed i32 division. If `m == 0`, retain `v`;
ordinary coincident coordinates give `(0,0)`. Add components to start with low-16-bit wrap.
Without reference unit, use explicit destination directly, without scaling. No direct draw.
Extreme inputs are not given saturation or unbounded arithmetic. A nonzero `m` is positive
or the minimum i32 value, never -1; zero skips division. Signed division therefore has no
fault case within this declared numeric domain.

**KNB-R09a Planned path writes** [verified in 1.13c code (static)]. Store requested endpoint
in first waypoint storage, call clip with start/requested endpoint, store supplied clipped
endpoint as planned endpoint and clear reference unit. If clipped equals start, return 0
without writing count or replacing the stored requested waypoint. Otherwise overwrite first
waypoint with clipped endpoint, set count 1, return 1. Actual position is not written here.
No direct draw outside clip dependency.

**KNB-R09b Collision and actual motion** [unresolved]. Collision endpoint selection and
later movement are KNB-Q07. Monster start continues to mode request after a supplied zero
planner return, but a complete blocked-motion result also needs waypoint initialization,
collision and movement integration evidence.

**KNB-R10a Draw boundaries** [verified in 1.13c code (static)]. Item chance draws once on
attacker after positive-stat guards; missile chance once on missile after its two guards.
Local selection/start/endpoint instructions add no direct draw. Supplied operations are
excluded. HRC-R15's unresolved monster mode-change elemental path may draw; KNB-Q06 gates the
complete event total. There is no universal zero-draw claim for full mode integration.

**KNB-R10b Missile order** [verified in 1.13c code (static)]. Consume SBA-R11 and MHD-R12:
missile chance precedes defender block/avoidance, later owner monster multiplication
(DPO-R22), hit-conditional direct application and the subsequent unconditional direct
reaction. Chance steps neither defender nor owner. Upstream hit checks are outside this
boundary. Block need not erase KB; reaction priority still chooses the operation. V24 supplies
later-stage operations to check ordering; MHD-R12 owns their real eligibility and call path.

**KNB-R10c Item dispatch integration** [unresolved]. Text selects `domeleedamage` and
`domissiledamage` with selector 7. Compiled registration and event-name numeric binding are
KNB-Q01/KNB-Q02. Text alone does not establish callback timing relative to block, crushing
blow or deadly strike. Supplied callback invocation is independently KNB-R01.

**KNB-R10d GH-veto order** [verified in 1.13c code (static)]. Queued reaction follows
application through DPO-R26; missile direct reaction follows its conditional application
through MHD-R12. Only GH selection reaches HRC-R06's zero to two defender draws, after the
earlier missile chance. KB selection has no GH-veto call. Item-callback timing remains KNB-Q02.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| KNB-R01 | Item callback | verified in 1.13c code (static) | none for supplied invocation | KNB-V01, KNB-V02, KNB-V03, KNB-V14, KNB-V15, KNB-V16, KNB-V31, KNB-V35 |
| KNB-R02a | Missile chance | verified in 1.13c code (static) | none | KNB-V04, KNB-V05, KNB-V06, KNB-V07, KNB-V17, KNB-V18 |
| KNB-R02b | Missile state gate | verified in 1.13c code (static) | none | KNB-V12 |
| KNB-R03b | Town gate | verified in 1.13c code (static) | none for supplied valid room/record | KNB-V27 |
| KNB-R05a | State-54 calls | verified in 1.13c code (static) | KNB-Q11: STL-Q07 disguise bookkeeping / RLI-Q02 update queue only | KNB-V12, KNB-V27(k) |
| KNB-R03a | Kind/mode gate | verified in 1.13c code (static) | none | KNB-V26 |
| KNB-R06a | Fallback | verified in 1.13c code (static) | none | KNB-V10, KNB-V19, KNB-V32 |
| KNB-R06c | Base conversion | verified in 1.13c code (static) | none | KNB-V09, KNB-V20, KNB-V32 |
| KNB-R04 | Death vs KB | verified in 1.13c code (static) | downstream death effects outside selection | KNB-V11 |
| KNB-R06b | Priority | verified in 1.13c code (static) | HRC-R04b/HRC-R11: full player effects only | KNB-V19, KNB-V28 |
| KNB-R06d | Monster start | verified in 1.13c code (static) | KNB-Q06/KNB-Q07: dependent effects only | KNB-V13, KNB-V21, KNB-V33 |
| KNB-R07a | Distance parameter | verified in 1.13c code (static) | none | KNB-V08, KNB-V09, KNB-V22 |
| KNB-R08 | Endpoint arithmetic | verified in 1.13c code (static) | none for supplied coordinates | KNB-V08, KNB-V23, KNB-V29, KNB-V34, KNB-V36 |
| KNB-R09a | Path writes | verified in 1.13c code (static) | KNB-Q07: actual clipping/motion only | KNB-V13, KNB-V23, KNB-V30, KNB-V34, KNB-V36 |
| KNB-R09b | Collision/motion | unresolved | KNB-Q07 | none |
| KNB-R10a | Local draws | verified in 1.13c code (static) | KNB-Q06: full draw count only | KNB-V14, KNB-V16, KNB-V17, KNB-V21 |
| KNB-R10b | Missile order | verified in 1.13c code (static) | none for supplied later stages | KNB-V24 |
| KNB-R10c | Item timing | unresolved | KNB-Q01/KNB-Q02 | none |
| KNB-R10d | GH-veto order | verified in 1.13c code (static) | none for supplied GH effect | KNB-V25 |

`KNB-R02` denotes the KNB-R02a/KNB-R02b family.

## 5. Outputs and side effects

| Write or effect | Rule |
|---|---|
| Attacker seed step, KB OR, callback return 0/1 | KNB-R01 |
| Missile seed step, KB OR | KNB-R02a |
| Continue or stop reaction at the town gate | KNB-R03b |
| `SetState(target,92,1)` request, with membership effects owned by RLI-R03/STL-R07 | KNB-R05a |
| Replace KB by GH, or add KB to GH | KNB-R06a/KNB-R06c |
| Select/request one reaction operation | KNB-R04/KNB-R06b |
| Path type/distance; ordered planner/preparation/mode calls; return 0/1 | KNB-R06d/KNB-R07a |
| Requested waypoint, clip call, planned endpoint, reference removal, conditional count/waypoint write, return 0/1 | KNB-R09a |
| Ordered later-stage/GH-veto invocation | KNB-R10b/KNB-R10d |

HRC-R14 supplies downstream knockback animation rate and path speed 4096. Actual motion,
dependent state/entry/mode effects, messages and persistence are not additional local writes.
V13c's modeChanged/update assertions use the supplied RLI-R04a model; actual update
bookkeeping and queue effects retain RLI-Q02's integration boundary.

## 6. Branches and edge cases

Size affects item chance only; player threshold is 64. Failed chance checks preserve result.
Fallback precedes base conversion, which can re-add KB even without `mKB`. Freeze prevents
conversion; stun affects only the GH branch. No difficulty/PvP branch exists in the local
specified operations; upstream eligibility and damage remain supplied by their owners.
Only LoD Expansion 1.13c is established. Static source coordinates preserve i32 width; endpoint
addition wraps at 16 bits. A zero-length clipped plan neither moves the unit nor clears an
existing count. No local subtype test proves full hireling/summon/NPC equivalence.

## 7. Open questions

Attempts and exact blocked assertions are in the [gap file](../plans/gaps/combat-knockback.md#open-evidence-gaps).

- **KNB-Q01** Integration: establish compiled stat/event-selector correspondence to the named
  text cells through the matching compiled record and loader mapping.
- **KNB-Q02** Integration: bind text event names to numeric dispatch, registration and timing
  relative to block/crushing blow/deadly strike. IEP-R09a/DPO-R13 supply numeric sites only.
- **KNB-Q04** Observation: measure completed displacement; NRF-R10 supplies coordinate scale,
  not a knockback measurement. Close with KNB-O03.
- **KNB-Q06** Integration: complete mode-dispatch refusal consequences for `mKB` without
  `mGH`, mode entry preparation and downstream state/RNG effects including HRC-R15. Handler return 0
  is verified independently.
- **KNB-Q07** Behavior/integration: finish collision endpoint selection, waypoint initialization,
  player mode-19 setup/gates and later actual movement. The player helper reaches type-8 setup;
  full caller/planner/motion reading remains necessary.
- **KNB-Q08** Outside-boundary: establish complete subtype producers/records/mode paths for
  hirelings, summons and NPCs; local lack of a subtype branch proves only local routing.
- **KNB-Q10** Outside-boundary: resolve simultaneous movement conflicts and client correction
  in movement/network owners. No complete client search supports absence.
- **KNB-Q11** Integration: only state-92 disguise bookkeeping and update-queue effects remain
  STL-R07a/STL-Q07 and RLI-Q02. SetState routing/membership is owned by RLI-R03/STL-R07;
  ordinary death-mode execution is not this request.
- **KNB-Q12** Observation: gameplay confirmation remains pending for all behavioral rules;
  KNB-O01–KNB-O06 are procedures, not observations.

## 8. Test vectors

### 8.1 Executable vectors

All vectors have `runtime_observed: no`. Each case starts independently. Defaults: result 0;
attacker/missile seed `{1,0}`, defender `{9,0}`; item owner/target/record present, supplied stat
81 value 1, monster size record present with neither flag; missile state 54 absent, chance 0.
Reaction defaults: monster mode 1, non-town room, valid base 0, both availability flags true,
no state 1/54 or avoidance/block. Path defaults: start and actual position `(100,100)`, dynamic
player source `(93,100)` present, `B=5`, count 0, planned endpoint `(7,7)`.

Supplied operations persist only stated effects and log calls in order: clip returns its stated
endpoint; planner returns its stated value with no writes; prepareModeEntry logs only; setMode stores
the requested mode and returns 1. Other reaction operations log only and return 1. All draw
nothing unless stated. A GH-veto operation logs and performs only stated draws. This fixture
does not verify these dependencies. Unlisted seeds and fields remain unchanged.

- **KNB-V01** Small item (KNB-R01; boundary: isolated rule)
  - Inputs: item defaults, small true.
  - Seeds: attacker `{1,0}`.
  - Draws: `(attacker,mask 128,69)`.
  - Trace: low `0x6AC690C5`; `69 < 128`.
  - Outputs: result 8, return 1, seed `{0x6AC690C5,0}`; threshold accepts 128/128 outcomes.
  - Discriminates: threshold 64 would reject this roll.
- **KNB-V02** Normal item (KNB-R01; boundary: isolated rule)
  - Inputs: independent item defaults with player target in a/b.
  - Seeds: a attacker `{1,0}`; b attacker `{0,48}`.
  - Draws: a `(attacker,mask 128,69)`; b `(attacker,mask 128,48)`.
  - Trace: a `69 >= 64`; b new low `0 * multiplier + 48 = 48`, and `48 < 64`.
  - Outputs: a result 0/return 0/seed `{0x6AC690C5,0}`; b result 8/return 1/seed `{48,0}`; threshold accepts 64/128 outcomes.
  - Discriminates: guaranteed positive-stat KB wrongly accepts a; threshold 32 wrongly rejects b.
- **KNB-V03** Large precedence (KNB-R01; boundary: isolated rule)
  - Inputs: independent item defaults with both size flags true in a/b.
  - Seeds: a attacker `{1,0}`; b attacker `{0,48}`.
  - Draws: a `(attacker,mask 128,69)`; b `(attacker,mask 128,48)`.
  - Trace: large chooses 32; a `69 >= 32`; b new low 48, `48 >= 32`.
  - Outputs: a result 0/return 0/seed `{0x6AC690C5,0}`; b result 0/return 0/seed `{48,0}`; threshold accepts 32/128 outcomes.
  - Discriminates: small-first would set 8 in both cases; a wrong default threshold 64 would set 8 in b.
- **KNB-V04** Missile 33 (KNB-R02a; boundary: isolated rule)
  - Inputs: missile defaults, chance 33.
  - Seeds: missile `{2,0}`.
  - Draws: `(missile,100,70)`.
  - Trace: low `0xD58D218A` mod 100 is 70; not below 33.
  - Outputs: result 0, seed `{0xD58D218A,0}`; 33/100 remainder values succeed.
  - Discriminates: guaranteed-on-positive-byte would set 8.
- **KNB-V05** Missile 75 (KNB-R02a; boundary: isolated rule)
  - Inputs: missile defaults, chance 75.
  - Seeds: missile `{2,0}`.
  - Draws: `(missile,100,70)`.
  - Trace: `70 < 75`.
  - Outputs: result 8, seed `{0xD58D218A,0}`; 75/100 remainder values succeed.
  - Discriminates: treating chance as a boolean misses V04/V05's differing results on the same roll.
- **KNB-V06** Missile 1 (KNB-R02a; boundary: isolated rule)
  - Inputs: independent missile defaults with chance 1 in a/b.
  - Seeds: (a) missile `{0,0}`; (b) missile `{1,15}`.
  - Draws: each `(missile,100,0)`.
  - Trace: a new low 0; b new low `0x6AC690C5 + 15 = 0x6AC690D4`, remainder 0; each `0 < 1`.
  - Outputs: a result 8/seed `{0,0}`; b result 8/seed `{0x6AC690D4,0}`; 1/100 remainder values succeed.
  - Discriminates: rejecting byte 1 as too small fails; b also rejects omitting the RNG step, which a's absorbing seed alone cannot distinguish.
- **KNB-V07** Zero byte (KNB-R02a/KNB-R10a; boundary: isolated rule)
  - Inputs: independent missile defaults with supplied compiled chance byte 0; a initial result 4, b initial result 12.
  - Seeds: missile `{1,0}` in each case; no draw.
  - Draws: none.
  - Trace: zero-byte guard skips without writing result.
  - Outputs: a result 4; b result 12; each seed `{1,0}`.
  - Discriminates: drawing before guard changes seed; clearing existing KB would change b's result to 4.
- **KNB-V08** Default endpoint (KNB-R07a/KNB-R08/KNB-R09a; boundary: supplied-input stage)
  - Inputs: path defaults; clip returns requested `(103,100)`.
  - Seeds: no draw.
  - Draws: none.
  - Trace: B 5 -> S 3; v `(7,0)`, m 7, offset `(3,0)`.
  - Outputs: requested/planned/waypoint `(103,100)`, count 1, reference absent, return 1; actual `(100,100)`.
  - Discriminates: using byte as step gives `(105,100)`.
- **KNB-V09** Base exception (KNB-R06c/KNB-R07a; boundary: supplied-input stage)
  - Inputs: reaction defaults, base 78, result 4; KB selection explicitly invokes the monster start stage. In this case only, planner explicitly invokes KNB-R08/KNB-R09a on path defaults with the distance byte just stored by start; clip returns requested `(106,100)` and generator return is planner return. This replaces the default planner's no-write effect.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unfrozen conversion gives 12; start chooses B 10 -> S 6.
  - Outputs: result 12, selected mode 13, requested endpoint `(106,100)`, supplied setMode stores 13.
  - Discriminates: default byte would request `(103,100)`.
- **KNB-V10** Fallback (KNB-R06a; boundary: supplied-input stage)
  - Inputs: reaction defaults, mKB false, result 8; GH operation log-only.
  - Seeds: no draw.
  - Draws: none.
  - Trace: `(8 & ~8)|4 = 4`, reaches GH.
  - Outputs: result 4, log `[GH]`, no KB call or actual displacement.
  - Discriminates: clearing KB without adding GH loses the call.
- **KNB-V11** Lethal priority (KNB-R04; boundary: supplied-input stage)
  - Inputs: (a,b) independent monster and player reaction defaults, result 10; (c,d) independent monster and player reaction defaults, result 26 (lethal+KB+block); no avoidance; selected operations log-only.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a/b lethal precedes KB; c monster lethal precedes KB/block; d player block precedes lethal/KB.
  - Outputs: a/b result 10, log `[death]`; c result 26, log `[death]`; d result 26, log `[block]`; no KB call/mode write/displacement in any case.
  - Discriminates: KB-first logs KB; death-first for both kinds wrongly logs death in d.
- **KNB-V12** State gates (KNB-R02b/KNB-R05a; boundary: supplied-input stage)
  - Inputs: (a) missile chance 255, state 54, result 4; (b) reaction defaults/state 54/result 8; (c) reaction defaults/state 54/result 10; (d) missile chance 255, state 54, result 12, with the missile chance stage only.
  - Seeds: no draw; missile `{1,0}` unchanged.
  - Draws: none.
  - Trace: a/d skip chance without clearing result; b returns; c requests lethal state then returns.
  - Outputs: a result 4/log `[]`; b result 8/log `[]`; c result 10/log `[SetState(target,92,1)]`; d result 12/log `[]` with missile seed `{1,0}` unchanged; no KB/GH/death-mode call in any case. The supplied SetState operation logs only; membership effects are tested by its owning contract.
  - Discriminates: ordinary death substitution changes c's log; clearing KB in the missile state gate changes d's result to 4.
- **KNB-V13** Ignored dependency returns (KNB-R06d/KNB-R09a; boundary: supplied-input stage)
  - Inputs: (a) monster start defaults, planner returns 0; (b) path defaults, clip returns start `(100,100)`, count 0; (c) monster start defaults except current mode 13/modeChanged false, initial path type 2/B 77, planner returns 0. In c only, setMode logs its invocation then executes RLI-R04a's actual equal-mode branch with a supplied MarkForUpdate operation that logs current mode/modeChanged and has no queue effects; it replaces the default return-1 setter. List processing and animation-reset operations log if called.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a ignores planner return; b clipped equals start; c equal mode 13 invokes update marking, stores modeChanged true and returns 0 from the setter, then the start handler returns 1.
  - Outputs: a log `[setPathType(8),setDistance(5),planner(target,0),prepareModeEntry,setMode(13)]`, mode 13, return 1; b planned `(100,100)`, stored requested waypoint `(103,100)`, count 0, reference absent, return 0, actual `(100,100)`; c log `[setPathType(8),setDistance(5),planner(target,0),prepareModeEntry,setMode(13),MarkForUpdate(mode=13,modeChanged=false)]`, path type 8/B 5/mode 13, supplied-model modeChanged true, setter return 0 but handler return 1, no list-processing or animation-reset call. Actual update effects remain RLI-Q02.
  - Discriminates: aborting start on zero planner omits mode call; forwarding the equal-mode setter's 0 as handler result fails c. Completed blocked motion is gated.
- **KNB-V14** Seed owner (KNB-R01/KNB-R10a; boundary: isolated rule)
  - Inputs: independent small-only, normal, large-only item cases.
  - Seeds: attacker `{1,0}`, defender `{9,0}` each.
  - Draws: each `(attacker,mask 128,69)`.
  - Trace: 69 below 128, not below 64/32.
  - Outputs: respective result/return `(8,1)`, `(0,0)`, `(0,0)`; each attacker `{0x6AC690C5,0}`, defender `{9,0}`.
  - Discriminates: defender-seed roll changes wrong seed.
- **KNB-V15** Positive magnitude (KNB-R01; boundary: isolated rule)
  - Inputs: (a–c) independent small-only, normal, large-only item cases, stat 1; (d) normal item case, stat 100.
  - Seeds: attacker `{2,0}` in a–c; `{1,0}` in d.
  - Draws: a–c each `(attacker,mask 128,10)`; d `(attacker,mask 128,69)`.
  - Trace: a–c low `0xD58D218A`, 10 below every threshold; d low `0x6AC690C5`, 69 not below 64 despite stat 100.
  - Outputs: a–c each result 8/return 1/seed `{0xD58D218A,0}`; d result 0/return 0/seed `{0x6AC690C5,0}`.
  - Discriminates: scaling threshold by stat would wrongly accept d.
- **KNB-V16** Nonpositive stat (KNB-R01/KNB-R10a; boundary: isolated rule)
  - Inputs: independent item stat 0 and -1; result 4.
  - Seeds: attacker `{1,0}`; no draw.
  - Draws: none.
  - Trace: signed nonpositive guard.
  - Outputs: each result 4, return 0, seed `{1,0}`.
  - Discriminates: nonzero-only guard draws for -1.
- **KNB-V17** Missile strict boundary (KNB-R02a/KNB-R10a; boundary: isolated rule)
  - Inputs: independent missile chances 85/86/100/255.
  - Seeds: missile `{1,0}` each.
  - Draws: each `(missile,100,85)`.
  - Trace: low `0x6AC690C5` mod 100 = 85.
  - Outputs: respective result 0/8/8/8; each seed `{0x6AC690C5,0}`.
  - Discriminates: inclusive comparison wrongly accepts 85.
- **KNB-V18** Second missile boundary (KNB-R02a; boundary: isolated rule)
  - Inputs: independent chances 33/34.
  - Seeds: missile `{5,0}` each.
  - Draws: each `(missile,100,33)`.
  - Trace: product low `0x15E0D3D9`, high 2; low mod 100 = 33.
  - Outputs: chance 33 result 0; chance 34 result 8; each seed `{0x15E0D3D9,2}`.
  - Discriminates: inclusive comparison wrongly accepts 33.
- **KNB-V19** KB vs GH (KNB-R06a/KNB-R06b; boundary: supplied-input stage)
  - Inputs: (a) reaction defaults/result 12; (b) reaction defaults/mKB false/result 8.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) KB before GH; (b) fallback gives 4.
  - Outputs: (a) result 12/log `[KB(13)]`; (b) result 4/log `[GH]`.
  - Discriminates: invoking both GH and KB changes (a)'s log.
- **KNB-V20** Freeze conversion veto (KNB-R06c; boundary: supplied-input stage)
  - Inputs: independent reaction cases base 78/result 4; state 1 absent/present; GH log-only.
  - Seeds: no draw.
  - Draws: none.
  - Trace: absent freeze adds 8; present freeze leaves 4.
  - Outputs: absent result 12/log `[KB(13)]`; present result 4/log `[GH]`.
  - Discriminates: testing stunned instead of freeze fails present case.
- **KNB-V21** Start refusal (KNB-R06d/KNB-R10a; boundary: supplied-input stage)
  - Inputs: independent start defaults with (a) mGH false/mKB true, (b) mGH true/mKB false; each mode 1, path type 2, B 77, count 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a fails first availability check; b passes first but fails second.
  - Outputs: each return 0, log `[]`, mode 1/type 2/B 77/count 0 unchanged.
  - Discriminates: mKB-only guard proceeds in a; mGH-only guard proceeds in b.
- **KNB-V22** Distance decoding (KNB-R07a; boundary: isolated rule)
  - Inputs: independent bytes 5/10.
  - Seeds: no draw.
  - Draws: none.
  - Trace: `(5>>1)+1=3`; `(10>>1)+1=6`.
  - Outputs: respective S 3/6.
  - Discriminates: rounding up before adding gives 4 for byte 5.
- **KNB-V23** Direction/truncation (KNB-R08/KNB-R09a; boundary: supplied-input stage)
  - Inputs: independent path cases (a) source `(93,102)`; (b) source `(102,93)`; (c) start/actual `(50,50)`, source `(54,54)`, B 10; (d) source `(100,100)`; defaults otherwise, clip returns requested, count 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) v `(7,-2)` -> offset `(3,0)`; (b) `(-2,7)` -> `(0,3)`; (c) `(-4,-4)` -> `(-6,-6)`; (d) m 0 -> `(0,0)`.
  - Outputs: requested/planned (a) `(103,100)`, (b) `(100,103)`, (c) `(44,44)`, (d) `(100,100)`; all reference absent; (a–c) count 1/waypoint endpoint/return 1; (d) count 0/waypoint `(100,100)`/return 0. Actual positions retain starts.
  - Discriminates: floor division yields -1 on minor components in (a)/(b).
- **KNB-V24** Missile stage order (KNB-R10b; boundary: supplied-input stage)
  - Inputs: result initially 1 after MHD-R12's hit preset, chance 86; later block/avoidance/multiplier/apply/reaction log-only, each records missile seed/result.
  - Seeds: missile `{1,0}`, defender `{9,0}`, owner `{2,0}`.
  - Draws: `(missile,100,85)`.
  - Trace: chance ORs 8 into the preset hit result 1 before all supplied stages; hit remains set for conditional apply.
  - Outputs: log `[block,avoidance,multiplier,apply,reaction]`; each sees result 9/missile `{0x6AC690C5,0}`; defender `{9,0}`, owner `{2,0}` unchanged.
  - Discriminates: chance after block changes block's recorded seed/result.
- **KNB-V25** Conditional GH veto (KNB-R10d; boundary: supplied-input stage)
  - Inputs: independent missile chance (a) 86, (b) 85, (c) 85; result initially 5 (MHD-R12's preset hit 1 plus GH 4); supplied apply/reaction, GH/KB log-only. GH veto returns 0 with no draw in a/b; in c it explicitly draws once through NRF-R02 RandMask(defender,4), logs that draw, then returns 0. The supplied c operation does not test HRC-R06's own threshold selection.
  - Seeds: missile `{1,0}` each; defender `{9,0}` in a/b, `{1,0}` in c.
  - Draws: a/b each `(missile,100,85)`; c in order `(missile,100,85)`, `(defender,mask 4,1)`.
  - Trace: a result 13 selects KB; b/c result 5 reaches supplied GH-veto after apply; hit 1 remains set for conditional apply in every case; c defender low `0x6AC690C5 & 3 = 1`.
  - Outputs: a result 13/log `[apply,KB(13)]`; b result 5/log `[apply,GH-veto,GH]`; c result 5/log `[apply,GH-veto,defenderDraw(mask4,1),GH]`; every missile `{0x6AC690C5,0}`; defender remains `{9,0}` in a/b and becomes `{0x6AC690C5,0}` in c.
  - Discriminates: unconditional GH veto adds a call in a; stepping defender during missile chance changes c's ordered draw/log assertions.
- **KNB-V26** Type/dead-mode gates (KNB-R03a; boundary: supplied-input stage)
  - Inputs: independent null target, type-3 target, monster mode 0, monster mode 12; result 8, earlier gates passed.
  - Seeds: no draw.
  - Draws: none.
  - Trace: null/type or dead-mode return.
  - Outputs: each result 8/log `[]`, no path/mode write.
  - Discriminates: every-non-player-is-monster routing calls KB for type 3.
- **KNB-V27** Town gate (KNB-R03b/KNB-R05a; boundary: supplied-input stage)
  - Inputs: gate-only cases (a) room level 1/player source; (b) 40/monster inTown false; (c) 75/monster inTown true; (d) 2/player; (e) no room/player; (f) 0/player; (g) 109/player; (h) 110/player; (i) level 1/monster whose MonStats lookup supplies no valid record, with a detached supplied inTown flag true; (j) level 1/source absent. Valid records are present in b/c; i's detached flag is not a valid record. Case k instead composes reaction entry through town and state checks: reaction defaults, room level 1, source absent, target state 54 present, result 10, mode 1, path type 2/B 5/count 0/planned `(7,7)`/actual `(100,100)`; supplied SetState and reaction/path operations log calls only.
  - Seeds: no draw.
  - Draws: none.
  - Trace: set membership true for a/b/c/g/i/j/k; permitted monster and present valid record required only there. In k the town rejection occurs before state-54 handling, so the lethal state request is never reached.
  - Outputs: continuation false a/b/g/i/j/k, true c/d/e/f/h; a–j no result/path/mode write in gate-only stage; k result 10, log `[]`, no SetState(92,1) request, mode 1/path type 2/B 5/count 0/planned `(7,7)`/actual `(100,100)` unchanged, and all seeds unchanged.
  - Discriminates: treating whole bounded range as towns rejects d; checking only the detached inTown flag wrongly permits i; missing source cannot satisfy permission in j; executing the lethal state-54 branch before town rejection adds SetState(92,1) to k's log.
- **KNB-V28** Player block vs monster KB (KNB-R06b; boundary: supplied-input stage)
  - Inputs: (a,b) independent player/monster reaction defaults, result 24, no avoidance; (c) player with result 152 (avoidance 128 + block 16 + KB 8). Selected terminal operations log only; c supplies an avoidance selection operation without asserting its gated skill/start effects.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a player block precedes KB; b monster KB precedes block; c player avoidance precedes both.
  - Outputs: a log `[block]`, b log `[KB(13)]`, c log `[avoidance]`; results remain 24, 24 and 152 respectively, with no mode/motion write in the supplied selection stage.
  - Discriminates: one shared priority cannot produce a/b; swapping player avoidance behind block or KB changes c's selected operation.
- **KNB-V29** Static width (KNB-R08; boundary: supplied-input stage)
  - Inputs: path defaults, source kind 2/full i32 position `(65536,100)`; clip returns requested.
  - Seeds: no draw.
  - Draws: none.
  - Trace: v `(-65436,0)`, m 65436, offset `(-3,0)`.
  - Outputs: requested/planned/waypoint `(97,100)`, count 1, reference absent, return 1, actual `(100,100)`.
  - Discriminates: source u16 truncation wrongly requests `(103,100)`.
- **KNB-V30** Existing count/explicit point (KNB-R09a/KNB-R08; boundary: supplied-input stage)
  - Inputs: (a) path defaults/count 2/clip returns start; (b) path defaults/count 2/reference absent/explicit `(120,120)`/clip returns `(101,102)`.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) requested stored before clip; (b) explicit bypasses scaling.
  - Outputs: (a) planned `(100,100)`, first waypoint `(103,100)`, count 2, return 0; (b) clip call `(start=(100,100),requested=(120,120))`, planned/first waypoint `(101,102)`, count 1, return 1; both reference absent/actual `(100,100)`.
  - Discriminates: clear-on-zero fails a; scaling explicit destination changes b's clip request.
- **KNB-V31** Callback guards and absent size record (KNB-R01; boundary: isolated rule)
  - Inputs: item defaults except (a) owner absent, (b) target absent, (c) damage record absent, (d) monster size record absent. Each present damage record initially has result 4; stat 1 where readable. The stat-read operation logs calls, returns 1 and draws nothing if reached.
  - Seeds: attacker `{1,0}` in every case; no draw in a/b/c.
  - Draws: a/b/c none; d `(attacker,mask 128,69)`.
  - Trace: a/b/c return before stat read; d stat read succeeds, absent size record selects threshold 64 and 69 fails.
  - Outputs: a/b return 0, result 4, stat-read log `[]`; c return 0, no record write, stat-read log `[]`; d return 0, result 4, stat-read log `[readStat(81)]`; seed `{1,0}` in a/b/c, `{0x6AC690C5,0}` in d.
  - Discriminates: reading stat before reference guards changes a/b/c logs; treating absent size record as an early return misses d's draw.
- **KNB-V32** Fallback retains flags and precedes conversion (KNB-R06a/KNB-R06c; boundary: supplied-input stage)
  - Inputs: reaction defaults with mKB false in (a,b); a result 24/base 0; b result 8/base 78; c mKB true/result 4/invalid base lookup. Selected operations log-only.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a fallback yields 20 (block+GH), then block wins; b fallback gives 4, conversion re-adds 8, KB wins; c invalid lookup leaves 4.
  - Outputs: a result 20/log `[block]`; b result 12/log `[KB(13)]`; c result 4/log `[GH]`; no mode/motion write in these selection-only cases.
  - Discriminates: replacing whole result with 4 loses a's block; conversion before fallback loses b's KB; converting invalid bases changes c.
- **KNB-V33** Start success without local writes (KNB-R06d; boundary: supplied-input stage)
  - Inputs: independent monster-start cases: (a) both availability flags true, target mode 0/path type 2/B 77/count 2/planned `(7,7)`; (b) mGH false/mKB true, otherwise as a; (c) both availability flags supplied true, target absent and no target path.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a availability checks pass, then mode-0 guard returns success; b availability refuses before the mode-0 guard; c availability passes, then missing-unit guard returns success.
  - Outputs: return 1, 0 and 1 respectively, each with log `[]` and no local path/mode write; a/b mode 0/type 2/B 77/count 2/planned `(7,7)` unchanged; c has no target path to mutate.
  - Discriminates: treating every no-op as failure changes a/c; testing mode 0 before availability changes b; omitting the missing-unit guard attempts path/mode writes in c.
- **KNB-V34** Endpoint integer widths (KNB-R08/KNB-R09a; boundary: supplied-input stage)
  - Inputs: path defaults except independent cases (a) start/actual `(65535,100)`, dynamic source `(65534,100)`; (b) static kind-2 source `(2147483647,100)`; (c) start/actual `(0,0)`, static kind-2 source `(-2147483648,0)`; each B 5/count 0, clip returns requested.
  - Seeds: no draw.
  - Draws: none.
  - Trace: a v `(1,0)`, offset `(3,0)`, low16 sum 2; b v `(-2147483547,0)`, m 2147483547, wrapping product -2147483345, quotient 0; c wrapping v/abs X -2147483648, signed max with 0 gives m 0, retained v adds to low16 `(0,0)`.
  - Outputs: a requested/planned/first waypoint `(2,100)`, count 1, return 1, actual `(65535,100)`; b requested/planned/first waypoint `(100,100)`, count 0, return 0, actual `(100,100)`; c requested/planned/first waypoint `(0,0)`, count 0, return 0, actual `(0,0)`; reference absent in all.
  - Discriminates: nonwrapping endpoints fail a; unbounded multiply gives b offset -3; unbounded absolute value/scaling changes c's endpoint.
- **KNB-V35** Successful OR and supplied stat identity (KNB-R01; boundary: isolated rule)
  - Inputs: independent item defaults with (a) supplied stat identity 81/value 1, result 16; (b) supplied synthetic stat identity 82/value 1, stat 81/value 0, result 4. Neither case has size flags. The stat-read operation logs the identity read and returns its explicitly supplied current value; identity 82 asserts no original-profile name or registration binding.
  - Seeds: attacker `{2,0}` each.
  - Draws: each `(attacker,mask 128,10)`.
  - Trace: 10 below default threshold 64; a ORs 8 into 16, b reads supplied identity 82 then ORs 8 into 4.
  - Outputs: a result 24/return 1/stat-read log `[readStat(81)]`; b result 12/return 1/stat-read log `[readStat(82)]`; each attacker `{0xD58D218A,0}`.
  - Discriminates: assignment instead of OR loses a's 16 or b's 4; hardcoding stat 81 rejects b because its supplied value is 0.
- **KNB-V36** Missing dynamic source path (KNB-R08/KNB-R09a; boundary: supplied-input stage)
  - Inputs: path defaults except reference unit kind 0 is present with no dynamic path; explicit destination `(7,7)` is supplied but unused while reference exists; clip returns requested `(103,103)`.
  - Seeds: no draw.
  - Draws: none.
  - Trace: missing source path supplies `(0,0)`; v `(100,100)`, m 100, B 5 -> S 3, offset `(3,3)`.
  - Outputs: requested/planned/first waypoint `(103,103)`, count 1, reference absent, return 1, actual `(100,100)`.
  - Discriminates: rejecting a source without a path would return 0; treating it as absent reference would use an explicit destination instead of the zero-coordinate source.

### 8.2 Observation vectors

All are pending. [Reproduction procedures](../research/spec-evidence/combat-knockback.md#pending-observation-procedures)
state controls, actions and readings; no gameplay observation occurred in this regeneration.

| Id | Measure | Expected assertion |
|---|---|---|
| KNB-O01 | Item roll by controlled size | Thresholds 128/64/32; distinguish accepted draw from motion |
| KNB-O02 | Missile bytes 1/33/75 | Remainder comparison; one eligible missile draw |
| KNB-O03 | Default/base-78 endpoint and actual motion | Constructed step 3/6; separately measure plan/completed travel |
| KNB-O04 | Collision behind target | Compare clipped plan/count/mode/motion; KNB-Q07 remains open |
| KNB-O05 | State 54 and GH veto | State gate blocks KB; GH veto only if GH selected |
| KNB-O06 | Hireling/summon/NPC/dead target | Local kind/mode routing; complete subtype equivalence KNB-Q08 |

## 9. Acceptance and dependencies

Partial static acceptance requires KNB-V01–KNB-V36's exact assertions at their stated isolated
or supplied-input boundaries. The verified Id-map rows are available at those boundaries;
unresolved KNB-R09b/KNB-R10c and dependent effects are not invented for a partial pass.
Full reproduction also needs compiled item binding/timing, mode/player/collision/motion
integration and matching gameplay observations for in-scope rules. A stage pass or text-cell
read establishes neither complete parity nor runtime confirmation.

**Depends on:** NRF-R01/NRF-R02 seed/draw; NRF-R10 coordinates; RLI-R03 states and Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).
RLI-R02a/RLI-R02b selected-array/minimum stat lookup;
RLI-R04a mode write; GDA-R08/GDA-R09/GDA-R10/GDF-R03/GDF-R05/GDF-R06/GDF-R08/GDA-R16 names;
SBA-R11 missile chance position; MHD-R12 missile owner-multiply/direct-apply/direct-reaction
path; DPO-R22 owner monster multiplication; DPO-R13/DPO-R26 event and queued-reaction order;
IEP-R02/IEP-R03/IEP-R06/IEP-R07/IEP-R09a selector, invocation and numeric dispatch, with
IEP-R04 integration unresolved; HRC-R04a/HRC-R05/HRC-R06/HRC-R14 priority, GH and speed/rate.
HRC-R04b/HRC-R11/HRC-R15 retain their full avoidance/block/mode-effect limits.
State-on routing/membership consumes RLI-R03 and
[state-timed-state-lifecycle](state-timed-state-lifecycle.md) STL-R07; its disguise/update
effects retain STL-R07a/STL-Q07 and RLI-Q02 integration limits.

**Depended on by:** SBA-R11, HRC-R01b/HRC-R04a, missile-hit-and-damage-calculation's handoff,
and numeric foundations' seed-owner references. Supporting citations add no prerequisite.

## 10. Related specs

- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): damage/event stages.
- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md): selector and registration.
- [combat-hit-recovery](combat-hit-recovery.md): priority, GH veto and downstream animation speed.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): checks following missile chance.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): freeze and shared movement.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): state-on membership behind the uninterruptible lethal branch.
- [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md): source/result handoff.

## Appendix A. Provenance

Evidence is LoD Expansion 1.13c Windows x86, artifacts `LOD113C-D2GAME` and `LOD113C-D2COMMON`
in the [source manifest](../research/source-manifest.md). Compatible 2026-10-11/12 static
readings retain their dates and limits. Targeted 2026-10-15 disassembly and original-image
table/encoding checks support the closures/corrections in the [public companion](../research/spec-evidence/combat-knockback.md).
Named ItemStatCost and States text cells were read through the game's loader on 2026-10-15 with the
recorded profile limits, without compiled-table or gameplay confirmation. No behavioral
vector is runtime-confirmed; status remains draft.
