# Fractional Target Defense Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make an eligible player's additive stat-116 source reduce defense used by the authoritative basic melee chance and draw.

**Architecture:** A pure function in the existing melee policy computes the HIT-R02d reduction after the stat-115 adjustment. The existing passive and equipment stat-source paths publish the value; the melee system resolves it at impact and sends the resulting defense through the current event, chance, and hit draw. No new authority or component is added.

**Tech Stack:** Lua authority systems and policy, synthetic Lua test suites driven by `internal/mod/d2legacy` Go harness.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-603-fractional-target-defense-design.md`; linked issue #603; source contract `docs/specs/combat-hit-roll.md` HIT-R02d/HIT-V12.

## Global Constraints

- Original 1.13c static evidence supports HIT-R02d; 1.14d equivalence is inferred. Never describe synthetic tests as an owned original-game observation.
- Apply positive stat 116 after Ignore Target's Defense; halve for a player defender, boss, active hireling, or future supplied `monsterDefenseHalf` classification; cap after halving at 100.
- Keep targeting, combat event schema, hit-chance arithmetic, RNG stream and draw count, and existing source lifecycle unchanged.
- Current integration recognizes player, boss and active hireling only; do not guess the unresolved HIT-Q03 producer or add PvP targeting/missiles.
- Do not commit original game data, captures, saves, or credentials. Report exact checks on the submitted commit.

## Review Focus

- A huge positive defense at 1,048,577 must divide before multiplication; Task 1 adds the boundary vector.
- Negative defense multiplication must wrap to signed 32 bits and divide toward zero; Task 1 adds truncation and overflow vectors.
- Halving a percentage before the 100 cap must preserve protected-target behavior; Task 1 adds value 250 and value 1 vectors.
- A stale source after a weapon-set change would silently skew the next attack; Task 2 checks resolved stat and Task 3 checks authoritative chance on each transition.
- A monster attacker or inactive hireling marker must not receive the wrong player/protected treatment; Task 3 tests both.

---

### Task 1: Pure HIT-R02d Defense Arithmetic

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/policy/melee.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/policy/melee_test.lua`

**Interfaces:**
- Consumes: signed prepared defense, signed additive stat-116 value, and a supplied boolean `half`.
- Produces: `melee.fractional_defense(defense, value, half) -> signed adjusted defense` for Task 3. It has no ECS or RNG access.

- [ ] **Step 1: Write failing policy tests.** Assert `(500,50,false)->250`, `(500,50,true)->375`, `(500,250,false)->0`, `(500,250,true)->0`, `(500,1,true)->500`, zero and negative values leave 500, `(1048577,50,false)->524327`, `(-501,50,false)->-251`, `(-2147483648,50,false)->-2147483648`, and `fractional_defense(effective_defense(500,1,true,true,false,false),50,false)->0`. Keep the existing 37/95 hit-chance assertions.
- [ ] **Step 2: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect the new test to fail because `fractional_defense` is missing.
- [ ] **Step 3: Implement `fractional_defense` in `policy/melee.lua`.** For `value<=0` return defense. For `half`, truncate positive `value/2` toward zero, then `p=min(value,100)`. If defense>1048576 use signed low-word `(trunc(defense/100)*p)`; else use `trunc(signed32(defense*p)/100)`; return `signed32(defense-cut)`. Local `signed32` and truncate helpers may follow established `policy/party_gold.lua` conventions. No random calls.
- [ ] **Step 4: Verify GREEN.** Run the same fast Lua command; expect exit 0 and `ok .../internal/mod/d2legacy`.
- [ ] **Step 5: Commit.** Stage only the two task files; commit `feat: calculate fractional target defense for melee`.

### Task 2: Publish Additive Stat-116 Sources

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/commands/enter_player.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/equipment.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/player_melee_test.lua`
- Create: `internal/content/d2legacy/lua/d2legacy/systems/fractional_target_defense_test.lua`

**Interfaces:**
- Consumes: current `passive_stat_sources`, equipment modifier-source publication, and `stat_sources.resolve(entities, player, "item_fractionaltargetac", 0)`.
- Produces: an additive authoritative stat-116 value at impact; Task 3 adds its melee consumer. The new test suite owns a synthetic weapon with stat-116 value 50 and no stat-115 source.

- [ ] **Step 1: Write failing admission tests.** In `player_melee_test.lua`, admit `item_fractionaltargetac=50` with operation `add` and assert resolved value 50. In an isolated `pcall(require("d2legacy.commands.enter_player").apply, {tick=1,payload=entry})` case, assert passive `percent` is rejected. Follow the neighboring stat-115 test pattern so a raised error cannot abort the suite.
- [ ] **Step 2: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect the additive admission test to fail.
- [ ] **Step 3: Implement passive admission.** Add the exact stat name `item_fractionaltargetac` to allowed passive stats in `enter_player.lua`; accept nil/add and reject percent for this stat only. Preserve other passive operations.
- [ ] **Step 4: Verify GREEN.** Run the same fast/integration Lua command; expect exit 0.
- [ ] **Step 5: Write failing equipment lifecycle test.** In `fractional_target_defense_test.lua`, provide a minimal inventory weapon with one additive `item_fractionaltargetac=50` modifier, no Ignore Target's Defense, and no attack-rating bonus. Equip it, switch to set 1, switch back to set 0, then remove it; after each structural update assert resolved stat `50/0/50/0` and source provenance/cleanup. Use admissible command ticks as in `item_modifier_sources_test.lua`.
- [ ] **Step 6: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect the new source assertion to fail before equipment publication.
- [ ] **Step 7: Implement equipment publication.** Add `item_fractionaltargetac` to the active modifier publication and no-layout cleanup lists in `systems/equipment.lua`; preserve item modifier admission behavior.
- [ ] **Step 8: Verify GREEN.** Run the same fast/integration Lua command; expect exit 0.
- [ ] **Step 9: Commit.** Stage only the four task files; commit `feat: publish fractional target defense sources`.

### Task 3: Apply the Reduction in Authoritative Melee

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/melee.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/fractional_target_defense_test.lua`

**Interfaces:**
- Consumes: Task 1 `policy.fractional_defense(defense,value,half)` and Task 2 resolved stat-116 source.
- Produces: one adjusted defense recorded in `melee_event` and `attack_result`, passed to `policy.hit_chance` and `policy.hits`.

- [ ] **Step 1: Write failing authority tests.** In the new suite use seed 36 and separate cases with player level 20, AR 300, monster level 20, DEF 500, and first hit draw 51. An ordinary target with passive stat 50 must record DEF 250/chance 54/hit; boss and active hireling must record DEF 375/chance 44/miss; zero and negative stat must record DEF 500/chance 37/miss. A monster attacker bearing stat 116 must retain player DEF 500/chance 37/miss. Add an inactive hireling-marker target that uses the ordinary DEF 250/chance 54/hit path. Assert `melee_event` and `attack_result` agree. Keep player-target halving in Task 1's supplied policy inputs.
- [ ] **Step 2: Extend the dedicated equipment fixture's lifecycle case.** After each `50/0/50/0` source assertion, issue a direct basic melee request against a level-one DEF 500 target. With no weapon attack-rating bonus the player's base AR is 65, so exact defense/chance should be `250/20`, `500/11`, `250/20`, `500/11`. Assert both event representations agree after each attack. Use the #601 lifecycle timing pattern and leave the seed-36 suite's separate ordinary/boss cases to prove distinct hit outcomes.
- [ ] **Step 3: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect ordinary DEF 500 rather than 250 and/or lifecycle chance 11 rather than 20.
- [ ] **Step 4: Implement the melee consumer.** In `systems/melee.lua`, only for a player attacker resolve `item_fractionaltargetac` at impact, compute `half` from player identity on target, existing boss marker, or active owned-unit category `hireable`, then apply Task 1's policy after `effective_defense` and before event/chance/draw. The unresolved `monsterDefenseHalf` producer stays absent. Keep the existing ECS read set and RNG calls.
- [ ] **Step 5: Verify GREEN.** Run the same fast/integration Lua command; expect exit 0. Run `make test-lua` and expect exit 0.
- [ ] **Step 6: Commit.** Stage only `systems/melee.lua` and the focused test file; commit `feat: apply fractional target defense to melee hits`.

## Final Verification and Delivery

- [ ] Use fresh read-only subagents to review the whole branch against issue #603, source evidence, architecture, test strength, determinism, security, and performance. Fix Critical/Important findings and seek re-review approval.
- [ ] On the final commit run `go test ./...`, `make architecture`, `make test-lua-hardening`, `make test-lua-format`, `make test-lua-syntax`, and `git diff --check origin/main...HEAD`; report each exact result plus real-asset/race local status.
- [ ] Rebase on current `origin/main`, rerun affected verification if the submitted commit changes, fill every section of `.github/PULL_REQUEST_TEMPLATE.md`, push, open a PR with `Closes #603`, wait for all required CI jobs, then merge and verify closure.
