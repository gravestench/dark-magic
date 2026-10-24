# Issue #656: Den of Evil cave entry and return

## Intent and boundary

Blood Moor should contain an authored cave entrance that a player deliberately uses to enter the Den of Evil. The Den is a separate indoor level, shown after relocation rather than beside the outdoor terrain. Its exit returns the player to the same Blood Moor session and preserves the existing seamless outdoor joins. This issue completes the map and warp journey only. Den monsters, quest progress, rewards, music, town portals, waypoints, and other cave branches belong to later issues.

The issue is [#656](https://github.com/gravestench/dark-magic/issues/656), under Act I roadmap #354. Work begins from merged outdoor topology #641/#648. The player-facing behavior is the acceptance boundary; no generic replacement of the transition or map systems is proposed.

## Evidence and confidence

Owned Expansion 1.14d `patch_d2.mpq` `Levels.txt` links Blood Moor level 2 `Vis3=8/Warp3=0` to Den of Evil level 8 `Vis0=2/Warp0=4`. Level 8 has maze DRLG type 1, cave level type 3, and `IsInside=1`. `LvlMaze.txt` requests one 24×24 room at all three difficulties. `LvlPrest.txt` has outdoor Cave Entrance definition 51 and cave previous-level definitions 83–86. Base `d2exp.mpq` `LvlWarp.txt` ID 0 is `Act 1 Wilderness to Cave Cliff L`; ID 4 is `Act 1 Cave Up`. These establish authored identities and reciprocal level relationships. They do not establish original placement, selection rectangle transforms, exact entry offset, one-room generation call order, or 1.13c equivalence.

Dark Magic's current cave generator handles multi-room graphs but assigns the Den's single room `previous-level` with connection mask zero; `cave_rooms.preset_definition` currently rejects that mask for special rooms. The production entry world contains levels 1–4 and three outdoor seams. `world_warp` already owns intentional, paired interaction and uses `gameplay.transition.apply` with kind `warp`, but production bootstrap supplies no static cave endpoints and currently requires portal appearance tokens. These are inspected implementation facts.

The chosen layout and safe interaction/arrival offsets are provisional deterministic Dark Magic policy. Record them as such in code, tests, and PR evidence; do not claim original seed parity.

## Approaches considered

1. **Extend the existing paired world-warp path (chosen).** Admit static cave endpoints from the prepared maps and authored records; allow a selectable endpoint without a portal sprite. This reuses operation admission, transition kind, connected projection, and existing portal behavior. It requires a new prepared-topology version and nonconsecutive level handling.
2. **Build a second cave-only transition system.** This would duplicate pairing, interaction admission, and relocation ownership without an evidenced rule difference. It increases the risk that portals and cave warps diverge.
3. **Model the cave as another outdoor seam.** The seam system relocates on approach and presents adjacent terrain. It would show the Den outdoors or cross without intentional use, contradicting the requested warp behavior.

The architecture reviewer selected approach 1 on the user's delegated review authority. Preserve the warp-lab and existing outdoor-seam paths while adding the Den path.

## Prepared map and static cave entrance

Lua remains the map policy owner. A new entry-world recipe version includes the existing four zones and a separate level-8 maze zone. Blood Moor gets one deterministic, reserved 8×8 Cave Entrance definition-51 stamp. The choice must avoid the existing town/Cold Plains route and edge join, leave a player-sized approach from the outdoor route, and survive generation at each difficulty. A named stream chooses placement/variant without advancing earlier map or population streams. Its decoded special-tile facts and the `Levels`/`LvlWarp` pair identify the outward endpoint; image pixels and filenames are not destination logic.

The Den's one-room `LvlMaze` request remains one room. Lua selects one authored previous-level cave chamber (definitions 83–86) and variant from a deterministic eligible set recorded in the recipe, using an explicit orientation policy. Do not insert a fictitious next-level chamber or silently change the room count. Resolve the unique decoded cave-up marker by orientation plus MainIndex/SubIndex and the `Levels`/`LvlWarp` pair; orientation 10/11 alone also represents other semantics. A safe Den arrival must be reachable from that marker and must not immediately reactivate the exit. Go validates the chosen materialized variant and rejects an absent, duplicate, or blocked marker; it does not select a replacement map policy after materialization.

Go materializes the nonconsecutive zone identities in canonical order, then resolves both cave endpoints from the admitted zone facts and typed level/warp records. It checks reciprocal identities, distinct maps, bounded and player-sized points, safe approach/return routes, and unique markers before returning `Prepared`. No partially published map or endpoint may survive a failed validation.

## Authoritative operation and presentation

Production bootstrap publishes the two validated endpoints to `d2legacy.world_warps`. The endpoint is selectable and the existing `d2legacy.world.warp` operation applies the relocation only after ordinary owner, level, range, and target admission. Reject wrong-level, far, stale, and mismatched-pair requests. Entry and exit keep stable IDs across return and same-binary restore. A trigger strip or accidental proximity must never operate the cave warp.

Portal appearances remain available for dynamic portal/warp-lab endpoints. Authored cave endpoints carry no generic TP/PP appearance token; their DS1 tiles provide the visible entrance/exit. The warp snapshot/presentation path must tolerate no appearance component while retaining the selectable target, and must exclude tokenless cave endpoints from portal-only COF rendering. After the authoritative level changes, the client presents the separate Den map and centers the camera on the relocated player; after return, it presents the same Blood Moor map. The outdoor scene continues to compose its adjacent levels before and after this trip.

## Versioned topology and restore

Introduce a new prepared-topology version that serializes zones 1, 2, 3, 4, and 8, the same three outdoor seams, and a reciprocal Den warp specification (or enough immutable authored facts to derive and validate it exactly). The checksum includes level 8 and the cave link. Fresh generation and restore run the same prepublication validation. A malformed level set, missing or duplicate endpoint, mismatched destination, or corrupt warp geometry fails without mutating a live session.

Existing v1/v2/v3 snapshots retain their original zone sets and seam counts; restore must not synthesize the Den into them. Current-binary v4 checkpoint restore retains the cave map, endpoint IDs, active level, player position, and earlier outdoor resident identities. Cross-binary migration from a changed topology is not claimed.

## Verification

Test the recipe and endpoint resolver first with synthetic records/geometry, including three difficulties, one-room preservation, marker absence/duplication, collision rejection, reciprocal links, and old-version snapshot stability. Owned 1.14d tests then exercise the generated stamp, materialized cave and Blood Moor, safe route, deterministic checksum, and all difficulty modes. Production-path offline and connected tests intentionally use entry/return, reject stale/far/wrong-level requests, verify public level/position and scene/camera ownership, reconnect, and restore a same-binary cave checkpoint. Test that the four outdoor maps remain visible across their seams and preserve residents after return. Explicitly verify owned tests skip without MPQs. Native rendered pixels remain unverified if no GPU display is available; do not treat a synthetic scene tree as a pixel capture.

Use Superpowers test-driven development for each bounded implementation task, independent task review, whole-branch correctness/evidence review, and a `refactor-for-readability` style reviewer. Run exact tests and required repository gates on the submitted commit before PR submission.
