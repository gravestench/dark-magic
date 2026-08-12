# Dark Magic architecture

Dark Magic is one application, not a collection of independently supported Go
libraries. `cmd/darkmagic` is the composition root. The reusable engine is
implemented primarily in Go; the first-party `d2legacy` mod owns Diablo II game
rules and mod-specific presentation in Lua. A package remains under `pkg`
only when the project deliberately promises it as a stable, independently useful
API. No current Go package has that commitment; an acceptance test rejects
accidental public Go source until a deliberate API and compatibility policy are
documented.

## Dependency and ownership rules

- Commands parse process configuration, construct the host, and translate final
  errors to exit status. They contain no game, rendering, or protocol behavior.
- Foundation packages may not import presentation, Lua adapters, `d2legacy`, or
  game rules.
- Engine capabilities may depend on foundations, but not on the application
  host or a global registry.
- Game-data decoders, engine simulation mechanisms, platform adapters, and Lua
  modules may depend inward on engine contracts. Generic engine packages never
  depend on `d2legacy` or other mod policy.
- Lua modules use explicit, versioned capabilities. They never become service
  locators and do not own native renderer, audio, filesystem, or persistence
  lifetimes. They may own authoritative gameplay rules and serialized gameplay
  state.
- Raylib and other platform libraries stay behind internal renderer/input/audio
  boundaries. Headless contracts must remain testable without native startup.
- Tools and test applications may compose internal packages but production code
  never imports them.
- Independently versioned file-format codecs remain separate Go modules.

The intended dependency direction is:

```text
cmd / server / tools / test apps
                 |
                 v
       host + platform adapters
                 |
       +---------+------------------+
       |                            |
       v                            v
Lua runtime + engine APIs      first-party d2legacy mod
       |                            |
       +-------------+--------------+
                     v
 generic deterministic engine mechanisms
                     |
                     v
 content, typed records, handles, cache, paths, logging
```

The mod-to-engine arrow is one-way: `d2legacy` consumes engine capabilities.
The engine does not import the mod or encode its rules.

The deployment boundary is equally important:

```text
Client
  Lua: UI, presentation, optional prediction
  Go:  input, rendering, transport
             |
             | authenticated semantic commands / authoritative projections
             v
Game server
  Go:  session host, networking, ECS, clock, replay, durable-state mechanisms
  Lua: authoritative d2legacy gameplay decisions
             |
             | revisioned durable character/account results
             v
Realm services
  Go:  authentication, lobby, matchmaking, allocation, storage,
       coordination, leases, and version negotiation
  Lua: optional realm/season policy only where deliberately moddable
```

Authoritative Lua runs inside the trusted, headless game-server/session process.
Moving D2 policy from Go to Lua does not move authority to the client, weaken
the server boundary, or require realm services to load the gameplay mod. A game
server may host multiple isolated sessions, each pinned to its own compatible
mod package. Realm services select a capable worker and coordinate durable
results; they do not resolve combat ticks.

## Engine and mod ownership boundary

“Authoritative” describes who is allowed to decide and commit gameplay state;
it does **not** mean “implemented in Go.” Trusted Lua command handlers and ECS
systems may execute authoritative gameplay logic. Determinism comes from stable
module and configuration identity, controlled APIs, fixed scheduling,
deterministic RNG streams, restricted side effects, and serialized state—not
from the implementation language.

The state invariant is:

```text
Lua owns gameplay decisions; the engine owns durable state.
```

Authoritative Lua operates on generic ECS components and explicitly registered,
versioned state stores. State that can affect a future tick may not hide in Lua
globals, closures, unserializable userdata, native handles, or presentation
resources. The engine snapshots, checksums, restores, replicates, and persists
registered state; Lua decides how D2 rules transform it through controlled
capabilities.

The Go engine owns reusable mechanisms:

- Lua runtime ownership, isolation, sandboxing, capabilities, instruction/time
  and memory budgets, and resource lifetimes;
- deterministic fixed-tick scheduling, ECS storage, queries, declared access,
  structural barriers, and command buffers;
- command admission and transport-neutral replay infrastructure;
- deterministic RNG primitives and named stream management;
- serialization, checkpoint, restore, checksum, and persistence primitives;
- networking, replication, interest-management, and protocol primitives;
- mod discovery, dependency resolution, package hashing, capability-version
  negotiation, and session identity pinning;
- VFS and file access, codecs, typed record decoding, validation, and immutable
  data generations;
- rendering, audio, input, localization, and platform adapters;
- generic collision, geometry, spatial indexing, navigation, and reusable
  algorithms that contain no Diablo-specific policy; and
- capability enforcement, quotas, cancellation, and native resource lifetime
  management.

The first-party `d2legacy` Lua mod owns Diablo II policy:

- combat formulas, damage policy, hit resolution, and death consequences;
- monster selection, spawning policy, AI, and encounter population;
- skills, casts, costs, targeting rules, and missile behavior;
- item generation, treasure classes, quality, affixes, equipment rules, and
  container policy;
- vendors, prices, services, crafting, and economy rules;
- character classes, stats, progression, experience, and difficulty rules;
- quests, NPC behavior, interactions, transitions, and world progression;
- Diablo II-specific map-generation policy and hard-coded legacy
  relationships; and
- presentation choices that are part of the mod rather than engine mechanics.

Mechanism and policy must be separated deliberately. A generic fixed-point
number, collision sweep, spatial query, stat-source container, or transactional
inventory primitive may remain in Go. Diablo schemas and file decoding may also
remain in Go. Deciding what a decoded record means to Diablo gameplay belongs
in `d2legacy`. Migration reviews must not mechanically transliterate a Go
package into Lua when its useful remainder is a smaller generic primitive.

This is the target architecture, not the current state. Much of the working
Diablo simulation is presently implemented under `internal/game`. Those
packages are migration sources until each file is classified and either moved
to `d2legacy`, reduced to an engine mechanism, retained as a data boundary, or
deleted. Older research documents remain valuable behavioral evidence and
current-state handoffs, but any recommendation that permanent Diablo policy or
authority must live in Go is superseded by this document.

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
| `internal/runtime/lua` | Sandboxed Lua ownership, deterministic execution, capabilities, and scoped resources | command, reload, session | Application/scopes | Keep; add authoritative registration/state boundaries |
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
| `internal/game/{combat,skill,missile}` | Current D2 combat/cast/missile policy plus some reusable numeric and movement mechanisms | session, acceptance | Game session | Transitional migration sources; move D2 policy to `d2legacy`, extract only justified primitives |
| `internal/game/{loot,item}` | Current D2 item generation, containers, equipment, vendors, and services | session, Lua, tests | Game session | Transitional migration sources; retain only generic transaction/storage primitives and typed data boundaries |
| `internal/game/{monster,player,stats,state,action}` | Current D2 actor, AI, progression, stat, timed-state, and action policy | session, skills, combat | Game session | Classify file by file; D2 policy moves to `d2legacy` |
| `internal/game/{interaction,transition}` | Current D2 interaction, commerce, and world-transition policy | session, Lua | Game session | Transitional migration sources; generic command/spatial mechanisms may remain |
| `internal/game/mapgen` | Current D2 preset/outdoor/maze policy mixed with reusable generation algorithms | world, tests | Game session | Move D2 selection/relationship policy to `d2legacy`; keep generic algorithms only |
| `internal/game/ownedunit` | Current owner/category/limit/lifecycle relation and attribution | combat, skills, transitions, UI | Game session | Split generic relation/attribution mechanism from D2 pet, summon, and hireling policy |
| `internal/game/world` | Immutable map facts, collision, geometry, navigation, transforms, and current D2 zone composition | session, Lua, presentation | Game session | Keep generic spatial mechanisms; migrate D2 act/level/map policy |
| `internal/game/ecs` | Deterministic Akara-backed phases, queries, access contracts, and structural barriers | command, Lua | Game session | Keep internal |
| `internal/game/session` | Command admission, fixed stepping, checkpointing, and replay recording for Go- or Lua-owned systems | client/server composition | Game session | Keep transport-neutral and policy-neutral |
| `internal/game/data/model` | Diablo TSV schemas and legacy enums | game data | Application data | Keep internal |
| `internal/paths` | Cross-platform host-path expansion | command and tools | Stateless | Migrated; guarded |
| `internal/logging` | Process log formatting | command | Application | Migrated; guarded |
| `internal/presentation/scene` | Headless scene state | command, Lua, world | Scene | Keep internal; merge with navigation when lifecycles converge |

## Target feature layout

Moves should be incremental and behavior-preserving:

```text
cmd/darkmagic              client composition root
internal/app               host wiring and configuration
internal/content           VFS, records, localization, bundled mods
internal/assets            decode, cache, inspection, cataloging
internal/platform/raylib   renderer, input, audio/video native adapters
internal/runtime/lua       sandbox, authoritative registration, capabilities, scopes
internal/shell             shared sessions, Lua evaluator, terminal adapter
internal/presentation      navigation, scenes, controls, transitions
internal/game/data         typed Diablo records and validation (data boundary)
internal/game/ecs          generic entity schedule and structural barriers
internal/game/simulation   commands, deterministic RNG, replay primitives
internal/game/session      policy-neutral authoritative session host
internal/engine/*          extracted reusable gameplay-adjacent mechanisms
internal/content/.../d2legacy
                           first-party Lua gameplay systems and D2 policy
internal/persistence       saves and future realm storage contracts
internal/network           client/session/realm protocols
internal/dev               profiling, capture, tools, test applications
```

Package moves must not introduce compatibility aliases unless a real external
consumer is identified. Before deleting transitional code, verify its callers,
tests, Git history, and any preserved stash notes. Every move must keep unit,
real-asset, race, and relevant interactive acceptance checks green.

## Finding the main execution paths

The client boots in `cmd/d2/main.go`. It parses process configuration,
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
25 Hz clock that is independent of renderer cadence. `engine.ecs/v1`
adapts Lua schemas and scoped system callbacks to that engine contract; Akara
does not import Lua or Dark Magic. Lua may mutate declared component fields
immediately, while entity creation and component add/remove operations are
deferred until the current system barrier.

The shared game-session owner admits commands by stable actor identity, target
tick, per-actor sequence, declared authority class, kind, and payload policy.
Player, administrator, and system authority must be granted by each trusted
handler. A trusted handler may be registered by Go or by an identified,
sandboxed Lua mod. Administrative and gameplay Lua use explicit handlers rather
than a generic ECS mutation backdoor.

Go command admission remains the outer trust boundary. Transport authenticates
the connection, actor, session, authority class, sequence number, and target
tick before invoking Lua policy. Clients submit intents, never outcomes: damage,
drops, item movement, quest completion, and other results are recomputed by the
pinned server mod against canonical state.

The authoritative session checksum covers the ECS snapshot, registered stable-ID
state participants, and the identity/configuration digest of every authoritative
Lua module. A Go or Lua subsystem whose handlers mutate state outside Akara must
provide deterministic snapshot and atomic restore operations and register before
the first command or tick. Replay and restore reject module/configuration drift
unless an explicit state migration is selected. Lua hot reload cannot silently
replace authoritative code in a running replayable session.

Authoritative execution has deterministic handler/system ordering, controlled
numeric and iteration semantics, engine-supplied named RNG streams, explicit
read/write declarations, and no ambient clock, unrestricted filesystem/network,
OS entropy, or other nondeterministic host API. Script errors and budget
exhaustion fail according to a documented atomic-tick policy: no partial
authoritative mutation is published. Resource limits apply per runtime/session
so one mod cannot starve sibling sessions.

## Mod identity, networking, and prediction

Every game session pins an authoritative mod identity containing at least:

- mod ID and semantic contract/version;
- package/content hash and authoritative Lua source or bytecode hash;
- the complete dependency graph and dependency hashes;
- gameplay configuration identity; and
- required engine capability/API and network protocol versions.

The realm selects a worker that supports this identity and binds it into session
creation and matchmaking. The game server validates it during handshake,
reconnect, late join, replay, checkpoint restore, and explicit session
migration. Replay headers and checkpoints carry it. Durable characters carry
the rule/schema compatibility metadata needed to decide whether admission or
migration is legal. No path may silently restore, join, or replay a session with
different authoritative code or configuration.

Live sessions keep their pinned identity. Changed scripts apply to new sessions
by default. Changing an active production session requires an explicit,
versioned state migration whose input/output identities and failure behavior are
recorded; incompatible versions fail clearly. Development hot reload may use a
separate non-replayable policy, but it must never masquerade as a realm-safe
session.

Clients may use one of three prediction levels:

1. no gameplay prediction: submit intent and render authoritative results;
2. limited generic prediction: predict movement/presentation and reconcile to
   canonical snapshots; or
3. shared `d2legacy` prediction: run compatible Lua locally with rollback and
   reconciliation.

Limited movement and presentation prediction is the initial realm-capable
design. Client Lua remains untrusted even when its hash matches the server mod;
only the game server admits commands and publishes canonical outcomes. Shared
rule prediction is optional later work, never an authority transfer.

Executable-era relationships recovered by Riiablo live verbatim under
`internal/content/shim/data/recovered/riiablo`, accompanied by provenance. The
`internal/game/data/recovered` catalog validates and normalizes those files;
`engine.quest_catalog/v1` exposes identifiers to Lua while localization and audio
remain separate capabilities responsible for resolving strings and assets.

The production game-world scene defines hero position, velocity, bounds, player
control, and camera-follow components in Lua through `engine.ecs/v1`. Its
`d2/gameplay/components` modules group small related schemas, while
`d2/gameplay/systems` gives each update rule its own documented file.
`world.lua` is their composition root and retains only player binding plus
presentation-safe snapshot helpers. Native input is normalized into one admitted
`player.move` command per active fixed tick; the session-owned handler applies
velocity before Lua movement, collision, and camera systems run. The retained
scene only reads component snapshots to update presentation nodes. The older
`engine.simulation/v1` adapter remains available to compatibility tests and shell
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

Live selectable entities attach `d2.world.selectable` beside their authoritative
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
