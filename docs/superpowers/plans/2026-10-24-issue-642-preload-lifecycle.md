# Issue #642 Preload Job Teardown Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Release every outdoor game scene preload handle safely, including jobs still running when an adapter or scene closes.

**Architecture:** `assetPreloader` remains the sole job registry. A pending release marks a job invisible to callers and deletes it after final worker completion; the existing mutex orders completion and release. The first-party Lua map adapter and `game_world` scene release their own handles on completion and teardown, while retained hero art survives a seamless crossing.

**Tech Stack:** Go 1.24, Gopher Lua, existing `engine.render/v1`, first-party `d2legacy` Lua tests, synthetic `fs.FS` fixtures.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-642-preload-lifecycle-design.md`

## Global Constraints

- Linked issue is #642. Change only preload-job lifetime in the outdoor `game_world` scene and its sparse map adapters; do not alter seam geometry, zone admission, asset decoding, native texture ownership, or gameplay authority.
- `Pending()` counts released but unfinished work until the worker finishes; `Status()` hides released IDs immediately; release returns true exactly once for a known job.
- Background decode workers remain bounded. Tests and commits use synthetic content only; never commit Diablo II game data or captures.
- Treat renderer nodes and preload records as separate resources. A seamless crossing retains hero art; a warp to a non-retained component releases the old pending hero handle.

## Review Focus

- Release before a worker enters `load`: the worker must finish without panic and delete the released record. Task 1 tests this with a blocked worker.
- Release racing the final completion lock: exactly one release succeeds and no record remains after either order. Task 1 tests both orders under `-race`.
- Destroy with an unfinished initial map job or an unfinished tile job: every adapter-owned handle releases; a later update cannot admit a node. Task 2 tests these states separately, plus a double-handle teardown fixture.
- Scene entity disappears before its art finishes: monster, missile, overlay, and hero handles release on removal/teardown. Task 3 tests tracked jobs.
- Scene entity finishes art before disappearing: completion releases the handle rather than losing the ID; repeated scene cycles remain bounded. Task 3 tests this and Task 4 verifies suites.

---

### Task 1: Native Pending Release

**Files:**
- Modify: `internal/runtime/lua/render_preload.go`
- Modify: `internal/runtime/lua/render_preload_test.go`
- Modify: `internal/runtime/lua/render_module.go` (command help only)

**Interfaces:**
- Consumes: existing `assetPreloader.Start`, `Status`, `Pending`, `Forget` and `engine.render/v1.preload_release(job)`.
- Produces: `Forget(id uint64) bool` accepts one release of a pending job; `Status` hides released jobs; worker completion auto-deletes released records; `Pending` continues to count work in flight.

- [ ] **Step 1: Write failing native tests.** Replace the old `TestPreloaderForgetsOnlyCompletedJobs` expectation. Add `TestPreloaderReleasePendingJobAfterWorkerCompletes` with a controlled pre-load barrier: fill `p.work` before `Start`, release while `Pending()==1` before freeing a worker slot, assert first release true, second false, and `Status` absent; then free the slot and assert `Pending()==0` and the locked job-map length is zero. Add completed/empty/unknown cases and a multi-request case that must retain the record until the final completion. Hold `p.mu` for every direct job-map length read while workers may run.
- [ ] **Step 2: Run red tests.** `go test ./internal/runtime/lua -run 'TestPreloader(Release|Forgets)' -count=1` must fail on current pending-release behavior.
- [ ] **Step 3: Implement native release.** Add a release marker to `assetPreloadJob`. Under `p.mu`, make `Forget` mark pending or delete completed jobs, make `Status` return absent for released jobs, and delete a released job after its final worker completion. Keep `Pending` unchanged for running work. Change render module help to describe release of pending or completed jobs.
- [ ] **Step 4: Verify both lock orders.** Run the focused tests and `go test -race ./internal/runtime/lua -run 'TestPreloader(Release|Forgets)' -count=1`. Use a bounded wait and an explicit completion barrier in tests; do not use random timing as the only race evidence.
- [ ] **Step 5: Commit.** Stage only the three files and commit `fix: release pending native preload jobs after completion`.

### Task 2: Sparse Adapter Ownership and End-to-End Teardown

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/presentation/chunked_map.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/presentation/chunked_map_test.lua`
- Create: `internal/runtime/lua/render_preload_lua_test.go`

**Interfaces:**
- Consumes: Task 1 `engine.render/v1.preload_release(job)` accepts unfinished jobs; `ContentRequire`, `NewRenderCapability`, and `RunScoped` exist for the Go/Lua integration harness.
- Produces: `chunked_map.destroy(state)` releases all `state.job`/`state.pending` IDs and clears ownership; `chunked_map.update` cannot admit after destroy.

- [ ] **Step 1: Write failing adapter tests.** Add separate Lua cases for an unfinished initial job and an unfinished tile job, plus an explicitly seeded double-handle teardown fixture. Destroy twice and assert each ID released once, `state.job==nil`, `state.pending` empty, shared camera root alive, and a later `update` creates no node. Keep the existing sparse residency case intact.
- [ ] **Step 2: Write failing real bridge test.** `TestChunkedMapDestroyReleasesBlockedNativePreload` registers `ContentRequire(content.D2Legacy(), "lua")` and the real render capability. A synthetic `fs.FS.Open` gate blocks `ds1_chunks` for `map.ds1`; wait with a bounded timeout until the worker reaches the gate. In a scoped Lua call, create a non-shared adapter from the real first-party module and retain its job ID. Destroy it through Lua while the gate is closed, assert native `Status(id)` absent and `Pending()==1`, then open the gate and wait until the locked job-map length is zero and `Pending()==0`. Re-arm the gate and repeat create/destroy, checking the native map after every cycle. Register `t.Cleanup` to unblock each gate even after failure or timeout, and bound completion waits. Hold `p.mu` for every direct job-map length read while workers may run. Missing synthetic DS1 data after the gate opens is acceptable; the lifecycle is the assertion.
- [ ] **Step 3: Run red tests.** Run `go test ./internal/runtime/lua -run TestChunkedMapDestroyReleasesBlockedNativePreload -count=1` and `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/presentation/chunked_map' -count=1`; verify they fail on pending cleanup/post-destroy behavior.
- [ ] **Step 4: Implement adapter cleanup.** Set a destroyed flag before clearing nodes. Release initial and tile jobs without checking status, clear IDs, keep the shared root alive, and make `update` return before preload/admission after destruction. Make repeated destroy harmless.
- [ ] **Step 5: Verify and commit.** Re-run focused Go and Lua tests, then stage only the three files and commit `fix: release sparse map jobs on teardown`.

### Task 3: Outdoor Scene Job Ownership

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/screens/game_world.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/screens/game_world_seams_test.lua` or create a focused neighboring `game_world_preload_test.lua` if the fixture would become unwieldy.

**Interfaces:**
- Consumes: Task 1 early-release semantics and Task 2 adapter cleanup.
- Produces: monster, missile, overlay, and hero preload handles release before nil/removal/scene destruction; a pending hero handle remains attached during seamless retained-hero crossing and releases at non-retained warp.

- [ ] **Step 1: Write failing tracked-job scene tests.** Extend the existing synthetic scene fixture to record `preload` IDs and `preload_release` calls. Assert pending monster/missile/overlay/hero IDs release on entity removal and scene close; completed IDs release before nil. Assert retained hero ID survives a 1↔2 seam crossing, but a non-retained level-4 warp releases it. Check each handle is released once and no removed node is drawn afterward.
- [ ] **Step 2: Run red test.** `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/screens/game_world' -count=1` must fail at the missing release assertions.
- [ ] **Step 3: Implement narrow ownership cleanup.** Reuse small local helpers where the same monster/missile removal occurs both on level change and ordinary disappearance. Release IDs at completion, removal, and `destroy`, nil them after release, and release the old hero job only when its node is not retained. Keep all render recipes, scene transitions, and camera logic unchanged.
- [ ] **Step 4: Verify and commit.** Run the focused Lua scene suite and `make test-lua-format test-lua-syntax`, then stage only the scene and test files and commit `fix: release outdoor scene preload handles`.

### Task 4: Final Evidence, Review, and PR

**Files:**
- Update only issue #642 research or ownership text if code evidence requires it; the committed spec already marks original-game behavior as unclaimed.
- Prepare `.github/PULL_REQUEST_TEMPLATE.md` sections in a temporary PR body file before submission.

- [ ] **Step 1: Run exact-head checks.** `make test`, `make architecture`, `make test-lua-hardening`, `make test-race`, `make build-client-backends`, pinned `make lint`, focused Go/Lua tests, and `git diff --check origin/main...HEAD`. Report owned-data skips accurately; #642's synthetic tests require no MPQ.
- [ ] **Step 2: Independent review.** Request whole-branch correctness, issue/evidence, and formatting/readability reviews. The style reviewer must use `refactor-for-readability`. Fix and re-review every blocking finding within #642.
- [ ] **Step 3: Submit and merge.** Fill every PR template section with exact checks on the submitted commit, commit/push the branch, attach the PR, await green CI, and merge using the user's standing authorization. Fast-forward the clean main checkout and confirm issue #642 closes.
