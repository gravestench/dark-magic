<h1 align="center">Dark Magic</h1>
<h3 align="center">An Open-Source Diablo 2 Engine Rewrite</h3>

The maintenance policy for the standalone OpenDiablo2-derived format
libraries is documented in [CODECS.md](CODECS.md).

## About

Dark Magic is a clean-room, open-source Diablo II engine rewrite. It aims to
preserve the original game's authored mechanics and presentation while replacing
the abandoned OpenDiablo2 architecture with deterministic simulation,
data-driven rules, scriptable gameplay, and explicit native-resource ownership.

Dark Magic does not distribute Blizzard assets. Running the client or real-asset
tools requires a legally obtained Diablo II installation or MPQ set.

## Research and community acknowledgements

Dark Magic exists because decades of Diablo II modders, preservationists,
reverse engineers, tool authors, and independent engine developers documented
behavior that the original executable kept implicit. Thank you to the entire
Diablo II community—and especially to the people who published their findings,
test cases, tools, and source so later projects could verify rather than guess.

Our principal research and corroboration sources include:

* [Paul Siramy's historical research and tools](http://paul.siramy.free.fr/)
  and the wider [Phrozen Keep community](https://d2mods.info/), particularly for
  DS1, DT1, map editing/generation, animation formats, and engine behavior.
* [ThePhrozenKeep/D2MOO](https://github.com/ThePhrozenKeep/D2MOO), the pinned
  reverse-engineered Diablo II 1.10f runtime baseline used throughout the
  gameplay and map research.
* [jaenster/libd2](https://github.com/jaenster/libd2), both for independent
  1.14d implementation evidence and its unusually strong retail-capture,
  holdout, save-round-trip, item-roll, and pathfinding verification discipline.
* [Riiablo](https://github.com/collinsmith/riiablo), including its format,
  composite, rendering, UI, and recovered declarative quest/dialogue/object
  work.
* [OpenDiablo2](https://github.com/OpenDiablo2/OpenDiablo2), from which this
  project descends, and [OpenD2](https://github.com/eezstreet/OpenD2) and
  [AbyssEngine](https://github.com/AbyssEngine/AbyssEngine) as independent
  behavioral and presentation references.
* [nokka/d2s](https://github.com/nokka/d2s) for independent save-format
  documentation, the historical Diablo II Data File Guide bundled for research,
  and the many community-authored manuals, patch notes, forum posts, and tools
  cited in the individual research ledgers.

These projects are evidence, not code-generation templates. Dark Magic restates
behavior independently, respects source licenses, labels version conflicts and
uncertainty, and prefers repeatable observations of lawfully owned game data.
The detailed provenance, inspected versions, confidence rules, and source paths
live in the [format source matrix](docs/formats/SOURCE_MATRIX.md) and
[gameplay-systems source matrix](docs/research/SYSTEMS_SOURCE_MATRIX.md). The
[research index](docs/research/GAME_SYSTEMS_INDEX.md) distinguishes a documented
baseline from behavior that has actually been validated.

## Directory structure

* `cmd` contains the client, game-session server, and realm composition roots.
  Commands perform wiring and process configuration, not gameplay.
* `internal/game` contains generic deterministic ECS/session/world mechanisms
  and typed Diablo record boundaries. Production Diablo gameplay policy lives
  in the first-party `d2legacy` Lua mod, with narrow Go adapters under
  `internal/mod/d2legacy` where host integration is unavoidable.
* `internal/content` owns the layered directory/MPQ/ZIP VFS and redistributable
  first-party `d2legacy` Lua mod.
* `internal/runtime/lua` adapts explicit, versioned capabilities into serialized
  Lua runtimes with disposable resource scopes.
* `internal/presentation` defines backend-neutral retained rendering and scene
  navigation; `internal/platform/raylib` owns native Raylib integration.
* `internal/dev/tools` and `internal/dev/testapps` contain repository-private
  diagnostics and harnesses. They are not product binaries.

There is intentionally no public Go API under `pkg`. Standalone Diablo file
codecs remain independently versioned repositories; see [CODECS.md](CODECS.md).

## Architecture

The executable uses an explicit internal application host for native lifecycle,
a dynamic manager for native and Lua-defined components, a layered content VFS,
and versioned Lua capabilities. The Raylib backend remains isolated beneath
`internal/platform/raylib`; scripts author scenes and overlays through retained rendering,
input, audio, records, locale, save, simulation, and navigation capabilities.
The `engine.ecs/v1` capability additionally lets trusted scripts define validated
component schemas and deterministic, scope-owned systems over the shared Akara
world. Systems declare their query and read/write access; structural mutations
are applied at phase barriers rather than during query iteration.
The `d2legacy` mod also preserves Riiablo's recovered quest hierarchy,
speech-to-string relationships, DS1 definitions, and act-local object mappings.
Go validates these immutable recovered catalogs; narrow `d2legacy` adapters
expose their facts to Lua, which alone decides their gameplay meaning.

### Product binaries

Dark Magic targets two primary executables under `cmd`:

* **Client** — runs presentation and input for offline single-player,
  self-hosted multiplayer, and realm-connected games. Offline play may host the
  same authoritative simulation in-process, but client code must not become the
  authority for a remote realm character.
* **Realm** — provides the Battle.net-like control plane: authentication,
  accounts, trusted character persistence, game creation and discovery, and
  assignment to authoritative game sessions. It publishes the exact mod
  manifest required by a game and serves or identifies the corresponding
  redistributable payloads. User-supplied Diablo II data is never distributed.

The deterministic game-session server is shared infrastructure. It can run
in-process for offline play, as a client-created listen server, as a standalone
self-hosted server, or under realm orchestration. These modes must share
simulation commands, snapshots, validation, and persistence contracts rather
than grow separate gameplay implementations.

The client also includes an in-game Lua console. It targets an explicitly
selected Lua runtime rather than a process-global VM. Console evaluation is
serialized through that runtime's owner, uses a dedicated disposable resource
scope, and sees only the capabilities permitted for that runtime. Offline and
development sessions may enable mutation/debug capabilities; realm-connected
clients cannot use the console to bypass authoritative simulation or obtain
server-only capabilities.

Press the grave/backtick key (`` ` ``) to open or close the in-game shell.
F1 selects the Lua editor and F2 selects structured application logs. In Lua,
Enter evaluates, Shift+Enter inserts a newline, Tab completes names, and the
arrow/Home/End keys edit or browse history. In Logs, arrows and Page Up/Down
control independent scrollback.
Lua `print(...)`, `printregs()`/`_printregs()`, and the bounded structured
application-log tail appear in their respective modal views; normal process
log output remains available outside the game window as well. Every Lua view
opens with a target- and policy-specific message of the day. The `d2legacy`
root provides discoverable, policy-filtered capability access: use
`d2legacy.help()` and `d2legacy.capabilities()`, friendly names such as
`d2legacy.app`, or `d2legacy.modules["engine.app/v1"]` for an exact versioned module ID.
Pass a module, command, or path to help for progressively more detail—for
example `d2legacy.help(d2legacy.audio)`, `d2legacy.help(engine.audio.play)`, or
`d2legacy.help("engine.audio.play")`. Lua module registrations own the summaries, usage,
parameters, returns, and examples used by both help and completion. Existing
commands without authored metadata are still listed with fallback help.

Shell presentation settings are live Lua runtime values. For example:

```lua
engine.shell.values()                  -- current native-resolution shell settings
engine.shell.set("font_size", 22)     -- apply immediately for this process
engine.shell.set_many({                -- validate and apply atomically
    console_height = 0.7,
    opacity = 0.85,
    transcript_limit = 4000,
    animation_speed = 1.5,
})
engine.shell.defaults()                -- inspect built-in values
engine.shell.reset()                   -- restore defaults in memory
engine.shell.reload()                  -- discard edits and reload the saved file
engine.shell.save()                    -- persist the active values
engine.shell.status()                  -- persistence path and dirty state
```

Settings default to `shell.json` under the platform user-configuration
directory. `DARK_MAGIC_SHELL_CONFIG` selects another host path and supports
home-directory aliases. Multiline Lua values retain line breaks, indentation,
and tabular spacing in both graphical and terminal shell views.

Game preferences are separate from developer-shell presentation. The authored
in-game sound and music sliders update mixer buses immediately through
`engine.settings/v1` and save to `preferences.json` under the platform
user-configuration directory when the overlay closes. Set
`DARK_MAGIC_PREFERENCES` to use another file.

Player-profile characters used by single-player, listen-server, and self-hosted
dedicated-server play default to `player-profile.json` under the platform
user-configuration directory. Set `DARK_MAGIC_PLAYER_PROFILE` to use another
file. A missing file starts with an empty roster and is created on clean
shutdown. Development character fixtures are always ephemeral and cannot
overwrite this file. Realm characters remain account-owned and never load from
this player-controlled profile.

Renderer residency diagnostics use the same persistent preference path and can
be controlled directly from the Lua shell:

```lua
engine.settings.set("debug_texture_residency", true) -- native-resolution cache overlay
engine.settings.set("texture_upload_budget_mb", 16)  -- optional warm uploads per frame
engine.settings.set("texture_cache_budget_mb", 512)   -- retained native texture capacity
engine.settings.set("camera_follow_strategy", "instant") -- default: no smoothing
engine.settings.set("camera_follow_duration", 0.0)        -- tween seconds per target update
engine.settings.set("camera_follow_param_1", 0.0)         -- strategy-specific value
engine.settings.save()                               -- retain these across launches

engine.render.diagnostics() -- decoded cache plus pending CPU/GPU warm work
```

Camera follow strategies are `instant`, `linear`, `quad_in`, `quad_out`,
`quad_in_out`, `cubic_in`, `cubic_out`, `cubic_in_out`, `exponential_out`, and
`back_out`. `exponential_out` uses parameter 1 as its exponent (default 10),
while `back_out` uses it as overshoot (default 1.70158). Parameters 2 and 3 are
persisted for strategies that need additional tuning values. Changes apply to
the live game-world camera and remain presentation-only.

Texture creation remains on the graphics-owner thread. Asset reads and bitmap
preparation run in bounded workers; immutable textures are then uploaded within
the configured frame budget and retained by content identity across scenes.

Use `d2legacy.apropos("music")` to search the permitted module and command
descriptions. `d2legacy.docs()` renders Markdown for the session's complete permitted
Lua API from the same registration metadata used by help and completion.
Built-in capability conformance tests reject public module functions that do
not provide an authored summary and usage signature. Audio, video, and render
userdata methods are documented and checked the same way. Lua tables render as
stable, indented structures with bounded depth and cycle detection. Page Up and
Page Down navigate retained Lua output in the graphical console; the terminal
viewport provides matching scrollback and basic semantic highlighting.

The game always renders into its configured logical-resolution target while
the console renders afterward in native window pixels. Consequently, resizing
the window never makes shell glyphs inherit the game's low-resolution scaling.
Game presentation defaults to aspect-preserving centered letterboxing:

```shell
go run -tags ffmpeg ./cmd/client --viewport-fit contain
```

Use `--viewport-fit stretch` to fill the entire window without preserving the
game aspect ratio. `DARK_MAGIC_VIEWPORT_FIT` provides the equivalent environment
setting. Mouse input is mapped back into logical game coordinates; input in
letterbox regions is excluded from the game viewport.

Use `--fullscreen` for a maximized borderless window that retains desktop
window semantics instead of switching the monitor into an exclusive video
mode. Set `DARK_MAGIC_FULLSCREEN=true` for the environment equivalent.

## Join the Quest

Are you ready to embark on a journey into the heart of darkness? Unite with 
fellow adventurers and follow development on the
[community Discord server](https://discord.gg/gT9vTKfV8G).

Gather your courage, for a new era of sanctuary awaits. 
Embrace the magic, rewrite the destiny - with Dark Magic.


## Project status

The project is under active architectural reconstruction and is not yet a
playable Diablo II replacement. The application host, layered content system,
retained renderer boundary, typed data catalog, deterministic loot foundation,
Lua-defined Akara ECS, DS1/DT1 world decoding, and developer inspection tools
are operational. The playable world presents DS1 maps through sparse,
camera-culled texture chunks whose placement follows authoritative ECS camera
coordinates; it does not allocate a full-map GPU texture. Player movement is
fixed-tick and command-driven, with normalized diagonal speed and a subtile
collision footprint rather than a presentation-sprite hit test. Full world generation, Diablo combat and progression,
authoritative networking, and end-to-end gameplay remain in progress.

See [ROADMAP.md](ROADMAP.md) for the canonical milestone backlog and
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for package ownership, dependency
direction, and guidance on where new work belongs.

For the focused playable-character acceptance loop, point `MPQ_DIRECTORY` at
legally obtained game data and run `make play-game-world`. The command selects
one deterministic character fixture and enters the real world scene; pointer
movement exercises screen-to-world targeting, fixed-tick authority, run/walk
mode, composite facing, camera chunks, and DT1 collision together. Keyboard
input remains available for hotkeys, modifiers, text, and cancellation.
`make capture-game-world` records a
fully settled initial frame under `CAPTURE_DIR` (default `captures/frontend`).
These targets intentionally do not bundle or guess a game-data location.

## How to contribute
We welcome contributions from developers, artists, designers, and Diablo 
enthusiasts of all levels of expertise. If you want to be part of this journey, 
check out our [CONTRIBUTING.md](https://github.com/gravestench/dark-magic/blob/main/CONTRIBUTING.md) guide to get started.

Follow the lifecycle and dependency rules in
[the architecture guide](docs/ARCHITECTURE.md) when adding a long-lived engine
component.

## Development

Run the complete test suite with:

```shell
make test
make test-race
```

Client logging defaults to `info`. Select `debug`, `info`, `warn`, or `error`
with either the command-line flag or its environment equivalent:

```shell
go run -tags ffmpeg ./cmd/client --log-level debug
DARK_MAGIC_LOG_LEVEL=warn go run -tags ffmpeg ./cmd/client
```

Quantize the final composed display with an optional GPU post-process. The
lookup cube emits only colors from the selected palette while preserving source
alpha:

```shell
go run -tags ffmpeg ./cmd/client \
  --output-palette data/global/Palette/fechar/pal.dat
```

The equivalent environment variable is `DARK_MAGIC_OUTPUT_PALETTE`.

Lua can apply the same effect to one retained texture or animation instead of
the whole screen:

```lua
node:set_palette_quantization("palettes/black-white.json")
node:clear_palette_quantization()
```

Palette files may be normal Diablo `pal.dat` files, GIMP `.gpl` files, or any
non-empty sequence of packed BGR triples. Tiny palettes are intentional: three
bytes force one color, while six bytes can define black and white. JSON palettes
are also supported as either `["#000000", "#ffffff"]` or
`{"colors":["#000000","#ffffff"]}`.

Capture CPU activity for the full client run and a live-heap snapshot at clean
shutdown with:

```shell
MPQ_DIRECTORY=/path/to/diablo-ii \
  go run -tags ffmpeg ./cmd/client --profile-dir ./profiles/menu-review
```

Exit the client normally after exercising the behavior of interest. The output
directory contains `cpu.pprof`, `heap.pprof`, `cpu.pdf`, and `heap.pdf`. Raw
profiles can also be opened interactively with `go tool pprof`. PDF generation
requires Graphviz (`dot`) on `PATH`; if it is unavailable, Dark Magic preserves
the raw profiles and reports the rendering error. `DARK_MAGIC_PROFILE_DIR`
provides the same opt-in configuration for launchers that cannot pass flags.
Profile output under `profiles/` is intentionally ignored by Git.

Add `--profile-scenes title,main_menu,character_create` (or
`DARK_MAGIC_PROFILE_SCENES`) to generate filtered CPU reports and retained-heap
snapshots for individual scenes. Use `all` to capture every visited scene:

```shell
go run -tags ffmpeg ./cmd/client \
  --profile-dir ./profiles/frontend-review \
  --profile-scenes all
```

Scene artifacts are written beneath `scenes/<scene-id>/`. `cpu.pdf` combines
all visits to that scene using pprof labels. Each visit gets a numbered
`heap-NNN.pprof` and `heap-NNN.pdf` snapshot taken immediately before the
scene's owned resources are released.

Every scene snapshot also receives `diagnostics-NNN.json`, containing decoded
cache residency and hit/miss/eviction counts, cumulative decode time, retained
render resources and estimated RGBA bytes, texture upload volume, and bounded
per-scene p50/p95/p99 frame-interval and update-work timing. A
repeatable frontend acceptance run is available with:

```shell
MPQ_DIRECTORY=/path/to/diablo-ii make profile
make profile-check
```

`make profile-acceptance` runs every budgeted scene deterministically and exits
through capture completion. See [docs/PERFORMANCE.md](docs/PERFORMANCE.md) for
the measured baseline, benchmark commands, and interpretation rules.

`profile-check` requires every scene listed in `docs/profile-budgets.json` to
have been visited and rejects snapshots exceeding the tracked budgets. Override
`PROFILE_DIR` or `PROFILE_BUDGETS` to compare another run or budget set.

Capture locally reviewable scene screenshots after ten stable presented frames:

```shell
MPQ_DIRECTORY=/path/to/diablo-ii make capture
```

The default captures `loading` and `title`; override `CAPTURE_SCENES` with a
comma-separated list or pass `--capture-scenes` directly. `--capture-settle-frames`
controls stabilization. Use `START_SCENE=character_create make capture` or
`--start-scene character_create` to enter a registered scene directly for local
development review. Passing `--capture-scenes` without `--capture-dir` writes to
`./captures/frontend`, so a single overlay can be captured with
`--start-scene=death --capture-scenes=death`. Capture exits automatically once every requested scene has
been recorded. The capture directory contains numbered PNGs plus a
`report.json` recording scene names, dimensions, and SHA-256 hashes. Capture
output under `captures/` is ignored by Git so Blizzard imagery remains local.
Saved-character presentation can be reviewed without modifying a real save
directory by combining `START_SCENE=character_select`,
`CAPTURE_SCENES=character_select`, and `FIXTURE_CHARACTERS=10`. The equivalent
development-only CLI option is `--fixture-characters 10`; its deterministic
in-memory records exercise both columns, paging, and expansion/hardcore labels.
The default presentation profile remains `lod-english-800x600`. The scoped
`lod-english-640x480-gameplay` profile selects the original 640-wide in-game
control panel and logical viewport across gameplay, side/full overlays, pause,
options, death, and loading transitions. It deliberately does not claim the
800x600 expansion frontend screens have been converted. Select it with
`--presentation-profile lod-english-640x480-gameplay` or
`PRESENTATION_PROFILE=lod-english-640x480-gameplay make capture`.
When starting directly in a gameplay scene such as `game_world`, `inventory`,
`character`, `stash`, or `help`, the first fixture is selected automatically so
asset-backed presentation can be captured in isolation.
Character sources may also attach an immutable appearance snapshot containing
the resolved COF, palette, direction, and component DCC paths. Character
selection renders that equipment-aware composite when available and falls back
to the class presentation for legacy metadata-only records.

Open the integrated composite animation laboratory against mounted MPQs with:

```shell
MPQ_DIRECTORY=~/d2_english_mpq go run ./cmd/client \
  --start-scene=composite_lab
```

Use Left/Right to increment or decrement the logical `0..15` player direction,
Up/Down for NU/WL/RN mode, Page Up/Down
for player class, Space to pause/play, Home/End to step frames, and Enter to cycle deterministic coherent
recipes. Press F to open the shared type-to-filter fuzzy recipe picker. The lab owns its defaults and selection state; no composite-specific
arguments leak through the production client composition root.

Browse a DT1 tileset through the engine's lazy tile decoder with:

```shell
MPQ_DIRECTORY=~/d2_english_mpq go run ./cmd/client \
  --start-scene=dt1_lab
```

The lab incrementally lays out every tile as a centered, labeled gallery cell
and initially fits the complete grid. Tab toggles between that grid and a
readable 1x view centered on the selected tile. Arrow keys or pointer dragging pan
the current view, scroll and Home/End zoom it, Space returns to and recenters
the fitted grid, Page Up/Down cycles Act palettes, and Enter selects another mounted DT1
through a deterministic random sequence. Hovering a tile shows an unscaled,
75%-opaque pointer tooltip reporting its index,
semantic orientation/type, main/style index, sub/sequence index, matching
variant count, direction, rarity, block count, and source dimensions. Cells
with the same orientation/main/sub lookup key share a subtle background color.
Press F to select an exact mounted DT1 through the shared fuzzy picker. Files
using the archived preliminary `4.1` layout remain discoverable but produce an
explicit unsupported-layout diagnostic; they are never decoded with `7.6`
offsets.

Open a DS1 stamp with its authored DT1 sources using:

```shell
MPQ_DIRECTORY=~/d2_english_mpq go run ./cmd/client \
  --start-scene=ds1_lab
```

Arrow keys pan, Home/End zooms, Space fits the complete stamp, Page Up/Down
cycles Act palettes, and Enter selects a mounted DS1 plus exactly the DT1
libraries declared by that stamp (including historical `.tg1` name mapping).
Both map labs discover their mounted assets themselves and open with
a useful selection. The scenes can be captured as `dt1_lab` and `ds1_lab`
through the ordinary capture flags.
Press F to find a specific mounted DS1 by any subsequence of its path. Monster
and missile labs expose the same picker for their typed record IDs.

`--start-scene=combat_lab` is a production-world combat instrument, not a
second miniature combat implementation. `make play-combat-lab` selects a
development character, enters the generated Act I Blood Moor, admits its typed
hostile population, and then delegates the complete scene lifecycle to
`game_world`. The lab relocates up to three members of that real population to
reachable open subtiles within the initial viewport, and admission keeps the
data-backed general Attack (skill ID 0) on the left mouse button. Pointer
traversal and hostile selection therefore exercise the
same fixed-tick intents, collision-aware A*, sparse tile residency and culling,
camera, entity depth ordering, player/monster composites, missiles, HUD,
damage, death, and loot cues used by ordinary play. F3 toggles decoded
collision, F4 tile geometry, F5 entity origins, and F6 the lab-only read-only
combat panel and hostile markers. The panel reports authoritative positions,
animation/attack phase, equipment-derived melee range and damage, nearest
hostile health/AI, and the latest semantic combat event. Run
`make capture-combat-lab` for a local MPQ-backed review image; game assets and
captures remain outside the repository.

`--start-scene=mapgen_lab` opens the pointer-driven generation proof. It shows
the typed Act I Cave Level 1 room topology, chamber recipe, checksum, and
decision trace without materializing renderer assets. Click its seed controls
to produce another deterministic zone; DS1 Lab remains the separate, lazy
materialization tool. The `engine.mapgen/v1` capability also retains the earlier
typed Act I Tristram preset proof.

`--start-scene=warp_lab` opens an argument-free spatial transition proof using
two visibly different mounted Act I DS1 stamps placed more than one viewport apart.
The player starts beside the western portal. Click ordinary ground to publish
a traversal intent, or click the portal to publish an interaction intent;
fixed-tick ECS authority walks into range, resolves the
portal's explicit paired entity, teleports to the eastern endpoint, and moves
the camera to the second stamp. This is not a special lab-only map renderer:
both destinations use the game world's sparse tile residency, viewport culling,
camera clamping, depth ordering, and subtile projection adapter. Player ground
movement and portal approach also select the ordinary NU/WL composite modes and
preserve animation playback through facing and destination changes. Pointer
targets are planned by the same deterministic, player-footprint-aware A* over
the decoded DT1 collision map; unreachable clicks are rejected instead of
letting the lab actor pass through scenery. A completed
warp immediately applies the legacy-style full-screen black mask and fades it
away over roughly 100 ms, hiding the destination camera/residency handoff. The
endpoints use the shipped animated blue
town-portal and red permanent-portal COF/DCC composites with luminous screen
composition over the world. Run it with `make play-warp-lab`, or create a
local production-asset image with `make capture-warp-lab`. Blizzard-owned
capture output remains ignored by Git.

Generated zone recipes are decoded one stamp at a time by the renderer-neutral
world materializer. It reuses DT1 catalogs across matching rooms, reports safe
loading progress, and only publishes the completed tile/object/collision map.
The extra shared terminal edge present in production DS1 room stamps is clipped
to the generated room footprint during composition.

DS1 Lab composes sparse 512x512 CPU chunks split into floor, lower-wall,
shadow, upper-wall, and roof passes. Tab isolates those passes, including the
paired north-corner records required by legacy stamps. The lab admits at most
two newly visible native textures per frame and releases retained chunk nodes
outside a viewport margin. Its status line reports logical dimensions,
demand-resident/total chunks, and authored object placements. Objects remain
simulation facts rather than pixels baked into a tile chunk. The command-line
PNG inspector intentionally retains its full-image path.
Press F3 in DS1 Lab to lazily overlay the authoritative DT1 subtile collision
grid: red is walk blocking, orange player-only blocking, blue line-of-sight,
magenta jump, and yellow light blocking. This diagnostic uses the same
bottom-to-top DT1 row transform consumed by simulation.

Inspect a legally obtained Diablo II asset without starting the renderer:

```shell
go run ./internal/dev/tools/asset_inspect \
  -source /path/to/d2data.mpq \
  -asset data/global/ui/Loading/loadingscreen.dc6 \
  -preview ./loading.png
```

Verify the curated screen-asset knowledge against a complete MPQ directory and
generate a JSON report plus palette-applied DC6 contact sheets:

```shell
go run ./internal/dev/tools/asset_catalog \
  -mpq-dir /path/to/diablo-ii \
  -out ./asset-catalog
```

Validate the installation against the redistributable structural fixture
without extracting or committing any Blizzard-owned pixels:

```shell
go run ./internal/dev/tools/asset_catalog \
  -mpq-dir /path/to/diablo-ii \
  -no-sheets \
  -fixture internal/content/d2legacy/manifests/asset-fixture.v1.json
```

View a Bink cinematic directly from an MPQ directory (FFmpeg/`ffplay` is
required, and the temporary extracted file is removed when playback exits):

```sh
go run ./internal/dev/tools/bik_view \
  -source ~/d2_english_mpq \
  -asset data/local/video/New_Bliz640x480.bik
```

Standalone files are supported with `-file movie.bik`. The equivalent Make
target accepts `MPQ_DIRECTORY` and `BIK_ASSET` environment variables.

Build the client with `-tags ffmpeg` and link FFmpeg development libraries via
`pkg-config` to decode Bink video and audio in-process inside the Dark Magic
window:

```sh
MPQ_DIRECTORY=/path/to/diablo-ii go run -tags ffmpeg ./cmd/client
```

Portable builds retain `ffplay` as a separate-window diagnostic fallback. If it
is absent, startup follows the manifest's failure policy and skips video without
preventing the client from loading.

The report records missing and disputed paths, the archive layer that supplied
each asset, hashes, direction/frame counts, dimensions, and stored offsets. Use
`-manifest custom.json` to verify additional hypotheses or `-no-sheets` for a
metadata-only pass. The command is read-only with respect to the MPQs.

The optional community discovery index can be audited independently:

```shell
go run ./internal/dev/tools/asset_catalog \
  -mpq-dir /path/to/diablo-ii \
  -listfile ./docs/Diablo2UberListfile.txt \
  -no-sheets \
  -out ./asset-catalog
```

Its `listfile-report.json` deliberately distinguishes paths merely listed by
community research from paths actually resolvable in the selected MPQ set.

Host filesystem paths accepted by Dark Magic consistently expand `~`, `~/`,
`~\`, `$NAME`, `${NAME}`, and `%NAME%` forms on every supported platform.
This applies to environment configuration and command input/output paths. MPQ
asset paths are virtual paths and are intentionally not expanded.

When the debug web server is running, the same metadata and selected DC6/DCC
frames are available through the asset routes:

```text
GET /asset/inspect/data/global/ui/Loading/loadingscreen.dc6
GET /asset/preview/data/global/ui/Loading/loadingscreen.dc6?direction=0&frame=0
```

Run the standalone interactive scene slice:

```shell
go run ./internal/dev/testapps/scene_demo
```

The main engine also starts the integrated scene service. Set `MPQ_DIRECTORY`
to one or more comma-separated content directories. Each directory is mounted
for ordinary filesystem assets and scanned for the supported Diablo MPQs. The
listed paths are ordered from highest to lowest priority and support
cross-platform aliases:

```shell
MPQ_DIRECTORY="~/my-mod,~/d2_english_mpq" go run ./cmd/client
```

To explore a real DS1 layout from an MPQ:

```shell
go run ./internal/dev/testapps/scene_demo \
  -source /path/to/d2data.mpq \
  -map data/global/tiles/Act1/BARRACKS/barE.ds1 \
  -dt1 data/global/tiles/Act1/BARRACKS/floor.dt1,data/global/tiles/Act1/BARRACKS/basewall.dt1,data/global/tiles/Act1/BARRACKS/barset.dt1 \
  -palette data/global/palette/ACT1/pal.pl2
```

Use WASD or the arrow keys to move, F5 to save, and F9 to load. See
[ROADMAP.md](ROADMAP.md) for restart progress and the next integration work.

## License

This project is licensed under the MIT License - see the 
[LICENSE](LICENSE) file for details.

## Acknowledgments
* [Open Diablo 2](https://github.com/opendiablo2/opendiablo2)
* [Gin](https://github.com/gin-gonic/gin)
* [Gopher LUA](https://github.com/yuin/gopher-lua)

---
*Dark Magic is not affiliated with or endorsed by Blizzard Entertainment, Inc. Diablo is a registered trademark of Blizzard Entertainment, Inc. All in-game content, imagery, and lore are the property of Blizzard Entertainment, Inc.*
