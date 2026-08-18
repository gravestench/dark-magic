# Dark Magic roadmap

Dark Magic is building an increasingly complete, clean-room implementation of
**Diablo II: Lord of Destruction 1.14d** on a deterministic, scriptable engine.
The project is past its initial engine and networking foundations; current work
is concentrated on completing authoritative gameplay without creating parallel
offline, server, or skill-specific implementations.

## Where progress is tracked

Execution state no longer lives in this file:

- the [Dark Magic Roadmap project](https://github.com/users/gravestench/projects/1)
  is the live view of priority, ownership, scheduling, and delivery;
- [GitHub issues](https://github.com/gravestench/dark-magic/issues) define
  actionable acceptance slices, research probes, and unresolved behavior; and
- [GitHub milestones](https://github.com/gravestench/dark-magic/milestones)
  group issues into outcome-level gameplay gates.

Issue state and acceptance evidence are authoritative. A milestone is not
complete merely because related pull requests merged, and a pull request should
link the issue it advances. The Roadmap project's fields organize the work; they
do not replace the issue's completion rule.

This document is intentionally a stable, high-level overview. It does not carry
task checklists, per-slice evidence ledgers, completion counters, target dates,
or a "next implementation cursor." Those details change frequently and belong
in the live trackers above. Edit this file only when the product boundary,
technical direction, or high-level program shape changes—not when an issue is
started, completed, deferred, or reordered.

## Product boundary

Dark Magic has one supported behavioral target: the latest original Diablo II
expansion release, **Lord of Destruction 1.14d**.

- Classic mode and behavior from earlier patches are out of scope.
- Diablo II: Resurrected behavior is not silently back-ported into the target.
- Vanilla client/server protocol compatibility is not a goal.
- Vanilla save-game import/export is not a goal.
- Compatibility with old community tools or their private data formats is not a
  goal.
- Dark Magic's own modding, session, and durable-character formats may evolve as
  versioned native contracts; they do not need to imitate legacy formats.

The project does not distribute Blizzard assets. Real-asset clients, labs, and
verification tools require legally obtained Diablo II data.

## Technical direction

The implementation is organized around a few durable constraints:

- **One authoritative simulation.** Offline play, listen servers, dedicated
  servers, replay, checkpointing, and Realm-managed games share gameplay
  mechanisms and command boundaries.
- **Lua owns Diablo policy; Go owns reusable mechanisms.** The first-party
  `d2legacy` package interprets Diablo records and rules. Go owns the engine,
  ECS, transport, persistence, rendering, audio, and lifecycle infrastructure.
- **ECS facts compose behavior.** Combat results, stat sources, states,
  relationships, motion, residency, and observation markers remain reusable
  facts instead of becoming special callbacks for individual skills or units.
- **Authored data drives content.** TXT records, linked identifiers, locale TBL
  strings and replacement tokens, animation data, sounds, overlays, and asset
  membership are joined before behavior is inferred or implemented.
- **Behavior families precede content breadth.** Representative skills,
  missiles, auras, curses, objects, items, and monsters validate reusable
  mechanisms. A named content record configures a family; it should not create
  a one-off runtime authority.
- **Determinism and bounded projections are product requirements.** Commands,
  RNG, checkpoints, reconnects, and player-scoped views must remain testable and
  reproducible as gameplay breadth grows.

## High-level progress

The engine/application foundation, layered content system, Lua policy boundary,
ECS simulation, real-asset presentation, deterministic session/replay path,
network admission, reconnect, Realm control-plane foundation, and native client
labs are established. The closed and active GitHub milestones contain the exact
acceptance record.

Current development is centered on combat fidelity and reusable
skill/state/missile behavior families. Existing vertical slices already
exercise authoritative movement, world residency, monsters, melee and missile
damage, death attribution, skills, states, auras, transitions, loot, items,
quests, checkpointing, and connected presentation. These slices prove the
architecture; they do not imply complete Diablo II content or exact behavior in
every unresolved ordering case.

The broader program proceeds from combat and skills into the item lifecycle,
kill-to-ground-item acceptance, reusable world objects, monster quality and
bosses, a coherent Act I, durable native characters, trade, hirelings and owned
units, the economy, and remaining campaign/class/UI/audio breadth. The Roadmap
project and milestones determine the actual ordering and current status.

## Acceptance and evidence

Behavior claims use explicit confidence labels:

- **verified**: reproduced against an owned Lord of Destruction 1.14d runtime or
  directly established by owned target data;
- **high-confidence recovered**: strongly supported by pinned recovered code or
  multiple independent sources, but not yet confirmed against the target
  runtime;
- **inferred**: the best current interpretation of incomplete evidence;
- **synthetic Dark Magic policy**: deliberate scaffolding used to validate an
  engine mechanism without claiming retail behavior; and
- **unresolved**: intentionally blocked on a specific probe or source conflict.

An implementation issue should identify the behavioral outcome, source and
confidence, existing authority, acceptance tests, and explicit deferrals. Pull
requests should update durable research or architecture documents when their
claims change, while progress, follow-ups, and remaining tasks stay on the
linked issue and milestone.

The long-term acceptance loop remains:

```text
Realm allocates one pinned game
  -> multiple authenticated players join one authoritative Session
  -> immutable rules plus checkpointed policies govern the world
  -> locomotion, combat, party context, loot, items, objects, and quests run
  -> reconnect reproduces player-scoped state
  -> one revisioned native character commit succeeds
```

## Related documentation

- [README.md](README.md) — product and repository orientation
- [CONTRIBUTING.md](CONTRIBUTING.md) — issue-first contribution workflow
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — package ownership and dependency
  rules
- [docs/GAMEPLAY_OWNERSHIP.md](docs/GAMEPLAY_OWNERSHIP.md) — Go/Lua gameplay
  authority
- [docs/research/GAME_SYSTEMS_INDEX.md](docs/research/GAME_SYSTEMS_INDEX.md) —
  gameplay research and evidence map
- [docs/research/SYSTEMS_SOURCE_MATRIX.md](docs/research/SYSTEMS_SOURCE_MATRIX.md)
  — source provenance and confidence
- [docs/realm/README.md](docs/realm/README.md) — Realm, server, and persistence
  direction
