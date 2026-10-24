# Issue #612: Death Sentry Corpse Roll Uses an Exclusive Upper Bound

## Intent and evidence

Death Sentry's supported corpse explosion currently calls the shared inclusive `damage.roll(minimum, maximum)`. The merged [corpse-damage specification](../../specs/skill-assassin-death-sentry-corpse-damage.md) records 1.13c static code evidence in DSC-R04: for an ordinary ordered range, the corpse roll draws from `[minimum, maximum - 1]`; a nonpositive range adds zero without a draw. Issue #612 asks for that one parity correction without changing other damage rolls or the independently truncated physical/fire split from #610. Applying the 1.13c result to the supported 1.14d path is an inference, not a runtime observation.

## Goals and acceptance boundary

- Only `explode_corpse` in `d2legacy.systems.trap_skill` uses an exclusive upper bound for ordinary ordered endpoints. Keep the existing endpoint computations, split, target selection, radius, mitigation, combat event, corpse consumption, and checkpoint behavior.
- When `maximum > minimum`, consume one draw from the existing `d2legacy.combat.damage.roll` authoritative stream with bound `maximum - minimum`, then add it to `minimum`. The maximum itself is unreachable.
- When `maximum <= minimum` for ordinary nonwrapping inputs, return `minimum` without calling the random service. Equal and reversed endpoints must not assert.
- The shared `d2legacy.policy.damage.roll` remains inclusive for its other callers.
- Add an authority-path test whose fixed seed would hit the inclusive-only maximum, and assert the staged physical/fire channels and combat-event total. Cover equal and reversed endpoints plus draw consumption or deterministic continuation as the existing harness permits. Preserve checkpoint parity.

## Approaches

1. **Chosen: a small local corpse-roll helper in `trap_skill.lua`.** It calls the existing authoritative random service only when the range is positive. This keeps the special rule adjacent to the only owner and does not expand the shared direct-damage API.
2. Add an `exclusive` flag to shared `damage.roll`. This introduces a mode on every caller's common interface and makes an unrelated direct-damage policy carry Death Sentry's exception.
3. Add a new policy module and random stream for corpse rolls. This creates another owner and changes replay stream topology for one call site, neither required by #612.

## Design

Add `engine.authority_random/v1` to the trap system's imports and a named local function near `explode_corpse`. Its guard returns the lower endpoint for equal/reversed ordinary ranges. Its ordered branch uses `random.integer("d2legacy.combat.damage.roll", maximum - minimum)` and adds that result to the lower endpoint. Replace only the corpse explosion's current `damage.roll` call. Do not change the later `damage.roll` call used by another trap path or the policy module.

Use the existing synthetic authority suite, with its fixed session seed 42 and checkpoint parity, to distinguish inclusive from exclusive behavior through the real trap transaction. Change its fixed corpse helper percentages to 50/53. A corpse with max health 202 then yields endpoints 101/107. The first shared damage-stream value is `1022979460928719132`: bound 7 (current inclusive call) yields offset 6, amount 107, independently truncated channels 53/53, and event total 106; bound 6 yields offset 4, amount 105, channels 52/52, and event total 104. Assert both staged channels and the event total, not merely the selected scalar. A second case in that suite uses max health 20 to obtain equal endpoints 10/10 and independently truncated channels 5/5. A small neighboring authority suite uses fixed skill ID 276 with a reversed 53/50 helper record to obtain endpoints 107/101. After checkpoint parity, draw a controlled next value from `d2legacy.combat.damage.roll` with bound 1000 for each nonpositive-range case: it is 132 if the explosion consumed no draw, versus 46 after an unwanted draw. The authority bootstrap admits only coverage-listed trap IDs, so merely adding synthetic sentry IDs to one suite would not exercise them. Keep the helper records fixed within each suite; mutating runtime definitions would bypass checkpoint ownership and weaken replay evidence. Share only small test fixture helpers if duplication would otherwise obscure the cases.

## Verification and limits

Write tests first and observe the ordered case fail against today's inclusive path; equal/reversed cases should show the current assertion or unwanted draw. Run the focused Lua trap suite, checkpoint parity, Lua formatting/syntax/hardening, repository tests and race gate, architecture, lint, and both client builds on the submitted commit. Record owned game-data test status exactly. Review the whole diff from correctness, spec/evidence, and style angles, with the style reviewer using `refactor-for-readability`.

Wrapped 32-bit endpoint arithmetic, corpse-owned original-game seed topology, source of monster max life, and other #420 gaps remain outside #612. This slice does not claim native 1.14d observation or finish parent #420.
