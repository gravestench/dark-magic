# Level-aware Monster Collision Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make each opted-in monster validate native fixed-tick movement against its authoritative generated level map.

**Architecture:** Extend the existing velocity integrator with a level-aware registration path sharing its single system ID, phase, and integration core. Resolve a configured location field from the current ECS world on each tick and select a prepared immutable map by level. Wire offline and listen hosts identically; pin the native behavior version in the existing session configuration hash.

**Tech Stack:** Go ECS/Akara, renderer-independent `internal/game/world`, d2legacy client composition, Go synthetic and owned-data integration tests.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-623-level-aware-monster-collision-design.md`

## Global Constraints

- Work only on #623: level-aware collision for existing generated Act I maps. Keep one movement authority and all existing movement-phase collision arithmetic.
- Engine code uses configurable schema and field names; `d2legacy.world.location.level_id` is supplied only at client composition.
- Preserve Lua-owned AI, population, location schema, map generation, and world-transition policy.
- No game assets, captures, saves, or private data in commits.
- Label Dark Magic invariants and owned 1.14d integration facts separately from unverified retail 1.13c/1.14d behavior.

## Review Focus

- Missing location/map at a same-cell step must error before the fast path (Task 1).
- A level change on the same mover must select the new map on the next tick (Task 1).
- ECS restore/rollback must refresh the location store as well as spatial stores (Task 1).
- Old two-zone recovery must register only its prepared maps and fail closed for level 3 (Task 2).
- A real-map test must use a step where level-2 and level-3 collision disagree; a walkable-only test cannot detect this bug (Task 3).

---

### Task 1: Select collision by each mover's current level

**Files:** Modify `internal/game/world/velocity_movement.go` and `velocity_movement_test.go`; create `internal/game/world/velocity_movement_levels.go` if it keeps the registration/selector focused.

**Interfaces:** Preserve `RegisterVelocityMovement(engine *gameecs.Engine, paths VelocityPathFinder, components VelocityComponents) error`. Add `type VelocityLevelLocation struct { Component, Field string }` and `RegisterVelocityMovementByLevel(engine *gameecs.Engine, levels map[int64]VelocityPathFinder, components VelocityComponents, location VelocityLevelLocation) error`. Both register only `VelocityMovementSystemID`; only one is called per engine. The new path declares `location.Component` in Read, leaves its component out of All, and selects before `velocityDestinationAllowed`. Copy the level lookup at registration and reject absent/nil paths. Refresh the location store after `Engine.Restore`/rollback.

- [ ] Add `TestVelocityMovementByLevelUsesCurrentMoverLocation`: two 40×40 maps, a blocked synthetic tile at `(2,2)` in one map, movers at `(9.4,12)` with radius `0.1` and velocity `(25,0)` for `40ms`. Assert the open-level mover reaches `10.4`, the blocked-level mover remains `9.4` and both velocity fields become zero. Swap blocked maps and change one mover's level, then assert the inverse outcome.
- [ ] Add `TestVelocityMovementByLevelLongStepUsesSelectedMap`: a velocity step crossing multiple collision cells must call `FindPath` with the mover's chosen level map and actual radius; the same start/goal must succeed on one synthetic map and block on the other. This case is included by the focused `TestVelocityMovement` command below.
- [ ] Add `TestVelocityMovementByLevelRejectsUnknownLocationBeforeSameCell`: missing location, invalid field/level, absent map, and nil map each return an error with no position write, including a move entirely within one collision cell.
- [ ] Add `TestVelocityMovementByLevelUsesRestoredLocation`: snapshot/restore and late session rollback replace stores; a changed level afterward selects the matching map. Preserve the existing single-map, adjacent-step, long-step, and velocity-consumption assertions.
- [ ] Run `go test ./internal/game/world -run 'TestVelocityMovement' -count=1`; expect the new cases to fail on the old registration path.
- [ ] Implement the shared registration and per-entity selector without changing `moveVelocityEntity`'s collision/write rules. Run the same command and `go test -race ./internal/game/world -run 'TestVelocityMovement' -count=1`; expect PASS.
- [ ] Commit Task 1.

### Task 2: Wire prepared maps and pin native semantics

**Files:** Modify `internal/app/clientapp/offline_commands.go`, `network_host.go`, `session_bootstrap.go`, `session_bootstrap_test.go`, and focused client composition tests as needed.

**Interfaces:** Add a single `preparedVelocityPaths(worlds map[int]*gameworld.Map) (map[int64]gameworld.VelocityPathFinder, error)` in clientapp. Both offline and listen registrations call Task 1's `RegisterVelocityMovementByLevel` with `VelocityLevelLocation{Component:"d2legacy.world.location", Field:"level_id"}`. Introduce `gameworld.VelocityMovementLevelMapVersion = "level-aware-v1"` beside Task 1's API, and include `"engine.world.velocity_movement_version": gameworld.VelocityMovementLevelMapVersion` in `sessionInitialData()`; do not add a Lua capability name.

- [ ] Write failing client tests: the map adapter exposes every prepared nonnil level exactly once, rejects an empty/nil map set, and a two-zone topology has no level-3 fallback. Test `sessionInitialData` includes the version and an otherwise identical configuration without it hashes to a different runtime identity.
- [ ] Run focused `go test ./internal/app/clientapp -run 'Test.*(Velocity|SessionInitialData)' -count=1`; expect new assertions to fail.
- [ ] Implement both production registrations with the same adapter/configuration. Verify a current identity accepts the configured runtime while an otherwise identical old-version identity is rejected by existing recovery/compatibility checks, without mutating the target session.
- [ ] Run focused client tests, `go test ./internal/mod/d2legacy ./internal/app/clientapp -run 'Test.*(Velocity|RuntimeIdentity|Recovery)' -count=1`, and `make architecture`; expect PASS.
- [ ] Commit Task 2.

### Task 3: Prove owned-map movement and offline/listen continuation

**Files:** Add `internal/app/clientapp/cold_plains_movement_real_test.go`; extend existing `cold_plains_population_real_test.go` only where its helpers need a narrow shared assertion.

**Interfaces:** Reuse `newColdPopulationFixture`, `moveAcrossColdPlainsSeam`, `advancePopulationRecoveryTick`, acknowledged host checkpoint projection, and `app.gameWorlds[2/3]`. Keep test-only probes in test files. No new production movement or population API.

- [ ] Add an owned-data fixture that finds a deterministic adjacent step at which the generated level-2 and level-3 maps disagree for a monster's real radius. The **ordinary spawned Cold Plains monster**, using the production movement registration, must attempt that step during `chase`; fail with an explicit fixture diagnostic if its AI target/state/velocity or the discriminating step cannot be established. Assert the expected position and blocked velocity consumption against the level-3 map, and demonstrate that level-2 validation would give the opposite result. Let the existing fixture helper skip only when owned data is absent.
- [ ] Exercise that same spawned-monster case through offline and listen authorities with equivalent admitted commands and completed ticks. Compare corresponding monster state and position plus the acknowledged connected view, not asynchronous wall-clock samples. Both production registrations must fail this test if either still uses level 2.
- [ ] From a Cold Plains checkpoint, continue both source and restored authorities through 3→2→3 with the same commands; assert monster positions, full checkpoints, replay command sequence, and checksums match. Include a rejection check for a checkpoint whose only identity difference is the absent native movement version.
- [ ] Run focused tests first without `MPQ_DIRECTORY` and confirm the owned cases report SKIP. Then run with `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/app/clientapp -run 'Test.*ColdPlains.*(Movement|Recovery|Population)' -count=1 -v`; confirm the new owned cases RUN and PASS. Keep protected data and captures local.
- [ ] Commit Task 3.

### Task 4: Final verification, independent review, and PR

**Files:** Only issue #623 files and the PR description template; open a separate linked issue for any unrelated finding.

- [ ] Run exact-head `make test`, `make architecture`, `make test-race`, `make lint`, `make test-lua-hardening`, focused owned-data tests, and `git diff --check origin/main...HEAD`; record results and skips. Run `make build-client-backends` if native composition changed build paths.
- [ ] Have independent correctness/evidence reviewers compare full diff with #623 acceptance and actual production call paths. Have a separate style reviewer use `refactor-for-readability` and check formatting. Fix and re-review any Critical/Important finding.
- [ ] Fill every `.github/PULL_REQUEST_TEMPLATE.md` section with exact submitted-commit checks, 1.14d evidence vs inference, AI review, compatibility effect, and remaining limits. Rebase on current `main`, push, attach PR, watch CI, and merge only after green checks.
