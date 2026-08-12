-- Arrange vendor stock into deterministic pages.
--
-- Diablo vendors auto-place stock. A player names a semantic category while
-- this policy chooses page and cell coordinates in stable code/identity order.

local ecs = require("engine.ecs/v1")
local placement = require("d2legacy.items.placement")
local M = {}

local function before(left, right)
    local left_code = left.item:get("code")
    local right_code = right.item:get("code")
    if left_code ~= right_code then return left_code < right_code end
    return left.item:get("id") < right.item:get("id")
end

local function vendor_items(layout_entity, category, excluded, entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local item = ecs.get(entity, "d2legacy.item.identity")
        local placed = ecs.get(entity, "d2legacy.item.placement")
        local included = item and placed
            and item:get("owner"):id() == layout_entity:id()
            and (not excluded or entity:id() ~= excluded:id())
            and placed:get("container") == "vendor"
            and placed:get("slot") == category
        if included then
            result[#result + 1] = {
                entity = entity,
                item = item,
                placed = placed,
            }
        end
    end
    table.sort(result, before)
    return result
end

local function overlaps(candidate, x, y, page, occupied)
    for _, other in ipairs(occupied) do
        local horizontal = x < other.x + other.width
            and other.x < x + candidate:get("width")
        local vertical = y < other.y + other.height
            and other.y < y + candidate:get("height")
        if other.page == page and horizontal and vertical then return true end
    end
    return false
end

local function find_position(layout, category, item, occupied)
    local max_x = layout:get("vendor_width") - item:get("width")
    local max_y = layout:get("vendor_height") - item:get("height")
    assert(max_x >= 0 and max_y >= 0, "item cannot fit a vendor page")

    for page = 0, 1000 do
        for x = 0, max_x do
            for y = 0, max_y do
                if not overlaps(item, x, y, page, occupied) then
                    occupied[#occupied + 1] = {
                        x = x,
                        y = y,
                        width = item:get("width"),
                        height = item:get("height"),
                        page = page,
                    }
                    return {
                        container = "vendor",
                        slot = category,
                        page = page,
                        x = x,
                        y = y,
                    }
                end
            end
        end
    end
    error("vendor page limit exhausted")
end

function M.apply(layout, layout_entity, category, added, excluded, entities)
    local values = vendor_items(layout_entity, category, excluded, entities)
    if added then
        values[#values + 1] = added
        table.sort(values, before)
    end

    -- Compute every placement before mutating any component. A failure cannot
    -- expose a half-rearranged vendor catalog to checkpointing or rendering.
    local arranged = {}
    local occupied = {}
    for _, value in ipairs(values) do
        arranged[#arranged + 1] = {
            component = value.placed,
            destination = find_position(layout, category, value.item, occupied),
        }
    end
    for _, value in ipairs(arranged) do
        placement.set(value.component, value.destination)
    end
end

return M
