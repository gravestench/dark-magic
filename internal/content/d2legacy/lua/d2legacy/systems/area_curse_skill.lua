-- Materialize point-centered timed curse effects from admitted casts.

local ecs = require("engine.ecs/v1")
local area_targets = require("d2legacy.gameplay.area_targets")
local progression = require("d2legacy.policy.skill_progression")
local M = {}

local function target_stat_value(definition, target)
    local value = definition.stat_value
    local defense = ecs.get(target, "d2legacy.combat.defense")
    if defense and definition.immune_divisor > 0 and value < 0 and defense:get("base_physical_resist") >= 100 then
        return math.ceil(value / definition.immune_divisor)
    end
    return value
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.state.apply_area_curse_skill",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            any = { "d2legacy.skill.cast", "d2legacy.world.selectable" },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.player.identity",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "d2legacy.monster.stats",
            "d2legacy.combat.defense",
        },
        write = {
            "d2legacy.skill.cast",
            "d2legacy.state.request",
            "d2legacy.state.stat_request",
            "d2legacy.skill.cast_event",
        },
        update = function(context, entities, structural)
            for _, caster in ipairs(entities) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = cast and definitions[cast:get("skill_id")] or nil
                if
                    definition
                    and definition.behavior == "state.point-area-curse"
                    and not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick")
                then
                    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
                    local level = cast:get("skill_level")
                    local radius = progression.linear(definition.radius_base, definition.radius_per_level, level)
                    local duration = progression.linear(definition.duration_base, definition.duration_per_level, level)
                    local owner_source_id = "skill:" .. identity:get("player") .. ":" .. cast:get("skill_id")
                    for _, target in
                        ipairs(
                            area_targets.hostile_monsters(
                                entities,
                                caster,
                                cast:get("target_x"),
                                cast:get("target_y"),
                                radius
                            )
                        )
                    do
                        structural:create({
                            ["d2legacy.state.request"] = {
                                operation = "apply",
                                target = target.entity,
                                state_id = definition.state_id,
                                source_id = owner_source_id,
                                duration = duration,
                                policy = "refresh_same_source",
                                exclusive_group = definition.exclusive_group,
                                replacement_priority = level,
                                reject_lower_priority = definition.reject_lower_priority,
                            },
                        })
                        structural:create({
                            ["d2legacy.state.stat_request"] = {
                                target = target.entity,
                                owner_source_id = owner_source_id,
                                source_id = owner_source_id .. ":" .. definition.stat,
                                stat = definition.stat,
                                operation = definition.stat_operation,
                                value = target_stat_value(definition, target.entity),
                                order = 300,
                            },
                        })
                    end
                    structural:create({
                        ["d2legacy.skill.cast_event"] = {
                            kind = "skill_effect",
                            tick = context.tick,
                            player = identity:get("player"),
                            skill_id = cast:get("skill_id"),
                            skill_level = level,
                            behavior = definition.behavior,
                            target_x = cast:get("target_x"),
                            target_y = cast:get("target_y"),
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
