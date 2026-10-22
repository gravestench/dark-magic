# Attack hit roll

**Description**

Before a defender can block or dodge, a melee attack or a missile that checks accuracy first rolls to see whether it hits at all. The attacker's attack rating, the defender's defense and both characters' levels set a hit chance that never falls below 5 percent or rises above 95 percent. Some player items raise attack rating against demons or undead, or ignore or reduce a monster's defense. A player hit with a prevent-monster-heal item also asks for that effect to be placed on most monsters.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: HIT-R, HIT-Q, HIT-V, HIT-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-hit-roll.md)
Open evidence gaps: [gap file](../plans/gaps/combat-hit-roll.md)

## 1. Scope

This contract covers the attack hit/miss roll: contact defense with its total-defense and Holy Shield terms, player and other attack rating, attack-rating percentages, ignore and reduce defense, demon/undead and monster-type bonuses, the negative swap, ratio, level term, 5/95 cap, the attacker draw, and the post-hit prevent-healing request. Whether the roll runs, the arguments its callers pass, shield block, avoidance, result codes/flags and the attack's overall draw order belong to [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md). Damage and later effects belong to [combat-damage-pipeline-order](combat-damage-pipeline-order.md). Mastery, monster-type relation, neutral monster condition, shield and state-applier producers have explicit supplied-input integration boundaries below.

## 2. Trigger and ownership

D2Game (server) computes the hit roll and the post-hit request. D2Common computes total defense and the player attack base. D2Game UNIT/SUnitDmg is a source-folder lead rather than an assert-string identification.

The roll is reached from two callers owned by [shield block and avoidance](combat-shield-block-and-avoidance.md). The melee result (SBA-R05) runs it after its preconditions (SBA-R12) when the hit flag is clear, with isMissile 0 and the caller's bonus. Missile unit contact (SBA-R06) runs it when Missiles.ToHit is nonzero, with isMissile 1 and bonus equal to the missile's Stat 19. SBA-R07a states when each caller skips the roll. Representative melee chain: server skill event -> do dispatcher -> representative handler 46 -> skill hit-damage calculation -> melee result -> hit roll. Other callers remain SBA-Q08. Original locations and argument proofs are in the companion.

The roll runs once per attack contact, before shield block (SBA-R08 owns the attack's complete draw order). On a successful hit the hook runs before the roll returns. No client message is sent; subsequent damage/reaction processing conveys the outcome. No periodic cadence belongs here; the hook's state applier can schedule expiry under its dependency contract.

## 3. Inputs and data

Arithmetic is `i32`, wrapping addition, subtraction, multiplication and shift; division truncates toward zero, except unsigned seed modulus. Stat(u, id) is the selected layer-zero read under RLI-R02a/RLI-R02b. Layer enumeration uses the aggregate array under RLI-R02b, sorted by `u16` layer. Own-list reads are named explicitly. A supplied operation logs requests and has only the effects declared by a vector. Supplied results never establish their producer.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| unit identity, class | Logical unit/game fields; NRF | `i32` and nullable associations | Types 0 player,1 monster,3 missile |
| levels, Dexterity | 12 level,2 dexterity | `i32` | Attack and defense inputs |
| attack/defense | 19 tohit,31 armorclass,32 armorclass_vs_missile,33 armorclass_vs_hth | `i32` | Base and contact terms |
| defense percentages | 171 skill_armor_percent,16 item_armor_percent,182 armor_override_percent | `i32` percent | Total-defense steps |
| attack percentages | 119 item_tohit_percent,115 item_ignoretargetac,116 item_fractionaltargetac | `i32` | Percent and defense adjustments |
| target bonuses | 123 item_demon_tohit,124 item_undead_tohit | `i32` | Flat AR additions |
| monster-type bonus | 179 attack_vs_montype | Up to 128 (`u16` layer, `i32` value) entries | Sum matching values; supplied signed `i16` monsterTypeIndex/count/relation matrix |
| mastery | 342 passive_mastery_melee_th; supplied selected weapon/meleeExtra | `i32` percent | Producer HIT-Q01 |
| state/list inputs | 101 holyshield; matching list's own 350 modifierlist_skill and 351 modifierlist_level | Bit, nullable list, `i32` | Holy Shield evaluation |
| shield result | Supplied shield result | Boolean | Holy Shield requirement; producer SBA-R01e, gated by SBA-Q10 |
| class factor | CharStats.ToHitFactor | `i32` | Supplied by class |
| monster flags | MonStats.boss, demon, lUndead, hUndead; hireling classification | Bits/booleans | Definition and class predicates |
| neutral monster conditions | monsterDefenseExclusion, monsterDefenseHalf | Booleans; half is first component of exclusion | Unknown producer semantics HIT-Q03; absent monster data means false |
| attack arguments | bonus, isMissile | `i32` | Player bonus percent; other bonus flat; isMissile selects the contact defense stat and skips mastery |
| seeds | Attacker | Pair of `u32` | NRF-R01/NRF-R02 |
| prevent-healing request | stat 117 item_preventheal; target class; supplied state applier | `i32`, logical request | State 52, level 1, duration 120000 frames, effectIdentity 0, aux stat 31/value 0; effectIdentity meaning is HIT-Q04 |

| Mapping | Origin | Binding |
|---|---|---|
| Stat/state selectors, caps 5/95 | Engine code | Fixed consumer identities, not inferred data remapping |
| Class factor | Table data | Supplied row/column values |
| Names 117/123/124/342/350/351 | Table data | Explicit ItemStatCost.ID/Stat named cells; companion Table data |
| Monster-type relation/mastery/shield producers | Unresolved | Supplied results, with precise gates below |

Synthetic vectors supply values rather than assert actual table cells. Observed single named cells and their read limits remain in the companion.

## 4. Rules (execution order)

Execution: contact defense and attack-rating terms -> percent -> negative swap -> ratio -> level term -> cap -> attacker draw -> hook on success. No local rule draws unless it says so. Expression/mastery/applier producers have separately owned draw contracts. For a player hit-roll path, evaluate HIT-R02a, then HIT-R02c/HIT-R02d/HIT-R02g, then acquire the optional HIT-R02f result and form HIT-R02b, before HIT-R03 and HIT-R04; the numbered sub-rules group the input terms rather than changing this call order.

**HIT-R01** Contact defense [verified in 1.13c code (static)]. DEF=TotalDefense(defender)+Stat(defender, isMissile ?32:33). No draw.

**HIT-R01a** Total defense [verified in 1.13c code (static)]. For a non-null unit, base=Stat(31)+Stat(2)/4; pct=Stat(171)+Stat(16)+holyTerm; prod=pct*base. If base>0 add prod/100; otherwise subtract prod/100. Read f=Stat(182); when nonzero add (base/100)*f if base>1048576, else (f/100)*base if f>65536, else (base*f)/100. Missing list reads zero. Null unit faults on a later unconditional unit access. No local draw.

**HIT-R01b** Holy Shield [verified in 1.13c code (static)]. holyTerm starts 0. Proceed only for types 0/1/3, active state 101, present aggregate-eligible base list and a matching state list. Read its own 350 and 351; both must be positive. A valid skill definition and supplied shield are required. Evaluate `Skills.calc1` in context(defender, skill, level) under SCX-R11, add the supplied signed result to pct before the product. Shield execution is SBA-Q10; requested expression producers retain their SCX gates/draws. No direct draw.

**HIT-R02a** Player base [verified in 1.13c code (static)]. AR=Stat(19)+5*Stat(2)-35+CharStats.ToHitFactor. Invalid class omits class term. Null/non-player helper input aborts error 4223; dispatcher calls only for players. No draw.

**HIT-R02b** Player percent [verified in 1.13c code (static)]. pct=meleeExtra+Stat(119)+bonus; meleeExtra 0 for isMissile!=0, otherwise HIT-R02f. A monster defender with monsterTypeIndex>0 enables up to 128 stat 179 entries. Add each value whose layer and index are both >0 and <count and whose supplied relation bit is set. Other defenders add none. No local draw. Producer integration HIT-Q01/HIT-Q02.

**HIT-R02c** Ignore defense [verified in 1.13c code (static)]. Player adjustment only: Stat 115!=0, monster defender, monsterDefenseExclusion=false, boss=false, hireling=false sets DEF 0. Otherwise retain DEF. Zeroed defense still goes through the cap and the attacker draw. Missing monster-specific data makes neutral conditions false. No draw.

**HIT-R02d** Reduce defense [verified in 1.13c code (static)]. After HIT-R02c, v=Stat 116; v<=0 changes nothing. Halve positive v for a player or a monster with boss, monsterDefenseHalf or hireling true. p=min(v,100). cut=(DEF/100)*p when DEF>1048576, else (DEF*p)/100; subtract cut. Negative DEF is not floored here. No draw.

**HIT-R02e** Other attackers [verified in 1.13c code (static)]. Non-player AR=Stat 19+bonus+5*Stat 2, pct=Stat 119. Skip player adjustments, mastery and type bonuses. No draw.

**HIT-R02f** Mastery producer [unresolved]. A selected weapon can produce meleeExtra. An absent selected weapon contributes zero. Exact selection and special-mastery results remain HIT-Q01; supplied meleeExtra supports caller arithmetic independently.

**HIT-R02g** Demon/undead [verified in 1.13c code (static)]. Player adjustment after HIT-R02d: add nonzero Stat 123 for demon, then nonzero Stat 124 for lUndead or hUndead. Both may add; negatives add unchanged. Non-monster/invalid definition predicates false. No draw.

**HIT-R03** Apply percent [verified in 1.13c code (static)]. Add mulDiv(pct, AR,100), NRF-R05 argument order: (AR/100)*pct if AR>1048576; else (pct/100)*AR if pct>65536; else (pct*AR)/100. No draw.

**HIT-R04** Negative swap [verified in 1.13c code (static)]. In order: DEF<0 adds -DEF to AR and sets DEF 0; AR<0 adds -AR to DEF and sets AR 0; DEF<0 then sets DEF 0. No draw.

**HIT-R05** Ratio [runtime-confirmed (1.13c)]. AR+DEF==0 yields 100, otherwise ratio=(AR*100)/(AR+DEF). No draw.

**HIT-R06** Level term [runtime-confirmed (1.13c)]. chance=((ratio*aLvl)<<1)/(aLvl+dLvl). Zero denominator faults. No draw.

**HIT-R07** Hit cap [verified in 1.13c code (static)]. chance<=5 becomes 5; else chance>=95 becomes 95. No draw.

**HIT-R08** Hit draw [verified in 1.13c code (static)]. Step attacker once, unsigned lo mod 100, return 1 if roll<chance, otherwise 0. A hit calls HIT-R09a before return. No defender draw.

**HIT-R09a** Hit hook [verified in 1.13c code (static)]. Successful hit: only player attacker, monster defender, Stat 117!=0, class outside inclusive 704..709 calls state applier(attacker, defender, state 52, level 1, duration 120000 frames, auxStat 31, value 0, effectIdentity 0, default removal callback). No local draw.

**HIT-R09b** Applier [asm-only]. Request can create/refresh timed prevent-healing list; complete state effects depend on STL-R06a/STL-R06b/STL-R06c/STL-R11 with their gates. Null target aborts 2602. Complete integration HIT-Q04; supplied applier verifies only request/order.

**HIT-R10** Attacker type differences [verified in 1.13c code (static)]. Player attackers alone use HIT-R02a/b/c/d/f/g; others HIT-R02e. HIT-R02c/d state exact player/boss/hireling/neutral-condition differences. No additional PvP arithmetic occurs locally.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| HIT-R01 | Contact defense | verified in 1.13c code (static) | none with supplied inputs | HIT-V05 |
| HIT-R01a | Total defense | verified in 1.13c code (static) | SBA-Q10 shield producer only | HIT-V05/HIT-V06/HIT-V07/HIT-V15/HIT-V17 |
| HIT-R01b | Holy Shield | verified in 1.13c code (static) | SBA-Q10 shield; SCX requested producer gates | HIT-V15/HIT-V18 |
| HIT-R02a | Player base | verified in 1.13c code (static) | none with supplied inputs | HIT-V08/HIT-V19 |
| HIT-R02b | Player percent | verified in 1.13c code (static) | HIT-Q01/HIT-Q02 producer integration only | HIT-V08/HIT-V16 |
| HIT-R02c | Ignore defense | verified in 1.13c code (static) | HIT-Q03 producer only | HIT-V10 |
| HIT-R02d | Reduce defense | verified in 1.13c code (static) | HIT-Q03 producer only | HIT-V12 |
| HIT-R02e | Other attackers | verified in 1.13c code (static) | none with supplied inputs | HIT-V09 |
| HIT-R02f | Mastery producer | unresolved | HIT-Q01 mastery | none |
| HIT-R02g | Demon/undead | verified in 1.13c code (static) | none with supplied inputs | HIT-V13 |
| HIT-R03 | Apply percent | verified in 1.13c code (static) | none with supplied inputs | HIT-V08/HIT-V09/HIT-V14 |
| HIT-R04 | Negative swap | verified in 1.13c code (static) | none with supplied inputs | HIT-V04 |
| HIT-R05 | Ratio | runtime-confirmed (1.13c) | none with supplied inputs | HIT-V01/HIT-V02/HIT-V03/HIT-V04/HIT-V08/HIT-V10 |
| HIT-R06 | Level term | runtime-confirmed (1.13c) | none with supplied inputs | HIT-V01/HIT-V02/HIT-V03/HIT-V08/HIT-V20 |
| HIT-R07 | Hit cap | verified in 1.13c code (static) | none with supplied inputs | HIT-V02/HIT-V03/HIT-V10 |
| HIT-R08 | Hit draw | verified in 1.13c code (static) | none with supplied inputs | HIT-V01/HIT-V08/HIT-V10/HIT-V11 |
| HIT-R09a | Hit hook | verified in 1.13c code (static) | HIT-Q04 actual applier only | HIT-V11 |
| HIT-R09b | Applier | asm-only | HIT-Q04 effects | none |
| HIT-R10 | Attacker type differences | verified in 1.13c code (static) | none with supplied inputs | HIT-V08/HIT-V09 |

## 5. Outputs and side effects

| Output/effect | Rule |
|---|---|
| Hit return 0/1 and attacker seed advance | HIT-R08 |
| Ordered applier request; actual timed-state effects gated | HIT-R09a/HIT-R09b |
| Fatal 4223 for invalid player-base input | HIT-R02a |
| Zero level denominator fault; null defense unit fault | HIT-R06/HIT-R01a |

AR, DEF, pct, chance and helper returns are intermediates, not stored writes. No local message, unit creation or unit destruction occurs.

## 6. Branches and edge cases

Player attackers alone use the player base, defense adjustments, mastery and monster-type terms; other attackers add their bonus flat. Raw Run player melee defenders and missiles with ToHit 0 skip this roll entirely (SBA-R07a). No difficulty term exists and no additional local PvP arithmetic occurs. Missing lists give zero/empty stats but retain class terms. Invalid monster definitions test false. Negative defense after reduction is not floored before the negative swap. A zero level sum faults. Signed wrapped products and truncation remain applicable. Unresolved producer values are supplied explicitly, never inferred defaults.

## 7. Open questions

The [gap file](../plans/gaps/combat-hit-roll.md) records attempted readings and exact boundaries. The shield producer that Holy Shield requires is SBA-Q10, and other callers of this roll are SBA-Q08.

| Question | Status | Evidence that would settle it |
|---|---|---|
| **HIT-Q01** Melee mastery | behavior | Full selected-weapon and special mastery paths |
| **HIT-Q02** Monster relation producer | integration | Named loader field and matrix producer |
| **HIT-Q03** Neutral monster conditions | outside-boundary | Spawn flag writers; supplied conditions usable |
| **HIT-Q04** State application | integration | Close CSA/STL attach/write/removal for this request |
| **HIT-Q05** Gameplay coverage | observation | Pending original-game observations beyond the recorded ratio/level cases |

## 8. Test vectors

### 8.1 Executable vectors

Boundary for all vectors: supplied-input stage unless the vector stops at an isolated arithmetic result. Inputs may name an intermediate AR, DEF or successful hit to start at that stage; earlier producers are not executed. Defaults: non-null monster attacker/player defender, valid supplied records, raw/effective mode 1, no states, all unlisted stats zero, entries empty and shield absent. An ordinary monster means type 1 with boss, hireling, demon, lUndead, hUndead and both neutral defense conditions false and valid class 5; unlisted monster flags use those defaults. Supplied operations log requests in order, draw nothing, and do not persist producer effects; the state applier logs its request only. No-weapon/supplied shield outcomes verify consumers only. Final seed is the successor after the listed draw count; every other seed remains unchanged. All numeric seeds/results below are exact.

Seed aliases in the vectors are exact pairs and successors: A={1,0x29A}, A1={0x6AC6935F,0}, A2={0x2F2ED81B,0x2C890AFD}; C={5,0x29A}, C1={0x15E0D673,2}. Draw notation is owner/modulus/result, ordered left to right. Numeric outputs use decimal unless prefixed 0x. Case(a)/(b) labels below give independent input overrides, seeds, draws and outputs; no case inherits another case's outputs.

- **HIT-V01** (HIT-R05/HIT-R06/HIT-R08; boundary: supplied-input stage)
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

- **HIT-V02** (HIT-R05/HIT-R06/HIT-R07; boundary: supplied-input stage)
  - Inputs: AR=100, DEF=5000, aLvl=dLvl=30.
  - Seeds: attacker A.
  - Draws: attacker/100/51.
  - Trace: ratio=10000/5100=1; chance 1 capped 5.
  - Outputs: chance 5, return 0, attacker A1.
  - Discriminates: chance 5 rejects omitted lower cap.
  - runtime_observed: no

- **HIT-V03** (HIT-R05/HIT-R06/HIT-R07; boundary: supplied-input stage)
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

- **HIT-V04** (HIT-R04/HIT-R05; boundary: isolated rule)
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

- **HIT-V05** (HIT-R01a/HIT-R01; boundary: isolated rule)
  - Inputs: defender Stat 31=100, Stat 2=50, Stat 171=0, Stat 16=50, Stat 182=0, state 101=false, Stat 33=0, Stat 32=20.
  - Seeds: no draw.
  - Draws: none.
  - Trace: base 100+50/4=112; pct 50; prod 5600; base 168.
  - Outputs: TotalDefense 168, melee DEF 168, missile DEF 188.
  - Discriminates: missile 188 rejects using melee defense 168.
  - runtime_observed: no

- **HIT-V06** (HIT-R01a; boundary: isolated rule)
  - Inputs: Stat 31=-20, Stat 2=0, Stat 171=50, Stat 16=0, Stat 182=0, state 101=false.
  - Seeds: no draw.
  - Draws: none.
  - Trace: prod -1000; negative base subtracts -10.
  - Outputs: TotalDefense -10.
  - Discriminates: negative base -10 rejects adding percent to obtain -30.
  - runtime_observed: no

- **HIT-V07** (HIT-R01a; boundary: supplied-input stage)
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

- **HIT-V08** (HIT-R02a/HIT-R02b/HIT-R03/HIT-R05/HIT-R06/HIT-R08/HIT-R10; boundary: supplied-input stage)
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

- **HIT-V09** (HIT-R02e/HIT-R03/HIT-R10; boundary: isolated rule)
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

- **HIT-V10** (HIT-R02c/HIT-R07/HIT-R05/HIT-R08; boundary: supplied-input stage)
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

- **HIT-V11** (HIT-R09a/HIT-R08; boundary: supplied-input stage)
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

- **HIT-V12** (HIT-R02d; boundary: isolated rule)
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

- **HIT-V13** (HIT-R02g; boundary: isolated rule)
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

- **HIT-V14** (HIT-R03; boundary: isolated rule)
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

- **HIT-V15** Holy Shield request (HIT-R01b/HIT-R01a; boundary: supplied-input stage)
  - Inputs: player, Stat 31=120, Dex 0, Stat 16=50, Stat 171=0, Stat 182=50, state 101 active, aggregate base present, matching list own 350=10, own 351=2, valid skill 10 with `Skills.calc1` reference 7, shield=true, supplied expression result 25.
  - Seeds: no draw in supplied expression.
  - Draws: none.
  - Trace: request Evaluate(defender,7,10,2); pct 75; base 120+90=210; final add 105.
  - Outputs: defense 315, one ordered evaluation request, zero local draws.
  - Discriminates: treating Holy Shield 25 as a flat addition after the first percent would produce 307, rather than 315.
  - runtime_observed: no

- **HIT-V16** Type sum/bounds (HIT-R02b; boundary: supplied-input stage)
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

- **HIT-V17** Large final defense modifiers (HIT-R01a; boundary: supplied-input stage)
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

- **HIT-V18** Holy Shield gates (HIT-R01b; boundary: supplied-input stage)
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

- **HIT-V19** Player base failure branches (HIT-R02a; boundary: supplied-input stage)
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

- **HIT-V20** Zero level denominator (HIT-R06; boundary: supplied-input stage)
  - Shared inputs: Isolated level-adjustment stage.
  - (a)
    - Inputs: ratio 66, aLvl=0, dLvl=0.
    - Seeds: no draw.
    - Draws: none.
    - Trace: Numerator 0; denominator 0; unguarded signed division faults.
    - Outputs: divide-by-zero fault, no chance result and no draw.
  - Discriminates: An inserted level floor or zero-denominator default would return a chance instead of faulting.
  - runtime_observed: no

### 8.2 Observation vectors

Tool-neutral procedures and exact dated scope are in the [companion](../research/spec-evidence/combat-hit-roll.md#observations).

- **HIT-O01** Observed 116 calls, chance 47, hit roll<=46, miss>=47, no defender step on miss.

## 9. Acceptance and dependencies

The verified supplied-input subset consists of HIT-V01 through HIT-V20 at their stated boundaries, including exact intermediates, requests and seed changes. Passing it does not establish unresolved mastery, monster-relation, neutral-condition, shield or state-applier integration. Full reproduction requires those gates closed and pending gameplay observations. Observation gaps alone do not gate verified supplied-input arithmetic.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md) NRF-R01/NRF-R02 steps/modulus, NRF-R05 ordered mulDiv, RLI-R03/RLI-R02a/RLI-R02b list/state/stat reads; [data fields](core-game-data-fields.md) GDF-R01/GDA-R06a/GDA-R08/GDF-R02/GDF-R03/GDF-R08/GDA-R16 names/schema/flags; [expressions](skill-calc-expressions.md) SCX-R11 and requested producers with their gates/draws; [timed states](state-timed-state-lifecycle.md) STL-R06a/STL-R06b/STL-R06c/STL-R11 for actual prevent-heal effects; [shield block and avoidance](combat-shield-block-and-avoidance.md) SBA-R01e for the supplied shield result, gated by SBA-Q10. The callers in section 2 supply bonus and isMissile; SBA-Q08 retains incomplete caller integration. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** [shield block and avoidance](combat-shield-block-and-avoidance.md) melee and missile result words (SBA-R05/SBA-R06); [missile damage](missile-hit-and-damage-calculation.md) contact hit roll; [PvP](combat-pvp-damage-and-kill-credit.md) defense adjustments and attacker differences; [item event triggers](combat-item-event-triggers-and-procs.md) demon predicate.

## 10. Related specs

[Shield block and avoidance](combat-shield-block-and-avoidance.md) owns the callers, skips, result flags and complete draw order; [skill dispatch](skill-cast-start-and-do-dispatch.md) and [leveled damage/to-hit](skill-leveled-damage-tohit-and-elements.md) locate the melee root and supply the bonus; [missile movement](missile-movement-and-collision.md) and [missile damage](missile-hit-and-damage-calculation.md) locate missile contact; [PvP](combat-pvp-damage-and-kill-credit.md) owns hostility.

## Appendix A. Provenance

Original LoD 1.13c Windows x86 D2Game, D2Common, Fog evidence uses the [source manifest](../research/source-manifest.md). The [companion](../research/spec-evidence/combat-hit-roll.md) preserves dated 2026-10-11..13 static/runtime evidence and correction rationale. Targeted 2026-10-15 reading used matched C where available, otherwise disassembly, including decoding checked against manifest-matching original bytes. Narrow ItemStatCost names were read through the effective loader 2026-10-15. Observed gameplay scope covers the ratio and level-term cases only; no new gameplay campaign ran.
