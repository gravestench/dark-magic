# Dark Magic Restart Roadmap

This roadmap favors complete, testable vertical slices over adding more isolated
services. Diablo II data is never distributed with Dark Magic; users must supply
their own legally obtained game data.

## M0: Recovery and inventory

- [x] Inventory the active branch, working tree, and stashes.
- [x] Preserve all existing Git state without resetting or dropping work.
- [x] Review and name the four historical stashes before any are applied.

## M1: Reproducible baseline

- [x] Remove absolute filesystem dependency replacements.
- [x] Make `go test ./...` pass on Go 1.24.
- [x] Add repeatable Make targets and CI.
- [x] Expand tests for configuration, file loading, records, and service startup.

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
- [x] Replace diagnostic HUD strings with localized table keys and game UI art.
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
- [x] Shut down cleanly with race detection enabled and without Service Mesh.

## Later gameplay milestones

The completed milestones above establish architecture and headless acceptance;
they do not mean the corresponding screens or gameplay are faithful Diablo II
implementations. The remaining work is tracked explicitly below.

## M15: Verified Blizzard-asset knowledge

- [x] Consolidate community research for MPQ paths, palettes,
  frames, coordinates, sounds, localization keys, and screen composition.
- [x] Add a read-only catalog command with provenance, hashes, frame metadata,
  and palette-applied DC6 contact sheets.
- [x] Verify the initial 90 hypotheses against a complete English Diablo II/LOD
  MPQ stack without committing derived game imagery.
- [ ] Move verified screen facts from Go literals into versioned shim manifests
  with confidence, game-version, language, and resolution fields.
- [ ] Catalog every front-end screen, HUD/panel sheet, cursor, font, cinematic,
  sound cue, and relevant TXT/TBL dependency.
- [ ] Add comparison fixtures that verify expected dimensions, frame counts,
  offsets, and hashes without distributing original asset bytes.

### M15 execution checkpoints

- [x] M15.1: Define a versioned presentation-manifest schema with provenance,
  confidence, game version, language, resolution, palette, layout, and timing.
- [x] M15.2: Load manifests through a read-only versioned Lua capability and
  reject malformed or incompatible documents with actionable errors.
- [ ] M15.3: Migrate front-end, character, HUD, panel, cursor, font, audio, and
  cinematic facts from compiled Go declarations into shim-owned manifests.
- [ ] M15.4: Generate and validate hash/dimension/frame fixtures from manifests
  without storing proprietary decoded pixels.

## M16: MPQ-backed presentation primitives

- [x] Decode DC6 assets with an explicit palette through managed render
  resources exposed to Lua by checked handles.
- [x] Decode DCC assets with an explicit palette through the same managed path.
- [x] Support DC6 frame/direction selection, placement offsets, anchoring,
  front-end tiling, deterministic layer order, and alpha/additive/multiply blend
  modes.
- [ ] Support native animation timing, loop modes, clipping, general tiling,
  and nine-slice composition.
- [ ] Decode Diablo font TBL/DC6 pairs and render localized, colored, aligned,
  wrapped text.
- [ ] Add Lua-authored buttons, text fields, checkboxes, scrollbars, focus order,
  controller navigation, hit testing, hover/pressed/disabled states, and cursor
  sounds.
- [ ] Add music/SFX/UI/ambience/speech buses, looping and streaming, sound-record
  lookup, grouped variants, fades, pan, and scope-owned playback.
- [ ] Cache decoded CPU assets separately from renderer/audio resources and
  invalidate them by VFS generation without leaking owner-thread resources.

### M16 execution checkpoints

- [x] M16.1: Make animation a renderer-owned resource with per-frame durations,
  loop/once/ping-pong modes, pause/resume/seek, and deterministic clocks.
- [ ] M16.2: Add palette-aware DCC decoding and COF-driven component direction,
  mode, weapon-class, layer, transform, shadow, and event composition.
  Palette-aware DCC frame/animation handles and checked COF metadata for layer
  priority, weapon class, draw effects, transparency, selection, speed, and
  frame events are implemented; assembled multi-component rendering remains.
- [x] M16.3: Decode Diablo font TBL/DC6 pairs and expose measured localized text
  layout with color, alignment, wrapping, clipping, and fallback behavior.
- [x] M16.4: Implement reusable Lua controls with pointer/controller focus,
  hit-testing, state visuals, activation sounds, and accessibility metadata.
  Buttons, UTF-8 text fields, checkboxes, and scrollbars provide pointer
  hit-testing, scoped directional focus, disabled/hidden states, activation,
  change/sound and visual-state callbacks, and accessibility snapshots.
- [ ] M16.5: Implement scoped music, UI, SFX, ambience, and speech buses with
  streaming, looping, fades, pan, grouping, and record-driven lookup.
  Generation-checked sounds now route through independently adjustable named
  buses with per-sound volume/pan, owner-thread looping, deterministic fades,
  and grouped stopping; streaming and record-driven lookup remain.
- [ ] M16.6: Add generation-aware CPU/native caches, budgets, diagnostics, and
  leak/race/rapid-scene-transition acceptance tests.

## M17: Authentic Lua-authored front end

- [ ] Implement startup/trademark/cinematic sequencing from verified BIK and UI
  assets with skip and failure behavior.
- [x] Replace the placeholder title and main-menu rectangles with verified,
  palette-aware tiled backgrounds and the anchored layered animated logo.
- [ ] Add localized front-end buttons, legal/version labels, cursor behavior,
  and title audio.
- [ ] Implement single-player, multiplayer/TCP-IP, credits, cinematics, dialogs,
  and navigation using verified front-end assets.
- [ ] Implement character creation with all seven class animation sets, hit
  regions, narration, name validation, expansion/hardcore flags, and cancel.
- [ ] Implement saved-character selection with paging, scrollbar, composite
  previews, metadata labels, deletion confirmation, and double activation.
- [ ] Implement progressive loading screens driven by real dependency progress.
- [ ] Add screenshot/composition tests for every state without checking Blizzard
  imagery into the repository.

### M17 execution checkpoints

- [ ] M17.1: Complete title/main-menu labels, controls, music, cursor, legal and
  version text, dialogs, and single/multiplayer/credits/cinematics navigation.
  The localized Single Player control now uses the verified split DC6 button,
  Exocet TBL/DC6 text, pointer/focus activation, and selection sound; remaining
  controls and sibling screens remain.
- [ ] M17.2: Implement startup/trademark/cinematic sequencing and skip/failure
  behavior.
- [ ] M17.3: Implement all seven character-creation presentations and rules.
- [ ] M17.4: Implement saved-character presentation, paging, deletion, composite
  previews, and activation.
- [ ] M17.5: Implement dependency-driven loading and front-end composition tests
  across supported resolution, language, and game-version variants.

## M18: Authentic in-game shell

- [ ] Replace the compatibility HUD and placeholder hero with Lua orchestration
  over real world and character presentation handles.
- [ ] Assemble the 640x480 and 800x600 control panels, globes, stamina and
  experience bars, belt, skill buttons, minipanel, tooltips, and cursor states.
- [ ] Implement interactive inventory/equipment, character stats, skill trees,
  automap, quest log, waypoint, party, hireling, vendor, stash, cube, options,
  help, chat, pause, and escape panels from verified assets and records.
- [ ] Make panel geometry data-driven from Inventory.txt, SkillDesc.txt, and
  related records; keep only presentation corrections in shim manifests.
- [ ] Implement correct focus/input routing when world, HUD, modal, cursor,
  transition, and debug layers coexist.
- [ ] Add an in-game Lua console with command history, multiline editing,
  source-aware errors, structured value inspection, and log output.
- [ ] Provide capability-aware autocompletion for Lua keywords, console-local
  names, permitted modules, and table/userdata members; Tab and Shift-Tab cycle
  candidates, a shared prefix is inserted first, and candidate browsing never
  executes code or invokes metamethods with side effects.
- [ ] Bind each console tab to an explicit Lua runtime and dedicated disposable
  resource scope; serialize evaluation through the runtime owner and clearly
  display the selected runtime, scope, and capability set.
- [ ] Gate mutation/debug capabilities by session policy: full local developer
  access when enabled, read-only inspection where appropriate, and no path from
  a realm-connected client to server-authoritative state or server-only APIs.
- [ ] Tear down console-created handles, callbacks, subscriptions, and render
  nodes on reset, runtime reload, disconnect, scene change when configured, and
  client shutdown.

## M19: Character, item, and save fidelity

- [ ] Implement COF/DCC composite animation with equipment components, weapon
  classes, directions, modes, layer order, transforms, shadows, and events.
- [ ] Replace placeholder character storage with versioned Diablo II save parsing,
  validation, creation, migration, and non-destructive writing.
- [ ] Implement authoritative inventory grids, body slots, belt, cursor item,
  alternate weapons, stash/cube/vendor transfers, and placement validation.
- [ ] Connect deterministic loot instances to item serialization, names, colors,
  requirements, durability, sockets, runewords, set state, and tooltip stats.
- [ ] Recalculate character attributes and presentation after every equipment,
  skill, level, state, and difficulty change.

## M20: World fidelity

- [ ] Import DS1/DT1 collision, orientation, material, warp, object, shadow, roof,
  and subtile flags into deterministic chunks and zones.
- [ ] Implement entity-size-aware A* navigation, path smoothing, collision,
  interaction range, line of sight, and movement modes.
- [ ] Implement deterministic preset, maze, outdoor, substitution, warp, object,
  monster, and waypoint generation for all acts and difficulties.
- [ ] Add animated map tiles, doors, breakables, shrines, chests, portals,
  missiles, overlays, selectable bounds, and automap discovery.
- [ ] Stream and cull zones without changing deterministic simulation results.

## M21: Diablo simulation

- [ ] Implement the complete stat/value system, ItemStatCost encoding, derived
  stats, caps, per-level values, state modifiers, and description functions.
- [ ] Implement player and monster modes, animation events, targeting, AI,
  spawning, packs, bosses, pets, hirelings, corpses, and death/respawn.
- [ ] Implement attacks, hit checks, block, defense, damage types, resistance,
  absorb, leech, regeneration, durability, experience, leveling, and difficulty.
- [ ] Implement skills, skill trees, missiles, auras, states, procs, charges,
  cooldowns, mana, targeting shapes, and server-authoritative validation.
- [ ] Implement NPC interaction, vendors, gambling, repair, identify, cube,
  shrines, quests, waypoints, acts, difficulties, and end-game transitions.
- [ ] Implement record-driven music, ambience, speech, monster, object, item, UI,
  and combat sound selection with deterministic event identity where required.

## M22: Client and game-session networking

- [ ] Define deterministic simulation snapshots, commands, event logs, RNG
  streams, checksums, replay files, and desync diagnostics.
- [ ] Add a first-class `cmd` client for offline single-player, self-hosted
  listen/dedicated multiplayer, and realm-connected games.
- [ ] Run one authoritative deterministic game-session implementation in-process,
  as a listen server, as a standalone self-hosted server, or under a realm.
- [ ] Implement discovery/direct connect, authentication boundaries, snapshot
  transfer, command replication, rollback or correction, and reconnect.
- [ ] Preserve offline characters safely and separate trusted server characters
  from client-controlled saves.
- [ ] Add long-running soak, malformed-data, fuzz, latency/loss, save round-trip,
  performance, and race tests across supported platforms.

## M23: Realm, mod delivery, and trusted persistence

- [ ] Add a first-class `cmd` realm providing accounts, authentication, game
  creation/discovery, session assignment, and operational administration.
- [ ] Define signed/versioned mod manifests containing dependency order, engine
  and capability compatibility, payload identities, hashes, sizes, and trust
  policy.
- [ ] Let clients negotiate, fetch, verify, cache, and activate redistributable
  realm mod payloads before joining; never serve Blizzard-owned game data.
- [ ] Persist realm-owned characters, items, progression, social/account records,
  and session handoff state through transactional versioned storage contracts.
- [ ] Ensure authoritative sessions validate commands and commit character state;
  connected clients must never write trusted realm characters directly.
- [ ] Support realm-managed game lifecycle, capacity/health reporting, graceful
  draining, crash recovery, reconnect tokens, and auditable persistence events.
- [ ] Add compatibility, tampering, interrupted-download, duplicate-session,
  migration, failover, and recovery acceptance tests.

## M24: Packaging and release acceptance

- [ ] Package the engine and shim without Blizzard assets and verify first-run
  installation discovery, configuration, diagnostics, mod isolation, and update
  behavior.

## Gameplay acceptance milestone

- [ ] Starting from legally supplied MPQs, watch/skip startup, navigate the real
  front end, create/select a character, and enter a generated Act I world.
- [ ] Fight monsters using skills, gain experience and loot, equip and store
  items, use town services, complete quests, change areas, save, quit, and resume.
- [ ] Run the same deterministic session as offline play, replay, listen-server,
  and dedicated-server play with matching simulation checksums.
- [ ] Complete all acceptance flows without placeholder rectangles,
  compatibility services, leaked handles, data races, or bundled Blizzard data.

## Performance priorities

- [x] Stop rebuilding and uploading the diagnostic HUD texture every moving frame.
- [x] Avoid redundant pixel conversion and uploads when creating a texture.
- [x] Chunk large DS1 maps into culled textures instead of one full-map GPU texture.
- [x] Cache scene child ordering until topology or Z-index changes.
- [x] Avoid unconditional per-frame callback and input-map allocations.
- [x] Pre-index dynamic loot candidates by item type and level.
- [x] Replace coarse startup sleeps with lifecycle events or bounded waits.

## Parallel codec maintenance

- [x] Audit the codec repositories and record ownership strategy.
- [x] Baseline all eleven modules online and remove filesystem path replacements.
- [x] Add decoder fuzzing, synthetic fixtures, and optional real-asset tests.
- [x] Standardize headless inspection and conversion commands.
- [x] Tag stable codec releases and update Dark Magic dependencies.

See [CODECS.md](CODECS.md) for the module-by-module plan. Codec repositories
remain independent rather than being copied into this engine.

## Current handoff

All restart milestones and the architectural acceptance path are implemented.
The repository builds against tagged codec releases, boots through the internal
host and layered shim, runs the Lua-authored shell and world orchestration, and
passes the complete package suite under race detection. Historical stashes
remain preserved and documented in `STASHES.md`; do not apply or drop them
without a separate comparison task.
