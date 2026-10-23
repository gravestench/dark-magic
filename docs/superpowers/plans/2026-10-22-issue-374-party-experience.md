# Issue #374 Party Experience Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Award monster-kill experience to eligible nearby party members through the authoritative death path.

**Architecture:** `d2legacy.policy.party_experience` reads the existing party snapshot and ECS entities, freezes eligible recipients, and calculates provisional raw shares. `monster_death.lua` keeps kill ownership and progression writes; it calls that policy and awards each share through `experience_gain`. NoDrop keeps its separate same-level count.

**Tech Stack:** Lua 5.1 policy and ECS, Go session integration tests, Superpowers TDD.

**Spec:** `docs/superpowers/specs/2026-10-22-issue-374-party-experience-design.md`

## Global Constraints

- Issue #374: this PR delivers monster-kill party XP only, uses `Refs #374`, and leaves the issue open.
- Keep party membership in `policy/party.lua`, death and owner credit in `systems/monster_death.lua`, and scaling/capping in `policy/experience_gain.lua`.
- Mark coordinate distance and share rounding as provisional 1.14d inferences from 1.13c static evidence.
- Never change NoDrop's distance-independent same-level count or commit game assets.

## Review Focus

- Missing victim position while party exists: retain the credited player's existing solo award (Task 2 test).
- Credited player has no identity or progress: no award and no error (Task 2 test).
- A roster member lacks progress, location, or position: omit that member before summing levels (Task 1 test).
- Death marker exists while health is still positive: omit that member (Task 1 test).
- Replayed monster death event: no second XP write (Task 2 test).

---

### Task 1: Eligible party recipients and provisional raw shares

**Files:**
- Create: `internal/content/d2legacy/lua/d2legacy/policy/party_experience.lua`
- Create: `internal/content/d2legacy/lua/d2legacy/policy/party_experience_test.lua`

**Interfaces:**
- Consumes: `party.snapshot()` with `membership[player_id]` and ordered `parties[party_id].members`; `ecs.get(entity, component)` from the death system's entity snapshot.
- Produces: `party_experience.shares(entities, credited_player, victim_position, raw_experience)`. Returns `nil` for an unpartied player or missing victim position, credited-player position, or credited-player level location (caller uses solo), otherwise an ordered array of `{ entity = entity, raw = integer }`. An empty array means no party award. For one admitted member, return `{ entity = credited_player, raw = raw_experience }`, even when that admitted member differs from the credited player.

- [ ] **Step 1: Write failing policy cases.** Use real ECS fixture entities and party authority state. Assert PXL-V09's three level-10 recipients at victim (100,100), XP 1000 each receive raw 565; two members with second at (180,100) receive raw 673 each; at (181,100) only credited receives raw 1000. Assert ordered recipients and one-member credited-player route, including credited player outside the radius while another member at the victim is sole admitted member; only credited receives raw 1000. For the explicitly provisional Lua arithmetic, levels 71/50/20 and X=4,049,741 give pool 6,865,576 and a level-71 raw share of 3,457,134; a stored float32 quotient would instead give 3,457,133. Label this as a discriminator, not original-game parity.
- [ ] **Step 2: Run focused Lua suite to observe the missing-module failure.** `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/party_experience' -count=1`.
- [ ] **Step 3: Implement `shares` with a snapshot of membership and entities.** Filter identity, progress, health > 0, absent death marker, same level, position X != 0, squared victim distance <= 6400. Freeze levels before computing `X + trunc((n-1)*X*89/256)` and level-proportional truncated shares. Reject more than eight admitted members; nonpositive level sum returns no awards. Keep arithmetic in this file.
- [ ] **Step 4: Write edge-case tests, then run the focused suite and observe each intended failure.** Assert far, other-level, member X=0 and victim X=0 separately, absent member position/location/progress, health zero, positive health plus death marker, no admitted members, no party, missing victim or credited-player position/location, zero sum, and changed membership. Assert the ninth admitted member fails before any award. Add one fractional-position case that pins the current f64-world-coordinate inference while explicitly noting original dynamic getter u16 integer conversion may differ.
- [ ] **Step 5: Complete only the missing policy guards.** Run the focused suite again and confirm all cases pass.
- [ ] **Step 6: Commit the tested policy and cases.**

### Task 2: Route monster death XP through the party policy

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/monster_death.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/policy/experience_gain.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/policy/experience_gain_test.lua`
- Create: `internal/mod/d2legacy/party_experience_test.go`

**Interfaces:**
- Consumes: `party_experience.shares` from Task 1, `experience_gain.gain` and `capped_total`, and the existing credited owner from `attribution.resolve`.
- Produces: one progress write per eligible recipient at death, without changing death fact `xp`, loot counts, or non-XP credit events.

- [ ] **Step 1: Write failing production-session tests.** Arrange nearby and far players in a party and kill one monster; assert each player's stored XP, unchanged NoDrop same-level count, and death fact's original monster XP. Repeat with a member dead, on another level, leaving or joining; assert owned-unit credit uses the same recipient path. Restore a checkpoint, change membership before the kill, and assert the restored session awards only current recipients without a duplicate death award. Assert solo kill, absent credited identity/progress, missing victim position, and repeated death processing retain current behavior.
- [ ] **Step 2: Run focused Go test and observe the XP assertion fail.** `go test ./internal/mod/d2legacy -run 'TestPartyExperience' -count=1`.
- [ ] **Step 3: Integrate the policy in `credit_experience`.** Keep the current solo award helper; for a party share list, call `experience_gain.gain` and `capped_total` for each frozen recipient, resolving that recipient's class and item XP bonus. Add `d2legacy.world.position` to the system read set. Do not change loot's `party.additional_living_members_in_same_level` call.
- [ ] **Step 4: Add the PXL-R05a nonpositive raw-share case to `experience_gain_test.lua` and watch it fail.** Then change `experience_gain.gain` to return one before level-cap logic when party supplies a nonpositive share; retain the existing zero monster-XP guard in the death path. Run the focused Lua suite and session test until both pass.
- [ ] **Step 5: Commit the integrated behavior and tests.**

### Task 3: Verify the delivered boundary

**Files:**
- Review: all changed files and `docs/specs/player-party-experience-sharing.md`
- Update: issue #374 and PR description after merge, with exact results and unresolved evidence.

- [ ] **Step 1: Run full verification on the candidate commit.** `make test`, `make architecture`, `make test-race`, and `make test-lua-hardening`; record exact results and asset-dependent skips.
- [ ] **Step 2: Request a separate review of the full diff.** Check actual XP assertions and production route, owned-unit attribution, NoDrop independence, deterministic order/replay, numerical inference labels, and issue scope. Fix findings with a fresh test cycle.
- [ ] **Step 3: Rebase on current main, verify the submitted commit, write every PR-template section, push, and open PR with `Refs #374`.** Include what remains unverified in 1.14d and complete review evidence.
- [ ] **Step 4: Merge only after CI passes; update #374 with the delivered XP slice and leave remaining reward work open.**
