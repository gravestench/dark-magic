# Seamless Cold Plains–Stony Field Border Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Cold Plains–Stony Field outdoor border visible and walkable in both directions in new Act I games.

**Architecture:** Extend Lua's admitted outdoor recipes and reciprocal seam policy to level 4, then publish a canonical version-3 `EntryWorld`. Reuse the prepared connected maps, authoritative seam movement, and shared-camera scene; preserve version-1/2 snapshots and room lifecycles.

**Tech Stack:** d2legacy Lua mapgen/authority/scene, Go worldgen transport and snapshot validation, Go offline/connected integration tests, owned 1.14d MPQ materialization.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-648-stony-field-live-seam-design.md`

## Global Constraints

- Issue #648 owns this work; #646 supplies the level-4 headless recipe. Original endpoint coordinates, border art, layout, and seed call order remain unverified; the new placement is provisional Dark Magic policy.
- Outdoor level 3↔4 is a reciprocal visible seam. Cave entrances and portals remain discrete warps. The already-shipped level 1↔2 and 2↔3 outcomes remain stable.
- New games publish four ordered zones (levels 1–4) and three seams as snapshot version 3. Version-1 and version-2 snapshot bytes, checksums, zone sets, and restore behavior remain unchanged.
- Lua chooses roles, directions, and recipes. Go transports and validates the admitted topology; neither Go nor the generic host chooses Diablo placement.
- Level-4 room residency, imported-item attachment, activation, and restoration operate without level-4 monster creation or population random draws. Preserve levels 2/3 monster outcomes and streams.
- Do not commit game data, extracts, captures, private saves, credentials, or dependency updates.

## Review Focus

- A v3 snapshot with one valid-looking but swapped generated endpoint must reject before world recovery; Task 2 tests endpoint identity and order.
- A v2 snapshot restored after v3 generation exists must retain exactly its original three maps and checksum; Task 2 pins its bytes and Task 5 checks recovery.
- Level-4 room residency must work without consuming level-2/3 population random draws; Task 4 checks both.
- Crossing just inside the new seam and immediately reversing must not retrigger or camera-snap; Tasks 3 and 5 test arrival separation and continuity.
- Static art and floors from both zones must remain correctly ordered at an off-center border crossing; Task 6 tests actual scene placements and bounded tile residency.

---

### Task 1: Place the Provisional Cold Plains Exit and Lua Seam

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/mapgen/cold_plains.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/mapgen/entry_world.lua`
- Create: `internal/mod/d2legacy/adapter/mapgen/entry_world_v3_policy_test.go`

**Interfaces:**
- Consumes: `cold_plains.generate(3, seed, source_direction, difficulty)` and `stony_field.generate(4, seed, source_direction, difficulty)` from #646.
- Produces: `cold_plains.generate_with_stony_exit(level_id, seed, source_direction, difficulty)` and `entry_world.cold_plains_v3(seed, difficulty, blood_moor)` for v3 composition. Also produces `entry_world.stony_field(seed, difficulty, plains)` and `entry_world.stony_field_seam(plains, stony)` returning the existing generated `SeamSpec` facts contract. The old `cold_plains.generate` and `entry_world.cold_plains` keep their single-entry output.

- [ ] **Step 1: Add failing focused tests.** In `entry_world_v3_policy_test.go`, call the explicit v3 policy for all four Blood Moor approach directions and difficulties 0–2 using synthetic records. Assert exactly one `stony-field-exit` targeting level 4 at the midpoint of the edge opposite `blood-moor-entry`, a connected dirt route to the interior, unchanged fill stamps/variants under the same seed, and unchanged single-entry result from the old direct `cold_plains.generate` call. Assert Stony Field's `cold-plains-entry` faces the exit and the described seam binds both unique warp IDs, roles, directions, coordinates, and destinations. Test absent/duplicate/mistargeted exit rejection.
- [ ] **Step 2: Prove RED.** Run `go test ./internal/mod/d2legacy/adapter/mapgen -run '^TestEntryWorldV3Policy' -count=1 -v`; expect the v3 policy call or endpoint to be absent.
- [ ] **Step 3: Implement the minimal Lua policy.** Share the current Cold Plains fill/variant loop between its old function and `generate_with_stony_exit` without changing stream names or draw order. Add the v3-only edge warp and `act1_dirt_path.apply` corridor, then derive Stony Field's source direction from the admitted exit. Bind the reciprocal generated seam using the existing geometry values (trigger radius 2, maximum arrival distance 10, inset 6); avoid a second seam authority.
- [ ] **Step 4: Prove GREEN.** Repeat the focused command, run `make test-lua-format test-lua-syntax`, and `git diff --check`.
- [ ] **Step 5: Commit** only the Task 1 files.

### Task 2: Canonical Version-3 Topology and Durable Validation

**Files:**
- Modify: `internal/mod/d2legacy/adapter/mapgen/entry_world.go`
- Create: `internal/mod/d2legacy/adapter/mapgen/entry_world_v3_test.go`
- Modify: `internal/mod/d2legacy/adapter/mapgen/entry_world_v2_test.go`
- Modify: `internal/mod/d2legacy/adapter/mapgen/entry_world_restore_test.go`
- Modify only if needed for preflight: `internal/app/clientapp/entry_world_recovery_test.go`

**Interfaces:**
- Consumes: Task 1's explicit Lua v3 generation and seam functions; `EntryWorld{Zones map[int]*worldgen.Zone, Seams []SeamSpec, Version int}`.
- Produces: `GenerateEntryWorld(...)` with ordered levels 1–4 and seams 1↔2, 2↔3, 3↔4, plus `Snapshot()`, `RestoreEntryWorld([]byte)`, and `Checksum()` support for version 3. Existing public signatures remain unchanged.

- [ ] **Step 1: Add failing transport/restore tests.** Record v1/v2 serialized bytes and checksums from the current generator before changing it, then use those pinned bytes with `RestoreEntryWorld` for legacy tests instead of calling the future v3 generator with the three-level `entryWorldV2Records()` fixture. Adapt `entry_world_v2_test.go` and `entry_world_restore_test.go` accordingly; keep the old Lua `entry_world.cold_plains` policy test. Add a four-level synthetic fixture and assert new generation has four admitted zones, three reciprocal seams, and stable v3 checksum. Table-test v3 rejection for missing/extra/reordered zones, seed/difficulty drift, duplicate or missing seam, wrong role/destination/direction/warp ID, out-of-bounds or mismatched tile, and invalid trigger/inset before returning a world. Assert a valid v3 restore retains its saved recipes/checksum even if Lua generation changes later.
- [ ] **Step 2: Prove RED.** Run `go test ./internal/mod/d2legacy/adapter/mapgen -run 'TestEntryWorldV3|TestEntryWorldLegacySnapshotBytes' -count=1 -v`; expect v3 cases to fail while recorded legacy bytes remain stable.
- [ ] **Step 3: Implement the v3 envelope and validation.** Add canonical ordered definitions and seams without modifying version-1/2 marshal branches. Re-admit every saved definition, validate four identities and shared seed/difficulty, require first seam authored and the other two generated, and bind each generated side to exactly one admitted warp. Call Task 1 Lua functions only after their source zone is admitted. Reuse existing seam parsing and generic helpers where their contracts fit.
- [ ] **Step 4: Prove GREEN.** Repeat focused tests and `go test ./internal/mod/d2legacy/adapter/mapgen -count=1`; run `git diff --check`.
- [ ] **Step 5: Commit** only the Task 2 files.

### Task 3: Materialize the Four-Zone Prepared World

**Files:**
- Modify only if a failing test requires it: `internal/mod/d2legacy/adapter/entryworld/build.go`, `internal/mod/d2legacy/adapter/entryworld/population.go`, `internal/mod/d2legacy/adapter/transition/seam.go`
- Create: `internal/mod/d2legacy/adapter/entryworld/stony_field_real_test.go`
- Extend: `internal/mod/d2legacy/adapter/entryworld/entryworld_test.go`
- Extend: `internal/mod/d2legacy/adapter/entryworld/cold_plains_real_test.go`

**Interfaces:**
- Consumes: Task 2's version-3 `GenerateEntryWorld`/restore and the existing `Prepared` maps and seam transitions.
- Produces: prepared maps for levels 1–4 and a collision-safe, reciprocal level-3/4 transition with level-4 room geometry available to `PopulationData`.

- [ ] **Step 1: Add failing synthetic and owned-data tests.** Assert `Prepared` contains four maps, three bound seams, and ordered population geometry for levels 2–4. Update `TestBuildColdPlainsReal` and related current-new-world assertions from three maps/two seams/four directed transitions to four maps/three seams/six directed transitions, while retaining their old-pair endpoint assertions and legacy-v1 case. With `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq`, materialize seed 17 Normal and each difficulty's source/arrival points; assert in-bounds unblocked points, an interior route from each side, reciprocal visual alignment, and an inset return point outside the reverse trigger. The owned test explicitly skips when MPQs are absent.
- [ ] **Step 2: Run the focused prepared tests and record RED or an already-passing generic path.** Use `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy/adapter/entryworld -run 'TestStonyField|TestPopulationDataOrdersAdmittedLevels|TestBuildColdPlainsReal' -count=1 -v`, then repeat without the environment variable to confirm explicit skips.
- [ ] **Step 3: Fix only an actual prepared-world gap.** Reuse generic composition and safe-inset logic. Do not add placement policy to Go or expand to level-4 monster population.
- [ ] **Step 4: Prove GREEN** with the focused tests, explicit owned-data execution/skip, and `git diff --check`.
- [ ] **Step 5: Commit** Task 3 tests and any necessary scoped fix.

### Task 4: Keep Level-4 Rooms Active Without Monster Rolls

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua`
- Extend: `internal/mod/d2legacy/cold_plains_population_test.go`

**Interfaces:**
- Consumes: Task 3's ordered level-2/3/4 `PopulationData`.
- Produces: a validated level-4 room plan usable by `resident_at` and `activate_zone`, with no level-4 monster creation or population RNG calls.

- [ ] **Step 1: Add failing Lua-authority tests.** Admit levels 2–4 atomically, reject reordered/duplicate level 4 and missing level 3, while preserving a valid legacy level-2/3 plan. Attach a preexisting level-4 item to its room, deactivate/reactivate that room and verify residency/return. Assert zero level-4 monster entities and zero level-4 population random calls across first activation and return. Replay equivalent level-2/3 commands with and without the level-4 zone and assert identical monster identities/positions and random continuation.
- [ ] **Step 2: Prove RED.** Run `go test ./internal/mod/d2legacy -run 'TestStonyFieldPopulation' -count=1 -v`; expect level-4 validation or room behavior failure.
- [ ] **Step 3: Extend only the room policy.** Admit canonical level 4 and apply its room lifecycle. Bypass candidate loading, density/family/group/seed/life rolls, and monster creation for level 4 before any population RNG call. Leave level-2/3 branches and stream names intact.
- [ ] **Step 4: Prove GREEN.** Repeat focused tests, existing `TestColdPlainsPopulation` tests, Lua format/syntax/hardening checks, and `git diff --check`.
- [ ] **Step 5: Commit** only Task 4 files.

### Task 5: Production Crossing, Recovery, and Connected Authority

**Files:**
- Create: `internal/app/clientapp/stony_field_transition_test.go`
- Extend: `internal/app/clientapp/cold_plains_transition_test.go`, `internal/app/clientapp/cold_plains_population_real_test.go`
- Extend only when a failing production test requires it: `internal/app/clientapp/entry_world_recovery.go`, `internal/app/clientapp/world_generation.go`, `internal/app/clientapp/entry_world_recovery_test.go`

**Interfaces:**
- Consumes: Tasks 2–4's v3 prepared topology and room plan; existing production movement/transition and offline/connected fixture helpers.
- Produces: new-game and restored/reconnected crossing 3→4→3 with one authoritative level change per crossing and stable camera/hero position.

- [ ] **Step 1: Add failing production-path tests.** Use the same movement commands as `TestLiveColdPlainsRoundTrip`: cross 3→4→3 offline and listen/connected, test a stale level-3 movement sample after arrival, then save/recover a v3 checkpoint and reconnect before reversing. Assert the client/server topology checksum, active map, location and position agree. Update current-new-game assumptions in `cold_plains_transition_test.go` from two seams to three and in `cold_plains_population_real_test.go` from two population zones to three, while preserving their existing level-2/3 claims. Restore a recorded v2 checkpoint and assert it remains a three-map world without a level-4 seam; corrupt a v3 seam/zone and assert preflight rejects it without mutating the session.
- [ ] **Step 2: Prove RED or record the already-working generic path.** Run `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/app/clientapp -run 'TestStonyField|TestLegacyEntryWorldSnapshotComposesWithoutColdPlains|TestLiveColdPlainsRoundTrip|TestLiveColdPlainsOrdinaryPopulation' -count=1 -v`; run without the environment variable to record explicit skips.
- [ ] **Step 3: Correct only observed production gaps.** Preserve the existing movement transaction and stale-level guard; modify recovery/composition only if v3 validation or map publication fails.
- [ ] **Step 4: Prove GREEN.** Repeat focused tests, existing Cold Plains crossing tests, owned-data execution/skip, and `git diff --check`.
- [ ] **Step 5: Commit** Task 5 files.

### Task 6: Advance Visibility and Shared-Camera Scene

**Files:**
- Modify if needed: `internal/content/d2legacy/lua/d2legacy/screens/game_world.lua`
- Extend: `internal/content/d2legacy/lua/d2legacy/screens/game_world_seams_test.lua`
- Extend if needed: `internal/content/d2legacy/lua/d2legacy/presentation/outdoor_seams_test.lua`

**Interfaces:**
- Consumes: Task 5's prepared version-3 maps/seams and existing shared camera root.
- Produces: both border-adjacent maps visibly placed before either crossing, continuous hero/camera, correct level-4 art ordering, and bounded resident tiles.

- [ ] **Step 1: Add failing scene tests.** Extend the existing seam fixture. For an off-center approach on each side, assert actual level-4 floor and static-art nodes are in the viewport while the actor remains in Cold Plains, and actual level-3 nodes are visible before the return crossing. Cross and reverse while checking camera/hero screen continuity, static-art/floor depth against the hero, no black/loading frame, and bounded sparse tile residency. Assert pointer projection and collision commands use only the active map. Keep a discrete cave/portal scene transition test that does not carry outdoor neighbor art.
- [ ] **Step 2: Prove RED.** Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/screens/game_world|TestLuaSuites/d2legacy/presentation/outdoor_seams' -count=1 -v`; identify any level-4 priority/placement failure.
- [ ] **Step 3: Correct the smallest scene gap.** Extend the current `id <= 3` depth rule only as required by actual overlap tests; retain one camera root and existing connected-map residency. Do not create a new visual handoff.
- [ ] **Step 4: Prove GREEN.** Repeat focused scene tests, `make test-lua-format test-lua-syntax`, and `git diff --check`.
- [ ] **Step 5: Commit** Task 6 files.

## Final Verification and Delivery

- [ ] Rebase onto current `main` after #646 lands. Run focused v3 mapgen, prepared owned-data, population, crossing/recovery, and scene tests on the submitted commit. Run `make test`, `make architecture`, `make test-race` alone, Lua format/syntax/hardening, pinned lint, both client backends, and `git diff --check`; record actual results and owned-data availability.
- [ ] Attempt both approach-direction graphical captures with owned 1.14d data when a monitor is available. Inspect produced images; report renderer initialization limits separately if no monitor exists. Commit no captures.
- [ ] Obtain independent whole-diff correctness/evidence reviews and a formatting/readability review that applies `refactor-for-readability`. Resolve scoped findings and rerun gates on the final commit.
- [ ] Fill every PR-template section, include `Closes #648` and `Refs #354`, push, attach the PR, and merge after required CI succeeds.
