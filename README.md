<h1 align="center">Dark Magic</h1>
<h3 align="center">An open-source Diablo II: Lord of Destruction engine rewrite</h3>

Dark Magic is a clean-room engine rewrite targeting **Diablo II: Lord of
Destruction 1.14d**. The project aims to reproduce the game's authored behavior
and presentation on a modern, deterministic, scriptable engine while making the
systems underneath it easier to inspect, test, extend, and mod.

Dark Magic is under active development and is **not yet a complete Diablo II
replacement**. It does not distribute Blizzard assets; running the real-asset
client and tools requires legally obtained Diablo II game data.

## How Dark Magic is developed

Most of Dark Magic has been implemented with repository-aware LLM coding agents
under active maintainer direction. Agents are used for codebase analysis,
research synthesis, implementation, tests, documentation, and repeated review
passes. The maintainer sets the target, curates evidence, decides architecture,
and remains accountable for every merged change.

This is not a bulk-generated “ship whatever the model wrote” codebase. Work is
scoped to explicit roadmap and acceptance boundaries, checked against existing
ownership rules, exercised by tests and real-asset tooling where appropriate,
and revised or discarded when it does not hold up. Review focuses on the
resulting evidence and maintainability, not whether every line was typed by
hand.

Performance is handled deliberately. Egregious problems—unbounded work,
repeated asset I/O, accidental quadratic paths, runaway allocation, or long
owner-thread stalls—are corrected when discovered. Broader profiling and
optimization passes normally follow coherent foundations, completed acceptance
work, or milestone closure, when the code is stable enough to measure and worth
tuning. Contributors are encouraged to use the same workflow described in
[CONTRIBUTING.md](CONTRIBUTING.md).

> **TODO: gameplay GIF** — short loop showing real-asset `game_world` movement,
> combat, drops, UI, and world presentation.
>
> **TODO: labs GIF** — short montage of Composite Lab, DT1/DS1 Lab, Monster Lab,
> Missile Lab, Mapgen Lab, Combat Lab, and Warp Lab.

## Project direction

Dark Magic has one authoritative, renderer-independent game simulation shared by
offline play, listen servers, dedicated servers, and Realm-managed games. The
current focus is completing Diablo II gameplay on that foundation rather than
building parallel implementations for each mode.

The project deliberately targets:

- **Expansion 1.14d only.** Classic mode and earlier patch behavior are not
  product targets.
- **Faithful, evidence-driven behavior.** Mechanics are reconstructed from game
  data, repeatable observation, reverse-engineering research, and independent
  implementations rather than guessed where evidence exists.
- **One authoritative simulation.** Networking, replay, checkpointing, offline
  play, and Realm play share the same gameplay boundaries.
- **Lua-owned Diablo policy, Go-owned engine mechanisms.** The first-party
  `d2legacy` package defines Diablo rules; Go owns reusable simulation,
  rendering, transport, storage, audio, and lifecycle infrastructure.
- **First-class modding.** Versioned capabilities, deterministic content
  layering, and extension packages are part of the architecture rather than an
  afterthought.

For the canonical implementation status and ordered gameplay work, see
[ROADMAP.md](ROADMAP.md). For package ownership and dependency rules, see
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Client, server, and Realm

Dark Magic has three cooperating process roles. They share versioned content and
session contracts, but each owns a different boundary.

| Process | Responsibility |
| --- | --- |
| [`client`](cmd/client) | Native player-facing application. It owns input, windowing, audio, UI, rendering, local profiles, and asset/mod mounting. Offline and listen-server play can compose the same authority locally; remote play consumes authenticated, player-scoped views from a server. |
| [`server`](cmd/server) | Headless authoritative game-session process. It runs one deterministic simulation, validates player commands, owns canonical world state, and produces checkpoints and bounded per-player projections. It can run standalone for self-hosted play or as a Realm-supervised worker. |
| [`realm`](cmd/realm) | Account and session control plane. It owns verified accounts, trusted Realm characters, channels and named games, worker allocation, admission, leases, checkpoints, durable commits, recovery, and audit. It allocates and supervises server workers but never executes gameplay ticks. |

Realm is deliberately topology-neutral. Its first deployment contract uses
PostgreSQL and ordinary `server` worker processes on one machine. The deferred
cloud target adapts that same contract to a Cloudflare/HTTPS edge, Kubernetes
ingress, managed PostgreSQL, and an Agones warm fleet of one-game server
workers. Protected game assets remain outside container images and public
storage and are mounted read-only only by game workers; native gameplay
continues over QUIC/UDP. Cloudflare, Kubernetes, Agones, and the chosen cloud
provider are deployment adapters rather than dependencies of Realm's domain
logic.

Cloud deployment is a target, not a claim that production manifests or hosted
infrastructure already exist. See the [Realm overview](docs/realm/README.md),
[cloud deployment contract](docs/realm/CLOUD_DEPLOYMENT.md), and
[networking architecture](docs/NETWORKING.md).

## Run the game and labs

Set `MPQ_DIRECTORY` to a legally obtained Diablo II installation or MPQ
directory. The repository currently uses Go 1.25.

```sh
export MPQ_DIRECTORY=/path/to/diablo-ii
make play-game-world
```

Raylib is the default native client backend. An experimental Ebitengine build
uses the same gameplay, Lua scenes, retained composer, logical input, and
capture path:

```sh
go run -tags ebitengine ./cmd/client --start-scene ui_lab
```

Use `make build-client-backends` to compile both choices and
`MPQ_DIRECTORY=/path/to/diablo-ii make profile-render-backends` for a matched,
audio-muted A/B capture. See [docs/PERFORMANCE.md](docs/PERFORMANCE.md) for the
current compile/runtime measurements, known Ebitengine parity gaps, and artifact
layout.

The development labs are intentionally small entry points into production
systems. They are useful both for contributors and for exploring what the engine
can already do.

| Entry point | What it exercises | Run it |
| --- | --- | --- |
| Game World | Authored world, player movement, camera, collision, UI | `make play-game-world` |
| Combat Lab | Production world + authoritative combat diagnostics | `make play-combat-lab` |
| Monster Lab | Typed monster records and composites | `make play-monster-lab` |
| Missile Lab | Typed missile records and presentation | `make play-missile-lab` |
| Warp Lab | Production movement, interaction, and level transitions | `make play-warp-lab` |
| Composite Lab | Player composite animation assembly | `go run ./cmd/client --start-scene=composite_lab` |
| DT1 Lab | Tileset decoding and inspection | `go run ./cmd/client --start-scene=dt1_lab` |
| DS1 Lab | Map stamps, authored tiles, objects, and collision | `go run ./cmd/client --start-scene=ds1_lab` |
| Mapgen Lab | Deterministic level-generation topology and decisions | `go run ./cmd/client --start-scene=mapgen_lab` |

Most visual scenes can also be captured locally through the Makefile capture
targets. Captures remain outside the repository so original game imagery is not
committed.

## Where to look

The README is intentionally only an orientation guide. Detailed design,
implementation notes, evidence, acceptance criteria, configuration, and tooling
live with the documents and packages that own them.

| If you are interested in... | Start here |
| --- | --- |
| Current progress and next work | [ROADMAP.md](ROADMAP.md) |
| Engine boundaries and package ownership | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) |
| Diablo gameplay rules | [`internal/content/d2legacy`](internal/content/d2legacy) and its Lua modules |
| Deterministic game/session mechanisms | [`internal/game`](internal/game) |
| Client presentation and native rendering | [`internal/presentation`](internal/presentation) and [`internal/platform/desktop`](internal/platform/desktop) |
| Realm, accounts, sessions, persistence, and deployment | [docs/realm/README.md](docs/realm/README.md) |
| Mod architecture and authoring | [docs/MODS.md](docs/MODS.md), [docs/MODDING_API.md](docs/MODDING_API.md), and the [`mod_template`](internal/content/mod_template/README.md) |
| Game-system research and fidelity evidence | [docs/research/GAME_SYSTEMS_INDEX.md](docs/research/GAME_SYSTEMS_INDEX.md) |
| File-format provenance and codec policy | [docs/formats/SOURCE_MATRIX.md](docs/formats/SOURCE_MATRIX.md) and [CODECS.md](CODECS.md) |
| Developer inspection/capture tools | [`internal/dev/tools`](internal/dev/tools) and the Makefile |
| Process configuration | [docs/CONFIGURATION.md](docs/CONFIGURATION.md) |

At a high level, `cmd` contains composition roots, `internal/game` contains
generic deterministic game mechanisms, and `internal/content/d2legacy` contains
the first-party Diablo II rules/content package. There is intentionally no
public Go API under `pkg`; independently versioned Diablo file codecs live in
separate repositories.

## Development

Run the main test suites with:

```sh
make test
make test-race
```

Before adding a long-lived engine component, read the
[architecture guide](docs/ARCHITECTURE.md). Contribution expectations are kept
in [CONTRIBUTING.md](CONTRIBUTING.md).

## Research, lineage, and community

Dark Magic exists because decades of Diablo II modders, preservationists,
reverse engineers, tool authors, and independent engine developers documented
behavior that the original executable kept implicit. The project descends from
[OpenDiablo2](https://github.com/OpenDiablo2/OpenDiablo2), but its current
architecture and gameplay implementation are being rebuilt around Dark Magic's
own boundaries and evidence requirements.

Principal research and corroboration sources include:

- [Paul Siramy's research and tools](http://paul.siramy.free.fr/) and the wider
  [Phrozen Keep](https://d2mods.info/) community, especially for maps, formats,
  editing/generation, animations, and engine behavior.
- [ThePhrozenKeep/D2MOO](https://github.com/ThePhrozenKeep/D2MOO), a major
  reverse-engineered Diablo II 1.10f runtime reference.
- [jaenster/libd2](https://github.com/jaenster/libd2), including independent
  1.14d implementation evidence and strong retail-capture/holdout verification.
- [Riiablo](https://github.com/collinsmith/riiablo), including format,
  composite, rendering, UI, quest, dialogue, and object research.
- [OpenD2](https://github.com/eezstreet/OpenD2) and
  [AbyssEngine](https://github.com/AbyssEngine/AbyssEngine) as independent
  behavioral and presentation references.
- [nokka/d2s](https://github.com/nokka/d2s), community data-format guides,
  historical patch notes, forum research, and the many tools and references
  cited in the repository's research ledgers.

These sources are **evidence, not code-generation templates**. Dark Magic
restates behavior independently, respects source licenses, records version
conflicts and uncertainty, and prefers repeatable observations of lawfully owned
game data. Detailed provenance and confidence tracking live in the
[format source matrix](docs/formats/SOURCE_MATRIX.md),
[gameplay-systems source matrix](docs/research/SYSTEMS_SOURCE_MATRIX.md), and
[research index](docs/research/GAME_SYSTEMS_INDEX.md).

If you have contributed research, tools, code, testing, documentation, or
knowledge to the Diablo II community: thank you. Dark Magic is built on that
shared body of work.

Join development and discussion on the
[Dark Magic Discord](https://discord.gg/eGfN7Faxur).
