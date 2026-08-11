# Dark Magic Restart Roadmap

This roadmap favors complete, testable vertical slices over adding more isolated
services. Diablo II data is never distributed with Dark Magic; users must supply
their own legally obtained game data.

## Milestone status and delivery policy

Checkboxes describe repository evidence, not aspiration. A milestone remains
open while any unchecked acceptance item remains, even when its underlying
architecture is already usable. Long notes under an unchecked item record
partial delivery; they do not make the item complete.

Milestone work uses a dedicated branch and pull request for each independently
reviewable checkpoint:

1. Branch from a clean, current `main` as `codex/mNN-checkpoint-name`.
2. Keep one behavioral objective and its tests, documentation, and migration in
   that pull request. Do not combine unrelated later-milestone cleanup.
3. Require `make test`, `make architecture`, and `make test-race`. Add the
   relevant real-asset capture/profile matrix when presentation or native media
   changes.
4. Record the completed checkpoint here in the same pull request. Merge before
   branching the next dependent checkpoint.

The milestone-number order is historical, not a strict dependency graph. The
active queue below closes bounded earlier gaps before item/gameplay work:

1. **M15.5 — presentation catalog coverage:** inventory every runtime-consumed
   presentation asset/dependency and report manifest or fixture gaps.
2. **M18.1 — authoritative overlay input routing:** make world/HUD/overlay/modal/
   cursor/debug focus ownership explicit and acceptance-tested.
3. **M18.2 + M27 — functional 800x600 HUD controls:** finish belt, skill, tooltip,
   cursor, and shared styling needed by item interaction.
4. **M19.1 + M29 gameplay-command integration — authoritative item containers:**
   implement inventory/body/belt/cursor commands and connect the Lua presentation
   without client authority.
5. Resume M18/M19 vertically per overlay; schedule independent M26 native-path
   work when it does not destabilize gameplay PRs.

The creature-authoring program in M31-M43 is a parallel, reusable asset track.
It may begin with research and synthetic fixtures without displacing the active
gameplay queue. Its dependency order and complete per-milestone contracts live
in [docs/CREATURE_ASSET_ROADMAP.md](docs/CREATURE_ASSET_ROADMAP.md), including a
candidate-tool radar for manual authoring, rigging, validation, and image-to-3D
experiments behind the same canonical import boundary.

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
- [x] Keep item and loot generation renderer-independent and driven by typed
  record catalogs rather than a lifecycle-managed generator service.

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
- [x] Provide a backend-neutral retained rendering contract with Raylib isolated
  as a native platform adapter rather than exposing a global renderer service.
- [x] Preserve chunking, culling, child-order caching, and allocation-free hot paths
  from the current renderer while replacing its ownership model.
- [x] Add a headless render-command backend for scene tests and golden composition
  fixtures where practical.
- [x] Prove the boundary by rendering one complete Lua-authored screen using only
  versioned capabilities.
- [x] Retire the exported Raylib `Renderable` object graph and unused direct-native
  world adapter. The backend now keeps private native state only, advances
  animations outside nodes, caches dirty transforms, intersects hierarchical
  clips, culls hidden subtrees, reports native work, and drains final commands
  before native shutdown.

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
- [x] Move verified screen facts from Go literals into versioned shim manifests
  with confidence, game-version, language, and resolution fields.
- [ ] Catalog every front-end screen, HUD/panel sheet, cursor, font, cinematic,
  sound cue, and relevant TXT/TBL dependency (M15.5).
- [x] Add comparison fixtures that verify expected dimensions, frame counts,
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
- [x] M15.5: Generate a deterministic coverage report joining runtime-consumed
  asset paths, presentation-manifest entries, and verified fixture entries.
  Classify deliberate dynamic paths separately, fail CI for unexplained static
  gaps, and document the remaining screen/HUD/font/audio/TXT/TBL inventory. The
  headless report currently tracks 139 manifest paths (73 verified, 66 pending),
  47 code-owned migration paths, one dynamic prefix, and 17 catalog-only paths;
  CI pins the complete classified-set fingerprint and rejects catalog/fixture
  drift.

## M16: MPQ-backed presentation primitives

- [x] Decode DC6 assets with an explicit palette through managed render
  resources exposed to Lua by checked handles.
- [x] Decode DCC assets with an explicit palette through the same managed path.
- [x] Support DC6 frame/direction selection, placement offsets, anchoring,
  front-end tiling, deterministic layer order, and alpha/additive/multiply blend
  modes.
- [x] Support native animation timing, loop modes, and clipping.
- [ ] Support reusable general tiling and nine-slice composition.
- [x] Decode Diablo font TBL/DC6 pairs and render localized, colored, aligned,
  wrapped text.
- [x] Add Lua-authored buttons, text fields, checkboxes, scrollbars, focus order,
  controller navigation, hit testing, hover/pressed/disabled states, and cursor
  sounds.
- [x] Add music/SFX/UI/ambience/speech buses, looping and streaming, sound-record
  lookup, grouped variants, fades, pan, and scope-owned playback.
- [x] Cache decoded CPU assets separately from renderer/audio resources and
  invalidate them by VFS generation without leaking owner-thread resources.
- [x] Add palette quantization to either the final display or individual
  retained textures and animations. A CPU-built 32³ nearest-color lookup cube
  is uploaded as a point-sampled shader texture, so every output RGB value
  belongs to an arbitrary non-empty raw-BGR, GIMP GPL, or JSON palette without
  a search per fragment. Monotone and black/white palettes are supported; alpha
  and resizable viewport geometry survive the off-screen composition pass.

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
- [x] Implement character creation with all seven class animation sets, hit
  regions, narration, name validation, expansion/hardcore flags, and cancel.
- [x] Implement saved-character selection with paging, scrollbar, composite
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
  paths and verified frame counts. Activating a class now plays its one-shot
  forward walk before entering the selected idle and opening name entry;
  switching classes concurrently returns the previous class with its back walk.
  Shipped base/effect pairs share composite bounds and one manually synchronized
  clock. Idle/hover/selected loops retain their 15 FPS cadence while forward and
  back walks use the 25 Hz frontend rate that aligns with selection/deselection
  cues. Manifest-owned stage depth keeps overlapping characters stable across
  resource changes, including Paladin in front of Barbarian. Input is gated during those transitions, and successful creation proceeds to
  loading only after selection presentation has completed. Headless acceptance
  verifies the walk-before-dialog order and a cross-class switch.
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

M18 is partially implemented, not complete. The merged UI compatibility work
establishes reusable controls and visually reviewable shells; an overlay is not
complete until its actions are driven by authoritative game state and commands.

- [x] Replace the compatibility HUD and placeholder hero with Lua orchestration
  over real world and character presentation handles. The fake hero rectangle
  is gone: the world scene now binds an optional selected-character COF/DCC
  appearance snapshot and keeps metadata-only characters invisible instead of
  fabricating world art. The Lua-authored world now owns an MPQ-backed DS1
  presentation adapter decoded from manifest-selected DT1 tiles and an Act
  palette; camera-near chunks follow absolute authoritative camera coordinates
  without transferring filesystem or native renderer ownership into Lua. Tile
  generation and entities remain pending M20. The unused direct-Raylib compatibility world
  package and its public renderer-object dependency have now been removed.
- [ ] Assemble the 640x480 and 800x600 control panels, globes, stamina and
  experience bars, belt, skill buttons, minipanel, tooltips, and cursor states.
  The supported 800x600 profile now assembles the six authored panel segments,
  two 48px skill wells, clipped health/mana liquids and overlays, generic skill
  icons, and value-driven stamina/experience fills from manifest-owned facts.
  The authored Run/Walk and mini-panel toggle states are interactive through the
  shared focus manager; the seven-button single-player mini-panel routes Character,
  Inventory, Skill Tree, Automap, Messages, Quest Log, and Options overlays with
  verified up/down states and localized tooltips. The belt exposes authoritative
  capacity and cursor presentation covers the interaction states currently admitted
  by game authority. Item contents and transfers remain M19 dependencies, learned
  skill selection remains an M21 dependency, and the 640x480 profile remains.
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
  The recovered Escape/options hierarchy now uses verified `OptBarC.dc6` and
  `OptSkull.dc6` controls for live, persistent sound and music preferences via
  `dm.settings/v1`. Unsupported 3D bias, gamma, contrast, and key-binding rows
  remain explicitly unavailable until their engine effects and recovered ranges
  are implemented rather than displaying fabricated values.
- [ ] Make panel geometry data-driven from Inventory.txt, SkillDesc.txt, and
  related records; keep only presentation corrections in shim manifests. The
  selected class's expansion Inventory.txt row now owns inventory grid and
  equipment hit geometry. Only the verified DC6 frame order and expansion panel
  origin correction remain in the presentation manifest.
- [ ] Implement correct focus/input routing when world, HUD, modal, cursor,
  transition, and debug layers coexist.
- [x] Add a shared shell with command history, multiline editing, source-aware
  errors, structured value inspection, and log output. The renderer-independent
  session core, persistent scoped Lua evaluator, and rich Charmbracelet v2
  terminal adapter are complete. The Raylib client now provides a grave-key
  overlay using the same session, suppresses scene input while focused, and
  remains legible above palette quantization. Its embedded Go Mono TrueType
  face and Unicode-safe transcript wrapping are platform-independent. Opening
  elastically settles from above while opacity fades in; closing slides upward
  on a normalized cubic curve with a quick nonlinear fade and retains input
  focus until gone. Lua print/register diagnostics and a bounded structured
  `slog` tail feed revision-cached F1 Lua and F2 Logs modal views in both the
  Raylib and Charm frontends. The graphical editor also supports a movable
  Unicode cursor, insertion/deletion, Home/End, visible completion candidates,
  and independent paged log scrollback. Each Lua view begins with a shared,
  target-aware MOTD, and exposes the policy-filtered `dm`/`darkmagic` root with
  friendly lazy capability aliases, help, discovery, and exact versioned module
  access. Versioned module registrations now own structured module and command
  help (usage, parameters, returns, and examples); `dm.help` accepts module or
  command values and string paths, while legacy commands receive generated
  fallback documentation. Multiline values now preserve newlines and indentation
  through both adapters. The shared `dm.shell/v1` capability exposes validated
  live presentation settings; font-size edits apply immediately and explicit
  defaults, reset, dirty status, and atomic platform-native persistence are
  available from client, server, and realm shells. All built-in capability
  functions now carry authored summaries and usage signatures checked against
  their actual exported Lua tables. Policy-filtered `dm.apropos` searches the
  registry and `dm.docs()` renders Markdown from that same source of truth.
  Structured table output now uses deterministic multiline formatting with
  cycle/depth protection. Lua output has retained Page Up/Page Down scrollback
  and semantic heading/code/value coloring in both adapters. Persistent settings
  additionally control console height, opacity, transcript retention, and
  animation speed, with atomic `set_many` and explicit `reload`. Audio sound,
  video playback, and all render-node userdata methods participate in help,
  Markdown generation, and metatable-level documentation conformance checks.
  Game composition now renders into a fixed logical-resolution target while
  shell overlays render afterward in native window pixels. Presentation can
  either preserve aspect ratio with centered letterboxing (`contain`, default)
  or fill the window (`stretch`), and pointer input applies the inverse viewport
  transform while rejecting letterbox regions.
- [x] Provide capability-aware autocompletion for Lua keywords, console-local
  names, permitted modules, and table/userdata members; Tab and Shift-Tab cycle
  candidates, a shared prefix is inserted first, and candidate browsing never
  executes code or invokes metamethods with side effects. Keywords, locals,
  modules, globals, raw table members, and raw userdata method tables are
  implemented without invoking completion-time metamethods.
- [ ] Bind each console tab to an explicit Lua runtime and dedicated disposable
  resource scope; serialize evaluation through the runtime owner and clearly
  display the selected runtime, scope, and capability set. Sessions already own
  the explicit runtime scope and the terminal header exposes target and policy;
  graphical multi-tab presentation remains.
- [x] Expose this exact shell/session contract through the graphical client,
  standalone game server, and realm targets. Headless targets use the shared
  Charmbracelet modal TUI; the in-game console supplies only a different view
  adapter. Initial `darkmagic-server` and `darkmagic-realm` composition roots
  now own that same renderer-free administration surface.
- [ ] Gate mutation/debug capabilities by session policy: full local developer
  access when enabled, read-only inspection where appropriate, and no path from
  a realm-connected client to server-authoritative state or server-only APIs.
  Module discovery and `require` now enforce the session allowlist and
  read-only policies reject shell-global assignment; target-specific
  read-only capability modules remain to be defined with networking.
- [ ] Tear down console-created handles, callbacks, subscriptions, and render
  nodes on reset, runtime reload, disconnect, scene change when configured, and
  client shutdown.

### M18 execution checkpoints

- [x] M18.1: Define and test one focus owner across world, HUD, overlay, modal,
  cursor, transition, and debug layers, including cancellation and scene change.
  Each native frame now carries an explicit scene/debug/none owner. The debug
  shell captures before scene publication, and the Lua input capability gives
  nonfocused callbacks either no input or an explicitly filtered gameplay-only
  channel while their simulation updates continue. HUD and cursor remain
  presentation owned by the focused scene rather than competing consumers.
- [ ] M18.2: Complete the functional 800x600 control panel, belt, skill wells,
  tooltips, cursor states, and minipanel behavior using authoritative snapshots.
  Inventory, character, skills, automap, help, quests, and party overlays now
  explicitly declare whether gameplay passes beneath them and whether the
  visible world frames the player in the left, right, or centered region.
  The health/mana globes and stamina/experience bars now consume a frame-stable
  value snapshot: live vitals and progression come from the admitted player ECS
  entity, while fields not yet modeled by the simulation fall back to immutable
  selected-save metadata. Hover tooltips report the same values. Belt contents,
  authoritative skill selection, cursor modes, and remaining minipanel behavior
  remain subsequent checkpoints.
  Run/walk is no longer a Lua-local visual toggle: the HUD and R shortcut queue
  local intent into the fixed-tick movement source, the replayable movement
  command applies speed and movement mode, and the button plus tooltip update
  only from the resulting ECS snapshot.
  Left/right skill wells now use the original 48px desktop geometry recovered
  from OpenDiablo2 and corroborated by riiablo, resolve generic or class skill
  sheets through typed Skills/SkillDesc records, and display localized hover
  tooltips. Assignments pass through a replayable command into the player ECS.
  The desktop selector opens upward in the five recovered logical rows, applies
  left/right eligibility and passive filtering, and closes its opposite side.
  Choices come from session-owned `dm.player.learned_skill` entities seeded by
  typed general/start-skill records; Lua only arranges and describes them.
  Save-imported purchased skills and charged item skills remain M19 work.
  Riiablo's mobile controls were explicitly
  excluded from the fixed 800x600 profile; the local AbyssEngine checkout has
  no corresponding skill-well implementation, and OpenD2 is not present.
  The belt now has an authoritative 16-slot identity component with a four-slot
  no-belt baseline. Its retained desktop control uses the original 4x4, 31px
  geometry and `ctrlpnl_popbelt.DC6`, revealing only capacity granted by the
  equipped belt. Item entities, cursor transfers, and consumption remain M19
  dependencies and are not simulated with presentation-owned placeholders.
  Cursor presentation now distinguishes the original animated five-FPS in-game
  pointer from static normal, pressed, and interaction-hand modes. OpenDiablo2,
  riiablo, and AbyssEngine provenance is retained in the asset catalog; item and
  vendor cursor modes remain gated by their authoritative M19/M21 interactions.
  The single-player mini-panel now uses every verified even/odd up/down frame,
  localized hover/focus tooltips, stateful open/close feedback, and a live
  non-pausing Messages shell. Party remains correctly absent from the compact
  single-player panel instead of adopting riiablo's multiplayer-width layout.
- [ ] M18.3: Convert inventory/equipment and the first dependent item overlays
  from visual shells to command-backed interfaces; continue with character,
  skills, automap, quests, waypoint, NPC, hireling, social, and help surfaces.
  Replace the current single-top-scene UI focus assumption with spatial overlay
  input routing. The persistent HUD and mini-panel must remain interactive while
  a gameplay panel is open, including using the originating mini-panel button to
  close that panel. A left-side and right-side panel may coexist and independently
  receive pointer input in their respective regions. Opening a new panel must
  atomically replace the active panel in the same left/right slot; opening a
  full-screen panel must close both side slots. Hotkeys and mini-panel actions
  must use the same slot-aware toggle operation so they cannot produce divergent
  stacks or duplicate overlays. Keyboard/gamepad focus remains singular and
  deterministic even when pointer input is routed to multiple visible surfaces.
  The navigation manager now owns those left/right/full slot transactions and
  routes focused, pointer-only, gameplay-only, or combined HUD/gameplay input to
  each visible scene. Hotkeys and the mini-panel share the Lua toggle capability;
  regression coverage exercises opposite-side coexistence, same-side replacement,
  non-top closure, full-panel eviction, and persistent HUD pointer delivery.
- [x] M18.4: Complete data-driven panel geometry and the supported 640x480 profile.
  Manifest profiles now select the native logical render target and deep-merge
  sparse Lua-facing presentation overrides. A scoped 640x480 gameplay profile
  uses the archive-verified six-frame `ctrlpnl7.dc6` panel, recovered desktop
  HUD placements, and a profile-relative world-input split. Inventory now selects
  the unsuffixed 640x480 `Inventory.txt` records, while character, skill, quest,
  and party panels apply profile-owned classic origins. Their close controls use
  named overlay-slot transactions so either visible side can close independently.
  Help now selects and composes the archive-verified eight-frame 640x480 border;
  stash, cube, hireling, vendor, and waypoint now share manifest-owned fixed-panel
  geometry with archive-verified structural fixtures and classic `(0,4)` origins.
  Pause and options derive their dim surface, menu center, retained visuals, and
  pointer hit regions from the selected logical viewport. Automap placeholder
  bounds and the archive-matched death typography are now profile-owned, with
  the classic death copy centered at 320 while retaining its recovered vertical
  rhythm. Startup and game-loading backdrops now follow the selected logical
  viewport around their independently managed movie/loading presenters. The
  expansion frontend remains deliberately excluded because its backgrounds and
  interaction geometry are authored for the 800x600 shell. M18.4 is complete for
  the profile's explicit gameplay scope and does not claim complete 640x480
  application support.
- [ ] M18.5: Complete multi-runtime console tabs, policy gating, and deterministic
  teardown of console-created resources.

## M19: Character, item, and save fidelity

- [ ] Implement COF/DCC composite animation with equipment components, weapon
  classes, directions, modes, layer order, transforms, shadows, and events.
  The playable-character path now resolves typed item-owned component and weapon
  class recipes, preserves one playback clock across direction/equipment changes,
  reads rates and frame events from `AnimData.d2`, caches complete composed
  direction/mode resources, and emits crossed events even when one update skips
  multiple frames. Its first shadow pass honors COF shadow flags. Exact legacy
  shadow projection, PL2 transforms, non-player families, and complete imported
  equipment recipes keep the broader milestone open.
- [ ] Replace placeholder character storage with versioned Diablo II save parsing,
  validation, creation, migration, and non-destructive writing.
- [ ] Implement authoritative inventory grids, body slots, belt, cursor item,
  alternate weapons, stash/cube/vendor transfers, and placement validation.
- [ ] Connect deterministic loot instances to item serialization, names, colors,
  requirements, durability, sockets, runewords, set state, and tooltip stats.
- [ ] Recalculate character attributes and presentation after every equipment,
  skill, level, state, and difficulty change.

### M19 execution checkpoints

- [ ] M19.1: Define authoritative inventory, body, belt, and cursor-item state;
  admit deterministic pickup/place/move/equip/drop commands with validation,
  replay coverage, and audited administrator equivalents. Treat the legacy
  in-hand item as a persistent authoritative `held` container rather than UI
  cursor state, so disconnect/reconnect and realm handoff preserve it. Model
  backpack, stash, cube, player/hireling equipment, belt, vendor/quest-service
  escrow, held, and world as explicit transfer locations. Keep animated
  world-drop DC6 presentation separate from the front-facing inventory DC6 used
  by panels and the held cursor. Backpack, stash, and cube placement validates
  the complete item footprint: no overlap places normally, exactly one
  overlapping item performs an atomic held-item swap, and multiple overlaps
  reject without mutation. Quest and reward panels expose authored spatial item
  sockets. Vendor stock is shown in paged category grids, but selling is an
  authoritative transaction that auto-arranges the transferred item instead of
  a pointer drop into a player-selected vendor cell. Current progress:
  fixed-tick authority, copied Lua snapshots,
  backpack, stash, and cube grid interaction, canonical body-location
  interaction, four-slot belt interaction, and MPQ-backed item presentation
  are implemented. Hireling equipment now uses an independent authoritative
  slot domain with MPQ-backed slot presentation. Named quest-service escrow
  slots support the same atomic held-item replacement rule; the primitive can
  also address authority-assigned vendor catalog positions, but selling and
  buying remain distinct transactions. Alternate hand-equipment sets now have
  independent `rarm`/`larm` occupancy, a replayable fixed-tick selection
  command, Lua snapshot/intent bindings, and the canonical W-key presentation
  path; armor and jewelry remain shared across both sets. Container authority
  now exports deterministic, checksummed versioned archives and restores them
  atomically through the ordinary placement validators, preserving held items,
  belt contents, escrow, and the active alternate-weapon set across reconnect
  or realm handoff. Vendor stock now uses authority-owned category/page grids:
  replayable buy/sell intents cannot provide coordinates, and the server sorts,
  packs, pages, and compacts real item footprints using the recovered Riiablo
  `VendorGrid` traversal order. Durable Diablo II save integration, NPC price
  and player-gold ledger rules now keep separate carried/stashed balances,
  derive server-side prices from item base cost plus authoritative
  NPC.txt 1024-based buy/sell multipliers and sale caps, and commit gold with
  item movement atomically; command payloads cannot provide prices. Durable
  Diablo II save integration remains M19.3 work. Quest/vendor service commands
  now resolve server-authored target/material sockets and costs, validate every
  input before mutation, consume materials, debit carried gold, and record the
  applied operation on the surviving item atomically. Rich socket/imbue/repair
  stat materialization remains coupled to M19.2 item properties, and service
  presentation remains an M18 dependency; those keep this checkpoint open.
  Session checkpoints now hash item authority together with ECS state, restore
  initial item archives for verification, and reject mismatched trade/service
  rule identities, so replay coverage verifies transaction results rather than
  merely retaining the submitted commands.
- [ ] M19.2: Connect loot materialization, world drops, item names/tooltips, and
  Lua UI snapshots to M19.1 without making presentation authoritative.
- [ ] M19.3: Add versioned save parsing/writing and non-destructive round trips for
  character identity, progression, equipment, and item containers.
- [ ] M19.4: Complete equipment-driven COF/DCC appearance and derived-stat
  recalculation, including durability, sockets, sets, and runewords.

## M20: World fidelity

- [ ] Import DS1/DT1 collision, orientation, material, warp, object, shadow, roof,
  and subtile flags into deterministic chunks and zones. DS1 dimensions,
  authored objects, and combined DT1 subtile collision flags now decode into a
  renderer-independent immutable map exposed through `dm.world/v1`; remaining
  orientation/material/warp/roof facts and chunk ownership are pending.
- [ ] Implement entity-size-aware A* navigation, path smoothing, collision,
  interaction range, line of sight, and movement modes.
- [ ] Implement deterministic preset, maze, outdoor, substitution, warp, object,
  monster, and waypoint generation for all acts and difficulties. Use D2MOO as
  the reverse-engineered 1.10f runtime baseline and jaenster/libd2 as a
  clean-room 1.14d implementation/verification source; record version conflicts
  and reproduce claims with independent deterministic probes.
- [ ] Materialize generated maps as deterministic session-owned zones from typed
  level records and decoded DS1/DT1 facts; asset loading and presentation must
  remain consumers rather than owners of world generation.
- [ ] Add animated map tiles, doors, breakables, shrines, chests, portals,
  missiles, overlays, selectable bounds, and automap discovery.
- [ ] Stream and cull zones without changing deterministic simulation results.

### M20 execution checkpoints

- [x] M20.1: Extract one immutable DT1 catalog keyed by
  orientation/main-index/sub-index; select rarity-weighted physical records
  deterministically for each DS1 coordinate; and preserve source identity,
  material, collision, roof, and presentation-pass metadata without decoding
  graphics. World collision and future presentation adapters now consume the
  same selected placements. Selection no longer collapses coordinates on the
  zero axes or admits zero-rarity alternatives when positive weights exist.
- [x] M20.2: Add sparse 512x512 map composition and migrate DS1 Lab to
  incrementally admitted, viewport-culled chunk nodes while retaining the full
  image composer for diagnostic PNG export. CPU preload never creates native
  textures; at most two newly visible chunks become demand-resident per frame,
  and chunks outside the viewport margin release their retained nodes. The game
  world migration remains the explicit M20.4 checkpoint.
- [x] M20.3: Present floor, lower-wall, shadow, upper-wall, roof, and north-corner
  pairs explicitly. DS1 Lab isolates each tile pass with Tab, reports
  demand-resident versus total chunks and authored object placements, and is
  verified against the Act I Barracks fixture. Objects remain semantic
  placements for authoritative systems rather than pixels baked into chunks.
- [x] M20.4: Migrate `game_world` to the shared chunked adapter. CPU composition
  is queued off-thread, camera-near textures are admitted gradually and culled
  against the unobscured left/right view, and map placement derives exclusively
  from absolute authoritative camera coordinates without a scene-entry baseline.
- [x] M20.4a: Replace eager whole-zone raster composition with a lightweight
  DT1 placement/depth index. The camera queues only chunks intersecting its
  unobscured viewport plus a prefetch margin; bounded workers rasterize those
  chunks before Lua admits their retained nodes, and conservative index bounds
  are replaced with exact transparent-trimmed bounds at admission.
- [x] M20.4b: Replace runtime depth-chunk RGBA copies with one shared immutable
  texture per selected physical DT1 graphic plus lightweight placement nodes.
  Viewport culling now admits and releases occurrences independently while a
  reference-counted render resource pool retains each visible source graphic
  only once. DS1 Lab keeps composed chunks as an intentional diagnostic view.
- [x] M20.4c: Spatially bucket shared DT1 placement commands in deterministic
  512-pixel map cells. Camera updates reverse-project the unobscured viewport,
  visit only intersecting buckets, deduplicate edge-spanning placements, and
  release retained nodes and finished preload jobs outside the culling margin
  instead of scanning every placement in the assembled zone each frame. Map
  residency covers the full physical viewport behind transparent panel artwork;
  overlay `world_view` remains solely a camera-framing and input-routing policy.
  Runtime residency also prefetches one viewport-quarter camera shift plus a
  legacy tile width, so opening, closing, or combining side panels cannot expose
  a one-frame black strip while visible DT1 jobs catch up.
- [x] M20.4d: Extend the persistent texture-residency overlay with exact
  last-frame draw-call, scene-node traversal, subtree-culling, and texture-update
  counters. Backend batching or instancing is now gated on measured game-world
  workload instead of total map-placement counts.
- [x] M20.4e: Clamp overlay-shifted presentation cameras to finite map-canvas
  bounds. Side panels still reframe the player into the visible half, but can no
  longer pull an authored zone edge onscreen as a persistent black strip.
- [x] M20.4f: Add development capture fixtures for left, right, simultaneous
  side, and full-screen overlays above a production-asset game world. Fixture
  startup uses the ordinary navigation stack, while camera-boundary tests keep
  pointer reverse projection aligned with the clamped presentation.
- [x] M20.4g: Add bounded game-world diagnostics: F3 authoritative collision,
  F4 tile/subtile projection geometry, and F5 entity origins. Each map overlay
  rasterizes only a bounded authoritative subtile region around the player and shares the map's
  camera transform. It avoids the hundreds-of-megabytes full outdoor overlay
  while making DT1 art/collision alignment directly inspectable.
- [x] M20.4h: Use the center of one DT1 subtile—not half of an entire tile—as
  the authoritative projection origin. Collision diamonds, entity origins,
  pointer reverse projection, and rendered floor art now share one coordinate plane.
- [x] M20.4i: Add backend-neutral texture pivots and anchor COF composites at
  their authored logical origin (the feet/ground-contact point), rather than
  centering changing character and shadow bounds on the world position.
- [x] M20.5: Correct DT1's bottom-to-top 5x5 subtile-row mapping at the world
  boundary, add exhaustive mapping/projection probes, and provide an F3 DS1 Lab
  collision overlay. Red marks walk blocking, orange player-only blocking, blue
  LOS, magenta jump, and yellow light blocking.
- [x] M20.6: Establish a versioned deterministic generator request and immutable
  zone contract. Canonical zones retain tile-space bounds, authored DS1/DT1
  recipes, rooms, links, warps, spawns, and a generation trace without owning
  images or native renderer resources. Purpose-named random streams isolate
  topology, preset, and population decisions, while canonical serialization and
  checksums make generation replay-verifiable across server and client.
- [x] M20.7: Prove one Act I preset level through typed Levels, LvlPrest, and
  LvlTypes records. The generator validates DRLG type and act, selects a seeded
  DS1 variant, applies the authored DT1 mask, and emits the immutable zone
  recipe and decision trace. `mapgen_lab` exposes that result through pointer
  seed controls without forcing expensive renderer materialization; DS1 Lab
  remains the separate lazy consumer of emitted recipes.
- [x] M20.8: Generate the first deterministic LvlMaze room topology for Act I
  Cave Level 1, including difficulty room counts, orthogonal growth, authored
  merge probability, normalized bounds, connectivity, and stable checksums.
  Every ordinary room's W/E/S/N mask selects the verified contiguous Act I Cave
  LvlPrest definition (53..67) and a purpose-stream variant. Mapgen Lab draws
  the topology without decoding assets. Other LevelType chamber sets remain
  explicit follow-up asset-interpretation work. Dark Magic intentionally does
  not target legacy seed compatibility or reproduction of original layouts.
- [x] M20.9: Reserve deterministic maze endpoints for previous-level and
  next-level roles, keep merge loops from consuming those one-door leaves, and
  replace ordinary rooms with the direction-compatible Act I Cave Prev/Next
  LvlPrest definitions 83..90. Mapgen Lab distinguishes both roles while the
  immutable zone recipe exposes them to future warp materialization.
- [x] M20.10: Materialize generated zones incrementally, one authored DS1 stamp
  per cancellable step, while caching immutable DT1 catalogs shared by rooms.
  The completed authoritative map is published atomically and contains no
  images or native renderer resources. Composition clips the shared terminal
  DS1 row and column to each generated room's logical LvlPrest footprint,
  preventing neighboring collision, object, and tile facts from overlapping.
- [x] M20.11: Replace scene-local coordinate arithmetic with shared tile,
  subtile, world-pixel, screen-pixel, and camera transforms. Derive collision
  queries and the F3 diagnostic overlay from the same transform contract. The
  world capability now exposes forward and inverse camera projection for future
  pointer intents; gameplay movement uses the same centered-cell sampling rule
  as authoritative map queries.
- [x] M20.12: Sort map passes, authored objects, characters, missiles, shadows,
  roofs, and overlays through one explicit world-depth contract. Floor remains
  a globally batchable background; occluding map tiles are partitioned by their
  projected baseline and share a retained parent with the player. The contract
  supplies stable pass offsets around the same `x+y` entity baseline that future
  objects, missiles, and world overlays consume.
- [x] M20.13: Route pointer-authored ground movement through the shared inverse
  camera transform and fixed-tick authoritative command stream. Held pointer
  targets remain semantic subtiles, are replayable, derive velocity/facing from
  live ECS position, and continue through the existing authoritative collision
  integrator. HUD space, obscured overlay halves, and held-item input do not
  leak movement.
- [x] M20.14: Materialize selectable world entities and resolve pointer targets
  into move, interact, primary-skill, or secondary-skill commands by authority,
  range, collision/path, and line-of-sight policy. Do not fake interaction by
  treating every world click as non-spatial UI.
  - [x] M20.14a: Turn resolved authored DS1 objects into stable selectable
    footprints, index them spatially, resolve overlapping hits deterministically,
    and admit coordinate-authored interactions only after authoritative range
    and DT1 line-of-sight checks. Primary clicks approach a selected object and
    stop before requesting interaction; Lua never chooses the trusted target ID.
  - [x] M20.14b: Normalize the secondary mouse button and dispatch its world
    target as a replayable assigned-skill command without coupling the device to
    combat rules. Authority resolves the currently assigned, learned, side-
    allowed skill and records its semantic coordinate/optional entity target in
    ECS state. Empty skill wells safely no-op; invalid non-empty assignments are
    rejected. Hostile/item/portal classification remains M20.14c input.
  - [x] M20.14c: Replace direct approach with deterministic entity-size-aware
    eight-way A*, diagonal corner protection, occupied-target stopping radii,
    unreachable-target cancellation, and hover presentation sourced from the
    same selector. Fixed-tick commands record successive authoritative
    waypoints, while DT1 collision remains the movement and path source.
  - [x] M20.14d: Define authoritative spawned-entity components for player,
    hostile, NPC, item, portal, missile, and scenery kinds; materialize players
    into the contract; resolve overlaps by copied ECS facts; filter the local
    player; and dispatch hostile primary clicks to assigned primary-skill intent.
    Future monster, drop, portal, and missile spawners attach these components;
    no hostility is inferred from a DS1 number or display string.
- [ ] M20.15: Generate and enter Act I Rogue Encampment as the first production
  session-owned zone.
  - [x] M20.15a: Correct static `LvlPrest.Files=0` alternative handling, select
    all four `TownN1/E1/S1/W1` layouts deterministically, retain their cardinal
    wilderness-exit role, and derive an open player-entry point near the
    authored RogueBonfire instead of the map center.
  - [x] M20.15b: Replace the fixed presentation fixture with the generated town
    recipe, materialize it incrementally, and publish it atomically to the
    authoritative session and chunked renderer.
  - [x] M20.15c: Route character creation, game join, and same-act game entry to
    the town campfire anchor; preserve character/session identity independently
    of the selected town layout.
    The entry command now records the server-selected act and level in an
    authoritative `dm.world.location` component; creation and offline join use
    the generated Act I town and campfire anchor. The production-asset client
    seam now proves a frontend-created/selected character is admitted by the
    fixed-tick session at that exact anchor with Act 1 / level 1 authority.
    Offline selection and remote/realm join admission now share one validated
    server-selected destination and command builder. Clients cannot choose the
    spawn or mint the system/admin-authority admission command.
  - [ ] M20.15d: Join the selected cardinal town edge to the first generated
    wilderness zone and validate bidirectional transition, collision, camera,
    depth, and pointer navigation across the seam.
    The first Blood Moor strategy now builds its production 80x80 extent as a
    deterministic 10x10 graph of authored 8x8 outdoor stamps and carries an
    explicit town-entry warp on the edge opposite the selected town exit.
    Structural rivers/cliffs/bridges and presentation seam validation remain
    before this checkpoint is complete.
    Hidden DS1 orientation-10/11 exit cells are now retained as semantic-only
    tile placements, and a production-asset seam contract verifies the selected
    town anchor meets the generated opposite Blood Moor edge in walkable
    subtile coordinates without asking presentation to infer either endpoint.
    A trusted fixed-tick `system.world.transition` command now requires the
    player's authoritative source level and proximity to that seam, then
    atomically updates destination level, inset arrival position, world bounds,
    and velocity. The local source detects the crossed endpoint without trusting
    Lua and arrival inset prevents immediate bounce-back. The client now swaps
    authoritative navigation, collision, spatial interaction selection, Lua
    collision, camera projection facts, and sparse assembled-world rendering
    when that command commits. Structural outdoor rules and an asset-backed
    interactive seam capture remain before this checkpoint is complete.
    The outdoor recipe also reserves a deterministic contiguous coarse-cell
    route between explicit town-entry and next-level edge anchors. The route is
    rasterized into validated tile-resolution `PathTile` facts in the immutable
    zone checksum. Materialization applies D2MOO's verified 3x3-neighbor
    floor-identity pass and rebuilds collision from the selected DT1 metadata.
    Structural rivers/cliffs/bridges and an asset-backed interactive seam
    capture remain before this checkpoint is complete.

## M21: Diablo simulation

M21 now starts with the playable-character vertical slice. Spatial NPC service
work proved the command/authority boundary, but further interactions are ordered
after one controlled character can enter a map, receive routed input, advance at
the fixed tick, and drive a real legacy composite from authoritative state.

Gameplay input follows the legacy pointer-first contract: the mouse targets
movement, interaction, item use, and combat, while keyboard events provide
hotkeys, text, modifiers, and cancellation. Modern controller navigation and
world targeting are future adapters over those semantic intents; controller
focus must not replace pointer coordinates as the authoritative client request.

- [ ] **M21.0: Complete the authoritative playable-character vertical slice.**
  The first increment now materializes separate composite appearance and
  animation components, maps normalized movement commands to authoritative
  NU/WL/RN modes and legacy eight-way facing, projects the ECS position through
  the map camera, and resolves an unarmed HTH COF plus default body DCC layers.
  Complete since that increment: composed direction/mode resources are cached;
  direction and equipment changes retain animation progress; typed
  `AnimData.d2` records supply rates and frame events; crossed events propagate
  through the presentation controller; COF shadow flags produce a distinct
  shadow pass; and authoritative equipment slots select component appearance
  codes and active-hand weapon classes. Semantic movement facing now follows the
  separate legacy logical-to-COF and COF-to-DCC direction tables corroborated by
  Riiablo and OpenDiablo2. Shadow-enabled body parts are first assembled into one
  shared alpha mask, then projected leftward at half height from one baseline so
  limbs cannot produce detached, full-height component shadows. The item
  snapshot carries only validated presentation recipes, never renderer paths.
  Diagonal velocity is normalized
  to the same walk/run speed as cardinal movement, and an authoritative
  `dm.world.collider` radius now probes the player's full footprint while
  retaining axis-separated wall sliding. COF shadow-enabled components now use
  the baseline-anchored half-height/shear projection corroborated by Riiablo and
  OpenDiablo2, with center-preserving animation canvases and exact pixel probes.
  Remaining acceptance is native live-input/capture coverage, broader
  collision/pathing refinement, richer imported equipment recipes under M19.4, and enough
  play testing to make the slice comfortable and visually stable. Continue
  NPC/quest/vendor interaction breadth only after that point.
  `make play-game-world` and `make capture-game-world` now provide the focused
  asset-backed manual/play and settled-frame acceptance entry points with a
  deterministic selected character; automated command-to-ECS input coverage
  remains in the embedded shell and runtime integration suites.
  Offline entry now derives authoritative bounds and a deterministic open spawn
  subtile from the manifest-selected DS1/DT1 map. The retired 4096x4096
  placeholder bounds can no longer place the camera thousands of subtiles away
  from every demand-resident map chunk.

- [ ] Implement the complete stat/value system, ItemStatCost encoding, derived
  stats, caps, per-level values, state modifiers, and description functions.
- [ ] Implement player and monster modes, animation events, targeting, AI,
  spawning, packs, bosses, pets, hirelings, corpses, and death/respawn.
- [ ] Materialize player and monster ECS archetypes from validated typed records,
  with character creation expressed as authoritative simulation commands rather
  than a stateful character/monster generator service. The privileged,
  replayable `system.player.enter` command now atomically creates stable player
  identity, progression, vitals, appearance, ownership, transform, velocity,
  and bounds components; typed starting-stat resolution, scene handoff, monsters,
  and persistence synchronization remain.
- [ ] Implement attacks, hit checks, block, defense, damage types, resistance,
  absorb, leech, regeneration, durability, experience, leveling, and difficulty.
- [ ] Implement skills, skill trees, missiles, auras, states, procs, charges,
  cooldowns, mana, targeting shapes, and server-authoritative validation.
- [ ] Implement NPC interaction, vendors, gambling, repair, identify, cube,
  shrines, quests, waypoints, acts, difficulties, and end-game transitions.
  The first NPC interaction boundary is now fixed-tick and replayable: clients
  request a known target ID, while session authority resolves the NPC, vendor,
  allowed catalog categories, and service names. Item commerce and services
  reject requests outside that active context, and Lua receives only copied
  facts plus open/close intent functions. The Akara development fixture now
  drives a real MPQ-backed vendor catalog panel with the four legacy category tabs,
  Monster `Inventory.txt` grid geometry, authority-arranged stock, carried item
  cursor state, gold display, and command-backed purchases. Riiablo's recovered
  tab frames and `gridLeft - invLeft` placement corroborate the implementation.
  Interaction opens now resolve the controlled player from authoritative ECS
  components and reject unknown, missing, or out-of-range targets using the
  server-owned target anchor and radius. Every commerce/service command repeats
  that current-position check, so walking away immediately expires permission
  even if presentation has not closed the panel. Zone-aware and moving NPC
  entity targets, record-built target catalogs, sell/repair modes, price
  tooltips, gambling refresh, and the remaining NPC services stay open.
- [ ] Integrate deterministic item materialization and treasure-class rolls with
  authoritative monster/chest events, world drops, inventories, equipment, and
  persistence without duplicating the completed `internal/game/loot` rules.
- [ ] Implement record-driven music, ambience, speech, monster, object, item, UI,
  and combat sound selection with deterministic event identity where required.

## M22: Client and game-session networking

M22 has a working deterministic session/replay foundation; transport, remote
authority, persistence separation, and resilience acceptance remain open.

- [x] Define deterministic simulation snapshots, purpose-named RNG streams,
  admitted command logs, canonical per-tick execution, checksums, restoration,
  replay verification, and first-field desync diagnostics.
- [ ] Add bounded, versioned on-disk replay containers and migration policy for
  manifests, initial snapshots, admitted commands, events, and checkpoints.
- [x] Keep `cmd/darkmagic` as the first-class client composition root and move
  its offline fixed-step ECS advancement behind the shared authoritative
  session owner without coupling that owner to Raylib or Lua.
- [ ] Run one authoritative deterministic game-session implementation in-process,
  as a listen server, as a standalone self-hosted server, or under a realm.
  The `cmd/darkmagic-server` composition root and shared administration shell
  now exist. `internal/game/session` now owns transport-independent admission,
  fixed stepping, canonical command ordering, checkpointing, and replay export;
  the client uses it for offline stepping and the standalone server owns its
  timed lifecycle plus read-only `dm.session/v1` diagnostics. Listen-server,
  realm orchestration, gameplay-system composition, and transports remain.
- [ ] Implement discovery/direct connect, authentication boundaries, snapshot
  transfer, command replication, rollback or correction, and reconnect.
- [ ] Preserve offline characters safely and separate trusted server characters
  from client-controlled saves.
- [ ] Add long-running soak, malformed-data, fuzz, latency/loss, save round-trip,
  performance, and race tests across supported platforms.

## M23: Realm, mod delivery, and trusted persistence

- [ ] Add a first-class `cmd` realm providing accounts, authentication, game
  creation/discovery, session assignment, and operational administration. The
  `cmd/darkmagic-realm` composition root and shared administration shell now
  exist; realm services and capability modules remain.
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
  The Raylib adapter now wraps contiguous engine RGBA storage directly instead
  of using raylib-go's per-pixel cgo converter, and releases fallback conversion
  images immediately after upload.
- [x] Evaluate indexed GPU textures with palette lookup for DC6/DCC assets after
  the simpler lifetime and upload improvements are measured.
- [x] Stream long-lived frontend music when doing so reduces residency without
  compromising gapless looping, synchronization, or shutdown behavior.
- [x] Add repeatable profiling acceptance runs and budgets for startup, title,
  main menu, character selection, character creation, and in-game scenes.
- [x] Add residency-aware capture readiness and warm-queue backpressure. The
  vendor probe observed a burst of roughly 1,000 resident entries—not a hard
  entry ceiling—while hundreds of startup warm requests remained queued.
  Captures now restart their settle window when retained node topology changes,
  so late-created UI pieces cannot produce a partial capture while animation,
  streaming textures, and refreshed text remain eligible to settle.
  Optional warming runs after presentation and is admitted only while it fits
  without eviction; demand uploads remain unrestricted and therefore retain
  priority over speculative residency.
- [x] Decode only the requested DCC direction, composite palette indexes directly
  into the final RGBA surface, give known frames generation-qualified semantic
  texture identities, split encoded/direction/composed CPU cache budgets, and
  report cold work by stage. The real unarmed-player fixture fell from roughly
  1.4 seconds to about 0.7 seconds on the development machine.
- [x] Give generated-world rasters an independent 96 MiB working-set cache and
  keep whole-zone RGBA out of preload residency. Production Blood Moor now
  indexes without expanded pixels, while a bounded camera-near sample stays
  below 16 MiB and cold chunk rasterization remains off the Lua update thread.

## M26: Native frame-path profiling follow-up

M26 is independent performance/lifecycle debt. Its shutdown-order checkpoint is
safety-sensitive; the remaining upload and accounting work should stay in
separate profiling PRs unless a measured gameplay budget requires it sooner.

- [ ] Upload contiguous FFmpeg `image.NRGBA` frames directly when their byte
  layout is already compatible with raylib, retaining conversion for genuinely
  incompatible color models and padded subimages.
- [ ] Measure cinematic upload bandwidth and distinguish frame-update traffic
  from resident GPU texture bytes in scene diagnostics and budgets.
- [x] Drain final composition commands on the renderer owner thread, release all
  backend nodes and palette effects, clear GPU caches, and only then close the
  native renderer. Architecture tests reject restoration of the legacy renderer
  object API and direct-native world adapter.
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

M28's structural package migration is complete. The milestone remains open only
for the remaining typed-record admissions/consumers and final migration
acceptance/documentation cleanup; it is no longer a general refactor mandate.

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
  now live under `internal/platform/raylib`, keeping window, input, renderer, and
  audio-device code visibly outside the engine contracts; previous adapter paths
  are guarded. The unused transitional world adapter and exported native-object
  scene graph are retired, and an acceptance guard prevents their revival.
  Repository-private CLIs
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
  an ordered typed `SoundRecords` view from the client-owned catalog instead of
  reparsing raw TSV column maps or caching a second Sounds generation. The view
  decodes only its admitted table so partial-content tools remain useful, while
  hot-reload invalidation still reaches typed records, generic Lua rows, and audio
  through the same source generation.
  The Lua records capability now uses the typed catalog as its generic-row
  gateway, so script-requested reloads also drop the typed snapshot rather than
  bypassing catalog invalidation through the raw store. The versioned
  `dm.game_data/v1` capability now supplies copied, typed character starting
  attributes and ordered unique-title fragments without exposing Go records or
  coupling authored scripts to arbitrary TSV columns.
  The Lua loot capability likewise consumes typed `TreasureClassEx` records and
  exposes `roll(class, seed)`; arbitrary TSV paths and duplicate parsing no
  longer cross the runtime boundary, while deterministic event seeds and roll
  results retain their existing simulation-owned representation.
- [x] Add package documentation and a concise newcomer architecture guide showing
  the boot path, frame path, scene/mod boundary, asset path, and where new code of
  each kind belongs. Keep examples aligned with the resulting structure.
- [x] Enforce the intended architecture with dependency tests or static checks,
  package-level tests at moved boundaries, and CI checks that reject new accidental
  public packages, forbidden imports, compatibility-service resurrection, and
  business logic added directly to `cmd`. Acceptance checks now enforce inward
  dependency direction for foundations, game, presentation, and platform layers;
  prevent production dependencies on developer tooling; freeze the reviewed
  command composition helpers; and run as an explicit CI stage.
- [ ] Complete the migration as behavior-preserving commits with the full unit,
  race, real-asset, and interactive acceptance suites green; then remove temporary
  migration notes and verify a newcomer can locate and extend a representative
  renderer, asset, Lua capability, scene, tool, and server feature from the guide.

## M29: Scriptable archetype ECS

M29's ECS/runtime foundation is complete. The remaining work is integration:
recovered executable datasets, command-backed gameplay actions, listen-server
replay verification, and narrow audited administrator handlers.

- [x] Redesign Akara around typed stores, interned archetypes, cached queries,
  deterministic systems, and explicit lifecycle rather than background ticks.
- [x] Add VM-neutral runtime component schemas with validated primitive fields,
  stable names, versions, defaults, and generation-checked references.
- [x] Add transactional schema migration that preserves the previous world when
  any entity conversion fails.
- [x] Add deterministic Dark Magic phases, dependency ordering, stable query
  snapshots, declared read/write access, and structural command barriers.
- [x] Expose checked entities and components through `dm.ecs/v1`, including
  Lua-defined schemas, systems, queries, field mutation, and deferred spawning,
  addition, removal, and destruction.
- [x] Tie Lua system registrations to disposable component scopes and prove that
  teardown removes systems without destroying persistent world state.
- [x] Advance gameplay at a bounded fixed 25 Hz clock independent of renderer
  frame duration, with stable tick and delta values exposed to Lua systems.
- [x] Replace the production compatibility hero/world state with Lua-defined ECS
  components and systems for input intent, transform, bounded movement, camera
  follow, and presentation snapshots. The shim now keeps related component
  schemas and individual systems in documented modules; `world.lua` only
  registers them and owns the narrow scene-facing binding/snapshot facade.
- [x] Admit a renderer-matched, round-tripped isometric pixel/subtile transform,
  then use axis-separated DT1 collision in ECS movement without mixing in
  presentation pixels.
- [x] Preserve and validate Riiablo's recovered quest and speech tables in the
  shim, including quest prerequisites, stage localization keys, and logical
  sound-to-localization joins exposed through `dm.quest_catalog/v1`.
- [x] Resolve DS1 static objects through recovered act-local `obj.txt` mappings
  and dynamic objects through act-local `MonPreset.txt` ordering before exposing
  immutable world records to Lua.
- [ ] Import the remaining executable-derived Riiablo datasets as they are
  identified, with provenance, typed schemas, cross-table validation, and
  representative quest/dialogue/audio behavior tests.
- [x] Add purpose-named deterministic RNG streams plus canonical, versioned ECS
  snapshot serialization and replay checksums with typed float-bit encoding.
- [x] Define transport-neutral replay commands and checkpoints, exact snapshot
  restoration, first-field desync diagnostics, and transactional authoritative
  command admission by tick, player sequence, kind, and payload policy.
- [x] Route normalized player movement through one validated `player.move`
  command per active fixed tick; apply velocity through the authoritative
  session before Lua ECS movement/collision and retain it in replay logs.
- [ ] Route skills, interactions, inventory, UI-confirmed gameplay actions, and
  resulting events through admitted commands in the shared session; then add
  listen-server mode and verify exported client/server replays.
- [x] Distinguish player, administrator, and system command authority at handler
  registration; reject privilege escalation and retain executed privileged
  commands as a defensive, replay-correlated audit exposed by `dm.session/v1`.
- [ ] Add explicit audited administrator handlers for item/loot, monster, world,
  character, and session operations as their authoritative ECS archetypes and
  validation rules become available; never add arbitrary component mutation.

## M30: Deferred shell ergonomics

- [ ] Support user-selected font families and safely rebuild native glyph atlases
  when font path, size, DPI, or Unicode coverage changes.
- [ ] Add persistent named color themes shared by the Raylib and Charm adapters,
  including accessible high-contrast and color-vision-friendly presets.
- [ ] Add mouse selection, clipboard copy/paste, searchable command history,
  reverse search, transcript search, and explicit transcript/API-doc export.
- [ ] Render structured help and Markdown with richer code-block, signature,
  table, link, and source-location presentation without changing evaluator text.
- [ ] Add explicit multi-runtime tabs with visible scope/capability identity,
  per-tab history and settings, reset/reload lifecycle controls, and resource
  teardown guarantees across scene changes, disconnects, and runtime restarts.
- [ ] Add optional detachable/remote administration frontends over the same
  policy-aware shell session contract without exposing server authority to a
  realm-connected gameplay client.

## M31-M43: Creature authoring, animation, and generated representations

This is a dependency-aware program rather than one undifferentiated "3D
support" milestone. The canonical plan, package ownership, diagrams, synthetic
fixtures, risks, prototype gates, Blender migration findings, and objective
acceptance criteria are in
[docs/CREATURE_ASSET_ROADMAP.md](docs/CREATURE_ASSET_ROADMAP.md).

- [ ] **M31 — evidence, boundaries, and executable schemas:** reconcile the
  existing DCC/DC6/COF/PL2 research and historical Blender 2.80 addon, publish
  versioned semantic schemas, and make `dm-asset validate/inspect --json` useful.
- [ ] **M32 — glTF import and golden poses:** import a redistributable humanoid
  and non-humanoid GLB, including a measured Dust3D-authored fixture, without
  making glTF or any authoring tool the runtime creature abstraction.
- [ ] **M33 — creature laboratory and reference 3D playback:** load, play, pause,
  scrub, inspect bones, and visualize bounds/sockets through a backend-neutral
  creature runtime with an initial Raylib adapter.
- [ ] **M34 — components, sockets, variants, and equipment:** share one skeleton
  and primary clock across arbitrary components; preserve normalized animation
  time across visual swaps.
- [ ] **M35 — semantic animation:** add time-based clips, skipped-interval-safe
  typed events, facing, root-motion policy, masks/layers extension points, and a
  deliberately small data-driven state model.
- [ ] **M36 — modern Blender addon:** provide supported-version authoring,
  validation, preview, and package export as a frontend over shared contracts;
  do not embed codecs or make `.blend` files runtime assets.
- [ ] **M37 — materials and palette parameters:** validate instance-local tint,
  alpha, emissive, region/mask, palette, and PL2-compatible transform behavior
  after prototype evidence selects the first 3D palette technique.
- [ ] **M38 — deterministic directional sprite generation:** sample an explicit
  timeline/camera/light profile into aligned RGBA component and flattened
  frames, preserving common origin, bounds, events, and reproducibility data.
- [ ] **M39 — modern sprite composition:** load a human-readable semantic
  composite representation with arbitrary layers, directions, timing, ordering,
  palette parameters, attachments, and representation comparison in the lab.
- [ ] **M40 — quantization and legacy exporters:** make palette conversion an
  explicit stage and export representable DC6/DCC/COF through independent codec
  APIs, with documented loss reports and no private Blender codecs.
- [ ] **M41 — retargeting:** validate semantic skeleton maps, rest-pose and root
  conversion, proportions, missing bones, masks, and actionable incompatibility
  diagnostics without requiring one universal or humanoid rig.
- [ ] **M42 — build graph, cache, and live iteration:** add incremental,
  content-addressed builds, file invalidation, on-demand development rendering,
  hot reload, and deterministic Blender-to-engine CI metadata.
- [ ] **M43 — measured advanced runtime:** profile before admitting GPU skinning,
  instancing, LOD/alternate representations, advanced animation, or parallel
  generation; production cache-miss generation remains separately gated.

The smallest convincing vertical proof spans M31-M38: a synthetic GLB with a
skeleton, body, replaceable weapon, idle/attack clips, socket, and `attack.hit`
event plays in the creature lab, swaps the weapon at 42% normalized time without
restarting, and emits aligned eight-direction sprites carrying the same event.

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
  DS1 Lab and `game_world` now share sparse, layered CPU composition and
  demand-resident retained nodes. The full-image composer remains only for the
  command-line diagnostic PNG exporter.
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
- [x] Add incremental reader/writer APIs to sequential codecs and lazy
  `io.ReaderAt` views to DC6, DCC, DT1, and TBL without removing compatibility
  APIs.
- [x] Make MPQ archive and entry access positional and concurrency-safe, verify
  it under the race detector against an owned real archive, and adopt the new
  APIs at Dark Magic's content/decode boundary.
- [x] Retain lazy codec files at scene/world boundaries so DC6 frame requests
  decode one frame, DC6/DCC animations decode one requested direction, world
  collision consumes metadata-only DT1 tiles, and DS1 previews decode block
  pixels only for keys used by the stamp. Encoded, directional, and composed
  residency is measured in separate M25 cache tiers.
- [x] Provide integrated `dt1_lab` and `ds1_lab` scenes using the production Lua
  render capability. DT1 Lab browses lazy tiles and metadata across composite,
  floor, and wall views; DS1 Lab renders configured stamps with palette-aware
  DT1 sources plus fit, zoom, pan, CLI recipes, and capture registration. Both
  labs can deterministically select mounted assets and cycle Act palettes from
  hotkeys without restarting the client.

See [CODECS.md](CODECS.md) for the module-by-module plan. Codec repositories
remain independent rather than being copied into this engine.

## Current handoff

The architectural acceptance path and M0–M14, M17, M25, and the foundational
portion of M29 are implemented. M15 and M16 retain bounded catalog/composition
gaps; M18–M24, M26–M27, the typed-data tail of M28, and M29's gameplay-command
integration remain open as described above. M31–M43 now define the independent
creature-authoring and generated-representation program; none of that program is
claimed implemented. The repository builds against
tagged codec releases, boots through the internal host and layered shim, runs
the Lua-authored shell and world orchestration, and passes the complete package
suite under race detection. Historical stashes remain preserved and documented
in `STASHES.md`; do not apply or drop them without a separate comparison task.
