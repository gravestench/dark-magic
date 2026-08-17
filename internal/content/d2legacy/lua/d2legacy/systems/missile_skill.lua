-- Materialize straight projectiles for casts whose definition selects this family.

local ecs = require("engine.ecs/v1")
local geometry = require("d2legacy.policy.geometry")
local population = require("d2legacy.bootstrap.population")

local M = {}

local function selectable_id(entity)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    if selectable then
        return selectable:get("id")
    end
    local identity = ecs.get(entity, "d2legacy.player.identity")
    if identity then
        return "player:" .. identity:get("player")
    end
    return "entity:" .. entity:id()
end

local function projectile_components(caster, cast, definition)
    local position = ecs.get(caster, "d2legacy.world.position")
    local location = ecs.get(caster, "d2legacy.world.location")
    local dx, dy =
        geometry.normalized_direction(position:get("x"), position:get("y"), cast:get("target_x"), cast:get("target_y"))
    local owner_id = selectable_id(caster)
    local projectile_id = "projectile:"
        .. owner_id
        .. ":skill:"
        .. cast:get("skill_id")
        .. ":effect:"
        .. cast:get("effect_tick")
    local components = {
        ["d2legacy.world.position"] = { x = position:get("x"), y = position:get("y") },
        ["d2legacy.world.location"] = location:snapshot(),
        ["d2legacy.missile.projectile"] = {
            owner_id = owner_id,
            target_x = cast:get("target_x"),
            target_y = cast:get("target_y"),
            velocity_x = dx * definition.speed_per_tick,
            velocity_y = dy * definition.speed_per_tick,
            previous_x = position:get("x"),
            previous_y = position:get("y"),
            remaining_ticks = definition.lifetime_ticks,
            collision_radius = definition.collision_radius,
            knockback_value = definition.knockback_value,
            minimum_damage_raw = definition.minimum_damage_raw,
            maximum_damage_raw = definition.maximum_damage_raw,
            damage_channel = definition.damage_channel,
            missile_id = definition.missile_id,
            dcc = definition.dcc,
            palette = definition.palette,
            travel_sound = definition.travel_sound,
            hit_sound = definition.hit_sound,
            directions = definition.directions,
            frames_per_second = definition.frames_per_second,
            loop = definition.loop,
            offset_x = definition.offset_x,
            offset_y = definition.offset_y,
            offset_z = definition.offset_z,
        },
    }
    local resident =
        population.resident_at(projectile_id, location:get("level_id"), position:get("x"), position:get("y"))
    if resident then
        components["d2legacy.world.room_resident"] = resident
    end
    return components
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.missile.spawn_straight",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            all = {
                "d2legacy.skill.cast",
                "d2legacy.world.position",
                "d2legacy.world.location",
            },
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
        },
        update = function(context, casters, structural)
            for _, caster in ipairs(casters) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = definitions[cast:get("skill_id")]
                if
                    definition
                    and definition.behavior == "missile.straight"
                    and not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick")
                then
                    structural:create(projectile_components(caster, cast, definition))
                    cast:set("effect_emitted", true)
                end
            end
        end,
    })
end

return M
