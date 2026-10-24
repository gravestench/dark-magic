# Issue #642: Release unfinished render preload jobs after scene teardown

## Intent

A sparse outdoor map adapter can close while its initial or tile decode job is still running. The native preloader currently rejects release of unfinished jobs, and the Lua adapter drops the job identifiers at teardown. Completed bookkeeping then remains in the shared preloader. Release the bookkeeping without touching outdoor seam geometry, gameplay authority, or native texture ownership.

## Goals and boundary

- `engine.render/v1.preload_release(job)` accepts a valid unfinished job exactly once, makes it unavailable to callers immediately, and removes its record after every worker in that job finishes. It still releases completed and empty jobs immediately. Unknown or already released IDs return false.
- Worker decoding may finish after release. The preloader keeps accurate in-flight counts until completion, records no late status for a released job, and cannot panic or resurrect a released record when completion races release.
- `chunked_map.destroy` releases both its initial job and every tile job regardless of completion, drops its identifiers, and leaves no tile node to admit after teardown. Repeated create/destroy cycles leave no completed job records.
- Use synthetic controlled jobs for native concurrency tests and Lua adapter tests. No game assets or screenshots are needed.

## Non-goals and constraints

- No worker cancellation, cache eviction, texture disposal, new scene-wide ownership system, or change to rendering/admission of a live adapter.
- Keep decoding on bounded workers and native texture uploads on the renderer thread. `Pending()` continues to count actual unfinished work, including jobs released by their Lua owner, so capture diagnostics still see work in progress.
- Other callers of `preload_release` retain their current behavior for completed jobs; allowing early release is a compatible extension. Scene code must drop an ID after release because status intentionally becomes unavailable.

## Approaches

1. **Chosen: deferred native forget.** Mark a pending job released under the preloader mutex, hide its status, and delete the record when the final worker completes. The map adapter unconditionally releases all IDs in `destroy`. This is a small change at the two existing owners and preserves pending-work diagnostics.
2. **Worker cancellation.** Propagate cancellation through queued work and asset decoders. This would touch many decode paths and still needs careful completion bookkeeping; #642 does not require stopping CPU work already begun.
3. **Scene-scope registry.** Transfer every job ID to a new scene owner that polls after teardown. This adds a second bookkeeping mechanism and periodic cleanup outside the existing preloader.

## Design and verification

Add a release marker to each native job. `Forget` returns false if the ID is absent or already released; if complete it deletes immediately, otherwise it marks released. `Status` returns absent for a released job. Worker completion updates its own job under the mutex and deletes it when all requests finish and it is released. `Pending` still sums unfinished requests until that point. The existing lock serializes completion and release, so both orders converge on deletion.

In Lua, `chunked_map.destroy` releases every retained job ID without first checking status, nils the initial ID, clears the tile-job map, and destroys existing nodes. The native preloader handles safe eventual removal. A destroyed adapter is never updated or used to admit tiles.

Tests first establish failure on current code: a blocked native worker is released before completion, then allowed to finish; status vanishes immediately and the job map becomes empty. Repeat the cycle and race release against completion. A Lua mock keeps an unfinished initial and tile job while `destroy` runs; it must receive both releases, clear ownership, preserve a shared camera root, and create no tile node afterward. Existing sparse-residency and render tests must continue to pass.

## Evidence status

This is a Dark Magic engine lifecycle invariant inferred from the current `assetPreloader` and `chunked_map` code and issue #642. It makes no claim about original Diablo II 1.13c or 1.14d preload internals. The visual seam target from #641 remains unchanged.
