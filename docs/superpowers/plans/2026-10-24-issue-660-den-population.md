# Den of Evil Ordinary Population Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Put ordinary, authoritative hostile residents in the generated Den of Evil on first entry, with safe placement and stable recovery.

**Architecture:** Append validated level-8 cave geometry to the existing Go population command, then let the existing Lua population plan choose and activate residents. A room point is published only when the player can reach it from the validated Den arrival; Lua checks each monster's full footprint and both endpoint clearances before creating it.

**Tech Stack:** Go entry-world adapter and tests; embedded Lua d2legacy population owner; owned Expansion 1.14d MPQs; existing offline and connected client fixtures.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-660-den-population-design.md`

## Global Constraints

- Issue #660 is the only code scope. Preserve three visible outdoor joins and the click-activated Den warp.
- Lua remains the sole monster-selection, density, activation, and durable-resident authority. Keep `d2legacy.population.plan/v6` and current-binary legacy plan readability.
- Owned 1.14d rows are evidence; existing density/group/RNG policy is provisional Dark Magic policy. Do not commit game data or captures.
- Publish only `[2,3,4,8]` when the Den exists; old `[2]`, `[2,3]`, and `[2,3,4]` payloads remain admitted.

## Review Focus

- A seed with an open but disconnected center pocket must publish arrival-reachable points; cover owned seed 42 in Task 1.
- A partial or mismatched cave pair must reject the entire command, with no level-8-only fallback; cover in Task 1.
- A malformed level-8 exclusion (missing, zero, nonfinite, or too small) must be rejected before state changes; cover in Task 2.
- A level-8 command following a truncated outdoor prefix must be rejected, while old complete prefixes still work; cover in Task 2.
- Leaving and returning after a Den monster dies must preserve that death and its identity; cover in Task 3.

---

### Task 1: Publish validated and reachable Den population geometry

**Files:**
- Modify: `internal/mod/d2legacy/adapter/entryworld/population.go`
- Modify: `internal/mod/d2legacy/adapter/entryworld/cave_warp.go` (share its bounded arrival-component flood)
- Test: `internal/mod/d2legacy/adapter/entryworld/population_test.go`
- Test: `internal/mod/d2legacy/adapter/entryworld/den_of_evil_real_test.go`
- Update existing expectations: `internal/mod/d2legacy/adapter/entryworld/stony_field_real_test.go`, `cold_plains_real_test.go` (current v4 topology only)

**Interfaces:**
- Consumes: `Prepared.CaveWarp`, `Prepared.Zones[8]`, `Prepared.Worlds[8]`, `CaveWarpPair.First.ArrivalX/Y`, `CaveWarpPair.Second.X/Y`, and `Map.WalkableStep` with radius 1.
- Produces: `Prepared.PopulationData(nearby int) (map[string]any,error)` with optional final level-8 zone; its level-8 `exclusions` array has arrival `{x,y,radius:1}` then return source `{x,y,radius:3.5}`. Existing zone payloads remain byte-equivalent.

- [ ] **Step 1: Write failing publisher tests.** `TestPopulationDataDenRequiresCompleteCave` checks correct `[2,3,4,8]` order, exact endpoint exclusions, and rejection of absent/mismatched level-8 zone/map or malformed pair. `TestPopulationDataDenSkipsDisconnectedCenter` makes an open disconnected center plus an arrival-connected ring and asserts every published candidate has a radius-1 path from arrival; its first four candidates must come from the reachable component. `TestDenPopulationOwnedRowsReal` pins level-8 `Levels` fields (`NumMon=3`, `MonDen* = 600`, six candidate fields), candidate `MonStats` ordinary flags, and definitions 83–86 `LvlPrest.Populate=1`. `TestDenPopulationGeometryReal` checks Normal/Nightmare/Hell, all observed chamber variants including seed 42, deterministic room-local open and arrival-reachable candidates, endpoint exclusions, and no topology drift.
- [ ] **Step 2: Run red:** `go test ./internal/mod/d2legacy/adapter/entryworld -run 'TestPopulationDataDen' -count=1` and `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy/adapter/entryworld -run 'TestDenPopulation' -count=1`. Require a clear missing-level-8 failure in each; the owned row assertions themselves should already pass.
- [ ] **Step 3: Implement minimal Go publication.** Keep seam loop unchanged for levels 2–4. Validate cave endpoint levels/dimensions and keyed map/zone before appending level 8. Extract the existing bounded Den flood into `denReachableCells(world *gameworld.Map, arrival gameworld.Point) []bool` and let `denHasReachableInterior` consume it, preserving its admission result. Share the deterministic room-ring scanner; the cave call filters candidate cells through that one radius-1 reachable component and continues until four accepted points or exhaustion. Use actual stamp `Populate` and room links. Publish the two fixed clearance envelopes. Update only entryworld v4 zone-count expectations invalidated by this addition.
- [ ] **Step 4: Run green:** repeat both focused commands and the existing entryworld Stony/Cold owned cases using `DARK_MAGIC_TEST_MPQ_DIRECTORY`. Require no owned-data skips. Client integration awaits Task 2's Lua admission.
- [ ] **Step 5: Commit:** `git add` only the listed files, then `git commit -m 'feat: publish reachable Den population geometry for #660'`.

### Task 2: Admit level 8 and materialize safe residents in the Lua owner

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua`
- Test: `internal/mod/d2legacy/cold_plains_population_test.go` (existing synthetic session fixtures)
- Test: `internal/mod/d2legacy/monster_difficulty_session_test.go` if difficulty-specific synthetic rows are needed
- Test: `internal/app/clientapp/den_of_evil_population_real_test.go` (new pointer-entry case)
- Update existing expectation: `internal/app/clientapp/cold_plains_population_real_test.go` (current v4 plan has four zones)

**Interfaces:**
- Consumes: Task 1's final level-8 zone and ordered `exclusions` array.
- Produces: one `d2legacy.population.plan/v6` with optional level-8 zone; level-qualified `level:8:room:<id>:monster:<n>` residents; `room.population_skip_reason = "no_safe_point"` if an eligible selected room cannot create a safe resident.

- [ ] **Step 1: Write failing command and pointer-entry tests.** Submit `[2,3,4,8]` with valid exclusions, then assert plan installation and a level-8 hostile after a level-8 player enters. Reject `[2,8]`, `[2,3,8]`, duplicate 8, missing or malformed exclusion fields, zero/nonfinite radius, arrival radius below 1, return radius below 3.5, and difficulty mismatch. Assert each rejected command leaves the plan uninstalled. Preserve successful old `[2]`, `[2,3]`, `[2,3,4]` fixtures. In the owned client fixture, use the actual pointer click into Den and assert an eligible authoritative hostile with level-qualified ID is visible there and absent from outdoor projection.
- [ ] **Step 2: Run red:** `go test ./internal/mod/d2legacy -run 'TestDenPopulation' -count=1` and `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/app/clientapp -run 'TestDenOfEvilPopulationEntryReal' -count=1`. Expect rejection of valid level 8 or absence of the Den resident.
- [ ] **Step 3: Implement command admission and streams.** Accept only the four stated orders. Validate the level-8 envelope with finite numbers and required positive minimum clearances, while leaving old zone shapes unchanged. Add a level-8-specific random stream prefix without changing existing level 2–4 streams. Keep plan schema v6.
- [ ] **Step 4: Add a failing placement test.** Place a candidate at each exclusion and another safe candidate; assert only the latter becomes a resident, and an all-unsafe selected room records `population_skip_reason = "no_safe_point"`. Check the spawned collider against actual level-8 collision and the two clearances, including its own radius.
- [ ] **Step 5: Run red:** same focused command; expect an overlap or missing inspectable skip reason.
- [ ] **Step 6: Implement minimal placement.** Extend `safe_point` for cave exclusions, using the existing `overlaps` full-footprint rule. Record a skip reason only for an eligible room whose selected points cannot produce a resident; do not change old outdoor room state. Existing activation/deactivation logic owns Den entities automatically.
- [ ] **Step 7: Run green and regressions:** `go test ./internal/mod/d2legacy -run 'TestDenPopulation|TestColdPlainsPopulation|TestGeneratedMonsterUsesSessionDifficulty' -count=1`; repeat the owned pointer-entry command and existing clientapp Stony/Cold population cases using `MPQ_DIRECTORY`; run repository Lua syntax/format checks. Expect PASS.
- [ ] **Step 8: Commit:** `git add` only the listed files, then `git commit -m 'feat: activate ordinary Den residents for #660'`.

### Task 3: Prove owned pointer journeys and recovery

**Files:**
- Extend: `internal/app/clientapp/den_of_evil_population_real_test.go`
- Extend legacy recovery cases: `internal/app/clientapp/entry_world_recovery_test.go`, `internal/app/clientapp/stony_field_transition_test.go`
- Extend prepared-topology recovery case: `internal/mod/d2legacy/adapter/entryworld/cold_plains_real_test.go`
- Use existing helpers from: `den_of_evil_transition_real_test.go`, `cold_plains_population_real_test.go`, `stony_field_population_real_test.go`
- Modify production files from Tasks 1–2 only if these tests expose a scoped defect.

**Interfaces:**
- Consumes: existing `newDenPointerClient`, `denClick`, `denMoveTo`, `denCheckpoint`, `checkpointPopulationIDs`, and player-scoped world projections. Add a scoped checkpoint inspector for Den collider/health/death fields because `populationResidentIDs` requires a different fixture type.
- Produces: end-to-end evidence for Normal/Nightmare/Hell offline and connected entry/return, reconnect, current-binary checkpoint restore, resident identity/combat/death persistence, and unchanged outdoor joins.

- [ ] **Step 1: Extend the owned pointer-entry test into journeys.** Assert each Den resident's actual collider footprint is safe, arrival-reachable, and clear of both endpoints. Submit production `player.use_skill` attacks against an identified Den resident, as in `hostile_lifecycle_restore_test.go`, until the checkpoint shows damage and then death. Leave, return, reconnect, and restore a current-binary checkpoint; assert that resident's ID, health, and death component are retained. Use a checkpoint-based Den inspector, not the fixture-specific `populationResidentIDs`. Run two independent fresh sessions with the same seed/rules and compare normalized level-8 plan plus resident IDs. Cover seed 42 and all available Den chamber variants without inserting protected fixtures. These assertions characterize the existing generic resident lifecycle in the new Den topology; no production edit is justified if they pass immediately.
- [ ] **Step 2: Run the owned journey checks:** `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/app/clientapp -run 'TestDenOfEvilPopulation' -count=1`. If a new assertion fails, identify the scoped defect before changing code.
- [ ] **Step 3: Resolve only #660 defects exposed by that red test.** Keep geometry/admission changes in Task 1's files and activation changes in Task 2's file; do not alter scene or warp semantics. If a failure belongs to another issue, record it instead of expanding scope.
- [ ] **Step 4: Run green:** same focused owned-data command; require all cases to execute without owned-data skips.
- [ ] **Step 5: Check old recovery shapes.** Extend the existing entry-world v1–v3 and client recovery fixtures to restore current-binary prepared topology plus an existing v6 plan without level 8; assert no level-8 zone or residents are synthesized. Run those focused recovery tests with the owned-data environment each suite requires.
- [ ] **Step 6: Run a no-MPQ check:** `go test ./internal/mod/d2legacy/adapter/entryworld ./internal/mod/d2legacy ./internal/app/clientapp -run 'TestDenPopulation|TestDenOfEvilPopulation' -count=1`; expect synthetic cases to pass and owned cases to report explicit skips.
- [ ] **Step 7: Commit:** `git add` only the scoped test and fix files, then `git commit -m 'test: verify Den population journeys and recovery for #660'`.

### Task 4: Final verification, review, and integration

**Files:**
- Modify only scoped files if a reviewer finds a #660 defect; fill the existing PR template in the PR body.

**Interfaces:**
- Consumes: Tasks 1–3 implementation and tests.
- Produces: a complete issue-linked PR with exact submitted-commit results and independent correctness/evidence/style review.

- [ ] **Step 1: Run full local gates on final head:** `make test architecture test-lua-format test-lua-syntax test-lua-hardening lint build-client-backends`; `make test-race`; focused owned 1.14d tests with `MPQ_DIRECTORY` as above; `git diff --check`.
- [ ] **Step 2: Request three independent final reviews:** issue acceptance/correctness, owned-data/evidence plus scope, and a read-only formatting/style review using the repository's `refactor-for-readability` skill. Apply only #660-scoped feedback and rerun affected checks on the amended head.
- [ ] **Step 3: Write every PR template section before submission.** Include `Closes #660`, direct code/evidence/test links, exact checks and skips from final commit, AI assistance and review, unresolved retail questions, and native-pixel limitation if still unverified. Review the complete diff yourself.
- [ ] **Step 4: Push, open PR, wait for CI, rebase on current main if needed, repeat verification/review after changes, mark ready, and merge only after required CI passes.** Attach the PR artifact to this task.
