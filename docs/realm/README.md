# Realm documentation

The Dark Magic Realm is the account and session control plane. It owns trusted
accounts, Realm characters, the public channel and named-game directory,
worker allocation, game admission, durable commits, and operational audit. It
does not execute gameplay ticks and it never distributes user-supplied game
assets.

Realm behavior is independent of deployment topology. The same product must
run on one machine with ordinary game-worker processes or in a cluster with an
external worker allocator. Kubernetes, Agones, Cloudflare, and any particular
cloud provider are deployment adapters, not Realm domain dependencies.

## Status language

These documents distinguish three states:

- **Implemented** describes behavior with repository evidence.
- **M23 target** describes the topology-neutral Realm work that is required
  before returning to the remaining game-system milestones.
- **Deferred to M44** describes cloud deployment and operations work that must
  not be pulled into the M23 implementation.

Planned commands, manifests, and services are not operator instructions until
their sections are marked implemented.

## Core invariants

- Offline, listen-server, and self-hosted dedicated play use player-controlled
  local profiles. Realm play uses account-owned Realm characters.
- A connected client never writes a trusted Realm character.
- A game worker validates gameplay commands and produces canonical state; the
  Realm coordinates revisioned leases and durable commits.
- A worker is addressed through a versioned private protocol. Realm code does
  not depend on a local Go pointer to the worker implementation.
- The embedded `d2legacy` package, ordered extension lock, authoritative
  configuration, engine API versions, and validated asset-set identity form one
  complete session identity.
- User-supplied game assets may be mounted by validated game workers but are
  never copied into images, caches, mod distribution, backups, or public object
  storage.
- Single-machine operation is a supported Realm topology, not a disposable
  development approximation.

## Documents

- [Architecture](ARCHITECTURE.md) defines service ownership and the topology
  boundary.
- [Accounts and security](ACCOUNTS_AND_SECURITY.md) defines identity, email,
  credentials, sessions, recovery, and audit requirements.
- [Characters and persistence](CHARACTERS_AND_PERSISTENCE.md) defines trusted
  ownership, revisions, leases, checkpoints, and PostgreSQL responsibilities.
- [Workers and game lifecycle](WORKERS_AND_GAME_LIFECYCLE.md) defines the
  process-independent worker contract and allocation lifecycle.
- [Local deployment](LOCAL_DEPLOYMENT.md) defines the required single-machine
  topology and its acceptance gate.
- [Cloud deployment](CLOUD_DEPLOYMENT.md) records the deferred M44 Kubernetes
  and edge plan.
- [Operations](OPERATIONS.md) defines schema evolution, backup, recovery, key rotation,
  observation, and incident runbook requirements.

The transport-level contracts remain in
[the networking guide](../NETWORKING.md). The historical evidence and source
assessment remain under [`docs/research`](../research/); research notes do not
override this product contract or claim implementation.
