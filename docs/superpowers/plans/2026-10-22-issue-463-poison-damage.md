# Poison Damage Over Time Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a landed, duration-bearing poison hit drain a player or monster for exactly its active frames, with replacement, Cleansing, death credit, and green presentation.

**Architecture:** Extend the existing typed damage event with poison length, then let one effects-phase observer request the existing timed poison state. A pre-simulation consumer spends its mitigated raw rate; the existing death owners handle lethal events. Local and connected presentation derive tint from the active state ID, without receiving private damage or attribution data.

**Tech Stack:** d2legacy Lua/ECS, Go world-view adapter and Lua harness, owned 1.14d MPQ record tests.

**Spec:** [Issue #463 design](../specs/2026-10-22-issue-463-poison-damage-design.md); [issue #463](https://github.com/gravestench/dark-magic/issues/463).

## Global Constraints

- Stay inside issue #463: shared poison rate/duration behavior, basic-melee stat entry, Cleansing, death credit, and tint.
- Do not implement individual poison skills, poison-length reduction items, monster modifiers, curses, or pierce.
- Treat 1.14d owned table identities as verified; label first/last tick, player rounding, town suppression, RGB tint, and 1.13c parity as interpretations pending original-runtime evidence.
- Preserve poison outside the immediate-health total and apply player/monster poison resistance once before the timed state.
- Never commit Diablo II game data, captures, credentials, or private saves.

## Review Focus

- A poison-only hit with `damage_raw=0` still creates a state; Task 2 tests the event-to-state path.
- A second hit from the same attacker resets `applied_tick` and expiry; Task 2 tests unique application identity.
- Two poison hits in one frame select the stronger rate deterministically, and equal rate can replace; Task 2 tests ranked requests.
- A player takes subpoint damage each frame without losing or duplicating raw units; Task 3 tests `health*256-poison_remainder_raw`.
- Connected poison does not enter aura rotation or expose attacker/rate; Task 4 tests the projection and Task 5 tests tint.

---

## File ownership map

- `policy/damage_bundle.lua`, `components/skill_actions.lua`, and `systems/melee.lua`: typed rate/length production; no state lifecycle.
- `systems/poison_application.lua`, `components/shared.lua`, and `systems/timed_state.lua`: event observation, ranked state request, and explicit attacker attribution.
- `systems/poison_tick.lua`, `systems/aura_pulse.lua`, and `authoritative.lua`: frame order, life debit, and ordinary lethal combat event.
- `adapter/player/world_view.go`, `adapter/player/client_view.go`, `app/clientapp/remote_events.go`, and presentation schema: bounded connected state projection.
- `gameplay/state_overlay_presentation.lua` and `screens/game_world.lua`: record-driven green tint and neutral reset for rendered actors.
- Each owner gets tests beside its implementation; the real-data check lives in `internal/mod/d2legacy` and reads the owned MPQ only when configured.

### Task 1: Carry poison rate and duration from basic melee

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/policy/damage_bundle.lua`, `components/skill_actions.lua`, and `systems/melee.lua`. Test in `policy/damage_bundle_test.lua` and `systems/player_melee_test.lua`.

**Interfaces:** `damage_bundle.stage_component(rolled, mitigated, poison_length_frames)` adds integer `poison_length_frames` (default 0) beside the six rate channels. Melee reads `poisonmindam`, `poisonmaxdam`, and `poisonlength` through `stat_sources.resolve`, rolls once when maximum and length are positive, and passes length to the staged component. `damage.resolve` keeps poison out of immediate life.

- [ ] Add failing tests: a melee hit with equal 128-raw poison endpoints and length 3 stages rate 128 and length 3; no poison stats stage zeros; a fire 256/poison 128 bundle removes only 256 raw immediate life. Assert the `d2legacy.combat.damage_bundle` fields from a production melee event.
- [ ] Run `make test-lua`; confirm the new assertions fail on the missing duration/poison producer.
- [ ] Add the stage field and minimal melee producer, with nonnegative range/length validation and no new skill formula.
- [ ] Run `make test-lua`; confirm the new tests pass.
- [ ] Commit the task, including its tests.

### Task 2: Admit and rank poison as a timed state

**Files:** Create `internal/content/d2legacy/lua/d2legacy/systems/poison_application.lua` and `systems/poison_application_test.lua`. Modify `components/shared.lua`, `components/skill_actions.lua`, `systems/timed_state.lua`, and `authoritative.lua`.

**Interfaces:** `poison_application.register()` observes a co-composed combat event and damage bundle once, resolving its public target ID. It requests `state_id="poison"`, `stat="hpregen"`, `stat_value=-poison_mitigated_raw`, `exclusive_group="poison"`, `replacement_priority=poison_mitigated_raw`, `reject_lower_priority=true`, and an application-unique source ID. Add `source_attacker_id` to state request/instance and copy it in `timed_state`; add an empty `d2legacy.combat.poison_observed` marker to prevent replay. `timed_state` runs after the observer.

- [ ] Add failing system tests: a poison-only event with mitigated rate 100 and length 5 makes `hpregen=-100` and expiry `t+5` despite immediate damage 0; immune/zero-length/dead targets make no state. Apply rate 90/length 10 next and assert 100/old expiry remains; apply equal rate 100/length 2 and assert new expiry `t+2` and application identity. Repeat from the same attacker, submit rates 90 and 120 in one frame, and advance one frame; assert reset, winner 120, and no replay. Assert actual instance and attribution, not only request counts.
- [ ] Run `make test-lua`; confirm failures on the missing observer/state fields.
- [ ] Add the observer, schemas, registration, and ordering. Keep the public event ID separate from the attacker ID; do not infer it from the source string.
- [ ] Run `make test-lua`; confirm all Task 2 cases pass.
- [ ] Commit the task, including its tests.

### Task 3: Spend raw poison each frame and credit deaths

**Files:** Create `internal/content/d2legacy/lua/d2legacy/systems/poison_tick.lua` and `systems/poison_tick_test.lua`. Modify `components/shared.lua`, `systems/aura_pulse.lua`, and `authoritative.lua`; extend `systems/aura_skill_test.lua` for Cleansing.

**Interfaces:** `poison_tick.register()` runs in `pre_simulation` before `d2legacy.skill.aura_periodic_effect`. It debits active `poison` instances only for `applied_tick < context.tick <= expires_tick`. Monster health is raw. Player vitals add `poison_remainder_raw` in `[0,255]`; effective raw life is `health*256-poison_remainder_raw`. A lethal tick emits the ordinary `d2legacy.combat.event` with `kind="unit_died"`, `source_kind="poison"`, and stored `source_attacker_id`. Town uses existing `movement_rules.is_town(level_id)` and suppresses damage while expiry advances.

- [ ] Add failing frame tests: a player starting at 2 life with rate 128/length 3 has `(health,remainder)` `(2,128)`, `(1,0)`, `(1,128)` on ticks `t+1..t+3`; rate 1 accumulates across frames; a monster starting at 512 raw loses 128 per frame. Town suppresses loss; dead/inactive targets do not tick; rate 256 against 1 life makes exactly one lethal event. Assert one player and one monster death reaches the existing death owner with the original attacker ID.
- [ ] Extend the Cleansing test to apply a real poison state, pulse it, and assert shortened expiry and fewer future poison debits.
- [ ] Run `make test-lua`; confirm the new cases fail without the consumer.
- [ ] Implement the consumer, bounded player remainder, lethal event, and explicit system ordering. Do not alter other whole-life damage or healing formulas.
- [ ] Run `make test-lua`; confirm Task 3 cases pass.
- [ ] Commit the task, including its tests.

### Task 4: Project the active poison state to connected clients

**Files:** Modify `internal/mod/d2legacy/adapter/player/world_view.go`, `adapter/player/client_view.go`, `app/clientapp/remote_events.go`, and `internal/content/d2legacy/lua/d2legacy/components/skill_actions.lua`, `gameplay/world.lua`. Test in `adapter/player/world_view_test.go`, `adapter/player/client_view_test.go`, `app/clientapp/remote_events_test.go`, and `gameplay/world_test.lua`.

**Interfaces:** `WorldState` gains `Aura bool` (`json:"aura"`). Aura rows keep positive `PeriodTicks`; a timed poison row uses `Aura=false`, `PeriodTicks=0`. `projectWorldStates` admits only active poison instances attached to visible public targets, deduplicates by target/state ID, and sends no rate, source, attacker, or expiry. The `d2legacy.presentation.state` mirror gains `aura:bool`; `gameplay.world.state_snapshots()` reads it so non-aura poison never joins `aura_overlay_cycle`.

- [ ] Add failing Go/Lua tests for visible poison inclusion, hidden/truncated target omission, duplicate/malformed state rejection, no private fields in JSON, client validation, mirror removal when poison expires, and preserved aura rotation semantics.
- [ ] Run `go test ./internal/mod/d2legacy/adapter/player ./internal/app/clientapp` and `make test-lua`; confirm failures on the missing timed projection/aura distinction.
- [ ] Add the narrow projection, validation, mirror/schema field, and snapshot distinction. Keep unrelated timed states off this issue's connected projection.
- [ ] Run `go test ./internal/mod/d2legacy/adapter/player ./internal/app/clientapp` and `make test-lua`; confirm Task 4 cases pass.
- [ ] Commit the task, including its tests.

### Task 5: Render and clear poison green

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/gameplay/state_overlay_presentation.lua`, `screens/game_world.lua`, and their neighboring Lua tests. Add a focused tint-selection test if `game_world_test.lua` cannot observe retained nodes directly.

**Interfaces:** `state_overlay_presentation.from_rows(...)` exposes the `green` flag from the owned `States.txt` row. `game_world` selects a presentation RGB tint of `(128,255,128)` for any visible actor target with that active state, applies it to `self.hero`, retained remote-player nodes, and retained monster nodes, and restores `(255,255,255)` when the state disappears. The existing cast overlay remains record-driven; tint does not mutate gameplay state.

- [ ] Add failing tests for green flag decoding, hero/remote-player/monster tint, removal reset, and actor/world reuse without stale tint. Include another active aura alongside poison to ensure the tint remains visible.
- [ ] Run `make test-lua`; confirm the new cases fail on missing tint behavior.
- [ ] Implement record-driven tint selection and apply/reset through existing `node:set_tint` calls, without adding a second state store.
- [ ] Run `make test-lua`; confirm Task 5 cases pass.
- [ ] Commit the task, including its tests.

### Task 6: Pin owned evidence and verify the complete issue

**Files:** Add `internal/mod/d2legacy/poison_damage_real_test.go` and any necessary synthetic support fixture under `internal/content/d2legacy/lua/d2legacy/tests/support/`. Keep implementation fixes with their owning files if the full-path test exposes a gap.

**Interfaces:** `TestOwnedPoisonDamageRecords` loads owned `ItemStatCost.txt` and `States.txt` only when `DARK_MAGIC_TEST_MPQ_DIRECTORY` is set, asserting the four stat IDs and poison `green`/`poisonhit` fields. One production-path Lua test starts with a melee poison hit, then verifies resisted rate, ranked state, raw life ticks, Cleansing, and death credit across players and monsters.

- [ ] Add the owned-record and production-path tests; run both and fix any uncovered issue-scoped gap.
- [ ] Run `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy -run '^TestOwnedPoisonDamageRecords$' -count=1`; confirm it ran rather than skipped. Run `make test-lua` for the final production path.
- [ ] Review the branch against each #463 criterion, evidence grade, deterministic event ordering, connected privacy, and the diff scope; fix only issue-scoped gaps.
- [ ] Commit the final issue-scoped changes, then run `make test`, `make architecture`, `make test-lua-hardening`, `make test-race`, and `git diff HEAD^ HEAD --check` on that exact commit. Record command results and any skips.
- [ ] Fill every PR-template section with scope, evidence, checks, unverified 1.14d behavior, and AI use; push, open the PR, wait for CI, rebase on current `main` if needed, rerun checks on the submitted commit, and merge only after CI passes. Close #463 only if all acceptance criteria are met.
