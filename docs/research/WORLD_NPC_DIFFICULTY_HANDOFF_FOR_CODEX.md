# World-object, NPC, difficulty, transition, and endgame research handoff for Codex

Status: implementation handoff for the fourth gameplay research tranche.

## Read first

1. [WORLD_OBJECTS_AND_INTERACTIONS.md](WORLD_OBJECTS_AND_INTERACTIONS.md)
2. [SHRINES_WELLS_AND_TEMPORARY_EFFECTS.md](SHRINES_WELLS_AND_TEMPORARY_EFFECTS.md)
3. [WARPS_PORTALS_WAYPOINTS.md](WARPS_PORTALS_WAYPOINTS.md)
4. [NPCS_DIALOGUE_AND_SERVICES.md](NPCS_DIALOGUE_AND_SERVICES.md)
5. [DIFFICULTY_AND_GAME_MODES.md](DIFFICULTY_AND_GAME_MODES.md)
6. [ENDGAME_AND_SPECIAL_EVENTS.md](ENDGAME_AND_SPECIAL_EVENTS.md)
7. [QUEST_RUNTIME_MODEL.md](QUEST_RUNTIME_MODEL.md)
8. [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md)

## Existing owners to preserve

- `internal/game/world` remains map/collision/semantic world state.
- `internal/game/data/worldobjects` remains recovered/authored object resolution.
- `internal/game/interaction` remains trusted player-to-target context and range revalidation.
- `internal/game/transition` remains the only authority that commits cross-level player movement.
- `internal/game/item` remains item/service/quest escrow authority.
- quest runtime remains quest progression authority.
- `internal/game/session` remains fixed-tick/replay/checkpoint authority.
- typed game-data catalogs remain immutable definitions.
- presentation/audio observe semantic state/events after commit.

## Highest-leverage implementation order

### 1. General interaction target kind

Generalize the current NPC-centric interaction target to support semantic kinds/capabilities:

```text
NPC
Object
Door
Chest
Shrine/Well
Waypoint
Portal
TransitionObject
QuestObject
```

Preserve current selector/range/LOS revalidation and vendor/service behavior.

### 2. Authoritative ObjectInstance state

Materialize stable world object identities with definition ID, level/position, mode, collision/selectability, seed and scheduled state.

Do not add rendering handles.

### 3. Door vertical slice

Operate a door through interaction authority, transition authoritative mode/collision, let movement replan, then drive animation/audio from object state.

### 4. Chest vertical slice

Operate one deterministic chest:

```text
object seed/state
 -> loot generation context
 -> stable world item identity
 -> chest used mode
```

### 5. Generic temporary effect source

Use the shared state/stat-source system for one shrine buff and one well/resource cleanup behavior.

### 6. General transition endpoint model

Replace the hard-coded Act I 1<->2 transition switch with immutable semantic endpoints/links while preserving all current seam behavior/tests.

### 7. Stair/transition object

Make one authored stair/trapdoor object request a transition through the same authority.

### 8. Per-difficulty waypoint state

Add semantic waypoint unlocks to durable character state, map legacy save bits later, and implement one waypoint activation/travel command.

### 9. Dynamic portal pair

Implement a synthetic Town Portal with stable object identity, owner/access policy, origin/town endpoints and return behavior.

### 10. NPC capability/dialogue snapshot

Server derives ordered Talk/Trade/Gamble/Repair/Identify/Heal/Hire/Resurrect/Quest choices from NPC + player + quest/game state.

### 11. NPC quest dialogue event

Implement one real/synthetic dialogue choice whose availability and result are quest-controller-owned.

### 12. Immutable GameRules

Introduce one session rules object containing at minimum difficulty, the fixed
expansion target, hardcore, ladder/content-era, admission capacity, and the
content/rules fingerprint. Keep the mutable effective player-count override in
separate checkpointed authority state.

### 13. Difficulty vertical slices

Wire resistance penalty, life/mana steal divisors, one monster state-duration divisor and Static Field cap through their existing owners.

### 14. Encounter controller primitive

Implement a synthetic encounter controller with activation, spawned actor IDs, phase/timer state and completion event.

### 15. Campaign/special portal vertical slice

Have an encounter/quest completion create an authoritative portal that uses transition authority.

### 16. Cow Level composition

Only after Cube, quest, portal, special area and ruleset components exist, implement the target-version Cow Portal/eligibility flow.

## Do not implement these shortcuts

Do not:

- put object state in Lua or renderer nodes;
- let object animations toggle collision by themselves;
- make chests generate loot with a private RNG/item subsystem;
- make traps subtract HP directly;
- create separate player relocation code for stairs, waypoints, portals and quests;
- trust destination coordinates or waypoint unlock bits from the client;
- treat waypoint state as global across all difficulties;
- let a UI-open NPC menu prove the player is still in range;
- make dialogue text itself the stable choice ID;
- let speech playback completion own quest timing;
- scatter `if Hell` rules across unrelated systems;
- conflate Hardcore, Expansion, Ladder or player-count settings with Difficulty;
- implement Uber/later content as if it were 1.10f just because source structs contain conditional fields;
- implement special areas in a parallel minigame engine.

## Object operation result pattern

Establish a semantic result/request vocabulary that lets object behavior delegate to existing owners:

```text
ObjectMutations
LootRequests
MonsterSpawnRequests
Missile/CombatRequests
State/ResourceMutations
QuestEvents
ItemService/ConsumptionRequests
Transition/PortalRequests
Audio/PresentationEvents
```

The exact Go interface is not prescribed. Avoid cross-package giant mutable structs if smaller typed requests/events fit better.

## Transition invariants

Every cross-level transition should:

- verify source entity/endpoint and actor eligibility;
- choose destination/arrival server-side;
- update location/position atomically;
- reset/cancel incompatible movement/interaction state;
- update active world/room state;
- apply owned-unit transition policy;
- emit post-commit transition event;
- allow loading/camera/audio to observe but not veto.

## GameRules invariants

`GameRules` is immutable after game/session creation. Mutable game population,
`/players X`, and event facts are separately registered authority state.

It should participate in replay/content/network fingerprints.

Consumers request relevant rule values rather than parsing UI/menu state.

## Recommended PR granularity

```text
generalize interaction target kind
object instance state
one door
one chest
one shrine/well
transition endpoint registry
one stair
waypoint durable state
waypoint travel
Town Portal pair
NPC dialogue snapshot
one quest dialogue event
GameRules object
resistance/leech difficulty wiring
encounter controller
campaign final portal
Cow Portal eligibility
```

## Definition of a strong world-system vertical slice

```text
player clicks authored chest
 -> selector resolves stable Object target
 -> movement stops in operate range
 -> fixed-tick interaction command
 -> object authority validates unused state
 -> object seed feeds loot generation
 -> stable world item(s) materialize
 -> chest mode becomes used
 -> collision/selectability update if relevant
 -> semantic audio cue plays
 -> replay/checkpoint produces identical object/item state
```

A transition slice:

```text
player operates authored stair
 -> object emits semantic destination request
 -> transition authority validates/chooses arrival
 -> level/position/active world commit
 -> interaction/path reset
 -> owned unit follows according to category policy
 -> presentation loads destination and soundscape after commit
```

## Verification discipline

Use [WORLD_NPC_DIFFICULTY_VERIFICATION_QUEUE.md](WORLD_NPC_DIFFICULTY_VERIFICATION_QUEUE.md). Every behavior ID, quest gate, transition coordinate, timer, difficulty scalar and later-version special-content rule should remain a probe until source/runtime evidence is sufficient.
