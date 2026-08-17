-- Apply one admitted keyboard or point-and-click movement sample.
--
-- The host only reports normalized input and a pathfinding waypoint. This mod
-- owns the Diablo rules that interpret it: walk/run speed, animation mode,
-- facing direction, and whether a new click replaces an attack action.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local movement_rules = require("d2legacy.movement_rules/v1")

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

local function update_animation(entity, x, y, running, tick)
    local animation = ecs.get(entity, "d2legacy.player.animation")
    if not animation then
        return
    end

    local moving = x ~= 0 or y ~= 0
    local mode = moving and (running and "RN" or "WL") or "NU"
    if animation:get("mode") ~= mode then
        animation:set("mode", mode)
        animation:set("start_tick", tick)
    end
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

    local payload = command.payload
    local explicit = payload.target ~= nil or payload.x ~= 0 or payload.y ~= 0

    -- An idle native input sample means "no new decision." It must not cancel
    -- a click-and-hold attack which currently owns velocity and animation.
    if explicit then
        cancel_attack(entity)
    elseif attack_owns_movement(entity) then
        return
    end

    local position = assert(ecs.get(entity, "d2legacy.world.position"))
    local identity = assert(ecs.get(entity, "d2legacy.player.identity"))
    local stats = assert(ecs.get(entity, "d2legacy.player.movement_stats"))
    local vitals = assert(ecs.get(entity, "d2legacy.player.vitals"))
    local running = payload.running and vitals:get("stamina_raw") > 0
    local velocity_x, velocity_y, moving =
        movement_rules.velocity(position:get("x"), position:get("y"), identity:get("class"), {
            x = payload.x,
            y = payload.y,
            target = payload.target,
            running = running,
        }, stats:get("velocitypercent"), stats:get("item_fastermovevelocity"))
    local velocity = assert(ecs.get(entity, "d2legacy.world.velocity"))
    velocity:set("x", velocity_x)
    velocity:set("y", velocity_y)

    local mode = ecs.get(entity, "d2legacy.player.movement_mode")
    if mode then
        mode:set("running", running)
    end
    update_animation(entity, moving and velocity_x or 0, moving and velocity_y or 0, running, command.tick)
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
