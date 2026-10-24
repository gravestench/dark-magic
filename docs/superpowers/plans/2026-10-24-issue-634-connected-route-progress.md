# Issue #634 Connected Route Progress Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make connected pointer routes advance and stop according to authenticated local movement even when the displayed player position lags.

**Architecture:** The existing `MovementSource` keeps using the presentation ECS to find paths and read collision radius. Connected reconciliation supplies its latest authenticated HUD owner, level, and position as a guarded route-progress sample. The source consumes waypoints from that sample and emits a targetless stop when it completes a connected route; offline routing keeps its current behavior.

**Tech Stack:** Go movement adapter and `clientapp` reconciliation, owned 1.14d bridge integration tests.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-634-connected-route-progress-design.md`

## Global Constraints

- Issue #634 only: connected pointer-route progress at the town bridge; no collision, seam, host arrival, network protocol, or general movement UI changes.
- Preserve `0.3` adapter waypoint tolerance and `0.2` host arrival tolerance.
- Never commit owned Diablo II game data, captures, or private saves.

## Review Focus

- A stale authenticated sample from another player must not advance the active player's route, while an unchanged-owner refresh must retain the sample; covered in Task 1's owner tests.
- A sample from another level must not advance a route after a seam; covered in Task 1's level-mismatch test.
- A presented position close to a corner while authority is still distant must not skip that corner; covered in Task 1's visual-near counterexample.
- Completing a route from authenticated progress must not pull the host backward to the lagging displayed position; covered in Task 1's targetless-stop test.
- Rebinding an engine or different map must discard the previous sample, while reinstalling the same map must preserve an active route; covered in Task 1's binding tests.

---

### Task 1: Guarded authenticated waypoint progress

**Files:**
- Modify: `internal/mod/d2legacy/adapter/movement/source.go`
- Modify: `internal/mod/d2legacy/adapter/movement/path.go`
- Test: `internal/mod/d2legacy/adapter/movement/movement_test.go`

**Interfaces:**
- Produces: `(*MovementSource).SetAuthenticatedPosition(player string, level int, position gameworld.Point)` for the connected reconciliation caller. It stores a sample under the source mutex; owner, engine, and map changes invalidate it.
- Produces: `pathWaypoint` selecting the same-owner, same-level sample only for waypoint consumption. When this sample completes the route, return no target; without it, retain the existing displayed-position stationary target.

- [ ] **Step 1: Write failing behavior tests.** Add `TestMovementSourceAdvancesConnectedWaypointFromAuthenticatedPosition` with displayed `(250.667,84)`, authenticated `(250.875,84)`, path through `(251,84)` and `(252,84)`, and assert emitted target `(252,84)`. Add `TestMovementSourceStopsCompletedConnectedRouteWithoutReverseTarget` with authenticated progress at the final waypoint and lagging display; assert nil target and zero axes. Add table cases for visually near but authority far, wrong owner, wrong level, and absent sample; assert the appropriate existing waypoint remains. Add rebind cases for `SetPlayer`, `SetEngine`, and `SetNavigation` with changed/same map pointer, including an unchanged-owner refresh, asserting subsequent commands use only valid progress.
- [ ] **Step 2: Run RED.** `go test -count=1 ./internal/mod/d2legacy/adapter/movement -run 'TestMovementSource(AdvancesConnectedWaypointFromAuthenticatedPosition|StopsCompletedConnectedRouteWithoutReverseTarget|GuardsAuthenticatedPosition)'` must fail for the missing behavior or API.
- [ ] **Step 3: Implement minimal adapter behavior.** Add `SetAuthenticatedPosition` and guarded sample invalidation in `source.go`; use it only to consume reached waypoints and choose targetless connected completion in `path.go`. Keep pathfinder start, collider, route geometry, and offline behavior unchanged.
- [ ] **Step 4: Run GREEN.** Run the RED command, then `go test -count=1 ./internal/mod/d2legacy/adapter/movement`; both must pass.
- [ ] **Step 5: Commit.** Commit adapter code and behavior tests as `fix: advance connected routes from authenticated progress`.

### Task 2: Supply authenticated progress from connected reconciliation

**Files:**
- Modify: `internal/app/clientapp/client_reconciliation.go`
- Test: `internal/app/clientapp/client_route_progress_test.go` for a deterministic HUD-to-command wiring assertion.
- Test: `internal/mod/d2legacy/adapter/movement/movement_test.go` for the offline no-HUD completion characterization.

**Interfaces:**
- Consumes: `(*MovementSource).SetAuthenticatedPosition(player string, level int, position gameworld.Point)` from Task 1, using the validated `hud` passed to `presentLocalPrediction`.

- [ ] **Step 1: Confirm the existing RED evidence and add wiring proof.** Record the earlier owned full-package failure at presented `(250.667,84)`, authenticated `(250.875,84)`, with accepted/pending waypoint `(251,84)` while clicking `(252,84)`. Add `TestConnectedReconciliationFeedsAuthenticatedRouteProgress` with a small open map and accepted route planned from `(10,10)` to `(13,10)`, then force displayed `(11.667,10)` and authenticated HUD `(11.875,10)` near waypoint `(12,10)`; it must emit target `(13,10)`. Temporarily remove the setter call and verify this test emits `(12,10)`; restore it before GREEN. Add a no-HUD completed-route case asserting the old stationary target.
- [ ] **Step 2: Wire the authenticated HUD sample.** In `presentLocalPrediction`, supply owner, level, and position to `movementSource` before applying the smoothed `presented` position to the replica. Do not pass predicted or interpolated coordinates.
- [ ] **Step 3: Verify the owned route.** With `MPQ_DIRECTORY` and `DARK_MAGIC_TEST_MPQ_DIRECTORY` pointing to owned 1.14d data, run `go test -count=5 ./internal/app/clientapp -run '^TestTownBridgePointerFromEntryConnected$'`, then `go test -count=3 ./internal/app/clientapp`. Both must pass, including the existing offline route and round-trip tests in package order. Report any unrelated failures by name.
- [ ] **Step 4: Commit.** Commit the reconciliation wiring and any narrowly justified acceptance-test change as `fix: feed connected route progress from authenticated HUD`.

### Task 3: Verify and prepare the issue PR

**References:**
- Read: `.github/PULL_REQUEST_TEMPLATE.md` for the PR body; do not change the file.
- Update: issue #634 with the demonstrated cause and exact verification evidence.

**Interfaces:**
- Consumes: committed Task 1 and Task 2 behavior.

- [ ] **Step 1: Run final checks on the candidate commit.** `go test -count=1 ./...`, `make architecture`, `make test-race`, `make lint`, `make test-lua-hardening`, `make build-client-backends`, and `git diff --check`; preserve exact outputs and any inability to run.
- [ ] **Step 2: Review the whole diff with independent correctness, evidence, and formatting/style subagents; the style reviewer must use `refactor-for-readability`.** Fix any important finding with a failing regression followed by a passing check.
- [ ] **Step 3: Prepare the PR description.** Fill every template section with scope, the observed route-progress cause, exact checks and results on submitted HEAD, licensed-data local execution versus CI skips, remaining native/manual uncertainty, and AI/subagent use. Link `Closes #634` only if acceptance is satisfied.
- [ ] **Step 4: Submit, wait for CI, and integrate.** Push branch, create and attach PR, wait for required checks on its exact head, merge only when green, then fast-forward the user's clean `~/src/dark-magic` main checkout and report the merged commit.
