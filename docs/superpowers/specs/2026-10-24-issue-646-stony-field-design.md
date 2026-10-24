# Issue #646: Deterministic Stony Field Zone Design

## Intent and evidence

The Act I production journey currently reaches Cold Plains, but no Stony Field zone can be generated. Issue #646 is a child of #354 and establishes one admitted, materializable level-4 recipe before a separate issue places the Cold Plains exit and installs a live seam. It does not make Stony Field playable yet.

Recovered [1.13 level data](https://github.com/fabd/diablo2/blob/master/code/d2_113_data/Levels.txt) and locally inspected owned Expansion 1.14d `patch_d2.mpq` records agree that Stony Field is Act I level 4, outdoor DRLG type 3, level type 2, and 80×80 tiles in Normal, Nightmare, and Hell. The owned 1.14d `LvlTypes` row 2 names Act I wilderness DT1 libraries. Its `LvlPrest` definitions 29 and 30 are generic 8×8 Act I stone fill pieces with three DS1 variants each. They are eligible authored assets for a first structural recipe. Choosing only these generic pieces, their arrangement, and the edge entry coordinate are **Dark Magic policy**, not verified 1.13c/1.14d layout or original seed call order. The owned table also lists Cairn Stones and Tower Tome presets, but their original placement is not established here; #646 does not emit those set pieces or quest interactions.

The [Act I route](../../research/QUESTS_AND_WORLD.md) records Cold Plains→Stony Field as ordinary physical traversal and Stony Field→Tristram via Cairn Stones as a portal. The [warp research](../../research/WARPS_PORTALS_WAYPOINTS.md) leaves links beyond level 3 open. A [1.14d clean-room implementation](https://github.com/jaenster/libd2/blob/main/docs/usage/node.md) exposes generated border exits and distinguishes them from placed warps; its [verification](https://github.com/jaenster/libd2/blob/main/docs/VERIFICATION.md) reports original-game map comparisons. This is a research comparison source, not evidence that the provisional Dark Magic layout matches original Stony Field.

## Approaches

1. **Chosen: dedicated `d2legacy.mapgen.stony_field` headless strategy.** It validates owned records, accepts the future Cold Plains exit direction explicitly, and emits a deterministic `worldgen.Zone` with one opposite-facing entry and a route into the interior. It follows the existing Lua policy/Go transport boundary and keeps the current live three-zone world untouched. Independent architecture and evidence reviewers approved this scoped approach.
2. Recreate or import the full 1.14d outdoor generator. That would require substantially broader source analysis, seed call-order fidelity, all Act I outdoor special pieces, and whole-act comparison. It cannot be justified by this one issue's checks.
3. Generate Stony Field and install its live Cold Plains seam in one change. That would require a placed Cold Plains exit, durable topology compatibility, authoritative crossing, camera continuity, and connected-client proof. Those are a separate independently closable issue after this recipe exists.

## Behavior and ownership

`Runtime.Generate(ctx, "stony_field", 4, seed, coldPlainsExitDirection, difficulty)` remains the public headless entry. The Go runtime transports plain values; a new Lua strategy owns all Stony Field-specific facts. The explicit source direction is a future seam input, not an invented existing Cold Plains endpoint. Synthetic callers may supply each cardinal direction. A later live issue must obtain it from one admitted, correctly targeted Cold Plains exit rather than hard-code it.

Before publishing a zone, reject a wrong level ID, difficulty outside 0–2, non-cardinal or absent direction, mismatched Act/DRLG/level type, absent or non-grid difficulty dimensions, absent or malformed fill candidates, declared DS1 variants with invalid references, and missing/invalid mask-selected DT1 references. Check **every** selectable fill candidate before any seed draw. Asset identity checks occur in headless generation; actual DS1/DT1 loading and collision are established only by owned-data materialization.

Use the validated 8×8 stone-fill candidates from definitions 29 and 30 with Stony Field-specific named deterministic streams for fill and variant choices. Emit canonically ordered stamps, rooms, and orthogonal links in bounds derived from `Levels.txt`. Emit exactly one warp with role `cold-plains-entry`, destination level 3, on the edge opposite the source direction. A semantic dirt path connects that entry to a distinct interior point. The trace should name the inspected 1.14d evidence behind the candidate policy, without asserting that any particular runtime record source is 1.14d. This grid and path are provisional Dark Magic structural policy. They must not create a portal, waypoint, Underground Passage exit, Cairn Stones, Tower Tome, monsters, or quest state.

Do not change the current `entry_world` recipes, durable snapshots, active 1↔2 or 2↔3 seams, renderer, movement authority, or monster population. No new generic Go Diablo policy is needed. Reuse existing record, asset-path, deterministic, and dirt-path helpers where appropriate; avoid a broad refactor of the live Cold Plains strategy to introduce this sibling.

## Verification

- Synthetic headless tests prove all three difficulties and four source directions produce bounds, connected room graph, one correctly facing/targeted edge entry, and a semantic path to the interior. Same inputs repeat the same checksum; changing only direction changes it; `worldgen.NewZone(zone.Definition())` retains the checksum.
- Admission-failure tests cover wrong level, bad or absent direction, invalid difficulty, wrong Act/DRLG/type, malformed dimensions in Normal and at least one other difficulty, missing candidates, invalid declared DS1 variant, and missing mask-selected DT1 path. A one-cell synthetic fixture can prove that a malformed candidate fails even when its selected seed chooses the other candidate.
- With owned 1.14d MPQs, materialize one generated Stony Field zone. Require nonempty tiles, an in-bounds unblocked arrival within a fixed small distance of the semantic entry, and a collision path to a distinctly interior point. Data-dependent tests skip clearly in CI and the PR states whether they ran locally. No owned game data, extracted tables, captures, or private saves enter Git.
- Run focused generator and world materialization tests, architecture, Lua quality and hardening, full tests, race, lint, and both client backends on the submitted commit. Review correctness, evidence, and formatting/readability; the style reviewer uses `refactor-for-readability`.

## Limits and follow-up

This slice proves Dark Magic can generate and materialize a connected Stony Field using authored assets. It does not prove exact original seed/layout parity, placed special-preset positions, or a visible/traversable live Cold Plains seam. Open a focused child of #354 for the Cold Plains exit, reciprocal Stony Field seam, snapshot compatibility, connected-client movement, and continuous neighboring terrain after #646 lands. Keep the Cairn Stones portal and quest set pieces with their own evidence and implementation issues.
