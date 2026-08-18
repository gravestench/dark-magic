-- Execute record-decoded deployed-device and weapon-trap transactions.
--
-- The systems below know reusable ECS shapes only. Definitions carry the
-- joined Skills/Missiles/MonStats/PetType facts, and ownership preserves kill
-- credit plus the shared five-trap replacement policy.

local ecs = require("engine.ecs/v1")
local geometry = require("d2legacy.policy.geometry")
local progression = require("d2legacy.policy.skill_progression")
local projectile_spawn = require("d2legacy.gameplay.projectile_spawn")
local monster_data = require("d2legacy.data.monster")
local spawn_monster = require("d2legacy.commands.spawn_monster")
local summon_policy = require("d2legacy.policy.summon")
local limits = require("d2legacy.owned_units.limits")
local damage = require("d2legacy.policy.damage")
local damage_bundle = require("d2legacy.policy.damage_bundle")

local M = {}

local function selectable_id(entity)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    if selectable then
        return selectable:get("id")
    end
    local identity = ecs.get(entity, "d2legacy.player.identity")
    return identity and ("player:" .. identity:get("player")) or ("entity:" .. entity:id())
end

local function learned_levels(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned then
            local owner = learned:get("owner"):id()
            result[owner] = result[owner] or { effective = {}, hard = {} }
            local id = learned:get("skill_id")
            result[owner].effective[id] = learned:get("level")
            local hard = ecs.get(entity, "d2legacy.player.skill_hard_level")
            result[owner].hard[id] = hard and hard:get("level") or learned:get("level")
        end
    end
    return result
end

local function owned_candidates(entities, owner)
    local result = {}
    for _, entity in ipairs(entities) do
        local relation = ecs.get(entity, "d2legacy.owned_unit")
        if relation and relation:get("owner"):id() == owner:id() then
            result[#result + 1] = {
                entity = entity,
                category = relation:get("category"),
                group = relation:get("group"),
                active = relation:get("active"),
                created_tick = relation:get("created_tick"),
            }
        end
    end
    return result
end

local function weapon_damage_range(source, definition, skill_level)
    local minimum, maximum = progression.damage_range(definition, skill_level, 0)
    local profile = ecs.get(source, "d2legacy.combat.melee_profile")
    if profile and (definition.weapon_fraction or 0) > 0 then
        minimum = minimum + math.floor(profile:get("physical_min") * definition.weapon_fraction / 128)
        maximum = maximum + math.floor(profile:get("physical_max") * definition.weapon_fraction / 128)
    end
    return minimum, maximum
end

local function projectile_components(source, definition, values)
    local minimum, maximum
    if (definition.weapon_fraction or 0) > 0 then
        minimum, maximum = weapon_damage_range(values.weapon_source or source, definition, values.skill_level)
    end
    return projectile_spawn.components(source, definition, {
        owner_id = values.owner_id,
        cast_id = values.cast_id,
        projectile_id = values.projectile_id,
        target_x = values.target_x,
        target_y = values.target_y,
        velocity_x = values.velocity_x,
        velocity_y = values.velocity_y,
        skill_level = values.skill_level,
        elemental_damage_percent = values.elemental_damage_percent,
        minimum_damage_raw = minimum,
        maximum_damage_raw = maximum,
    })
end

local function spawn_toward(source, definition, values, structural)
    local position = assert(ecs.get(source, "d2legacy.world.position"))
    local dx, dy = geometry.normalized_direction(position:get("x"), position:get("y"), values.target_x, values.target_y)
    values.velocity_x = dx * definition.speed_per_tick
    values.velocity_y = dy * definition.speed_per_tick
    structural:create(projectile_components(source, definition, values))
end

local function direct_projectile(context, caster, cast, definition, structural)
    local id = "trap-projectile:"
        .. selectable_id(caster)
        .. ":skill:"
        .. cast:get("skill_id")
        .. ":tick:"
        .. context.tick
    spawn_toward(caster, definition, {
        owner_id = selectable_id(caster),
        cast_id = id,
        projectile_id = id,
        target_x = cast:get("target_x"),
        target_y = cast:get("target_y"),
        skill_level = cast:get("skill_level"),
        elemental_damage_percent = cast:get("elemental_damage_percent"),
        weapon_source = caster,
    }, structural)
end

local function field_count(definition, level, hard)
    return definition.field_count_base
        + math.floor(level / definition.field_count_levels_per_add)
        + math.floor((hard[definition.field_count_synergy_skill_id] or 0) / definition.field_count_synergy_divisor)
end

local function deploy_fields(context, caster, cast, definition, hard, structural)
    local count = field_count(definition, cast:get("skill_level"), hard)
    local owner_id = selectable_id(caster)
    for instance = 1, count do
        local angle = (instance - 1) * 2 * math.pi / math.max(count, 1)
        local radius = instance == 1 and 0 or 0.75
        local x = cast:get("target_x") + math.cos(angle) * radius
        local y = cast:get("target_y") + math.sin(angle) * radius
        local id = "trap-field:"
            .. owner_id
            .. ":skill:"
            .. cast:get("skill_id")
            .. ":tick:"
            .. context.tick
            .. ":instance:"
            .. instance
        local components = projectile_components(caster, definition, {
            owner_id = owner_id,
            cast_id = id,
            projectile_id = id,
            target_x = x,
            target_y = y,
            velocity_x = 0,
            velocity_y = 0,
            skill_level = cast:get("skill_level"),
            elemental_damage_percent = cast:get("elemental_damage_percent"),
        })
        components["d2legacy.world.position"].x = x
        components["d2legacy.world.position"].y = y
        components["d2legacy.missile.projectile"].previous_x = x
        components["d2legacy.missile.projectile"].previous_y = y
        structural:create(components)
    end
end

local function deploy_returning_weapon(context, entities, caster, cast, definition, structural)
    local victims = limits.victims(owned_candidates(entities, caster), {
        id = definition.category,
        group = definition.category_group,
        base_max = definition.category_base_max,
        replacement = "replace_oldest",
    })
    for _, victim in ipairs(victims) do
        structural:destroy(victim.entity)
    end
    local owner_id = selectable_id(caster)
    local id = "trap-patrol:" .. owner_id .. ":skill:" .. cast:get("skill_id") .. ":tick:" .. context.tick
    local duration =
        progression.linear(definition.duration_base, definition.duration_per_level, cast:get("skill_level"))
    local components = projectile_components(caster, definition, {
        owner_id = owner_id,
        cast_id = id,
        projectile_id = id,
        target_x = cast:get("target_x"),
        target_y = cast:get("target_y"),
        velocity_x = 0,
        velocity_y = 0,
        skill_level = cast:get("skill_level"),
        elemental_damage_percent = cast:get("elemental_damage_percent"),
        weapon_source = caster,
    })
    local projectile = components["d2legacy.missile.projectile"]
    projectile.remaining_ticks = duration
    projectile.destroy_on_contact = false
    components["d2legacy.owned_unit"] = {
        owner = caster,
        owner_id = owner_id,
        ultimate_owner_id = owner_id,
        category = definition.category,
        group = definition.category_group,
        limit = definition.category_base_max,
        replacement = "replace_oldest",
        created_tick = context.tick,
        expires_tick = context.tick + duration,
        durable_id = "",
        durable = false,
        unsummon = false,
        warp_with_owner = false,
        range_limited = false,
        active = true,
        survives_owner_death = false,
    }
    components["d2legacy.trap.returning_weapon"] = {
        owner = caster,
        target_x = cast:get("target_x"),
        target_y = cast:get("target_y"),
        speed_per_tick = definition.speed_per_tick,
        outbound = true,
        expires_tick = context.tick + duration,
    }
    structural:create(components)
end

local function sentry_components(context, caster, cast, definition, levels)
    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
    local location = assert(ecs.get(caster, "d2legacy.world.location"))
    local progress = assert(ecs.get(caster, "d2legacy.player.progress"))
    local summon_id = "trap:" .. identity:get("player") .. ":" .. definition.skill_id .. ":" .. context.tick
    local monster = monster_data.summon(
        definition.monster_id,
        summon_policy.pet_level(cast:get("skill_level"), progress:get("level"))
    )
    local components = spawn_monster.components({
        tick = context.tick,
        payload = {
            spawn_id = summon_id,
            x = cast:get("target_x"),
            y = cast:get("target_y"),
            act = location:get("act"),
            level_id = location:get("level_id"),
            seed = summon_id,
            definition = monster,
        },
    }, 1)
    local hard = levels.hard or {}
    local extra_shots = definition.shot_synergy_skill_id > 0
            and math.floor((hard[definition.shot_synergy_skill_id] or 0) / definition.shot_synergy_divisor)
        or 0
    components["d2legacy.owned_unit"] = {
        owner = caster,
        owner_id = "player:" .. identity:get("player"),
        ultimate_owner_id = "player:" .. identity:get("player"),
        category = definition.category,
        group = definition.category_group,
        limit = definition.category_base_max,
        replacement = "replace_oldest",
        created_tick = context.tick,
        expires_tick = 0,
        durable_id = "",
        durable = false,
        unsummon = false,
        warp_with_owner = false,
        range_limited = definition.range_limited,
        active = true,
        survives_owner_death = false,
    }
    components["d2legacy.world.stationary"] = {}
    components["d2legacy.trap.autonomous"] = {
        owner = caster,
        owner_id = "player:" .. identity:get("player"),
        skill_id = definition.skill_id,
        skill_level = cast:get("skill_level"),
        shots_remaining = definition.shots_base + extra_shots,
        next_fire_tick = context.tick + definition.fire_interval,
        fire_interval = definition.fire_interval,
        target_radius = definition.target_radius,
        cast_serial = 0,
    }
    components["d2legacy.monster.granted_skill"] = {
        skill = tostring(definition.helper_skill_id),
        level = cast:get("skill_level"),
    }
    components["d2legacy.world.selectable"].owner = identity:get("player")
    components["d2legacy.world.occupancy"].blocks_movement = false
    return components
end

local function deploy_sentry(context, entities, caster, cast, definition, levels, structural)
    local victims = limits.victims(owned_candidates(entities, caster), {
        id = definition.category,
        group = definition.category_group,
        base_max = definition.category_base_max,
        replacement = "replace_oldest",
    })
    for _, victim in ipairs(victims) do
        structural:destroy(victim.entity)
    end
    structural:create(sentry_components(context, caster, cast, definition, levels))
end

local function apply_periodic_state(context, caster, cast, definition, structural)
    local source_id = "skill:" .. selectable_id(caster) .. ":" .. definition.skill_id
    local duration =
        progression.linear(definition.duration_base, definition.duration_per_level, cast:get("skill_level"))
    structural:create({
        ["d2legacy.state.request"] = {
            operation = "apply",
            target = caster,
            state_id = definition.state_id,
            source_id = source_id,
            duration = duration,
            policy = "refresh_same_source",
        },
    })
    structural:set(caster, "d2legacy.combat.periodic_weapon", {
        owner = caster,
        owner_id = selectable_id(caster),
        skill_id = definition.skill_id,
        skill_level = cast:get("skill_level"),
        source_id = source_id,
        radius = definition.radius,
        weapon_fraction = definition.weapon_fraction,
        next_tick = context.tick + definition.period_ticks,
        period_ticks = definition.period_ticks,
        expires_tick = context.tick + duration,
    })
end

local function emit_cast_effect(context, entities, caster, cast, definition, levels, structural)
    local shape = definition.shape
    if shape == "lobbed_payload" or shape == "repeat_weapon_missile" then
        direct_projectile(context, caster, cast, definition, structural)
    elseif shape == "returning_weapon_patrol" then
        deploy_returning_weapon(context, entities, caster, cast, definition, structural)
    elseif shape == "persistent_field" then
        deploy_fields(context, caster, cast, definition, levels.hard or {}, structural)
    elseif shape == "stationary_sentry" then
        deploy_sentry(context, entities, caster, cast, definition, levels, structural)
    elseif shape == "periodic_weapon_state" then
        apply_periodic_state(context, caster, cast, definition, structural)
    else
        error("unsupported trap-family runtime shape " .. tostring(shape))
    end
    cast:set("effect_emitted", true)
end

local function update_returning_weapons(context, entities, structural)
    for _, entity in ipairs(entities) do
        local returning = ecs.get(entity, "d2legacy.trap.returning_weapon")
        local projectile = ecs.get(entity, "d2legacy.missile.projectile")
        local position = ecs.get(entity, "d2legacy.world.position")
        if returning and projectile and position then
            local owner_position = ecs.get(returning:get("owner"), "d2legacy.world.position")
            if not owner_position or context.tick >= returning:get("expires_tick") then
                structural:destroy(entity)
            else
                local target_x = returning:get("outbound") and returning:get("target_x") or owner_position:get("x")
                local target_y = returning:get("outbound") and returning:get("target_y") or owner_position:get("y")
                local dx, dy = target_x - position:get("x"), target_y - position:get("y")
                local distance = math.sqrt(dx * dx + dy * dy)
                local speed = returning:get("speed_per_tick")
                if distance <= speed then
                    if returning:get("outbound") then
                        returning:set("outbound", false)
                        target_x, target_y = owner_position:get("x"), owner_position:get("y")
                        dx, dy = target_x - position:get("x"), target_y - position:get("y")
                        distance = math.sqrt(dx * dx + dy * dy)
                    else
                        structural:destroy(entity)
                        distance = 0
                    end
                end
                if distance > 0 then
                    projectile:set("velocity_x", dx / distance * speed)
                    projectile:set("velocity_y", dy / distance * speed)
                end
            end
        end
    end
end

local function target_candidates_at(entities, source, center_x, center_y, radius)
    local result = {}
    local position = ecs.get(source, "d2legacy.world.position")
    local location = ecs.get(source, "d2legacy.world.location")
    if not position or not location then
        return result
    end
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        local target_position = ecs.get(entity, "d2legacy.world.position")
        local target_location = ecs.get(entity, "d2legacy.world.location")
        local stats = ecs.get(entity, "d2legacy.monster.stats")
        if
            entity:id() ~= source:id()
            and selectable
            and selectable:get("kind") == "hostile"
            and stats
            and stats:get("health") > 0
            and target_position
            and target_location
            and target_location:get("act") == location:get("act")
            and target_location:get("level_id") == location:get("level_id")
        then
            local dx = target_position:get("x") - center_x
            local dy = target_position:get("y") - center_y
            local distance = math.sqrt(dx * dx + dy * dy)
            if distance <= radius then
                result[#result + 1] = { entity = entity, distance = distance, id = selectable:get("id") }
            end
        end
    end
    table.sort(result, function(left, right)
        return left.distance < right.distance or left.distance == right.distance and left.id < right.id
    end)
    return result
end

local function target_candidates(entities, source, radius)
    local position = ecs.get(source, "d2legacy.world.position")
    if not position then
        return {}
    end
    return target_candidates_at(entities, source, position:get("x"), position:get("y"), radius)
end

local function corpse_candidates(entities, source, radius)
    local result = {}
    local position = ecs.get(source, "d2legacy.world.position")
    local location = ecs.get(source, "d2legacy.world.location")
    for _, entity in ipairs(entities) do
        local death = ecs.get(entity, "d2legacy.monster.death")
        local target_position = ecs.get(entity, "d2legacy.world.position")
        local target_location = ecs.get(entity, "d2legacy.world.location")
        if
            death
            and death:get("corpse_usable")
            and not death:get("active")
            and target_position
            and target_location
            and target_location:get("act") == location:get("act")
            and target_location:get("level_id") == location:get("level_id")
        then
            local dx = target_position:get("x") - position:get("x")
            local dy = target_position:get("y") - position:get("y")
            local distance = math.sqrt(dx * dx + dy * dy)
            if distance <= radius then
                result[#result + 1] =
                    { entity = entity, distance = distance, x = target_position:get("x"), y = target_position:get("y") }
            end
        end
    end
    table.sort(result, function(left, right)
        return left.distance < right.distance
            or left.distance == right.distance and left.entity:id() < right.entity:id()
    end)
    return result
end

local function create_damage_event(structural, context, owner_id, target_id, source_kind, resolved)
    structural:create({
        ["d2legacy.combat.event"] = {
            kind = resolved.lethal and "unit_died" or "damage_applied",
            tick = context.tick,
            attacker_id = owner_id,
            target_id = target_id,
            source_kind = source_kind,
            damage_channel = resolved.channel,
            rolled_damage_raw = resolved.rolled_damage_raw,
            damage_raw = resolved.damage_raw,
            remaining_health_raw = resolved.remaining_health_raw,
        },
        ["d2legacy.combat.damage_bundle"] = damage_bundle.stage_component(resolved.rolled, resolved.mitigated),
    })
end

local function explode_corpse(context, entities, trap, autonomous, definition, corpse, structural)
    local corpse_stats = assert(ecs.get(corpse.entity, "d2legacy.monster.stats"))
    local minimum = math.floor(corpse_stats:get("max_health") * definition.corpse_minimum_percent / 100)
    local maximum = math.floor(corpse_stats:get("max_health") * definition.corpse_maximum_percent / 100)
    local amount = damage.roll(minimum, maximum)
    local fire = math.floor(amount * definition.corpse_fire_percent / 100)
    local physical = amount - fire
    local radius = progression.linear(
        definition.corpse_radius_base,
        definition.corpse_radius_per_level,
        autonomous:get("skill_level")
    )
    structural:destroy(corpse.entity)
    for _, candidate in ipairs(target_candidates_at(entities, trap, corpse.x, corpse.y, radius)) do
        local resolved = damage.resolve(candidate.entity, { physical = physical, fire = fire }, ecs)
        create_damage_event(structural, context, autonomous:get("owner_id"), candidate.id, "trap_corpse", resolved)
    end
end

local function burst_count(definition, autonomous, hard)
    local count =
        progression.linear(definition.burst_count_base, definition.burst_count_per_level, autonomous:get("skill_level"))
    if definition.burst_synergy_skill_id > 0 then
        count = count + math.floor((hard[definition.burst_synergy_skill_id] or 0) / definition.burst_synergy_divisor)
    end
    return math.max(count, 1)
end

local function fire_sentry(context, trap, autonomous, definition, target, hard, structural)
    local source_position = ecs.get(trap, "d2legacy.world.position")
    local target_position = ecs.get(target.entity, "d2legacy.world.position")
    local base_angle = math.atan(
        target_position:get("y") - source_position:get("y"),
        target_position:get("x") - source_position:get("x")
    )
    local count = burst_count(definition, autonomous, hard)
    local serial = autonomous:get("cast_serial") + 1
    autonomous:set("cast_serial", serial)
    for instance = 1, count do
        local spread = count == 1 and 0 or (instance - (count + 1) / 2) * math.pi / 36
        local angle = base_angle + spread
        local id = "trap-shot:" .. autonomous:get("owner_id") .. ":" .. trap:id() .. ":" .. serial .. ":" .. instance
        structural:create(projectile_components(trap, definition, {
            owner_id = autonomous:get("owner_id"),
            cast_id = id,
            projectile_id = id,
            target_x = target_position:get("x"),
            target_y = target_position:get("y"),
            velocity_x = math.cos(angle) * definition.speed_per_tick,
            velocity_y = math.sin(angle) * definition.speed_per_tick,
            skill_level = autonomous:get("skill_level"),
            elemental_damage_percent = 0,
        }))
    end
end

local function update_sentries(context, entities, definitions, levels, structural)
    for _, trap in ipairs(entities) do
        local autonomous = ecs.get(trap, "d2legacy.trap.autonomous")
        if autonomous and context.tick >= autonomous:get("next_fire_tick") then
            if autonomous:get("shots_remaining") <= 0 then
                structural:destroy(trap)
            else
                local definition = assert(definitions[autonomous:get("skill_id")])
                local owner_levels = levels[autonomous:get("owner"):id()] or { hard = {} }
                local fired = false
                if definition.operation == "corpse_or_projectile" then
                    local radius = progression.linear(
                        definition.corpse_radius_base,
                        definition.corpse_radius_per_level,
                        autonomous:get("skill_level")
                    )
                    local corpses = corpse_candidates(entities, trap, radius)
                    if corpses[1] then
                        explode_corpse(context, entities, trap, autonomous, definition, corpses[1], structural)
                        fired = true
                    end
                end
                if not fired then
                    local targets = target_candidates(entities, trap, autonomous:get("target_radius"))
                    if targets[1] then
                        fire_sentry(
                            context,
                            trap,
                            autonomous,
                            definition,
                            targets[1],
                            owner_levels.hard or {},
                            structural
                        )
                        fired = true
                    end
                end
                if fired then
                    autonomous:set("shots_remaining", autonomous:get("shots_remaining") - 1)
                end
                autonomous:set("next_fire_tick", context.tick + autonomous:get("fire_interval"))
            end
        end
    end
end

local function update_periodic_weapons(context, entities, definitions, structural)
    for _, source in ipairs(entities) do
        local periodic = ecs.get(source, "d2legacy.combat.periodic_weapon")
        if periodic then
            if context.tick >= periodic:get("expires_tick") then
                structural:remove(source, "d2legacy.combat.periodic_weapon")
            elseif context.tick >= periodic:get("next_tick") then
                local definition = assert(definitions[periodic:get("skill_id")])
                local minimum, maximum = weapon_damage_range(source, definition, periodic:get("skill_level"))
                for _, candidate in ipairs(target_candidates(entities, source, periodic:get("radius"))) do
                    local amount = damage.roll(minimum, maximum)
                    local resolved = damage.resolve(candidate.entity, damage_bundle.single("physical", amount), ecs)
                    create_damage_event(
                        structural,
                        context,
                        periodic:get("owner_id"),
                        candidate.id,
                        "periodic_weapon",
                        resolved
                    )
                end
                periodic:set("next_tick", context.tick + periodic:get("period_ticks"))
            end
        end
    end
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.trap.returning_weapon",
        phase = "pre_simulation",
        query = {
            all = { "d2legacy.trap.returning_weapon", "d2legacy.missile.projectile", "d2legacy.world.position" },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.trap.returning_weapon",
            "d2legacy.missile.projectile",
            "d2legacy.world.position",
        },
        write = {
            "d2legacy.trap.returning_weapon",
            "d2legacy.missile.projectile",
            "d2legacy.world.position",
        },
        update = function(context, entities, structural)
            update_returning_weapons(context, entities, structural)
        end,
    })

    ecs.system({
        id = "d2legacy.trap.deploy",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            any = {
                "d2legacy.skill.cast",
                "d2legacy.player.learned_skill",
                "d2legacy.player.skill_hard_level",
                "d2legacy.owned_unit",
            },
            none = { "d2legacy.player.death", "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.player.learned_skill",
            "d2legacy.player.skill_hard_level",
            "d2legacy.player.identity",
            "d2legacy.player.progress",
            "d2legacy.combat.melee_profile",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.selectable",
            "d2legacy.owned_unit",
        },
        write = {
            "d2legacy.skill.cast",
            "d2legacy.missile.projectile",
            "d2legacy.missile.effect",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.velocity",
            "d2legacy.world.facing",
            "d2legacy.world.collider",
            "d2legacy.world.occupancy",
            "d2legacy.world.selectable",
            "d2legacy.world.stationary",
            "d2legacy.world.room_resident",
            "engine.world.velocity_mover",
            "d2legacy.monster.identity",
            "d2legacy.monster.stats",
            "d2legacy.monster.appearance",
            "d2legacy.monster.ai",
            "d2legacy.monster.granted_skill",
            "d2legacy.combat.melee_profile",
            "d2legacy.combat.knockback_target",
            "d2legacy.owned_unit",
            "d2legacy.trap.autonomous",
            "d2legacy.trap.returning_weapon",
            "d2legacy.combat.periodic_weapon",
            "d2legacy.state.request",
        },
        update = function(context, entities, structural)
            local levels = learned_levels(entities)
            for _, caster in ipairs(entities) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = cast and definitions[cast:get("skill_id")] or nil
                if definition and not cast:get("effect_emitted") and context.tick >= cast:get("effect_tick") then
                    emit_cast_effect(
                        context,
                        entities,
                        caster,
                        cast,
                        definition,
                        levels[caster:id()] or { effective = {}, hard = {} },
                        structural
                    )
                end
            end
        end,
    })

    ecs.system({
        id = "d2legacy.trap.autonomous_attack",
        phase = "combat",
        query = {
            any = {
                "d2legacy.trap.autonomous",
                "d2legacy.world.selectable",
                "d2legacy.monster.death",
                "d2legacy.player.learned_skill",
                "d2legacy.player.skill_hard_level",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.trap.autonomous",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "d2legacy.monster.stats",
            "d2legacy.monster.death",
            "d2legacy.player.learned_skill",
            "d2legacy.player.skill_hard_level",
            "d2legacy.combat.defense",
        },
        write = {
            "d2legacy.trap.autonomous",
            "d2legacy.missile.projectile",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.room_resident",
            "d2legacy.monster.stats",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
        },
        update = function(context, entities, structural)
            update_sentries(context, entities, definitions, learned_levels(entities), structural)
        end,
    })

    ecs.system({
        id = "d2legacy.combat.periodic_weapon",
        phase = "combat",
        query = {
            any = { "d2legacy.combat.periodic_weapon", "d2legacy.world.selectable" },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.periodic_weapon",
            "d2legacy.combat.melee_profile",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.monster.stats",
            "d2legacy.combat.defense",
        },
        write = {
            "d2legacy.combat.periodic_weapon",
            "d2legacy.monster.stats",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
        },
        update = function(context, entities, structural)
            update_periodic_weapons(context, entities, definitions, structural)
        end,
    })
end

return M
