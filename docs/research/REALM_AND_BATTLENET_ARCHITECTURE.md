# Realm, game-server, persistence, and Battle.net responsibility research

> Architecture note: the authoritative game server hosts the pinned
> `d2legacy` Lua rules inside its trusted session process. “Server-authoritative”
> does not require native Go gameplay rules. Go owns authentication, admission,
> clocks, ECS/state, networking, replay, persistence, and sandbox enforcement;
> Lua owns D2 gameplay decisions. Client Lua remains untrusted.

Status: implementation-oriented architecture baseline. This document uses legacy Diablo II/PvPGN service decomposition as behavioral evidence for responsibilities, while designing a modern Dark Magic control plane around the existing transport-independent authoritative session. It does **not** prescribe byte-compatible BNCS/MCP/D2GS implementation yet.

## Executive conclusion

Keep gameplay simulation and realm/account/persistence concerns separated:

```text
Client
  |
  +--> Gateway/Auth/Social adapter
  |      account auth / realm selection / chat presence / session token
  |
  +--> Realm Control Plane
  |      character list/create/delete
  |      game create/join/list
  |      game-server allocation
  |      character lease/lock
  |      durable commit coordination
  |      ladder/season metadata
  |
  +--> Game Server
         one or more authoritative `game/session.Session`
         gameplay command admission
         fixed-tick simulation
         per-client snapshots/events
         transient game checkpoints
         durable character delta handoff

Persistence service/store
  account/character durable records
  revisioned writes
  backups/audit
  ladder/ranking projections
```

Legacy protocols become adapters into these semantic services.

Realm services generally do not load the full gameplay mod. They negotiate and
allocate by immutable mod identity; the selected game server loads and executes
the package. Optional realm Lua is limited to deliberately moddable control
policy such as a season or matchmaking mode, never character inventory or
combat mutation.

## Current Dark Magic composition roots

### `server`

Current `cmd/server` already creates:

- one Akara ECS engine;
- one authoritative `internal/game/session.Session`;
- player and movement command handlers;
- a fixed-tick background server loop;
- a mutable local admin shell.

This is the correct **game-session worker** direction.

The server should gain transport/client membership/snapshot adapters around the session rather than embed networking into ECS/gameplay packages.

### `realm`

Current `cmd/realm` is intentionally a realm control-plane composition root, but today it only runs a mutable headless admin shell.

That is a good placeholder. Grow it into account/character/game-directory/allocation/persistence coordination, not gameplay simulation.

### Runtime API

`internal/app/runtimeapi` is explicitly a local development component-management API, not a player/realm protocol.

Do not repurpose `/v1/components` into a public realm endpoint. Keep lifecycle admin/control traffic separate from player-facing authenticated services.

## Legacy service decomposition evidence

PvPGN's Diablo II server implementation is useful independent corroboration of the broad historical split.

Its `d2cs` configuration describes a **Diablo II game control server** that:

- has a realm identity;
- connects to Battle.net/bnetd;
- knows a list of game servers;
- selects/falls back to least-loaded game servers;
- controls Classic/LoD realm compatibility;
- allows/disallows character creation/conversion;
- maintains character save/info/ladder paths in that implementation;
- limits character count/game list;
- handles queues/timeouts;
- checks game-server version/checksum/password;
- applies game lifetime/level policy.

Its separate `d2dbs` is explicitly a **Diablo II game database server** with:

- game-server allowlist;
- character save/info storage;
- backups;
- ladder storage/update policy;
- connection/keepalive/timeouts.

This is not proof of Blizzard's internal implementation, but it strongly corroborates the protocol-era separation between control/directory, game simulation workers and durable character storage.

## Semantic service boundaries

### 1. Account/auth gateway

Responsibilities:

- authenticate account/session principal;
- issue short-lived realm/game join credentials;
- account bans/roles/admin restrictions;
- social/chat/presence if implemented;
- select realm/region.

It must not be able to mutate character inventory/XP directly.

### 2. Realm control plane

Responsibilities:

- list/create/delete/rename/convert characters according to policy;
- validate Hardcore/Expansion/Ladder mode compatibility;
- game directory create/list/join metadata;
- allocate/select a healthy game server;
- reserve/lease a character for one active game session;
- obtain durable character revision/snapshot;
- coordinate final durable save/checkpoint commit;
- ladder/ranking updates;
- realm-wide special-event inputs;
- moderation/admin metadata.

It should not execute combat ticks.

### 3. Game server

Responsibilities:

- own authoritative GameRules/content generation for game lifetime;
- pin and sandbox the exact authoritative mod package for each session;
- admit authenticated session players/characters;
- run fixed-tick game simulation;
- validate semantic gameplay commands;
- own transient world/game/portal/event/party/trade state;
- produce per-client snapshots/events;
- checkpoint reconnectable game state;
- report durable character mutations/revisions to persistence/control plane;
- release leases on clean leave/commit.

### 4. Durable character store

Responsibilities:

- semantic durable CharacterRecord storage;
- opaque/lossless legacy save fields when needed;
- revision/CAS writes;
- backup/history/audit;
- account ownership and character-mode constraints;
- ladder projection inputs.

It does not resolve movement, loot or combat.

## Character lease / active-session lock

Prevent the same durable character from being mutated concurrently by two game servers.

Use a lease:

```text
CharacterLease
  CharacterID
  DurableRevision
  GameID
  GameServerID
  LeaseToken
  ExpiresAt / renewal
```

Join flow:

1. realm authenticates account/character eligibility;
2. acquire lease on current durable revision;
3. provide signed/opaque admission token to selected game server;
4. game server admits exactly that character/revision;
5. server renews lease while active;
6. final durable commit uses lease token + expected revision;
7. release lease.

Stale/duplicate servers cannot overwrite newer character state.

## Durable revision versus game checkpoint

Keep separate:

```text
DurableCharacterRevision
  survives game destruction

GameCheckpoint
  full transient game state for crash/reconnect
```

A game checkpoint may include monsters, portals, party/trade, active states and player live ECS state. It must not automatically become the durable character save.

Final save extracts/commits only the durable character projection.

## Server crash recovery

Possible policy:

- game worker periodically persists a recoverable checkpoint to ephemeral/shared storage;
- character lease remains associated with GameID;
- replacement worker can restore checkpoint if policy allows;
- clients reconnect using fresh admission credentials;
- only one active worker owns the game generation/lease epoch.

If recovery is not implemented initially, fail closed: release/repair lease through realm timeout and restore character from last committed durable revision rather than accepting conflicting writes.

## Client disconnect versus character save

Disconnect is not necessarily game exit.

Need states:

```text
Connected
DisconnectedGrace
Reconnected
Leaving
Committed
```

During grace, the authoritative player entity may remain in world state according to target rules. Item trade/held state, death and combat continue under server policy.

Do not immediately serialize whatever the client last reported.

## Admission token

Client should not be able to choose:

- GameServerID after allocation beyond the issued target;
- CharacterID it does not own;
- durable revision;
- player/session identity;
- GameRules/content fingerprint.

Realm issues an admission token binding:

```text
Account/Principal
CharacterID
CharacterRevision
GameID
GameServerID
GameRulesFingerprint
ContentFingerprint
expiry/nonce
```

Game server verifies before joining.

## Game creation and allocation

GameCreate request includes semantic player choices such as:

- game name/password/private/public policy;
- difficulty;
- expansion/classic restrictions derived from character;
- Hardcore mode derived/validated;
- max players;
- optional mod/ruleset.

Realm validates and assigns:

- GameID;
- GameRules/content generation;
- selected worker based on capacity/region/version/content support;
- directory visibility.

The game server creates the actual authoritative session.

## Game list / join

Directory entries are projections, not game state authority.

Expose only safe metadata:

```text
GameID/name
current/max players
difficulty/mode
age/region
password/private indicator
compatible content/rules version
```

Joining still revalidates eligibility with the authoritative realm/game worker.

A stale listing cannot grant entry.

## Content and simulation compatibility

Dark Magic is mod-capable. A game must pin a content/rules generation.

Network handshake/admission should compare:

- engine/protocol schema;
- semantic game rules fingerprint;
- simulation-affecting table/content fingerprint;
- mod/plugin IDs/versions/hashes where required.

The stable authoritative identity includes mod ID/contract version, package and
Lua source/bytecode hashes, the full dependency graph and hashes, gameplay
configuration identity, and required engine capability/API versions. Realm
allocation, admission tokens, handshakes, replay headers, checkpoints,
reconnect/late join, recovered sessions, and persistence compatibility checks
all bind or validate that same identity. Changed scripts apply to new sessions;
an active session changes only through an explicit versioned state migration.

Presentation-only mods can be negotiated separately if safe.

Never allow peers to silently simulate different Skills/MonStats/Items/map rules.

## Authoritative command transport

Existing `Session.Submit` is a strong target.

Transport adapter responsibilities:

- authenticate connection -> SessionPlayerID;
- decode message schema;
- enforce rate/size/sequence windows;
- set/bind `Player` and `Authority` server-side;
- translate to semantic `simulation.Command`;
- call `Session.Submit`;
- return rejection/events/snapshot deltas.

Do not expose arbitrary command authority or admin/system command kinds to ordinary clients.

## Tick and input timing

Game server owns current tick.

Client sends semantic inputs with sequence/timing hints under a bounded lead/prediction policy. The server canonicalizes/admit/rejects based on the existing `Admitter` window.

Do not let packet arrival order decide same-tick results.

Prediction has three explicit levels: no gameplay prediction; limited generic
movement/presentation prediction; and optional shared-`d2legacy` prediction with
rollback/reconciliation. Start with limited movement/presentation prediction.
Even identical client scripts are untrusted; canonical server snapshots and
events always win.

## Snapshot/event projection

The game server should project per-client state from one canonical simulation.

Separate:

```text
StateSnapshot/Delta
  current authoritative facts

SemanticEvents
  ordered occurrences since cursor/ack
```

Examples of state:

- nearby entities/positions/modes/resources;
- player's inventory/equipment/skills/quests;
- party/hostility relations;
- world object state;
- game rules/tick.

Events:

- hit/death;
- item drop/pickup;
- quest update;
- chat/system message;
- sound/presentation cue identity where not locally derivable.

## Interest management

Do not send every entity in the game to every client.

Interest can derive from authoritative active rooms/level/party/portal context:

- nearby active rooms/entities;
- party roster info beyond local area where original UI needs it;
- private player state only to owner;
- shared game/quest event summaries.

Interest filtering is network projection, not simulation culling.

## Private information filtering

Never expose to unrelated clients:

- account identifiers/tokens;
- private inventory/stash unless trading/inspection policy requires it;
- hidden gamble seeds/final outcomes;
- unrevealed map state;
- private quest/NPC flags;
- admin/audit details.

Use typed per-client projections rather than serializing raw ECS snapshots over the network.

## Persistence commit projection

At save/leave, game server builds a semantic durable character update from authoritative state:

- progression/stats/skills;
- inventory/equipment/stash/personal state;
- quest/waypoint/NPC state;
- hireling durable record;
- Hardcore death/status;
- metadata needed for legacy save round-trip.

Transient monsters/party/trade/portal state is not included in character durable state.

Commit uses expected durable revision/lease.

## Ladder/ranking

Ladder is a derived realm projection from committed durable character state, not an input that the game server can arbitrarily edit.

On qualifying commit:

```text
CharacterRevision committed
 -> ranking/ladder projection update
```

Keep ranking rebuildable from durable records/events where possible.

## Realm-wide special events

Later realm-wide events can enter games as authenticated semantic event activations from realm/event service.

Do not let game clients increment global counters or spawn special bosses directly.

See `ENDGAME_AND_SPECIAL_EVENTS.md`.

## Chat/presence

Chat/social presence may live in gateway/realm services rather than the game simulation. In-game chat messages still need authenticated sender, ignore/squelch filtering and game/party/channel routing.

Chat timing/content generally does not affect deterministic gameplay and need not enter replay checksum unless a gameplay command explicitly derives from it.

## Legacy protocol adapters

If Dark Magic later supports original clients, create adapter services/modules for historical protocol families:

```text
BNCS-like gateway/auth/chat
MCP/realm-like character/game directory
D2GS-like gameplay transport
```

Adapters translate packet state machines into the semantic services above.

Do not make internal character IDs, GameRules, item storage or ECS layouts match packet layouts merely for compatibility.

## Security model

Trust boundaries:

- client is untrusted;
- gateway authenticates principal;
- realm authorizes character/game/worker;
- game server is authoritative simulation;
- authoritative `d2legacy` executes only inside that trusted server/session
  boundary; client copies never gain command or outcome authority;
- persistence trusts only authenticated realm/game service identities + revision/lease checks;
- admin APIs use separate privileged credentials/policies.

All cross-service mutation should be authenticated, replay-resistant and auditable.

## Observability

Use stable IDs across logs/traces:

```text
AccountID (restricted)
CharacterID
GameID
GameServerID
SessionPlayerID
Command sequence/tick
Durable revision
Checkpoint checksum
Content/rules fingerprints
```

Support deterministic replay export for debugging without exposing private account secrets.

## Suggested implementation slices

### R1 — semantic IDs and game directory model

Define GameID/GameServerID/CharacterID/session admission structures with no network listener yet.

### R2 — character lease/revision service

Implement in-memory lease + CAS durable character repository behind interfaces and crash/stale-write tests.

### R3 — remote command adapter

Build a local loopback/test transport that binds a connection principal to a session player and submits ordinary movement commands through `Session.Submit`.

### R4 — snapshot/event projection

Expose one typed per-client authoritative snapshot/delta stream with private-state filtering.

### R5 — realm game create/join allocation

Have `realm` manage worker registration/capacity, create a game on `server`, and issue opaque admission tokens.

### R6 — disconnect/reconnect/checkpoint

Implement grace period + checkpoint restoration/rejoin token flow.

### R7 — durable leave commit

Extract authoritative CharacterRecord, CAS commit expected revision, release lease and update directory/ladder projection.

### R8 — legacy protocol adapters later

Only after semantic APIs are stable, research/implement BNCS/MCP/D2GS packet compatibility as adapters.

## Verification backlog

1. Historical BNCS/MCP/D2GS connection/state-machine responsibilities by target client version.
2. Realm character create/delete/convert restrictions and status bits.
3. Original game create/list/join metadata and filters.
4. Game-server allocation/queue behavior.
5. Character lock/load/save/release protocol and crash behavior.
6. Original reconnect/disconnect persistence window.
7. D2GS command/input sequencing/anti-cheat validation.
8. Snapshot/unit visibility/interest rules.
9. Party/hostility/trade replication fields.
10. Game destruction/idle/lifetime policy.
11. Hardcore dead-character realm visibility/join restrictions.
12. Ladder update/season semantics.
13. Realm-wide event inputs in later patches.
14. Original-client content/version checksum negotiation.
15. Chat/presence/ignore routing responsibilities.
16. Rate limits/malformed packet behavior for compatibility adapters.

## Primary sources inspected

- Current Dark Magic `cmd/server`, `cmd/realm`, `internal/game/session`, persistence/player/item/game-rules research and local runtime admin API.
- PvPGN pinned repository commit `9cd173f4e02ba3d9f8f15a67ca308b5eb78723e4`, especially `d2cs` and `d2dbs` configuration/source split as independent historical implementation evidence.
- D2MOO gameplay server/client/session code for game-side authority behavior.

PvPGN is corroborating implementation evidence, not proof of Blizzard's private server topology. The modern Dark Magic service decomposition is intentionally semantic and protocol-independent.
