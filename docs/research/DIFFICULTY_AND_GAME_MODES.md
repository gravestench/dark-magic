# Difficulty progression and game-mode research

Status: implementation-oriented research baseline. Dark Magic loads a rich
`DifficultyLevels.txt` model, persists character status/progression facts, and
checkpoints immutable expansion 1.14d `GameRules`. Remaining work is migrating
every relevant authoritative consumer and verifying unresolved formulas.

## Executive conclusion

Difficulty is not `monsterHP *= 2`.

Use a session-level rules context:

```text
GameRules
  difficulty = Normal | Nightmare | Hell
  expansion/classic
  hardcore/softcore
  ladder/season/content-era
  admission capacity (not gameplay scaling)
  game type / realm/offline policy
  content generation fingerprint
```

Subsystems consume this context plus immutable table rows:

```text
combat
monster population/stats/AI
loot/quality/TC
skills/states
hirelings/pets
vendors/gambling
quests/waypoints
player death
special areas/events
```

Do not scatter `if difficulty == Hell` throughout unrelated packages when the rule is data-driven.

## Current Dark Magic DifficultyLevels model

The typed model already exposes a broad modern table surface including:

- resistance penalty, plus non-expansion resistance penalty;
- death experience penalty;
- monster skill bonus;
- monster freeze/cold divisors;
- AI curse divisor;
- life/mana steal divisors;
- unique/champion damage bonuses;
- player-vs-player, player-vs-mercenary and player-vs-Prime-Evil scalars;
- hit-react buffer values;
- mercenary PvP/boss damage scalars and max stun length;
- Prime Evil damage scalars against player/mercenary/pet;
- pet damage versus player;
- monster Corpse Explosion and Fire Enchant explosion scalars;
- Static Field minimum/cap parameter;
- gambling Rare/Set/Unique/Exceptional/Elite odds.

This should remain the canonical typed content view. Some fields are later-data revisions beyond the pinned 1.10f `D2DifficultyLevelsTxt`, so runtime support must be target-version aware rather than assuming every field existed with identical semantics in all LoD patches.

## Pinned 1.10f DifficultyLevels structure

D2MOO's 1.10f `D2DifficultyLevelsTxt` reconstructs at least:

```text
ResistPenalty
DeathExpPenalty
UberCodeOddsNorm / UberCodeOddsGood
MonsterSkillBonus
MonsterFreezeDiv
MonsterColdDiv
AiCurseDiv
UltraCodeOddsNorm / UltraCodeOddsGood
LifeStealDiv
ManaStealDiv
UniqueDmgBonus
ChampionDmgBonus
HireableBossDmgPercent
MonsterCEDmgPercent
StaticFieldMin
GambleRare
GambleSet
GambleUniq
GambleUber
GambleUltra
```

This is strong evidence that item-grade odds, combat, monster control effects, hireling/boss interaction and gambling all consult the same difficulty row.

## Difficulty belongs to game/session state

A character can exist independently of the current game difficulty. The current session chooses the difficulty and validates that the character is eligible to join it.

Keep separate:

```text
CharacterProgression
  completed/unlocked difficulties
  quest/waypoint state per difficulty

GameRules
  selected current difficulty
```

Do not mutate a character's permanent "difficulty" field on every join.

## Difficulty unlock/progression

Classic LoD progression generally gates Nightmare/Hell behind act-completion requirements, but exact eligibility depends on expansion/classic and quest progression.

Server/host game creation should validate:

- character mode/status;
- completed prerequisite difficulty/act quest;
- expansion/classic compatibility;
- hardcore compatibility;
- ladder/realm restrictions if applicable.

The client difficulty button is only a request.

Exact quest flags used for each unlock belong in the probe queue.

## Resistance penalty

Difficulty resistance penalty should be represented as a rules/stat source applied to player effective resistance, not by rewriting base item/character resistance values.

Conceptually:

```text
base + character/item/skill/state sources
+ difficulty resistance penalty source
= effective resistance before caps/other combat rules
```

Expansion versus non-expansion may use different authored values in later tables.

This source changes when entering a game of another difficulty and should disappear cleanly when leaving the session.

## Death experience penalty

Death XP penalty is a player-death policy, not a generic XP gain multiplier.

The player death transaction needs:

- current level and XP within current level;
- current difficulty row;
- softcore/hardcore;
- PvP/other death context if rules differ;
- corpse recovery state.

Do not permanently encode the penalty into the experience curve.

Exact percentage/base and corpse-recovery behavior remain dedicated combat/death probes.

## Monster level and stat scaling

`Levels.txt`, `MonStats`, `MonLvl` and difficulty context interact to produce effective monster stats and population.

Difficulty can affect:

- level/area level;
- life/defense/attack/damage;
- resistances/immunities;
- monster skill levels;
- champion/unique damage bonuses;
- density/unique counts and replacement lists;
- AI/control-effect durations.

Keep the effective monster-materialization policy centralized rather than letting combat independently guess a Hell multiplier.

## Monster skill bonus

The difficulty row exposes a monster skill-level bonus. Monster skill materialization should apply it through the normalized skill definition/learned-level model.

Do not bake the bonus into individual skill handlers.

## Cold/freeze/curse scaling

Difficulty-specific divisors for monster freeze/cold and AI curse behavior are evidence that timed-state duration depends on both effect and target/difficulty context.

The state application pipeline should receive `GameRules` and target kind so duration policy can be centralized/tested.

## Life and mana steal

The combat research already found player leech divided by difficulty-specific `LifeStealDiv` and `ManaStealDiv` in the pinned 1.10f path.

These values should be consumed at the leech stage of combat, not pre-adjust item stat values.

The character sheet can still show raw item life-steal percentage while gameplay applies difficulty effectiveness.

## Unique/champion/boss damage rules

Difficulty rows supply special damage scalars for unique/champion/mercenary/boss interactions in different table eras.

Model these as combat-context rules keyed by attacker/defender classification.

Avoid permanently increasing the spawned monster's base damage if the same bonus is context-specific rather than a universal stat. Probe the original application point before choosing representation.

## Static Field

Static Field's minimum target-life behavior changes by difficulty through `StaticFieldMin` in the table.

The skill handler should query `GameRules.Difficulty` / difficulty row at effect time and apply the verified cap. Do not hard-code separate Normal/Nightmare/Hell skill implementations.

## Gambling

DifficultyRows carry gambling quality/base-grade odds. The gambling item-generation context must use the current game difficulty row together with player level and vendor/offer state.

This is independent of monster TreasureClass Magic Find.

## Item base-grade odds

Pinned 1.10f difficulty data includes exceptional/elite (`Uber`/`Ultra`) odds. Exact consumers vary between generation contexts.

Do not assume these fields are ordinary unique/set quality chances. Preserve semantic names and trace the actual caller before wiring them into vendor/gamble/drop code.

## Difficulty and quests/waypoints

Quest bits and waypoint unlocks are stored per difficulty in legacy character state. The semantic durable model should therefore key these collections by difficulty.

Entering a Nightmare game must not expose Normal quest/waypoint state as if the same bits were global.

Shared invariant:

```text
character identity global
inventory/equipment mostly global
difficulty quest/waypoint progression partitioned by difficulty
```

Exact exceptional cross-difficulty quest-item rules are handled by `QUEST_ITEMS.md`.

## Difficulty and hirelings

Hireling initialization/materialization depends on current difficulty and authored hireling rows. Vendor offerings/archetypes can also differ by difficulty.

The durable mercenary identity remains the same while its current effective stats are materialized under verified rules.

Do not recreate/reroll a durable hireling simply because the player entered another difficulty unless original behavior requires it.

## Difficulty and world generation

Level size/layout, monster lists, preset substitutions, object/shrine eligibility and special areas may differ by difficulty/content mode.

The generated world fingerprint must include the current difficulty and other simulation-affecting game-rule parameters so replay/network peers cannot disagree on world content.

## Classic versus Expansion

Treat Classic/Expansion as an explicit game/content mode independent from difficulty.

It influences systems such as:

- classes/content availability;
- Act V;
- runes/runewords/charms/jewels and expansion item records;
- hirelings and their persistence/equipment;
- exceptional/elite item rules;
- quests/levels;
- save format/item format;
- endgame/special areas;
- some formulas/flags.

Do not express expansion as `difficulty > X` or infer it solely from presence of Act V assets.

A character's expansion/classic status and a game's expansion/classic mode must be compatible.

## Hardcore versus Softcore

Hardcore is a durable character/game-mode constraint, not a difficulty.

It affects:

- death permanence;
- game joining/matchmaking compatibility;
- character select/status;
- persistence commit policy;
- trading/party/realm separation where applicable.

Combat damage itself should generally not branch on Hardcore unless verified; the death transaction does.

## Ladder/season/content era

Ladder is also orthogonal to difficulty.

It can affect:

- item/runeword/Cube eligibility;
- realm/game join compatibility;
- ranking/economy rules;
- later patch special content.

Dark Magic should model it as a content/game-rules flag with a generation/fingerprint, not sprinkle `ladder=true` only in loot code.

Offline mods may choose to enable ladder content; that is policy separate from reproducing an historical realm season.

## Player-count setting

The served game's live authoritative population influences monster scaling,
loot NoDrop, and possibly other systems. A host-authorized `/players X` command
forces the gameplay count independently of that live population.

Keep the admission capacity in immutable `GameRules`, but represent the
optional effective-player-count override in separate mutable, checkpointed game
state. With no override, consumers read live join/leave population. Neither the
admission cap nor a client-supplied drop result is a gameplay multiplier.

Subsystems should receive the effective count they need from one service, not parse chat/UI commands themselves.

## Game type / PvP / arena

D2MOO contains arena/game-type branches in combat and AI. Future game modes may include ordinary PvM, hostile PvP, arena/test modes or mod-defined rules.

Do not overload Difficulty with these. Introduce a separate rules-capability layer if/when needed.

## Game-rules fingerprint

Replay/network compatibility should fingerprint at least:

```text
content generation
rules schema version
difficulty
expansion/classic
hardcore/softcore
ladder/season flags
modded rule parameters
```

Two peers with different difficulty rows or ladder content must not share a deterministic simulation unnoticed.
The mutable player-count override is checkpoint/replay authority state and may
change during a session, so it is not part of the immutable admission
fingerprint.

## Suggested implementation slices

### D1 — immutable GameRules

Add an authoritative immutable game/session rules object and thread it through session/player/world generation without changing formulas yet.

### D2 — difficulty player stat source

Apply resistance penalty from typed difficulty data as a session-only effective stat source.

### D3 — combat difficulty terms

Wire verified life/mana steal divisors, Static Field minimum and one monster control-duration divisor through combat/skill/state tests.

### D4 — progression gating

Represent per-difficulty quest/waypoint state and validate creation/join of a higher-difficulty game.

### D5 — content/game-mode fingerprint

Include difficulty/expansion/hardcore/ladder/player-count policy in replay/network/content fingerprints.

## Verification backlog

1. Exact Normal->Nightmare->Hell unlock quest requirements for Classic/Expansion.
2. Resistance penalties across Classic/Expansion and patch versions.
3. Death XP penalty and corpse-recovery XP rules.
4. MonLvl/MonStats difficulty stat scaling and player-count interaction.
5. MonsterSkillBonus application point and caps.
6. Freeze/Cold/AiCurse divisor exact duration arithmetic.
7. Life/ManaStealDiv exact fixed-point ordering.
8. Unique/Champion damage bonus application point.
9. Mercenary/boss/Prime Evil/PvP scalars by target version.
10. StaticFieldMin units and effect threshold.
11. Gamble Rare/Set/Unique/Uber/Ultra odds and generation context.
12. Exceptional/Elite odds outside gambling and exact callers.
13. Per-difficulty quest/waypoint save mapping.
14. Difficulty-specific quest-item validity.
15. Vendor/hireling offerings by difficulty.
16. Monster lists/density/superunique changes by difficulty.
17. `/players X` monster/loot/XP scaling and bounds.
18. Hardcore join/death/persistence constraints.
19. Classic/Expansion content/formula differences.
20. Ladder/season eligibility and offline policy.
21. Game-type/arena rule branches worth preserving or discarding.

## Primary sources inspected

- Current Dark Magic typed `DifficultyLevels.txt`, progression/persistence/world/loot/combat/hireling/vendor foundations.
- D2MOO pinned 1.10f `D2DifficultyLevelsTxt` plus combat/monster/player/NPC/skill call sites.
