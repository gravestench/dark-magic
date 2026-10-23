# Demon and Undead Attack Rating Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make player basic melee honor additive demon and undead attack-rating sources from original stats 123 and 124.

**Architecture:** The existing MonStats loader and spawn materializer publish immutable monster type facts. A pure melee policy applies target-specific flat bonuses, while the authoritative melee system resolves pre-percentage player attack rating at impact, adds those bonuses, then uses the existing percentage and selected-hand steps. Derived target-independent combat stats remain unchanged.

**Tech Stack:** Lua authority systems/policy and synthetic test suites, Go Lua test harness and owned-MPQ local test.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-605-demon-undead-attack-rating-design.md`; issue #605; `docs/specs/combat-hit-roll.md` HIT-R02g/HIT-V13/HIT-R03.

## Global Constraints

- Original 1.13c static evidence verifies HIT-R02g; 1.14d runtime applicability is inferred. Owned 1.14d MonStats checks are table observations only.
- Player-only stat 123 adds for `demon`; stat 124 adds once for `lUndead` or `hUndead`; both can add and negative values subtract unchanged.
- Target-specific flat addition precedes existing `item_tohit_percent` resolution; preserve the current selected-hand delta after that percentage.
- Keep the target-independent derived `combat_stats.attack_rating`, the monster-attacker path, targeting, event schema, RNG stream/draw count, and source lifecycle owners.
- The owned-row local test is required before #605 closes. Never commit licensed game data, captures, credentials, or saves.

## Review Focus

- Both undead bits true must add stat 124 once: Task 1 tests it, and Task 2 publishes both flags without collapsing them.
- A missing or invalid monster definition must add no target bonus: Tasks 1 and 4 test supplied and authority inputs.
- A flat +50 bonus with +20% sources must produce AR 420, not 410: Task 4 uses DEF 400 and draw 50 to distinguish hit from miss.
- Hand selection must retain current post-percent behavior: Task 4 tests primary/secondary 360/380 without matching flags and 420/440 with a demon.
- Weapon-set switching or removal must clear the bonus before the next impact: Tasks 3 and 4 test source values and event AR/chance on each transition.

---

### Task 1: Pure HIT-R02g Attack-Rating Policy

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/policy/melee.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/policy/melee_test.lua`

**Interfaces:**
- Consumes: base signed AR; signed stat-123 and stat-124 values; booleans for player attacker, monster defender, demon, low undead, high undead.
- Produces: `melee.target_type_attack_rating(ar, demon_bonus, undead_bonus, player_attacker, monster_defender, demon, low_undead, high_undead) -> ar` for Task 4. No ECS or RNG access.

- [ ] **Step 1: Write failing tests.** Assert AR 300 with bonuses +50/+40 yields 350 for demon, 340 for low undead or high undead, 390 for demon+high undead, and 340 with both undead bits. Assert 300 for neither, missing/invalid monster predicate, non-player attacker, and zero bonuses; assert 210 for demon+high undead with -50/-40. Compare seed state or use a mock to establish no draw.
- [ ] **Step 2: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect missing `target_type_attack_rating`.
- [ ] **Step 3: Implement policy.** Add only eligible nonzero bonuses, demon first and undead once, preserving signed values. Reuse existing policy integer conventions without broad HIT-R03 changes.
- [ ] **Step 4: Verify GREEN.** Run the same fast Lua command; expect exit 0.
- [ ] **Step 5: Commit.** Stage the two task files and commit `feat: calculate demon and undead melee attack rating`.

### Task 2: Publish Authored Monster Type Facts

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/data/monster.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/components/shared.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/commands/spawn_monster.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/data/monster_catalog_test.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/commands/spawn_monster_resistances_test.lua` or a focused new classification suite
- Create: `internal/mod/d2legacy/monster_classification_real_test.go`
- Create: `internal/content/d2legacy/lua/d2legacy/tests/support/monster_classification_real.lua`

**Interfaces:**
- Consumes: MonStats `demon`, `lUndead`, `hUndead` columns in the existing immutable record catalog; shared `runtime_definition` used by `load`, `summon`, and `revive`.
- Produces: definition booleans `demon`, `low_undead`, `high_undead` and one `d2legacy.monster.type_flags` component with those boolean fields, for Task 4. Missing fields decode false.

- [ ] **Step 1: Write failing synthetic loader/spawn tests.** Use one row with demon=1, lUndead=0, hUndead=1, a low-undead-only row with lUndead=1, and another with all zero/blank; assert `monster.load`, `monster.summon`, and `monster.revive` preserve the three booleans under their existing admission gates. Assert `spawn_monster.components` publishes exactly those fields and absence defaults false. No licensed rows in fixtures.
- [ ] **Step 2: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect missing flags/component.
- [ ] **Step 3: Implement loader and component.** Select the three columns in `monster_catalog`; decode in `runtime_definition`. Register `d2legacy.monster.type_flags` with three boolean fields in `components/shared.lua`, and publish it in `spawn_monster.components`. Preserve shared owned-unit construction without new record reads.
- [ ] **Step 4: Verify synthetic GREEN.** Run the same fast Lua command; expect exit 0.
- [ ] **Step 5: Add required owned 1.14d row check.** Following `monster_resistances_real_test.go`, load local MonStats/MonStats2/MonLvl through `DARK_MAGIC_TEST_MPQ_DIRECTORY`; first inspect the owned row flags and choose one demon, one low-undead, one high-undead, plus one neither row. In Go assert exact `demon`, `lUndead`, `hUndead` cells, then call a Lua support helper that runs `monster.load` and `spawn_monster.components` and asserts the same authoritative type component. Keep all asset bytes and captures out of Git.
- [ ] **Step 6: Verify owned GREEN.** Run `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy -run TestOwnedMonsterClassification -count=1 -v`; require executed passing cases, not a skip. If the archive path is unavailable or rows conflict, stop and report issue criterion unverified.
- [ ] **Step 7: Commit.** Stage only task files and commit `feat: publish monster demon and undead type facts`.

### Task 3: Admit and Retire Flat Target-Bonus Sources

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/commands/enter_player.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/equipment.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/player_melee_test.lua`
- Create: `internal/content/d2legacy/lua/d2legacy/systems/target_type_attack_rating_test.lua` (also used by Task 4)

**Interfaces:**
- Consumes: existing passive admission and active equipment stat-source publication/cleanup.
- Produces: resolved additive `item_demon_tohit` and `item_undead_tohit` sources at impact, with the equipment fixture tracking equip/set1/set0/remove and no stat-115/116 source.

- [ ] **Step 1: Write failing passive tests.** Admit `item_demon_tohit=50` and `item_undead_tohit=40` with operation `add`; assert resolved values. Isolated `pcall` cases reject passive `percent` for each stat.
- [ ] **Step 2: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect admission failure.
- [ ] **Step 3: Implement admission.** Add both exact names to allowed passive stats, requiring nil/add operations for each; preserve all other operations.
- [ ] **Step 4: Write failing equipment lifecycle test.** A synthetic weapon carries additive demon +50 and undead +40 only. Equip, set1, set0, remove; assert resolved values `50/0/50/0` and `40/0/40/0` plus source provenance/cleanup.
- [ ] **Step 5: Verify RED, implement publication, verify GREEN.** Run the same fast/integration command before and after adding both names to equipment's active publication and no-layout cleanup lists; expect pre-change source 0 and final exit 0.
- [ ] **Step 6: Commit.** Stage only the four task files and commit `feat: publish target-type melee bonus sources`.

### Task 4: Apply Target Bonuses in Authoritative Basic Melee

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/melee.lua`
- Test: `internal/content/d2legacy/lua/d2legacy/systems/target_type_attack_rating_test.lua`

**Interfaces:**
- Consumes: Task 1 pure `target_type_attack_rating`, Task 2 monster type component, Task 3 stat sources.
- Produces: one adjusted player AR passed to current event/chance/hit paths; no persistent target-specific AR or new draw.

- [ ] **Step 1: Write failing authority cases.** Use separate seed-429 fixtures for each controlled first draw of 50; subsequent draws differ. At equal levels, pre-percent player AR 300, target DEF 400, passive demon bonus +50, and attack-rating percent +20, assert demon AR 420/chance 51/hit and nonmatching AR 360/chance 47/miss; both melee and attack-result records agree. Add demon+undead, undead-only, zero/negative, monster-attacker and missing-component cases. Make target health high enough for cases that deliberately issue repeated events.
- [ ] **Step 2: Write failing hand and lifecycle cases.** With primary pre-percent AR 300, secondary delta +20, and percent +20, assert primary/secondary AR 360/380 against nonmatching and 420/440 against demon. For the Task 3 equipment fixture, assert authoritative AR/chance after every `50/0/50/0` source transition; verify no stale source on switch or removal.
- [ ] **Step 3: Verify RED.** Run `env DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run TestLuaSuites -count=1`; expect target-specific event AR/chance mismatch.
- [ ] **Step 4: Implement impact-time player AR.** Declare `d2legacy.monster.type_flags` in melee reads. For a player attacker, resolve `attack_rating` sources over `combat_stats.base_attack_rating`; apply Task 1 policy using stat 123/124 and target flags; then resolve existing `item_tohit_percent`; finally call existing `hand_attack_rating` so selected-hand delta stays after percent. Keep current non-player `combat_values` path and target-independent derived stat. Apply after #601/#603 defense adjustment and before event/chance/draw.
- [ ] **Step 5: Verify GREEN.** Run the same fast/integration command and `make test-lua`; expect exit 0. Confirm the controlled draw with the new event outcome rather than only a pure probability assertion.
- [ ] **Step 6: Commit.** Stage only melee and focused authority test files; commit `feat: apply demon and undead melee attack rating`.

## Final Verification and Delivery

- [ ] Use fresh read-only subagents for whole-branch code/architecture and evidence/acceptance review. Fix Critical/Important findings and seek re-review approval.
- [ ] Rebase on current `origin/main` before final verification. If main advances again before submission, rebase again and repeat every check claimed in the PR on that submitted commit.
- [ ] On final submitted commit run `go test ./...`, `make architecture`, `make test-lua-hardening`, `make test-lua-format`, `make test-lua-syntax`, `git diff --check origin/main...HEAD`, and required `TestOwnedMonsterClassification` with local owned 1.14d archive. Report exact results and whether local race ran.
- [ ] Fill every PR-template section with evidence/version limits and exact submitted-commit checks, push, open a PR with `Closes #605`, wait for all required CI jobs, merge, and verify #605 closure.
