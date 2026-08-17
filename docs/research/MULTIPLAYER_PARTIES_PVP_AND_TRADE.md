# Multiplayer parties, hostility, PvP, trading, and shared-progress research

Status: implementation-oriented research baseline. Dark Magic has a
transport-independent authoritative fixed-tick session, stable item
identity/containers, trusted player/item/world commands, deterministic
replay/checkpoints, and a checkpointed `d2legacy.party/v1` authority for
invite/cancel/accept/leave, stable membership identity, same-level living-member
queries, game-departure cleanup, reconnect continuity, and the first shared
reward hook through credited-owner same-level party NoDrop context. Party XP,
quest/gold sharing, projection, hostility, and trade remain. This document
defines those multiplayer gameplay systems without copying the legacy D2GS
protocol.

## Executive conclusion

Multiplayer should add **shared authoritative relationships and transactions**, not a second simulation path.

```text
one authoritative Game Session
  players
  party relationships
  hostility/loot permission relationships
  trade transactions
  shared quest/event state
  per-player quest/waypoint/NPC state
  shared world/monster/item/object state
        |
        v
same movement/skill/combat/item/quest/world handlers
        |
        v
per-client filtered snapshots/events
```

Offline/local play should remain one player using the same game authority.

## Current Dark Magic session is already network-shaped

`internal/game/session` has several useful invariants:

- one transport/renderer-independent authoritative simulation;
- untrusted command validation separated from trusted mutation;
- explicit player/system/admin authority classes;
- bounded command lead time;
- network arrival order does not define execution order;
- commands are canonically sorted within a tick by player, sequence and kind;
- fixed-step ECS update;
- non-ECS `StateParticipant` snapshots;
- replay command log and composite checkpoints/checksums;
- remote commands can enter through `Submit`; only local intent uses `CommandSource`.

This is a strong multiplayer simulation core. Build transport/session membership around it rather than rewriting gameplay around sockets.

## Player identity

Separate three concepts:

```text
AccountIdentity        // realm/account layer
CharacterIdentity      // durable character/save identity
SessionPlayerIdentity  // one joined player in one game
```

Current command `Player` strings are sufficient for local/testing but future remote authority should bind an authenticated connection/session principal to exactly the allowed session-player ID.

Never trust a network payload's arbitrary `Player` string.

## Party state

D2MOO 1.10f reconstructs explicit game-owned party structures and helpers for:

- allocate/free party control;
- create/join party IDs;
- leave party;
- iterate all party members;
- iterate same-level members;
- count living same-level party members;
- share gold drops;
- resolve a unit owner's party ID.

Party state therefore participates in gameplay, not merely UI roster grouping.

Recommended semantic model:

```text
PartyState
  PartyID
  MemberPlayerIDs ordered/stable

PartyInvite
  From
  To
  CreatedTick
  ExpiresTick/policy
```

Player -> PartyID can be derived/indexed.

## Party invite/join/leave

Pinned 1.10f `PartyScreen` shows explicit invite/cancel/join/leave relationship transitions and client notifications.

Dark Magic should expose semantic commands:

```text
party.invite
party.cancel_invite
party.accept
party.leave
```

Authority validates current relationship, player presence, hostility and any future policy before changing membership.

Do not let the party UI submit a party ID to join directly.

## Same-level party context matters

Original code uses same-level living party member iteration/counting for systems such as NoDrop and likely quest/XP behavior.

Keep helpers that answer semantic questions:

```text
LivingPartyMembersInLevel(player, level)
EligiblePartyMembersNear(event)
```

Do not infer eligibility only from party membership.

## Gold sharing

D2MOO has explicit party gold-sharing helpers. Exact range/level/cap behavior needs probing.

Gold pickup should be an item/economy transaction that can emit a `GoldPickup` result to a party distribution policy. Each player's carried/stashed limits and overflow behavior remain authoritative.

Do not have the client split a pile into displayed shares.

## Quest sharing

Quest research already shows player-specific difficulty quest bits plus live shared quest-controller state and party credit.

Multiplayer quest events should compute an explicit eligible-recipient set at event time from:

- party membership;
- level/area/proximity;
- alive/dead state;
- prerequisite quest state;
- event-specific rules.

Commit those per-player quest transitions deterministically.

Do not loop over current party membership later after players may have moved/left.

## XP sharing

Monster death needs a stable XP-credit transaction rather than each client computing experience.

Inputs can include:

```text
monster level/XP
killer + ultimate owner
party membership
same-level/proximity eligibility
member levels
player count/game rules
```

D2MOO damage/party code includes party XP helper structures and same-level party iteration, but exact XP share/scaling needs dedicated vectors.

Keep XP calculation in progression/combat reward authority, not party UI.

## Hardcore corpse-loot permission

Pinned 1.10f `PartyScreen` has explicit per-player lootability toggles that only operate for Hardcore clients. It stores a relationship flag between the two players and emits roster/event updates.

This is separate from ordinary party membership and from item ownership.

Semantic relation:

```text
CorpseLootPermission
  deceased/owner player
  permitted player
  enabled
```

Exact directionality, persistence and when the permission may be changed require probes.

## Ignore/squelch

The legacy party-screen code also stores per-player ignore/squelch-like relationship flags.

These are communication/presentation policy, not gameplay combat authority. Keep them at social/connection profile or game-session relationship level and filter chat/voice/events accordingly.

Do not let ignore change simulation targeting unless original gameplay explicitly says so.

## Hostility

Pinned 1.10f evidence from `PARTYSCREEN_ToggleHostile` is concrete:

- declarer must be in town;
- both players must be at least level 9;
- hostility is a server relationship flag/state;
- removing hostility uses a separate server operation;
- opening hostility is subject to a delay;
- the inspected code uses `GetTickCount() + 60000`, i.e. a 60-second legacy host-clock delay;
- hostile declaration against a party member can propagate against that target's party;
- a player declaring hostility against someone in the same party is removed from that party.

### Dark Magic timing policy

Preserve the intended cooldown duration/rule but represent future eligibility as authoritative session/game time, e.g. `HostileAvailableTick`, for replay determinism.

Legacy `GetTickCount()` is evidence of duration semantics, not a requirement to import nondeterministic wall clock into the new fixed-tick simulation.

## Hostility is directional versus mutual combat eligibility

The original player-list/friendly machinery has relationship state in both directions and operations that open mutual hostility in the inspected path.

Dark Magic should distinguish:

```text
DeclaredHostility(A -> B)
CombatHostile(A,B)
```

rather than assume one boolean on a player.

This allows version/mod policies and clean UI explanations.

## PvP target legality

Combat target validation should consult a faction/relationship service:

```text
CanDamage(attacker, defender, skill/source)
```

Inputs include:

- same player/self;
- party/alliance;
- hostility;
- town/safe-zone restrictions;
- level requirements;
- game mode;
- pet/hireling/trap ownership;
- skill-specific friendly/hostile policy.

Do not teach every skill handler how to inspect party flags.

## PvP damage

The typed DifficultyLevels data contains explicit player/mercenary/pet/boss damage scalars in later data revisions, while pinned combat code also has arena/PvP branches.

PvP damage belongs in the shared combat pipeline as context modifiers after legality resolution.

Exact target-version values/order/leech/CB/Open Wounds/poison rules remain probes.

## Town/safe-zone behavior

Town is authoritative level/room metadata. Hostility may be declared there in 1.10f, but combat/missile/skill behavior can remain restricted while in town.

Separate:

```text
relationship may change in town
combat action legality in town
```

Do not infer safe-zone from whether the town UI is displayed.

## Player trade is a two-party transaction state machine

D2MOO's `PlrTrade` exposes explicit functions for:

- requesting/starting trade;
- changing trade state/button actions;
- stopping all interactions;
- tracking items entering/leaving trade;
- trade save/check buffers;
- validation/commit helpers.

The exact structures are less important than the architectural rule: trade is **not two independent inventory moves**.

Recommended state:

```text
TradeSession
  ID
  PlayerA / PlayerB
  state
  offered item IDs by player
  offered gold by player
  acceptance flags/revisions
  created/last-change tick
```

## Trade escrow

Extend item authority with a trade escrow placement/capability rather than letting offered items remain independently movable.

Possible:

```text
ContainerTrade
  TradeID
  OwnerPlayerID
  Slot/grid
```

When an item/gold offer changes:

- clear both acceptance confirmations as required;
- validate ownership/locks;
- update authoritative offer revision.

## Atomic trade commit

Final accept must atomically validate:

- both participants still connected/eligible/in range if required;
- exact offer revision accepted by both;
- all item identities still in escrow;
- destination inventory capacity/requirements;
- gold balances/limits;
- quest/soulbound/nontradeable restrictions;
- no item is simultaneously socketed/held/service-locked/corpse-locked;
- resulting ownership/placements are valid.

Then transfer all items/gold in one commit or none.

A crash/reconnect checkpoint must not expose one-sided trade completion.

## Trade item identity and persistence

Trade transfers the same stable item identity. It does not serialize->reroll->recreate an equivalent item.

Legacy protocol/save buffers may be used for validation/transmission, but Dark Magic semantic authority keeps the item archive object and updates owner/placement.

## Trade anti-scam/stale acceptance

The classic UX clears acceptance when an offer changes. Build this structurally using offer revision hashes:

```text
OfferRevision = canonical hash(items + gold + relevant visible facts)
PlayerAccepts(revision)
```

Any offer mutation invalidates prior acceptance automatically.

This is more robust than UI timing tricks and maps cleanly to networked play.

## Disconnect during trade

On disconnect/leave game:

- cancel trade;
- atomically return escrowed items/gold to their owners according to a deterministic overflow policy;
- checkpoint state;
- only then remove the session player.

Do not leave authoritative items stranded in a client-owned cursor/trade panel.

## Player death and PvP rewards

Player death can involve:

- corpse ownership;
- Hardcore corpse-loot permission;
- gold/XP loss;
- ear drop in some PvP contexts;
- hostility/party context;
- kill attribution.

Death transaction should emit semantic PvP/death results consumed by item/party/persistence systems.

D2MOO has an explicit player-ear item creation path. Preserve this as a target-version/game-mode rule, not a generic monster loot drop.

## Per-player visibility and private state

Not every authoritative fact belongs in every client's snapshot.

Examples of restricted information:

- hidden gamble outcome seed;
- other players' private inventory/stash;
- trade offer only when relevant;
- unrevealed map/automap data;
- private quest/NPC flags;
- account/session metadata.

The server should project per-client views from canonical state rather than duplicate simulations.

## Suggested implementation slices

### MP1 — multiplayer player registry

Bind authenticated/local principals to stable session-player IDs and expose join/leave lifecycle without changing simulation handlers.

### MP2 — party state

Implement invite/accept/leave plus deterministic party membership snapshots and same-level member queries.

### MP3 — shared reward hook

Use party state for one real policy such as NoDrop count or synthetic XP/gold sharing.

Implemented for monster-death NoDrop using ultimate owned-unit attribution and
living same-level party membership. Exact 1.14d proximity remains queued.

### MP4 — hostility relationship

Implement level/town validation, deterministic 60-second-equivalent cooldown, party removal/propagation and combat target legality.

### MP5 — trade escrow state

Add two-player trade session and stable item/gold escrow with offer revisions.

### MP6 — atomic commit/disconnect

Implement dual acceptance, capacity validation, one-transaction exchange and safe cancellation on disconnect.

### MP7 — Hardcore corpse-loot permission

Add the semantic relationship and integrate with death/corpse item access after target-version probes.

## Verification backlog

1. Party invite/cancel/accept relationship-state semantics and edge cases.
2. Party size limits and ID lifecycle.
3. Same-level/proximity rules for XP, quests, gold and NoDrop.
4. Exact party XP formula/player level weighting.
5. Gold sharing radius/level/overflow behavior.
6. Party quest credit by representative quests.
7. Hardcore corpse-loot permission directionality/change/persistence.
8. Ignore/squelch/chat filtering behavior.
9. Hostility level/town restrictions across patches.
10. Hostility cooldown duration/reset and wall-clock versus game-time semantics.
11. Hostility propagation against party members and party ejection.
12. Removing hostility conditions/delay.
13. Town PvP/skill/missile restrictions.
14. PvP damage/leech/CB/Open Wounds/poison/hit-recovery rules.
15. Pet/hireling/trap hostility and kill attribution.
16. PvP death gold/XP/corpse/ear behavior.
17. Trade initiation range/town/restrictions.
18. Trade item placement/grid/offer rules.
19. Acceptance reset delay/revision behavior.
20. Trade gold limits and destination capacity failure.
21. Quest/nontradeable item rules.
22. Disconnect/death/hostility during trade.
23. Private state visibility in network snapshots.

## Primary sources inspected

- Current Dark Magic fixed-tick session/replay, item authority, interaction/world/combat/game-rules research.
- D2MOO pinned 1.10f `UNIT/Party.*`, `PLAYER/PartyScreen.cpp`, `PLAYER/PlrTrade.*`, Friendly/PlayerList and combat/item/death call sites.
