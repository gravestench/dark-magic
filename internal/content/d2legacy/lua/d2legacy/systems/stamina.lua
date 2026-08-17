-- Advance authoritative Diablo II stamina at the 25 Hz simulation cadence.
--
-- Stamina is held in the same 8.8 fixed-point units used by life and mana.
-- CharStats supplies the class drain constant; named stat sources supply armor,
-- item drain, and recovery modifiers. Presentation receives only the projected
-- display integers and can never decide whether running remains available.

local ecs = require("engine.ecs/v1")
local movement_rules = require("d2legacy.movement_rules/v1")
local player_motion = require("d2legacy.gameplay.player_motion")
local M = {}

local function moving(velocity)
    return velocity:get("x") ~= 0 or velocity:get("y") ~= 0
end

local function update(entity, context)
    local vitals = ecs.get(entity, "d2legacy.player.vitals")
    local stats = ecs.get(entity, "d2legacy.player.movement_stats")
    local mode = ecs.get(entity, "d2legacy.player.movement_mode")
    local velocity = ecs.get(entity, "d2legacy.world.velocity")
    local location = ecs.get(entity, "d2legacy.world.location")
    local animation = ecs.get(entity, "d2legacy.player.animation")
    local is_moving = moving(velocity)
    local in_town = movement_rules.is_town(location:get("level_id"))
    local is_running = mode:get("running") and is_moving

    local animation_mode = animation and animation:get("mode") or ""
    local can_recover = animation_mode == "NU" or (is_moving and not is_running) or (in_town and is_moving)
    can_recover = can_recover or stats:get("staminarecoverybonus") >= 1000
    local current, force_walk = movement_rules.stamina(
        vitals:get("stamina_raw"),
        vitals:get("max_stamina_raw"),
        stats:get("run_drain"),
        stats:get("armor_run_drain"),
        stats:get("item_staminadrainpct"),
        stats:get("staminarecoverybonus"),
        is_running,
        is_moving,
        in_town,
        can_recover
    )
    vitals:set("stamina_raw", current)
    if force_walk then
        player_motion.force_walk(entity, context.tick)
    end

    vitals:set("stamina", math.floor(vitals:get("stamina_raw") / 256))
    vitals:set("max_stamina", math.floor(vitals:get("max_stamina_raw") / 256))
end

function M.register()
    ecs.system({
        id = "d2legacy.player.stamina",
        phase = "pre_simulation",
        after = { "d2legacy.player.resolve_motion" },
        query = {
            all = {
                "d2legacy.player.identity",
                "d2legacy.player.vitals",
                "d2legacy.player.movement_stats",
                "d2legacy.player.movement_mode",
                "d2legacy.player.motion",
                "d2legacy.world.velocity",
                "d2legacy.world.location",
            },
            none = { "d2legacy.player.death" },
        },
        read = {
            "d2legacy.player.identity",
            "d2legacy.player.movement_stats",
            "d2legacy.world.location",
            "d2legacy.player.animation",
            "d2legacy.player.motion",
        },
        write = {
            "d2legacy.player.vitals",
            "d2legacy.player.movement_mode",
            "d2legacy.player.motion",
            "d2legacy.world.velocity",
            "d2legacy.player.animation",
        },
        update = function(context, entities)
            for _, entity in ipairs(entities) do
                update(entity, context)
            end
        end,
    })
end

return M
