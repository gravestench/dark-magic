# Issue #613: authorize player vendor trades against the active NPC

## Intent and boundary

Ordinary player buy/sell commands currently trust payload vendor and category without rechecking the active NPC trade interaction. This permits stale or mismatched commands to mutate item placement and gold after the player closes the page or leaves range. The existing permanent Akara offer path does check active trading, so the missing gate belongs in `commands/item_vendor.lua`. This design completes child issue #613 under vendor roadmap #360. It changes repository command authority, not an original Diablo II pricing rule.

## Approaches

1. **Local apply-time player trade guard (chosen by independent architecture and content reviewers).** For player commands, check admitted player position/location, active target position/location, same act and level, range, `trade` page, vendor match, and exact category membership before mutating anything. Use `interactions.context.active` for target liveness and distance, adding explicit player/location checks because its general helper permits missing admission/location. This is bounded to vendor transactions.
2. Tighten `interactions.context.in_range` globally, then reuse it. This changes other interaction and administrator/bootstrap behavior beyond #613.
3. Check only during command validation. The interaction or location can change between validation and application; the command must be checked at commit time.

## Command path

Keep syntactic payload validation and `owner.resolve` unchanged. `validate` currently calls `terms(payload.vendor)` for every sell; defer that lookup for player sells until after the apply-time guard, so an unknown or mismatched vendor is a no-op rather than a validation error. Preserve the administrator/bootstrap terms validation and existing item/currency validation.

At the start of each player `apply_buy`/`apply_sell`, resolve the owner and perform one guard using current authoritative state. A missing player, player position/location, context, target/location, active `trade` page, vendor match, or category match returns without mutation. Compare act and level explicitly, then reuse the existing context range/target-liveness check. For buy, take the category from the authoritative item's vendor placement slot; for sell, use the validated payload category. Do not trust a client-supplied buy category. Exact comma-delimited token membership avoids substring matches. The check binds the submitted vendor to the active NPC; ordinary stock has no vendor identity, so it does not prove the item belongs to that NPC.

Keep the existing permanent-offer checks as an additional restriction. A supported player purchase still goes through `offers.trading` and its Normal-town/death/discount rules. Administrator commands retain their present path and need no NPC context; do not alter the general interaction helper. The currently player-facing trade page is Akara only. Enabling Charsi or another NPC trade page is separate work.

## Tests and verification

Use synthetic ordinary Akara stock and a real `interaction.open`/`interaction.choose` sequence in the Lua authority suite. Prove valid player buy and sell, then refusals after close, distance or level change, missing player admission/location, wrong vendor, disallowed category, and unavailable ordinary buy item. Each stale/mismatched command should no-op at application time, leave gold/placement/offer state unchanged, and allow a later valid session command. Retain checkpoint/replay parity. Existing invalid-held-sale assertions remain after a valid trade gate; missing buys remain no-ops.

The Go item acceptance fixture currently submits Charsi sales as player commands without an interaction. Move those fixture operations to administrator authority when they are only exercising item lifecycle and bootstrap behavior, retaining their outcome/restore assertions. Add or retain a production authority path for the player Akara trade. Run focused Lua/Go tests, full Go, architecture, Lua hardening, race, and complete-diff review on the submitted commit. Report owned-game-data test status explicitly.

## Unfinished work

This issue does not add vendor-specific stock ownership, stock generation, Charsi player dialogue, gambling, repair, discounts, or original-game vendor behavior verification. Follow-up issues under #360 own those gaps.
