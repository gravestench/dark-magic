# Issue #625: align the Rogue Encampment seam with its visible exit

## Intent and boundary

A player who walks to the exit shown by the selected Rogue Encampment layout must cross into Blood Moor and be able to return. The existing transition transaction, topology version, wilderness generation, and authority ownership stay in place. This fix changes only how the town side of the first seam is resolved and how that choice is verified. Original 1.13c/1.14d trigger geometry remains unverified; this is a bounded Dark Magic policy grounded in the owned DS1 art and collision map.

## Evidence that exposed the defect

The current resolver selects the cardinal-most orientation-10/11 special tile. In owned `TownE1.ds1`, the three such tiles are `(30,13)`, `(32,13)`, and `(34,20)` in tile coordinates. The chosen one resolves to `(173,103)` in subtiles. A new-character route to the visible southeast bridge at `(276,84)` is collision-safe and remains at least 26 subtiles from that trigger. A local capture at `(270,84)` matches the user's screenshots, while a capture at the current seam shows a different opening near tents. A direct controller movement and a one-click diagnostic fixture aimed specifically at the old trigger both cross; their tests missed the visible bridge.

The owned `TownE1` bridge uses floor tiles from `data/global/tiles/Act1/Outdoors/Bridge.dt1` at tile X `47..55`, Y `15..18`. At one tile from the east edge, its player-footprint-safe floor corridor is subtile Y `81..86`. The other owned town variants selected by seeds 2, 5, and 9 face south, west, and north. Their visible exits differ, but each has a walkable authored-floor point one tile inside its named edge reachable from that variant's ordinary spawn. The normal production seed is 1; variant checks protect the supported generator choices.

## Approaches considered

1. **Project the authored exit guide to the named walkable edge (chosen).** Keep the selected DS1 special marker as a guide, then scan one tile inside the direction named by the selected town stamp. Admit only points on authored floor with a collision-safe player footprint, an interior route, and a straight collision-safe continuation on authored floor to the outermost footprint-safe border cell. Group adjacent edge points into corridors, select the corridor nearest the guide, and place the trigger on the guide's perpendicular coordinate if it lies inside that corridor, otherwise at its midpoint. This reaches the east bridge and works for other town directions without a source-asset filename special case. It adds bounded geometry validation but does not imply original-game coordinate parity.
2. Match `Bridge.dt1` explicitly. This is strong evidence for `TownE1` but cannot resolve north, south, and west, where the owned maps have no bridge tile reference.
3. Store four layout-specific coordinates. This would make current owned assets pass but would duplicate the DS1 geometry, drift with replacement assets, and bypass the generator's role data.

## Resolver contract

`ResolveSeam` continues to distinguish the authored town side from generated seams. The town resolver receives the selected cardinal direction and materialized map. It uses the existing cardinal-most special marker only as a guide. It scans exactly the collision row or column one tile inside that named edge, requiring:

- a decoded floor placement at the candidate's tile;
- a radius-1 player footprint clear in authoritative collision;
- a `FindPath` route from the open guide position to the candidate using that footprint.
- a straight, radius-1 walkable sequence across authored floor from that candidate to the outermost footprint-safe cell of the named map edge.

Adjacent eligible candidates form corridors. Rank corridors by perpendicular distance from the guide, then by stable row/column order; select a point inside the chosen corridor as described above. If no eligible corridor exists, preparation fails before publishing the entry world. This prevents a generic transparent or apparently open border, and an interior blind alley, from becoming an exit. An owned-data review found that the original interior-only rule would select blind alleys for west and north variants, so outward continuity is a required admission condition before corridor ranking. The endpoint remains on walkable floor and retains the existing trigger radius and arrival inset. Validate the town-side return route from its inset arrival to the chosen trigger as part of preparation. The resulting endpoint is still a resolved runtime fact, so checkpoint and network topology policy do not change.

The first-seam endpoint moves for newly generated topologies. Existing checkpoint identities include the resolved topology's source recipe, not the materialized coordinate; a saved checkpoint prepared under the old runtime may fail compatibility preflight. Do not weaken preflight or add migration in this issue.

## Verification

1. Write a red synthetic resolver test with an interior orientation-10 guide and a narrow edge floor corridor. It must assert the edge endpoint, all four cardinal directions, a blocked/no-floor failure, and a reachable inset arrival. Existing tests that modeled a whole open map without floor placements will be updated to represent an authored corridor.
2. Add an owned-data test for all four selected town variants. Assert each seam is on authored walkable floor at the named edge, has a straight collision-safe authored-floor continuation to the outermost footprint-safe border, and that the ordinary spawn can route to it. For seed 1, assert the resolved east endpoint sits on the `Bridge.dt1` corridor and that the route to the visible bridge crosses the seam radius; the old `(173,103)` endpoint must fail this test.
3. Exercise one ordinary screen-space click through the production capture/input path from a near-exit fixture. The resulting capture must show committed Blood Moor after the click. Retain the local capture only as uncommitted evidence; game assets and captures must not enter Git.
4. Rerun offline and connected seam round trips, recovery, full Go/Lua/architecture/race/build/lint gates on the final commit. State precisely which owned-data checks ran locally and which skip in CI.

## Limits

The edge placement rule is an implementation choice, not a claim about retail `LvlWarp` selection, transition rectangle, or seed order. The issue does not alter Blood Moor or Cold Plains generation, monsters, pathfinding, or renderer projection. If research reveals original version-specific exit coordinates or an authoritatively typed DS1 exit marker, replace this policy in a focused follow-up.
