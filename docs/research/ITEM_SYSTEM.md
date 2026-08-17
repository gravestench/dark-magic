# Base items, equipment, and inventory

Status: implementation-oriented research baseline. Dark Magic already has authoritative item identity/placement, atomic cursor swaps, weapon-set selection, vendor transactions, service escrow, and deterministic loot generation. This document identifies the remaining item-runtime fidelity work.

## Executive result

Preserve the current split between:

```text
BaseItemDefinition      immutable game-data record
GeneratedItemFacts      immutable roll results/properties
ItemInstance            stable durable/runtime identity + mutable condition
ItemPlacement           exactly one authoritative location
ItemPresentation        derived world/inventory/composite recipe
```

The current `internal/game/item.Item` intentionally contains only placement-relevant facts. Do not overload it into the complete item model by adding every stat/save field directly; compose a richer instance record while keeping placement validation focused.

## Current Dark Magic baseline

Current container identities include world, inventory, stash, cube, player equipment, hireling equipment, belt, held/in-hand, vendor stock, and quest/service sockets. Placement includes grid coordinates, named slot, belt slot, weapon set, and vendor page. State construction validates unique IDs and placements deterministically. Held placement can atomically displace one legal occupant into the hand. Weapon-set changes selection without moving both equipped pairs. Vendor buy/sell and service completion are separate authoritative commands.

The first world-placement residency slice is also executable. An imported item
with explicit act/level/subtile coordinates receives generic world position and
location plus a pre-plan ECS room-attachment request. Generated population-plan
admission resolves that request to the ordinary stable room resident; the
shared inactive marker then checkpoints the same item entity and filters local
and owner-private item projections. Existing placement commands remove spatial
and residency components on pickup and reacquire them on re-drop. This is a
synthetic player-layout-owned fixture proving state transitions. Public ground
ownership, generated-loot materialization, legal drop coordinates, pickup
range/path, reservation/allocation, and target 1.14d ground lifetime remain
unimplemented and must not be inferred from it.

This is a strong foundation and should be extended rather than replaced.

## Original item-instance evidence

Pinned D2MOO `Items.cpp` reconstructs item data exposing owner GUID, body location, per-item seed and start seed, quality, auto affix, multiple magic prefix/suffix IDs, rare name IDs, flags/command flags, item level, inventory page/cell, ear metadata, and variable inventory graphics index.

Confidence: **high** for 1.10f structure/behavior, while not every field is necessarily durable save state.

## Base definition versus instance

A base item definition describes code/type hierarchy, dimensions, base damage/defense/block, durability/quantity, speed/range, body locations, handedness/class restrictions, requirements, belt eligibility, appearance, normal/exceptional/elite family, and base value.

An instance describes what happened to one object:

```text
ItemInstance
  stable ID
  base code
  item level
  quality
  generation seed/start seed provenance
  quality-specific record IDs
  affix IDs / rolled properties
  semantic flags
  durability/current-max
  quantity
  socket count and ordered socket children
  personalization / ear metadata where applicable
  applied services
  location
  mutable revision
```

Generated stat entries belong in the item/stat model, not duplicated as presentation strings.

## Base-item namespaces

Weapons, armor, and misc are separate authored tables but participate in one item-code universe in many consumers. Resolve one code to exactly one admitted base record or diagnose ambiguity.

`ItemTypes.txt` defines type inheritance/category behavior. Do not flatten it to one category string if eligibility rules need parent chains or exclusions.

Research each use of base code, type/equivalent type, body location, weapon class, store page, compact save code/index, and unique/set base item link.

## Item location is a state machine

Exactly one authoritative location owns each item. A complete model must eventually cover:

```text
World(level/zone, position, room/context)
Inventory(owner, container, x, y)
Equipment(owner, body location, weapon set)
Hireling(owner/hireling, body location)
Belt(owner, slot)
Held(owner)
VendorStock(vendor, category, page, x, y)
TradeEscrow(trade ID, side/socket)
QuestService(owner, service ID, socket)
Corpse(owner, corpse ID, body location/container facts)
Golem(owner)
Deleted/Consumed(reason)
```

Extend the existing `Container` model only when a real behavior needs a new owner. Do not encode mutually exclusive locations as independent booleans.

## Atomic movement

Every transfer is a transaction:

```text
validate source ownership and expected revision
validate destination
compute displacement / side effects
apply placement and stat-source changes
record/audit
emit snapshot/event
```

A failed transaction leaves old placement and active stat sources untouched. Clients submit desired destination; authority computes displacement, price, compatibility, body-slot rules, service output, and gold effects.

Add stale-revision/expected-source checks for networked transfers so delayed commands cannot move an item that has already changed location or owner.

## Held item / cursor

The held container is authoritative, not UI state. Consequences:

- reconnect/UI reload does not lose the item;
- only one held item per owner;
- software hand hides because held item art becomes the pointer;
- placing on one occupied legal destination can swap atomically;
- closing a panel does not imply dropping/deleting the item.

Death, disconnect, trade cancellation, and save/exit still need explicit held-state policies.

## Equipment and weapon sets

Current state separates weapon set 0/1 for arm slots while shared body slots remain outside the swap. Continue that design.

Remaining research: body-location aliases; two-hand/offhand clearing; bow/crossbow/ammo rules; barbarian dual wield; shield/offhand compatibility; class restrictions; inactive weapon-set stat activation; requirement failure after stat loss; hireling restrictions.

These are authority rules, never UI rules.

## Durability, quantity, consumption

Mutable condition should be replayable:

```text
current durability / max durability
current quantity / max stack
replenish/repair metadata
indestructible state
```

Research weapon/armor durability loss, throwing/ammo consumption, replenish cadence, repair restrictions/cost, consumables, tome/key stacks, broken-item behavior, and stat suppression.

## Identification, personalization, ethereal, socketing

Model semantic state for identified, ethereal, socketed/socket count, personalized, runeword, broken/indestructible, and quest/service flags. Preserve unknown legacy bits in save adapters separately.

## Socket ownership

Sockets are containment, not ordinary grid placement. A socket child has exactly one parent item and ordered index. Moving the parent moves children implicitly.

Socket insertion validates parent/slot/item, removes child from its old location, attaches it, and updates derived stat/runeword state atomically. Removal/destruction policy belongs to socket/cube research.

## World drops

World placement uses semantic zone/level + subtile/world position, optional reservation/owner, and pickup flags. Presentation selects world art. Pickup is an authoritative transfer validated for range/path/interaction.

## Gold

Current `Layout.Gold` distinguishes carried/stashed balances. Gold transactions must be atomic with items for vendor buy/sell, repair, gambling, stash transfer, player trade, and death behavior.

## Trade and escrow

Player trade needs a dedicated escrow owner:

```text
TradeSession
  participants + revisions
  escrow items per side
  offered gold
  accept versions
```

Any offer mutation clears prior acceptance. Final commit revalidates ownership/destinations and transfers both sides atomically. Disconnect/cancel returns escrow deterministically.

## Corpse/death

Legacy saves can contain corpse item sections. Exact death transfer/reconnect behavior still needs tracing. Canonical item IDs must survive corpse movement; death changes location, not identity. Multiple-corpse behavior belongs to combat/persistence research.

## Vendor stock

Current Dark Magic correctly treats vendor stock as authority-arranged pages/categories and uses transactions. Remaining work: stock generation/refresh owner, per-player/shared behavior, gambling, buyback, quantity/repair, eligibility/pricing, persistence across town/reconnect.

## Replication

Separate public ground summary, owner inventory/equipment details, participant-only trade escrow, player-specific vendor stock, and hidden server generation fields. Clients do not decide true affixes, legal placement, price, or transaction results from UI state.

## Anti-duplication invariants

- One stable item ID exists at most once in authority.
- One item has exactly one location.
- A socket child is not also top-level.
- Transfer completes completely or not at all.
- Retry with same transaction identity is idempotent.
- Stale source/revision cannot commit.
- Recovery paths cannot create a second copy.
- Save export enumerates each durable item exactly once.
- Import detects duplicate identities/overlapping impossible placements.

## Implementation slices

1. canonical rich item instance composed with current placement state;
2. location completion for corpse/trade/socket/golem as dependencies require;
3. equipment rule resolver;
4. condition/durability/quantity system;
5. socket containment and derived-stat update;
6. trade escrow;
7. death/corpse integration;
8. legacy item save adapter independent of runtime types.

## Acceptance criteria

- Item cannot occupy two containers.
- Held swap is atomic/deterministic.
- Weapon-set selection changes activation without moving identities.
- Equip/unequip updates stat sources in same transaction as placement.
- Socket children are not serialized twice.
- Failed vendor/trade/service preserves item/gold state.
- Save/reload retains identity, immutable rolls, condition, location, socket graph.
- Disconnect during held/trade/service has deterministic recovery.
- World position is semantic world data, not screen pixels.
- Mod-added base codes do not require core placement changes.

## Verification backlog

1. Complete body-location/equipment compatibility matrix.
2. Trace two-handed/offhand behavior across classes/families.
3. Trace weapon-swap effects on stats/auras/charges/action state.
4. Trace durability, broken, repair, ethereal, indestructible, replenish.
5. Trace quantity/stack consumption for throw/ammo/keys/tomes/potions.
6. Trace identify/personalize flags and naming through save/network/UI.
7. Trace sockets and runeword activation/removal order.
8. Trace held-item behavior on save/exit/death/disconnect/level transition.
9. Trace corpse and multi-corpse edge cases.
10. Trace player trade acceptance/reset/commit/disconnect.
11. Trace vendor/buyback/stock identity/refresh.
12. Map legacy seed/quality/flags/body/inv-page fields to save bits.

## Sources

- Dark Magic `internal/game/item/types.go`, `state.go`, and `commands.go`.
- Dark Magic M6 loot/property roadmap and `internal/game/loot`.
- [D2MOO `Items.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Common/src/Items/Items.cpp).
- [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md).
- [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md).
- Bundled Data File Guide sections for Weapons/Armor/Misc/ItemTypes/Inventory/Belts and quality tables.
