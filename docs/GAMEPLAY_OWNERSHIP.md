# Gameplay ownership migration inventory

Status: enforced M21.14 ownership contract. This inventory classifies the
post-migration production code and prevents the deleted Go policy boundary from
returning.

The machine-checked source is
[`docs/architecture/gameplay-ownership.tsv`](architecture/gameplay-ownership.tsv).
Every production Go file under `internal/game` and `internal/runtime/lua`, plus
every bundled Lua file, must match exactly one rule. New files fail architecture
CI until their ownership is intentional.

## Classification vocabulary

| Class | Meaning | Destination |
| --- | --- | --- |
| `mechanism` | Mod-neutral scheduling, storage, math, geometry, transactions, or infrastructure | Generic Go engine |
| `d2-policy` | A Diablo II decision, formula, relationship, or behavior | First-party `d2legacy` Lua mod |
| `data` | Codec, schema, typed decode, immutable record, or validation boundary | Go data layer |
| `adapter` | Narrow Go/Lua, presentation, command, or platform translation | Boundary package; never policy owner |
| `transitional` | Mixed or subsystem-shaped API that must be split/replaced | Delete after consumers migrate |

`Lua owns gameplay decisions; the engine owns durable state.` A `d2-policy`
classification therefore does not imply arbitrary Lua globals: migrated state
must use ECS or registered, versioned engine stores.

## Current domain disposition

| Current area | Current contents | Intended result |
| --- | --- | --- |
| `internal/game/ecs`, `simulation`, `session` | ECS schedule, command envelopes, RNG, transactional ticks, checkpoints, and replay | Generic mechanisms only |
| `internal/game/data/store`, `typed`, `model` | Generic rows, optional schema binding, and lossless D2 row schemas | Keep mechanisms/schemas in Go; no global catalog; Lua selects and interprets tables |
| former Go gameplay packages | Deleted migration sources for combat, skills, missiles, monsters, loot, items, players, progression, interactions, transitions, owned units, and D2 map generation | Production policy now lives under `d2legacy` Lua |
| `internal/game/world`, `worldgen` | Decoded map facts, geometry, collision, navigation, materialization, and immutable recipe contracts | Keep reusable mechanisms/data; Lua chooses D2 topology and population policy |
| `internal/runtime/lua` | Sandboxed runtime, versioned capabilities, authoritative registration, and resource scopes | Keep generic and mod-neutral |
| bundled Lua content | Canonical first-party `d2legacy` authority, presentation, and labs | One mod and one `d2legacy.*` namespace |

## Migration execution policy

Migration PRs optimize for coherent ownership, not continuously runnable
intermediate commits. A related domain may move as one large swath, and
intermediate commits may intentionally fail to build or boot. Each PR must say
where that transition begins and ends, avoid presenting a broken midpoint as a
release, and finish with its stated architecture and deterministic acceptance
checks green. The tagged baseline
`baseline/pre-d2legacy-lua-migration-2026-08-11` is the recovery point before
this architectural refactor.

Do not preserve compatibility wrappers merely to keep old internal APIs alive.
There are no external consumers. Prefer deleting superseded Go policy and
changing call sites together.

## Authoritative Lua readability contract

`d2legacy` is intended to teach as well as run. Do not replace large Go files
with large Lua files. Keep component schemas, registered state, command
handlers, deterministic systems, and domain policy in separate purpose-named
modules. Composition roots only import and register those modules.

Functions should do one small job, use descriptive names, and stay short enough
to understand without scrolling through unrelated behavior. Comments explain
ownership, ordering, state lifetime, legacy evidence, numeric units, and why a
rule exists in plain language suitable for a reader new to the engine. Avoid
comments that merely repeat syntax. Shared helpers are extracted only when they
make the rule easier—not to hide control flow behind abstraction.

Each migrated domain must include a short README or module-level guide showing:

1. which engine capabilities it receives;
2. which state it reads and writes;
3. command-to-system execution order;
4. where decoded D2 records become gameplay decisions; and
5. which tests prove replay/checkpoint behavior.

## First migration sequence

1. Make the inventory exhaustive and enforce its dependency direction.
2. Extract the authoritative Lua runtime/state contract without moving D2 rules.
3. Pin mod identity in session/replay/checkpoint and headless server composition.
4. Use Fire Bolt as the first proof of the reusable cast/straight-missile
   behavior family; never retain a skill-specific subsystem after the boundary
   is proven.
5. Move tightly coupled domains in coherent groups rather than maintaining two
   authorities during a prolonged file-by-file port.
