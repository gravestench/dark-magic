-- Apply one admitted keyboard or point-and-click movement sample.
--
-- The host only reports normalized input and a pathfinding waypoint. This mod
-- owns the Diablo rules that interpret it: walk/run speed, animation mode,
-- facing direction, and whether a new click replaces an attack action.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")

local M = {}

local WALK_SPEED = 10
local RUN_SPEED = 15
local ARRIVAL_DISTANCE = 0.2
local DIAGONAL_SCALE = 0.7071067811865476

local function is_finite(value)
    return type(value) == "number"
        and value == value
        and value ~= math.huge
        and value ~= -math.huge
end

local function controlled_player(player_name)
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.world.player_control" },
    })) do
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

local function vector_to_target(entity, target)
    local position = assert(ecs.get(entity, "d2legacy.world.position"))
    local x = target.x - position:get("x")
    local y = target.y - position:get("y")
    local distance = math.sqrt(x * x + y * y)

    if distance <= ARRIVAL_DISTANCE then
        return 0, 0
    end
    return x / distance, y / distance
end

local function normalized_vector(entity, payload)
    if payload.target then
        return vector_to_target(entity, payload.target)
    end
    if payload.x ~= 0 and payload.y ~= 0 then
        return payload.x * DIAGONAL_SCALE, payload.y * DIAGONAL_SCALE
    end
    return payload.x, payload.y
end

local function update_animation(entity, x, y, running)
    local animation = ecs.get(entity, "d2legacy.player.animation")
    if not animation then return end

    local moving = x ~= 0 or y ~= 0
    if not moving then
        animation:set("mode", "NU")
        return
    end

    animation:set("mode", running and "RN" or "WL")
end

function M.validate(command)
    local payload = command.payload
    assert(type(payload) == "table", "movement payload is required")
    assert(type(payload.x) == "number" and payload.x >= -1 and payload.x <= 1,
        "horizontal movement must be normalized")
    assert(type(payload.y) == "number" and payload.y >= -1 and payload.y <= 1,
        "vertical movement must be normalized")
    assert(type(payload.running) == "boolean", "movement mode is required")

    if payload.target then
        assert(is_finite(payload.target.x) and is_finite(payload.target.y),
            "movement target must be finite")
    end
end

function M.apply(command)
    local entity = controlled_player(command.player)
    if not entity then return end

    local payload = command.payload
    local explicit = payload.target ~= nil or payload.x ~= 0 or payload.y ~= 0

    -- An idle native input sample means "no new decision." It must not cancel
    -- a click-and-hold attack which currently owns velocity and animation.
    if explicit then
        cancel_attack(entity)
    elseif attack_owns_movement(entity) then
        return
    end

    local x, y = normalized_vector(entity, payload)
    local speed = payload.running and RUN_SPEED or WALK_SPEED
    local velocity = assert(ecs.get(entity, "d2legacy.world.velocity"))
    velocity:set("x", x * speed)
    velocity:set("y", y * speed)

    local mode = ecs.get(entity, "d2legacy.player.movement_mode")
    if mode then mode:set("running", payload.running) end
    update_animation(entity, x, y, payload.running)
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
