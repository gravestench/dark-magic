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
- [x] M15.3: Migrate front-end, character, HUD, panel, cursor, font, audio, and
  cinematic facts from compiled Go declarations into shim-owned manifests. The
  verified asset catalog and every presentation fact consumed by the runtime
  now live in the embedded shim. The unused legacy MPQ path constant dump was
  removed so Go no longer presents a competing asset-knowledge source.
- [x] M15.4: Generate and validate hash/dimension/frame fixtures from manifests
  without storing proprietary decoded pixels. The 90 verified entries now have
  source hashes, byte sizes, decoder types, direction/frame counts, dimension
  ranges, and deterministic full-frame-metadata hashes. The catalog tool can
  generate or compare the fixture and reports every mismatch in one pass.

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
- [x] M16.2: Add palette-aware DCC decoding and COF-driven component direction,
  mode, weapon-class, layer, transform, shadow, and event composition.
  Palette-aware DCC frame/animation handles and checked COF metadata cover layer
  priority, weapon class, draw effects, transparency, selection, speed, and
  frame events. Lua can assemble supplied equipment-component DCCs into static
  or managed composite animations with component offsets, COF ordering,
  shadow/alpha transforms, timing, and returned frame events.
- [x] M16.3: Decode Diablo font TBL/DC6 pairs and expose measured localized text
  layout with color, alignment, wrapping, clipping, and fallback behavior.
- [x] M16.4: Implement reusable Lua controls with pointer/controller focus,
  hit-testing, state visuals, activation sounds, and accessibility metadata.
  Buttons, UTF-8 text fields, checkboxes, and scrollbars provide pointer
  hit-testing, scoped directional focus, disabled/hidden states, activation,
  change/sound and visual-state callbacks, and accessibility snapshots.
- [x] M16.5: Implement scoped music, UI, SFX, ambience, and speech buses with
  streaming, looping, fades, pan, grouping, and record-driven lookup.
  Generation-checked sounds now route through independently adjustable named
  buses with per-sound volume/pan, owner-thread looping, deterministic fades,
  and grouped stopping. Long-form assets use owner-thread streaming, while
  deterministic `Sounds.txt` lookup resolves paths, routing, volume ranges,
  grouped variants, looping, and streaming intent through layered content.
- [x] M16.6: Add generation-aware CPU/native caches, budgets, diagnostics, and
  leak/race/rapid-scene-transition acceptance tests.
  The shared weighted LRU now supports generation namespaces, stale-entry
  invalidation, live budget enforcement, eviction callbacks, and hit/miss/weight
  diagnostics. Renderer/audio handle diagnostics and a race-tested 50-cycle
  scene/overlay transition check detect leaked resources. Render decoding caches
  now discard DC6, DCC, COF, and font entries on VFS generation changes and are
  governed by a weighted decoded-data budget. Versioned Lua diagnostics expose
  decoded cache, retained renderer, and audio lifecycle state without exposing
  native payloads.

## M17: Authentic Lua-authored front end

- [x] Implement startup/trademark/cinematic sequencing from verified BIK and UI
  assets with skip and failure behavior.
- [x] Replace the placeholder title and main-menu rectangles with verified,
  palette-aware tiled backgrounds and the anchored layered animated logo.
- [x] Add localized front-end buttons, legal/version labels, cursor behavior,
  and title audio.
- [x] Implement single-player, multiplayer/TCP-IP, credits, cinematics, dialogs,
  and navigation using verified front-end assets.
- [ ] Implement character creation with all seven class animation sets, hit
  regions, narration, name validation, expansion/hardcore flags, and cancel.
- [ ] Implement saved-character selection with paging, scrollbar, composite
  previews, metadata labels, deletion confirmation, and double activation.
- [x] Implement progressive loading screens driven by real dependency progress.
- [x] Add screenshot/composition tests for every state without checking Blizzard
  imagery into the repository.

### M17 execution checkpoints

- [x] M17.1: Complete title/main-menu labels, controls, music, cursor, legal and
  version text, dialogs, and single/multiplayer/credits/cinematics navigation.
  The localized Single Player control now uses the verified split DC6 button,
  Exocet TBL/DC6 text, pointer/focus activation, and selection sound. Manifest-
  backed Multiplayer, Credits, and Cinematics controls now navigate to their
  verified sibling backgrounds and return safely. The verified hand cursor is
  shared across front-end scenes, and Exit requests orderly client shutdown
  through `dm.app/v1`. Localized legal copy and the real Go build version render
  through the bitmap-font path. A reusable focus-isolated text-entry modal now
  drives manifest-backed TCP/IP Host/Join intent controls. The supplied
  `Sounds.txt` resolves menu music through `ESOUND_MUSIC_DIABLO`, preserving its
  record-owned streaming/loop/volume behavior. Expansion credits now decode the
  verified UTF-16 text through the VFS and paginate it over the palette-correct
  credits background. The Cinematics scene now exposes all seven verified
  English 640x292 BIK entries as VFS-aware localized controls; playback and
  startup sequencing proceed under M17.2.
- [x] M17.2: Implement startup/trademark/cinematic sequencing and skip/failure
  behavior. A scoped backend-neutral `dm.video/v1` capability now keeps large
  MPQ payloads outside Lua and exposes deterministic playing/complete/failed/
  stopped lifecycle state. The verified Blizzard and Blizzard North startup
  BIKs sequence with explicit skip/failure policy, and the Cinematics selector
  drives the same contract. A strict BIK container parser now bounds dimensions,
  timing, frame indexes, packet sizes, and audio-track configuration before a
  native decoder receives data. An optional FFplay backend now provides real
  Bink video/RDFT audio playback, lifecycle polling, skip/stop, and temporary
  extraction cleanup while retaining failure-safe behavior when FFmpeg is not
  installed. The macOS client now keeps Cocoa/GLFW on the process main thread,
  uses the current raylib 6 bindings, and safely transitions from both startup
  movies into the verified expansion trademark composition. That scene uses
  the distinct `trademarkscreenEXP.dc6` background and the synchronized black
  left/right plus additive fire left/right 30-frame logo layers, then accepts
  click, Space, Enter, or Escape before entering the main menu. Embedded
  in-window playback is tracked in M17.6. Engine-level acceptance now drives
  both startup entries through playing, failed, complete, and skipped states,
  verifies manifest ordering, and asserts scope-owned playback cleanup.
- [x] M17.3: Implement all seven character-creation presentations and rules.
  The expansion creation screen now composes the verified tiled background and
  animates all seven class unselected/hover/selected assets through manifest-owned
  hit regions and anchors. DC6 frame offsets are normalized onto a shared canvas
  so cropped animation frames retain a stable world anchor instead of jittering.
  Empty character lists enter creation instead of silently
  fabricating a hero. A focus/pointer-driven name dialog requests engine-owned
  save identities, while `savecore` canonicalizes the seven supported classes,
  enforces the 2–15 character naming grammar, and rejects duplicate names.
  Verified per-class selection/deselection narration and manifest-backed
  expansion/hardcore checkboxes now feed immutable creation metadata through
  the save capability. Every class now uses manifest-owned forward/back walk
  paths and verified frame counts: deselection returns to idle after a one-shot
  back walk, while successful creation defers loading until the selected
  character's forward walk completes. Headless acceptance verifies that delay.
  Native comparison against established reference captures calibrated the
  seven-character stage and hit regions, restored the foreground campfire,
  heading, focused class copy, deferred creation options, and visible Exit
  control, and corrected the frontend background and character palettes.
  Resource-less grouping nodes are now never drawn, eliminating raylib's
  default-texture artifact without sacrificing retained parent transforms.
- [x] M17.4: Implement saved-character presentation, paging, deletion, composite
  previews, and activation. The manifest-backed two-column/four-row list now
  provides stable focus slots, vertical paging, metadata labels, selection art,
  explicit New/Delete/Exit/OK controls, and the observed 1.25-second pointer
  double-activation policy. Deletion uses an isolated confirmation dialog and a
  narrow engine-owned save operation that clears stale selection safely.
  Class-only and legacy saves now render clipped, scaled, animated previews from
  the verified selected-class assets, including synchronized overlay layers
  where authored. Native real-asset calibration now supplies the authored title
  font, stable segmented focus frame, Font16 metadata baselines, frontend
  palette, scrollbar frames, and Tall/Medium footer controls at their observed
  positions. A development-only deterministic fixture source exercises filled
  pages without touching player saves. Save records can now carry an optional,
  defensively copied appearance snapshot containing the authoritative COF,
  palette, direction, and component DCC paths. The shim feeds that snapshot to
  the generic compositor while legacy records retain their class-only fallback;
  this keeps save decoding and asset resolution out of presentation code.
- [x] M17.5: Implement dependency-driven loading and front-end composition tests
  across supported resolution, language, and game-version variants. An
  engine-owned coordinator now executes named character, loading-asset, and
  world readiness tasks and exposes immutable progress through
  `dm.loading/v1`. The Lua loading screen animates toward that real progress,
  treats its configured sweep duration as presentation smoothing only, reports
  failures, and cannot enter the world while a dependency remains blocked.
  Coordinator, capability, and end-to-end blocked-transition tests cover the
  contract. The manifest now explicitly declares the supported 800x600,
  English, Lord of Destruction profile. Its asset-independent composition
  matrix checks every screen's geometry, exact frontend tile coverage, and all
  referenced localized strings; runtime validation rejects malformed future
  profiles. Existing headless scene acceptance covers every state and resource
  lifetime, while opt-in native captures keep proprietary pixels local.

### M17.6: Embedded, single-window Bink playback

- [x] Decode Bink video and audio in-process behind `video.Backend`; keep
  codec and FFmpeg details out of Lua and scene definitions. The build-tagged
  libav decoder now demuxes seekable VFS input, decodes video packets, converts
  native frames to RGBA, and emits timestamped frames. The same decoder now
  normalizes the preferred Bink audio track to timestamped interleaved stereo
  S16 PCM, with bounded backend scheduling and owner-thread audio streaming.
- [x] Add checked streaming texture updates that transfer decoded frames onto
  the renderer owner thread without replacing the cinematic scene node.
- [x] Present decoded frames on `LayerTransition` with aspect-preserving
  letterboxing inside the existing Dark Magic window.
- [x] Stream decoded PCM through the cinematic audio bus and use audio timestamps
  as the synchronization clock, with bounded frame dropping when rendering lags.
  The audio core and raylib owner-thread backend now support checked S16 PCM
  stream creation, incremental writes, playback, and teardown. The audio owner
  thread now reports generation-checked consumed-frame progress; video switches
  from startup monotonic time to that device clock and retains bounded late-frame
  dropping.
- [x] Implement stop, skip, completion, decode failure, device loss, resize, and
  shutdown cleanup through the existing `dm.video/v1` lifecycle contract. The
  embedded scheduler now uses bounded decode queues, one monotonic media clock,
  late-video dropping, lossless audio ordering, cancellation, and checked
  presenter/audio teardown. Resizable windows now publish owner-thread viewport
  changes and refit every active cinematic without retaining completed
  playbacks. A stalled audio-device clock now fails through the existing video
  lifecycle instead of leaving cinematic playback hung indefinitely.
- [x] Prefer the embedded backend in native FFmpeg client builds and retain
  FFplay as the portable developer/diagnostic fallback.
- [x] Verify every discovered BIK variant and capture startup-to-title
  composition screenshots without committing Blizzard-owned media. A native
  integration matrix now requires successful video and audio decoding for both
  startup movies and all seven localized act cinematics. An opt-in owner-thread
  capture session now waits for stable presented scene frames, writes numbered
  local PNGs, and records dimensions plus hashes in `report.json`. Native review
  verified a clean embedded startup-to-title transition and the centered title
  composition at 800x600. The loading capture also exposed its intentionally
  incomplete progressive sweep; completing that composition remains in M17.5.

## M18: Authentic in-game shell

- [ ] Replace the compatibility HUD and placeholder hero with Lua orchestration
  over real world and character presentation handles. The fake hero rectangle
  is gone: the world scene now binds an optional selected-character COF/DCC
  appearance snapshot and keeps metadata-only characters invisible instead of
  fabricating world art. The real world-tile handle remains pending M20.
- [ ] Assemble the 640x480 and 800x600 control panels, globes, stamina and
  experience bars, belt, skill buttons, minipanel, tooltips, and cursor states.
  The supported 800x600 profile now assembles the six authored panel segments,
  two 48px skill wells, clipped health/mana liquids and overlays, generic skill
  icons, and value-driven stamina/experience fills from manifest-owned facts.
  The authored Run/Walk and mini-panel toggle states are interactive through the
  shared focus manager; the seven-button single-player mini-panel routes implemented
  Character, Inventory, Skill Tree, Automap, and Options overlays while retaining
  unavailable Messages and Quest Log controls as disabled. Belt contents,
  tooltips, cursor states, and the 640x480 profile remain.
- [ ] Implement interactive inventory/equipment, character stats, skill trees,
  automap, quest log, waypoint, party, hireling, vendor, stash, cube, options,
  help, chat, pause, and escape panels from verified assets and records. The
  inventory placeholder is now replaced by the four authored expansion panel
  quadrants, a real close control, ten equipment focus regions, and all 40 item
  cells. The character-sheet placeholder is likewise replaced by the authored
  left-panel quadrants, real close control, localized labels, and an immutable
  engine-owned stat snapshot for progression, attributes, defense, life, mana,
  and stamina. Derived combat and resistance values remain part of M21; item
  instances and transfers remain part of M19.
- [ ] Make panel geometry data-driven from Inventory.txt, SkillDesc.txt, and
  related records; keep only presentation corrections in shim manifests. The
  selected class's expansion Inventory.txt row now owns inventory grid and
  equipment hit geometry. Only the verified DC6 frame order and expansion panel
  origin correction remain in the presentation manifest.
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

## M25: Scene performance and asset residency

- [x] Attribute CPU samples across each complete renderer frame, including
  deferred composition draining, texture upload, native drawing, and media work.
- [x] Report per-scene cache residency, resource counts, upload volume, decode
  time, and eviction so heap growth can be distinguished from intentional reuse.
- [x] Bound decoded DC6/DCC residency using retained-buffer weights instead of
  compressed source size and release presentation resources with scene scopes.
- [x] Share exact duplicate animation surfaces and texture handles while
  retaining timing entries for every source frame.
- [x] Replace per-pixel `image.Image` conversion in texture upload with cached or
  directly decoded GPU-ready buffers and benchmark the loading/front-end path.
- [x] Evaluate indexed GPU textures with palette lookup for DC6/DCC assets after
  the simpler lifetime and upload improvements are measured.
- [x] Stream long-lived frontend music when doing so reduces residency without
  compromising gapless looping, synchronization, or shutdown behavior.
- [x] Add repeatable profiling acceptance runs and budgets for startup, title,
  main menu, character selection, character creation, and in-game scenes.

## M26: Native frame-path profiling follow-up

- [ ] Upload contiguous FFmpeg `image.NRGBA` frames directly when their byte
  layout is already compatible with raylib, retaining conversion for genuinely
  incompatible color models and padded subimages.
- [ ] Measure cinematic upload bandwidth and distinguish frame-update traffic
  from resident GPU texture bytes in scene diagnostics and budgets.
- [ ] Destroy scenes and drain final composition commands on the renderer owner
  thread before native renderer shutdown; require zero pending commands afterward.
- [ ] Repeat the interactive acceptance profile through `game_loading` and
  `game_world`, enforce every tracked scene budget, and compare CPU/heap PDFs
  against the latest frontend-only baseline.

## M27: Diablo UI styling fidelity

- [x] Preserve palette-authored bitmap-font shading, DC6 frame offsets, and a
  shared glyph origin while applying sprite-style color modulation.
- [ ] Replace screen-local font paths and approximate colors with semantic text
  styles pairing the verified bitmap fonts with units, static, sky, and other
  context-specific palettes. The shared style vocabulary and character panel
  migration are complete. Frontend footer and character-select title/metadata
  now use their verified Formal12/Static, Font42/Units, and Font16 pairings with
  separate gold, white, and green PL2 runs. Character creation now uses the
  verified Font30/Units heading and class name, Font16/Units description and
  gold option labels through the shared style path; remaining screens still need
  migration. Credits now preserve their authored heading markers as red/gold PL2
  runs using Formal10/Sky, leaving no direct screen-level `set_text` calls.
- [ ] Implement reusable authored button families with their correct DC6 assets,
  segment composition, fonts, text offsets, normal/hover/pressed/disabled states,
  keyboard focus, activation sounds, and hit bounds. The shared segmented-button
  and text-only label-button components, pressed-state input, and main-menu,
  TCP/IP, cinematics, character-select, and character-creation migrations are
  complete. Both single-frame and segmented definitions are supported with
  authored family text offsets, and inventory/character icon-only close controls
  share the same state and sound contract. Text-entry and confirmation dialogs
  now use visible shared label buttons and semantic text styles.
- [ ] Implement label buttons, multiline alignment, inline Diablo color tokens,
  per-run colors, and optional glyph-sized background highlights. Named inline
  color tokens now select the authentic PL2 text-color index transforms, with
  RGB modulation retained only for transform-less mod content. Optional
  real-archive coverage verifies all 13 transforms and distinct non-empty
  white/gold glyph output; background highlights remain.
- [ ] Implement clamped, origin-aware multiline tooltips with centered labels and
  the authentic translucent-black backing box, then connect item, skill, HUD,
  and panel-control tooltip content. The shared tooltip and inventory/character
  close-control help are complete; item, skill, and HUD content remain.
- [ ] Implement reusable frames, boxes, scrollbars, text boxes, list controls,
  toggles, and modal focus rules from verified Blizzard assets and records.
- [ ] Add native screenshot fixtures for every shared component state and migrate
  every shim screen and overlay away from one-off styling code. The interactive
  five-page `font_lab` now isolates semantic screen styles, bitmap families, PL2
  color slots and contexts, and text layout behavior; automated native screenshot
  fixtures and the remaining component migrations are still outstanding.

## M28: Repository architecture and legacy cleanup

- [x] Inventory every package, its responsibility, importers, lifecycle owner,
  stability, and whether it is engine infrastructure, game/shim implementation,
  tooling, or an intentionally supported external API. Record obsolete,
  duplicate, transitional, and misleadingly named code before moving anything.
- [x] Define and document a strict boundary policy: implementation details live
  under `internal`; `pkg` contains only deliberately supported, independently
  useful APIs with clear compatibility commitments. Default ambiguous legacy
  packages to `internal` rather than exposing accidental public surface area.
- [x] Design a small, feature-oriented top-level layout with obvious homes for
  the application host, runtime capabilities, rendering, assets, audio, video,
  input, persistence, networking, Lua/mod execution, authored content, developer
  tools, and acceptance fixtures. Replace historical `*core`, `services`, and
  other implementation-era names where they obscure ownership or data flow.
- [x] Establish import-direction and ownership rules so low-level facilities do
  not depend on presentation or game logic, Lua-facing adapters do not become
  global service locators, platform backends remain behind engine contracts, and
  commands stay thin composition roots.
- [x] Move accidental public packages from `pkg` into their final `internal`
  homes incrementally, updating imports and tests without compatibility aliases
  unless a real external consumer is identified. Keep independently versioned
  codec modules outside this repository. Host-path expansion is now owned by
  `internal/paths`; all callers migrated atomically and an architecture test
  rejects resurrection of the retired public import path. Process log formatting
  is likewise owned by `internal/logging`, with its public import path retired.
  The weighted generation-aware cache is now `internal/cache`; renderer and Lua
  consumers migrated together and the old public import is guarded.
  The unused service template and empty service-era package are removed;
  contributor guidance now points to the authoritative architecture rules, and
  acceptance tests reject service-mesh imports.
  The unreferenced but useful tween curves are preserved under
  `internal/presentation/easing`; the accidental public package is retired
  without discarding the historical implementation. Renderer-independent world
  scene state is now owned by `internal/presentation/scene`; command, Lua,
  acceptance, test-app, and Raylib callers migrated together, and the retired
  public import path is guarded. Engine-specific decoding, cataloging, and
  inspection now form one `internal/assets` subsystem; runtime, presentation,
  content, test-app, and developer-tool callers migrated atomically, with guards
  for all three retired public paths. Diablo loot generation and TSV adaptation
  now live under `internal/game/loot`; Lua and diagnostic applications migrated
  with the package, and its former public path is rejected. The complete typed
  Diablo TSV schema and legacy lookup corpus is preserved under
  `internal/game/data/model`; the game-data catalog and real-asset validation
  tests now consume that internal authority, and the old public path is guarded.
  No Go packages remain under `pkg`, and CI rejects new public Go source unless
  the architecture policy is deliberately revised.
- [x] Consolidate duplicate models, caches, paths, inspection helpers, service
  remnants, test applications, and tools around one authoritative implementation
  per concept. Delete superseded code only after callers and preserved historical
  work have been audited. Generic table storage and the typed catalog have been
  audited as complementary layers and consolidated under `internal/game/data`:
  `store` owns layered TSV bytes and immutable generic rows, while `catalog` owns
  the one atomic typed snapshot and indexes. The client continues to construct a
  single shared store for typed records, Lua, and audio rather than duplicating
  caches or source resolution. Four small renderer-independent capabilities now
  use feature ownership instead of historical `*core` names: `inputstate`,
  `loading`, `localization`, and `persistence`; all composition, Lua, acceptance,
  and Raylib callers migrated together and the retired paths are guarded. The
  engine-side audio command boundary, sound catalog, buses, and playback state
  now live in `internal/audio`; the native Raylib audio device remains a platform
  adapter and the previous `audiocore` path is rejected. Backend-neutral retained
  nodes, resources, handles, animations, and composition commands now live under
  `internal/presentation/render`; Lua and video produce that contract while the
  Raylib backend consumes it, and the old `rendercore` path is guarded. Cinematic
  contracts, in-process decode coordination, synchronized playback, presentation,
  and the fallback player now live in `internal/video`; FFmpeg stays optional,
  and the retired `videocore` import path is rejected. All native Raylib adapters
  now live under `internal/platform/raylib`, keeping window, input, renderer,
  audio-device, and transitional world presentation code visibly outside the
  engine contracts; previous adapter paths are guarded. Repository-private CLIs
  and manual diagnostic applications now live together under `internal/dev`,
  every README, Make target, and root usage example points to the new commands,
  and an acceptance constraint rejects recreation of the ambiguous old roots.
  Application lifecycle, filesystem observation, transactional reload, and the
  local management API now live together under `internal/app`; the command stays
  their composition root and all previous top-level import paths are guarded.
  Serialized Lua VM ownership, capability modules, dynamic services, scopes,
  diagnostics, console-facing runtime behavior, and tests now live together at
  `internal/runtime/lua`; the existing `modruntime` package name remains explicit
  at call sites while the old top-level import path is rejected. Scene and overlay
  stack ownership now lives at `internal/presentation/navigation` beside retained
  rendering and headless scene state; command, acceptance, and Lua callers moved
  together and the former top-level path is guarded. Native-run screenshot capture
  and CPU, heap, and scene profiling now live under `internal/dev`, alongside the
  tools that inspect their artifacts; the client retains explicit opt-in wiring
  and the former top-level paths are rejected. The typed TSV model layer no longer
  imports the Lua VM: an unused legacy `SoundEntry` table-export method was
  removed, current data exposure remains in `internal/runtime/lua`, and an
  architecture test prevents this dependency inversion from returning. The
  repository audit now finds one authority for each named concept, no surviving
  service-era implementation, and no duplicate developer command roots; future
  changes should be feature work rather than compatibility consolidation.
- [ ] Restore the complete typed Diablo TSV record layer as an internal game-data
  catalog. Preserve and verify every existing table struct and CSV tag, recover
  useful loading, indexing, lookup, validation, hot-reload, and Lua exposure
  behavior from the retired record-manager work, and replace its service wrappers
  with explicit stores/catalogs owned by the application host. Generic row access
  remains available for mods and unknown columns, but is not a substitute for the
  typed records that drive simulation, rendering, audio, world generation, and UI.
  The internal reflection decoder, deterministic primary-key indexing, atomic
  snapshot/invalidation contract, and real-archive verification are complete for
  CharStats, Levels, Objects, Skills, Sounds, and TreasureClassEx. Admission also
  recovered the legacy grouped-column tag convention and corrected monster-code
  fields that real records prove are strings; remaining schemas and consumers
  still need admission and migration. Typed binding now delegates to the existing
  independent TSV codec; Dark Magic retains layered source ownership, generic
  rows, compatibility overlays, diagnostics, and immutable catalog policy. The
  base-item spine now admits Armor, Weapons, Misc, and ItemTypes with deterministic
  code indexes. Generic rows preserve shipped duplicate and unnamed columns under
  stable synthetic keys instead of rejecting the affected tables. The dependent
  item-rule layer now also admits ItemRatio, ItemStatCost, Properties, UniqueItems,
  and SetItems, with immutable name/code/index lookups where the authored schema
  provides a stable key. Ordered typed generations now also preserve MagicPrefix,
  MagicSuffix, AutoMagic, RarePrefix, and RareSuffix; deliberately repeated display
  names remain ordered records rather than being forced into lossy lookup maps.
  Gems, Runes, CubeMain, and Sets now form the next admitted typed generation,
  with stable code/index lookups for gems and set definitions and ordered recipes
  where descriptions or display names are not authoritative identities. The world
  dependency layer now admits LvlTypes, LvlPrest, LvlMaze, LvlWarp, and LvlSub,
  including real-data semantic checks that corrected the legacy LvlSub `Type`
  binding rather than accepting a silently zero-filled field. MonStats, MonStats2,
  MonLvl, MonProp, MonSounds, and MonEquip now provide the admitted typed monster
  foundation with immutable ID indexes and ordered scaling/equipment records.
  Missiles, States, Overlay, and PetType now cover combat presentation/state data;
  fixed-size grouped legacy fields are rebound element-by-element so PetType's
  `mclass` and `micon` arrays retain their authored values. The client composition
  root now owns and eagerly validates the complete admitted catalog generation;
  hot reload invalidates through that owner so generic and typed records cannot
  drift into different generations. Experience, Inventory, Belts, Hireling, and
  Difficultylevels now supply the admitted character progression, layout,
  mercenary, and difficulty-rule configuration layer. SkillDesc, SoundEnviron,
  and AutoMap now admit skill-tooltip layout, area audio environments, and
  automap-cell presentation data with stable skill and environment indexes.
  Npc, Shrines, MonPreset, and Gamble now preserve town economics, world
  interactables, preset population, and gambling-pool configuration. ObjType,
  ObjGroup, and ObjMode now admit object tokens, population distributions, and
  animation-mode metadata. QualityItems, WeaponClass, and Books now preserve
  superior-item modifier rules, animation weapon classes, and scroll/book
  casting economics. MonSeq, MonUMod, UniqueAppellation, UniquePrefix, and
  UniqueSuffix now preserve ordered animation events, special-monster modifier
  rules, and the authored name-part pools. TreasureClass and HireDesc now
  preserve the classic drop graph and hireling voice-selection metadata
  alongside their expansion-era counterparts.
  The bundled compressed Data File Guide is now the required semantic reference
  for further schema and consumer work, while real layered MPQs remain the
  compatibility authority when documentation and shipped bytes disagree.
  Guide-reviewed SuperUniques now provides stable encounter and hardcoded-ID
  indexes plus documented class, modifier, group, transform, and treasure-class
  relationships; undocumented patch columns remain available through generic
  rows rather than speculative typed fields. Guide-referenced LowQualityItems,
  BodyLocs, StorePage, Composit, and HitClass now expose their authored
  name/code dictionaries for item naming, equipment placement, vendor UI,
  composite animation, and impact presentation.
  PlayerClass, PlrMode, PlrType, MonMode, and MonPlace now preserve the actor
  class/type/mode token dictionaries and ordered preset-placement codes cited
  by the guide and dependent tables.
  Colors, CompCode, ElemTypes, Events, MissCalc, and SkillCalc now expose the
  guide-referenced transform, component, element, callback-event, and formula
  dictionaries without embedding those authored descriptions in engine code.
  ArmType and CubeMod now expose the remaining unambiguous hardcoded armor and
  cube-modifier dictionaries. Direct classic-archive inspection resolved the
  guide's broken UniqueTitle section: the shipped table is an ordered 16-row
  title-fragment list whose `Name` column is always the repeated `unused`
  sentinel, so it is admitted without a lossy index. CharTemplate remains
  deferred because it is absent from the supplied classic patch data and has no
  usable guide section or surviving schema. Audio record resolution now consumes
  the client-owned shared table store instead of constructing and indefinitely
  caching a second Sounds generation; hot-reload invalidation therefore reaches
  typed records, generic Lua rows, and audio through the same source generation.
- [x] Add package documentation and a concise newcomer architecture guide showing
  the boot path, frame path, scene/mod boundary, asset path, and where new code of
  each kind belongs. Keep examples aligned with the resulting structure.
- [ ] Enforce the intended architecture with dependency tests or static checks,
  package-level tests at moved boundaries, and CI checks that reject new accidental
  public packages, forbidden imports, compatibility-service resurrection, and
  business logic added directly to `cmd`.
- [ ] Complete the migration as behavior-preserving commits with the full unit,
  race, real-asset, and interactive acceptance suites green; then remove temporary
  migration notes and verify a newcomer can locate and extend a representative
  renderer, asset, Lua capability, scene, tool, and server feature from the guide.

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
