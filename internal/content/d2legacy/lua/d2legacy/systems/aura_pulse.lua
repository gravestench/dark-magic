-- Apply checkpointed direct effects for selected-aura target relationships.
--
-- Aura selection, party/radius eligibility, state arbitration, and visuals stay
-- in the shared selected-aura system. This consumer executes each due pulse's
-- authored operations and advances its durable schedule before doing any work,
-- so a repeated step cannot apply the same pulse twice.

local ecs = require("engine.ecs/v1")
local progression = require("d2legacy.policy.skill_progression")
local resources = require("d2legacy.policy.resources")
local M = {}

local function pulse_targets(entities, emitter)
    local result = {}
    local seen = {}
    for _, entity in ipairs(entities) do
        local effect = ecs.get(entity, "d2legacy.skill.aura_effect")
        if effect and effect:get("emitter"):id() == emitter:id() then
            local target = effect:get("target")
            local vitals = ecs.get(target, "d2legacy.player.vitals")
            if
                vitals
                and vitals:get("health") > 0
                and not ecs.get(target, "d2legacy.player.death")
                and not ecs.get(target, "d2legacy.world.inactive")
                and not seen[target:id()]
            then
                seen[target:id()] = true
                result[#result + 1] = target
            end
        end
    end
    table.sort(result, function(left, right)
        local left_identity = ecs.get(left, "d2legacy.player.identity")
        local right_identity = ecs.get(right, "d2legacy.player.identity")
        return left_identity:get("player") < right_identity:get("player")
    end)
    return result
end

local function pulse_effects(entities, emitter, source_id)
    local result = {}
    for _, entity in ipairs(entities) do
        local effect = ecs.get(entity, "d2legacy.skill.aura_pulse_effect")
        if effect and effect:get("emitter"):id() == emitter:id() and effect:get("source_id") == source_id then
            result[#result + 1] = effect
        end
    end
    table.sort(result, function(left, right)
        return left:get("order") < right:get("order")
    end)
    return result
end

local function heal_life(target, value)
    if value <= 0 then
        return false
    end
    local vitals = ecs.get(target, "d2legacy.player.vitals")
    local health = vitals:get("health")
    local healed = math.min(health + value, vitals:get("max_health"))
    if healed == health then
        return false
    end
    vitals:set("health", healed)
    return true
end

local function affected_timed_states(entities, target, state_policy)
    local result = {}
    for _, entity in ipairs(entities) do
        local instance = ecs.get(entity, "d2legacy.state.instance")
        if instance and instance:get("target"):id() == target:id() then
            local state_id = instance:get("state_id")
            if state_id == state_policy.poison_state_id or state_policy.duration_reduced_states[state_id] then
                result[#result + 1] = entity
            end
        end
    end
    table.sort(result, function(left, right)
        local left_state = ecs.get(left, "d2legacy.state.instance")
        local right_state = ecs.get(right, "d2legacy.state.instance")
        if left_state:get("state_id") ~= right_state:get("state_id") then
            return left_state:get("state_id") < right_state:get("state_id")
        end
        if left_state:get("source_id") ~= right_state:get("source_id") then
            return left_state:get("source_id") < right_state:get("source_id")
        end
        return left:id() < right:id()
    end)
    return result
end

local function scale_remaining_timed_state(entities, target, tick, percentage, state_policy)
    local changed = false
    for _, entity in ipairs(affected_timed_states(entities, target, state_policy)) do
        local instance = ecs.get(entity, "d2legacy.state.instance")
        local expires = instance:get("expires_tick")
        if expires > tick then
            local remaining = expires - tick
            local scaled = math.floor(remaining * percentage / 100)
            if tick + scaled ~= expires then
                instance:set("expires_tick", tick + scaled)
                changed = true
            end
        end
    end
    return changed
end

local function apply_effect(entities, target, tick, effect, state_policy)
    local kind = effect:get("kind")
    if kind == "heal_life" then
        return heal_life(target, effect:get("value"))
    end
    if kind == "scale_remaining_timed_state" then
        return scale_remaining_timed_state(entities, target, tick, effect:get("value"), state_policy)
    end
    error("unsupported aura pulse effect " .. kind)
end

local function execute_pulse(context, entities, emitter, pulse, state_policy)
    local owner_vitals = ecs.get(emitter, "d2legacy.player.vitals")
    if not owner_vitals then
        return false
    end
    local cost = pulse:get("mana_cost_raw")
    if resources.mana_raw(owner_vitals) < cost then
        return false
    end
    local effects = pulse_effects(entities, emitter, pulse:get("source_id"))
    assert(#effects > 0, "aura pulse has no composed effects")
    local changed = false
    for _, target in ipairs(pulse_targets(entities, emitter)) do
        for _, effect in ipairs(effects) do
            changed = apply_effect(entities, target, context.tick, effect, state_policy) or changed
        end
    end
    if changed and cost > 0 then
        assert(resources.spend_mana(owner_vitals, cost))
        return true
    end
    return false
end

local function reconcile_suppression(entities, emitter, source_id, desired, structural, removed)
    local existing
    for _, entity in ipairs(entities) do
        local suppression = ecs.get(entity, "d2legacy.resource.mana_regen_suppression")
        if suppression and not removed[entity:id()] and suppression:get("target"):id() == emitter:id() then
            if suppression:get("source_id") == source_id and not existing then
                existing = entity
            elseif suppression:get("source_id") == source_id or suppression:get("source_id"):match("^aura:") then
                structural:destroy(entity)
            end
        end
    end
    if desired and not existing then
        structural:create({
            ["d2legacy.resource.mana_regen_suppression"] = {
                target = emitter,
                source_id = source_id,
            },
        })
    elseif not desired and existing then
        structural:destroy(existing)
    end
end

local function remove_stale_suppressions(entities, structural)
    local removed = {}
    for _, entity in ipairs(entities) do
        local suppression = ecs.get(entity, "d2legacy.resource.mana_regen_suppression")
        if suppression and suppression:get("source_id"):match("^aura:") then
            local target = suppression:get("target")
            local pulse = ecs.get(target, "d2legacy.skill.aura_pulse")
            if not pulse or pulse:get("source_id") ~= suppression:get("source_id") then
                structural:destroy(entity)
                removed[entity:id()] = true
            end
        end
    end
    return removed
end

function M.register(state_policy)
    assert(state_policy and state_policy.poison_state_id, "periodic aura state policy is required")
    ecs.system({
        id = "d2legacy.skill.aura_periodic_effect",
        phase = "pre_simulation",
        after = { "d2legacy.skill.selected_party_aura" },
        query = {
            any = {
                "d2legacy.skill.aura_pulse",
                "d2legacy.skill.aura_pulse_effect",
                "d2legacy.skill.aura_effect",
                "d2legacy.player.vitals",
                "d2legacy.state.instance",
                "d2legacy.resource.mana_regen_suppression",
            },
        },
        read = {
            "d2legacy.skill.aura_pulse_effect",
            "d2legacy.skill.aura_effect",
            "d2legacy.player.identity",
            "d2legacy.player.death",
            "d2legacy.world.inactive",
            "d2legacy.resource.mana_regen_suppression",
        },
        write = {
            "d2legacy.skill.aura_pulse",
            "d2legacy.player.vitals",
            "d2legacy.state.instance",
            "d2legacy.resource.mana_regen_suppression",
        },
        update = function(context, entities, structural)
            local removed = remove_stale_suppressions(entities, structural)
            for _, emitter in ipairs(entities) do
                local pulse = ecs.get(emitter, "d2legacy.skill.aura_pulse")
                if pulse and context.tick >= pulse:get("next_tick") then
                    pulse:set("next_tick", progression.next_periodic_tick(context.tick, pulse:get("period_ticks")))
                    local suppress = execute_pulse(context, entities, emitter, pulse, state_policy)
                    reconcile_suppression(entities, emitter, pulse:get("source_id"), suppress, structural, removed)
                end
            end
        end,
    })
end

return M
