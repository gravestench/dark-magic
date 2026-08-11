# Quest runtime model research

Status: architecture notes derived primarily from D2MOO 1.10f. This is not yet
a Dark Magic API contract.

## Why a finite-state enum is insufficient

D2MOO exposes 41 saved quest-status slots. Visible quests, invisible prologues,
act-completion records, unused gossip controllers, and expansion quests share
that storage domain. Each quest slot is a bit record, while a live game also
holds quest-controller state and quest-specific extension data.

A useful conceptual split is:

```text
character + difficulty
  -> durable quest bits
       reward granted, reward pending, primary goal done,
       completed before/now, started, left town, entered area,
       quest-specific custom bits

game session
  -> global quest flags
  -> live controller state / last state
  -> quest-specific counters, GUID lists, object references, timers

generated world
  -> object modes, portals, room/preset availability, monster encounters
```

The same player-facing stage can correspond to several valid bit combinations.
Conversely, `primary goal done` and `reward granted` are deliberately separate.
Saving only an enum such as `not_started / active / completed` would lose
unclaimed rewards, permanent reward consumption, party eligibility, act travel,
and quest-item recovery behavior.

## Event surface recovered from source

Individual controllers subscribe only to relevant events. Across the quest
corpus, the important inputs are:

- NPC activate/deactivate;
- scroll/dialogue message acknowledged;
- player changed level;
- item picked up or dropped;
- player dropped from a game while holding a quest item;
- monster killed;
- player started or joined a game;
- player left a game;
- object initialization, operation, and mode change through quest-specific
  hooks;
- scheduled/timed callbacks for encounter and presentation sequencing.

The dialogue acknowledgement event is behavioral: specific message IDs grant
rewards, advance controller state, open services, or transition an act. It is
not merely UI text playback.

## Proposed implementation boundary

```text
QuestDefinition (data)
  id, act/order, visible, title/stage keys, declared prerequisite
  objective/reward descriptors

QuestRecord (durable, per character+difficulty)
  named common flags
  quest-specific durable flags
  permanent-reward consumption
  act travel and introduction records

QuestController (authoritative game session)
  event subscriptions
  live global phase and counters
  eligibility/party-credit policy
  references to spawned encounters and world objects

WorldQuestPort
  stable level/object/monster/item identifiers
  operate/set-mode/spawn/despawn/create-portal/unlock-transition commands

QuestPresentation (client)
  derive quest-log stage from record + controller snapshot
  resolve NPC message selection
  play localized speech and update log markers
```

Quest scripts should emit commands through authoritative interfaces; they
should not mutate rendering nodes or generated room data directly. Object and
monster systems should report stable semantic events rather than require quest
code to scan the ECS.

## Minimum event envelope

Every event that can award progress should include:

- game/session identifier and difficulty;
- initiating player and party snapshot;
- target stable ID and runtime GUID;
- source and destination level IDs for travel;
- world position/room only as supporting context;
- quest-item owner/container and item code;
- cause (direct interaction, kill credit, party credit, scripted consequence);
- monotonic sequence number for idempotency.

Quest mutation should be idempotent because object callbacks, reconnects, party
credit, and reward dialogue can converge on the same durable flag.

## Implementation order

1. Add named common quest flags and opaque custom-bit preservation.
2. Persist records per difficulty, including act transition/introduction slots.
3. Build the event dispatcher and deterministic controller lifecycle.
4. Implement one representative quest from each shape:
   Den of Evil (population), Cain (ordered objects/portal), Horadric Staff
   (multi-item recipe/world gate), Golden Bird (opportunistic drop/NPC exchange),
   Terror's End (multi-object encounter), Ancients (resettable gated encounter).
5. Add multiplayer join/leave and party-credit traces before expanding to all
   27 visible quests.
6. Derive quest-log and NPC presentation last; never use displayed stage text
   as the authority for gameplay state.

## Source map

- Common IDs and state slots: [D2MOO `Quests.h`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/include/QUESTS/Quests.h).
- Generic allocation, dispatch, status cycling, global state, party iteration,
  and log updates: [D2MOO `Quests.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS/Quests.cpp).
- Quest record bit storage: [D2MOO `D2QuestRecord.h`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Common/include/D2QuestRecord.h).
- Concrete callback/event combinations: [D2MOO quest source tree](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS).
- Independent save-layout check: [nokka/d2s](https://github.com/nokka/d2s#quests).

