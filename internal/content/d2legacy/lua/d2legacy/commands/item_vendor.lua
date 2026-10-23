-- Authoritative vendor pricing, buying, selling, and page arrangement.
--
-- Vendors are not player-controlled grids. Selling removes the held item from
-- the cursor and lets the vendor arranger choose its page and slot. Buying
-- performs the reverse transaction and refuses to overwrite an already-held
-- item. Every price is integer arithmetic using the legacy 1024 scale.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local owner = require("d2legacy.items.command_owner")
local find = require("d2legacy.items.find")
local placement = require("d2legacy.items.placement")
local arrangement = require("d2legacy.items.vendor_arrangement")
local offers = require("d2legacy.items.vendor_offers")
local contexts = require("d2legacy.interactions.context")
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
    if not buying and command.authority ~= "player" then terms(payload.vendor) end
    if buying then
        assert(payload.to_inventory == nil or type(payload.to_inventory) == "boolean", "inventory purchase choice must be boolean")
    end
end

local function allowed_category(categories, wanted)
    for category in string.gmatch(categories, "[^,]+") do
        if category == wanted then return true end
    end
    return false
end

local function player_trading(command, owner_id, category)
    if command.authority ~= "player" then return true end
    local player_location
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity:get("player") == owner_id then
            if not ecs.get(entity, "d2legacy.world.position") then return false end
            player_location = ecs.get(entity, "d2legacy.world.location")
            break
        end
    end
    if not player_location then return false end
    local context, target = contexts.active(owner_id)
    if not context or context:get("dialogue_page") ~= "trade" then return false end
    local target_location = ecs.get(context:get("target"), "d2legacy.world.location")
    if not target_location or player_location:get("act") ~= target_location:get("act")
        or player_location:get("level_id") ~= target_location:get("level_id") then return false end
    return string.lower(target:get("vendor")) == string.lower(command.payload.vendor)
        and allowed_category(target:get("categories"), category)
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
    local owner_id = owner.resolve(command)
    if not player_trading(command, owner_id, payload.category) then return end
    local entities = all_items()
    local layout_entity, layout = assert(find.layout(
        owner_id,
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

    -- The arranger owns vendor pagination. A seller chooses a category, never
    -- an exact vendor cell.
    arrangement.apply(layout, layout_entity, payload.category, {
        entity = entity,
        item = item,
        placed = placed,
    }, nil, entities)
    layout:set("carried_gold", layout:get("carried_gold") + price)
end

local function buy_offer(command, layout_entity, layout, entity, item, offer, entities)
    local payload = command.payload
    if not payload.to_inventory or not offers.trading(owner.resolve(command), string.lower(payload.vendor))
        or offer:get("vendor") ~= string.lower(payload.vendor) then return end
    local placed = ecs.get(entity, "d2legacy.item.placement")
    if not placed or placed:get("container") ~= "vendor" or placed:get("slot") ~= "misc"
        or ecs.get(entity, "d2legacy.world.inactive") then return end
    if layout:get("carried_gold") < offer:get("price") then return end
    if placement.slot_occupant(entity, item, { container = "held", slot = "" }, entities) then return end
    for y = 0, layout:get("inventory_height") - item:get("height") do
        for x = 0, layout:get("inventory_width") - item:get("width") do
            local destination = { container = "inventory", x = x, y = y }
            if #placement.overlaps(entity, item, destination, entities) == 0 then
                placement.validate(layout, entity, item, destination, entities)
                offers.purchase(layout_entity, entity, item, destination)
                layout:set("carried_gold", layout:get("carried_gold") - offer:get("price"))
                return
            end
        end
    end
end

function M.apply_buy(command)
    local payload = command.payload
    local entities = all_items()
    local owner_id = owner.resolve(command)
    local layout_entity, layout = find.layout(
        owner_id,
        entities
    )
    if not layout_entity then return end
    local entity, item = find.item(
        layout_entity,
        payload.item_id,
        entities
    )
    if not entity then return end
    if command.authority == "player" then
        local placed = ecs.get(entity, "d2legacy.item.placement")
        if not placed or placed:get("container") ~= "vendor"
            or not player_trading(command, owner_id, placed:get("slot")) then return end
    end
    local offer = ecs.get(entity, "d2legacy.vendor.offer")
    if offer then
        buy_offer(command, layout_entity, layout, entity, item, offer, entities)
        return
    end
    -- Inventory purchases are restricted to the supported permanent offers.
    if payload.to_inventory then return end
    local placed = assert(ecs.get(entity, "d2legacy.item.placement"))
    if placed:get("container") ~= "vendor" then return end

    -- A purchased item first enters the durable held location. The player can
    -- then place it in inventory, belt, stash, cube, or equipment normally.
    local held_destination = { container = "held", slot = "" }
    local occupied = placement.slot_occupant(
        entity,
        item,
        held_destination,
        entities
    )
    if occupied then return end

    local price = scaled_price(item, terms(payload.vendor):get("sell_multiplier"))
    if layout:get("carried_gold") < price then return end

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
