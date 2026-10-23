# NoDrop Outcome Session Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove that issue #373's eligible party count changes the stored monster loot outcome in an authoritative session.

**Architecture:** A new test-local fixture starts two independent sessions with the same seed and synthetic monster/treasure rows. One session joins the players into a party before the kill; the other preserves the same ticks without joining. The test compares both death-time NoDrop inputs and parsed durable drops.

**Tech Stack:** Go integration tests, Akara dynamic component stores, D2Legacy Lua authority, synthetic Excel records.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-373-nodrop-outcome-design.md`

## Global Constraints

- Stay within issue #373's NoDrop outcome path; edit product code only if the test exposes a scoped defect.
- Use seed `314`, treasure class `fallen-drop` with `Picks=1`, `NoDrop=3`, `Item1=rin`, `Prob1=1`.
- Keep two living players and one monster, with identical spawn count and kill timing across sessions.
- Assert stored `drops` as parsed JSON: empty array without a party, one `rin` code with a party. Do not assert quality or full JSON encoding.
- Do not commit game data, captures, credentials, or private saves.

## Review Focus

- Player population differs at spawn: the test must assert `monster_player_count=2` in both runs.
- A different player gets death credit: the test must assert `credited_id=player:alice` in both runs.
- Different tick timing advances the loot stream: the test must drive the same number of ticks and kill attempts in both runs.
- Death may be recorded more than once: the test must assert exactly one durable death in each run.
- Party membership might not affect eligibility: the test must assert `no_drop_player_count` is 1 versus 2 before comparing drops.

---

### Task 1: Pin Party-Dependent Durable Loot Outcome

**Files:**
- Create: `internal/mod/d2legacy/party_nodrop_outcome_test.go`
- Conditional modify only if a scoped failure proves necessary: `internal/content/d2legacy/lua/d2legacy/systems/monster_death.lua` or `internal/content/d2legacy/lua/d2legacy/loot/treasure_class.lua`

**Interfaces:**
- Consumes: `generatedHostileRecords() fixtureRecords`, `StartWithConfig(...)`, `generatedPlayerPayload(...)`, `submitPartyCommand(...)`, `stepPartySession(...)`, `stepSession(...)`, `akara.GetDynamicStore(...)`.
- Produces: `TestPartyEligibilityChangesDurableMonsterLootOutcome(t *testing.T)` and test-local fixture/result helpers only; no new product API.

- [ ] **Step 1: Write the integration test.** Build two fresh sessions from `generatedHostileRecords()`, overriding only the `treasureclassex.txt` row. Bootstrap the same room and two players at tick 1. At ticks 2 and 3, join Alice and Bob only in the party run; advance identical ticks in the other run. Kill the single monster through Alice's skill at matching ticks. For each run, assert one death, Alice's credit, spawn count 2, game count 2, `no_drop_player_count` 1 or 2, and parsed `drops` as `[]` or one object with `code: rin`.
- [ ] **Step 2: Verify test sensitivity.** Run `go test ./internal/mod/d2legacy -run '^TestPartyEligibilityChangesDurableMonsterLootOutcome$' -count=1`. If current product code passes, temporarily pass `no_drop_player_count=1` into `loot.roll` in `monster_death.lua` while preserving the unmodified context used for durable death fields. Rerun and confirm the party case fails specifically at parsed `drops` (while its stored count remains 2); then revert the temporary mutation and rerun green. Never commit the sensitivity mutation.
- [ ] **Step 3: Fix a scoped defect only if the untouched production path fails.** Use the failing assertion and MSD-R21–R23 in `docs/specs/monster-scaling-by-players-and-difficulty.md` to make the smallest correction to the death-time count, treasure weight, or stored drop result. Do not change original seed/draw-order policy.
- [ ] **Step 4: Review and verify before commit.** Ask independent agents to review issue boundary and behavioral evidence, and address actionable feedback. On the resulting diff, run the focused test, `make test`, `make architecture`, `make test-lua-hardening`, and `make test-race`. Record exact results and whether `MPQ_DIRECTORY` was set; repeat affected checks after any later edit.
- [ ] **Step 5: Commit and reverify.** Commit the approved test and any scoped correction. Rebase on current `main` if needed, then rerun the applicable checks on that final submitted commit.
- [ ] **Step 6: Submit the PR.** Before `gh pr create`, write the full `.github/PULL_REQUEST_TEMPLATE.md` description with the issue boundary, evidence, exact final-commit check results, unverified behavior, and AI use/review. Push, create and attach the PR, then merge only after required CI passes.
