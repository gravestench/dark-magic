# Warps, portals, waypoints, teleportation, and level-transition research

Status: implementation-oriented research baseline. Dark Magic already has a trusted `internal/game/transition` authority for the Act I town/Blood Moor seam, semantic map seams/warps, fixed-tick commands, persistent player location, and generated world boundaries. Warp Lab materializes a paired authoritative interaction/transition entity set; both endpoints now acquire stable level/room residency from generated geometry and are pinned by the mounted-asset lab. This document generalizes that boundary to the full Diablo II transition vocabulary.

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

The existing package already establishes several correct principles:

- command kind is system/admin-authority rather than presentation-authority;
- destination is validated server-side;
- player must be near the authored source endpoint;
- level ID and world position update atomically in ECS;
- velocity is reset;
- presentation/cache observer fires only after authoritative commit and cannot veto;
- the source generates a system transition when the player crosses the current Act I seam.

Keep these properties while replacing the current hard-coded `1 <-> 2` endpoint limitation with data/instance-driven transitions.

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

Replace the hard-coded Act I seam endpoint switch with immutable endpoint/link definitions while preserving current town/Blood Moor behavior/tests.

### T2 — authored stair/warp

Materialize one preset stair/exit object and route it through generic transition authority.

Current foundation: synthetic paired Warp Lab endpoints already use that
production path and the generic room-resident component. Room inactivation does
not need a warp-specific archive or recreated entity identity.

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
