# Shield block and avoidance

**Description**

An attack first determines whether it hits, then gives the defender a chance to block or avoid it. Shields are checked before passive avoidance. Movement changes which defenses apply, and missiles can use different checks from melee attacks.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: SBA-R, SBA-Q, SBA-V, SBA-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-shield-block-and-avoidance.md)
Open evidence gaps: [gap file](../plans/gaps/combat-shield-block-and-avoidance.md)

## 1. Scope

This contract covers attack hit/miss with its defense and attack-rating terms, caps and attacker draw; shield block and Dodge, Avoid, Evade and weapon block; result codes/flags; the post-hit prevent-healing request; seed ownership and draw order. It includes melee, missile unit contact and avoidance-only application. Damage and later effects belong to [combat-damage-pipeline-order](combat-damage-pipeline-order.md); reactions/life consequences to [combat-hit-recovery](combat-hit-recovery.md); knockback effects to [combat-knockback](combat-knockback.md); later durability draws to [item-durability-loss](item-durability-loss.md). Hostility, range, owner resolution, equipment classification and state-applier effects have explicit supplied-input integration boundaries below.

## 2. Trigger and ownership

D2Game (server) computes hit, block, avoidance and result flags. D2Common computes defense, player attack base and block chance. D2Game UNIT/SUnitDmg is a source-folder lead rather than an assert-string identification.

Melee runs once per attack contact in a skill do-function: server skill event -> do dispatcher -> representative handler 46 -> skill hit-damage calculation -> melee result -> hit roll -> block/avoidance. Missile contact runs in the missile step: contact handler -> optional hit roll -> damage fill -> missile result application. Avoidance-only application calls the avoidance consumer without shield block when the supplied result word is nonzero. Other callers and the actual hostility/range/owner integration remain SBA-Q22. Original locations and argument proofs are in the companion.

The hit-roll hook runs before shield block. Missile application presets hit and rolls missile knockback before defender checks. Local consumers send no client message; subsequent damage/reaction processing conveys the outcome. No periodic cadence belongs here; the hook's state applier can schedule expiry under its dependency contract.

## 3. Inputs and data

Arithmetic is `i32`, wrapping addition, subtraction, multiplication and shift; division truncates toward zero, except unsigned seed modulus. Stat(u, id) is the selected layer-zero read under NRF-R12b/NRF-R12c. Layer enumeration uses the aggregate array under NRF-R12c, sorted by `u16` layer. Own-list reads are named explicitly. A supplied operation logs requests and has only the effects declared by a vector. Supplied results never establish their producer.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| unit identity, class, mode, effective mode, inventory | Logical unit/game fields; NRF; GDF-R14a/GDF-R15a | `i32` and nullable associations | Types 0 player,1 monster,3 missile; effective mode uses its override when present |
| levels, Dexterity | 12 level,2 dexterity | `i32` | Attack, defense and block inputs |
| attack/defense | 19 tohit,31 armorclass,32 armorclass_vs_missile,33 armorclass_vs_hth | `i32` | Base and contact terms |
| defense percentages | 171 skill_armor_percent,16 item_armor_percent,182 armor_override_percent | `i32` percent | Total-defense steps |
| block/attack percentages | 20 toblock,119 item_tohit_percent,115 item_ignoretargetac,116 item_fractionaltargetac | `i32` | Block, percent and defense adjustments |
| target bonuses | 123 item_demon_tohit,124 item_undead_tohit | `i32` | Flat AR additions |
| monster-type bonus | 179 attack_vs_montype | Up to 128 (`u16` layer, `i32` value) entries | Sum matching values; supplied signed `i16` monsterTypeIndex/count/relation matrix |
| mastery | 342 passive_mastery_melee_th; supplied selected weapon/meleeExtra | `i32` percent | Producer SBA-Q07 |
| Dodge/Avoid/Evade | 338 passive_dodge,339 passive_avoid,340 passive_evade | `i32` percent | Positive values enable draws |
| weapon block | 348 passive_weaponblock; supplied `w4`/`w5`, item-type matches and classification | Up to 32 (`u16` layer, `i32` value); `i32` classification | Fold entries; classification 13 enables draw |
| state/list inputs | 54 uninterruptable,101 holyshield; matching list's own 350 modifierlist_skill and 351 modifierlist_level | Bits, nullable list, `i32` | Reaction gates and Holy Shield evaluation |
| class factors | CharStats.ToHitFactor, CharStats.BlockFactor | `i32`, `u8` | Supplied by class |
| monster flags | MonStats.boss, NoShldBlock, demon, lUndead, hUndead; hireling classification | Bits/booleans | Definition and class predicates |
| neutral monster conditions | monsterDefenseExclusion, monsterDefenseHalf | Booleans; half is first component of exclusion | Unknown producer semantics SBA-Q18; absent monster data means false |
| attack arguments | bonus, isMissile, nDamageParam, blockEnable, rangeAdjustment | `i32` | Player bonus percent; other bonus flat; nDamageParam 0 melee/nonzero missile |
| game/preconditions | expansion, frame, hostility/range results, owner-link validity, resolved owner, valid missile record | `i32`/booleans/unit | Supplied integration boundary |
| missile data | Missiles.ToHit, KnockBack, GetHit, SoftHit; missile stat 19 | `u8`, `u8`, bits, `i32` | Optional hit, knockback and flags |
| damage fields | resultWord, physicalDamage | `u16` flags, `i32` raw physical damage | Prior flags; physical !=0 enables missile shield |
| seeds | Attacker, defender, missile | Pairs of `u32` | NRF-R01/NRF-R02 |
| prevent-healing request | stat 117 item_preventheal; target class; supplied state applier | `i32`, logical request | State 52, level 1, duration 120000 frames, effectIdentity 0, aux stat 31/value 0; effectIdentity meaning is SBA-Q27 |

| Mapping | Origin | Binding |
|---|---|---|
| Stat/state selectors, outcome codes, caps 5/95/75, movement divisor 3, classification 13 | Engine code | Fixed consumer identities, not inferred data remapping |
| Class factors/missile fields | Table data | Supplied row/column values |
| Names 117/123/124/342/350/351 | Table data | Explicit ItemStatCost.ID/Stat named cells; companion Table data |
| Type/equipment/mastery/token producers | Unresolved | Supplied results, with precise gates below |

Synthetic vectors supply values rather than assert actual table cells. Observed single named cells and their read limits remain in the companion.

## 4. Rules (execution order)

Execution: supplied preconditions -> hit roll (including hook on success) -> shield block -> avoidance -> result map. Missile knockback precedes defender checks. No local rule draws unless it says so. Expression/mastery/applier/precondition producers have separately owned draw contracts. For a player hit-roll path, evaluate R02a, then R02c/R02d/R02g, then acquire the optional R02f result and form R02b, before R03 and R04; the numbered sub-rules group the input terms rather than changing this call order.

**SBA-R01** Contact defense [verified in 1.13c code (static)]. DEF=TotalDefense(defender)+Stat(defender, isMissile ?32:33). No draw.

**SBA-R01a** Total defense [verified in 1.13c code (static)]. For a non-null unit, base=Stat(31)+Stat(2)/4; pct=Stat(171)+Stat(16)+holyTerm; prod=pct*base. If base>0 add prod/100; otherwise subtract prod/100. Read f=Stat(182); when nonzero add (base/100)*f if base>1048576, else (f/100)*base if f>65536, else (base*f)/100. Missing list reads zero. Null unit faults on a later unconditional unit access. No local draw.

**SBA-R01b** Holy Shield [verified in 1.13c code (static)]. holyTerm starts 0. Proceed only for types 0/1/3, active state 101, present aggregate-eligible base list and a matching state list. Read its own 350 and 351; both must be positive. A valid skill definition and supplied shield are required. Evaluate `Skills.calc1` in context(defender, skill, level) under SCX-R11, add the supplied signed result to pct before the product. Shield execution is SBA-Q24; requested expression producers retain their SCX gates/draws. No direct draw.

**SBA-R02a** Player base [verified in 1.13c code (static)]. AR=Stat(19)+5*Stat(2)-35+CharStats.ToHitFactor. Invalid class omits class term. Null/non-player helper input aborts error 4223; dispatcher calls only for players. No draw.

**SBA-R02b** Player percent [verified in 1.13c code (static)]. pct=meleeExtra+Stat(119)+bonus; meleeExtra 0 for isMissile!=0, otherwise SBA-R02f. A monster defender with monsterTypeIndex>0 enables up to 128 stat 179 entries. Add each value whose layer and index are both >0 and <count and whose supplied relation bit is set. Other defenders add none. No local draw. Producer integration SBA-Q07/SBA-Q16.

**SBA-R02c** Ignore defense [verified in 1.13c code (static)]. Player adjustment only: Stat 115!=0, monster defender, monsterDefenseExclusion=false, boss=false, hireling=false sets DEF 0. Otherwise retain DEF. Missing monster-specific data makes neutral conditions false. No draw.

**SBA-R02d** Reduce defense [verified in 1.13c code (static)]. After R02c, v=Stat 116; v<=0 changes nothing. Halve positive v for a player or a monster with boss, monsterDefenseHalf or hireling true. p=min(v,100). cut=(DEF/100)*p when DEF>1048576, else (DEF*p)/100; subtract cut. Negative DEF is not floored here. No draw.

**SBA-R02e** Other attackers [verified in 1.13c code (static)]. Non-player AR=Stat 19+bonus+5*Stat 2, pct=Stat 119. Skip player adjustments, mastery and type bonuses. No draw.

**SBA-R02f** Mastery producer [unresolved]. A selected weapon can produce meleeExtra. An absent selected weapon contributes zero. Exact selection and special-mastery results remain SBA-Q07; supplied meleeExtra supports caller arithmetic independently.

**SBA-R02g** Demon/undead [verified in 1.13c code (static)]. Player adjustment after R02d: add nonzero Stat 123 for demon, then nonzero Stat 124 for lUndead or hUndead. Both may add; negatives add unchanged. Non-monster/invalid definition predicates false. No draw.

**SBA-R03** Apply percent [verified in 1.13c code (static)]. Add mulDiv(pct, AR,100), NRF-R05 argument order: (AR/100)*pct if AR>1048576; else (pct/100)*AR if pct>65536; else (pct*AR)/100. No draw.

**SBA-R04** Negative swap [verified in 1.13c code (static)]. In order: DEF<0 adds -DEF to AR and sets DEF 0; AR<0 adds -AR to DEF and sets AR 0; DEF<0 then sets DEF 0. No draw.

**SBA-R05** Ratio [runtime-confirmed (1.13c)]. AR+DEF==0 yields 100, otherwise ratio=(AR*100)/(AR+DEF). No draw.

**SBA-R06** Level term [runtime-confirmed (1.13c)]. chance=((ratio*aLvl)<<1)/(aLvl+dLvl). Zero denominator faults. No draw.

**SBA-R07** Hit cap [verified in 1.13c code (static)]. chance<=5 becomes 5; else chance>=95 becomes 95. No draw.

**SBA-R08** Hit draw [verified in 1.13c code (static)]. Step attacker once, unsigned lo mod 100, return 1 if roll<chance, otherwise 0. A hit calls R18a before return. No defender draw.

**SBA-R09a** Player block [runtime-confirmed (1.13c)]. Supplied shield absent returns 0 before stat/class reads. Otherwise T=Stat 20+CharStats.BlockFactor. expansion!=0 uses block=((Stat 2-15)*T)/(max(Stat 12,1)*2), else block=T. Cap signed block>=75 to 75; no lower floor. Missing list still adds class factor; invalid class faults. No draw. Recorded observation covers the scaled branch only.

**SBA-R09b** Monster/other block [verified in 1.13c code (static)]. Monster NoShldBlock true enables raw block; otherwise require supplied special result nonzero. Eligible block=min(Stat 20,75), signed; otherwise 0. Null/other types 0. No class/scaling/movement penalty. No draw.

**SBA-R09c** Special monster gate [unresolved]. Special classification can enable monster block when NoShldBlock false. Exact class/definition selection SBA-Q12; supplied result supports R09b independently.

**SBA-R09d** Expansion [verified in 1.13c code (static)]. Both melee/missile pass game.expansion unchanged to block scaling. No draw; classic observation pending, not an arithmetic gate.

**SBA-R09e** Shield selection [asm-only]. Supplied selection tries first candidate then second; candidate must match Any Shield(type 51) and pass two item exclusion conditions. Producer meanings/slots/type-equivalence remain SBA-Q24/SBA-Q26.

**SBA-R10** Movement block [verified in 1.13c code (static)]. After testing block>0, a player with effective mode 2/3/6 and raw mode!=2 divides capped block by 3. Others retain block. Positive 1/2 still draw after becoming 0. No direct draw.

**SBA-R11** Shield roll [runtime-confirmed (1.13c)]. blockEnable 0 or block<=0 before R10 skips to avoidance. Otherwise apply R10, step defender once, unsigned lo mod 100, assert attacker non-null (error 2602 after draw), then roll<block returns code 1 immediately; failure runs avoidance. No other defender mode/death test.

**SBA-R12a** Moving [verified in 1.13c code (static)]. Raw player modes 2/3 or other-unit 2/15 are moving. Read only Stat 340; c<=0 returns 0, no draw. Else draw defender mod 100, assert attacker, return 8 if roll<c else 0. No weapon/Dodge/Avoid test.

**SBA-R12b** Standing weapon block [verified in 1.13c code (static)]. Fold up to 32 sorted Stat 348 entries starting at v=0: layer 0 assigns value; layer>0 tests supplied `w5`/type match, else `w4`/type match; a match sets max(v, value). Only v>0 requests classification with defender, inventory, current-mode selector -1, enabled 1, initial output 0. Only output 13 draws defender mod 100 and asserts attacker; roll<v returns 16. Failure/disabled proceeds R12c. Equipment/token producers remain SBA-Q06/SBA-Q26; supplied fold/classification is independently verified.

**SBA-R12c** Standing Dodge/Avoid [verified in 1.13c code (static)]. nDamageParam 0 selects Stat 338/code 4, nonzero Stat 339/code 2. c<=0 returns 0, no draw. Else draw defender mod 100, assert attacker, return selected code if roll<c else 0.

**SBA-R12d** Consumer cap [verified in 1.13c code (static)]. No cap applies to c/v in these consumers: supplied 100/101 always succeed after a draw. Upstream passive production/capping is SBA-Q03, outside this consumer boundary.

**SBA-R12e** Token producer [unresolved]. Equipment and mode produce classification;13 enables weapon block. Full classifier/item selection is SBA-Q06. No descriptive meaning for 13 is assumed.

**SBA-R13** Melee word [verified in 1.13c code (static)]. After preconditions, word 1 if raw Run 3 player defender, else 0. If hit clear run hit roll, set hit on return 1. If hit set, run block/avoidance with blockEnable 1, nDamageParam 0. OR code 1=>16,2=>256,4=>128,16=>32768;8 maps no flag. Any nonzero code clears hit; block flag also clears hit. Remaining hit adds 4 unless state 54 active. Return low 16 bits.

**SBA-R14** Other maps [verified in 1.13c code (static)]. Missile contact optional hit uses isMissile 1, bonus=missile Stat 19. Accepted contact/valid owner+record: OR hit 1 into incoming word, run R20, block/avoidance with nDamageParam 1, blockEnable=(physicalDamage!=0). If code has bits in 31 and GetHit false and KnockBack 0, clear hit, OR 16384, discard code. Else map 8=>512,2=>256,1 or 16=>16; any nonzero code clears hit. Code 4 has no mapping and cannot arise with nDamageParam 1. Avoidance-only runs only for incoming word!=0, skips shield, uses nDamageParam 1; map 1=>16,2=>256,4=>128,16=>32768,8=>no flag; nonzero code clears hit. Prior unrelated flags survive.

**SBA-R15a** Skips [verified in 1.13c code (static)]. Raw Run 3 player melee defender skips hit roll, hook and attacker draw; defender checks still run. Missile ToHit 0 skips hit roll. Stat 115 only zeroes defense, retaining cap/draw/defender checks.

**SBA-R15b** Other skill callers [unresolved]. Additional skill policies may bypass this consumer or supply other arguments. Complete caller coverage SBA-Q09; no local never-miss flag/default is established.

**SBA-R16** RNG order [verified in 1.13c code (static)]. All named draws use one NRF step then unsigned mod 100. Local attacker 0/1 hit roll. Standing defender 0..3 in order shield, weapon, Dodge/Avoid, stop on success. Moving defender 0..2 shield, Evade. Missile 0/1 knockback before defender. Nonpositive chance skips except positive pre-penalty block reduced 0. Calculation/mastery/applier/precondition producers and later durability have separately owned draw contracts.

**SBA-R17a** Type differences [verified in 1.13c code (static)]. Player attackers alone use R02a/b/c/d/f/g; others R02e. Player/monster block follow R09a/b. R02c/d state exact player/boss/hireling/neutral-condition differences. No additional PvP arithmetic occurs locally.

**SBA-R17b** Upstream PvP [unresolved]. Additional caller/hostility policy remains SBA-Q10/SBA-Q22. Supplied preconditions do not verify it.

**SBA-R18a** Hit hook [verified in 1.13c code (static)]. Successful hit: only player attacker, monster defender, Stat 117!=0, class outside inclusive 704..709 calls state applier(attacker, defender, state 52, level 1, duration 120000 frames, auxStat 31, value 0, effectIdentity 0, default removal callback). No local draw.

**SBA-R18b** Applier [asm-only]. Request can create/refresh timed prevent-healing list; complete state effects depend on STL-R06a/STL-R06b/STL-R06c/STL-R11 with their gates. Null target aborts 2602. Complete integration SBA-Q27; supplied applier verifies only request/order.

**SBA-R19** Attacker assertion [verified in 1.13c code (static)]. After each enabled weapon/Dodge/Avoid/Evade draw, assert attacker before comparison; null aborts 2602 after seed step. Non-null adds no effect/draw. Ordinary melee gate rejects null before this isolated failure boundary.

**SBA-R20** Missile flags [verified in 1.13c code (static)]. After preset hit, perform the following only when state 54 is absent: if GetHit is true, OR 4 into the result word; otherwise, if SoftHit is true, OR 16384. Within that same state-54-absent branch, a nonzero KnockBack draws missile mod 100 and roll<KnockBack ORs 8. When state 54 is present, skip both reaction flags and the missile draw. These OR operations preserve existing bits and precede shield block.

**SBA-R21** Melee preconditions [asm-only]. Null/non-player/non-monster inputs return 0 flags, no local draw. Otherwise supplied hostility/range must pass before resolution. Full reach/distance/line/hostility integration SBA-Q22.

**SBA-R22** Owner preconditions [asm-only]. Require missile, valid owner link, resolved owner, valid missile record. Failure before preset leaves incoming flags untouched, no local draw. Owner is attacker for assertions; avoidance-only uses a separate owner resolver. Lookup integration SBA-Q22.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SBA-R01 | Contact defense | verified in 1.13c code (static) | none with supplied inputs | SBA-V13 |
| SBA-R01a | Total defense | verified in 1.13c code (static) | SBA-Q24 shield producer only | SBA-V13/SBA-V14/SBA-V15/SBA-V33/SBA-V37 |
| SBA-R01b | Holy Shield | verified in 1.13c code (static) | SBA-Q24 shield; SCX requested producer gates | SBA-V33/SBA-V38 |
| SBA-R02a | Player base | verified in 1.13c code (static) | none with supplied inputs | SBA-V16/SBA-V39 |
| SBA-R02b | Player percent | verified in 1.13c code (static) | SBA-Q07/SBA-Q16 producer integration only | SBA-V16/SBA-V34 |
| SBA-R02c | Ignore defense | verified in 1.13c code (static) | SBA-Q18 producer only | SBA-V18 |
| SBA-R02d | Reduce defense | verified in 1.13c code (static) | SBA-Q18 producer only | SBA-V27 |
| SBA-R02e | Other attackers | verified in 1.13c code (static) | none with supplied inputs | SBA-V17 |
| SBA-R02f | Mastery producer | unresolved | SBA-Q07 mastery | none |
| SBA-R02g | Demon/undead | verified in 1.13c code (static) | none with supplied inputs | SBA-V28 |
| SBA-R03 | Apply percent | verified in 1.13c code (static) | none with supplied inputs | SBA-V16/SBA-V17/SBA-V29 |
| SBA-R04 | Negative swap | verified in 1.13c code (static) | none with supplied inputs | SBA-V04 |
| SBA-R05 | Ratio | runtime-confirmed (1.13c) | none with supplied inputs | SBA-V01/SBA-V02/SBA-V03/SBA-V04/SBA-V16/SBA-V18 |
| SBA-R06 | Level term | runtime-confirmed (1.13c) | none with supplied inputs | SBA-V01/SBA-V02/SBA-V03/SBA-V16/SBA-V40 |
| SBA-R07 | Hit cap | verified in 1.13c code (static) | none with supplied inputs | SBA-V02/SBA-V03/SBA-V18 |
| SBA-R08 | Hit draw | verified in 1.13c code (static) | none with supplied inputs | SBA-V01/SBA-V16/SBA-V24 |
| SBA-R09a | Player block | runtime-confirmed (1.13c) | SBA-Q24 shield producer only | SBA-V05/SBA-V06/SBA-V07/SBA-V08/SBA-V31/SBA-V41 |
| SBA-R09b | Monster/other block | verified in 1.13c code (static) | SBA-Q12 special producer only | SBA-V09/SBA-V20/SBA-V21/SBA-V41 |
| SBA-R09c | Special monster gate | unresolved | SBA-Q12 special eligibility | none |
| SBA-R09d | Expansion | verified in 1.13c code (static) | none with supplied inputs | SBA-V31 |
| SBA-R09e | Shield selection | asm-only | SBA-Q24/SBA-Q26 shield execution | none |
| SBA-R10 | Movement block | verified in 1.13c code (static) | none with supplied inputs | SBA-V05/SBA-V06/SBA-V07/SBA-V19/SBA-V42 |
| SBA-R11 | Shield roll | runtime-confirmed (1.13c) | none with supplied inputs | SBA-V08/SBA-V10/SBA-V19/SBA-V25/SBA-V42 |
| SBA-R12a | Moving | verified in 1.13c code (static) | none with supplied inputs | SBA-V11/SBA-V12/SBA-V19/SBA-V30/SBA-V35/SBA-V43 |
| SBA-R12b | Standing weapon block | verified in 1.13c code (static) | SBA-Q06/SBA-Q26 token/equipment producers only | SBA-V22/SBA-V30/SBA-V32/SBA-V44 |
| SBA-R12c | Standing Dodge/Avoid | verified in 1.13c code (static) | none with supplied inputs | SBA-V10/SBA-V22/SBA-V30 |
| SBA-R12d | Consumer cap | verified in 1.13c code (static) | none with supplied inputs | SBA-V35 |
| SBA-R12e | Token producer | unresolved | SBA-Q06 token | none |
| SBA-R13 | Melee word | verified in 1.13c code (static) | SBA-Q22 actual preconditions only | SBA-V10/SBA-V22/SBA-V23 |
| SBA-R14 | Other maps | verified in 1.13c code (static) | SBA-Q22 actual contact/owner only | SBA-V12/SBA-V26/SBA-V30/SBA-V36 |
| SBA-R15a | Skips | verified in 1.13c code (static) | none with supplied inputs | SBA-V18/SBA-V19/SBA-V36 |
| SBA-R15b | Other skill callers | unresolved | SBA-Q09 callers | none |
| SBA-R16 | RNG order | verified in 1.13c code (static) | none with supplied inputs | SBA-V10/SBA-V11/SBA-V19/SBA-V22/SBA-V26 |
| SBA-R17a | Type differences | verified in 1.13c code (static) | none with supplied inputs | SBA-V05/SBA-V09 |
| SBA-R17b | Upstream PvP | unresolved | SBA-Q10/SBA-Q22 upstream | none |
| SBA-R18a | Hit hook | verified in 1.13c code (static) | SBA-Q27 actual applier only | SBA-V24 |
| SBA-R18b | Applier | asm-only | SBA-Q27 effects | none |
| SBA-R19 | Attacker assertion | verified in 1.13c code (static) | none with supplied inputs | SBA-V25 |
| SBA-R20 | Missile flags | verified in 1.13c code (static) | none with supplied inputs | SBA-V26 |
| SBA-R21 | Melee preconditions | asm-only | SBA-Q22 preconditions | none |
| SBA-R22 | Owner preconditions | asm-only | SBA-Q22 owner | none |

## 5. Outputs and side effects

| Output/effect | Rule |
|---|---|
| Hit return 0/1 and attacker seed advance | SBA-R08 |
| Block/avoidance code 0/1/2/4/8/16 and defender seed advance | SBA-R11/SBA-R12a/SBA-R12b/SBA-R12c |
| Melee returned word; missile/avoidance-only stored word | SBA-R13/SBA-R14 |
| Missile seed and reaction/knockback flags | SBA-R20 |
| Ordered applier request; actual timed-state effects gated | SBA-R18a/SBA-R18b |
| Fatal 2602 after defender draw; fatal 4223 for invalid player-base input | SBA-R11/SBA-R19/SBA-R02a |
| Zero level denominator fault; invalid block class/null defense faults | SBA-R06/SBA-R09a/SBA-R01a |

AR, DEF, pct, chance, block and helper returns are intermediates, not stored writes. No local message, unit creation or unit destruction occurs.

## 6. Branches and edge cases

Walk gets full block/moving Evade; Run gets one-third capped block/Evade; Town Walk gets one-third block/standing Dodge or Avoid. Block uses effective mode plus raw Walk exclusion; avoidance uses raw mode. Negative block returns unchanged and skips draw; positive 1/2 reduced 0 still draws. No difficulty term exists. Expansion changes player scaling; no additional local PvP arithmetic. Missing lists give zero/empty stats but retain class terms. Invalid monster definitions test false; invalid owner/missile record fails before preset; invalid player block class faults. Signed wrapped products and truncation remain applicable. Unresolved producer values are supplied explicitly, never inferred defaults.

## 7. Open questions

The [gap file](../plans/gaps/combat-shield-block-and-avoidance.md) records attempted readings and exact boundaries.

| Question | Status | Evidence that would settle it |
|---|---|---|
| **SBA-Q03** Passive producers | outside-boundary | Exact stat 338..340/348 setters/stacking/caps; none locally |
| **SBA-Q06** Classification 13 | integration | Complete mode/equipment classifier and item consumers |
| **SBA-Q07** Melee mastery | behavior | Full selected-weapon and special mastery paths |
| **SBA-Q09** Other skill skips | outside-boundary | Remaining root-to-consumer caller arguments; none locally |
| **SBA-Q10** Other PvP policy | outside-boundary | Upstream callers/hostility; none locally |
| **SBA-Q12** Special monster | behavior | Complete class selector and definition lookup |
| **SBA-Q16** Monster relation producer | integration | Named loader field and matrix producer |
| **SBA-Q18** Neutral monster conditions | outside-boundary | Spawn flag writers; supplied conditions usable |
| **SBA-Q19** Gameplay coverage | observation | Pending original-game observations |
| **SBA-Q20** Extra defender steps | observation | Subsequent durability trace, not seed correspondence alone |
| **SBA-Q22** Other roots/preconditions/owners | integration | Complete callers and hostile/range/owner contracts |
| **SBA-Q23** Elemental missile | observation | Physical-zero shield gate with physical control |
| **SBA-Q24** Shield selection | integration | Equipment candidates and exclusion consumers/producers |
| **SBA-Q26** Item types/slots | integration | Equivalence/secondary-type and slot population |
| **SBA-Q27** State application | integration | Close CSA/STL attach/write/removal for this request |

## 8. Test vectors

### 8.1 Executable vectors

Boundary for all vectors: supplied-input stage unless the vector stops at an isolated arithmetic/fold result. Inputs may name an intermediate AR, DEF, block or successful hit to start at that stage; earlier producers are not executed. Defaults: non-null monster attacker/player defender, valid supplied records, passing hostility/range/owner checks, raw/effective mode 1, no mode override, no states, all unlisted stats zero, entries empty, shield absent, type matches false, classification 0, expansion 1, resultWord 0, KnockBack 0, GetHit/SoftHit false, blockEnable 1 and nDamageParam 0. An ordinary monster means type 1 with boss, hireling, demon, lUndead, hUndead, NoShldBlock and both neutral defense conditions false, special eligibility result 0 and valid class 5; unlisted monster flags use those defaults. Supplied operations log requests in order, draw nothing, and do not persist producer effects; the state applier logs its request only. No-weapon/supplied shield/token/type outcomes verify consumers only. Final seed is the successor after the listed draw count; every other seed remains unchanged. A supplied hit starts result mapping at hit bit 1 without a prior attacker draw. All numeric seeds/results below are exact.

Seed aliases in the vectors are exact pairs and successors: A={1,0x29A}, A1={0x6AC6935F,0}, A2={0x2F2ED81B,0x2C890AFD}; B={4,0x29A}, B1={0xAB1A45AE,1}, B2={0x62FD7EE7,0x475D97EE}; C={5,0x29A}, C1={0x15E0D673,2}, C2={0xC997B681,0x09201110}; D={8,0x29A}, D1={0x563488C2,3}, D2={0xFD665D4D,0x23F49E01}; E={12,0x29A}, E1={0x014ECBD6,5}, E2={0x97CF3BB3,0x008BA415}; F={2,0x29A}, F1={0xD58D2424,0}. Draw notation is owner/modulus/result, ordered left to right. Numeric outputs use decimal unless prefixed 0x. Case(a)/(b) labels below give independent input overrides, seeds, draws and outputs; no case inherits another case's outputs.

- **SBA-V01** (SBA-R05/SBA-R06/SBA-R08; boundary: supplied-input stage)
  - (a)
    - Inputs: AR=1000, DEF=500, aLvl=dLvl=30.
    - Seeds: attacker A
    - Draws: attacker/100/51
    - Trace: ratio=100000/1500=66; chance=3960/60=66.
    - Outputs: return 1, hook called, no applier request, attacker A1
  - (b)
    - Inputs: AR=1000, DEF=500, aLvl=dLvl=30.
    - Seeds: attacker C.
    - Draws: attacker/100/99.
    - Trace: ratio=100000/1500=66; chance=3960/60=66.
    - Outputs: return 0, no hook, attacker C1.
  - Discriminates: defender seed unchanged rejects using defender for hit draw.
  - runtime_observed: no

- **SBA-V02** (SBA-R05/SBA-R06/SBA-R07; boundary: supplied-input stage)
  - Inputs: AR=100, DEF=5000, aLvl=dLvl=30.
  - Seeds: attacker A.
  - Draws: attacker/100/51.
  - Trace: ratio=10000/5100=1; chance 1 capped 5.
  - Outputs: chance 5, return 0, attacker A1.
  - Discriminates: chance 5 rejects omitted lower cap.
  - runtime_observed: no

- **SBA-V03** (SBA-R05/SBA-R06/SBA-R07; boundary: supplied-input stage)
  - (a)
    - Inputs: monster attacker, player defender, AR=5000, DEF=0, aLvl 40, dLvl 10.
    - Seeds: attacker A
    - Draws: attacker/100/51
    - Trace: ratio 100; raw chance 160 capped 95.
    - Outputs: chance 95, return 1, A1
  - (b)
    - Inputs: monster attacker, player defender, AR=5000, DEF=0, aLvl 40, dLvl 10.
    - Seeds: attacker C.
    - Draws: attacker/100/99.
    - Trace: ratio 100; raw chance 160 capped 95.
    - Outputs: chance 95, return 0, C1.
  - Discriminates: roll 99 misses despite raw chance 160.
  - runtime_observed: no

- **SBA-V04** (SBA-R04/SBA-R05; boundary: isolated rule)
  - (a)
    - Inputs: AR 800, DEF -200
    - Seeds: no draw.
    - Draws: none.
    - Trace: add 200 to AR
    - Outputs: AR 1000, DEF 0, ratio 100
  - (b)
    - Inputs: AR -300, DEF 500
    - Seeds: no draw.
    - Draws: none.
    - Trace: add 300 to DEF
    - Outputs: AR 0, DEF 800, ratio 0
  - (c)
    - Inputs: AR -100, DEF -50
    - Seeds: no draw.
    - Draws: none.
    - Trace: AR becomes -50 then add 50 to DEF
    - Outputs: AR 0, DEF 50, ratio 0
  - (d)
    - Inputs: AR 0, DEF 0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: zero ratio denominator branch.
    - Outputs: AR 0, DEF 0, ratio 100.
  - (e)
    - Inputs: AR 0, DEF -2147483648.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Negating the minimum i32 wraps to itself: first transfer makes AR -2147483648, DEF 0; second makes AR 0, DEF -2147483648; final negative-defense clamp sets DEF 0.
    - Outputs: AR 0, DEF 0, ratio 100.
  - Discriminates: case(c) rejects flooring negatives independently.
  - runtime_observed: no

- **SBA-V05** (SBA-R09a/SBA-R10/SBA-R17a; boundary: supplied-input stage)
  - Inputs: player defender, shield=true, Stat 12=30, Stat 2=60, Stat 20=20, CharStats.BlockFactor 20, expansion 1, raw/effective Walk 2.
  - Seeds: no draw.
  - Draws: none.
  - Trace: T40;45*40/60=30; raw Walk excludes penalty.
  - Outputs: block 30.
  - Discriminates: Walk retains 30 instead of 10.
  - runtime_observed: no

- **SBA-V06** (SBA-R09a/SBA-R10; boundary: supplied-input stage)
  - Inputs: SBA-V05 inputs with raw/effective Run 3.
  - Seeds: no draw.
  - Draws: none.
  - Trace: block 30 then 30/3.
  - Outputs: block 10.
  - Discriminates: Run produces 10 instead of 30.
  - runtime_observed: no

- **SBA-V07** (SBA-R09a/SBA-R10; boundary: supplied-input stage)
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

- **SBA-V08** (SBA-R09a/SBA-R11; boundary: supplied-input stage)
  - Inputs: SBA-V05 inputs with Dex 10; all avoidance chances zero.
  - Seeds: no draw.
  - Draws: none.
  - Trace: -5*40/60=-3; positive block check fails.
  - Outputs: block -3, code 0, all seeds unchanged.
  - Discriminates: returned -3 rejects floor 0 or round -4.
  - runtime_observed: no

- **SBA-V09** (SBA-R09b/SBA-R17a; boundary: isolated rule)
  - Inputs: monster, NoShldBlock=true, Stat 20=80, raw/effective mode 15.
  - Seeds: no draw.
  - Draws: none.
  - Trace: min(80,75)=75; player movement penalty inapplicable.
  - Outputs: block 75.
  - Discriminates: monster 75 rejects player run penalty.
  - runtime_observed: no

- **SBA-V10** (SBA-R11/SBA-R12b/SBA-R12c/SBA-R13/SBA-R16; boundary: supplied-input stage)
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

- **SBA-V11** (SBA-R12a/SBA-R16; boundary: supplied-input stage)
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

- **SBA-V12** (SBA-R12a/SBA-R14; boundary: supplied-input stage)
  - Inputs: missile apply, player raw/effective Walk 2, physicalDamage 0, Stat 339=50, Stat 340=0, GetHit=true, KnockBack 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: preset 1, GetHit adds 4; shield disabled, moving Evade 0.
  - Outputs: code 0, word 5, all seeds unchanged.
  - Discriminates: Avoid 50 produces no draw while moving.
  - runtime_observed: no

- **SBA-V13** (SBA-R01a/SBA-R01; boundary: isolated rule)
  - Inputs: defender Stat 31=100, Stat 2=50, Stat 171=0, Stat 16=50, Stat 182=0, state 101=false, Stat 33=0, Stat 32=20.
  - Seeds: no draw.
  - Draws: none.
  - Trace: base 100+50/4=112; pct 50; prod 5600; base 168.
  - Outputs: TotalDefense 168, melee DEF 168, missile DEF 188.
  - Discriminates: missile 188 rejects using melee defense 168.
  - runtime_observed: no

- **SBA-V14** (SBA-R01a; boundary: isolated rule)
  - Inputs: Stat 31=-20, Stat 2=0, Stat 171=50, Stat 16=0, Stat 182=0, state 101=false.
  - Seeds: no draw.
  - Draws: none.
  - Trace: prod -1000; negative base subtracts -10.
  - Outputs: TotalDefense -10.
  - Discriminates: negative base -10 rejects adding percent to obtain -30.
  - runtime_observed: no

- **SBA-V15** (SBA-R01a; boundary: supplied-input stage)
  - Shared inputs: base after first percent 168
  - (a)
    - Inputs: Stat 182=50
    - Seeds: no draw.
    - Draws: none.
    - Trace: small product branch adds 84.
    - Outputs: 252
  - (b)
    - Inputs: Stat 182=-50.
    - Seeds: no draw.
    - Draws: none.
    - Trace: small product branch subtracts 84.
    - Outputs: 84.
  - Discriminates: final 252 rejects replacing defense with modifier 84.
  - runtime_observed: no

- **SBA-V16** (SBA-R02a/SBA-R02b/SBA-R03/SBA-R05/SBA-R06/SBA-R08; boundary: supplied-input stage)
  - Shared inputs: player attacker, Stat 19=100, Stat 2=50, CharStats.ToHitFactor 20, Stat 119=20, bonus 10, stats 115/116/123/124/117 zero, no 179 entries, player defender, DEF 168, aLvl=dLvl 20
  - (a)
    - Inputs: isMissile 1
    - Seeds: attacker A independently in case.
    - Draws: attacker/100/51 in case.
    - Trace: AR 335, pct 30; mulDiv 100; AR 435; ratio 43500/603=72; chance 72.
    - Outputs: return 1, AR 435, pct 30, ratio 72, chance 72, attacker A1; no applier request.
  - (b)
    - Inputs: isMissile 0, selected weapon absent.
    - Seeds: attacker A independently in case.
    - Draws: attacker/100/51 in case.
    - Trace: AR 335, pct 30; mulDiv 100; AR 435; ratio 43500/603=72; chance 72.
    - Outputs: return 1, AR 435, pct 30, ratio 72, chance 72, attacker A1; no applier request.
  - (c)
    - Inputs: isMissile 1, selected weapon present, supplied mastery operation would return meleeExtra 30 if requested.
    - Seeds: attacker A.
    - Draws: attacker/100/51.
    - Trace: Missile skips the supplied mastery operation; AR 335, pct 30, percent amount 100, AR 435, ratio 72, chance 72.
    - Outputs: return 1, AR 435, pct 30, ratio/chance 72, attacker A1; no mastery or applier request.
  - (d)
    - Inputs: isMissile 0, selected weapon present, supplied mastery operation returns meleeExtra 30 with no draw.
    - Seeds: attacker A.
    - Draws: attacker/100/51.
    - Trace: Request mastery once; AR 335, pct 30+30=60, percent amount 201, AR 536; ratio 53600/704=76, chance 76.
    - Outputs: return 1, AR 536, pct 60, ratio/chance 76, attacker A1; one mastery request before percentage arithmetic, no applier request.
  - Discriminates: Player bonus 10 enters percent rather than flat AR. Cases c/d additionally reject requesting/applying a nonzero melee mastery term on missiles: missile AR 435 differs from melee AR 536.
  - runtime_observed: no

- **SBA-V17** (SBA-R02e/SBA-R03; boundary: isolated rule)
  - Shared inputs: monster attacker, Stat 19=200, Stat 2=10, bonus 15; stop before negative swap.
  - (a)
    - Inputs: Stat 119=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: AR=200+15+50=265; percent 0 adds 0.
    - Outputs: AR 265, pct 0.
  - (b)
    - Inputs: Stat 119=20.
    - Seeds: no draw.
    - Draws: none.
    - Trace: AR=265; ordered percent amount=(20*265)/100=53; add 53.
    - Outputs: AR 318, pct 20.
  - Discriminates: monster bonus 15 is flat, while a nonzero Stat 119 still scales the resulting AR.
  - runtime_observed: no

- **SBA-V18** (SBA-R02c/SBA-R07/SBA-R15a/SBA-R05; boundary: supplied-input stage)
  - Shared inputs: player attacker, Stat 115=1, Stat 117=0, DEF 500 before adjustment, AR 300 after percent, aLvl=dLvl 20
  - (a)
    - Inputs: ordinary monster, exclusion=false, boss=false, hireling=false
    - Seeds: attacker A independently in case.
    - Draws: attacker/100/51.
    - Trace: DEF 0, ratio 100, chance 95
    - Outputs: DEF 0, chance 95, return 1, A1
  - (b)
    - Inputs: boss monster
    - Seeds: attacker A independently in case.
    - Draws: attacker/100/51.
    - Trace: DEF 500, ratio 37, chance 37.
    - Outputs: DEF 500, chance 37, return 0, A1.
  - (c)
    - Inputs: hireling monster
    - Seeds: attacker A independently in case.
    - Draws: attacker/100/51.
    - Trace: DEF 500, ratio 37, chance 37.
    - Outputs: DEF 500, chance 37, return 0, A1.
  - (d)
    - Inputs: exclusion=true monster
    - Seeds: attacker A independently in case.
    - Draws: attacker/100/51.
    - Trace: DEF 500, ratio 37, chance 37.
    - Outputs: DEF 500, chance 37, return 0, A1.
  - (e)
    - Inputs: player.
    - Seeds: attacker A independently in case.
    - Draws: attacker/100/51.
    - Trace: DEF 500, ratio 37, chance 37.
    - Outputs: DEF 500, chance 37, return 0, A1.
  - Discriminates: protected defender retains 500 and misses, ordinary defense 0 still draws.
  - runtime_observed: no

- **SBA-V19** (SBA-R15a/SBA-R10/SBA-R11/SBA-R12a/SBA-R16; boundary: supplied-input stage)
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

- **SBA-V20** (SBA-R09b; boundary: isolated rule)
  - Inputs: monster, NoShldBlock=true, Stat 20=30.
  - Seeds: no draw.
  - Draws: none.
  - Trace: flag enables 30 without special classifier.
  - Outputs: block 30, no special-classifier request.
  - Discriminates: flag alone enables raw 30.
  - runtime_observed: no

- **SBA-V21** (SBA-R09b; boundary: supplied-input stage)
  - Inputs: monster, NoShldBlock=false, supplied special result 0, Stat 20=30, avoidance 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: both eligibility routes fail.
  - Outputs: block 0, code 0, no shield draw.
  - Discriminates: stat 30 alone does not enable block.
  - runtime_observed: no

- **SBA-V22** (SBA-R12b/SBA-R12c/SBA-R13/SBA-R16; boundary: supplied-input stage)
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

- **SBA-V23** (SBA-R13; boundary: supplied-input stage)
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

- **SBA-V24** (SBA-R18a/SBA-R08; boundary: supplied-input stage)
  - Shared inputs: player attacker, Stat 117=1, monster defender, AR 1000, DEF 500, aLvl=dLvl 30
  - Shared trace: chance 66, hit
  - Shared outputs: each return 1, A1
  - (a)
    - Inputs: class 5
    - Seeds: attacker A.
    - Draws: attacker/100/51.
    - Trace: class outside 704..709
    - Outputs: one logged state-applier request(attacker, defender, state 52, level 1, duration 120000, auxStat 31, value 0, effectIdentity 0, default callback)
  - (b)
    - Inputs: class 704
    - Seeds: attacker A.
    - Draws: attacker/100/51.
    - Trace: excluded endpoints.
    - Outputs: no request; no stored-state assertion.
  - (c)
    - Inputs: class 709.
    - Seeds: attacker A.
    - Draws: attacker/100/51.
    - Trace: excluded endpoints.
    - Outputs: no request; no stored-state assertion.
  - (d)
    - Inputs: class 703.
    - Seeds: attacker A.
    - Draws: attacker/100/51.
    - Trace: 703 is immediately below the inclusive excluded range.
    - Outputs: return 1, attacker A1; one logged state-applier request(attacker, defender, state 52, level 1, duration 120000 frames, auxStat 31, value 0, effectIdentity 0, default callback); no stored-state assertion.
  - (e)
    - Inputs: class 710.
    - Seeds: attacker A.
    - Draws: attacker/100/51.
    - Trace: 710 is immediately above the inclusive excluded range.
    - Outputs: return 1, attacker A1; one logged state-applier request(attacker, defender, state 52, level 1, duration 120000 frames, auxStat 31, value 0, effectIdentity 0, default callback); no stored-state assertion.
  - Discriminates: Excluded 704/709 request no applier while adjacent 703/710 do; the hook adds no local draw.
  - runtime_observed: no

- **SBA-V25** (SBA-R11/SBA-R19; boundary: isolated rule)
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

- **SBA-V26** (SBA-R20/SBA-R14/SBA-R16; boundary: supplied-input stage)
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

- **SBA-V27** (SBA-R02d; boundary: isolated rule)
  - Shared inputs: player attacker, DEF 500
  - (a)
    - Inputs: Stat 116=50, ordinary monster.
    - Seeds: no draw.
    - Draws: none.
    - Trace: p=50, cut 250
    - Outputs: DEF 250
  - (b)
    - Inputs: Stat 116=50, player or hireling.
    - Seeds: no draw.
    - Draws: none.
    - Trace: half 25, cut 125
    - Outputs: 375
  - (c)
    - Inputs: Stat 116=250, ordinary monster.
    - Seeds: no draw.
    - Draws: none.
    - Trace: cap 100, cut 500
    - Outputs: 0
  - (d)
    - Inputs: Stat 116=50, ordinary monster, DEF 2000000.
    - Seeds: no draw.
    - Draws: none.
    - Trace: divide first, cut 1000000
    - Outputs: 1000000
  - (e)
    - Inputs: Stat 116=1, player.
    - Seeds: no draw.
    - Draws: none.
    - Trace: half 0, cut 0.
    - Outputs: 500.
  - (f)
    - Inputs: Stat 116=0, ordinary monster.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Nonpositive stat bypasses the reduction arithmetic.
    - Outputs: DEF 500.
  - (g)
    - Inputs: Stat 116=-50, ordinary monster.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Negative stat also bypasses reduction; it does not increase defense.
    - Outputs: DEF 500.
  - (h)
    - Inputs: Stat 116=50, ordinary monster, DEF -500.
    - Seeds: no draw.
    - Draws: none.
    - Trace: p=50; cut=(-500*50)/100=-250; DEF=-500-(-250).
    - Outputs: DEF -250.
  - Discriminates: halved 1 becomes 0; player defense 375 differs ordinary 250.
  - runtime_observed: no

- **SBA-V28** (SBA-R02g; boundary: isolated rule)
  - Shared inputs: player attacker, AR 300, Stat 123=50, Stat 124=40
  - (a)
    - Inputs: demon monster
    - Seeds: no draw.
    - Draws: none.
    - Trace: add 50 then 40 only when predicate matches.
    - Outputs: AR 350
  - (b)
    - Inputs: lUndead monster
    - Seeds: no draw.
    - Draws: none.
    - Trace: add 50 then 40 only when predicate matches.
    - Outputs: 340
  - (c)
    - Inputs: demon+hUndead monster
    - Seeds: no draw.
    - Draws: none.
    - Trace: add 50 then 40 only when predicate matches.
    - Outputs: 390
  - (d)
    - Inputs: neither monster
    - Seeds: no draw.
    - Draws: none.
    - Trace: add 50 then 40 only when predicate matches.
    - Outputs: 300
  - (e)
    - Inputs: player.
    - Seeds: no draw.
    - Draws: none.
    - Trace: add 50 then 40 only when predicate matches.
    - Outputs: 300.
  - (f)
    - Inputs: Monster defender with demon and hUndead true; Stat 123=-50 and Stat 124=-40, overriding their shared positive values.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Both nonzero values are added unchanged:300-50-40.
    - Outputs: AR 210.
  - Discriminates: both flags add to 390 rather than exclusive branch.
  - runtime_observed: no

- **SBA-V29** (SBA-R03; boundary: isolated rule)
  - (a)
    - Inputs: AR 2000050, pct 33
    - Seeds: no draw.
    - Draws: none.
    - Trace: 20000*33=660000
    - Outputs: AR 2660050
  - (b)
    - Inputs: AR 150, pct 70050.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 700*150=105000.
    - Outputs: AR 105150.
  - (c)
    - Inputs: AR 1048576, pct 33.
    - Seeds: no draw.
    - Draws: none.
    - Trace: The strict AR>1048576 test is false. Direct i32 product 34603008 divided by 100 truncates to 346030; add to AR.
    - Outputs: percent amount 346030, AR 1394606.
  - (d)
    - Inputs: AR 150, pct 65536.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Both strict threshold tests are false. Direct i32 product 9830400 divided by 100 is 98304; add to AR.
    - Outputs: percent amount 98304, AR 98454.
  - Discriminates: Cases a/b reject multiply-first arithmetic above the thresholds. Exact-boundary c/d reject >= branches: premature division would give amounts 346005 and 98250, rather than 346030 and 98304.
  - runtime_observed: no

- **SBA-V30** (SBA-R14/SBA-R12a/SBA-R12b/SBA-R12c; boundary: supplied-input stage)
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

- **SBA-V31** (SBA-R09d/SBA-R09a; boundary: supplied-input stage)
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

- **SBA-V32** (SBA-R12b; boundary: isolated rule)
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

- **SBA-V33** Holy Shield request (SBA-R01b/SBA-R01a; boundary: supplied-input stage)
  - Inputs: player, Stat 31=120, Dex 0, Stat 16=50, Stat 171=0, Stat 182=50, state 101 active, aggregate base present, matching list own 350=10, own 351=2, valid skill 10 with `Skills.calc1` reference 7, shield=true, supplied expression result 25.
  - Seeds: no draw in supplied expression.
  - Draws: none.
  - Trace: request Evaluate(defender,7,10,2); pct 75; base 120+90=210; final add 105.
  - Outputs: defense 315, one ordered evaluation request, zero local draws.
  - Discriminates: treating Holy Shield 25 as a flat addition after the first percent would produce 307, rather than 315.
  - runtime_observed: no

- **SBA-V34** Type sum/bounds (SBA-R02b; boundary: supplied-input stage)
  - Shared inputs: player attacker, isMissile 1, Stat 119=20, bonus 10, supplied typeIndex 2, count 4, entries 179[(0,99),(1,7),(3,-2),(4,50)].
  - (a)
    - Inputs: Monster defender, relation(2,1)=true, relation(2,3)=true.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Ignore layers 0/4; sum 7-2=5; pct=20+10+5.
    - Outputs: typeBonus 5, pct 35, no relation read for layers 0/4.
  - (b)
    - Inputs: Monster defender, relation(2,1)=false, relation(2,3)=true.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Ignore layers 0/4; false relation skips 7; add -2.
    - Outputs: typeBonus -2, pct 28, no relation read for layers 0/4.
  - (c)
    - Inputs: Player defender, same supplied relation bits as case a.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Non-monster defender skips the stat-179 relation loop.
    - Outputs: typeBonus 0, pct 30, no relation read.
  - (d)
    - Inputs: Player attacker, isMissile 1, Stat 119=20, bonus 10, monster defender, typeIndex 2, count 130. Replace the shared entries with a supplied stat-179 enumerator fixture having 129 available sorted entries: layers 1 through 129, each value 1; it logs the requested maximum and returns the first min(maximum,129) entries. Every supplied relation(2,L) for L in 1..129 is true.
    - Seeds: no draw.
    - Draws: none.
    - Trace: The consumer requests maximum 128; the supplied enumerator returns layers 1..128. Each passes bounds and relation; sum=128, pct=20+10+128.
    - Outputs: typeBonus 128, pct 158; enumeration maximum log[128], no read of layer 129.
  - Discriminates: summation differs from maximum; false relation and non-monster defender suppress the corresponding additions; invalid bounds never access the supplied matrix.
  - runtime_observed: no

- **SBA-V35** No consumer cap (SBA-R12d/SBA-R12a; boundary: isolated rule)
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

- **SBA-V36** Missile skip/avoidance-only (SBA-R15a/SBA-R14/SBA-R12b; boundary: supplied-input stage)
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

- **SBA-V37** Large final defense modifiers (SBA-R01a; boundary: supplied-input stage)
  - Shared inputs: cases a/b start after the first defense percentage stage; cases c/d use the isolated TotalDefense operation stated in their inputs.
  - (a)
    - Inputs: base 2000050, Stat 182=33.
    - Seeds: no draw.
    - Draws: none.
    - Trace: base>1048576 uses (2000050/100)*33=660000.
    - Outputs: TotalDefense 2660050.
  - (b)
    - Inputs: base 150, Stat 182=70050.
    - Seeds: no draw.
    - Draws: none.
    - Trace: base threshold false, modifier>65536 uses (70050/100)*150=105000.
    - Outputs: TotalDefense 105150.
  - (c)
    - Inputs: Isolated TotalDefense operation with a null unit.
    - Seeds: no draw.
    - Draws: none.
    - Trace: The non-null-unit requirement fails on the unconditional unit access; no arithmetic result is returned.
    - Outputs: memory fault, no defense result, no draw.
  - (d)
    - Inputs: Isolated TotalDefense operation, non-null player, Stat 31=0, Stat 2=0, Stat 171=50, Stat 16=0, Stat 182=0, state 101 inactive.
    - Seeds: no draw.
    - Draws: none.
    - Trace: base=0, pct=50, prod=0; the nonpositive-base path subtracts 0; final modifier is zero.
    - Outputs: TotalDefense 0.
  - Discriminates: The final modifier retains its divide-first branches rather than using mathematically exact multiply-first amounts 660016/105075.
  - runtime_observed: no

- **SBA-V38** Holy Shield gates (SBA-R01b; boundary: supplied-input stage)
  - Shared inputs: Isolated Holy Shield contribution: player, active state 101, present aggregate-eligible base list, matching state list own 350=10/351=2, valid skill 10 and `Skills.calc1` reference 7, shield=true; supplied evaluation would return 25 if requested.
  - (a)
    - Inputs: Unit type 2.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Unsupported type stops before state/list work.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - (b)
    - Inputs: State 101 inactive.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Inactive state stops before list/skill work.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - (c)
    - Inputs: Base list aggregate eligibility false.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Ineligible list storage stops before matching-list reads.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - (d)
    - Inputs: Matching state list absent.
    - Seeds: no draw.
    - Draws: none.
    - Trace: No list means no identity/level values.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - (e)
    - Inputs: Matching list own stat 350=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Nonpositive identity skips lookup/evaluation.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - (f)
    - Inputs: Matching list own stat 351=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Nonpositive level skips lookup/evaluation.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - (g)
    - Inputs: Skill definition lookup for id 10 returns null.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Failed definition lookup skips shield/evaluation.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - (h)
    - Inputs: Supplied shield result false.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Shield failure skips evaluation.
    - Outputs: holyTerm 0; no expression-evaluation request.
  - Discriminates: Ignoring any gate would request the supplied evaluation and add 25 rather than returning contribution 0.
  - runtime_observed: no

- **SBA-V39** Player base failure branches (SBA-R02a; boundary: supplied-input stage)
  - Shared inputs: Isolated player-base helper, Stat 19=100, Stat 2=50.
  - (a)
    - Inputs: Player with invalid class -1.
    - Seeds: no draw.
    - Draws: none.
    - Trace: 100+5*50-35; omit class factor.
    - Outputs: AR 315, no fault.
  - (b)
    - Inputs: Null unit.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Reject before class/stat calculation.
    - Outputs: abort error 4223, no return value.
  - (c)
    - Inputs: Monster unit.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Reject non-player helper input.
    - Outputs: abort error 4223, no return value.
  - Discriminates: Invalid class omits a factor; invalid unit kind is fatal rather than returning the same base.
  - runtime_observed: no

- **SBA-V40** Zero level denominator (SBA-R06; boundary: supplied-input stage)
  - Shared inputs: Isolated level-adjustment stage.
  - (a)
    - Inputs: ratio 66, aLvl=0, dLvl=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Numerator 0; denominator 0; unguarded signed division faults.
    - Outputs: divide-by-zero fault, no chance result and no draw.
  - Discriminates: An inserted level floor or zero-denominator default would return a chance instead of faulting.
  - runtime_observed: no

- **SBA-V41** Block eligibility branches (SBA-R09a/SBA-R09b; boundary: supplied-input stage)
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

- **SBA-V42** Effective/raw modes and disabled shield (SBA-R10/SBA-R11; boundary: supplied-input stage)
  - Shared inputs: Player, block 30 before movement penalty, no avoidance.
  - (a)
    - Inputs: R10-only stage, raw mode 2, effective mode 3 with override present.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Effective moving true, but raw Walk excludes penalty.
    - Outputs: effective block 30.
  - (b)
    - Inputs: R10-only stage, raw mode 1, effective mode 3 with override present.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Effective moving true and raw mode not Walk:30/3.
    - Outputs: effective block 10.
  - (c)
    - Inputs: R10-only stage, raw mode 3, effective mode 1 with override present.
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

- **SBA-V43** Other moving units and nonpositive Evade (SBA-R12a; boundary: supplied-input stage)
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

- **SBA-V44** Weapon candidate fallback/short circuit (SBA-R12b; boundary: supplied-input stage)
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
- **SBA-O07** Observed 116 calls, chance 47, hit roll<=46, miss>=47, no defender step on miss.
- **SBA-O08** Observed 57 scaled-block calls return 62.
- **SBA-O09** Observed 34 blocks/27 failures, strict threshold 62, one defender step on block.
- **SBA-O10** Observed 4 total defender steps after failed block, subsequent attribution pending.
- **SBA-O11** Physical-zero missile skips shield; physical control reaches it.

## 9. Acceptance and dependencies

The verified supplied-input subset consists of SBA-V01 through SBA-V44 at their stated boundaries, including exact intermediates, flags, requests and seed changes. Passing it does not establish unresolved mastery, special-monster, shield, classification, caller, precondition, owner or state-applier integration. Full reproduction requires those gates closed and pending gameplay observations. Observation gaps alone do not gate verified supplied-input arithmetic.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md) NRF-R01/NRF-R02 steps/modulus, NRF-R05 ordered mulDiv, NRF-R12a/NRF-R12b/NRF-R12c list/state/stat reads; [data fields](core-game-data-fields.md) GDF-R03/GDF-R05/GDF-R07/GDF-R11/GDF-R12/GDF-R14a/GDF-R15a/GDF-R16/GDF-R17/GDF-R18 names/schema/modes; [expressions](skill-calc-expressions-and-synergies.md) SCX-R11 and requested producers with their gates/draws; [timed states](state-timed-state-lifecycle.md) STL-R06a/STL-R06b/STL-R06c/STL-R11 for actual prevent-heal effects. Section 2 supplies upstream trigger/bonus/contact/precondition results; SBA-Q22 retains incomplete integration.

**Depended on by:** [damage pipeline](combat-damage-pipeline-order.md) result flags; [hit recovery](combat-hit-recovery.md) reactions; [knockback](combat-knockback.md) knockback flag; [durability](item-durability-loss.md) standing hit result and seed position.

## 10. Related specs

[Skill dispatch](skill-cast-start-and-do-dispatch.md) and [leveled damage/to-hit](skill-leveled-damage-tohit-and-elements.md) locate melee roots; [missile movement](missile-movement-and-collision.md) and [missile damage](missile-hit-and-damage-calculation.md) locate contact/fill; [PvP](combat-pvp-damage-and-kill-credit.md) owns hostility; [movement/freeze](state-movement-modifiers-and-freeze.md) shares mode identities.

## Appendix A. Provenance

Original LoD 1.13c Windows x86 D2Game, D2Common, Fog evidence uses the [source manifest](../research/source-manifest.md). The [companion](../research/spec-evidence/combat-shield-block-and-avoidance.md) preserves dated 2026-10-11..13 static/runtime evidence and correction rationale. Targeted 2026-10-15 reading used matched C where available, otherwise disassembly, including decoding checked against manifest-matching original bytes. Narrow ItemStatCost names were read through the effective loader 2026-10-15. Observed gameplay scope covers ratio/level, scaled player block and shield draw cases only; no new gameplay campaign ran.
