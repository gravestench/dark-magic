# PvP damage, hostility and kill credit

**Description**

Attacks between players use special damage scaling and rules that decide whether one character can hurt another. Towns prevent player-versus-player damage, and an attack made by a missile or summoned creature can involve its owner when checking opponents or assigning kill credit. Damage scaling, permission to attack, and credit for a kill are separate decisions; this specification describes those decisions and identifies the parts that still need confirmation.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: PVP-R, PVP-Q, PVP-V, PVP-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-pvp-damage-and-kill-credit.md)
Open evidence gaps: [gap file](../plans/gaps/combat-pvp-damage-and-kill-credit.md)

## 1. Scope

This specification owns the server damage-scale selector and its field arithmetic, the hostility and early room/town gates, the owner resolution used by those gates, and the selection of a player for kill-experience credit. It also identifies the experience-routing requests made after credit selection. The existing boundary includes the selector's non-player cases. Damage construction, mitigation and final resource changes belong to [damage pipeline](combat-damage-pipeline-order.md) and [mitigation](combat-damage-mitigation.md); hit rolls/block/avoidance belong to [shield block and avoidance](combat-shield-block-and-avoidance.md); experience amounts belong to [party experience](player-party-experience-sharing.md); the hostility-declaration request that sets the relation flag read here belongs to [hostility declaration](player-hostility-declaration.md). Loot and ear generation, complete party effects, and unspecified caller behavior are outside the verified slice.

## 2. Trigger and ownership

D2Game computes the damage-scale and credit decisions. D2Common supplies class predicates, alignment reads, room/town predicates and the stat-list helpers described below. A connected attack reaches these decisions through the server skill/damage path; missile contact passes its resolved creator as attacker. The queued damage path runs scaling at build time and later applies the stored record without repeating that stage. Direct callers select whether the stage runs again (PVP-R01/PVP-R01a).

Credit selection is reached from the death path with the killer and victim supplied by that path; it does not discover the source of damage independently. The companion records the original dispatch/call chains. Only named requests are covered for experience updates; client delivery and complete scoring effects are not established here.

## 3. Inputs and data

All operation names are illustrative. Integer arithmetic uses the foundations' i32/u32, wrapping and truncation conventions. Damage magnitudes use raw 8.8 life units; leech inputs are integer percentages. Boolean predicates consume their documented nonzero result unless an exact raw result is stated. Units and lists are logical objects, not required memory layouts.

| Input | Semantic source | Width/format | Meaning |
|---|---|---|---|
| game and units | supplied game, attacker, defender, killer, victim; optional unit | identity; type/class/id i32 | Player type 0, monster 1, missile 3; other types remain distinct |
| monster ownership | optional monster data, owner link, owner game/type/id; unit lookup | optional identities; type/id i32 | Link lookup uses the game stored in that link; absent link/game or missing unit produces no replacement |
| missile ownership | owner-link-valid condition, owner type/id and supplied game | boolean; i32 | Owner is the creator; a missing lookup retains the missile in the isolated hostility resolver |
| damage record | physical, fire, lightning, cold, magic, poison magnitude, lifeDrain, manaDrain, staminaDrain; poison/cold/freeze length; burn magnitude/length; stun length | i32 | Exact scaled and retained fields are PVP-R03 |
| damage/result flags | hostilityOverride; result word with hit mask 1 and lethal mask 2 | boolean; u16 result | Other result bits survive the PVP-R06 guards |
| scale inputs | defender/attacker type, hireling predicate; attacker.scaleCondition and defender.scaleCondition | logical predicates | The neutral scaleCondition's hit-time producer remains PVP-Q01/PVP-Q02 |
| monster class flags | MonStats.boss, primeevil, inTown and Align | booleans; Align u8 | GDA-R08a/GDA-R16; invalid/missing class gives false flag predicates |
| difficulty | supplied game's difficulty and corresponding DifficultyLevels.HireableBossDamagePercent | difficulty u8; field i32 | Selected only for hireling versus boss, after hireling-versus-hireling priority; missing difficulty record returns factor 100 |
| relations | acting player's ordered relation nodes: other unit id, flags; player data | id i32, flags u32 | Directional match; first matching node; missing match gives zero. The declaration request for this flag is HOS-R02; its stored effects are HOS-R03 |
| alignment | state 105 alignment list, stat 172 alignment at layer 0; base/extended-list eligibility | state/stat identities; i32 value | GDF-R02/GDF-R03; read through PVP-R05a with RLI-R02a/RLI-R02b |
| contact | Missiles.CollideFriend | u8 | Zero rejects a non-hostile contact before apply |
| defense adjustment | stat 115 item_ignoretargetac, stat 116 item_fractionaltargetac; monsterDefenseExclusion and monsterDefenseHalf | i32; booleans | Supplied inputs to HIT-R02c/HIT-R02d; producers remain HIT-Q03 |
| credit eligibility | optional extended base list, active state membership, credit-state mask, state-definition count | boolean/sets; count i32 | Intersect active states with the supplied credit-state mask; its wider producer meaning is PVP-Q02 |
| credit redirect | ordered primary list chain, each list's creditRedirect condition, owner type/id | optional lists; boolean; i32 | First matching list; base list and secondary chain are not candidates |
| death gates | victim.skipKillExperience; victim stat 13 experience | boolean; i32 | Unknown flag producer is PVP-Q02; positive victim experience is required for routing |
| awards | credited player's data/party id; PetType count; supplied active-hireling lookup; victim/credited/hireling levels; supplied amount helper | optional identity; party u16; count/levels/amount i32 | No-party id 65535; PetType ownership GDA-R04b; amount contracts PXL-R03/PXL-R04/PXL-R05/PXL-R08 |
| scoring | optional game scoring context, killer identity | optional identity | This contract proves only the absent-context no-op and request placement |
| dependency operations | supplied tier-creation requests, award/level and scoring operations, lookups and source predicates | named results and ordered call log | Stored effects must be explicitly supplied; local call routing does not verify their implementations |

Implementation-relevant identities and constants:

| Mapping/constant | Origin | Binding |
|---|---|---|
| Stat/state names and ids above | table data: effective data/global/excel/itemstatcost.txt columns ID/Stat; states.txt columns id/state, named rows in GDF-R02/GDF-R03 | Original consumers select the stated numeric identities in code |
| Hireling classes 271 roguehire, 338 act2hire, 359 act3hire, 560 act5hire1, 561 act5hire2 | Class names/data correspondence GDF-R08; class membership engine code | Fixed class comparisons; not a table-controlled hireling category |
| MonStats.boss/primeevil/inTown/Align; DifficultyLevels.HireableBossDamagePercent; Missiles.CollideFriend | table fields, GDA-R05/GDA-R08a/GDA-R10/GDA-R16 | Named prepared fields selected by code; code defines how they affect a hit |
| Factors 17, 25, 100, 200, 400; alignment conversion; party sentinel 65535 | engine code | Fixed by original consumers |
| scaleCondition, skipKillExperience, credit-state mask and redirect-list producer roles | unresolved wider origin | Known consumer behavior uses supplied conditions; producers remain PVP-Q01/PVP-Q02 |

Original bindings describe the evidence profile; support for custom data remapping is a separate capability. No named original difficulty cell is asserted by this specification: vector field values are explicitly synthetic inputs.

## 4. Rules

**PVP-R01. Scale placement** [verified in 1.13c code (static)]. The shared scale/mitigation stage computes flat-reduction slots, then selects/scales fields by PVP-R02a/PVP-R03/PVP-R04, then reaches defender event 11 subject to DMI-R03's hit condition, before per-type resistance/absorb. DPO-R08a/DPO-R09 and DMI-R03 own the surrounding operations; IEP-R07 owns event traversal. Scaling does not itself subtract life or award leech. The apply entry runs this stage only when its supplied stage-enable argument is nonzero. An already built record is not automatically recognized or exempted.

**PVP-R01a. Runs per route** [verified in 1.13c code (static)].

| Route | Runs of the stage on the same record |
|---|---|
| Queued connected-hit build followed by queued apply, including basic attack do-selector 1 | One at build; queued apply passes enable 0 |
| Direct missile contact or other direct enabled apply, without queued build | One at apply, enable nonzero |
| Direct-stage-then-enabled-apply route | Two: direct stage, then enable 1 |
| Direct-stage-then-disabled-apply route | One direct stage; apply enable 0 |
| Disabled self-apply routes without queued build | Zero |

The route names identify the caller behavior, not mandatory APIs. The companion identifies the original callers and the limits of the recorded caller search. No rule assumes every arbitrary caller runs scaling exactly once. Which Skills rows choose a given do-selector is outside this routing contract.

### PVP-R02. Factor selection

**PVP-R02a. Factor selection** [verified in 1.13c code (static)]. Return signed i32 P, in this order:

```
if attacker is absent or defender is absent or attacker == defender: return 100
if defender is a player:
    if attacker is a player or hireling(attacker) or attacker.scaleCondition: return 17
    return 100
if hireling(attacker) and hireling(defender): return 25
if hireling(attacker) and boss(defender):
    return difficulty.HireableBossDamagePercent if difficulty record exists else 100
if primeevil(attacker) and defender.scaleCondition:
    return 200 if hireling(defender) else 400
return 100
```

hireling is true only for the five monster classes in section 3. boss/primeevil are false for non-monsters, invalid classes or missing records. The data-selected factor is i32 and is not locally capped to the immediate constants' range.

**PVP-R02b. Hit-time scale-condition production** [unresolved]. PVP-Q01/PVP-Q02 retain which summons or other non-hireling units carry scaleCondition at the relevant hit. With that boolean supplied, R02a is independently executable. boss and primeevil are the named table flags, not guessed descriptions of this condition.

**PVP-R03. Field selection** [verified in 1.13c code (static)]. When P != 100, visit physical, fire, lightning, cold, magic, poisonMagnitude, lifeDrain, manaDrain and staminaDrain and apply R04 to positive values. Retain poisonLength, coldLength and freezeLength. burnMagnitude, burnLength and stunLength are absent from the scale table and unchanged. P == 100 skips field writes. There is no draw.

**PVP-R04. Scale arithmetic** [verified in 1.13c code (static)]. For each selected i32 field v > 0, write NRF-R05 mulDiv(P, v, 100), in that operand order. Thus v > 1048576 uses low32((v/100)*P); otherwise P > 65536 uses low32((P/100)*v); otherwise divide the signed wrapped 32-bit product by 100. Division truncates toward zero. Values v <= 0 are untouched. There is no local cap, positive-result floor or whole-life rounding. The familiar P=17 case is 2560 -> 435 raw; adjacent inputs 1048576 and 1048577 produce 178257 and 178245 respectively.

### PVP-R05. Hostility and alignment

**PVP-R05a. Hostility predicate** [verified in 1.13c code (static)]. Resolve attacker and defender as R07. Equal identities return false, including two absent identities. Two players return true only for a matching directional relation node on the attacker's list whose relation eligibility flag is set (the declaration request is HOS-R02; the flag's meaning is HOS-R04); no reverse check occurs. Other pairs return NOT sameTierNeutral(a,b). sameTierNeutral returns false if either unit is absent, true if both non-null identities are equal, otherwise true only when both alignment tiers equal 0 or both equal 2. Tier 1 never qualifies.

Tier read: absent unit or absent base list -> 0. A non-player/non-monster with a present base list -> 2. For player or monster, require an extended base list, find state 105 under RLI-R03 on the primary chain, then secondary if not found, and read that list's own stat 172/layer 0 under RLI-R02a/RLI-R02b. Missing list yields 0. State-list presence, rather than the active-state bit, controls this read. This consumes the existing-array stat rules and does not establish state-list creation.

**PVP-R05b. Tier selections at creation** [verified in 1.13c code (static)]. The recorded sentry/Blade Sentinel/hireling creation sites request tier 2. At the summoned-monster site, MonStats.Align 1 requests tier 2 and sets alignmentCreationCondition; Align 2 requests tier 1; other byte values request tier 0. These are the caller's selected values and, for Align 1, its direct unit-condition write. Persisting a newly allocated/attached alignment list, propagation and notifications are integration PVP-Q07; a supplied existing alignment result is available to R05a independently.

**PVP-R05c. Player-tier production** [unresolved]. The player/monster tier reader is R05a. Which creation or later writers establish a player's actual alignment tier against monsters remains outside-boundary producer PVP-Q04; do not assume every player has tier 2.

**PVP-R06. Early damage gates** [verified in 1.13c code (static)]. At apply entry, run R05a first. If false and hostilityOverride is false, clear hit and lethal and return before scaling or resource changes. Next require defender room; missing room has the same two-bit clearing. If the defender room is a town, proceed only for a monster attacker with a present valid class record and inTown true; otherwise clear lethal only and return. Unrelated result bits and damage magnitudes survive these early returns. Override bypasses hostility alone. Party membership and level are not tested here. Upstream missile contact skips a non-hostile contact if CollideFriend is zero; it does not bypass the apply guards. Passing these guards is permission to continue DPO-R13, not a promise that later guards permit damage.

**PVP-R07. Owner resolution** [verified in 1.13c code (static)]. Hostility's attacker takes one monster-owner hop, or one missile-owner hop when its owner-link-valid condition is true. Hostility's defender repeatedly takes monster-owner hops until the current unit is non-monster or the link/game/lookup is missing, or owner == current unit. A failed replacement retains the current unit. There is no depth cap and no multi-unit cycle detector. The missile-contact path resolves its creator before invoking apply/results; a sentry-created bolt therefore supplies the sentry, and the hostility predicate can take the additional sentry-to-player hop. Absent contact owner can stop contact under SBA-R13; the isolated resolver's missing-owner fallback does not override that caller.

### PVP-R08. Related effect inputs

**PVP-R08a. Leech inputs** [verified in 1.13c code (static)]. lifeDrain/manaDrain/staminaDrain are among R03's scaled fields. With P=17, positive inputs 3/8/12 become 0/1/2 before the later leech computation. Non-monster leech-effectiveness input is 100 under LSE-R02. This proves the input transformation, not a final amount healed or mana gained.

**PVP-R08b. Defense inputs** [verified in 1.13c code (static)]. Consume HIT-R02c and HIT-R02d for the player-attacker defense-adjustment slice: stat 115 does not ignore a player's defense; positive stat 116 is halved against a player, hireling, boss or monsterDefenseHalf monster and capped at 100. HIT-R02c's exclusion condition controls stat 115 on other monsters. Full attack rating, defense reduction, hit probability and draws remain HIT-R01 through HIT-R08; this rule adds no PvP hit roll.

**PVP-R08c. Other effects** [unresolved]. Block/avoidance behavior is owned by SBA-R09a, Crushing Blow by [Crushing Blow](combat-crushing-blow.md), and Open Wounds by [Open Wounds](combat-open-wounds.md). Their current labels, gates and target-specific arithmetic apply at their own boundaries. This rule supplies no additional PvP multiplier or claim that such effects are absent. Complete combined effect/credit behavior is PVP-Q05.

### PVP-R09. Kill credit

**PVP-R09a. Player credit and death eligibility** [verified in 1.13c code (static)]. The reached death-experience wrapper skips when killer/victim is absent or victim.skipKillExperience is true. Award routing then requires killer type player/monster, a present victim and victim.experience >0. The isolated credited-player resolver returns a player killer immediately, without testing either unit's redirect lists; a non-monster/non-player or absent killer gives none. For a monster killer, start with its one-hop monster owner (or none), apply R09b, and retain the result only if it is a player. A minion's monster owner does not automatically resolve a second hop. Missile contact's prior creator resolution is a separate R07 boundary.

**PVP-R09b. Redirect selection** [verified in 1.13c code (static)]. For a monster killer only, set selectedList absent. If creditEligible(killer), assign firstCreditList(killer), including absent. If victim exists and creditEligible(victim), assign firstCreditList(victim), including absent; this replaces the earlier selection. If selectedList exists, replace the one-hop owner result with lookup(game argument, selectedList.ownerType, selectedList.ownerId), including a missing lookup result. If selectedList is absent, retain the original one-hop owner result. Then R09a requires player type.

creditEligible returns false for absent unit, missing/non-extended base list or absent active-state membership; otherwise true iff any of the scanned membership words intersects the supplied credit-state mask. The word count is trunc(i32(low32(stateCount+31))/32); nonpositive count scans nothing and returns false. For ordinary nonnegative counts this is ceiling(stateCount/32). firstCreditList requires an extended base list, then visits primary-chain entries in order, returning the first creditRedirect entry; it never tests the base list or secondary chain. Owner type/id are read from that selected list. No ownership lookup or eligibility operation draws a random number within this defined slice.

**PVP-R09c. Redirect producers** [asm-only]. R09b establishes the active-state predicate, first-list search and selected owner reads. Which original gameplay effects create the selected lists or populate the credit-state mask remains outside-boundary PVP-Q02. No game-effect name is inferred from a flag or accessor label.

### PVP-R10. Experience and scoring

**PVP-R10a. Experience/scoring routing** [verified in 1.13c code (static)]. After R09a selects a player, the credited player's active-hireling lookup is attempted only with present player data and signed PetType count >7. A found hireling receives an experience-update request using the supplied scaled amount; if it is not the killer, apply R10b's share first. Next, party id 65535 selects solo-amount then level-update requests; other party ids select party-distribution request. This does not perform the award inside the supplied callees. Separately, the reached death path with a killer requests scoring; absent scoring context returns immediately before score writes. Full scoring when present and complete active-hireling lookup are PVP-Q09.

**PVP-R10b. Experience amounts** [verified in 1.13c code (static)]. PXL-R03/PXL-R04/PXL-R05/PXL-R08 own solo/party amount calculations and their existing gates. For a supplied nonnegative hireling amount x when that hireling is not the killer, the local share is trunc(i32(low32(x*86))/256); do not replace this with an unbounded 64-bit multiplication. When the hireling is the killer, retain x. This specification's executable share example uses x=1000 and produces 335; it does not establish a complete integrated award.

**PVP-R11. Same-party predicate** [verified in 1.13c code (static)]. Independently replace each monster input by at most one found owner, using its link's game. Equal resulting identities return true, including two absent inputs. For two players, read their party id only with present game context; otherwise use 65535. Return true only for equal ids other than 65535. Other pairs return sameTierNeutral from R05a, without inversion. The damage guards do not call this predicate.

**PVP-R12. Local RNG** [verified in 1.13c code (static)]. The factor selector, scale arithmetic, relation/tier/owner predicates, credit resolver, share arithmetic and routing decisions above make no direct random step. Their supplied getters/lookups and notification/award operations are explicit execution boundaries. Additional RNG in caller damage build, callbacks, state appliers, full party/award operations or other unverified callees is owned there. A no-draw fixture for this slice does not assert that a complete PvP attack/death leaves every seed unchanged.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| PVP-R01 | Stage placement | verified in 1.13c code (static) | none at stated stage/request boundary | PVP-V01, PVP-V08 |
| PVP-R01a | Runs per route | verified in 1.13c code (static) | none for recorded routes | PVP-V01, PVP-V08 |
| PVP-R02a | Factor selection | verified in 1.13c code (static) | PVP-Q01/PVP-Q02 actual scaleCondition producer only | PVP-V01, PVP-V06, PVP-V09, PVP-V22 |
| PVP-R02b | Hit-time condition producer | unresolved | PVP-Q01/PVP-Q02 actual carrier | PVP-V06 supplied condition only |
| PVP-R03 | Selected fields | verified in 1.13c code (static) | none | PVP-V01, PVP-V05, PVP-V10, PVP-V22 |
| PVP-R04 | Operand order and scale arithmetic | verified in 1.13c code (static) | none | PVP-V01, PVP-V02, PVP-V11, PVP-V22 |
| PVP-R05a | Directional hostility and tier read | verified in 1.13c code (static) | PVP-Q04 actual player tier producer only | PVP-V04, PVP-V12, PVP-V24 |
| PVP-R05b | Creation tier selection/request | verified in 1.13c code (static) | PVP-Q07 stored list/propagation integration | PVP-V13 |
| PVP-R05c | Player tier producer | unresolved | PVP-Q04 producer | none |
| PVP-R06 | Early legality/room/town guards | verified in 1.13c code (static) | none for supplied predicates | PVP-V03, PVP-V04, PVP-V14 |
| PVP-R07 | Owner depth and fallback | verified in 1.13c code (static) | contact callers retain SBA-Q08 integration | PVP-V06, PVP-V07, PVP-V15 |
| PVP-R08a | Leech input transformation | verified in 1.13c code (static) | LSE gates only for final leech effects | PVP-V05 |
| PVP-R08b | Defense-input dependency | verified in 1.13c code (static) | HIT-Q03 producers only | PVP-V16 |
| PVP-R08c | Combined effect integration | unresolved | PVP-Q05 complete cross-effect parity | none |
| PVP-R09a | Credit selection and death gates | verified in 1.13c code (static) | PVP-Q03 gameplay observation only | PVP-V07, PVP-V17, PVP-V23, PVP-V25 |
| PVP-R09b | Active-state redirect selection | verified in 1.13c code (static) | PVP-Q02 supplied mask/list producers only | PVP-V17, PVP-V23 |
| PVP-R09c | Wider redirect-list meaning | asm-only | PVP-Q02 original gameplay producer | none |
| PVP-R10a | Award/scoring requests | verified in 1.13c code (static) | PVP-Q09 full lookup/scoring; PXL dependency gates | PVP-V18, PVP-V25 |
| PVP-R10b | Supplied share and amount owner | verified in 1.13c code (static) | PXL gates for integrated amounts only | PVP-V19 |
| PVP-R11 | Same-party predicate | verified in 1.13c code (static) | PVP-Q04 tier producer only | PVP-V20, PVP-V24 |
| PVP-R12 | Owned direct no-draw slice | verified in 1.13c code (static) | external operations retain own draw contracts | PVP-V21 |

## 5. Outputs and side effects

| Output/write/request | Rule |
|---|---|
| Selected signed factor; selected positive damage/leech fields replaced; unselected fields retained | PVP-R02a/PVP-R03/PVP-R04 |
| Directional eligibility/same-tier result and resolved identities; no unit replacement is persisted | PVP-R05a/PVP-R07/PVP-R11 |
| Creation tier request and Align-1 alignmentCreationCondition=true | PVP-R05b; attached/stat propagation gated PVP-Q07 |
| Hit/lethal bit clearing on early return; no life/leech writes at this boundary | PVP-R06 |
| Defense/leech inputs transformed before their owning consumers | PVP-R08a/PVP-R08b |
| Credited player or none; selected redirect list and owner lookup arguments | PVP-R09a/PVP-R09b |
| Active-hireling lookup, hireling-update, solo-amount, level-update, party-distribution and scoring requests; supplied hireling share | PVP-R10a/PVP-R10b |
| No direct seed mutation in owned operations | PVP-R12 |

Resource subtraction, final leech, experience stat writes, stored relation activation, state-list attachment, scoring with a present context and client delivery require their owning contracts and gates. No persistence format, unit allocation or event schedule is introduced by the verified local slice.

## 6. Branches and edge cases

Scale selection does not itself check hostility or town. Guard rejection at apply can occur after a queued record has already been scaled. Nonpositive magnitudes are retained even for zero/negative P; positive values can become zero or negative. Missing difficulty records fall back to 100. Arithmetic at the large-value threshold is discontinuous and wraps where R04 states.

Hostility and same-party use different owner depths and opposite tier results; a party match is not a damage guard. A directional relation need not exist in reverse. Owner chains can be deeper than one on the defender side; a cycle longer than one can fail to terminate. Invalid/missing monster class records disable boss/primeevil/inTown predicates. A missing contact creator is not the same boundary as the isolated hostility resolver's fallback.

A player killer bypasses redirect tests. A qualifying victim with no matching credit list erases the killer-side list selection and retains the original owner; a present redirect whose unit lookup fails produces no credited player. No additional classic/expansion branch exists in the local selector or predicates; dependencies retain their own differences.

## 7. Open questions

| Id | Kind/status | Missing evidence and affected boundary |
|---|---|---|
| **PVP-Q01** | outside-boundary / open | Establish scaleCondition carried by each sentry/summon at hit time, using applicable creation/update paths or an attributable hit. This does not gate R02a with supplied true/false |
| **PVP-Q02** | outside-boundary / open | Establish wider producer meanings of scaleCondition, skipKillExperience, credit-state mask and creditRedirect lists. Read actual producers and trace their use; known consumer predicates/owner reads remain executable |
| **PVP-Q03** | observation / pending | Observe a player-victim kill's actual experience routing with known raw victim experience and attribution. The positive supplied-input route remains statically available |
| **PVP-Q04** | outside-boundary / open | Identify player alignment-list writers and their current values against monsters; R05a's supplied-list reader is independent |
| **PVP-Q05** | integration / open | Establish complete combined effects/credit outcomes through the current Crushing Blow/Open Wounds/avoidance owners, preserving their gates. Loot and ear production remain outside this spec and unestablished; no completed absence search is claimed |
| **PVP-Q06** | observation / pending | Observe a plain player melee hit at both build and apply, expecting one 17-percent scale for PVP-V01/PVP-V08. Stage-location evidence is static |
| **PVP-Q07** | integration / open | Complete alignment-list allocation/attachment/propagation and creation notifications after R05b's selected tier. Reuse STL/NRF state/stat contracts where applicable; source requests do not prove their stored effects |
| **PVP-Q08** | observation / pending | Reproduce the verified behavioral assertions on original 1.13c with attributable inputs, reached stage and measured outputs. No executable vector below is newly gameplay-observed |
| **PVP-Q09** | integration / open | Complete active-hireling lookup and score writes with a present scoring context, and dependency-backed experience updates. R10a's supplied lookup/routing and absent-context no-op remain available |

Detailed attempts and remaining limits are in the [gap file](../plans/gaps/combat-pvp-damage-and-kill-credit.md#open-evidence-gaps).

## 8. Test vectors

### 8.1 Executable vectors

Each vector/case starts afresh. Unless overridden: distinct present player A/B, type 0, level 9, field rooms present/not town; A has a directional eligible relation to B; damage physical 2560, hit/lethal result 3 with other bits zero; all other damage fields/stat values/conditions zero; monster-specific data and definitions present for supplied valid monsters; no ownership link; no party (65535); scoring context absent; no skipKillExperience; victim experience 1000 when a credit/award route is invoked. Base/extended lists and counts are supplied explicitly where read. Definitions for cited stats exist with min-adjustment disabled; stat definition count is 359 and state definition count 185 where required. Supplied MonStats count is 734: a fixture described as a hireling has class 271; other valid monster fixtures have class 0 unless overridden. These are synthetic class records, with flags supplied by each case, not new original table-cell claims. Tier fixtures use extended base lists with one matching state-105 list and its own stat 172/layer 0 equal to the supplied tier; unspecified state-list chains are empty. Player A's eligible node has flags 8, and an ineligible node has flags 0 unless a case states 256. Unstated callbacks/award/notification operations are supplied no-op/log operations with no draws; they record call arguments/order and persist nothing. Actual owned factor/arithmetic/owner/relation/tier/credit helpers execute rather than being replaced by their expected values. These fixtures do not execute damage construction, mitigation, callback effects, complete hireling search or experience-award callees. All cases have Seeds: no draw; Draws: none; runtime_observed: no. Logical field names express the fixture values; the companion maps them to original binary evidence.

- **PVP-V01** Single-route player scaling (PVP-R01/PVP-R01a/PVP-R02a/PVP-R03/PVP-R04; boundary: supplied-input stage)
  - Inputs: Independently run queued build+disabled apply, direct missile-enabled apply and other direct-enabled apply, with defaults and no later resource operations.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Each route reaches selection once: P=17; 2560*17/100=435. Log scale positions relative to build/apply.
  - Outputs: physical=435 before resistance/flat subtraction; queued log build, scale, enqueue, apply(enable 0); direct log apply(enable 1), scale. No second scale.
  - Discriminates: No melee exemption or second queued scale.
  - runtime_observed: no.

- **PVP-V02** Large magnitude (PVP-R04; boundary: isolated rule)
  - Inputs: P=17; physical=1048577.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 1048577/100=10485; 10485*17=178245.
  - Outputs: physical=178245.
  - Discriminates: Divide-first result differs from multiply-first 178258.
  - runtime_observed: no.

- **PVP-V03** Town rejection (PVP-R06; boundary: supplied-input stage)
  - Inputs: Defaults with defender town=true. Stop at early guards.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Hostility passes; room present; player attacker fails monster-inTown exception.
  - Outputs: result=1 (hit retained, lethal cleared); physical=2560; no scale/life/leech request.
  - Discriminates: Town clears lethal only and stops before resources.
  - runtime_observed: no.

- **PVP-V04** Missing directional relation (PVP-R05a/PVP-R06; boundary: supplied-input stage)
  - Inputs: A has no node for B, hostilityOverride=false; B may have an eligible reverse node.
  - Seeds: no draw.
  - Draws: none.
  - Trace: R05a false; first guard clears both low result bits.
  - Outputs: result=0; physical=2560; no scale/resource request.
  - Discriminates: Reverse relation does not grant permission.
  - runtime_observed: no.

- **PVP-V05** Leech percentage input (PVP-R03/PVP-R08a; boundary: isolated rule)
  - Inputs: lifeDrain=8, physical=2560. Case (a) P=17; case (b) P=100.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) 8*17/100=1; (b) P = 100 skips writes.
  - Outputs: (a) lifeDrain=1, physical=435; (b) lifeDrain=8, physical=2560. No final heal asserted.
  - Discriminates: Scale percentage input before leech arithmetic.
  - runtime_observed: no.

- **PVP-V06** Sentry creator and supplied condition (PVP-R02a/PVP-R02b/PVP-R07; boundary: supplied-input stage)
  - Inputs: Missile creator is monster S; S has one-hop owner A. Contact supplies S to scaling; defender B player. S not a hireling. Independent S.scaleCondition true/false.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Hostility resolves S once to A; scaling keeps S as its attacker.
  - Outputs: true: P = 17, physical = 435; false: P = 100, physical = 2560; hostility true in both. Actual sentry condition producer remains gated.
  - Discriminates: Do not replace scaling attacker with player owner.
  - runtime_observed: no.

- **PVP-V07** Sentry kill credit (PVP-R07/PVP-R09a; boundary: supplied-input stage)
  - Inputs: Killer monster S has valid link to player A; victim monster V. No active states, no credit-mask intersection; base lists absent.
  - Seeds: no draw.
  - Draws: none.
  - Trace: One owner lookup yields A; eligibility false; final player test passes.
  - Outputs: credited=A; exactly one owner lookup; no redirect lookup; no second hop.
  - Discriminates: Credit uses immediate monster owner.
  - runtime_observed: no.

- **PVP-V08** Route multiplicity (PVP-R01/PVP-R01a; boundary: supplied-input stage)
  - Inputs: Defaults, no mitigation or resource operations. Independent routes: queued; disabled self-apply; enabled direct apply; direct-stage then enabled; direct-stage then disabled. The self-apply case uses attacker == defender; its unchanged magnitude assertion does not require passing the hostility guard. Route calls are logged and the reached scale body executes; complete caller construction is not executed.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Each reached scale maps 2560->435; next scale maps 435->73.
  - Outputs: queued 435; disabled self 2560; enabled 435; direct+enabled 73 with two logged scales; direct+disabled 435 with one.
  - Discriminates: Stage-enable controls apply; no automatic already-scaled marker.
  - runtime_observed: no.

- **PVP-V09** Other selector factors (PVP-R02a; boundary: isolated rule)
  - Inputs: physical 2560, distinct identities unless absent. Cases (a) attacker absent; (b) both hireling monsters; (c) primeevil monster attacker, ordinary non-hireling monster defender scaleCondition=true; (d) same with hireling defender.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Factors 100/25/400/200; apply R04 once.
  - Outputs: (a) 2560; (b) 640; (c) 10240; (d) 5120; exact factors 100/25/400/200.
  - Discriminates: Preserve factor priority and distinguish 200 from 400.
  - (e) Inputs: fresh distinct hireling monsters, each class 271 with valid records; defender.boss=true; available difficulty.HireableBossDamagePercent=37; physical=2560. Draws: none. Trace: both-hireling branch returns 25 before the hireling-versus-boss branch can read its field. Outputs: P=25, physical=640; no HireableBossDamagePercent field read. This rejects prioritizing the boss factor 37 (which would produce physical=947).
  - (f) Inputs: fresh hireling monster attacker, class 271 with valid record, scaleCondition=false; distinct player defender; physical=2560. Draws: none. Trace: player-defender branch recognizes the hireling before consulting the neutral scale condition. Outputs: P=17, physical=435. This rejects treating a hireling with scaleCondition=false as an ordinary unscaled monster attacker.
  - runtime_observed: no.

- **PVP-V10** Selected fields and exclusions (PVP-R03; boundary: isolated rule)
  - Inputs: P = 17; poisonLength 200, coldLength 100, freezeLength 75, burnMagnitude 512, burnLength 23, stunLength 50, fire 1000; other scaled magnitudes zero.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Only fire positive in scaled fields:17000/100=170.
  - Outputs: fire 170; lengths 200/100/75, burn 512/23, stun 50 unchanged; all other scaled fields remain zero.
  - Discriminates: Duration/burn/stun are not scaled.
  - (b) Inputs: fresh P=17; physical=2560, fire=1000, lightning=1100, cold=1200, magic=1300, poisonMagnitude=1400, lifeDrain=8, manaDrain=18, staminaDrain=28; poisonLength=200, coldLength=100, freezeLength=75, burnMagnitude=512, burnLength=23, stunLength=50. Draws: none. Trace: multiply each selected positive input by 17, divide by 100 with truncation; the products are 43520, 17000, 18700, 20400, 22100, 23800, 136, 306 and 476 in the selected field order. Outputs: physical=435, fire=170, lightning=187, cold=204, magic=221, poisonMagnitude=238, lifeDrain=1, manaDrain=3, staminaDrain=4; excluded lengths remain 200/100/75, burn 512/23 and stun 50. Each selected field has its own output assertion, rejecting an omitted field or swapped outputs.
  - runtime_observed: no.

- **PVP-V11** Nonpositive and threshold arithmetic (PVP-R04; boundary: isolated rule)
  - Inputs: Independent P = 17, v=0, -300, 1, 1048576.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Nonpositive untouched; 17/100=0; 17825792/100=178257.
  - Outputs: 0, -300, 0, 178257 respectively.
  - Discriminates: No positive floor; multiply-first includes threshold equality.
  - (b) Inputs: fresh isolated scale arithmetic, P=65535, positive v=65535. Draws: none. Trace: neither threshold is exceeded; the product 4294836225 wraps to low32 4294836225, interpreted as i32 -131071; signed division by 100 truncates to -1310. Outputs: scaled field=-1310. A wide unwrapped product would yield 42948362 and is rejected; no positive floor applies.
  - runtime_observed: no.

- **PVP-V12** Directional relation and equal tiers (PVP-R05a; boundary: isolated rule)
  - Inputs: Cases (a) players A relation flags 8 only; (b) player node with flags 256 only; (c) distinct monsters tiers 1/1; (d) 0/0; (e) 2/2; (f) 0/2. For monsters: extended base lists and matching state 105 lists own stat 172 supplied as stated, stateCount 185, statCount 359. No active-state bit is required.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Player relation directly; monsters invert sameTierNeutral.
  - Outputs: true, false, true, false, false, true respectively.
  - Discriminates: Tier 1 is not neutral even when equal; unrelated relation flag is insufficient.
  - (g) Inputs: fresh distinct players A/B; A's relation chain has two consecutive nodes for B in this order: flags 0, then flags 8. No ownership links. Draws: none. Trace: the first matching node returns raw 0; the later node is not examined for eligibility. Outputs: hostility=false. This rejects accepting any eligible duplicate rather than the first matching node.
  - (h) Inputs: attacker and defender are the same present player A; its synthetic relation list contains an eligible self-node (otherId=A.id, flags=8). Draws: none. Trace: equal resolved identities require false regardless of the self-node flag. Outputs: hostility=false. This rejects granting permission from the eligible self-node.
  - runtime_observed: no.

- **PVP-V13** Summon alignment selection (PVP-R05b; boundary: supplied-input stage)
  - Inputs: Fresh summoned monster, alignmentCreationCondition=false; synthetic MonStats.Align 1, 2, 0, 3; tier request logs only.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Select 1->2, 2->1, other ->0. Align 1 sets direct creation condition.
  - Outputs: requested tiers 2, 1, 0, 0 respectively; condition true only for Align 1; no attached-list result asserted.
  - Discriminates: Do not copy Align directly or invent stored effects.
  - runtime_observed: no.

- **PVP-V14** Town exception and override (PVP-R06; boundary: supplied-input stage)
  - Inputs: Attacker M is a monster with valid class and no owner; defender B is a player. Both have the tier-list inputs from the preamble. Cases (a)/(b): M tier 0, B tier 2, so the actual hostility helper returns true; defender is in town; M.inTown is true/false respectively. Cases (c)/(d): both tiers 0, so the actual helper returns false; hostilityOverride=true, M.inTown=false; defender is in the field/town respectively. Stop at the early guards.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) exception passes; (b) town clears lethal; (c) override skips hostility rejection; (d) override does not skip town.
  - Outputs: (a) continue, result 3; (b) return, result 1; (c) continue, result 3; (d) return, result 1. No final life/leech asserted.
  - Discriminates: Override alone cannot authorize town damage.
  - (e) Inputs: fresh case-(a) tier-list inputs (M tier 0, B tier 2), defender town, hostilityOverride=false, M.inTown=true supplied in both independent runs; MonStats count 734. Run 1: M.class=-1. Run 2: M.class=734, equal to the supplied count and therefore out of range. Draws: none. Trace: either out-of-range class makes the actual inTown predicate false despite the supplied flag. Outputs in each run: return before scale/resource operations, result=1 and physical=2560. This rejects using the inTown flag without validating the class index.
  - (f) Inputs: fresh valid-class case-(b) tier inputs (M tier 0, B tier 2), defender town, M.inTown=false, hostilityOverride=false; incoming result=7, physical=2560. Draws: none. Trace: town rejection clears lethal mask 2 only: 7 AND NOT 2 = 5. Outputs: return before scale/resource operations, result=5, physical=2560. This rejects clearing unrelated result bit 4 or overwriting the word with 1.
  - (g) Inputs: fresh monster M and player B with actual tier-list values 0/0, so hostility=false; hostilityOverride=true; B has no room; incoming result=7, physical=2560. Draws: none. Trace: override passes the hostility gate, then missing defender room clears hit and lethal: 7 AND NOT 3 = 4. Outputs: early return, result=4, physical=2560; no town test, scale or resource operation. This rejects applying override to the room requirement or clearing only lethal.
  - runtime_observed: no.

- **PVP-V15** Asymmetric owner depth (PVP-R07; boundary: isolated rule)
  - Inputs: Monster M2->monster M1->player A, all valid links and lookup results; independently resolve defender and attacker. Another case: valid missile link whose lookup is absent.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Defender takes 2 hops; attacker 1; failed missile lookup retains missile.
  - Outputs: defender=A with 2 lookups; attacker=M1 with 1; missing owner=original missile with 1 lookup.
  - Discriminates: No shared ultimate-owner resolver for both sides.
  - (b) Inputs: fresh defender monster M2 has present monster data and owner link with a present owner game and stored M1 type/id, but the lookup returns absent. Draws: none. Trace: the first replacement fails and the defender loop stops. Outputs: resolved defender=M2, exactly one lookup, no lookup of M1's owner. This rejects replacing a failed defender lookup with null.
  - (c) Inputs: fresh attacker missile X has owner-link-valid=false and stored type/id for present player A; the lookup would return A if requested. Draws: none. Trace: the invalid-link condition bypasses lookup. Outputs: resolved attacker=X; lookup log empty. This rejects resolving stored owner fields without testing link validity.
  - (d) Inputs: fresh defender monster M3 -> monster M2 -> monster M1 -> player A, each link and owner game present, all three lookups successful. Draws: none. Trace: defender resolution continues through all three links until the player. Outputs: resolved defender=A, lookup order M2, M1, A; exactly three lookups. This rejects a two-hop cap that would stop at M1.
  - (e) Inputs: fresh defender monster M with valid owner link and present owner game; its stored owner type/id lookup returns M itself. Draws: none. Trace: owner == current unit stops the loop after that successful lookup. Outputs: resolved defender=M; exactly one lookup and no repeat. This rejects repeatedly traversing a self-owner link.
  - runtime_observed: no.

- **PVP-V16** Player defense adjustments (PVP-R08b; boundary: supplied-input stage)
  - Inputs: Invoke HIT-R02c/HIT-R02d slice only. Fresh DEF = 1000, stat 115=0, stat 116 values:50 vs player; 250 vs player; 50 vs ordinary monster; 50 vs hireling; 51 vs boss monster. Ordinary monster has boss/monsterDefenseHalf/monsterDefenseExclusion false. Extra cases stat 115=1, stat 116=0 vs player and vs ordinary monster with monsterDefenseExclusion=true.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Half positive values where required: 50 -> 25, 250 -> 125, 51 -> 25; cap 125 to 100. Compute cut = DEF * p / 100.
  - Outputs: effective percentages 25, 100, 50, 25, 25; DEF = 750, 0, 500, 750, 750. Last stat 115 cases DEF = 1000 unchanged.
  - Discriminates: Odd 51 halves to 25; stat 115 does not bypass players/excluded monsters.
  - runtime_observed: no.

- **PVP-V17** Player bypass and victim redirect (PVP-R09a/PVP-R09b; boundary: supplied-input stage)
  - Inputs: Player P1 killer; victim has eligible primary credit list owned P2; killer list owned P3. Then monster M killer one-hop owner P1 with eligible primary list owned P3, victim eligible list owned P2. Eligibility uses extended base lists, active state 1 and supplied mask {1}, stateCount = 32. All owner lookups succeed; list creditRedirect=true.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Player returns immediately. Monster selects killer list then victim list.
  - Outputs: player kill=P1 with no eligibility/list calls; monster kill=P2 with owner lookup P1 then redirect lookup P2, no lookup P3.
  - Discriminates: Victim selection supersedes killer selection; player bypass is before both.
  - runtime_observed: no.

- **PVP-V18** Solo routing and absent scoring (PVP-R10a; boundary: supplied-input stage)
  - Inputs: Credited player A, party 65535, playerData present, PetType count 7 (skip hireling), victim experience 1000; supplied solo amount 321; scoring context absent.
  - Seeds: no draw.
  - Draws: none.
  - Trace: No hireling search; request solo amount then level update; separate scoring call returns at context guard.
  - Outputs: ordered award log soloAmount(A, V), levelUpdate(A, 321); no party/hireling update or score write.
  - Discriminates: Solo/party routing distinct; no absent-context scoring.
  - runtime_observed: no.

- **PVP-V19** Hireling share (PVP-R10b; boundary: isolated rule)
  - Inputs: Supplied x 1000, hireling not killer; independent hireling==killer.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 86000/256 truncates 335; identity case bypasses share.
  - Outputs: 335; 1000 respectively.
  - Discriminates: Use 86/256 rather than one third or applying share to killer hireling.
  - (b) Inputs: fresh supplied amount x=50000000; hireling is not the killer. Draws: none. Trace: x*86=4300000000; low32 product=5032704, positive as i32; trunc(5032704/256)=19659. Outputs: share=19659. A wide multiplication without the required product wrap would return 16796875 and is rejected.
  - runtime_observed: no.

- **PVP-V20** Same-party inputs (PVP-R11; boundary: isolated rule)
  - Inputs: Cases (a) players party 3/3, game present; (b) 65535/65535; (c) same unit A; (d) monster owner A vs A; (e) player A game absent, other party 3. All owner links are valid when provided.
  - Seeds: no draw.
  - Draws: none.
  - Trace: At most 1 hop each, identity check, then party ids.
  - Outputs: true, false, true, true, false; case (d) one owner lookup.
  - Discriminates: Sentinel is not a shared party; identity follows owner resolution.
  - (f) Inputs: fresh first input monster M -> monster M1 -> player A with both owner links valid and all requested lookups successful; second input is player A. M1 and A each have the preamble's tier-list inputs with own stat 172=1. Draws: none. Trace: resolve M only once to M1; A remains A; distinct identities and a monster/player pair use sameTierNeutral, which is false for tiers 1/1. Outputs: sameParty=false, exactly one owner lookup returning M1; no lookup of M1's owner. A recursive owner resolver would reach A and wrongly return true on identity.
  - runtime_observed: no.

- **PVP-V21** No direct RNG (PVP-R12; boundary: supplied-input stage)
  - Inputs: Execute selector/scale, hostility/owner, same-party, credited-player and routing slices with the exact PVP-V01/PVP-V07/PVP-V18/PVP-V20 fixtures independently. The declaration checks' no-draw assertion is HOS-V03. Every external operation is the stated no-draw logger; initialize involved game/unit seed pairs {1, 666}.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Only owned decisions and no-draw supplied operations execute.
  - Outputs: All supplied seeds remain {1, 666}; zero direct random-step requests. No complete attack/death seed claim.
  - Discriminates: Reject adding a local random draw without asserting unexecuted callback behavior.
  - runtime_observed: no.

- **PVP-V22** Data-selected scale factor (PVP-R02a/PVP-R03/PVP-R04; boundary: supplied-input stage)
  - Inputs: Hireling attacker, boss non-hireling monster defender; supplied difficulty.HireableBossDamagePercent=70001, physical=1234. Fresh cases field = -50 and 0; missing difficulty record; attacker == defender.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 70001>65536:70001/100*1234=700*1234=863800; negative factor uses product/100; missing/self P = 100.
  - Outputs: P = 70001->863800; P = -50->-617; P = 0 ->0; missing record P = 100->1234; self P = 100->1234. With P = 100, write log empty.
  - Discriminates: Data factor is not capped to 65536; operand-order/divide-first differs from 863812.
  - (b) Inputs: fresh same hireling/boss types and available difficulty record; HireableBossDamagePercent=70001, physical=1048577. Draws: none. Trace: both thresholds are exceeded, so the v-first branch takes precedence: trunc(1048577/100)=10485; 10485*70001=733960485. Outputs: P=70001, physical=733960485. The wrong P-first branch would produce trunc(70001/100)*1048577=734003900 and is rejected.
  - runtime_observed: no.

- **PVP-V23** Victim null erasure and eligibility (PVP-R09a/PVP-R09b; boundary: supplied-input stage)
  - Inputs: Monster M one-hop owner P1. Extended base with primary lists[nonredirect, redirect(owner P3), redirect(owner P4)], secondary redirect P5; active state 1 / mask {1}, count 32. Victim(a) eligible with no matching list; (b) ineligible; (c) eligible matching list owner P2 whose lookup fails. All players distinct; P1/P3 lookup succeeds.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Killer list chosen is first primary P3. (a) victim selection absent clears it; (b) skip victim selection; (c) victim list replaces it, then missing lookup clears result.
  - Outputs: (a) P1, no redirect lookup; (b) P3, redirect lookup P3 once; (c) none, redirect lookup P2 once. No lookup P4/P5. Additional independent eligibility controls: active {1}, mask {2}->false; missing membership -> false; stateCount = 0 -> false; non-extended base -> false.
  - Discriminates: Absent victim list differs from missing owner lookup; secondary/base are not searched.
  - (d) Inputs: isolated creditEligible call, fresh unit with extended base list and present active membership; stateCount=33; active states {32}; supplied credit-state mask {32}. Draws: none. Trace: trunc((33+31)/32)=2 words; both first words are 0, both second words are 1, so the second intersection is nonzero. Outputs: creditEligible=true. This rejects truncating the scan to one word or ignoring the partial final word.
  - (e) Inputs: fresh monster killer M with owner-link game X, owner type player/id 1; lookup in X for id 1 returns player P1. The supplied kill-credit game is G, distinct from X. M is creditEligible with an extended base, active state 1/mask {1}, count 32 and a first primary creditRedirect list whose owner type/id is player/42; victim is ineligible. Registry G[player,42] returns P2; registry X[player,42] would return distinct P3. Draws: none. Trace: obtain initial owner P1 through X, then selected-list owner through the supplied game G. Outputs: credited=P2; ordered lookup log (X,player,1)->P1 then (G,player,42)->P2; no (X,player,42) lookup. This rejects reusing the monster link's game for redirect credit.
  - (f) Inputs: fresh monster killer M, one-hop owner player P1; creditEligible=true through extended base, active state 1/mask {1}, stateCount=32. The extended base itself has creditRedirect=true and owner P6, whose lookup would succeed. Its primary chain has only nonredirect entries; secondary chain empty; victim ineligible. Draws: none. Trace: firstCreditList searches primary attached entries rather than the extended base, finds no match, and leaves selectedList absent. Outputs: credited=P1. Selecting the base would redirect to distinct P6 and is rejected.
  - runtime_observed: no.

- **PVP-V24** Tier-reader presence and identity (PVP-R05a/PVP-R11; boundary: isolated rule)
  - Inputs: Stat definition count 359, state definition count 185, minimum adjustments disabled. Cases (a) player A extended base list, primary matching state 105 with own stat 172=2, active state 105 false; distinct monster B has its own matching state-105 list with stat 172=2; (b) A has no matching state-105 list; B retains its matching list/tier 2; (c) other-type unit with base present; (d) other-type unit with base missing; (e) two absent inputs.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) read the matching list despite inactive state bit; (b) tier 0; (c) tier 2; (d) tier 0; (e) identity test before fallback.
  - Outputs: (a) hostility = false, sameTierNeutral = true; (b) hostility = true against B; (c) tier 2; (d) tier 0; (e) hostility = false, sameParty = true.
  - Discriminates: Do not require active state 105 or give every other-type unit tier 2 without a base. The consumed state-list lookup's secondary fallback is separately discriminated by RLI-V05 case (g), where a nonmatching primary chain and matching secondary chain return the secondary list; RLI-V05 case (a) rejects preferring secondary over a matching primary. R05a supplies state 105 to that same RLI-R03 operation.
  - runtime_observed: no.

- **PVP-V25** Death and routing guards (PVP-R09a/PVP-R10a; boundary: supplied-input stage)
  - Inputs: Independent cases:skipKillExperience = true; victim experience 0; killer type 3; credited player A with PetType count 8 and supplied hireling H found, x 1000, not the killer, party 3; scoring context absent. Supply exact hireling lookup result, amount and player data.
  - Seeds: no draw.
  - Draws: none.
  - Trace: The first three cases stop before awards. Last requests hireling update 335 then party distribution; scoring no-op.
  - Outputs: First three:no award requests. Last ordered log hirelingLookup(A), amount(H, V)=1000, hirelingUpdate(H, 335), partyDistribution(A, V); no solo/level update or score write.
  - Discriminates: Strict count >7, positive experience and request order; supplied lookup does not verify search.
  - (b) Inputs: two fresh death-experience wrapper runs: (1) killer absent, victim present with experience 1000 and skipKillExperience=false; (2) killer player A present, victim absent. Award routing operations log their requests. Draws: none. Trace: the wrapper's absent-unit guard returns before award routing in either run. Outputs in each run: no award-routing, hireling, solo/party or level-update request. This rejects proceeding with a missing killer or victim.
  - runtime_observed: no.

### 8.2 Observation vectors

**PVP-O01. Player combat and credit cross-checks.** Pending original 1.13c observation. Use matched attacker/defender/gear and recorded raw inputs. Compare player attacks, summon-created attacks, sentry bolts and hireling attacks against player and monster controls; include field/town, directional relation enabled/disabled and party differences; declaration level and timer cross-checks are HOS-O01. Measure the reached scale count and raw field values, guard result bits, actual leech changes separately from scaled inputs, and kill-experience recipients. Expected local assertions are PVP-V01 through PVP-V07/PVP-V08; actual condition/award integration assertions remain their gates. No observation is recorded as matched. A later build cannot confirm a 1.13c rule.

## 9. Acceptance and dependencies

Partial static acceptance comprises the verified assertions of PVP-V01 through PVP-V25 at their named execution boundaries. PVP-V06's actual sentry condition, PVP-V13's stored tier effects and complete dependency-backed awards are excluded until their gates close. Exact stage counts, factor/field arithmetic, guard writes, owner depth, directional relation results, supplied tier/list reads, null-erasing credit selection, request order and local no-draw assertions are available with explicit inputs. Passing those assertions does not establish complete original-game parity.

Full reproduction also requires the in-scope behavior/integration questions settled with source-complete contracts and distinguishing executable vectors, plus matching original 1.13c gameplay observations for every verified behavioral rule (PVP-Q08). Producer questions do not block a supplied-input consumer, and pending observations do not erase static evidence. User approval is separate.

**Depends on:** NRF-R05/R06a for ordered percent arithmetic and wrapping; RLI-R03 for state-list lookup and RLI-R02a/RLI-R02b for selected-list stat reads; GDA-R05/GDA-R08a/GDA-R10/GDF-R02/GDF-R03/GDF-R08/GDA-R16/GDA-R04b for named fields/identities/count ownership; DPO-R08a/R09/R13 for surrounding stage/apply boundaries; DMI-R03 for stage placement; IEP-R07 for event traversal; HIT-R02c/HIT-R02d/SBA-R09a/SBA-R13 for defense differences and contact preconditions; LSE-R02 for leech effectiveness; PXL-R02a/R03/R04/PXL-R05/PXL-R08 for compatible credit/experience assertions at their documented evidence boundaries. Complete implementations of dependencies are not asserted by supplied fixture outcomes. Owners of the cited HIT, GDA, RLI ids: [combat-hit-roll](combat-hit-roll.md) (HIT); [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** damage-pipeline ordering and missile/contact/skill consumers requiring factor selection, hostility or player-credit decisions. Party-experience consumers may reuse the credit resolver, while this specification consumes their amount contracts; these are different operation boundaries. Crushing Blow/Open Wounds share predicates but own their own effects. A supporting citation alone does not create an execution prerequisite.

## 10. Related specs

- [Damage pipeline](combat-damage-pipeline-order.md): build/apply order and resource writes.
- [Mitigation](combat-damage-mitigation.md): resistance/absorb and defender event placement.
- [Shield block and avoidance](combat-shield-block-and-avoidance.md): defense adjustments, hit/block/avoidance.
- [Leech and status effects](combat-leech-and-drain.md): final leech arithmetic.
- [Crushing Blow](combat-crushing-blow.md) and [Open Wounds](combat-open-wounds.md): effect-specific target rules.
- [Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md): supplied sentry attacker on explosion callbacks.
- [Hostility declaration](player-hostility-declaration.md): request that sets the relation eligibility flag read by PVP-R05a.
- [Party experience sharing](player-party-experience-sharing.md): credit/award dependencies and remaining ownership gaps.
- [Timed-state lifecycle](state-timed-state-lifecycle.md): list/propagation integration context.

## Appendix A. Provenance

Evidence build is original Lord of Destruction Expansion 1.13c. D2Game/D2Common original specimens are identified by the [source manifest](../research/source-manifest.md). Compatible static readings retain their 2026-10-11/12 dates and limits. Targeted 2026-10-15 disassembly/original-image checks and available byte-matched reconstructions support the redirect and predicate corrections recorded in the [public companion](../research/spec-evidence/combat-pvp-damage-and-kill-credit.md); declaration evidence is in the [hostility declaration companion](../research/spec-evidence/player-hostility-declaration.md). Shared data identities reuse the dated GDF records; no new table-cell or gameplay observation is claimed. No executable vector is runtime-confirmed here.
