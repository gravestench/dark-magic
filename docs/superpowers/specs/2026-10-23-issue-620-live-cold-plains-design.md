# Issue #620: Live Blood Moor–Cold Plains Crossing Design

## Intent and evidence

Issue #620 makes the already generated Cold Plains recipe from #618 reachable through ordinary movement from Blood Moor, and lets a player return. The current production builder publishes two maps, two zones, trusted spawns, and one town↔Blood Moor seam as one unit. The Lua transition system already accepts a list of directed seams, but Go's prepared world and client application still hold only one pair. The result must work in offline and connected sessions, survive checkpoint restoration, and remain deterministic.

Recovered 1.13 data identifies Cold Plains as Act I level 3. The #618 recipe and owned-asset test establish one provisional Blood Moor-facing entry and a route inward. Exact 1.13c/1.14d exit coordinates, trigger radius, inset distance, and seed call order are unverified. New placement and seam geometry in this issue are Dark Magic policy until supported by original-game observations or a pinned generator trace.

## Approaches

- **Chosen: eager three-zone preparation.** Extend the current entry-world owner to generate and materialize town, Blood Moor, and Cold Plains, then resolve both seam pairs before returning a prepared result. This reuses the authority publication barrier, keeps failure before admission, and gives replay one topology identity.
- Generate Cold Plains only when a player first crosses. That introduces loading states, retry/cancellation, and checkpointing of a pending transition inside the tick loop; it cannot meet this issue's pre-publication validation rule without broader changes.
- Keep Cold Plains in a separate session manager. That duplicates map, seam, collision, and restoration ownership and risks a client seeing different topology from the authority.

Two independent code/evidence reviewers approved eager preparation. The user delegated these design reviews to subagents.

## Topology and generation

The in-memory entry-world topology owns zones by level ID and a canonically ordered list of seam specifications. New generation uses the existing map-generation seed and selected difficulty for all zones; the current client composition root supplies an explicit seed of 1, and this issue does not change seed ownership. Generate Blood Moor first, then inspect its admitted, serialized zone facts in d2legacy Lua for exactly one `next-level-exit` with destination level 3. Reject missing, duplicate, wrong-target, or non-cardinal exit facts before calling the #618 Cold Plains strategy. Pass that exit's direction explicitly to Cold Plains; do not infer it from coordinates or hard-code it in Go.

The town↔Blood Moor seam retains its authored-town/generated-Moor resolution. The Blood Moor↔Cold Plains pair uses both admitted generated warp coordinates. A dedicated generated-to-generated resolver checks reciprocal facing and destination IDs, converts tile centers to collision subtiles, requires walkable arrival points within a bounded distance, confirms each arrival can reach an interior point, and keeps arrivals outside the return trigger. D2legacy Lua seam policy supplies the trigger radius, search bound, and inset/clearance distances as explicit facts. Go consumes the geometry facts for collision resolution, and `world_transition.lua` consumes the same per-seam trigger radius instead of its current fixed local constant. It returns two verified endpoints. Resolve both pairs before publishing maps, zones, spawns, or Lua initial data; failure leaves the previous world untouched.

One prepared result owns all maps, zones, spawns, and seam pairs. Lua policy describes pair identities and directions; Go resolves materialized collision coordinates. Stable list order is town↔Blood Moor followed by Blood Moor↔Cold Plains, with each pair emitting its two directed entries. Duplicate source/destination edges and mismatched level IDs fail before publication. The existing Lua world-transition system receives four directed seams and continues to call the shared transition transaction. It remains the sole authority for changing entity level, position, bounds, velocity, and motion state.

## Durable state and restoration

New entry-world snapshots use a version-2 envelope with admitted definitions indexed in canonical level order and both seam specifications in canonical pair order. The decoder validates version, unique level IDs, reciprocal pair identities, referenced levels, direction and coordinate facts, and matching seed/difficulty before materialization. Version-1 two-zone snapshots still decode to the same in-memory topology shape with one pair and retain their original two-zone representation on re-serialization; they do not acquire Cold Plains implicitly.

The simulation checkpoint contains dynamic ECS and Lua participant state, while entry-world topology is a separate durable envelope. The current offline/listen composition owner in `internal/app/clientapp` captures them as one recovery value: the entry-world snapshot, its canonical checksum, and `Session.RecoveryCheckpoint`. Topology is immutable after publication, so it can be captured before the dynamic checkpoint without changing tick order. Restoration validates and materializes the saved topology first, installs exactly its directed seams (two for v1, four for v2), and puts the canonical topology checksum into `sessionInitialData` before composing the runtime. That configuration is hashed into the existing d2legacy runtime identity participant. The recovery composition preflights the checkpoint participant's identity against the freshly composed identity before calling `RestoreRecoveryCheckpoint`; this rejects a different valid topology before saved ECS state is applied or a tick advances. The test also changes only the topology in a recovery pair to prove this rejection. A valid integration test checkpoints while the player is in Cold Plains, reconstructs both topology and session, returns to Blood Moor with the same command sequence, and compares the complete authoritative checkpoint checksum, participants, player identity, location, and motion state with an uninterrupted continuation. A topology checksum round-trip by itself is insufficient.

## Presentation and connected play

Client preparation and bootstrap use the same prepared topology as the authority; the current client application stores one seam and must store the canonical pair list. Existing town and Blood Moor spawns remain as before, and Cold Plains gains a trusted arrival spawn. Both peers include the canonical topology checksum in `sessionInitialData`, so the existing runtime-identity handshake rejects a client whose generated map differs from the authority before connection. On committed player-location changes, the application switches active map and navigation/collision to that level. A connected client must observe the authority's level and position after crossing and render and navigate Cold Plains from the corresponding map. Its local projection cannot authorize a crossing.

The authored warp-lab fixture remains tied to the town↔Blood Moor pair. Cold Plains does not gain monster population, waypoint activation, or outward exits in this issue.

## Acceptance and verification

- Synthetic records: three difficulties and all four reciprocal orientations; unique source-exit extraction; malformed/missing/duplicate/wrong-target exits; mismatched or duplicate seam facts; v1 restore compatibility and v2 canonical replay/checksum.
- Owned MPQ: all three zones materialize; both outdoor endpoints have bounded, collision-valid arrivals and paths inward; both seams resolve before publication; no game-data bytes enter the repository. Data-dependent tests skip clearly in CI and run locally when the owner provides assets.
- Production path: ordinary movement crosses town→Blood Moor→Cold Plains→Blood Moor→town; no immediate bounce, identity loss, stale navigation, or partial publication. Offline and connected client observation use the same authority topology.
- Recovery: checkpoint in Cold Plains, rebuild topology, restore session, and compare a return crossing's complete authoritative state with uninterrupted execution. A different valid topology is rejected before session restoration. Failure cases reject incomplete/blocked destinations.
- Report exact commands and results on the submitted commit, including owned-data status. Measure one eager preparation with owned assets to catch a material startup regression, without adding a cache or streaming framework.

## Scope and open evidence

This issue updates only the entry-world/mapgen/transition and client composition paths required for level-3 travel, their tests, and relevant map/warp documentation. It does not change generic movement rules, population, quests, waypoint behavior, other exits, or dependency versions. Physical seam geometry and original trigger behavior remain unverified Dark Magic policy; document any later original-game evidence before changing them.

## Sources

- [Issue #620](https://github.com/gravestench/dark-magic/issues/620) and its dependency [#618](https://github.com/gravestench/dark-magic/issues/618).
- [Recovered 1.13 Levels.txt](https://raw.githubusercontent.com/fabd/diablo2/master/code/d2_113_data/Levels.txt), level 3 `Id`, `Act`, `DrlgType`, and `LevelType`.
- [Map-generation architecture](../../MAP_GENERATION.md) and [warp research](../../research/WARPS_PORTALS_WAYPOINTS.md).
