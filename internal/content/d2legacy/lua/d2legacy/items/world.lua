-- Ground-item attachment to generic authoritative world/ECS state.
--
-- Item placement decides whether the item is in the world. This module owns
-- only the reusable spatial component transition. Room policy remains in the
-- population owner and no item-specific inactive archive exists.

local ecs = require("engine.ecs/v1")
local population = require("d2legacy.bootstrap.population")

local M = {}

local function resident_id(owner, item_id)
    return "item:" .. owner .. ":" .. item_id
end

function M.initial_components(owner, item)
    if item.container ~= "world" or type(item.level_id) ~= "number" or item.level_id <= 0 then
        return {}
    end
    assert(type(item.act) == "number" and item.act > 0, "world item act is required")
    assert(type(item.x) == "number" and type(item.y) == "number", "world item position is required")
    return {
        ["d2legacy.world.position"] = { x = item.x, y = item.y },
        ["d2legacy.world.location"] = { act = item.act, level_id = item.level_id },
        ["d2legacy.world.room_attach"] = { id = resident_id(owner, item.id) },
    }
end

function M.place(entity, owner, item_id, location, x, y)
    assert(location, "world item owner location is required")
    local id = resident_id(owner, item_id)
    local level_id = location:get("level_id")
    ecs.set(entity, "d2legacy.world.position", { x = x, y = y })
    ecs.set(entity, "d2legacy.world.location", location:snapshot())
    ecs.remove(entity, "d2legacy.world.inactive")
    local resident = population.resident_at(id, level_id, x, y)
    if resident then
        ecs.set(entity, "d2legacy.world.room_resident", resident)
        ecs.remove(entity, "d2legacy.world.room_attach")
    else
        ecs.remove(entity, "d2legacy.world.room_resident")
        ecs.set(entity, "d2legacy.world.room_attach", { id = id })
    end
end

function M.clear(entity)
    for _, component in ipairs({
        "d2legacy.world.inactive",
        "d2legacy.world.room_resident",
        "d2legacy.world.room_attach",
        "d2legacy.world.location",
        "d2legacy.world.position",
    }) do
        ecs.remove(entity, component)
    end
end

return M
