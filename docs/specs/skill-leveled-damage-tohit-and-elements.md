# Skill Leveled Damage, To-Hit and Elements

**Description**

Skills can gain damage and accuracy as their level rises. These shared calculations turn a skill's damage range, bonuses, and duration into the damage used by a cast or attack. They also select the damage kind and draw a value from its range before later combat calculations apply.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: SLD-R, SLD-Q, SLD-V, SLD-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md)
Open evidence gaps: [gap file](../plans/gaps/skill-leveled-damage-tohit-and-elements.md)

## 1. Scope

This contract covers shared skill-level physical/elemental endpoints, elemental duration, the skill to-hit bonus, and the server helpers that roll and route damage into a combat record. It includes level bands, source-damage scaling with supplied equipment results, synergy/mastery order, and the caller sequence for `Skills.srvdofunc` 38 and 53. Calculation-language execution belongs to [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md); Kick damage production, equipment selection, other callers, and skill-specific extra terms are outside this boundary. Later damage processing belongs to [combat-damage-pipeline-order](combat-damage-pipeline-order.md), hit resolution to [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md), and duration application to [combat-status-application](combat-status-application.md).

## 2. Trigger and ownership

D2Common computes the shared endpoints, duration, and to-hit value. D2Game rolls the endpoints and writes the combat record during skill execution, before area damage or the subsequent damage pipeline. Server do-functions 38 and 53 call physical fill before elemental fill; do-function 46 reaches the to-hit hand-off and elemental fill. These are calls within an execution, not independently scheduled per-frame operations. Dispatch and output anchors are in the [companion](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md#dispatch-and-functions). Client display and messages are outside this contract.

## 3. Inputs and data

All arithmetic is signed i32 with low-32-bit wrapping unless a narrower or unsigned value is named. Signed division truncates toward zero. `mulDiv(percent,value,100)` consumes NRF-R05 in that argument order; source division consumes NRF-R06a. Damage uses raw 8.8 units after the shift when `HitShift = 8` (NRF-R08a); durations use frames. Semantic record fields consume foundations section 3.2.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| skill identity, available record count, record presence | selected Skills entry | signed i32 identities/count; presence condition | Lookup and rejection |
| level `L` | caller's level for this skill | i32 | Progression input |
| `Kick` | Skills flag | condition | Physical override request |
| `MinDam`, `MaxDam` | Skills columns | i32 | Physical bases |
| `MinLevDam1` through `MinLevDam5`, `MaxLevDam1` through `MaxLevDam5` | Skills columns | five i32 increments per endpoint | Physical level bands |
| `SrcDam` | Skills column | u8; ratio in 1/128 | Source contribution |
| include-source condition | physical endpoint caller | zero/nonzero | Enables source contribution |
| selected weapon, source classification | equipment operations on caster | optional item; i32 classification | Supplied results; Q02 limits integration |
| caster stats 21 `mindamage`, 22 `maxdamage`; item stats 23 `secondary_mindamage`, 24 `secondary_maxdamage` | layer-zero stat reads | i32 | Source endpoints |
| supplied Kick result | Kick producer, selected minimum/maximum | i32 | Returned directly by override |
| `HitShift` | Skills column | u8; effective shift count masked by 31 | Left shift of damage |
| `DmgSymPerCalc`, `EDmgSymPerCalc`, `ELenSymPerCalc`, `ToHitCalc` | Skills calculation references | u32 bit pattern; `0xFFFFFFFF` means empty | Optional calculation |
| calculation pool and size; evaluation operation | selected skill pool; context `(unit,skill,L)` | optional bytes, u32 size, i32 result | Calculation dependency |
| `EType` | Skills linked element selector | u8 | Engine-defined routing/mastery selection |
| `EMin`, `EMax`, `EMinLev1` through `EMinLev5`, `EMaxLev1` through `EMaxLev5` | Skills columns | i32 bases and five increments | Elemental level bands |
| `ELen`, `ELevLen1` through `ELevLen3` | Skills columns | i32 frames | Elemental duration progression |
| mastery stats 329 `passive_fire_mastery`, 330 `passive_ltng_mastery`, 331 `passive_cold_mastery`, 332 | caster at layer zero | i32 | Percentage inputs |
| to-hit addend, stat 325 | caster at layer zero | i32 | Added on do-function 46 path |
| `ToHit`, `LevToHit` | Skills columns | i32 | Linear to-hit terms |
| unit presence, stat list/definitions and selected arrays | NRF-R12b / NRF-R12c | semantic stat inputs | Mastery/source getters |
| caster seed | unit supplied to the fill/router | u32 low/high pair | Range and random-element draws |
| combat record and optional outputs | caller | foundations section 3.2; optional i32 outputs | Additive damage, replaced durations/hit class, resist and selected-element results |
| target and combat context | to-hit hand-off caller | opaque dependency inputs | Forwarded with the caster and computed bonus |
| `skillHitClassValue` | selected Skills record; neutral name, column identity SLD-Q10 | i32 | Do-function 53 hit-class replacement when nonzero |

| Mapping or constant | Origin | Binding |
|---|---|---|
| Level bands, shift semantics, divisors 128/100, routing selectors, random-element choices/default duration | engine code | Fixed by original code |
| Mastery selector-to-stat mapping; source stat identities; to-hit addend 325 | engine code | Fixed by original code; names consume GDF-R11 |
| Skills fields and linked `EType` | table data | Values resolved from the selected Skills record; selector meaning fixed by handlers |
| Stat 325 text name `progressive_tohit`; stat 332 text name `passive_pois_mastery` | table data | Named effective text cells only; [record and profile limits](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md#named-table-cells) |
| ElemTypes text names and codes | table data | Named text cells; compiled index correspondence remains Q07 |

## 4. Rules (execution order)

The physical path takes the override before ordinary progression; the numbered progression rules describe the ordinary path. Calculation hooks can themselves draw under SCX-R15. A stated `no draw` below describes this rule's arithmetic, and does not erase draws performed by a requested calculation.

**SLD-R01a** Skill and level rejection [verified in 1.13c code (static)]. A negative skill identity, an identity at least the signed record count, or an absent selected record fails lookup. Elemental endpoints, duration and to-hit return 0; physical minimum returns 1; fills leave the record and seed unchanged, and elemental fill returns 0. Elemental endpoints, duration and to-hit also return 0 for signed `L <= 0`. Physical endpoints have no nonpositive-level rejection: absent an override, progression at `L <= 1` is the level-1 value. No calculation or fill draw follows rejection.

**SLD-R01b** Invalid physical maximum [verified in 1.13c code (static)]. Failed skill lookup returns the unshifted integer 2, with no downstream operation or draw.

**SLD-R02** Five level bands [verified in 1.13c code (static)]. For increments `a[0..4]`, the progression `bp` is zero for an absent array or signed `L <= 1`. Otherwise, using wrapping i32 products/additions:

```
L <= 8:   (L-1)*a[0]
L <= 16:  7*a[0] + (L-8)*a[1]
L <= 22:  7*a[0] + 8*a[1] + (L-16)*a[2]
L <= 28:  7*a[0] + 8*a[1] + 6*a[2] + (L-22)*a[3]
otherwise: 7*a[0] + 8*a[1] + 6*a[2] + 6*a[3] + (L-28)*a[4]
```

No draw occurs.

**SLD-R03** Physical base [verified in 1.13c code (static)]. Minimum adds `MinDam + bp(L,MinLevDam1..5)` to source share `w`; maximum uses the corresponding maximum columns. This wrapping sum `v` proceeds to synergy and shift. No draw occurs.

**SLD-R04a** Source share [verified in 1.13c code (static)]. With either `SrcDam == 0` or include-source condition zero, use `w = 0` and make no source request. Otherwise, for selected endpoint `x` from R04b, compute `w = low32(SrcDam*x)/128`, signed division truncating toward zero. `SrcDam` is zero-extended u8. The local arithmetic draws nothing.

**SLD-R04b** Source endpoint selection [verified in 1.13c code (static)]. With source inclusion enabled, first request the selected weapon for `(unit,0)`. If present, request the source classification for that unit and its inventory association. Classification 2 chooses the selected item's stat 23 for minimum or 24 for maximum. An absent item or any other classification chooses caster stat 21 or 22 respectively. All reads are at layer zero and consume NRF-R12b/NRF-R12c's selected-array/minimum semantics, including the item wrappers. No classification request occurs for an absent weapon. Q02 gates production of the weapon/classification and their resulting integrated choice; supplied operation results establish the selection and scaling contract independently.

**SLD-R05** Physical synergy [verified in 1.13c code (static)]. If `DmgSymPerCalc` is nonempty, request R19's calculation in context `(unit,skill,L)`. For result `s != 0`, replace `v` with `v + mulDiv(s,v,100)`. The source share is included in `v`. This occurs before the shift, with no local draw beyond any requested calculation.

**SLD-R06** Physical shift [verified in 1.13c code (static)]. Return `low32(v << (HitShift & 31))`; discarded high bits do not survive. There is no saturation or sign-preserving shift. No draw occurs.

**SLD-R07** Kick bypass [verified in 1.13c code (static)]. After valid skill lookup, a set `Kick` condition requests the Kick producer on the same unit with minimum selection 1 or maximum selection 0. Return its result unchanged; omit source share, level progression, synergy and shift. This establishes dispatch/return behavior with a supplied Kick result. Kick arithmetic and its resulting endpoint remain the outside-boundary producer in Q08.

**SLD-R08** Elemental base and shift [verified in 1.13c code (static)]. For a valid skill and positive level, compute `e = low32((EMin + bp(L,EMinLev1..5)) << (HitShift & 31))`, or the corresponding maximum fields. Unlike physical damage, the shift precedes synergy. No draw occurs.

**SLD-R09** Elemental synergy [verified in 1.13c code (static)]. Evaluate nonempty `EDmgSymPerCalc` only when its pool is present and its unsigned reference is less than pool size. Minimum has an additional pre-evaluation condition: signed `e > 256` or `EMinLev1 != 0`. Maximum has no such condition. For calculation result `s != 0`, set `e = e + mulDiv(s,e,100)`; otherwise retain `e`. A rejected minimum makes no calculation request and consumes none of that calculation's draws.

**SLD-R10** Elemental mastery [verified in 1.13c code (static)]. After R09, when the caller's mastery condition is nonzero and the unit exists, add the SCX-R31 mastery amount to `e`, using that synergy-increased `e`. Select stat 329 for `EType` 1; 330 for 2; 331 for 4 or 12; 332 for 5. Other selectors add zero. For 329/330/332 use the ordinary layer-zero unit getter. The cold path requires a present list, a definition count greater than 331, and an available definition; read stat 331/layer zero from that list's selected array with NRF-R12b/NRF-R12c's conditional minimum adjustment. It does not force the own array. With value `m`, add zero for `m == 0`, otherwise `mulDiv(m,e,100)`. No local draw occurs. Stat names do not control this code-fixed mapping.

**SLD-R11** Elemental duration [verified in 1.13c code (static)]. For valid skill and positive level, use three signed level bands without a damage shift:

```
L <= 8:  t = (L-1)*ELevLen1
L <= 16: t = 7*ELevLen1 + (L-8)*ELevLen2
otherwise: t = 7*ELevLen1 + 8*ELevLen2 + (L-16)*ELevLen3
length = ELen + t
```

For a valid nonempty `ELenSymPerCalc` calculation returning `s != 0`, set `length = length + mulDiv(s,length,100)`. Products/additions wrap; output is i32 frames. No local draw occurs beyond requested calculations.

**SLD-R12** Skill to-hit value [verified in 1.13c code (static)]. This supplies an attack-rating percentage bonus to hit resolution. For valid skill and positive level, empty `ToHitCalc` returns wrapping `ToHit + (L-1)*LevToHit`, with no level bands. Nonempty `ToHitCalc` replaces the linear result completely: return zero if the pool is absent or unsigned reference is at least size; otherwise return the SCX evaluation result unchanged for `(unit,skill,L)`. A negative calculation result is not clamped. No local draw occurs beyond calculation evaluation.

**SLD-R13** To-hit hand-off [verified in 1.13c code (static)]. On do-function 46's path, add caster stat 325 at layer zero to R12's result with wrapping i32 addition. Supply this sum as the bonus to melee result resolution with the same caster, target and combat context, and missile condition zero. If that resolver reaches its hit-roll operation, it forwards the bonus unchanged. Eligibility, hit outcome and downstream draws belong to SBA-R02b and the hit-resolution contract; this hand-off does not guarantee that a roll is reached.

**SLD-R14** Physical fill [verified in 1.13c code (static)]. For a valid skill, request physical minimum then maximum for `(unit,skill,L)` with source inclusion zero. Set `r = RandAlt(caster seed,low32(max-min))` using NRF-R02, then add `min+r` to `record.physical`. There is at most one local range draw; a nonpositive signed range returns zero without a seed step. Q08 gates an actual Kick endpoint only; supplied endpoints and non-Kick calls require no Kick result.

**SLD-R15** Elemental fill [verified in 1.13c code (static)]. For a valid skill, request duration, elemental minimum and elemental maximum in that order, each with mastery condition 1 where the operation takes that condition. Then set `d = min + RandAlt(caster seed,low32(max-min))` and route `(EType,d,length)` under R16/R17, with both optional outputs absent. Return `d` unchanged after routing. The local range draw occurs after all three requests and before any random-element draw. Calculation-owned draws occur inside their requests.

**SLD-R16** Routing [verified in 1.13c code (static)]. Use the following code-fixed table. Damage additions wrap in i32; durations/hit class replace the existing value. Fields not listed retain their incoming value. Write the resist output only when present. For type 10 first execute R17 and then use its selected type.

| `EType` | Combat-record writes | Hit class | Resist output |
|---|---|---|---|
| 1 | `fire += d` | 32 | 39 `fireresist` |
| 2 | `lightning += d` | 64 | 41 `lightresist` |
| 3 | `magic += d` | unchanged | 37 `magicresist` |
| 4 | `cold += d`; `coldLength = length` | 48 | 43 `coldresist` |
| 5 | `poison += d`; `poisonLength = length` | 80 | 45 `poisonresist` |
| 6 | `lifeLeech += d` | unchanged | -1 |
| 7 | `manaBurn += d` | unchanged | -1 |
| 8 | `staminaBurn += d` | unchanged | -1 |
| 9 | `stunLength += d + length` | 96 | -1 |
| 11 | `burnMagnitude += d`; `burnLength = length` | unchanged | -1 |
| 12 | `cold += d`; `freezeLength = length` | 48 | 43 |
| 0 or greater than 12 | `physical += d` | unchanged | -1 |

No seed step occurs in this table except the type-10 preprocessing. Non-random types do not write the optional selected-element output.

**SLD-R17** Random element [verified in 1.13c code (static)]. For type 10, advance the supplied unit's seed once under NRF-R01, take `k = newLow & 3`, and select `{1,2,4,5}[k]`. If signed `length <= 0`, replace it with 50 frames. Write the selected type to the optional selected-element output when present, then route it under R16. This draw follows any R15 range draw; the selection cannot recurse to 10.

**SLD-R18** Caller order and damage hand-off [verified in 1.13c code (static)]. At the record-construction stage of do-function 38, clear the combat record, perform physical fill then elemental fill, set the supplied-physical flag and hit result, and pass that record to area damage. Their numeric flag/result values are 1 and 1. Do-function 53 clears the record, sets its already-built flag (value 2), then performs physical fill followed by elemental fill; after the elemental fill and before hand-off, a nonzero `skillHitClassValue` replaces the record's hit class, and zero leaves the routed hit class. Do-function 38 has no such replacement. On 38's supplied-physical path the downstream build omits physical/critical draws under DPO-R02a and DPO-R03. The local draw sequence is physical range, elemental range, then random element when requested, interleaved with any calculation draws inside endpoint requests. These claims start after caller eligibility and stop at record hand-off; other skill-specific effects are outside scope.

**SLD-R19** Calculation boundary [verified in 1.13c code (static)]. Empty references are skipped by callers. For a requested reference, an absent pool or unsigned reference at least its byte size yields zero. Otherwise evaluate the remaining bytes from that reference in context `(unit,skill,L)` and return its signed result. SCX-R01 through SCX-R11, SCX-R13 through SCX-R33 and their sub-rule evidence labels/gates own evaluation, parameter operations and any expression draws. Q04 applies only when the expression requests an unresolved producer. A supplied result supports the local consumer arithmetic without claiming that producer is verified.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SLD-R01a | Lookup/level failures | verified in 1.13c code (static) | none | SLD-V02, SLD-V05, SLD-V07, SLD-V08, SLD-V09 |
| SLD-R01b | Invalid physical max is 2 | verified in 1.13c code (static) | none | SLD-V09 |
| SLD-R02 | Five level bands | verified in 1.13c code (static) | none | SLD-V01 |
| SLD-R03 | Physical base and source sum | verified in 1.13c code (static) | Q02 only for produced source endpoint; supplied share is available | SLD-V02, SLD-V04 |
| SLD-R04a | Source scaling | verified in 1.13c code (static) | Q02 only for equipment-produced input; supplied x is available | SLD-V04 |
| SLD-R04b | Source choice with supplied equipment results | verified in 1.13c code (static) | Q02: complete equipment selection/classification integration | SLD-V17 |
| SLD-R05 | Synergy before physical shift | verified in 1.13c code (static) | Q04 only for unresolved expression producer results/draws | SLD-V03 |
| SLD-R06 | Wrapping shift | verified in 1.13c code (static) | none with supplied v | SLD-V02, SLD-V03 |
| SLD-R07 | Kick dispatch and unchanged return | verified in 1.13c code (static) | Q08: actual Kick result; supplied-result dispatch available | SLD-V18 |
| SLD-R08 | Element shift before synergy | verified in 1.13c code (static) | none | SLD-V05 |
| SLD-R09 | Element synergy/minimum gate | verified in 1.13c code (static) | Q04 only for unresolved expression producer results/draws | SLD-V05, SLD-V06 |
| SLD-R10 | Mastery mapping and addition | verified in 1.13c code (static) | none for shared getter/supplied stat contract | SLD-V14 |
| SLD-R11 | Three duration bands | verified in 1.13c code (static) | Q04 only for unresolved expression producer results/draws | SLD-V07 |
| SLD-R12 | Linear or calculated to-hit | verified in 1.13c code (static) | Q04 only for unresolved expression producer results/draws | SLD-V08 |
| SLD-R13 | Summed to-hit request | verified in 1.13c code (static) | Q06: downstream hit outcome only; hand-off available | SLD-V19 |
| SLD-R14 | Physical range fill | verified in 1.13c code (static) | Q08 only for actual Kick endpoint; supplied endpoints available | SLD-V10 |
| SLD-R15 | Duration/min/max/range/router order | verified in 1.13c code (static) | Q04 only for unresolved requested expression results/draws | SLD-V11, SLD-V15 |
| SLD-R16 | Record routing | verified in 1.13c code (static) | none with numeric EType; Q07 is text/compiled identity only | SLD-V11, SLD-V16 |
| SLD-R17 | Random element and default duration | verified in 1.13c code (static) | none | SLD-V12 |
| SLD-R18 | Local fill order and hand-off | verified in 1.13c code (static) | Q04 for complete expression-dependent trace; Q10: column identity of the do-function 53 replacement value only; downstream DPO retains its labels | SLD-V13 |
| SLD-R19 | Pool bounds/context | verified in 1.13c code (static) | Q04: dependency-produced expression results only | SLD-V03, SLD-V08, SLD-V20 |

## 5. Outputs and side effects

| Output/effect | Rule |
|---|---|
| Endpoint/duration/to-hit return, including invalid-input defaults | R01a/R01b, R03 through R12, R19 |
| Source, Kick, calculation and to-hit requests with specified arguments/order | R04b/R05/R07/R09/R11/R12/R13/R19 |
| Added physical damage | R14; R16 fallback |
| Added elemental damage, burn magnitude, leech/burn quantities or stun duration; replaced duration/hit-class fields; optional resist/selected-element outputs | R16/R17 |
| Elemental fill returns its rolled d | R15 |
| Cleared record and supplied-physical/hit or already-built flags; do-function 53 hit-class replacement; downstream record request | R18 |
| Caster seed advanced for each positive range and each random-element choice | R14/R15/R17 |

No stat, state, event or message is written by the local rules. Calculation and combat dependencies retain their own effects.

## 6. Branches and edge cases

- Physical level zero uses the level-1 progression; elemental damage, duration and to-hit reject it.
- Negative percentages can reduce damage/duration; wrapping arithmetic and truncation are retained.
- `HitShift` 32 acts as shift zero. Physical synergy precedes the shift; elemental synergy follows it.
- Equal or reversed endpoints use minimum without a range step. Type 10 still steps once even when the range did not draw.
- The fills disable ordinary physical source inclusion and enable elemental mastery. This does not suppress the Kick bypass.
- A physical endpoint needing source inclusion does not reject an absent unit before its equipment request; equipment lookup faults on an absent unit. No zero-source fallback is defined for that case.
- Numeric `EType` zero routes elemental-fill damage into physical. Type 9 adds both d and length to stun duration.
- Difficulty and PvP are not direct arithmetic inputs here; later combat/status owners apply their rules.

## 7. Open questions

| Id | Kind | Exact outstanding boundary and settling evidence |
|---|---|---|
| SLD-Q02 | integration | Equipment-produced weapon/classification results in R04b. Inventory/item contracts must establish selection and classification 2. Supplied-result choice/scaling remains available. |
| SLD-Q04 | integration | Expression requests for SCX-R16b skill-entry production, SCX-R18b special to-hit, SCX-R27 parameter results or SCX-R32b physical source/Kick retain those dependency gates. Close the particular requested producer to establish its endpoint result and draw trace; supplied results and verified constant expressions remain available. Compilation is upstream of the supplied-bytecode boundary. |
| SLD-Q05 | outside-boundary | Other physical endpoint callers with source inclusion enabled are not enumerated here; an explicitly scoped caller audit would establish those contexts. None locally is gated. |
| SLD-Q06 | outside-boundary | Hit eligibility, hit outcome and downstream draws remain the SBA contract, including its gates. R13's exact summed request is established. |
| SLD-Q07 | integration | ElemTypes text row/code evidence has no compiled selector-index correspondence record; correlate the loaded linked selector with the named effective cells. Numeric routing/mastery is independently established. |
| SLD-Q08 | outside-boundary | The Kick producer's actual calculation is outside scope. A complete Kick contract would enable integrated Kick endpoints; supplied-result bypass behavior is available. |
| SLD-Q09 | observation | No matching gameplay observations exist for this spec's behavioral rules. Reproduce recorded vector assertions with controlled original-game inputs. |
| SLD-Q10 | integration | The Skills column loaded into `skillHitClassValue` (R18) is not identified by a descriptor or table-cell record; a Skills descriptor reading that maps the field, plus a named cell for a do-function 53 skill, would settle it. The replacement with a supplied value is established. |

## 8. Test vectors

### 8.1 Executable vectors

Each fixture is independent. Unless replaced explicitly: skill 0 of count 1 is present; unit U exists; `Kick` is false; include-source/mastery conditions are zero; bases, increments and `SrcDam` are zero; `HitShift = 8`; calculation references are empty. A supplied calculation uses reference 0 of pool `[0]` of size 1 and returns the stated constant without draws in context `(U,0,L)`; this tests the consumer with an evaluator stub, not those bytes' expression meaning. A supplied endpoint stage tests caller requests and returned-input consumption; it does not verify the endpoint producer. Record fields start at zero unless stated. All vectors have `runtime_observed: no`. A `no draw` fixture retains any supplied seed unchanged.

- **SLD-V01** (R02; isolated progression). Inputs: `a=[10,20,30,40,50]`, levels `[0,1,2,8,9,16,17,22,23,28,29,30]`. Seed: `no draw`; draws: none. Trace: cumulative tiers `0;0;10;7*10;70+20;70+8*20;230+30;230+6*30;410+40;410+6*40;650+50;650+2*50`. Outputs: `[0,0,10,70,90,230,260,410,450,650,700,750]`. Absent array at L30 returns 0. Discriminates all tier transitions and level-1 baseline.

- **SLD-V02** (R01a/R03/R06; endpoints). Inputs: `MinDam=3`, min increments `[2,3,4,5,6]`, `MaxDam=5`, max increments `[3,4,5,6,7]`, levels 10, 1 and 0 independently. Seed: `no draw`; draws: none. Trace at 10: min `3+7*2+2*3=23`, max `5+7*3+2*4=34`, then shift 8. Outputs at L10: minimum `0x1700`, maximum `0x2200`; at L1 and L0: `0x300`, `0x500`. Additional shift-only cases: supplied v `0x40000001`, shift 2 returns 4; v 3, shift 32 returns 3. Discriminates no physical level-zero rejection, wrap and masked count.

- **SLD-V03** (R05/R19; physical endpoint consumer). Inputs: V02 at L10; valid physical calculation returning 50. Seed: `no draw`; draws: none. Trace: min `23+mulDiv(50,23,100)=34`, max `34+17=51`, then shift 8. Outputs: `0x2200`, `0x3300`; one evaluation per endpoint, context `(U,0,10)`. Independent minimum cases: result -20 gives `23+(-460/100)=19`, output `0x1300`; absent pool gives `0x1700`, no evaluation. With supplied source x 10, SrcDam 64/include-source 1, percentage 50: `w=5`, `v=28`, `v+14=42`, output `0x2A00`. Discriminates synergy before shift, truncation toward zero and inclusion of source share.

- **SLD-V04** (R03/R04a; supplied-source endpoint). Inputs: V02 minimum at L10, SrcDam 64, include-source 1; supplied selection chooses caster stat 21 value 10. Seed: `no draw`; draws: none. Trace: `w=640/128=5`; `(5+23)<<8`. Output: `0x1C00`. Independent source-share cases: SrcDam 128/x -3 gives product -384 and w -3; SrcDam 64/x -3 gives product -192 and w -1. With SrcDam 0 or include-source 0: w 0 and no source request. Discriminates signed division and both inclusion gates; equipment production is Q02.

- **SLD-V05** (R01a/R08/R09; elemental endpoints). Inputs: `EMin=3`, min increments `[2,3,4,5,6]`, `EMax=5`, max increments `[3,4,5,6,7]`, L10. Seed: `no draw`; draws: none. Trace: shifted endpoints `23<<8=0x1700`, `34<<8=0x2200`. Outputs with no calculation: `0x1700`,`0x2200`. Independent valid calculation 50: min `5888+2944=8832 (0x2280)`, max `0x2200+0x1100=0x3300`; context `(U,0,10)` each. At L0 both return 0 and make no evaluation. Discriminates elemental shift-before-synergy from V03.

- **SLD-V06** (R09; minimum calculation gate). Inputs: EMin 1/EMax 1, all increments 0, L5, valid calculation 50. Seed: `no draw`; draws: none. Trace: shifted min 256 is not greater than 256, first min increment 0; max evaluates. Outputs: min `0x100` with no evaluation, max `0x180` with one evaluation. Independent L1, EMinLev1=1: bp 0, min gate passes and returns `0x180`. Discriminates strict comparison and increment exception.

- **SLD-V07** (R01a/R11; duration). Inputs: ELen 25, increments `[5,3,1]`, no calculation; levels 20,1,9,0 independently. Seed: `no draw`; draws: none. Trace: at 20 `25+7*5+8*3+4*1=88`; at 9 `25+35+3=63`. Outputs: `[88,25,63,0]` frames. Independent L20 valid calculation 100: `88+88=176`. Discriminates three bands, no damage shift and duration synergy.

- **SLD-V08** (R01a/R12/R19; to-hit). Inputs: ToHit 15, LevToHit 5, levels 10,1,0 independently. Seed: `no draw`; draws: none. Empty calculation outputs `[60,15,0]`; trace at 10 `15+9*5`. Independent L10 valid result 120 returns 120; valid result -7 returns -7. Reference 1/size 1 or absent pool returns 0 without evaluation. Discriminates replacement, equality bound and unchanged signed return.

- **SLD-V09** (R01a/R01b; rejection). Inputs: count 1, identity -1 then 1; independently identity 0 with absent record. L10; seed `{1,0x29A}`; record physical 7/fire 9. Draws: none. Trace: reject before calculations/endpoints/fills. Outputs in every case: physical minimum 1, maximum 2, elemental min/max 0, duration 0, to-hit 0; both fills preserve physical 7/fire 9 and seed `{1,0x29A}`; elemental fill returns 0; request log empty. Discriminates inclusive upper bound and asymmetric physical defaults.

- **SLD-V10** (R14; supplied-endpoint fill). Inputs: U/skill 0/L1; supplied physical min `0x500`, max `0x900`; record.physical `0x100`; seed `{1,0x29A}`. Draws: range step 1 to `{0x6AC6935F,0}`, range `0x400`, r `0x35F`. Trace: request minimum then maximum, both include-source 0; add `0x500+0x35F`. Outputs: physical `0x95F`, exact two-request log and final seed above. Independent max `0x500` then `0x400`: no draw, physical `0x600`, seed unchanged in each. Discriminates exclusive range, additive write and no draw for equal/reversed bounds.

- **SLD-V11** (R15/R16; supplied-endpoint elemental fill). Inputs: U/skill 0/L1, supplied min `0x300`, max `0x700`, zero record, seed `{1,0x29A}`; separate (EType,length) cases below. Draws: one range step to `{0x6AC6935F,0}`, r `0x35F`. Trace: duration, minimum, maximum requests in that order, all applicable mastery conditions 1; d `0x65F`; route with optional outputs absent. Outputs: type 1/length 0 fire `0x65F`, hitClass 32; type 4/length 50 cold `0x65F`, coldLength 50, hitClass 48; type 9/length 50 stunLength `0x691`, hitClass 96; type 3/length 0 magic `0x65F`, hitClass 0; type 0/length 0 physical `0x65F`. Every case returns d `0x65F` and the stated final seed, preserving other zero fields. Discriminates request order, mastery flags, d+length and unchanged return after type 9 routing.

- **SLD-V12** (R17; random routing). Inputs: supplied elemental endpoints both `0x400`, EType 10, length 0, L1, zero record. Case a seed `{2,0x29A}`; one selection step to `{0xD58D2424,0}`, k 0/type 1. Outputs: fire `0x400`, hitClass 32. Case b seed `{1,0x29A}`; one selection step to `{0x6AC6935F,0}`, k 3/type 5. Outputs: poison `0x400`, poisonLength 50, hitClass 80. No range draw in either. Case c endpoints `0x300`,`0x700`, seed `{1,0x29A}`: range step r `0x35F`, then selection step to `{0x2F2ED81B,0x2C890AFD}`, k 3/type 5; output poison `0x65F`, poisonLength 50, hitClass 80. Each fill returns its d; other zero fields persist. Discriminates selection after range and default duration even without a range draw.

- **SLD-V13** (R18; record-construction stage with supplied endpoints). Inputs: do-function 38, U/skill 0/L1, supplied physical endpoints `0x500`,`0x900`, elemental endpoints `0x300`,`0x700`, duration 0/EType 1; supplied endpoint operations make no other requests/draws; seed `{1,0x29A}`. Draws: physical step to `{0x6AC6935F,0}`, r1 `0x35F`; elemental step to `{0x2F2ED81B,0x2C890AFD}`, r2 `0x01B`. Trace: clear record; physical `0x85F`; fire `0x31B`; set flags/result; hand off. Outputs: physical `0x85F`, fire `0x31B`, hitClass 32, flags 1, resultFlags 1, other fields 0; final seed above; request log physical-min/max then elemental-duration/min/max then area hand-off. Downstream supplied-physical build has no physical/critical draw under DPO's contract. Independent do-function 53 with identical inputs and `skillHitClassValue` 0: same damage/seed/order, hitClass 32, flags 2, resultFlags 0 at this stage. Do-function 53 with `skillHitClassValue` 7, otherwise identical: same damage/seed/order/flags, hitClass 7 (replaced after the elemental fill's 32). Discriminates fill order, distinct caller flags and the post-fill replacement (a replacement before the fill would leave 32); this does not test caller eligibility or other effects.

- **SLD-V14** (R10; mastery integration with shared getter). Inputs: EMin 10/EMax 10, zero increments, L1, valid elemental synergy 50, mastery condition 1, EType 1, caster stat 329=30 at layer 0; other mastery stats 0. Seed: `no draw`; draws: none. Trace: `10<<8=2560`, synergy `+1280=3840`, mastery `mulDiv(30,3840,100)=1152`. Outputs: both endpoints 4992 and read log stat 329/layer 0. Independent EType 3 returns 3840 with no mastery read; mastery condition 0 or absent unit also returns 3840 without a mastery read. Mapping-only cases with damage input 1000 and selected stat value 30: EType 2/stat 330, 5/stat 332, 4/stat 331 and 12/stat 331 each add 300. Cold case: present list/definition count 359, definition 331 present with fMin false/MinAccr 0/ValShift 0; usesResolvedStats true, owner type 0, own stat 331=10 and resolved stat 331=30, active count 1 each. Output mastery amount 300, not 100. Cold missing list, count 331, or absent definition returns 0. Discriminates post-synergy input, exact selectors and selected-array source.

  Independent minimum-only rounding case (R09/R10): skill 0 of count 1 is present, unit U exists, L1, EMin 257, all five minimum increments 0, HitShift 0, EType 1, mastery condition 1. EDmgSymPerCalc references byte 0 of pool `[0]` of size 1; the supplied evaluator returns 50 without a draw in context `(U,0,1)`. The supplied ordinary stat getter returns 30 for stat 329/layer 0; other vector defaults apply. Seed: `no draw`; draws: none. Trace: shifted base 257 passes the minimum synergy gate; `mulDiv(50,257,100)=128`, giving 385; `mulDiv(30,385,100)=115`, giving 500. Both calls take NRF-R05's narrow branch because their first operands are at most `0x10000` and their second operands are at most `0x100000`; the products fit signed i32. Outputs: elemental minimum 500; one evaluation in context `(U,0,1)`, followed by one stat 329/layer 0 read. Discriminates operation order through truncation: applying mastery first gives `257+77=334`, then synergy gives `334+167=501` and fails this assertion. runtime_observed: no.

- **SLD-V15** (R15; supplied endpoint with a draw). Inputs: U/skill 0/L1, zero record, EType 1, seed `{1,0x29A}`; supplied duration operation returns 0 without a draw, minimum operation consumes one NRF-R01 step then returns `0x300`, maximum returns `0x700` without a draw. Draws: dependency step to `{0x6AC6935F,0}` during minimum, then range step to `{0x2F2ED81B,0x2C890AFD}`, r `0x01B`. Trace: duration -> minimum/dependency step -> maximum -> range -> route. Outputs: fire `0x31B`, hitClass 32, return `0x31B`, final seed above; other fields 0. Discriminates range drawing after endpoint execution; the supplied dependency's draw does not verify an expression producer.

- **SLD-V16** (R16/R17; isolated router). Inputs: d 7, length 11, every damage/drain/duration field 100, hitClass 9, present resist output initially 999 and selected-element output initially 999; each non-random type independent. Seed `{1,0x29A}`; draws none. Trace: apply one table row. Outputs: types 1/2/3/4/5/6/7/8/11/12/0/13 change respectively fire/lightning/magic/cold/poison/lifeLeech/manaBurn/staminaBurn/burnMagnitude/cold/physical/physical to 107. Type 9 changes stunLength to 118. Types 4/5/11/12 replace coldLength/poisonLength/burnLength/freezeLength with 11. HitClass for types 1/2/4/5/9/12 is 32/64/48/80/96/48; otherwise 9. Resist outputs for 1/2/3/4/5/12 are 39/41/37/43/45/43, otherwise -1. Selected-element output 999 and seed remain unchanged; all other fields 100. Additional type 10 case: d 7, length -1, all record fields 100, seed `{1,0x29A}` -> `{0x6AC6935F,0}`; selected type 5, poison 107, poisonLength 50, hitClass 80, resist 45, selected-element output 5. Discriminates every branch, additive versus replacement writes and optional outputs.

- **SLD-V17** (R04b; supplied equipment operations/shared getter). Inputs: minimum/maximum source requests on U; selected weapon W, supplied classification 2; caster selected stat 21=10/stat 22=20, W selected stat 23=30/stat 24=40, own stat 23=3/stat 24=4; both lists use resolved arrays with active count 2, W own count 2, U own count 0; definition count 359, required definitions present with fMin false. Seed: `no draw`; draws: none. Trace: weapon(U,0), classification(U,inventory association), selected item stat read. Outputs: min 30/max 40, exact request order; no caster endpoint read. Independent classification 1 uses caster 10/20. Absent weapon uses caster 10/20 without classification. Discriminates item versus caster, endpoint identity, selected array and absent-weapon short circuit; producing W/classification remains Q02.

- **SLD-V18** (R07; supplied-result bypass). Inputs: valid Kick skill, L0, physical base 100/increments 10, HitShift 8, SrcDam 128/include-source 1, nonempty synergy reference; Kick operation returns 37 for selection 1 and 91 for selection 0. Seed: `no draw`; supplied Kick operation draws none. Trace: lookup then Kick request. Outputs: min 37/max 91; exactly `(U,1)` or `(U,0)` request; no source/progression/calculation/shift. Discriminates direct return and bypass despite level zero. This does not verify Kick production.

- **SLD-V19** (R13; hand-off with supplied resolution operations). Inputs: do-function 46 path, caster U/type 0, target D/type 1/mode 1, combat context G, skill 0/L10, R12 output 60, caster stat 325=-7; both prior eligibility operations return 1 without draws, the third resolver condition is 0, and the hit-roll operation returns supplied miss 0. Seed: `no draw`; supplied dependencies draw none. Trace: `60+(-7)=53`; request melee resolution `(U,D,G,bonus 53,condition 0)`; guards/eligibility pass and hit-roll request is `(U,D,bonus 53,missile 0)`. Outputs: those exact ordered requests; resolver result 0; no subsequent block/avoidance request. Independent R12 output `0x7FFFFFFF`, addend 1 forwards signed -2147483648. Discriminates sum, wrap, and unchanged forwarding. No real hit outcome is asserted.

- **SLD-V20** (R19; calculation front end with supplied evaluator). Inputs: U/skill 0/L10, pool bytes `[11,22,33,44]` of size 4, reference 1, evaluator returns -9 without draws. Seed: `no draw`; draws: none. Trace: unsigned 1<4; evaluate remaining bytes `[22,33,44]` with context `(U,0,10)`. Outputs: -9, one evaluation with those bytes/reference 1/length 3/context. Independent reference 4 or `0xFFFFFFFE`, or absent pool: output 0, no evaluation. Empty reference skips evaluation at the consumer. Discriminates equality/unsigned bound, remaining length and unchanged signed result. This supplied evaluator fixture does not assert valid expression encoding.

### 8.2 Observation vectors

- **SLD-O01** (R02/R06/R14): compare physical endpoint returns across levels 8/9,16/17,22/23,28/29, and the resulting fill; [procedure](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md#level-observation). Pending.
- **SLD-O02** (R05/R09): compare raw physical and elemental endpoints with controlled nonintegral synergy percentages; [procedure](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md#synergy-observation). Pending.
- **SLD-O03** (R17): compare recorded seed steps and selected random element/duration; [procedure](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md#random-element-observation). Pending.
- **SLD-O04** (R10): compare mastery-disabled/enabled elemental results with measured mastery inputs; [procedure](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md#mastery-observation). Pending. Other rules still require attribution of their own matched assertions under Q09.

## 9. Acceptance and dependencies

Partial static acceptance consists of the verified assertions in SLD-V01 through SLD-V20. Supplied equipment, Kick, expression and endpoint results establish only their explicitly stated consumers. Complete integrated source/Kick/parameterized-expression endpoints retain Q02/Q08/Q04, and downstream damage/hit effects retain their owners' labels. Q07 does not block numeric routing/mastery with supplied EType. Full reproduction requires closing the applicable integration assertions and matching observations for every behavioral rule under Q09; a partial pass or a named archive cell establishes no complete parity claim.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), NRF-R01/R02 for seed/range operations, NRF-R05 for argument-ordered percentage arithmetic, NRF-R06a for signed source division, NRF-R08a for fixed-point units, NRF-R12b/NRF-R12c for selected stat reads; [core-game-data-fields](core-game-data-fields.md), GDF-R08 for Skills fields and GDF-R11 for named stat identities; [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md), SCX-R01 through SCX-R11 and SCX-R13 through SCX-R33 for evaluation/context/parameter results with their sub-rule gates, including SCX-R31 for mastery; [combat-damage-pipeline-order](combat-damage-pipeline-order.md), DPO-R02a/DPO-R03 for the downstream supplied-physical result only. Equipment and Kick production are explicit missing upstream contracts.

**Depended on by:** [combat-damage-pipeline-order](combat-damage-pipeline-order.md) consumes incoming combat-record damage; [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md) consumes R13's bonus; skill handlers using these fills consume endpoint/record results. Reverse consumers do not become prerequisites for the local supplied-input arithmetic.

## 10. Related specs

- [combat-status-application](combat-status-application.md): consumes routed durations.
- [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md): overlapping endpoint arithmetic and expression execution.
- [skill-assassin-death-sentry-corpse-damage](skill-assassin-death-sentry-corpse-damage.md): a separate skill-specific damage formula.
- [core-game-data-fields](core-game-data-fields.md): field and text-name evidence.

## Appendix A. Provenance

The evidence target is Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies original D2Common/D2Game binaries. Compatible original-instruction and byte-matched reconstruction readings dated 2026-10-12 are reused within their recorded boundaries; targeted original-instruction readings and hash-checked image decoding dated 2026-10-15 cover the additional closure records in the [companion](../research/spec-evidence/skill-leveled-damage-tohit-and-elements.md). Named text cells were read through the original loader on 2026-10-15 with the profile/compiled-correspondence limits recorded there. No gameplay observation is claimed. Pending integration and observation evidence keeps the document draft.
