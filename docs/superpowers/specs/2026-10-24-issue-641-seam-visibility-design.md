# Issue #641: visible outdoor seams

## Intent and evidence

The player can reach the southeast Rogue Encampment bridge and cross to Blood Moor, but the area beyond the bridge is black until the active level changes. The player expects the neighboring outdoor terrain to be visible while approaching and to remain visible as the crossing completes. The same rule applies to the prepared Blood Moor–Cold Plains seam. Cavern entrances and portals remain discrete warps. This is a requested presentation behavior for the project's 1.13c/1.14d parity target; exact original-game edge placement and draw order have not been independently verified.

Issue [#641](https://github.com/gravestench/dark-magic/issues/641) defines this implementation slice. Prepared entry-world maps and reciprocal seams already exist. `clientapp.currentWorld` exposes only the active map to `engine.world/v1`; `game_world.lua` destroys its map node when the active level changes; `chunked_map.update` clamps each map independently to the screen. The existing crossing and admission paths remain authoritative.

## Goals

- Show generated terrain and static map art from the neighboring outdoor zone whenever the viewport crosses a prepared seam, before and after authority changes level.
- Use one continuous camera placement across the crossing, including overlay-panel camera anchors, without a black strip or single-map clamp at the seam.
- Keep tile decoding, preload, and retained node counts bounded to the viewport and its margin.
- Preserve one active gameplay map for collision, commands, units, warps, and pointer reverse projection; changing the visible layers must not create a second gameplay authority.
- Cover both directions of the prepared 1↔2 and 2↔3 seams in synthetic acceptance tests and run an owned-data scene/capture path locally.

## Non-goals and anti-goals

- This slice does not project dynamic units or objects from the inactive level, alter spawn/AI/collision, change warp semantics, or generate new zones.
- A backdrop painted beyond the edge without the actual prepared neighboring map would fail the requested behavior.
- Combining maps into one collision world would expand the change beyond issue #641.
- A crossing that still destroys visible neighbor terrain or jumps the camera visibly is incomplete, even if authority switches correctly.

## Approaches

1. **Selected: shared presentation coordinates for separate maps.** Expose a copied, read-only snapshot of prepared outdoor maps and directed seam geometry to the scene. Place each map's sparse tile root at a seam-derived offset under one camera. Keep the active map as the only gameplay map. This reuses prepared content and sparse rendering, with a small additive world capability and scene composition change.
2. **Stitch maps into a giant world.** This simplifies drawing but changes map generation, collision, transition ownership, and checkpoint assumptions. It is outside this issue.
3. **Draw a fixed edge extension.** This is smaller but cannot show the generated neighboring layout and would produce a visible switch later.

## Design

The application publishes a consistent, immutable presentation snapshot: current level, prepared level maps with their existing DS1/DT1 recipes, and the two reciprocal directed seams for each outdoor pair. The snapshot is copied at the Lua boundary; it exposes map handles as read-only map userdata and copied scalar/table fields. No mutable application registry is exposed. A presentation helper validates each reciprocal pair and computes per-level isometric pixel offsets from the prepared endpoint geometry. It aligns the source transition point with its destination arrival and checks the reverse mapping's residual. If the two mappings cannot share one placement within a tight pixel tolerance, the helper reports the mismatch rather than silently selecting a layout that teleports the camera. The TDD fixture and owned-data check will determine whether this alignment is valid for the existing generated maps.

The game scene holds one sparse `chunked_map` instance per prepared outdoor map while this scene is active. Each has a root offset in the shared presentation space. All roots receive the same camera transform; `chunked_map` gets an explicit placement mode that skips its single-map edge clamp while retaining its normal viewport-bucket residency, preload, and per-frame admission limits. The active map root still owns the hero, units, warps, debug overlays, and ground items. On a level change, those active-level children are replaced and rebound to the new active root; the terrain roots remain. Layer/depth ordering must make the outgoing authored bridge visible over any overlap and let the destination terrain appear past the bridge edge, with a deterministic active/inactive ordering at each crossing. The scene computes pointer projection from the actual shared camera anchor; it does not interpret clicks as commands in the inactive level.

At the outer boundary of the prepared three-zone group, the camera may still clamp to available terrain. At an internal seam, the clamp must use the union of visible prepared terrain, not the active map's own canvas. The scene must not stall while adjacent tiles are first decoded: preload into the bounded margin before they enter the physical viewport. The scene releases all retained roots and jobs on exit.

## Acceptance and verification

- Synthetic tests exercise bidirectional seam offsets, reverse residual validation, level-switch camera continuity, active-map-only pointer projection, edge clamp behavior, and bounded adjacent-map tile residency.
- A real-data test or local capture using legally obtained MPQs exercises the town bridge and the Blood Moor–Cold Plains seam on the production scene path, checking that neighboring terrain becomes visible before crossing and remains so after crossing. Captures and game data stay local.
- Run focused tests, applicable Lua tests, `make test`, `make architecture`, `make test-race`, and client backend build. Report exact submitted-commit results and any skipped data-dependent tests in the PR.

## Builder decisions still open

- Choose the smallest read-only `engine.world/v1` API shape that publishes prepared maps and seams without exposing authority state for mutation.
- Determine the renderer's root/depth rule for overlapping terrain from its existing draw contract; encode the rule in a scene test.
- Measure reciprocal alignment residuals against the owned 1↔2 and 2↔3 maps. If the prepared transition points do not permit continuous placement, report the mismatch and revisit this design before changing authority geometry.

