# Dark Magic Modding API Guide

Status: `dark-magic.mod/v1` and the currently exposed `engine.* /v1` Lua modules.
The project is pre-release; incompatible changes require a manifest/API version
change and a migration note.

## Package contract

Dark Magic always supplies embedded `d2legacy`. A user mod is an `extension`
layered above it. Start by copying `internal/content/mod_template`, then rename
every `mod_template` path and ID to a stable lowercase package namespace.

```json
{
  "schema": "dark-magic.mod/v1",
  "id": "example",
  "name": "Example Extension",
  "version": "0.1.0",
  "kind": "extension",
  "engine_api": "v1",
  "redistributable": true,
  "content_roots": ["assets", "locales"],
  "entrypoints": {
    "client_components": ["example.boot"],
    "authority_components": ["example.authority"]
  },
  "dependencies": [{"id": "d2legacy", "version": "0.1.0"}]
}
```

Set `redistributable` to `false` unless every archive file may legally and
intentionally be shared by a game host. Never package Blizzard assets, MPQ
contents, saves, credentials, or keys.

## Layout, VFS, and namespaces

```text
example/
  mod.json
  boot.lua
  components/authority.lua
  lua/example/policy/example.lua
  lua/example/policy/example_test.lua
  assets/...
  locales/...
```

The archive is private at `mods/example/`. Module
`require("example.policy.example")` resolves only from the owning package. The
loader clears cached modules when a session changes or removes that package.

Only `content_roots` also appear at the shared VFS root. `components`, `lua`,
and `mods` cannot be exported. Dependencies activate before dependents; shared
lookup is reversed so a dependent may intentionally override an exported asset
without overriding private code. Use shared roots sparingly for assets, data,
locales, or versioned manifests.

## Components and lifecycle

`boot.lua` and each `components/*.lua` return one managed definition:

```lua
return {
    id = "example.boot",
    api = 1,
    depends_on = { "d2legacy.boot" },

    start = function(self)
        -- Allocate subscriptions/handles and register scenes here.
    end,

    stop = function(self)
        -- Scope-owned native resources are reclaimed automatically.
    end,
}
```

Component IDs and entrypoints must begin with the package ID. Dependencies may
refer only to the same package or a manifest dependency. Discovery never starts
a component. Package IDs must not overlap as dot namespaces with another
active package. Composition starts declared entrypoints and their dependency
closure, then stops them in reverse.

Use `client_components` for UI, audio, input, and presentation. Use
`authority_components` for deterministic commands, systems, and state.
Authority components run in offline, listen, dedicated, and realm-managed game
authorities—not in a connected graphical client.

Resource-producing APIs require an active component scope. Do not hide native
handles in globals or transfer them between component instances.

## Engine APIs and trust

Engine APIs are exact versioned modules such as `engine.scene/v1`,
`engine.render/v1`, `engine.records/v1`, `engine.animdata/v1`, `engine.ecs/v1`,
and deterministic authority modules. The read-only AnimData capability exposes
format facts from the session-pinned authoritative generation without assigning
them skill or combat meaning. `d2legacy.help()` and `d2legacy.capabilities()` expose
metadata for modules registered in the current runtime.

The manifest intentionally has no `capabilities` permission field. Modules are
currently selected by runtime domain, not granted per package. Extension Lua is
executable code and must be trusted. A future permission format requires actual
isolated-runtime enforcement; decorative declarations are not security.

## Authoritative gameplay and ECS

Lua may decide gameplay, but durable state belongs to registered engine
mechanisms:

- execute on fixed ticks, never wall-clock time;
- use purpose-named deterministic RNG streams;
- validate semantic commands before mutation;
- declare ECS query and read/write access;
- make structural changes through barriers/command buffers;
- store future-tick facts in registered ECS or authority state;
- keep presentation, filesystem, network, and ambient randomness out of
  deterministic handlers; and
- test checkpoint/restore parity for behavior spanning ticks.

Commands validate external transitions. Systems advance tick state. Pure policy
modules calculate without owning ECS, scheduling, or I/O.

Component IDs are durable schema identities under the package namespace, such
as `example.regeneration.progress`. Do not reuse a d2legacy ID for different
data. Prefer this one-way path:

```text
semantic command -> validated request/state -> deterministic system
                 -> durable components -> passive client projection
```

The archive digest/order, authoritative configuration, engine API, network
protocol, and capability-contract versions are pinned by the runtime recipe.
Admission, reconnect, checkpoints, and replays reject drift.

## UI composition

Scenes own lifecycle; controls own focus/hit testing; widgets own retained
nodes. UI observes authenticated snapshots and submits semantic intents. It
never mutates authoritative ECS state directly.

Use root-scene replacement for whole screens and overlays for gameplay panels.
Store instance state on `self`, use scope-owned subscriptions, keep asset/string/
layout facts in manifests, and make keyboard/controller focus explicit.

## Lua tests

Place `foo_test.lua` beside `foo.lua`. The production harness discovers it and
exposes each Lua case as a Go subtest. Choose the narrowest profile:

| Profile | Use |
| --- | --- |
| `module` | Pure policy, records, and helpers without ECS. |
| `ecs` | Component schemas and focused ECS behavior. |
| `authority` | Commands, systems, checkpoints, vertical gameplay. |

```lua
local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("adds the authored bonus", {
            test.run(function()
                local policy = require("example.policy.example")
                test.expect(policy.resolve(2, 3), "resolved value"):equals(5)
            end),
        }),
    },
})
```

Tests should read as arrange, act, assert. Require production modules and use
production schemas, constructors, and systems; never create replicas that can
drift. Add restore parity for stateful authority behavior.

```sh
go test ./internal/mod/d2legacy -run 'TestLuaSuites/example/path/foo'
make test-lua
make test-lua-hardening
go test ./...
```

## Installation and networking

Extension archives are immutable ZIP blobs in the platform cache. The
extension-only `mods.json` selects local defaults. Installed does not mean
enabled; names resolve to exact descriptors before mounting.

A direct host advertises its complete runtime recipe over pinned TLS. Before
character admission, a client may stream missing redistributable extensions to
quarantine. Full digest/size/archive/manifest/dependency verification precedes
atomic promotion and client activation. The client recomputes the recipe from
local bytes before joining. See [MODS.md](MODS.md) for exact precedence,
transfer limits, concurrent versions, and remaining realm work.

## Compatibility and migration

- Never change released bytes while retaining version/digest identity.
- Treat component/state IDs, schemas, commands, and entrypoints as durable.
- Add a new engine-module version instead of changing `/v1` semantics.
- Declare exact dependency versions when behavior depends on them.
- State migration must name source/destination identities and run at an explicit
  safe boundary.
- Revocation blocks new admission; it does not rewrite historical replay
  identity.

## Pre-publication checklist

- Manifest validates as an extension of d2legacy.
- IDs, namespaces, components, and dependencies are package-owned.
- Shared roots contain only intentional, legally distributable overrides.
- Client and authority entrypoints are accurate.
- Deterministic code has no ambient time/random/I/O and stores durable state.
- Tests use production APIs and restore parity where applicable.
- Clean-profile resolution succeeds and tampered bytes fail.
- Listen and dedicated authorities execute the same authority behavior.
- Connected clients run client components only.
- Redistributability is conservative and documented.
