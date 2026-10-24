# Stony Field Ordinary Population Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Populate eligible Stony Field rooms with ordinary monsters through the existing authoritative room plan, without changing earlier-area outcomes.

**Architecture:** Extend the level-4 branch of Lua `bootstrap.population` to use the level-3 room-local placement and return safeguards, with its own named random streams. Publish collision-safe room-centered points for level 4 through the current Go adapter. Keep the same entity, room, and checkpoint owners.

**Tech Stack:** Go host/entryworld, d2legacy Lua authority, Go offline/connected integration tests, owned 1.14d MPQs.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-654-stony-field-population-design.md`

## Global Constraints

- Issue #654 is the only implementation boundary. It depends on #648; rebase onto merged `main` before submitting this issue's PR.
- Level 2 retains its unqualified random streams and placement; level 3 retains its `level3` names and outcomes. Level 4 gets `d2legacy.population.level4.{density,family,group,seed,life}`.
- Lua owns candidate, density, rarity, group, and random policy. Go publishes geometry only. Do not add a parallel spawn owner or population-plan schema.
- Owned 1.14d row facts are verified; exact retail spawn arithmetic, special packs, and cross-binary checkpoint migration are outside the issue.
- Do not commit game data, captures, saves, credentials, dependencies, or unrelated refactors.

## Review Focus

- First visit to a populated level-4 room creates monsters once; returning or restoring does not duplicate them. Task 1 and Task 3 test this.
- A blocked candidate or an inactive returning resident cannot overlap a fresh spawn. Task 1 tests collider footprints and reservation.
- Visiting level 4 must not advance levels 2/3 or the legacy monster-life stream. Task 1 compares earlier residents' IDs, positions, health, and stream values.
- Difficulty selection must be distinguishable despite identical owned Stony Field columns. Task 1 uses distinct synthetic lists/densities; Task 2 checks owned facts.
- A legacy v1/v2 topology restored under the new binary must not gain level 4 population. Task 3 tests it; cross-binary stream migration is not claimed.

---

### Task 1: Lua ordinary population and isolated randomness

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua`
- Modify: `internal/mod/d2legacy/identity.go`
- Modify: `internal/mod/d2legacy/cold_plains_population_test.go`

**Interfaces:**
- Consumes: existing ordered `system.population.bootstrap` zones, `monster.load`, `spawn.components`, `collision.destination_clear`, and room residency state.
- Produces: first-activation level-4 monsters under `level:4:room:<id>:monster:<member>` identities and five dedicated level-4 random streams; existing plan schema and command interface remain unchanged.

- [ ] **Step 1: Write failing synthetic authority tests.** Extend `coldPopulationSession` to install level-4 collision geometry alongside levels 2 and 3 before expecting collision-safe level-4 spawns. Replace `TestStonyFieldPopulationItemLifecycleWithoutRandomDraws` with item-plus-monster lifecycle assertions. Use distinct `mon*`/`nmon*` in a synthetic level-4 row to prove Normal versus Nightmare/Hell families. Use at least two eligible active level-4 rooms and density columns of `0` versus `100000` to prove difficulty density selection in the later room; assert the first room creates a resident so `plan.created` is nonzero before evaluating the later room. Cover `NumMon` truncation, ordinary eligibility, no candidates, invalid candidate admission, blocked/duplicate/out-of-room points, queued collider overlap, inactive returning footprints, stable IDs/health after return and same-binary checkpoint restore. Extend `TestStonyFieldPopulationPreservesLevelTwoAndThree` so the player enters level 4, then compare earlier-area IDs, positions, health and only their relevant stream values.
- [ ] **Step 2: Prove RED.** Run `go test ./internal/mod/d2legacy -run 'TestStonyFieldPopulation' -count=1 -v`; expect the new level-4 monster assertions to fail under the bypass.
- [ ] **Step 3: Implement the smallest Lua/registration change.** Remove the level-4 first-activation bypass; `stream` returns level-4 names only for level 4. Apply level-3 `safe_point`, point validation, queued occupancy, returning-footprint reservation, and named life stream to level 4. Register the five level-4 streams in `identity.go`. Do not change level-2/3 branches beyond a shared `level_id >= 3` predicate where semantics are identical.
- [ ] **Step 4: Prove GREEN.** Repeat the focused test; run `make test-lua-format test-lua-syntax test-lua-hardening` and `git diff --check`.
- [ ] **Step 5: Commit** only Task 1 files.

### Task 2: Room-local Stony Field geometry and owned admission

**Files:**
- Modify: `internal/mod/d2legacy/adapter/entryworld/population.go`
- Modify: `internal/mod/d2legacy/adapter/entryworld/stony_field_real_test.go`
- Modify or create only if useful: `internal/mod/d2legacy/adapter/entryworld/population_test.go`

**Interfaces:**
- Consumes: #648 `Prepared.PopulationData(nearby)` and `roomPopulationPoints`.
- Produces: four or fewer distinct open, room-local subtile candidates per level-4 room, retaining the existing zone envelope and order.

- [ ] **Step 1: Write failing geometry tests.** For level 4, assert candidate points lie within their own room, are open in the prepared map, and are deterministic across repeated calls; use an all-blocked room to assert an empty list. In the owned 1.14d fixture, inspect all difficulties and confirm the admitted level-4 zone has populated rooms with valid candidate points and no duplicate points.
- [ ] **Step 2: Prove RED.** Run `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy/adapter/entryworld -run 'TestStonyField|TestPopulationData' -count=1 -v`; record any already-passing generic assertions and the expected failing room-local case.
- [ ] **Step 3: Implement geometry selection.** Use `roomPopulationPoints` for levels 3 and 4, leaving level 2's anchor policy intact. Do not choose monster families or spawn timing in Go.
- [ ] **Step 4: Prove GREEN.** Repeat focused tests, run with owned MPQ variables unset to verify explicit skips, and run `git diff --check`.
- [ ] **Step 5: Commit** Task 2 files.

### Task 3: Production population, restore, and research evidence

**Files:**
- Modify: `internal/app/clientapp/stony_field_transition_test.go` or create a focused `stony_field_population_real_test.go`
- Modify: `internal/app/clientapp/cold_plains_population_real_test.go` only for shared fixture reuse
- Modify: `docs/research/MONSTERS_SPAWNING_AND_AI.md`

**Interfaces:**
- Consumes: Tasks 1–2 level-4 room population and #648 offline/connected crossing and recovery fixtures.
- Produces: owned-data production proof of authoritative/public level-4 monsters and uninterrupted/restored continuation, without adding production API.

- [ ] **Step 1: Write failing owned-data tests.** Find and pin a deterministic seed with eligible Stony Field rooms and residents; cross from level 3 offline and connected, assert the same level-qualified monster IDs/positions in authority and public projection, checkpoint in level 4, restore/reconnect, leave and return, and compare residents and random continuation to an uninterrupted session. Keep the existing level-4 item residency assertions. Restore a v1/v2 topology under the new binary and assert it has no level-4 zone or residents. Explicitly skip without owned MPQs.
- [ ] **Step 2: Prove RED or an already-working path.** Run `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/app/clientapp -run 'TestStonyField.*Population|TestStonyFieldLegacyV2Recovery' -count=1 -v` and record the exact results.
- [ ] **Step 3: Fix only observed production gaps.** Reuse current authority, public projection, and recovery paths. Add no second spawn or special-case UI. Update monster research with inspected row facts, current interpreter policy, test evidence, and unknown retail formulas.
- [ ] **Step 4: Prove GREEN.** Repeat focused owned-data tests and explicit no-MPQ skips; run `git diff --check`.
- [ ] **Step 5: Commit** Task 3 files.

## Final Verification and Delivery

- [ ] Rebase onto merged `main`; on the submitted commit run focused synthetic and owned-data tests, `make test`, `make architecture`, standalone `make test-race`, Lua format/syntax/hardening, pinned lint, both client backends, and `git diff --check`. Record exact results, skips, and any unverified graphical behavior.
- [ ] Obtain independent correctness/evidence reviews and a formatting/style review applying `refactor-for-readability`; resolve findings and rerun affected gates.
- [ ] Fill the complete PR template, use `Closes #654` and `Refs #354`, push, attach the PR, then merge only after required CI passes.
