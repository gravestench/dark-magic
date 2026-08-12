-- Presentation-facing item API for the first-party mod.
--
-- Screens read snapshots from Lua-owned ECS state. Every requested change is
-- queued through the engine's generic command mailbox. The authoritative item
-- command runs on a later fixed tick, validates the request, and owns mutation.

local intents = require("engine.command_intent/v1")
local snapshots = require("d2legacy.items.snapshot")

local M = { api = 1 }

function M.snapshot()
    return snapshots.local_player()
end

function M.move(item_id, destination, place_held)
    intents.submit("item.move", {
        item_id = item_id,
        destination = destination,
        place_held = place_held == true,
    })
end

function M.select_weapon_set(set)
    intents.submit("item.weapon_set", { set = set })
end

function M.sell_held(item_id, vendor, category)
    intents.submit("item.vendor_sell", {
        item_id = item_id,
        vendor = vendor,
        category = category,
    })
end

function M.buy_to_held(item_id, vendor)
    intents.submit("item.vendor_buy", {
        item_id = item_id,
        vendor = vendor,
    })
end

function M.complete_service(service)
    intents.submit("item.service_complete", { service = service })
end

return M
