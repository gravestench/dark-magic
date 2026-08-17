# d2legacy authoritative gameplay

The broader architecture—including presentation composition, ECS domain maps,
system scheduling, UI ownership, checkpoint behavior, and feature placement—is
documented in [`../../ARCHITECTURE.md`](../../ARCHITECTURE.md).

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

## Straight-missile skill execution order

1. `manifests/skill-behavior-coverage.v1.json` admits exact skill IDs to reviewed
   reusable families for the Expansion 1.14d target.
2. Go admits `d2legacy.skill.cast` by identity, tick, sequence, and authority.
3. `commands/cast.lua` validates its payload and creates a cast request.
4. `data/missile_skills.lua` validates each explicitly supported 1.14d record
   pair into one immutable `missile.straight` definition.
5. `systems/cast.lua` validates learned skill, target, and mana against the
   selected definition, then schedules its effect and completion ticks.
6. `systems/missile_skill.lua` creates the configured projectile when the
   effect tick arrives.
7. `systems/projectile.lua` moves it, finds first contact, and expires it.
8. `policy/damage.lua` resolves its configured damage channel and consequences.

Fire Bolt is currently the first supported definition and integration fixture.
It is not an authority boundary: new verified straight-missile configurations
join the same modules, while missiles with different impact/motion behavior
receive another reusable family rather than a per-skill system.

Run `MPQ_DIRECTORY=/path/to/mpqs make skill-behavior-coverage` to print the
target-locked server-function/missile linkage inventory. The report never
infers support from matching row shapes: each supported consumer must have its
own exact-ID declaration and evidence status. Generated reports stay local.

Run `MPQ_DIRECTORY=/path/to/mpqs make skill-evidence` to join the default exact
skill IDs to SkillDesc and English base/Expansion/patch TBL text. Override
`SKILL_IDS=...` for the skill under investigation or pass another locale to the
underlying tool. The report retains localized replacement tokens and resolves
cross-skill `.blvl`/`.lvl` formulas to exact IDs; it is required evidence for
synergies and skills that modify other skills, not a gameplay formula engine.

All mutable facts live in ECS or registered engine state. Random rolls use a
purpose-named engine stream, so replay and checkpoint restore reproduce them.

## Melee-action skill execution order

1. The target-locked behavior manifest admits exact skill ID 0 to
   `action.melee`; matching function IDs never admit another skill.
2. `data/melee_skills.lua` validates the owned Expansion 1.14d Attack row into
   a zero-mana immutable definition.
3. `commands/cast.lua` emits the same cast request used by other skill families,
   and `systems/cast.lua` verifies the learned level and zero resource cost.
4. `systems/melee_skill.lua` emits the configured action effect.
5. `systems/player_melee.lua` owns approach, selected hand, and the impact/
   completion schedule read from the session-pinned composite AnimData record;
   `systems/melee.lua` owns factual hit and damage resolution.

Attack is therefore a skill configuration, not a special command path. Exact
target/range/LOS policy remains incomplete; other melee skills need their own
reviewed declaration and behavior evidence.

## Timed self-state/stat execution order

1. The target-locked behavior manifest admits one exact skill ID.
2. `data/state_skills.lua` validates the owned server-function, state/stat, cost,
   level, and hard-point-synergy fields into a `state.self-timed-stat` definition.
3. The shared cast lifecycle validates the learned level and available mana;
   funded casts pay exactly once, while underfunded requests have no effect and
   preserve the partial balance.
4. `systems/state_skill.lua` computes the authored duration/stat value and emits
   one source-tagged state request at the effect tick.
5. `systems/timed_state.lua` applies or refreshes the state and its named stat
   source together; state-group replacement, explicit removal, or expiration
   removes that exact source.
6. `systems/reactive_state.lua` turns a factual successful melee hit into the
   active state's configured response, never a skill-name callback.
7. Generic disabled-action state facts stop monster AI and motion until the
   timed state expires; `systems/derived_stats.lua` independently rebuilds
   effective defense.

Frozen Armor is the first configuration. Its defense, armor duration, PvM melee-
hit freeze response, difficulty divisor, and cold-armor exclusion are backed by
owned Expansion 1.14d rows and Blizzard's official table. PvP chill conversion,
target cold resistance/immunity and monster-class duration modifiers, exact
integer/tick ordering, presentation, and cast animation timing remain explicitly
incomplete.

## Player population and `/players X`

`policy/game_rules.lua` checkpoints immutable expansion 1.14d session facts.
Its `maximum_players` field is only the admission cap. It is deliberately not
an effective gameplay player count.

`policy/player_count.lua` owns the separate `d2legacy.player_count/v1` state.
Without an override, monster and reward consumers use the number of present
authoritative player entities at the moment their policy runs. The privileged
`game.player_count.override` command forces `/players X` behavior from 1 through
8 without changing admission; `game.player_count.follow_population` returns to
live join/leave behavior. Monsters snapshot this as `spawn_player_count`, while
death/NoDrop facts retain live, effective, nearby-party, spawn, and final
eligible counts separately.

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
tests must retain the production deadline. Host inputs are copied before each
case boots. Every case gets a fresh ECS engine, session, deterministic streams,
and Lua VM. Authority cases use the same `StartWithConfig` composition root as
the headless server.

Profiles enforce the production boundary under test:

- `module` installs content loading, records, deterministic helpers, worldgen,
  initial data, and checkpoint-compatible random streams, but no ECS or command
  authority.
- `ecs` adds the production ECS capability and production shared component
  schemas, but no commands or authoritative systems.
- `authority` boots the complete renderer-free production authority.

Asset-backed renderer and interactive-client tests remain native integration
tests because a headless Lua suite must not claim it booted those environments.
Tiers are `fast`, `integration`, `real_assets`, and `stress`; normal `go test` runs
`fast,integration`. Select tiers with `DARK_MAGIC_LUA_TEST_TIERS`, and repeat
every isolated case with `DARK_MAGIC_LUA_TEST_REPEAT` to expose hidden state
dependencies. `DARK_MAGIC_LUA_TEST_ORDER_SEED` shuffles case discovery with a
reproducible seed. Use `test.array()` when structured input needs an empty JSON
array; an unmarked empty Lua table is an object. Use `test.property` with
`samples`, explicit `seeds`, and the composable `integer`, `one_of`, `map`, and
`tuple` generators for deterministic input matrices. Use `test.expect` for
labeled, path-aware failures. Install narrow dependency doubles with
`test.mock_module`; it validates the declared function contract and fails on
unexpected reads instead of silently returning `nil`.

The phased contract is intentional. Calling `session.Step` from a Lua callback
would re-enter the serialized runtime while its owner goroutine was already
executing. Go therefore performs host actions between callbacks, while the
complete scenario and all gameplay assertions remain authored in Lua.

## Lua style guide

Lua tests and production modules are living documentation. A reader should be
able to open one file and understand one idea without memorizing Lua syntax.

- Keep every line at or below 120 columns.
- Prefer short declarations, one idea per line, and whitespace over dense
  inline tables.
- Use explanatory comments for legacy concepts, state lifetimes, units, and
  ordering rules. Explain *why* a value matters, not just what the syntax does.
- Extract repeated setup and assertions into named local helpers. Tests should
  read like a short story: arrange, act, assert.
- Break long payloads and record tables into smaller named pieces or helper
  functions. Do not compress suites into giant one-line tables merely because
  Lua accepts them.
- Keep test data semantically meaningful. Named fixtures belong in
  `d2legacy.tests.support`; scenario-specific setup belongs in short local
  helpers beside the test that uses them.
- Use `test.expect` and `test.property` for labeled, path-aware failures when
  the assertion is part of the documented contract.

Run every suite with:

```text
go test ./internal/mod/d2legacy -run TestLuaSuites
make test-lua
```

Select one suite or case with the normal Go test path, for example:

```text
go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/mitigation'
```

The same format, syntax, repeat, and seeded-order checks required by CI run with
`make test-lua-hardening`. The `integration` tier contains multi-system
authority scenarios; `real_assets` and `stress` are opt-in classifications and
must not be assigned until a suite actually supplies those resources or load.

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
