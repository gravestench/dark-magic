# Issue #642: Release unfinished render preload jobs after scene teardown

## Intent

The outdoor game scene can close while its map tiles, monsters, missiles, overlays, or hero art are still decoding. The native preloader rejects release of unfinished jobs, while several Lua paths clear a job ID without releasing it even after completion. The shared preloader then retains completed bookkeeping across repeated scene visits. Release these records without changing visible outdoor seam behavior, gameplay authority, or native texture ownership.

## Goals and boundary

- `engine.render/v1.preload_release(job)` accepts a valid unfinished job exactly once, hides its status immediately, and removes its record after every worker in that job finishes. Completed and empty jobs release immediately; unknown or already released IDs return false.
- Worker decoding may finish after release. `Pending()` still counts actual unfinished work, including released jobs, until completion. Release/completion races cannot panic or resurrect records.
- `chunked_map.destroy` releases its initial job and every tile job regardless of completion, clears its handles, and does not admit nodes if update is mistakenly called after destruction. The shared camera root remains owned by the scene.
- The `game_world` scene releases its own monster, missile, overlay, and hero preload IDs before clearing them on normal completion, entity removal, non-retained world change, or scene teardown. Existing warp and ground-item release paths use the new native early-release semantics. A seamless crossing that retains the hero keeps its current pending art job.
- Repeated game-world and adapter create/destroy cycles leave no completed preload records. Synthetic tests use no game assets or captures.

## Non-goals and constraints

- No worker cancellation, cache eviction, native texture disposal, new scene-wide job registry, changes to other screens' long-lived preload policy, or change to a live adapter's draw/admission behavior.
- Keep decoding on bounded workers and native texture uploads on the renderer thread. Retained render nodes remain owned by the Lua scene and are destroyed separately from job bookkeeping.
- The issue covers the outdoor `game_world` scene and its sparse adapters. Other screens that intentionally retain preload handles are outside this acceptance slice.

## Approaches

1. **Chosen: deferred native forget plus owner cleanup.** Mark a pending job released under the preloader mutex, hide its status, and delete its record on final worker completion. Each outdoor scene owner releases its own handles at completion and teardown. This uses existing ownership boundaries and preserves pending-work diagnostics.
2. **Worker cancellation.** Propagate cancellation through queues and asset decoders. This crosses many decode paths and still needs completion bookkeeping; #642 does not require stopping CPU work already begun.
3. **Scene-scope registry.** Transfer every job ID to a new scene owner that polls after teardown. This duplicates the preloader's registry and creates periodic cleanup outside current owners.

## Design

Add a release marker to each native job. `Forget` returns false for an absent or already released ID; it deletes a completed record or marks a pending one. `Status` treats released jobs as absent. Worker completion updates the job under the same mutex and deletes it when all requests finish and it is released. `Pending` continues to sum unfinished requests until deletion. The lock serializes release before/after final completion.

In Lua, `chunked_map.destroy` releases every retained ID without a status check, nils the initial ID, clears the tile-job map, and sets a destroyed flag before removing nodes. `update` returns without admitting nodes when destroyed. In `game_world`, add release-before-nil at normal completion for monster, missile, overlay, and hero jobs. Release pending IDs when their entity/node is removed, when a non-retained hero is replaced at a warp, and at scene teardown. Reuse narrow cleanup helpers for the corresponding removal paths; do not change AI, rendering recipes, seam layout, or camera policy. Warp and ground-item owners already call release on removal.

## Verification

- Native tests use controlled blocked workers to prove release before and after completion, exactly one successful release, unavailable status, `Pending()` transition, and zero retained records after completion; repeat the cycle and run a race-instrumented ordering test.
- One synthetic Go/Lua integration test registers the real `engine.render/v1` capability and the first-party `chunked_map` module. It starts an unfinished initial preload through the adapter, destroys that adapter through Lua, lets the native worker complete, and checks the real preloader's job map is empty. Repeating the cycle proves bounded records. A missing synthetic DS1 may fail decoding; the lifecycle result is the assertion.
- Lua scene tests use tracked synthetic jobs for monster, missile, overlay, and hero completion/removal/teardown, including a retained hero across a seamless crossing and release across a discrete warp. Adapter tests assert both initial and tile releases, idempotent destruction, shared-root preservation, and no post-destroy admission.
- Existing sparse-residency, seam presentation, Go render, and Lua hardening suites must pass on the submitted commit.

## Evidence status

This is a Dark Magic engine lifecycle invariant inferred from `assetPreloader`, `chunked_map`, `game_world`, and issue #642. It makes no claim about original Diablo II 1.13c or 1.14d preload internals. The visual seam target from #641 remains unchanged.
