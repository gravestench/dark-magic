# Issue #636: bridge-lip pointer route design

## Intent and evidence

A native character can remain on Rogue Encampment's southeast bridge after an attempted exit. The user's screenshots establish that visible symptom, but do not preserve the clicked pixel or prove the failing layer. Issue #636 owns the reproduction and any narrowly demonstrated input/navigation fix. The prepared town seam and Blood Moor map already exist; #634 separately owns an intermittent connected route stop before the bridge.

Owned 1.14d seed-one map diagnostics on `722fd7e4` found a concrete gap: the ordinary spawn can reach bridge shoulder positions `(269,80)` and `(269,81)`. From `(269,80)`, a radius-one path to the nearest published lane at `(275,82)` succeeds in seven points, while a path to `(275,80)` is unreachable. `(269,81)` can also reach `(275,82)`. The published trigger strip is Y82..86. `seam_geometry.pointer_target` currently retargets outside-map clicks only when their ray leaves through Y82..86, so a visually plausible eastward click from the shoulder keeps an unreachable goal. `MovementSource.replaceAcceptedRoute` then rejects that goal and stops. This is a Dark Magic interaction design issue; the precise retail 1.13c/1.14d trigger shape remains unverified.

## Approaches

1. **Widen the authoritative trigger.** This does not solve the unreachable radius-one outward route in shoulder lanes, and it risks accepting collision or floor under parapet art.
2. **Choose a reachable safe lane for near-opening outside clicks.** Keep the authored trigger and collision facts unchanged. When a town pointer ray first leaves the map on the named edge within two tangential subtiles of the published lane strip, and also crosses the trigger plane within that margin, choose the nearest strip lane at the prepared trigger plane. The existing pathfinder still rejects a blocked approach. This is the chosen approach because it changes only presentation targeting and uses the same prepared geometry that authority consumes.
3. **Snap every blocked click to a nearest walkable cell.** This would affect all maps and ordinary wall clicks and exceeds #636.

## Boundary

- A ground click beyond the town edge near the named bridge may route to the nearest published safe lane; a click beyond an unrelated border remains rejected.
- Clicks whose ray already resolves to a valid prepared strip lane keep their current destination. The exit trigger, collision flags, generated-area seams, and authoritative transition logic do not change.
- Selectable objects, attacks, stand-still clicks, and in-map ground clicks keep their current input precedence. This issue does not solve #634 or broaden map generation.
- The two-subtile shoulder margin is an inferred Dark Magic usability policy supported by the owned map's reachable Y80/81 cells. It is not presented as original Diablo II geometry.

## Verification

- First add a Lua regression with a production-shaped seam `(275,84)`, strip Y82..86, and source map 280×200, so `(269,80)` plus exact outside target `(281,80)` must return `(275,82)` rather than the old `(281,80)`. Pin the two-subtile margin with Y80 corrected and Y79 unchanged above the strip, plus synthetic lower-side Y88 corrected and Y89 unchanged. Cover both reachable upper shoulder lanes, an already valid strip ray, other town borders, in-map clicks, and a non-town level. A distant non-exit border target remains unchanged by this helper.
- Add an owned-data production-scene pointer test: reach `(269,80)` from a seam fixture, click beyond the east lip in lane 80, and assert the authoritative player enters level 2 in offline and connected sessions. This must fail on the old code for the intended reason before changing targeting. The test reads the presentation camera's ECS position and derives the effective click anchor from the production canvas and clamp rule. The seed-one dimensions predict no clamp at `(269,80)`; measure that precondition rather than assuming the hero equals the camera or hardcoding `(400,330)` without checking it.
- During the failing and passing test runs, record the logical cursor pixel, production reverse-projected subtile, radius-one path results for raw and corrected goals, authoritative position and level, and whether the prepared strip matches the approach. The Lua unit test directly records the requested goal from `pointer_target`: old code leaves exact input `(281,80)` uncorrected, while new code returns `(275,82)`. The production test asserts the rounded logical cursor still projects to the outside-east collision cell near lane 80; it logs the actual continuous coordinate and observes movement and committed level without claiming access to the private movement mailbox. The old-code RED must show route rejection, not just a crossing timeout; the new code must move into the strip and commit level 2. These observations separate click admission, route rejection, and seam failure without a new product diagnostic API.
- Run the existing town bridge round trips and owned corridor checks. CI has no licensed MPQs, so the owned-data tests must be reported as local execution and skips in CI.
- Compare the final native screenshot/click evidence if the user provides the clicked pixel or mode. If it points to a different failure boundary, keep this fix tied to its own demonstrated regression and update #636 with the remaining case.
