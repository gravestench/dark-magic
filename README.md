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

## Directory structure

* `cmd` contains the client, game-session server, and realm composition roots.
  Commands perform wiring and process configuration, not gameplay.
* `internal/game` owns typed Diablo records, deterministic ECS scheduling,
  simulation primitives, world decoding, and loot rules.
* `internal/content` owns the layered directory/MPQ/ZIP VFS and redistributable
  first-party Lua shim.
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
The `dm.ecs/v1` capability additionally lets trusted scripts define validated
component schemas and deterministic, scope-owned systems over the shared Akara
world. Systems declare their query and read/write access; structural mutations
are applied at phase barriers rather than during query iteration.
The shim also preserves Riiablo's recovered quest hierarchy and speech-to-string
relationships as validated data exposed through `dm.quest_catalog/v1`; scripts
do not need to recreate executable-era Diablo II rules as hard-coded branches.
Recovered DS1 definition and act-local object mappings are available separately
through `dm.map_catalog/v1`.

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
opens with a target- and policy-specific message of the day. The `dm` root
(`darkmagic` is an alias) provides discoverable, policy-filtered capability
access: use `dm.help()` and `dm.capabilities()`, friendly names such as
`dm.app`, or `dm.modules["dm.app/v1"]` for an exact versioned module ID.
Pass a module, command, or path to help for progressively more detail—for
example `dm.help(dm.audio)`, `dm.help(dm.audio.play)`, or
`dm.help("dm.audio.play")`. Lua module registrations own the summaries, usage,
parameters, returns, and examples used by both help and completion. Existing
commands without authored metadata are still listed with fallback help.

Shell presentation settings are live Lua runtime values. For example:

```lua
dm.shell.values()                  -- current native-resolution shell settings
dm.shell.set("font_size", 22)     -- apply immediately for this process
dm.shell.set_many({                -- validate and apply atomically
    console_height = 0.7,
    opacity = 0.85,
    transcript_limit = 4000,
    animation_speed = 1.5,
})
dm.shell.defaults()                -- inspect built-in values
dm.shell.reset()                   -- restore defaults in memory
dm.shell.reload()                  -- discard edits and reload the saved file
dm.shell.save()                    -- persist the active values
dm.shell.status()                  -- persistence path and dirty state
```

Settings default to `shell.json` under the platform user-configuration
directory. `DARK_MAGIC_SHELL_CONFIG` selects another host path and supports
home-directory aliases. Multiline Lua values retain line breaks, indentation,
and tabular spacing in both graphical and terminal shell views.

Game preferences are separate from developer-shell presentation. The authored
in-game sound and music sliders update mixer buses immediately through
`dm.settings/v1` and save to `preferences.json` under the platform
user-configuration directory when the overlay closes. Set
`DARK_MAGIC_PREFERENCES` to use another file.

Use `dm.apropos("music")` to search the permitted module and command
descriptions. `dm.docs()` renders Markdown for the session's complete permitted
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
go run -tags ffmpeg ./cmd/darkmagic --viewport-fit contain
```

Use `--viewport-fit stretch` to fill the entire window without preserving the
game aspect ratio. `DARK_MAGIC_VIEWPORT_FIT` provides the equivalent environment
setting. Mouse input is mapped back into logical game coordinates; input in
letterbox regions is excluded from the game viewport.

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
are operational. Full world generation, Diablo combat and progression,
authoritative networking, and end-to-end gameplay remain in progress.

See [ROADMAP.md](ROADMAP.md) for the canonical milestone backlog and
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for package ownership, dependency
direction, and guidance on where new work belongs.

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
go run -tags ffmpeg ./cmd/darkmagic --log-level debug
DARK_MAGIC_LOG_LEVEL=warn go run -tags ffmpeg ./cmd/darkmagic
```

Quantize the final composed display with an optional GPU post-process. The
lookup cube emits only colors from the selected palette while preserving source
alpha:

```shell
go run -tags ffmpeg ./cmd/darkmagic \
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
  go run -tags ffmpeg ./cmd/darkmagic --profile-dir ./profiles/menu-review
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
go run -tags ffmpeg ./cmd/darkmagic \
  --profile-dir ./profiles/frontend-review \
  --profile-scenes all
```

Scene artifacts are written beneath `scenes/<scene-id>/`. `cpu.pdf` combines
all visits to that scene using pprof labels. Each visit gets a numbered
`heap-NNN.pprof` and `heap-NNN.pdf` snapshot taken immediately before the
scene's owned resources are released.

Every scene snapshot also receives `diagnostics-NNN.json`, containing decoded
cache residency and hit/miss/eviction counts, cumulative decode time, retained
render resources and estimated RGBA bytes, and texture upload volume. A
repeatable frontend acceptance run is available with:

```shell
MPQ_DIRECTORY=/path/to/diablo-ii make profile
make profile-check
```

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
When starting directly in `game_world`, `inventory`, or `character`, the first fixture is
selected automatically so asset-backed presentation can be captured in
isolation.
Character sources may also attach an immutable appearance snapshot containing
the resolved COF, palette, direction, and component DCC paths. Character
selection renders that equipment-aware composite when available and falls back
to the class presentation for legacy metadata-only records.

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
  -fixture internal/content/shim/manifests/asset-fixture.v1.json
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
MPQ_DIRECTORY=/path/to/diablo-ii go run -tags ffmpeg ./cmd/darkmagic
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
MPQ_DIRECTORY="~/my-mod,~/d2_english_mpq" go run ./cmd/darkmagic
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
