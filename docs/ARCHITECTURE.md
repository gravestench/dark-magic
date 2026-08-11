# Dark Magic architecture

Dark Magic is one application, not a collection of independently supported Go
libraries. `cmd/darkmagic` is the composition root; engine implementation and
Diablo-specific behavior belong under `internal`. A package remains under `pkg`
only when the project deliberately promises it as a stable, independently useful
API. No current Go package has that commitment; an acceptance test rejects
accidental public Go source until a deliberate API and compatibility policy are
documented.

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
| `internal/dev/profiling` | CPU/heap/scene profile capture | command | Application/run | Keep under developer support |
| `internal/dev/capture` | Screenshot fixture writing | command | Run | Keep under developer support |
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
| `internal/game/player` | Authoritative player entry and stable ECS archetype materialization | session composition | Game session | Keep internal |
| `internal/game/world` | Immutable map facts, generated-zone composition, collision, and shared tile/subtile/pixel/camera transforms | session, Lua, presentation | Game session | Keep renderer-neutral |
| `internal/game/ecs` | Deterministic Akara-backed phases, queries, access contracts, and structural barriers | command, Lua | Game session | Keep internal |
| `internal/game/session` | Authoritative command admission, fixed stepping, checkpointing, and replay recording | client/server composition | Game session | Keep transport-neutral |
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
internal/runtime/lua       Lua ownership, capabilities, scopes
internal/shell             shared sessions, Lua evaluator, terminal adapter
internal/presentation      navigation, scenes, controls, transitions
internal/game/data         typed Diablo records and validation
internal/game/loot         deterministic item generation
internal/game/player       authoritative player archetypes and entry
internal/game/ecs          deterministic entity schedule and structural barriers
internal/game/simulation   replay contracts, RNG, and higher-level gameplay rules
internal/game/session      shared authoritative session owner
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
through `internal/platform/raylib/input`; the transport-neutral authoritative
owner in `internal/game/session` advances `internal/game/ecs`; Lua scene updates
run through `internal/runtime/lua`; and retained presentation commands cross
`internal/presentation/render` before `internal/platform/raylib/renderer` executes them. Game rules
must remain usable without this native frame loop.

Akara owns entity identity, typed and runtime-defined component storage,
archetypes, and subscriptions. Dark Magic owns named simulation phases, system
ordering, read/write declarations, command-buffer barriers, and a bounded fixed
25 Hz clock that is independent of renderer cadence. `dm.ecs/v1`
adapts Lua schemas and scoped system callbacks to that engine contract; Akara
does not import Lua or Dark Magic. Lua may mutate declared component fields
immediately, while entity creation and component add/remove operations are
deferred until the current system barrier.

The shared game-session owner admits commands by stable actor identity, target
tick, per-actor sequence, declared authority class, kind, and payload policy.
Player, administrator, and system authority must be granted by each trusted
handler. Administrative Lua may inspect replay and audit records, but concrete
privileged mutations must remain explicit handlers; no generic ECS mutation
backdoor is part of the administration contract.

The authoritative session checksum covers both the ECS snapshot and registered,
stable-ID state participants. A subsystem whose command handlers mutate state
outside Akara must provide deterministic snapshot and atomic restore operations
and register before the first command or tick. Item authority uses its versioned
container archives for this boundary and includes the identity of server-owned
trade and service rules, so replay rejects configuration drift rather than
silently accepting different transaction results.

Executable-era relationships recovered by Riiablo live verbatim under
`internal/content/shim/data/recovered/riiablo`, accompanied by provenance. The
`internal/game/data/recovered` catalog validates and normalizes those files;
`dm.quest_catalog/v1` exposes identifiers to Lua while localization and audio
remain separate capabilities responsible for resolving strings and assets.

The production game-world scene defines hero position, velocity, bounds, player
control, and camera-follow components in Lua through `dm.ecs/v1`. Its
`darkmagic/gameplay/components` modules group small related schemas, while
`darkmagic/gameplay/systems` gives each update rule its own documented file.
`world.lua` is their composition root and retains only player binding plus
presentation-safe snapshot helpers. Native input is normalized into one admitted
`player.move` command per active fixed tick; the session-owned handler applies
velocity before Lua movement, collision, and camera systems run. The retained
scene only reads component snapshots to update presentation nodes. The older
`dm.simulation/v1` adapter remains available to compatibility tests and shell
examples but is no longer registered by the client.

Scene navigation belongs to `internal/presentation/navigation`; renderer-independent scene
state belongs to `internal/presentation`; and authored screen behavior belongs in
the shim Lua scripts under `internal/content/shim`. Lua modules expose explicit
capabilities but do not own native resources or discover arbitrary services.

Assets enter through `internal/content`, which resolves layered directory, MPQ,
ZIP, and shim sources. `internal/assets/decode` converts supported formats,
`internal/presentation/render` describes retained resources, and the Raylib adapter owns
uploads and disposal. Inspection and catalog tools reuse the same content and
decode paths under `internal/assets/inspect` and `internal/assets/catalog`.

World-space draw ordering is a renderer-neutral policy in `internal/game/world`.
Map passes and standing entities receive comparable projected-baseline keys;
the chunk adapter may batch equal-depth facts but must not put the entire map
under a parent that prevents walls and entities from interleaving.

Pointer gameplay crosses the presentation boundary as a world-subtile target,
not a screen pixel or direct velocity mutation. Lua applies the shared inverse
camera transform, then the fixed-tick session records and admits the target;
authoritative ECS position and collision determine movement. Future entity hit
testing must turn that same target into explicit interaction or skill commands.

Resolved authored DS1 objects enter a uniform-grid selection index as stable
semantic footprints. Presentation may submit world coordinates and display the
selected facts, but authoritative interaction re-runs selection, owner range,
and DT1 line-of-sight checks before changing context. Client-provided target IDs
remain a compatibility/admin surface, not the pointer gameplay trust boundary.

Native mouse buttons are normalized before Lua sees them. Secondary world input
submits a semantic coordinate and optional selected-entity ID; the fixed-tick
skill command resolves the authoritative assignment and learned/side-allowed
skill, then records the intent in ECS for later cast/cost/timing systems.

Pointer movement paths are deterministic eight-way subtile paths over the same
DT1 collision queried by integration. Entity radius affects passability,
diagonals cannot cut blocked corners, occupied interaction targets use reachable
stopping rings, and unreachable requests cancel explicitly. Selection kinds
must come from spawned authoritative definitions—not DS1-name heuristics.

Live selectable entities attach `dm.world.selectable` beside their authoritative
position. Its explicit kind is one of player, NPC, hostile, item, portal,
missile, or scenery. The targeting capability returns copied hit facts and
filters the locally owned player in presentation; it does not expose ECS stores
or infer classifications from asset labels.

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
