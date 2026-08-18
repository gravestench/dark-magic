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
   pair into one immutable `missile.straight` definition, preserving authored
   fields such as `Missiles.KnockBack` without assigning unverified semantics.
5. `systems/cast.lua` validates learned skill, target, and mana against the
   selected definition, then schedules its effect and completion ticks.
6. `systems/missile_skill.lua` creates the configured projectile when the
   effect tick arrives.
7. `systems/projectile.lua` moves it, finds first contact, and expires it.
8. `policy/damage.lua` resolves its configured damage channel and consequences.

Fire Bolt remains a production acceptance fixture, not a standalone
implementation. Straight, radial, area-impact, and on-hit-state missile records
compose the same cast, projectile, contact, damage, and timed-state mechanisms;
only genuinely different motion or impact shapes receive another family.

## Golem-family summon execution order

1. The target-locked manifest admits Clay, Blood, Iron, and Fire Golem together
   to `summon.golem`; a matching server function never admits another record.
2. `data/golem_summon_skills.lua` joins all four `Skills.txt` rows with Golem
   Mastery, Summon Resist, PetType, SkillDesc synergy keys, and Fire Golem's
   granted Holy Fire row. It validates the family as one modifier graph.
3. `systems/cast.lua` applies shared learned-skill, target, and fixed-point mana
   policy. Hard-point levels remain separate from effective cast levels so
   equipment bonuses do not become synergy points.
4. Point targets and Iron Golem's ground item pass the same resolver before
   mana payment and at the effect tick. An invalidated item cast keeps its paid
   mana but neither consumes the item nor replaces the existing golem.
5. `systems/golem_summon_skill.lua` materializes an ordinary friendly monster,
   applies record-derived life, offense, defense, resistance, absorb, movement,
   ownership, and PetType replacement facts, then commits replacement and item
   consumption as one success path.
6. Durable ECS reaction, intrinsic-stat, periodic-damage, and item-provenance
   components let generic consumers execute Clay slow, Blood healing exchange,
   Iron item properties, thorns, Fire absorb, and Holy Fire without branching
   on a skill name or numeric ID.

The family acceptance test casts all four members through Spell Lab's production
assignment and command path. Decoder tests validate the four definitions
together so a change to one cross-skill modifier cannot pass in isolation.

Run `MPQ_DIRECTORY=/path/to/mpqs make skill-behavior-coverage` to print the
target-locked server-function/missile linkage inventory. The report never
infers support from matching row shapes: each supported consumer must have its
own exact-ID declaration and evidence status. Generated reports stay local.

Run `MPQ_DIRECTORY=/path/to/mpqs make skill-evidence` to join the default exact
skill IDs to SkillDesc and English base/Expansion/patch TBL text. Override
`SKILL_IDS=...` for the skill under investigation or pass another locale to the
underlying tool. The report retains localized replacement tokens and resolves
generic cross-skill formula selectors such as `.blvl`, `.lvl`, `.edns`, and
`.edmn` to exact IDs; it is required evidence for synergies and skills that
modify other skills, not a gameplay formula engine.

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

## Reactive melee-reflection execution order

1. `systems/melee.lua` resolves an attack into co-composed factual
   `attack_result`, `melee_event`, `combat.event`, and `damage_bundle`
   components.
2. `systems/reflected_damage.lua` runs before generic death observation and
   selects only successful melee results with committed physical damage.
3. It immediately adds the empty `reflection_observed` component. Independent
   consumers can retain the same result entity without allowing reflection to
   run twice after another simulation step or checkpoint restore.
4. The system resolves `thorns_percent` from the defender's ordinary stat
   sources. It does not inspect a selected skill ID or English skill name.
5. The reflected basis is post-defender-mitigation physical damage capped by
   the damage actually committed to life. The configured percentage is then
   resolved as a new physical hit against the attacker through the shared
   damage policy, including the attacker's mitigation.
6. The returned hit emits an ordinary combat result with the defender as its
   source, so the existing death consumer owns lethal attribution.

This is currently a PvM boundary. Player attackers are excluded until
Expansion 1.14d hostility, hireling classification, and the documented reduced
PvP return are verified. Missiles and non-physical melee damage do not satisfy
the factual input. The front/back aura presentation is shared with other aura
states; the distinct `hit_thorns` reaction overlay remains a presentation
follow-up.

## Periodic selected-aura execution order

1. The target-locked manifest admits an exact skill ID to
   `aura.selected-party-periodic`; matching server functions do not admit other
   skills.
2. `data/periodic_aura_skills.lua` validates the owned aura, state, maintained-
   stat/direct-effect columns, fixed-point cost, radius, target filter, and
   period fields into immutable facts.
3. `systems/aura_skill.lua` reuses the selected-right emitter and current
   party/radius relationships, then co-composes one durable `aura_pulse`
   schedule and ordered `aura_pulse_effect` entities. Selection changes remove
   the whole set through the same source lifecycle.
4. Maintained columns become ordinary keyed `stat.source` entities. Periodic
   columns remain ordered `aura_pulse_effect` entities on the same definition.
5. `systems/aura_pulse.lua` advances the checkpointed schedule, gathers the
   emitter's eligible relationships in stable player order, and requires the
   entire authored mana cost before changing any target.
6. Each target receives the effects in authored column order. Direct healing
   clamps to maximum life; duration transforms reschedule the current remaining
   lifetime using deterministic integer arithmetic.
7. A funded schedule spends mana once only when at least one value changes. An
   underfunded or all-full-health Prayer pulse consumes nothing;
   `policy/resources.lua` keeps that transaction shared with ordinary casts.
   Useful paid pulses own a mana-regeneration suppression relationship until a
   later pulse clears it or selection removes its source.
8. `systems/mana.lua` consumes class `ManaRegen`, resolved percentage/flat
   sources, and suppression relationships once per 25 Hz tick. It owns fixed-
   point integer ordering and clamping; aura systems never add mana directly.

Prayer supplies one paid healing effect. Cleansing supplies a free ordered pair:
current-duration scaling for poison, curable curses, and the officially
documented `shrine_*` family, followed by healing derived from the owner's
learned Prayer level. Meditation combines a maintained `manarecoverybonus`
source with the same free learned-Prayer pulse; its owned 73729 filter and the
official reference both exclude hirelings. Other 73731 auras currently stop at
living same-level player party members. Exact 1.14d pulse/shrine/stat-regeneration
event ordering and selection-switch timing remain explicit evidence gaps.

## Corpse-periodic selected-aura execution order

1. The manifest admits an exact skill ID to
   `aura.selected-corpse-periodic`; sharing server-do 82 never enables another
   record.
2. `data/corpse_aura_skills.lua` validates the owned owner-state, corpse filter,
   radius/chance/recovery formulas, zero cost, and period into a generic target
   policy plus ordered operations.
3. Monster construction projects `MonStats2.corpseSel` into the empty
   `monster.corpse_selectable` capability. Death separately creates mutable
   `corpse_usable` state, so authored eligibility and one-time consumption are
   not conflated.
4. `systems/aura_pulse.lua` excludes town and enumerates usable same-level,
   in-radius, active corpses in stable spawn-ID order. A purpose-named
   checkpointed stream rolls the evaluated chance once for each candidate.
5. Each success applies its definition's ordered owner-resource operations,
   clamps through shared resource policy, and then consumes the corpse. Full
   resources do not prevent consumption because success is about redeeming the
   corpse, not whether either resource changed.
6. The committed operation emits `skill.aura_pulse_event`, a durable semantic
   result that presentation can map to the owned target assets without reading
   internal death/resource components or rerunning gameplay.

Redemption is the first admitted configuration. Its exact Expansion 1.14d
records and localized TBL intent pin the behavior shape and level vectors;
owned State/Overlay/Missiles assets pin presentation vocabulary separately.
The semantic event is not yet rendered, and exact radius units, target-runtime
RNG ordering, and same-tick corpse eligibility remain probe-gated.

## Timed self-state/reactive execution order

1. The target-locked behavior manifest admits exact skill IDs; a shared decoder
   joins Skills, States, Missiles, SkillDesc, and localized TBL evidence before
   constructing a `state.self-timed-reactive` definition.
2. The cast lifecycle validates assignment, learned level, and mana. A funded
   cast pays once; an underfunded request creates neither a cast nor an effect.
3. `systems/state_skill.lua` evaluates effective-level defense and duration
   separately from authored `.blvl` hard-point synergies, then snapshots the
   reaction values into a source-tagged state request.
4. `systems/timed_state.lua` refreshes a same-source state or replaces another
   member of its States group. The state and exact named stat source share one
   lifecycle across replacement, removal, expiration, and checkpoint restore.
5. `systems/reactive_state.lua` observes immutable combat results and selects a
   decoded reaction by record event/function pair. It never dispatches by skill
   name or numeric ID.
6. Direct reaction damage uses the common mitigation and combat-event path.
   Chilling Armor's return bolt uses the ordinary projectile materializer and
   collision path. Authored state overlays, reaction overlays, and sounds leave
   authority as bounded semantic presentation cues.

Frozen Armor, Shiver Armor, and Chilling Armor are one admitted family. Their
owned Expansion 1.14d rows define defense, duration, mana, mutual exclusion,
hard-point synergies, event triggers, cold damage, and return-missile assets.
Frozen Armor requires a successful damaging melee hit; Shiver Armor reacts to a
melee attempt even when it misses; Chilling Armor reacts only to a missile hit.
Players and monster classes carrying the empty `monster.freeze_immune` ECS
capability receive chill instead of a hard freeze. Cold resistance changes
effect length, immunity suppresses both damage and chill, and monster cold
lengths use the immutable Normal/Nightmare/Hell divisor. Exact action-frame
timing and any unrepresented monster-quality source remain explicit follow-up
work rather than hidden skill-specific exceptions.

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
