# Issue #597 Shared Act-Town Levels Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the existing shared town predicate return the five original 1.13c act-town level IDs and prove the corrected answer reaches current gameplay consumers.

**Architecture:** Keep `movement.IsTownLevel(int64) bool` and its `d2legacy.movement_rules/v1` bridge as the single existing classification path. Change its fixed set in place; authority, connected prediction, and poison retain their current calls and policies.

**Tech Stack:** Go, gopher-lua, authoritative Lua suite, synthetic ECS/player fixtures.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-597-town-levels-design.md`; issue #597; `docs/specs/player-stamina-drain-while-running.md` SDR-R02/V06; `docs/research/spec-evidence/combat-knockback.md` KNB-R03b.

## Global Constraints

- Issue #597 only: town membership and tests of existing consumers. Do not change stamina arithmetic, room lookup, poison/aura/corpse policies, or dependencies.
- Original 1.13c membership is `{1,40,75,103,109}`; applying it to 1.14d remains inferred without owned observation.
- No game assets, real captures, private saves, credentials, or proprietary data in the commit.
- PR description must complete every template section and report exact checks on the submitted commit, including whether real-asset tests ran.

## Review Focus

- Level 1 still receives town recovery/protection: direct predicate and existing regression tests.
- Later towns 40/75/103/109 classify as town: direct predicate and bridge tests; level 40 production cases.
- Former false positives 45/80/108/114 classify as outside: direct predicate and bridge tests; level 45 production cases.
- Zero, negatives, neighbors, and above-range IDs remain outside: direct predicate exhaustive 0..120 plus negative/out-of-range assertions.
- Authority and connected prediction agree on one running tick: explicit 2,602/2,520 raw stamina cases in both paths.

---

### Task 1: Correct and prove shared town classification

**Files:**
- Modify: `internal/mod/d2legacy/adapter/movement/rules.go` (`IsTownLevel` comment and body only)
- Modify: `internal/mod/d2legacy/adapter/movement/rules_test.go` (rename stale `TestTownLevelsAreTheFiveExpansion114dActTowns` and replace expected set)
- Modify: `internal/mod/d2legacy/adapter/movement/testdata/module_test.lua` (Lua bridge assertions)
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/player_stamina_test.lua` (authority cases)
- Modify: `internal/app/clientapp/client_prediction_test.go` (connected prediction case)
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/poison_tick_test.lua` (representative non-stamina case)

**Interfaces:**
- Consumes: existing `IsTownLevel(level int64) bool`, `d2legacy.movement_rules/v1`.is_town, `newPositionPrediction(...).advanceTick(uint64)`, and current Lua suite fixtures.
- Produces: no new public interface; corrected membership through those existing calls.

- [ ] **Step 1: Write direct predicate and bridge tests.** Rename the Go test to claim only recovered 1.13c membership. Assert exactly `{1,40,75,103,109}` for every integer level 0..120, plus negative and large out-of-range values. In `module_test.lua`, assert the five true values and old false positives through `rules.is_town`.
- [ ] **Step 2: Write production regression tests.** In `player_stamina_test.lua`, enter a player at level 40 or 45, step the entry tick, then set `stamina_raw` to 2,560 before submitting one running movement tick. With max stamina 84 and RunDrain 20, assert 2,602 at level 40 and 2,520 at level 45. In `client_prediction_test.go`, seed a HUD with `StaminaRaw=2560`, `MaxStaminaRaw=84*256`, `RunDrain=20`, `Running=true`, moving velocity, and each level; call `newPositionPrediction(...).advanceTick(11)` and assert the same two `staminaRaw` results. In `poison_tick_test.lua`, add otherwise identical active poisoned players in levels 40 and 45 to the existing town case, asserting protection at 40 and health loss at 45 after two ticks while retaining dead/inactive assertions.
- [ ] **Step 3: Run the new tests red before changing production.** Run `go test ./internal/mod/d2legacy/adapter/movement -run 'TestTownLevels|TestRulesModule' -count=1`, `go test ./internal/app/clientapp -run TestPredictionTownStamina -count=1`, and `DARK_MAGIC_LUA_TEST_TIERS=fast go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`. Each affected path must fail for the old town set; inspect any unrelated failure before proceeding.
- [ ] **Step 4: Implement the minimal correction.** Make `IsTownLevel` return true exactly for 1, 40, 75, 103, 109. Update its comment to cite recovered 1.13c membership and avoid a verified 1.14d claim. Do not modify current callers.
- [ ] **Step 5: Run focused tests green.** Re-run the three Step 3 commands; all must exit 0. Run `make test-lua` and `go test ./...` for regression coverage. Fix only issue #597 failures, with new failing tests first.
- [ ] **Step 6: Verify and commit.** Run `make architecture`, `make test-lua-format`, and `git diff --check`; inspect the complete diff against issue #597, then commit the six test/production files. Record exact command results and unrun real-asset tests for the PR.

## Integration

Request independent architecture and evidence/test review of the final branch. Apply substantive findings with a new failing test where behavior changes. Rebase on current `origin/main`, run the final applicable checks on the submitted commit, write every PR template section before submission, push, attach the PR, wait for all CI checks, and merge only when green and mergeable as already authorized by the user.
