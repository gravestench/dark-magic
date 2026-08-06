# Dark Magic architecture

Dark Magic is one application, not a collection of independently supported Go
libraries. `cmd/darkmagic` is the composition root; engine implementation and
Diablo-specific behavior belong under `internal`. A package remains under `pkg`
only when the project deliberately promises it as a stable, independently useful
API. No current `pkg` package has that commitment or an in-repository external
consumer, so each is a migration candidate rather than an implied public API.

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
| `internal/host` | Ordered component lifecycle | command, runtime API, Lua | Application | Keep |
| `internal/content` | Layered directory/MPQ/ZIP/shim VFS | command, reload, Lua, tools | Application | Keep |
| `internal/recordstore` | Generic immutable TSV generations | game data, audio, Lua | Application | Keep |
| `internal/gamedata` | Typed Diablo data snapshots and indexes | command | Application | Keep; split by domain |
| `internal/rendercore` | Retained renderer contracts and handles | Lua, raylib, video | Application/scopes | Keep |
| `internal/audiocore` | Audio buses, records, playback state | command, Lua, video | Application/scopes | Keep |
| `internal/videocore` | Cinematic decode/playback orchestration | command, Lua | Scene | Keep |
| `internal/inputcore` | Serialized input state | command, Lua, raylib | Application | Keep |
| `internal/localecore` | TBL-backed localization | command, Lua | Application | Keep |
| `internal/loadcore` | Observable loading progress | command, Lua | Application | Keep |
| `internal/savecore` | Current character persistence boundary | command, Lua | Application | Keep; replace format |
| `internal/navigation` | Scene/overlay navigation | command, Lua | Application | Keep |
| `internal/modruntime` | Serialized Lua runtimes and capabilities | command, reload | Application/scopes | Keep; split adapters |
| `internal/hotreload` | Transactional script/content reload | command | Application | Keep |
| `internal/filewatch` | Filesystem change observation | command | Application | Keep |
| `internal/runtimeapi` | Local runtime-management HTTP API | command | Application | Keep |
| `internal/profiling` | CPU/heap/scene profile capture | command | Application/run | Keep |
| `internal/capture` | Screenshot fixture writing | command | Run | Keep |
| `internal/raylib/common` | Native adapter logging | raylib adapters | Application | Fold into platform |
| `internal/raylib/input` | Raylib input adapter | command, world | Application | Keep under platform |
| `internal/raylib/renderer` | Raylib renderer/audio owner thread | command, world | Application | Keep under platform |
| `internal/raylib/world` | Legacy native world presentation | command/tests | Scene | Transitional; replace |
| `internal/acceptance` | Cross-system acceptance fixtures | tests | Test | Keep |
| `internal/tools/*` | Asset, profile, shim, and extraction CLIs | developer | Process | Keep |
| `internal/testapps/*` | Manual diagnostics and experiments | developer | Process | Keep |
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
