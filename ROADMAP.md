# Dark Magic Restart Roadmap

This roadmap favors complete, testable vertical slices over adding more isolated
services. Diablo II data is never distributed with Dark Magic; users must supply
their own legally obtained game data.

## M0: Recovery and inventory

- [x] Inventory the active branch, working tree, and stashes.
- [x] Preserve all existing Git state without resetting or dropping work.
- [x] Review and name the four historical stashes before any are applied.

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
- [x] Resolve random class-skill function 36 against Skills records.
- [x] Add ladder-season eligibility to special-item selection.
- [x] Connect monster/chest events to deterministic loot seeds.

## M7: Explicit application host

- [x] Add an internal application host with explicit `Start(context.Context)` and
  `Stop(context.Context)` lifecycle contracts.
- [x] Construct native engine dependencies explicitly instead of discovering them
  by scanning a service registry.
- [x] Start dependencies before dependents and stop them in reverse order with
  bounded shutdown contexts.
- [x] Propagate startup, runtime, and shutdown errors to the process entry point.
- [x] Give the renderer/main-thread loop and background workers distinct,
  documented ownership.
- [x] Add lifecycle tests covering order, failure rollback, cancellation, timeout,
  and idempotent shutdown.
- [x] Keep the new host under `internal` until its contracts have survived the
  engine migration.

## M8: Runtime component management

- [x] Register definitions for every available native and scripted component
  without necessarily starting each one.
- [x] Track desired and actual state through `disabled`, `enabling`, `enabled`,
  `disabling`, and `failed` transitions.
- [x] Serialize enable, disable, restart, and reload transitions through one
  reconciler.
- [x] Validate dependencies and reject dependency cycles at registration time.
- [x] Reject disabling a dependency that still has active dependents, with an
  explicit opt-in cascading operation.
- [x] Drive desired state from startup configuration and a runtime management API.
- [x] Publish typed observational lifecycle events without using events for
  dependency injection or startup coordination.
- [x] Add deterministic tests for concurrent requests, partial failures, retries,
  and dependency cascades.

## M9: Layered game-content filesystem

- [x] Define one normalized VFS contract for directories, MPQs, zip archives, and
  embedded content.
- [x] Mount sources in deterministic override order: user mods, the Dark Magic
  shim archive, Diablo II patches/expansion data, then base game data.
- [x] Preserve source provenance so diagnostics can report which layer supplied an
  asset or script.
- [x] Add enumeration, existence, and invalidation behavior suitable for script
  loading and development-time reload.
- [x] Package a redistributable `darkmagic` shim archive containing engine-owned
  scripts and assets, while continuing to require users to supply Diablo II data.
- [x] Load a minimal `boot.lua` from the layered VFS and verify the complete boot
  path headlessly.

## M10: Capability-based Lua runtime

- [x] Replace the mutable global export table with versioned modules such as
  `dm.vfs/v1`, `dm.render/v1`, `dm.input/v1`, and `dm.audio/v1`.
- [x] Expose narrow engine capabilities rather than the application host or
  component registry itself.
- [x] Make one goroutine own each Lua state and route every invocation, callback,
  timer, and reload through its serialized inbox.
- [x] Represent native objects as stable checked handles with type and generation
  validation instead of leaking Go object ownership into Lua.
- [x] Give every script component a resource scope that owns its subscriptions,
  timers, callbacks, render nodes, routes, and native handles.
- [x] Release the entire resource scope automatically on disable, failure, reload,
  or shutdown.
- [x] Keep persistent game/mod state outside disposable Lua states and define a
  versioned serialization boundary.
- [x] Add API conformance tests, stack/error diagnostics, and source-aware Lua
  tracebacks.

## M11: Script-defined components and hot reload

- [x] Let trusted Lua modules declare stable IDs, dependencies, configuration,
  and lifecycle callbacks using the same runtime definition model as Go code.
- [x] Load all available script definitions at startup, then enable only those
  selected by configuration or runtime state.
- [x] Adapt Lua lifecycle callbacks to the host without pretending screens and
  short-lived UI objects are application services.
- [x] Reload transactionally: build and validate a replacement, transfer approved
  persistent state, switch ownership, then tear down the old scope.
- [x] Preserve the working instance when replacement compilation or startup fails.
- [x] Begin with one trusted first-party Lua state and isolated module
  environments; evaluate per-mod states for untrusted third-party content later.
- [x] Add execution budgets or an isolatable runtime boundary before claiming
  support for untrusted scripts.

## M12: Rendering and composition core

- [x] Replace renderer access spread across services with one main-thread-owned
  render capability and a thread-safe command boundary.
- [x] Define retained render nodes with explicit parentage, transform, Z order,
  visibility, clipping, blend mode, and lifetime.
- [x] Make textures, palettes, fonts, animations, and render targets managed
  resources referenced by checked handles.
- [x] Separate asset decoding from GPU upload and make upload/destruction occur on
  the renderer thread.
- [x] Support deterministic world, HUD, modal, cursor, debug, and transition layers.
- [x] Preserve chunking, culling, child-order caching, and allocation-free hot paths
  from the current renderer while replacing its ownership model.
- [x] Add a headless render-command backend for scene tests and golden composition
  fixtures where practical.
- [x] Prove the boundary by rendering one complete Lua-authored screen using only
  versioned capabilities.

## M13: Lua-authored Diablo shell

- [x] Add a scene/navigation manager distinct from long-lived engine components.
- [x] Define scene lifecycle callbacks for `create`, `enter`, `update`, `render`,
  `exit`, and `destroy`.
- [x] Implement the loading screen from the shim archive.
- [x] Implement the title and main game menus from the shim archive.
- [x] Implement character selection/creation as Lua scenes backed by native save
  and record capabilities.
- [x] Implement the interactive game-world scene as Lua orchestration over native
  rendering, input, assets, simulation, and audio.
- [x] Implement inventory, character, skill, automap, options, and pause UI as
  layered overlays rather than independent application services.
- [x] Add deterministic navigation tests for screen transitions, overlay stacking,
  input focus, cancellation, and cleanup.

## M14: Service Mesh retirement

- [x] Migrate one simple native service to explicit construction and host lifecycle
  as a contract test for the new architecture.
- [x] Migrate the renderer, Lua runtime, VFS/assets, input, audio, records, locale,
  configuration, file watching, debug web, and gameplay systems incrementally.
- [x] Replace `ResolveDependencies` polling, readiness flags, and coarse startup
  sleeps with constructors, lifecycle results, or bounded waits at real async
  boundaries.
- [x] Replace `LuaPlugin` discovery with explicit capability registration.
- [x] Update commands, tests, documentation, and the service template as each
  subsystem moves.
- [x] Remove the `servicemesh` dependency only after the main engine and utility
  commands run entirely on the internal host.
- [x] Rename or reorganize `pkg/services` around stable engine capabilities,
  gameplay systems, and internal adapters instead of the old framework shape.

## Architectural acceptance milestone

- [x] Boot the executable through the internal host.
- [x] Mount legally supplied Diablo II archives plus the Dark Magic shim archive.
- [x] Execute `boot.lua`, enter the Lua-authored main menu, select a character, and
  enter an interactive world.
- [x] Open and close inventory as an overlay without disrupting the world scene.
- [x] Enable, disable, restart, and transactionally reload a scripted component at
  runtime without leaking callbacks, renderer resources, or native handles.
- [ ] Shut down cleanly with race detection enabled and without Service Mesh.

## Later gameplay milestones

1. Monster generation and AI.
2. Combat and progression.
3. Audio and music integration.
4. Expanded map generation and developer/mod packaging tools.
5. Save compatibility, multiplayer synchronization, and replayable deterministic
   simulation.

## Performance priorities

- [x] Stop rebuilding and uploading the diagnostic HUD texture every moving frame.
- [x] Avoid redundant pixel conversion and uploads when creating a texture.
- [x] Chunk large DS1 maps into culled textures instead of one full-map GPU texture.
- [x] Cache scene child ordering until topology or Z-index changes.
- [x] Avoid unconditional per-frame callback and input-map allocations.
- [x] Pre-index dynamic loot candidates by item type and level.
- [x] Replace coarse startup sleeps with lifecycle events or bounded waits.

## Parallel codec maintenance

- [x] Audit the sibling codec repositories and record ownership strategy.
- [x] Baseline all eleven modules online and remove local path replacements.
- [x] Add decoder fuzzing, synthetic fixtures, and optional real-asset tests.
- [x] Standardize headless inspection and conversion commands.
- [x] Tag stable codec releases and update Dark Magic dependencies.

See [CODECS.md](CODECS.md) for the module-by-module plan. Codec repositories
remain independent rather than being copied into this engine.

## Current handoff

The repository builds and tests cleanly with portable dependencies. The engine
now runs a real DS1/DT1 scene, polls input on the renderer thread, safely reloads
Lua mods, and exposes records on demand. The current gameplay seam is M6: adapt
supplied TreasureClassEx records to `pkg/loot`, then resolve rolled item codes
and quality. Do not apply or drop historical stashes until their Lua experiments
have been compared with the current implementation.
