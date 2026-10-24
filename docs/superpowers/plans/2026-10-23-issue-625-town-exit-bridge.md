# Issue #625 Town Exit Bridge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make walking through the visible Rogue Encampment exit cross into Blood Moor and permit the return trip.

**Architecture:** Resolve the town endpoint from authored floor and collision at the stamp-named edge, using the existing special marker to select among walkable corridors. Keep the published seam and transition transaction unchanged; verify geometry with synthetic and owned maps, then verify a normal screen-space click in a local capture.

**Tech Stack:** Go, owned Diablo II 1.14d MPQ fixtures, existing Lua game-world input, native capture fixture.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-625-town-exit-bridge-design.md`

## Global Constraints

- Stay within issue #625: first town seam resolution, its tests, and only the fixture support needed for the pointer acceptance capture.
- Do not commit game data, its captures, credentials, or private saves.
- Treat edge placement as Dark Magic policy; no claim of exact 1.13c/1.14d trigger geometry.
- Preserve generated seam resolution, transition authority, and topology format.

## Review Focus

- Multiple floor corridors at the named edge: the guide selects the nearest corridor deterministically; Task 1 tests it.
- A transparent edge without an authored floor placement: preparation fails; Task 1 tests it.
- A floor corridor blocked by walls, disconnected from the guide, or ending short of the border: preparation fails or selects a continuous corridor; Task 1 tests it.
- Return arrival exists but cannot route to the trigger: preparation fails; Task 1 tests it.
- A normal click on the visible east bridge may aim outside collision bounds: Task 3 tests screen-space clicks from inside the bridge to the seam in offline and connected play and checks committed Blood Moor and return.

---

### Task 1: Resolve the town seam at its authored edge

**Files:**
- Modify: `internal/mod/d2legacy/adapter/transition/seam.go`
- Modify: `internal/mod/d2legacy/adapter/transition/seam_test.go`

**Interfaces:**
- Consumes: `Map.AuthoredExitAnchors`, `Map.Tiles`, `Map.WalkableStep`, `Map.FindPath`, `Map.OpenPointNearExit`, `SeamSpec.FirstDirection`.
- Produces: `resolveAuthoredEdge(spec d2mapgen.SeamSpec, world *gameworld.Map) (float64, float64, error)` returning a walkable point one tile inside the selected named edge, with a walkable authored-floor lane to its border.

- [ ] Write red table tests for all four directions on synthetic 8x8 tile maps with an interior special guide and authored floor at the intended edge. Assert edge coordinate, corridor preference and stable tie, floor absence, blocked/disconnected and outward blind-alley corridors, and an unreachable return arrival.
- [ ] Run `go test -count=1 ./internal/mod/d2legacy/adapter/transition -run 'TestTown|TestResolveAuthored'` and confirm the new tests fail against the old interior endpoint.
- [ ] Implement `resolveAuthoredEdge` with a one-tile-inset collision row/column, authored floor membership, radius-1 footprint check, guide route, straight radius-1 authored-floor continuation to the outermost footprint-safe border, adjacent corridor grouping, and deterministic corridor/point selection. Validate the town-side inset return route in `ResolveSeam` before constructing `Seam`.
- [ ] Run `go test -count=1 ./internal/mod/d2legacy/adapter/transition` and confirm pass.
- [ ] Commit resolver and synthetic tests.

### Task 2: Pin the owned town layouts and ordinary walking route

**Files:**
- Create: `internal/mod/d2legacy/adapter/entryworld/town_exit_real_test.go` (or the existing owned entry-world test file if its fixture helpers are needed)

**Interfaces:**
- Consumes: `entryworld.Build`, `Prepared.Worlds`, `Prepared.Spawns`, `Prepared.Seams`, `Map.FindPath`.
- Produces: a data-gated regression for seeds 1, 2, 5, and 9.

- [ ] Write the owned-data test: each selected town seam lies on authored floor at its named edge, has a straight radius-1 authored-floor continuation to the footprint-safe border, has a route from its ordinary spawn and from its return arrival, and preserves round-trip directions. Seed 1 must lie on the visible `Bridge.dt1` floor corridor; a route from spawn to the bridge must enter the seam's two-subtile radius.
- [ ] Run the test with `MPQ_DIRECTORY` and `DARK_MAGIC_TEST_MPQ_DIRECTORY` set to the owned fixture directory; confirm it fails at seed 1 before Task 1 if available, or record the red result from the Task 1 test.
- [ ] Make only fixture/assertion corrections needed to match the approved design; rerun the same owned test and confirm pass.
- [ ] Commit the owned regression.

### Task 3: Prove the normal pointer path and integrated round trips

**Files:**
- Modify only if needed: `internal/app/clientapp/world_generation.go`, `internal/app/clientapp/pointer_acceptance.go` (development fixture behavior only)
- Test: `internal/app/clientapp/cold_plains_transition_test.go` and a focused pointer fixture test if Task 3 adds fixture code.

**Interfaces:**
- Consumes: prepared first seam endpoint and existing `FixtureWorldSpawn="seam"`, `FixturePointerMove`, `pointerMovementAcceptance.Frame`.
- Produces: screen-space clicks through normal Lua input from near the visible exit in offline and connected play, with an uncommitted capture showing Blood Moor.

- [ ] Run the existing offline and connected town/Blood Moor round trips against owned data after Task 2. Record any failure before changing fixture support.
- [ ] Add a seam-target fixture mode only if the existing acceptance click cannot target the visible bridge; test that its cursor maps to the resolved seam without bypassing normal input.
- [ ] Run native `make capture` with the near-exit fixture and one injected pointer click; inspect the capture and authority level. Keep generated captures outside Git.
- [ ] Exercise the same ordinary pointer-click path in an owned offline and connected client regression, asserting committed Blood Moor and return in both modes; direct controller movement does not satisfy this step.
- [ ] Run the focused owned client tests; confirm offline and connected pointer-click round trips.
- [ ] Commit only required fixture code/tests, if any.

### Task 4: Final verification and PR

**Files:**
- Update issue #625 PR body from `.github/PULL_REQUEST_TEMPLATE.md`; do not include captures in Git.

**Interfaces:**
- Produces: reviewed PR whose submitted commit passes exact local checks and CI.

- [ ] Ask independent task/evidence and formatting/style subagents to review the final diff; the style reviewer uses `refactor-for-readability`. Fix actionable findings and repeat reviews.
- [ ] Rebase onto current `main`; run `git diff --check`, focused owned-data tests, `go test -count=1 ./...`, `make test-lua-hardening`, `make architecture`, `make test-race`, `make build-client-backends`, and lint on the final commit. Record exact outputs and any asset-gated skips.
- [ ] Complete every PR template section, push, open and attach the PR, wait for required CI, merge when green, and update the primary checkout's `main`.
