-- Move projectiles, resolve first contact, and expire them.
--
-- Movement and contact are separate fixed-tick systems. Movement records the
-- old point, then contact checks the entire swept line. This prevents a fast
-- missile from jumping over a target between two 25 Hz simulation ticks.

local ecs = require("engine.ecs/v1")
local damage = require("d2legacy.policy.damage")
local damage_bundle = require("d2legacy.policy.damage_bundle")
local geometry = require("d2legacy.policy.geometry")
local population = require("d2legacy.bootstrap.population")
local cold_duration = require("d2legacy.policy.cold_duration")
local game_rules = require("d2legacy.policy.game_rules")

local M = {}

local function move_projectiles(projectiles)
    for _, entity in ipairs(projectiles) do
        local projectile = ecs.get(entity, "d2legacy.missile.projectile")
        local position = ecs.get(entity, "d2legacy.world.position")
        local x, y = position:get("x"), position:get("y")
        projectile:set("previous_x", x)
        projectile:set("previous_y", y)
        position:set("x", x + projectile:get("velocity_x"))
        position:set("y", y + projectile:get("velocity_y"))
        projectile:set("remaining_ticks", projectile:get("remaining_ticks") - 1)
    end
end

local function target_snapshot(entity)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    local position = ecs.get(entity, "d2legacy.world.position")
    local location = ecs.get(entity, "d2legacy.world.location")
    if not selectable or not position or not location then
        return nil
    end
    if not ecs.get(entity, "d2legacy.monster.stats") and not ecs.get(entity, "d2legacy.player.vitals") then
        return nil
    end
    local collider = ecs.get(entity, "d2legacy.world.collider")
    return {
        entity = entity,
        id = selectable:get("id"),
        x = position:get("x"),
        y = position:get("y"),
        act = location:get("act"),
        level_id = location:get("level_id"),
        radius = collider and collider:get("radius") or selectable:get("radius"),
    }
end

local function contact_key(projectile, target_id)
    return projectile:get("cast_id") .. "\0" .. target_id
end

local function first_contact(projectile_entity, projectile, position, location, targets, locks)
    local best, best_along
    for _, target in ipairs(targets) do
        if
            target.entity:id() ~= projectile_entity:id()
            and target.id ~= projectile:get("owner_id")
            and target.act == location:get("act")
            and target.level_id == location:get("level_id")
            and not locks[contact_key(projectile, target.id)]
        then
            local distance, along = geometry.segment_distance(
                projectile:get("previous_x"),
                projectile:get("previous_y"),
                position:get("x"),
                position:get("y"),
                target.x,
                target.y
            )
            local touching = distance <= projectile:get("collision_radius") + target.radius
            if touching and (not best or along < best_along or (along == best_along and target.id < best.id)) then
                best, best_along = target, along
            end
        end
    end
    return best, best_along
end

local function emit_hit(context, projectile, target, result, structural)
    structural:create({
        ["d2legacy.combat.event"] = {
            kind = result.lethal and "unit_died" or "damage_applied",
            tick = context.tick,
            attacker_id = projectile:get("owner_id"),
            target_id = target.id,
            source_kind = "missile",
            damage_channel = result.channel,
            rolled_damage_raw = result.rolled_damage_raw,
            damage_raw = result.damage_raw,
            remaining_health_raw = result.remaining_health_raw,
        },
        ["d2legacy.combat.damage_bundle"] = damage_bundle.stage_component(result.rolled, result.mitigated),
    })
end

local function apply_hit(context, projectile, target, structural)
    local amount = damage.roll(projectile:get("minimum_damage_raw"), projectile:get("maximum_damage_raw"))
    local bundle = damage_bundle.single(projectile:get("damage_channel"), amount)
    local result = damage.resolve(target.entity, bundle, ecs)
    emit_hit(context, projectile, target, result, structural)
    if projectile:get("hit_sound") ~= "" then
        structural:create({
            ["d2legacy.presentation.effect_cue"] = {
                kind = "missile_hit",
                tick = context.tick,
                target = target.entity,
                overlay_id = "",
                sound = projectile:get("hit_sound"),
            },
        })
    end
    return result
end

local function apply_on_hit_state(projectile, target, damage_result, structural)
    local state_id = projectile:get("on_hit_state_id")
    if state_id == "" or damage_result.lethal then
        return
    end
    local duration = projectile:get("on_hit_state_duration")
    local policy = projectile:get("on_hit_state_duration_policy")
    if policy == "monster_cold" then
        if not ecs.get(target.entity, "d2legacy.monster.stats") then
            return
        end
        duration = cold_duration.target_frames(duration, target.entity, game_rules.difficulty(), ecs)
    elseif policy == "cold" then
        duration = cold_duration.target_frames(duration, target.entity, game_rules.difficulty(), ecs)
    else
        assert(policy == "", "unsupported on-hit state duration policy")
    end
    if duration <= 0 then
        return
    end
    structural:create({
        ["d2legacy.state.request"] = {
            operation = "apply",
            target = target.entity,
            state_id = state_id,
            source_id = projectile:get("on_hit_state_source_id"),
            duration = duration,
            policy = "refresh_same_source",
            exclusive_group = projectile:get("on_hit_state_exclusive_group"),
            action_disabled = projectile:get("on_hit_state_action_disabled"),
        },
    })
end

local function impact_targets(projectile, location, targets, x, y)
    local result = {}
    for _, target in ipairs(targets) do
        if
            target.id ~= projectile:get("owner_id")
            and target.act == location:get("act")
            and target.level_id == location:get("level_id")
        then
            local dx, dy = target.x - x, target.y - y
            local radius = projectile:get("impact_radius") + target.radius
            if dx * dx + dy * dy <= radius * radius then
                result[#result + 1] = target
            end
        end
    end
    return result
end

local function emit_impact(context, projectile, target, location, x, y, structural)
    local effect_id = projectile:get("cast_id") .. ":impact:" .. target.id .. ":tick:" .. context.tick
    local components = {
        ["d2legacy.world.position"] = { x = x, y = y },
        ["d2legacy.world.location"] = location:snapshot(),
        ["d2legacy.missile.effect"] = {
            owner_id = projectile:get("owner_id"),
            cast_id = projectile:get("cast_id"),
            remaining_ticks = projectile:get("impact_lifetime_ticks"),
            missile_id = projectile:get("impact_missile_id"),
            dcc = projectile:get("impact_dcc"),
            palette = projectile:get("impact_palette"),
            travel_sound = projectile:get("impact_sound"),
            hit_sound = projectile:get("hit_sound"),
            directions = projectile:get("impact_directions"),
            logical_direction = 0,
            frames_per_second = projectile:get("impact_frames_per_second"),
            loop = projectile:get("impact_loop"),
            transparency_mode = projectile:get("impact_transparency_mode"),
            offset_x = 0,
            offset_y = 0,
            offset_z = 0,
        },
    }
    local resident = population.resident_at(effect_id, location:get("level_id"), x, y)
    if resident then
        components["d2legacy.world.room_resident"] = resident
    end
    structural:create(components)
end

local function age_effects(effects, structural)
    for _, entity in ipairs(effects) do
        local effect = ecs.get(entity, "d2legacy.missile.effect")
        effect:set("remaining_ticks", effect:get("remaining_ticks") - 1)
        if effect:get("remaining_ticks") <= 0 then
            structural:destroy(entity)
        end
    end
end

local function resolve_contacts(context, entities, structural)
    local targets = {}
    local locks = {}
    for _, entity in ipairs(entities) do
        local target = target_snapshot(entity)
        if target then
            targets[#targets + 1] = target
        end
        local lock = ecs.get(entity, "d2legacy.missile.contact_lock")
        if lock then
            if context.tick >= lock:get("expires_tick") then
                structural:destroy(entity)
            else
                locks[lock:get("cast_id") .. "\0" .. lock:get("target_id")] = true
            end
        end
    end
    table.sort(targets, function(a, b)
        return a.id < b.id
    end)

    for _, entity in ipairs(entities) do
        local projectile = ecs.get(entity, "d2legacy.missile.projectile")
        if projectile then
            local position = ecs.get(entity, "d2legacy.world.position")
            local location = ecs.get(entity, "d2legacy.world.location")
            local target, along = first_contact(entity, projectile, position, location, targets, locks)
            if target then
                local previous_x, previous_y = projectile:get("previous_x"), projectile:get("previous_y")
                local impact_x = previous_x + (position:get("x") - previous_x) * along
                local impact_y = previous_y + (position:get("y") - previous_y) * along
                if projectile:get("impact_radius") > 0 then
                    for _, area_target in ipairs(impact_targets(projectile, location, targets, impact_x, impact_y)) do
                        local damage_result = apply_hit(context, projectile, area_target, structural)
                        apply_on_hit_state(projectile, area_target, damage_result, structural)
                    end
                else
                    local damage_result = apply_hit(context, projectile, target, structural)
                    apply_on_hit_state(projectile, target, damage_result, structural)
                end
                if projectile:get("impact_missile_id") ~= "" then
                    emit_impact(context, projectile, target, location, impact_x, impact_y, structural)
                end
                local delay = projectile:get("next_hit_delay")
                if delay > 0 then
                    local key = contact_key(projectile, target.id)
                    locks[key] = true
                    structural:create({
                        ["d2legacy.missile.contact_lock"] = {
                            cast_id = projectile:get("cast_id"),
                            target_id = target.id,
                            expires_tick = context.tick + delay,
                        },
                    })
                end
                if projectile:get("destroy_on_contact") then
                    structural:destroy(entity)
                end
            elseif projectile:get("remaining_ticks") <= 0 then
                structural:destroy(entity)
            end
        end
    end
end

function M.register()
    ecs.system({
        id = "d2legacy.missile.move",
        phase = "movement",
        query = {
            all = { "d2legacy.missile.projectile", "d2legacy.world.position" },
            none = { "d2legacy.world.inactive" },
        },
        read = { "d2legacy.missile.projectile", "d2legacy.world.position" },
        write = { "d2legacy.missile.projectile", "d2legacy.world.position" },
        update = function(_, entities)
            move_projectiles(entities)
        end,
    })

    ecs.system({
        id = "d2legacy.missile.contact",
        phase = "combat",
        query = {
            any = {
                "d2legacy.missile.projectile",
                "d2legacy.missile.contact_lock",
                "d2legacy.world.selectable",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.missile.projectile",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.selectable",
            "d2legacy.world.collider",
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.combat.defense",
            "d2legacy.missile.contact_lock",
        },
        write = {
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
            "d2legacy.missile.contact_lock",
            "d2legacy.missile.effect",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.room_resident",
            "d2legacy.state.request",
            "d2legacy.presentation.effect_cue",
        },
        update = resolve_contacts,
    })

    ecs.system({
        id = "d2legacy.missile.effect_lifetime",
        phase = "movement",
        query = {
            all = { "d2legacy.missile.effect" },
            none = { "d2legacy.world.inactive" },
        },
        read = { "d2legacy.missile.effect" },
        write = { "d2legacy.missile.effect" },
        update = function(_, effects, structural)
            age_effects(effects, structural)
        end,
    })
end

return M
