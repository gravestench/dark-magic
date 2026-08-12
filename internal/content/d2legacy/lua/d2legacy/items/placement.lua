-- Diablo II item footprint, named-slot, belt, held-item, and swap policy.
--
-- A placement is the item's current home. Grid homes use x/y plus the item's
-- width and height. Named homes use a body slot, belt slot, or service slot.
-- The invisible "held" home represents the item attached to the cursor. This
-- module only validates and writes homes; commands decide who may move them.

local ecs = require("engine.ecs/v1")
local M = {}

local grid_containers = {
    inventory = true,
    stash = true,
    cube = true,
}

local held_destinations = {
    equipment = true,
    hireling = true,
    belt = true,
    quest_service = true,
}

local function has_token(csv, wanted)
    for token in string.gmatch(csv or "", "[^,]+") do
        if token == wanted then return true end
    end
    return false
end

local function rectangles_overlap(left, right)
    local horizontal = left.x < right.x + right.width
        and right.x < left.x + left.width
    local vertical = left.y < right.y + right.height
        and right.y < left.y + left.height
    return horizontal and vertical
end

local function belongs_to_same_owner(left, right)
    return left:get("owner"):id() == right:get("owner"):id()
end

function M.grid_size(layout, container)
    if not grid_containers[container] then return nil, nil end
    return layout:get(container .. "_width"),
        layout:get(container .. "_height")
end

function M.overlaps(item_entity, item, destination, entities)
    local result = {}
    local candidate = {
        x = destination.x,
        y = destination.y,
        width = item:get("width"),
        height = item:get("height"),
    }
    for _, entity in ipairs(entities) do
        if entity:id() ~= item_entity:id() then
            local other = ecs.get(entity, "d2legacy.item.identity")
            local placed = ecs.get(entity, "d2legacy.item.placement")
            if other and placed
                and belongs_to_same_owner(other, item)
                and placed:get("container") == destination.container then
                local occupied = {
                    x = placed:get("x"),
                    y = placed:get("y"),
                    width = other:get("width"),
                    height = other:get("height"),
                }
                if rectangles_overlap(candidate, occupied) then
                    result[#result + 1] = entity
                end
            end
        end
    end
    return result
end

local function same_named_slot(placed, destination)
    if destination.container == "belt" then
        return placed:get("belt_slot") == destination.belt_slot
    end
    if placed:get("slot") ~= destination.slot then return false end
    local is_hand = destination.container == "equipment"
        and (destination.slot == "rarm" or destination.slot == "larm")
    return not is_hand or placed:get("weapon_set") == destination.weapon_set
end

function M.slot_occupant(item_entity, item, destination, entities)
    for _, entity in ipairs(entities) do
        if entity:id() ~= item_entity:id() then
            local other = ecs.get(entity, "d2legacy.item.identity")
            local placed = ecs.get(entity, "d2legacy.item.placement")
            if other and placed
                and belongs_to_same_owner(other, item)
                and placed:get("container") == destination.container
                and same_named_slot(placed, destination) then
                return entity
            end
        end
    end
    return nil
end

local function normalize(destination)
    destination.x = destination.x or 0
    destination.y = destination.y or 0
    destination.slot = destination.slot or ""
    destination.belt_slot = destination.belt_slot or 0
    destination.weapon_set = destination.weapon_set or 0
    destination.page = destination.page or 0
end

local function validate_grid(layout, item_entity, item, destination, entities,
    allow_one_overlap)
    local width, height = M.grid_size(layout, destination.container)
    if not width then return nil, false end
    local fits = destination.x >= 0
        and destination.y >= 0
        and destination.x + item:get("width") <= width
        and destination.y + item:get("height") <= height
    assert(fits, "item does not fit grid")
    -- One overlap is intentional during inventory reordering: the incoming
    -- held item takes the footprint and the displaced item becomes held.
    local overlaps = M.overlaps(item_entity, item, destination, entities)
    assert(#overlaps == 0 or allow_one_overlap and #overlaps == 1,
        "item footprint is occupied")
    return overlaps[1], true
end

local function validate_body_slot(item, destination)
    assert(destination.slot ~= "", "body slot is required")
    assert(has_token(item:get("body_slots"), destination.slot),
        "item cannot use body slot")
    local is_hand = destination.container == "equipment"
        and (destination.slot == "rarm" or destination.slot == "larm")
    if is_hand then
        assert(destination.weapon_set == 0 or destination.weapon_set == 1,
            "invalid weapon set")
    else
        assert(destination.weapon_set == 0,
            "shared body slot cannot use weapon set")
    end
end

local function validate_named_slot(layout, item, destination)
    if destination.container == "equipment"
        or destination.container == "hireling" then
        validate_body_slot(item, destination)
    elseif destination.container == "belt" then
        local valid = item:get("belt_eligible")
            and destination.belt_slot >= 0
            and destination.belt_slot < layout:get("belt_capacity")
        assert(valid, "item cannot use belt slot")
    elseif destination.container == "quest_service" then
        assert(destination.slot ~= "", "service slot is required")
    elseif destination.container ~= "world"
        and destination.container ~= "held" then
        error("unsupported item container")
    end
end

function M.validate(layout, item_entity, item, destination, entities,
    allow_one_overlap)
    assert(type(destination) == "table", "item destination is required")
    assert(type(destination.container) == "string",
        "item destination container is required")
    normalize(destination)

    -- Grid and named-slot containers have different occupancy rules, but both
    -- return the single item that should be swapped into the held home.
    local overlap, is_grid = validate_grid(
        layout,
        item_entity,
        item,
        destination,
        entities,
        allow_one_overlap
    )
    if is_grid then return overlap end

    validate_named_slot(layout, item, destination)
    local occupant = M.slot_occupant(
        item_entity,
        item,
        destination,
        entities
    )
    assert(not occupant or allow_one_overlap, "item slot is occupied")
    return occupant
end

function M.set(component, destination)
    normalize(destination)
    local fields = {
        "container",
        "x",
        "y",
        "slot",
        "belt_slot",
        "weapon_set",
        "page",
    }
    for _, field in ipairs(fields) do
        component:set(field, destination[field])
    end
end

function M.is_grid(container)
    return grid_containers[container] or false
end

function M.is_held_destination(container)
    return grid_containers[container]
        or held_destinations[container]
        or false
end

return M
