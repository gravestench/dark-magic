-- Build ordinary authoritative projectile entities from a record definition.
--
-- Cast skills and reactive states differ only in why a projectile starts. Once
-- admitted, both use this boundary so movement, swept contact, mitigation,
-- checkpointing, and presentation consume the same component contract.

local ecs = require("engine.ecs/v1")
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

function M.selectable_id(entity)
    return selectable_id(entity)
end

function M.components(source, definition, values)
    local position = assert(ecs.get(source, "d2legacy.world.position"), "projectile source has no position")
    local location = assert(ecs.get(source, "d2legacy.world.location"), "projectile source has no location")
    local owner_id = values.owner_id or selectable_id(source)
    local minimum_damage = values.minimum_damage_raw
    local maximum_damage = values.maximum_damage_raw
    if minimum_damage == nil or maximum_damage == nil then
        minimum_damage, maximum_damage = progression.damage_range(
            definition,
            assert(values.skill_level, "projectile skill level is required"),
            values.elemental_damage_percent or 0
        )
    end
    local components = {
        ["d2legacy.world.position"] = { x = position:get("x"), y = position:get("y") },
        ["d2legacy.world.location"] = location:snapshot(),
        ["d2legacy.missile.projectile"] = {
            owner_id = owner_id,
            cast_id = assert(values.cast_id, "projectile cast ID is required"),
            target_x = values.target_x or position:get("x") + values.velocity_x * definition.lifetime_ticks,
            target_y = values.target_y or position:get("y") + values.velocity_y * definition.lifetime_ticks,
            velocity_x = values.velocity_x,
            velocity_y = values.velocity_y,
            previous_x = position:get("x"),
            previous_y = position:get("y"),
            remaining_ticks = definition.lifetime_ticks,
            collision_radius = definition.collision_radius,
            destroy_on_contact = definition.destroy_on_contact,
            impact_on_expiry = definition.impact_on_expiry or false,
            next_hit_delay = definition.next_hit_delay,
            impact_radius = progression.linear(
                definition.impact_radius or 0,
                definition.impact_radius_per_level or 0,
                values.skill_level or 1
            ),
            impact_missile_id = definition.impact_missile_id or "",
            impact_dcc = definition.impact_dcc or "",
            impact_palette = definition.impact_palette or "",
            impact_lifetime_ticks = definition.impact_lifetime_ticks or 0,
            impact_directions = definition.impact_directions or 1,
            impact_frames_per_second = definition.impact_frames_per_second or 1,
            impact_loop = definition.impact_loop or false,
            impact_transparency_mode = definition.impact_transparency_mode or 0,
            impact_sound = definition.impact_sound or "",
            on_hit_state_id = values.on_hit_state_id or definition.on_hit_state_id or "",
            on_hit_state_source_id = values.on_hit_state_source_id
                or ("skill:" .. owner_id .. ":" .. tostring(definition.skill_id or 0) .. ":on-hit"),
            on_hit_state_duration = values.on_hit_state_duration or 0,
            on_hit_state_duration_policy = values.on_hit_state_duration_policy
                or definition.on_hit_state_duration_policy
                or "",
            on_hit_state_action_disabled = values.on_hit_state_action_disabled
                or definition.on_hit_state_action_disabled
                or false,
            on_hit_state_exclusive_group = values.on_hit_state_exclusive_group
                or definition.on_hit_state_exclusive_group
                or "",
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
            transparency_mode = definition.transparency_mode or 0,
            offset_x = definition.offset_x,
            offset_y = definition.offset_y,
            offset_z = definition.offset_z,
        },
    }
    local resident = population.resident_at(
        values.projectile_id or values.cast_id,
        location:get("level_id"),
        position:get("x"),
        position:get("y")
    )
    if resident then
        components["d2legacy.world.room_resident"] = resident
    end
    return components
end

function M.sound_cue(target, tick, kind, sound)
    if not sound or sound == "" then
        return nil
    end
    return {
        ["d2legacy.presentation.effect_cue"] = {
            kind = kind,
            tick = tick,
            target = target,
            overlay_id = "",
            sound = sound,
        },
    }
end

return M
