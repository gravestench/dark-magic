# Diablo II gameplay systems research program

Status: research baseline complete / implementation and verification program active. This document is not a runtime API contract.

All 28 workstreams in [GAME_SYSTEMS_INDEX.md](GAME_SYSTEMS_INDEX.md) now have implementation-oriented baseline documents. That means ownership, major state/data relationships, implementation slices, and explicit unknowns are documented. It does **not** mean all legacy behavior is implemented or empirically validated.

Dark Magic also moved substantially while this research was being assembled: the authoritative simulation now includes shared stat-source provenance, fixed-point combat vocabulary, ordinary hostile materialization, scheduled basic monster AI, a semantic melee transaction, immutable cast requests, and a checkpointed generic skill cast lifecycle. Use [NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md](NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md) together with `ROADMAP.md` to determine the current implementation cursor.

Dark Magic already has substantial engine foundations: a typed game-data catalog, a fixed-step authoritative session, replay/checkpoint support, authoritative item containers, deterministic loot generation, renderer-neutral world facts, Lua capability boundaries, provenance-preserving stat sources, and the beginning of a typed combat/monster/skill simulation. The purpose of this research program is therefore **not** to invent parallel systems. It identifies the Diablo II fidelity work that remains inside or beside those owners.

This program was distilled from the project research handoff supplied by the maintainer and reconciled repeatedly against the moving repository before new models were proposed.

## Research rules

For every system, answer four questions separately:

1. What behavior is observable in Diablo II?
2. Which declarative records and stable identifiers drive it?
3. Which Dark Magic subsystem should own the authoritative state and transition?
4. Which details are still uncertain enough to require an owned-game trace, save diff, packet trace, or binary/source corroboration?

A researched behavior is not automatically a compatibility contract. Use these labels:

- **verified**: direct owned-game observation or agreement among multiple strong sources.
- **high**: reconstructed original-runtime behavior with clear evidence, normally pinned D2MOO behavior.
- **medium**: one credible implementation, format document, or historical source.
- **uncertain**: inference, source disagreement, incomplete reversal, or patch-sensitive behavior.

Implementation may use clearly labeled Dark Magic scaffolding while an exact legacy rule remains unresolved, as M21.5 does for the first synthetic hit policy. Never silently promote scaffolding to a compatibility claim.

## Evidence order

Prefer:

1. lawful owned-game observations and repeatable traces;
2. ThePhrozenKeep/D2MOO pinned to the inspected 1.10f commit;
3. contemporary Phrozen Keep research and the bundled Data File Guide;
4. Blizzard manuals and patch notes;
5. independently implemented/recovered engines such as jaenster/libd2 and Riiablo;
6. lawful TXT/BIN/TBL/save inspection and independently documented formats;
7. player guides and wikis only as discovery aids.

GPL projects may be used as behavioral or service-responsibility corroboration, never as a source to mechanically translate implementation code.

## Compatibility dimensions

Every document should explicitly distinguish dimensions that matter:

- Classic versus Lord of Destruction;
- 1.10f, 1.13c/1.13d, 1.14d, and current Resurrected where evidence exists;
- Normal, Nightmare, and Hell;
- single player, TCP/IP/open multiplayer, legacy realm behavior, and modern online behavior;
- softcore/hardcore;
- Ladder/non-Ladder;
- client presentation, authoritative game-session behavior, realm persistence, and durable character state.

Do not silently combine these into one invented universal rule.

## Required sections for system research

A completed baseline should cover:

- player-visible behavior;
- relevant TXT/BIN/TBL/save/network records and cross-links;
- stable identifiers and runtime domain model;
- state machine/lifecycle;
- authoritative inputs and consequences;
- persistence representation;
- join/leave/disconnect/concurrency concerns;
- client/server/realm ownership;
- failure, rollback, idempotency, duplication, and recovery;
- patch/difficulty/mode variations;
- dependencies on other systems;
- concrete invariants and acceptance tests;
- unresolved questions and a verification backlog.

Formula work must state widths, fixed-point scale, signedness, rounding order, caps, and overflow assumptions. Procedural work must state RNG owner, seed derivation, stream state, and call-order sensitivity.

## Relationship to current Dark Magic architecture

The following are **existing owners to extend**, not proposals to replace:

| Concern | Current owner | Research consequence |
| --- | --- | --- |
| layered content bytes / generic TSV | `internal/game/data/store` + VFS | research exact legacy parsing/linkage/precedence while retaining raw provenance |
| typed immutable game-data generation | `internal/game/data/catalog` | add tables/index semantics only after evidence; do not parse tables independently in gameplay systems |
| fixed simulation clock, commands, replay, checkpoints | `internal/game/session`, `internal/game/simulation`, `internal/game/ecs` | add Diablo timers/RNG as deterministic state participants or ECS state |
| parameterized stat/effect provenance | `internal/game/stats` | extend M21.1 source identity/aggregation; do not flatten removable equipment/state/aura/charm contributions into permanent totals |
| typed fixed-point combat vocabulary | `internal/game/combat` | extend M21.2/M21.5 transactions and events while keeping exact legacy formulas evidence-gated |
| monster definition/materialization and basic AI | current monster/ECS simulation owners | extend M21.3/M21.4 with population, quality, lifecycle and behavior families; do not move AI into presentation |
| skill admission, immutable cast requests, and cast lifecycle | `internal/game/session` plus current skill simulation owner | extend M21.6/M21.7 with timed states and missiles; raw Skills formulas/server-function IDs remain explicit promotion/verification boundaries |
| deterministic loot generation | `internal/game/loot` | research remaining quality/stat/lifecycle fidelity instead of writing a second generator |
| authoritative item placement and transactions | `internal/game/item` | extend item instance/location semantics and runtime effects |
| selected-character shell state | `internal/persistence` | keep separate from legacy `.d2s` codec and from live ECS authority |
| live player state | `internal/game/player` + ECS | progression research expands admitted/live components and commands |
| map/world facts and DRLG | `internal/game/world` and current map generation work | reuse the existing map research; add gameplay systems that consume semantic world facts |
| cross-level relocation | `internal/game/transition` | generalize the trusted seam/endpoint authority for stairs, waypoints and portals instead of creating parallel travel systems |
| audio resources/backend | `internal/audio` | add semantic gameplay/world cue and soundscape policy above the existing owner-thread mixer; audio is never gameplay authority |
| Lua | versioned capabilities under `internal/runtime/lua` | presentation reads copied/revisioned snapshots and submits intent; it is not gameplay authority |
| standalone game worker | `cmd/darkmagic-server` + `internal/game/session` | networking must transport the same semantic commands/snapshots rather than fork simulation logic |
| realm control-plane shell | `cmd/darkmagic-realm` | add allocation, leases, durable-character and directory services without moving gameplay simulation into the realm |

The current authoritative session runs a bounded fixed clock and records canonical commands/checkpoints. Research or implementation that introduces hidden clocks, global gameplay RNG, direct Lua mutation, mutable data-catalog lookups during a live session, or transport-specific game logic conflicts with the current architecture and requires a very strong reason.

## Completed research order

The baseline work was assembled in this dependency order:

1. game data loading/linkage/modability;
2. time/RNG/determinism;
3. save format and persistent character state;
4. item stats/property evaluation;
5. complete runtime item model;
6. character progression;
7. combat/damage;
8. skills/missiles/states;
9. monsters/AI;
10. owned units/hirelings;
11. movement/collision/pathfinding;
12. item qualities/sockets/charms/cube/economy/quest items;
13. world objects, temporary effects, transitions, NPC services, difficulty/endgame;
14. multiplayer/realm responsibilities;
15. gameplay UI and audio contracts.

Every indexed workstream now meets the baseline definition below. Further research should be driven by a concrete implementation blocker, compatibility claim, or executable verification queue item rather than broad unscheduled surveying.

## Definition of “research complete”

A workstream is not complete because its TXT columns were enumerated. It is complete enough for implementation when:

- ownership is unambiguous;
- identifiers survive save/network/mod boundaries;
- state transitions and failure paths are described;
- formulas and RNG are either specified with test vectors or explicitly blocked on a probe;
- persistence and multiplayer consequences are covered;
- conflicts are visible;
- remaining unknowns are executable probes rather than vague TODOs.

All 28 indexed workstreams have reached that **baseline** threshold. The next maturity step is **validated**, earned only when important unresolved behavior is backed by owned-game traces, byte-exact fixtures, or multiple strong independent sources.

## How research proceeds from here

Do not reopen broad workstreams by default. Instead:

1. take the next implementation checkpoint from `ROADMAP.md` and [NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md](NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md);
2. read the relevant baseline plus its verification queue;
3. identify only the exact unresolved behavior needed by that checkpoint;
4. run the smallest lawful probe or source comparison that can answer it;
5. record version/confidence and synthetic or owned-game vectors;
6. implement the bounded checkpoint in the existing owner;
7. update roadmap evidence and, when warranted, promote the specific research claim from baseline to validated.

See [GAME_SYSTEMS_INDEX.md](GAME_SYSTEMS_INDEX.md) for the complete system map and [SYSTEMS_SOURCE_MATRIX.md](SYSTEMS_SOURCE_MATRIX.md) for the current evidence inventory.