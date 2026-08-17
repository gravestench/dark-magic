# Character creation, stats, levels, and progression

Status: implementation-oriented research baseline. Dark Magic already has character creation UI, typed `CharStats`/`Experience` data, system-authority player admission, live level/experience/vitals, and learned skill entities. Exact per-class numeric vectors must be generated from lawfully mounted game data and original behavior before this becomes a fidelity contract.

## Executive result

Character progression should be driven by immutable class/progression data plus durable base state:

```text
CharStats + Experience + Skills + quest rewards
                 |
                 v
DurableCharacter base progression
  level / XP
  base attributes
  unspent stat/skill points
  learned skill base levels
  permanent rewards
                 |
                 v
authoritative session player components
                 |
                 v
derived stats through the shared stat resolver
```

Do not persist a giant snapshot of every derived combat stat as canonical character state. Save base/permanent facts and reconstruct derived values from the pinned content generation.

## Current Dark Magic baseline

`internal/game/data/model.CharStats` currently models starting Strength, Dexterity, Intelligence/Energy, Vitality, stamina, HP add, mana regeneration, to-hit factor, walk/run velocities, run drain, Life/Stamina/Mana per level, Life/Stamina per Vitality, Mana per Magic/Energy, stat points per level, class animation terms, block factor, starting skill, class skill identifiers, base weapon class, and up to ten starting items with locations/counts.

`ExperienceData` contains class-specific cumulative experience columns plus `ExpRatio`.

`internal/mod/d2legacy/adapter/player.Entry` currently admits level, experience,
Vitality, current/max life/mana/stamina, class/identity, appearance, and world
location. Admitted Vitality is retained in checkpointed
`d2legacy.player.stamina_progression`; the live model still lacks the complete
four-attribute/unspent-point/permanent-reward and skill-progression contracts.

## Class definition

Build normalized immutable data from `CharStats.txt`:

```text
ClassDefinition
  stable class ID/token
  starting attributes
  starting life/mana/stamina terms
  per-level gains
  per-vitality / per-energy gains
  stat points per level
  walk/run/drain terms
  block factor
  start skill
  class skill namespace/list
  base weapon class
  starting item recipes
```

Do not hard-code seven Go switch blocks for values already authored in data. Stable class token/legacy numeric ID mappings still need explicit save/network/composite adapters.

The source model calls the authored `int` column `Intelligence`; player-facing terminology should use Energy where appropriate while retaining raw mapping.

## Per-class formulas

The class data exposes the symbolic inputs. Maximum stamina now has an
implemented fixed-point interpretation pinned by the owned Expansion 1.14d
record graph and recovered executable structure; life/mana and remaining
progression transactions still need equivalent vectors.

Research these relationships:

```text
creation max life
  = class starting term + vitality contribution + HPAdd semantics
life gain on level
  = LifePerLevel using legacy fixed-point scale
life gain from vitality
  = LifePerVitality using legacy fixed-point scale
mana gain on level
  = ManaPerLevel
mana gain from energy
  = ManaPerMagic
maximum stamina raw (implemented)
  = starting stamina << 8
  + (level - 1) * StaminaPerLevel << 6
  + (base Vitality - class starting Vitality) * StaminaPerVitality << 6
  + direct maxstamina and ItemStatCost op contributions
```

Do not divide/round until the exact scale is proven. Owned-data tests should generate non-distributable vectors for all seven classes: creation, +1 Vitality, +1 Energy, one level with no allocation, and one level plus allocated Vitality/Energy.

## Durable progression state

```text
CharacterProgression
  class
  level
  experience
  base strength
  base dexterity
  base vitality
  base energy
  unspent stat points
  unspent skill points
  learned skill base levels
  permanent reward modifiers
  respec availability/usage
```

Current health/mana/stamina are live resource values. Maximums are derived from progression plus active stat sources. A legacy save may store redundant current/max fields; importer preserves bytes and validates canonical relationships.

## Experience thresholds

Treat `Experience.txt` as the authored threshold table. Determine whether rows are cumulative thresholds for entering levels and how the cap is represented in target patches.

A level-up transaction:

1. authoritative XP increases;
2. compare against next class threshold;
3. advance across thresholds as original rules allow;
4. grant class-authored stat points and skill-point award;
5. update base per-level resource terms;
6. emit semantic level-up event;
7. invalidate/recompute derived stats;
8. mark durable progression changed.

UI never writes the new level/points.

## Experience award policy

Separate threshold logic from XP award formulas. Exact research is still needed for monster/player level penalties, party division/range eligibility, high-level penalties, difficulty, player count, shrines, mercenary XP, quest/Ancients awards, and death penalties/recovery.

## Stat-point allocation

An authoritative command requests increments, not a final value:

```text
player.allocate_stat(stat, count)
```

Validate positive bounded count, sufficient unspent points, valid base attribute. Mutation atomically decrements unspent points and increments base attribute; derived values update through the stat resolver. A client cannot submit `strength = 400`.

## Skills

Separate purchased/base learned level from item bonuses, oskills, charges, and temporary effects. Durable progression stores base class skill levels. Skill allocation validates unspent points, tree/class membership, character-level requirement, prerequisites, and max base level.

Current `d2legacy.player.learned_skill` entities are a useful live shape but should eventually distinguish base purchased level from effective/display level.

## Starting skills and equipment

`CharStats` contains start skill and up to ten starting item recipes. Creation should use these records rather than hard-coded UI loadouts.

Creation transaction:

1. validate name/mode/class;
2. pin content generation;
3. load class definition;
4. construct base attributes/resources/level/XP/points;
5. create starting learned skills;
6. mint starting items from authored recipes;
7. place through item authority;
8. create durable character atomically;
9. select/admit only after success.

Missing required item/skill causes failure with no partial character.

## Name validation

Current Dark Magic validates 2-15 ASCII letters with limited internal punctuation. Independent save documentation corroborates 16-byte storage and 2-15 visible length but differs on punctuation and does not establish all realm rules.

Separate canonical Dark Magic naming, `.d2s` encodability, original offline rules, legacy realm rules, and modern realm uniqueness/reserved-name policy.

## Appearance

Durable state should store semantic class/equipment/cosmetic inputs, not only resolved DCC paths. COF/mode/weapon class/component paths are derived presentation resources. This aligns with the composite/creature asset roadmap.

## Permanent quest rewards

Permanent growth belongs in durable quest/progression state, not unexplained max-stat edits. Model stable reward-consumption identity, for example:

```text
quest-reward:<difficulty>:<quest>:<reward>
```

Quest durable bits prove consumption; the stat resolver can expose the resulting permanent source. Reconnect/dialogue retries cannot grant twice.

## Respecs

Research Akara reward, token respec, patch applicability, exactly which stats/skills reset, and permanent quest effects. Respec must compute a complete replacement progression state and commit atomically; do not remove skills first and risk failure halfway through.

## Movement and stamina

`CharStats` walk/run velocity, run drain, stamina drain/recovery, and the core
level/Vitality/source-derived maximum are implemented. Movement owns integration;
progression owns the base class/resource terms. Remaining work here is the
environment-period `item_stamina_bytime` operand, live base-Vitality allocation
ordering, and verified armor/shield/chill/freeze modifier ordering.

## Blocking and breakpoints

`BlockFactor` is a class input, but final block depends on shield/item stats, Dexterity, level, skills, and caps. Animation breakpoints cross class parameters, equipment weapon classes, skills/states, and speed stats. Progression supplies base inputs; combat/animation research owns final formulas.

## Difficulty, titles, hardcore, Ladder

Keep durable hardcore flag, edition, Ladder/season identity, highest completed difficulty/act, current/last difficulty, and derived title/progression display distinct. `.d2s` summary fields do not replace deeper quest/act/realm state. Hardcore death is a persistence/game-mode transition, not ordinary current HP.

## Ownership boundary

| Concern | Owner |
| --- | --- |
| creation form/preview | Lua presentation |
| name/class/mode validation | authoritative creation repository/service |
| class definitions/XP thresholds | pinned typed data generation |
| XP awards | combat/party/quest authority |
| level/stat/skill point mutation | player progression authority |
| effective stats | stat resolver |
| durable progression | persistence repository |
| title/progression display | derived presentation |
| uniqueness/hardcore/ladder policy | realm service |

## Failure/idempotency

- Character creation is atomic.
- Repeated creation transaction cannot mint duplicate starting items.
- Level-up awards occur once per crossed threshold.
- Stat/skill allocation cannot overspend under concurrent commands.
- Quest permanent rewards have stable consumption identity.
- Respec replaces progression atomically.
- Save/reload cannot convert item bonuses into permanent base attributes.
- Content-generation mismatch is diagnosed before reconstructing progression with different formulas.

## Implementation slices

1. normalized class definition from CharStats;
2. base-attribute/unspent-point/base-skill progression components;
3. immutable XP threshold view + authoritative level-up transaction;
4. stat allocation command;
5. skill allocation command/prerequisites;
6. authored starting loadout through item authority;
7. exact fixed-point life/mana vectors and remaining stamina time/allocation vectors;
8. respec after quest/item dependencies;
9. difficulty/title/hardcore integration.

## Acceptance criteria

- All seven classes load from catalog with no hard-coded numeric starts.
- Creation uses owned-data starting attributes/skills/items.
- XP one below/at/above thresholds is exact.
- Multi-threshold grant cannot skip/duplicate awards.
- Stat allocation is deterministic and cannot overspend.
- Skill allocation rejects missing requirements.
- Base vs item/quest-derived stats remain distinguishable after save/reload.
- Permanent quest rewards are idempotent.
- Respec cannot duplicate/delete points on failure.
- Renderer FPS cannot affect progression.
- Hardcore/edition/difficulty remain durable and separate from current life.

## Verification backlog

1. Extract all seven CharStats rows into non-distributable test vectors and document scaling.
2. Verify creation max life/mana and `HPAdd` semantics; maximum stamina core is pinned.
3. Verify life/mana per-level and per-Vitality/Energy rounding plus stamina
   base-Vitality allocation callback ordering.
4. Verify Experience row semantics and level cap.
5. Trace one XP grant crossing multiple levels.
6. Trace party/level/difficulty XP penalties and integer rounding.
7. Trace shrine/Ancients/death/player-count effects.
8. Verify points awarded per level and patch differences.
9. Trace skill prerequisites/max base level/respec reconstruction.
10. Verify all starting item/location/count and start skill records.
11. Establish exact offline/realm name rules.
12. Trace difficulty/title/Classic->Expansion/hardcore/Ladder transitions.

## Sources

- Dark Magic `internal/game/data/model/charstats.go`, `experience.go`, `internal/game/data/catalog`, and `internal/game/player`.
- [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md).
- [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md).
- [QUEST_RUNTIME_MODEL.md](QUEST_RUNTIME_MODEL.md).
- Bundled Data File Guide sections for `CharStats.txt`, `Experience.txt`, and `Skills.txt`.
- [nokka/d2s README](https://github.com/nokka/d2s/blob/master/README.md) as independent save evidence, subordinate to owned-save probes.
