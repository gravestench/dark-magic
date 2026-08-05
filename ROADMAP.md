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
- [ ] Load one embedded and one external example mod.
- [x] Materialize embedded example mods without overwriting user files.
- [x] Load and initialize an external example mod end to end.
- [x] Install and initialize an embedded smoke mod end to end.
- [x] Run the embedded terminal UI mod end to end against renderer/UI/tween APIs.
- [ ] Demonstrate logging, record access, and reload behavior.
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

## Later gameplay milestones

1. Item and loot generation.
2. Monster generation and AI.
3. Combat and progression.
4. Audio and music integration.
5. Expanded map generation and developer/mod packaging tools.

## Parallel codec maintenance

- [x] Audit the sibling codec repositories and record ownership strategy.
- [ ] Baseline all eleven modules online and remove local path replacements.
- [ ] Add decoder fuzzing, synthetic fixtures, and optional real-asset tests.
- [ ] Standardize headless inspection and conversion commands.
- [ ] Tag stable codec releases and update Dark Magic dependencies.

See [CODECS.md](CODECS.md) for the module-by-module plan. Codec repositories
remain independent rather than being copied into this engine.

## Current handoff

The repository builds and tests cleanly with portable dependencies. Continue by
binding `pkg/scene.State` to the existing renderer service, replacing the demo
grid with DS1/DT1 output, and providing the terminal mod's Lua callbacks through
the serialized Lua-state API. Do not apply or drop historical stashes until
their Lua experiments have been compared with the current implementation.
