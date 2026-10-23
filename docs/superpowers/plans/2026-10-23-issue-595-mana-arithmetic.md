# Issue #595 Player Mana Arithmetic Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the existing authoritative mana tick match PMR-R04/R05/R07a arithmetic for signed-32 resource inputs admitted as nonnegative current and maximum mana.

**Architecture:** Keep the single mana writer in `systems/mana.lua`. Add private signed-32 and `mulDiv(factor,base,100)` arithmetic there, consuming the same ECS facts and leaving the state/admission API unchanged. The authority-profile Lua suite checks the full entry → stat source → tick → vitals path.

**Tech Stack:** d2legacy Lua, ECS authority profile, Go test runner.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-595-mana-arithmetic-design.md`; game rules in PMR-R04/R05/R07a and NRF-R05.

## Global Constraints

- Linked issue #595 owns this slice; do not change event scheduling, potion cleanup, class admission, or other resource systems.
- Verified source is original 1.13c static code. Applying it to the 1.14d target remains inferred.
- No original game assets, captures or private saves may be committed.
- Only signed-32 raw resource/stat inputs are within the recovered arithmetic claim. Player-entry validation permits wider i64 values; do not claim original parity for them.
- For this PMR call, `c=100`, `base<=floor(INT_MAX/25)`, and the NRF-R05 branch order keep reachable multiplication intermediates below 2^53, so Lua's numeric multiplication is exact in this domain.

## Review Focus

- Factor 65,537 with base 10,000 must divide the factor first and request 6,550,000 raw units (Task 1).
- A flat INT_MAX contribution must wrap the regen sum before the valid-state lower clamp (Task 2).
- Factor addition beyond INT_MAX must wrap before multiplication (Task 2).
- Negative percent must truncate toward zero, while suppression leaves flat recovery active (Task 2 regressions).
- Authored ManaRegen zero must keep the existing 7,500-tick fallback (Task 2 regression).

---

### Task 1: PMR divide-first percent arithmetic

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/systems/mana_test.lua` and `internal/content/d2legacy/lua/d2legacy/systems/mana.lua`.

**Interfaces:** Consume existing `d2legacy.player.resource_stats` and `d2legacy.player.vitals`. Produce the same vitals fields; no new public API.

- [ ] Write an authority-profile `test.case` that enters an Amazon, sets `max_mana_raw=30000000` and `mana_raw=0`, supplies `manarecoverybonus=65437`, advances one tick, and observes `mana_raw=6550000` (PMR-V14(a)).
- [ ] Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/systems/mana' -count=1`; confirm the new assertion fails with 6,553,700 or another specifically diagnosed current result.
- [ ] Add local signed-32 conversion and a `mulDiv(factor,base,100)` implementation that tests NRF-R05's `b` branch before `a`, uses signed truncation, and wraps products/returns as specified. Replace the existing `integer_percentage(base,...)` call with `mulDiv(i32(100+bonus),base,100)`.
- [ ] Run the focused command again; confirm the new assertion and existing mana cases pass.
- [ ] Commit this tested increment with a message referencing #595.

### Task 2: Flat wrap, ordered valid-state clamp and regressions

**Files:** Modify the same two files only.

**Interfaces:** Same vitals inputs and outputs; no new public API.

- [ ] Add production-tick cases for PMR-V14(b) with `current=100,max=25600,flat=INT_MAX` (result 0) and PMR-V14(c) with `current=0,max=25600,bonus=INT_MAX` (result 7). Add separate valid-state upper cap (`current=25599,max=25600,base=8` produces 25600) and lower drain (`current=1,max=25600,percent=-130` produces 0) cases, plus suppression/flat and authored `ManaRegen=0` regressions through the authority path. Valid nonnegative resource inputs cannot distinguish upper/lower clamp order; that order comes from PMR-R07a static evidence.
- [ ] Run the focused command before editing production code. PMR-V14(b) should fail from unwrapped flat addition. PMR-V14(c) should already pass after Task 1 wraps the factor; upper/lower and other regression cases may also pass immediately. Record each result rather than treating an already-green regression as a new red test.
- [ ] Apply signed-32 wrap after adding flat recovery. Compute `room=i32(max-current)`, then cap regen above room, then cap below `i32(-current)`, and add the resulting nonzero delta once. Preserve the current display projection and suppression ownership. Keep behavior beyond signed-32 resource inputs unclaimed.
- [ ] Run the focused suite and inspect the complete diff. Commit the tested increment; record the submitted commit SHA.
- [ ] On that submitted commit, run `make test-lua`, `go test ./...`, `make architecture`, and `git diff --check origin/main...HEAD`. Record exact results, skips, failures, and checks not run (including race and real-asset tests) in the PR.

### Delivery and review

- [ ] Have fresh independent subagents review evidence, architecture, test quality and full diff. Fix valid findings and rerun affected checks on the final submitted commit.
- [ ] Fill every PR-template section with issue #595, exact checks, 1.13c static evidence and 1.14d observation limit; push, attach PR, wait for CI, then merge if green.
