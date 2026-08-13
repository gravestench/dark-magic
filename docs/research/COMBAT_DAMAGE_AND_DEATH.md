# Combat, damage, defense, and death research

> Architecture note: the behavioral findings and deterministic vectors in this
> document remain requirements. Recommendations that `internal/game/combat` or
> another Go package permanently own Diablo policy are superseded by
> [the engine/`d2legacy` boundary](../ARCHITECTURE.md) and M21.14. Migration must
> preserve one authority and one stat/item model while moving D2 policy into
> authoritative Lua; it must not create a parallel simulation.

Status: implementation-oriented research baseline. The strongest runtime evidence in this document is D2MOO's reconstruction of Diablo II 1.10f. It is not yet a Dark Magic compatibility contract. Exact arithmetic and patch-sensitive behavior must graduate through owned-game traces or other independent verification before becoming golden compatibility requirements.

This document extends [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md), [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md), and the current authoritative session/ECS architecture. It does not propose a parallel combat world outside `internal/game/session` and Akara.

## Executive conclusion

Dark Magic should model combat as a deterministic **transaction pipeline**, not as a single `damage = attack - defense` formula and not as presentation-driven animation callbacks.

A useful conceptual flow is:

```text
admitted action / missile contact
        |
        v
attacker + defender + source snapshot
        |
        v
hit eligibility / target legality
        |
        v
attack resolution
  miss / hit / block / avoid / critical family
        |
        v
raw typed damage bundle
  physical / fire / lightning / cold / magic / poison / drain / durations
        |
        v
source modifiers + conversion
        |
        v
defender mitigation
  resist / max resist / pierce / absorb / reductions / special immunities
        |
        v
final damage + secondary effects
        |
        +--> HP/mana/stamina mutation
        +--> leech/heal
        +--> cold/freeze/poison/stun/burn states
        +--> knockback/hit-recovery intent
        +--> durability / proc / unit events
        `--> death transaction when HP reaches zero
```

The output should be a replayable result record or equivalent deterministic mutations. Presentation consumes semantic events after authority decides them.

## Current Dark Magic boundary

Dark Magic already has the pieces a combat subsystem should plug into:

- `internal/game/session` owns ordered fixed-tick command admission, replay, checkpoints, and trusted mutation;
- `internal/game/player` materializes live player stats and skill knowledge into ECS;
- `internal/game/item` owns equipment/container identity and transaction state;
- `internal/game/loot` already evaluates a substantial portion of Diablo item-property semantics;
- `internal/game/targeting` provides stable semantic target kinds;
- `internal/game/world` owns collision and navigation facts;
- the typed game-data catalog already loads `ItemStatCost`, `Properties`, `Skills`, `Missiles`, `States`, `MonStats`, `MonLvl`, `DifficultyLevels`, and related tables.

The next combat work must preserve these owners' tested invariants while M21.14
migrates D2 policy into `d2legacy`. Do not create a second stat store, a second
item/equipment owner, or a parallel combat simulation; replace each
transitional Go owner through a complete vertical migration.

## D2MOO 1.10f damage pipeline evidence

D2MOO's `SUnitDmg.cpp` is high-value evidence because it sits at the intersection of stats, skills, items, monsters, missiles, states, events, quests, pets, and difficulty rules.

### Fixed-point values are part of the behavior

The runtime repeatedly shifts damage and life-like quantities by eight bits. Examples include weapon min/max damage and hit points. Later application paths treat values below `256` as zero. This is strong evidence that many combat quantities use 8-bit fractional precision internally.

Dark Magic should not casually flatten all combat math to `float64`. Prefer explicitly named integer/fixed-point types or conversion helpers so rounding boundaries remain visible and testable.

### Raw physical damage is not just weapon min/max

`SUNITDMG_ApplyDamageBonuses` shows a layered physical calculation:

- primary or secondary weapon damage can be selected;
- unarmed/default minimums are synthesized when needed;
- `item_normaldamage` contributes to both ends of the range;
- generic enhanced damage contributes;
- item strength and dexterity bonus coefficients contribute according to the equipped weapon;
- weapon mastery contributes;
- min- and max-specific percentage stats may diverge;
- the resulting range is rolled from the **attacker unit RNG stream**;
- source-damage scaling uses `128` as the full-value reference.

This strongly supports the layered-stat model proposed in `ITEM_STATS_AND_AFFIXES.md`: combat needs both source provenance and parameterized stats, not only final character-sheet totals.

### Target type changes physical damage

The 1.10f path applies target-sensitive bonuses before final physical resolution. Evidence includes:

- damage versus demons;
- damage versus undead;
- a built-in blunt-weapon bonus against undead;
- damage versus nested monster types.

Monster taxonomy therefore participates in combat semantics. `MonType` relationships are not merely UI grouping.

### Critical Strike and Deadly Strike are a family, not additive damage stats

The observed path checks weapon mastery critical chance, then passive critical strike, then item deadly strike. A successful roll doubles physical damage and marks a critical-style result flag.

Important implementation consequence: do not model all three as simultaneously additive percentages. Preserve their ordering/exclusivity until a broader trace proves the exact rule for every source.

### Damage conversion happens before final typed mitigation

The runtime can convert a percentage of physical damage to a selected element. Random conversion uses the attacker RNG. Cold/poison/freeze conversions also establish minimum effect durations.

Dark Magic therefore needs an intermediate typed `DamageBundle` or equivalent. Converting only after reducing all damage to one scalar loses behavior.

## Typed mitigation table

`SUNITDMG_CalculateTotalDamage` contains a table-driven mapping between damage fields and defender stats. The recovered rows include at least:

| Channel/effect | Defender stat family observed |
| --- | --- |
| physical | physical/damage resistance |
| fire | fire resist, max fire resist, fire pierce, percent/flat fire absorb |
| lightning | lightning resist, max lightning resist, lightning pierce, percent/flat lightning absorb |
| cold | cold resist, max cold resist, cold pierce, percent/flat cold absorb |
| magic | magic resist, max magic resist, percent/flat magic absorb |
| cold duration | cold resistance family |
| freeze duration | cold resistance family |
| poison duration | poison-length resistance family |

This is evidence for a **table/data-driven mitigation stage**. It is not evidence that Dark Magic should reproduce the exact C struct layout.

### Resist and absorb ordering needs explicit probes

The source establishes that resistance, max resistance, pierce, reductions, and absorbs are distinct phases/inputs. Exact caps, negative-resistance behavior, PvP differences, and the order of flat versus percentage absorb must be recorded with executable vectors before implementation is declared compatible.

## Drains and leech

The observed application path distinguishes:

- life drain/leech;
- mana drain/leech;
- stamina drain;
- monster `Drain` effectiveness by difficulty;
- player difficulty divisors for life and mana steal;
- unit-type differences, including hirelings;
- healing capped by missing life;
- defender mana/stamina capped at available values.

Leech therefore belongs after successful damage resolution with access to:

```text
attacker type
source type
defender type
defender drain effectiveness
difficulty
actual physical damage dealt
attacker missing resources
```

It should not be implemented as a generic item proc detached from final damage.

## Timed secondary effects

Damage can carry effect magnitude and duration separately.

Observed examples include:

- cold duration;
- freeze duration;
- poison damage plus poison duration;
- stun duration;
- burn fields, although D2MOO marks parts of burn calculation as unresolved/TODO;
- knockback and hit-recovery related result flags elsewhere in the damage path.

The existing `States.txt` catalog should become the durable vocabulary for timed unit state, but combat should emit semantic state-application requests rather than directly manipulating presentation modes.

### Stun evidence

The 1.10f path:

- caps some stun durations;
- applies different restrictions to monsters;
- prevents ordinary boss stun through explicit conditions;
- gives hirelings a special cap in the observed path;
- stores expiration as a future game-frame event;
- reschedules an existing stun state rather than always allocating a duplicate.

This is useful evidence for future state stacking/refresh semantics.

## Unit events are part of combat

Successful hit resolution triggers semantic unit events. The recovered path distinguishes melee and missile events and has attacker-side and defender-side event families.

This matters because Diablo item/skill behavior frequently hangs secondary effects off "on attack", "on striking", "when struck", kill, and related events.

Recommended Dark Magic boundary:

```text
CombatResolution
  -> deterministic CombatEvent stream
       ATTACK_ATTEMPT
       HIT
       MELEE_DAMAGE
       MISSILE_DAMAGE
       DAMAGED
       KILL
       DEATH
       ...
  -> registered trusted effect handlers
```

Do not have Lua scan ECS after the fact to infer that a hit occurred.

## Hit chance, block, dodge, and avoidance

D2MOO's pinned 1.10f `SUNITDMG_IsHitSuccessful` now provides the exact final
rating/level arithmetic and roll boundary. Dark Magic preserves it in
`combat.LegacyHitChance` and executable integer vectors:

```text
ratingFactor = 100 * attackRating / (attackRating + defense)
chance = 2 * attackerLevel * ratingFactor / (attackerLevel + defenderLevel)
chance = clamp(chance, 5, 95)
hit = random % 100 < chance
```

These are ordered integer operations: rating division truncates before the
level factor. Negative defense transfers its magnitude into attack rating;
negative attack rating transfers its magnitude into defense, then both are
floored at zero. When the combined divisor is zero, the recovered function
retains a rating factor of 100.

This verifies the final resolver, not every upstream stat contribution.

Pinned D2MOO commit `3b21043b99e987bad41cf0f7b49f1f246db52d5c`
also supplies the upstream player inputs used by the production d2legacy
resolver:

- `UNITS_GetAttackRate` in `D2Common/src/Units/Units.cpp` computes
  `STAT_TOHIT + 5 * (Dexterity - 7) + CharStats.ToHitFactor`;
- `UNITS_GetDefense` computes `Dexterity / 4 + STAT_ARMORCLASS`, then applies
  the combined item/skill armor percentage with integer truncation;
- `SUNITDMG_IsHitSuccessful` adds percentage attack-rating sources before the
  final rating/level calculation above;
- `sub_6FC7C7B0` and `D2Common_10434` make hand selection a `Skills.txt`
  `weapsel` decision and restrict player dual wielding to Barbarian and
  Assassin, rather than alternating merely because both hand slots are full.

Dark Magic projects those inputs as named flat or percentage sources, resolves
them in a deterministic flat-then-percentage phase, and records the resolved
attack rating, defense, chance, and selected hand on every melee event.

Before implementing a compatibility claim, capture:

- attacker attack rating;
- defender defense;
- attacker and defender levels;
- class/Dexterity base attack rating and flat/percent modifiers;
- block chance and animation/state prerequisites;
- Amazon dodge/avoid/evade and similar skill-specific avoidance ordering;
- effects that ignore target defense or always hit;
- melee versus missile distinctions;
- PvP differences.

Treat familiar community formulas as leads until pinned against the targeted runtime/data revision.

## Hit recovery and control effects

Do not make hit recovery a UI-only animation decision. It is a combat/simulation state that can change whether the defender may act.

Research still required:

- exact damage threshold(s) and unit-specific rules that cause hit recovery;
- Faster Hit Recovery interaction and frame breakpoints;
- knockback distance/path collision and failure cases;
- chill/freeze velocity effects;
- stun interaction with monster AI scheduling;
- interruption rules for active skills.

The animation system should receive an authoritative mode/event after these rules decide the result.

## Death is a transaction, not `hp <= 0` alone

The 1.10f damage path sets a `WILLDIE` result after applying damage and timed effects, then triggers kill/death event handling. Other systems participate in player/monster death.

Dark Magic should separate:

```text
lethal damage detection
        |
        v
death eligibility / special prevention
        |
        v
DeathTransaction
  killer / assists / party
  unit type
  hardcore/softcore
  current difficulty
  corpse policy
  gold/experience policy
  quest credit
  loot credit
  owned-unit consequences
        |
        +--> authoritative state mutation
        +--> spawned corpse / dropped gold / loot
        +--> quest/XP events
        `--> presentation event
```

### Monster death

At minimum future monster death needs to coordinate:

- final killer / last attacker;
- party and summon ownership credit;
- XP award;
- loot seed/drop event;
- quest kill notification;
- corpse creation and corpse-usability flags;
- death effects and unique modifiers;
- resurrection restrictions;
- removal/inactivation timing.

### Softcore player death

Research separately:

- corpse equipment transfer;
- carried-gold loss and dropped-gold rules;
- experience penalty by difficulty;
- corpse recovery and experience recovery behavior;
- town respawn position;
- multiple corpse edge cases;
- save persistence boundaries.

Do not implement this by deleting and recreating the selected `persistence.Character`. The durable identity remains the same character while live session state transitions through dead/corpse/recovered states.

### Hardcore death

Hardcore permanence is a **durable character transition**, not simply a different death animation. The authoritative session should produce a persistence event that permanently records the death only after the death transaction commits.

Reconnect/crash semantics around that commit need explicit design and tests.

## PvP and hostility

The generic combat pipeline should not bake `hostile = unit type is player` into hit resolution. Player-versus-player legality, party membership, town restrictions, and hostility are separate authority inputs.

PvP damage scaling, leech differences, minion/trap attribution, and death penalties must be researched before multiplayer compatibility is claimed.

## Recommended Dark Magic architecture

A possible package shape, subject to repository review:

```text
internal/game/combat
    types.go        typed fixed-point damage/result values
    resolve.go      hit + raw damage + mitigation pipeline
    events.go       semantic combat event records
    death.go        unit-type-specific death transaction orchestration
    tests/...       synthetic vectors
```

The package should consume narrow interfaces/snapshots from:

- game data;
- stats/effects;
- item equipment;
- world/targeting;
- session RNG/state;

and mutate only through trusted session/ECS boundaries.

Do not let `combat` import renderer, Lua, or platform packages.

## Suggested implementation slices

### C1 — fixed-point combat vocabulary

Add explicit damage channels, fixed-point helpers, result flags, and deterministic test vectors without changing live attacks.

Acceptance:

- integer/fixed-point rounding is explicit;
- damage bundles preserve all supported channels independently;
- marshal/checkpoint form is deterministic.

### C2 — basic melee transaction

Implement one ordinary player/monster melee path:

```text
legal target -> hit/miss -> physical damage -> HP -> hit/death event
```

Keep criticals, elemental damage, states, leech, and advanced modifiers deferred but structurally supported.

### C3 — typed mitigation

Add elemental/magic/physical resistance and verified cap/pierce ordering using table-driven vectors.

### C4 — secondary damage effects

Add critical/deadly strike, cold/freeze, poison, stun, leech, absorb, and damage conversion in independently testable layers.

### C5 — death transaction

Implement monster death first, then softcore player death, then hardcore durable transition. Every step must be replay/checkpoint-safe.

## Verification plan

Synthetic tests should record every intermediate stage, not merely final HP.

Example golden record:

```text
attacker stats hash
source identity
defender stats hash
RNG before
hit result
raw bundle
conversion result
mitigation result
absorbed/healed
state applications
final resource deltas
RNG after
events emitted
```

Owned-game probes should cover boundary values such as:

- resistance around caps and negative values;
- min/max damage one unit apart;
- exactly lethal versus one fixed-point unit surviving;
- critical and deadly strike ordering;
- zero/immune drain;
- poison/cold duration rounding;
- blocking and dodge families;
- PvP scaling.

## Open verification backlog

1. Exact 1.10f chance-to-hit formula, clamps, and integer rounding.
2. Exact block ordering, cap, movement penalty, and animation requirements.
3. Dodge/Avoid/Evade and other avoidance ordering versus block/hit.
4. Full resistance/max-resistance/pierce/absorb order with negative values.
5. Physical damage reduction percentage/flat ordering.
6. Crushing blow rules by attacker/defender type and difficulty/PvP.
7. Open Wounds duration/damage and refresh behavior.
8. Hit-recovery trigger and FHR timing.
9. Knockback path/collision semantics.
10. Poison source combination and refresh/overwrite behavior.
11. Freeze/chill immunity, duration, and boss/champion modifiers.
12. Thorns/reflection and attribution ordering.
13. PvP damage, leech, minion/trap ownership, and hostility rules.
14. Softcore corpse, gold, XP-loss, and corpse-recovery details.
15. Hardcore commit/persistence behavior across disconnect/reconnect.
16. Durability loss and ethereal/indestructible rules on hit/death.
17. Exact kill/assist/party credit for XP, quests, loot, and summons.

## Primary sources inspected

- D2MOO pinned 1.10f `source/D2Game/src/UNIT/SUnitDmg.cpp` and `include/UNIT/SUnitDmg.h`.
- D2MOO `D2StatList`, `D2Items`, `DifficultyLevels`, monster and skill table usage reachable from the damage path.
- Current Dark Magic `internal/game/session`, `internal/game/item`, `internal/game/player`, targeting/world code, and foundation research docs.

Source code from D2MOO is evidence. Dark Magic implementation should restate the behavior independently and should promote delicate arithmetic only after executable vectors or owned-runtime traces exist.
