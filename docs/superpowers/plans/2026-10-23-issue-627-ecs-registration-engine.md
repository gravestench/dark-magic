# Lua ECS Registration Engine Ownership Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Release every Lua-registered ECS system from the engine where it registered, even after connected presentation rebinds the capability.

**Architecture:** `registerSystem` captures the current engine once and uses it for registration, scope cleanup, and failed scope transfer rollback. The test uses two real engines and the same system ID to detect silent removal from the wrong one. The hosted client test stops suppressing the #627 cleanup error.

**Tech Stack:** Go, gopher-lua, Dark Magic ECS runtime, owned 1.14d client fixture.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-627-ecs-registration-engine-design.md`

## Global Constraints

- Linked issue: #627. Keep production changes in `internal/runtime/lua/ecs_module.go`; change only the focused Lua test and the hosted fixture's #627 suppression outside it.
- Do not change engine scheduling, bridge movement, or pointer-route issue #634.
- Do not commit Diablo II game data, captures, credentials, or private saves.
- After implementation, use independent correctness and formatting/readability subagent reviews; the style reviewer must use `refactor-for-readability`.

## Review Focus

- If both engines have the same system ID, closing the original scope must remove only the original registration; the new unit regression asserts both engine lists.
- If the replica has no matching system, closing the original scope must still succeed; the unit regression also covers the no-replica-system state before registering on the replica.
- A system registered after `SetEngine` must belong to the replica; closing its own scope must clear only the replica.
- If `scope.Add` rejects an already closed scope, immediate rollback must use the captured registration engine; inspect this line in review, and keep the existing rollback behavior.
- A hosted `game_world` scene close must surface any ECS unregister error; remove the suppression and run a connected bridge fixture with owned data.

---

### Task 1: Pin registration-engine ownership

**Files:**
- Modify: `internal/runtime/lua/ecs_module_test.go`
- Modify: `internal/runtime/lua/ecs_module.go`

**Interfaces:**
- Consumes: `NewECSCapability`, `SetEngine`, `Runtime.RunScoped`, `Scope.Close`, and `gameecs.Engine.Systems`.
- Produces: the existing `registerSystem` behavior with registration-engine lifetime ownership; no API change.

- [ ] **Step 1: Write `TestLuaECSSystemsCloseOnRegistrationEngineAfterRebind`.** Use one runtime and capability, register `scene.follow` in an original scope, rebind to a second real engine, register `scene.follow` in a replica scope, close the original scope, and assert the original `Systems()` is empty while the replica retains `scene.follow`; close the replica scope and assert it is empty. Check the original scope also closes cleanly before the replica registration, using a separate subcase or registration pair so the missing-ID path is exercised.
- [ ] **Step 2: Verify red.** Run `go test -count=1 ./internal/runtime/lua -run '^TestLuaECSSystemsCloseOnRegistrationEngineAfterRebind$'`; it must fail on wrong-engine release, not test setup.
- [ ] **Step 3: Make the minimal production edit.** Capture `registrationEngine := capability.engine` immediately before `Register` and use it for `Register`, the scope release closure, and the `scope.Add` failure rollback. Do not alter `SetEngine` or system updates.
- [ ] **Step 4: Verify green.** Run the focused test and `go test -count=1 ./internal/runtime/lua`; both must pass.
- [ ] **Step 5: Commit** the focused test and production fix with an issue-scoped message.

### Task 2: Enforce hosted scene cleanup

**Files:**
- Modify: `internal/app/clientapp/town_bridge_pointer_real_test.go`

**Interfaces:**
- Consumes: `newBridgePointerClient` test cleanup and the fixed Lua registration owner.
- Produces: an owned-data hosted client check that fails if `app.scenes.Close` reports any error.

- [ ] **Step 1: Remove the #627 special case.** In `newBridgePointerClient` cleanup, report every scene-close error with `t.Errorf`, and remove the now-unused `strings` import. Do not change movement assertions or the separate #626 network cleanup path.
- [ ] **Step 2: Run a hosted owned-data check.** Set `MPQ_DIRECTORY` to the operator's licensed 1.14d directory and run `go test -v -count=1 ./internal/app/clientapp -run '^TestTownBridgePointerBeyondEastLipConnected$'`. Confirm `=== RUN` and `--- PASS` appear with no `--- SKIP` or camera-follow unregister message; if assets are unavailable or #634 manifests, report that separately rather than weaken cleanup.
- [ ] **Step 3: Commit** the hosted cleanup assertion with the spec and plan.

### Task 3: Review and integrate

**Files:**
- Create: PR body using `.github/PULL_REQUEST_TEMPLATE.md`.

**Interfaces:**
- Consumes: Task 1 and 2 commits and exact test outputs.
- Produces: reviewed PR closing #627, merged only after all CI jobs pass.

- [ ] **Step 1: Run broad checks.** On final HEAD run `go test -count=1 ./...`, `make architecture`, `make test-lua-hardening`, `make build-client-backends`, `go test -race -count=1 ./internal/runtime/lua -run '^TestLuaECSSystemsCloseOnRegistrationEngineAfterRebind$'`, `GOLANGCI_LINT=/private/tmp/golangci-lint-2.12.2-darwin-arm64/golangci-lint make lint`, `test -z "$(gofmt -l internal/runtime/lua/ecs_module.go internal/runtime/lua/ecs_module_test.go internal/app/clientapp/town_bridge_pointer_real_test.go)"`, and `git diff origin/main...HEAD --check`. Record failures by name.
- [ ] **Step 2: Obtain independent correctness and `refactor-for-readability` style reviews.** Resolve concrete in-scope feedback and rerun every check claimed in the PR on final HEAD.
- [ ] **Step 3: Write every PR-template section, push, create/attach the PR, and wait for all CI jobs.** Merge after all required jobs pass, then fast-forward the clean primary `main` checkout.
