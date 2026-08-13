# Networking architecture

Dark Magic separates semantic protocols from their transports. Gameplay code
depends on `internal/game/session`; authenticated adapters depend on
`internal/app/gameserver.Endpoint`. Neither package knows about sockets, HTTP,
protobuf, or legacy Diablo II packets.

## Transport choices

Use different transports for traffic with different requirements:

| Boundary | Preferred transport | Reason |
| --- | --- | --- |
| Client to live game session | QUIC with TLS 1.3 | One secure connection can carry reliable independent streams and latency-sensitive unreliable datagrams without TCP head-of-line blocking between flows. |
| Realm, account, directory, and administration | Connect/gRPC-compatible protobuf over HTTP/2 | These are typed request/response and low-rate streaming APIs where tooling, inspection, and interoperability matter more than datagrams. |
| Browser client, if added | WebTransport over HTTP/3 | It exposes reliable streams and unreliable datagrams with the browser security model, but remains a developing standard and is not the native-client baseline. |
| Original Diablo II client | Dedicated legacy adapters | Historical packet formats translate into the same semantic endpoint and never define core APIs. |

The first native game adapter should use `quic-go`. Join, reconnect, critical
commands, inventory changes, chat, and full/correction snapshots use reliable
streams. Replaceable movement inputs, acknowledgements, and high-frequency
state deltas may use QUIC datagrams only after measurement proves that loss is
preferable to late delivery. Correctness must never depend on a datagram.

This is a direction, not a claim that QUIC is universally fastest. Benchmarks
must cover representative tick rates, message sizes, loss, jitter, NATs, CPU,
and memory before tuning encodings or assigning a message class to datagrams.

Primary references:

- [QUIC transport, RFC 9000](https://www.rfc-editor.org/rfc/rfc9000)
- [QUIC DATAGRAM, RFC 9221](https://www.rfc-editor.org/rfc/rfc9221)
- [gRPC performance guidance](https://grpc.io/docs/guides/performance/)
- [quic-go streams](https://quic-go.net/docs/quic/streams/)
- [WebTransport](https://www.w3.org/TR/webtransport/)

## Session trust boundary

The endpoint authenticates a credential into a server-owned principal. It binds
that principal to the session player ID and character ID, checks the exact mod
runtime identity, and issues an unpredictable bearer credential held in the
server membership table. Client command messages contain only tick, sequence,
kind, and action payload. The endpoint supplies `Player` and player authority.

Join and reconnect return versioned per-player semantic projections plus a
canonical tick and checksum. They do not expose raw ECS snapshots or hidden
server facts. Reconnect rotates the bearer credential so a successfully used
old credential cannot be replayed.

## Join-time mod acquisition

Realm admission advertises a signed, versioned mod manifest before a client is
allowed into a game session. Missing redistributable packages can be fetched
from the realm, game server, or peers into a content-addressed cache. Peers are
untrusted blob sources; the authoritative manifest decides acceptable hashes,
sizes, dependency order, engine compatibility, and capabilities.

The acquisition pipeline must:

1. reserve bounded temporary/quarantine storage;
2. request independently hash-addressed chunks from one or more sources;
3. enforce per-file, package, decompression, concurrency, and time limits;
4. verify every chunk and the complete package against the signed manifest;
5. validate dependency and capability policy in a sandbox;
6. atomically promote the package into the cache only after verification; and
7. pin the verified package identity for admission and replay.

Partial downloads are resumable and cannot be loaded as mods. Cache entries are
immutable by digest, reference-counted while sessions use them, and evicted by
a bounded least-recently-used policy. Revocation metadata prevents newly
joining a known-bad digest without mutating historical replay identity.

Only packages explicitly marked redistributable may enter this flow. Dark Magic
must never serve or peer-distribute Blizzard game data, MPQs, user saves,
credentials, or other private/proprietary content.
