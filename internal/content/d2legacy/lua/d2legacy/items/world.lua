-- Ground-item attachment to generic authoritative world/ECS state.
--
-- Item placement decides whether the item is in the world. This module owns
-- only the reusable spatial component transition. Room policy remains in the
-- population owner and no item-specific inactive archive exists.

local ecs = require("engine.ecs/v1")
local population = require("d2legacy.bootstrap.population")
local collision = require("d2legacy.gameplay.collision")

local M = {}

function M.selectable(owner, item)
    local label = item.name_key and item.name_key ~= "" and item.name_key or item.code
    return { id = "item:" .. owner .. ":" .. item.id, kind = "item", label = label,
        owner = "", priority = 15, radius = 3.5 }
end

local function active_player(player)
    if not player or ecs.get(player, "d2legacy.world.inactive") or ecs.get(player, "d2legacy.player.death") then
        return false
    end
    local vitals = ecs.get(player, "d2legacy.player.vitals")
    return vitals ~= nil and vitals:get("health") > 0
end

function M.player(player_id)
    for _, entity in ipairs(ecs.query({ all = {
        "d2legacy.player.identity", "d2legacy.world.position", "d2legacy.world.location",
    } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == player_id
            and active_player(entity) then return entity end
    end
end

function M.within_reach(player, location, x, y)
    if not active_player(player) or not location then return false end
    local from_location = ecs.get(player, "d2legacy.world.location")
    if not from_location then return false end
    if from_location:get("act") ~= location:get("act")
        or from_location:get("level_id") ~= location:get("level_id") then return false end
    local from = ecs.get(player, "d2legacy.world.position")
    if not from then return false end
    local dx, dy = from:get("x") - x, from:get("y") - y
    return dx * dx + dy * dy <= 3.5 * 3.5
        and collision.line_clear(from_location:get("level_id"), from:get("x"), from:get("y"), x, y)
end

function M.can_pick_up(player, item)
    if not item or not active_player(player) or ecs.get(item, "d2legacy.world.inactive") then return false end
    local position = ecs.get(item, "d2legacy.world.position")
    return position and M.within_reach(player, ecs.get(item, "d2legacy.world.location"),
        position:get("x"), position:get("y")) or false
end

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
        ["d2legacy.world.selectable"] = M.selectable(owner, item),
        ["d2legacy.world.position"] = { x = item.x, y = item.y },
        ["d2legacy.world.location"] = { act = item.act, level_id = item.level_id },
        ["d2legacy.world.room_attach"] = { id = resident_id(owner, item.id) },
    }
end

function M.place(entity, owner, item_id, location, x, y)
    assert(location, "world item owner location is required")
    local id = resident_id(owner, item_id)
    local level_id = location:get("level_id")
    local identity = assert(ecs.get(entity, "d2legacy.item.identity"))
    ecs.set(entity, "d2legacy.world.selectable", M.selectable(owner, {id=item_id,code=identity:get("code")}))
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
        "d2legacy.world.selectable",
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
