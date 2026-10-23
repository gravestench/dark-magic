# Issue #618: Cold Plains Zone Generation Design

## Intent and evidence

Issue #618 is a child of the G14 Act I campaign gate. The generated Blood Moor already contains a next-level exit naming level 3, but Dark Magic cannot yet generate a Cold Plains zone. This slice makes level 3 an admitted, reproducible map recipe that can be materialized and inspected before a later issue installs it in a live session.

Recovered 1.13 level data identifies Cold Plains as Act I level 3, outdoor DRLG type 3 and level type 2, with an 80×80 tile Normal map and a waypoint flag. A 1.14d clean-room generation example reports connections to Blood Moor, Stony Field, Burial Grounds, and Cave Level 1. These facts establish identity and future topology, not the exact level-3 room layout, physical exit coordinates, or original seed call order. The first structural layout and entry placement below are explicitly Dark Magic policy.

## Approaches

- **Chosen: a dedicated `d2legacy.mapgen.cold_plains` Lua strategy.** It receives a Blood Moor `next-level-exit` direction as an explicit input, interprets recovered records, and emits one immutable `worldgen.Zone` through the existing headless runtime. The caller extracts that direction from an admitted Blood Moor zone and verifies the warp targets level 3; Lua validates the cardinal string but cannot authenticate its source. This keeps Diablo-specific policy in the mod and avoids changing the production entry world.
- Extend the current `outdoor.lua` for both levels. That module currently assumes Blood Moor-specific fill roles, river/cliff structures, and one opposite next-level exit. Sharing its whole strategy would silently turn those assumptions into Cold Plains rules or require a broad refactor.
- Put level-3 policy in Go. The generic map engine and adapter intentionally own transport, validation, and materialization, while the d2legacy mod owns level policy. A Go strategy would cross that boundary.

Two independent repository/evidence reviews selected the dedicated strategy.

## Behavior and boundaries

`Runtime.Generate(ctx, "cold_plains", levelID, seed, bloodMoorExitDirection, difficulty)` remains the public headless entry. The Lua strategy accepts only level 3 and cardinal source directions; it rejects absent or mismatched recovered level/type/preset facts before publishing. The caller or integration test obtains the direction from a `next-level-exit` warp targeting level 3 on an admitted Blood Moor `Zone`; direct synthetic tests may supply each cardinal direction. It derives bounds from the difficulty-specific `Levels` size fields, requires an 8-tile grid, and uses normalized, nonempty DS1/DT1 references from level type 2 and LvlPrest definitions 29 and 30. Those rows are 8×8 `LevelId=0` Act I outdoor stone-fill pieces in recovered 1.13 data, but selecting only these two for Cold Plains is provisional Dark Magic policy, not an original placement claim. The implementation plan fixes named deterministic streams and tests; the policy must not call Blood Moor river/cliff or its one-opposite-exit route module.

The zone contains canonically ordered 8×8 room and stamp facts, a connected structural room graph, and one semantic Blood Moor entry warp at an edge opposite the supplied Blood Moor exit direction. A deterministic tile route connects that entry to an interior point. The entry coordinate and route shape are Dark Magic structural policy, not a claim about original Cold Plains generation. No usable exits to Stony Field, Burial Grounds, or Cave Level 1 are emitted until their placement is separately researched. If useful, trace may name those known future connections without turning them into active endpoints.

The same explicit inputs produce the same zone checksum. A changed source direction changes the facing edge and checksum. The admitted `Definition` is the durable representation; `worldgen.NewZone` can restore it with the same checksum without rerunning Lua. Asset paths are checked as references during generation; actual DS1/DT1 loading and collision are tested at materialization, which requires owned MPQ files. No game data or captures enter the repository.

## Acceptance and verification

Synthetic-record tests call the production headless strategy. They cover Normal, Nightmare, and Hell dimensions; each cardinal source exit and opposite-facing Cold Plains entry; stable output/checksum; definition restore; connected room/route facts; and clear rejection of wrong level, bad direction, invalid dimensions, missing preset candidates, and empty asset references. The owned-asset test derives the source exit from an admitted Blood Moor zone, generates Cold Plains with installed game data, materializes the zone, and checks map dimensions, nonempty tiles, an open entry arrival subtile within a small fixed distance of the authored edge warp, and a collision-walkable path from that arrival to a distinctly interior point. An in-bounds but blocked, far-away, or isolated entry fails the test. It skips explicitly when the MPQ fixture is unavailable, and the PR reports whether it ran locally.

This issue does not change `EntryWorld` snapshot v1, session maps, player admission, population, waypoint state, existing town↔Blood Moor seam, or transition authority. A later #353/#354 child will install level 3 in the production session, resolve and test the live Blood Moor↔Cold Plains crossing, then add researched outward exits and waypoint use in further focused slices.

## Implementation constraints

Go remains a typed transport and generic materialization boundary. Lua owns all level-specific selection and source-direction interpretation. Every choice without original 1.13c/1.14d placement evidence must be called Dark Magic policy in code comments, tests, and PR. No edits outside #618's acceptance boundary, no dependency changes, and no committed game data.

## Sources

- [Recovered 1.13 `Levels.txt`](https://raw.githubusercontent.com/fabd/diablo2/master/code/d2_113_data/Levels.txt) identifies level 3 and its level/waypoint fields; it does not directly enumerate all placed outdoor exits.
- [Recovered 1.13 `LvlPrest.txt`](https://github.com/fabd/diablo2/blob/master/code/d2_113_data/LvlPrest.txt) identifies definitions 29/30 as generic Act I stone-fill records.
- [libd2 1.14d example](https://github.com/jaenster/libd2) reports Cold Plains' four connections.
- [`docs/MAP_GENERATION.md`](../../MAP_GENERATION.md) and [`docs/research/WARPS_PORTALS_WAYPOINTS.md`](../../research/WARPS_PORTALS_WAYPOINTS.md) describe Dark Magic's policy/mechanism boundary and transition ownership.
