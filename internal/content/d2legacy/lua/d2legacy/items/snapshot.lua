-- Build value-only item snapshots for presentation.
-- No returned table is authoritative; commands must still mutate ECS state.

local ecs = require("engine.ecs/v1")
local M = {}

local function local_owner()
    local available, network = pcall(require, "engine.network/v1")
    if available then
        local status = network.status()
        if status.player_id and status.player_id ~= "" then
            return status.player_id
        end
    end
    return "local-player"
end

local function layout_for(owner)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.items.layout" } })) do
        local value = ecs.get(entity, "d2legacy.items.layout")
        if value:get("owner") == owner then
            return entity, value
        end
    end
end

function M.for_owner(owner)
    local owner_entity, layout = layout_for(owner)
    assert(layout, "unknown item owner")
    local result = {
        belt_capacity = layout:get("belt_capacity"),
        active_weapon_set = layout:get("active_weapon_set"),
        carried_gold = layout:get("carried_gold"),
        stashed_gold = layout:get("stashed_gold"),
        vendor_grid = { width = layout:get("vendor_width"), height = layout:get("vendor_height") },
        grids = {
            inventory = { width = layout:get("inventory_width"), height = layout:get("inventory_height") },
            stash = { width = layout:get("stash_width"), height = layout:get("stash_height") },
            cube = { width = layout:get("cube_width"), height = layout:get("cube_height") },
        },
        items = {},
    }
    for _, entity in
        ipairs(
            ecs.query({ all = { "d2legacy.item.identity", "d2legacy.item.placement", "d2legacy.item.presentation" } })
        )
    do
        local item = ecs.get(entity, "d2legacy.item.identity")
        if item:get("owner"):id() == owner_entity:id() then
            local placement = ecs.get(entity, "d2legacy.item.placement"):snapshot()
            local presentation = ecs.get(entity, "d2legacy.item.presentation"):snapshot()
            local value = {
                id = item:get("id"),
                code = item:get("code"),
                width = item:get("width"),
                height = item:get("height"),
                body_slots = item:get("body_slots"),
                belt_eligible = item:get("belt_eligible"),
                base_cost = item:get("base_cost"),
                applied_services = item:get("applied_services"),
                inventory_dc6 = presentation.inventory_dc6,
                world_dc6 = presentation.world_dc6,
                world_animated = presentation.world_animated,
                composite = presentation.composite,
                weapon_class = presentation.weapon_class,
            }
            for key, field in pairs(placement) do
                value[key] = field
            end
            result.items[#result.items + 1] = value
        end
    end
    table.sort(result.items, function(a, b)
        return a.id < b.id
    end)
    return result
end
function M.local_player()
    return M.for_owner(local_owner())
end
return M
