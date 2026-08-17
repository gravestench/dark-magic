-- Apply one admitted keyboard or point-and-click movement sample.
--
-- The host only reports normalized input and a pathfinding waypoint. This mod
-- owns the Diablo rules that interpret it: walk/run speed, animation mode,
-- facing direction, and whether a new click replaces an attack action.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local player_motion = require("d2legacy.gameplay.player_motion")

local M = {}

local function is_finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

local function controlled_player(player_name)
    for _, entity in
        ipairs(ecs.query({
            all = { "d2legacy.world.player_control" },
        }))
    do
        local control = ecs.get(entity, "d2legacy.world.player_control")
        if control:get("player") == player_name then
            return entity
        end
    end
    return nil
end

local function cancel_attack(entity)
    ecs.remove(entity, "d2legacy.combat.attack_approach")
    ecs.remove(entity, "d2legacy.combat.attack_animation")
end

local function attack_owns_movement(entity)
    return ecs.get(entity, "d2legacy.combat.attack_approach") ~= nil
        or ecs.get(entity, "d2legacy.combat.attack_animation") ~= nil
end

function M.validate(command)
    local payload = command.payload
    assert(type(payload) == "table", "movement payload is required")
    assert(type(payload.x) == "number" and payload.x >= -1 and payload.x <= 1, "horizontal movement must be normalized")
    assert(type(payload.y) == "number" and payload.y >= -1 and payload.y <= 1, "vertical movement must be normalized")
    assert(type(payload.running) == "boolean", "movement mode is required")

    if payload.target then
        assert(is_finite(payload.target.x) and is_finite(payload.target.y), "movement target must be finite")
    end
end

function M.apply(command)
    local entity = controlled_player(command.player)
    if not entity then
        return
    end
    if ecs.get(entity, "d2legacy.player.death") then
        return
    end

    local payload = command.payload
    local explicit = payload.target ~= nil or payload.x ~= 0 or payload.y ~= 0

    -- An idle native input sample means "no new decision." It must not cancel
    -- a click-and-hold attack which currently owns velocity and animation.
    if explicit then
        cancel_attack(entity)
    elseif attack_owns_movement(entity) then
        return
    end

    player_motion.locomotion(entity, payload)
end

function M.register()
    commands.register({
        kind = "player.move",
        authorities = { "player" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
