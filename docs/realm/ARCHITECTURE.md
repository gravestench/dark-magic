# Realm architecture

## Product boundary

The Realm owns durable identity and coordinates authoritative game sessions.
The game worker owns the live simulation. The client owns presentation and
input but cannot assert trusted durable state.

```text
Client
  | HTTPS: account, characters, channel, game directory
  v
Realm control plane --------------------> PostgreSQL
  | private authenticated worker protocol       |
  v                                             | durable revisions
Game worker <-----------------------------------+
  |
  | QUIC: admission, commands, projections, reconnect
  v
Client
```

The control plane may use multiple internal processes, but those processes are
one product boundary. Splitting them is an operational choice, not permission
to duplicate account, character, or allocation rules.

## Current implementation

The repository has a `cmd/realm` composition root, bounded TLS HTTP API,
PostgreSQL account/session, character/lease, named-game, allocation, mail, and
audit repositories, public
channel and named-game primitives, structured audit events, character leases,
and Realm-to-session admission machinery.

Realm admission sees only `WorkerRegistry`, `WorkerClient`, and `TicketIssuer`.
Runtime description, durable admission, canonical projection, drain, ticket
issue, and revocation cross `RealmWorkerControl/v1` over pinned TLS and a
separate per-worker bearer secret. The in-process manager implements the same
semantics for focused tests. `ProcessAllocator` now supervises ordinary
`cmd/server --realm-worker` children, uses owner-only per-worker credentials and
readiness state, shares ticket revocation with the worker's QUIC authority, and
performs bounded shutdown and cleanup. Public named-game create/join now
reserves directory capacity, allocates the worker, leases and admits the
selected character, and returns the private assignment atomically with bounded
rollback. Lua sees only lobby state; the native controller consumes the ticket
and endpoint directly. Each real worker prepares the shared d2legacy entry-world
contract before publishing readiness, and its description—not Realm flags—owns
the admission spawn. Private worker status now drives lease renewal,
reconnect-expiry and unclaimed-admission commits, capacity observation, and
three-strike dead-worker reconciliation. Allocation intent and lifecycle are
durable. On startup, the local allocator authenticates to and fences the exact
surviving allocation generation before checkpoint restoration. Unprovable
ownership remains fail-closed without accepting client state. Cluster worker
fencing or reattachment remains M23 work. Canonical checkpoint capture and durable
latest-checkpoint persistence now cross the same worker boundary. Membership
admission and departure receipts are
durable; the canonical character update, lease consumption, and receipt write
share one PostgreSQL transaction. In-memory repositories remain
deterministic test fixtures and are never a Realm deployment mode.

## M23 target

M23 makes the Realm complete on one machine before adding cluster machinery.
It introduces:

- PostgreSQL repositories and a mutable pre-production baseline schema;
- explicit verified account creation and recovery;
- a small HTTPS verification and password-recovery portal plus explicit
  native-client credential login;
- a versioned private Realm-to-worker protocol;
- an in-process worker adapter for fast tests;
- an ordinary-process worker allocator for supported single-machine operation;
- durable game/allocation state, checkpoints, reconnect, draining, and
  reconciliation; and
- an end-to-end acceptance that crosses real Realm, worker, gameplay, and
  persistence boundaries.

The domain-facing abstractions are conceptually:

```go
type WorkerAllocator interface {
	Allocate(context.Context, GameSpec) (Allocation, error)
	Release(context.Context, AllocationID, ReleaseReason) error
	Status(context.Context, AllocationID) (WorkerStatus, error)
}

type WorkerClient interface {
	PrepareGame(context.Context, PrepareGameRequest) (PreparedGame, error)
	AdmitCharacter(context.Context, AdmitCharacterRequest) (Admission, error)
	Checkpoint(context.Context, CheckpointRequest) (CheckpointResult, error)
	Drain(context.Context, DrainRequest) error
}
```

These names illustrate ownership rather than freezing the final Go API. The
wire contract must be bounded, versioned, authenticated, idempotent where
retried, and independent of Kubernetes types.

## Supported topologies

| Topology | Allocation adapter | Storage | Status |
| --- | --- | --- | --- |
| Unit and focused acceptance | In-process adapter | In-memory repositories | Implemented |
| Single machine | Child `cmd/server` processes | PostgreSQL | Process boundary, live lifecycle reconciliation, create/join/leave wiring, durable allocation/membership/checkpoint state, retryable departures, and positively fenced checkpoint restoration implemented |
| Cluster | Agones-backed workers | Managed PostgreSQL | Deferred to M44 |

All topologies share session identity, admission, gameplay, persistence, and
protocol contracts. A clustered deployment may not introduce a second
game-creation or character-commit implementation.

## Service ownership

### Realm API and account portal

Owns authentication, account policy, Realm character selection, channel and
game-directory APIs, plus browser-based verification and recovery. Native login
always requires explicit credentials. It never mounts protected game
assets or executes gameplay Lua.

### Allocation and reconciliation

Transitions durable games through `requested`, `allocating`, `ready`,
`draining`, and terminal states. External allocation cannot be part of a
PostgreSQL transaction, so every transition is idempotent and a reconciler
repairs interrupted work.

### Game worker

Loads the exact pinned runtime, validates the configured asset set, runs the
authoritative fixed-step session, admits Realm-issued principals, projects
client state, checkpoints, and reports canonical durable results.

### Mail worker

Claims transactional outbox rows, submits mail through a provider-neutral
adapter, and records bounded retry outcomes. It cannot read game assets or
character payloads.

### PostgreSQL

Owns transactional Realm records. Workers do not receive general database
credentials; they use the private Realm boundary to renew membership and
commit durable results.

## Deferred cluster boundary

M44 may provide an `AgonesAllocator`, cloud secret integration, shared storage,
and edge ingress. It must implement the same allocator and worker contracts
proven by the single-machine topology. See [cloud deployment](CLOUD_DEPLOYMENT.md).
