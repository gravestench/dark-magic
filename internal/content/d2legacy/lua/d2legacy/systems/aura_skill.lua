-- Reconcile right-selected party auras into explicit ECS relationships.
--
-- Assignment owns activation; right-button use never casts the aura. Each
-- winning target relationship is co-composed with its stat source so losing
-- range, party membership, life, level, or selection removes the modifier by
-- destroying one entity. Multiple different state IDs compose, while one
-- deterministic highest-level source wins for an overlapping state ID.

local ecs = require("engine.ecs/v1")
local party = require("d2legacy.policy.party")
local progression = require("d2legacy.policy.skill_progression")
local M = {}

local function player_maps(entities)
    local by_id, levels = {}, {}
    for _, entity in ipairs(entities) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity then
            by_id[identity:get("player")] = entity
        end
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned then
            local owner = learned:get("owner"):id()
            levels[owner] = levels[owner] or {}
            levels[owner][learned:get("skill_id")] = learned:get("level")
        end
    end
    return by_id, levels
end

local function emitter_values(context, player, definition, level, player_id, current)
    local activated = current and current:get("skill_id") == definition.skill_id and current:get("activated_tick")
        or context.tick
    return {
        source_id = "aura:" .. player_id .. ":" .. definition.skill_id,
        skill_id = definition.skill_id,
        skill_level = level,
        state_id = definition.state_id,
        target_state_id = definition.target_state_id,
        radius = progression.linear(definition.radius_base, definition.radius_per_level, level),
        stat = definition.stat,
        operation = definition.stat_operation,
        value = progression.linear(definition.stat_value_base, definition.stat_value_per_level, level),
        refresh_delay = definition.record_refresh_delay,
        activated_tick = activated,
        entity = player,
        player_id = player_id,
    }
end

local function emitter_changed(current, values)
    return not current
        or current:get("source_id") ~= values.source_id
        or current:get("skill_id") ~= values.skill_id
        or current:get("skill_level") ~= values.skill_level
        or current:get("state_id") ~= values.state_id
        or current:get("target_state_id") ~= values.target_state_id
        or current:get("radius") ~= values.radius
        or current:get("stat") ~= values.stat
        or current:get("operation") ~= values.operation
        or current:get("value") ~= values.value
        or current:get("refresh_delay") ~= values.refresh_delay
end

local function desired_emitters(context, entities, definitions, levels, structural)
    local result = {}
    for _, player in ipairs(entities) do
        local identity = ecs.get(player, "d2legacy.player.identity")
        if identity then
            local current = ecs.get(player, "d2legacy.skill.aura_emitter")
            local assignment = ecs.get(player, "d2legacy.player.skill_assignment")
            local definition = assignment and definitions[assignment:get("right")] or nil
            local level = definition and (levels[player:id()] or {})[definition.skill_id] or 0
            local vitals = ecs.get(player, "d2legacy.player.vitals")
            local active = definition
                and definition.activation == "selected_right"
                and level > 0
                and vitals
                and vitals:get("health") > 0
                and not ecs.get(player, "d2legacy.player.death")
                and not ecs.get(player, "d2legacy.world.inactive")
            if active then
                local values = emitter_values(context, player, definition, level, identity:get("player"), current)
                result[#result + 1] = values
                if emitter_changed(current, values) then
                    structural:set(player, "d2legacy.skill.aura_emitter", {
                        source_id = values.source_id,
                        skill_id = values.skill_id,
                        skill_level = values.skill_level,
                        state_id = values.state_id,
                        target_state_id = values.target_state_id,
                        radius = values.radius,
                        stat = values.stat,
                        operation = values.operation,
                        value = values.value,
                        refresh_delay = values.refresh_delay,
                        activated_tick = values.activated_tick,
                    })
                end
            elseif current then
                structural:remove(player, "d2legacy.skill.aura_emitter")
            end
        end
    end
    table.sort(result, function(left, right)
        return left.source_id < right.source_id
    end)
    return result
end

local function in_radius(source, target, radius)
    local source_position = ecs.get(source, "d2legacy.world.position")
    local target_position = ecs.get(target, "d2legacy.world.position")
    if not source_position or not target_position then
        return false
    end
    local dx = target_position:get("x") - source_position:get("x")
    local dy = target_position:get("y") - source_position:get("y")
    return dx * dx + dy * dy <= radius * radius
end

local function candidate_wins(candidate, current)
    if not current or candidate.skill_level ~= current.skill_level then
        return not current or candidate.skill_level > current.skill_level
    end
    if candidate.value ~= current.value then
        return candidate.value > current.value
    end
    return candidate.source_id < current.source_id
end

local function desired_effects(entities, emitters, players)
    local result = {}
    for _, emitter in ipairs(emitters) do
        for _, player_id in ipairs(party.living_members_in_same_level(emitter.player_id, entities)) do
            local target = players[player_id]
            if
                target
                and not ecs.get(target, "d2legacy.world.inactive")
                and in_radius(emitter.entity, target, emitter.radius)
            then
                local key = target:id() .. ":" .. emitter.target_state_id
                local candidate = {
                    key = key,
                    emitter = emitter.entity,
                    target = target,
                    source_id = emitter.source_id,
                    skill_id = emitter.skill_id,
                    skill_level = emitter.skill_level,
                    state_id = emitter.target_state_id,
                    stat = emitter.stat,
                    operation = emitter.operation,
                    value = emitter.value,
                    refresh_delay = emitter.refresh_delay,
                }
                if candidate_wins(candidate, result[key]) then
                    result[key] = candidate
                end
            end
        end
    end
    return result
end

local function effect_key(effect)
    return effect:get("target"):id() .. ":" .. effect:get("state_id")
end

local function sorted_keys(values)
    local result = {}
    for key in pairs(values) do
        result[#result + 1] = key
    end
    table.sort(result)
    return result
end

local function state_event(structural, kind, tick, value, reason)
    structural:create({
        ["d2legacy.state.event"] = {
            kind = kind,
            tick = tick,
            target = value.target,
            state_id = value.state_id,
            source_id = value.source_id,
            expires_tick = 0,
            reason = reason,
        },
    })
end

local function create_effect(structural, context, value)
    structural:create({
        ["d2legacy.skill.aura_effect"] = {
            emitter = value.emitter,
            target = value.target,
            source_id = value.source_id,
            skill_id = value.skill_id,
            skill_level = value.skill_level,
            state_id = value.state_id,
            refresh_delay = value.refresh_delay,
        },
        ["d2legacy.stat.source"] = {
            target = value.target,
            source_id = value.source_id,
            stat = value.stat,
            operation = value.operation,
            value = value.value,
            order = 250,
        },
    })
    state_event(structural, "state_applied", context.tick, value, "aura_entered")
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.skill.selected_party_aura",
        phase = "pre_simulation",
        query = {
            any = {
                "d2legacy.player.identity",
                "d2legacy.player.learned_skill",
                "d2legacy.skill.aura_emitter",
                "d2legacy.skill.aura_effect",
            },
        },
        read = {
            "d2legacy.player.identity",
            "d2legacy.player.learned_skill",
            "d2legacy.player.skill_assignment",
            "d2legacy.player.vitals",
            "d2legacy.player.death",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.inactive",
            "d2legacy.skill.aura_emitter",
            "d2legacy.skill.aura_effect",
            "d2legacy.stat.source",
        },
        write = {
            "d2legacy.skill.aura_emitter",
            "d2legacy.skill.aura_effect",
            "d2legacy.stat.source",
            "d2legacy.state.event",
        },
        update = function(context, entities, structural)
            local players, levels = player_maps(entities)
            local emitters = desired_emitters(context, entities, definitions, levels, structural)
            local desired = desired_effects(entities, emitters, players)
            local existing, duplicate = {}, {}
            for _, entity in ipairs(entities) do
                local effect = ecs.get(entity, "d2legacy.skill.aura_effect")
                if effect then
                    local key = effect_key(effect)
                    if existing[key] then
                        duplicate[#duplicate + 1] = entity
                    else
                        existing[key] = entity
                    end
                end
            end
            for _, entity in ipairs(duplicate) do
                local effect = ecs.get(entity, "d2legacy.skill.aura_effect")
                state_event(structural, "state_removed", context.tick, {
                    target = effect:get("target"),
                    state_id = effect:get("state_id"),
                    source_id = effect:get("source_id"),
                }, "aura_reconciled")
                structural:destroy(entity)
            end
            for _, key in ipairs(sorted_keys(desired)) do
                local value = desired[key]
                local entity = existing[key]
                local effect = entity and ecs.get(entity, "d2legacy.skill.aura_effect") or nil
                if effect and effect:get("source_id") == value.source_id then
                    effect:set("emitter", value.emitter)
                    effect:set("target", value.target)
                    effect:set("skill_id", value.skill_id)
                    effect:set("skill_level", value.skill_level)
                    effect:set("refresh_delay", value.refresh_delay)
                    local source = assert(ecs.get(entity, "d2legacy.stat.source"), "aura effect has no stat source")
                    source:set("target", value.target)
                    source:set("source_id", value.source_id)
                    source:set("stat", value.stat)
                    source:set("operation", value.operation)
                    source:set("value", value.value)
                    source:set("order", 250)
                    existing[key] = nil
                else
                    if effect then
                        state_event(structural, "state_removed", context.tick, {
                            target = effect:get("target"),
                            state_id = effect:get("state_id"),
                            source_id = effect:get("source_id"),
                        }, "aura_replaced")
                        structural:destroy(entity)
                        existing[key] = nil
                    end
                    create_effect(structural, context, value)
                end
            end
            for _, key in ipairs(sorted_keys(existing)) do
                local entity = existing[key]
                if entity then
                    local effect = ecs.get(entity, "d2legacy.skill.aura_effect")
                    state_event(structural, "state_removed", context.tick, {
                        target = effect:get("target"),
                        state_id = effect:get("state_id"),
                        source_id = effect:get("source_id"),
                    }, "aura_left")
                    structural:destroy(entity)
                end
            end
        end,
    })
end

return M
