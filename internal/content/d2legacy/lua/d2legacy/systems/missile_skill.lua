-- Materialize definition-driven straight and radial projectile families.

local ecs = require("engine.ecs/v1")
local geometry = require("d2legacy.policy.geometry")
local progression = require("d2legacy.policy.skill_progression")
local projectile_spawn = require("d2legacy.gameplay.projectile_spawn")

local M = {}

local function cast_id(caster, cast)
    return "projectile:"
        .. projectile_spawn.selectable_id(caster)
        .. ":skill:"
        .. cast:get("skill_id")
        .. ":effect:"
        .. cast:get("effect_tick")
end

local function projectile_components(caster, cast, definition, velocity_x, velocity_y, instance, target_x, target_y)
    local shared_cast_id = cast_id(caster, cast)
    local projectile_id = shared_cast_id
    if definition.trajectory == "radial" then
        projectile_id = projectile_id .. ":instance:" .. instance
    end
    return projectile_spawn.components(caster, definition, {
        cast_id = shared_cast_id,
        projectile_id = projectile_id,
        velocity_x = velocity_x,
        velocity_y = velocity_y,
        target_x = target_x,
        target_y = target_y,
        skill_level = cast:get("skill_level"),
        elemental_damage_percent = cast:get("elemental_damage_percent"),
        on_hit_state_duration = cast:get("effect_duration_ticks"),
    })
end

local function spawn_straight(context, caster, cast, definition, structural)
    local position = ecs.get(caster, "d2legacy.world.position")
    local dx, dy =
        geometry.normalized_direction(position:get("x"), position:get("y"), cast:get("target_x"), cast:get("target_y"))
    structural:create(
        projectile_components(
            caster,
            cast,
            definition,
            dx * definition.speed_per_tick,
            dy * definition.speed_per_tick,
            1,
            cast:get("target_x"),
            cast:get("target_y")
        )
    )
    local sound = projectile_spawn.sound_cue(caster, context.tick, "missile_spawn", definition.travel_sound)
    if sound then
        structural:create(sound)
    end
end

local function spawn_radial(context, caster, cast, definition, structural)
    local count =
        progression.linear(definition.missile_count_base, definition.missile_count_per_level, cast:get("skill_level"))
    for instance = 1, count do
        local angle = (instance - 1) * 2 * math.pi / count
        structural:create(
            projectile_components(
                caster,
                cast,
                definition,
                math.cos(angle) * definition.speed_per_tick,
                math.sin(angle) * definition.speed_per_tick,
                instance
            )
        )
        local sound = projectile_spawn.sound_cue(caster, context.tick, "missile_spawn", definition.travel_sound)
        if sound then
            structural:create(sound)
        end
    end
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.missile.spawn",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            all = { "d2legacy.skill.cast", "d2legacy.world.position", "d2legacy.world.location" },
            none = { "d2legacy.player.death" },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.selectable",
            "d2legacy.player.identity",
        },
        write = {
            "d2legacy.skill.cast",
            "d2legacy.missile.projectile",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.room_resident",
            "d2legacy.presentation.effect_cue",
        },
        update = function(context, casters, structural)
            for _, caster in ipairs(casters) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = definitions[cast:get("skill_id")]
                if definition and not cast:get("effect_emitted") and context.tick >= cast:get("effect_tick") then
                    if definition.trajectory == "straight" then
                        spawn_straight(context, caster, cast, definition, structural)
                        cast:set("effect_emitted", true)
                    elseif definition.trajectory == "radial" then
                        spawn_radial(context, caster, cast, definition, structural)
                        cast:set("effect_emitted", true)
                    end
                end
            end
        end,
    })
end

return M
