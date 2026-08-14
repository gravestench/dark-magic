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

The native game adapter uses `quic-go`. Join, reconnect, critical
commands, inventory changes, chat, and full/correction snapshots use reliable
streams. Compact, replaceable transform samples use QUIC datagrams because a
new sample is more useful than a late one. Correctness, entity membership,
private state, and removals never depend on a datagram.

QUIC begins with conservative 1200-byte packets and keeps path-MTU discovery
enabled. Application datagrams must fit the transport's negotiated maximum and
an additional schema-specific ceiling (initially at most 1000 application
bytes, leaving headroom inside a 1200-byte packet); they are never split at the
application layer or sent at an assumed Ethernet MTU. Oversized semantic state
uses bounded reliable streams, allowing QUIC to packetize it safely. Transform
frames have a 1000-byte application ceiling, prioritize the projection's
nearest entities, and are truncated rather than fragmented when the bounded
world view does not fit. Reliable wire frames remain limited to 4 MiB and
command payloads to 8 KiB.

This is a direction, not a claim that QUIC is universally fastest. Benchmarks
must cover representative tick rates, message sizes, loss, jitter, NATs, CPU,
and memory before tuning encodings or assigning a message class to datagrams.
The reliable baseline is exercised through caller-provided production packet
connections under deterministic bidirectional packet loss and cyclic delay.
That acceptance covers handshake, join, the long-lived correction stream,
command submission, reconnect, and leave. It proves bounded recovery behavior;
the separate sustained harness described below adds multi-client load, explicit
reordering, active credential rotation, socket redial, and platform coverage.
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
server facts. The client treats even a pinned host as an untrusted decoder
boundary: `ClientView/v4` rejects unknown fields, inconsistent nested ticks,
non-finite numbers, duplicate identities, invalid private interaction shapes,
and collections or strings beyond their schema limits. Reconnect rotates the
bearer credential so a successfully used old credential cannot be replayed.

The native server enables QUIC only when `-quic-listen`, `-tls-cert`,
`-tls-key`, and `-admission-key` are supplied together. TLS uses an explicitly
configured certificate; clients must use an explicit trust root. The admission
key file must contain at least 32 bytes, is bounded to 4096 bytes, and must not
be group/world accessible. A realm and its allocated worker share this key to
sign and consume short-lived, session-bound, one-use admission tickets. The
standalone server does not expose a ticket-minting endpoint.

The current `d2legacy` remote projection is `PlayerHUD/v5`. It derives from the
canonical checkpoint rather than rereading the live ECS and selects the entity
bound to the authenticated player. Its field allowlist includes identity,
vitals, progression, combat display values, position, location, movement mode,
skill assignments, learned skills, and belt contents. `PrivateView/v1` separately
projects only that authenticated player's layout, items, and active interaction.
Both exclude other players' private state, raw component stores, and hidden
server facts. Realm admission must load and lease the durable
character and submit the trusted player-entry command before join; a client
credential never materializes authoritative character state.

`ClientView/v4` envelopes `PlayerHUD/v5`, `PrivateView/v1`, and `WorldView/v1`
at one canonical tick. The world projection includes only explicitly public selectable fields
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
server admission session/runtime and decodes exactly `ClientView/v4`. Reconnect
rotates the session credential and atomically replaces the correction view. An
unexpected transport loss suspends the membership for a ten-second lease,
rejects commands on the disconnected credential, and permits a fresh pinned
QUIC connection to rotate that credential before deterministic player removal;
the client retries with bounded exponential backoff. A reconnect nonce makes
credential rotation idempotent when the server response itself is lost. Exact command retransmits
are idempotently acknowledged while a conflicting payload for an accepted
sequence remains rejected. Command submission, refresh, reconnect, and close
serialize credential access.
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
network member. Transport starts only after character selection. Connected
presentation uses an entity-empty, schema-compatible client ECS rather than the
frozen offline authority: authenticated HUD state creates the local entity,
`WorldView/v1` creates nearby public entities, owner-private projections rebuild
the local inventory/interaction graph, and Lua binds to the authenticated
session player ID. Input is sampled on the 25 Hz simulation clock, scheduled
against an extrapolated server tick, retained until contiguous acknowledgement,
and replayed as limited local movement prediction. The same production movement
rules and collision integrator serve authoritative Lua and prediction, so policy
does not drift into a second client implementation. Canonical corrections reset
the prediction baseline. A separate render transform decays only reconciliation
error, keeping new local input immediate instead of filtering all owner motion.

The client maintains two server-derived timelines. The prediction timeline
estimates current authority using the latest tick and half the smoothed RTT. The
remote interpolation timeline stays behind by a bounded, jitter-adaptive delay
and slews between 0.75x and 1.25x to absorb timing error without pausing or
jumping. A 32-snapshot immutable ring interpolates peers and extrapolates them
for at most the configured outage window; it then freezes instead of allowing a
runaway projection. Lifecycle and non-transform metadata change only at
canonical snapshot boundaries.

The network goroutine decodes and merges corrections into immutable presentation
snapshots published through an atomic pointer. The render thread never clones a
live authoritative world or waits for JSON decoding. Disposable 25 Hz transform
datagrams use a one-element latest-wins channel, while complete 10 Hz reliable
views repair loss and carry structural/private state. Player animation carries
an authoritative `start_tick`; clients derive playback phase from the matching
prediction or interpolation timeline instead of accumulating render-frame time.

Listen and dedicated authorities install every generated level's collision map
and route movement collision by each entity's authoritative level. Public
interest filtering requires the same act and level before applying distance.
One canonical checkpoint is captured per server tick and shared by all per-player
projections. Each membership owns at most one long-lived correction stream.
Explicit leave revokes membership immediately. QUIC connection loss starts the
reconnect lease; expiration revokes membership and submits the same deterministic
`system.player.leave` command, which removes the player, admission-owned state,
and its marked null interaction target without leaking shared target definitions.
Self-host profile throttling is per normalized remote IP
and refills over time; one abusive address cannot exhaust a process-global join
counter.

Single-player, listen-host, direct-join, and dedicated-server characters remain
player-profile data, never realm authority. Clean single-player shutdown
projects session-owned fields from the canonical checkpoint. Clean listen-host
and direct-join shutdown first requests a reliable canonical correction, then
merges its authenticated HUD into the selected local baseline. Dedicated
servers perform the equivalent projection from their canonical checkpoint.
All three paths preserve profile-only fields and reject any attempt to replace
the selected character identity before `Profile/v1` is written atomically.

Authenticated refresh carries a complete bounded view over a reliable QUIC
stream, from which the client derives `WorldDelta/v1`. A long-lived correction
stream sends an immediate view and then changed views at no more than 10 Hz. Its
one-element application channel deliberately propagates a slow consumer back
to QUIC flow control instead of building an unbounded queue. Each membership
has independent token buckets allowing a burst of 64 commands at a sustained
32 commands/second and a burst of 4 correction requests at 2 requests/second.
Reconnect preserves those budgets so credential rotation cannot reset them.
The client permits at most eight independent reliable command streams in flight,
so input cadence is not serialized behind one network round trip; credential
rotation takes an exclusive gate and waits for those old-credential requests.
Wire requests have one strict operation-specific shape and reject ambiguous
fields. These reliable paths remain the correctness baseline beneath compact
lossy datagrams; correctness and removals never depend on a datagram.

## Network test strategy

The renderer-free acceptance in
`internal/app/clientapp/network_motion_acceptance_test.go` drives the production
clock and snapshot buffer at 60 Hz while applying deterministic latency, jitter,
20 percent loss, and reordering. It asserts monotonic peer motion, a bounded
per-frame displacement, few stationary frames during continuous motion, and a
bounded freeze after an outage. Add new presentation policies to this harness
as deterministic schedules with perceptual invariants rather than sleeps or
wall-clock tolerances.

`internal/app/gameserver/sessionquic/impairment_test.go` exercises the real QUIC
adapter through an injected packet connection. Its sustained acceptance runs
three independently authenticated clients at the production 25 Hz command
cadence under bidirectional drop, cyclic delay, jitter, and explicit packet
reordering. One active member rotates credentials and later loses its socket,
redials, reconnects, and continues its uninterrupted command sequence. The test
requires every command to apply, correction and transform streams to converge,
all application channels to remain one-slot bounded, and actual impairment plus
RTT/transform telemetry to be observed.

The normal suite runs 80 ticks. Set `DARK_MAGIC_NETWORK_SOAK_TICKS` from 20 to
15000, run `make test-network-soak` for the 1500-tick preset, or run
`make test-network-hardening NETWORK_SOAK_TICKS=<ticks>` for another bounded
duration. `.github/workflows/network-hardening.yml` runs the production-path
packages on macOS, Linux, and Windows for pull requests and runs the long preset
weekly or on demand. The same workflow continuously fuzzes reliable frame,
transform datagram, and client projection decoders. `make test-network-fuzz`
runs those fuzz boundaries locally. The transform codec benchmark and allocation
budget pin the MTU-bounded hot path.

`internal/app/clientsession/live_gameworld_test.go` boots the production
`d2legacy` Lua authority and verifies that the live transport reaches a generated
game world. Package tests beside `networkclock`, the presentation buffer, local
prediction, shared movement rules, and transform codec pin their smaller
contracts. Run `make test` and `make test-race` for the complete gate; use focused
package tests while iterating. Run `make test`, `make test-race`, and
`make test-network-hardening` for the complete local gate.

## Join-time mod acquisition

The production direct-host path now advertises the complete canonical runtime
recipe over pinned TLS before profile admission. It includes embedded
`d2legacy`, the ordered extension descriptors, engine/protocol contracts, and
authoritative/configuration hashes. A joiner downloads missing redistributable
extensions from the host, recomposes its VFS and client components, recomputes
the same recipe locally, and only then offers the selected character and starts
the ordinary authenticated join.

Package bytes travel on bounded reliable QUIC streams in 32 KiB application
chunks. The connection uses a per-peer byte burst/refill limit. Client recipes
are bounded by extension count, per-package size, and total bytes. A download
streams into quarantine and is promoted atomically only after full SHA-256,
size, ZIP-safety, manifest identity/version, kind, dependency/order, and
redistributability checks. Exact content-addressed versions may coexist across
concurrent sessions without rewriting the user's enabled profile.

Only a package in the host's recipe and explicitly marked redistributable can
be served. Blizzard game data/MPQs, saves, credentials, and private keys never
enter the protocol. Lua package caches are invalidated when a recipe removes or
changes an extension, so previously loaded code cannot survive a lock change.
If admission or later join startup fails after recomposition, the client
restores the extension recipe selected at process startup; downloaded blobs
remain installed but are not silently enabled in `mods.json`.

M23 still owns realm-signed publication and revocation, resumable/multi-source
downloads, cache reference counting/eviction, and interrupted-download resume.
Those future sources remain untrusted mirrors; no source may replace the exact
realm/session recipe.
