# Realm workers and game lifecycle

## Principle

A Realm game worker is the same authoritative session host used by offline,
listen-server, and standalone-server modes, placed behind a process-independent
control boundary. Deployment topology must not fork gameplay behavior.

## Current implementation

`internal/app/realm.Manager` creates local `gameserver.Host` instances behind an
in-process `WorkerClient` for focused tests. Admissions depends on a
`WorkerRegistry`, semantic worker requests, and a context-aware `TicketIssuer`;
it never inspects a host allocation/session or accepts a concrete
ticket-authority pointer. Worker descriptions are defensively copied and their
runtime digest is revalidated before a character lease proceeds. Canceled joins
use an independent bounded cleanup context so ticket revocation and lease
release still run.

`RealmWorkerControl/v1` carries those operations over bounded strict JSON on
pinned TLS with a separate high-entropy per-worker bearer secret. A supervised
`ProcessAllocator` launches `cmd/server --realm-worker` without a shell, waits
for an owner-only atomic readiness rendezvous, verifies the child's certificate
and runtime identity, captures bounded logs, detects early exit, rejects
duplicate allocation, drains with a deadline, force-terminates when necessary,
and removes per-worker credentials. The worker control service and QUIC server
share one ticket authority, so remote revocation is effective at gameplay
admission.

Public create/join orchestration now crosses this boundary: game creation
allocates a worker, registers admission, atomically reserves roster capacity,
leases and admits the selected character, and returns a native-only assignment.
Joining an existing named game follows the same reservation/admission/commit
sequence. Every partial failure revokes tickets, removes a queued player,
releases its character lease and worker, and removes uncommitted directory
state as applicable. The client verifies the worker's pinned TLS identity,
acquires the exact extension recipe, and consumes the assignment without
placing its endpoint or ticket in Lua-visible state.

Workers use the same shared d2legacy entry-world adapter as the interactive
client. Before readiness, `cmd/server` generates and materializes the entry
zones, incorporates transition/interaction state in the runtime identity,
installs every collision map, queues population, and publishes its actual town
spawn in `WorkerDescription`. Realm does not invent worker geometry.

`WorkerStatus` now reports readiness, canonical tick, active simulated players,
and player IDs whose QUIC reconnect grace expired. The public transport and
private worker-control adapters share a process-local membership tracker. A
disconnect does not delete a Realm entity: after grace expires the tracker
reports it, Realm projects and commits canonical character state, consumes the
lease, asks the worker to remove the entity, updates the public roster, and
releases an empty allocation. Explicit native leave follows the same path and
keeps a durable receipt so a lost HTTP response or Realm restart can be retried
without incrementing the character revision twice.

A worker admission also carries the Realm-owned one-use ticket deadline. A
successful initial join/reconnect is observed inside the authenticated endpoint;
an admission that is never claimed is reported after its ticket can no longer
be used and follows the same canonical departure path. This covers a client
process dying between HTTP handoff and QUIC join.

The Realm maintenance loop probes every two seconds and requires three
consecutive unavailable/not-ready results before abandoning an allocation. A
healthy probe clears the failure count. When authority has already been lost,
Realm never invents a checkpoint: it retains the last successfully validated
complete recovery checkpoint and the last committed character
revision, releases the leases, removes the directory entry, and audits the
reconciliation. Healthy probes also renew active character leases once half of
their configured lifetime has elapsed. The first healthy probe captures a
complete recovery checkpoint; later successful captures are cadence-limited (15 seconds
by default). Failed captures retry on the next maintenance pass without
classifying an otherwise healthy worker as dead.

Allocation requests are durably recorded before external worker creation, with
stable allocation and idempotency identities. Readiness persists the public
endpoint and exact runtime identity; health, completion, and bounded failure
reasons are durable. The process allocator can start a replacement worker from
a validated recovery checkpoint. It transfers that checkpoint through an
owner-only bounded file; the worker composes the exact runtime and world first,
verifies the pinned runtime identity, restores simulation plus input-admission
state transactionally, and skips fresh-world population. After the configured
health threshold, the owning Realm process now terminates a failed worker,
restores its last checkpoint under the same durable allocation generation,
seeds its active player IDs, updates the private endpoint, and retains the game
and leases. Authenticated clients retry Realm for a fresh endpoint and one-use
ticket, then atomically replace their transport while resubmitting exact pending
inputs. Single-machine startup recovery now uses an owner-only v2 readiness
record containing the game, allocation generation, and worker PID. The new
Realm authenticates to the exact surviving worker with its persisted bearer
token and pinned TLS identity, verifies the worker-reported game identity, and
requests a drain that closes public QUIC before acknowledging. Only after the
old control endpoint is unreachable does it restore the checkpoint, rotate
membership leases, and reopen the game. Missing, malformed, mismatched, or
unreachable fencing evidence remains deliberately fail-closed: Realm releases
only the game leases, hides the game, retains its latest checkpoint, and keeps
an auditable failed allocation record. Active memberships and
departure receipts are durable. Departure atomically
commits the canonical character, consumes its lease, and records the retry
result before worker/roster cleanup. Periodic game checkpoints, live automatic
replacement, recovered membership seeding, and authenticated reconnect handoff
and safely fenced single-machine restart recovery are implemented. The
loopback-only, independently authenticated operator drain durably closes
admission and commits every canonical character before worker retirement.
Cluster fencing/reattachment remains M23 work.

The durable membership repository also has an authority-resume operation. For
PostgreSQL it atomically locks all active game memberships, replaces their lost
process-private lease tokens with fresh digest-only leases, and returns the raw
tokens only to Admissions. Admissions can rehydrate its local index and the
replacement endpoint before traffic resumes. Startup invokes it only after the
exact surviving local authority has been positively fenced.

## M23 worker boundary

M23 introduces a bounded, authenticated, versioned private worker protocol. Its
semantic operations include:

- prepare an allocated game from an immutable `GameSpec`;
- return endpoint, certificate identity, runtime identity, and readiness;
- admit a leased canonical character;
- report and renew active memberships;
- produce reviewed durable character results and game checkpoints;
- expose bounded capacity and health;
- begin graceful drain; and
- terminate an allocation idempotently.

The protocol carries semantic values, never Go pointers, Kubernetes resources,
raw ECS stores, credentials belonging to another service, or user-supplied
asset bytes.

The Realm issues or authorizes client admission. The worker validates the
one-use ticket and binds it to the exact game, player, character revision, and
runtime identity. A worker cannot create an account or independently replace a
Realm character.

## Allocation state machine

```text
requested -> ready -> completed
     \          \-> failed
      \-----------> failed
```

External allocation and PostgreSQL cannot share one transaction. Every request
therefore has an idempotency key, every external allocation has a durable
identity, and a reconciler repairs interrupted transitions or releases orphaned
workers.

## Single-machine allocator

The single-machine adapter starts ordinary `cmd/server` processes with:

- an assigned private control endpoint;
- a bounded public QUIC endpoint/port;
- explicit TLS and workload credentials;
- the validated game-asset directory;
- the exact embedded-base and extension recipe; and
- cancellation and graceful-shutdown ownership.

It captures bounded worker logs, detects unexpected exit, applies bounded drain
and forced termination, and cannot execute through a shell-constructed command
string. Live-process health, durable allocation records, periodic durable
checkpoints, and positively fenced single-machine restart restoration are
implemented. Cluster fencing/reattachment remains. Worker processes remain
independently debuggable.

An in-process adapter preserves fast unit and acceptance tests, but it must call
the same semantic worker boundary instead of exposing `gameserver.Host` to Realm
business logic.

## Session identity and assets

A prepared worker validates one complete identity:

```text
embedded d2legacy identity
ordered extension lock
engine and capability versions
authoritative configuration
asset-set identity
```

The generated asset-set manifest contains configured-root order, normalized
relative file names, sizes, and SHA-256 hashes. Its public identity is only the
path-independent manifest digest; absolute paths remain in an owner-local
metadata cache used to avoid rehashing unchanged multi-gigabyte archives.
`DARK_MAGIC_ASSET_SET_REHASH=1` forces byte revalidation. Realm computes the
expected digest independently and rejects a child whose reported runtime names
a different set before that child becomes allocatable.
Single-machine workers receive a read-only directory. Cluster workers later
receive the same logical directory through an operator-provided volume. Assets
are never transferred by the worker protocol.

## Deferred to M44

Agones implements the allocator contract with Ready and Allocated workers,
Fleet warm capacity, autoscaling, pod draining, and node-failure recovery. It
does not change `GameSpec`, admission, persistence, or worker-control semantics.
