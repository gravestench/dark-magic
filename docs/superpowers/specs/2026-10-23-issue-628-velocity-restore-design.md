# Issue #628: keep velocity movement bound to the restored ECS world

## Intent and boundary

An authoritative listen session must keep moving opted-in entities after a late input or recovery replaces its ECS world. An intermittent owned-data Cold Plains run panicked while the native velocity integrator read a position component. The stale-store mechanism is a confirmed engine restore defect consistent with that panic; the failing run was not instrumented to prove that a restore immediately preceded it. The fix stays within the native velocity system, its focused regression tests, and removal of the now-redundant offline re-registration workaround. It does not change movement speed, collision, monster population, or the ECS update API.

## Root cause and evidence

`RegisterVelocityMovement` captures three `*akara.DynamicStore` handles in its `Update` closure. `Engine.Restore` creates a new world and remaps system filters and subscriptions, but preserves that closure. The next subscription supplies entities from the new world while the closure reads the old stores. An entity absent from the old store returns a nil component and panics at `velocity_movement.go:133`; an entity present in both worlds can be written only in the old world, leaving the restored state unchanged. Session late-command rollback and durable recovery both call `Engine.Restore`. The offline entry-world path currently works around this by unregistering and re-registering native movement after restore; listen-host rollback does not.

The intermittent clientapp panic stack and test order are recorded in issue #628. Independent read-only code reviews confirmed the old-world handle mismatch as a failure mechanism, while the exact cause of that intermittent run remains unproven. No original-game evidence is needed for this engine invariant: the governing documents are `docs/ARCHITECTURE.md` (one authoritative, checkpointed ECS) and `docs/NETWORKING.md` (transactional rollback/replay).

## Approaches considered

1. **Refresh the velocity callback's stores when its engine world changes (chosen).** Capture the engine and three schema names alongside the initial handles. At each batch callback, compare the current world identity with the cached one; only after a change, resolve the three stores by name and replace the cache. `Engine.Update` and `Engine.Restore` serialize through `updateMu`, so a callback cannot observe a half-restored world. Normal ticks incur only the identity comparison, and the entity loop keeps direct store handles.
2. Pass the current world through every ECS `UpdateFunc` context. This makes world identity explicit across all systems, but changes a broad engine API for one known native callback and is outside this issue.
3. Re-register movement around every session restore. The offline path already does this, but rollback, failed-transaction rollback, replay, and recovery have multiple restore sites. Repeating the workaround would be fragile and would leave stale closures possible elsewhere.

The user previously delegated design review and approval to subagents. Two independent reviews selected approach 1 based on the code and architecture. Remove the offline-only unregister/re-register block once its existing recovery regression passes, so offline and hosted restore use the same callback behavior.

## Contract and failure handling

The callback retains the original schema configuration and pathfinder. Restore is accepted only when the replacement world has schemas required by the registered system filters, so refreshed stores should exist. If a store is unexpectedly absent, return a named error from the system update rather than dereferencing nil. The per-entity component lookup should likewise return a useful error if an opted-in entity no longer has a required component. These checks do not alter successful movement or collision behavior.

## Verification

1. Add a synthetic red regression in `internal/game/world/velocity_movement_test.go`: restore a snapshot with an existing mover, tick, and assert the **restored world** position advances; restore again and repeat. The current closure silently writes the old world.
2. Add a red case that restores an empty registered world, then creates a mover in the new world and ticks. The current closure panics on the old store's absent entity; the new behavior advances without panic.
3. Add a deterministic Session-level regression in the world test package: run a registered velocity mover for two ticks, submit an ordinary command late enough to invoke Session rollback/replay, assert the ECS world was replaced, then advance one more tick and assert the restored world's position advances. This pins the affected authoritative rollback path without game data.
4. Remove the offline re-registration workaround. Run `internal/game/world`, `internal/game/ecs`, and the existing offline recovery test; with owned 1.14d data run the connected Cold Plains population test and affected `clientapp` package. Run repository CI gates on the submitted commit. CI-owned-data tests skip without licensed assets; report that explicitly.

## Limits

This corrects the known velocity callback. Other callbacks that capture world-bound stores are a separate audit if found. The original intermittent panic can still depend on timing in a broad test order, so the deterministic restored-world tests are the primary regression evidence.
