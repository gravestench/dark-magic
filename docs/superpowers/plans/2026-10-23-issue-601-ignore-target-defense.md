# Issue #601 Ignore Target's Defense Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make an eligible player's Ignore Target's Defense source zero ordinary monster defense for the authoritative basic melee chance and draw.

**Architecture:** A pure melee policy function decides eligibility from supplied classifications. Existing passive and equipment stat-source paths publish stat 115, and the authoritative melee system resolves it at impact and passes one adjusted defense to chance, draw, and event recording.

**Tech Stack:** First-party d2legacy Lua, ECS v1 capabilities, Go-hosted Lua test harness.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-601-ignore-target-defense-design.md`

## Global Constraints

- Issue #601 only. HIT-R02c is verified in original 1.13c static code; HIT-V10 is synthetic; 1.14d equivalence remains inferred.
- Preserve targeting, RNG stream and draw count, equipment source architecture, combat event schema, and current hit chance arithmetic.
- Do not implement the unresolved HIT-Q03 exclusion producer, stat 116, missiles, or PvP admission.
- Never commit original game data, captures, credentials, or private saves.

## Review Focus

- A negative nonzero stat value must enable the gate; Task 1 tests it.
- A boss or active hireling with the stat present must retain defense; Tasks 1 and 3 test this.
- A player defender or monster attacker must retain defense even if stat 115 is supplied; Task 1 tests both without changing targeting.
- An item switched away or removed must stop affecting the next resolved melee attack; Task 2 tests source cleanup and Task 3 tests the attack result after cleanup.
- Chance, actual hit draw, and event defense must all use the same adjusted value; Task 3 tests a controlled draw that separates 37% from 95%.

---

### Task 1: Pure HIT-R02c Eligibility

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/policy/melee.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/policy/melee_test.lua`

**Interfaces:**
- Produce: `melee.effective_defense(defense, ignore_target_ac, attacker_is_player, defender_is_monster, boss, hireling) -> number` for Task 3.

- [ ] Add a fast policy case with prepared defense 500 and stat values 0, 1, and -1. Assert only nonzero stat for player versus ordinary monster returns 0. With stat 1, separately assert boss, hireling, player defender, and monster attacker return 500; `hit_chance(20,20,300,0)` is 95 while `hit_chance(20,20,300,500)` is 37.
- [ ] Run `DARK_MAGIC_LUA_TEST_TIERS=fast go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect the new case to fail because the function is absent.
- [ ] Implement the pure function with no RNG or ECS access and no defense mutation for ineligible inputs.
- [ ] Rerun the same command; expect pass. Commit policy and test with `feat: gate Ignore Target's Defense by attacker and target`.

### Task 2: Publish and Clear Stat 115 Sources

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/commands/enter_player.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/equipment.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/item_modifier_sources_test.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/player_melee_test.lua`

**Interfaces:**
- Produce: resolved `item_ignoretargetac` from `stat_sources.resolve(entities, player, "item_ignoretargetac", 0)` for Task 3.

- [ ] Add a passive-entry test admitting integer `item_ignoretargetac` with operation `add`. In a separate isolated case, reject `percent` via `pcall(require("d2legacy.commands.enter_player").apply, { tick = 1, payload = entry })`, following `enter_player_difficulty_test.lua` so the expected error cannot abort the suite. Add an equipment lifecycle test that equips a synthetic item modifier for the stat, then switches away and removes it, asserting the resolved player stat is 1, then 0, with no duplicate or stale source after structural updates.
- [ ] Run `DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect the new cases to fail at passive admission or source resolution.
- [ ] Admit stat 115 in `create_passive_sources` with add-only operation. Add it to equipment's active modifier publication list and no-layout cleanup list, following adjacent named stats without a new source store.
- [ ] Rerun the same integration-inclusive command; expect pass. Commit the source lifecycle and tests with `feat: publish Ignore Target's Defense stat sources`.

### Task 3: Authoritative Melee Consumption

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/melee.lua`
- Create: `internal/content/d2legacy/lua/d2legacy/systems/ignore_target_defense_test.lua`

**Interfaces:**
- Consume: Task 1 `melee.effective_defense` and Task 2 stat source.

- [ ] Create a focused authority integration suite with `seed = 36` (its first basic-melee hit draw is 51) and a direct basic melee request, so the existing `player_melee_test.lua` seed and expectations stay intact. Set player entry `level = 20` and `base_attack_rating = 300 - 5 * (dexterity - 7) - CharStats.ToHitFactor`; set monster level 20 and defense 500. Assert event attack rating 300 and defender level fixture 20 before relying on the 37%/95% expectations. An ordinary hostile target with passive stat 1 records defense 0/chance 95 and hits; zero stat records defense 500/chance 37 and misses. Add a boss case and a synthetic hostile target carrying active `owned_unit.category = "hireable"`; both retain defense 500/chance 37. The latter tests supplied classification without asserting that a player can target a friendly hireling. Add a monster-attacker case with an explicit stat-115 source on the attacker; it still retains target defense. Assert `attack_result` and `melee_event` agree. In the existing item modifier fixture, assert post-swap and post-removal attacks return to defense 500/chance 11 at base level-one attack rating 65, while equipped attacks use defense 0/chance 95. The controlled seed-36 cases separately prove the discriminating hit draw.
- [ ] Run `DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect the new production cases to fail before the integration.
- [ ] In `systems/melee.lua`, declare `d2legacy.monster.boss` in the ECS read set. Resolve stat 115 for player attackers, classify existing boss and active `owned_unit` category `hireable` targets, call the pure policy function, and use its one result for event defense, `hit_chance`, and `hits`.
- [ ] Rerun the same integration-inclusive command, then `make test-lua`, `go test ./...`, `make architecture`, `make test-lua-format`, and `git diff --check`; expect all to pass. Commit with `feat: apply Ignore Target's Defense to melee hits`.

## Final review and delivery

- [ ] Review the final diff against issue #601, the design, HIT-R02c/HIT-V10, source architecture, security, determinism, and test call paths with at least two independent agents; resolve any findings and rerun affected checks on the final commit.
- [ ] Fill every pull request template section with direct file links, versioned evidence, exact final-commit checks, local real-asset/race status, AI use, and `Closes #601` / `Refs #348`. Rebase onto current main, push, open and attach the PR, wait for all CI checks, then merge and verify issue closure.
