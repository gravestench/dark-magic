# Issue #648: Seamless Cold Plains–Stony Field Border

## Intent and evidence

A player in a newly created Act I game should see Stony Field terrain while approaching the Cold Plains border, walk across without a loading/black frame or camera jump, and see Cold Plains before crossing back. The level change remains an authoritative movement event. Cavern entrances and portals retain their discrete transition presentation. This is the user-visible rule already implemented for Rogue Encampment↔Blood Moor and Blood Moor↔Cold Plains, extended to the next outdoor pair.

The [Act I route](../../research/QUESTS_AND_WORLD.md) identifies Cold Plains→Stony Field as physical traversal, while the Cairn Stones lead to Tristram through a portal. [Warp research](../../research/WARPS_PORTALS_WAYPOINTS.md) documents the current two prepared outdoor seam pairs and leaves later links open. Issue #646 supplies an admitted headless Stony Field recipe with a semantic `cold-plains-entry`; its `worldgen.Zone.warps` storage does not mean a clickable portal. The exact original 1.13c/1.14d endpoint coordinates, border art, room layout, and seed call order remain unverified. The first placed Cold Plains exit and corridor are therefore **Dark Magic structural policy**, subject to later original-game comparison.

## Approaches

1. **Chosen: extend the existing reciprocal generated seam and shared-camera path.** Add a matched Cold Plains exit, generate Stony Field from its direction, publish a four-zone topology, and let the current seam authority and scene display the prepared connected component. Independent architecture and gameplay/evidence reviewers selected this approach because it has one transition authority and already renders neighboring tiles before a crossing.
2. Stitch the zones into one collision map. That changes coordinate, collision, command, and checkpoint ownership beyond #648.
3. Add a separate lazy handoff and rendering path at the border. A handoff after crossing fails the advance-visibility rule; preloading before crossing recreates the existing connected-map mechanism with duplicate ownership.

## New-world topology and legacy preservation

New games admit four ordered zones, levels 1–4, and three reciprocal seams. The new durable topology version is 3. The existing version-1 two-zone and version-2 three-zone snapshot encodings, checksums, zone sets, and restore paths stay byte-for-byte stable. Restoration consumes saved definitions rather than regenerating a Cold Plains recipe with new endpoints. A v1/v2 snapshot continues to contain only its original zones and does not gain Stony Field. A new v3 snapshot restores all four admitted zone definitions and all three bound seams with the same checksum; invalid zone order, seed/difficulty drift, missing or duplicate endpoints, wrong destination/role/direction, and mismatched seam geometry reject before mutable world recovery.

The current Go `EntryWorld` transport and validation remain generic host mechanisms. Lua map policy chooses the Cold Plains exit, Stony Field recipe, and 3↔4 seam roles. The v3 Go envelope adds only the minimum canonical storage and validation needed for this exact four-zone topology; it does not move Diablo placement policy into Go or alter v1/v2 serialization.

## Placed outdoor border and movement

In the v3 composition path, Cold Plains gains one `stony-field-exit` targeting level 4. Its edge faces opposite the admitted `blood-moor-entry`, at that edge's midpoint. A semantic dirt route connects the existing interior to this exit. The direct Cold Plains generator's current call shape and older topology definitions retain their old single-entry output; the new endpoint/path are selected explicitly for v3 composition. Existing Cold Plains fill/variant stream names and draw order stay stable. The exit's edge and midpoint are provisional Dark Magic policy, not a recovered original coordinate.

Stony Field receives the **admitted Cold Plains exit direction**, rather than choosing an edge independently. Its `cold-plains-entry` targets level 3 on the opposite edge. Bind each 3↔4 seam end to its exact admitted warp ID, role, direction, tile position, and destination, as the existing 2↔3 generated seam does. Reuse the established source trigger, collision-safe inset arrival, reverse-trigger separation, and movement transaction. Crossing 3→4 or 4→3 changes level once; stale connected movement samples stamped with the old source level are ignored. A new or restored v3 session keeps this behavior offline and connected after reconnect.

## Presentation and population boundary

The scene prepares all maps connected by published reciprocal seams under one camera root. Stony Field tiles and static map art enter the viewport before the player crosses from Cold Plains, and Cold Plains remains visible before the return crossing. Both maps remain placed while the active gameplay map changes. The camera and hero position carry continuously, including off-center crossings, repeated reversal, and panel camera anchors. Pointer projection and collision/commands continue to use only the active map. Audit the existing level-1–3 depth priority rule for a fourth map and prove correct ordering of overlapping floor/static art and hero; keep sparse tile residency bounded. A discrete cave/portal transition still changes to its separate map without outdoor visual carry.

The generic population adapter currently exposes every seam destination. Adding level 4 would otherwise feed it to Lua population policy that shares earlier-zone random streams. Issue #648 must prevent accidental Stony Field monster spawns and changes to existing-zone monster outcomes while monster population remains separately scoped. Preserve existing 1–3 population behavior and stream state; a later issue will establish level-4 monster policy from evidence.

## Verification and limits

- Synthetic mapgen tests prove a unique, correctly targeted Cold Plains exit, the opposite-facing Stony Field entry, dirt-route connectivity, all cardinal orientations and difficulties, deterministic v3 checksum, and rejection of mismatched endpoints. Pin v1/v2 snapshot bytes/checksums and restore behavior before changing generation; test restored v3 definitions without regeneration.
- Owned 1.14d materialization checks both 3↔4 sides for in-bounds walkable source/arrival points, reachable interior routes, reciprocal visual alignment, and a return point outside the reverse trigger. Keep MPQs, extracts, captures, and private saves out of Git.
- Production-path tests move an actor across 3→4→3 offline and connected, including a new-v3 checkpoint restore/reconnect and stale old-level samples. The existing 1↔2 and 2↔3 crossings remain unchanged.
- Scene tests show actual prepared neighbor terrain before both crossings, continuous camera/hero and static-art depth through off-center crossing/reversal, bounded tile residency, active-map-only pointer projection, and a discrete warp regression. Attempt local owned-data captures of both approaches when a graphical monitor is available; report any environmental capture limitation separately from the scene and authority results.
- Preserve v1/v2 map sets and checksums. New v3 worlds do not claim migration of legacy snapshots or original-game endpoint/layout parity. Cairn Stones, Underground Passage, quests, set pieces, and level-4 monster population remain separate issues.

## Builder decisions

Choose whether the v3-only Cold Plains exit is passed as an explicit argument to its strategy or selected by a narrow v3-specific Lua wrapper. Avoid a global default that changes old direct calls or saved definitions. Choose the smallest v3 envelope implementation that keeps v1/v2 byte encoding untouched. Determine the fourth map's depth priority using real overlap tests; the current scene's `id <= 3` branch must not silently decide it. Confirm the population skip at the Lua policy boundary without consuming level-1–3 random draws.
