# Room Population Player Count Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make first room activation pin the same eligible-player count as direct monster spawning for issue #373.

**Architecture:** In `bootstrap/population.lua`, pass active player entities from the room system snapshot to the existing `policy.player_count.live_population`; keep `spawn.components` as the sole monster materializer. Declare the health and death reads required by that policy. Use a Go integration test through authoritative room activation and direct spawn, backed by synthetic records.

**Tech Stack:** Lua `d2legacy` authoritative systems, Go `internal/mod/d2legacy` integration test, existing ECS and session fixtures.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-373-population-player-count-design.md`

## Global Constraints

- Linked issue: #373, partial live-player-context criterion only. Use `Refs #373` in PR.
- The 1.13c MSD-R05 rule is static verified; 1.14d applicability and full unit-condition mapping remain inferred.
- Preserve existing room selection, inactive resident restoration, RNG order, difficulty, loot, and party behavior.
- Do not commit game assets or captures.

## Review Focus

- A committed dead player remains joined: room spawn pins 1 with one living peer, not 2.
- Two living players: room spawn pins 2 with 50% life/XP bonus.
- A living but inactive player: the room count excludes it, matching direct spawn's `none` query.
- `live_population` accesses vitals/death only after the room system declares both reads.
- An explicit override floor and already activated rooms continue to use existing `spawn.components`/room state behavior; inspect the production path and existing tests for these, without expanding this slice.

---

### Task 1: Align room activation with the shared count policy

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua`
- Create: `internal/mod/d2legacy/population_player_count_test.go`

**Interfaces:**
- Consumes: `player_count.live_population(entities)` and existing `spawn.components(command, game_player_count)`.
- Produces: an active-player-filtered count at first room activation; no new exported interface or authority.

- [ ] **Step 1: Write the failing integration test.** Use `startPartyFixture`, `generatedHostileRecords`, `system.population.bootstrap`, and two `system.player.enter` commands, with a populated room that stays unactivated until a player enters it. In subtests, stage Bob's committed death before Alice enters the room, keep both alive, or mark living Bob inactive. Assert the resulting monster's `spawn_player_count` and exact fixture life/XP (`3*256` and `5` for count 1; `4*256` and `7` for count 2). In the dead and inactive cases, submit `system.monster.spawn` with a matching fixed definition and assert the same count and scaled fields. Existing room and direct-spawn tests are examples of the command and store helpers.
- [ ] **Step 2: Run the focused test and verify RED.** `GOCACHE=/private/tmp/dark-magic-issue373-gocache GOMAXPROCS=2 GOFLAGS=-p=2 go test ./internal/mod/d2legacy -run '^TestRoomPopulationUsesLivePlayerCount$' -count=1`. Expected: dead/inactive room spawns pin 2 while direct spawn pins 1.
- [ ] **Step 3: Implement the minimum Lua change.** Add `local player_count = require("d2legacy.policy.player_count")`; replace local raw-identity count with a small helper that filters `entities` by player identity and absence of `d2legacy.world.inactive`, then calls `player_count.live_population`. Add `d2legacy.player.vitals` and `d2legacy.player.death` to `d2legacy.population.active_rooms.read`; keep its broad query and all structural writes intact.
- [ ] **Step 4: Run focused test and verify GREEN.** Repeat Step 2; inspect each subtest result and `git diff --check`.
- [ ] **Step 5: Commit this test and code.** `git add internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua internal/mod/d2legacy/population_player_count_test.go && git commit -m 'fix: share live player count with room population'`.

### Task 2: Verify the submitted boundary and update its issue

**Files:**
- Modify only if correction is necessary: files from Task 1.
- Update: GitHub issue #373 after PR merge.

**Interfaces:**
- Consumes: Task 1's committed change and the issue's still-open difficulty/evidence gates.
- Produces: a PR whose description fills every template section and reports exact checks on its submitted commit.

- [ ] **Step 1: Run relevant verification.** `make test`, `make architecture`, `make test-lua-hardening`, `make test-race`, and `git diff --check`, with `GOCACHE`, `GOMAXPROCS`, and `GOFLAGS` set as above. Report skipped real-asset checks if `MPQ_DIRECTORY` is absent.
- [ ] **Step 2: Request independent reviews.** Ask separate evidence and production-path reviewers to compare the full diff, assertions, and 1.13c/1.14d claims. Correct findings and rerun impacted checks on the final commit.
- [ ] **Step 3: Open, push, and merge.** Rebase on `origin/main`, push branch, open a PR with all template sections and `Refs #373`, attach it to the task, wait for every required CI check to pass, merge, then update issue #373 with delivered scope and remaining gates.
