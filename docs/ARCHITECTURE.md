# Dark Magic architecture

Dark Magic is one application, not a collection of independently supported Go
libraries. `cmd/darkmagic` is the composition root; engine implementation and
Diablo-specific behavior belong under `internal`. A package remains under `pkg`
only when the project deliberately promises it as a stable, independently useful
API. No current Go package has that commitment, so `pkg` contains artwork only;
an acceptance test rejects accidental public Go source until a deliberate API
and compatibility policy are documented.

## Dependency and ownership rules

- Commands parse process configuration, construct the host, and translate final
  errors to exit status. They contain no game, rendering, or protocol behavior.
- Foundation packages may not import presentation, Lua adapters, or game rules.
- Engine capabilities may depend on foundations, but not on the application
  host or a global registry.
- Game-data and simulation packages may depend on engine contracts; platform
  adapters and Lua modules depend inward on those contracts.
- Lua modules adapt explicit capabilities. They never become service locators
  and do not own native renderer, audio, filesystem, or persistence lifetimes.
- Raylib and other platform libraries stay behind internal renderer/input/audio
  boundaries. Headless contracts must remain testable without native startup.
- Tools and test applications may compose internal packages but production code
  never imports them.
- Independently versioned file-format codecs remain separate Go modules.

The intended dependency direction is:

```text
cmd / tools / test apps
          |
          v
host + Lua adapters + platform adapters
          |
          v
engine capabilities + game systems
          |
          v
content, records, handles, cache, paths, logging
```

## Current package inventory

`Application` means the client composition root owns the lifetime. `Scene`
means a resource scope owns it. `Stateless` means there is no runtime lifecycle.

| Package | Responsibility | Main importers | Owner | Disposition |
| --- | --- | --- | --- | --- |
| `cmd/darkmagic` | Client composition root | executable | Process | Keep thin |
| `internal/app/host` | Ordered component lifecycle | command, runtime API, Lua | Application | Keep |
| `internal/content` | Layered directory/MPQ/ZIP/shim VFS | command, reload, Lua, tools | Application | Keep |
| `internal/game/data/store` | Generic immutable TSV generations | typed catalog, audio, Lua | Application | Keep internal |
| `internal/game/data/catalog` | Typed Diablo data snapshots and indexes | command | Application | Keep; split consumers by domain |
| `internal/presentation/render` | Retained renderer contracts and handles | Lua, raylib, video | Application/scopes | Keep internal |
| `internal/audio` | Audio buses, records, playback state | command, Lua, video | Application/scopes | Keep |
| `internal/video` | Cinematic decode/playback orchestration | command, Lua | Scene | Keep |
| `internal/inputstate` | Serialized input state | command, Lua, raylib | Application | Keep |
| `internal/localization` | TBL-backed localization | command, Lua | Application | Keep |
| `internal/loading` | Observable loading progress | command, Lua | Application | Keep |
| `internal/persistence` | Current character persistence boundary | command, Lua | Application | Keep; replace format |
| `internal/presentation/navigation` | Scene/overlay navigation | command, Lua | Application | Keep |
| `internal/runtime/lua` | Serialized Lua runtimes and capabilities | command, reload | Application/scopes | Keep; split adapters by feature when useful |
| `internal/app/hotreload` | Transactional script/content reload | command | Application | Keep |
| `internal/app/filewatch` | Filesystem change observation | command | Application | Keep |
| `internal/app/runtimeapi` | Local runtime-management HTTP API | command | Application | Keep |
| `internal/profiling` | CPU/heap/scene profile capture | command | Application/run | Keep |
| `internal/capture` | Screenshot fixture writing | command | Run | Keep |
| `internal/platform/raylib/common` | Native adapter logging | raylib adapters | Application | Keep under platform |
| `internal/platform/raylib/input` | Raylib input adapter | command, world | Application | Keep under platform |
| `internal/platform/raylib/renderer` | Raylib renderer/audio owner thread | command, world | Application | Keep under platform |
| `internal/platform/raylib/world` | Legacy native world presentation | command/tests | Scene | Transitional; replace |
| `internal/acceptance` | Cross-system acceptance fixtures | tests | Test | Keep |
| `internal/dev/tools/*` | Asset, profile, shim, and extraction CLIs | developer | Process | Keep |
| `internal/dev/testapps/*` | Manual diagnostics and experiments | developer | Process | Keep |
| `internal/assets/decode` | Engine-specific decoded asset helpers | Lua, video, tools | Stateless/cache | Keep internal |
| `internal/assets/catalog` | Presentation asset research/catalog output | tools | Stateless | Keep internal |
| `internal/assets/inspect` | Asset metadata and preview helpers | tools, world | Stateless | Keep internal |
| `internal/cache` | Weighted generation-aware LRU | Lua, renderer | Application | Migrated; guarded |
| `internal/presentation/easing` | Preserved tween easing functions | future presentation runtime | Stateless | Migrated; guarded |
| `internal/game/loot` | Diablo loot rules and TSV compatibility | Lua, test apps | Session | Keep internal |
| `internal/game/data/model` | Diablo TSV schemas and legacy enums | game data | Application data | Keep internal |
| `internal/paths` | Cross-platform host-path expansion | command and tools | Stateless | Migrated; guarded |
| `internal/logging` | Process log formatting | command | Application | Migrated; guarded |
| `internal/presentation/scene` | Headless scene state | command, Lua, world | Scene | Keep internal; merge with navigation when lifecycles converge |

## Target feature layout

Moves should be incremental and behavior-preserving:

```text
cmd/darkmagic              client composition root
internal/app               host wiring and configuration
internal/content           VFS, records, localization, authored shim
internal/assets            decode, cache, inspection, cataloging
internal/platform/raylib   renderer, input, audio/video native adapters
internal/runtime/lua       Lua ownership, capabilities, scopes, console
internal/presentation      navigation, scenes, controls, transitions
internal/game/data         typed Diablo records and validation
internal/game/loot         deterministic item generation
internal/game/simulation   world and gameplay rules
internal/persistence       saves and future realm storage contracts
internal/network           client/session/realm protocols
internal/dev               profiling, capture, tools, test applications
```

Package moves must not introduce compatibility aliases unless a real external
consumer is identified. Before deleting transitional code, verify its callers,
tests, Git history, and any preserved stash notes. Every move must keep unit,
real-asset, race, and relevant interactive acceptance checks green.

## Finding the main execution paths

The client boots in `cmd/darkmagic/main.go`. It parses process configuration,
opens the layered content filesystem, constructs shared application capabilities,
registers Lua modules, and gives those components to `internal/app/host` for ordered
startup and shutdown. Keep this file as wiring: capability behavior belongs in
the package that owns it.

Each frame begins at the Raylib renderer owner thread. Native input is translated
through `internal/platform/raylib/input`, Lua scene updates run through
`internal/runtime/lua`, and retained presentation commands cross
`internal/presentation/render` before `internal/platform/raylib/renderer` executes them. Game rules
must remain usable without this native frame loop.

Scene navigation belongs to `internal/presentation/navigation`; renderer-independent scene
state belongs to `internal/presentation`; and authored screen behavior belongs in
the shim Lua scripts under `internal/content/shim`. Lua modules expose explicit
capabilities but do not own native resources or discover arbitrary services.

Assets enter through `internal/content`, which resolves layered directory, MPQ,
ZIP, and shim sources. `internal/assets/decode` converts supported formats,
`internal/presentation/render` describes retained resources, and the Raylib adapter owns
uploads and disposal. Inspection and catalog tools reuse the same content and
decode paths under `internal/assets/inspect` and `internal/assets/catalog`.

Diablo TSV bytes and generic rows are owned by `internal/game/data/store`. The
typed, atomic generation and indexes live in `internal/game/data/catalog`, using
schemas from `internal/game/data/model`. Consult `docs/GAME_DATA_RECORDS.md` and
the bundled Data File Guide before admitting or interpreting another table, then
verify assumptions against real layered MPQs.

New developer-only executables belong under `internal/dev/tools` or
`internal/dev/testapps`; production entry points belong under `cmd`. A new engine
capability should expose a renderer-independent contract under the relevant
feature directory, receive explicit ownership from the composition root, and
gain both focused tests and a cross-system acceptance test when appropriate.
