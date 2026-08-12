# UI-visible gameplay state and intent contracts

> Architecture note: presentation still reads snapshots and submits intents,
> but the gameplay authority consuming those intents may be a trusted,
> deterministic Lua system. “Engine-owned authority” below means the
> policy-neutral session host owns admission and consistency; it does not
> require Diablo rules to be implemented in Go.

Status: implementation-oriented research baseline. Dark Magic's Lua shim already documents and enforces the core principle: presentation reads safe snapshots, submits semantic intent, and never owns authoritative gameplay state. This document turns that principle into explicit contracts for every major gameplay surface so multiplayer/realm work does not accidentally reintroduce client authority.

## Executive conclusion

Every gameplay UI feature should answer two questions:

1. **What immutable/copy snapshot or ordered event stream may the UI read?**
2. **What semantic intent/command may the UI submit?**

The UI does not mutate game state directly.

```text
authoritative session/world/item/player/etc.
        |
        +--> PlayerViewSnapshot
        +--> InventorySnapshot
        +--> TargetSnapshot
        +--> QuestSnapshot
        +--> PartySnapshot
        +--> VendorSnapshot
        +--> World/AutomapSnapshot
        +--> ordered SemanticEvents
        |
        v
Lua presentation/widgets
        |
        +--> movement target/input intent
        +--> use skill / assign skill
        +--> item move/equip/drop/use
        +--> interact with stable target
        +--> NPC/dialogue choice
        +--> vendor/service/trade action
        +--> party/hostility action
        +--> waypoint/portal destination
        `--> UI-only local actions
```

## Existing Dark Magic presentation contract is already correct

The embedded Lua shim explicitly states:

- engine owns save data, authoritative simulation, networking and native state;
- Lua reads safe snapshots;
- Lua submits intent/commands;
- UI does not directly change authoritative item/game tables;
- versioned `dm.* /v1` capabilities form a stable modding boundary;
- retained render/audio handles are scope-owned presentation resources;
- gameplay can keep simulating under overlays while input ownership changes;
- pointer-first presentation converts clicks to movement/interaction/item/combat intents;
- headless tests can run presentation logic without proprietary assets.

Preserve this architecture as gameplay expands.

## Snapshot design rules

### Snapshot values are copies/projections

Lua/network clients should not hold pointers/references to mutable Go/ECS data.

A snapshot should contain stable semantic IDs and plain values:

```text
entity/item/quest/skill IDs
numbers/strings/enums
ordered arrays
booleans/ticks/revisions
```

Native handles are presentation-only capabilities and never gameplay identity.

### Snapshot order must be stable

For UI lists, provide an explicit order:

- authored order;
- body-slot order;
- grid coordinates;
- party join/roster order;
- quest/act order;
- skill tree order;
- vendor page/order;
- stable entity sorting.

Do not ask Lua to sort arbitrary maps to reconstruct gameplay meaning.

### Include revisions where stale intent matters

Item grids, trade offers, vendor stock, dialogue options and other mutable lists should expose a revision/generation/hash.

Intent can reference:

```text
stable item/choice ID
expected revision
```

Authority rejects stale requests instead of applying them to a visually old state.

### Distinguish public/private/debug snapshots

Do not expose all authoritative state to every client/mod API.

Examples:

- owner inventory/stash: private;
- nearby monster HP/name: public according to gameplay rules;
- hidden monster AI state/RNG: debug/admin only;
- gamble outcome seed: server private;
- account token: never presentation;
- quest internal controller state: expose only UI-relevant projection.

## Semantic events versus snapshots

Use snapshots for current truth and events for ordered occurrences.

Examples:

```text
Snapshot: HP = 37/100
Event: DamageApplied(source, 22)

Snapshot: item is in inventory at (2,3)
Event: ItemPickedUp(itemID)

Snapshot: quest objective = complete
Event: QuestAdvanced(...)
```

UI can recover from missed events by rebuilding from a snapshot. Do not make event delivery the sole state storage.

## HUD contract

HUD reads a `PlayerHUDSnapshot` containing only authoritative display facts:

```text
life/max life
mana/max mana
stamina/max stamina
level/experience progress
current level/area name ID
left/right assigned skill IDs
skill cooldown/delay presentation facts
selected target summary
active status/buff summary
belt slots/quantity
party/mercenary/pet summaries
current game difficulty/mode
```

HUD sends no `set_hp`/`set_mana` mutations.

Inputs include semantic skill use, run toggle, potion/belt item use, panel toggles and chat/UI actions.

## Resource interpolation

Presentation may animate orb/bar values between snapshots for smoothness, but interpolation is cosmetic.

Damage/heal authoritative values should snap/correct when a newer tick arrives.

Do not feed interpolated HP into skill/potion eligibility.

## Skill UI contract

Skill tree/assignment UI reads:

```text
class/skill tree definitions
skill points available
learned hard/effective levels
requirements/prerequisite satisfaction
left/right assignment eligibility
skill description/calculated display values
cooldown/delay state if relevant
```

Intent:

```text
learn skill(skillID, expected point/revision state)
assign skill(side, skillID)
use skill(side, semantic target)
```

The UI cannot grant a point or choose final mana/damage/result.

### Skill tooltip calculations

Prefer server/shared deterministic display calculation from the same normalized formula evaluator used by gameplay.

Do not duplicate complex skill formulas in Lua where they can drift from authority.

Presentation can format localized text around returned semantic values.

## Inventory/equipment/stash/Cube contract

UI reads an `ItemViewSnapshot`/container snapshot:

```text
item ID
container/slot/page/x/y
footprint
base/quality/name/display identity
identified/ethereal/broken/socket/runeword/personalized flags
quantity/durability/charges
visible property lines/stat values
requirements/usable reason
nested socket child display
container dimensions/revision
```

It submits semantic item actions using stable IDs/destinations:

```text
pick up / hold
place container cell
swap/equip/unequip
move stash/inventory/Cube
insert socket filler
use consumable
identify
repair/service request
drop to world
```

Authority validates capacity, requirements, ownership, container rules and resulting item-source changes.

### Hidden item facts

Unidentified items still have rolled properties server-side, but normal UI should not receive undisclosed property lines if legacy behavior hides them.

The owner client may receive enough opaque identity/version data for rendering while the server keeps hidden attributes until identified.

For a fully trusted offline client this separation is still valuable because it preserves network architecture and mod semantics.

## Item tooltip contract

Do not make tooltip code reverse-engineer effective properties from raw item flags.

Provide normalized semantic property lines or a structured tooltip model:

```text
name line(s)
base damage/defense
requirements
quality/set/runeword identity
durability/quantity/charges/socket count
properties grouped/order by compatibility rules
set inactive/active partial bonus visibility
price/value in current context if requested
```

Localization/style lives in Lua; gameplay property evaluation/order comes from engine data.

## Ground item labels

World snapshot supplies nearby/selectable item IDs and display facts allowed to the client.

UI decides label layout/filter/highlight, but pickup intent references the stable world item ID/semantic target and authority revalidates distance/ownership/capacity.

A loot filter mod can hide labels without deleting drops.

## Vendor contract

Vendor UI reads:

```text
interaction target/NPC/vendor identity
stock item snapshots
paged/category placement
stock revision
server-computed buy/sell/repair/service quotes
player carried gold
service capability list
```

Intent:

```text
buy itemID
sell held/itemID according to authority workflow
repair itemID / repair all
identify/heal/service ID
gamble offer ID
```

Never send trusted price, final gamble quality or vendor-generated item stats from Lua.

## Cube contract

Cube UI reads ordinary Cube container item state plus a `can_transmute`/match summary if useful.

Intent is simply `transmute` for the current Cube revision or a semantic operation request.

Authority matches recipe/inputs and generates outputs. UI does not submit a recipe row ID unless the original/mod UI genuinely chooses among recipes rather than deriving the only match.

## NPC dialogue contract

Dialogue snapshot:

```text
NPC instance ID
ordered dialogue/capability choices
choice stable IDs
localized text/speech IDs
availability/disabled reason if exposed
current conversation state/revision
```

Intent:

```text
npc.choose(targetID, choiceID, expected revision)
npc.close(targetID)
```

Server revalidates quest/items/services/range.

Lua can play/display speech/text after receiving selected-line events.

## Quest log contract

Quest UI reads a **projection**, not raw internal quest bits/controller memory:

```text
act/order
quest ID
status = unavailable/available/in-progress/completed/reward-pending
ordered objective lines
current objective state/progress where intended
reward state
optional map target hints
```

Raw legacy bit offsets remain save-codec/internal implementation detail.

Quest interaction intent normally occurs through NPC/object/item/world actions, not a `complete_quest` button.

## Waypoint UI contract

Snapshot:

```text
current waypoint/source ID
current difficulty
ordered acts/waypoint destinations
unlocked/locked/current state
server-known destination level IDs/name keys
revision/content generation
```

Intent:

```text
waypoint.travel(destinationWaypointID)
```

Server validates current waypoint interaction and unlock state, then invokes transition authority.

## Party screen contract

Snapshot:

```text
player roster IDs/names/classes/levels
party membership/invite relationship
same-game presence
current area/level summary according to legacy visibility rules
hostility relationship
Hardcore corpse-loot permission
ignore/squelch state
health/party-visible summaries where applicable
```

Intent:

```text
invite
cancel invite
accept
leave
hostile/unhostile
toggle corpse-loot permission
ignore/squelch/chat policy
```

UI never edits party IDs or friendliness directly.

## Trade UI contract

Snapshot is a projection of an authoritative `TradeSession`:

```text
trade ID
counterparty summary
your escrow items/gold
their visible offered item/gold projection
offer revision/hash
both acceptance flags
validation/status message
```

Intent:

```text
add/remove item through item authority
set offered gold
accept current revision
cancel
```

Final transfer is server atomic. Changing an offer invalidates acceptance in authority; UI simply reflects it.

## Mercenary/pet UI contract

Snapshot:

```text
owned-unit ID/category
name/class/level
life/max life
state alive/dead
skills summary
equipment projection
resurrection quote/availability
```

Intent:

```text
hire/resurrect service
move equipment through item authority
unsummon if category allows
use potion/heal interaction if supported
```

No direct pet HP/equipment mutation.

## Target/hover contract

Presentation hit-tests against an authoritative/selectable world projection containing stable semantic targets and select radii/bounds.

UI may choose among overlapping labels/nodes using presentation policy, but final interact/attack/skill command references a stable target or world point and authority rechecks:

- target existence;
- hostility/alignment;
- range;
- LOS/collision;
- alive/object state;
- skill/action policy.

## Automap contract

Automap is a presentation of authoritative/discovered world topology plus player-specific exploration state.

Separate:

```text
WorldMapTopology
  level geometry/rooms/exits/waypoints/object markers

ExplorationState
  which cells/rooms are revealed to this player

DynamicMarkers
  party players/quests/portals/objects according to visibility rules
```

Do not reveal undiscovered server map data to a remote client if compatibility/security requires hidden exploration.

Offline implementation may have full map locally but should still respect the projection contract.

## Chat/system message contract

Chat UI sends authenticated message intent to game/party/channel/social service. Server binds sender identity and applies ignore/squelch/rate/policy.

System messages are typed events with localization/message IDs and parameters where practical, not arbitrary trusted client strings.

Chat is normally outside deterministic gameplay checksum.

## Death/respawn UI contract

Death snapshot/event exposes only permitted choices:

- dead state;
- Hardcore permanence;
- respawn/town option;
- corpse presence/location summary as allowed;
- XP/gold loss result;
- corpse-loot permission state.

Intent chooses a valid respawn/rejoin action. Server commits death/corpse/persistence consequences.

## Save/character-select contract

Frontend character list is a projection from durable character/realm state:

```text
CharacterID
name/class/level
Hardcore/dead
Expansion/Classic
Ladder/season
progression/difficulty availability
appearance summary
last played metadata
```

Client selects CharacterID; realm validates ownership/revision/game compatibility.

Never load gameplay authority from Lua-side cached character data.

## Game list/create/join contract

UI reads realm directory projections and submits semantic create/join parameters.

Server decides GameID/worker/admission token and validates character mode/difficulty/content compatibility.

See `REALM_AND_BATTLENET_ARCHITECTURE.md`.

## UI-only state

Keep UI presentation state local and out of gameplay snapshots when it does not affect simulation:

- open panel/overlay;
- selected tab;
- scroll position;
- tooltip hover;
- label filter;
- controller focus;
- window layout;
- animation interpolation;
- audio volume/preferences;
- chat input buffer.

Input-focus state can decide which intents Lua sends, but the server does not need to replicate it as gameplay state.

## Presentation prediction

Allowed examples:

- animate character toward submitted movement target;
- start cast animation immediately;
- highlight a prospective item cell;
- gray out estimated insufficient-mana skill.

Canonical server snapshot may correct prediction. UI must tolerate rejection cleanly.

Never create an authoritative item/quest/HP state solely to make prediction convenient.

## Error/rejection contract

Command rejection should be structured enough for UI/logging:

```text
kind
stable reason code
human/localization parameters
current revision/tick if relevant
```

Avoid forcing Lua to parse English Go error strings for gameplay decisions.

## Versioned capability/schema contracts

Continue the existing `/v1` capability philosophy.

For complex snapshots, prefer explicit schema versions and additive evolution. A mod can declare supported versions and fail clearly rather than silently misreading changed fields.

Network schema and Lua capability schema may be separate adapters over the same semantic view model.

## Accessibility/testing

The control layer already exposes accessibility snapshots. Gameplay view contracts should use semantic names/roles/values so:

- screen-reader/accessibility adapters can describe items/quests/resources;
- headless tests can assert UI-visible state without rendering pixels;
- automation can exercise the same intents a human UI emits.

## UI compatibility research

Exact Diablo II UI presentation/layout/order belongs in presentation compatibility docs/manifests, but gameplay-visible semantics needing probes include:

- what stats are shown/hidden and rounding;
- tooltip property ordering;
- unidentified item information;
- party roster information visibility;
- automap exploration/party markers;
- quest status/objective text transitions;
- vendor quote/stock refresh timing;
- skill tooltip computed values;
- target health/name visibility.

Presentation research may reproduce the old look without moving authority into UI code.

## Suggested implementation slices

### UI1 — explicit PlayerHUD snapshot

Create a versioned immutable view from player/game state and migrate HUD resource/skill display to it.

### UI2 — structured command rejection

Add stable reason codes/revisions for one existing item or skill command and show the result through Lua without string parsing.

### UI3 — ItemView model

Centralize tooltip/container-visible item fields and hidden/identified policy.

### UI4 — Quest/Waypoint projections

Expose semantic quest status/objectives and waypoint unlock list without raw bitfields.

### UI5 — Party/Trade projections

Add versioned roster/invite/hostility/trade offer view models before building network UI.

### UI6 — network projection reuse

Use the same semantic view builders for local Lua and remote per-client snapshot adapters where appropriate, with private-field filtering.

## Verification backlog

1. Exact HUD resource/stat rounding and displayed fields.
2. Character sheet derived stat formulas/labels/breakpoints.
3. Skill tooltip value calculation and known client/server display differences.
4. Item tooltip property ordering and localization.
5. Identified/unidentified visible fields by quality/type.
6. Set/runeword/socket tooltip inactive/active lines.
7. Ground label name/quality/ownership visibility.
8. Quest log objective/status transitions and party state.
9. Waypoint ordering/lock visibility.
10. NPC dialogue option ordering/disabled versus absent choices.
11. Vendor stock/price refresh timing visible to UI.
12. Party roster location/health/hostility/lootability fields.
13. Trade acceptance delay/status and item visibility.
14. Automap exploration persistence/markers/party sharing.
15. Target monster name/HP/resistance/immunity/unique-mod display rules.
16. Death/corpse/Hardcore UI state.
17. Game list/create/join visible metadata.
18. Client prediction/correction UX for movement/skills/items.
19. Structured command rejection taxonomy.
20. Original-client packet fields needed for legacy UI adapters without leaking internal ECS state.

## Primary sources inspected

- Current Dark Magic embedded Lua shim living architecture documentation, versioned capabilities, retained UI/overlay model and existing gameplay item/interaction/movement/skill APIs.
- Current authoritative session/item/player/world/quest/transition/audio research.
- Pinned D2MOO multiplayer/party/NPC/client-event structures as evidence for UI-visible semantic events, without treating packet structs as internal API design.
