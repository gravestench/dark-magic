# Monster MonStats Resistances Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make spawned monsters use their six difficulty-specific MonStats resistances for authoritative damage mitigation.

**Architecture:** `data/monster.lua` selects the six columns using immutable game difficulty. `commands/spawn_monster.lua` stores their base and current values in `d2legacy.combat.defense`. `policy/mitigation.lua` uses monster resistance without a player maximum, while retaining player behavior. No client projection gains the monster fields.

**Tech Stack:** d2legacy Lua, Go Lua-harness and owned-MPQ record tests.

**Spec:** Issue #461; `docs/specs/combat-damage-mitigation.md` and `docs/research/spec-evidence/combat-damage-mitigation.md`. The owned 1.14d MonStats table supplies numeric evidence.

## Global Constraints

- Stay within issue #461. Do not implement modifiers, curses, pierce, or poison damage over time.
- Treat 1.14d owned data as verified and 1.13c parity as inferred.
- Never commit game data or captures.

## Review Focus

- A monster resistance of 100 or more makes each damage channel immune; test this at and above 100.
- A monster resistance of 80 remains 80 rather than taking the player's 75 cap; test the resulting damage.
- A negative monster resistance increases damage; test a representative value.
- Players keep their existing 75/95 cap and magic passthrough; test both.
- Normal, Nightmare, and Hell select their own MonStats values even when area level differs; test distinct values.

---

### Task 1: Decode and materialize MonStats resistances

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/data/monster.lua`, `commands/spawn_monster.lua`, and `components/shared.lua`. Add `data/monster_resistances_test.lua` and `commands/spawn_monster_resistances_test.lua`.

**Interfaces:** `monster.load(id, area)` produces six `*_resist` numeric fields. `spawn_monster.components(command, game_player_count)` stores six matching base/current fields on `d2legacy.combat.defense`, including a new authority-only `magic_resist` field.

- [ ] Add fixture tests with distinct six-channel values for difficulties 0, 1, and 2 and assert the spawned component values, not a mock's calls.
- [ ] Run `make test-lua` and confirm failure from missing resistance values/component.
- [ ] Implement the six difficulty lookups and defense component, without adding client projection fields.
- [ ] Run `make test-lua` and confirm green.
- [ ] Commit the task.

### Task 2: Apply monster mitigation

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/policy/mitigation.lua` and `policy/mitigation_test.lua`.

**Interfaces:** `mitigation.apply(raw, channel, defense, monster_target)` returns nonnegative integer damage. Monster physical/magic/elemental channels use their respective resistance; elemental and magic values >=100 are immune, while values below 100 bypass the player maximum. Player branches retain their current rules.

- [ ] Add literal damage vectors for all six channels, 80/100/above-100/negative monster resistance, and player controls.
- [ ] Run `make test-lua` and confirm expected failing vectors.
- [ ] Implement the minimal monster branch in mitigation.
- [ ] Run `make test-lua` and confirm green.
- [ ] Commit the task.

### Task 3: Pin owned 1.14d data and verify the submitted feature

**Files:** Add `internal/mod/d2legacy/monster_resistances_real_test.go` and update only issue-scoped tests if the real table exposes a fixture mistake.

**Interfaces:** The owned-MPQ test loads `data/global/excel/monstats.txt` when `DARK_MAGIC_TEST_MPQ_DIRECTORY` is set and pins representative columns for all three difficulties, including a Hell poison-immune row.

- [ ] Add the test with exact values observed from the owned table; confirm it fails if a selected column is wrong.
- [ ] Run the real-data test against `/Users/dylanknuth/d2_english_mpq` and confirm green.
- [ ] Run `make test`, `make architecture`, `make test-lua-hardening`, `make test-race`, and formatting checks on the final commit; record exact results and skipped data tests.
- [ ] Review the whole diff, commit the task, and prepare the complete PR template.
