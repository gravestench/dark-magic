# Dark Magic Restart Roadmap

This roadmap favors complete, testable vertical slices over adding more isolated
services. Diablo II data is never distributed with Dark Magic; users must supply
their own legally obtained game data.

## M0: Recovery and inventory

- [x] Inventory the active branch, working tree, and stashes.
- [x] Preserve all existing Git state without resetting or dropping work.
- [ ] Review and name the four historical stashes before any are applied.

## M1: Reproducible baseline

- [x] Remove absolute local dependency replacements.
- [x] Make `go test ./...` pass on Go 1.24.
- [x] Add repeatable Make targets and CI.
- [ ] Expand tests for configuration, file loading, records, and service startup.

## M2: Stable engine core

- [x] Serialize new access to the shared Lua state.
- [x] Close the Lua state during graceful shutdown.
- [x] Stop exporting Lua plugins concurrently against an unprotected state.
- [x] Remove the legacy direct `State` API and add lifecycle smoke tests.

## M3: Asset inspection slice

- [x] Add a headless directory/MPQ asset inspector with actionable errors.
- [x] Decode and describe DC6, DCC, DS1, and DT1 asset metadata.
- [x] Export a selected DC6 frame as a headless PNG preview.
- [x] Export a structural isometric DS1 map preview without renderer startup.
- [x] Expose decoded metadata through the asset debug API.
- [x] Describe tabular record files and locale TBL entries.
- [x] Render selected DC6/DCC frames and expose previews through the debug API.

## M4: Lua modding slice

- [x] Validate required manifest fields, API keys, and requirements.
- [x] Validate dependency ordering.
- [x] Load one embedded and one external example mod.
- [x] Materialize embedded example mods without overwriting user files.
- [x] Load and initialize an external example mod end to end.
- [x] Install and initialize an embedded smoke mod end to end.
- [x] Run the embedded terminal UI mod end to end against renderer/UI/tween APIs.
- [x] Demonstrate logging and on-demand record access through the terminal mod.
- [x] Add cache-safe per-table record reload behavior for Lua mods.
- [x] Reload edited mod scripts with optional shutdown and serialized Lua access.
- [x] Demonstrate renderer, tween, and serialized input APIs through the terminal mod.
- [x] Serialize Lua tween callbacks and implement start/update/repeat/complete bindings.

## M5: Interactive scene

- [x] Make tween updates non-blocking, race-safe, time-correct, and stoppable.
- [x] Load a deterministic DS1 map stamp from a configured MPQ.
- [x] Provide a standalone interactive grid scene with a controllable placeholder hero.
- [x] Load a real DS1 stamp from an MPQ into the interactive scene.
- [x] Compose matching DT1 floor/wall graphics with an act-specific PL2 palette.
- [x] Add deterministic headless hero movement, camera tracking, bounds, and save/load.
- [x] Render a configured DS1/DT1 map in the engine scene with a safe fallback.
- [x] Add a camera-anchored scene HUD with locale and map status.
- [ ] Replace diagnostic HUD strings with localized table keys and game UI art.
- [x] Connect deterministic scene movement and camera tracking to Raylib input/rendering.

## M6: Deterministic item and loot foundation

- [x] Extract renderer-independent treasure-class selection.
- [x] Preserve positive-pick, NoDrop, negative-pick, and nested-class behavior.
- [x] Make rolls reproducible from an explicit seed.
- [x] Detect malformed classes, recursive cycles, and runaway nesting.
- [x] Add a headless loot-roll test application with redistributable example data.
- [x] Parse supplied TreasureClass/TreasureClassEx records into the loot catalog.
- [x] Resolve direct terminal codes into base item records and retain quality modifiers.
- [x] Expand generic and dynamic item-type codes such as `weap3` and `armo33` by level.
- [x] Parse and select Classic/Expansion ItemRatio rule sets.
- [x] Apply level, magic-find, minimum, and TreasureClass modifiers to item quality.
- [x] Apply unique/set availability fallback and item-type quality restrictions.
- [x] Select concrete UniqueItems/SetItems records by eligibility and rarity weight.
- [x] Parse and deterministically select eligible magic/rare prefixes and suffixes.
- [x] Roll affix property values and materialize portable item instances.
- [x] Parse Properties/ItemStatCost and interpret common property functions.
- [x] Interpret damage, proc, skill, state, charged-skill, and boolean functions.
- [ ] Resolve random class-skill function 36 against Skills records.
- [ ] Add ladder-season eligibility to special-item selection.
- [ ] Connect monster/chest events to deterministic loot seeds.

## Later gameplay milestones

1. Monster generation and AI.
2. Combat and progression.
3. Audio and music integration.
4. Expanded map generation and developer/mod packaging tools.

## Performance priorities

- [x] Stop rebuilding and uploading the diagnostic HUD texture every moving frame.
- [x] Avoid redundant pixel conversion and uploads when creating a texture.
- [ ] Chunk large DS1 maps into culled textures instead of one full-map GPU texture.
- [x] Cache scene child ordering until topology or Z-index changes.
- [ ] Avoid unconditional per-frame callback and input-map allocations.
- [ ] Pre-index dynamic loot candidates by item type and level.
- [ ] Replace coarse startup sleeps with lifecycle events or bounded waits.

## Parallel codec maintenance

- [x] Audit the sibling codec repositories and record ownership strategy.
- [ ] Baseline all eleven modules online and remove local path replacements.
- [ ] Add decoder fuzzing, synthetic fixtures, and optional real-asset tests.
- [ ] Standardize headless inspection and conversion commands.
- [ ] Tag stable codec releases and update Dark Magic dependencies.

See [CODECS.md](CODECS.md) for the module-by-module plan. Codec repositories
remain independent rather than being copied into this engine.

## Current handoff

The repository builds and tests cleanly with portable dependencies. The engine
now runs a real DS1/DT1 scene, polls input on the renderer thread, safely reloads
Lua mods, and exposes records on demand. The current gameplay seam is M6: adapt
supplied TreasureClassEx records to `pkg/loot`, then resolve rolled item codes
and quality. Do not apply or drop historical stashes until their Lua experiments
have been compared with the current implementation.
