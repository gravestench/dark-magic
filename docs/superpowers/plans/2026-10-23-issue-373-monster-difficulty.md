# Monster Difficulty Session Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove that the immutable session difficulty reaches generated monster level, life, armor, XP, and treasure-class selection for issue #373.

**Architecture:** Use separate authoritative sessions with synthetic MonStats, MonLvl, Levels, and MonStats2 rows. Activate the same room with one or two players in each difficulty and inspect the resulting monster ECS facts. A separate session submits an invalid bootstrap difficulty to prove the immutable rule is not overwritten.

**Tech Stack:** Go integration tests in `internal/mod/d2legacy`, existing Lua `game_rules`, monster catalog, room population, and spawn materializer.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-373-monster-difficulty-design.md`

## Global Constraints

- Linked issue: #373, partial difficulty acceptance; use `Refs #373` and keep it open for 1.14d runtime evidence and loot outcomes.
- The rule is verified in 1.13c static code and inferred for 1.14d. No protected game data or captures may be committed.
- If production-path assertions already pass, commit the test without unrelated behavior edits.

## Review Focus

- Nightmare and Hell must take `Levels.MonLvl2Ex` and `MonLvl3Ex`, not MonStats `Level(N)/(H)`.
- HP, AC, and XP must come from the selected level's respective MonLvl difficulty columns and MonStats percentages.
- Two living players must add 50% to life and XP in all three difficulties without changing level or armor.
- The selected `TreasureClass1`, `(N)`, or `(H)` is a spawn input; do not claim a loot-roll result.
- A mismatched bootstrap difficulty must fail before changing the immutable game rules or population plan.

---

### Task 1: Assert difficulty through generated-monster production

**Files:**
- Create: `internal/mod/d2legacy/monster_difficulty_session_test.go`
- Modify only if a test fails within this boundary: `internal/content/d2legacy/lua/d2legacy/data/monster.lua`, `bootstrap/population.lua`, or `commands/spawn_monster.lua`.

**Interfaces:**
- Consumes: `StartWithConfig`, `system.population.bootstrap`, `system.player.enter`, `monsterEntity`, and ECS dynamic stores.
- Produces: no new product interface; an executable authority-session matrix.

- [ ] **Step 1: Write the matrix test.** Add synthetic Levels row `Id=2`, `NumMon=1`, `mon1=nmon1=fallen`, `MonDen`, `MonDen(N)`, `MonDen(H)` all positive, `MonLvl2Ex=4`, `MonLvl3Ex=5`; ordinary Expansion MonStats `noRatio=0`, levels `1/2/3`, fixed equal min/max HP percentages `100/120/140`, AC percentages `100/120/150`, XP percentages `100/150/200`, and distinct `TreasureClass1` values. MonLvl level 1 has Normal HP=10, AC=20, XP=50; level 4 has Nightmare HP=20, AC=30, XP=80; level 5 has Hell HP=30, AC=40, XP=120. Include the graphics row and any minimal record tables needed to start the authority. For each difficulty 0/1/2 and live count 1/2, start `game_rules` with that difficulty, admit the players in the same tick as a matching population bootstrap, and assert one generated monster's level, raw health, defense, XP, pinned count, and treasure class.
- [ ] **Step 2: Check the exact matrix.** One-player expected `(level, health/256, defense, XP)` values: Normal `(1,10,20,50)`, Nightmare `(4,24,36,120)`, Hell `(5,42,60,240)`. Two-player expected health/256 and XP: Normal `(15,75)`, Nightmare `(36,180)`, Hell `(63,360)`; level, defense, and treasure class remain the same. Run `GOCACHE=/private/tmp/dark-magic-issue373-difficulty-gocache GOMAXPROCS=2 GOFLAGS=-p=2 go test ./internal/mod/d2legacy -run '^TestGeneratedMonsterUsesSessionDifficulty$' -count=1 -v`. If GREEN on existing code, retain test-only production coverage. If RED due a real within-boundary defect, keep failure output and fix the minimum behavior before repeating the test.
- [ ] **Step 3: Add mismatch test in its own session.** Start Hell game rules, submit `system.population.bootstrap` with difficulty 1 and level 2, expect command-apply failure, and read `d2legacy.game_rules` plus `d2legacy.population.plan` to prove difficulty 2 and `installed=false` remain. Do not continue this failed-command session.
- [ ] **Step 4: Run the focused matrix and mismatch tests, then commit.** Run `git diff --check` and the two focused tests. Commit only the test and a within-boundary correction if Step 2/3 exposed one.

### Task 2: Verify and publish the issue slice

**Files:**
- Update: GitHub issue #373 after merge.
- Modify only if a check or review finds a scoped defect: files from Task 1.

**Interfaces:**
- Consumes: Task 1's committed result.
- Produces: a PR with exact evidence, checks, and remaining gates.

- [ ] **Step 1: Run broad gates on submitted commit.** `make test`, `make architecture`, `make test-lua-hardening`, `make test-race`, and `git diff --check`, with the task GOCACHE/GOMAXPROCS/GOFLAGS above. State whether `MPQ_DIRECTORY` allowed real-asset tests.
- [ ] **Step 2: Have independent agents review the final diff.** Require production-path and evidence reviews; correct any concrete findings and rerun relevant gates on final commit.
- [ ] **Step 3: Rebase, push, submit, and merge.** Rebase on latest `origin/main` after PR #558 merges, push, open PR with every template section and `Refs #373`, attach the PR artifact, wait for all required CI checks, merge, then update issue #373's delivered evidence and remaining 1.14d/loot questions.
