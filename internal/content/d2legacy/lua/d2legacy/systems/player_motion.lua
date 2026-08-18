-- Resolve one explicit player-motion owner into execution velocity and mode.

local ecs = require("engine.ecs/v1")
local movement_rules = require("d2legacy.movement_rules/v1")
local M = {}

local function update_animation(entity, moving, running, tick)
    local animation = ecs.get(entity, "d2legacy.player.animation")
    if not animation then
        return
    end
    local mode = moving and (running and "RN" or "WL") or "NU"
    if animation:get("mode") ~= mode then
        animation:set("mode", mode)
        animation:set("start_tick", tick)
    end
end

local function resolve(entity, context)
    local motion = ecs.get(entity, "d2legacy.player.motion")
    local velocity = ecs.get(entity, "d2legacy.world.velocity")
    local mode = ecs.get(entity, "d2legacy.player.movement_mode")
    if not motion:get("active") then
        velocity:set("x", 0)
        velocity:set("y", 0)
        mode:set("running", false)
        if motion:get("owner") == "locomotion" then
            update_animation(entity, false, false, context.tick)
        end
        return
    end

    local owner = motion:get("owner")
    assert(owner == "locomotion" or owner == "attack_approach", "unknown player motion owner " .. owner)
    local position = ecs.get(entity, "d2legacy.world.position")
    local identity = ecs.get(entity, "d2legacy.player.identity")
    local stats = ecs.get(entity, "d2legacy.player.movement_stats")
    local vitals = ecs.get(entity, "d2legacy.player.vitals")
    local running = owner == "locomotion" and motion:get("running") and vitals:get("stamina_raw") > 0
    local payload = {
        x = motion:get("x"),
        y = motion:get("y"),
        running = running,
    }
    if motion:get("has_target") then
        payload.target = { x = motion:get("target_x"), y = motion:get("target_y") }
    end
    local velocity_x, velocity_y, moving = movement_rules.velocity(
        position:get("x"),
        position:get("y"),
        identity:get("class"),
        payload,
        stats:get("velocitypercent"),
        stats:get("item_fastermovevelocity")
    )
    velocity:set("x", velocity_x)
    velocity:set("y", velocity_y)
    mode:set("running", running)
    update_animation(entity, moving, running, context.tick)
end

function M.register()
    ecs.system({
        id = "d2legacy.player.resolve_motion",
        phase = "pre_simulation",
        after = {
            "d2legacy.combat.player_melee_approach",
            "d2legacy.combat.player_melee_animation",
        },
        query = {
            all = {
                "d2legacy.player.motion",
                "d2legacy.player.identity",
                "d2legacy.player.vitals",
                "d2legacy.player.movement_stats",
                "d2legacy.player.movement_mode",
                "d2legacy.world.position",
                "d2legacy.world.velocity",
            },
            none = { "d2legacy.player.death", "d2legacy.skill.cast_action" },
        },
        read = {
            "d2legacy.player.motion",
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
            "d2legacy.player.movement_stats",
            "d2legacy.world.position",
            "d2legacy.player.death",
            "d2legacy.skill.cast_action",
        },
        write = {
            "d2legacy.world.velocity",
            "d2legacy.player.movement_mode",
            "d2legacy.player.animation",
        },
        update = function(context, entities)
            for _, entity in ipairs(entities) do
                resolve(entity, context)
            end
        end,
    })
end

return M
