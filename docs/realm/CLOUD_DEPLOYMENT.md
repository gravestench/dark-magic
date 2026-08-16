# Realm cloud deployment

Status: **deferred to M44**. This document records the deployment contract and
sequencing gate. It is not a claim that manifests or cloud operation exist.

## Gate

Cloud work begins only after:

- the M23 single-machine acceptance passes through a real worker process and
  PostgreSQL;
- Realm admissions no longer depend on `gameserver.Host` pointers;
- account verification, recovery, character leases, durable commits, and game
  reconciliation are complete; and
- the current gameplay milestones and gameplay acceptance have returned to the
  agreed completion point.

Container and protocol smoke tests may validate seams earlier. They may not
pull cluster orchestration into Realm domain packages.

## Target topology

```text
Cloudflare DNS / HTTPS proxy / WAF
                  |
           Kubernetes ingress
                  |
     Realm API / portal / allocator / mailer
          |                   |
  managed PostgreSQL      Agones Fleet
                              |
                      allocated workers
                              |
                 read-only game-assets volume

Native clients ---------------- QUIC/UDP ----------------> workers or gateway
```

Cloudflare is an edge provider, not the Kubernetes host and not a Realm domain
dependency. Native gameplay QUIC initially uses a cloud UDP load balancer or
worker endpoints. An optional stable QUIC gateway or UDP edge product is added
only after direct operation is measured.

## Kubernetes and Agones

M44 provides an `AgonesAllocator` implementing the worker contract already
proven locally. A Fleet maintains workers that have loaded immutable runtime
content and reported Ready. Game creation allocates from that warm pool; cold
pod or node startup replenishes capacity asynchronously rather than blocking
ordinary users.

One pod hosts one game initially. Multi-session packing requires measurement
showing that process or pod overhead is material and must not alter session
isolation.

M44 benchmarks warm allocation, cold readiness, asset validation, image pull,
node scale-up, draining, and crash recovery before setting service objectives.

## PostgreSQL

Production prefers managed PostgreSQL in the same region and private network as
the Realm. It requires high availability, encrypted connections, point-in-time
recovery, migration discipline, tested restore, narrowly scoped roles, and
capacity observation. Running PostgreSQL inside Kubernetes is acceptable for
local integration and deliberately self-managed installations, not the default
cloud recommendation.

## Protected game assets

The chart accepts an existing volume claim. The backing storage must support
read-only mounts on every eligible worker node. An asset-validation job may
mount controlled write/import access; ordinary workers mount `/game-assets`
read-only. Realm web, mail, and public API pods do not mount it.

Assets are excluded from images, registries, CI artifacts, mod distribution,
public object storage, platform backups, and logs. Only the manifest and
`AssetSetID` participate in session identity. If shared filesystem reads become
a measured bottleneck, M44 may add a restricted node-local cache without
changing the worker-visible path or redistribution policy.

## Local Kubernetes integration

M44 includes a separate `kind-dark-magic` cluster rather than modifying an
unrelated current context. The integration environment provides:

- ingress on local HTTP/HTTPS ports;
- a trusted local certificate issuer;
- CloudNativePG or equivalent disposable PostgreSQL;
- Mailpit;
- Agones with a small warm Fleet;
- a fixed mapped UDP range; and
- an operator-configured host asset mount or local shared-volume fixture.

The local Kubernetes environment tests deployment behavior. It does not replace
the faster native M23 development loop.

## Edge, mail, and redistributable objects

Cloudflare may provide DNS, proxied HTTPS, WAF, rate limiting, certificate DNS
challenges, and delivery of redistributable extension objects. Protected game
assets are never placed there. Production email remains behind the same mailer
interface used by Mailpit; provider selection does not alter account semantics.

## M44 acceptance

- A clean cluster deploys from pinned, asset-free images and migrations.
- Warm allocation and replacement stay within measured budgets.
- Worker/node termination drains or recovers without character corruption.
- Database failover and restored backups preserve documented invariants.
- Asset mounts are unavailable to unrelated workloads and never writable by
  ordinary workers.
- Network policies and service accounts enforce least privilege.
- Secret and certificate rotation do not require character migration.
- HTTPS account, Realm directory, worker admission, QUIC gameplay, checkpoint,
  reconnect, and commit pass end to end.
- The native single-machine topology continues to pass the same semantic
  acceptance.
