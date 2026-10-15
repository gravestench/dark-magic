# Shield block and avoidance

**Description**

After an attack's hit roll succeeds, the defender gets a chance to block it with a shield or avoid it with Dodge, Avoid, Evade or weapon block. Shields are checked before passive avoidance. Movement changes which defenses apply, and missiles can use different checks from melee attacks. The result decides whether the attack lands, was blocked or was avoided.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: SBA-R, SBA-Q, SBA-V, SBA-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-shield-block-and-avoidance.md)
Open evidence gaps: [gap file](../plans/gaps/combat-shield-block-and-avoidance.md)

## 1. Scope

This contract covers what follows and frames the attack hit roll: whether the hit roll runs and with which arguments, shield block and Dodge, Avoid, Evade and weapon block, result codes/flags, missile reaction and knockback flags, seed ownership and the attack's complete draw order. It includes melee, missile unit contact and avoidance-only application. The hit roll itself (attack rating, defense, ratio, chance, cap, attacker draw and post-hit prevent-healing request) belongs to [combat-hit-roll](combat-hit-roll.md). Damage and later effects belong to [combat-damage-pipeline-order](combat-damage-pipeline-order.md); reactions/life consequences to [combat-hit-recovery](combat-hit-recovery.md); knockback effects to [combat-knockback](combat-knockback.md); later durability draws to [item-durability-loss](item-durability-loss.md). Hostility, range, owner resolution and equipment classification have explicit supplied-input integration boundaries below.

## 2. Trigger and ownership

D2Game (server) computes block, avoidance and result flags. D2Common computes block chance. D2Game UNIT/SUnitDmg is a source-folder lead rather than an assert-string identification.

Melee runs once per attack contact in a skill do-function: server skill event -> do dispatcher -> representative handler 46 -> skill hit-damage calculation -> melee result -> hit roll ([combat-hit-roll](combat-hit-roll.md)) -> block/avoidance. Missile contact runs in the missile step: contact handler -> optional hit roll -> damage fill -> missile result application. Avoidance-only application calls the avoidance consumer without shield block when the supplied result word is nonzero. Other callers and the actual hostility/range/owner integration remain SBA-Q08. Original locations and argument proofs are in the companion.

The hit roll, including its post-hit hook, runs before shield block. Missile application presets hit and rolls missile knockback before defender checks. Local consumers send no client message; subsequent damage/reaction processing conveys the outcome. No periodic cadence belongs here.

## 3. Inputs and data

Arithmetic is `i32`, wrapping addition, subtraction, multiplication and shift; division truncates toward zero, except unsigned seed modulus. Stat(u, id) is the selected layer-zero read under RLI-R02a/RLI-R02b. Layer enumeration uses the aggregate array under RLI-R02b, sorted by `u16` layer. A supplied operation logs requests and has only the effects declared by a vector. Supplied results never establish their producer. Hit-roll inputs are listed in [combat-hit-roll](combat-hit-roll.md) section 3.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| unit identity, class, mode, effective mode, inventory | Logical unit/game fields; NRF; GDF-R05a/GDF-R06a | `i32` and nullable associations | Types 0 player,1 monster,3 missile; effective mode uses its override when present |
| levels, Dexterity | 12 level,2 dexterity | `i32` | Block inputs |
| block | 20 toblock | `i32` | Block chance term |
| Dodge/Avoid/Evade | 338 passive_dodge,339 passive_avoid,340 passive_evade | `i32` percent | Positive values enable draws |
| weapon block | 348 passive_weaponblock; supplied `w4`/`w5`, item-type matches and classification | Up to 32 (`u16` layer, `i32` value); `i32` classification | Fold entries; classification 13 enables draw |
| state inputs | 54 uninterruptable | Bit | Reaction gates |
| class factor | CharStats.BlockFactor | `u8` | Supplied by class |
| monster flags | MonStats.NoShldBlock | Boolean | Definition predicate |
| attack arguments | bonus, nDamageParam, blockEnable, rangeAdjustment | `i32` | Bonus passed to the hit roll; nDamageParam 0 melee/nonzero missile |
| game/preconditions | expansion, frame, hostility/range results, owner-link validity, resolved owner, valid missile record | `i32`/booleans/unit | Supplied integration boundary |
| missile data | Missiles.ToHit, KnockBack, GetHit, SoftHit; missile stat 19 | `u8`, `u8`, bits, `i32` | Optional hit, knockback and flags |
| damage fields | resultWord, physicalDamage | `u16` flags, `i32` raw physical damage | Prior flags; physical !=0 enables missile shield |
| hit result | [combat-hit-roll](combat-hit-roll.md) HIT-R08 | `i32` 0/1 | Supplied hit-roll return |
| seeds | Defender, missile | Pairs of `u32` | NRF-R01/NRF-R02 |

| Mapping | Origin | Binding |
|---|---|---|
| Stat/state selectors, outcome codes, cap 75, movement divisor 3, classification 13 | Engine code | Fixed consumer identities, not inferred data remapping |
| Class factor/missile fields | Table data | Supplied row/column values |
| Type/equipment/token producers | Unresolved | Supplied results, with precise gates below |

Synthetic vectors supply values rather than assert actual table cells. Observed single named cells and their read limits remain in the companion.

## 4. Rules (execution order)

Execution: supplied preconditions -> hit roll ([combat-hit-roll](combat-hit-roll.md), including hook on success) -> shield block -> avoidance -> result map. Missile knockback precedes defender checks. No local rule draws unless it says so. Mastery/applier/precondition producers have separately owned draw contracts.

**SBA-R01a** Player block [runtime-confirmed (1.13c)]. Supplied shield absent returns 0 before stat/class reads. Otherwise T=Stat 20+CharStats.BlockFactor. expansion!=0 uses block=((Stat 2-15)*T)/(max(Stat 12,1)*2), else block=T. Cap signed block>=75 to 75; no lower floor. Missing list still adds class factor; invalid class faults. No draw. Recorded observation covers the scaled branch only.

**SBA-R01b** Monster/other block [verified in 1.13c code (static)]. Monster NoShldBlock true enables raw block; otherwise require supplied special result nonzero. Eligible block=min(Stat 20,75), signed; otherwise 0. Null/other types 0. No class/scaling/movement penalty. No draw.

**SBA-R01c** Special monster gate [unresolved]. Special classification can enable monster block when NoShldBlock false. Exact class/definition selection SBA-Q05; supplied result supports SBA-R01b independently.

**SBA-R01d** Expansion [verified in 1.13c code (static)]. Both melee/missile pass game.expansion unchanged to block scaling. No draw; classic observation pending, not an arithmetic gate.

**SBA-R01e** Shield selection [asm-only]. Supplied selection tries first candidate then second; candidate must match Any Shield(type 51) and pass two item exclusion conditions. Producer meanings/slots/type-equivalence remain SBA-Q10/SBA-Q11.

**SBA-R02** Movement block [verified in 1.13c code (static)]. After testing block>0, a player with effective mode 2/3/6 and raw mode!=2 divides capped block by 3. Others retain block. Positive 1/2 still draw after becoming 0. No direct draw.

**SBA-R03** Shield roll [runtime-confirmed (1.13c)]. blockEnable 0 or block<=0 before SBA-R02 skips to avoidance. Otherwise apply SBA-R02, step defender once, unsigned lo mod 100, assert attacker non-null (error 2602 after draw), then roll<block returns code 1 immediately; failure runs avoidance. No other defender mode/death test.

**SBA-R04a** Moving [verified in 1.13c code (static)]. Raw player modes 2/3 or other-unit 2/15 are moving. Read only Stat 340; c<=0 returns 0, no draw. Else draw defender mod 100, assert attacker, return 8 if roll<c else 0. No weapon/Dodge/Avoid test.

**SBA-R04b** Standing weapon block [verified in 1.13c code (static)]. Fold up to 32 sorted Stat 348 entries starting at v=0: layer 0 assigns value; layer>0 tests supplied `w5`/type match, else `w4`/type match; a match sets max(v, value). Only v>0 requests classification with defender, inventory, current-mode selector -1, enabled 1, initial output 0. Only output 13 draws defender mod 100 and asserts attacker; roll<v returns 16. Failure/disabled proceeds SBA-R04c. Equipment/token producers remain SBA-Q02/SBA-Q11; supplied fold/classification is independently verified.

**SBA-R04c** Standing Dodge/Avoid [verified in 1.13c code (static)]. nDamageParam 0 selects Stat 338/code 4, nonzero Stat 339/code 2. c<=0 returns 0, no draw. Else draw defender mod 100, assert attacker, return selected code if roll<c else 0.

**SBA-R04d** Consumer cap [verified in 1.13c code (static)]. No cap applies to c/v in these consumers: supplied 100/101 always succeed after a draw. Upstream passive production/capping is SBA-Q01, outside this consumer boundary.

**SBA-R04e** Token producer [unresolved]. Equipment and mode produce classification;13 enables weapon block. Full classifier/item selection is SBA-Q02. No descriptive meaning for 13 is assumed.

**SBA-R05** Melee word [verified in 1.13c code (static)]. After preconditions, word 1 if raw Run 3 player defender, else 0. If hit clear run hit roll, set hit on return 1. If hit set, run block/avoidance with blockEnable 1, nDamageParam 0. OR code 1=>16,2=>256,4=>128,16=>32768;8 maps no flag. Any nonzero code clears hit; block flag also clears hit. Remaining hit adds 4 unless state 54 active. Return low 16 bits.

**SBA-R06** Other maps [verified in 1.13c code (static)]. Missile contact optional hit uses isMissile 1, bonus=missile Stat 19. Accepted contact/valid owner+record: OR hit 1 into incoming word, run SBA-R11, block/avoidance with nDamageParam 1, blockEnable=(physicalDamage!=0). If code has bits in 31 and GetHit false and KnockBack 0, clear hit, OR 16384, discard code. Else map 8=>512,2=>256,1 or 16=>16; any nonzero code clears hit. Code 4 has no mapping and cannot arise with nDamageParam 1. Avoidance-only runs only for incoming word!=0, skips shield, uses nDamageParam 1; map 1=>16,2=>256,4=>128,16=>32768,8=>no flag; nonzero code clears hit. Prior unrelated flags survive.

**SBA-R07a** Skips [verified in 1.13c code (static)]. Raw Run 3 player melee defender skips hit roll, hook and attacker draw; defender checks still run. Missile ToHit 0 skips hit roll. Stat 115 does not skip the hit roll: it only zeroes defense under HIT-R02c, and defender checks still run.

**SBA-R07b** Other skill callers [unresolved]. Additional skill policies may bypass this consumer or supply other arguments. Complete caller coverage SBA-Q03; no local never-miss flag/default is established.

**SBA-R08** RNG order [verified in 1.13c code (static)]. All named draws use one NRF step then unsigned mod 100. Local attacker 0/1 hit roll (HIT-R08). Standing defender 0..3 in order shield, weapon, Dodge/Avoid, stop on success. Moving defender 0..2 shield, Evade. Missile 0/1 knockback before defender. Nonpositive chance skips except positive pre-penalty block reduced 0. Calculation/mastery/applier/precondition producers and later durability have separately owned draw contracts.

**SBA-R09a** Defender type differences [verified in 1.13c code (static)]. Player/monster block follow SBA-R01a/b. Attacker type differences are HIT-R10. No additional PvP arithmetic occurs locally.

**SBA-R09b** Upstream PvP [unresolved]. Additional caller/hostility policy remains SBA-Q04/SBA-Q08. Supplied preconditions do not verify it.

**SBA-R10** Attacker assertion [verified in 1.13c code (static)]. After each enabled weapon/Dodge/Avoid/Evade draw, assert attacker before comparison; null aborts 2602 after seed step. Non-null adds no effect/draw. Ordinary melee gate rejects null before this isolated failure boundary.

**SBA-R11** Missile flags [verified in 1.13c code (static)]. After preset hit, perform the following only when state 54 is absent: if GetHit is true, OR 4 into the result word; otherwise, if SoftHit is true, OR 16384. Within that same state-54-absent branch, a nonzero KnockBack draws missile mod 100 and roll<KnockBack ORs 8. When state 54 is present, skip both reaction flags and the missile draw. These OR operations preserve existing bits and precede shield block.

**SBA-R12** Melee preconditions [asm-only]. Null/non-player/non-monster inputs return 0 flags, no local draw. Otherwise supplied hostility/range must pass before resolution. Full reach/distance/line/hostility integration SBA-Q08.

**SBA-R13** Owner preconditions [asm-only]. Require missile, valid owner link, resolved owner, valid missile record. Failure before preset leaves incoming flags untouched, no local draw. Owner is attacker for assertions; avoidance-only uses a separate owner resolver. Lookup integration SBA-Q08.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SBA-R01a | Player block | runtime-confirmed (1.13c) | SBA-Q10 shield producer only | SBA-V01/SBA-V02/SBA-V03/SBA-V04/SBA-V17/SBA-V21 |
| SBA-R01b | Monster/other block | verified in 1.13c code (static) | SBA-Q05 special producer only | SBA-V05/SBA-V10/SBA-V11/SBA-V21 |
| SBA-R01c | Special monster gate | unresolved | SBA-Q05 special eligibility | none |
| SBA-R01d | Expansion | verified in 1.13c code (static) | none with supplied inputs | SBA-V17 |
| SBA-R01e | Shield selection | asm-only | SBA-Q10/SBA-Q11 shield execution | none |
| SBA-R02 | Movement block | verified in 1.13c code (static) | none with supplied inputs | SBA-V01/SBA-V02/SBA-V03/SBA-V09/SBA-V22 |
| SBA-R03 | Shield roll | runtime-confirmed (1.13c) | none with supplied inputs | SBA-V04/SBA-V06/SBA-V09/SBA-V14/SBA-V22 |
| SBA-R04a | Moving | verified in 1.13c code (static) | none with supplied inputs | SBA-V07/SBA-V08/SBA-V09/SBA-V16/SBA-V19/SBA-V23 |
| SBA-R04b | Standing weapon block | verified in 1.13c code (static) | SBA-Q02/SBA-Q11 token/equipment producers only | SBA-V12/SBA-V16/SBA-V18/SBA-V24 |
| SBA-R04c | Standing Dodge/Avoid | verified in 1.13c code (static) | none with supplied inputs | SBA-V06/SBA-V12/SBA-V16 |
| SBA-R04d | Consumer cap | verified in 1.13c code (static) | none with supplied inputs | SBA-V19 |
| SBA-R04e | Token producer | unresolved | SBA-Q02 token | none |
| SBA-R05 | Melee word | verified in 1.13c code (static) | SBA-Q08 actual preconditions only | SBA-V06/SBA-V12/SBA-V13 |
| SBA-R06 | Other maps | verified in 1.13c code (static) | SBA-Q08 actual contact/owner only | SBA-V08/SBA-V15/SBA-V16/SBA-V20 |
| SBA-R07a | Skips | verified in 1.13c code (static) | none with supplied inputs | SBA-V09/SBA-V20 |
| SBA-R07b | Other skill callers | unresolved | SBA-Q03 callers | none |
| SBA-R08 | RNG order | verified in 1.13c code (static) | none with supplied inputs | SBA-V06/SBA-V07/SBA-V09/SBA-V12/SBA-V15 |
| SBA-R09a | Defender type differences | verified in 1.13c code (static) | none with supplied inputs | SBA-V01/SBA-V05 |
| SBA-R09b | Upstream PvP | unresolved | SBA-Q04/SBA-Q08 upstream | none |
| SBA-R10 | Attacker assertion | verified in 1.13c code (static) | none with supplied inputs | SBA-V14 |
| SBA-R11 | Missile flags | verified in 1.13c code (static) | none with supplied inputs | SBA-V15 |
| SBA-R12 | Melee preconditions | asm-only | SBA-Q08 preconditions | none |
| SBA-R13 | Owner preconditions | asm-only | SBA-Q08 owner | none |

## 5. Outputs and side effects

| Output/effect | Rule |
|---|---|
| Block/avoidance code 0/1/2/4/8/16 and defender seed advance | SBA-R03/SBA-R04a/SBA-R04b/SBA-R04c |
| Melee returned word; missile/avoidance-only stored word | SBA-R05/SBA-R06 |
| Missile seed and reaction/knockback flags | SBA-R11 |
| Fatal 2602 after defender draw | SBA-R03/SBA-R10 |
| Invalid block class fault | SBA-R01a |

Block and helper returns are intermediates, not stored writes. No local message, unit creation or unit destruction occurs.

## 6. Branches and edge cases

Walk gets full block/moving Evade; Run gets one-third capped block/Evade; Town Walk gets one-third block/standing Dodge or Avoid. Block uses effective mode plus raw Walk exclusion; avoidance uses raw mode. Negative block returns unchanged and skips draw; positive 1/2 reduced 0 still draws. No difficulty term exists. Expansion changes player scaling; no additional local PvP arithmetic. Missing lists give zero/empty stats but retain class terms. Invalid monster definitions test false; invalid owner/missile record fails before preset; invalid player block class faults. Signed wrapped products and truncation remain applicable. Unresolved producer values are supplied explicitly, never inferred defaults.

## 7. Open questions

The [gap file](../plans/gaps/combat-shield-block-and-avoidance.md) records attempted readings and exact boundaries.

| Question | Status | Evidence that would settle it |
|---|---|---|
| **SBA-Q01** Passive producers | outside-boundary | Exact stat 338..340/348 setters/stacking/caps; none locally |
| **SBA-Q02** Classification 13 | integration | Complete mode/equipment classifier and item consumers |
| **SBA-Q03** Other skill skips | outside-boundary | Remaining root-to-consumer caller arguments; none locally |
| **SBA-Q04** Other PvP policy | outside-boundary | Upstream callers/hostility; none locally |
| **SBA-Q05** Special monster | behavior | Complete class selector and definition lookup |
| **SBA-Q06** Gameplay coverage | observation | Pending original-game observations |
| **SBA-Q07** Extra defender steps | observation | Subsequent durability trace, not seed correspondence alone |
| **SBA-Q08** Other roots/preconditions/owners | integration | Complete callers and hostile/range/owner contracts |
| **SBA-Q09** Elemental missile | observation | Physical-zero shield gate with physical control |
| **SBA-Q10** Shield selection | integration | Equipment candidates and exclusion consumers/producers |
| **SBA-Q11** Item types/slots | integration | Equivalence/secondary-type and slot population |

## 8. Test vectors

### 8.1 Executable vectors

Boundary for all vectors: supplied-input stage unless the vector stops at an isolated arithmetic/fold result. Inputs may name an intermediate block or successful hit to start at that stage; earlier producers are not executed. Defaults: non-null monster attacker/player defender, valid supplied records, passing hostility/range/owner checks, raw/effective mode 1, no mode override, no states, all unlisted stats zero, entries empty, shield absent, type matches false, classification 0, expansion 1, resultWord 0, KnockBack 0, GetHit/SoftHit false, blockEnable 1 and nDamageParam 0. An ordinary monster means type 1 with boss, hireling, demon, lUndead, hUndead, NoShldBlock and both neutral defense conditions false, special eligibility result 0 and valid class 5; unlisted monster flags use those defaults. Supplied operations log requests in order, draw nothing, and do not persist producer effects. No-weapon/supplied shield/token/type outcomes verify consumers only. Final seed is the successor after the listed draw count; every other seed remains unchanged. A supplied hit starts result mapping at hit bit 1 without a prior attacker draw. All numeric seeds/results below are exact.

Seed aliases in the vectors are exact pairs and successors: B={4,0x29A}, B1={0xAB1A45AE,1}, B2={0x62FD7EE7,0x475D97EE}; C={5,0x29A}, C1={0x15E0D673,2}, C2={0xC997B681,0x09201110}; D={8,0x29A}, D1={0x563488C2,3}, D2={0xFD665D4D,0x23F49E01}; E={12,0x29A}, E1={0x014ECBD6,5}, E2={0x97CF3BB3,0x008BA415}; F={2,0x29A}, F1={0xD58D2424,0}. Draw notation is owner/modulus/result, ordered left to right. Numeric outputs use decimal unless prefixed 0x. Case(a)/(b) labels below give independent input overrides, seeds, draws and outputs; no case inherits another case's outputs.

- **SBA-V01** (SBA-R01a/SBA-R02/SBA-R09a; boundary: supplied-input stage)
  - Inputs: player defender, shield=true, Stat 12=30, Stat 2=60, Stat 20=20, CharStats.BlockFactor 20, expansion 1, raw/effective Walk 2.
  - Seeds: no draw.
  - Draws: none.
  - Trace: T40;45*40/60=30; raw Walk excludes penalty.
  - Outputs: block 30.
  - Discriminates: Walk retains 30 instead of 10.
  - runtime_observed: no

- **SBA-V02** (SBA-R01a/SBA-R02; boundary: supplied-input stage)
  - Inputs: SBA-V01 inputs with raw/effective Run 3.
  - Seeds: no draw.
  - Draws: none.
  - Trace: block 30 then 30/3.
  - Outputs: block 10.
  - Discriminates: Run produces 10 instead of 30.
  - runtime_observed: no

- **SBA-V03** (SBA-R01a/SBA-R02; boundary: supplied-input stage)
  - Shared inputs: player, shield=true, Stat 12=30, Stat 2=100, Stat 20=40, BlockFactor 20, expansion 1
  - (a)
    - Inputs: raw/effective mode 2.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 85*60/60=85 capped 75; effective mode 2 is moving, but raw Walk 2 excludes the division.
    - Outputs: 75
  - (b)
    - Inputs: raw/effective mode 3.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 85*60/60=85 capped 75; moving Run with raw mode not Walk divides 75 by 3.
    - Outputs: 25
  - (c)
    - Inputs: raw/effective mode 6.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 85*60/60=85 capped 75; moving Town Walk with raw mode not Walk divides 75 by 3.
    - Outputs: 25
  - (d)
    - Inputs: raw/effective mode 1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 85*60/60=85 capped 75; effective mode 1 is not moving, so no division.
    - Outputs: 75.
  - Discriminates: Run 25 rejects dividing 85 before cap.
  - runtime_observed: no

- **SBA-V04** (SBA-R01a/SBA-R03; boundary: supplied-input stage)
  - Inputs: SBA-V01 inputs with Dex 10; all avoidance chances zero.
  - Seeds: no draw.
  - Draws: none.
  - Trace: -5*40/60=-3; positive block check fails.
  - Outputs: block -3, code 0, all seeds unchanged.
  - Discriminates: returned -3 rejects floor 0 or round -4.
  - runtime_observed: no

- **SBA-V05** (SBA-R01b/SBA-R09a; boundary: isolated rule)
  - Inputs: monster, NoShldBlock=true, Stat 20=80, raw/effective mode 15.
  - Seeds: no draw.
  - Draws: none.
  - Trace: min(80,75)=75; player movement penalty inapplicable.
  - Outputs: block 75.
  - Discriminates: monster 75 rejects player run penalty.
  - runtime_observed: no

- **SBA-V06** (SBA-R03/SBA-R04b/SBA-R04c/SBA-R05/SBA-R08; boundary: supplied-input stage)
  - (a)
    - Inputs: melee result stage, supplied hit=true, player raw/effective 1, block 30 after penalty, blockEnable 1, nDamageParam 0, Stat 338=20, entries 348[(0,20)] and classification 13.
    - Seeds: defender B
    - Draws: defender/100/10
    - Trace: 10<30 block
    - Outputs: code 1, word 16, B1; no weapon-fold or token request
  - (b)
    - Inputs: melee result stage, supplied hit=true, player raw/effective 1, block 30 after penalty, blockEnable 1, nDamageParam 0, Stat 338=20.
    - Seeds: defender C
    - Draws: defender/100/99 then defender/100/73
    - Trace: 99>=30,73>=20
    - Outputs: code 0, word 5, C2
  - (c)
    - Inputs: melee result stage, supplied hit=true, player raw/effective 1, block 30 after penalty, blockEnable 1, nDamageParam 0, Stat 338=20.
    - Seeds: defender D.
    - Draws: defender/100/58 then defender/100/5.
    - Trace: 58>=30,5<20 Dodge.
    - Outputs: code 4, word 128, D2; no token request in case.
  - (d)
    - Inputs: melee result stage, supplied hit=true, player raw/effective mode 1, block 1 after penalty, blockEnable 1, nDamageParam 0, Stat 348 entries[(0,1)], classification 13, Stat 338=1.
    - Seeds: defender C.
    - Draws: defender/100/99, then defender/100/73, then defender/100/5.
    - Trace: shield 99>=1 fails; positive weapon fold 1 requests classification 13, then 73>=1 fails; Dodge 5>=1 fails.
    - Outputs: code 0, word 5, defender {0x134E1255,0x54152EE9}; one classification request between shield and weapon draws, no further defender draw.
  - (e)
    - Inputs: Melee result stage, supplied hit=true, standing player raw/effective mode 1, block 0, entries 348 empty, classification 0, Stat 338=10.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Shield and weapon block are disabled; Dodge roll 10 equals chance 10, so the strict comparison fails; remaining hit sets flag 4.
    - Outputs: code 0, word 5, defender B1; no classification request.
  - (f)
    - Inputs: Melee result stage, supplied hit=true, standing player raw/effective mode 1, block 0, entries 348 empty, classification 0, Stat 338=-1.
    - Seeds: defender B.
    - Draws: none.
    - Trace: Nonpositive Dodge skips its draw; remaining hit sets flag 4.
    - Outputs: code 0, word 5, defender B unchanged; no classification request.
  - Discriminates: successful block consumes one defender draw and makes no weapon-fold/token request despite an enabled weapon-block fixture; three failed standing checks consume exactly three draws in shield/weapon/Dodge order. Case e rejects <= for Dodge, and case f retains the seed for a negative selected chance.
  - runtime_observed: no

- **SBA-V07** (SBA-R04a/SBA-R08; boundary: supplied-input stage)
  - (a)
    - Inputs: melee, player raw/effective Run 3, block 0, Stat 340=20, Stat 338=50.
    - Seeds: defender B
    - Draws: defender/100/10
    - Trace: Run presets hit; moving reads Evade 20, not Dodge 50.
    - Outputs: code 8, word 0, B1
  - (b)
    - Inputs: melee, player raw/effective Run 3, block 0, Stat 340=20, Stat 338=50.
    - Seeds: C; attacker unchanged.
    - Draws: defender/100/99.
    - Trace: Run presets hit; moving reads Evade 20, not Dodge 50.
    - Outputs: code 0, word 5, C1; zero attacker draws in each.
  - (c)
    - Inputs: Melee, player raw/effective Run 3, block 0, Stat 340=10, Stat 338=50.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Run presets hit; moving selects Evade 10. Equality 10<10 is false, and Dodge is not evaluated; remaining hit sets flag 4.
    - Outputs: code 0, word 5, defender B1; no attacker draw.
  - Discriminates: The successful Evade case maps no melee Evade flag; equality case c rejects <= and confirms moving failure does not fall through to standing Dodge.
  - runtime_observed: no

- **SBA-V08** (SBA-R04a/SBA-R06; boundary: supplied-input stage)
  - Inputs: missile apply, player raw/effective Walk 2, physicalDamage 0, Stat 339=50, Stat 340=0, GetHit=true, KnockBack 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: preset 1, GetHit adds 4; shield disabled, moving Evade 0.
  - Outputs: code 0, word 5, all seeds unchanged.
  - Discriminates: Avoid 50 produces no draw while moving.
  - runtime_observed: no

- **SBA-V09** (SBA-R07a/SBA-R02/SBA-R03/SBA-R04a/SBA-R08; boundary: supplied-input stage)
  - Shared inputs: melee, player raw/effective Run 3, blockEnable 1, Stat 340=20
  - Shared trace: preset hit; effective block 10 or 0
  - (a)
    - Inputs: block 30 before penalty
    - Seeds: defender B
    - Draws: defender/100/10 then defender/100/63
    - Trace: 10<10 false, Evade 63 fails
    - Outputs: word 5, B2
  - (b)
    - Inputs: block 30 before penalty
    - Seeds: defender E; attacker unchanged.
    - Draws: defender/100/6
    - Trace: 6<10 true
    - Outputs: word 16, E1
  - (c)
    - Inputs: block 2 before penalty.
    - Seeds: defender E; attacker unchanged.
    - Draws: defender/100/6 then defender/100/51.
    - Trace: pre-penalty 2 still draws with effective 0, Evade 51 fails.
    - Outputs: word 5, E2; zero attacker draws case.
  - Discriminates: block 2 still consumes its failed draw after penalty 0.
  - runtime_observed: no

- **SBA-V10** (SBA-R01b; boundary: isolated rule)
  - Inputs: monster, NoShldBlock=true, Stat 20=30.
  - Seeds: no draw.
  - Draws: none.
  - Trace: flag enables 30 without special classifier.
  - Outputs: block 30, no special-classifier request.
  - Discriminates: flag alone enables raw 30.
  - runtime_observed: no

- **SBA-V11** (SBA-R01b; boundary: supplied-input stage)
  - Inputs: monster, NoShldBlock=false, supplied special result 0, Stat 20=30, avoidance 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: both eligibility routes fail.
  - Outputs: block 0, code 0, no shield draw.
  - Discriminates: stat 30 alone does not enable block.
  - runtime_observed: no

- **SBA-V12** (SBA-R04b/SBA-R04c/SBA-R05/SBA-R08; boundary: supplied-input stage)
  - Shared inputs: melee result stage, supplied hit=true, player mode 1, block 0, w5 matches type T, w4 absent, Stat 348 entries[(0,5),(T,20)], Stat 338=30
  - Shared trace: for the unchanged shared entries, v assigns 5 then max 20; a case that replaces the entries uses its own fold trace.
  - (a)
    - Inputs: classification 13
    - Seeds: defender B
    - Draws: defender/100/10
    - Trace: weapon succeeds
    - Outputs: word 32768, B1
  - (b)
    - Inputs: classification 13
    - Seeds: defender C
    - Draws: defender/100/99 then defender/100/73
    - Trace: both fail
    - Outputs: word 5, C2
  - (c)
    - Inputs: classification 13
    - Seeds: defender D
    - Draws: defender/100/58 then defender/100/5
    - Trace: Dodge succeeds
    - Outputs: word 128, D2
  - (d)
    - Inputs: classification 0.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: weapon disabled, Dodge first draw succeeds.
    - Outputs: word 128, B1; v20, one token request.
  - (e)
    - Inputs: Stat 348 entries[(0,10)], classification 13, Stat 338=0; other shared inputs unchanged.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Layer zero assigns v=10; classification 13 enables one weapon draw; equality 10<10 is false, and zero Dodge draws nothing.
    - Outputs: v 10, code 0, word 5, defender B1; one classification request.
  - Discriminates: fold 20 rejects sum 25 and classification mismatch uses Dodge first draw. Case e rejects <= for weapon block without a later Dodge draw masking its result.
  - runtime_observed: no

- **SBA-V13** (SBA-R05; boundary: supplied-input stage)
  - Shared inputs: melee supplied hit=true, block 0, all avoidance 0
  - (a)
    - Inputs: state 54=false
    - Seeds: no draw.
    - Draws: none.
    - Trace: hit adds 4
    - Outputs: word 5
  - (b)
    - Inputs: state 54=true
    - Seeds: no draw.
    - Draws: none.
    - Trace: state suppresses 4
    - Outputs: word 1
  - (c)
    - Inputs: supplied code 8 at mapping, independent result-stage fixture.
    - Seeds: no draw.
    - Draws: none.
    - Trace: nonzero code clears hit without map.
    - Outputs: word 0.
  - (d)
    - Inputs: supplied code 2 at mapping, independent result-stage fixture.
    - Seeds: no draw.
    - Draws: none.
    - Trace: OR flag 256 and clear hit; no reaction flag is added.
    - Outputs: word 256.
  - Discriminates: melee code 8 maps no Evade 512.
  - runtime_observed: no

- **SBA-V14** (SBA-R03/SBA-R10; boundary: isolated rule)
  - Shared inputs: isolated defender-check operation, null attacker, standing melee defender
  - (a)
    - Inputs: block 20, blockEnable 1
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: enabled block or Dodge steps defender before null assertion.
    - Outputs: abort error 2602, defender B1, no returned code/word.
  - (b)
    - Inputs: block 0, Stat 338=30.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: enabled block or Dodge steps defender before null assertion.
    - Outputs: abort error 2602, defender B1, no returned code/word.
  - (c)
    - Inputs: Standing player defender, block 0, entries 348[(0,20)], classification 13, Stat 338=0.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Weapon chance 20 enables a draw; null attacker aborts before the otherwise successful 10<20 comparison.
    - Outputs: abort error 2602, defender B1, no returned code/word.
  - (d)
    - Inputs: Moving player defender raw/effective Run 3, block 0, Stat 340=20.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Evade chance 20 enables a draw; null attacker aborts before the otherwise successful 10<20 comparison.
    - Outputs: abort error 2602, defender B1, no returned code/word.
  - Discriminates: Block, Dodge, weapon block and Evade each abort after stepping the defender, before returning even when roll 10 would succeed.
  - runtime_observed: no

- **SBA-V15** (SBA-R11/SBA-R06/SBA-R08; boundary: supplied-input stage)
  - Shared inputs: missile apply, KnockBack 50, GetHit=true, physicalDamage 1, block 20, standing defender, avoidance 0
  - (a)
    - Inputs: state 54=false
    - Seeds: missile F, defender B.
    - Draws: missile/100/36 then defender/100/10
    - Trace: word 1->5->13, block OR 16 clears hit
    - Outputs: word 28, missile F1, defender B1
  - (b)
    - Inputs: state 54=true.
    - Seeds: missile F, defender B.
    - Draws: defender/100/10.
    - Trace: reaction/knockback skipped, block succeeds.
    - Outputs: word 16, missile F unchanged, defender B1.
  - (c)
    - Inputs: state 54=false, GetHit=false, SoftHit=true, KnockBack=0, block 0, all avoidance zero.
    - Seeds: missile F, defender B.
    - Draws: none.
    - Trace: Preset hit 1, then SoftHit OR 16384; no missile or defender check is enabled.
    - Outputs: word 16385, missile F unchanged, defender B unchanged.
  - (d)
    - Inputs: state 54=false, GetHit=true, SoftHit=true, KnockBack=0, block 0, all avoidance zero.
    - Seeds: missile F, defender B.
    - Draws: none.
    - Trace: Preset hit 1, GetHit OR 4; the else-SoftHit branch is skipped.
    - Outputs: word 5, missile F unchanged, defender B unchanged.
  - (e)
    - Inputs: Incoming resultWord 256, state 54=false, GetHit=true, SoftHit=false, KnockBack=0, block 0, all avoidance zero.
    - Seeds: missile F, defender B.
    - Draws: none.
    - Trace: Preset hit ORs 1 into 256, giving 257; GetHit ORs 4, giving 261; no later check changes it.
    - Outputs: word 261, missile F unchanged, defender B unchanged.
  - Discriminates: missile draw precedes block and its flag survives block.
  - runtime_observed: no

- **SBA-V16** (SBA-R06/SBA-R04a/SBA-R04b/SBA-R04c; boundary: supplied-input stage)
  - Shared inputs: missile apply, standing player, block 0, Stat 339=50, entries 348[(0,20)], classification 13, KnockBack 0, physicalDamage 1
  - (a)
    - Inputs: GetHit=false, SoftHit=false
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: weapon 16 discarded to 16384
    - Outputs: word 16384
  - (b)
    - Inputs: GetHit=true, entries empty
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Avoid 2 maps 256
    - Outputs: 260
  - (c)
    - Inputs: GetHit=true
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: weapon 16 maps 16
    - Outputs: 20
  - (d)
    - Inputs: GetHit=true, raw/effective Run 3, Stat 340=20
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Evade 8 maps 512
    - Outputs: 516
  - (e)
    - Inputs: GetHit=true, physicalDamage 0, block 60, entries empty.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: shield skipped, Avoid first draw.
    - Outputs: 260; defender B1.
  - (f)
    - Inputs: GetHit=false, SoftHit=false, KnockBack=1, physicalDamage=0, block 0, entries 348[(0,20)], classification 13.
    - Seeds: missile F, defender B.
    - Draws: missile/100/36 then defender/100/10.
    - Trace: Missile knockback fails 36<1, leaving word 1. Weapon block succeeds 10<20. Nonzero KnockBack disables the discard branch even though the knockback roll failed and GetHit is false. Map code 16 to block flag 16 and clear hit.
    - Outputs: word 16, missile F1, defender B1.
  - Discriminates: weapon maps block 16 on missile; physical zero skips shield.
  - runtime_observed: no

- **SBA-V17** (SBA-R01d/SBA-R01a; boundary: supplied-input stage)
  - Shared inputs: player, shield=true, Dex 60, level 30, expansion 0, BlockFactor 20
  - (a)
    - Inputs: Stat 20=20, raw/effective 2
    - Seeds: no draw.
    - Draws: none.
    - Trace: T40
    - Outputs: 40
  - (b)
    - Inputs: Stat 20=70, raw/effective 2
    - Seeds: no draw.
    - Draws: none.
    - Trace: T90 capped 75
    - Outputs: 75
  - (c)
    - Inputs: Stat 20=20, raw/effective 3.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 40/3=13.
    - Outputs: 13.
  - Discriminates: expansion 0 returns 40 rather than scaled 30.
  - runtime_observed: no

- **SBA-V18** (SBA-R04b; boundary: isolated rule)
  - Shared inputs: fold-only stage, w5 matches T, w4 absent
  - (a)
    - Inputs: entries 348[(0,30),(T,10)]
    - Seeds: no draw.
    - Draws: none.
    - Trace: initialize v=0; layer 0 assigns 30; matching T gives max(30,10)=30; stop at the fold return, before classification or drawing.
    - Outputs: v=30
  - (b)
    - Inputs: [(0,-5)]
    - Seeds: no draw.
    - Draws: none.
    - Trace: initialize v=0; layer 0 assigns -5; v<=0 skips classification and the weapon draw.
    - Outputs: v=-5, no token call
  - (c)
    - Inputs: [(U,25)], U unmatched.
    - Seeds: no draw.
    - Draws: none.
    - Trace: initialize v=0; U matches neither w5 nor w4, so v remains 0; v<=0 skips classification and the weapon draw.
    - Outputs: v=0, no token call.
  - Discriminates: maximum 30 rejects sum 40; negative layer 0 skips token.
  - runtime_observed: no

- **SBA-V19** No consumer cap (SBA-R04d/SBA-R04a; boundary: isolated rule)
  - Shared inputs: player raw/effective Run 3, block 0; isolated moving-avoidance stage.
  - (a)
    - Inputs: Stat 340=100.
    - Seeds: defender C.
    - Draws: defender/100/99.
    - Trace: 99<100 succeeds.
    - Outputs: code 8, defender C1.
  - (b)
    - Inputs: Stat 340=101.
    - Seeds: defender C.
    - Draws: defender/100/99.
    - Trace: 99<101 succeeds.
    - Outputs: code 8, defender C1.
  - Discriminates: a local cap 95 would fail roll 99 in either case.
  - runtime_observed: no

- **SBA-V20** Missile skip/avoidance-only (SBA-R07a/SBA-R06/SBA-R04b; boundary: supplied-input stage)
  - (a)
    - Inputs: missile contact ToHit 0, accepted contact, physicalDamage 0, all avoidance zero, GetHit/SoftHit false, KnockBack 0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: skip hit roll; missile application presets hit 1; no subsequent check draws.
    - Outputs: word 1, zero attacker draws, no hit hook.
  - (b)
    - Inputs: avoidance-only word 1, standing player, Stat 339=20.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Avoid returns 2; map flag 256 and clear hit.
    - Outputs: word 256, defender B1.
  - (c)
    - Inputs: avoidance-only word 0, standing player, Stat 339=20.
    - Seeds: defender B.
    - Draws: none.
    - Trace: zero-word guard bypasses avoidance.
    - Outputs: word 0, defender B unchanged.
  - (d)
    - Inputs: avoidance-only word 1, standing player, Stat 348 entries[(0,20)], classification 13, Stat 339=0.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: shield is not requested; weapon fold 20 and supplied classification 13 enable the successful weapon roll; avoidance-only maps code 16 to 32768 and clears hit.
    - Outputs: code 16, word 32768, defender B1; one classification request, no shield or missile-knockback request.
  - Discriminates: ToHit 0 skips rather than fails; zero-word guard prevents an otherwise enabled Avoid draw; avoidance-only weapon block uses 32768 rather than the missile map's 16 or discard result.
  - runtime_observed: no

- **SBA-V21** Block eligibility branches (SBA-R01a/SBA-R01b; boundary: supplied-input stage)
  - Shared inputs: Block-chance helper only.
  - (a)
    - Inputs: Player, shield=false, invalid class -1, Stat 20=80, Dex=100, level=30.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Shield check fails before invalid class access.
    - Outputs: block 0, no fault, no class/stat read.
  - (b)
    - Inputs: Monster, NoShldBlock=false, special eligibility result 1, Stat 20=30, raw/effective mode 15.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Special result enables raw stat 30; no player movement penalty.
    - Outputs: block 30.
  - (c)
    - Inputs: Player, shield=true, no stat list, valid class with CharStats.BlockFactor 20, expansion 1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: All stat reads return zero, but class factor remains:T=20. Dex 0 and level floor 1 give ((0-15)*20)/(1*2)=-150; no lower clamp.
    - Outputs: block -150.
  - (d)
    - Inputs: Player, shield=true, invalid class -1, Stat 20=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Shield passes; invalid class has no usable class record, so class-factor access faults.
    - Outputs: memory fault, no block result, no draw.
  - (e)
    - Inputs: Monster, NoShldBlock=true, Stat 20=-3.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Eligible raw block -3 is below the upper cap and is returned unchanged.
    - Outputs: block -3.
  - Discriminates: Shield absence guards class access; the monster special result is an alternative to the definition flag.
  - runtime_observed: no

- **SBA-V22** Effective/raw modes and disabled shield (SBA-R02/SBA-R03; boundary: supplied-input stage)
  - Shared inputs: Player, block 30 before movement penalty, no avoidance.
  - (a)
    - Inputs: SBA-R02-only stage, raw mode 2, effective mode 3 with override present.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Effective moving true, but raw Walk excludes penalty.
    - Outputs: effective block 30.
  - (b)
    - Inputs: SBA-R02-only stage, raw mode 1, effective mode 3 with override present.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Effective moving true and raw mode not Walk:30/3.
    - Outputs: effective block 10.
  - (c)
    - Inputs: SBA-R02-only stage, raw mode 3, effective mode 1 with override present.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Effective moving false despite raw Run.
    - Outputs: effective block 30.
  - (d)
    - Inputs: Block/avoidance stage, raw/effective mode 1, blockEnable 0.
    - Seeds: defender B.
    - Draws: none.
    - Trace: Disabled shield bypasses positive chance; zero avoidance also skips.
    - Outputs: code 0, defender B unchanged.
  - Discriminates: Mode disagreement separates the effective test from raw Walk exclusion; disabled positive block makes no draw.
  - runtime_observed: no

- **SBA-V23** Other moving units and nonpositive Evade (SBA-R04a; boundary: supplied-input stage)
  - Shared inputs: Isolated avoidance stage, block 0.
  - (a)
    - Inputs: Monster raw/effective mode 15, Stat 340=20.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Other-unit Run counts as moving;10<20.
    - Outputs: code 8, defender B1.
  - (b)
    - Inputs: Monster raw/effective mode 15, Stat 340=0.
    - Seeds: defender B.
    - Draws: none.
    - Trace: Nonpositive Evade skips the draw.
    - Outputs: code 0, defender B unchanged.
  - (c)
    - Inputs: Player raw/effective mode 3, Stat 340=-1.
    - Seeds: defender B.
    - Draws: none.
    - Trace: Negative Evade also skips the draw.
    - Outputs: code 0, defender B unchanged.
  - Discriminates: Other-unit Run uses Evade; zero/negative chances retain the seed.
  - runtime_observed: no

- **SBA-V24** Weapon candidate fallback/short circuit (SBA-R04b; boundary: supplied-input stage)
  - Shared inputs: Isolated standing avoidance, player raw/effective mode 1, block 0, T=7, entries 348[(T,20)], both w5 and w4 present, classification 13, Dodge/Avoid zero; supplied match operations log candidate and type.
  - (a)
    - Inputs: w5 does not match T, w4 matches T.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Query w5/T then w4/T;fold max(0,20)=20;request classification once;10<20.
    - Outputs: v 20, code 16, defender B1; type query log[w5/T,w4/T], one subsequent classification request.
  - (b)
    - Inputs: w5 matches T, w4 also matches T.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Query w5/T only;fold 20;request classification once;10<20.
    - Outputs: v 20, code 16, defender B1; type query log[w5/T], no w4 query, one subsequent classification request.
  - (c)
    - Inputs: Replace the shared entries with a supplied stat-348 enumerator fixture containing 33 sorted entries: layers 1..32 each value 1, layer 33 value 99. The fixture logs the requested maximum and returns the first min(maximum,33) entries. w5 matches each of these layers; w4 is present but is never needed. Classification 13, Dodge/Avoid zero.
    - Seeds: defender B.
    - Draws: defender/100/10.
    - Trace: Consumer requests maximum 32, so only layers 1..32 fold; v=1. Each w5 match skips w4. Classification 13 enables one draw;10<1 is false, and no standing passive draws.
    - Outputs: v 1, code 0, defender B1; enumeration maximum log[32], type-query log contains exactly w5/1 through w5/32 in order, one classification request; layer 33 never tested.
  - Discriminates: A w4-only match must work; a w5 match must skip w4 rather than query both.
  - runtime_observed: no

### 8.2 Observation vectors

Tool-neutral procedures and exact dated scope are in the [companion](../research/spec-evidence/combat-shield-block-and-avoidance.md#observations).

- **SBA-O01** Walk versus Run, block 62 versus 20 after caller penalty.
- **SBA-O02** Shield/Dex 10, negative chance, no shield draw.
- **SBA-O03** Shield before Dodge, success stops avoidance.
- **SBA-O04** Moving Dodge-only, no avoidance.
- **SBA-O05** Moving Avoid-only, no avoidance.
- **SBA-O06** Physical missile, failed block before standing Avoid.
- **SBA-O07** Observed 57 scaled-block calls return 62.
- **SBA-O08** Observed 34 blocks/27 failures, strict threshold 62, one defender step on block.
- **SBA-O09** Observed 4 total defender steps after failed block, subsequent attribution pending.
- **SBA-O10** Physical-zero missile skips shield; physical control reaches it.

## 9. Acceptance and dependencies

The verified supplied-input subset consists of SBA-V01 through SBA-V24 at their stated boundaries, including exact intermediates, flags, requests and seed changes. Passing it does not establish unresolved special-monster, shield, classification, caller, precondition or owner integration. Full reproduction requires those gates closed, the [combat-hit-roll](combat-hit-roll.md) contract, and pending gameplay observations. Observation gaps alone do not gate verified supplied-input arithmetic.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md) NRF-R01/NRF-R02 steps/modulus, RLI-R03/RLI-R02a/RLI-R02b list/state/stat reads; [data fields](core-game-data-fields.md) GDF-R01/GDA-R06a/GDA-R08/GDA-R10/GDF-R02/GDF-R03/GDF-R05a/GDF-R06a/GDF-R07/GDA-R16 names/schema/modes; [hit roll](combat-hit-roll.md) HIT-R01 through HIT-R10 for the hit return, attacker draw and post-hit request. Section 2 supplies upstream trigger/bonus/contact/precondition results; SBA-Q08 retains incomplete integration. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** [hit roll](combat-hit-roll.md) supplied shield result; [damage pipeline](combat-damage-pipeline-order.md) result flags; [hit recovery](combat-hit-recovery.md) reactions; [knockback](combat-knockback.md) knockback flag; [durability](item-durability-loss.md) standing hit result and seed position.

## 10. Related specs

[Hit roll](combat-hit-roll.md) owns attack rating, defense, chance and the attacker draw; [skill dispatch](skill-cast-start-and-do-dispatch.md) and [leveled damage/to-hit](skill-leveled-damage-tohit-and-elements.md) locate melee roots; [missile movement](missile-movement-and-collision.md) and [missile damage](missile-hit-and-damage-calculation.md) locate contact/fill; [PvP](combat-pvp-damage-and-kill-credit.md) owns hostility; [movement/freeze](state-movement-modifiers-and-freeze.md) shares mode identities.

## Appendix A. Provenance

Original LoD 1.13c Windows x86 D2Game, D2Common, Fog evidence uses the [source manifest](../research/source-manifest.md). The [companion](../research/spec-evidence/combat-shield-block-and-avoidance.md) preserves dated 2026-10-11..13 static/runtime evidence and correction rationale. Targeted 2026-10-15 reading used matched C where available, otherwise disassembly, including decoding checked against manifest-matching original bytes. Observed gameplay scope covers scaled player block and shield draw cases only; no new gameplay campaign ran.
