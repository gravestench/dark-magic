# Missile Damage Parameters

**Description**

When a missile is created, the game prepares the damage ranges and effect durations that it will carry. These can come from the missile's own damage values, a skill, its source, or a parent missile. The preparation also selects bypass properties and can decide a critical strike before contact; later hit processing uses the resulting missile stats.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MDP-R, MDP-Q, MDP-V, MDP-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/missile-damage-parameters.md)
Open evidence gaps: [gap file](../plans/gaps/missile-damage-parameters.md)

## 1. Scope

This contract covers creation-time missile damage parameters: column and skill endpoints, source or parent contribution, creation critical decision, elemental routing and scaling, and ordered stat-write requests. Equipment/mastery and attack-preparation results have explicit supplied-result boundaries where their production remains gated. [Missile creation](missile-creation-and-seeding.md) owns the call and [missile initialization](missile-initialization-and-registry.md) owns unit initialization and ownership; [missile hit and damage](missile-hit-and-damage-calculation.md) owns contact, hit-time rolls and later damage processing. Shared skill endpoints and calculation-language execution belong to their named dependencies. No original internal object representation is required by this contract.

## 2. Trigger and ownership

D2Common computes the parameters and issues the writes from its missile subsystem. D2Game requests this once within missile creation, after fresh stat-list setup, creation callback, skill and level assignment, and before owner attribution (MIR-R11 through MIR-R14). The semantic builder inputs are source, optional parent missile, new missile and requested level; the caller then passes the resulting parameter record to the store stage. The [companion](../research/spec-evidence/missile-damage-parameters.md#dispatch-and-functions) gives the original caller/import/output chain. This operation has no independent frame schedule and sends no client message; subsequent server hit processing consumes the missile stats.

## 3. Inputs and data

Unless specified otherwise, arithmetic is signed i32 with low-32-bit wrapping products/additions and signed division truncating toward zero. `share(v,p)` means `i32(low32(v*p))/128` under NRF-R06a. `mulDiv(p,v,100)` consumes NRF-R05 in that exact operand order. Shift counts are masked by 31; shifted damage is raw 8.8 when `HitShift = 8` (NRF-R08a). Durations are signed frame counts. Stat reads use layer zero and RLI-R02a/RLI-R02b, including absent-list/definition and selected-array/minimum behavior. No stat getter here implies a verified stat writer.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| source, parent, missile | caller-provided units | optional units, type and class identity i32 | Source selection and seed ownership |
| missile table/count and selected record presence | effective Missiles collection | signed class/count, optional record | Lookup rejection |
| `MinDamage`, `MaxDamage`, `EMin`, `EMax` | Missiles columns | i32 | Column endpoint bases |
| `MinLevDam1..5`, `MaxLevDam1..5`, `MinELev1..5`, `MaxELev1..5` | Missiles columns | five i32 increments for each endpoint | Level bands |
| `DmgSymPerCalc`, `EDmgSymPerCalc` | Missiles calculation references | u32 pattern; all-one pattern is empty | Optional percentage expression |
| `ELen`, `ELevLen1..3` | Missiles columns | i32 frames | Element length progression |
| `EType`, `HitShift`, `SrcDamage`, `SrcMissDmg`, `Holy` | Missiles columns | u8; shares in 1/128 | Routing, shift, source/parent shares, bypass selectors |
| `Skill` | Missiles linked skill | signed i16 | Skill damage source |
| `ApplyMastery`, `MissileSkill`, `Half2HSrc` | Missiles flags | conditions | Mastery, stored-skill selection, conditional share halving |
| stored missile skill and level | new missile's associated data | signed i16 each; optional data | Skill override and nonpositive endpoint-level fallback |
| requested identity, level `L` | endpoint caller | i32 | Explicit selection and progression |
| selected skill record, `SrcDam`, `EType` | Skills collection/columns | identity/count i32; two u8 columns | Skill branch source share and element selector |
| source stats 0, 2, 17, 18, 21, 22, 23, 24, 25, 103, 104, 106, 120, 121, 122, 141, 159, 160, 337 | source's layer-zero stats | i32 | Attribute, endpoint, percent, bypass and critical inputs |
| parent stats 17, 18, 21, 22, 25, 120, 121, 122, 141, 337 | parent's layer-zero stats | i32; damage already raw | Parent contribution/critical inputs |
| source/parent stats 48..60, 62, 101, 315, 316, 317, 326 | selected contributing unit | i32 | Elemental additions under R11a; only the listed mapping is read |
| source mastery stats 329, 330, 331, 332 | source at layer zero | i32 | Elemental percentage; GDF-R02 and SLD-R10 names |
| missile stat 25 | new missile at layer zero | i32 | Added to physical percent |
| selected weapon, item classification and predicates | equipment/source operations | optional item; i32/conditions | Supplied producer results under Q04 |
| strength/dexterity coefficients | item-record operations | low16 reinterpreted as signed i16 | Attribute percentages under R07a |
| mastery result | request `(unit,weapon,context 0,selector 1 or 2)` | i32 | Physical bonus or third critical chance |
| preparation operation | source and condition 1 | side-effecting dependency | Runs before source elemental reads; Q04 |
| unit seed | contributing source or parent | two u32 words | NRF-R01 critical draws; calculation draws have their own owner |
| source stat-180 enumeration eligibility/results | source list and enumeration operation | eligibility condition; ordered entries with u16 layer and i32 value; capacity 128 | Copied replacing-write requests |
| parameter record and missile metadata | builder/store | semantic fields listed below; u32 markers | Transient results and requested writes |

The parameter record contains physical minimum/maximum; fire minimum/maximum and `fireLength`; lightning minimum/maximum; magic minimum/maximum; cold minimum/maximum and `coldLength`; poison minimum/maximum, `poisonLength`, `poisonCount`; life-drain, mana-drain and stamina-drain minimum/maximum; `stunLength`; burn minimum/maximum and `burnLength`; percent fields for stat 121, stat 122, stat 120 and physical stat 25; and conditions `sourcePhysical`, `critical`, `secondaryMarker`, `bypassUndead`, `bypassDemons`, `bypassBeasts`. All numeric fields are i32 and start at zero. These semantic names impose no storage layout. `secondaryMarker` is an input to storage; the builder does not set it.

| Mapping or constant | Origin | Binding |
|---|---|---|
| Missiles fields and flags; Skills `SrcDam` and `EType` | table data | Values selected from prepared records; column identities established by original descriptor code, [record](../research/spec-evidence/missile-damage-parameters.md#field-bindings) |
| `Holy` bits 0/1/2 to undead/demon/beast bypass; EType routing; mastery stat selection | engine code | Fixed by handlers; text names do not establish handler execution |
| Bands, divisors 100/128, physical-percent floor -90, type test 57, copied-entry capacity 128 | engine code | Fixed by original code |
| Stat identities and hireling-category class membership | engine code | Fixed numeric reads; shared names consume GDF-R02; category consumes CSA-R02 |

All numeric row values in the vectors are synthetic supplied records, not claims about archive cells. Table preparation and effective text-to-compiled correspondence retain GDF's evidence limits.

## 4. Rules (execution order)

**MDP-R01a** Column endpoints [verified in 1.13c code (static)]. A present missile argument of any type other than missile (3) returns 0. A negative supplied class uses that unit's class only when the missile argument exists. With a present missile and signed `L <= 0`, use its stored signed i16 level, or 0 if associated missile data is absent. Without a missile argument, leave the supplied level unchanged. Negative/out-of-range class or absent selected row returns 0. For physical minimum/maximum select `MinDamage`/`MaxDamage`, the corresponding five physical increments and `DmgSymPerCalc`; for elemental minimum/maximum select `EMin`/`EMax`, corresponding five element increments and `EDmgSymPerCalc`:

```
v = Base + bandProgression(L,Steps)
if calculation reference is not empty:
    p = missileCalculation(missile,source,reference,class,L)
    if p != 0: v = v + mulDiv(p,v,100)
return low32(v << (HitShift & 31))
```

The arithmetic itself draws nothing. Calculation execution can draw and retains R01b's integration limits.

**MDP-R01b** Missile calculation boundary [verified in 1.13c code (static)]. Consume SCX-R29a for the missile pool, unsigned reference bound, remaining-length evaluation and context `(missile,source,class,L)`, and the shared interpreter's verified token rules. Its own negative-only level fallback differs from R01a's nonpositive fallback. Parameter/resolver and owner-entry production retain SCX-R29b, SCX-Q04/SCX-Q07 and MDP-Q02. Supplied calculation results verify local arithmetic without verifying their producer.

**MDP-R02** Level bands [verified in 1.13c code (static)]. Consume SLD-R02's five-band progression with the selected endpoint's five increments and level from R01a. Its signed i32 wrapping arithmetic, no-draw behavior and absent-array result 0 apply unchanged. R01a owns which columns and level reach this shared operation.

**MDP-R03** Element length [verified in 1.13c code (static)]. Invalid class/record returns 0. Use the requested level directly, with no missile-level fallback. If `L <= 0`, return `ELen`; if `1 <= L <= 8`, return `ELen+(L-1)*ELevLen1`; through 16 return `ELen+7*ELevLen1+(L-8)*ELevLen2`; thereafter return `ELen+7*ELevLen1+8*ELevLen2+(L-16)*ELevLen3`. No expression, shift or draw occurs.

**MDP-R04** Builder selection and endpoint request order [verified in 1.13c code (static)]. Clear every parameter field/condition first. Absent source, absent/wrong-type missile, or invalid missile class/record returns with the cleared record and no further operation. With signed `Skill <= 0` and `MissileSkill` false, take physical share from `SrcDamage`, parent share from `SrcMissDmg`, and request in order physical minimum, physical maximum, element minimum, element maximum (R01a), then element length (R03), passing new missile/source/class/requested level. Take `EType` from the missile row. If `ApplyMastery`, add R05's bonus first to element minimum, then to maximum.

Otherwise start with the signed linked `Skill`. When `MissileSkill` is true, request the missile's stored skill, and use that result only if signed positive; otherwise keep the linked skill. Missing/invalid selected Skills entry returns with the zero record before bypass processing. For a valid skill set physical share to its unsigned `SrcDam`, except missile `SrcDamage == 255` forces it to 0; parent share is 0. Request, in order, physical minimum and maximum with source inclusion 0, then element minimum and maximum with mastery condition 1, then duration, always passing `(source,skill,requested level)`. Consume SLD-R01 through SLD-R11 and SLD-R19, including their endpoint sentinels, source/Kick/calculation gates. Use the skill's `EType`. These local requests add no draws; requested operations retain their own draws/gates. The builder returns a record, with no success integer specified.

**MDP-R05** Column elemental mastery [verified in 1.13c code (static)]. Element types 1/2/4-or-12/5 select source stats 329/330/331/332 respectively; all other types give bonus 0. For selected mastery `m`, return 0 if `m == 0`, otherwise `mulDiv(m,v,100)`. R04 adds the bonus to each endpoint with wrapping addition. No local draw. Ordinary selected stat reads consume RLI-R02a/RLI-R02b, including the optimized cold getter's same selected/minimum behavior.

**MDP-R06** Row bypass conditions [verified in 1.13c code (static)]. After successful endpoint selection, `Holy` bit 0 sets `bypassUndead`, bit 1 `bypassDemons`, bit 2 `bypassBeasts`; other bits have no effect here. No draw. This selects parameter conditions; R12a owns write requests and Q06 gates stored effects.

**MDP-R07a** Source physical contribution with supplied operations [verified in 1.13c code (static)]. When physical share is nonzero, set `sourcePhysical`. For a player request weapon selection `(source,1)`; for a monster with inventory request the monster weapon; otherwise use no weapon. Run R08a with the source seed and selected weapon; success sets `critical`. Initialize attribute/mastery percent and category-dexterity percent to zero.

With a selected item of type item (4), if `Half2HSrc` and the supplied two-handed predicate is nonzero, halve the physical share with signed truncation toward zero. Start the attribute/mastery percent from mastery request `(source,weapon,0,1)`. For each nonzero signed coefficient, add `trunc(low32(source strength stat 0 * strengthCoefficient)/100)` then the corresponding dexterity stat 2 term. The item coefficient helpers return 100 on absent/wrong-type item, absent item record, or stored coefficient 1; otherwise their low16 result is interpreted as signed i16. With nonzero supplied alternate-endpoint predicate, select source stats 159/160. Otherwise, if source inventory exists and supplied source classification equals 2, select source stats 23/24; otherwise source stats 21/22. Shift each endpoint left by 8. A nonzero supplied item-type relationship test `(weapon,57)` adds 50 to the stat-122 percent field.

Without an item weapon, use source stats 21/22 shifted by 8; a monster with nonzero fixed hireling category (CSA-R02) adds source dexterity stat 2 to the separate category percent. Scale the shifted endpoints with `share(endpoint,physicalShare)`. Contribution percent is category percent + attribute/mastery percent + source stat 25. Retain source stats 18 and 17 as two competing physical bonuses. Add source stats 121/122/120 to the corresponding percent fields, in that order; nonzero source stats 103/104/106 set the corresponding bypass conditions. No local draw beyond R08a; supplied operations do not establish producer effects.

**MDP-R07b** Equipment and mastery production [unresolved]. Producing the selected equipment, two-handed/type/classification/alternate-endpoint results and mastery values, including their effects and draw budgets, remains Q04. R07a specifies exact requests and consumption with supplied results.

**MDP-R08a** Creation critical decision [verified in 1.13c code (static)]. For a present contributing unit, read critical chance from stat 337 (zero for absent list/definition), advance that unit's seed once under NRF-R01, and take unsigned new low modulo 100. Compare the remainder to signed chance; return 1 on strict less-than. Otherwise read stat 141. If nonzero, draw again on the same seed and return 1 on strict less-than. Otherwise continue without that draw. If weapon exists, request mastery `(unit,weapon,0,2)`; if its result is nonzero, draw once more and compare likewise. Return 0 after all reached failures. Negative nonzero chances still consume their conditional draw; chance at least 100 succeeds on its reached draw. Short-circuit on success. The first draw always occurs for a valid unit, even with every chance zero. An absent unit faults at first seed access and has no integer return. Source contribution uses the source seed, parent contribution the parent seed. Calculation draws are separate. Q04 gates producing the weapon-mastery result, not this supplied-chance decision.

**MDP-R08b** Weapon mastery critical production [unresolved]. The third chance and any producer effects remain Q04; returning 1 only sets the parameter's `critical` condition. R12a requests stat 141 = 1, and hit-time doubling belongs to MHD-R10f with its own evidence label.

**MDP-R09** Parent physical contribution [verified in 1.13c code (static)]. Choose this branch only when the physical share entering R07a is zero, parent share nonzero and parent present; run R08a on the parent with no weapon. R07a's later share halving does not revisit this branch. Scale parent stat 21/22 directly with `share(value,parentShare)`; do not shift them by 8. Contribution percent is parent stat 25; competing bonuses are parent stats 18/17. Add parent stats 121/122/120 to the corresponding percent fields. This path does not set `sourcePhysical` and does not copy parent bypass stats. Missing parent or zero share leaves these contributions zero. No local draw beyond the critical decision.

**MDP-R10** Physical assembly and elemental routing [verified in 1.13c code (static)]. Add contribution percent, new missile stat 25 and the signed maximum of the competing bonuses; floor the wrapping result at -90. Write it as physical percent. Physical minimum/maximum are contribution endpoints + selected row/skill endpoints. Route selected element minimum, maximum and length by the following numeric selector; fields not named remain unchanged. No local draw. Selector semantics are fixed code bindings, independent of text-name correspondence.

| EType | Assigned fields |
|---|---|
| 1 | fire minimum, maximum, fireLength |
| 2 | lightning minimum, maximum |
| 3 | magic minimum, maximum |
| 4 or 12 | cold minimum, maximum, coldLength |
| 5 | poison minimum, maximum, poisonLength |
| 6 | life-drain minimum, maximum |
| 7 | mana-drain minimum, maximum |
| 8 | stamina-drain minimum, maximum |
| 9 | stunLength only |
| 11 | burn minimum, maximum, burnLength |
| 0, 10, greater than 12 | none |

This stores the physical percentage for subsequent use; it does not apply that percentage to the assembled physical endpoints here.

**MDP-R11a** Elemental additions and scaling with supplied prepared stats [verified in 1.13c code (static)]. Test the current physical share, including any halving by R07a. If nonzero, first request source preparation `(source,1)`, then use that source and shift 8. Otherwise, with nonzero parent share and a present parent use the parent and shift 0; otherwise do nothing. Thus an initial physical share 1 halved to 0 can select parent elemental additions even though source physical contribution ran and parent physical contribution did not. Read/add stats in this order: 48/49 to fire min/max, 50/51 to lightning min/max, 52/53 to magic min/max, 54/55 to cold min/max, each shifted by the selected shift; then 56 to coldLength, 57/58 to poison min/max, unshifted. If stat 101 is positive, add it to poisonLength and skip stats 59/326. Otherwise add stat 59 to poisonLength then stat 326 to poisonCount. Add, unshifted, stat 60 to life-drain minimum, 62 to mana-drain minimum, 316/317 to burn min/max, then 315 to burnLength. Every addition wraps.

Scale the whole resulting fire/lightning/magic/cold min/max, coldLength, poison min/max, poisonCount, life-drain minimum, mana-drain minimum, burn min/max and burnLength by `share(field,selectedShare)`, including the row/skill contributions already present. Share 128 skips scaling. Do not scale fireLength, poisonLength, drain maxima, stamina-drain values, stunLength or any percent field here. No local draw; preparation and expression draws retain their owners' gates.

**MDP-R11b** Source preparation effects [unresolved]. The preparation dependency can change effective source stats through equipment/list ownership; Q04 gates actual effects and whole-operation draw guarantees. R11a's supplied post-preparation stat contract remains available.

**MDP-R12a** Missile metadata and ordered replacing-write requests [verified in 1.13c code (static)]. An absent/wrong-type destination returns without requests. For a missile with associated metadata, retain its existing markers, set the source-physical marker if `sourcePhysical`, and set the second marker if `secondaryMarker`. Without metadata, continue issuing stats. Issue replacing requests on the destination list, layer 0, in exactly this order, including zero values:

| Stat | Value |
|---|---|
| 12 | requested level L |
| 21, 22 | physical minimum, maximum |
| 48, 49, 315 | fire minimum, maximum, fireLength |
| 50, 51 | lightning minimum, maximum |
| 52, 53 | magic minimum, maximum |
| 54, 55, 56 | cold minimum, maximum, coldLength |
| 57, 58, 59, 326 | poison minimum, maximum, poisonLength, poisonCount |
| 60, 61 | life-drain minimum, maximum |
| 62, 63 | mana-drain minimum, maximum |
| 64, 65 | stamina-drain minimum, maximum |
| 66 | stunLength |
| 316, 317, 315 | burn minimum, maximum, burnLength; second stat-315 request |
| 121, 122, 120, 25 | corresponding three percent fields, then physical percent |

Then, in order, for each true `bypassUndead`, `bypassDemons`, `bypassBeasts`, `critical`, request unit replacing stat 103/104/106/141 = 1 at layer 0. False conditions issue no corresponding request and do not request clearing an existing value. If source is present and its list satisfies enumeration eligibility, request up to 128 entries for stat 180, then issue replacing requests on the destination list for each returned entry in enumeration order, retaining its unsigned 16-bit layer and signed value. Otherwise skip enumeration/copy. Enumeration and replacing writes are supplied operations for request-log acceptance; no stored-state assertion follows from a logged request. There is no specified success return. This local sequence contains no RNG step; dependency effects remain Q06.

**MDP-R12b** Stat mutation and duplicate replacement [unresolved]. Actual stat propagation, replacement/removal by zero, and enumeration production consume RLI-Q03/STL-R06c and MDP-Q06. In particular, two requests to stat 315 do not prove the final stored fire/burn duration. The ordinary getters RLI-R02a/RLI-R02b cannot close this write contract.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MDP-R01a | Column endpoints | verified in 1.13c code (static) | MDP-Q02: actual unresolved calculation results only | MDP-V01, MDP-V02, MDP-V08 |
| MDP-R01b | Calculation boundary | verified in 1.13c code (static) | MDP-Q02: resolver/owner-entry integration | MDP-V02, SCX-V29 |
| MDP-R02 | Five level bands | verified in 1.13c code (static) | none | MDP-V03, SLD-V01 |
| MDP-R03 | Element length | verified in 1.13c code (static) | none | MDP-V04 |
| MDP-R04 | Builder selection/requests | verified in 1.13c code (static) | MDP-Q02/MDP-Q04: requested producer results only | MDP-V09, MDP-V16 |
| MDP-R05 | Elemental mastery | verified in 1.13c code (static) | none | MDP-V10 |
| MDP-R06 | Holy bypass conditions | verified in 1.13c code (static) | none | MDP-V05 |
| MDP-R07a | Source contribution consumer | verified in 1.13c code (static) | MDP-Q04: equipment/mastery production | MDP-V11, MDP-V16, MDP-V17 |
| MDP-R07b | Equipment/mastery producers | unresolved | MDP-Q04: actual results/effects | none |
| MDP-R08a | Critical decision/seed | verified in 1.13c code (static) | MDP-Q04: third chance producer only | MDP-V06 |
| MDP-R08b | Third-chance production | unresolved | MDP-Q04: mastery result/effects | none |
| MDP-R09 | Parent contribution | verified in 1.13c code (static) | none | MDP-V12, MDP-V16 |
| MDP-R10 | Clamp and routing | verified in 1.13c code (static) | none | MDP-V13, MDP-V16 |
| MDP-R11a | Element addition/scaling | verified in 1.13c code (static) | MDP-Q04: source preparation integration | MDP-V07, MDP-V14, MDP-V16, MDP-V18 |
| MDP-R11b | Preparation effects | unresolved | MDP-Q04: effective source/draw effects | none |
| MDP-R12a | Metadata/write requests | verified in 1.13c code (static) | MDP-Q06: stored effects/enumeration production | MDP-V05, MDP-V15 |
| MDP-R12b | Actual stored mutation | unresolved | MDP-Q06: stat propagation and zero replacement | none |

## 5. Outputs and side effects

| Output/effect | Rule |
|---|---|
| Endpoint, length, mastery, critical returns and transient semantic parameter fields | MDP-R01a through MDP-R11a |
| Contributing source or parent seed changes, one to three reached steps | MDP-R08a; weapon producer effects remain Q04 |
| Calculation-context seed changes when expression requests a draw | MDP-R01b via SCX-R15/SCX-R29a; exact context owns seed |
| Destination metadata source-physical and second marker OR updates | MDP-R12a |
| Ordered replacing stat requests for 12, 21, 22, 25, 48..66, 103, 104, 106, 120..122, 141, 180, 315..317, 326 as explicitly mapped | MDP-R12a; mutation gated by MDP-R12b |
| Source preparation and equipment/mastery dependency requests | MDP-R07a, MDP-R08a, MDP-R11a; actual effects gated by Q04 |

The temporary record is not independently persistent. There is no local unit creation, event scheduling, persistence write or client message. This does not eliminate side effects of unresolved dependencies.

## 6. Branches and edge cases

Source versus parent physical contribution is chosen before source-share halving. The later elemental stage chooses again using the modified share; an initial source share 1 halved to 0 can therefore use a parent's elemental stats and share without a parent critical draw or source-preparation request. Skill branches force parent share zero. `SrcDamage == 255` suppresses source contribution only on the skill branch; it is an unsigned 255 share on the column branch. Percent floor -90 applies to stored physical percent rather than clamping endpoint magnitudes. Endpoint level fallback is nonpositive and conditional on a present missile, length never falls back, and calculation-front-end fallback is negative-only. Share multiplication wraps before division; negative nonmultiples truncate toward zero. No local classic/expansion, difficulty or PvP branch appears in the inspected parameter operations; supplied records/stats/producers can still depend on those contexts. Store flags are conditional writes, so absence does not establish clearing. Fire length and burn length share a destination stat request with unresolved actual replacement behavior.

## 7. Open questions

Attempted evidence and exact blocked assertions are in the [gap file](../plans/gaps/missile-damage-parameters.md).

| Id | Status and boundary | Evidence that would settle it |
|---|---|---|
| **MDP-Q02** | integration: actual calculation parameter/owner-entry results | Close SCX-Q04/SCX-Q07/SCX-Q08 for the expressions requested by these endpoints; no gate on a supplied calculation's arithmetic consumer |
| **MDP-Q04** | integration: equipment/mastery/preparation and skill override results | Complete producer contracts for R07b/R08b/R11b and SLD's source/Kick dependencies; quantify actual effects/draw ownership for the reached path |
| **MDP-Q06** | integration: replacing writes, zero replacement/removal and stat-180 enumeration | Close RLI-Q03/STL-R06c plus the enumeration producer; distinguish final stat 315 after fireLength then burnLength=0 requests |
| **MDP-Q07** | observation: gameplay confirmation pending for every behavioral rule | Observe valid 1.13c creations/endpoints with recorded inputs, ordered writes/returns/seeds and evidence identity; existing availability failure supplies no match |

## 8. Test vectors

### 8.1 Executable vectors

All cases have `runtime_observed: no`. Unspecified fields and stats are 0; source/missile kinds are player/missile, valid selected row and present associated data unless overridden. Getters have valid definitions and return the supplied layer-zero values with no minimum substitution. Supplied selection, mastery, preparation, expression and stat-enumeration operations return the stated results and draw nothing. Preparation is a logged no-op only within a supplied-input vector; it does not model real integration. Replacing operations append an ordered `(destination,stat,layer,value)` log and do not claim stored mutation. Semantic record updates persist inside each case. Cases are independent and inherit no earlier case's outputs.

- **MDP-V01** Column base and band (MDP-R01a/MDP-R02; boundary: isolated rule)
  - Inputs: minimum base 3, increments `[1,2,3,4,5]`, empty calculation, HitShift 8, L10
  - Seeds: `no draw`
  - Draws: none
  - Trace: `7*1+2*2=11`; base+progression=14; `14<<8`
  - Outputs: 3584
  - Discriminates: wrong second-band length or selected endpoint base
  - runtime_observed: no

- **MDP-V02** Supplied calculation percentage (MDP-R01a/MDP-R01b; boundary: supplied-input stage)
  - Inputs: V01 record, nonempty reference 0, missile pool `[11,22,33,44]` of byte size 4, context missile M/source S/class 0/L10; per-case supplied evaluator result
  - Seeds: `no draw`
  - Draws: none; supplied evaluator draws none
  - Trace: unshifted value 14; apply `mulDiv(p,14,100)` before shift; log evaluation context
  - Outputs: per case
  - Discriminates: shifting before synergy or treating zero as an added base
  - runtime_observed: no
  - (a) Inputs: p=50. Draws: none. Outputs: `14+7=21`, endpoint 5376; one evaluation with context `(M,S,0,10)`.
  - (b) Inputs: p=0. Draws: none. Outputs: endpoint 3584; one evaluation, no percentage addition.
  - (c) Inputs: minimum base 100001, all increments 0, L1, HitShift 0, supplied p 65537. Draws: none. Trace: `mulDiv(65537,100001,100)` takes the first-operand-large divide-first branch: `655*100001=65500655`; add base 100001. Outputs: endpoint 65600656; logged percentage arguments `(65537,100001,100)`. Reversing operands would produce 65637001 and fails this assertion.
  - (d) Inputs: direct R01b front-end calls with class 0, missile stored level 10, valid pool/reference 0, supplied evaluator returns 73; independent requested L-1 and L0. Draws: none. Outputs: 73 in each case; evaluation contexts have level 10 and level 0 respectively. This is a front-end stage fixture, separate from R01a's earlier nonpositive fallback.
  - (e) Inputs: direct R01b front end, L10; independent reference 4, reference `0xFFFFFFFE`, or absent pool. Draws: none. Outputs: 0, no evaluator request in every case; pool bound is unsigned and equality fails.
  - (f) Inputs: four independent endpoint calls with the same valid missile record/class 0/L10/HitShift 8; physical minimum base 3/increments `[1,2,3,4,5]`, physical maximum base 5/increments `[2,3,4,5,6]`, element minimum base 7/increments `[3,4,5,6,7]`, element maximum base 11/increments `[4,5,6,7,8]`; physical calculation reference 0 returns 20, element reference 1 returns 50 from the shared supplied evaluator. Draws: none. Trace: base-plus-progression values are 14/25/36/49; percentage additions 2/5/18/24 give 16/30/54/73 before shifting. Outputs: physical minimum 4096, maximum 7680, element minimum 13824, maximum 18688; each physical call evaluates reference 0 and each element call reference 1 in context `(M,S,0,10)`. This distinguishes all four column/step selections and physical versus element calculation binding.
  - (g) Inputs: V01 record and supplied percentage 1, HitShift 8, L10. Draws: none. Trace: unshifted value 14, `mulDiv(1,14,100)=0`, then `14<<8=3584`. Outputs: endpoint 3584; ordered percentage arguments `(1,14,100)`. Applying the percentage after shifting would add `trunc(3584/100)=35` and return 3619, which fails this case.

- **MDP-V03** All progression boundaries (MDP-R02; boundary: isolated rule)
  - Inputs: increments `[1,2,3,4,5]`; independent L `[0,1,2,8,9,16,17,22,23,28,29,30]`
  - Seeds: `no draw`
  - Draws: none
  - Trace: completed band totals 7,23,41,65; then add active-band increments
  - Outputs: corresponding `[0,0,1,7,9,23,26,41,45,65,70,75]`; independent absent array/L30 returns 0
  - Discriminates: band thresholds and absent-array handling
  - runtime_observed: no

- **MDP-V04** Length levels (MDP-R03; boundary: isolated rule)
  - Inputs: ELen 25, length increments `[5,3,1]`, HitShift 8; independent levels `[0,1,8,9,16,20]`, stored missile level 30
  - Seeds: `no draw`
  - Draws: none
  - Trace: base 25, completed first band 35, completed second 24; L20 adds 4
  - Outputs: corresponding `[25,25,60,63,84,88]` despite HitShift 8; invalid class returns 0. Independent ELen -25 with the same increments and levels `[-1,1,8,20]` returns `[-25,-25,10,38]`: nonpositive level returns the signed base; L20 adds `35+24+4`. Independent ELen 2147483647, first increment 1, L2 returns -2147483648 by wrapping addition. Independent class 0/count 1 with absent selected row and L10 returns 0, without reading length columns or drawing.
  - Discriminates: no stored-level fallback, no HitShift
  - runtime_observed: no

- **MDP-V05** Bypass and request boundaries (MDP-R06/MDP-R12a; boundary: supplied-input stage)
  - Inputs: zero record, Holy 5; destination M, no source enumeration, L1
  - Seeds: `no draw`
  - Draws: none
  - Trace: bits 0/2 set undead/beast conditions; store after base request sequence
  - Outputs: per case; base requests follow R12a, stat 12=1 and all other base values 0, with no source enumeration
  - Discriminates: row-mask mapping and no absent-flag clear; actual stored stat 103/106=1 and absence of stat 104 remain MDP-Q06 integration assertions
  - runtime_observed: no
  - (a) Inputs: Holy 5. Draws: none. Outputs: bypassUndead/bypassBeasts true, bypassDemons false; ordered conditional requests `(M,103,0,1)`, `(M,106,0,1)`; no stat 104 request.
  - (b) Inputs: Holy 2. Draws: none. Outputs: bypassDemons true, bypassUndead/bypassBeasts false; only conditional request `(M,104,0,1)`, no stat 103/106 request.
  - (c) Inputs: Holy 128. Draws: none. Outputs: all three bypass conditions false; no conditional stat 103/104/106 request. The unrelated high bit has no effect.

- **MDP-V06** Critical draw order (MDP-R08a; boundary: supplied-input stage)
  - Inputs: source S, stat 337=0, no weapon; per-case stat 141
  - Seeds: S `{1,0x29A}`
  - Draws: per case, source S modulo 100
  - Trace: first NRF step gives `{0x6AC6935F,0}` and remainder 51; second gives `{0x2F2ED81B,0x2C890AFD}` and remainder 31
  - Outputs: per case
  - Discriminates: unconditional first draw, conditional second draw, seed high-word carry and short circuit; stored missile stat 141=1 remains MDP-Q06
  - runtime_observed: no
  - (a) Inputs: stat 141=40. Draws: `(S,100,51)`, `(S,100,31)`. Outputs: return 1, critical true, final S `{0x2F2ED81B,0x2C890AFD}`; R12a requests stat 141=1.
  - (b) Inputs: stat 141=0. Draws: `(S,100,51)`. Outputs: return 0, critical false, final S `{0x6AC6935F,0}`; no conditional stat 141 request.
  - (c) Inputs: stat 337=100,stat 141=40. Draws: `(S,100,51)`. Outputs: return 1, final S `{0x6AC6935F,0}`; stat 141/weapon not requested.
  - (d) Inputs: stat 141=-1, weapon W; supplied selector 2 mastery 100. Draws: `(S,100,51)`, `(S,100,31)`, `(S,100,12)`. Outputs: return 1, final S `{0x280687C4,0x13ADFE65}`; one mastery request `(S,W,0,2)` after failed second draw.
  - (e) Inputs: stat 337=-1, stat 141=0, no weapon. Draws: `(S,100,51)`. Outputs: return 0, final S `{0x6AC6935F,0}`; the negative first chance fails its signed comparison after the required first draw.
  - (f) Inputs: direct critical operation with absent contributing unit, no weapon. Seeds: none. Draws: no completed step. Trace: initial absent-unit chance is 0, then the first seed access is unguarded. Outputs: fault at seed access, no integer return, seed write, stat 141 read or weapon request.

- **MDP-V07** Share scaling (MDP-R11a; boundary: isolated scaling stage)
  - Inputs: fire minimum 300,fireLength 50, selected share per case; all other fields 0
  - Seeds: `no draw`
  - Draws: none
  - Trace: product wraps to i32 then signed /128; fireLength excluded
  - Outputs: per case
  - Discriminates: arithmetic right shift would floor a negative nonmultiple and wrongly scaling fireLength changes 50
  - runtime_observed: no
  - (a) Inputs: share 64. Draws: none. Outputs: fire minimum 150,fireLength 50.
  - (b) Inputs: share 64,fire minimum -3. Draws: none. Outputs: fire minimum -1 from `-192/128`,fireLength 50.
  - (c) Inputs: share 128. Draws: none. Outputs: fire minimum 300,fireLength 50 unchanged.

- **MDP-V08** Endpoint rejection and fallback (MDP-R01a; boundary: isolated rule)
  - Inputs: synthetic minimum base 3,increments `[1,2,3,4,5]`,empty calculation,HitShift 8; missile class 0/stored level 10
  - Seeds: `no draw`
  - Draws: none
  - Trace: level 0 with present missile falls back to 10; absent unit leaves level 0 and zero progression
  - Outputs: present missile,class -1,L0 ->3584; absent missile,class 0,L0 ->768; wrong-type unit ->0; absent missile,class -1 ->0; absent associated data,L0 ->768
  - Discriminates: fallback presence/type and nonpositive condition
  - runtime_observed: no

- **MDP-V09** Builder source selection and operation order (MDP-R04; boundary: supplied-input stage)
  - Inputs: source S,missile M,class 0,L3,valid records; supplied physical endpoints 10/20,element 30/40,length 50; mastery false; per-case Skill/MissileSkill/SrcDamage/SrcMissDmg
  - Seeds: `no draw`
  - Draws: none
  - Trace: clear record, choose branch, request five values in stated order
  - Outputs: per case; all fields still zero until the later assembly stage
  - Discriminates: source inclusion/mastery arguments, positive-only skill override and sentinel 255 branch
  - runtime_observed: no
  - (a) Inputs: Skill 0,MissileSkill false,SrcDamage 64,SrcMissDmg 32,EType 4. Draws: none. Outputs: ordered column requests min/max/element min/max/length `(M,S,0,3)`; selected values 10/20/30/40/50,type 4,physical share 64,parent share 32.
  - (b) Inputs: Skill 2,MissileSkill true,stored skill 7,SrcDamage 255,skill 7 SrcDam 128/EType 5. Draws: none. Outputs: stored-skill request, skill lookup 7; skill physical min/max `(S,7,3,includeSource0)`, then element min/max `(S,7,3,mastery1)`, then duration `(S,7,3)`; five selected values 10/20/30/40/50,type 5,physical/parent shares 0.
  - (c) Inputs: Skill 2,MissileSkill true,stored skill 0,SrcDamage 64,skill 2 SrcDam 128. Draws: none. Outputs: stored-skill request, lookup 2, same five ordered skill requests using skill 2; physical share 128,parent share 0.
  - (d) Inputs: missing source or wrong missile type or invalid missile class. Draws: none. Outputs: zero record, no endpoint/bypass requests.
  - (e) Inputs: Skill 2,missing skill record. Draws: none. Outputs: zero record, no endpoint/bypass requests after lookup.
  - (f) Inputs: Skill 0, MissileSkill false, SrcDamage 0, SrcMissDmg 0, EType 1, ApplyMastery true, source stat 329=50. Draws: none. Trace: after the five column requests, mastery minimum bonus `mulDiv(50,30,100)=15`, then maximum bonus `mulDiv(50,40,100)=20`. Outputs: selected physical endpoints 10/20, element endpoints 45/60, length 50, type 1, both shares 0; exact bonus request order `(S,1,30)`, `(S,1,40)`.
  - (g) Inputs: independent linked Skill -1 with MissileSkill false (column branch), or linked Skill 2 with MissileSkill true/stored skill -1, skill count 2. Draws: none. Outputs: first case makes the five column requests and selects 10/20/30/40/50 with both shares 0; second case keeps linked skill 2, lookup rejects equality with count, zero record and no endpoint/bypass requests.
  - (h) Inputs: linked Skill 2, MissileSkill true, stored skill -1, skill count 3, selected skill 2 present with SrcDam 128/EType 5, missile SrcDamage 64. Draws: none. Trace: stored result -1 is not positive, so keep linked skill 2, which is below count 3. Outputs: stored-skill request then lookup 2, physical minimum/maximum requests `(S,2,3,includeSource0)`, element minimum/maximum requests `(S,2,3,mastery1)`, duration request `(S,2,3)`; selected values 10/20/30/40/50, type 5, physical share 128, parent share 0. Choosing stored -1 would instead reject lookup and fail this successful request/value assertion.

- **MDP-V10** Mastery selection and rounding (MDP-R05; boundary: isolated rule)
  - Inputs: element value 7,source mastery 329=50,330=-30,331=25,332=100
  - Seeds: `no draw`
  - Draws: none
  - Trace: `mulDiv(50,7,100)=3`; `mulDiv(-30,7,100)=-2`; `mulDiv(25,7,100)=1`; `mulDiv(100,7,100)=7`
  - Outputs: per case
  - Discriminates: type 12 cold mastery and truncation before endpoint addition
  - runtime_observed: no
  - (a) Inputs: the listed type/value/mastery cases. Draws: none. Outputs: bonuses for independent types 1/2/4/12/5/3 respectively 3/-2/1/1/7/0; augmented values 10/5/8/8/14/7.
  - (b) Inputs: type 1, value 100001, source stat 329=65537. Draws: none. Trace: bonus `mulDiv(65537,100001,100)=655*100001=65500655`; add 100001. Outputs: bonus 65500655, augmented endpoint 65600656, ordered percentage arguments `(65537,100001,100)`; reversed operands yield augmented 65637001 and fail.
  - (c) Inputs: type 12, value 7, source stat 331=0. Draws: none. Outputs: bonus 0, augmented value 7; no percentage request.

- **MDP-V11** Source physical arithmetic (MDP-R07a; boundary: supplied-input stage)
  - Inputs: player S,physical share 128,`Half2HSrc` true,selected item W,two-handed result 1,mastery selector 1 result 10,strength 101,dexterity 51,signed coefficients 50/25,alternate-endpoint result 0,inventory present,classification 2,source stats 23=10/24=20,25=3,18=7,17=9,121=4,122=5,120=6,103=1/104=0/106=-1,type 57 result 1; critical chances 0
  - Seeds: S `{1,0x29A}`
  - Draws: `(S,100,51)`; supplied selector 2 mastery 0
  - Trace: weapon(S,1); critical fail; share becomes 64; percent 10+trunc(5050/100)+trunc(1275/100)+3=75; endpoints `(10<<8)*64/128=1280`, `(20<<8)*64/128=2560`
  - Outputs: sourcePhysical true,critical false,physical share 64; contribution endpoints 1280/2560,percent 75,competing bonuses 7/9,stat 121/122/120 fields 4/55/6,bypassUndead/Beasts true,Demons false; seed `{0x6AC6935F,0}`
  - Discriminates: source stats rather than weapon stats, weapon-selection condition 1, coefficient truncation and type 57 addend
  - runtime_observed: no

- **MDP-V12** Parent share and precedence (MDP-R09; boundary: supplied-input stage)
  - Inputs: physical share 0,parent share 64,parent P present,stats 21=300/22=500,25=4,18=7,17=9,121=2,122=3,120=5; critical chances 0
  - Seeds: P `{1,0x29A}`; S `{9,0}` unchanged
  - Draws: `(P,100,51)`
  - Trace: no weapon; `share(300,64)=150`, `share(500,64)=250`
  - Outputs: per case
  - Discriminates: parent endpoints already raw, parent seed owner and absent-parent short circuit
  - runtime_observed: no
  - (a) Inputs: listed parent inputs/chances 0. Draws: `(P,100,51)`. Outputs: the listed 150/250, percent 4, bonuses 7/9, fields 2/3/5, sourcePhysical/critical false; P `{0x6AC6935F,0}`, S unchanged.
  - (b) Inputs: listed inputs except parent absent. Draws: none. Outputs: contribution endpoints/percent/competing bonuses/three percent fields 0, sourcePhysical/critical false, both seeds unchanged.
  - (c) Inputs: parent present, share 2, parent stat 21=2147483647, stat 22=-2147483648, stat 337=100; other listed stats unchanged. Draws: `(P,100,51)`. Trace: endpoint products wrap to -2 and 0, then /128 gives 0/0; first critical draw succeeds. Outputs: endpoints 0/0, critical true, sourcePhysical false, percent 4, competing bonuses 7/9, percent fields 2/3/5; P `{0x6AC6935F,0}`, S unchanged; no stat 141 read.
  - (d) Inputs: parent P present with the listed nonzero stats, physical share 0, parent share 0. Draws: none. Outputs: contribution endpoints/percent/competing bonuses/three percent fields 0, sourcePhysical/critical false, P `{1,0x29A}` and S `{9,0}` unchanged; no parent critical request or parent stat read.

- **MDP-V13** Assembly and complete selector map (MDP-R10; boundary: isolated rule)
  - Inputs: contribution endpoints 10/20,row endpoints 3/4,contribution percent -120,missile stat 25=5,competing bonuses 7/9,element endpoints 30/40,length 50; per-case EType
  - Seeds: `no draw`
  - Draws: none
  - Trace: sum `-120+5+9=-106`, floor -90; endpoint sums 13/24; assign exactly the routing-table fields
  - Outputs: per case
  - Discriminates: all twelve table entries, no type 10 random selection, percent stored without endpoint multiplication
  - runtime_observed: no
  - (a) Inputs: the listed independent selector cases. Draws: none. Outputs: physical endpoints 13/24,physical percent -90 for every case. Independent types 1/2/3/4/5/6/7/8/11/12 assign listed min/max 30/40; types 1/4/5/11/12 assign listed length 50; type 9 assigns only stunLength 50; 0/10/13/255 assign no element field. Every unassigned field 0.
  - (b) Inputs: contribution percent 2147483647, missile stat 25=1, competing bonuses -7/-9, EType 0; listed endpoints unchanged. Draws: none. Trace: first addition wraps to -2147483648, then adding signed max -7 wraps to 2147483641. Outputs: physical percent 2147483641, physical endpoints 13/24, every element field 0.
  - (c) Inputs: contribution percent 0, missile stat 25=0, competing bonuses -7/-9, EType 0; listed endpoints unchanged. Draws: none. Outputs: physical percent -7, physical endpoints 13/24, every element field 0; the negative maximum is used rather than zero.

- **MDP-V14** Element additions, poison branch and scaling (MDP-R11a; boundary: supplied-input stage)
  - Inputs: physical share 64,source S,fire minimum 300,fireLength 50,poisonLength 20; source stats 48=2,56=10,57=30,58=50,101=7,59=99,326=8,60=12,62=14,316=16,317=18,315=20; other fields/stats 0
  - Seeds: `no draw`
  - Draws: none
  - Trace: log preparation(S,1); fire `(300+(2<<8))*64/128=406`; coldLength 10/2=5; poison min/max 30/2,50/2; positive 101 adds 7,skips 59/326; scale drain/burn
  - Outputs: fire minimum 406,fireLength 50,coldLength 5,poison min/max 15/25,poisonLength 27,poisonCount 0,life-drain minimum 6,mana-drain minimum 7,burn min/max 8/9,burnLength 10; one preparation request before stat 48 read. Independent stat 101=0 case: poisonLength 119,poisonCount 4; other outputs as above.
  - Discriminates: scaled preexisting elemental value, shifted vs unshifted additions and skipped poison count
  - runtime_observed: no

- **MDP-V15** Store order, markers and copied layers (MDP-R12a; boundary: supplied-input stage)
  - Inputs: destination M/type 3/associated metadata with both markers false; record sourcePhysical/secondaryMarker true,fireLength 50,burnLength 0,bypassUndead/Beasts/critical true,other fields 0; L7; source enumeration eligible and supplies ordered stat 180 entries `(layer2,value9)`, `(layer65535,value-3)`
  - Seeds: `no draw`
  - Draws: none; all supplied write/enumeration operations draw none
  - Trace: set both markers; log base replacing sequence, conditional unit writes, enumerate/copy
  - Outputs: per case
  - Discriminates: duplicate 315 order, zero request retention, no implied stored replacement, layer width/order
  - runtime_observed: no
  - (a) Inputs: listed record/destination/eligible enumeration. Draws: none. Outputs: markers true; base log stat ids `[12,21,22,48,49,315,50,51,52,53,54,55,56,57,58,59,326,60,61,62,63,64,65,66,316,317,315,121,122,120,25]`, all layer 0,stat 12 value 7,first 315 value 50,second 315 value 0,others 0; then unit requests 103=1,106=1,141=1 (no 104); enumerate(source,180,capacity 128), then list requests 180/layer 2=9 and 180/layer 65535=-3.
  - (b) Inputs: absent or wrong-type destination. Draws: none. Outputs: no marker write, no replacing or enumeration request.
  - (c) Inputs: listed record but destination metadata absent. Draws: none. Outputs: R12a's 31 base requests at layer 0, stat 12=7, first stat 315=50, second stat 315=0, others 0; then conditional unit requests 103=1,106=1,141=1 and no 104; enumeration `(source,180,capacity128)`, then list requests 180/layer 2=9 and 180/layer 65535=-3; no metadata write.
  - (d) Inputs: both destination markers already true; sourcePhysical/secondaryMarker and all bypass/critical conditions false; all numeric record fields 0; L7; source present but enumeration eligibility false. Draws: none. Outputs: both markers remain true; only R12a's 31 base requests occur with stat 12=7 and every other value 0, including both stat 315 requests; no conditional unit write, enumeration or stat 180 copy.
  - (e) Inputs: the shared record and destination, source absent; supplied eligibility/enumeration operations would report eligible/two entries if requested. Draws: none. Outputs: both destination markers set true; R12a's 31 base requests at layer 0 with stat 12=7, first stat 315=50, second stat 315=0, others 0; then unit requests 103=1,106=1,141=1 and no 104; no source eligibility query, enumeration or stat 180 copy.

- **MDP-V16** Halving to zero separates physical and elemental selection (MDP-R04/MDP-R07a/MDP-R09/MDP-R10/MDP-R11a; boundary: supplied-input builder stage)
  - Inputs: source player S, new missile M/class 0, parent missile P; L1; synthetic missile row Skill 0, MissileSkill false, SrcDamage 1, SrcMissDmg 64, `Half2HSrc` true, ApplyMastery false, physical bases 0/0, element bases 300/500, all increments 0, all calculations empty, HitShift 0, EType 1, ELen 50. Source selected weapon W/type item, two-handed predicate 1, alternate-endpoint predicate 0, inventory present, classification 1, type-57 predicate 0; mastery selectors 1/2 return 0 and both coefficients 0. Source stats 21=10,22=20,48=1000,49=2000,337=0,141=0. Parent stats 21=300,22=500,48=10,49=20,337=100; all other stats 0.
  - Seeds: S `{1,0x29A}`, P `{9,0}`
  - Draws: `(S,100,51)` only
  - Trace: original source share 1 selects source physical path and its one failed critical draw; halving stores current share 0, so source physical endpoints become 0 and parent physical path is not revisited. Routing supplies fire 300/500 and length 50. Element selection tests current share 0, takes parent share 64 with shift 0: `(300+10)*64/128=155`, `(500+20)*64/128=260`.
  - Outputs: sourcePhysical true, critical false, physical minimum/maximum 0/0, physical percent 0, current physical share 0, fire minimum/maximum 155/260, fireLength 50, every other numeric parameter field 0 and bypass/secondaryMarker false; S `{0x6AC6935F,0}`, P `{9,0}` unchanged. No parent critical request, no parent physical stat 21/22 read, no source-preparation request; elemental helper receives P/shift 0 and scaling receives 64.
  - Discriminates: a remembered physical source selection would consume source elemental stats/preparation; rerunning parent physical selection would consume the parent's seed and set critical. Both fail this fixture.
  - runtime_observed: no

- **MDP-V17** Source branches and coefficient interpretation (MDP-R07a; boundary: supplied-input stage)
  - Inputs: per case; defaults player S, share 128, selected item W, inventory present, `Half2HSrc` false, alternate predicate 0, classification 1, type-57 predicate 0, mastery selectors 1/2 return 0, signed coefficients 0, all stats and record fields 0
  - Seeds: S `{1,0x29A}` except coefficient-getter-only case (c), which has `no draw`
  - Draws: `(S,100,51)` except (c), none
  - Trace: per case; in (a)/(b)/(d), nonzero initial share sets sourcePhysical, the first critical draw fails and no further chance draws are needed; (c) executes only the isolated coefficient operations
  - Outputs: per case; all unlisted percent fields/competing bonuses/bypass/secondaryMarker are 0/false
  - Discriminates: no-item category dexterity, alternate endpoint priority, signed coefficient conversion and signed truncation
  - runtime_observed: no
  - (a) Inputs: source monster class 271, inventory absent, no weapon, source stats 21=3,22=5,2=17,25=4. Draws: `(S,100,51)`. Trace: CSA-R02 category nonzero; category dexterity 17 plus stat 25=4. Outputs: sourcePhysical true, critical false, contribution endpoints 768/1280, contribution percent 21, S `{0x6AC6935F,0}`; no weapon-selection, mastery, attribute-coefficient or classification request.
  - (b) Inputs: player defaults, alternate predicate 1, source stats 159=7,160=9,23=50,24=60. Draws: `(S,100,51)`. Outputs: sourcePhysical true, critical false, contribution endpoints 1792/2304, contribution percent 0, S `{0x6AC6935F,0}`; no classification request or source stat 23/24 read.
  - (c) Inputs: isolated coefficient helpers only, valid item/record with strength raw word 1 and dexterity raw word 65535; independent missing item record. Seeds: `no draw`. Draws: none. Trace: raw 1 maps 100; raw 65535 survives helper then caller sign-extension makes -1. Outputs: helper results 100/65535, caller coefficients 100/-1; missing record yields 100/100. This supplied-record fixture does not verify item-record production.
  - (d) Inputs: player defaults, strength stat 0=101, dexterity stat 2=51, signed coefficients -1/-100. Draws: `(S,100,51)`. Trace: `trunc(-101/100)=-1`, `trunc(-5100/100)=-51`. Outputs: sourcePhysical true, critical false, contribution endpoints 0/0, contribution percent -52, S `{0x6AC6935F,0}`.
  - (e) Inputs: source monster class 271, inventory present, monster weapon selector returns N/type player (not item), source stats 21=3,22=5,2=17,25=4; weapon critical mastery result 0. Draws: `(S,100,51)`. Trace: request monster weapon `(S)`, run critical with N, then the non-item weapon takes the no-item physical branch. Outputs: sourcePhysical true, critical false, contribution endpoints 768/1280, contribution percent 21, S `{0x6AC6935F,0}`; one selector-2 mastery request `(S,N,0,2)` within critical, no selector-1 mastery, coefficient, alternate-endpoint or classification request.

- **MDP-V18** Parent elemental additions and every scaling group (MDP-R11a; boundary: supplied-input stage)
  - Inputs: physical share 0, parent share 64, parent P present; every numeric parameter field 100 except life/mana/stamina drain maxima 101; all conditions false. Parent stats 48=2,49=4,50=6,51=8,52=10,53=12,54=14,55=16,56=18,57=20,58=22,101=0,59=24,326=26,60=28,62=30,316=32,317=34,315=36; other stats 0.
  - Seeds: P `{9,0}`, S `{1,0x29A}`
  - Draws: none
  - Trace: use parent/shift 0, add each mapped stat, then divide the listed scaled groups by 2; poisonLength adds 24 without scaling, poisonCount adds 26 before scaling
  - Outputs: per case
  - Discriminates: parent additions are unshifted, every listed scaling group is covered, and drain maxima/stamina/stun/physical/percent/poisonLength/fireLength exclusions remain unchanged.
  - runtime_observed: no
  - (a) Inputs: the listed parent share/stats and parameter record. Draws: none. Outputs: fire 51/52, lightning 53/54, magic 55/56, cold 57/58, coldLength 59, poison 60/61, poisonLength 124, poisonCount 63, life-drain minimum 64/maximum 101, mana-drain minimum 65/maximum 101, stamina-drain minimum 100/maximum 101, burn 66/67, burnLength 68; fireLength/stunLength/physical minimum/maximum/all four percent fields remain 100; conditions and both seeds unchanged; no source-preparation or critical request.
  - (b) Inputs: the same supplied record and present parent stats, physical share 0, parent share 0. Draws: none. Outputs: every numeric parameter field remains 100 except the three drain maxima remain 101; all conditions false, P `{9,0}` and S `{1,0x29A}` unchanged; no source preparation, parent stat read, scaling operation or critical request.

### 8.2 Observation vectors

- **MDP-O01** Creation critical decision (MDP-R08a/MDP-R12a). Read the contributing source seed, actual chances and conditional write requests across one eligible missile creation; compare the exact reached draw sequence and final seed, including one-draw failure and two-draw success cases. Stored stat 141=1 is an additional assertion requiring Q06's write evidence. [Tool-neutral reproduction procedure](../research/spec-evidence/missile-damage-parameters.md#critical-observation). Pending; this observation alone does not confirm all rules in Q07.

## 9. Acceptance and dependencies

Partial static acceptance consists of the verified supplied-input assertions in MDP-V01 through MDP-V18, including exact record results, requests and local seed changes. It excludes actual unresolved calculation/skill/equipment/mastery/preparation production, stat-180 enumeration and stat-write propagation. MDP-Q02, MDP-Q04 and MDP-Q06 gate those exact integrations; no gate on their separately supplied consumer arithmetic is implied. Full reproduction requires closing applicable integration assertions and matching observations for every behavioral rule under Q07. Passing a request log does not prove stored missile stats or complete parity; an archive cell is not gameplay observation.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), NRF-R01/R02 for seeds/draw reduction, NRF-R05 for operand-ordered mulDiv, NRF-R06a for signed wrapped share division, NRF-R08a for units, RLI-R02a/RLI-R02b for selected stat reads, RLI-Q03 for replacing-write integration; [core-game-data-fields](core-game-data-fields.md), GDA-R10/GDA-R11/GDF-R02 for fields and available stat names; [skill-calc-expressions](skill-calc-expressions.md), SCX-R29a/SCX-R29b and their referenced interpreter contracts; [skill-leveled-damage-tohit-and-elements](skill-leveled-damage-tohit-and-elements.md), SLD-R01 through SLD-R11/SLD-R19 for supplied skill endpoints and bands with their own gates; [combat-status-application](combat-status-application.md), CSA-R02 for fixed hireling category; [state-timed-state-lifecycle](state-timed-state-lifecycle.md), STL-R06c for replacing-write integration; [missile-initialization-and-registry](missile-initialization-and-registry.md), MIR-R12/MIR-R13 for caller identity, stored skill/level and invocation. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md), MHD-R10 for consuming the resulting missile stats; creation consumes the builder/store hand-off. Reverse consumers are not prerequisites for supplied arithmetic.

## 10. Related specs

- [missile-creation-and-seeding](missile-creation-and-seeding.md): creation sequence.
- [missile-initialization-and-registry](missile-initialization-and-registry.md): initialization and ownership.
- [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md): contact-time rolling and critical application.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): downstream stored-stat roll boundary and creation consumers.

## Appendix A. Provenance

Evidence targets Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies original D2Common and D2Game binaries. Compatible instruction/reconstruction records dated 2026-10-12 are reused with their recorded limits; targeted matching-reconstruction-first readings, original-image instruction/data decoding and descriptor corroboration dated 2026-10-15 support the additions in the [companion](../research/spec-evidence/missile-damage-parameters.md). Named field/stat dependencies retain their own data-profile limits. No gameplay observation is claimed; integration gaps and pending observations keep this document draft.
