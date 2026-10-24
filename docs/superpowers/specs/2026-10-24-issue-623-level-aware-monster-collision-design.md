# Level-aware native monster collision — issue #623

## Intent and boundary

Ordinary Cold Plains monsters already have authoritative `d2legacy.world.location.level_id`, position, velocity, and collider components. Their AI requests velocity in the intent phase. The one native movement system integrates that velocity in the movement phase, but offline and listen hosts currently give it only the Blood Moor map. A level-3 monster can therefore be stopped by level-2 terrain or move through level-3 terrain. The fix is to select the prepared collision map from each mover's current authoritative level on each tick.

This is a Dark Magic authority correction for existing generated Act I maps. It preserves one native movement system, its phase and ID, existing path/step rules, rejected-move velocity consumption, Lua-owned monster policy, and current generated maps. It does not redesign AI, monster paths, density, generation, warps, or original Diablo II collision arithmetic. Owned 1.14d maps are integration fixtures; they do not prove 1.13c or 1.14d retail movement parity.

## Approaches considered

1. **One integrator with an explicit level lookup (chosen).** Give the existing system a configured location component/field and an immutable level-to-pathfinder lookup. Declare the location read, refresh its store after ECS world replacement, and resolve a map before the collision fast path. Offline and listen composition supply the same prepared maps. This keeps the dependency visible and preserves tick ordering.
2. **An entity-aware selector adapter.** Keep the integrator unaware of location fields and let a callback read ECS state. This can work but hides an ECS dependency in a closure and is easier to get wrong after checkpoint restore.
3. **One movement system per level.** This duplicates system registration and complicates stable ID/order. It does not fit the issue's single-authority requirement.

Independent architecture and evidence subagents selected approach 1. The evidence review corrected an initial identity idea: `authoritativeCapabilityVersions()` exactly matches installed Lua modules and cannot list a native-only capability.

## Design

Keep the existing single-map `RegisterVelocityMovement` contract for generic callers. Add one level-aware registration entry point that shares its registration and per-entity integration core with the single-map path; it must not create a second system or a second movement algorithm. Its configuration names the location component and integer level field rather than hardcoding `d2legacy` in the engine. The engine reads that component for every opted-in mover, selects from a copied immutable level-to-`VelocityPathFinder` lookup, and then calls the existing collision and write path.

The level-aware registration declares location as a read dependency but does not put it in the query's `All` set: a mover missing location must produce an explicit error rather than silently disappear from movement. A missing component, invalid field/type/level, or absent/nil map fails the tick before that mover’s position is written. Selection happens before same-cell bypass, so missing geometry never becomes collision-free movement. The existing adjacent-step validator, long-step path search, radius, and velocity-zeroing on blocked moves remain unchanged. On `Engine.Restore` or rollback, refresh the location store with the other cached component stores before reading the next mover. A location change between ticks selects the new map immediately.

The client app builds one level lookup from its already prepared, published `gameWorlds`, including the two-zone recovery topology when applicable. Both `registerOfflineMovementSystem` and `installListenWorlds` call the same level-aware contract. No mutable copy of location or map state is introduced. Lua's existing `d2legacy.world.location` schema and population lifecycle remain the authority.

The native behavior change affects replay and recovery compatibility even though it changes no Lua source. Add an immutable native movement policy version field to `sessionInitialData()`, shared by offline identity, listen-host construction, and connected-client identity. That data already contributes to `RuntimeIdentity.ConfigurationHash`. Identical peers and restored sessions must agree on it; an otherwise identical old identity without the field must be rejected. Do not add a fake Lua capability or loosen `validateCapabilityIdentity`. No migration of old authoritative checkpoints is part of #623.

## Verification

- A focused engine test uses two synthetic maps with opposite collision at the same adjacent step. Level-2 and level-3 movers start at `(9.4,12)` with radius `0.1` and velocity `(25,0)` for `40ms`, crossing into cell 10. One advances; the blocked one stays and consumes velocity. Reverse which map is blocked and repeat, including a location change on one entity. Test same-cell missing-map failure as well as adjacent and long-step behavior.
- Engine restore and session rollback tests replace the ECS world, then move a level-3 entity and prove the selector reads its current location and map. Preserve existing single-map tests.
- Production offline and listen-host tests use the existing Cold Plains population fixture, real admitted seam movement, matched completed ticks, and a moving spawned monster. Assert AI chase state/target/velocity, then choose a deterministic local step where prepared levels 2 and 3 differ and assert the position/velocity outcome against the entity's own level. The listen client is compared to an acknowledged host checkpoint, not an unacknowledged frame.
- From a Cold Plains checkpoint, uninterrupted and restored authorities receive the same later commands across a 3→2→3 crossing. Compare positions, complete checkpoints, replay commands, and checksums. Prove the native version field changes the runtime identity and compatibility checks reject the old identity.
- Run focused, full Go, architecture, race, Lua, and owned-data tests that exercise these paths; state skips and unverified graphics/original-game details precisely. No owned game assets or captures enter the commit.

## Risks and uncertainty

The real maps may have few convenient opposing local steps near a live monster. The owned fixture must search deterministically for a reachable discriminating step or fail with a clear fixture diagnostic; it must not turn a missing case into a passing skip. The synthetic tests establish the selector independently of the current generated geography. The new identity field intentionally makes old single-map checkpoints incompatible with new authoritative behavior. Retail Diablo II monster path types, collision masks, speed, and tick cadence remain unverified by this issue.
