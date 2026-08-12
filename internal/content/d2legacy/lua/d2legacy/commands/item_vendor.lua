-- Authoritative vendor pricing, buying, selling, and page arrangement.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local owner = require("d2legacy.items.command_owner")
local find = require("d2legacy.items.find")
local placement = require("d2legacy.items.placement")
local arrangement = require("d2legacy.items.vendor_arrangement")
local M = {}

local PRICE_SCALE = 1024

local function all_items()
    return ecs.query({
        any = {
            "d2legacy.items.layout",
            "d2legacy.item.identity",
        },
    })
end

local function terms(vendor)
    local wanted = string.lower(vendor)
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.vendor.terms" },
    })) do
        local value = ecs.get(entity, "d2legacy.vendor.terms")
        if value:get("vendor") == wanted then return value end
    end
    error("unknown vendor")
end

local function validate(command, buying)
    local payload = command.payload
    assert(type(payload) == "table", "vendor payload is required")
    assert(type(payload.item_id) == "string" and payload.item_id ~= "",
        "item is required")
    assert(type(payload.vendor) == "string" and payload.vendor ~= "",
        "vendor is required")
    if not buying then
        local category = payload.category
        assert(type(category) == "string" and category ~= "",
            "category is required")
        assert(not string.find(category, "/", 1, true),
            "category cannot contain a slash")
    end
    owner.resolve(command)
    terms(payload.vendor)
end

local function held_item(layout_entity, item_id, entities)
    local entity, item = assert(find.item(layout_entity, item_id, entities))
    local placed = assert(ecs.get(entity, "d2legacy.item.placement"))
    assert(placed:get("container") == "held", "item is not held")
    return entity, item, placed
end

local function scaled_price(item, multiplier)
    return math.floor(item:get("base_cost") * multiplier / PRICE_SCALE)
end

function M.apply_sell(command)
    local payload = command.payload
    local entities = all_items()
    local layout_entity, layout = assert(find.layout(
        owner.resolve(command),
        entities
    ))
    local entity, item, placed = held_item(
        layout_entity,
        payload.item_id,
        entities
    )
    local rule = terms(payload.vendor)
    local price = scaled_price(item, rule:get("buy_multiplier"))
    if rule:get("max_buy") > 0 then
        price = math.min(price, rule:get("max_buy"))
    end

    arrangement.apply(layout, layout_entity, payload.category, {
        entity = entity,
        item = item,
        placed = placed,
    }, nil, entities)
    layout:set("carried_gold", layout:get("carried_gold") + price)
end

function M.apply_buy(command)
    local payload = command.payload
    local entities = all_items()
    local layout_entity, layout = assert(find.layout(
        owner.resolve(command),
        entities
    ))
    local entity, item = assert(find.item(
        layout_entity,
        payload.item_id,
        entities
    ))
    local placed = assert(ecs.get(entity, "d2legacy.item.placement"))
    assert(placed:get("container") == "vendor", "item is not vendor stock")

    local held_destination = { container = "held" }
    local occupied = placement.slot_occupant(
        entity,
        item,
        held_destination,
        entities
    )
    assert(not occupied, "held item already exists")

    local price = scaled_price(item, terms(payload.vendor):get("sell_multiplier"))
    assert(layout:get("carried_gold") >= price, "insufficient carried gold")

    arrangement.apply(
        layout,
        layout_entity,
        placed:get("slot"),
        nil,
        entity,
        entities
    )
    placement.set(placed, held_destination)
    layout:set("carried_gold", layout:get("carried_gold") - price)
end

function M.register()
    commands.register({
        kind = "item.vendor_sell",
        authorities = { "player", "administrator" },
        validate = function(command) validate(command, false) end,
        apply = M.apply_sell,
    })
    commands.register({
        kind = "item.vendor_buy",
        authorities = { "player", "administrator" },
        validate = function(command) validate(command, true) end,
        apply = M.apply_buy,
    })
end

return M
