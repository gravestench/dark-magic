# Issue #597: shared act-town level IDs

## Intent

Make the existing shared town predicate use the five level IDs established by original Diablo II 1.13c code. Correcting that one predicate lets authoritative stamina, connected stamina prediction, poison ticks, aura pulses, and corpse targeting receive the right town answer through their current interfaces. For the 1.14d target this is an inferred carryover until an owned observation confirms it.

## Boundary

The issue covers level-ID membership in `movement.IsTownLevel`, its Lua bridge, and production-path evidence that the corrected answer reaches stamina, connected prediction, and one non-stamina consumer. It does not change room lookup, movement mode, poison/aura/corpse policy, or stamina drain arithmetic. It does not claim an original 1.14d observation.

## Approaches

1. **Correct the existing shared predicate (chosen).** Replace the five values in `internal/mod/d2legacy/adapter/movement/rules.go`; retain the current Lua bridge and prediction call. This is the smallest change that preserves one answer for every current consumer. Rename the stale test and update the stale 1.14d comment to state the 1.13c evidence and 1.14d inference.
2. Duplicate a town table in each Lua and Go consumer. This would let each consumer choose its own answer and make future correction harder.
3. Migrate town classification and all movement policy into Lua now. That is a broader architecture change, including connected prediction, beyond issue #597.

## Acceptance and evidence

- `IsTownLevel` accepts exactly `{1,40,75,103,109}` among relevant level IDs. It rejects the prior false positives `{45,80,108,114}`, adjacent IDs, zero, and negative/out-of-range IDs. The direct Go test and Lua bridge test exercise the same predicate.
- A synthetic authoritative player tick starts at 2,560 raw stamina with an 84-point maximum and moves while running. At level 40 it follows the existing town recovery path and reaches 2,602 raw; at level 45 it takes the ordinary 40-unit outside-town drain and reaches 2,520 raw. Use the current player-entry/test fixture rather than a helper-only call.
- A connected-prediction tick at levels 40 and 45 observes the corresponding 2,602/2,520 raw results through the existing HUD and prediction path.
- A poison tick test supplies otherwise identical active poisoned players in levels 40 and 45: level 40 receives town protection, level 45 spends poison damage. Preserve the existing dead/inactive checks.
- Run the focused Go and Lua suites, architecture and formatting checks, then broader tests. Record exact results on the final commit and whether real-asset tests ran.

The original 1.13c table is traced in `docs/research/spec-evidence/combat-knockback.md` and consumed by KNB-R03b. `docs/specs/player-stamina-drain-while-running.md` SDR-R02/SDR-V06 corroborates the set and town stamina exemption. No community recollection upgrades 1.14d to verified.

## Risks and limits

Several existing consumers will change behavior from one shared predicate correction. The tests target representative live paths; they do not revalidate every aura or corpse rule. Existing town predicate location in the Go adapter is an established shared prediction seam; this issue does not add another policy owner. If the connected-prediction harness cannot exercise a tick through its current interface, add the narrowest reachable production-path test without changing the interface, and document that limit.
