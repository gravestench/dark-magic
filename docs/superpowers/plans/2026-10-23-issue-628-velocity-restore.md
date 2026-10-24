# Velocity Movement Restore Safety Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep the native velocity integrator bound to the authoritative ECS world after direct restore and Session rollback, without changing movement behavior.

**Architecture:** `Engine.Restore` replaces its Akara world while preserving registered system callbacks. The velocity callback will compare the engine's current world with its cached world at each batch, re-resolve its three named stores only after a replacement, and report missing required components as errors. The offline recovery path can then use this same behavior instead of re-registering the system.

**Tech Stack:** Go, Akara ECS, `internal/game/ecs`, `internal/game/session`, `internal/game/world`.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-628-velocity-restore-design.md`

## Global Constraints

- Linked issue: #628. Change only native velocity restore safety, its tests, and the offline workaround it replaces.
- Preserve movement speed, collision, monster population, and the ECS update API.
- No new dependencies or committed game data.
- User delegated design and final review to independent subagents; the final style reviewer must use `refactor-for-readability`.

## Review Focus

- Existing mover after one restore must advance in the **new** world, not merely return a successful update; Task 1 test with `1e-9` tolerance.
- Existing mover after a second restore must advance again; Task 1 test.
- Mover created only after restore must not read the old world's absent entity; Task 1 test.
- Missing required component passed to the movement helper must yield an error rather than panic; Task 1 direct helper test.
- Late authoritative input must restore and replay without losing subsequent movement; Task 2 test.

---

### Task 1: Bind velocity stores to the current ECS world

**Files:**
- Modify: `internal/game/world/velocity_movement.go`
- Test: `internal/game/world/velocity_movement_test.go`

**Interfaces:**
- Consumes: `(*gameecs.Engine).World() *akara.World`, `akara.GetDynamicStore(*akara.World, string)`, `(*gameecs.Engine).Restore(gameecs.Snapshot) error`.
- Produces: `velocityMovementUpdate(engine *gameecs.Engine, paths VelocityPathFinder, components VelocityComponents, stores velocityStores) gameecs.UpdateFunc`; caller passes the registration engine and schema names.

- [ ] **Step 1: Write direct restore tests.** Create fixture mover at `(10, 10)` with velocity `(1, 0)` and `40ms` step. Snapshot and restore, then assert new-world `x` is within `1e-9` of `10.04` after one update; snapshot, restore, update again and assert `x` is within `1e-9` of `10.08`. In a separate case, snapshot an empty registered world, restore, create the same mover only in the new world, update, and assert new-world `x` is within `1e-9` of `10.04`. Call `moveVelocityEntity` directly with a store bundle whose entity has no position component and assert a named component error instead of a panic; removing position before `Update` would remove the entity from Akara's subscription and not exercise this guard.
- [ ] **Step 2: Write the Session rollback regression.** Create the fixture mover before `gamesession.New` captures its initial snapshot. Register a valid no-op ordinary command handler, run two `40ms` ticks, submit a tick-1 sequence-1 command through `SubmitNetwork`, assert `engine.World()` changed, then `Step` once and assert new-world mover `x` is within `1e-9` of `10.12`. Close the Session in cleanup.
- [ ] **Step 3: Run red tests separately.** Run each named restore, missing-component, and Session rollback regression with `go test -count=1 ./internal/game/world -run '^<TestName>$'`. Record each expected failure independently: the new-world case intentionally panics on current code and would abort a combined process before the other tests run.
- [ ] **Step 4: Implement restore-aware callback.** Change `velocityMovementUpdate` to the stated signature. Cache the world alongside stores; on a different `engine.World()` pointer, resolve position, velocity and collider stores by their configured names once, return a named error if any is absent, then use the refreshed stores for the batch. In `moveVelocityEntity`, check each `Get(entity)` result before dereferencing and return a named error for missing required components. Preserve existing field mutation order and collision path.
- [ ] **Step 5: Run green tests.** `go test -count=1 ./internal/game/world ./internal/game/ecs ./internal/game/session`; expected pass.
- [ ] **Step 6: Commit.** `git add internal/game/world/velocity_movement.go internal/game/world/velocity_movement_test.go && git commit -m 'fix(world): rebind velocity stores after ECS restore'`.

### Task 2: Remove offline re-registration

**Files:**
- Modify: `internal/app/clientapp/entry_world_recovery.go`

**Interfaces:**
- Consumes: Task 1's restore-aware callback through `RegisterVelocityMovement` and its Session rollback regression.
- Produces: offline recovery path that calls `RestoreRecoveryCheckpoint` then `syncActiveWorldFromPlayer` without system re-registration.

- [ ] **Step 1: Run the existing offline recovery tests with the workaround still present.** Record baseline results from the owned-data clientapp recovery tests.
- [ ] **Step 2: Remove the offline-only workaround.** Delete `Unregister(VelocityMovementSystemID)` and `registerOfflineMovementSystem()` after `RestoreRecoveryCheckpoint`; update the function comment. Do not alter topology preflight or active-world sync.
- [ ] **Step 3: Run focused and owned-data checks.** `go test -count=1 ./internal/game/world ./internal/game/session`; run clientapp's offline recovery and connected Cold Plains population tests with owned 1.14d data, then `go test -count=1 ./internal/app/clientapp` and record asset skips precisely.
- [ ] **Step 4: Commit.** `git add internal/app/clientapp/entry_world_recovery.go && git commit -m 'fix(clientapp): use restore-safe native movement'`.

### Task 3: Review and submit the issue-scoped fix

**Files:**
- Add: `docs/superpowers/specs/2026-10-23-issue-628-velocity-restore-design.md`
- Add: `docs/superpowers/plans/2026-10-23-issue-628-velocity-restore.md`
- Create: pull request description using `.github/PULL_REQUEST_TEMPLATE.md`.

**Interfaces:**
- Consumes: Task 1 and 2 commits and their exact verification results.
- Produces: reviewed PR referencing #628, merged only after CI passes. Keep #628 open because its connected owned-data acceptance check fails on clean main as well; #630 tracks that separate test instability.

- [ ] **Step 1: Run independent correctness and formatting/style reviews.** Delegate correctness/evidence and `refactor-for-readability` style reviews to separate subagents; resolve material feedback within issue scope.
- [ ] **Step 2: Run final verification on the submitted commit.** Run `gofmt` check, `git diff --check`, owned-data affected tests, repository CI gates available locally, and record exact outcomes, including data-gated skips.
- [ ] **Step 3: Commit the approved spec and plan, complete every PR-template section with `Refs #628` and #630's connected-test evidence, push the branch, create and attach the PR.** Re-run any required checks if the submitted commit changes.
- [ ] **Step 4: Wait for CI, fix any failures, and merge once all required checks pass.** Update `~/src/dark-magic` main to the merge commit if it remains clean.
