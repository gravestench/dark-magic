# Issue #625 Bridge Follow-up Implementation Plan

> For agentic workers: use `superpowers:executing-plans` or `superpowers:subagent-driven-development` task by task. The linked issue is the acceptance boundary.

**Goal:** Make the displayed Rogue Encampment exits usable with ordinary screen clicks from normal entry, including clicks beyond the east bridge lip and all four supported town directions.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-625-bridge-followup-design.md`

**Ownership:** The resolver publishes the selected town corridor. Shared seam geometry lets the ordinary pointer choose a reachable movement point and lets authority detect the same trigger. Only the Lua collision-phase transition changes levels. Generated wilderness seams and the reverse Blood Moor trigger keep their circular shape.

## Task 1 — Reproduce the missed player route

**Files:** `internal/app/clientapp/town_bridge_pointer_real_test.go`, `internal/mod/d2legacy/adapter/entryworld/town_exit_real_test.go`

- [x] Add a real scene pointer click beyond the east bridge map edge. Verify it is projected inside the gameplay viewport and begins outside the transition trigger. On unchanged production code, record the red position and pending movement state.
- [x] Check the northmost radius-one walkable bridge lane. Y=82 crosses on the baseline; Y=81 is blocked, so retain the positive lane guard without changing the east trigger radius.
- [x] Inspect the four owned town layouts. Reject the floor-only wide-border diagnostic because it includes ground beneath visible fences.

## Task 2 — Publish only the selected visible town aperture

**Files:** `internal/mod/d2legacy/adapter/transition/seam.go`, `seam_test.go`, `internal/mod/d2legacy/adapter/entryworld/bootstrap.go`, `entryworld_test.go`, `town_exit_real_test.go`, `internal/content/d2legacy/lua/d2legacy/systems/world_transition.lua`, `world_transition_test.lua`

- [x] Test a synthetic floor strip extending under neighboring fence art, then reproduce exact safe lane bounds for seeds 1, 2, 5, and 9 in owned data. A raw wall-tile filter is insufficient: bridge parapet tiles overlap traversable floor and fence sprites extend beyond materialized wall tiles.
- [x] After guide-nearest corridor selection, intersect it with an inferred three-tile policy window around the selected lane, inset one cell at either side. Publish `TriggerStrip` only on the authored town endpoint and validate both arrival clearances.
- [x] Serialize optional lane bounds and source direction. Keep the reverse and generated triggers circular. Test serialized directions and bounds.
- [x] Test all published lanes for a radius-one spawn route and an outward safe walk. Require radius-one return paths from town and moor arrivals. On authoritative Lua ticks, accept every synthetic east strip lane, reject both adjacent lanes, and exercise north, south, and west trigger orientation.

## Task 3 — Route only deliberate outside-map exit clicks

**Files:** `internal/content/d2legacy/lua/d2legacy/systems/world_transition.lua`, `world_transition_test.lua`, `internal/content/d2legacy/lua/d2legacy/screens/game_world.lua`, `internal/app/clientapp/town_bridge_pointer_real_test.go`

- [x] Add a red Lua test for the beyond-map pointer target. Put shared seam geometry in a dependency-free Lua module so screen-only tests remain independent of ECS. Project only ordinary ground movement toward the prepared town seam; leave selection, attacks, interactions, and skill targeting unchanged.
- [x] Apply the review finding: require the first map-boundary intersection to be the seam's authored edge and a published lane. Test a ray that grazes the trigger but exits far from the opening and one that first exits another boundary.
- [x] Verify cardinal pointer projections in Lua and the east bridge beyond-lip and return in offline and connected sessions, a held pointer crossing, a non-exit border negative case, and a normal-entry sequence of viewport clicks across and back offline and connected. Alternate stamps have geometry and authority orientation coverage, not full screen-click coverage.

## Task 4 — Review, verify, and integrate

- [ ] Ask independent correctness/evidence and formatting/style subagents to review the final diff. The style reviewer loads `refactor-for-readability`. Iterate findings until both approve.
- [ ] Commit the design, plan, implementation, and tests. Rebase on current `main` and run the focused owned-data tests, `go test -count=1 ./...`, `make test-lua-hardening`, `make architecture`, `make build-client-backends`, and formatting/diff checks on the submitted commit. Record actual results and skips.
- [ ] Fill every `.github/PULL_REQUEST_TEMPLATE.md` section, push and attach the PR, wait for CI, and merge only after all required checks pass. Fast-forward the clean `~/src/dark-magic` checkout so the user's build contains the fix.

## Limits

The exact original Diablo II trigger geometry is unverified. The three-tile aperture is a documented Dark Magic policy supported by owned DS1 previews and the four selected layout regressions. Never commit game data, visual captures made from it, credentials, or private saves.
