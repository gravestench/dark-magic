# Issue #601 Ignore Target's Defense Design

## Intent and evidence

Honor an equipped or admitted `item_ignoretargetac` (stat 115) source in the authoritative basic melee hit roll. The bounded result is that a player attacking a classified ordinary monster uses defense zero for both the recorded hit chance and the hit draw. This implements the consumer described by HIT-R02c in `docs/specs/combat-hit-roll.md` without claiming that all original monster exclusions are classified yet.

HIT-R02c is verified in original 1.13c static code; HIT-V10 is a supplied-input vector, not an original-game observation. Applying it to 1.14d is inferred pending an owned observation. For equal levels with prepared AR 300 and defense 500, current chance is 37%; eligible Ignore Target's Defense makes it 95% after the existing cap. The original `monsterDefenseExclusion` producer is unresolved under HIT-Q03, so this change recognizes the known boss and hireling exclusions and records the remaining gap.

## Scope and constraints

- Issue: #601, within G8 combat fidelity. The issue is complete only when stat admission, equipped source lifecycle, eligibility, and authoritative melee results are covered.
- Preserve the current targeting rules, hit chance arithmetic, RNG stream and draw count, equipment source model, and combat event schema.
- Only a player attacker with a nonzero resolved stat 115 source may zero defense, and only for a monster defender that is neither boss nor active hireling. A player defender and a monster attacker retain defense.
- Use the existing `d2legacy.stat.source` store for passive and equipment facts; do not add persistent combat state or a second stat authority.
- Leave stat 116 defense reduction, unidentified monster exclusion flags, PvP combat admission, missiles, other hit-roll terms, and 1.14d observation for their own issues.

## Approaches

1. **Pure eligibility policy plus existing melee integration (chosen).** A small function in `d2legacy.policy.melee` accepts the prepared defense, resolved stat value, and supplied attacker/defender classifications. The combat system supplies classifications from existing components, uses the returned effective defense for chance and draw, and records it in the event. This gives the unavailable player-versus-player case a supplied-input test without altering targeting. It adds one narrow policy function and no new state.
2. **Inline the gate in `systems/melee.lua`.** This changes fewer production lines, but eligibility would be coupled to ECS and the player-defender exclusion could only be tested by bypassing normal targeting or adding an unrelated targeting mode.
3. **Generalize all defense adjustments now.** A reusable pipeline could later host stat 116 and unknown exclusion flags, but those inputs are unresolved or outside #601. Building it now would exceed the issue and make the result harder to review.

## Data flow

`enter_player.lua` admits additive integer `item_ignoretargetac` passive sources for controlled player fixtures and rejects the generic `percent` passive operation for this stat. `equipment.lua` includes the same stat in its existing modifier-source publication and clear lists so equip, swap, and removal propagate through the normal structural lifecycle. The melee system resolves the player's stat with `stat_sources.resolve(entities, attacker, "item_ignoretargetac", 0)` at impact; it supplies the attacker and defender types, boss marker, and active `d2legacy.owned_unit` category `hireable` to the pure eligibility function. Its declared ECS read set includes the boss marker. The effective defense replaces the original defense before `policy.hit_chance` and `policy.hits` and is written to the event's defense field. A non-eligible attack retains the original defense. The function itself has no random draw or state mutation.

## Acceptance evidence

- Pure supplied-input cases pin stat 0, positive and negative nonzero stat, ordinary monster, boss, hireling, player defender, and monster attacker. They retain or zero defense as HIT-R02c requires for the known classifications.
- Production melee cases assert the event defense and chance for ordinary (0, 95%) versus protected (500, 37%) equal-level targets at AR 300, and assert the actual draw outcome on a controlled random input where 37% misses and 95% hits. The same effective defense must be used by both chance and draw.
- Item modifier tests cover equip, switch away, and removal through existing `sync_sources`; player-entry tests cover passive admission and its production effect. The tests assert authoritative defense and hit chance after each equipment change, with no stale source on the next authoritative tick. The separate controlled-draw cases prove that eligible and protected defenses lead to distinct hit outcomes.
- Run focused Lua tests, the Lua harness, `go test ./...`, architecture, and Lua format checks on the final commit. State explicitly whether real game assets and race tests ran locally. CI must pass before merge.

## Remaining decision

HIT-Q03 needs an independently evidenced producer for `monsterDefenseExclusion`. Until then, an ordinary classified monster is eligible by the known predicates; this design does not claim complete 1.13c or 1.14d monster classification.
