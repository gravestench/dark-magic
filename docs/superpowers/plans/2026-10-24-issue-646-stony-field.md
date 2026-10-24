# Deterministic Stony Field Zone Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Admit and materialize a deterministic, connected Stony Field level-4 zone from authored Act I assets.

**Architecture:** Add a dedicated Lua map policy reachable through the existing `Runtime.Generate` transport. It receives the future Cold Plains exit direction explicitly, validates all candidate records, and emits one provisional structural recipe; the current live entry world and seams stay untouched.

**Tech Stack:** d2legacy Lua map policy, deterministic/worldgen modules, Go headless runtime tests, owned 1.14d MPQ materialization test.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-646-stony-field-design.md`

## Global Constraints

- Issue #646 alone owns this change. The future Cold Plains exit, live seam, snapshots, monsters, and quest set pieces remain separate.
- Stony Field is Act I level 4, outdoor DRLG type 3, level type 2, 80×80 tiles in all three difficulties in inspected owned Expansion 1.14d and recovered 1.13 tables.
- Generic 8×8 outdoor `LvlPrest` definitions 29 and 30 have three DS1 variants each in the inspected owned 1.14d table. Their selection and arrangement are provisional Dark Magic policy, not exact original placement.
- Exactly one semantic `cold-plains-entry` targets level 3 on the edge opposite the explicitly supplied source direction. Do not infer that direction from an unplaced Cold Plains exit.
- Use the existing Go transport and generic world materializer. Do not add Diablo policy to Go, change dependencies, or commit game data, extracts, captures, or saves.

## Review Focus

- A valid but unselected fill candidate with a missing declared DS1 variant must reject admission; Task 2 tests it.
- A selected DT1 mask bit whose `LvlTypes` path is missing must reject admission; Task 2 tests it.
- Missing or diagonal source direction and difficulty outside 0–2 must fail before any draw; Task 2 tests them.
- Rotating the source direction must rotate the sole entry, keep it on the boundary, and alter the checksum; Task 1 tests it.
- A semantic path alone can cross blocked tiles; Task 3 tests the materialized arrival and collision path into the interior.

---

### Task 1: Stony Field Recipe and Positive Headless Tests

**Files:**
- Create: `internal/content/d2legacy/lua/d2legacy/mapgen/stony_field.lua`
- Create: `internal/mod/d2legacy/adapter/mapgen/stony_field_test.go`

**Interfaces:**
- Consumes: `Runtime.Generate(ctx, "stony_field", float64(4), float64(seed), sourceDirection, float64(difficulty))` and existing `d2legacy.mapgen.records`, `assets`, `act1_dirt_path`, `engine.deterministic/v1`, and `engine.worldgen/v1`.
- Produces: `M.generate(level_id, seed, source_direction, difficulty)` returning an admitted `worldgen.Zone` with role `cold-plains-entry` targeting level 3.

- [ ] **Step 1: Write failing positive tests.** Add a small fixed record fixture with `Levels` row 4 (Act 0, DRLG 3, type 2, 80×80 in all modes), `LvlTypes` row 2, and `LvlPrest` 29/30 with three valid 8×8 outdoor DS1 variants and valid DT1 mask. `TestStonyFieldProducesConnectedZone` loops difficulties 0–2 and four source directions; assert request, bounds, ordered room/stamp count, connected links, exactly one edge warp of role `cold-plains-entry` targeting 3 with opposite direction, a semantic path to the center, and trace naming the inspected 1.14d source and provisional defs 29/30 rationale. `TestStonyFieldChecksumSurvivesRestore` asserts repeated checksum, `worldgen.NewZone(first.Definition())` parity, and changed-direction checksum.
- [ ] **Step 2: Prove RED.** Run `go test ./internal/mod/d2legacy/adapter/mapgen -run 'TestStonyField(ProducesConnectedZone|ChecksumSurvivesRestore)$' -count=1 -v`; expect module `d2legacy.mapgen.stony_field` missing.
- [ ] **Step 3: Implement the positive strategy.** In `stony_field.lua`, add `M.generate(level_id, seed, source_direction, difficulty)` with the row/size/candidate parsing needed for the valid fixture, unique named `d2legacy.mapgen.stony_field.fill` and `.fill_variant` streams, canonical 8×8 grid, orthogonal links, one opposite-facing midpoint edge entry, and `act1_dirt_path.apply` route to center. Name the 1.14d source and why defs 29/30 qualify in trace; label arrangement and entry as Dark Magic policy. Reuse existing mapgen helpers; leave `entry_world.lua` and Cold Plains untouched. Task 2 supplies the complete rejection contract.
- [ ] **Step 4: Prove GREEN.** Repeat the focused command, confirm both named tests ran and passed, then run `git diff --check`.
- [ ] **Step 5: Commit.** Stage only these two files and commit the tested slice.

### Task 2: Reject Malformed Inputs and Unselected Assets

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/mapgen/stony_field.lua`
- Modify: `internal/mod/d2legacy/adapter/mapgen/stony_field_test.go`

**Interfaces:**
- Consumes: Task 1's `M.generate` and fixed synthetic records.
- Produces: strict pre-draw admission failures without changing the successful recipe contract.

- [ ] **Step 1: Write admission cases.** Add `TestStonyFieldRejectsInvalidInputsAndRecords` with table cases for level 3, direction empty/diagonal, difficulty -1/3, wrong Act/DRLG/type, size 0/81, missing type or either preset, wrong preset level/outdoor flag/size, declared `Files=2` with blank `File2` on one candidate, a directory or absolute DS1 path, DT1 mask 3 with missing `File 2`, and invalid/zero mask. Require an error for each before any seed draw so no candidate can hide by not being selected.
- [ ] **Step 2: Check the admission cases before this task's guard changes.** Run `go test ./internal/mod/d2legacy/adapter/mapgen -run '^TestStonyFieldRejectsInvalidInputsAndRecords$' -count=1 -v`; identify any malformed inputs that Task 1 still admits. If Task 1 already rejects all cases, record that result rather than manufacture a failure.
- [ ] **Step 3: Tighten admission.** Validate difficulty and direction before drawing. Validate every candidate's level/outdoor/grid fields, declared DS1 count and each path, mask range and every selected DT1 path before `assets.preset_files` / `assets.masked_tiles` normalize them. Reject any malformed candidate even when unselected. Preserve Task 1's positive output.
- [ ] **Step 4: Prove GREEN.** Repeat the admission command and Task 1's focused positive command; run `git diff --check`.
- [ ] **Step 5: Commit.** Stage only the two Task 2 files and commit.

### Task 3: Owned-Asset Collision Route

**Files:**
- Create: `internal/game/world/stony_field_real_test.go`

**Interfaces:**
- Consumes: `Runtime.Generate(ctx, "stony_field", 4, seed, "east", 0)` and existing `openRealWorldContent`, `openMapgenRuntime`, `materializeRealZone`, `OpenPointNearSubtile`, `OpenPointNearCenter`, and `FindPath` helpers.
- Produces: one owned 1.14d materialization and collision-route regression, skipped when the MPQ fixture is unavailable.

- [ ] **Step 1: Write the owned-data test.** `TestGeneratedStonyFieldMaterializesAConnectedEntry` follows `cold_plains_real_test.go`: generate seed 17 Normal with explicit east source direction, materialize, assert 80×80 and nonempty tiles, exactly one west-facing `cold-plains-entry` targeting 3, an in-bounds unblocked arrival within 10 subtiles of the warp, an open center at least 8 tiles from arrival, and nonempty `FindPath` from arrival to center. State that the direction and layout are provisional policy; do not assert an existing Cold Plains exit.
- [ ] **Step 2: Prove RED or observe the materialization failure.** With `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq`, run `go test ./internal/game/world -run '^TestGeneratedStonyFieldMaterializesAConnectedEntry$' -count=1 -v`. If it already passes against Task 2, record that the test adds owned-asset coverage rather than inventing a production failure; do not force a false RED.
- [ ] **Step 3: Correct only issue-scoped materialization defects if found.** Keep authored asset selection and route within the #646 strategy; do not change generic collision or the live Cold Plains seam for this test.
- [ ] **Step 4: Prove GREEN and skip behavior.** Repeat the owned-data command and run the same test without `DARK_MAGIC_TEST_MPQ_DIRECTORY` to confirm explicit skip; run `git diff --check`.
- [ ] **Step 5: Commit.** Stage only the new owned test and any necessary #646 strategy correction; commit.

## Final Verification and Delivery

- [ ] Run focused generator and owned world tests, `make test`, `make architecture`, `make test-race`, `make test-lua-format test-lua-syntax test-lua-hardening`, pinned `make lint GOLANGCI_LINT='go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2'`, and `make build-client-backends` on the final submitted commit. Record actual results, including owned-data execution and CI skip.
- [ ] Obtain independent correctness/evidence reviews of the whole issue diff and a formatting/readability review using `refactor-for-readability`; address concrete findings within #646.
- [ ] Complete every PR-template section, use `Closes #646` and `Refs #354`, rebase on current `main`, push, and merge only after required CI passes. Follow with a separate live Cold Plains↔Stony Field seam issue.
