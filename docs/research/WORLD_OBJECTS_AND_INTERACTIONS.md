# World objects and environmental interaction research

Status: implementation-oriented research baseline. Dark Magic already resolves authored DS1/object semantics, has semantic world selection/interaction authority, generated/preset map collision, fixed-tick commands, item service escrow, transition authority, and renderer-independent presentation state. This document defines the missing runtime object state/operation layer.

## Executive conclusion

A Diablo II world object is an authoritative stateful unit with a **behavior family**, not merely a map decoration or click hotspot.

Conceptually:

```text
ObjectDefinition
  Objects.txt identity
  modes/animation
  operate/event function IDs
  collision/selectability
  sounds
  trap/loot/quest/transition metadata
        |
        v
ObjectInstance
  stable entity ID
  level/room/world position
  current mode/state
  seed/RNG where required
  cooldown/used/locked flags
  quest/owner/runtime parameters
        |
        v
semantic interaction request
        |
        v
ObjectOperation transaction
  validate range/LOS/state/quest/items
  dispatch behavior family
  mutate object/world/item/player state
  schedule follow-up events
  emit semantic presentation/audio events
```

The object animation never decides whether the chest opened, the door became passable, or the waypoint activated.

## Current Dark Magic foundation

Dark Magic already has important pieces:

- typed `Objects.txt` / object-function metadata;
- recovered DS1 object resolution under `internal/game/data/worldobjects`;
- semantic map selectables;
- `internal/game/interaction` authoritative target context with current-world filtering, range revalidation and LOS-aware world selection;
- `internal/game/transition` trusted zone movement;
- `internal/game/item` stable world/container item identity and quest/service escrow;
- quest runtime/controller research;
- world collision and diagnostics;
- presentation/audio layers separated from authority.

The next object work should extend these boundaries rather than add a Lua-owned object game state.

## Original behavior-family evidence

D2MOO's pinned 1.10f `ObjMode` reconstructs a large operation dispatch surface. Named handlers include, among others:

- doors and secret/slime doors;
- trap doors and stairs;
- teleport pads;
- Harrogath gate;
- chests and exploding chests;
- multiple trap families;
- shrines and obelisks;
- caskets, urns, baskets, jars and corpses;
- Evil Urns;
- Wirt's body;
- jungle stash;
- bookshelves;
- armor stands and weapon racks;
- barrels/exploding barrels;
- torches;
- waypoints/portals and quest objects elsewhere in the same object system.

This is strong evidence for **data-selected behavior families**. Dark Magic should normalize numeric operation/event IDs to stable semantic handlers rather than write `switch objectID` throughout gameplay.

## Object definition versus object instance

Keep immutable content separate from runtime state.

Possible normalized definition:

```text
ObjectDefinition
  class ID / token
  operation behavior ID
  event behavior ID
  allowed modes
  collision per mode
  selectable/target flags
  interaction radius
  frame/event timing metadata
  sound roles
  lighting/overlay/presentation metadata
```

Runtime instance:

```text
ObjectState
  EntityID
  DefinitionID
  LevelID / room identity
  Position
  Mode
  Used/Locked/Disabled flags
  Seed
  ScheduledEvents
  BehaviorParams
```

Do not put renderer handles or decoded DC6/DCC frames in authoritative object state.

## Mode/state changes

Many objects have stateful mode transitions such as:

```text
closed -> operating/opening -> open
closed door -> open door
armed trap -> triggered -> spent
unused shrine -> used
inactive waypoint -> activated/interacting
intact barrel -> operating -> destroyed
```

A mode change may affect:

- collision;
- selection/interaction eligibility;
- animation;
- sound;
- lighting;
- scheduled events;
- loot/trap spawn;
- quest state;
- level transition availability.

Authority commits the state transition first. Presentation observes it.

## Object events and scheduled behavior

The object subsystem includes explicit event-function invocation in D2MOO. Some objects do work after a delay or through repeated events rather than entirely inside the click operation.

Dark Magic should use the fixed-tick scheduler/session model:

```text
ObjectScheduledEvent
  object ID
  behavior/event kind
  due tick
  parameters
```

This state must survive replay/checkpoint and room streaming if it can affect future gameplay.

## Doors and collision

Door state is authoritative collision state.

Recommended flow:

```text
interact closed door
 -> validate actor/object
 -> commit object mode = opening/open
 -> update semantic collision footprint at defined authoritative tick
 -> movement paths may replan
 -> emit animation/audio cue
```

Do not toggle collision when an animation visually reaches a frame unless that frame has been converted into an authoritative scheduled event by the object behavior model.

Monster `OpenDoors`/AI should use the same semantic operate action where possible.

## Chests and containers

A chest operation can involve:

- eligibility/used state;
- lock/key policy;
- deterministic trap roll;
- TreasureClass/loot generation;
- item ground placement;
- quest/object special cases;
- sound/animation;
- one-time or refill behavior.

Use a transactional result:

```text
ObjectOperationResult
  ObjectMutations
  GeneratedLootRequests
  Trap/Monster/MissileRequests
  Player/StateMutations
  QuestEvents
  TransitionRequests
  Audio/PresentationEvents
```

Loot remains owned by `internal/game/loot`; the object handler supplies source/context/seed.

## Object traps

D2MOO exposes several trap-handler families and explicit monster-spawning trap behavior.

Object traps should reuse:

- missile system;
- monster spawn system;
- combat/damage;
- states;
- audio;

rather than implementing direct player HP subtraction inside object code.

The object behavior decides **what trap action to request** and owns trap-used state.

## Racks and stands

Weapon racks/armor stands use special item generation rules distinct from ordinary monster TreasureClasses.

They should produce a generation context such as:

```text
ItemGenerationSource = WeaponRack | ArmorStand
source object seed
area/item level context
```

Do not route them through monster TC just because both ultimately create world items.

## Destructible objects

Barrels and similar objects can have:

- health/destruction or direct operate-to-break semantics;
- explosion/trap effects;
- loot chance;
- collision removal;
- sound/animation.

The engine should support object damage as a semantic pathway if the target runtime allows attacks against the object. Where an object is only operated, keep that distinction.

## Quest objects

Quest objects need a generic hook surface, not hard-coded quest imports inside presentation.

The object transaction can query/emit:

```text
QuestPredicate
QuestEvent
QuestItemRequirement/Consumption
```

Examples include altars, stones, seals, gates, quest chests, special corpses and act set pieces.

Quest runtime remains the owner of quest progression bits/controllers.

## Transition objects

Stairs, trap doors, gates, pads, portals and waypoints may request level/zone transitions.

The object handler should not directly rewrite player position. It returns a semantic transition request to `internal/game/transition`, which validates destination/arrival/quest context and commits the move.

This preserves one trusted transition boundary.

## Object collision footprint and selection footprint

Keep separate:

- collision footprint;
- interaction radius;
- selection footprint/radius;
- visual bounds.

Dark Magic's current interaction authority already distinguishes range/select radius. Extend it to arbitrary object targets rather than assuming every target is an NPC/vendor.

## Interaction target model extension

Current `interaction.Target` requires an `NPC` field and exposes vendor/categories/services. General world objects need a broader semantic target type.

Recommended direction:

```text
InteractionTarget
  ID
  Kind = NPC | Object | Waypoint | Portal | Shrine | Item | ...
  Definition/semantic ID
  Position
  SelectRadius
  OperateRadius
  Capabilities
```

NPC-specific vendor/service context can remain a capability attached to NPC targets.

Do not make object interaction masquerade as a fake NPC.

## Room streaming

Object instance state affecting future gameplay must survive room inactivation:

- current mode;
- used/opened/destroyed state;
- seed/loot/trap state;
- scheduled events;
- quest-related flags;
- collision mode.

Presentation residency is independent.

## Audio integration

Objects can emit:

- ambient loops;
- operate/open sounds;
- trap sounds;
- destruction sounds;
- portal/waypoint loops.

Object state transitions emit semantic audio cue roles to [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md). Audio does not decide object state.

## Suggested implementation slices

### WO1 — generic authoritative object instance

Materialize one synthetic object with stable ID, mode, collision/selectability and interaction behavior independent of Lua/presentation.

### WO2 — door vertical slice

Operate a door, update collision authoritatively, replan movement, and show animation/audio from state.

### WO3 — chest vertical slice

Operate one chest using object seed -> loot request -> stable world item identities -> used state.

### WO4 — trap object

Trigger one object trap through missile/monster/combat systems.

### WO5 — quest/transition object

Implement one object whose operation produces a quest event and/or trusted zone transition request.

### WO6 — inactive room persistence

Deactivate/restore used/open object state and scheduled object events deterministically.

## Verification backlog

1. Objects.txt operate/event function mapping and target-version differences.
2. Object mode transition timing and collision change point.
3. Door lock/key/monster-open behavior.
4. Chest seed derivation, TC/item level, lock/key and used-state behavior.
5. Chest/object trap probability/type selection and RNG stream.
6. Rack/stand item-generation rules.
7. Barrel/destructible object damage/loot/explosion behavior.
8. Object reset/refill behavior and room reload.
9. Secret doors/slime doors/gates special conditions.
10. Teleport pads/stairs/trap doors destination resolution.
11. Quest object predicate/event ordering.
12. Object operation party credit/proximity semantics.
13. Interaction radius/LOS and path-to-operate stopping behavior.
14. Object collision footprints by mode.
15. Object scheduled events across room inactivation.
16. Ambient/operate sound mapping and loop lifetime.
17. Multiplayer simultaneous object operation conflict ordering.
18. Object state replication/snapshot fields.

## Primary sources inspected

- Current Dark Magic recovered world-object resolver, typed Objects/object-function data, interaction authority, transition/item/quest/world foundations.
- D2MOO pinned 1.10f `OBJECTS/ObjMode.h/.cpp` behavior families and related object/quest/item/missile/monster paths.
