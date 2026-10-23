# Issue #607: shared monster spawn ECS write contracts

## Intent and boundary

Repair access declarations for systems that already materialize shared monster components. The immediate source-level mismatch is `d2legacy.trap.deploy`: `sentry_components` returns `spawn_monster.components` with `d2legacy.combat.defense`, then `deploy_sentry` passes that table to `structural:create`, while `trap.deploy.write` omits defense. Shared spawning can also produce boss, freeze-immune, and prime-evil markers, and the four system writers omit those marker names. This design does not change which monsters spawn, which facts their authored definitions contain, combat math, or original-game parity claims.

## Approaches

1. **Explicit per-system declarations and authority tests (chosen by two independent reviewers).** Compare the final component table from each supported spawn path with its ECS write list. Add exact missing names and test actual system execution. This follows current code style and has a small behavioral surface; future drift is addressed with focused tests.
2. Strip components that a caller does not declare. This can silently remove defense or authored classification facts and changes gameplay behavior, so it does not fit #607.
3. Introduce a shared manifest that grants all spawn writers a common access list. This could reduce duplication but broadens permissions and abstracts caller ownership beyond the demonstrated repair. Revisit only if inventory/testing shows repeated drift that cannot be contained locally.

## Existing path inventory

`spawn_monster.components` always returns monster identity/stats/type flags, combat defense/melee profile/knockback target, appearance/AI, world position/velocity/facing/location/collider/occupancy/selectable, and velocity mover. It conditionally adds corpse-selectable, revivable, boss, freeze-immune (with boss), and prime-evil markers. Population adds room-resident. Corpse summon, golem, and trap add owned/summon-specific components.

`monster.load` admits enabled `isSpawn` non-NPC rows for population. `monster.summon` admits enabled non-NPC authored rows for fixed skill IDs, then forces corpse-selectable and revivable false. `monster.revive` admits enabled non-NPC revive-capable rows and also strips those two corpse capabilities from the owned copy. Thus golem and trap need no corpse-capability write unless their final table can acquire one elsewhere. Population and corpse writers already declare those capabilities. No caller currently declares boss, freeze-immune, or prime-evil writes. The ordinary data reachability of those flags and any eligible summoned/revived type must be checked without treating synthetic fixtures as shipped-game observations.

## Implementation and tests

First, add an authority-level sentry-deployment test using the existing trap fixture as a base and a real cast through `d2legacy.trap.deploy`. Extend its synthetic MonStats/graphics/level records with the `enabled` and runtime fields required by `monster.summon`; the current fixture omits them because its tests install sentries directly. Confirm RED specifically with undeclared `d2legacy.combat.defense`, then declare that write and assert the created sentry retains defense. Keep existing direct-installed sentry tests; they bypass deployment and do not prove this contract.

Next, use synthetic admitted rows/recipes to exercise conditional marker materialization through population and owned summon/Revive paths. Confirm the access failure before declaration where possible, add boss/freeze-immune/prime-evil to each writer whose supported definition can carry them, then assert retained markers. If a path cannot reach such a definition under its current admission rules, record that exclusion and do not add an unnecessary declaration. Verify the final component-key inventory against each writer, including caller-specific additions. Use existing checkpoint/replay tests where the production path already has them; no new persistence mechanism or behavior is introduced.

Run focused spawn, trap, summon, and population suites, then full Go, architecture, Lua hardening/format/syntax checks on the submitted commit. Report whether an owned data test or local race check ran. An owned MonStats query may establish which 1.14d authored rows carry the flags; it cannot prove original 1.14d combat behavior.

## Uncertainty and follow-up

The sentry failure is presently inferred from the builder and write list, awaiting a production-path reproduction. Conditional marker reachability across shipped definitions has not been established. Unexpected gameplay eligibility or component-content changes discovered during tests require a separate issue rather than expansion of #607. Original-game 1.13c/1.14d behavior is unchanged by these access declarations and is not claimed verified by them.
