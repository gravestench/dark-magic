-- Turn one admitted skill request into a definition-driven timed cast.
--
-- Mana is stored in 8.8 fixed-point units: 256 means one visible mana point.
-- The cost is paid exactly once when a cast starts. An underfunded request is
-- rejected before the action exists and preserves the available mana. The
-- system then remembers effect and completion ticks so replay does not depend
-- on animation frames.

local ecs = require("engine.ecs/v1")
local skill_progression = require("d2legacy.policy.skill_progression")

local M = {}

local function learned_levels(entities)
    local levels = {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned then
            local owner = learned:get("owner"):id()
            levels[owner] = levels[owner] or {}
            levels[owner][learned:get("skill_id")] = learned:get("level")
        end
    end
    return levels
end

local function elemental_damage_percent(definition, levels)
    local total_levels = 0
    for _, skill_id in ipairs(definition.damage_synergy_skill_ids or {}) do
        total_levels = total_levels + (levels[skill_id] or 0)
    end
    return total_levels * (definition.damage_synergy_percent_per_level or 0)
end

local function effect_duration(definition, level, levels)
    local base =
        skill_progression.linear(definition.effect_duration_base or 0, definition.effect_duration_per_level or 0, level)
    local synergy_levels = 0
    for _, skill_id in ipairs(definition.effect_duration_synergy_skill_ids or {}) do
        synergy_levels = synergy_levels + (levels[skill_id] or 0)
    end
    local multiplier = 100 + synergy_levels * (definition.effect_duration_synergy_percent_per_level or 0)
    return math.floor(base * multiplier / 100)
end

local function begin_cast(context, player, request, definitions, levels, commands)
    local vitals = ecs.get(player, "d2legacy.player.vitals")
    local available = vitals:get("mana_raw")
    if available == 0 then
        available = vitals:get("mana") * 256
    end

    local player_levels = levels[player:id()] or {}
    local known_level = player_levels[request:get("skill_id")] or 0
    local definition = definitions[request:get("skill_id")]
    local valid = request:get("request_tick") <= context.tick
        and definition ~= nil
        and known_level > 0
        and available >= skill_progression.mana_cost(definition, known_level)

    if valid then
        local remaining = available - skill_progression.mana_cost(definition, known_level)
        vitals:set("mana_raw", remaining)
        vitals:set("mana", math.floor(remaining / 256))
        commands:set(player, "d2legacy.skill.cast", {
            skill_id = request:get("skill_id"),
            skill_level = known_level,
            target_x = request:get("target_x"),
            target_y = request:get("target_y"),
            target_id = request:get("target_id"),
            effect_tick = context.tick + definition.effect_delay,
            complete_tick = context.tick + definition.complete_delay,
            elemental_damage_percent = elemental_damage_percent(definition, player_levels),
            effect_duration_ticks = effect_duration(definition, known_level, player_levels),
            effect_emitted = false,
        })
    end
    commands:remove(player, "d2legacy.skill.cast_request")
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.skill.cast_lifecycle",
        phase = "pre_simulation",
        query = {
            any = {
                "d2legacy.skill.cast_request",
                "d2legacy.skill.cast",
                "d2legacy.player.learned_skill",
            },
            none = { "d2legacy.player.death" },
        },
        read = {
            "d2legacy.skill.cast_request",
            "d2legacy.skill.cast",
            "d2legacy.player.learned_skill",
            "d2legacy.player.vitals",
            "d2legacy.player.death",
        },
        write = {
            "d2legacy.skill.cast_request",
            "d2legacy.skill.cast",
            "d2legacy.player.vitals",
        },
        update = function(context, entities, structural)
            local levels = learned_levels(entities)
            for _, player in ipairs(entities) do
                local request = ecs.get(player, "d2legacy.skill.cast_request")
                local cast = ecs.get(player, "d2legacy.skill.cast")
                if request and not cast then
                    begin_cast(context, player, request, definitions, levels, structural)
                elseif cast and context.tick >= cast:get("complete_tick") then
                    structural:remove(player, "d2legacy.skill.cast")
                end
            end
        end,
    })
end

return M
