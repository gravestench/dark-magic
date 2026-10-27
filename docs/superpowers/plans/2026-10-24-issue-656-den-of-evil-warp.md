# Den of Evil Cave Warp Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a player intentionally enter the authored Den of Evil cave from Blood Moor and return, with a separate indoor map and durable authoritative state.

**Architecture:** Lua owns the deterministic outdoor cave stamp and one-room Den recipe. Go admits and materializes a v4 five-map topology, resolves one reciprocal cave-warp pair, and publishes it to the existing Lua world-warp operation. Presentation swaps to a separate Den component on the authoritative `warp` relocation; the outdoor seam component remains intact.

**Tech Stack:** Go, Lua, owned Diablo II Expansion 1.14d MPQs, deterministic worldgen recipes, ECS interaction admission, headless and native client adapters.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-656-den-of-evil-warp-design.md`

## Global Constraints

- Work only on issue #656; Den population, quest, loot, audio, waypoints, and other caves are outside this PR.
- Use the inspected 1.14d rows as data evidence, and label original placement, selection transforms, RNG order, and 1.13c equivalence unresolved.
- Preserve existing v1/v2/v3 topology restoration and all three visible outdoor seams; the Den is a discrete warp.
- Lua selects map policy; Go validates and materializes it. No game data, captures, saves, or credentials enter Git.
- Every task has a failing test before feature code, a scoped pass, a commit, and independent review. The final branch gets correctness, evidence, and `refactor-for-readability` style reviews.

## Review Focus

1. A one-room Den with no usable special marker must fail preparation before publication; Task 1/3 owned tests pin it.
2. An authored special tile with orientation 10/11 but the wrong MainIndex/SubIndex must not become a cave warp; Task 3 resolver tests pin it.
3. A far, wrong-level, stale, or mismatched-pair click must not relocate a player; Task 4 authority tests pin it.
4. A v1/v2/v3 checkpoint must restore its original level set without synthesizing Den; Task 2/3 snapshot tests pin it.
5. A cave trip and return must not hide neighboring outdoor terrain or regenerate earlier residents; Task 5 client acceptance pins it.

---

### Task 1: Deterministic one-room Den recipe

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/mapgen/maze.lua`, `cave_rooms.lua`, and `maze_test.lua`; add `internal/mod/d2legacy/adapter/mapgen/den_of_evil_real_test.go` for owned DS1 marker inspection.

**Interfaces:** Consume `maze.generate(level_id, seed, difficulty)` and existing `worldgen.admit`. Produce a level-8 `worldgen.Zone` with exactly one 24×24 authored previous-level stamp, a stable `den-exit` semantic warp targeting level 2, and an eligible variant chosen by a Den-specific named stream. Preserve multi-room cave behavior.

- [ ] **Step 1: Write RED tests.** `TestDenOfEvilOneRoomRecipe` (synthetic Lua) requires level ID 8, one room/stamp at 24×24 in each difficulty, a single `den-exit` warp to 2, and stable checksum for a fixed seed. A second test rejects absent/invalid authored variants and proves an ordinary multi-room cave recipe is unchanged. `TestDenOfEvilOneRoomReal` (owned Go) materializes the selected 1.14d DS1 and requires exactly one full-identity cave-up marker and a passable player-sized point near it.
- [ ] **Step 2: Run the targeted Lua mapgen tests.** Use `make test-lua`; confirm the new one-room test fails from the special-room mask assertion before implementation.
- [ ] **Step 3: Implement the one-room policy.** Select one of authored defs 83–86 and its variant in Lua; record orientation/variant in the admitted recipe; do not create a second room. Keep special-room connection-mask handling for multi-room mazes unchanged. The `den-exit` warp records the authored reciprocal level relationship, not a guessed selector rectangle.
- [ ] **Step 4: Run `make test-lua` and `DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy/adapter/mapgen -run TestDenOfEvilOneRoomReal -count=1`.** Require a one-room materializable Den with a unique full-identity exit marker; if no authored variant supports the contract, stop and revise the issue/design before continuing.
- [ ] **Step 5: Commit the scoped Lua recipe/tests.**

### Task 2: Authored Blood Moor entrance and v4 recipe policy

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/mapgen/outdoor.lua`, `entry_world.lua`, `outdoor_test.lua`, and add `entry_world_v4_test.lua`.

**Interfaces:** Consume Task 1's `maze.generate(8, seed, difficulty)`. Produce Blood Moor with one definition-51 cave-entrance stamp and semantic `den-entrance` warp targeting 8; new Lua entry recipe returns zones 1/2/3/4/8, the same three outdoor seam specifications, and one reciprocal cave-warp policy. Version 1–3 Lua entry paths remain unchanged.

- [ ] **Step 1: Write RED tests.** `TestBloodMoorDenEntranceRecipe` requires one 8×8 authored cave-entrance stamp, a unique `den-entrance`, a player-sized route from the established outdoor path, no overlap with town/Cold Plains reserved joins, deterministic seed replay, all difficulties, and unchanged earlier stream outputs. `TestEntryWorldV4Recipe` checks exact level set and unchanged three seams.
- [ ] **Step 2: Run `make test-lua`; inspect the expected missing cave stamp/v4 failure.**
- [ ] **Step 3: Add a Den-specific named placement/variant stream and v4 Lua composition.** Choose an eligible cell away from the two outdoor joins, join its authored DS1/DT1 resources, and reserve a route. Do not use image pixels or filename parsing to choose destination. Older recipe functions keep their original outputs.
- [ ] **Step 4: Run Lua mapgen and generated-zone checks, then commit.**

### Task 3: v4 snapshot, materialization, and reciprocal endpoint validation

**Files:** Modify `internal/mod/d2legacy/adapter/mapgen/entry_world.go`; add `entry_world_v4_test.go`; modify `internal/mod/d2legacy/adapter/entryworld/build.go`, `entryworld.go`; add `cave_warp.go`, `cave_warp_test.go`, and owned `den_of_evil_real_test.go`.

**Interfaces:** Add `mapgen.CaveWarpSpec` (levels 2/8, roles `den-entrance`/`den-exit`, level/warp slots 3/0) to `EntryWorld` and v4 snapshot. Add `entryworld.CaveWarpPair` with two trusted endpoint IDs, source/arrival coordinates, target level, and destination dimensions to `Prepared`. `resolveCaveWarp(records, zones, worlds, spec) (CaveWarpPair, error)` validates full marker identities and both player-sized routes. Its validated Den arrival becomes `Prepared.Spawns[8]`, so `Prepared.Destination(8)` and restored admission use the same trusted point. `Prepared.WarpData()` will publish the pair in Task 4.

- [ ] **Step 1: Write RED snapshot tests.** A v4 round trip checks canonical bytes/checksum and exact level set `{1,2,3,4,8}` plus three seams and reciprocal cave link. Corrupt level IDs, duplicate/missing markers, bad destination/slot, and mismatched pairs fail before publication. `Prepared.Destination(8)` returns the validated Den arrival on fresh preparation and restore. Existing v1/v2/v3 fixture snapshots still restore their original sets.
- [ ] **Step 2: Run focused `go test ./internal/mod/d2legacy/adapter/mapgen ./internal/mod/d2legacy/adapter/entryworld -run 'Test.*Den|Test.*V4|Test.*Legacy' -count=1`; confirm RED.**
- [ ] **Step 3: Implement v4 encoding/validation and canonical nonconsecutive materialization order.** Validate the expected level IDs explicitly rather than looping `1..len(Zones)`. Do not synthesize Den when restoring older versions. Keep the v1–v3 snapshot bytes and semantics intact.
- [ ] **Step 4: Resolve the pair after materialization.** Join unique `world.Map.SpecialTiles` by orientation/MainIndex/SubIndex to the semantic zone warps and typed `Levels`/`LvlWarp` entries; do not use a provisional Lua warp tile coordinate as the collision anchor. Fail closed on missing/ambiguous markers, unsafe positions/routes, or asymmetric links. Record the validated Den arrival in `Spawns[8]`; only return `Prepared` after all checks succeed.
- [ ] **Step 5: Run focused synthetic and owned 1.14d tests for all difficulties, plus legacy snapshot tests; commit.**

### Task 4: Production static-warp authority and tokenless presentation

**Files:** Modify `internal/mod/d2legacy/adapter/entryworld/bootstrap.go`, `internal/app/clientapp/application.go`, `world_generation.go`, `world_transitions.go`, `session_bootstrap.go`; modify `internal/content/d2legacy/lua/d2legacy/systems/world_warp.lua`, `screens/game_world.lua`, and focused Lua/Go tests.

**Interfaces:** `Prepared.WarpData() map[string]any` emits two static endpoints from Task 3. `application` retains the prepared pair and `warpBootstrapData()` supplies it in production, while `warp_lab` keeps its fixture pair. `world_warp.create` accepts tokenless static endpoints, still creates selectable and interaction components, and creates `warp_appearance` only for token-bearing portals.

- [ ] **Step 1: Write RED authority and presentation tests.** Static endpoints appear once in production bootstrap with reciprocal IDs and no appearance token; warp-lab portals still have their tokens. Wrong-level, far, stale, and mismatched-pair operations do not relocate; explicitly mutate or remove the reciprocal static endpoint after bootstrap to prove runtime pair validation, since bootstrap validation alone cannot prevent drift. Admitted use sets kind `warp` and correct destination. A tokenless cave endpoint remains selectable but yields no portal COF render request.
- [ ] **Step 2: Run focused Go and Lua tests and observe the production-empty/token-required failures.**
- [ ] **Step 3: Wire `Prepared.WarpData()` through publication and extend `world_warp` only for tokenless static endpoints.** Keep existing town-portal owner/expiry rules and ordinary interaction admission. Recheck that the clicked static endpoint still has a live reciprocal partner and matching destination immediately before `transition.apply`; do not rely solely on bootstrap validation. The scene renderer skips portal sprite construction for tokenless endpoints; DS1 geometry supplies the cave graphic.
- [ ] **Step 4: Run the focused tests, `make test-lua-format test-lua-syntax test-lua-hardening`, and commit.**

### Task 5: Production journey, restore, and evidence

**Files:** Add `internal/app/clientapp/den_of_evil_transition_real_test.go`; extend focused entry-world acceptance where needed; update `docs/research/WARPS_PORTALS_WAYPOINTS.md` with verified rows and unknowns.

**Interfaces:** Consume the v4 `Prepared`, static warp pair, existing offline/connected app harnesses, and the unchanged outdoor seams. No new gameplay owner.

- [ ] **Step 1: Write RED owned-data acceptance.** For each difficulty, prepare from the owned 1.14d MPQs. Walk near/onto the cave entrance without clicking and assert no relocation. Use the ordinary pointer/selectable click path to enter and exit in offline and connected sessions; direct operation alone is insufficient. After each click/warp assert authority/public level and safe positions, distinct indoor/outdoor scene components, and camera centered on the committed player. Assert wrong-level/far/stale requests fail, earlier resident IDs and all three seamless outdoor joins survive return, connected reconnect agrees, and a current-binary checkpoint restores an active Den and a later return.
- [ ] **Step 2: Run the focused owned tests and inspect the first missing behavior, then implement only integration gaps discovered inside #656.** No fixture-only substitute may satisfy the production tests.
- [ ] **Step 3: Run focused owned acceptance while developing.** Use `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq DARK_MAGIC_TEST_MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq go test ./internal/mod/d2legacy/adapter/mapgen ./internal/mod/d2legacy/adapter/entryworld ./internal/app/clientapp -run 'TestDenOfEvil' -count=1`; name new owned tests with that prefix. Use `env -u MPQ_DIRECTORY -u DARK_MAGIC_TEST_MPQ_DIRECTORY go test ./internal/mod/d2legacy/adapter/mapgen ./internal/mod/d2legacy/adapter/entryworld ./internal/app/clientapp -run 'TestDenOfEvil' -count=1 -v` and inspect explicit `SKIP` lines.
- [ ] **Step 4: Commit acceptance and research, request independent correctness/evidence/style reviews, fix and re-review findings, and commit every fix.** On the last submitted commit run focused owned and no-MPQ checks, `make test`, `make architecture test-lua-format test-lua-syntax test-lua-hardening lint build-client-backends GOLANGCI_LINT='go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2'`, standalone `make test-race`, and `git diff --check main...HEAD`; rerun affected checks after any later commit. Record exact results in the complete PR template. If no native GPU display is available, state that pixel-level cave/bridge appearance remains unverified. Push a draft PR and merge only after required CI passes, as the user authorized.
