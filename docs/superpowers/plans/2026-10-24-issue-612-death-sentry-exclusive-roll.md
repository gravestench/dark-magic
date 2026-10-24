# Death Sentry Exclusive Corpse Roll Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Death Sentry's corpse explosion draw below its upper endpoint without changing other damage rolls.

**Architecture:** Keep the exception in `d2legacy.systems.trap_skill`, adjacent to `explode_corpse`. Use the existing authoritative damage-roll stream and fixed skill records in authority suites; leave the shared inclusive policy intact.

**Tech Stack:** First-party Lua gameplay, deterministic `engine.authority_random/v1`, Go-hosted Lua authority suites.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-612-death-sentry-exclusive-roll-design.md`

## Global Constraints

- Issue #612 alone owns this change; #610's split, other trap rolls, and parent #420 remain untouched.
- 1.13c static evidence supports the exclusive bound; applying it to 1.14d remains an inference.
- The shared stream name is exactly `d2legacy.combat.damage.roll`; equal and reversed ranges consume no draw.
- Do not add a product API, change supported skill coverage, or commit owned game data, captures, or saves.

## Review Focus

- Ordered 101/107 endpoints must never produce 107: Task 1 asserts the deterministic 104 event total.
- A range narrowed by percentage truncation to 10/10 must not draw: Task 1 asserts a next-stream value of 132.
- Reversed 107/101 endpoints must return 107 without asserting or drawing: Task 2 asserts damage and next-stream value.
- Splitting an odd selected amount must truncate both channels separately: Task 1 asserts 52/52 from 105.
- A consumed corpse and its RNG state must survive checkpoint continuation: both tasks assert corpse removal and checkpoint parity.

---

### Task 1: Ordered and Equal Corpse Transactions

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill.lua:1-18,471-478`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill_test.lua:65-91,285-390`

**Interfaces:**
- Consumes: `random.integer(stream, positive_bound)` from `engine.authority_random/v1`.
- Produces: local `roll_corpse_damage(minimum, maximum)` returning the selected scalar; no exported API.

- [ ] **Step 1: Write the failing authority tests.** Fix helper 904 at `Param1=50`, `Param2=53`; make `install_transaction(corpse_max_health)` use 202 by default and the supplied value for the corpse's max health. In the ordered transaction (202), assert corpse removal, `physical_rolled_raw=52`, `fire_rolled_raw=52`, and `rolled_damage_raw=104`. Add `test.case("death_sentry_equal_corpse_range_skips_random_draw", ...)` with max health 20; assert 5/5, event total 10, corpse removal, checkpoint parity, and then `random.integer("d2legacy.combat.damage.roll", 1000) == 132`.
- [ ] **Step 2: Prove RED.** Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/systems/trap_skill' -count=1 -v`; confirm the ordered case reports 106 rather than 104 and the equal case detects the unwanted draw.
- [ ] **Step 3: Implement the local helper.** In `trap_skill.lua`, import `engine.authority_random/v1`. `roll_corpse_damage(minimum, maximum)` returns `minimum` when `maximum <= minimum`; otherwise returns `minimum + random.integer("d2legacy.combat.damage.roll", maximum - minimum)`. Replace only `explode_corpse`'s call to `damage.roll`.
- [ ] **Step 4: Prove GREEN and inspect execution.** Repeat the focused command with `-v`, confirm `death_sentry_centers_ordered_corpse_damage_on_the_consumed_corpse` and `death_sentry_equal_corpse_range_skips_random_draw` ran and passed, then run `git diff --check`.
- [ ] **Step 5: Commit.** Stage only the two files above and commit the tested slice.

### Task 2: Reversed Endpoint Authority Regression

**Files:**
- Create: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill_reversed_corpse_test.lua`

**Interfaces:**
- Consumes: Task 1's local corpse-roll behavior through the registered authoritative trap system.
- Produces: one supplemental authority suite using fixed supported skill ID 276, not a new gameplay interface.

- [ ] **Step 1: Write the reversed fixture.** Add `test.case("death_sentry_reversed_corpse_range_skips_random_draw", ...)`. Copy the minimum fixed records needed from the existing trap suite: Death Sentry 276, corpse helper 904, lightning helper 905, `death-shot` missile, trap PetType, sentry MonStats/MonStats2 under both uppercase and lowercase table paths, level 1 MonLvl, and Death Sentry SkillDesc. The uppercase monster tables feed trap decoding; the lowercase aliases feed monster admission. Set helper 904 to `Param1=53`, `Param2=50`. Install one sentry, one usable corpse with max health 202, and one victim through the authority harness. Assert the corpse is consumed, staged physical/fire values are 53/53, event total is 106, checkpoint parity holds, and the next `random.integer("d2legacy.combat.damage.roll", 1000)` is 132. Keep the suite small; share test-only fixture code only if duplication obscures these assertions.
- [ ] **Step 2: Prove the test detects the regression.** Temporarily compare against the pre-fix `damage.roll` call in a throwaway worktree or uncommitted one-line reversal, run the focused new suite, observe its reversed-range assertion fail, and restore Task 1 before committing. Do not carry the deliberate failure into the commit.
- [ ] **Step 3: Prove GREEN.** Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/systems/trap_skill' -count=1 -v`; confirm `death_sentry_reversed_corpse_range_skips_random_draw` ran and passed, then run `git diff --check`.
- [ ] **Step 4: Commit.** Stage only the new test file and commit the tested slice.

## Final Verification and Delivery

- [ ] Run `make test`, `make architecture`, `make test-race`, `make test-lua-format test-lua-syntax`, `make test-lua-hardening`, `make lint GOLANGCI_LINT='go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2'`, and `make build-client-backends` on the final submitted commit. Record exact results and whether owned-data tests ran.
- [ ] Have independent correctness/evidence reviewers compare the complete diff and issue claims against the exercised path; have a style reviewer use `refactor-for-readability` and address concrete findings within #612.
- [ ] Complete every PR-template section, link `Closes #612` and `Refs #420`, push, and merge only after required CI passes.
