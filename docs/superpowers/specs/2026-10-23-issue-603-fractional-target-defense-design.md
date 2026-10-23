# Issue #603: Fractional Target Defense for player melee

## Intent and evidence

Issue #603 is the next bounded HIT-R02d consumer after #601. A player's additive `item_fractionaltargetac` (original stat 116) source should lower the prepared defense used by an eligible basic melee hit. Players see the corresponding hit chance change when they equip, swap, or remove the source. Original 1.13c static code verifies the rule and its order after Ignore Target's Defense; HIT-V12 is supplied-input evidence. Applying the rule to 1.14d remains inferred. HIT-Q03 leaves the `monsterDefenseHalf` condition producer unknown, so this design recognizes only player defenders, boss monsters, and active hirelings for the known halving branch.

## Goals, limits, and constraints

- Use one adjusted defense for the attack event, current hit-chance arithmetic, and its single hit draw. Keep targeting, event schema, RNG stream/draw count, and existing item source lifecycle.
- Match signed 32-bit and truncation-toward-zero arithmetic in HIT-R02d: positive source only, halve for protected targets, cap after halving at 100, divide defense first only when it exceeds 1,048,576, then subtract the cut. Apply this after stat-115's zero-defense gate; do not floor a negative result here.
- Admit additive player-entry passive stat-116 sources and publish additive equipment modifiers through the current stat-source owner. Reject `percent` in player-entry passive admission only; equipment modifier admission is outside this issue.
- Leave total-defense reconstruction, HIT-Q03's original condition producer, PvP targeting, missiles, and other attack-rating modifiers for their owners. Player-target halving is a supplied-input policy test, not a claim of reachable combat.

## Approaches

1. **Extend the current melee policy (chosen).** Add a pure `fractional_defense(defense, value, half)` operation beside `effective_defense`. The melee system calls it for player attackers after the existing Ignore Target's Defense step, supplying `half` from a player defender, boss marker, or active hireling relation. The resulting defense flows to the existing event, chance, and draw. This keeps arithmetic and eligibility inputs reviewable and requires no new state.
2. Put arithmetic directly in `systems/melee.lua`. This would shorten a call path but mix original-game integer policy with ECS queries and make supplied-input edge cases harder to test.
3. Cache another combat-defense component. This would duplicate a value that can change with current equipment and risk stale state across weapon switches; it is unnecessary for one impact-time read.

## Behavior and data flow

`enter_player.lua` admits `item_fractionaltargetac` for additive passive sources and rejects its passive `percent` operation. `equipment.lua` includes the same name in its active modifier publication and no-layout cleanup lists. The melee system resolves stat 116 only when the attacker has player identity, after resolving stat 115. It uses existing target identity, `monster.boss`, and active owned-unit category `hireable` to decide the known half branch. For an ordinary monster the full value applies. A monster attacker never gets this player-only adjustment. A player defender would use half if targeting supported it; supplied policy tests lock that rule. The unknown `monsterDefenseHalf` producer remains false in the current integration and is not assigned a guessed meaning.

The policy takes a signed prepared defense and additive value, applies signed i32 low-word arithmetic with truncation toward zero, and returns the adjusted defense. A nonpositive value returns the input unchanged. A value greater than 100 still halves first for protected targets, then caps. A zero defense from Ignore Target's Defense remains zero after reduction. No new random call occurs.

## Acceptance evidence

Pure policy cases reproduce HIT-V12's DEF 500 reductions, nonpositive values, cap, high-defense branch, and negative input. Derived cases distinguish divide-first at DEF 1,048,577/value 50 -> 524,327, negative truncation at DEF -501/value 50 -> -251, signed low-word multiplication at DEF -2,147,483,648/value 50 -> -2,147,483,648, and half-before-cap at protected DEF 500/value 250 -> 0. A composition case sends stat-115 zero defense through stat-116 reduction.

A seed-36 authority suite uses equal level 20, AR 300, monster DEF 500 and first hit draw 51: ordinary value 50 -> DEF 250/chance 54/hit; boss or active hireling -> DEF 375/chance 44/miss; zero/negative value -> DEF 500/chance 37/miss. It compares melee and attack-result events and checks monster attackers cannot apply the player stat. A dedicated equipment fixture with no stat-115 source verifies resolved stat 116 = 50/0/50/0 and authoritative defense/chance after equip, switch away, switch back, and removal; player-entry tests cover passive admission and rejection. No real-asset assertions are needed for this source consumer.

## Remaining uncertainty

The original `monsterDefenseHalf` producer and 1.14d runtime behavior need later evidence. Current tests establish the 1.13c static rule on supplied and synthetic production inputs, not a full original-game observation. The existing hit-chance policy handles negative-defense normalization after this adjustment; its broader reconstruction and overflow behavior are outside #603 except the signed low-word arithmetic in this local reduction.
