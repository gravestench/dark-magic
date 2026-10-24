# Warps, portals, waypoints, teleportation, and level-transition research

Status: implementation-oriented research baseline. Dark Magic now prepares Rogue
Encampment, Blood Moor, Cold Plains, and Stony Field together and crosses their
three seam pairs through an authoritative Lua collision-phase system. The Den of
Evil is a separate indoor map reached through an intentional reciprocal cave
interaction. The generated topology and per-seam trigger policy are pinned by a
runtime checksum; original
1.13c/1.14d seam coordinates remain unverified. Warp Lab also materializes a
paired authoritative interaction/transition entity set with stable level/room
residency from generated geometry. This document generalizes those boundaries
to the full Diablo II transition vocabulary.

## Executive conclusion

All cross-level movement should converge on one trusted transition transaction:

```text
TransitionRequest
  actor
  source kind/entity
  source level/position
  destination level/endpoint policy
  quest/waypoint/portal context
        |
        v
transition authority validates
  source state/range
  destination eligibility
  quest/difficulty/game-mode gates
  arrival position/collision
  owned-unit policy
        |
        v
atomic world move
  level/room/position
  velocity/path reset
  active-room/world residency
  owned-unit follow policy
  interaction context reset
        |
        v
semantic transition event
  presentation/loading/audio respond
```

Stairs, level seams, doors, red/blue portals, town portals, waypoints, teleport pads and quest gates differ in **how they build/validate the request**, not in who owns final player relocation.

## Current Dark Magic transition authority

The generic `internal/game/transition` command path and the Lua collision-phase
Act I seam path both keep final relocation in the authority:

- the generic command path requires system/admin authority rather than
  presentation authority; Act I seams run directly in the authoritative Lua
  tick without submitting a transition command;
- prepared Act I seam destinations are validated against materialized map
  bounds and collision;
- the player must be within the admitted seam's source radius;
- level ID and world position update atomically in ECS;
- velocity is reset;
- presentation/cache observer fires only after authoritative commit and cannot veto;
- Lua's collision-phase system applies each prepared Act I seam after movement enters its source radius;
- connected movement samples stamped with their source level are ignored after an authoritative seam crossing.

The `1↔2`, `2↔3`, and `3↔4` seamless joins use prepared endpoint data. The
`2↔8` cave pair uses the existing interaction admission and world-warp operation,
then commits through `d2legacy.gameplay.transition.apply` with kind `warp`.
Other level links still need data/instance-driven transition coverage.

## Den of Evil evidence and implementation boundary (#656)

Owned Expansion 1.14d data supplies the following authored facts. These are data
observations, not observations of the original executable's generation algorithm.
No protected data or rendered captures are stored with the tests.

| Evidence | Authored fact | Executable coverage |
| --- | --- | --- |
| `patch_d2.mpq` `Levels.txt`, levels 2 and 8 | Blood Moor `Vis3=8/Warp3=0`; Den `Vis0=2/Warp0=4`; Den is an indoor cave maze | `TestDenOfEvilPreparedReal` and the reciprocal resolver tests in `internal/mod/d2legacy/adapter/entryworld` |
| `LvlMaze.txt` level 8 | One 24×24 room in Normal, Nightmare, and Hell | `TestDenOfEvilOneRoomReal` in `internal/mod/d2legacy/adapter/mapgen` |
| `LvlPrest.txt` and decoded DS1 special tiles | Outdoor Cave Entrance definition 51; previous-level cave definitions 83–86. Full marker identities are orientation/MainIndex/SubIndex `10/6/24` outside and `10/0/21` inside | Owned map materialization and `cave_warp_test.go` reject missing, duplicate, wrong-identity, and blocked endpoints |
| Base `d2exp.mpq` `LvlWarp.txt` IDs 0 and 4 | Wilderness-to-cave and cave-up relationships, offset and exit-walk fields | Prepared endpoints require bounded radius-1 source/arrival routes and a reachable cave interior |

Lua's v4 entry recipe selects the authored chamber and Blood Moor stamp with
separate named random streams. It retains levels 1/2/3/4 and their three seams,
adds level 8, and records one reciprocal cave relationship. Go validates the
materialized marker identities and collision before publishing either endpoint.
Current-binary v4 restore validates those same facts; v1/v2/v3 restoration retains
its original map set. The static endpoints have selectable identities but no
portal appearance token: cave art comes from the authored map.

`internal/app/clientapp/den_of_evil_transition_real_test.go` exercises owned-data
production clients at all three difficulties. The round-trip test walks onto
both cave endpoints without relocation, uses the ordinary `game_world` pointer
and selectable path for entry and return, and checks authoritative/public levels,
safe arrival, `warp` relocation facts, route reset, and camera position. It
observes the live Lua scene's separate `{8}` indoor and `{1,2,3,4}` outdoor
components, and retained renderer handles confirm old terrain is destroyed on
each cave warp. Both offline and listen-server journeys retain map and resident
identities and all three scene seam alignments; the offline journey additionally
walks all three joins in both directions after return, verifies `seam` facts,
keeps the outdoor terrain root, and requires terrain nodes from both sides.
The connected journey stops its old correction stream before credential recovery,
then checks the unchanged Den authority and completes the clicked return.

The checkpoint test closes the transient scene through its normal lifecycle
while gameplay remains in level 8, captures that authority, and requires an exact
recovery checksum before opening a fresh scene. It then checks the saved Den
arrival and ordinary clicked return, including outdoor resident IDs. Capturing
and rebinding a *live* presentation camera is not covered: the initial experiment
restored the saved camera and created a second one on scene entry. That generic
scene/checkpoint ownership issue is tracked in
[#657](https://github.com/gravestench/dark-magic/issues/657).

Original-game cave placement, chamber/variant RNG call order, authored selection
rectangle transforms, exact arrival/automatic-walk behavior, and 1.13c equivalence
remain unresolved. The deterministic layout, radius-3.5 selectable, bounded
radius-1 arrival adjustment, and reachable central-half interior are Dark Magic
policy. They establish a usable cave journey, not retail seed parity. Native GPU
pixels and cave/bridge visual fidelity remain unverified by the headless scene
assertions. Den population, quests, rewards, music, other caves, and general
waypoint/portal behavior remain outside #656.

## Transition source kinds

Normalize source semantics:

```text
LevelSeam
Stair
TrapDoor
Door/Gate
TeleportPad
Waypoint
TownPortal
QuestPortal
RedPortal
CubePortal
ScriptedEncounterExit
SkillTeleport       // same-level relocation policy, not necessarily level transition
```

A source can be a map seam or a world object/entity.

## Static warps from map data

Map-generation research already identifies DS1/DT1 logical tiles, exits, links and warps as semantic data.

For a generated/preset zone, materialization should produce `TransitionEndpoint` facts containing:

- source level/region/coordinates;
- destination level identity;
- matching destination endpoint/link identity;
- source/destination arrival offsets;
- selection/trigger geometry;
- direction/type metadata;
- quest/gate metadata where applicable.

Do not rediscover level links from rendered exit sprites.

## LvlWarp data

D2MOO's `LvlWarp` representation contains fields for selection dimensions, exit-walk coordinates, offsets, lit version and direction/tile data. These are evidence that visual/selection/arrival geometry are distinct authored concepts.

Dark Magic should preserve normalized fields such as:

```text
SelectBounds
ArrivalOffset / ExitWalk
Direction
VisualVariantMetadata
```

Exact coordinate interpretation must remain tied to the map/coordinate research and verified with owned map fixtures.

### Verified warp coordinate domains

Riiablo's server map construction supplies a special cell at
`tileCoordinate * 5` subtiles, then adds `LvlWarp.OffsetX/Y` to create the warp
entity position. Arrival initially uses the destination warp entity position;
`ExitWalkX/Y` is added to that position to form the automatic post-arrival walk
target. These values therefore belong to authoritative world-subtile space.

`SelectX/Y/DX/DY` is different. Riiablo copies it directly into a client-side
local bounding box. A production Act I cave fixture resolves values
`(-30, -120, 120, 150)`, disproving any interpretation as world subtiles. Dark
Magic retains this rectangle in authored client-local units until the
presentation adapter's screen-space transform is verified. It must never enter
collision, pathfinding, or authoritative movement arithmetic.

## Stairs, trap doors, pads and gates

D2MOO object behavior has distinct operation handlers for stairs, trap doors, teleport pads, gates and specialized doors.

These handlers should construct transition requests through the generic object-operation layer:

```text
operate object
 -> object validates mode/quest state
 -> TransitionRequest(source object, destination semantic link)
 -> transition authority commits player move
 -> object/quest state updates if required
```

Some transitions require a click; ordinary open level seams may trigger by crossing/standing near an endpoint.

## Waypoints are durable player unlock state

Waypoint discovery/activation is distinct from traveling through a waypoint.

A player needs durable per-difficulty waypoint state:

```text
WaypointState
  difficulty
  unlocked waypoint IDs/levels
```

The legacy `.d2s` research already identifies waypoint bit blocks. The semantic persistent model should store unlocked waypoint identities, while the save codec maps to/from legacy bits.

Do not make waypoint unlock global to the account or transient to one game unless a mod explicitly chooses that rule.

## Waypoint activation

Operating an authored waypoint object should:

1. validate player/object range and object state;
2. resolve the waypoint identity from authored level/object data;
3. mark the waypoint unlocked for the current difficulty if newly discovered;
4. emit a semantic activation event;
5. open/present waypoint destination UI as a client-facing snapshot of server-known reachable waypoints.

Presentation does not grant the bit merely by opening the panel.

## Waypoint travel

Client submits a stable destination waypoint ID/level. Authority validates:

- source interaction is an active waypoint;
- destination is unlocked or otherwise permitted;
- same game/difficulty/act conditions;
- destination exists in the current content generation;
- arrival coordinate is legal.

Then it performs the same trusted world transition transaction as stairs/seams.

## Waypoint destination list

The UI-visible list should be derived from:

```text
current difficulty
persistent unlocked waypoint set
level/act data
quest/game-mode restrictions
current content generation
```

Keep list ordering authored/stable, not Go-map order.

## Town portals

Town Portal is a paired dynamic transition with ownership/lifecycle.

Potential state:

```text
PortalPair
  stable portal ID
  owner/player or party context
  origin level/position
  town destination level/position
  return endpoint
  open/closed state
  usage policy
  created tick/source skill/item
```

The portal visual/object exists in world state; its two endpoints route through transition authority.

Research needs to pin whether/when the origin portal disappears, who may use it, and how a return transition moves/invalidates the pair.

## Portal ownership and party use

Do not assume every portal is private or public.

Rules can depend on:

- player owner;
- party;
- hostility;
- game mode;
- quest/event type;
- level restrictions.

The portal instance should carry an explicit access policy. The UI/render color is not authority.

## Red/quest/Cube portals

Special portals are usually one-way or destination-specific and may use quest/Cube progression.

Cube portal operations and quest object handlers should request portal creation through a world/transition API:

```text
CreatePortalRequest
  source event/actor
  origin level/position
  destination level
  return/access/lifetime policy
```

The result is an authoritative world object/portal pair, not a direct scene switch.

## Cow Level and special destinations

The Cube `COWPORTAL` operation is explicit evidence that special areas enter through the Cube/world transition stack.

Eligibility belongs in Cube/quest/game-mode policy, while portal instance creation and use belong in world/transition.

Later patch features such as Uber portal operations must be version-gated and are covered in `ENDGAME_AND_SPECIAL_EVENTS.md`.

## Teleport skill versus level transition

Sorceress/monster teleport is normally same-level relocation, not a level transition. It should reuse coordinate/destination validation primitives but not necessarily tear down/rebuild zone residency.

The implemented exact-ID Sorceress Teleport slice now keeps the same durable ECS
entity and level/location component, atomically changes position after static
and dynamic destination checks, stops competing motion, and emits a generic
relocation fact. This is deliberately separate from the level-transition
transaction. Owned Levels.txt supplies policies 0/1/2; the current policy-2
BlockLOS interpretation remains conservative pending target-runtime vectors,
alongside exact visibility range, invalid-target mana/fallback, room-edge
ordering, owned-unit following, and presentation timing.

Separate:

```text
RelocateWithinLevel
TransitionBetweenLevels
```

Both are authoritative movement transactions; only the latter changes level/room graph/world residency.

## Arrival placement

Destination entry is more than a level ID.

Authority should choose an arrival point using:

- authored endpoint/warp destination;
- source direction/exit-walk metadata;
- unit collider/footprint;
- static and dynamic collision;
- deterministic nearby-free-coordinate fallback;
- player/owned-unit spacing.

The client must not submit arbitrary arrival coordinates.

## Owned units during transition

Hirelings/summons need per-category transition policy from the owned-unit research.

The transition transaction should produce a follow-up plan:

```text
FollowAndPlaceNearOwner
DeactivateInOldZone
DestroyOnTransition
PersistAndRestore
```

Do not simply copy their old coordinates into the new level.

## Interaction and movement reset

After cross-level transition:

- clear/validate current NPC/object interaction context;
- cancel ordinary path target;
- reset velocity/forced movement as appropriate;
- retarget camera/presentation world after commit;
- refresh active zone collision/selection index;
- reconstruct zone soundscape/music.

The current transition observer already models the post-commit presentation boundary.

## Room/world streaming

Transition changes the active player-centered room/world set. Server world state decides what activates/inactivates; presentation loading is an observer.

Do not delay authoritative arrival until every texture/sound is loaded. The client can use loading presentation while authoritative session state is already transitioning under a controlled protocol.

For networked play, server may gate input until the client acknowledges readiness, but the server still owns the destination.

## Persistence

Persist durable facts such as:

- current character progression/difficulty;
- waypoint unlocks;
- quest bits.

Do **not** persist temporary game portal instances into a legacy character save unless the target behavior explicitly does so. Realm game checkpoints may retain them for reconnect if desired.

## Audio/presentation

Transition events should drive:

- loading screen/scene changes;
- music/soundscape crossfade;
- portal/waypoint operate sounds;
- area-entry cues.

See [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md). Audio/load completion never chooses destination eligibility.

## Suggested implementation slices

### T1 — generalize transition endpoints

The three prepared Act I seams establish this boundary for Rogue Encampment,
Blood Moor, Cold Plains, and Stony Field. Extend endpoint/link definitions to other
levels while preserving the existing seam tests.

### T2 — authored stair/warp

Materialize one preset stair/exit object and route it through generic transition authority.

The authored Blood Moor/Den pair now uses that production path through
ordinary pointer interaction. Synthetic Warp Lab endpoints retain portal
appearance and generic room residency. Other stairs, trap doors, and pads still
need authored evidence and acceptance of their own operation rules.

### T3 — waypoint unlock/travel

Add per-difficulty durable waypoint state, one activation object and trusted waypoint destination command.

### T4 — dynamic portal pair

Implement a synthetic Town Portal pair with owner/access/lifetime and return endpoint.

### T5 — quest/Cube portal

Create one special portal from a semantic Cube/quest event without direct presentation scene changes.

### T6 — owned-unit/network lifecycle

Transition player + hireling/summon policies and expose clean snapshot/correction boundaries.

## Verification backlog

1. LvlWarp selection/offset/exit-walk coordinate semantics.
2. DS1 logical exit/warp link resolution and paired endpoint selection.
3. Stair/trapdoor/pad operation timing/range.
4. Arrival/free-coordinate fallback rules.
5. Waypoint object identity -> waypoint bit/index mapping.
6. Waypoint activation and per-difficulty save bits.
7. Waypoint list ordering and destination eligibility.
8. Same-act/cross-act waypoint behavior and quest restrictions.
9. Town Portal origin/town destination/return placement.
10. Town Portal ownership, party access, hostile-player access and lifetime.
11. Portal close/replace behavior when casting another portal.
12. Portal persistence across owner death/disconnect/level travel.
13. Red/quest portal one-way/return/access semantics.
14. Cow Portal eligibility and destination lifecycle.
15. Later Uber portal version-gated semantics.
16. Teleport destination validation and no-teleport level flags.
17. Owned-unit follow/placement across transitions.
18. Interaction/path/skill state reset across levels.
19. Active-room streaming and inactive-unit restoration at arrival.
20. Audio/music/loading transition ordering.
21. Multiplayer server/client loading acknowledgment and stale transition commands.

## Primary sources inspected

- Current Dark Magic `internal/game/transition`, world seam/mapgen/selection, player/item/interaction/persistence systems.
- D2MOO pinned 1.10f object stair/pad/door/waypoint/portal paths, `LvlWarp`, player waypoint/save data, Cube portal operations and quest call sites.
