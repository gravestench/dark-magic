# Bridge-Lip Pointer Route Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Route a normal outside-map click from Rogue Encampment's reachable bridge shoulder into the published safe exit strip and cross to Blood Moor.

**Architecture:** First-party Lua corrects only pointer targets near the prepared town opening. Authoritative trigger, collision, and Go pathfinding stay unchanged. An owned-data test drives the production scene, camera, input, movement, and transition paths.

**Tech Stack:** Go 1.25 tests, first-party Lua `d2legacy`, owned 1.14d MPQs.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-636-bridge-lip-design.md`

## Global Constraints

- Work only on #636; #634 connected return routing and #627 cleanup are separate.
- Keep the town trigger Y82..86, collision map, generated seams, and other borders unchanged.
- The two-subtile pointer margin is inferred Dark Magic behavior, not a verified retail trigger.
- Never commit owned game data, captures, credentials, or saves.

## Review Focus

- Wrong edge, in-map click, or non-town level: leave target unchanged (Task 1 Lua cases).
- More than two lanes outside the strip: leave Y79/Y89 unchanged (Task 1 Lua cases).
- Oblique ray near the edge but far at the trigger plane: leave target unchanged (Task 1 Lua case).
- Already valid strip ray: preserve exact target (Task 1 Lua case).
- Camera lag or canvas clamp: use live camera and effective anchor, then assert reverse projection (Task 1 owned test).

---

### Task 1: Red tests, narrow correction, green tests

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/world_transition_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/gameplay/seam_geometry.lua`
- Modify: `internal/app/clientapp/town_bridge_pointer_real_test.go`

**Interfaces:**
- Consumes and preserves: `M.pointer_target(level, hero_x, hero_y, target_x, target_y, width, height)`.
- Produces: `TestTownBridgeShoulderPointerCrossesOffline` and `TestTownBridgeShoulderPointerCrossesConnected`.

- [ ] **Step 1: Write Lua RED.** Add a production-shaped east seam `(275,84)` with strip Y82..86 to the existing test fixture. With source map 280×200, require exact `(269,80) -> (281,80)` to target `(275,82)` and Y81 to target Y82; lower Y88 targets Y86. Pin Y79/Y89 unchanged, plus existing in-strip, oblique, other-border, in-map, and non-town cases.
- [ ] **Step 2: Run Lua RED.** Run `make test-lua`; confirm the old helper returns `(281,80)` for the first shoulder case.
- [ ] **Step 3: Write owned-data scene RED.** Stage the controlled player at `(269,80)` from the seam fixture. Read the presentation camera ECS position and compute its effective anchor using the production canvas/clamp rule. Project `(281,80)` to an integer logical cursor above the HUD, reverse-project it, and assert its collision cell is outside the east edge near lane 80; log the actual continuous coordinate. Log camera, anchor, independently calculated goals, radius-one path results, authoritative position/level, and strip match. Assert the raw projected goal is rejected and `(275,82)` is reachable. Actual goal selection is asserted by the Lua unit test; this scene test must observe movement into the strip and offline/connected authority committing level 2 without reading private mailbox state.
- [ ] **Step 4: Run scene RED.** `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test -v -count=1 ./internal/app/clientapp -run '^TestTownBridgeShoulderPointerCrosses(Offline|Connected)$'`. Record both old-code stops if feasible: the raw outside goal has no route and authority remains in town.
- [ ] **Step 5: Implement one Lua correction.** In `pointer_target`'s strip branch, require both the boundary lane and trigger-plane lane within `LaneMin-2..LaneMax+2`; clamp only the returned trigger-plane lane into the published strip. Preserve other guards and `matching_seam`.
- [ ] **Step 6: Run GREEN.** Run the focused Lua suite and both owned scene cases; require `(275,82)` route acceptance and canonical level 2, including connected arrival. Commit the localized Lua and test changes.

### Task 2: Verify, review, and integrate

**Files:**
- Complete the PR template and update issue #636 only from observed results.

**Interfaces:**
- Consumes: Task 1's unchanged Lua API and executable acceptance evidence.
- Produces: a reviewable issue-linked PR, merged only after CI passes.

- [ ] **Step 1: Run final checks on the current branch head.** `go test -count=1 ./...`, `make architecture`, `make test-race`, `make test-lua-hardening`, `stylua --check` on changed Lua files, and owned-data `TestTownBridge` plus `TestOwnedTownExitCorridorsReachVisibleBoundary`. State exact results and CI licensed-data skips.
- [ ] **Step 2: Obtain independent correctness, evidence, and formatting/readability reviews.** The style subagent uses `refactor-for-readability`; resolve blocking findings and rerun claimed checks after any edit.
- [ ] **Step 3: Rebase on current `origin/main`, rerun claimed checks on the submitted commit, fill every PR-template section, push, attach the PR, wait for all required CI jobs, and merge manually.** Fast-forward `~/src/dark-magic` after merge; never use early auto-merge.
