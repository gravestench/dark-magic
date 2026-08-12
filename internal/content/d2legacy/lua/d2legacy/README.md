# d2legacy authoritative gameplay

This tree is the first-party Diablo II rules mod. Its canonical name and Lua
namespace are `d2legacy`. The repository embeds this canonical first-party mod;
there is no separate gameplay package or competing first-party identity.

The Go host supplies generic tools—fixed ticks, ECS storage, command admission,
deterministic random streams, checkpoints, decoded records, rendering, and
platform adapters. The Lua files here decide what those tools mean for Diablo II.

The directory is intentionally split by responsibility:

- `bootstrap/` is composition only. It imports modules and registers them.
- `components/` declares durable ECS data shapes without behavior.
- `data/` turns decoded legacy records into small reviewed rule definitions.
- `commands/` validates player intent and writes authoritative requests.
- `systems/` advances requests during fixed simulation phases.
- `policy/` contains formulas and decisions that do not own scheduling.
- `items/`, `loot/`, `owned_units/`, and `mapgen/` keep larger domains split
  into small rule-focused helpers rather than hiding a subsystem in one file.
- `save/` owns Diablo character-roster policy over a narrow opaque Go store.
- `gameplay/`, `presentation/`, `screens/`, `overlays/`, and `ui/` consume safe
  authoritative facts; they never become a second gameplay authority.

A new reader should be able to open one file and learn one idea. Functions stay
short, module dependencies are explicit, and comments explain legacy meaning,
state lifetime, units, and ordering rather than restating Lua syntax.

## Fire Bolt execution order

1. Go admits `d2legacy.skill.cast` by identity, tick, sequence, and authority.
2. `commands/cast.lua` validates its payload and creates a cast request.
3. `systems/cast.lua` validates learned skill, target, and mana, then schedules
   the effect and completion ticks.
4. `systems/fire_bolt.lua` creates a projectile when the effect tick arrives.
5. `systems/projectile.lua` moves it, finds first contact, and expires it.
6. `policy/damage.lua` resolves Fire damage and death consequences.

All mutable facts live in ECS or registered engine state. Random rolls use a
purpose-named engine stream, so replay and checkpoint restore reproduce them.

## Testing ownership

Lua-owned policy and authoritative behavior tests live beside their production
modules as `*_test.lua`. Cross-domain scenarios may live under `tests/`. The Go
entry point in `internal/mod/d2legacy/lua_suite_test.go` discovers every suite
from the embedded production mod and exposes each Lua case as a named Go
subtest.

A suite uses `d2legacy.tests/v1`. New tests should use `test.case`; the builder
makes production phases explicit without exposing the Go bridge:

```lua
local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("moves the player", function(t)
            t:system_command("system.player.enter", fixtures.amazon_entry, { tick = 1, sequence = 1 })
            t:step()
            t:command("player.move", { x = 1, y = 0 }, {
                tick = 2, sequence = 1, player = "alice",
            })
            t:step()
            t:check(function()
                local entity = test.only_entity_with("d2legacy.world.velocity")
                local velocity = require("engine.ecs/v1").get(entity, "d2legacy.world.velocity")
                test.expect(velocity:get("x"), "horizontal velocity"):equals(15)
            end)
        end),
    },
})
```

Test code follows the same living-documentation standard as production Lua.
Use named local helpers for fixture construction and multi-part assertions,
keep each suite focused on its sibling module or one cross-domain scenario,
share repeated semantic fixtures through `d2legacy.tests.support`, and keep
lines at or below 120 columns. Do not compress suites into generated one-line
tables merely because Lua accepts them.

The builder supports `arrange`/`run`, `check`, `step`, `update`, `command`,
`system_command`, `restore_checkpoint`, and `expect_checkpoint_parity`.
Commands and suite `initial_data` and `records` are ordinary Lua tables; the
host owns the JSON boundary. `restore_checkpoint` takes the latest checkpoint, tears down
the authority, reconstructs the ECS and every deterministic participant, and
continues the remaining Lua-authored phases in the replacement runtime. A suite
may also provide `seed`, `initial_data`, and `records`. Large bounded
offline vectors may set `disable_execution_budget = true`; ordinary gameplay
tests must retain the production deadline. Host inputs are copied before each case boots. Every case gets a
fresh ECS engine, session, deterministic streams, state store, component tree,
and Lua VM through the same `StartWithConfig` composition root used by the
headless server.

Profiles declare the production boundary under test: `authority`, `policy`,
`presentation`, `client`, or `real_assets`. The default is `authority`. Tiers
are `fast`, `integration`, `real_assets`, and `stress`; normal `go test` runs
`fast,integration`. Select tiers with `DARK_MAGIC_LUA_TEST_TIERS`, and repeat
every isolated case with `DARK_MAGIC_LUA_TEST_REPEAT` to expose hidden state
dependencies. `DARK_MAGIC_LUA_TEST_ORDER_SEED` shuffles case discovery with a
reproducible seed. Use `test.array()` when structured input needs an empty JSON
array; an unmarked empty Lua table is an object. Use `test.property` for deterministic input matrices and
`test.expect` for labeled, path-aware failures.

The phased contract is intentional. Calling `session.Step` from a Lua callback
would re-enter the serialized runtime while its owner goroutine was already
executing. Go therefore performs host actions between callbacks, while the
complete scenario and all gameplay assertions remain authored in Lua.

Run every suite with:

```text
go test ./internal/mod/d2legacy -run TestLuaSuites
```

Select one suite or case with the normal Go test path, for example:

```text
go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/mitigation'
```

CI can repeat each case or opt into expensive tiers:

```text
DARK_MAGIC_LUA_TEST_REPEAT=3 go test ./internal/mod/d2legacy -run TestLuaSuites
DARK_MAGIC_LUA_TEST_TIERS=fast,integration,real_assets go test ./internal/mod/d2legacy
```

Go remains responsible for runtime construction, checkpoint serialization,
reconstruction, native adapters, and resource lifetime. Presentation and
composition integration drivers remain in `internal/mod/d2legacy` or
`internal/acceptance`, but their Lua programs live under `tests/integration/`
and execute with `Runtime.Execute` or `Runtime.ExecuteScoped`. This keeps native
fixtures and assertions at their appropriate boundary without embedding Lua in
Go strings. Generic runtime tests under `internal/runtime/lua`
deliberately boot synthetic mods instead. The checked coverage ledger at
`docs/architecture/d2legacy-test-coverage.tsv` records both Lua policy evidence
and the Go integration evidence that must remain at the host boundary.
