# Lua ECS System Registration Ownership

## Intent and boundary

Issue #627 asks that a Lua scene release each ECS system from the engine on which it registered, even after a connected client rebinds `engine.ecs/v1` to its presentation replica. Today `registerSystem` registers on `capability.engine`, but its scope cleanup reads that mutable field again. `SetEngine` can change the field before the scope closes, causing an unregister error or removal from the wrong engine.

This is a Dark Magic resource-lifetime rule. It does not assert Diablo II gameplay behavior. The issue's hosted camera-follow error and the current registration/cleanup code establish the need.

## Goals and limits

- A system's registration, scope cleanup, and immediate rollback all use its registration engine.
- Rebinding the capability remains available for later registrations; systems already registered keep their original owner.
- A regression exercises register, rebind, and scope close with distinct engines and the same system ID on both, so silent wrong-engine removal is detectable.
- Remove the existing #627 cleanup-error suppression in the hosted bridge pointer fixture so its scene close must succeed after the fix.
- Keep production changes in Lua ECS registration. Do not change engine scheduling, connected movement, bridge routing, or the separate pointer-stop issue #634.

## Approaches

1. **Capture the engine for each registration (selected).** Store `capability.engine` in a local variable at registration, and use that variable for `Register`, the scope's release closure, and `scope.Add` rollback. This matches the existing per-resource scope ownership and needs no new state.
2. **Track system-to-engine ownership in `ECSCapability`.** This could support introspection, but adds a mutable registry and synchronization/lifetime decisions that issue #627 does not need.

The selected approach has one owner for the complete registration transaction. It does not alter how `SetEngine` chooses the engine for future Lua calls.

## Verification

Write a focused Lua runtime test first. Register a scoped system on an original engine, rebind the capability to a replica engine, register a same-ID system in a second scope there, then close the original scope. The original engine must lose its system while the replica retains its own; closing the second scope must clear the replica. The test must fail against the current cleanup closure before implementation. Inspect the immediate `scope.Add` rollback path to confirm it uses the same captured engine; this rebind regression does not exercise that rollback branch.

Remove the narrow #627 suppression in `newBridgePointerClient` cleanup and run a hosted owned-data bridge test so scene close must succeed without the camera-follow unregister error. Run the Lua runtime package, Lua hardening, architecture, and broad Go checks. If owned data is unavailable, report the hosted completion case as unverified rather than calling issue #627 complete.
