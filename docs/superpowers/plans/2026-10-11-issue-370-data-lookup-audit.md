# Issue #370 Game-Data Lookup Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every currently admitted `d2legacy` table relationship reviewable by identity kind and protect representative divergent identities with consumer-level tests.

**Architecture:** Keep Diablo table-link policy in the existing Lua consumers and narrow recovered-data adapter. Record the audit in the existing game-data research document. Change a consumer only when a test demonstrates an incorrect row choice; do not add a generic catalog or infer links in the record store.

**Tech Stack:** Go 1.25, the `d2legacy.tests/v1` Lua suite, Markdown research ledger, Makefile checks.

**Spec:** `docs/superpowers/specs/2026-10-11-issue-370-data-lookup-audit-design.md`

## Global Constraints

- Issue #370 is the only acceptance boundary; #342 is its parent tracker.
- Keep source provenance and pinned generation attached to data, separate from lookup identity.
- Preserve uncertain Expansion 1.14d TXT-to-BIN behavior and unverified 1.13c bytes as explicit unknowns.
- Do not commit owned Diablo II data, captures, saves, or credentials.
- Avoid refactors, dependency changes, or generic record-store policy.

## Review Focus

- A numeric skill ID equal to another row's ordinal must select the row with matching `Skills.Id`; Task 2 adds a conflicting fixture.
- A symbolic skill-description reference must select matching `SkillDesc.skilldesc` despite differing row positions; Task 2 adds a conflicting fixture.
- A `LvlTypes` positional reference must keep the null row and zero-based offset even when an authored-looking `Id` disagrees; Task 2 adds a conflicting fixture.
- A sound redirect to a grouped row must start selection at the redirected position; Task 3 adds a failing consumer test and fixes the current index mix-up.
- An invalid or absent reference must not silently select an unrelated row; Tasks 2 and 3 add negative/sentinel assertions and Task 1 records each admitted rule.

---

### Task 1: Inventory and ledger

**Files:**
- Modify: `docs/research/GAME_DATA_LOADING_AND_LINKAGE.md`

**Interfaces:**
- Consumes: production `records.load` call sites under `internal/content/d2legacy/lua/d2legacy`, `internal/mod/d2legacy/adapter/worldobjects/resolver.go`, #369 identity types, and the research source matrix.
- Produces: an audit table with consumer, source field, target field, identity kind, sentinel, evidence grade/source, and a discriminating test or evidence gap.

- [ ] **Step 1: Enumerate call sites.** Run `rg -n 'records\.load\(' internal/content/d2legacy/lua/d2legacy --glob '*.lua' | rg -v '_test\.lua'` and inspect each consumer; record joins/positional lookups and separately list single-table reads that need no cross-table identity.
- [ ] **Step 2: Inspect each admitted relationship.** Trace source value to target lookup, including `Skills`/`SkillDesc`/`Missiles`/`States`, monster tables, map generation, loot/item types, sound redirects, presentation, and recovered DS1 objects. Group only identical identity and sentinel rules.
- [ ] **Step 3: Write the ledger.** Add a bounded section to `GAME_DATA_LOADING_AND_LINKAGE.md` with the exact fields in `Interfaces`. Distinguish current Dark Magic behavior from original-runtime evidence; mark unresolved original TXT-to-BIN semantics and 1.13c data comparison as unknown.
- [ ] **Step 4: Check coverage.** Re-run the call-site search and compare every production file against either a ledger row or the single-table-read inventory; resolve omissions before proceeding.
- [ ] **Step 5: Commit.** Commit only the research document for this task.

### Task 2: Divergent numeric, symbolic, and positional identities

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/data/skill_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/mapgen/records_test.lua`

**Interfaces:**
- Consumes: `d2legacy.data.skill.load(id)` and `learned_for_ids(ids, level)`; `d2legacy.mapgen.records.level_type(id)`.
- Produces: production-consumer assertions that select the intended row with conflicting ordinal, numeric, and symbolic candidates.

- [ ] **Step 1: Make the skill fixture adversarial.** Put `Skills.Id = "1"` at a row whose zero-based ordinal is not 1, place a different skill at ordinal 1, and put the selected `SkillDesc.skilldesc` at a different ordinal than the skill. Add a case calling `skill.load(1)` that asserts the selected icon/name key and a case for a nonexistent ID returning `nil`.
- [ ] **Step 2: Run the focused Lua suite.** Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/data/skill' -count=1`. The test may pass immediately because this task proves an existing correct lookup; record that result honestly.
- [ ] **Step 3: Make the map fixture adversarial.** In `mapgen/records_test.lua`, give a different row an authored-looking `Id = "2"` while the intended `LvlTypes` row remains at zero-based ordinal 2. Assert the intended `File1`, and assert an out-of-range ordinal returns `nil`.
- [ ] **Step 4: Run focused and complete Lua suites.** Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/mapgen/records' -count=1`, then `make test-lua`. Record exact results.
- [ ] **Step 5: Commit.** Commit the two test files only.

### Task 3: Correct redirected sound group position

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/audio_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/audio.lua`
- Modify: `docs/research/GAME_DATA_LOADING_AND_LINKAGE.md` if the audit found different sentinel or evidence detail.

**Interfaces:**
- Consumes: `d2legacy.audio.play_record(name, seed)`; `Sounds.txt` `Sound`, `Redirect`, `Group Size`, and `Group Weight`.
- Produces: sound selection based on the redirected zero-based row ordinal and that row's grouped range.

- [ ] **Step 1: Add a failing consumer test.** Use three or more synthetic sound rows: a named redirect source, an unrelated neighboring row, and a redirected group with two variants. Give the source and target different group weights and filenames; assert `play_record` selects the target group's path for a fixed seed. Also assert an invalid redirect does not select an unrelated row.
- [ ] **Step 2: Verify red.** Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/audio' -count=1`; confirm failure is the selected path, not setup.
- [ ] **Step 3: Apply the minimal correction.** In `audio.lua`, update the group-start index when a valid `Redirect` moves `row` to `rows[redirect + 1]`; preserve the existing zero-based conversion and unknown-redirect behavior.
- [ ] **Step 4: Verify green.** Re-run the focused suite and `make test-lua`; confirm both pass.
- [ ] **Step 5: Commit.** Commit the consumer, its test, and any directly corresponding ledger correction.

### Task 4: Final evidence, review, and PR

**Files:**
- Modify: only an in-scope ledger/test/consumer file if review finds a demonstrated issue.
- Prepare: `.github/PULL_REQUEST_TEMPLATE.md` content in the PR body.

**Interfaces:**
- Consumes: the spec, ledger, issue #370 criteria, and the final diff.
- Produces: verified commit and a complete PR that closes #370 and references #342 only if the whole child boundary is met.

- [ ] **Step 1: Review each ledger row.** Check the owning call path, sentinel, evidence grade, and test pointer. For any newly confirmed wrong lookup, add a failing consumer test first and correct only that owner; update the ledger.
- [ ] **Step 2: Run final gates.** On the intended submitted commit, run `make test`, `make architecture`, `make test-race`, and `make test-lua-hardening`; record exact results and skips. Run relevant real-asset checks only if lawful local data is configured.
- [ ] **Step 3: Perform separate review.** Compare issue claims with assertions and production paths; check architecture, determinism, source provenance, dead code, scope, and performance. Correct findings and repeat affected gates.
- [ ] **Step 4: Commit final corrections.** Verify `git status --short` and inspect the full diff.
- [ ] **Step 5: Write the complete PR body.** Fill every template section with scope, evidence, exact checks on the final commit, unverified behavior, and AI/review use. Use `Closes #370` and `Refs #342` only when the child is complete.
- [ ] **Step 6: Push and open the PR.** Create/use a `codex/` branch, push it, open a PR against `main`, and attach its URL to this task.
