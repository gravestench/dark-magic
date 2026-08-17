-- Materialize targetless timed-state/stat effects from generic admitted casts.

local ecs = require("engine.ecs/v1")
local M = {}

local function learned_levels(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned then
            local owner = learned:get("owner"):id()
            result[owner] = result[owner] or {}
            result[owner][learned:get("skill_id")] = learned:get("level")
        end
    end
    return result
end

local function level_value(base, per_level, level)
    return base + math.max(level - 1, 0) * per_level
end

local function synergy_levels(definition, levels)
    local result = 0
    for _, skill_id in ipairs(definition.duration_synergy_skill_ids) do
        result = result + (levels[skill_id] or 0)
    end
    return result
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.state.apply_skill",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = { any = { "d2legacy.skill.cast", "d2legacy.player.learned_skill" } },
        read = { "d2legacy.skill.cast", "d2legacy.player.learned_skill", "d2legacy.player.identity" },
        write = { "d2legacy.skill.cast", "d2legacy.state.request", "d2legacy.skill.cast_event" },
        update = function(context, entities, structural)
            local levels_by_owner = learned_levels(entities)
            for _, caster in ipairs(entities) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = cast and definitions[cast:get("skill_id")] or nil
                if
                    definition
                    and definition.behavior == "state.self-timed-stat"
                    and not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick")
                then
                    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
                    local levels = levels_by_owner[caster:id()] or {}
                    local skill_level = cast:get("skill_level")
                    local duration = level_value(definition.duration_base, definition.duration_per_level, skill_level)
                        + synergy_levels(definition, levels) * definition.duration_synergy_per_level
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
