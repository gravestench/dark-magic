# Death Sentry Independent Corpse Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the currently supported Death Sentry corpse explosion independently truncate its physical and fire raw damage shares.

**Architecture:** Keep the existing trap execution, mitigation, and event path. Add a controlled odd-unit integration fixture, then change only the physical calculation in `explode_corpse`.

**Tech Stack:** Lua gameplay and authority test harness, Go test runner.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-610-death-sentry-independent-split-design.md`

## Global Constraints

- Issue #610 only; the evidence is static 1.13c code, and 1.14d application remains an inference.
- Preserve corpse selection, roll, percent lookup, radius, mitigation, event creation, and consumption.
- Do not commit original game data, captures, credentials, or saves.

## Review Focus

- Odd raw amount with a 50% split: the damage bundle must show physical 50 and fire 50 from amount 101, with aggregate 100.
- Event attachment: assert channel values on `d2legacy.combat.damage_bundle`, aggregate and target on `d2legacy.combat.event`.
- Corpse lifecycle: the test must still see the consumed corpse removed.
- Replay: retain checkpoint parity after the transaction.
- Existing transaction behavior: retain the mixed-channel, target, consumption, and replay assertions while making its fixture deterministic.

---

### Task 1: Controlled split through the production trap path

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/trap_skill.lua:471-489`

**Interfaces:**
- Consumes: existing `install_transaction`, `test.case`, `d2legacy.combat.event`, and attached `d2legacy.combat.damage_bundle`.
- Produces: no new public interface; Death Sentry's existing event carries the independently truncated channel amounts.

- [ ] **Step 1: Write the failing test.** Make the existing `death_sentry_centers_ordered_corpse_damage_on_the_consumed_corpse` fixture deterministic: set its sole `install_transaction` corpse `max_health=202`, and the suite's Corpse Explosion Fixture `Param1=Param2=50`, `calc3=50`. In that case, assert bundle `physical_rolled_raw=50`, `fire_rolled_raw=50`; event `rolled_damage_raw=100`, `target_id="monster:victim"`, `source_kind="trap_corpse"`; retain corpse removal and checkpoint parity assertions. No other case calls `install_transaction`.
- [ ] **Step 2: Confirm RED.** Run `DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test -count=1 ./internal/mod/d2legacy -run TestLuaSuites`; expect the new physical assertion to report 51 instead of 50.
- [ ] **Step 3: Implement the repair.** In `explode_corpse`, replace `local physical = amount - fire` with independently truncated `math.floor(amount * (100 - definition.corpse_fire_percent) / 100)`; keep the fire line unchanged.
- [ ] **Step 4: Confirm GREEN.** Run the same focused Lua suite and require exit 0; inspect the changed test for literal independent expected values and no bypass of the real trap path.
- [ ] **Step 5: Commit.** Commit only the two Lua files with issue #610 in the message.

### Task 2: Verify and submit

**Files:**
- Read: `.github/PULL_REQUEST_TEMPLATE.md`, issue #610, and the complete branch diff.

**Interfaces:**
- Consumes: Task 1 commit and its production-path test.
- Produces: reviewed PR linked to #610, with exact submitted-commit evidence.

- [ ] **Step 1: Run required checks on the final commit.** Run `go test ./...`, `make architecture`, `make test-lua-hardening`, `make test-race`, and `git diff --check origin/main...HEAD`; record any owned-data test status accurately. If race cannot run, record the reason as unverified.
- [ ] **Step 2: Obtain independent code and evidence reviews.** Resolve concrete blockers within #610, rerun every check claimed in the PR description on the new final commit, and inspect the complete effective diff.
- [ ] **Step 3: Submit.** Fill every PR template section before creating the PR; push branch, create PR, attach it to this chat, and let CI run.
- [ ] **Step 4: Integrate.** When required CI checks pass and the PR remains within #610, merge it and record the merge result.
