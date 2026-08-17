-- Bootstrap and operate configured paired warp endpoints.
--
-- The fixture configuration supplies locations and appearance only. Operation
-- goes through ordinary interaction admission, and final movement uses the one
-- shared transition transaction also consumed by authored level seams.

local ecs = require("engine.ecs/v1")
local initial_available, initial = pcall(require, "engine.initial_data/v1")
local operations = require("d2legacy.interactions.operations")
local transition = require("d2legacy.gameplay.transition")

local M = {}

local function finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

local function player(owner)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.player_control" } })) do
        if ecs.get(entity, "d2legacy.world.player_control"):get("player") == owner then
            return entity
        end
    end
    return nil
end

local function validate(endpoint)
    assert(type(endpoint.id) == "string" and endpoint.id ~= "", "warp id is required")
    assert(type(endpoint.pair_id) == "string" and endpoint.pair_id ~= "", "warp pair id is required")
    assert(type(endpoint.token) == "string" and endpoint.token ~= "", "warp appearance token is required")
    assert(finite(endpoint.x) and finite(endpoint.y), "warp position must be finite")
    assert(finite(endpoint.level_id) and endpoint.level_id > 0, "warp level is invalid")
    assert(finite(endpoint.radius) and endpoint.radius > 0, "warp radius is invalid")
    transition.validate({
        level_id = endpoint.destination_level,
        x = endpoint.destination_x,
        y = endpoint.destination_y,
        width = endpoint.destination_width,
        height = endpoint.destination_height,
    })
end

local function create(endpoint)
    validate(endpoint)
    for _, entity in
        ipairs(ecs.query({ all = {
            "d2legacy.world.warp",
            "d2legacy.world.selectable",
        } }))
    do
        if ecs.get(entity, "d2legacy.world.selectable"):get("id") == endpoint.id then
            return
        end
    end
    ecs.create({
        ["d2legacy.world.position"] = { x = endpoint.x, y = endpoint.y },
        ["d2legacy.world.location"] = { act = 1, level_id = endpoint.level_id },
        ["d2legacy.world.selectable"] = {
            id = endpoint.id,
            kind = "dynamic-object",
            label = endpoint.label,
            owner = "",
            radius = endpoint.radius,
            priority = 100,
        },
        ["d2legacy.interaction.target"] = {
            id = endpoint.id,
            npc = endpoint.label,
            vendor = "",
            categories = "",
            services = "",
            x = endpoint.x,
            y = endpoint.y,
            radius = endpoint.radius,
        },
        ["d2legacy.world.warp"] = {
            pair_id = endpoint.pair_id,
            destination_level = endpoint.destination_level,
            destination_x = endpoint.destination_x,
            destination_y = endpoint.destination_y,
            destination_width = endpoint.destination_width,
            destination_height = endpoint.destination_height,
        },
        ["d2legacy.world.warp_appearance"] = { token = endpoint.token },
    })
end

function M.register()
    operations.register("d2legacy.world.warp", function(owner, _, warp)
        local entity = assert(player(owner), "warp owner has no admitted player")
        transition.apply(entity, {
            level_id = warp:get("destination_level"),
            x = warp:get("destination_x"),
            y = warp:get("destination_y"),
            width = warp:get("destination_width"),
            height = warp:get("destination_height"),
        })
    end)
    local configuration = initial_available and initial.get("d2legacy.world_warps") or {}
    local endpoints = configuration.endpoints or {}
    local by_id = {}
    for _, endpoint in ipairs(endpoints) do
        validate(endpoint)
        assert(not by_id[endpoint.id], "duplicate warp id " .. endpoint.id)
        by_id[endpoint.id] = endpoint
    end
    for _, endpoint in ipairs(endpoints) do
        local pair = assert(by_id[endpoint.pair_id], "unknown paired warp " .. endpoint.pair_id)
        assert(pair.pair_id == endpoint.id, "warp pair relationship must be reciprocal")
        assert(pair.level_id == endpoint.destination_level, "warp destination level must contain its pair")
        create(endpoint)
    end
end

return M
