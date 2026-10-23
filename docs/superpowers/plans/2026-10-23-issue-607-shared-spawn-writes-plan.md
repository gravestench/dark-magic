# Shared Monster Spawn Write Contracts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every supported shared monster spawn succeed through its ECS writer without losing existing components.

**Architecture:** Keep `spawn_monster.components` as the component owner and explicit write lists in each consuming system. Repair only missing access declarations. Prove them through production authority systems, then use source inventory to distinguish admitted synthetic recipes from owned 1.14d authored rows.

**Tech Stack:** Lua ECS authority systems and suites, Go integration tests, owned expansion 1.14d MPQ records for optional reachability checks.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-607-shared-spawn-writes-design.md`

## Global Constraints

- Linked issue: #607; no changes to monster eligibility, definition facts, combat math, or unrelated systems.
- Existing MonStats facts and spawn component contents remain as authored; only ECS access declarations and focused tests change.
- Synthetic admitted recipes prove code reachability, not shipped-game occurrence. Owned 1.14d records prove table cells, not original-runtime behavior.
- Never commit licensed game data, captures, credentials, or private saves.
- Rebase onto latest `origin/main` after PR #606 merges and run final checks on the submitted commit.

## Review Focus

- Sentry skill cast reaches `trap.deploy` and retains combat defense, rather than passing only a direct builder test (Task 1).
- Population `boss` and `prime_evil` booleans reach the authority system without losing boss/freeze/prime markers (Task 2).
- Revived and fixed summoned definitions preserve any admitted conditional markers without changing corpse eligibility (Task 3).
- Golem/trap summon definitions have `corpse_selectable=false` and `revivable=false`; avoid claiming those marker paths can occur (Task 3).
- Restore/replay of generated population remains deterministic with classification markers (Task 2).

---

### Task 1: Reproduce and Repair Sentry Defense Write

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill.lua`

**Interfaces:**
- Consumes: `spawn_monster.components` returns `d2legacy.combat.defense` unconditionally.
- Produces: `d2legacy.trap.deploy` may create a sentry with that defense component.

- [ ] **Step 1: Write failing authority test.** Extend the trap suite's synthetic Death Sentry MonStats/graphics/MonLvl records with the `enabled` and runtime fields required by `monster.summon`. Enter a player and issue `player.use_skill` for the sentry (or install its admitted cast fact), advance to effect tick, and assert one created `d2legacy.trap.autonomous` entity retains `d2legacy.combat.defense`. Existing direct-installed sentry cases do not satisfy this step.
- [ ] **Step 2: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/systems/trap_skill' -count=1`; require the undeclared-write error to name `d2legacy.combat.defense`. If fixture validation fails first, correct fixture records and rerun RED.
- [ ] **Step 3: Add exact write.** Add `d2legacy.combat.defense` to `d2legacy.trap.deploy.write` without changing `sentry_components` or the materializer.
- [ ] **Step 4: Verify GREEN.** Run the focused command and `make test-lua`; expect exit 0 and retained defense assertion.
- [ ] **Step 5: Commit.** Stage only Task 1 files; commit `fix: declare sentry spawn defense write`.

### Task 2: Reproduce and Repair Population Conditional Markers

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua`
- Create or modify: `internal/mod/d2legacy/monster_spawn_writer_test.go`

**Interfaces:**
- Consumes: `monster.load` admission and `spawn_monster.components` conditional `d2legacy.monster.boss`, `d2legacy.monster.freeze_immune`, and `d2legacy.monster.prime_evil` marker keys.
- Produces: `d2legacy.population.active_rooms` write declarations that admit these markers when a valid candidate definition carries them.

- [ ] **Step 1: Write failing generated-population test.** Reuse `generatedHostileRecords` and its authority/bootstrap helper with an enabled, `isSpawn` synthetic MonStats candidate carrying boss and prime-evil flags. Assert the generated entity retains boss, freeze-immune, and prime-evil components after room activation. Include one restore/replay assertion using the existing lifecycle helper where practical; if this requires unrelated persistence changes, keep the test to creation and report the limit.
- [ ] **Step 2: Verify RED.** Run `go test ./internal/mod/d2legacy -run TestGeneratedBossSpawnWriter -count=1 -v`; require a missing conditional-marker write error, not an invalid fixture.
- [ ] **Step 3: Add exact writes.** Add the three marker component names to `d2legacy.population.active_rooms.write`.
- [ ] **Step 4: Verify GREEN.** Run the focused Go test and existing generated-hostile lifecycle tests; expect exit 0 and retained marker assertions.
- [ ] **Step 5: Commit.** Stage only Task 2 files; commit `fix: declare population boss marker writes`.

### Task 3: Verify Owned Spawn Reachability and Declarations

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/corpse_summon_skill.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/golem_summon_skill.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/corpse_summon_skill_test.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill_test.lua`
- Create if needed: `internal/content/d2legacy/lua/d2legacy/systems/golem_summon_skill_test.lua`

**Interfaces:**
- Consumes: `monster.summon` and `monster.revive` preserve boss/prime-evil booleans but force corpse-selectable/revivable false; `spawn_monster.components` maps boss to both boss and freeze-immune markers.
- Produces: only the write declarations needed by code-admitted owned recipes, with authority-level proof and no change to those recipes.

- [ ] **Step 1: Inventory final component tables.** Compare the common builder's unconditional and conditional keys with each owned caller's additions/removals and write list. Record explicitly which paths can and cannot emit corpse-selectable/revivable, boss/freeze-immune, and prime-evil; use current admission functions and optional owned MonStats rows, not assumptions from synthetic flags.
- [ ] **Step 2: Write failing owned authority cases.** Use one admitted synthetic Revive or fixed summon recipe for boss/prime-evil marker retention through corpse summon, and admitted synthetic fixed rows for golem and sentry where their skill fixtures can execute production writers. Each case must fail at the specific undeclared marker before the write is added. If a caller cannot reach a marker, document that exclusion and omit its code change/test.
- [ ] **Step 3: Add exact writes.** Add only reachable conditional marker names to the corresponding owned system `write` lists. Do not strip marker components or broaden summon eligibility.
- [ ] **Step 4: Verify GREEN.** Run `make test-lua` and focused owned spawn suites; expect exit 0 and retained marker assertions. Check `git diff --check`.
- [ ] **Step 5: Commit.** Stage only Task 3 files; commit `fix: declare owned monster marker writes`.

## Final Verification and Delivery

- [ ] Obtain independent code/architecture and evidence/acceptance subagent reviews of the complete #607 branch; fix findings and seek re-review approval.
- [ ] Rebase on current `origin/main` after #606 merges. If main advances before submission, rebase again and repeat all reported checks on the submitted commit.
- [ ] On final submitted commit run `go test ./...`, `make architecture`, `make test-lua-hardening`, `make test-lua-format`, `make test-lua-syntax`, and `git diff --check origin/main...HEAD`. Run any owned data test used for authored reachability with a local MPQ path and state whether local race ran.
- [ ] Fill every `.github/PULL_REQUEST_TEMPLATE.md` section with exact check results, scope, evidence/version limits, and AI review; push, open PR with `Closes #607`, wait for all required CI jobs, merge, and verify closure.
