# Vendors, pricing, repair, gambling, and economy research

Status: implementation-oriented research baseline. Dark Magic already has a strong authoritative vendor/item transfer skeleton, stable vendor stock placement, carried/stashed gold, buy/sell commands, and server-owned service rules. The current price formula is intentionally much simpler than Diablo II's full transaction-cost logic.

## Executive conclusion

Keep vendor **authority** exactly where it is, but split the remaining work into independent policies:

```text
NPC/vendor availability
        |
        v
VendorStock generation + refresh policy
        |
        v
Authoritative vendor inventory
        |
        +--> Buy quote/transaction
        +--> Sell quote/transaction
        +--> Repair/recharge quote/transaction
        +--> Identify/heal/hire/resurrect services
        `--> Gamble inventory + hidden outcome policy
```

Lua should render quotes and submit semantic choices. It should never calculate authoritative cost or generate stock.

## Current Dark Magic foundation

`internal/game/item` already provides:

- authority-arranged paged vendor stock;
- deterministic vendor-grid placement;
- buy-to-held and sell-held transitions;
- fixed-tick item commands;
- carried and stashed gold balances;
- per-vendor `TradeTerms` from NPC-like buy/sell/max-buy data;
- atomic service escrow/consumption;
- vendor UI layered over those commands.

This is the correct security/network shape.

## Current price model is a baseline, not Diablo-complete

Dark Magic currently calculates:

```text
sell-to-vendor = baseCost * BuyMultiplier / 1024, capped by MaxBuy
buy-from-vendor = baseCost * SellMultiplier / 1024
```

This is useful scaffolding, but D2Common's `ITEMS_CalculateTransactionCost` shows the original transaction model depends on substantially more item/player/vendor state.

Do not delete the current commerce boundary. Replace/extend only the price policy behind it after vectors exist.

## Original transaction-cost inputs

The 1.10f common item routine accepts:

- player;
- item;
- difficulty;
- quest flags;
- vendor/NPC ID;
- transaction type.

The recovered body references or derives factors including:

- base item cost;
- quality-specific records/properties;
- magic affixes;
- set/unique item identity;
- item type;
- character level;
- vendor NPC price multipliers/max-buy;
- reduced-prices item stat;
- durability/current versus max durability;
- repairability;
- ethereal state;
- quantity/stack state;
- charged-skill recharge cost;
- start-item/free-cost behavior;
- quest/vendor conditions.

The exact formula contains known reverse-engineering TODO/rounding uncertainty, so promote it through test vectors instead of line-for-line translation.

## Transaction kinds are semantic

A shared pricing engine should accept a semantic transaction kind:

```text
VendorBuy
VendorSell
Repair
Recharge
Gamble
HirelingHire
HirelingResurrect
ServicePurchase
```

Some kinds use item value; others use separate service formulas.

Do not overload `BaseCost` to mean repair or resurrection price.

## Price quote versus commit

For network/stale-state safety, make quotes advisory but commit authoritative:

```text
Quote request
 -> server resolves current item/vendor/player state
 -> returns price + quote revision/facts

Commit request
 -> server revalidates authoritative state
 -> recomputes or validates quote revision
 -> atomically moves item/gold/service state
```

A client-provided `nCost` field from a legacy protocol should never be trusted as the actual server cost in a modern Dark Magic realm design.

## Vendor stock is authoritative generated item state

D2MOO reconstructs NPC control records with:

- persistent vendor inventory;
- per-player gamble inventory;
- initialization/refresh flags;
- level-refresh behavior;
- tick state;
- NPC GUID/trade state;
- a vendor-control RNG seed.

Functions explicitly exist for:

- generating store items;
- filling store inventory;
- creating vendor caches;
- filling gambling inventory;
- buying/selling;
- repair;
- identify-all;
- healing;
- hire/resurrection services.

This supports modeling stock as real generated item instances with authoritative IDs, not UI rows synthesized every frame.

## Vendor stock generation context

Vendor stock should use a dedicated item-generation context distinct from monster drops.

Potential inputs:

- vendor identity/act;
- difficulty;
- player level;
- vendor minimum/maximum levels;
- item code/type lists from base item tables/NPC data;
- magic/quality chances;
- quantity/stack initialization;
- content version/expansion;
- vendor refresh generation/seed.

Generated stock can reuse M6 affix/property materialization but should not use monster TreasureClass NoDrop/MF behavior.

## Stock refresh

Original NPC trade state contains level-refresh and refresh-inventory flags plus tick data. Exact refresh triggers need probes.

Dark Magic needs a deterministic semantic policy such as:

```text
VendorStockGeneration
  vendor ID
  player/session scope
  refresh counter/seed
  player level snapshot
  content generation
```

Potential refresh events to verify:

- entering/leaving town or level;
- player level change;
- time/tick threshold;
- game creation;
- per-player vendor interaction;
- explicit vendor-cache creation.

Do not refresh because the vendor UI panel closed unless verified.

## Sold items and buyback

D2MOO's vendor has an inventory and sell path, so player-sold items can become vendor-owned stock rather than disappearing abstractly.

Dark Magic already moves sold items into `ContainerVendor`, which is an excellent foundation.

Research remaining:

- which vendors/categories accept which items;
- where sold items appear;
- whether they remain until refresh;
- price when bought back;
- max stock/page behavior;
- unique/quest/personalized/socketed item rules.

Stable item identity should survive sale and buyback.

## Repair

Repair is a mutation of the same item identity.

D2MOO exposes a dedicated repair item function plus the common transaction-cost path.

Repair needs to consider:

- current/max durability;
- broken state;
- indestructible;
- ethereal restriction;
- quantity for throwable/stackable durability models;
- charged skills/recharge;
- vendor service capability;
- reduced-price stat;
- gold source.

One "repair all" command should be an atomic or clearly ordered collection with a server-computed total quote.

## Recharge

Charged skills have current/max charges and can contribute additional transaction cost.

Recharge should update the item stat source/charged-skill availability through the same item-stat system used by skills and item activation.

Do not store charges only in UI tooltip state.

## Identify services

The original NPC code includes identify-all and identify-bought-item paths.

Identification should mutate only identified state, not reroll properties.

Cain/scroll/vendor variants may differ in cost/quest availability; model those as service policies.

## Healing services

NPC heal functions operate on player and pet/hireling state. Healing is a character/owned-unit transaction triggered by NPC interaction, not an item transaction internally.

The vendor/NPC service layer should call the player/owned-unit state system and emit semantic events.

## Gambling

Gambling is a special per-player vendor inventory in D2MOO (`D2NpcGambleStrc`).

A gambling item displayed to the player should preserve whatever information the original reveals while hiding final quality/concrete identity if that is the target behavior.

Recommended modern architecture:

```text
GambleOffer
  stable offer ID
  visible base presentation
  authoritative hidden generation seed/context
  price

Purchase
  server consumes offer/gold
  materializes final item using Gamble generation policy
  moves resulting item to held/inventory
  removes/replaces offer according to verified refresh behavior
```

Do not let client-side hover or UI reopening reroll an offer.

## Reduced prices

Item stat `item_reducedprices` participates in original transaction pricing.

Treat it as an effective player stat from active equipment/container sources. Vendor quote computation reads the authoritative value at transaction time.

Quest discounts or vendor-specific quest flags should be separate inputs rather than faked as invisible equipment stats unless the original data model actually does so.

## Gold balances and overflow

Dark Magic already separates carried/stashed gold. That should remain canonical.

Future economy work needs verified limits and overflow rules for:

- pickup;
- vendor sale;
- stash deposit/withdrawal;
- death drop;
- trade;
- quest/service rewards.

D2MOO's gold transaction path can spill excess carried gold into world piles in some contexts. Context-specific behavior must be explicit.

## Multiplayer ownership and vendor concurrency

If vendor stock is shared or per-player depends on the original service/context.

The server must define:

- whether ordinary vendor inventory is shared globally or projected per player;
- gamble inventory is per-player in the observed structure;
- simultaneous purchase conflict resolution;
- sold-item visibility;
- stock refresh while another player has UI open;
- stale quote/item ID rejection.

Use stable server item IDs and item-authority revisions.

## Suggested implementation slices

### V1 — quote API over current trade terms

Separate `QuoteBuy/QuoteSell` from commit while retaining current simple formula. Expose diagnostic components.

### V2 — richer transaction cost evaluator

Add quality/affix/durability/reduced-price components using synthetic vectors; keep unsupported formula terms explicit.

### V3 — generated vendor stock

Generate one vendor's stock through a dedicated `VendorStock` item-generation context and stable refresh generation.

### V4 — repair/recharge

Implement one item repair and charged-skill recharge with server quotes and atomic mutation.

### V5 — gambling

Add per-player stable GambleOffer state and one purchase/materialization path.

### V6 — service integration

Identify/heal/hire/resurrect/quest services reuse the same quote/commit and trusted interaction boundaries where appropriate.

## Verification backlog

1. Exact buy/sell/repair transaction formulas and integer rounding.
2. Quality, affix, set/unique and socket contribution to price.
3. Vendor-specific buy/sell multipliers and max-buy interpretation.
4. Reduced-price stat ordering/cap.
5. Quest discounts and quest-flag inputs.
6. Ethereal/indestructible repair restrictions.
7. Repair cost for broken/partially damaged items.
8. Recharge cost for charged skills and mixed repair+recharge.
9. Throwable quantity/durability repair behavior.
10. Vendor stock item-level/quality/affix generation by player level/difficulty.
11. Vendor item-type lists and magic-level behavior.
12. Vendor refresh triggers, timer and deterministic seed/call order.
13. Player-sold item persistence/buyback/category placement.
14. Shared versus per-player ordinary vendor stock.
15. Gamble offer generation, visible information, price and final quality probabilities.
16. Gamble refresh/purchase replacement behavior.
17. Identify-all/bought-item and Cain quest conditions.
18. Heal player/pet/hireling conditions and costs.
19. Carried/stashed gold limits and overflow behavior.
20. Multiplayer simultaneous purchase/stale-item semantics.

## Primary sources inspected

- Current Dark Magic `internal/game/item/vendor.go`, `commerce.go`, service/command/source architecture.
- D2MOO pinned 1.10f `D2Common/src/Items/Items.cpp` transaction-cost routines and `D2Game/UNIT/SUnitNpc.*` vendor/gamble/repair/service architecture.
