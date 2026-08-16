# Realm characters and persistence

## Ownership classes

Dark Magic deliberately keeps two character trust domains:

| Character | Owner | Storage | Trusted by a Realm |
| --- | --- | --- | --- |
| Offline/listen/self-host profile | Player | Local `Profile/v1` | No |
| Realm character | Realm account | Realm repository | Yes |

Copying or offering a local profile character does not turn it into a Realm
character. Realm characters are created through Realm policy and are admitted
only from the Realm's current durable revision.

## Current implementation

Revisioned Realm character records, exclusive leases, canonical projection, and
PostgreSQL persistence exist. Membership admission keeps leases inside
the Realm boundary, and clients receive only a short-lived worker ticket and
public assignment facts. Graceful leave and reconnect-grace expiry both commit
only the worker's canonical projection, consume the lease, update the public
roster, and release an empty worker allocation. Retried graceful leave returns
the same durable receipt rather than committing a second revision.

The PostgreSQL character adapter stores defensive JSONB character projections,
runtime compatibility, revision compare-and-commit state, and digest-only lease
tokens. Row locks and a unique character lease enforce one winner under
concurrent acquisition. Account identities, digest-only sessions, characters,
and audit events survive a Realm pool/process restart in real PostgreSQL tests.

Named-game directory, allocation lifecycle, and active membership state are
durable. A canonical departure advances the character revision, consumes its
lease, and stores its retry receipt in one PostgreSQL transaction. Worker and
roster cleanup can then be retried—even after a Realm restart—without committing
the character twice. Realm startup positively fences a surviving
single-machine worker using its owner-only allocation-generation record,
bearer token, pinned TLS identity, and worker-reported game identity before
restoring it. If that proof cannot be completed, startup reconciles the
allocation fail-closed: it hides the game, releases only leases associated with
that game, and records a bounded failure without changing character revisions.
Realm also persists the
latest complete game-recovery checkpoint. Each record is pinned to the durable
allocation generation and complete runtime-identity digest, enforces monotonic
ticks, verifies the recovery checksum and stored-payload digest, and is removed
after clean game completion. In addition to canonical ECS and registered-runtime
state, it preserves accepted future inputs, recent exact applied inputs, and
per-player sequence acknowledgements so recovery can neither lose acknowledged
input nor double-apply an input whose acknowledgement was lost. A replacement
worker can consume this record through a bounded owner-only handoff after
recomposing and validating the exact runtime. While one Realm process still
owns the failed worker, it terminates that authority, starts a replacement with
the same durable allocation generation, seeds the recovered player IDs, and
keeps the game and character leases active. A client can then authenticate to
Realm for a fresh endpoint and one-use ticket; its native session retargets only
if game, runtime, player, and character identities are unchanged. The local
process allocator now performs that fencing and startup restoration; future
cluster allocators must provide an equivalent fencing proof or fail closed.

Membership persistence never stores raw character-lease bearer tokens. The
PostgreSQL recovery boundary now locks every active membership and lease for a
game, rotates all lease digests in one transaction, extends their deadlines,
and returns the new raw tokens only to the recovering Realm process. Admissions
can rebuild its process-local game and membership indexes from that result
before public traffic starts. Safely fenced startup recovery now uses this
path; ordinary live replacement retains the already valid raw
leases in the owning Realm process.

## M23 PostgreSQL responsibilities

PostgreSQL becomes the production and integration-test store for:

- accounts and verified email ownership;
- credential metadata and sessions;
- Realm characters and revision history policy;
- active character leases and renewal deadlines;
- named games, allocations, memberships, latest canonical checkpoints, and
  recovery state;
- transactional mail outbox rows;
- audit events; and
- schema identity and evolution policy.

Before production stability, the schema is one mutable baseline and local
incompatible changes use the guarded fresh-install workflow. Immutable forward
migrations begin only after the complete storage contract is declared stable.

Repository interfaces remain usable with small in-memory test implementations,
but the single-machine production acceptance uses PostgreSQL.

## Character lease

A character lease binds at least:

```text
AccountID
CharacterID
DurableRevision
GameID
WorkerID
LeaseTokenDigest
ExpiresAt
```

Acquisition verifies account ownership and atomically excludes another active
game. Renewal is performed through the trusted worker/Realm boundary. A stale,
foreign, expired, or replayed lease cannot commit.

## Join and commit

1. Realm authenticates the account and selected character.
2. Realm resolves a ready game and exact session identity.
3. Realm acquires the current character revision under a lease.
4. Realm asks the worker to admit the canonical character at a trusted
   destination.
5. Realm returns the bounded worker assignment and one-use ticket to the client.
6. The worker runs the authoritative session and renews membership.
7. Periodically, the worker returns its complete recovery checkpoint; on
   leave it separately projects the reviewed durable character subset.
8. Realm compares identity and expected revision, commits atomically, increments
   the revision, and releases or renews the lease as appropriate.

A connected client never supplies the replacement character record.

## Game checkpoints versus character commits

A game checkpoint preserves transient world, input-admission, and reconnect state. A character
commit updates long-lived account-owned character state. They have different
lifetimes and may not be conflated.

Worker loss may restore a compatible game checkpoint or terminate the game and
recover the last safe character revision according to explicit policy. It must
never guess a successful commit from missing acknowledgements; commit requests
are idempotent and identified durably.

## M23 acceptance

- Realm restart preserves accounts, characters, revisions, games requiring
  reconciliation, mail jobs, and audit records.
- Duplicate admission cannot lease the same character twice.
- Worker termination cannot accept a client-authored fallback save.
- Retried commit returns the original result or a classified conflict.
- Database failure leaves allocation and character state repairable.
- Offline profiles remain unchanged by Realm operations.

Backups, point-in-time recovery, managed PostgreSQL failover, and disaster
recovery exercises are completed under M44; the schemas and idempotency needed
for them are M23 work.
