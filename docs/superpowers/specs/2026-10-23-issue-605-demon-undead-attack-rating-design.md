# Issue #605: Demon and undead attack-rating bonuses in player melee

## Intent and evidence

The basic melee hit roll currently ignores `item_demon_tohit` and `item_undead_tohit`, so a player's target-specific attack-rating sources do not change the authoritative chance against matching monsters. Issue #605 implements the bounded HIT-R02g consumer after #601/#603's defense adjustments. Original 1.13c static evidence verifies the rule and HIT-V13 supplies isolated vectors; applying it to 1.14d remains inferred. No original-game runtime outcome is claimed from synthetic tests.

## Goals, limits, and constraints

- Carry authored MonStats `demon`, `lUndead`, and `hUndead` facts through the shared definition and spawn path. Ordinary, summoned, and revived monster definitions retain their own type facts. Missing/invalid classification yields false predicates.
- For player basic melee, add nonzero stat 123 when demon, then nonzero stat 124 once when either undead bit is true. Both may add; negative values subtract. No target-specific bonus applies to monster attackers or non-monster defenders.
- Preserve the current post-percentage selected-hand delta. This issue inserts the target-specific flat addition before the existing `item_tohit_percent` source resolver, without altering broader percent arithmetic or the general target-independent derived combat stat.
- Use one resulting attack rating in the melee event, attack result, hit chance, and existing hit draw. Preserve targeting, event schema, RNG stream/draw count, equipment source lifecycle, and existing ECS authority ownership.
- Leave mastery/HIT-Q01, relation-matrix/HIT-Q02, missiles, damage bonuses, broader HIT-R03 reconstruction, and original-game hand-order verification to their owners. If implementation needs a broad rewrite, split the prerequisite before code changes.

## Approaches

1. **Existing definition/spawn facts and impact-time calculation (chosen by two delegated reviewers).** Select and decode the three MonStats fields in `data/monster.lua`; publish a compact immutable classification component in `spawn_monster.components`, declared as a melee read. At a player impact, resolve existing base attack-rating sources, add eligible type bonuses, pass the result through the current percent resolver, and apply the existing selected-hand delta. This computes target-specific AR when a target exists and avoids another persistent authority.
2. Cache a target-specific attack rating in `derived_stats`. That system has no selected target and would need new invalidation rules for target, equipment, and source changes, so it is unsuitable.
3. Re-read MonStats at each impact. This would duplicate data interpretation and add record access in the combat path, so it is unsuitable.

## Behavior and data flow

The Go MonStats model already names the three boolean columns. The Lua monster catalog must select them; the shared `runtime_definition` should expose three booleans for `load`, `summon`, and `revive`. The existing spawn materializer publishes these as one authoritative classification fact on the monster. A missing fact is treated as all false for melee; this supports synthetic and older inputs without guessing from IDs. The spawn materializer is shared by ordinary and owned monster paths, so those authored facts survive without new lookup in combat. This design does not imply a player can target a friendly summon.

Extend the existing player-entry passive-source admission and equipment publication/cleanup lists to accept `item_demon_tohit` and `item_undead_tohit` as additive flat sources; reject player-entry passive `percent` for these stats. At impact, only for a player attacker, the melee system rebuilds pre-percentage AR from `combat_stats.base_attack_rating` through current `attack_rating` sources. Its pure policy applies the demon bonus, then the undead bonus once, using the target's classification. The system then applies current `item_tohit_percent` resolution and the existing hand delta. It leaves `combat_stats.attack_rating` intact for other target-independent consumers and keeps the non-player path unchanged. The defense adjustments from #601/#603 still precede the target bonus before the existing chance and draw.

## Acceptance evidence

Pure policy tests cover all distinct HIT-V13 results: base AR 300 with demon +50 -> 350, either undead flag +40 -> 340, demon plus an undead flag -> 390, neither/non-monster/invalid definition or non-player attacker -> 300, and both negative bonuses -> 210. Both undead bits together add stat 124 once; zero values do nothing. The operation draws no random number.

Synthetic loader and spawn tests prove the three fields survive MonStats -> definition -> authoritative component for ordinary and owned definitions. A local test against representative owned 1.14d MonStats rows must also prove the loader/spawn classification path; the available owned MPQ directory is exercised without committing data or captures. Record exact rows and results as table observations, which do not upgrade 1.14d runtime behavior to verified. If local data cannot be read, this criterion stays unverified and #605 is not closed. Passive admission, equipment equip/switch/switch-back/remove, and impact-time tests prove source values and subsequent AR/chance update without stale sources.

A controlled equal-level melee case uses pre-percentage AR 300, demon bonus 50, percent source 20, and target DEF 400. Seed 429 gives the first `d2legacy.combat.basic_melee.hit` draw of 50 under the current named-stream algorithm. Correct AR 420 yields chance 51 and hits; adding bonus after the percent produces AR 410/chance 50/miss. Both event records must agree. Selected-hand cases use primary pre-percent AR 300, secondary delta +20, and percent 20: a monster with neither type flag records primary/secondary AR 360/380, and a demon with +50 records 420/440. These preserve current hand behavior while testing target-specific bonus order. Monster attackers and nonmatching targets do not gain a bonus.

## Remaining uncertainty

Original 1.14d runtime behavior remains unobserved, and HIT-V13 is supplied-input evidence. Current selected-hand delta placement is preserved as implemented; this design does not claim that placement matches original Diablo II. The existing percentage resolver is reused without claiming complete HIT-R03 fidelity. Original-game evidence or a separately scoped issue must settle those gaps.

Before submission, run the existing melee and equipment suites and the appropriate broad, architecture, and Lua hardening checks on the submitted commit. The PR must give exact commands and results and state whether local real-asset and race checks ran. The owned-row check above is required to close this issue.
