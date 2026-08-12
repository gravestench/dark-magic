-- Admit opening and closing a Diablo NPC or world-object interaction.
--
-- Native input may name a target or point at world coordinates. The mod owns
-- target selection, ownership, and distance policy. The durable interaction
-- context is an ECS reference, so checkpoint and replay preserve it exactly.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")

local M = {}

local POINTER_RADIUS_SQUARED = 2.25

local function command_owner(command)
    local requested = command.payload.owner
    if command.authority == "player" then
        assert(not requested or requested == "" or requested == command.player,
            "cannot change another owner's interaction")
    end
    if requested and requested ~= "" then return requested end
    return command.player
end

local function interaction_context(owner)
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.interaction.context" },
    })) do
        local value = ecs.get(entity, "d2legacy.interaction.context")
        if value:get("owner") == owner then return value end
    end
    error("unknown interaction owner")
end

local function target_by_id(id)
    local wanted = string.lower(id)
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.interaction.target" },
    })) do
        local value = ecs.get(entity, "d2legacy.interaction.target")
        if value:get("id") == wanted then return entity, value end
    end
    return nil, nil
end

local function player_position(owner)
    for _, entity in ipairs(ecs.query({
        all = {
            "d2legacy.player.identity",
            "d2legacy.world.position",
        },
    })) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity:get("player") == owner then
            return ecs.get(entity, "d2legacy.world.position")
        end
    end
    return nil
end

local function squared_distance(x1, y1, x2, y2)
    local x = x1 - x2
    local y = y1 - y2
    return x * x + y * y
end

local function in_range(owner, target)
    local position = player_position(owner)
    if not position then return true end
    local distance = squared_distance(
        position:get("x"),
        position:get("y"),
        target:get("x"),
        target:get("y")
    )
    return distance <= target:get("radius") * target:get("radius")
end

local function target_at(x, y)
    local best_entity
    local best_target
    local best_distance

    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.interaction.target" },
    })) do
        local target = ecs.get(entity, "d2legacy.interaction.target")
        local distance = squared_distance(x, y, target:get("x"), target:get("y"))
        if distance <= POINTER_RADIUS_SQUARED
            and (not best_distance or distance < best_distance) then
            best_entity = entity
            best_target = target
            best_distance = distance
        end
    end
    return best_entity, best_target
end

local function requested_target(payload)
    if payload.at then return target_at(payload.x, payload.y) end
    return target_by_id(assert(payload.target, "target is required"))
end

function M.validate_open(command)
    command_owner(command)
    local payload = command.payload
    assert(payload.at or type(payload.target) == "string",
        "interaction target or point is required")
end

function M.open(command)
    local owner = command_owner(command)
    local entity, target = requested_target(command.payload)
    assert(entity and in_range(owner, target),
        "interaction target is unavailable or out of range")
    interaction_context(owner):set("target", entity)
end

function M.close(command)
    interaction_context(command_owner(command)):set("target", ecs.create())
end

function M.register()
    commands.register({
        kind = "interaction.open",
        authorities = { "player", "administrator" },
        validate = M.validate_open,
        apply = M.open,
    })
    commands.register({
        kind = "interaction.close",
        authorities = { "player", "administrator" },
        validate = command_owner,
        apply = M.close,
    })
end

return M
