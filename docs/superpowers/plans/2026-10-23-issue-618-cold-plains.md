# Cold Plains Zone Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Cold Plains level 3 an admitted, deterministic, asset-materializable zone while preserving the current two-zone production entry world.

**Architecture:** A dedicated d2legacy Lua strategy reads recovered records and emits a `worldgen.Zone` through the existing headless `mapgen.Runtime.Generate`. The caller supplies a direction extracted from Blood Moor's admitted level-3 exit; generic Go owns validation, checksum, restoration, materialization, and pathfinding.

**Tech Stack:** Lua d2legacy map policy, Go `internal/game/worldgen`, headless Lua adapter tests, optional owned-MPQ materialization test.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-618-cold-plains-design.md`

## Global Constraints

- Issue #618 and G14 define scope; no `EntryWorld` v1, live transition, waypoint, population, or dependency changes.
- LvlPrest definitions 29 and 30 are provisional generic Act I 8×8 stone fill, not a claim of original Cold Plains layout.
- Lua interprets d2legacy records; Go remains transport and generic world mechanism.
- No Diablo II data, captures, credentials, or private saves may be committed.

## Review Focus

- Missing/invalid Blood Moor direction: Lua rejects it, and the real-data caller verifies the source warp targets level 3.
- Empty DS1/DT1 references: generation rejects them before normalization can turn an empty value into a directory path.
- Blocked, distant, or isolated entry: the asset-backed test requires an open arrival close to the edge warp and a collision path to a distinctly interior point.
- Difficulty-specific sizes: each mode uses its own Levels columns, with invalid/misaligned dimensions rejected.
- Restore identity: admitted definition round-trip preserves checksum without rerunning Lua.

---

### Task 1: Headless Cold Plains recipe and synthetic acceptance

**Files:**
- Create: `internal/content/d2legacy/lua/d2legacy/mapgen/cold_plains.lua`
- Create: `internal/mod/d2legacy/adapter/mapgen/cold_plains_test.go`

**Interfaces:**
- Consumes: `Runtime.Generate(ctx, "cold_plains", float64(3), float64(seed), sourceDirection, float64(difficulty))`, `records.level`, `records.level_type`, `records.preset_by_definition`, `assets.preset_files`, `assets.masked_tiles`, `worldgen.admit`.
- Produces: one level-3 `worldgen.Zone` with 8×8 stamps/rooms, connected links, one `Warp{Role:"blood-moor-entry", DestinationLevel:2}` opposite `sourceDirection`, and a semantic path to an interior tile.

- [ ] **Step 1: Write failing synthetic tests.** Use `entryRecords` with level 3, level type 2, and LvlPrest 29/30. Assert per-difficulty bounds, all four source directions and opposite entry edges, connected room graph and route, repeat/checksum stability and a changed checksum when only direction changes, `worldgen.NewZone(zone.Definition())` checksum parity, and rejection of wrong level/direction, mismatched Act/DRLG/LevelType and preset LevelId/Outdoors/SizeX/SizeY, missing rows, bad dimensions, missing preset variants, and empty DS1/DT1 paths. Use test names that distinguish these outcomes.
- [ ] **Step 2: Run the focused test to confirm failure.** `go test ./internal/mod/d2legacy/adapter/mapgen -run ColdPlains -count=1`; expected failure because `d2legacy.mapgen.cold_plains` does not exist.
- [ ] **Step 3: Implement `cold_plains.generate(level_id, seed, source_direction, difficulty)` in Lua.** Validate Act 0/DRLG 3/LevelType 2 and 8-tile dimensions from the selected mode. Validate candidate rows 29/30 are generic outdoor 8×8 stone fills with nonempty variants and DT1 paths. Select candidates and variants with named Cold Plains deterministic streams and per-cell keys; emit a cardinal opposite-edge midpoint entry and a deterministic tile path to the map interior. Keep output order deterministic and label layout/entry placement as Dark Magic policy.
- [ ] **Step 4: Run the focused tests to green.** `go test ./internal/mod/d2legacy/adapter/mapgen -run ColdPlains -count=1`; expected PASS.
- [ ] **Step 5: Commit.** `git add internal/content/d2legacy/lua/d2legacy/mapgen/cold_plains.lua internal/mod/d2legacy/adapter/mapgen/cold_plains_test.go && git commit -m 'feat(mapgen): generate Cold Plains zone recipe'`.

### Task 2: Owned-asset materialization and collision path

**Files:**
- Create: `internal/game/world/cold_plains_real_test.go`
- Modify if the owned-asset test exposes an issue: `internal/content/d2legacy/lua/d2legacy/mapgen/cold_plains.lua`

**Interfaces:**
- Consumes: `GenerateEntryWorld`, Blood Moor `Zone.Warps()`, headless `Runtime.Generate`, `materializeRealZone`, `Map.OpenPointNearSubtile`, `Map.OpenPointNearCenter`, and `Map.FindPath`.
- Produces: an optional local integration proof using owned MPQ assets; no committed asset bytes.

- [ ] **Step 1: Write the owned-asset test.** Reuse `openRealWorldContent`, `openMapgenRuntime`, and `materializeRealZone`. Generate/admit Blood Moor, find exactly one `next-level-exit` with destination 3, pass its direction to the Cold Plains strategy, then assert 80×80 Normal map, nonempty tiles, convert the warp tile coordinate to its center subtile (`tile * SubtilesPerTile + 2.5`) and assert an open arrival no more than 10 subtiles from that center, an open goal at least one 8-tile room into the interior, and nonempty `FindPath` from arrival to that goal. `OpenPointNearSubtile` alone is insufficient because its search may roam across the map. The existing helper skips clearly when `DARK_MAGIC_TEST_MPQ_DIRECTORY` is unset.
- [ ] **Step 2: Run locally with owned data to expose any failure.** `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/game/world -run TestGeneratedColdPlains -count=1 -v`; expected PASS after any necessary issue-scoped recipe correction.
- [ ] **Step 3: If red, diagnose the first asset or collision failure, adjust only Cold Plains policy, and rerun the focused test.** Preserve the failing evidence in the task log; do not alter generic materialization or unrelated outdoor generation without splitting scope.
- [ ] **Step 4: Commit.** `git add internal/game/world/cold_plains_real_test.go internal/content/d2legacy/lua/d2legacy/mapgen/cold_plains.lua && git commit -m 'test(mapgen): verify Cold Plains owned-asset path'`.

### Task 3: Final verification, review, and PR

**Files:**
- Modify only if test-proven behavior requires it: `docs/MAP_GENERATION.md` (document this structural Cold Plains slice and remaining exit/seed gaps).

**Interfaces:**
- Consumes: issue #618, final diff, repository PR template, review findings.
- Produces: issue-scoped PR with exact checks and owned-data status, ready for CI-gated merge.

- [ ] **Step 1: Update the map-generation documentation if needed, then commit that issue-scoped change.** Distinguish structural Cold Plains generation from original layout and live travel.
- [ ] **Step 2: Request independent architecture/evidence/code and formatting/readability reviews.** The readability reviewer uses the `refactor-for-readability` skill. Resolve blockers within #618.
- [ ] **Step 3: Rebase on latest main and run final checks on the final head.** `go test ./...`, `make architecture`, `make test-lua-hardening`, `make test-race`, `git diff --check origin/main...HEAD`, and the owned-asset focused test. Record exact results and skips.
- [ ] **Step 4: Complete every PR-template section, push, open and attach the PR, wait for CI, then merge only when required checks pass.** Confirm #618 closes and #354 stays open.
