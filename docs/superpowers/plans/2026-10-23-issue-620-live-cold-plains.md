# Live Cold Plains Crossing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let an authoritative player cross Blood Moor↔Cold Plains through ordinary movement, restore that journey, and show it to a connected client.

**Architecture:** Eagerly prepare all three Act I zones and both seam pairs in the existing entry-world adapter. Carry one versioned topology and its checksum through collision preparation, Lua bootstrap, runtime identity, and session recovery. Reuse the current Lua transition transaction for relocation.

**Tech Stack:** Go mapgen/entryworld/transition adapters, d2legacy Lua map and transition policy, worldgen snapshots, game session recovery, owned MPQ integration tests.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-620-live-cold-plains-design.md`

## Global Constraints

- Issue #620 is the sole code scope. #618 supplies the Cold Plains recipe; #354 and #353 are parents/context.
- Current composition seed is explicit (1 in clientapp). Keep the same seed/difficulty across all three zones; do not add seed ownership.
- The level-3 layout, seam geometry, trigger radius, and arrival offsets are Dark Magic policy, not proven 1.13c/1.14d placement.
- v1 two-zone snapshots must restore without gaining Cold Plains; new three-zone worlds use a canonical v2 envelope.
- Keep d2legacy decisions in Lua and collision resolution in Go. Do not add population, waypoint, outward exits, quests, or dependencies.
- Never commit game data, MPQ bytes, captures, credentials, or saves.

## Review Focus

- Missing, duplicate, wrong-target, or non-cardinal Blood Moor exit: Task 1 rejects before level-3 generation.
- Duplicate or mismatched level/seam facts, including seed/difficulty drift: Task 1 rejects before snapshot restoration or publication.
- Blocked, distant, or immediate-return arrival: Task 1 rejects within bounded collision resolution.
- A valid checkpoint paired with a different valid map topology: Task 2 rejects before calling session restore.
- A connected client with a different topology checksum: Task 2 rejects at the existing runtime-identity handshake.

---

### Task 1: Prepare one three-zone topology with both seam pairs

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/mapgen/entry_world.lua`
- Modify: `internal/mod/d2legacy/adapter/mapgen/entry_world.go`
- Create: `internal/mod/d2legacy/adapter/mapgen/entry_world_v2_test.go`
- Modify: `internal/mod/d2legacy/adapter/mapgen/entry_world_restore_test.go`
- Modify: `internal/mod/d2legacy/adapter/transition/seam.go`
- Modify: `internal/mod/d2legacy/adapter/transition/seam_test.go`
- Modify: `internal/mod/d2legacy/adapter/entryworld/{build.go,entryworld.go,bootstrap.go,population.go,entryworld_test.go}`
- Create: `internal/mod/d2legacy/adapter/entryworld/cold_plains_real_test.go`
- Migrate direct callers/tests: `internal/app/clientapp/{application.go,world_generation.go,world_transitions.go,session_bootstrap.go,world_test.go,entry_real_warp_test.go}` and `internal/game/world/{materialize_real_test.go,cold_plains_real_test.go}`

**Interfaces:**
- `EntryWorld{Zones map[int]*worldgen.Zone, Seams []SeamSpec, Version int}` is the single in-memory topology. `GenerateEntryWorld(ctx, source, records, seed, difficulty)` generates v2; `RestoreEntryWorld(encoded)` accepts v1/v2; `Snapshot()` and `Checksum()` preserve each version's canonical representation. Lua `entry_world.cold_plains(seed, difficulty, bloodMoorFacts)` validates the unique level-3 exit and calls `cold_plains.generate`; Go only transports it.
- `SeamSpec` carries first/second level, directions, both generated tile coordinates when applicable, `FirstGenerated`, `TriggerRadius=2`, `MaxArrivalDistance=10`, and `ArrivalInset=6` (subtiles). `ResolveSeam(spec, firstMap, secondMap) (Seam, error)` returns `Seam{First, Second SeamEndpoint; TriggerRadius float64}`. Both directions receive the same radius in `TransitionData([]Seam)`.
- `Prepared{Worlds, Zones, Spawns, Seams []transition.Seam, Topology mapgen.EntryWorld, Difficulty}` is published only after all three maps and two seam pairs resolve. `RestorePrepared(ctx, content, source, records, resolver, snapshot) (*Prepared, error)` restores zone definitions without rerunning zone generation; it may rerun deterministic materialization and `ResolveTownEntry` spawn resolution. New worlds emit four directed Lua entries; restored v1 emits two. Migrate existing callers to these interfaces in this same task so the branch compiles at the commit.

- [ ] **Step 1: Write red policy/topology tests.** `TestEntryWorldV2SelectsUniqueBloodMoorExit` checks three difficulties, four directions, opposite entry, zones 1/2/3, pair order, and shared seed. `TestEntryWorldV2RejectsMalformedSourceExit` calls `entry_world.cold_plains` with missing, duplicate, wrong-target, and non-cardinal admitted warp facts. `TestEntryWorldV2SnapshotAndLegacyRestore` checks v2 canonical checksum, invalid duplicate/mismatched facts, and unchanged v1 snapshot/checksum with exactly two zones.
- [ ] **Step 2: Run `go test ./internal/mod/d2legacy/adapter/mapgen -run '(EntryWorldV2|LuaSelectedEntryWorld)' -count=1`.** Expected red on absent v2 behavior; the regex includes the existing `TestLuaSelectedEntryWorldRestoresToSameTopologyChecksum`.
- [ ] **Step 3: Write red resolver tests.** `TestResolveGeneratedSeamAllDirections` covers four reciprocal orientations and both destination IDs. `TestResolveGeneratedSeamRejectsUnsafeArrival` checks blocked, >10-subtile displacement, no interior path, and arrival within radius 2 of a return trigger. Run `go test ./internal/mod/d2legacy/adapter/transition -run 'Resolve(Generated|Seam)' -count=1`; expected red on the generated first edge.
- [ ] **Step 4: Implement Lua selection, v2/v1 topology, and generated-edge resolution.** For v1 decode, keep version 1 and old canonical serialization; never synthesize level 3. Use Lua-authored source and geometry facts, Go materialization/collision, and stable seam order. The authored town edge retains its current error precedence.
- [ ] **Step 5: Write red preparation tests.** Use owned MPQs for `TestBuildColdPlainsReal` and `TestRestorePreparedV1KeepsTwoZonesReal`: verify materialized levels, bounded arrivals/interior paths, four v2 or two v1 directed seams, preserved v1 checksum/format, and no level-3 synthesis. Add `TestBuildFailsBeforePublicationOnMissingOrBlockedLevelThree`: inject a missing level-3 destination and an unsafe level-3 arrival through the production `Build` boundary, require an error and nil `Prepared` in both cases. These tests skip without `DARK_MAGIC_TEST_MPQ_DIRECTORY`; do not manufacture a `Prepared` value as a substitute for `Build` or `RestorePrepared`. Synthetic malformed-map tests stay in mapgen/transition where their fixtures exercise the owning boundary.
- [ ] **Step 6: Implement `Build`, `RestorePrepared`, bootstrap, spawns, and all direct caller migrations.** Keep level-2 population and existing level-1/2 spawns; add trusted level-3 arrival. Measure one owned-data preparation time without an arbitrary threshold.
- [ ] **Step 7: Run `go test ./...` and `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy/adapter/entryworld -run '(BuildColdPlainsReal|RestorePreparedV1KeepsTwoZonesReal|BuildFailsBeforePublicationOnMissingOrBlockedLevelThree)' -count=1 -v`.** Rerun focused mapgen and transition tests from Steps 2–3. Commit the coherent compiling change as `feat(entryworld): prepare Cold Plains and both seams`.

### Task 2: Session identity, live crossing, and recovery

**Files:**
- Modify: `internal/app/clientapp/world_generation.go`
- Modify: `internal/app/clientapp/world_transitions.go`
- Modify: `internal/app/clientapp/session_bootstrap.go`
- Modify: `internal/app/clientapp/application.go`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/world_transition.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/world_transition_test.lua`
- Create: `internal/app/clientapp/cold_plains_transition_test.go`
- Create: `internal/app/clientapp/entry_world_recovery.go`
- Create: `internal/app/clientapp/entry_world_recovery_test.go`
- Modify: connected-session tests near `internal/app/clientapp/network_host_test.go`

**Interfaces:**
- Consumes: Task 1 `Prepared`, topology snapshot/checksum, and the existing `Session.RecoveryCheckpoint`.
- Produces: `sessionInitialData()` includes `d2legacy.entry_world_checksum` for both authority and peer identity; client state stores the canonical seam list. `entryWorldRecovery{Topology []byte, TopologyChecksum string, Session gamesession.RecoveryCheckpoint}` is captured by `app.entryWorldRecoveryCheckpoint()` and restored through `app.restoreOfflineEntryWorldRecovery(bundle)`; that composition path verifies the checkpoint participant's runtime identity digest against the freshly composed identity before calling `RestoreRecoveryCheckpoint`.

- [ ] **Step 1: Write failing offline and connected tests.** Owned-data `TestLiveColdPlainsRoundTrip` crosses 1→2→3→2→1 through ordinary movement, observes committed identity/position/motion and active navigation; owned-data `TestConnectedColdPlainsMatchesAuthority` performs an actual authority/client handshake, checks the committed remote level/position projection and rendered map/navigation switch, and rejects a different topology at handshake. `TestColdPlainsSetupFailurePreservesExistingWorld` injects a failed three-zone preparation into the client composition boundary and verifies that the prior published world, topology identity, and active navigation are unchanged. Use the existing clientapp real-fixture helper; owned-data tests skip clearly without `MPQ_DIRECTORY`. Lua transition tests assert each seam uses its supplied trigger radius and avoids immediate bounce.
- [ ] **Step 2: Run focused clientapp/Lua tests to red.** `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/app/clientapp -run '(ColdPlains|EntryWorldRecovery)' -count=1 -v` and `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/systems/world_transition' -count=1`.
- [ ] **Step 3: Publish the seam list and topology checksum.** Update client bootstrap, fixture town-pair selection, active-map navigation, and Lua per-seam trigger radius while retaining the shared relocation transaction. Offline/listen authority and remote client derive the same configuration digest.
- [ ] **Step 4: Write `TestColdPlainsRecoveryMatchesContinuation` and `TestColdPlainsRecoveryRejectsDifferentTopology`.** Capture topology and session checkpoint in Cold Plains, reconstruct from the saved topology, preflight the checkpoint runtime identity before any restore, continue with the same return-crossing commands, and compare complete authoritative checkpoint checksums/participants and motion facts. Wrong topology must fail before a tick or state restore. Run the focused clientapp command from Step 2 and confirm these new tests fail for the missing recovery composition before implementation.
- [ ] **Step 5: Implement composition-owned `entryWorldRecovery`, rerun the owned-data clientapp command and Lua command to green, and commit as `feat(client): traverse and restore Cold Plains`. Record that the clientapp cases skip in CI without owned MPQs.

### Task 3: Documentation, review, and integration

**Files:**
- Modify: `docs/MAP_GENERATION.md`
- Modify: `docs/research/WARPS_PORTALS_WAYPOINTS.md` only to correct implemented ownership and clearly label remaining original-geometry gaps.

- [ ] **Step 1: Document the live two-way boundary and original-game uncertainties, then commit issue-scoped docs.**
- [ ] **Step 2: Request independent code/architecture, game-evidence, and formatting/readability reviews.** The style reviewer applies `refactor-for-readability`; fix findings and get revised approval.
- [ ] **Step 3: Rebase on current main after #619 merges.** On the final commit run `go test ./...`, `make architecture`, `make test-lua-hardening`, `make test-race`, `git diff --check origin/main...HEAD`, and the owned-MPQ focused tests. Report exact results and skips.
- [ ] **Step 4: Fill every PR-template section with issue criteria, evidence, ownership, exact checks, unverified behavior, and AI use.** Push, open/attach the PR, wait for required CI, merge after green, and verify #620 closes while #354 remains open.
