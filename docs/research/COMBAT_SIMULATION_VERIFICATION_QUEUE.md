# Combat/simulation verification queue

This queue consolidates the highest-value empirical probes from combat, skills/states/missiles, monsters/AI, owned units, and movement/streaming research.

The point is to turn uncertain reverse-engineering claims into **small executable evidence PRs** rather than allow plausible community formulas to become undocumented engine constants.

## P0: architecture-shaping probes

- [ ] Revalidate the recovered chance-to-hit rating/level arithmetic, 5/95 clamps,
  modulo comparison, negative-rating normalization, and integer truncation
  order from `SUNITDMG_IsHitSuccessful`; executable vectors live in
  `internal/game/combat/hit_chance_test.go`. Those 1.10f-derived vectors are
  secondary evidence until confirmed against expansion 1.14d.
- [x] Pin and project the M21.13 player attack-rating and defense inputs:
  CharStats class factor, Dexterity bases, equipped weapon/armor, named
  affix/socket/passive flat sources, combined percentage sources, local
  enhanced defense, and skill-record hand selection. Broader block,
  ignore-defense, target-AC, mastery, aura, and PvP branches remain separately
  enumerated below instead of being hidden in Basic Attack.
- Pin one complete physical-damage transaction including fixed-point scale, min/max roll boundaries, resistance/reduction, exact lethal threshold, and RNG before/after.
- Enumerate the `Skills.txt` server start/do behavior IDs present in mounted LoD data and identify which shared behavior families are required for an initial playable Act I slice.
- Build a headless state-instance experiment proving refresh/expiration semantics for at least stun, chill, poison, and one aura/curse source.
- Capture one straight missile's spawn position, fixed-point velocity, lifetime/range, collision, damage snapshot, and removal boundary.
- Extract one ordinary Blood Moor monster from MonStats/MonLvl into a complete expected effective-stat vector for each difficulty.
- Trace one monster's AI-think cadence from activation through target acquisition and attack, including RNG consumption.
- [x] Pin all seven Expansion 1.14d `CharStats.txt` walk/run and stamina fields,
  implement shared 25 Hz 8.8 stamina/FRW policy, and cover authority/prediction,
  exhaustion, town recovery, item FRW, and equipment-source lifecycles. The
  level/Vitality/direct/skill/item-per-level maximum graph and proportional
  source transition are also implemented and target-record pinned. Time-of-day,
  live base-Vitality allocation, and owned-runtime cold/freeze classification,
  immunity, duration, and action-rate boundaries remain in the movement queue.
- Compare Dark Magic A* output with original path behavior for a small blocked room and identify where the original selects A*, IDA*, direct, or other path types.
- Define canonical checkpoint state for one active monster, one missile, one timed state, one pending cast, and one owned unit.

## P1: combat arithmetic

- [ ] Block chance, cap, movement penalty, and ordering relative to hit/
  avoidance. The Expansion-1.14d-only owned-runtime analyzer and matched-control
  template now live in `internal/dev/tools/defense_outcome_probe` and
  `docs/research/probes/defense-outcome-lod-114d-expansion.template.json`.
  Populate the matrix before promoting arithmetic.
- [ ] Dodge/Avoid/Evade and similar avoidance order. The same contract records
  exact effect-record identity, displayed chance, defender action state, visual
  outcome/reaction, and raw health deltas; it rejects Classic, earlier patches,
  servers, saves, community tools, and mismatched controls.
- Physical resistance/percentage reduction/flat reduction ordering.
- Fire/lightning/cold/magic resistance, max resistance, pierce, percentage absorb, and flat absorb ordering at negative/cap boundaries.
- Critical Strike, weapon mastery critical, and Deadly Strike ordering/exclusivity.
- Crushing Blow scaling by target type, difficulty, ranged/melee, and PvP.
- Open Wounds damage/duration/refresh.
- Hit recovery trigger and Faster Hit Recovery timing.
- [ ] Knockback chance, target categories/mode fallback, distance, collision,
  and failure semantics. The Expansion-1.14d-only owned-runtime capture
  contract and normalized older-source hypotheses now live in
  `internal/dev/tools/knockback_probe` and
  `docs/research/probes/knockback-lod-114d-expansion.template.json`; no gameplay
  constant is promoted until that matrix is populated.
- Life/mana leech difficulty divisors and monster drain effectiveness.
- Damage-to-mana and reflection/thorns attribution/order.
- Poison source combination, duration, overwrite, and per-tick rounding.
- Chill/freeze duration, immunity, velocity effects, boss/champion behavior.
- Durability loss on attack/hit/death.

## P1: skills and states

- Mana-cost formula and exact charge tick.
- Skill delay/cooldown groups and first reusable cast-timing vectors.
- Cast-speed relationship to action/effect tick and animation presentation.
- Interruption by hit recovery, block, knockback, stun, movement, and death.
- Target-policy flags for self/unit/point/corpse/item/summon/movement skills.
- Skill calculation expression semantics, overflow, and rounding.
- Hard-point versus soft-point synergy behavior.
- Weapon contribution and alternate weapon-set behavior.
- Aura pulse cadence, range metric, stacking, and source removal.
- Curse overwrite/priority behavior.
- Passive skill source lifetime.
- State persistence across death, level transition, shapeshift, dispel, save/exit.
- Corpse skill eligibility and consumption ordering.

## P1: missiles

- Velocity/range/lifetime stepping in fixed-point.
- `CollideType` -> movement/contact collision mask mapping.
- Pierce finite count, hit memory, and repeated-target prevention.
- Splash/explosion radius and candidate ordering.
- Guided/homing target loss/reacquisition.
- Bounce/chain/child-missile creation ordering.
- Source stat snapshot versus dynamic owner lookups by missile family.
- Trap/sentry ownership and proc/kill attribution.
- Client-sent versus server-derived missile fields for future networking.

## P1: monsters and AI

- MonStats/MonLvl effective stat construction by difficulty and flags.
- Level population density, group size, rarity, and room eligibility.
- Monster seed derivation and stable RNG stream usage.
- Champion/unique probabilities and modifier selection.
- Superunique/minion initialization and quest substitutions.
- AI ID -> behavior family and AI parameter interpretation.
- Think delay and rescheduling under chill/freeze/stun.
- Target acquisition, LOS, retention, fear/flee/taunt/confuse/attract behavior.
- Pack leader/minion command propagation.
- Door handling and special path selection.
- Corpse eligibility/lifetime and resurrection restrictions.
- XP/loot/quest credit for player, pet, hireling, trap, party, and environmental kills.

## P1: movement and streaming

- Exact A*/IDA* neighbor/tie-break/path-length behavior.
- Direct/toward/straight/wall-follow/circle path strategies.
- Unit collision-pattern mapping versus current radius approximation.
- Dynamic unit occupancy and simultaneous movement conflict ordering.
- Door collision changes and path replanning.
- Charge, Leap, Whirlwind and Teleport movement rules.
- Missile-specific path families including Charged Bolt and Blessed Hammer.
- Original active-room set/radius and room lifecycle sequence.
- Inactive monster archive fields and restoration order.
- Long-inactive HP restoration threshold/behavior across patches.
- Inactive corpse/item/object behavior.

## P1: hirelings and owned units

- Hire/reward unlock rules by act and difficulty.
- Initial hireling level/name/seed generation.
- All `Hireling.txt` fixed-point/per-level rounding vectors.
- Hireling XP thresholds and level limits.
- Act II aura variant rules and Act III skill selection.
- Equipment restrictions and replacement behavior.
- Resurrection cost and gold source.
- Death persistence through save/exit/rejoin.
- Follow/leash/teleport and town/zone-transition behavior.
- PetType maximum modification and deterministic excess removal.
- Summon replacement semantics by pet category.
- Revive source-monster inheritance and duration.
- Owner-attributed PvP/quest/XP/loot rules.

## P2: player death and PvP

- Softcore corpse equipment transfer.
- Carried/stashed gold loss/drop rules.
- Nightmare/Hell XP loss and corpse-recovery XP restoration.
- Multiple-corpse edge cases and save/exit behavior.
- Hardcore durable death commit point and disconnect semantics.
- Hostility declaration/cooldown/town restrictions.
- PvP damage scaling, leech, minion/trap attribution, and party restrictions.

## Probe artifact standard

A completed probe should preferably produce:

```text
probe ID / title
target version/source (expansion 1.14d)
owned input requirements
reproduction steps
raw captured values
normalized vector(s)
expected deterministic result
conflicts with older secondary sources, if relevant
confidence upgrade
which research document/section changed
```

When feasible, retain synthetic/normalized vectors in the repository and keep proprietary captures outside Git. For broad algorithm work, reserve at least one blind holdout case before implementation, following the verification discipline documented by libd2legacy.
