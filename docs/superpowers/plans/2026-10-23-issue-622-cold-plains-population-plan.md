# Cold Plains Ordinary Population Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or
superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for
tracking.

**Goal:** Populate live Cold Plains with ordinary data-selected monsters while preserving Blood Moor behavior
and deterministic recovery.

**Architecture:** The prepared entry world emits one ordered population envelope for its admitted wilderness
levels. The existing tick-1 system command installs one version-6 level-keyed Lua plan atomically. The current
room lifecycle activates both levels; level 3 receives separate random streams and full-radius terrain and
occupancy checks.

**Tech Stack:** Go entry-world adapter and application, d2legacy Lua authority, Go/Lua tests, owned MPQ
fixtures.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-622-cold-plains-population-design.md`

## Global Constraints

- Issue #622 only: ordinary level-3 population, stable level-2 behavior, production offline/connected proof, and
  directly relevant tests/docs.
- Keep one population authority and one privileged tick-1 bootstrap command; do not change generic state store,
  runtime identity enforcement, or dependencies.
- Preserve level-2 room and random-draw order. Give level 3 separate density, family, group, seed, and
  monster-life streams.
- Version the checkpointed plan as v6. Historical full-checkpoint migration is outside #622; current-runtime v1
  entry topology must bootstrap level 2 alone.
- Original 1.13c/1.14d density/group/room arithmetic remains unverified Dark Magic policy. Never commit owned
  game data or captures.

## Review Focus

- Missing level-3 map or zone must reject the entire bootstrap before command admission: Task 1 prepared-world
  test.
- Duplicate or out-of-order zone IDs and a second valid bootstrap must leave state and ECS unchanged: Task 1
  admission tests.
- Same numeric room ID in levels 2 and 3 must not cross-activate residents: Task 1 level-scoping test.
- A selected level-3 monster must not overlap terrain, an active collider, or another queued group member: Task
  1 placement tests.
- A Cold Plains visit must not change later Blood Moor random outcomes, including monster life: Task 1
  stream-isolation test.

---

### Task 1: One working multi-zone population authority

The Go envelope and Lua consumer land in one commit so Blood Moor remains populated at every commit boundary.

**Files:**
- Modify: `internal/mod/d2legacy/adapter/entryworld/population.go`, `bootstrap.go`, `entryworld_test.go`,
  `cold_plains_real_test.go`
- Modify: `internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua`, `commands/spawn_monster.lua`,
  `commands/item_use_test.lua`
- Modify: `internal/mod/d2legacy/identity.go`, `hostile_lifecycle_restore_test.go`,
  `monster_difficulty_session_test.go`, `population_player_count_test.go`
- Create: `internal/mod/d2legacy/cold_plains_population_test.go`
- Update other existing tests only where they read the v5 top-level plan shape or assume a flat prepared-world
  payload.

**Interfaces:**
- Consumes: `Prepared.Seams`, `Zones`, `Worlds`, `Spawns`;
  `gameplay.collision.destination_clear(level,x,y,radius)`; ECS position/location/collider facts; authority
  random streams.
- Produces: `(*Prepared).PopulationData(nearby int) (map[string]any, error)` with `{zones:[zone2,zone3]}` in
  seam order or `{zones:[zone2]}` for a restored v1 topology. `PopulationCommand` retains its tick-1
  `system.population.bootstrap` identity. Lua registers `d2legacy.population.plan/v6` with ordered `zones`,
  retaining `M.resident_at` and `d2legacy.population.active_rooms`. `spawn.components` accepts an optional
  life-stream argument and uses the current default stream when it is absent.

- [ ] **Step 1: Write failing Go envelope tests.** Add
      `TestPopulationDataOrdersAdmittedLevelsAndRejectsMissingMap`, `TestPopulationDataV1KeepsOnlyLevelTwo`, and
      `TestPopulationDataLevelThreePointsStayDistinctAndInRoom`. Assert level order 2,3, own room bounds/links,
      duplicate/missing zone failure, no command on error, level-2 point order, room-local distinct level-3
      centers, and zero-point blocked rooms. Update existing
      `TestPreparedWorldOwnsDestinationAndBootstrapLevelIdentity` and
      `TestPopulationDataPreservesRoomBoundsAndAdjacency` for the new `(map,error)` envelope; retain their
      level/room assertions. Extend `TestBuildColdPlainsReal` for owned level-3 facts and v1 level-2-only
      population.
- [ ] **Step 2: Write failing authority tests.** Add `TestColdPlainsPopulationAdmitsBothLevelsAtomically`,
      `TestColdPlainsPopulationRejectsDuplicateAndOutOfOrderZones`,
      `TestColdPlainsPopulationRejectsSecondBootstrapWithoutMutation`,
      `TestColdPlainsPopulationSparseAndEmptyCandidates`, `TestColdPlainsPopulationKeepsLevelTwoRandomSequence`,
      `TestColdPlainsPopulationRoomIDsAreLevelScoped`,
      `TestColdPlainsPopulationSkipsBlockedOrOccupiedFootprints`, and
      `TestColdPlainsPopulationReplayMatchesContinuation`. Cover all three difficulties with synthetic records;
      assert selected candidates, stable level-qualified IDs, no overlap with terrain/active colliders/queued
      group members, and unchanged level-2 IDs, health, and later random outcomes. Include `NumMon=0` (no
      candidate/no spawn) and a sparse `mon*`/`nmon*` list with a valid later slot (deterministic selection/no
      panic). Submit the same valid bootstrap twice and compare plan/ECS before and after rejection. Call the
      existing `Session.Replay()` for uninterrupted and restored runs, compare continuation commands from the
      checkpoint tick onward and final checkpoint checksum, and verify each replay through the existing
      simulation API. Existing flat single-zone synthetic commands must remain valid.
- [ ] **Step 3: Run red.** Run the command below. Require the new tests to fail under the current level-2-only
      path; record any owned-test skip separately.

      ```sh
      go test -count=1 ./internal/mod/d2legacy/adapter/entryworld ./internal/mod/d2legacy \
        -run 'Test(PopulationData|PreparedWorldOwns|BuildColdPlainsReal|ColdPlainsPopulation)'
      ```
- [ ] **Step 4: Implement the envelope and authority together.** Derive wilderness IDs from the prepared seam
      sequence, validate keyed zone request IDs and map presence (maps have no intrinsic level ID), preserve
      level-2 point order, and build level-3 distinct room-local center points. Validate all zones before one
      state replacement; normalize flat synthetic commands; register the v6 plan; make room/resident lookup
      level-scoped. Use distinct level-3 density/family/group/seed streams and an optional level-3 life stream
      in `spawn.components`, preserving its default for all old callers. Before each level-3 structural create,
      check full-radius terrain and room bounds, active same-level collider spacing, and queued same-tick spawn
      spacing. Zero safe points yield zero spawns. Keep level 2's materialization path unchanged.
- [ ] **Step 5: Run green and regression checks.** Run the focused command from Step 3, then the commands
      below. The full package command includes existing generated-hostile and population-inactivation tests.
      Update `commands/item_use_test.lua` and Go tests only as needed for the v6 plan shape, retaining their
      behavior assertions.

      ```sh
      go test -count=1 ./internal/mod/d2legacy
      make test-lua-hardening
      MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq \
        DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq \
        go test -count=1 ./internal/mod/d2legacy/adapter/entryworld \
        -run 'Test(BuildColdPlainsReal|RestorePreparedV1KeepsTwoZonesReal|PopulationData)'
      ```
- [ ] **Step 6: Review and commit.** Inspect the complete Task 1 diff and `git diff --check`; commit
      `feat(d2legacy): populate Cold Plains through shared room authority` only after the new and existing
      focused checks pass.

### Task 2: Production crossing, replay/recovery, and evidence

**Files:**
- Create: `internal/app/clientapp/cold_plains_population_real_test.go`
- Modify: `internal/app/clientapp/entry_world_recovery_test.go` and `entry_real_spells_test.go` only where
  existing assertions need v6 shape
- Modify: `docs/research/MONSTERS_SPAWNING_AND_AI.md` and `docs/MAP_GENERATION.md`
- Modify: `docs/research/NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md` only if an explicit #622 note is needed for
  its historical cursor

**Interfaces:**
- Consumes: Task 1 authoritative v6 plan and existing #620 seam movement/connected fixture helpers; no new
  transition, network, or presentation authority.
- Produces: owned-data production tests proving level-3 monster identity and visibility, level-2 state
  preservation, connected authority agreement, and a Cold Plains checkpoint/replay whose 3→2→3 continuation
  matches uninterrupted execution.

- [ ] **Step 1: Write production tests.** Add `TestLiveColdPlainsOrdinaryPopulation`,
      `TestConnectedColdPlainsPopulationMatchesAuthority`, and
      `TestColdPlainsPopulationRecoveryMatchesContinuation` using owned assets and ordinary movement across #620
      seams. Assert level-3 monster IDs/positions, terrain-safe nonoverlap, no duplicate creation on return, and
      Blood Moor plan/resident IDs and active/inactive state across 2→3→2. For connected play, compare the
      client's projection with its own host authority at an acknowledged completed tick; compare offline and
      connected authority outcomes after the same seed, commands, and completed-tick count rather than
      wall-clock arrival. Compare the restored 3→2→3 continuation's complete authoritative checkpoint, semantic
      events, and replay commands from the checkpoint tick onward with uninterrupted execution; compare final
      replay checkpoint checksums and verify each replay separately, since their initial baselines differ. Under
      the current runtime, a restored v1 entry topology must bootstrap level 2 alone.
- [ ] **Step 2: Run production tests.** With owned `MPQ_DIRECTORY` and `DARK_MAGIC_TEST_MPQ_DIRECTORY`, run
      `go test -count=1 ./internal/app/clientapp -run 'ColdPlains.*Population'`. If an assertion fails because
      the production path does not publish or observe the Task 1 plan, retain the failing test for Step 3. If
      all pass, record that no additional production wiring was needed.
- [ ] **Step 3: Wire only missing production details and document evidence.** Reuse offline/listen population
      command and authority snapshots. Update only code required by failing tests. Record the owned 1.14d
      `patch_d2.mpq` provenance: `Levels.txt` row `Id=3`, `NumMon`, exact Normal/Nightmare/Hell `MonDen` and
      `mon*`/`nmon*` columns selected by the interpreter, selected `MonStats.txt` ordinary rows, and
      `LvlPrest.txt Populate`. Distinguish interpreter selection from original-game inference and label current
      density/group/room arithmetic Dark Magic policy. Update `docs/MAP_GENERATION.md`'s statement that Cold
      Plains population remains future work.
- [ ] **Step 4: Check and commit the production proof.** Repeat Step 2's owned-data command, run
      `git diff --check`, inspect the Task 2 diff, and commit `test(d2legacy): prove live Cold Plains population`
      (or a scoped feature commit if wiring was needed).
- [ ] **Step 5: Obtain independent code, evidence, and formatting/readability reviews.** The style reviewer must
      read `refactor-for-readability`. Incorporate feedback in follow-up commits, then run exact CI lint package
      scope, `go test -count=1 ./...`, `make architecture`, `make test-lua-hardening`, issue-focused owned MPQ
      tests, and `make test-race` on the final commit. Record exact commands/results, note data-dependent CI
      skips, and inspect `git diff --check` and the complete issue-scoped diff. Any change after these checks
      requires rerunning affected gates and updating the recorded commit.

## Completion

Prepare every section of `.github/PULL_REQUEST_TEMPLATE.md`, including exact final-commit checks,
owned-game-data status, original-behavior uncertainty, and AI review. Push the issue branch, open a non-draft PR
with `Closes #622` and `Refs #354, #356`, attach it to this task, wait for required CI to pass, and merge. Do
not merge on a pending or failed check.
