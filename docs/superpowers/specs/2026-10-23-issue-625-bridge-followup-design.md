# Issue #625: finish Rogue Encampment's visible exit crossing

## Boundary

A new character on current `main` still stops at the southeast bridge after PR #629. Issue #625 requires ordinary screen clicks to cross the selected visible Rogue Encampment opening into Blood Moor and return, offline and connected. It also covers the four supported town stamp directions. This change does not alter wilderness generation, other warps, combat, or Diablo II data.

## Evidence

The existing client regression spawned beside the prepared seam and clicked its exact in-map point. Its passing result did not cover the user's click toward black space beyond the bridge. A new owned-data regression projects an outside-map point through the visible bridge into the gameplay viewport. On the unchanged code, the character remains at (269,84), six subtiles from the (275,84) trigger, with no movement target. The path planner correctly rejects an outside-map goal. A click on the northmost walkable east bridge lane, Y=82, already crosses; Y=81 has floor art but is not walkable for a radius-one player.

The four owned town stamps reveal another limit of the old circular trigger. East's bridge floor admits lanes Y82..86, within radius two of its selected point. South, west, and north have wider displayed exits with safe lanes X146..158, Y106..118, and X146..158. Enlarging the shared circle would reach the inset town arrival and risk immediate retrigger. Raw DS1 fence art is wider than materialized wall placements, while generic floor continues under some visible fence. The hidden authored exit markers are offset from the displayed edge in three layouts, so neither all open border cells nor an aperture centered on a marker identifies the opening.

## Design

Keep `resolveAuthoredEdge`'s guide-nearest, authored-floor, outward-walkable corridor selection. After selecting that corridor, bound its tangential span to a three-tile policy window around the selected lane, inset one player-radius cell on each side, and intersect it with the actual walkable corridor. Publish these inclusive lane bounds on the town-side endpoint only. The existing radius supplies the strip's normal depth. Reverse and generated seams remain circles. Validate that the town arrival is beyond strip depth and the Blood Moor arrival is outside its reverse circle. Serialize the source direction and optional lane bounds. A shared seam-geometry module reads those prepared facts for both presentation targeting and authoritative trigger checks; it has no ECS dependency. The collision-phase system remains the only owner of level relocation.

For ground movement only, the game-world scene may translate an outside-map pointer goal to a reachable point on the prepared town trigger. Require the hero-to-pointer ray to leave the map first through the selected authored edge, at a tangential lane within the published bounds, and to cross the trigger's normal plane in a published lane. An unrelated map border, an in-map goal, other levels, entity interactions, and skill targets retain their existing behavior. Held clicks must continue to target the bridge as the camera follows.

These lane widths and pointer projection are a bounded Dark Magic policy based on owned 1.14d geometry and the user's observed route. The exact original 1.13c/1.14d trigger shape remains unverified.

## Acceptance and evidence

Use the production `game_world` scene and authoritative tick in headless owned-data client tests. Prove the east bridge beyond-lip click and return offline and connected; prove a normal character can approach by ordinary viewport clicks and return offline and connected; prove an unrelated border click does not cross; prove a held bridge click crosses. In the owned resolver tests, check exact four-layout policy bounds, a radius-one route from spawn along every advertised lane to the outer edge, and radius-one return paths from both arrivals. In Lua tests, exercise every east strip lane and the adjacent rejected lanes on authority ticks, plus cardinal pointer rays and trigger ticks for the other three directions. The client tests do not simulate ordinary screen clicks on those three alternate stamps. Run focused tests, full Go and Lua gates, architecture and backend build checks, and CI on the submitted commit. Never commit MPQs, screenshots, saves, or credentials.
