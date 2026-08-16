# Realm operations

## Implemented local lifecycle

The repository provides `scripts/realm/up.sh`, `down.sh`, `fresh-install.sh`,
and `drain-game.sh`, with matching `make realm-*` targets. Bring-up uses a
bounded readiness probe and an owner-only PID/log/runtime layout. Teardown sends
SIGTERM to the verified Realm process so its allocator can drain supervised
workers, then applies a bounded forced-exit fallback. A stale or foreign PID is
never signalled.

Fresh install archives operational Realm state instead of deleting it. Its
PostgreSQL reset is separately authorized and local-only. These are developer
and single-host lifecycle tools; they do not replace database backup, remote
administration, service-manager, or cluster rollout procedures.

This is the index for operational contracts. Concrete commands become normative
only when their corresponding implementation is marked complete.

## M23 single-machine operations

M23 must provide and test runbooks for:

- initial PostgreSQL schema installation and version inspection;
- creating the first administrator without logging credentials;
- configuring public TLS and SMTP;
- validating and selecting an asset set;
- starting, observing, draining, and terminating worker processes;
- inspecting games, allocations, character leases, and mail-outbox backlog;
- revoking account sessions and recovering accounts;
- rotating Realm/worker workload credentials;
- retrying or quarantining failed mail jobs;
- reconciling interrupted allocations and memberships;
- exporting bounded diagnostics; and
- clean shutdown and restart.

Every mutating administration operation is authenticated, authorized, bounded,
idempotent where retries are expected, and audited. The interactive local shell
is not the eventual remote administration security model.

The implemented game-drain operation uses a separate TLS operator listener,
`127.0.0.1:6113` in the local profile. It is not routed through the player or
browser handler. Its bearer credential is generated once in an owner-only
runtime file and never stored in PostgreSQL or passed on the helper's command
line. Drain first durably changes the game from `active` to `draining`, which
atomically removes it from discovery and rejects new admission. It then
projects and transactionally commits each canonical character before removing
the worker and checkpoint. A partial failure remains `draining` and is safe to
retry:

```sh
make realm-drain-game GAME_ID=00000000-0000-0000-0000-000000000000
```

## Implemented PostgreSQL startup behavior

`cmd/realm --postgres-url` connects with a five-second establishment timeout,
pings the database, and initializes the embedded baseline schema before serving
traffic. Initialization is transactional and serialized with a PostgreSQL
advisory transaction lock. Dark Magic is pre-production, so this baseline is
mutable and no migration ledger is maintained yet. An incompatible development
schema change must fail startup and be applied with the guarded local
`make realm-fresh-install` workflow. Immutable forward migrations begin only
after the complete Realm storage contract is declared stable. There is no
file-backed Realm deployment mode; in-memory repositories exist only as
deterministic test fixtures.

Healthy game workers publish a bounded canonical checkpoint through the private
worker protocol. Realm stores only the latest validated checkpoint per game in
PostgreSQL, pinned to the allocation generation and runtime identity. The
capture cadence defaults to `15s` and can be changed with
`--checkpoint-interval` or `DARK_MAGIC_REALM_CHECKPOINT_INTERVAL`. Clean game
completion removes the checkpoint. A single-machine Realm restart uses it only
after positively fencing the exact surviving allocation generation; failed or
unprovable fencing retains the checkpoint for operator diagnosis.

## Backups and recovery

PostgreSQL contains valuable account and character data. The storage schema and
write paths therefore support:

- consistent database backups;
- point-in-time recovery in production;
- restoration into an isolated validation environment;
- documented recovery-point and recovery-time objectives;
- append-only or externally exported audit retention;
- explicit treatment of in-flight leases and allocations after restore; and
- encryption and access separation for backup material.

Protected game assets are operator-supplied installation data and are not copied
into Realm database backups. The operator preserves them separately and Realm
validates their `AssetSetID` after recovery.

M23 establishes durable and idempotent semantics. Automated cloud backup,
failover, and disaster-recovery exercises are M44.

Fail-closed allocation recovery marks active memberships abandoned after their
leases are released. A departure that had already committed retains its receipt
and is sealed as worker-removed once the former worker is proven stopped, so a
later client retry never targets an orphaned process or advances the character
again.

## Required observability

Structured logs, metrics, traces, and audit events use correlation identifiers
without exposing credentials or save payloads. Required signals include:

- authentication and recovery outcomes;
- active sessions and revocations;
- channel and named-game counts;
- allocation latency and state age;
- worker readiness, heartbeat, capacity, and exit classification;
- lease age, renewal, conflict, and expiry;
- checkpoint and durable commit latency/failure;
- database pool and transaction health;
- mail queue age, retries, and terminal failures; and
- asset validation identity and duration.

Alerts must describe an operator action. High-cardinality player or game labels
are bounded and must not turn metrics into a second audit database.

## M44 operations

The cloud milestone adds cluster rollout/rollback, autoscaling, node drain,
managed database failover, secret-manager integration, certificate automation,
network-policy verification, regional capacity, and disaster-recovery drills.
See [cloud deployment](CLOUD_DEPLOYMENT.md).
