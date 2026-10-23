# Issue #373: prove session difficulty reaches generated monsters

## Intent and boundary

Advance the remaining difficulty acceptance of [issue #373](https://github.com/gravestench/dark-magic/issues/373). A game started in Normal, Nightmare, or Hell should generate a monster from that difficulty's authored MonStats columns and preserve its selected loot class at spawn. The player-count life and XP bonus must compose with those difficulty-specific bases. This is a production-path verification slice: if the existing implementation passes, only executable coverage and an honest issue update are needed; a failed assertion justifies a correction within this boundary.

## Evidence and limits

`docs/specs/monster-scaling-by-players-and-difficulty.md` MSD-R06 says the count bonus itself has no difficulty input; MSD-R08/R10/R13 select monster level, life, armor, and XP from difficulty-indexed data in original 1.13c static code. `internal/content/d2legacy/lua/d2legacy/data/monster.lua` already reads `game_rules.difficulty()`, selects MonStats and MonLvl columns, and selects `TreasureClass1`, `(N)`, or `(H)`. `bootstrap/population.lua` uses that definition and `spawn.components` materializes the scaled monster. Module tests cover data interpretation, but not the complete authority-session path across all three difficulties. Applying the 1.13c rules to 1.14d remains inferred; owned 1.14d observations remain open.

## Approaches

1. **Production-session matrix with synthetic authored rows (chosen).** Start separate immutable sessions for difficulties 0, 1, and 2, generate the same monster through room activation with one and two living players, and assert exact level, life, XP, selected treasure class, and pinned count. It tests both game-rule selection and spawn materialization without protected data.
2. Expand only `monster_catalog_test.lua`. This is cheaper but cannot prove that the room generator and authoritative materializer use the selected definition.
3. Add a difficulty parameter to the monster spawn command. This would create a second mutable source of truth in conflict with `game_rules`.

## Behavior and constraints

Use an ordinary Expansion monster with `noRatio=0`, fixed equal min/max HP percentages, and distinct MonStats Normal/Nightmare/Hell percentages for HP, armor, and XP. Its MonStats `Level(N)` and `Level(H)` must differ from the area's `Levels.MonLvl2Ex` and `MonLvl3Ex`, and MonLvl must contain separate level rows and difficulty columns with discriminating HP, AC, and XP multipliers. For example, area levels 4 and 5 against MonStats levels 2 and 3 make a missed area override visible; choose exact integral bases and assert the spawned level, life, armor, and XP in each difficulty. The Levels fixture must include `mon1`, `nmon1`, and all three density columns so each session can populate the same class. Keep all three sessions otherwise equivalent; compare one-player and two-player results so the 50% count bonus is shown to operate on each difficulty's base. Include a bootstrap command with the matching difficulty and assert a mismatched bootstrap difficulty is rejected without changing the session rule in a separate session, since a failed command remains pending. Assert the spawned `d2legacy.monster.identity.treasure_class`; this proves selection of the loot class input, not an item drop or NoDrop formula. Reuse current Go synthetic-record and session fixtures. Do not change the shared count policy, loot roller, MonStats formula, or difficulty policy unless a failing production test identifies a concrete within-boundary defect.

## Remaining work

This test cannot establish original 1.14d behavior without an owned runtime observation and does not settle the user's dead-player loot-count question. Issue #373 remains open for those target-runtime and remaining loot observations. Real-asset tests require locally supplied legal game data, which is unavailable when `MPQ_DIRECTORY` is unset.
