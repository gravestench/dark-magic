# Issue #634: connected pointer route progress design

## Intent and evidence

An owned-data full `clientapp` package run intermittently stopped the connected normal-entry route before the southeast Rogue Encampment bridge. This is issue #634's boundary. The test starts at the normal town spawn, clicks a sequence toward the bridge, crosses to Blood Moor, then returns; the issue's earlier wording describes the route as beginning in Blood Moor, but the failing test begins in town.

A throwaway Go test overlay captured a stop while clicking `(252,84)`: the presented ECS position was `(250.667,84)`, the authenticated HUD at tick 480 was `(250.875,84)` with zero velocity, and the listen host at tick 481 had accepted another command toward waypoint `(251,84)`. The movement adapter's route progress check uses the presented position and a `0.3` arrival distance. That position is `0.333` from `(251,84)`, so it resends waypoint 251; the authenticated position is `0.125` away. The host movement rule's `0.2` arrival distance likely explains its zero velocity, although the host log and HUD are one tick apart. An independent camera diagnostic found that the scene click reverse-projected to the intended world target in all sampled steps. No product code was changed for these diagnostics.

This is a Dark Magic connected-input invariant, not a claimed retail Diablo II movement rule. The visible position can be smoothed for rendering, but a connected route should advance against authenticated movement progress.

## Approaches

1. **Increase the adapter's `0.3` waypoint tolerance.** This may consume a collision corner before the authoritative player reaches it, changing every route and leaving the disagreement between displayed and authenticated positions unresolved.
2. **Use authenticated local position only to consume connected waypoints.** Keep the presented ECS position and collision radius for route search and rendering; when the authenticated sample names the same player and level, use its position to decide whether the current waypoint has been reached. This is the chosen approach. It follows the movement authority without changing path geometry or offline input.
3. **Remove presentation smoothing or use prediction to advance the route.** That would change how connected movement looks or treat speculative position as progress, respectively, and exceeds the demonstrated input boundary.

## Boundary and ownership

- The client reconciliation path already receives a validated authenticated HUD. It passes a local player ID, level ID, and position sample to the existing `MovementSource`. The source stores only this local input progress sample, not an authoritative world state or a new movement authority.
- `MovementSource` uses the sample to consume waypoints and complete a connected route; its presented position remains the pathfinder start and collision radius source. When authenticated progress completes the final waypoint, it emits a targetless, zero-axis stop instead of targeting a lagging presented position. The host's existing `Resolve` rule makes that command stationary. Offline completion retains its current presented-position stop target. The source ignores a sample for a different player or level. Changing the controlled player, ECS engine, or map invalidates any saved sample; reinstalling the same map pointer retains the active route.
- Offline routing and connected routing before a HUD sample keep their current presented-position behavior. The `0.3` adapter tolerance, `0.2` host arrival tolerance, collision rules, click policy, seam transitions, and authority admission remain unchanged.
- The failure capture and regression tests make target choice observable in test diagnostics. A general product movement debug UI or network protocol addition is outside this issue because the host accepted the target and the demonstrated loss is route progress inside the client adapter.

## Verification

- First add a deterministic adapter regression with a displayed position `(250.667,84)`, authenticated position `(250.875,84)`, and accepted path through `(251,84)` to `(252,84)`. With same player and level, the emitted move command must target `(252,84)`; the old adapter targets `(251,84)`.
- Add a completion regression where authenticated progress is within `0.3` of the final waypoint but the presented position lags farther away. The emitted connected command must have no target and zero axes, so the host cannot be pulled back to a stale display coordinate. The no-HUD offline case must keep its existing stationary target.
- Exercise the production client reconciliation caller in a deterministic replica test: force a displayed position below the first waypoint's `0.3` threshold, provide an authenticated HUD within it, and assert the next movement command advances. This test must fail when the reconciliation setter call is removed; repeated owned-route green runs alone cannot establish that connection because the route sometimes passed before the fix.
- Guard the counterexample: a displayed position near a route corner must not consume it when authenticated position remains farther than `0.3`; wrong-player and wrong-level samples must not influence progress. Preserve the offline/no-HUD behavior and test invalidation on owner, engine, and map changes where relevant.
- Run the owned-data connected normal-entry bridge test repeatedly in isolation and in full `clientapp` package order. Keep the offline normal-entry route, southeast shoulder crossing, and round-trip coverage green. Report any CI licensed-data skips explicitly.
- Run repository Go tests, architecture, race, and appropriate lint/build checks on the submitted commit. The exact test commands and any unverified native reproduction belong in the PR description.
