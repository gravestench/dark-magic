# d2legacy mod architecture

This document describes the production architecture of the bundled `d2legacy`
mod. It is a map for contributors: where a rule belongs, which runtime owns it,
how ECS data moves through systems, how presentation observes authority, and
where tests execute.

For a smaller starting point, compare this mod with the sibling
[`mod_template`](../mod_template/README.md). For the Lua API tutorial and
recommended reading order, see [`lua/README.md`](lua/README.md). For detailed
test-harness syntax, see [`lua/d2legacy/README.md`](lua/d2legacy/README.md).

The installed package is privately addressable at `mods/d2legacy/`. Its
manifest explicitly projects `assets`, `data`, `locales`, and `manifests` into
the shared layered VFS; `boot.lua`, `components`, `lua`, and `mod.json` remain
namespaced and cannot collide with another mod. See the repository
[mod loading guide](../../../docs/MODS.md) for dependency and override order.

## Runtime boundaries

Dark Magic supplies mechanisms. The mod supplies meaning.

```text
native host
  mounts content and decoded records
  owns fixed ticks, ECS storage, checkpoints, rendering, audio, and I/O
       |
       +-- presentation runtime
       |     boot.lua
       |       -> scene registry -> screens / overlays / UI helpers
       |
       +-- authoritative runtime
             components/d2legacy.lua
               -> d2legacy.authoritative
                    -> components -> commands -> systems -> events/snapshots
```

Both runtimes execute Lua from this mod, but they have different authority:

- Presentation may create retained render nodes, navigate scenes, read copied
  snapshots, and submit intent. It must not change authoritative gameplay.
- Authoritative Lua may validate commands and mutate declared ECS state during
  deterministic phases. It must not depend on a renderer or local UI state.
- Go owns capability enforcement, resource lifetime, serialization, native
  adapters, and scheduling. Diablo-specific rules belong in Lua unless a
  native boundary is genuinely required.

An `engine.*/*` import is a versioned host capability. A `d2legacy.*` import is
ordinary mod-owned Lua. Do not import Go implementation details from mod code.

## Composition roots

### Presentation: `boot.lua`

The host discovers `boot.lua` and every top-level `components/*.lua` definition.
Discovery loads definitions but does not start them. The component manager then
creates scoped instances and calls their lifecycle methods.

`boot.lua` loads the presentation manifest, registers symbolic scene names,
creates the transition root, and selects the loading scene. It should remain
composition code; widgets and screen behavior belong in their own modules.

### Authority: `components/d2legacy.lua`

The authoritative host explicitly loads `components/d2legacy.lua`. Its small
lifecycle wrapper delegates to `d2legacy.authoritative`, which registers:

1. ECS component schemas;
2. immutable bootstrap data;
3. decoded record interpretations;
4. command handlers;
5. deterministic systems.

Registration order makes dependencies readable, while simulation phases and
declared system access determine execution. The composition root must not hide
formulas or gameplay mutations.

Immutable `d2legacy.game_rules/v2` contains session facts such as difficulty,
target, mode, and admission capacity. Mutable `/players X` policy is not a game
rule: `d2legacy.player_count/v1` checkpoints its optional override separately,
while consumers count present player entities by default. `maximum_players`
can reject entry but cannot become a monster or reward multiplier.

## Directory ownership

| Path | Responsibility |
| --- | --- |
| `boot.lua` | Presentation entry point and initial scene selection. |
| `components/` | Host-managed lifecycle definitions. |
| `lua/d2legacy/bootstrap/` | Composition, registries, and shared routing. |
| `lua/d2legacy/components/` | Durable ECS schemas only; no behavior. |
| `lua/d2legacy/commands/` | Validate admitted commands and create/change authoritative facts. |
| `lua/d2legacy/systems/` | Advance ECS state in fixed simulation phases. |
| `lua/d2legacy/policy/` | Pure formulas and decisions without scheduling or ECS queries. |
| `lua/d2legacy/data/` | Interpret decoded legacy records into reviewed definitions. |
| `lua/d2legacy/items/`, `loot/`, `owned_units/`, `mapgen/` | Larger rule domains split into focused modules. |
| `lua/d2legacy/gameplay/` | World presentation, snapshots, composites, and presentation systems. |
| `lua/d2legacy/screens/` | Root scene implementations. |
| `lua/d2legacy/overlays/` | Panels layered over root scenes. |
| `lua/d2legacy/ui/` | Reusable retained visuals and interaction controls. |
| `lua/d2legacy/tests/` | Harness API, shared fixtures, and cross-domain scenarios. |
| `manifests/` | Versioned, validated presentation and recovered-data facts. |
| `locales/` | Mod-owned localized strings. |

## ECS model

An entity is an identity with a set of typed components. Components contain
facts; systems give those facts behavior. A component schema is registered once
with `engine.ecs/v1` and then accessed through checked fields.

The main component families are:

| Family | Representative components | Meaning |
| --- | --- | --- |
| Player | `player.identity`, `player.progress`, `player.vitals`, `player.combat_stats` | Durable identity, progression, resources, and resolved combat values. |
| World | `world.position`, `world.velocity`, `world.location`, `world.collider`, `world.selectable` | Spatial facts shared by players, monsters, missiles, and interactions. |
| Combat | `combat.melee_profile`, attack requests/animations/events, defense | Inputs, scheduled actions, and factual outcomes. |
| Skills/states | learned skills, assignments, cast requests/events, timed states | Skill admission and deterministic lifecycle state. |
| Monsters | identity, stats, AI | Authored monster facts and current decision state. |
| Items | identity, placement, melee/armor facts, stat modifiers, layouts | Item provenance, container ownership, equipment, and generated properties. |
| Relations | stat sources, owned units | Explicit contributions and ownership rather than inferred proximity. |
| Quests/world transitions | quest state, requests, portals, transition state | Durable progression and cross-zone transactions. |

Component design rules:

- Store authoritative mutable state in ECS or another registered checkpoint
  participant, never in a module global or presentation object.
- Use explicit identities and relations. Do not rediscover ownership by scanning
  names, positions, or appearance.
- Preserve base facts separately from derived values. Named stat sources can
  then be added and removed without reconstructing history.
- Record units in names or comments when ambiguous: fixed-point raw values,
  world subtiles, ticks, percentages, and record IDs are not interchangeable.
- Events are immutable semantic facts. Requests are pending work. Components
  representing an active action are checkpointed state.

Schema files divide that catalog further: `shared.lua` owns player, monster,
world, defense, state, and generic stat-source facts; `skill_actions.lua` owns
generic cast and projectile facts; `melee.lua` owns approach, animation, impact requests, and
outcome events; `items.lua` owns layouts, item identity/placement/presentation,
equipment facts, modifiers, and vendor/service state; `owned_unit.lua` owns
explicit summon/hireling relations; and `quests.lua` owns quest definitions,
progress, requests, and events. Open the schema file before changing a system
so new state joins the correct lifetime and checkpoint contract.

## Systems and scheduling

A system declares a stable ID, phase, query, read set, write set, and update
function. The broad query supplies a deterministic entity view for that tick;
systems must not perform hidden world queries during iteration.

The engine's complete phase order is:

```text
input -> intent -> pre_simulation -> movement -> collision -> combat
      -> effects -> inventory -> presentation -> cleanup
```

d2legacy currently uses `intent` for AI decisions, `pre_simulation` for cast
and attack admission plus equipment/stat projection, `movement` and `collision`
for spatial work, `combat` for impacts, `effects` for death/progression/state
lifecycle, and `presentation` for safe derived-facing facts. Empty phases remain
real barriers available to future systems.

Use the phase for coarse ordering and explicit request/event components for
cross-system handoff. Do not depend on incidental Lua module load order.

### Authoritative system catalog

| System module | Phase(s) | Individual responsibility |
| --- | --- | --- |
| `cast` | `pre_simulation` | Advances admitted skill casts through start/effect/complete ticks. |
| `missile_skill` | `pre_simulation` | Turns a definition-selected straight-missile effect into an authored projectile. |
| `state_skill` | `pre_simulation` | Turns a definition-selected self-state effect into an authored timed state and removable stat source. |
| `player_melee` | `pre_simulation` | Owns approach, swing timing, selected hand, and impact requests. |
| `equipment` | `pre_simulation` | Projects the active equipment set into melee and named stat sources. |
| `derived_stats` | `pre_simulation` | Rebuilds effective stats from durable bases and removable sources. |
| `reactive_state` | `effects` | Converts factual melee hits into generic reactions declared by active state instances. |
| `timed_state` | `effects` | Applies, refreshes, group-replaces, and expires state instances with exact stat-source cleanup. |
| `monster_ai` | `intent` | Chooses deterministic monster intent from authoritative world facts. |
| `projectile` | `movement`, `combat` | Moves projectiles, selects first contact, resolves damage, and expires them. |
| `world_transition` | `collision` | Validates and executes spatial transition requests. |
| `melee` | `combat` | Selects eligible targets and resolves hit, mitigation, health, and events. |
| `monster_death` | `effects` | Converts lethal monster state into corpse, credit, XP, and loot consequences. |
| `progression` | `effects` | Applies experience and level progression definitions. |
| `owned_unit_lifecycle` | `effects` | Enforces ownership limits, expiry, transitions, and unsummon policy. |
| `facing` | `presentation` | Derives presentation-facing direction from authoritative motion. |

Commands form the admission edge for those systems. `enter_player` and
`spawn_monster` materialize trusted entities; `cast` and `move_player` admit
player action; the item command modules own movement, vendor, interaction, and
service transactions; quest and owned-unit commands validate their respective
domain transitions. Commands should finish immediately or create explicit
state for a system to finish later.

### Example: player Basic Attack

```text
UI submits player.use_skill
  -> commands/cast.lua validates assignment and writes skill.cast_request
  -> systems/cast.lua admits exact skill ID 0 at its row-derived zero mana cost
  -> systems/melee_skill.lua emits the generic action.melee effect
  -> systems/player_melee.lua selects pinned AnimData timing by actor/mode/weapon
     and owns approach plus attack-animation state
  -> impact creates combat.basic_attack_request
  -> systems/melee.lua resolves range, AR/defense, hit, damage, and health
  -> combat.melee_event records the outcome for presentation/diagnostics
```

Attack is a configuration of the shared skill transaction, not a command-level
exception. Its exact target row is admitted by the same target-locked behavior
manifest as missile and state skills. Approach, weapon-hand selection, action
timing, and melee resolution remain reusable mechanisms selected by the
definition's behavior family.

Equipment does not special-case the melee resolver. `systems/equipment.lua`
projects active item facts into the melee profile and named stat sources;
`systems/derived_stats.lua` resolves those sources before combat. This keeps
inventory location, stat arithmetic, and hit resolution independently testable.

### Commands versus systems versus policy

- A command answers: “is this externally requested transition allowed?”
- A system answers: “what advances on this deterministic tick?”
- A policy module answers: “given these values, what is the result?”

If a function needs neither ECS nor scheduling, put it in `policy/` and test it
with the `module` profile. If it interprets legacy table strings, put that
translation in `data/`. Keeping these seams narrow prevents tests from building
shadow implementations of production behavior.

## Checkpoints and deterministic behavior

The authoritative composition registers ECS, random streams, runtime identity,
and other state participants with the session. Restore constructs the same
runtime, registers the same schemas and systems, restores participants, and
continues on the fixed timeline.

Consequences for mod code:

- use purpose-named deterministic random streams;
- never use wall-clock time for gameplay;
- do not retain mutable gameplay facts only in Lua locals between ticks;
- give sources, requests, entities, and systems stable identities;
- test checkpoint parity for behavior that spans multiple ticks.

## Presentation and UI

Scenes are Lua tables with only the lifecycle callbacks they need: `create`,
`enter`, `update`, `render`, `exit`, and `destroy`. State stored on `self` belongs
to that scene instance. Checked render/audio/subscription handles belong to the
active component scope and are reclaimed when the scope closes.

Root screens use `engine.scene/v1` replacement or stack navigation. Gameplay
panels are overlays assigned to `left`, `right`, or `full` slots. Overlay policy
separately controls whether lower scenes update, receive input, and how the
world camera is framed.

UI has two deliberately separate layers:

- `ui.controls` owns focus, hit testing, pointer capture, keyboard navigation,
  editing, value changes, and accessibility state.
- Widgets such as `button`, `slider`, `item_grid`, and `text_field` own retained
  nodes and visual reactions to control state.

A UI path should be one-way:

```text
authoritative snapshot -> scene/widget draws facts
player interaction     -> intent/command submission
next authoritative snapshot -> scene/widget reflects accepted result
```

Never let a widget mutate an authoritative component. Conversely, authoritative
systems must not wait for an animation or renderer callback before progressing.

Presentation facts such as asset paths, palettes, layout, timings, and string
keys belong in versioned manifests when practical. Lua should describe how
those facts compose and behave.

## Test architecture

Production Lua tests are ordinary `*_test.lua` modules beside the code they
specify. `internal/mod/d2legacy/lua_suite_test.go` discovers them from the
embedded production tree and exposes every Lua case as a Go subtest. The test
scenario remains Lua-authored; Go owns runtime construction and host actions
between callbacks.

Profiles select the smallest production environment needed:

| Profile | Environment | Use for |
| --- | --- | --- |
| `module` | records, content, deterministic helpers; no ECS | Pure policy and data interpretation. |
| `ecs` | module profile plus production ECS schemas | Component contracts and focused ECS behavior. |
| `authority` | complete renderer-free d2legacy authority | Commands, systems, and vertical gameplay scenarios. |

Tiers describe cost/resources: `fast`, `integration`, `real_assets`, and
`stress`. Normal `go test` includes fast and integration suites. Native Go
integration tests remain appropriate for renderers, filesystem/native adapters,
real assets, process lifetime, and cross-runtime host composition.

### Writing a test

1. Place `foo_test.lua` beside `foo.lua` unless it is genuinely cross-domain.
2. Select the narrowest profile.
3. Build meaningful records or initial data; do not reimplement production
   constructors in the test.
4. Arrange commands and fixtures, advance explicit ticks, then inspect public
   ECS facts or semantic events.
5. Extract repeated setup and assertions into named helpers so the case reads
   as arrange, act, assert.
6. Add checkpoint parity when behavior survives across ticks or restoration.

```lua
local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("describes one rule", {
            test.run(function()
                local policy = require("d2legacy.policy.example")
                test.expect(policy.resolve(2, 3), "resolved value"):equals(5)
            end),
        }),
    },
})
```

Run focused and complete checks with:

```sh
go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/path/foo'
make test-lua
make test-lua-hardening
go test ./...
```

The coverage ledger at `docs/architecture/d2legacy-test-coverage.tsv` records
which behavior is proven in Lua and which evidence must remain at a native host
boundary.

## Adding a feature

Use this dependency direction:

```text
records/manifests -> data interpretation -> policy
                                      \-> command/system composition
components --------------------------------------^
authoritative ECS/events -> copied presentation snapshot -> UI
```

A typical contribution should:

1. document the authoritative facts and units;
2. add or extend focused component schemas;
3. interpret external records once in `data/`;
4. keep formulas in pure `policy/` modules;
5. admit external intent through `commands/`;
6. advance multi-tick behavior through declared systems;
7. expose semantic outcomes rather than presentation guesses;
8. add sibling Lua tests and only the native integration tests required by the
   boundary;
9. register the new modules visibly in the appropriate composition root.

If a change cannot be explained cleanly in those layers, pause and make the
ownership boundary explicit before adding more code.
