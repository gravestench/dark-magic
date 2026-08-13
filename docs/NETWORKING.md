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

QUIC begins with conservative 1200-byte packets and keeps path-MTU discovery
enabled. Application datagrams must fit the transport's negotiated maximum and
an additional schema-specific ceiling (initially at most 1000 application
bytes, leaving headroom inside a 1200-byte packet); they are never split at the
application layer or sent at an assumed Ethernet MTU. Oversized semantic state uses bounded
reliable streams, allowing QUIC to packetize it safely. The first adapter keeps
datagrams disabled until compact delta schemas and loss tests exist, limits a
wire frame to 4 MiB, and limits command payloads to 8 KiB.

This is a direction, not a claim that QUIC is universally fastest. Benchmarks
must cover representative tick rates, message sizes, loss, jitter, NATs, CPU,
and memory before tuning encodings or assigning a message class to datagrams.
The reliable baseline is exercised through caller-provided production packet
connections under deterministic bidirectional packet loss and cyclic delay.
That acceptance covers handshake, join, the long-lived correction stream,
command submission, reconnect, and leave. It proves bounded recovery behavior;
it is not a substitute for sustained load, diverse paths, or platform testing.
The adapter accepts at most 16 concurrent bidirectional streams per connection,
disables peer-initiated unidirectional streams, caps a stream receive window at
4 MiB and a connection receive window at 8 MiB, and explicitly releases both
sides of completed unary streams. A repeated malformed-stream test verifies
that 256 rejected requests neither exhaust stream credit nor prevent a later
valid session lifecycle on the same connection.

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

Listen-server TCP/IP play keeps its user flow address-only. A host name or IP
uses port `6112` unless a port is supplied. The application configuration
directory's `network` folder stores the persistent owner-only
`host-identity.pem`, its `host-certificate.pem`, and client
`known-hosts.json` trust-on-first-use pins keyed by normalized `host:port`.
First contact records the certificate fingerprint; a later identity change is
rejected rather than silently replacing it. Realm endpoint fingerprints remain
assignment-owned and separate from this direct-game trust store.

Join and reconnect return versioned per-player semantic projections plus a
canonical tick and checksum. They do not expose raw ECS snapshots or hidden
server facts. Reconnect rotates the bearer credential so a successfully used
old credential cannot be replayed.

The native server enables QUIC only when `-quic-listen`, `-tls-cert`,
`-tls-key`, and `-admission-key` are supplied together. TLS uses an explicitly
configured certificate; clients must use an explicit trust root. The admission
key file must contain at least 32 bytes, is bounded to 4096 bytes, and must not
be group/world accessible. A realm and its allocated worker share this key to
sign and consume short-lived, session-bound, one-use admission tickets. The
standalone server does not expose a ticket-minting endpoint.

The initial `d2legacy` remote projection is `PlayerHUD/v1`. It derives from the
canonical checkpoint rather than rereading the live ECS and selects the entity
bound to the authenticated player. Its field allowlist includes identity,
vitals, progression, combat display values, position, and location. It excludes
inventory/belt contents, other players' private state, raw component stores,
and hidden server facts. Realm admission must load and lease the durable
character and submit the trusted player-entry command before join; a client
credential never materializes authoritative character state.

`ClientView/v1` envelopes `PlayerHUD/v1` and `WorldView/v1` at one canonical
tick. The world projection includes only explicitly public selectable fields
within 80 subtiles, excludes the authenticated player's own entity, sorts by
distance then stable public ID, rejects malformed or duplicate IDs, and caps
the result at 256 entities. Monster health is exposed only through this reviewed
projection; AI targets, damage, raw components, far entities, inventories, and
other hidden facts remain excluded. `WorldDelta/v1` contains deterministic
upserts and removals. A truncated base or result forces a complete bounded
reset because removals cannot otherwise be proven.

Realm join is a transaction across the durable character repository and the
allocated worker. It acquires an exclusive revisioned lease owned by the
account, validates the character's pinned runtime compatibility, creates the
trusted next-tick player-entry command with a realm-owned monotonic sequence,
and returns the worker address, TLS fingerprint, exact runtime identity, public
character revision, and short-lived ticket. The ticket additionally binds the character revision
and runtime identity hash. Validation or submission failure releases the lease
and revokes the unused ticket. Active memberships renew their lease through the
realm; reconnect never accepts client-reported durable state.

Character ownership is independent of topology. Single-player, listen-server,
and self-hosted dedicated-server play use the player's local profile roster;
those hosts may admit player-controlled save data according to their own policy,
but that data never becomes realm-trusted. Realm play begins with account login
or initial account creation, loads only the characters associated with that
account, and requires character selection before browsing, creating through the
realm, or joining a realm game. A new realm account therefore begins with no
characters just as a new local profile does.

The character lease is a realm/worker capability and never crosses the client
assignment boundary. A trusted realm worker commit must present the active,
unexpired lease, preserve character identity, atomically replace a defensive
copy, increment the realm revision, and consume the lease. Empty, expired,
foreign, and replayed leases are rejected. The d2legacy selection store is
explicitly player-profile-owned and cannot implement or call this realm
repository contract, so copying a self-hosted character value does not confer
trusted realm write authority.

Player-profile persistence uses a separate `Profile/v1` envelope. It bounds a
profile to 4 MiB and 64 characters, rejects unknown fields, duplicate or absent
selected identities, unsupported versions, trailing data, oversized input, and
SHA-256 integrity mismatches. Writes use private `0600` temporary files, sync
file contents, atomically replace the destination, and sync its directory.
Decoded rosters, stats, appearance maps, and selection are defensively copied.
This durability does not change their player-controlled trust classification.

Realm persistence never accepts a client-authored character replacement. At
commit, Admissions captures the worker session's canonical checkpoint and
projects the leased player's durable subset by the server-bound player ID and
character ID. Session-owned identity, level, experience, health, mana, and
defense replace the leased baseline; expansion/hardcore, appearance, attributes,
resistances, and other fields not yet owned by that projection are preserved.
Identity mismatch, missing canonical components, invalid numeric state, stale
leases, and replayed commits fail without changing the durable record.

Self-hosted dedicated servers have a separate explicit profile path. Host
configuration selects a private `Profile/v1` file, stable player ID, and
authoritative spawn destination; startup queues the selected character through
the ordinary system-authority entry command. Clean shutdown projects the same
canonical session-owned subset back into the profile and atomically persists
it. No profile bytes are added to realm join messages, and this operation never
creates realm ownership or a realm lease.

Remote self-host admission remains a distinct pre-join operation. The client
sends one selected `CharacterOffer/v1` (not its full roster), bounded to 128 KiB
and protected by strict schema and SHA-256 integrity validation. The host
authenticates a protected configured credential, supplies principal/player
identity and destination itself, permits at most eight attempts and one
successful admission, queues the system entry, and returns a short-lived
one-use ordinary session ticket. QUIC exposes `profile_admit` only when this
self-host policy is installed; realm and default servers reject it. The command
server enables it with a protected `--remote-profile-key` file and the explicit
profile player/destination flags. TLS still protects the credential in transit.
The client verifies normal X.509 trust plus the pinned leaf fingerprint before
sending the profile credential, encodes only its selected defensive profile
copy, validates the returned session/runtime/character identities, and then
uses the same session object as realm connections. Because trusted entry is a
next-tick command, join permits only the explicit “player absent” projection
error to wait for readiness, polling for at most two seconds or the caller's
earlier cancellation; all other projection errors fail immediately.

The client treats the returned assignment as untrusted discovery data. It
accepts only a canonical host/port endpoint, validates the advertised runtime
identity locally, performs normal X.509 verification against an explicit trust
configuration, and additionally pins the leaf certificate to the realm's
`sha256:` fingerprint before sending the one-use ticket. It then verifies the
server admission session/runtime and decodes exactly `ClientView/v1`. Reconnect
rotates the session credential and atomically replaces the correction view;
command submission, refresh, reconnect, and close serialize credential access.
Corrections are monotonic: the client rejects an older tick and rejects a
different checksum for an already-installed tick.

Protocol v2 also publishes the authoritative fixed-step duration and highest
contiguous input sequence applied for that player. The client schedules inputs
two ticks ahead of a bounded extrapolation of the latest authoritative clock,
retains unacknowledged inputs in sequence order, and discards them only when
that contiguous acknowledgement advances.

Normal network reordering is not a simulation error. Network admission
validates identity, authority, schema, payload, and maximum lead independently
of arrival order, rejects duplicate player sequences, and sorts admitted input
canonically at execution. Input arriving after its intended tick restores a
complete ECS plus registered-runtime frame and deterministically replays at
most eight ticks. Restore/replay is transactional: failure restores the live
world, participant state, pending commands, replay log, checkpoints, history,
and acknowledgement state together. Older input is rejected and corrected by
the next canonical projection.

The live game-world acceptance in `internal/app/clientsession` crosses this
entire boundary without substituting a fake projection: it starts the embedded
production d2legacy Lua authority, advances the authoritative ECS session,
populates a deterministic Blood Moor hostile from synthetic TXT records,
offers a selected local-profile character over real QUIC, and observes both the
owner-private HUD and nearby public hostile from canonical checkpoints. The
fixture deliberately supplies redistributable synthetic record rows so normal
CI does not depend on a developer's owned Diablo II installation. It also sends
a real `player.move` intent and waits on the production correction stream for
the authoritative position change. The headless Lua authority owns position
integration; interactive world composition adds collision data and camera
behavior but is not required for movement to execute.

The interactive client's TCP/IP Host button uses the same topology rather than
starting a privileged Lua listener. `engine.network/v1` carries only host/join
intent and copied safe status. The client application owns an in-process listen
host, ephemeral TLS material, QUIC server, selected-profile admission, and the
hosting player's ordinary `clientsession.Session`. This makes the host a real
network member and establishes the lifecycle seam that remote peers will share;
the subsequent presentation slice will render and submit input through that
connected session instead of the offline ECS.

Authenticated refresh carries a complete bounded view over a reliable QUIC
stream, from which the client derives `WorldDelta/v1`. A long-lived correction
stream sends an immediate view and then changed views at no more than 10 Hz. Its
one-element application channel deliberately propagates a slow consumer back
to QUIC flow control instead of building an unbounded queue. Each membership
has independent token buckets allowing a burst of 32 commands at a sustained
16 commands/second and a burst of 4 correction requests at 2 requests/second.
Reconnect preserves those budgets so credential rotation cannot reset them.
Wire requests have one strict operation-specific shape and reject ambiguous
fields. These reliable paths remain the correctness baseline before compact
lossy datagrams; correctness and removals never depend on a datagram.

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
