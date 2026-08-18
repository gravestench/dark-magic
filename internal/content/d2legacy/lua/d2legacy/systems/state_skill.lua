-- Materialize targetless timed-state/stat effects from generic admitted casts.

local ecs = require("engine.ecs/v1")
local progression = require("d2legacy.policy.skill_progression")
local M = {}

local function learned_levels(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned then
            local owner = learned:get("owner"):id()
            result[owner] = result[owner] or { effective = {}, hard = {} }
            local skill_id = learned:get("skill_id")
            result[owner].effective[skill_id] = learned:get("level")
            local hard = ecs.get(entity, "d2legacy.player.skill_hard_level")
            result[owner].hard[skill_id] = hard and hard:get("level") or learned:get("level")
        end
    end
    return result
end

local function level_value(base, per_level, level)
    return base + math.max(level - 1, 0) * per_level
end

local function synergy_levels(skill_ids, levels)
    local result = 0
    for _, skill_id in ipairs(skill_ids or {}) do
        result = result + (levels[skill_id] or 0)
    end
    return result
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.state.apply_skill",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            any = {
                "d2legacy.skill.cast",
                "d2legacy.player.learned_skill",
                "d2legacy.player.skill_hard_level",
            },
            none = { "d2legacy.player.death" },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.player.learned_skill",
            "d2legacy.player.skill_hard_level",
            "d2legacy.player.identity",
        },
        write = { "d2legacy.skill.cast", "d2legacy.state.request", "d2legacy.skill.cast_event" },
        update = function(context, entities, structural)
            local levels_by_owner = learned_levels(entities)
            for _, caster in ipairs(entities) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = cast and definitions[cast:get("skill_id")] or nil
                if
                    definition
                    and definition.behavior == "state.self-timed-reactive"
                    and not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick")
                then
                    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
                    local levels = levels_by_owner[caster:id()] or { effective = {}, hard = {} }
                    local skill_level = cast:get("skill_level")
                    local duration = level_value(definition.duration_base, definition.duration_per_level, skill_level)
                        + synergy_levels(definition.duration_synergy_skill_ids, levels.hard)
                            * definition.duration_synergy_per_level
                    local reaction_duration = progression.banded(
                        definition.reaction_duration_base,
                        definition.reaction_duration_per_level,
                        skill_level
                    )
                    reaction_duration = math.floor(
                        reaction_duration
                            * (100 + synergy_levels(definition.reaction_duration_synergy_skill_ids, levels.hard) * definition.reaction_duration_synergy_percent)
                            / 100
                    )
                    local minimum_damage, maximum_damage =
                        progression.damage_range(definition, skill_level, cast:get("elemental_damage_percent"))
                    local source_id = "skill:" .. identity:get("player") .. ":" .. cast:get("skill_id")
                    structural:create({
                        ["d2legacy.state.request"] = {
                            operation = "apply",
                            target = caster,
                            state_id = definition.state_id,
                            source_id = source_id,
                            duration = duration,
                            policy = "refresh_same_source",
                            stat = definition.stat,
                            stat_operation = definition.stat_operation,
                            stat_value = level_value(definition.stat_base, definition.stat_per_level, skill_level),
                            stat_order = 300,
                            exclusive_group = definition.exclusive_group,
                            reaction = definition.reaction,
                            reaction_skill_id = definition.skill_id,
                            reaction_state_id = definition.reaction_state_id,
                            reaction_chill_state_id = definition.reaction_chill_state_id,
                            reaction_stat = definition.reaction_stat,
                            reaction_stat_value = definition.reaction_stat_value,
                            reaction_chill_stat = definition.reaction_chill_stat,
                            reaction_chill_stat_value = definition.reaction_chill_stat_value,
                            reaction_duration = reaction_duration,
                            reaction_disables_action = definition.reaction_disables_action,
                            reaction_minimum_damage_raw = minimum_damage,
                            reaction_maximum_damage_raw = maximum_damage,
                        },
                    })
                    structural:create({
                        ["d2legacy.skill.cast_event"] = {
                            kind = "skill_effect",
                            tick = context.tick,
                            player = identity:get("player"),
                            skill_id = cast:get("skill_id"),
                            skill_level = skill_level,
                            behavior = definition.behavior,
                            target_x = 0,
                            target_y = 0,
                            target_id = "",
                            reason = "",
                        },
                    })
                    cast:set("effect_emitted", true)
                end
            end
        end,
    })
end

return M
