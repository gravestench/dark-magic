-- Admit opening and closing a Diablo NPC or world-object interaction.
--
-- Native input may name a target or point at world coordinates. The mod owns
-- target selection, ownership, and distance policy. The durable interaction
-- context is an ECS reference, so checkpoint and replay preserve it exactly.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local operations = require("d2legacy.interactions.operations")

local M = {}

local POINTER_RADIUS_SQUARED = 2.25

local function command_owner(command)
    local requested = command.payload.owner
    if command.authority == "player" then
        assert(
            not requested or requested == "" or requested == command.player,
            "cannot change another owner's interaction"
        )
    end
    if requested and requested ~= "" then
        return requested
    end
    return command.player
end

local function interaction_context(owner)
    for _, entity in
        ipairs(ecs.query({
            all = { "d2legacy.interaction.context" },
        }))
    do
        local value = ecs.get(entity, "d2legacy.interaction.context")
        if value:get("owner") == owner then
            return value
        end
    end
    error("unknown interaction owner")
end

local function destroy_null_target(context)
    local target = context:get("target")
    local ok, marker = pcall(ecs.get, target, "d2legacy.interaction.null_target")
    if ok and marker then
        ecs.destroy(target)
    end
end

local function target_by_id(id)
    local wanted = string.lower(id)
    for _, entity in
        ipairs(ecs.query({
            all = { "d2legacy.interaction.target" },
            none = { "d2legacy.world.inactive" },
        }))
    do
        local value = ecs.get(entity, "d2legacy.interaction.target")
        if value:get("id") == wanted then
            return entity, value
        end
    end
    return nil, nil
end

local function player_entity(owner)
    for _, entity in
        ipairs(ecs.query({
            all = {
                "d2legacy.player.identity",
                "d2legacy.world.position",
            },
        }))
    do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity:get("player") == owner then
            return entity
        end
    end
    return nil
end

local function squared_distance(x1, y1, x2, y2)
    local x = x1 - x2
    local y = y1 - y2
    return x * x + y * y
end

local function in_range(owner, entity, target)
    local player = player_entity(owner)
    if not player then
        return true
    end
    local position = ecs.get(player, "d2legacy.world.position")
    local player_location = ecs.get(player, "d2legacy.world.location")
    local target_location = ecs.get(entity, "d2legacy.world.location")
    if player_location and target_location and player_location:get("level_id") ~= target_location:get("level_id") then
        return false
    end
    local distance = squared_distance(position:get("x"), position:get("y"), target:get("x"), target:get("y"))
    return distance <= target:get("radius") * target:get("radius")
end

local function target_at(owner, x, y)
    local best_entity
    local best_target
    local best_distance

    for _, entity in
        ipairs(ecs.query({
            all = { "d2legacy.interaction.target" },
            none = { "d2legacy.world.inactive" },
        }))
    do
        local target = ecs.get(entity, "d2legacy.interaction.target")
        local distance = squared_distance(x, y, target:get("x"), target:get("y"))
        if
            distance <= POINTER_RADIUS_SQUARED
            and in_range(owner, entity, target)
            and (not best_distance or distance < best_distance)
        then
            best_entity = entity
            best_target = target
            best_distance = distance
        end
    end
    return best_entity, best_target
end

local function requested_target(owner, payload)
    if payload.at then
        return target_at(owner, payload.x, payload.y)
    end
    return target_by_id(assert(payload.target, "target is required"))
end

function M.validate_open(command)
    command_owner(command)
    local payload = command.payload
    assert(payload.at or type(payload.target) == "string", "interaction target or point is required")
end

function M.open(command)
    local owner = command_owner(command)
    local entity, target = requested_target(owner, command.payload)
    -- Target existence, location, and range are mutable gameplay facts. A
    -- presentation prediction or queued network command can legitimately be
    -- stale by the time this admitted command applies. Reject that attempt as
    -- a no-op; malformed payloads remain validation errors, but ordinary
    -- latency must never terminate the authoritative session.
    if not entity or not in_range(owner, entity, target) then
        return
    end
    if operations.apply(owner, entity) then
        return
    end
    local context = interaction_context(owner)
    destroy_null_target(context)
    context:set("target", entity)
end

function M.close(command)
    local context = interaction_context(command_owner(command))
    destroy_null_target(context)
    context:set("target", ecs.create({ ["d2legacy.interaction.null_target"] = {} }))
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
