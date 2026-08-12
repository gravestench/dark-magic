# Gameplay ownership migration inventory

Status: M21.14.1 migration contract. This inventory classifies production code
by architectural destination; it does not claim that the move is complete.

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
| `internal/game/ecs`, `simulation`, `session` | ECS schedule, command envelopes, RNG/replay, plus some D2 handlers | Keep mechanisms; move embedded D2 handlers to `d2legacy` |
| `internal/game/data/**` | Typed D2 schemas, immutable catalogs, recovered records | Keep decoding/data in Go; Lua interprets gameplay meaning |
| `combat`, `skill`, `missile` | D2 hit/cast/missile policy mixed with reusable fixed-point/spatial work | Migrate policy together; extract only proven generic primitives |
| `monster`, `loot`, `item`, `player`, `stats`, `state`, `action` | Most current authoritative D2 simulation | Migration source, not permanent engine API |
| `mapgen`, `world`, `targeting` | Reusable geometry/navigation mixed with D2 act/level/population policy | Split mechanisms from D2 selection and relationships |
| `interaction`, `transition`, `ownedunit` | Generic relations/transactions mixed with D2 services and lifecycle policy | Split, then migrate D2 policy |
| `internal/runtime/lua` | Runtime mechanisms plus subsystem-shaped Go-owned D2 facades | Keep runtime/adapters; retire facades as policy moves into Lua |
| bundled `darkmagic` Lua | Presentation, labs, and early gameplay helpers | Separate generic examples from first-party `d2legacy` mod |

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
4. Move the complete Fire Bolt path as the first proof.
5. Move tightly coupled domains in coherent groups rather than maintaining two
   authorities during a prolonged file-by-file port.
