# Multiplayer, realm, and gameplay-UI research handoff for Codex

Status: implementation handoff for the final major gameplay research cluster.

## Read first

1. [MULTIPLAYER_PARTIES_PVP_AND_TRADE.md](MULTIPLAYER_PARTIES_PVP_AND_TRADE.md)
2. [REALM_AND_BATTLENET_ARCHITECTURE.md](REALM_AND_BATTLENET_ARCHITECTURE.md)
3. [GAMEPLAY_UI_CONTRACTS.md](GAMEPLAY_UI_CONTRACTS.md)
4. [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md)
5. [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md)
6. [DIFFICULTY_AND_GAME_MODES.md](DIFFICULTY_AND_GAME_MODES.md)
7. [ITEM_SYSTEM.md](ITEM_SYSTEM.md)

## Current architecture to preserve

- `internal/game/session.Session` remains the only game simulation authority.
- Local intent sources and remote/network commands both converge on semantic `simulation.Command` admission.
- Akara ECS plus registered `StateParticipant`s remain canonical game/checkpoint state.
- item/player/world/quest/etc. packages remain gameplay owners.
- `cmd/server` remains the game-worker composition root.
- `cmd/realm` remains the realm/control-plane composition root.
- local runtime component HTTP API remains development/admin-only, not a realm/player protocol.
- Lua presentation reads snapshots and submits intents through versioned capabilities.

## Highest-leverage implementation order

### 1. Semantic session player identity

Introduce stable `SessionPlayerID` distinct from account and durable character identity. Bind all ordinary player commands to it server-side.

Keep current string player IDs as adapters/tests if convenient, but make spoofing impossible at a transport boundary.

### 2. Party state participant

Add deterministic party/invite membership as session state with semantic commands:

```text
invite
cancel
accept
leave
```

Expose same-level living-member queries for loot/quest/XP consumers.

### 3. Shared reward integration

Use party state in one existing real path, preferably TreasureClass NoDrop effective player count, because that work is already researched and bounded.

### 4. Hostility relation

Add server-owned directional/mutual hostility relationships and deterministic future cooldown tick. Enforce town/level-9 rules for the targeted 1.10f policy after vectors.

Feed hostility into one centralized combat target-legality service.

### 5. Trade session/escrow

Add `TradeSession` as session state with stable item IDs, trade escrow placements, offered gold, offer revision and per-player acceptance.

Any offer change invalidates acceptance.

### 6. Atomic trade commit/cancel

Validate both inventories/gold/locks/revision then exchange all or none. Cancel safely on disconnect/death/hostility/leave according to verified rules.

### 7. Versioned semantic UI snapshots

Before remote networking gets large, formalize local Lua view models for:

- PlayerHUD
- Item/container/tooltip
- Quest
- Waypoint
- NPC/vendor
- Party
- Trade

Use stable ordered arrays/revisions and structured reason codes.

### 8. Character durable repository + lease

Create an interface/in-memory implementation supporting:

```text
Get(CharacterID)
AcquireLease(expected revision, GameID/worker)
Renew
CompareAndSwapCommit
Release
```

Test stale-worker/crash/concurrent-game scenarios.

### 9. Realm game directory/worker registry

`realm` owns:

- registered game workers/capacity;
- create/list/join directory;
- GameRules/content fingerprint;
- character eligibility/lease;
- opaque admission token.

No real network protocol is required for the first slice.

### 10. Loopback remote transport

Build an in-process/test transport adapter that:

- authenticates a synthetic principal;
- verifies an admission token;
- binds connection to SessionPlayerID;
- submits movement/skill commands;
- receives typed snapshots/events.

Prove gameplay code does not know/care whether input is local or remote.

### 11. Per-client projection/interest

Add owner/private and nearby/public projections without serializing raw ECS snapshots to clients.

Reuse semantic view builders with Lua where practical.

### 12. Disconnect/reconnect

Introduce disconnected grace state, checkpoint/rejoin token, trade cancellation policy and character lease renewal/release.

### 13. Durable leave commit

Extract semantic durable character state from authoritative session, CAS commit expected revision, update ladder projection if applicable and release lease.

### 14. Legacy protocol adapters last

Once semantic services are stable, create separate adapters for historical gateway/realm/game protocol families if original-client interoperability is desired.

Do not shape core APIs around packet structs.

## Trust matrix

| Actor | May request | May decide |
| --- | --- | --- |
| Lua/local UI | semantic gameplay/UI intent | local presentation state only |
| Remote client | authenticated semantic intent | prediction/presentation only |
| Game server/session | gameplay validation/mutation | all game/transient world state |
| Realm | character/game admission/allocation | leases, directory, game rules/content selection |
| Durable store | revisioned data persistence | CAS/lease enforcement, not game rules |
| Admin | explicit privileged control | only through audited admin authority |

## Multiplayer state participants

Candidate non-ECS state to register/checkpoint:

```text
GameRules
PartyState
Hostility/PlayerRelations
TradeSessions
UniqueDropOccurrence
Portal/GameEvent state
Vendor/Gamble state if outside ECS/item archive
Realm admission/player membership metadata needed for reconnect
```

Prefer ECS when state naturally belongs to entities; use StateParticipant for game-level indexed/transactional state.

## Command invariants

Network transport must set/bind these, never trust client values:

```text
Player/SessionPlayerID
Authority = player
GameID/session
allowed tick window
sequence policy
```

Payload contains only action-specific semantic input.

System/Admin commands require separate authenticated channels and audit.

## UI projection invariants

Every gameplay view model should:

- be immutable/copied;
- use stable semantic IDs;
- have deterministic ordering;
- carry revision/tick where stale actions matter;
- omit hidden/private server facts;
- remain serializable for local Lua and remote protocol adapters;
- never expose native pointers/ECS internals/renderer handles.

## Structured rejection

Before networking expands, introduce stable rejection codes for commands/transactions.

Example categories:

```text
stale_revision
out_of_range
invalid_target
not_authorized
insufficient_resource
inventory_full
incompatible_game_rules
cooldown
not_hostile
trade_offer_changed
```

Human text/localization is presentation. Lua/network clients should not parse Go error strings.

## Prediction policy

Start with movement prediction only.

Client/local presentation may predict:

- movement/interpolation;
- animation start;
- hover/drag prospective state.

Do not predict durable item transfers, loot generation, quest completion or trade commit as canonical state.

Add skill/combat prediction only after server correction/event semantics are stable.

## Realm/game-worker interfaces

Keep semantic interfaces protocol-neutral:

```text
Realm.CreateGame
Realm.ListGames
Realm.JoinGame
Realm.ListCharacters
Realm.AcquireCharacter

GameWorker.CreateSession
GameWorker.AdmitPlayer
GameWorker.RemovePlayer
GameWorker.SubmitCommand
GameWorker.SubscribeView
GameWorker.Checkpoint/Restore

CharacterStore.Get/CAS/Lease
```

Concrete gRPC/HTTP/custom-binary/in-process adapters can be chosen later.

## Legacy adapter rule

For historical client support:

```text
packet state machine
 -> semantic request
 -> modern authority
 -> semantic snapshot/event
 -> packet encoding
```

Never:

```text
legacy packet bytes
 -> direct ECS/item/save mutation
```

This keeps security, replay, mods and modern clients consistent.

## Recommended PR granularity

```text
session player identity
party state
NoDrop party integration
hostility state
trade session
trade atomic commit
structured command errors
PlayerHUD view model
ItemView model
Quest/Waypoint view model
character lease repository
game directory/worker registry
loopback remote command transport
per-client snapshot projection
disconnect/reconnect
durable leave commit
```

## Definition of a strong multiplayer vertical slice

```text
realm creates GameID + pins GameRules/content
 -> character lease acquired
 -> worker creates authoritative Session
 -> two loopback clients receive admission identities
 -> both submit movement through Session.Submit
 -> party invite/accept changes one shared PartyState
 -> monster drop NoDrop calculation observes same-level party state
 -> per-client HUD/world snapshots arrive
 -> one client disconnects/rejoins checkpointed game
 -> final character revision commits with lease/CAS
```

A strong trade slice:

```text
two players start trade
 -> item IDs move to server escrow
 -> each receives projected offer + revision
 -> both accept same revision
 -> authority validates destination capacity/gold
 -> one atomic item/gold exchange
 -> replay/checkpoint hash reproduces exact ownership
```

## Verification discipline

Use [MULTIPLAYER_REALM_UI_VERIFICATION_QUEUE.md](MULTIPLAYER_REALM_UI_VERIFICATION_QUEUE.md). Original protocol compatibility, party/PvP formulas, trade timing and exact UI-visible fields remain empirical work, while the modern authority split can proceed independently.
