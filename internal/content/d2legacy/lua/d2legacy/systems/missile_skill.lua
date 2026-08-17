-- Materialize definition-driven straight and radial projectile families.

local ecs = require("engine.ecs/v1")
local geometry = require("d2legacy.policy.geometry")
local population = require("d2legacy.bootstrap.population")
local progression = require("d2legacy.policy.skill_progression")

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

local function cast_id(caster, cast)
    return "projectile:"
        .. selectable_id(caster)
        .. ":skill:"
        .. cast:get("skill_id")
        .. ":effect:"
        .. cast:get("effect_tick")
end

local function projectile_components(caster, cast, definition, velocity_x, velocity_y, instance, target_x, target_y)
    local position = ecs.get(caster, "d2legacy.world.position")
    local location = ecs.get(caster, "d2legacy.world.location")
    local owner_id = selectable_id(caster)
    local shared_cast_id = cast_id(caster, cast)
    local projectile_id = shared_cast_id
    if definition.trajectory == "radial" then
        projectile_id = projectile_id .. ":instance:" .. instance
    end
    local minimum_damage, maximum_damage =
        progression.damage_range(definition, cast:get("skill_level"), cast:get("elemental_damage_percent"))
    local components = {
        ["d2legacy.world.position"] = { x = position:get("x"), y = position:get("y") },
        ["d2legacy.world.location"] = location:snapshot(),
        ["d2legacy.missile.projectile"] = {
            owner_id = owner_id,
            cast_id = shared_cast_id,
            target_x = target_x or position:get("x") + velocity_x * definition.lifetime_ticks,
            target_y = target_y or position:get("y") + velocity_y * definition.lifetime_ticks,
            velocity_x = velocity_x,
            velocity_y = velocity_y,
            previous_x = position:get("x"),
            previous_y = position:get("y"),
            remaining_ticks = definition.lifetime_ticks,
            collision_radius = definition.collision_radius,
            destroy_on_contact = definition.destroy_on_contact,
            next_hit_delay = definition.next_hit_delay,
            impact_radius = definition.impact_radius or 0,
            impact_missile_id = definition.impact_missile_id or "",
            impact_dcc = definition.impact_dcc or "",
            impact_palette = definition.impact_palette or "",
            impact_lifetime_ticks = definition.impact_lifetime_ticks or 0,
            impact_directions = definition.impact_directions or 1,
            impact_frames_per_second = definition.impact_frames_per_second or 1,
            impact_loop = definition.impact_loop or false,
            impact_sound = definition.impact_sound or "",
            on_hit_state_id = definition.on_hit_state_id or "",
            on_hit_state_source_id = "skill:" .. owner_id .. ":" .. definition.skill_id .. ":on-hit",
            on_hit_state_duration = cast:get("effect_duration_ticks"),
            on_hit_state_duration_policy = definition.on_hit_state_duration_policy or "",
            on_hit_state_action_disabled = definition.on_hit_state_action_disabled or false,
            on_hit_state_exclusive_group = definition.on_hit_state_exclusive_group or "",
            knockback_value = definition.knockback_value or 0,
            minimum_damage_raw = minimum_damage,
            maximum_damage_raw = maximum_damage,
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

local function spawn_straight(caster, cast, definition, structural)
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
end

local function spawn_radial(caster, cast, definition, structural)
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
        },
        update = function(context, casters, structural)
            for _, caster in ipairs(casters) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = definitions[cast:get("skill_id")]
                if definition and not cast:get("effect_emitted") and context.tick >= cast:get("effect_tick") then
                    if definition.trajectory == "straight" then
                        spawn_straight(caster, cast, definition, structural)
                        cast:set("effect_emitted", true)
                    elseif definition.trajectory == "radial" then
                        spawn_radial(caster, cast, definition, structural)
                        cast:set("effect_emitted", true)
                    end
                end
            end
        end,
    })
end

return M
