# Visible Outdoor Seams Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show prepared neighboring outdoor terrain before and after crossing the Rogue Encampment–Blood Moor and Blood Moor–Cold Plains seams.

**Architecture:** Publish a copied presentation snapshot of the prepared maps through `engine.world/v1` and reuse existing directed seam initial data. A Lua helper derives shared pixel/depth offsets; the scene places sparse map tiles and active entities as siblings under one camera root and carries visual camera/hero offsets across an authoritative level switch. Authority and discrete warps stay unchanged.

**Tech Stack:** Go world capability and client app; first-party `d2legacy` Lua scene and sparse map adapter; Go and Lua acceptance tests; owned MPQ local scene test.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-641-seam-visibility-design.md`

## Global Constraints

- Issue #641 only; do not change authoritative generation, collision, admission, or warp semantics.
- Do not commit owned game data, captures, credentials, or saves.
- Keep existing `chunked_map` single-map behavior and bounded preload/residency.
- Claims about 1.13c/1.14d behavior are user-reported until independently verified.

## Review Focus

- Off-center legal trigger and rapid reverse crossing: visual hero and camera remain continuous, with authority unchanged (Task 4).
- Preview margins or absent tiles at the seam: viewport pixels beyond the source edge receive actual destination tile draws, not just canvas bounds (Tasks 3 and 4).
- Equal-depth overlap and upper walls: stable source-side priority and interleaving with the active hero (Tasks 3 and 4).
- Side-panel camera anchor changes during a crossing: pointer reverse projection follows the rendered floor (Task 4).
- Old two-zone restored topology and absent session world: snapshot and scene use only prepared levels without indexing level 3 blindly (Tasks 1 and 4).

---

### Task 1: Publish prepared presentation maps

**Files:** Modify `internal/runtime/lua/world_module.go`, `internal/runtime/lua/world_module_test.go`, `internal/app/clientapp/world_generation.go`, `internal/app/clientapp/lua_runtime.go`; test in `internal/app/clientapp/world_test.go`.

**Interfaces:** Define `PresentationWorld{ActiveLevel int, Maps []CurrentWorld}` and `PresentationWorldProvider func() PresentationWorld` in `world_module.go`. Add `SessionWorldModuleWithPresentation(source fs.FS, current CurrentWorldProvider, presentation PresentationWorldProvider, resolvers ...gameworld.ObjectResolver) Module`, retaining the existing constructor for other callers. `engine.world/v1.presentation()` returns `{active_level, maps=[{level_id,map,ds1,dt1}]}` in level-ID order. Map userdata uses the existing immutable map methods; recipe tables/slices are copied. Keep `current()` and `current_level()` behavior. `app.presentationWorld()` holds its world read lock only while copying prepared references and values.

- [ ] Write Go/Lua boundary tests: mutation of returned recipe tables cannot mutate the provider; no-session returns an empty snapshot; app snapshot includes only the prepared 2 or 3 levels.
- [ ] Run focused `go test ./internal/runtime/lua ./internal/app/clientapp -run 'Test.*(Presentation|World)' -count=1` and confirm the new assertions fail.
- [ ] Implement the minimal provider and capability command using the existing world-map userdata registration.
- [ ] Run the focused tests and confirm all new assertions pass.
- [ ] Commit this task.

### Task 2: Align map coordinates and depth at a prepared seam

**Files:** Create `internal/content/d2legacy/lua/d2legacy/presentation/outdoor_seams.lua` and `outdoor_seams_test.lua`.

**Interfaces:** `outdoor_seams.layout(snapshot, directed_seams)` returns `levels[level_id]={map,recipe,offset_x,offset_y,depth_offset}` and shared camera bounds. `outdoor_seams.pixel(level, map_pixel_x, map_pixel_y, layout)` and `outdoor_seams.depth(level, local_depth, layout)` translate only presentation facts. Input is the Task 1 snapshot plus the existing `seam_geometry.configured_seams()` result. Validate reciprocal endpoint/arrival alignment with a small pixel tolerance and reject malformed/disconnected layouts. Keep floor/shadow/roof pass bands while translating comparable upper-wall/entity depth.

- [ ] Write synthetic Lua tests for 1↔2 and 2↔3 endpoint alignment, reverse residual, shuffled map order, two-zone topology, malformed or missing seam, and global depth ordering of neighboring upper walls and an entity.
- [ ] Run the focused Lua suite through the repository Lua test command and confirm the new tests fail.
- [ ] Implement only the layout/projection helper.
- [ ] Run focused Lua tests and confirm the new tests pass.
- [ ] Commit this task.

### Task 3: Render sparse neighboring tiles in one depth-ordered tree

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/presentation/chunked_map.lua` and `chunked_map_test.lua`.

**Interfaces:** Add optional `shared_root` plus `offset_x`, `offset_y`, `depth_offset` to `chunked_map.create(parent, recipe, options)`; in this mode tile nodes attach directly to `shared_root`, use global tile coordinates/depth, and `chunked_map.update` culls against the shared camera transform without per-map clamping. `chunked_map.destroy` releases only nodes/jobs it owns and never destroys `shared_root`. Existing default mode remains unchanged.

- [ ] Write Lua tests showing a destination tile is queued/admitted before its canvas enters the active-map camera center; actual tile rectangles cover the seam viewport region, not merely canvas bounds; nodes from both maps and an entity share a sortable parent; off-screen tiles and pending jobs remain bounded/released; destroying one adapter leaves its sibling/root alive.
- [ ] Run the focused Lua suite and confirm the new tests fail.
- [ ] Implement the shared-root mode without changing default-mode output.
- [ ] Run the focused and existing `chunked_map` tests; confirm pass.
- [ ] Commit this task.

### Task 4: Compose the scene and carry visual continuity

**Files:** Modify `internal/content/d2legacy/lua/d2legacy/screens/game_world.lua`, `internal/content/d2legacy/lua/d2legacy/gameplay/ground_items.lua`; create focused scene tests in `internal/content/d2legacy/lua/d2legacy/screens/game_world_seams_test.lua`; extend `internal/app/clientapp/cold_plains_transition_test.go` for owned-data production path.

**Interfaces:** Scene creation loads Task 1 snapshot and Task 2 layout, creates one shared camera root, and creates Task 3 adapters for prepared levels. Active entity nodes, including ground items, use the active level's global pixel/depth translation and remain siblings of tile nodes. On level change, retain all terrain adapters, replace only active entity nodes, and compute camera/hero visual carry offsets from last rendered positions; ease them to zero over a fixed short duration. Pointer inverse projection uses the effective rendered camera anchor. Scene exit releases all adapters and root.

- [ ] Write synthetic scene tests for visible neighboring tile before crossing, bidirectional centered/off-center crossing with no first-frame camera/hero jump, rapid reverse crossing, pointer under side-panel anchor during settling, active-only dynamic entities, depth interleaving, and cleanup. Extend an owned-data test to inspect actual admitted tile draws on the production scene before/after both seams.
- [ ] Run focused Lua and Go tests; confirm the new assertions fail.
- [ ] Implement scene composition, global entity placement, visual carry, pointer anchor, and union-edge clamp.
- [ ] Run focused tests with `MPQ_DIRECTORY=/Users/dylanknuth/d2_english_mpq` for the data-dependent test; confirm pass and record which cases ran.
- [ ] Commit this task.

### Task 5: Verify and submit issue #641

**Files:** Update only issue #641 design/plan if implementation evidence changes; fill `.github/PULL_REQUEST_TEMPLATE.md` in PR description.

- [ ] Run on final commit: `make test`, `make architecture`, `make test-race`, `make test-lua-hardening`, `make build-client-backends`, focused owned-data tests, and the `capture-act1-seam` path if its local inputs are available. Record exact results and skipped data-dependent cases; keep captures local.
- [ ] Have independent correctness/evidence and formatting/style reviewers inspect final diff; style reviewer uses `refactor-for-readability`. Fix blockers and rerun affected checks.
- [ ] Write complete PR description before creating PR, push branch, attach PR, wait for CI, and merge when all required checks pass. Fast-forward the main checkout.
