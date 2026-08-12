# d2legacy authoritative gameplay

This tree is the first-party Diablo II rules mod. Its canonical name and Lua
namespace are `d2legacy`. The repository's “Dark Magic shim” is the bundle that
contains this mod; it is not a second mod with a separate gameplay identity.

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
